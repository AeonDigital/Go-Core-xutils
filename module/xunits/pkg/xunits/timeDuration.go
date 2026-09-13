package xunits

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
	"github.com/AeonDigital/Go-Core-xutils/module/xstrings/pkg/xstrings"
)

const XERR_PKGCTX_TIMEDURATION xerrors.ErrorCode = "XUNITS.TIMEDURATION"

// TimeDuration extends standard time.Duration to add custom JSON
// serialization and support for day-based representations (e.g., "7d").
type TimeDuration struct {
	time.Duration
}

// MarshalJSON implements the json.Marshaler interface.
// It serializes the TimeDuration into its string representation.
func (o TimeDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses a JSON string into a TimeDuration, extending time.ParseDuration
// to natively support the "d" suffix for days (1d = 24h).
func (o *TimeDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Check if the duration string ends with the custom 'd' (days) suffix
	if strings.HasSuffix(s, "d") {
		var days int
		daysStr := strings.TrimSuffix(s, "d")

		days, err := strconv.Atoi(daysStr)
		if err != nil {
			errInfo := xstrings.BuildErrorInfo(s, "integer with 'd' suffix")

			return xerrors.NewError500(
				XERR_PKGCTX_TIMEDURATION,
				xerrors.XERR_INVALID_FORMAT,
				nil,
				"",
				errInfo,
			)
		}

		o.Duration = time.Duration(days) * 24 * time.Hour
		return nil
	}

	// Fallback to the standard standard Go duration parsing (e.g., "1s", "5m", "10h")
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	o.Duration = dur
	return nil
}

// String returns a string representation of the duration.
// If the duration is an exact multiple of 24 hours, it formats it using
// the "d" suffix (e.g., "2d"). Otherwise, it falls back to time.Duration.String().
func (o TimeDuration) String() string {
	// Check if the duration is an exact multiple of 24 hours
	const exactDayDuration = 24 * time.Hour
	if o.Duration >= exactDayDuration && (o.Duration%exactDayDuration) == 0 {
		days := o.Duration / exactDayDuration
		return fmt.Sprintf("%dd", days)
	}

	// Fallback to standard Go formatting (e.g., "1h30m0s")
	return o.Duration.String()
}
