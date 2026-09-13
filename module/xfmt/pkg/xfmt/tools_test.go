package xfmt_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xfmt/pkg/xfmt"
)

// TestPrint verifies that messages are correctly written to standard output.
func TestPrint(t *testing.T) {
	// 1. Capture os.Stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = w

	// 2. Execute the function with a single message (behaves like Println)
	xfmt.Print("hello world")

	// 3. Execute the function with arguments (behaves like Printf + \n)
	xfmt.Print("user %s has id %d", "john", 42)

	// Close the writer so we can read from the pipe
	w.Close()
	os.Stdout = oldStdout

	// 4. Read the captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read from stdout pipe: %v", err)
	}
	output := buf.String()

	// 5. Assert the results
	expected1 := "hello world\n"
	expected2 := "user john has id 42\n"

	if !strings.Contains(output, expected1) {
		t.Errorf("Print layout 1 failed.\nExpected to contain: %q\nFull output: %q", expected1, output)
	}
	if !strings.Contains(output, expected2) {
		t.Errorf("Print layout 2 failed.\nExpected to contain: %q\nFull output: %q", expected2, output)
	}
}

// TestPrintAsTable verifies that tabular data is correctly formatted and printed to standard output.
func TestPrintAsTable(t *testing.T) {
	// 1. Capture os.Stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = w

	// 2. Execute scenario 1: Empty headers to cover the early return block
	xfmt.PrintAsTable([]string{}, [][]string{{"data"}})

	// 3. Execute scenario 2: Valid table with a mix of valid rows and an invalid row to cover the len mismatch block
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"1", "Alice"},
		{"2"}, // Invalid row (mismatched length) - will be skipped
		{"3", "Bob"},
	}
	xfmt.PrintAsTable(headers, rows)

	// Close the writer so we can read from the pipe
	w.Close()
	os.Stdout = oldStdout

	// 4. Read the captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read from stdout pipe: %v", err)
	}
	output := buf.String()

	// 5. Assert the results
	// The tabwriter pads fields using three trailing spaces based on your configuration
	expectedHeader := "ID   NAME\n"
	expectedRow1 := "1    Alice\n"
	expectedRow2 := "3    Bob\n"
	unexpectedRow := "2"

	if !strings.Contains(output, expectedHeader) {
		t.Errorf("PrintAsTable failed.\nExpected to contain header: %q\nFull output: %q", expectedHeader, output)
	}
	if !strings.Contains(output, expectedRow1) {
		t.Errorf("PrintAsTable failed.\nExpected to contain row 1: %q\nFull output: %q", expectedRow1, output)
	}
	if !strings.Contains(output, expectedRow2) {
		t.Errorf("PrintAsTable failed.\nExpected to contain row 2: %q\nFull output: %q", expectedRow2, output)
	}
	if strings.Contains(output, unexpectedRow) {
		t.Errorf("PrintAsTable failed.\nDid not expect to contain skipped row containing: %q\nFull output: %q", unexpectedRow, output)
	}
}
