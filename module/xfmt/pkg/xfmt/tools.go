package xfmt

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintLog writes a message to the standard output.
// If no extra arguments (args) are provided, it behaves like fmt.Println, printing the message directly.
// If extra arguments are provided, it treats the message as a format string and behaves like fmt.Printf, automatically appending a newline character.
func Print(message string, args ...any) {
	if len(args) == 0 {
		// Ensures a clean, predictable line-break output layout without complex formatting overhead
		fmt.Fprintln(os.Stdout, message)
		return
	}

	// Normalizes trailing boundary spacing dynamically to prevent double-newline visual distortions
	format := message
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stdout, format, args...)
}

// PrintInline writes a message to the standard output.
// Unlike Print, this function does NOT automatically append a newline.
// The caller is responsible for including "\n" in the message if desired.
func PrintInline(message string, args ...any) {
	if len(args) == 0 {
		// When no variadic arguments are provided, print the message directly.
		// fmt.Fprint writes the string exactly as given, without adding a newline.
		fmt.Fprint(os.Stdout, message)
		return
	}

	// When variadic arguments are provided, treat 'message' as a format string.
	// fmt.Fprintf writes formatted output to os.Stdout without forcing a newline.
	fmt.Fprintf(os.Stdout, message, args...)
}

// PrintAsTable Prints data formatted in columns to the terminal based on the provided headers.
func PrintAsTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	tabW := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer tabW.Flush()

	format := strings.Repeat("%s\t", len(headers)-1) + "%s\n"

	// Converts headers from []string to []any for Fprintf
	headerArgs := make([]any, len(headers))
	for i, v := range headers {
		headerArgs[i] = v
	}
	fmt.Fprintf(tabW, format, headerArgs...)

	// Prints each line of data
	for _, row := range rows {
		// Ensures the line has the same number of columns as the format expects
		if len(row) != len(headers) {
			continue
		}

		rowArgs := make([]any, len(row))
		for i, v := range row {
			rowArgs[i] = v
		}

		fmt.Fprintf(tabW, format, rowArgs...)
	}
}
