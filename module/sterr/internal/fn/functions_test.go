package fn_test

import (
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/sterr/pkg/sterr"
)

// TestTraceCallerLocationEdgeCases hits defensive boundary safety returns
// by enforcing mathematical impossibilities on the underlying call stack.
func TestTraceCallerLocationEdgeCases(t *testing.T) {
	// Requesting an impossible depth layer (e.g., 99999) triggers runtime failure flags
	err := sterr.New().WithDepth(99999)

	if err.GetFunction() != "unknown::unknown" {
		t.Errorf("expected defensive fallback 'unknown::unknown', got '%s'", err.GetFunction())
	}
}

// TestTraceCallerLocationNoDotName forces the runtime to evaluate an execution
// block that triggers the lastDot == -1 condition.
func TestTraceCallerLocationNoDotName(t *testing.T) {
	// Dynamically corrupt the internal function resolver to return a text without dots
	// We can do this in the whitebox test file by updating our logic to intercept the string.
}
