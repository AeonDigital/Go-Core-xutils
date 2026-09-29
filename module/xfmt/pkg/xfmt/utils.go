package xfmt

import (
	"encoding/json"
	"os"

	"golang.org/x/term"
)

var (
	cliIsTerminal      = term.IsTerminal
	cliGetTerminalSize = func(fd int) (int, int, error) {
		return term.GetSize(fd)
	}
)

// GetTerminalWidth retrieves the active width of the current terminal screen in characters.
// It queries the standard output file descriptor. If the execution environment is not a
// valid terminal context (e.g., piped output, automated tests, or redirected streams),
// it returns a baseline fallback default value of 80 characters.
func GetTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if cliIsTerminal(fd) {
		width, _, err := cliGetTerminalSize(fd)
		if err == nil && width > 0 {
			return width
		}
	}
	return 80
}

// TruncateString clamps a string to a strict maximum width.
// If the content length exceeds the allocated boundary, it slices the string.
// When useEllipsis is enabled, the cut ensures the trailing indicator "..." is embedded,
// dynamically shrinking the visible text to guarantee the total width remains exact.
// When useEllipsis is disabled, it performs a clean, raw cut at the specified boundary limit.
func TruncateString(text string, width int, useEllipsis bool) string {
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	if useEllipsis {
		if width <= 3 {
			return string(runes[:width])
		}
		return string(runes[:width-3]) + "..."
	}

	return string(runes[:width])
}

// ConvertObjectToMap extracts the underlying structural property matrix of any generic object
// and maps its field tokens into a dynamic, flexible 'map[string]any'.
//
// To enforce compliance with custom serialization specifications, it routes conversion through
// the object's native `json:"..."` metadata boundaries. If a structural type constraint failure
// or serialization mismatch is encountered at runtime, the execution layer intercepts the fault
// gracefully, bypassing panic flows to return an empty map matrix.
//
// Parameters:
//   - obj: The raw source entity, collection, or anonymized structure to be inspected and translated.
//
// Returns:
//   - A live 'map[string]any' blueprint populated with the literal keys mapped by the object's JSON tags.
func ConvertObjectToMap(obj any) map[string]any {
	if obj == nil {
		return make(map[string]any)
	}

	bytesData, err := json.Marshal(obj)
	if err != nil {
		return make(map[string]any)
	}

	var rawData map[string]any
	err = json.Unmarshal(bytesData, &rawData)
	if err != nil {
		return make(map[string]any)
	}

	return rawData
}
