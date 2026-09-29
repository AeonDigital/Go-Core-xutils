package xfmt_test

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xfmt/pkg/xfmt"
)

func TestCliPrintTableCalculateColumnWidths(t *testing.T) {
	tests := []struct {
		name           string
		columns        []xfmt.CLIPrintTableColumnRule
		data           [][]string
		availableWidth int
		wantWidths     []int
		wantActive     []bool
	}{
		{
			name: "fixed columns clamped by maximum content length",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "ID", Width: 5},      // Content "10" (len 2) < Width (5) -> Allocates 2
				{Header: "Status", Width: 10}, // Header "Status" (len 6) > Content "OK" (len 2) -> Allocates 6
			},
			data: [][]string{
				{"10", "OK"},
			},
			availableWidth: 50,
			wantWidths:     []int{2, 6},
			wantActive:     []bool{true, true},
		},
		{
			name: "fixed columns clamped by their structural maximum limit",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Code", Width: 3}, // Content "12345" (len 5) > Width (3) -> Clamped to 3
			},
			data: [][]string{
				{"12345"},
			},
			availableWidth: 30,
			wantWidths:     []int{3},
			wantActive:     []bool{true},
		},
		{
			name: "dynamic columns sharing remaining space evenly with remainder on first",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "A", Width: 0},
				{Header: "B", Width: 5}, // Data is "y" (len 1). min(1, 5) = 1. Spacing per col (2) * 3 = 6. Total fixed cost = 7.
				{Header: "C", Width: 0},
			},
			data: [][]string{
				{"x", "y", "z"},
			},
			// Total space = 20. Remaining for dynamic = 20 - 7 = 13.
			// 13 / 2 columns = 6 base width. 13 % 2 = 1 remainder.
			// Column A (first dynamic): 6 + 1 = 7. Column B: 1. Column C: 6.
			availableWidth: 20,
			wantWidths:     []int{7, 1, 6},
			wantActive:     []bool{true, true, true},
		},
		{
			name: "trigger responsive cascade pruning dynamic columns first from right to left",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Fix1", Width: 10}, // Content "val" -> max(4,3)=4. min(4,10)=4. Cost = 4 + 2 = 6.
				{Header: "Dyn1", Width: 0},
				{Header: "Dyn2", Width: 0},
			},
			data: [][]string{
				{"val", "val", "val"},
			},
			// Spacing cost for 3 cols = 6. Fixed content cost = 4. Total minimal cost to exist without pruning = 10 + 2 (dynamic spaces) = 12.
			// Setting availableWidth to 7 forces immediate pruning of dynamic columns from right to left.
			availableWidth: 7,
			wantWidths:     []int{4, 0, 0},
			wantActive:     []bool{true, false, false},
		},
		{
			name: "trigger responsive cascade pruning fixed columns when no dynamic columns remain",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Fix1", Width: 10},
				{Header: "Fix2", Width: 10},
			},
			data: [][]string{
				{"data1", "data2"},
			},
			// Two fixed columns need at least (Header "Fix1"=4 + spacing=2) + (Header "Fix2"=4 + spacing=2) = 12 width.
			// Setting to 8 forces the rightmost fixed column ("Fix2") to be pruned.
			availableWidth: 8,
			wantWidths:     []int{5, 0}, // "data1" len is 5. max("Fix1", "data1") = 5. min(5,10) = 5.
			wantActive:     []bool{true, false},
		},
		{
			name: "absolute safety check fallback when terminal space cannot host any column layout",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "CriticalField", Width: 20},
			},
			data: [][]string{
				{"criticalerrorpayload"},
			},
			availableWidth: 1, // Impossible width layout constraint
			wantWidths:     []int{0},
			wantActive:     []bool{false},
		},
		{
			name: "absolute safety check constraint bypass with negative terminal size",
			columns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Fix", Width: 10},
				{Header: "Dyn", Width: 0},
			},
			data: [][]string{
				{"val1", "val2"},
			},
			availableWidth: -1, // Força remainingWidth a ser menor que zero mesmo após desativar tudo
			wantWidths:     []int{0, 0},
			wantActive:     []bool{false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidths, gotActive := xfmt.ExportCalcColumnWidths(tt.columns, tt.data, tt.availableWidth)

			if !reflect.DeepEqual(gotWidths, tt.wantWidths) {
				t.Errorf("%s: gotWidths = %v, want %v", tt.name, gotWidths, tt.wantWidths)
			}

			if !reflect.DeepEqual(gotActive, tt.wantActive) {
				t.Errorf("%s: gotActive = %v, want %v", tt.name, gotActive, tt.wantActive)
			}
		})
	}
}

