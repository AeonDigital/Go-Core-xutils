package xfmt

import "strings"

// CLIPrintTableObject defines the contract for any data model that can be
// extracted, structured, and rendered dynamically into a command-line interface (CLI) table layout.
type CLIPrintTableObject interface {
	// ToStringSlice converts the internal object fields into a raw row string sequence.
	ToStringSlice() []string

	// GetColumnRules returns the layout schema definitions and sizing constraints for each column.
	GetColumnRules() []CLIPrintTableColumnRule
}

// CLIPrintTableColumnRule outlines the structural schema constraints for a single grid table column.
// It maps identity bounds and dictates responsive width behaviors during terminal layout rendering.
type CLIPrintTableColumnRule struct {
	// Header specifies the visible text label displayed at the top of the column matrix.
	Header string

	// Width dictates the targeted layout character width allocation capacity constraint.
	// A value of '0' defines that this column acts as a dynamic/flexible block.
	Width int

	// ForceTruncate commands the rendering engine to always execute a single-line truncation
	// strategy with trailing ellipsis indicators for this specific field column block.
	// When true, it prevents the layout engine from exploding the cell content vertically,
	// even when the active terminal real estate switches over to multi-line Word Wrap mode.
	ForceTruncate bool
}

// cliPrintTableCalculateColumnWidths evaluates data matrices alongside structural column definitions
// to compute the precise grid width allocation for each column. It natively handles fixed constraints,
// evenly distributes available whitespace to dynamic columns (Width = 0), appends arithmetic remainders
// to the first dynamic column, and triggers a backward-purging adaptive responsive pruning cascade
// (from right to left) if the cumulative required structural boundary overflows the total terminal screen space.
//
// Parameters:
//   - columns: The ordered slice of CLIPrintTableColumnRule metadata definitions.
//   - data: The raw matrix rows representing cell strings.
//   - availableWidth: The maximum horizontal screen real estate constraint (terminal character width).
//
// Returns:
//   - allocatedWidths: An integer slice indicating the calculated character width for each active column.
//   - activeColumns: A boolean slice where true marks that the column at that index survived responsive pruning.
func cliPrintTableCalculateColumnWidths(
	columns []CLIPrintTableColumnRule,
	data [][]string,
	availableWidth int,
) ([]int, []bool) {
	numCols := len(columns)
	allocatedWidths := make([]int, numCols)
	activeColumns := make([]bool, numCols)

	// Assume all columns are active initially
	for i := range activeColumns {
		activeColumns[i] = true
	}

	// Internal loop for adaptive pruning simulation cascade
	for {
		totalFixedAndSpacing := 0
		dynamicColsCount := 0
		spacingPerColumn := 2 // Margin padding cost assumed per column layout by xfmt delimiters

		// Step 1: Compute maximum content overhead for active fixed-size fields
		for i := range numCols {
			if !activeColumns[i] {
				continue
			}

			totalFixedAndSpacing += spacingPerColumn
			if columns[i].Width > 0 {
				// Detect longest cell token in current column space
				maxContentLen := len(columns[i].Header)
				for _, row := range data {
					if i < len(row) && len(row[i]) > maxContentLen {
						maxContentLen = len(row[i])
					}
				}

				// Clamp to maximum constraint if lower than content data
				allocatedWidths[i] = min(maxContentLen, columns[i].Width)
				totalFixedAndSpacing += allocatedWidths[i]
			} else {
				dynamicColsCount++
			}
		}

		// Step 2: Validate boundary overflow or execute dynamic layout allocation matrix
		remainingWidth := availableWidth - totalFixedAndSpacing

		if remainingWidth >= dynamicColsCount {
			if dynamicColsCount > 0 {
				baseDynamicWidth := remainingWidth / dynamicColsCount
				remainder := remainingWidth % dynamicColsCount

				firstDynamicFound := false
				for i := range numCols {
					if !activeColumns[i] || columns[i].Width > 0 {
						continue
					}

					allocatedWidths[i] = baseDynamicWidth
					if !firstDynamicFound {
						allocatedWidths[i] += remainder
						firstDynamicFound = true
					}
				}
			}
			break // System allocation fits, successfully escaping simulation cascade loops
		}

		// Step 3: Trigger active fallback cascade strategy (Backward-purging pruning rule)
		// Prunes dynamic columns first, then fixed columns from right-to-left
		pruned := false
		for i := numCols - 1; i >= 0; i-- {
			if activeColumns[i] && columns[i].Width == 0 {
				activeColumns[i] = false
				allocatedWidths[i] = 0
				pruned = true
				break
			}
		}

		if !pruned {
			for i := numCols - 1; i >= 0; i-- {
				if activeColumns[i] {
					activeColumns[i] = false
					allocatedWidths[i] = 0
					pruned = true
					break
				}
			}
		}

		// Absolute safety check: If no elements can be evaluated, break layout constraints safely
		if !pruned {
			break
		}
	}

	return allocatedWidths, activeColumns
}

