package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
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

// convertEventToTimeZone converts an event's time to a target timezone.
// Returns formatted display strings for both original and converted times.
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

// ============================================================================
// DST Warning Helpers
// ============================================================================

// checkDSTWarning checks if an event time has DST warnings and returns formatted message.
// Returns empty string if no warning.
func checkDSTWarning(eventTime time.Time, tz string) string {
	if tz == "" {
		return ""
	}

	// Use timezone service to check for DST warnings
	tzService := timezone.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check for warnings within 7 days
	warning, err := tzService.CheckDSTWarning(ctx, eventTime, tz, 7)
	if err != nil || warning == nil {
		return ""
	}

	if !warning.IsNearTransition {
		return ""
	}

	// Format warning message with appropriate icon
	return formatDSTWarning(warning)
}

// formatDSTWarning formats a DST warning for display in the terminal.
func formatDSTWarning(warning *domain.DSTWarning) string {
	if warning == nil {
		return ""
	}

	var icon string
	switch warning.Severity {
	case "error":
		icon = "⛔"
	case "warning":
		icon = "⚠️"
	case "info":
		icon = "ℹ️"
	default:
		icon = "⚠️"
	}

	return fmt.Sprintf("%s %s", icon, warning.Warning)
}

// checkDSTConflict checks if an event time falls in a DST conflict.
// Returns the warning if there's a conflict, nil otherwise.
func checkDSTConflict(eventTime time.Time, tz string, duration time.Duration) (*domain.DSTWarning, error) {
	if tz == "" {
		return nil, nil
	}

	tzService := timezone.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check for DST warning at event time (no warning window, only exact conflicts)
	warning, err := tzService.CheckDSTWarning(ctx, eventTime, tz, 0)
	if err != nil {
		return nil, err
	}

	// Only return warning if it's an actual conflict (gap or duplicate)
	if warning != nil && (warning.InTransitionGap || warning.InDuplicateHour) {
		return warning, nil
	}

	return nil, nil
}

