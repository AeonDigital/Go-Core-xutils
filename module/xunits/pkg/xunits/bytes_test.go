package xunits_test

import (
	"encoding/json"
	"testing"

	"github.com/AeonDigital/Go-Core-xutils/module/xunits/pkg/xunits"
)

// TestBytes_UnmarshalJSON validates the parsing of various JSON string sizes into Bytes.
func TestBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    xunits.Bytes
		wantErr bool
	}{
		{"Plain bytes representation", `"512B"`, 512, false},
		{"No unit suffix defaults to bytes", `"1024"`, 1024, false},
		{"Kilobytes lowercase", `"2kb"`, 2 * 1024, false},
		{"Megabytes short uppercase", `"5M"`, 5 * 1024 * 1024, false},
		{"Gigabytes fractional decimal", `"1.5GB"`, xunits.Bytes(1.5 * 1024 * 1024 * 1024), false},
		{"Terabytes boundary", `"1TB"`, 1024 * 1024 * 1024 * 1024, false},
		{"Empty string defaults to zero", `""`, 0, false},
		{"Null JSON value defaults to zero", `"null"`, 0, false},
		// Error handling scenarios
		{"Invalid numeric structure error", `"abcMB"`, 0, true},
		{"Unknown unit suffix abbreviation error", `"100XYZ"`, 0, true},
		{"Malformed JSON syntax type error", `true`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bs xunits.Bytes
			err := json.Unmarshal([]byte(tt.jsonStr), &bs)

			if (err != nil) != tt.wantErr {
				t.Fatalf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && bs != tt.want {
				t.Errorf("json.Unmarshal() = %v, want %v", bs, tt.want)
			}
		})
	}
}

// TestBytes_String validates that Bytes formats correctly into human-readable strings.
func TestBytes_String(t *testing.T) {
	tests := []struct {
		name  string
		Bytes xunits.Bytes
		want  string
	}{
		{"Format pure bytes value", 450, "450B"},
		{"Format exact kilobytes boundary", 1024, "1.00KB"},
		{"Format fractional megabytes value", xunits.Bytes(2.75 * 1024 * 1024), "2.75MB"},
		{"Format exact gigabytes boundary", 1024 * 1024 * 1024, "1.00GB"},
		{"Format large terabytes boundary", 2 * 1024 * 1024 * 1024 * 1024, "2.00TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.Bytes.String()

			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
