package xreflect_test

import (
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xreflect/pkg/xreflect"
)

// TestNewInstanceOf verifies memory allocation for both value types and pointer types.
func TestNewInstanceOf(t *testing.T) {
	type DummyStruct struct {
		Value string
	}

	// Case 1: Value type allocation
	instanceValue := xreflect.NewInstanceOf[DummyStruct]()
	if instanceValue.Value != "" {
		t.Errorf("NewInstanceOf for value type should return an initialized zero-value struct")
	}

	// Case 2: Pointer type allocation (should allocate the underlying memory to prevent nil-pointer panics)
	instancePtr := xreflect.NewInstanceOf[*DummyStruct]()
	if instancePtr == nil {
		t.Fatalf("NewInstanceOf for pointer type returned nil, expected a valid memory allocation")
	}

	// Assigning a value ensures the pointer is safely dereferenceable. A nil pointer would crash the test here.
	instancePtr.Value = "testing pointer"
	if instancePtr.Value != "testing pointer" {
		t.Errorf("Allocated pointer could not receive values correctly")
	}
}
