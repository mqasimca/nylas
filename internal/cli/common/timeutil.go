package common

import "time"

// Standard date/time format constants.
const (
	DateFormat        = "2006-01-02"
	TimeFormat        = "15:04"
	DateTimeFormat    = "2006-01-02 15:04"
	DisplayDateFormat = "Jan 2, 2006"
	DisplayTimeFormat = "3:04 PM"
	DisplayDateTime   = "Jan 2, 2006 3:04 PM"
)

// Standard timeout constants.
const (
	DefaultTimeout  = 30 * time.Second
	ShortTimeout    = 10 * time.Second
	LongTimeout     = 60 * time.Second
	VeryLongTimeout = 5 * time.Minute
)

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// ParseTime parses a time string in HH:MM format.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(TimeFormat, s)
}

// FormatDate formats a time as a date string (YYYY-MM-DD).
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatDisplayDate formats a time for user display.
func FormatDisplayDate(t time.Time) string {
	return t.Format(DisplayDateTime)
}
