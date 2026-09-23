package xtime

import (
	"strings"
	"time"
)

// FormatGenericDateTimeToGolangLayout converts a universal date/time format string (e.g., "YYYY-MM-DD HH:mm:ss.SSS a") into Go's native reference time layout.
// It maps standard placeholders for year, month, day, hours, minutes, seconds, milliseconds, meridiem indicators (AM/PM), and timezones to their corresponding Go time layout tokens.
// Returns the translated string ready to be used with Go's time parsing and formatting functions.
func FormatGenericDateTimeToGolangLayout(universalLayout string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"DD", "02",
		"HH", "15", // 24h (00-23)
		"hh", "03", // 12h (01-12)
		"mm", "04",
		"ss", "05",
		"SSS", "000", // Milliseconds
		"A", "PM", // AM/PM (Uppercase)
		"a", "pm", // am/pm (Lowercase)
		"ZZ", "-0700",
		"Z", "Z07:00", // UTC timezone string placeholder in Go
	)

	return replacer.Replace(universalLayout)
}

// Now returns the current local date and time.
func Now() time.Time {
	return time.Now()
}

// NowString returns the current date and time formatted as "YYYY-MM-DD HH:mm:ss".
func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ToDateTimeString formats the given time as "YYYY-MM-DD HH:mm:ss".
func ToDateTimeString(dt time.Time) string {
	return dt.Format("2006-01-02 15:04:05")
}

// ToDateString formats the given time keeping only its date part as "YYYY-MM-DD".
func ToDateString(dt time.Time) string {
	return dt.Format("2006-01-02")
}

// ToTimeString formats the given time keeping only its time part as "HH:mm:ss".
func ToTimeString(dt time.Time) string {
	return dt.Format("15:04:05")
}

// ToHourString formats the given time keeping only its hour and minute as "HH:mm".
func ToHourString(dt time.Time) string {
	return dt.Format("15:04")
}
