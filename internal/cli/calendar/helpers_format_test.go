package calendar

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

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