func TestCliPrintTableWrapString(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{
			name:  "invalid or zero width layout constraint",
			text:  "should return empty slice segment",
			width: 0,
			want:  []string{""},
		},
		{
			name:  "empty or whitespace-only input string",
			text:  "   \t\n   ",
			width: 10,
			want:  []string{""},
		},
		{
			name:  "clean standard word wrapping on space boundaries",
			text:  "golang xfmt toolchain layout",
			width: 12,
			want:  []string{"golang xfmt", "toolchain", "layout"},
		},
		{
			name:  "monolithic single word exceeding maximum width boundary forcefully sliced",
			text:  "unknwownerrorpayload",
			width: 5,
			want:  []string{"unknw", "owner", "rorpa", "yload"},
		},
		{
			name:  "force slice long word when current layout line already contains pending tokens",
			text:  "go supercalifragilistic",
			width: 6,
			// "go" entra na linha 1. "superc..." é longo demais e força o flush de "go".
			// Em seguida, "supercalifragilistic" é fatiado em blocos de 6 caracteres.
			want: []string{"go", "superc", "alifra", "gilist", "ic"},
		},
		{
			name:  "word fits column width but exceeds remaining horizontal space cost",
			text:  "wrap text now",
			width: 6,
			// "wrap" (len 4). Próxima palavra: "text" (len 4).
			// "wrap" + 1 (espaço) + "text" = 9 > 6. Força "text" para a linha seguinte.
			want: []string{"wrap", "text", "now"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xfmt.ExportWrapString(tt.text, tt.width)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCliPrintTableBuildTableRows(t *testing.T) {
	tests := []struct {
		name              string
		originalColumns   []xfmt.CLIPrintTableColumnRule
		widths            []int
		data              [][]string
		minMultilineWidth int
		terminalWidth     int
		want              [][]string
	}{
		{
			name: "skip pruned columns and handle missing cell data gracefully",
			originalColumns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Active", Width: 5},
				{Header: "Pruned", Width: 0},
				{Header: "Missing", Width: 5},
			},
			widths: []int{5, 0, 5}, // Coluna do meio podada (width = 0)
			data: [][]string{
				{"OK"}, // Falta o dado da coluna 'Missing', deve virar ""
			},
			minMultilineWidth: 100,
			terminalWidth:     50, // Ativa Truncamento simples
			want: [][]string{
				{"OK", ""}, // Apenas as colunas com width > 0 aparecem
			},
		},
		{
			name: "strategy A - truncation mode when terminal is narrow",
			originalColumns: []xfmt.CLIPrintTableColumnRule{
				{Header: "FixedCol", Width: 5}, // Fixed -> Raw cut
				{Header: "DynCol", Width: 0},   // Dynamic -> Ellipsis cut
			},
			widths: []int{5, 8},
			data: [][]string{
				{"LongFixedText", "LongDynamicText"},
			},
			minMultilineWidth: 100,
			terminalWidth:     50, // terminalWidth <= minMultilineWidth -> Truncamento
			want: [][]string{
				{"LongF", "LongD..."},
			},
		},
		{
			name: "strategy B - multiline wrap mode with cell height equalization",
			originalColumns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Col1", Width: 5},
				{Header: "Col2", Width: 5},
			},
			widths: []int{5, 5},
			data: [][]string{
				{"Short", "Very Long Text"},
			},
			minMultilineWidth: 10,
			terminalWidth:     50, // terminalWidth > minMultilineWidth -> Multilinha
			// "Very Long Text" vira ["Very", "Long", "Text"] (3 linhas)
			// "Short" deve ser inflado com strings vazias nas linhas de baixo
			want: [][]string{
				{"Short", "Very"},
				{"", "Long"},
				{"", "Text"},
			},
		},
		{
			name: "force truncate override during multiline execution strategy",
			originalColumns: []xfmt.CLIPrintTableColumnRule{
				{Header: "Normal", Width: 5, ForceTruncate: false},
				{Header: "Forced", Width: 5, ForceTruncate: true}, // Deve truncar mesmo em multiline
			},
			widths: []int{5, 5},
			data: [][]string{
				{"Wrap Text Here", "Truncate This Field"},
			},
			minMultilineWidth: 10,
			terminalWidth:     50, // Cenário propício para multilinha
			// O comportamento esperado depende se a coluna forçada é considerada dinâmica (Width=0) ou fixa no modelo original.
			// Na nossa struct originalColumns[1].Width é 5 (Fixa), mas ForceTruncate é true.
			// O seu algoritmo faz: useEllipsis := originalColumns[colIdx].Width == 0 || originalColumns[colIdx].ForceTruncate
			// Como ForceTruncate é true, useEllipsis será true -> "Trunc..." (len 8, mas width é 5, vira "Tr...")
			want: [][]string{
				{"Wrap", "Tr..."},
				{"Text", ""},
				{"Here", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xfmt.ExportBuildTableRows(tt.originalColumns, tt.widths, tt.data, tt.minMultilineWidth, tt.terminalWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

//
//
//

// mockCLIObject implements the CLIPrintTableObject interface contract for integration testing.
type mockCLIObject struct {
	cells []string
	rules []xfmt.CLIPrintTableColumnRule
}

func (m mockCLIObject) ToStringSlice() []string {
	return m.cells
}

func (m mockCLIObject) GetColumnRules() []xfmt.CLIPrintTableColumnRule {
	return m.rules
}

func TestPrintCLITableAndCLIPrintTable(t *testing.T) {
	t.Run("PrintCLITable early return on empty headers", func(t *testing.T) {
		// Capture stdout to verify nothing gets written during empty header protection path
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		xfmt.PrintCLITable(nil, [][]string{{"data"}})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if buf.Len() > 0 {
			t.Errorf("expected no output on empty headers early return, got %q", buf.String())
		}
	})

	t.Run("PrintCLITable clean grid alignment and formatting", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		headers := []string{"ID", "Name"}
		rows := [][]string{
			{"1", "Alice"},
			{"2"}, // Malformed row (different column length), should be safely skipped by validation check
		}

		xfmt.PrintCLITable(headers, rows)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "ID") || !strings.Contains(output, "Alice") {
			t.Errorf("expected output to contain structured matrix tokens, got %q", output)
		}
		if strings.Contains(output, "2") {
			t.Errorf("expected row with mismatched column count to be completely discarded, got %q", output)
		}
	})

	t.Run("CLIPrintTable orchestrator early return on empty data slices", func(t *testing.T) {
		var emptyDataset []mockCLIObject
		// Should execute safely without panic or output triggering
		xfmt.CLIPrintTable(emptyDataset, 120)
	})

	t.Run("CLIPrintTable successful full layout execution flow", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		dataset := []mockCLIObject{
			{
				cells: []string{"100", "System Engineering Service Deployment"},
				rules: []xfmt.CLIPrintTableColumnRule{
					{Header: "Code", Width: 5},
					{Header: "Description", Width: 0},
				},
			},
		}

		// Trigger orchestration layer (will calculate widths, split rows, and call PrintCLITable)
		xfmt.CLIPrintTable(dataset, 120)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "Code") || !strings.Contains(output, "System Engineering") {
			t.Errorf("expected compiled matrix output from orchestrator payload, got %q", output)
		}
	})
}
