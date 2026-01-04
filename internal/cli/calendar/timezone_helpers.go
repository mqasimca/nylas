package calendar

import (
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/utilities/timezone"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

// ============================================================================
// Timezone Helpers
// ============================================================================

// getLocalTimeZone returns the user's local IANA timezone ID.
// Falls back to UTC if detection fails.
func getLocalTimeZone() string {
	local := time.Now().Location().String()

	// time.Local.String() returns "Local" which isn't an IANA ID
	// We need to get the actual IANA timezone name
	if local == "Local" || local == "" {
		// Try to load from system timezone
		// This works on Unix systems where /etc/localtime is a symlink
		// On macOS/Linux, we can read the timezone from system
		tz := getSystemTimeZone()
		if tz != "" {
			return tz
		}

		// Fallback to UTC
		return "UTC"
	}

	return local
}

// getSystemTimeZone attempts to detect the system's IANA timezone.
// Returns empty string if detection fails.
func getSystemTimeZone() string {
	// On Unix systems, check common environment variables
	// TZ environment variable often contains IANA timezone
	// This is a simplified implementation

	// For now, we'll use a heuristic based on UTC offset
	// In production, you might use a library or system call
	now := time.Now()
	_, offset := now.Zone()

	// Map common offsets to likely timezones
	// This is a simplified approach - in production, use proper detection
	offsetHours := offset / 3600

	switch offsetHours {
	case -8:
		return "America/Los_Angeles"
	case -7:
		return "America/Denver"
	case -6:
		return "America/Chicago"
	case -5:
		return "America/New_York"
	case 0:
		return "Europe/London"
	case 1:
		return "Europe/Paris"
	case 8:
		return "Asia/Singapore"
	case 9:
		return "Asia/Tokyo"
	default:
		// Return UTC as safe fallback
		return "UTC"
	}
}

// validateTimeZone checks if a timezone string is a valid IANA ID.
func validateTimeZone(tz string) error {
	if tz == "" {
		return common.NewUserError(
			"timezone cannot be empty",
			"Use IANA timezone IDs like 'America/Los_Angeles', 'Europe/London', etc.\nRun 'nylas timezone list' to see available timezones.",
		)
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return common.NewUserError(
			fmt.Sprintf("invalid timezone: %s", tz),
			"Use IANA timezone IDs like 'America/Los_Angeles', 'Europe/London', etc.\nRun 'nylas timezone list' to see available timezones.",
		)
	}
	return nil
}

// EventTimeDisplay represents formatted time display with timezone conversion.
type EventTimeDisplay struct {
	OriginalTime      string
	OriginalTimezone  string
	ConvertedTime     string
	ConvertedTimezone string
	ShowConversion    bool // true if original != converted
}

// formatEventTimeWithTZ formats an event's time with timezone conversion.
// If the event has timezone locking enabled, conversion is skipped and a lock indicator is shown.
func formatEventTimeWithTZ(event *domain.Event, targetTZ string) (*EventTimeDisplay, error) {
	display := &EventTimeDisplay{}
	when := event.When

	// For all-day events, no timezone conversion needed
	if when.IsAllDay() {
		start := when.StartDateTime()
		end := when.EndDateTime()
		if start.Equal(end) || end.IsZero() {
			display.OriginalTime = start.Format("Mon, Jan 2, 2006") + " (all day)"
		} else {
			display.OriginalTime = fmt.Sprintf("%s - %s (all day)",
				start.Format("Mon, Jan 2, 2006"),
				end.Format("Mon, Jan 2, 2006"))
		}
		display.ShowConversion = false
		return display, nil
	}

	// Get event times
	start := when.StartDateTime()
	end := when.EndDateTime()

	// Determine original timezone
	originalTZ := start.Location().String()
	if originalTZ == "Local" {
		originalTZ = getLocalTimeZone()
	}

	// Format original time
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		// Same day
		display.OriginalTime = fmt.Sprintf("%s, %s - %s",
			start.Format("Mon, Jan 2, 2006"),
			start.Format("3:04 PM"),
			end.Format("3:04 PM"))
	} else {
		display.OriginalTime = fmt.Sprintf("%s - %s",
			start.Format("Mon, Jan 2, 2006 3:04 PM"),
			end.Format("Mon, Jan 2, 2006 3:04 PM"))
	}

	// Get timezone abbreviations
	display.OriginalTimezone = start.Format("MST")

	// If event is timezone-locked, don't convert and show lock indicator
	if event.IsTimezoneLocked() {
		display.OriginalTime = display.OriginalTime + " 🔒"
		display.ShowConversion = false
		return display, nil
	}

	// Check if conversion is needed
	if targetTZ == "" || targetTZ == originalTZ {
		display.ShowConversion = false
		return display, nil
	}

	// Convert to target timezone
	tzService := timezone.NewService()
	ctx, cancel := common.CreateContext()
	defer cancel()

	convertedStart, err := tzService.ConvertTime(ctx, originalTZ, targetTZ, start)
	if err != nil {
		return nil, fmt.Errorf("timezone conversion failed: %w", err)
	}

	convertedEnd, err := tzService.ConvertTime(ctx, originalTZ, targetTZ, end)
	if err != nil {
		return nil, fmt.Errorf("timezone conversion failed: %w", err)
	}

	// Format converted time
	if convertedStart.Format("2006-01-02") == convertedEnd.Format("2006-01-02") {
		// Same day
		display.ConvertedTime = fmt.Sprintf("%s, %s - %s",
			convertedStart.Format("Mon, Jan 2, 2006"),
			convertedStart.Format("3:04 PM"),
			convertedEnd.Format("3:04 PM"))
	} else {
		display.ConvertedTime = fmt.Sprintf("%s - %s",
			convertedStart.Format("Mon, Jan 2, 2006 3:04 PM"),
			convertedEnd.Format("Mon, Jan 2, 2006 3:04 PM"))
	}

	display.ConvertedTimezone = convertedStart.Format("MST")
	display.ShowConversion = true

	return display, nil
}

// formatTimezoneBadge creates a formatted timezone badge for display.
// Returns a string like "[America/New_York]" or "[EST]" depending on format.
func formatTimezoneBadge(tz string, useAbbreviation bool) string {
	if tz == "" {
		return ""
	}

	if useAbbreviation {
		// Try to get timezone abbreviation
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return fmt.Sprintf("[%s]", tz)
		}
		abbr := time.Now().In(loc).Format("MST")
		return fmt.Sprintf("[%s]", abbr)
	}

	return fmt.Sprintf("[%s]", tz)
}

// getTimezoneColor returns a color code based on timezone offset.
// This helps visually distinguish different timezones in list views.
func getTimezoneColor(tz string) int {
	if tz == "" {
		return 7 // Default gray
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return 7
	}

	// Get UTC offset in hours
	_, offset := time.Now().In(loc).Zone()
	offsetHours := offset / 3600

	// Map offset ranges to colors
	// Using ANSI color codes: 31=red, 33=yellow, 32=green, 36=cyan, 34=blue, 35=magenta
	switch {
	case offsetHours <= -8: // Pacific and earlier
		return 34 // Blue
	case offsetHours <= -5: // Eastern, Central, Mountain
		return 36 // Cyan
	case offsetHours <= 0: // UTC and west
		return 32 // Green
	case offsetHours <= 3: // Europe
		return 33 // Yellow
	case offsetHours <= 12: // Asia and Pacific islands
		return 35 // Magenta
	default: // Edge cases
		return 31 // Red
	}
}
