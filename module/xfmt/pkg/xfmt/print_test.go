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

// TestPrintStream verifies that messages are written to standard output without an automatic newline.
func TestPrintStream(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = w

	xfmt.PrintInline("hello ")
	xfmt.PrintInline("%s has id %d", "john", 42)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read from stdout pipe: %v", err)
	}

	expected := "hello john has id 42"
	if output := buf.String(); output != expected {
		t.Errorf("PrintStream output mismatch.\nExpected: %q\nActual: %q", expected, output)
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
	xfmt.PrintCLITable([]string{}, [][]string{{"data"}})

	// 3. Execute scenario 2: Valid table with a mix of valid rows and an invalid row to cover the len mismatch block
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"1", "Alice"},
		{"2"}, // Invalid row (mismatched length) - will be skipped
		{"3", "Bob"},
	}
	xfmt.PrintCLITable(headers, rows)

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

func TestPrintMapAsJSON(t *testing.T) {
	t.Run("Success formatting with explicit key prioritization sequence", func(t *testing.T) {
		input := map[string]any{
			"status":  "active",
			"id":      int64(42),
			"version": "v1.2.0",
		}

		// Solicitamos que "id" e "status" venham primeiro, "version" deve ir para o final automaticamente
		orderList := []string{"id", "status"}
		jsonStr := xfmt.PrintMapAsJSON(input, orderList, false, "")

		if strings.HasPrefix(jsonStr, "[ERROR]") {
			t.Fatalf("expected clean JSON output, got formatting failure: %s", jsonStr)
		}

		// Como a saída é determinística e inline (pretty=false), validamos o padrão exato do texto
		wantInline := `{"id":42,"status":"active","version":"v1.2.0"}`
		if jsonStr != wantInline {
			t.Errorf("ordered inline serialization output mismatch\ngot : %s\nwant: %s", jsonStr, wantInline)
		}
	})

	t.Run("Success formatting remaining keys alphabetically when not listed in orderKeys", func(t *testing.T) {
		input := map[string]any{
			"charlie": 3,
			"alpha":   1,
			"bravo":   2,
		}

		// Nenhuma chave prioritária fornecida. O algoritmo deve ordenar todas alfabeticamente
		jsonStr := xfmt.PrintMapAsJSON(input, nil, false, "")

		wantAlphabetical := `{"alpha":1,"bravo":2,"charlie":3}`
		if jsonStr != wantAlphabetical {
			t.Errorf("alphabetical backup sorting sequence failed\ngot : %s\nwant: %s", jsonStr, wantAlphabetical)
		}
	})

	t.Run("Success formatting to pretty JSON adopting default 2-space padding", func(t *testing.T) {
		input := map[string]any{
			"id":   1,
			"name": "xfmt",
		}

		orderList := []string{"id", "name"}
		jsonStr := xfmt.PrintMapAsJSON(input, orderList, true, "")

		if !strings.Contains(jsonStr, "\n  \"id\": 1,") || !strings.Contains(jsonStr, "\n  \"name\": \"xfmt\"\n}") {
			t.Errorf("default 2-space pretty printing pattern mismatch, got:\n%s", jsonStr)
		}
	})

	t.Run("Success formatting to pretty JSON utilizing a custom tab character layout depth", func(t *testing.T) {
		input := map[string]any{"service": "core"}

		jsonStr := xfmt.PrintMapAsJSON(input, nil, true, "\t")

		if !strings.Contains(jsonStr, "\n\t\"service\": \"core\"\n}") {
			t.Errorf("explicit tab layout depth injection constraint failed, got:\n%s", jsonStr)
		}
	})

	t.Run("Success formatting deeply nested maps and array objects dynamically matching parent indent depth", func(t *testing.T) {
		input := map[string]any{
			"id": 100,
			"meta": map[string]any{
				"tag": "cli",
			},
		}

		orderList := []string{"id", "meta"}
		jsonStr := xfmt.PrintMapAsJSON(input, orderList, true, "  ")

		// O objeto interno "meta" deve ser quebrado verticalmente e suas propriedades internas
		// devem ser concatenadas contendo o recuo de margem acumulado proporcional da indentação pai ("    ")
		if !strings.Contains(jsonStr, "\n    \"tag\": \"cli\"") {
			t.Errorf("nested object structure indentation depth failed to cascade align, got:\n%s", jsonStr)
		}
	})

	t.Run("Graceful early exit preservation on nil input reference", func(t *testing.T) {
		jsonStr := xfmt.PrintMapAsJSON(nil, nil, false, "")
		if jsonStr != "{}" {
			t.Errorf("expected clean empty object serialization sequence on nil reference, got %q", jsonStr)
		}
	})

	t.Run("Fail execution path returning formatted error string when encountering non-marshallable value tokens", func(t *testing.T) {
		// Injetamos um tipo bruto impossível de converter para JSON (um canal ativo) como valor do mapa
		brokenInput := map[string]any{
			"valid_key": "ok",
			"bad_key":   make(chan int),
		}

		// CORREÇÃO: Chamando explicitamente a PrintMapAsJSON
		jsonStr := xfmt.PrintMapAsJSON(brokenInput, nil, false, "")

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected value serialization barrier to capture marshaling error, but got: %s", jsonStr)
		}
	})

	t.Run("Fail execution path returning formatted error string when key marshalling explicitly fails", func(t *testing.T) {
		// Salva o comportamento original e garante a restauração ao final do teste
		origKeyMarshal := *xfmt.ExportJsonMarshalKeyHook
		defer func() { *xfmt.ExportJsonMarshalKeyHook = origKeyMarshal }()

		// Sabota o marshal de chave para forçar um erro controlado
		*xfmt.ExportJsonMarshalKeyHook = func(v any) ([]byte, error) {
			return nil, os.ErrClosed
		}

		input := map[string]any{"trigger": true}
		jsonStr := xfmt.PrintMapAsJSON(input, nil, false, "")

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected key serialization barrier to capture failure token, but got: %s", jsonStr)
		}
	})
}

