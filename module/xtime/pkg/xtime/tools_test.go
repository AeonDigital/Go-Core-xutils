package xtime_test

import (
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xtime/pkg/xtime"
)

// TestNewErr verifies the dynamic creation of error objects.
// TestFormatGenericDateTimeToGolangLayout verifies translation of universal date tokens to Go layout tokens.
func TestFormatGenericDateTimeToGolangLayout(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard Date Format",
			input:    "YYYY-MM-DD",
			expected: "2006-01-02",
		},
		{
			name:     "Full Format with Time and Milliseconds",
			input:    "YYYY-MM-DD HH:mm:ss.SSS a",
			expected: "2006-01-02 15:04:05.000 pm",
		},
		{
			name:     "12h Format with Uppercase Meridiem",
			input:    "hh:mm A",
			expected: "03:04 PM",
		},
		{
			name:     "ISO Format with UTC Timezone",
			input:    "YYYY-MM-DDTHH:mm:ssZ",
			expected: "2006-01-02T15:04:05Z07:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := xtime.FormatGenericDateTimeToGolangLayout(tt.input)
			if output != tt.expected {
				t.Errorf("Test case '%s' failed.\nInput:    %s\nExpected: %s\nGot:   %s", tt.name, tt.input, tt.expected, output)
			}
		})
	}
}
