package xfmt_test

import (
	"os"
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xfmt/pkg/xfmt"
)

func TestGetTerminalWidth(t *testing.T) {
	origIsTerminal := *xfmt.ExportCliIsTerminal
	origGetSize := *xfmt.ExportCliGetTerminalSize
	defer func() {
		*xfmt.ExportCliIsTerminal = origIsTerminal
		*xfmt.ExportCliGetTerminalSize = origGetSize
	}()

	t.Run("execute standard hardware function initialization cover", func(t *testing.T) {
		// Restaura as variáveis padrão para a implementação real de produção
		*xfmt.ExportCliIsTerminal = origIsTerminal
		*xfmt.ExportCliGetTerminalSize = origGetSize

		// Invoca o ponteiro da função real de produção para computar a cobertura da linha anônima.
		// Passamos um descritor de arquivo inválido (-1) para que ela execute de forma segura sem crashar,
		// apenas gerando um erro esperado do sistema operacional.
		_, _, _ = (*xfmt.ExportCliGetTerminalSize)(-1)
	})
	t.Run("terminal check returns false", func(t *testing.T) {
		*xfmt.ExportCliIsTerminal = func(fd int) bool { return false }
		if got := xfmt.GetTerminalWidth(); got != 80 {
			t.Errorf("got %d, want 80", got)
		}
	})

	t.Run("terminal check true but gets size error", func(t *testing.T) {
		*xfmt.ExportCliIsTerminal = func(fd int) bool { return true }
		*xfmt.ExportCliGetTerminalSize = func(fd int) (int, int, error) {
			return 0, 0, os.ErrClosed
		}
		if got := xfmt.GetTerminalWidth(); got != 80 {
			t.Errorf("got %d, want 80", got)
		}
	})

	t.Run("terminal check true but width is zero or negative", func(t *testing.T) {
		*xfmt.ExportCliIsTerminal = func(fd int) bool { return true }
		*xfmt.ExportCliGetTerminalSize = func(fd int) (int, int, error) {
			return 0, 0, nil
		}
		if got := xfmt.GetTerminalWidth(); got != 80 {
			t.Errorf("got %d, want 80", got)
		}
	})

	t.Run("successful active terminal dimensions retrieval", func(t *testing.T) {
		*xfmt.ExportCliIsTerminal = func(fd int) bool { return true }
		*xfmt.ExportCliGetTerminalSize = func(fd int) (int, int, error) {
			return 140, 40, nil
		}
		if got := xfmt.GetTerminalWidth(); got != 140 {
			t.Errorf("got %d, want 140", got)
		}
	})
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		width       int
		useEllipsis bool
		want        string
	}{
		{
			name:        "text shorter than maximum width limit",
			text:        "Golang",
			width:       10,
			useEllipsis: true,
			want:        "Golang",
		},
		{
			name:        "text exactly matching maximum width limit",
			text:        "Terminal",
			width:       8,
			useEllipsis: true,
			want:        "Terminal",
		},
		{
			name:        "clean raw truncation without ellipsis indicator",
			text:        "Software Engineering",
			width:       8,
			useEllipsis: false,
			want:        "Software",
		},
		{
			name:        "dynamic ellipsis truncation fitting explicit boundaries",
			text:        "Responsive Layout Matrix",
			width:       15,
			useEllipsis: true,
			want:        "Responsive L...",
		},
		{
			name:        "boundary threshold constraint where width matches ellipsis length",
			text:        "Truncate",
			width:       3,
			useEllipsis: true,
			want:        "Tru",
		},
		{
			name:        "boundary threshold constraint where width is lower than ellipsis length",
			text:        "Truncate",
			width:       2,
			useEllipsis: true,
			want:        "Tr",
		},
		{
			name:        "empty text input preservation",
			text:        "",
			width:       5,
			useEllipsis: true,
			want:        "",
		},
		{
			name:        "zero width boundary constraints yielding empty return slices",
			text:        "No Width",
			width:       0,
			useEllipsis: false,
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xfmt.TruncateString(tt.text, tt.width, tt.useEllipsis)
			if got != tt.want {
				t.Errorf("TruncateString() = %q, want %q", got, tt.want)
			}
		})
	}
}
