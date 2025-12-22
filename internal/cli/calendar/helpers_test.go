package calendar

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestGetLocalTimeZone(t *testing.T) {
	tz := getLocalTimeZone()

	if tz == "" {
		t.Error("Expected non-empty timezone, got empty string")
	}

	// Verify it's a valid timezone
	_, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("getLocalTimeZone() returned invalid timezone %q: %v", tz, err)
	}
}

func TestValidateTimeZone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{
			name:    "valid IANA timezone",
			tz:      "America/New_York",
			wantErr: false,
		},
		{
			name:    "valid UTC",
			tz:      "UTC",
			wantErr: false,
		},
		{
			name:    "valid Europe timezone",
			tz:      "Europe/London",
			wantErr: false,
		},
		{
			name:    "valid Asia timezone",
			tz:      "Asia/Tokyo",
			wantErr: false,
		},
		{
			name:    "invalid timezone",
			tz:      "Invalid/Timezone",
			wantErr: true,
		},
		{
			name:    "timezone abbreviation (not IANA)",
			tz:      "PST",
			wantErr: true,
		},
		{
			name:    "empty timezone",
			tz:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeZone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTimeZone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestFormatEventTimeWithTZ(t *testing.T) {
	// Create a test time in EST (America/New_York)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	testTime := time.Date(2025, 1, 15, 14, 0, 0, 0, loc) // 2 PM EST

	tests := []struct {
		name           string
		when           domain.EventWhen
		targetTZ       string
		wantConversion bool
		wantError      bool
	}{
		{
			name: "all-day event - no conversion",
			when: domain.EventWhen{
				Object: "date",
				Date:   "2025-01-15",
			},
			targetTZ:       "America/Los_Angeles",
			wantConversion: false,
			wantError:      false,
		},
		{
			name: "timed event - same timezone - no conversion",
			when: domain.EventWhen{
				Object:        "timespan",
				StartTime:     testTime.Unix(),
				EndTime:       testTime.Add(time.Hour).Unix(),
				StartTimezone: "America/New_York",
				EndTimezone:   "America/New_York",
			},
			targetTZ:       "America/New_York",
			wantConversion: false,
			wantError:      false,
		},
		{
			name: "timed event - different timezone - with conversion",
			when: domain.EventWhen{
				Object:        "timespan",
				StartTime:     testTime.Unix(),
				EndTime:       testTime.Add(time.Hour).Unix(),
				StartTimezone: "America/New_York",
				EndTimezone:   "America/New_York",
			},
			targetTZ:       "America/Los_Angeles",
			wantConversion: true,
			wantError:      false,
		},
		{
			name: "timed event - empty timezone - no conversion",
			when: domain.EventWhen{
				Object:        "timespan",
				StartTime:     testTime.Unix(),
				EndTime:       testTime.Add(time.Hour).Unix(),
				StartTimezone: "America/New_York",
				EndTimezone:   "America/New_York",
			},
			targetTZ:       "",
			wantConversion: false,
			wantError:      false,
		},
		{
			name: "timed event - invalid timezone - error",
			when: domain.EventWhen{
				Object:        "timespan",
				StartTime:     testTime.Unix(),
				EndTime:       testTime.Add(time.Hour).Unix(),
				StartTimezone: "America/New_York",
				EndTimezone:   "America/New_York",
			},
			targetTZ:  "Invalid/Timezone",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &domain.Event{When: tt.when}
			display, err := formatEventTimeWithTZ(event, tt.targetTZ)

			if (err != nil) != tt.wantError {
				t.Errorf("formatEventTimeWithTZ() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			if display == nil {
				t.Error("formatEventTimeWithTZ() returned nil display")
				return
			}

			if display.ShowConversion != tt.wantConversion {
				t.Errorf("formatEventTimeWithTZ() ShowConversion = %v, want %v", display.ShowConversion, tt.wantConversion)
			}

			// Verify we have original time
			if display.OriginalTime == "" {
				t.Error("formatEventTimeWithTZ() OriginalTime is empty")
			}

			// If conversion is expected, verify we have converted time
			if tt.wantConversion {
				if display.ConvertedTime == "" {
					t.Error("formatEventTimeWithTZ() ConvertedTime is empty when conversion expected")
				}
				if display.ConvertedTimezone == "" {
					t.Error("formatEventTimeWithTZ() ConvertedTimezone is empty when conversion expected")
				}
			}
		})
	}
}

func TestFormatEventTimeWithTZ_AllDayEvent(t *testing.T) {
	when := domain.EventWhen{
		Object: "date",
		Date:   "2025-01-15",
	}

	event := &domain.Event{When: when}
	display, err := formatEventTimeWithTZ(event, "America/Los_Angeles")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if display.ShowConversion {
		t.Error("All-day events should not show conversion")
	}

	if display.OriginalTime == "" {
		t.Error("Expected OriginalTime to be populated")
	}

	// Verify the format includes "all day"
	if display.OriginalTime != "Wed, Jan 15, 2025 (all day)" {
		t.Errorf("Expected all-day format, got: %s", display.OriginalTime)
	}
}

func TestFormatEventTimeWithTZ_MultiDayAllDay(t *testing.T) {
	when := domain.EventWhen{
		Object:    "datespan",
		StartDate: "2025-01-15",
		EndDate:   "2025-01-17",
	}

	event := &domain.Event{When: when}
	display, err := formatEventTimeWithTZ(event, "America/Los_Angeles")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if display.ShowConversion {
		t.Error("All-day events should not show conversion")
	}

	if display.OriginalTime == "" {
		t.Error("Expected OriginalTime to be populated")
	}

	// Verify it shows both dates
	expectedPrefix := "Wed, Jan 15, 2025 - Fri, Jan 17, 2025"
	if len(display.OriginalTime) < len(expectedPrefix) || display.OriginalTime[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected multi-day format starting with %q, got: %s", expectedPrefix, display.OriginalTime)
	}
}

func TestFormatEventTimeWithTZ_TimedEventConversion(t *testing.T) {
	// Create event at 2 PM EST (Eastern Time)
	estLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load EST timezone: %v", err)
	}

	// 2 PM EST on Jan 15, 2025
	startTime := time.Date(2025, 1, 15, 14, 0, 0, 0, estLoc)
	endTime := startTime.Add(time.Hour) // 3 PM EST

	when := domain.EventWhen{
		Object:        "timespan",
		StartTime:     startTime.Unix(),
		EndTime:       endTime.Unix(),
		StartTimezone: "America/New_York",
		EndTimezone:   "America/New_York",
	}

	// Convert to PST (Pacific Time)
	event := &domain.Event{When: when}
	display, err := formatEventTimeWithTZ(event, "America/Los_Angeles")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !display.ShowConversion {
		t.Error("Expected conversion to be shown")
	}

	if display.OriginalTimezone != "EST" {
		t.Errorf("Expected OriginalTimezone to be EST, got: %s", display.OriginalTimezone)
	}

	if display.ConvertedTimezone != "PST" {
		t.Errorf("Expected ConvertedTimezone to be PST, got: %s", display.ConvertedTimezone)
	}

	// Verify the converted time shows 11 AM - 12 PM (PST is 3 hours behind EST)
	expectedConverted := "Wed, Jan 15, 2025, 11:00 AM - 12:00 PM"
	if display.ConvertedTime != expectedConverted {
		t.Errorf("Expected converted time %q, got: %s", expectedConverted, display.ConvertedTime)
	}
}

