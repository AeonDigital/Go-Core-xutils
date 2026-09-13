package xjson_test

import (
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xjson/pkg/xjson"
)

// TestDumpToJSON verifies serialization of structures into indented JSON strings.
func TestDumpToJSON(t *testing.T) {
	type Sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := Sample{Name: "Aeon", Age: 10}

	// 1. Test with empty indentation string (should default to 2 spaces)
	resultDefault := xjson.Dump(data, "")
	expectedDefault := "{\n  \"name\": \"Aeon\",\n  \"age\": 10\n}"
	if resultDefault != expectedDefault {
		t.Errorf("DumpToJSON() with empty indentation failed.\nExpected:\n%s\nGot:\n%s", expectedDefault, resultDefault)
	}

	// 2. Test with custom indentation string (4 spaces)
	resultCustom := xjson.Dump(data, "    ")
	expectedCustom := "{\n    \"name\": \"Aeon\",\n    \"age\": 10\n}"
	if resultCustom != expectedCustom {
		t.Errorf("DumpToJSON() with 4-space indentation failed.\nExpected:\n%s\nGot:\n%s", expectedCustom, resultCustom)
	}

	// 3. Test with an invalid type that causes a marshalling error (e.g., channels cannot be marshaled to JSON)
	ch := make(chan int)
	resultError := xjson.Dump(ch, "")
	if !strings.Contains(resultError, "json: unsupported type") {
		t.Errorf("DumpToJSON() with a channel should return an error message string, got: %s", resultError)
	}
}
