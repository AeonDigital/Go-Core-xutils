package xfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// PrintCLITable renders architectural string matrices into an aligned column grid directly to standard output.
// It leverages the low-level standard 'text/tabwriter' engine to format structural content by intercepting
// horizontal tab delimiters ('\t') and injecting dynamic runtime padding spacing based on a column-widest scanning rule.
//
// Because this function acts as the final terminal serialization barrier, it expects the upstream data layers
// to have already calculated structural word wrapping, multi-line row string equalization, and responsive
// column-dropping cascades. It forces horizontal column boundary alignment by verifying that every data row's
// structural layout exact match constraint complies with the total header slice size.
//
// Parameters:
//   - headers: The pre-calculated slice of visible column header label strings that survived responsive pruning.
//   - rows: The completely structured matrix of processed cell strings, already formatted via truncation or word wrap.
func PrintCLITable(headers []string, rows [][]string) {
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

// PrintObjectAsJSON marshals any structured data object into an indented, pretty-printed JSON string sequence.
// It applies a strict 2-space padding constraint ("  ") for nested properties to guarantee clean visual
// alignment when rendering serialized records within console output interfaces.
//
// To circumvent Go's compile-time static struct encoding constraints, this function pipelines serialization
// through a dynamic interception layer. It translates the object into an intermediary 'map[string]any',
// allowing the runtime to mutate the underlying property matrix and cleanly purge blacklisted keys before
// serializing the final visual payload.
//
// Parameters:
//   - obj: The source entity, collection, or structured data block to be serialized.
//   - ignoreFields: A collection of field identifier strings to be excluded from the final payload.
//     Crucially, these tokens must match the literal keys defined inside the struct's `json:"..."` tags
//     (e.g., "scenario_prompt" or "description") rather than the static Go struct field identifiers.
//   - pretty: If true, it applies a strict 2-space padding constraint ("  ") for nested properties; otherwise,
//     it outputs a single-line inline compact layout structure.
//
// Returns:
//   - A formatted, 2-space indented JSON block string stripped of any trailing newline noise.
//   - If any phase of the serialization or mapping pipeline encounters an encoding constraint violation,
//     it bypasses panic states to return a safe, formatted error token: "[ERROR] :: <message>".
func PrintObjectAsJSON(obj any, ignoreFields []string, pretty bool) string {
	var err error
	var bytesData []byte

	bytesData, err = json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("[ERROR] :: %s", err.Error())
	}

	// 1. Convert initial object matrix into a dynamic generic map map[string]any
	var rawData map[string]any
	err = json.Unmarshal(bytesData, &rawData)
	if err != nil {
		return fmt.Sprintf("[ERROR] :: %s", err.Error())
	}

	// 2. Intercept and purge unwanted property keys provided in the filter parameter
	for _, field := range ignoreFields {
		delete(rawData, field)
	}

	// 3. Serialize the filtered map matrix into a JSON string block
	var buf bytes.Buffer
	err = jsonPrettyEncoderExecute(&buf, rawData, pretty)
	if err != nil {
		return fmt.Sprintf("[ERROR] :: %s", err.Error())
	}

	return strings.TrimSpace(buf.String())
}

var jsonPrettyEncoderExecute = func(w io.Writer, v any, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(v)
}