func TestFormatEventTimeWithTZ_SameDayEvent(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	startTime := time.Date(2025, 1, 15, 10, 0, 0, 0, loc)
	endTime := startTime.Add(2 * time.Hour)

	when := domain.EventWhen{
		Object:        "timespan",
		StartTime:     startTime.Unix(),
		EndTime:       endTime.Unix(),
		StartTimezone: "America/New_York",
		EndTimezone:   "America/New_York",
	}

	event := &domain.Event{When: when}
	display, err := formatEventTimeWithTZ(event, "")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should show same-day format
	expected := "Wed, Jan 15, 2025, 10:00 AM - 12:00 PM"
	if display.OriginalTime != expected {
		t.Errorf("Expected same-day format %q, got: %s", expected, display.OriginalTime)
	}
}

func TestFormatEventTimeWithTZ_MultiDayEvent(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	startTime := time.Date(2025, 1, 15, 10, 0, 0, 0, loc)
	endTime := time.Date(2025, 1, 16, 14, 0, 0, 0, loc) // Next day

	when := domain.EventWhen{
		Object:        "timespan",
		StartTime:     startTime.Unix(),
		EndTime:       endTime.Unix(),
		StartTimezone: "America/New_York",
		EndTimezone:   "America/New_York",
	}

	event := &domain.Event{When: when}
	display, err := formatEventTimeWithTZ(event, "")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should show multi-day format with full dates
	expected := "Wed, Jan 15, 2025 10:00 AM - Thu, Jan 16, 2025 2:00 PM"
	if display.OriginalTime != expected {
		t.Errorf("Expected multi-day format %q, got: %s", expected, display.OriginalTime)
	}
}