// confirmDSTConflict displays a DST conflict warning and asks for user confirmation.
// Returns true if user wants to proceed, false if cancelled.
func confirmDSTConflict(warning *domain.DSTWarning) bool {
	if warning == nil {
		return true
	}

	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	fmt.Println()
	if warning.InTransitionGap {
		red.Println("⚠️  DST Conflict Detected!")
	} else {
		yellow.Println("⚠️  DST Conflict Detected!")
	}
	fmt.Println()

	fmt.Println(warning.Warning)
	fmt.Println()

	// Show suggested alternatives if available
	if warning.InTransitionGap {
		fmt.Println("Suggested alternatives:")
		fmt.Println("  1. Schedule 1 hour earlier (before DST)")
		fmt.Println("  2. Schedule at the requested time after DST")
		fmt.Println("  3. Use a different date")
		fmt.Println()
	} else if warning.InDuplicateHour {
		fmt.Println("Note: This time occurs twice due to falling back.")
		fmt.Println("The event will be created at the first occurrence.")
		fmt.Println()
	}

	// Ask for confirmation
	fmt.Print("Create anyway? [y/N]: ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)

	return strings.ToLower(confirm) == "y" || strings.ToLower(confirm) == "yes"
}

// ============================================================================
// Working Hours Validation
// ============================================================================

// checkWorkingHoursViolation checks if an event time falls outside working hours.
// Returns warning message if outside working hours, empty string otherwise.
func checkWorkingHoursViolation(eventTime time.Time, config *domain.Config) string {
	if config == nil || config.WorkingHours == nil {
		// No working hours configured, use defaults
		config = &domain.Config{
			WorkingHours: &domain.WorkingHoursConfig{},
		}
	}

	// Get schedule for the event's day of week
	weekday := strings.ToLower(eventTime.Weekday().String())
	schedule := config.WorkingHours.GetScheduleForDay(weekday)

	// If working hours not enabled for this day, no violation
	if schedule == nil || !schedule.Enabled {
		return ""
	}

	// Parse working hours
	startHour, startMin, err := parseTimeString(schedule.Start)
	if err != nil {
		return "" // Invalid config, skip validation
	}

	endHour, endMin, err := parseTimeString(schedule.End)
	if err != nil {
		return "" // Invalid config, skip validation
	}

	// Check if event time is outside working hours
	eventHour := eventTime.Hour()
	eventMin := eventTime.Minute()

	// Convert to minutes for easier comparison
	eventMinutes := eventHour*60 + eventMin
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	if eventMinutes < startMinutes || eventMinutes >= endMinutes {
		// Outside working hours
		var offset string
		if eventMinutes < startMinutes {
			hoursBefore := (startMinutes - eventMinutes) / 60
			minsBefore := (startMinutes - eventMinutes) % 60
			if minsBefore > 0 {
				offset = fmt.Sprintf("%dh %dm before start", hoursBefore, minsBefore)
			} else {
				offset = fmt.Sprintf("%d hour(s) before start", hoursBefore)
			}
		} else {
			hoursAfter := (eventMinutes - endMinutes) / 60
			minsAfter := (eventMinutes - endMinutes) % 60
			if minsAfter > 0 {
				offset = fmt.Sprintf("%dh %dm after end", hoursAfter, minsAfter)
			} else {
				offset = fmt.Sprintf("%d hour(s) after end", hoursAfter)
			}
		}

		return fmt.Sprintf("Event scheduled outside working hours (%s - %s) - %s",
			schedule.Start, schedule.End, offset)
	}

	return ""
}

// confirmWorkingHoursViolation displays a working hours warning and asks for confirmation.
// Returns true if user wants to proceed, false if cancelled.
func confirmWorkingHoursViolation(violation string, eventTime time.Time, schedule *domain.DaySchedule) bool {
	if violation == "" {
		return true
	}

	yellow := color.New(color.FgYellow, color.Bold)

	fmt.Println()
	yellow.Println("⚠️  Working Hours Warning")
	fmt.Println()

	fmt.Printf("This event is scheduled outside your working hours:\n")
	fmt.Printf("  • Your hours: %s - %s\n", schedule.Start, schedule.End)
	fmt.Printf("  • Event time: %s\n", eventTime.Format("3:04 PM MST"))
	fmt.Printf("  • %s\n", violation)
	fmt.Println()

	// Ask for confirmation
	fmt.Print("Create anyway? [y/N]: ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)

	return strings.ToLower(confirm) == "y" || strings.ToLower(confirm) == "yes"
}

// parseTimeString parses a time string in "HH:MM" format.
func parseTimeString(s string) (hour, min int, err error) {
	_, err = fmt.Sscanf(s, "%d:%d", &hour, &min)
	if err != nil {
		return 0, 0, err
	}
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return 0, 0, fmt.Errorf("invalid time")
	}
	return hour, min, nil
}

// checkBreakViolation checks if an event time falls during a break period.
// Returns error message if during break (hard block), empty string otherwise.
func checkBreakViolation(eventTime time.Time, config *domain.Config) string {
	if config == nil || config.WorkingHours == nil {
		return "" // No working hours or breaks configured
	}

	// Get schedule for the event's day of week
	weekday := strings.ToLower(eventTime.Weekday().String())
	schedule := config.WorkingHours.GetScheduleForDay(weekday)

	// If no schedule or breaks, no violation
	if schedule == nil || len(schedule.Breaks) == 0 {
		return ""
	}

	// Check each break period
	eventHour := eventTime.Hour()
	eventMin := eventTime.Minute()
	eventMinutes := eventHour*60 + eventMin

	for _, breakBlock := range schedule.Breaks {
		// Parse break start/end times
		startHour, startMin, err := parseTimeString(breakBlock.Start)
		if err != nil {
			continue // Skip invalid break config
		}

		endHour, endMin, err := parseTimeString(breakBlock.End)
		if err != nil {
			continue // Skip invalid break config
		}

		// Convert to minutes for comparison
		breakStart := startHour*60 + startMin
		breakEnd := endHour*60 + endMin

		// Check if event falls within this break period
		if eventMinutes >= breakStart && eventMinutes < breakEnd {
			return fmt.Sprintf("Event cannot be scheduled during %s (%s - %s)",
				breakBlock.Name, breakBlock.Start, breakBlock.End)
		}
	}

	return ""
}

// ============================================================================
// Natural Language Time Parsing
// ============================================================================

// ParsedTime represents a parsed natural language time expression.
type ParsedTime struct {
	Time     time.Time
	Timezone string
	Original string
}

// parseNaturalTime parses natural language time expressions.
// Supports formats like:
// - "tomorrow at 3pm"
// - "next Tuesday 2pm PST"
// - "Dec 25 10:00 AM"
// - "in 2 hours"
// - "2024-12-25 14:00"
func parseNaturalTime(input string, defaultTZ string) (*ParsedTime, error) {
	if input == "" {
		return nil, common.NewUserError(
			"time expression is empty",
			"Provide a time like 'tomorrow at 3pm' or 'Dec 25 10:00 AM'",
		)
	}

	// Default timezone if not specified
	if defaultTZ == "" {
		defaultTZ = getLocalTimeZone()
	}

	// Load the timezone
	loc, err := time.LoadLocation(defaultTZ)
	if err != nil {
		return nil, common.NewUserError(
			fmt.Sprintf("invalid timezone: %s", defaultTZ),
			"Use IANA timezone IDs like 'America/Los_Angeles'",
		)
	}

	now := time.Now().In(loc)
	normalizedInput := normalizeTimeString(input)

	// Try parsing in order of specificity
	// Note: Some parsers need normalized input, others need original
	parsers := []struct {
		fn            func(string, *time.Location, time.Time) (*ParsedTime, error)
		useNormalized bool
	}{
		{parseRelativeTime, true},
		{parseRelativeDayTime, true},
		{parseSpecificDayTime, true},
		{parseAbsoluteTime, false}, // Keep original for proper month name parsing
		{parseISOTime, false},      // Keep original for ISO formats
	}

	for _, parser := range parsers {
		inputToUse := input
		if parser.useNormalized {
			inputToUse = normalizedInput
		}
		result, err := parser.fn(inputToUse, loc, now)
		if err == nil && result != nil {
			result.Original = input
			return result, nil
		}
	}

	return nil, common.NewUserError(
		fmt.Sprintf("could not parse time: %s", input),
		"Try formats like:\n"+
			"  - tomorrow at 3pm\n"+
			"  - next Tuesday 2pm PST\n"+
			"  - Dec 25 10:00 AM\n"+
			"  - in 2 hours\n"+
			"  - 2024-12-25 14:00",
	)
}

// normalizeTimeString normalizes the input string for parsing.
func normalizeTimeString(s string) string {
	// Convert to lowercase for case-insensitive matching
	s = strings.ToLower(s)
	// Remove extra whitespace
	s = strings.TrimSpace(s)
	// Collapse multiple spaces into one
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// parseRelativeTime parses relative time expressions like "in 2 hours", "in 30 minutes".
func parseRelativeTime(input string, loc *time.Location, now time.Time) (*ParsedTime, error) {
	// Pattern: "in X hours/minutes/days"
	patterns := []struct {
		pattern string
		unit    time.Duration
	}{
		{"in %d hour", time.Hour},
		{"in %d hours", time.Hour},
		{"in %d minute", time.Minute},
		{"in %d minutes", time.Minute},
		{"in %d day", 24 * time.Hour},
		{"in %d days", 24 * time.Hour},
	}

	for _, p := range patterns {
		var amount int
		_, err := fmt.Sscanf(input, p.pattern, &amount)
		if err == nil {
			result := now.Add(time.Duration(amount) * p.unit)
			return &ParsedTime{
				Time:     result,
				Timezone: loc.String(),
			}, nil
		}
	}

	return nil, fmt.Errorf("not a relative time")
}

// parseRelativeDayTime parses relative day + time like "tomorrow at 3pm", "today at 2:30pm".
func parseRelativeDayTime(input string, loc *time.Location, now time.Time) (*ParsedTime, error) {
	relativeDays := map[string]int{
		"today":    0,
		"tomorrow": 1,
	}

	for day, offset := range relativeDays {
		if len(input) > len(day) && input[:len(day)] == day {
			// Extract the time part
			timePart := input[len(day):]
			timePart = strings.TrimSpace(timePart)

			// Remove "at" if present
			if len(timePart) > 3 && timePart[:3] == "at " {
				timePart = timePart[3:]
			}

			// Parse the time
			parsedTime, err := parseTimeOfDay(timePart, loc)
			if err != nil {
				return nil, err
			}

			// Set to the target day
			targetDay := now.AddDate(0, 0, offset)
			result := time.Date(
				targetDay.Year(),
				targetDay.Month(),
				targetDay.Day(),
				parsedTime.Hour(),
				parsedTime.Minute(),
				0, 0, loc,
			)

			return &ParsedTime{
				Time:     result,
				Timezone: loc.String(),
			}, nil
		}
	}

	return nil, fmt.Errorf("not a relative day time")
}

// parseSpecificDayTime parses specific weekday + time like "next Tuesday 2pm", "Monday at 10am".
func parseSpecificDayTime(input string, loc *time.Location, now time.Time) (*ParsedTime, error) {
	weekdays := map[string]time.Weekday{
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		"sunday":    time.Sunday,
	}

	// Check for "next" prefix
	isNext := false
	checkInput := input
	if len(input) > 5 && input[:5] == "next " {
		isNext = true
		checkInput = input[5:]
	}

	for dayName, weekday := range weekdays {
		if len(checkInput) > len(dayName) && checkInput[:len(dayName)] == dayName {
			// Extract time part
			timePart := checkInput[len(dayName):]
			timePart = strings.TrimSpace(timePart)

			// Remove "at" if present
			if len(timePart) > 3 && timePart[:3] == "at " {
				timePart = timePart[3:]
			}

			// Parse the time
			parsedTime, err := parseTimeOfDay(timePart, loc)
			if err != nil {
				return nil, err
			}

			// Find next occurrence of this weekday
			daysUntil := int(weekday - now.Weekday())
			if daysUntil <= 0 || isNext {
				daysUntil += 7
			}

			targetDay := now.AddDate(0, 0, daysUntil)
			result := time.Date(
				targetDay.Year(),
				targetDay.Month(),
				targetDay.Day(),
				parsedTime.Hour(),
				parsedTime.Minute(),
				0, 0, loc,
			)

			return &ParsedTime{
				Time:     result,
				Timezone: loc.String(),
			}, nil
		}
	}

	return nil, fmt.Errorf("not a specific day time")
}

// parseAbsoluteTime parses absolute dates like "Dec 25 10:00 AM", "December 25, 2024 2pm".
func parseAbsoluteTime(input string, loc *time.Location, now time.Time) (*ParsedTime, error) {
	// Common date/time formats - try both lowercase and titlecase
	formats := []string{
		// Lowercase formats (after normalization) - with leading zero for hours
		"jan 2 03:04 pm",
		"jan 2 03:04pm",
		"jan 2 3:04 pm",
		"jan 2 3:04pm",
		"jan 2, 2006 03:04 pm",
		"jan 2, 2006 3:04 pm",
		"january 2 03:04 pm",
		"january 2 3:04 pm",
		"january 2, 2006 03:04 pm",
		"january 2, 2006 3:04 pm",
		// Titlecase formats (original input)
		"Jan 2 03:04 PM",
		"Jan 2 3:04 PM",
		"Jan 2 03:04PM",
		"Jan 2 3:04PM",
		"Jan 2, 2006 03:04 PM",
		"Jan 2, 2006 3:04 PM",
		"January 2 03:04 PM",
		"January 2 3:04 PM",
		"January 2, 2006 03:04 PM",
		"January 2, 2006 3:04 PM",
		// Numeric formats
		"2006-01-02 15:04",
		"01/02/2006 03:04 PM",
		"01/02/2006 3:04 PM",
		"01/02/2006 15:04",
	}

	for _, format := range formats {
		t, err := time.ParseInLocation(format, input, loc)
		if err == nil {
			// If year is not in input, use current year
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			}
			return &ParsedTime{
				Time:     t,
				Timezone: loc.String(),
			}, nil
		}
	}

	return nil, fmt.Errorf("not an absolute time")
}

// parseISOTime parses ISO format times like "2024-12-25T14:00:00".
func parseISOTime(input string, loc *time.Location, now time.Time) (*ParsedTime, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}

	for _, format := range formats {
		t, err := time.ParseInLocation(format, input, loc)
		if err == nil {
			return &ParsedTime{
				Time:     t,
				Timezone: loc.String(),
			}, nil
		}
	}

	return nil, fmt.Errorf("not an ISO time")
}

// parseTimeOfDay parses time of day like "3pm", "2:30pm", "14:00".
func parseTimeOfDay(input string, loc *time.Location) (time.Time, error) {
	// Normalize to lowercase, then try both lowercase and uppercase formats
	originalInput := input
	lowerInput := strings.ToLower(input)

	formats := []string{
		"3pm",
		"3:04pm",
		"3 pm",
		"3:04 pm",
		"15:04",
	}

	// Try lowercase formats
	for _, format := range formats {
		t, err := time.ParseInLocation(format, lowerInput, loc)
		if err == nil {
			return t, nil
		}
	}

	// Try original input with uppercase formats (for backward compatibility)
	upperFormats := []string{
		"3PM",
		"3:04PM",
		"3 PM",
		"3:04 PM",
	}

	for _, format := range upperFormats {
		t, err := time.ParseInLocation(format, originalInput, loc)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", input)
}
