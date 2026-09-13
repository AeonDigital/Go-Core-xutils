package sterr_test

import (
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/sterr/pkg/sterr"
	"github.com/AeonDigital/Go-Core-xutils/module/sterr/pkg/sterrinterfc"
)

// TestNewErrorAutoLocation validates if sterr.New capturing engine successfully
// extracts the current execution context package and function name.
func TestNewErrorAutoLocation(t *testing.T) {
	// Execute the constructor. The context must be this exact function.
	err := sterr.New()

	expected := "sterr_test::TestNewErrorAutoLocation"
	if err.GetFunction() != expected {
		t.Errorf("expected function location '%s', got '%s'", expected, err.GetFunction())
	}
}

// TestNewWithExplicitFunc validates if sterr.NewWithFunc bypasses the runtime stack
// registers and enforces the exact string provided by the developer.
func TestNewWithExplicitFunc(t *testing.T) {
	customFunc := "custompkg::CustomFunction"
	err := sterr.NewWithFunc(customFunc)

	if err.GetFunction() != customFunc {
		t.Errorf("expected function location '%s', got '%s'", customFunc, err.GetFunction())
	}
}

// TestWithDepthModifier validates if WithDepth successfully steps back into the
// call stack execution chain to recalculate the frame pointer destination.
func TestWithDepthModifier(t *testing.T) {
	// Helper function simulating an internal library factory wrapper frame layer
	errorFactoryWrapper := func() sterrinterfc.CliError {
		// depth 1 would capture errorFactoryWrapper.
		// We pass depth 1 additional layer to capture TestWithDepthModifier.
		return sterr.New().WithDepth(1)
	}

	err := errorFactoryWrapper()
	expected := "sterr_test::TestWithDepthModifier"

	if err.GetFunction() != expected {
		t.Errorf("expected function location '%s', got '%s'", expected, err.GetFunction())
	}

	// Technical Coverage Edge Case: Send an invalid zero or negative depth.
	// The imutability pattern must return the exact same object state without crashing.
	baseErr := sterr.New()
	modifiedErr := baseErr.WithDepth(0)
	if baseErr != modifiedErr {
		t.Errorf("expected original instance when depth modifier is invalid")
	}
}

// TestErrorImutabilityAndSetters validates that every Set operation creates
// a completely new, isolated instance without modifying the parent error.
func TestErrorImutabilityAndSetters(t *testing.T) {
	// 1. Create a pristine base error
	baseErr := sterr.NewWithFunc("pkg::TestFunc")

	// 2. Apply SetDevMessage and verify formatting + isolation
	devErr := baseErr.SetDevMessage("database failure: %d", 500)
	if devErr == baseErr {
		t.Errorf("expected a new instance pointer, but got the same base pointer")
	}
	if devErr.GetDevMessage() != "database failure: 500" {
		t.Errorf("expected dev message to be formatted, got '%s'", devErr.GetDevMessage())
	}
	if baseErr.GetDevMessage() != "" {
		t.Errorf("mutation leaked: base error dev message should remain empty")
	}

	// 3. Apply SetUserMessage and verify formatting + isolation
	userErr := devErr.SetUserMessage("tente novamente mais tarde %s", "admin")
	if userErr == devErr {
		t.Errorf("expected a new instance pointer from user setter")
	}
	if userErr.GetUserMessage() != "tente novamente mais tarde admin" {
		t.Errorf("expected user message to be formatted, got '%s'", userErr.GetUserMessage())
	}
	if devErr.GetUserMessage() != "" {
		t.Errorf("mutation leaked: previous error user message should remain empty")
	}
}