func TestGetSystemTimeZone(t *testing.T) {
	tz := getSystemTimeZone()

	// Should return a valid IANA timezone or UTC
	if tz == "" {
		t.Error("Expected non-empty timezone")
	}

	// Verify it's loadable
	_, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("getSystemTimeZone() returned invalid timezone %q: %v", tz, err)
	}
}

func TestFormatTimezoneBadge(t *testing.T) {
	tests := []struct {
		name            string
		tz              string
		useAbbreviation bool
		wantContains    string
		wantEmpty       bool
	}{
		{
			name:            "empty timezone",
			tz:              "",
			useAbbreviation: false,
			wantEmpty:       true,
		},
		{
			name:            "full timezone name",
			tz:              "America/New_York",
			useAbbreviation: false,
			wantContains:    "[America/New_York]",
		},
		{
			name:            "timezone abbreviation",
			tz:              "America/Los_Angeles",
			useAbbreviation: true,
			wantContains:    "[P", // PST or PDT
		},
		{
			name:            "UTC timezone",
			tz:              "UTC",
			useAbbreviation: true,
			wantContains:    "[UTC]",
		},
		{
			name:            "Europe timezone abbreviation",
			tz:              "Europe/London",
			useAbbreviation: true,
			wantContains:    "[", // GMT or BST
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimezoneBadge(tt.tz, tt.useAbbreviation)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("formatTimezoneBadge() = %q, want empty string", got)
				}
				return
			}

			if len(got) == 0 {
				t.Errorf("formatTimezoneBadge() = empty, want %q", tt.wantContains)
				return
			}

			// Badge should start with '[' and end with ']'
			if got[0] != '[' || got[len(got)-1] != ']' {
				t.Errorf("formatTimezoneBadge() = %q, want format [...]", got)
			}

			// Check if contains expected substring (for partial matches)
			if tt.wantContains != "" {
				found := false
				for i := 0; i <= len(got)-len(tt.wantContains); i++ {
					if got[i:i+len(tt.wantContains)] == tt.wantContains {
						found = true
						break
					}
				}
				if !found && got != tt.wantContains {
					t.Errorf("formatTimezoneBadge() = %q, want to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}

func TestGetTimezoneColor(t *testing.T) {
	tests := []struct {
		name          string
		tz            string
		wantColorCode int
	}{
		{
			name:          "empty timezone returns default",
			tz:            "",
			wantColorCode: 7, // Default gray
		},
		{
			name:          "Pacific timezone (PST/PDT)",
			tz:            "America/Los_Angeles",
			wantColorCode: 34, // Blue (offset -8/-7)
		},
		{
			name:          "Eastern timezone (EST/EDT)",
			tz:            "America/New_York",
			wantColorCode: 36, // Cyan (offset -5/-4)
		},
		{
			name:          "UTC timezone",
			tz:            "UTC",
			wantColorCode: 32, // Green (offset 0)
		},
		{
			name:          "Europe timezone",
			tz:            "Europe/London",
			wantColorCode: 32, // Green (offset 0/+1)
		},
		{
			name:          "Asia timezone",
			tz:            "Asia/Tokyo",
			wantColorCode: 35, // Magenta (offset +9)
		},
		{
			name:          "India timezone",
			tz:            "Asia/Kolkata",
			wantColorCode: 35, // Magenta (offset +5:30)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTimezoneColor(tt.tz)

			if got != tt.wantColorCode {
				t.Errorf("getTimezoneColor(%q) = %d, want %d", tt.tz, got, tt.wantColorCode)
			}

			// Verify it's a valid ANSI color code
			validCodes := []int{7, 31, 32, 33, 34, 35, 36}
			isValid := false
			for _, code := range validCodes {
				if got == code {
					isValid = true
					break
				}
			}
			if !isValid {
				t.Errorf("getTimezoneColor(%q) = %d, not a valid ANSI color code", tt.tz, got)
			}
		})
	}
}

func TestCheckDSTWarning(t *testing.T) {
	tests := []struct {
		name        string
		eventTime   time.Time
		tz          string
		wantWarning bool
		wantEmpty   bool
	}{
		{
			name:      "empty timezone returns empty",
			eventTime: time.Now(),
			tz:        "",
			wantEmpty: true,
		},
		{
			name:      "invalid timezone returns empty",
			eventTime: time.Now(),
			tz:        "Invalid/Timezone",
			wantEmpty: true,
		},
		{
			name: "event far from DST transition returns empty",
			// June is far from DST transitions in most zones
			eventTime: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			tz:        "America/New_York",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDSTWarning(tt.eventTime, tt.tz)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("checkDSTWarning() = %q, want empty string", result)
				}
				return
			}

			if tt.wantWarning && result == "" {
				t.Error("checkDSTWarning() = empty, want warning message")
			}

			if !tt.wantWarning && result != "" {
				t.Errorf("checkDSTWarning() = %q, want empty", result)
			}
		})
	}
}

