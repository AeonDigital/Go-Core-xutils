package xunits_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AeonDigital/Go-Core-xutils/module/xunits/pkg/xunits"
)

// TestTimeDuration_UnmarshalJSON validates the parsing of JSON strings into TimeDuration.
func TestTimeDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    time.Duration
		wantErr bool
	}{
		{"Exact days conversion", `"1d"`, 24 * time.Hour, false},
		{"Multiple days conversion", `"7d"`, 7 * 24 * time.Hour, false},
		{"Standard Go hours format", `"10h"`, 10 * time.Hour, false},
		{"Standard Go minutes format", `"30m"`, 30 * time.Minute, false},
		// Reaches time.ParseDuration successfully
		{"Standard Go compound format", `"1h30m"`, 1*time.Hour + 30*time.Minute, false},
		// Reaches time.ParseDuration and returns an error
		{"Standard Go duration parse error", `"10x"`, 0, true},
		{"Invalid days format structure", `"abcd"`, 0, true},
		{"Invalid JSON data type", `123`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var td xunits.TimeDuration
			err := json.Unmarshal([]byte(tt.jsonStr), &td)

			if (err != nil) != tt.wantErr {
				t.Fatalf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && td.Duration != tt.want {
				t.Errorf("json.Unmarshal() = %v, want %v", td.Duration, tt.want)
			}
		})
	}
}

// TestTimeDuration_MarshalJSON validates the serialization of TimeDuration into JSON strings.
func TestTimeDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantJSON string
	}{
		{"Serializes exact days representation", 48 * time.Hour, `"2d"`},
		{"Serializes standard Go duration layout", 1*time.Hour + 30*time.Minute, `"1h30m0s"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := xunits.TimeDuration{Duration: tt.duration}

			gotBytes, err := json.Marshal(td)
			if err != nil {
				t.Fatalf("json.Marshal() failed unexpectedly: %v", err)
			}

			gotJSON := string(gotBytes)
			if gotJSON != tt.wantJSON {
				t.Errorf("json.Marshal() = %s, want %s", gotJSON, tt.wantJSON)
			}
		})
	}
}

// TestTimeDuration_String validates the output formatting of the String method.
func TestTimeDuration_String(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"One exact day string", 24 * time.Hour, "1d"},
		{"Multiple exact days string", 48 * time.Hour, "2d"},
		{"Fraction of a day output", 25 * time.Hour, "25h0m0s"},
		{"Standard Go minutes output", 45 * time.Minute, "45m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := xunits.TimeDuration{Duration: tt.duration}
			got := td.String()

			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
