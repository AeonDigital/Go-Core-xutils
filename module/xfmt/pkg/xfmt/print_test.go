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

func TestPrintObjectAsJSON(t *testing.T) {
	// Definimos uma struct local de testes para simular uma entidade de domínio padrão
	type mockUserEntity struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		RoleCode string `json:"role_code"`
	}

	t.Run("Success serialization to pretty JSON with indentation", func(t *testing.T) {
		user := mockUserEntity{
			ID:       101,
			FullName: "Alan Turing",
			RoleCode: "ADMIN",
		}

		jsonStr := xfmt.PrintObjectAsJSON(user, nil)

		// Garante que o retorno não disparou a máscara de erro
		if strings.HasPrefix(jsonStr, "[ERROR]") {
			t.Errorf("expected clean JSON output, got formatting failure: %s", jsonStr)
		}

		// Verifica se contém as chaves e valores esperados
		if !strings.Contains(jsonStr, `"id": 101`) || !strings.Contains(jsonStr, `"full_name": "Alan Turing"`) {
			t.Errorf("missing properties in serialized JSON payload: %s", jsonStr)
		}

		// Garante que a indentação de 2 espaços está ativa observando as quebras de linha com espaçamento
		if !strings.Contains(jsonStr, "\n  ") {
			t.Errorf("expected output to be pretty-printed with 2-space indentation format")
		}
	})

	t.Run("Success serialization stripping out blacklisted ignore fields using json tags", func(t *testing.T) {
		user := mockUserEntity{
			ID:       202,
			FullName: "Linus Torvalds",
			RoleCode: "KERNEL_DEV",
		}

		// Passamos os campos que queremos ignorar (devem bater com a tag JSON "role_code")
		ignoreList := []string{"role_code"}
		jsonStr := xfmt.PrintObjectAsJSON(user, ignoreList)

		if strings.HasPrefix(jsonStr, "[ERROR]") {
			t.Errorf("expected clean JSON output, got formatting failure: %s", jsonStr)
		}

		// O nome e ID devem persistir, mas o papel secreto deve sumir
		if !strings.Contains(jsonStr, `"full_name": "Linus Torvalds"`) {
			t.Errorf("expected 'full_name' key to persist in the payload")
		}
		if strings.Contains(jsonStr, `"role_code"`) || strings.Contains(jsonStr, "KERNEL_DEV") {
			t.Errorf("found blacklisted 'role_code' property key inside filtered JSON output: %s", jsonStr)
		}
	})

	t.Run("Fail execution path returning formatted error string on encoding constraints", func(t *testing.T) {
		// Passar uma função pura ou um canal direto dentro de um tipo anônimo ou de um map
		// quebra o json.Marshal nativo de forma garantida na primeira linha da função.
		brokenData := map[string]any{
			"unsupported_field": make(chan int),
		}

		jsonStr := xfmt.PrintObjectAsJSON(brokenData, nil)

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected engine barrier to capture serialization fault, but got: %s", jsonStr)
		}
	})

	t.Run("Fail execution path on unmarshal stage constraints", func(t *testing.T) {
		// Retorna uma string crua inválida como JSON objeto (ex: apenas um texto solto sem chaves)
		_ = func() ([]byte, error) {
			return []byte(`"not-an-object"`), nil
		}

		// Como Go exige conformidade estrita para chamar o Marshaler customizado, podemos emular
		// passando um tipo primitivo complexo que o Marshal aceite mas o Unmarshal para map rejeite.
		// Se passarmos um slice de inteiros: o Marshal gera "[1,2,3]".
		// O Unmarshal tenta converter um array JSON "[]" para um "map[string]any", o que é um erro de tipo inválido!
		sliceData := []int{1, 2, 3}

		jsonStr := xfmt.PrintObjectAsJSON(sliceData, nil)

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected unmarshal matrix constraint to trigger a failure token, but got: %s", jsonStr)
		}
	})

	t.Run("Fail execution path on final encoder stage constraints", func(t *testing.T) {
		// Salva o executor original e garante sua restauração ao final do subteste
		origEncoder := *xfmt.ExportJsonPrettyEncoderExecute
		defer func() { *xfmt.ExportJsonPrettyEncoderExecute = origEncoder }()

		// Sabota o executor final para forçar um erro controlado
		*xfmt.ExportJsonPrettyEncoderExecute = func(w io.Writer, v any) error {
			return os.ErrClosed // Simula um erro qualquer de I/O ou restrição
		}

		validData := map[string]any{"ok": true}
		jsonStr := xfmt.PrintObjectAsJSON(validData, nil)

		if !strings.HasPrefix(jsonStr, "[ERROR] ::") {
			t.Errorf("expected final encoder barrier to capture serialization fault, but got: %s", jsonStr)
		}
	})
}