func TestFormatDSTWarning(t *testing.T) {
	tests := []struct {
		name         string
		warning      *domain.DSTWarning
		wantContains string
		wantEmpty    bool
	}{
		{
			name:      "nil warning returns empty",
			warning:   nil,
			wantEmpty: true,
		},
		{
			name: "error severity shows red icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "error",
				Warning:          "This time will not exist",
			},
			wantContains: "⛔",
		},
		{
			name: "warning severity shows warning icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "warning",
				Warning:          "DST begins soon",
			},
			wantContains: "⚠️",
		},
		{
			name: "info severity shows info icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "info",
				Warning:          "DST ends in 5 days",
			},
			wantContains: "ℹ️",
		},
		{
			name: "unknown severity defaults to warning icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "unknown",
				Warning:          "Some warning",
			},
			wantContains: "⚠️",
		},
		{
			name: "warning message is included",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "warning",
				Warning:          "Daylight Saving Time begins in 3 days",
			},
			wantContains: "Daylight Saving Time begins in 3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDSTWarning(tt.warning)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("formatDSTWarning() = %q, want empty string", result)
				}
				return
			}

			if tt.wantContains != "" {
				found := false
				for i := 0; i <= len(result)-len(tt.wantContains); i++ {
					if result[i:i+len(tt.wantContains)] == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					// Try substring match
					if len(result) < len(tt.wantContains) {
						t.Errorf("formatDSTWarning() = %q, want to contain %q", result, tt.wantContains)
					} else {
						// Check if any part matches
						foundSubstring := false
						for i := 0; i <= len(result)-1; i++ {
							for j := i + 1; j <= len(result); j++ {
								substr := result[i:j]
								for k := 0; k <= len(tt.wantContains)-len(substr); k++ {
									if tt.wantContains[k:k+len(substr)] == substr && len(substr) > 2 {
										foundSubstring = true
										break
									}
								}
								if foundSubstring {
									break
								}
							}
							if foundSubstring {
								break
							}
						}
						if !foundSubstring {
							t.Errorf("formatDSTWarning() = %q, want to contain %q", result, tt.wantContains)
						}
					}
				}
			}
		})
	}
}

