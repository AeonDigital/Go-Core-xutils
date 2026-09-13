package sterr

import (
	"fmt"

	"github.com/AeonDigital/Go-Core-xutils/module/sterr/internal/fn"
	"github.com/AeonDigital/Go-Core-xutils/module/sterr/pkg/sterrinterfc"
)

// typedCliError implements the CliError interface providing immutability via cascading copies.
type typedCliError struct {
	function string
	devMsg   string
	userMsg  string
}

// New initializes a new structured error with automatic package and function trace detection.
//
// By default, it captures the immediate caller metadata at stack depth level 1.
func New() sterrinterfc.CliError {
	return &typedCliError{
		function: fn.TraceCallerLocation(1),
	}
}

// NewWithFunc initializes a new structured error enforcing an explicit, manually defined function location string.
//
// Arguments:
//   - functionName: The static text representation of the target scope location (e.g., "pkgname::FunctionName").
func NewWithFunc(functionName string) sterrinterfc.CliError {
	return &typedCliError{
		function: functionName,
	}
}

// Error formats the technical diagnostic telemetry string required for native Go error interfaces.
func (e *typedCliError) Error() string {
	return fmt.Sprintf("[FUNC: %s][MSG: %s]", e.function, e.devMsg)
}

// SetDevMessage overrides the current technical diagnostic text.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) SetDevMessage(format string, args ...any) sterrinterfc.CliError {
	return &typedCliError{
		function: e.function,
		devMsg:   fmt.Sprintf(format, args...),
		userMsg:  e.userMsg,
	}
}

// SetUserMessage overrides the current end-user friendly instruction.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) SetUserMessage(format string, args ...any) sterrinterfc.CliError {
	return &typedCliError{
		function: e.function,
		devMsg:   e.devMsg,
		userMsg:  fmt.Sprintf(format, args...),
	}
}

// AppendDevMessage concatenates a text payload into the current developer message.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) AppendDevMessage(format string, args ...any) sterrinterfc.CliError {
	payload := fmt.Sprintf(format, args...)
	if e.devMsg != "" {
		payload = e.devMsg + payload
	}
	return &typedCliError{
		function: e.function,
		devMsg:   payload,
		userMsg:  e.userMsg,
	}
}

// AppendUserMessage concatenates a text payload into the current end-user message.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) AppendUserMessage(format string, args ...any) sterrinterfc.CliError {
	payload := fmt.Sprintf(format, args...)
	if e.userMsg != "" {
		payload = e.userMsg + payload
	}
	return &typedCliError{
		function: e.function,
		devMsg:   e.devMsg,
		userMsg:  payload,
	}
}

// AppendLNDevMessage concatenates a text payload appending an explicit newline character at the end.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) AppendLNDevMessage(format string, args ...any) sterrinterfc.CliError {
	payload := fmt.Sprintf(format, args...) + "\n"
	if e.devMsg != "" {
		payload = e.devMsg + payload
	}
	return &typedCliError{
		function: e.function,
		devMsg:   payload,
		userMsg:  e.userMsg,
	}
}

// AppendLNUserMessage concatenates a text payload appending an explicit newline character at the end.
//
// Arguments:
//   - format: Standard print formatting string template base.
//   - args: Variable slice payloads to inject into placeholders.
func (e *typedCliError) AppendLNUserMessage(format string, args ...any) sterrinterfc.CliError {
	payload := fmt.Sprintf(format, args...) + "\n"
	if e.userMsg != "" {
		payload = e.userMsg + payload
	}
	return &typedCliError{
		function: e.function,
		devMsg:   e.devMsg,
		userMsg:  payload,
	}
}

// ClearDevMessage purges the developer message content, resetting it to empty string.
func (e *typedCliError) ClearDevMessage() sterrinterfc.CliError {
	return &typedCliError{
		function: e.function,
		devMsg:   "",
		userMsg:  e.userMsg,
	}
}

// ClearUserMessage purges the end-user message content, resetting it to empty string.
func (e *typedCliError) ClearUserMessage() sterrinterfc.CliError {
	return &typedCliError{
		function: e.function,
		devMsg:   e.devMsg,
		userMsg:  "",
	}
}

// WithDepth forces the engine to recalculate the runtime trace stack location using an offset.
//
// This is critical when encapsulating the error creation within helper factory patterns.
//
// Arguments:
//   - additionalDepth: Positive stack depth modifier index layer count.
func (e *typedCliError) WithDepth(additionalDepth int) sterrinterfc.CliError {
	// Protect against zero or negative depth alterations
	if additionalDepth <= 0 {
		return e
	}
	return &typedCliError{
		// +1 accounts for moving away from this wrapper call context scope inside the lib
		function: fn.TraceCallerLocation(1 + additionalDepth),
		devMsg:   e.devMsg,
		userMsg:  e.userMsg,
	}
}

// GetFunction returns the qualified target name tracking the package and functional scope.
func (e *typedCliError) GetFunction() string {
	return e.function
}

// GetDevMessage returns the technical diagnostics text string.
func (e *typedCliError) GetDevMessage() string {
	return e.devMsg
}

// GetUserMessage returns the actionable text instruction designed for human end-users.
func (e *typedCliError) GetUserMessage() string {
	return e.userMsg
}

// HasDevMessage verifies if a technical message payload has been populated.
func (e *typedCliError) HasDevMessage() bool {
	return e.devMsg != ""
}

// HasUserMessage verifies if a human-friendly instruction payload has been populated.
func (e *typedCliError) HasUserMessage() bool {
	return e.userMsg != ""
}

// HasErrors checks if any descriptive message fields contain active content tracking failures.
func (e *typedCliError) HasErrors() bool {
	return e.devMsg != "" || e.userMsg != ""
}