// cliPrintTableWrapString splits a continuous string block into a vertical slice of strings (lines),
// ensuring no individual line segment exceeds the maximum column width allocation constraint.
// It scans for space boundaries to perform clean whole-word wrapping whenever possible.
// If a single word is larger than the maximum width constraint, it slices the word forcefully.
func cliPrintTableWrapString(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		wordRunes := []rune(word)

		// If a single word is inherently wider than the entire column limit,
		// force-slice the word into sequential chunks
		if len(wordRunes) > width {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			for len(wordRunes) > width {
				lines = append(lines, string(wordRunes[:width]))
				wordRunes = wordRunes[width:]
			}
			word = string(wordRunes)
			wordRunes = []rune(word)
		}

		// Evaluate padding addition cost
		spaceCost := 0
		if currentLine.Len() > 0 {
			spaceCost = 1
		}

		if currentLine.Len()+spaceCost+len(wordRunes) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// cliPrintTableBuildTableRows orchestrates the data transformation layer, formatting raw matrix cells
// into a standardized tabular layout grid structure. It dynamically toggles execution strategies:
//   - Under the Truncation strategy (terminalWidth <= minMultilineWidth): cells are single-line,
//     applying raw cuts for fixed columns and ellipsis cuts for dynamic columns.
//   - Under the Multilinha strategy (terminalWidth > minMultilineWidth): long text is split vertically
//     via word wrapping, and shorter sibling cells are inflated with empty vertical space to match row height.
//   - If ForceTruncate is explicitly active on a column definition, the Multilinha strategy is bypassed,
//     forcing a single-line truncation with trailing ellipsis to preserve vertical screen workspace.
//
// Parameters:
//   - originalColumns: The baseline slice of original CLIPrintTableColumnRule definitions to check for dynamic sizing and truncation constraints.
//   - widths: The calculated exact character width limits matching the active headers.
//   - data: The source collection rows of unformatted data slices.
//   - minMultilineWidth: The configurable terminal width threshold (e.g., 120) that triggers Word Wrap.
//   - terminalWidth: The current measured screen width of the user console.
//
// Returns:
//   - formattedRows: A compiled matrix of strings ready for direct grid printing.
func cliPrintTableBuildTableRows(
	originalColumns []CLIPrintTableColumnRule,
	widths []int,
	data [][]string,
	minMultilineWidth int,
	terminalWidth int,
) [][]string {
	var formattedRows [][]string

	// Active strategy selection: evaluate if terminal real estate triggers multiline wrapping
	useMultiline := terminalWidth > minMultilineWidth

	for _, row := range data {
		cellLinesMap := make([][]string, len(widths))
		maxRowLines := 1

		for colIdx, width := range widths {
			if width == 0 {
				continue // Skip pruned columns
			}

			cellText := ""
			if colIdx < len(row) {
				cellText = row[colIdx]
			}

			// Intercept force truncate constraint to override multi-line strategy on this specific column
			if useMultiline && !originalColumns[colIdx].ForceTruncate {
				// Strategy B: Wrap text into vertical multi-line slices
				lines := cliPrintTableWrapString(cellText, width)
				cellLinesMap[colIdx] = lines
				if len(lines) > maxRowLines {
					maxRowLines = len(lines)
				}
			} else {
				// Strategy A: Truncate cell text into a single monolithic line block
				// Flexible columns or force-truncated columns receive trailing ellipsis (...)
				useEllipsis := originalColumns[colIdx].Width == 0 || originalColumns[colIdx].ForceTruncate

				truncated := TruncateString(cellText, width, useEllipsis)
				cellLinesMap[colIdx] = []string{truncated}
			}
		}

		// Equalize row blocks by inflating short cells with vertical space padding
		for lineIdx := range maxRowLines {
			compiledRowLine := make([]string, 0, len(widths))

			for colIdx, width := range widths {
				if width == 0 {
					continue // Ignore pruned fields
				}

				lines := cellLinesMap[colIdx]
				if lineIdx < len(lines) {
					compiledRowLine = append(compiledRowLine, lines[lineIdx])
				} else {
					compiledRowLine = append(compiledRowLine, "")
				}
			}

			formattedRows = append(formattedRows, compiledRowLine)
		}
	}

	return formattedRows
}

// CLIPrintTable orchestrates the comprehensive responsive rendering pipeline for structural data.
// It queries terminal hardware dimensions dynamically, extracts column properties from the
// CLIPrintTableObject layout metadata, executes pruning calculations, and structures text into
// truncated or wrapped tabular configurations based on a customizable terminal width constraint.
// Finally, it passes the organized header slice and rows matrix directly into the underlying print engine.
//
// Parameters:
//   - data: A slice of entities satisfying the modern CLIPrintTableObject interface contract.
//   - minMultilineWidth: The customizable terminal threshold (e.g., 120 chars) governing layout behavior.
func CLIPrintTable[T CLIPrintTableObject](data []T, minMultilineWidth int) {
	if len(data) == 0 {
		return
	}

	// Step 1: Initialize metadata boundaries and capture current hardware real estate
	originalColumns := data[0].GetColumnRules()
	terminalWidth := GetTerminalWidth()

	// Step 2: Extract and map raw unformatted records to matrix strings arrays
	rawMatrix := make([][]string, 0, len(data))
	for _, reg := range data {
		rawMatrix = append(rawMatrix, reg.ToStringSlice())
	}

	// Step 3: Run the mathematical distribution and backward pruning allocation cascade
	calculatedWidths, activeColumnsMask := cliPrintTableCalculateColumnWidths(originalColumns, rawMatrix, terminalWidth)

	// Step 4: Extract active headers that successfully survived the pruning stage
	var activeHeaders []string
	for idx, active := range activeColumnsMask {
		if active {
			activeHeaders = append(activeHeaders, originalColumns[idx].Header)
		}
	}

	// Step 5: Process and format grid cells using the custom multiline configuration limit
	formattedTableBody := cliPrintTableBuildTableRows(originalColumns, calculatedWidths, rawMatrix, minMultilineWidth, terminalWidth)

	// Step 6: Dispatch the standardized matrix payload directly into the print engine
	PrintCLITable(
		activeHeaders,
		formattedTableBody,
	)
}