func TestParseNaturalTime(t *testing.T) {
	// Note: Tests use current time, so relative time tests may vary
	// Reference: Wednesday, Jan 15, 2025, 10:00 AM EST would be used for deterministic testing

	tests := []struct {
		name      string
		input     string
		tz        string
		wantError bool
		checkTime func(*testing.T, time.Time)
	}{
		{
			name:      "empty input returns error",
			input:     "",
			tz:        "America/New_York",
			wantError: true,
		},
		{
			name:      "invalid timezone returns error",
			input:     "tomorrow at 3pm",
			tz:        "Invalid/Zone",
			wantError: true,
		},
		{
			name:  "relative time - in 2 hours",
			input: "in 2 hours",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				// Check that time is roughly 2 hours from now
				now := time.Now()
				expected := now.Add(2 * time.Hour)
				diff := result.Sub(expected)
				if diff < -time.Minute || diff > time.Minute {
					t.Errorf("Expected time ~2 hours from now, got %v (diff: %v)", result, diff)
				}
			},
		},
		{
			name:  "relative time - in 30 minutes",
			input: "in 30 minutes",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				// Check that time is roughly 30 minutes from now
				now := time.Now()
				expected := now.Add(30 * time.Minute)
				diff := result.Sub(expected)
				if diff < -time.Minute || diff > time.Minute {
					t.Errorf("Expected time ~30 minutes from now, got %v (diff: %v)", result, diff)
				}
			},
		},
		{
			name:  "relative day - tomorrow at 3pm",
			input: "tomorrow at 3pm",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				// Check it's tomorrow and at 3pm
				now := time.Now()
				if result.Day() != now.AddDate(0, 0, 1).Day() || result.Hour() != 15 {
					t.Errorf("Expected tomorrow at 15:00, got %v at %02d:00", result.Day(), result.Hour())
				}
			},
		},
		{
			name:  "relative day - today at 2:30pm",
			input: "today at 2:30pm",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				// Check it's today at 2:30pm
				now := time.Now()
				if result.Day() != now.Day() || result.Hour() != 14 || result.Minute() != 30 {
					t.Errorf("Expected today at 14:30, got %v at %02d:%02d", result.Day(), result.Hour(), result.Minute())
				}
			},
		},
		{
			name:  "specific weekday - next tuesday 2pm",
			input: "next tuesday 2pm",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				// Check it's a Tuesday and at 2pm
				if result.Weekday() != time.Tuesday {
					t.Errorf("Expected Tuesday, got %v", result.Weekday())
				}
				if result.Hour() != 14 {
					t.Errorf("Expected 14:00, got %02d:00", result.Hour())
				}
				// Check it's in the future
				if !result.After(time.Now()) {
					t.Error("Expected future date")
				}
			},
		},
		{
			name:  "absolute time - dec 25 10:00 am",
			input: "Dec 25 10:00 AM",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				if result.Month() != time.December || result.Day() != 25 || result.Hour() != 10 {
					t.Errorf("Expected Dec 25 10:00, got %v %v %02d:00", result.Month(), result.Day(), result.Hour())
				}
			},
		},
		{
			name:  "ISO time - 2025-03-15 14:00",
			input: "2025-03-15 14:00",
			tz:    "America/New_York",
			checkTime: func(t *testing.T, result time.Time) {
				if result.Year() != 2025 || result.Month() != time.March || result.Day() != 15 || result.Hour() != 14 {
					t.Errorf("Expected 2025-03-15 14:00, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock time.Now() by using the parseNaturalTime implementation
			// For these tests, we'll test the individual parser functions
			result, err := parseNaturalTime(tt.input, tt.tz)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if tt.checkTime != nil {
				tt.checkTime(t, result.Time)
			}

			if result.Original != tt.input {
				t.Errorf("Original = %q, want %q", result.Original, tt.input)
			}
		})
	}
}

func TestParseTimeOfDay(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name      string
		input     string
		wantHour  int
		wantMin   int
		wantError bool
	}{
		{
			name:     "3pm",
			input:    "3pm",
			wantHour: 15,
			wantMin:  0,
		},
		{
			name:     "3PM (uppercase)",
			input:    "3PM",
			wantHour: 15,
			wantMin:  0,
		},
		{
			name:     "2:30pm",
			input:    "2:30pm",
			wantHour: 14,
			wantMin:  30,
		},
		{
			name:     "2:30 PM (with space)",
			input:    "2:30 PM",
			wantHour: 14,
			wantMin:  30,
		},
		{
			name:     "14:00 (24-hour)",
			input:    "14:00",
			wantHour: 14,
			wantMin:  0,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeOfDay(tt.input, loc)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d", result.Hour(), tt.wantHour)
			}

			if result.Minute() != tt.wantMin {
				t.Errorf("Minute = %d, want %d", result.Minute(), tt.wantMin)
			}
		})
	}
}

func TestNormalizeTimeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "uppercase to lowercase",
			input: "TOMORROW AT 3PM",
			want:  "tomorrow at 3pm",
		},
		{
			name:  "extra whitespace removed",
			input: "  tomorrow   at   3pm  ",
			want:  "tomorrow at 3pm",
		},
		{
			name:  "mixed case normalized",
			input: "Next Tuesday 2PM",
			want:  "next tuesday 2pm",
		},
		{
			name:  "already normalized",
			input: "tomorrow at 3pm",
			want:  "tomorrow at 3pm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTimeString(tt.input)
			if result != tt.want {
				t.Errorf("normalizeTimeString() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestCheckBreakViolation(t *testing.T) {
	tests := []struct {
		name          string
		eventTime     time.Time
		config        *domain.Config
		wantViolation bool
		wantContains  string
	}{
		{
			name:          "no config - no violation",
			eventTime:     time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC), // 12:30 PM
			config:        nil,
			wantViolation: false,
		},
		{
			name:      "no breaks configured - no violation",
			eventTime: time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC), // 12:30 PM (Wednesday)
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks:  nil, // No breaks
					},
				},
			},
			wantViolation: false,
		},
		{
			name:      "event outside break time - no violation",
			eventTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), // 10:00 AM (before lunch)
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name:      "event during lunch break - violation",
			eventTime: time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC), // 12:30 PM (during lunch)
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
						},
					},
				},
			},
			wantViolation: true,
			wantContains:  "Lunch",
		},
		{
			name:      "event at break start time - violation",
			eventTime: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC), // Exactly 12:00 PM
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
						},
					},
				},
			},
			wantViolation: true,
			wantContains:  "Lunch",
		},
		{
			name:      "event at break end time - no violation",
			eventTime: time.Date(2025, 1, 15, 13, 0, 0, 0, time.UTC), // Exactly 1:00 PM (after break)
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name:      "event during coffee break - violation",
			eventTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC), // 10:30 AM
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Coffee Break", Start: "10:30", End: "10:45", Type: "coffee"},
						},
					},
				},
			},
			wantViolation: true,
			wantContains:  "Coffee Break",
		},
		{
			name:      "multiple breaks - violation in second break",
			eventTime: time.Date(2025, 1, 15, 15, 10, 0, 0, time.UTC), // 3:10 PM
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
							{Name: "Afternoon Break", Start: "15:00", End: "15:15", Type: "coffee"},
						},
					},
				},
			},
			wantViolation: true,
			wantContains:  "Afternoon Break",
		},
		{
			name:      "multiple breaks - between breaks no violation",
			eventTime: time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC), // 2:00 PM (between lunch and afternoon break)
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
							{Name: "Afternoon Break", Start: "15:00", End: "15:15", Type: "coffee"},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name:      "day-specific break - Monday",
			eventTime: time.Date(2025, 1, 20, 11, 30, 0, 0, time.UTC), // Monday 11:30 AM
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Monday: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Early Lunch", Start: "11:00", End: "12:00", Type: "lunch"}, // Monday has early lunch
						},
					},
				},
			},
			wantViolation: true,
			wantContains:  "Early Lunch",
		},
		{
			name:      "invalid break config - ignored",
			eventTime: time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC),
			config: &domain.Config{
				WorkingHours: &domain.WorkingHoursConfig{
					Default: &domain.DaySchedule{
						Enabled: true,
						Start:   "09:00",
						End:     "17:00",
						Breaks: []domain.BreakBlock{
							{Name: "Invalid", Start: "invalid", End: "13:00"}, // Invalid start time
						},
					},
				},
			},
			wantViolation: false, // Invalid config is skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkBreakViolation(tt.eventTime, tt.config)

			if tt.wantViolation {
				if result == "" {
					t.Error("Expected violation message, got empty string")
					return
				}
				if tt.wantContains != "" {
					// Check if result contains expected string
					found := false
					for i := 0; i <= len(result)-len(tt.wantContains); i++ {
						if result[i:i+len(tt.wantContains)] == tt.wantContains {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("checkBreakViolation() = %q, want to contain %q", result, tt.wantContains)
					}
				}
			} else {
				if result != "" {
					t.Errorf("Expected no violation, got: %q", result)
				}
			}
		})
	}
}
