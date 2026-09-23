package xtime_test

import (
	"strings"
	"testing"
	"time"

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

// TestNow verifies that Now returns a time value close to the actual current time.
func TestNow(t *testing.T) {
	before := time.Now()
	got := xtime.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("Now() returned %v, expected a value between %v and %v", got, before, after)
	}
}

// TestNowString verifies that NowString formats the current time as "YYYY-MM-DD HH:mm:ss".
func TestNowString(t *testing.T) {
	output := xtime.NowString()

	if _, err := time.Parse("2006-01-02 15:04:05", output); err != nil {
		t.Errorf("NowString() returned %q, which does not match layout \"2006-01-02 15:04:05\": %v", output, err)
	}

	if !strings.HasPrefix(output, xtime.ToDateString(time.Now())) {
		t.Errorf("NowString() returned %q, expected it to start with today's date", output)
	}
}

// TestToDateTimeString verifies formatting of a time value as "YYYY-MM-DD HH:mm:ss".
func TestToDateTimeString(t *testing.T) {
	dt := time.Date(2024, time.March, 5, 13, 45, 30, 0, time.UTC)
	expected := "2024-03-05 13:45:30"

	if output := xtime.ToDateTimeString(dt); output != expected {
		t.Errorf("ToDateTimeString() = %q, expected %q", output, expected)
	}
}

// TestToDateString verifies formatting of a time value keeping only its date part.
func TestToDateString(t *testing.T) {
	dt := time.Date(2024, time.March, 5, 13, 45, 30, 0, time.UTC)
	expected := "2024-03-05"

	if output := xtime.ToDateString(dt); output != expected {
		t.Errorf("ToDateString() = %q, expected %q", output, expected)
	}
}

// TestToTimeString verifies formatting of a time value keeping only its time part.
func TestToTimeString(t *testing.T) {
	dt := time.Date(2024, time.March, 5, 13, 45, 30, 0, time.UTC)
	expected := "13:45:30"

	if output := xtime.ToTimeString(dt); output != expected {
		t.Errorf("ToTimeString() = %q, expected %q", output, expected)
	}
}

// TestToHourString verifies formatting of a time value keeping only hour and minute.
func TestToHourString(t *testing.T) {
	dt := time.Date(2024, time.March, 5, 13, 45, 30, 0, time.UTC)
	expected := "13:45"

	if output := xtime.ToHourString(dt); output != expected {
		t.Errorf("ToHourString() = %q, expected %q", output, expected)
	}
}
