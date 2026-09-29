package xfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

// PrintMapAsJSON serializes a dynamic 'map[string]any' layout model into a clean JSON string sequence
// enforcing an explicit structural key positioning sequence dictated by the orderKeys parameters.
//
// To circumvent Go's runtime behavior which randomizes map key traversal, this function acts as an
// ordered stream serializer. It maps out prioritary keys found within orderKeys first, appends all
// remaining keys sorted alphabetically to maintain output determinism, and structures the final text.
// If pretty is true, it builds vertical whitespace padding alignment based on the provided indent sequence
// (defaulting to 2 spaces when left blank). If pretty is false, it outputs a single-line compact inline stream.
//
// Parameters:
//   - obj: The raw map matrix containing key-value data nodes to be ordered and formatted.
//   - orderKeys: An ordered collection of string tokens defining which properties must be rendered first at the top.
//   - pretty: Flag governing whether to apply vertical whitespace structural alignment and line breaks.
//   - indent: The literal string pattern used for line padding constraints (e.g., "\t", "    ", "  ").
//
// Returns:
//   - A clean ordered JSON string output stripped of any outer edge or trailing newline noise characters.
//   - If any object embedded inside the map cannot be serialized by the marshal layer, it returns: "[ERROR] :: <message>".
func PrintMapAsJSON(obj map[string]any, orderKeys []string, pretty bool, indent string) string {
	if obj == nil {
		return "{}"
	}

	// 1. Establish the default indentation token if left blank
	if pretty && indent == "" {
		indent = "  "
	}

	// 2. Map and identify all available keys inside the map matrix
	availableKeys := make(map[string]bool, len(obj))
	for k := range obj {
		availableKeys[k] = true
	}

	// 3. Extract prioritary ordered keys requested by the developer
	var finalOrderedKeys []string
	for _, k := range orderKeys {
		if availableKeys[k] {
			finalOrderedKeys = append(finalOrderedKeys, k)
			delete(availableKeys, k)
		}
	}

	// 4. Collect remaining keys and sort them alphabetically for absolute determinism
	var remainingKeys []string
	for k := range availableKeys {
		remainingKeys = append(remainingKeys, k)
	}
	sort.Strings(remainingKeys)
	finalOrderedKeys = append(finalOrderedKeys, remainingKeys...)

	// 5. Serialize the ordered pairs to buffer step-by-step
	var buf strings.Builder
	buf.WriteString("{")

	numKeys := len(finalOrderedKeys)
	for i, k := range finalOrderedKeys {
		// Apply line breaks and margin indents for pretty printing
		if pretty {
			buf.WriteString("\n")
			buf.WriteString(indent)
		}

		// Marshal key identifier securely
		keyBytes, err := jsonMarshalKeyHook(k)
		if err != nil {
			return fmt.Sprintf("[ERROR] :: %s", err.Error())
		}
		buf.Write(keyBytes)

		if pretty {
			buf.WriteString(": ")
		} else {
			buf.WriteString(":")
		}

		// Marshal the arbitrary value node
		valBytes, err := jsonMarshalKeyHook(obj[k])
		if err != nil {
			return fmt.Sprintf("[ERROR] :: %s", err.Error())
		}

		// If the value is a nested struct/map and we are in pretty mode,
		// we should format it to align nicely with the parent indentation.
		if pretty && (strings.HasPrefix(string(valBytes), "{") || strings.HasPrefix(string(valBytes), "[")) {
			var prettyValBuf bytes.Buffer
			encoder := json.NewEncoder(&prettyValBuf)
			encoder.SetIndent("", indent)
			err := encoder.Encode(obj[k])
			if err == nil {
				prettyVal := strings.TrimSpace(prettyValBuf.String())
				// Align nested child lines with the parent current indent depth margin
				prettyVal = strings.ReplaceAll(prettyVal, "\n", "\n"+indent)
				buf.WriteString(prettyVal)
			} else {
				buf.Write(valBytes)
			}
		} else {
			buf.Write(valBytes)
		}

		// Append field separator commas appropriately
		if i < numKeys-1 {
			buf.WriteString(",")
		}
	}

	if pretty && numKeys > 0 {
		buf.WriteString("\n")
	}
	buf.WriteString("}")

	return buf.String()
}

var jsonMarshalKeyHook = func(v any) ([]byte, error) {
	return json.Marshal(v)
}

// PrintObjectAsJSON marshals any generic structured data object into a clean JSON string sequence,
// features key blacklisting filter rules, and enforces an explicit horizontal property sequence.
//
// To circumvent compile-time static type boundaries, this function acts as an orchestration pipeline.
// It maps the input object into a flexible map, purges blacklisted keys at runtime, and delegates
// formatting to PrintMapAsJSON to bypass Go's native random map traversal behavior.
//
// Parameters:
//   - obj: The raw source entity, collection, or structured model to be serialized.
//   - orderKeys: An ordered collection of string tokens defining which JSON property keys must appear first at the top.
//   - ignoreKeys: A collection of field identifier strings to be completely stripped from the final payload.
//     Crucially, both orderKeys and ignoreKeys tokens must match the literal names defined inside the object's
//     struct `json:"..."` tags (e.g., "scenario_prompt" or "description") rather than internal Go identifiers.
//   - pretty: Flag governing whether to apply vertical whitespace structural alignment and line breaks.
//   - indent: The literal string pattern used for line padding constraints (e.g., "\t", "    ", "  ").
//
// Returns:
//   - A clean JSON string output stripped of any edge or trailing newline noise characters.
//   - If any phase of the underlying serialization pipeline fails, it returns: "[ERROR] :: <message>".
func PrintObjectAsJSON(
	obj any,
	orderKeys []string,
	ignoreKeys []string,
	pretty bool,
	indent string,
) string {
	if obj == nil {
		return "{}"
	}

	// Step 1: Extract the structural property matrix of the object into a live map
	rawData := ConvertObjectToMap(obj)
	if len(rawData) == 0 && obj != nil {
		if _, err := json.Marshal(obj); err != nil {
			return fmt.Sprintf("[ERROR] :: %s", err.Error())
		}
	}

	// Step 2: Intercept and purge unwanted property keys provided in the ignore filter parameters
	for _, field := range ignoreKeys {
		delete(rawData, field)
	}

	// Step 3: Dispatch the filtered map matrix into the ordered map serialization layer
	return PrintMapAsJSON(rawData, orderKeys, pretty, indent)
}
