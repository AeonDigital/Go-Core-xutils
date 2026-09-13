package xunits

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
	"github.com/AeonDigital/Go-Core-xutils/module/xstrings/pkg/xstrings"
)

const XERR_PKGCTX_BYTES xerrors.ErrorCode = "XUNITS.BYTES"

// Bytes represents a data size in bytes as an unsigned 64-bit integer.
type Bytes uint64

// Binary/IEC byte size constants based on powers of 1024.
const (
	B  Bytes = 1
	KB       = B * 1024
	MB       = KB * 1024
	GB       = MB * 1024
	TB       = GB * 1024
)

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses a JSON string containing human-readable data sizes
// (e.g., "10mb", "500K", "1.5GB") into a Bytes integer value.
func (bs *Bytes) UnmarshalJSON(b []byte) error {
	// Remove JSON string quotes and surrounding whitespace
	s := string(bytes.Trim(b, `"`))
	s = strings.TrimSpace(s)

	if s == "" || s == "null" {
		*bs = 0
		return nil
	}

	// Separate numeric characters from the unit suffix literal
	var numStr strings.Builder
	var unitStr strings.Builder

	for _, r := range s {
		if unicode.IsDigit(r) || r == '.' {
			numStr.WriteRune(r)
		} else if unicode.IsLetter(r) {
			unitStr.WriteRune(r)
		}
	}

	// Parse the numeric part into a float64 to correctly support fractional values like "1.5Mb"
	value, err := strconv.ParseFloat(numStr.String(), 64)
	if err != nil {
		errInfo := xstrings.BuildErrorInfo(
			"rawvalue", s, "float64", numStr.String(),
		)

		return xerrors.NewError500(
			XERR_PKGCTX_BYTES,
			xerrors.XERR_INVALID_FORMAT_PARSE,
			err,
			"",
			errInfo,
		)
	}

	// Normalize the unit suffix to uppercase for standardized matching
	unit := strings.ToUpper(strings.TrimSpace(unitStr.String()))

	var multiplier Bytes
	switch unit {
	case "", "B":
		multiplier = B
	case "K", "KB":
		multiplier = KB
	case "M", "MB":
		multiplier = MB
	case "G", "GB":
		multiplier = GB
	case "T", "TB":
		multiplier = TB
	default:
		errInfo := xstrings.BuildErrorInfo(
			"rawvalue", s, "B, K, KB, M, MB, G, GB, T, TB", unitStr.String(),
		)

		return xerrors.NewError500(
			XERR_PKGCTX_BYTES,
			xerrors.XERR_INVALID_OPTION,
			nil,
			"",
			errInfo,
		)
	}

	*bs = Bytes(value * float64(multiplier))
	return nil
}

// String implements the fmt.Stringer interface.
// It formats the Bytes value back into a human-readable string
// rounded to two decimal places (e.g., "1.50MB", "450B").
func (bs Bytes) String() string {
	switch {
	case bs >= TB:
		return fmt.Sprintf("%.2fTB", float64(bs)/float64(TB))
	case bs >= GB:
		return fmt.Sprintf("%.2fGB", float64(bs)/float64(GB))
	case bs >= MB:
		return fmt.Sprintf("%.2fMB", float64(bs)/float64(MB))
	case bs >= KB:
		return fmt.Sprintf("%.2fKB", float64(bs)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bs)
	}
}