// TestClearMethods validates if Clear operations correctly purge content
// by returning a new clean state while keeping original objects intact.
func TestClearMethods(t *testing.T) {
	// 1. Setup an error containing both payloads
	fullErr := sterr.NewWithFunc("pkg::TestFunc").
		SetDevMessage("tech message").
		SetUserMessage("human message")

	// 2. Clear dev message and check isolation
	noDevErr := fullErr.ClearDevMessage()
	if noDevErr.HasDevMessage() {
		t.Errorf("expected dev message to be empty after clear")
	}
	if !fullErr.HasDevMessage() {
		t.Errorf("mutation leaked: fullErr should still preserve its dev message")
	}

	// 3. Clear user message and check isolation
	noUserErr := fullErr.ClearUserMessage()
	if noUserErr.HasUserMessage() {
		t.Errorf("expected user message to be empty after clear")
	}
	if !fullErr.HasUserMessage() {
		t.Errorf("mutation leaked: fullErr should still preserve its user message")
	}
}

// TestAppendAndAppendLNMethods validates cumulative text concatenation behaviors
// ensuring correct spacing structure and tracking isolation between calls.
func TestAppendAndAppendLNMethods(t *testing.T) {
	baseErr := sterr.NewWithFunc("pkg::TestAppend")

	// 1. Test standard continuous Append on Dev Message
	firstAppend := baseErr.AppendDevMessage("first payload")
	secondAppend := firstAppend.AppendDevMessage(" second payload")

	if secondAppend.GetDevMessage() != "first payload second payload" {
		t.Errorf("expected seamless continuation, got '%s'", secondAppend.GetDevMessage())
	}

	// 2. Test AppendLN (with automatic trailing newline) on Dev Message
	lnAppend := baseErr.AppendLNDevMessage("line 1")
	lnSecond := lnAppend.AppendLNDevMessage("line 2")

	if lnSecond.GetDevMessage() != "line 1\nline 2\n" {
		t.Errorf("expected formatted newline stacking blocks, got %q", lnSecond.GetDevMessage())
	}

	// 3. Test standard continuous Append on User Message
	userFirst := baseErr.AppendUserMessage("hello")
	userSecond := userFirst.AppendUserMessage(" world")

	if userSecond.GetUserMessage() != "hello world" {
		t.Errorf("expected clean user message tracking, got '%s'", userSecond.GetUserMessage())
	}

	// 4. Test AppendLN (with automatic trailing newline) on User Message
	userLNFirst := baseErr.AppendLNUserMessage("step 1")
	userLNSecond := userLNFirst.AppendLNUserMessage("step 2")

	if userLNSecond.GetUserMessage() != "step 1\nstep 2\n" {
		t.Errorf("expected human-friendly multi-line payload, got %q", userLNSecond.GetUserMessage())
	}
}

// TestStateInspectionAndNativeError validates all structural state verification booleans
// alongside the standard technical output of the native Go error interface contract.
func TestStateInspectionAndNativeError(t *testing.T) {
	err := sterr.NewWithFunc("pkg::TestState")

	// 1. Verify initial pristine state (completely empty)
	if err.HasDevMessage() || err.HasUserMessage() || err.HasErrors() {
		t.Errorf("newly initialized error instance should report zero boolean state flags")
	}

	// 2. Hydrate only dev message and re-check flags
	errWithDev := err.SetDevMessage("low level crash")
	if !errWithDev.HasDevMessage() {
		t.Errorf("expected HasDevMessage to toggle true")
	}
	if !errWithDev.HasErrors() {
		t.Errorf("expected HasErrors to report true when tracking active failure descriptions")
	}
	if errWithDev.HasUserMessage() {
		t.Errorf("user message state check should remain false")
	}

	// 3. Hydrate user message and re-check flags
	errWithUser := err.SetUserMessage("friendly warning")
	if !errWithUser.HasUserMessage() {
		t.Errorf("expected HasUserMessage to toggle true")
	}
	if !errWithUser.HasErrors() {
		t.Errorf("expected HasErrors to report true when tracking user alerts")
	}

	// 4. Validate native Go Error() interface signature string formatting
	targetTechOutput := errWithDev.Error()
	expectedFormat := "[FUNC: pkg::TestState][MSG: low level crash]"
	if targetTechOutput != expectedFormat {
		t.Errorf("native Error() output mismatch. Expected %q, got %q", expectedFormat, targetTechOutput)
	}
}