func TestPrintObjectAsJSON_Orchestrator(t *testing.T) {
	type mockRoleEntity struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Persona string `json:"description"`
		Secret  string `json:"secret_token"`
	}

	t.Run("Success full orchestration pipeline execution with ordering and filtering constraints", func(t *testing.T) {
		role := mockRoleEntity{
			ID:      7,
			Name:    "System Architect",
			Persona: "AI Model Assistant",
			Secret:  "XYZ123TEST",
		}

		// Queremos que "name" venha antes de "id", e queremos apagar "secret_token" completamente
		orderList := []string{"name", "id"}
		ignoreList := []string{"secret_token"}

		jsonStr := xfmt.PrintObjectAsJSON(role, orderList, ignoreList, false, "")

		if strings.HasPrefix(jsonStr, "[ERROR]") {
			t.Fatalf("expected clear orchestrated JSON output, got error: %s", jsonStr)
		}

		// O token secreto deve ter sumido
		if strings.Contains(jsonStr, "secret_token") || strings.Contains(jsonStr, "XYZ123TEST") {
			t.Errorf("found blacklisted key 'secret_token' inside filtered output: %s", jsonStr)
		}

		// Validamos o padrão exato da string inline ordenada determinística
		wantInline := `{"name":"System Architect","id":7,"description":"AI Model Assistant"}`
		if jsonStr != wantInline {
			t.Errorf("orchestrated serialization output mismatch\ngot : %s\nwant: %s", jsonStr, wantInline)
		}
	})

	t.Run("Graceful early exit preservation on nil object input interface", func(t *testing.T) {
		jsonStr := xfmt.PrintObjectAsJSON(nil, nil, nil, false, "")
		if jsonStr != "{}" {
			t.Errorf("expected clean empty object serialization sequence on nil reference, got %q", jsonStr)
		}
	})

	t.Run("Fail execution path returning formatted error string when encountering non-marshallable objects", func(t *testing.T) {
		// Passar um tipo impossível de serializar (um canal ativo) força a branch de erro interno do Step 1
		brokenData := map[string]any{
			"bad_field": make(chan int),
		}

		jsonStr := xfmt.PrintObjectAsJSON(brokenData, nil, nil, false, "")

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected orchestrator barrier to capture serialization error, but got: %s", jsonStr)
		}
	})
}
