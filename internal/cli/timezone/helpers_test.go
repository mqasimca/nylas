package timezone

import (
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/cli/common"
)

func TestParseTimeZones(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single zone",
			input:    "America/New_York",
			expected: []string{"America/New_York"},
		},
		{
			name:     "multiple zones",
			input:    "America/New_York,Europe/London,Asia/Tokyo",
			expected: []string{"America/New_York", "Europe/London", "Asia/Tokyo"},
		},
		{
			name:     "zones with spaces",
			input:    "America/New_York, Europe/London , Asia/Tokyo",
			expected: []string{"America/New_York", "Europe/London", "Asia/Tokyo"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "America/New_York,Europe/London,",
			expected: []string{"America/New_York", "Europe/London"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimeZones(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d zones, got %d", len(tt.expected), len(result))
				return
			}

			for i, zone := range result {
				if zone != tt.expected[i] {
					t.Errorf("Expected zone[%d] = %q, got %q", i, tt.expected[i], zone)
				}
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "hours",
			input:    "1h",
			expected: 1 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "minutes",
			input:    "30m",
			expected: 30 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "hours and minutes",
			input:    "1h30m",
			expected: 90 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "seconds",
			input:    "45s",
			expected: 45 * time.Second,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := common.ParseDuration(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseWorkingHours(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		end       string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			name:      "valid hours",
			start:     "09:00",
			end:       "17:00",
			wantStart: "09:00",
			wantEnd:   "17:00",
			wantErr:   false,
		},
		{
			name:      "midnight to noon",
			start:     "00:00",
			end:       "12:00",
			wantStart: "00:00",
			wantEnd:   "12:00",
			wantErr:   false,
		},
		{
			name:    "invalid start format",
			start:   "abc",
			end:     "17:00",
			wantErr: true,
		},
		{
			name:    "invalid end format",
			start:   "09:00",
			end:     "5pm",
			wantErr: true,
		},
		{
			name:    "invalid hour",
			start:   "25:00",
			end:     "17:00",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			start:   "09:60",
			end:     "17:00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseWorkingHours(tt.start, tt.end)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if start != tt.wantStart {
				t.Errorf("Expected start = %q, got %q", tt.wantStart, start)
			}

			if end != tt.wantEnd {
				t.Errorf("Expected end = %q, got %q", tt.wantEnd, end)
			}
		})
	}
}

func TestFormatOffset(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "UTC",
			seconds:  0,
			expected: "UTC+0",
		},
		{
			name:     "positive hours only",
			seconds:  5 * 3600,
			expected: "UTC+5",
		},
		{
			name:     "negative hours only",
			seconds:  -8 * 3600,
			expected: "UTC-8",
		},
		{
			name:     "positive with minutes",
			seconds:  5*3600 + 30*60,
			expected: "UTC+5:30",
		},
		{
			name:     "negative with minutes",
			seconds:  -9*3600 - 30*60,
			expected: "UTC-9:30",
		},
		{
			name:     "positive 12 hours",
			seconds:  12 * 3600,
			expected: "UTC+12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOffset(tt.seconds)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNormalizeTimeZone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PST abbreviation",
			input:    "PST",
			expected: "America/Los_Angeles",
		},
		{
			name:     "EST abbreviation",
			input:    "EST",
			expected: "America/New_York",
		},
		{
			name:     "IST abbreviation",
			input:    "IST",
			expected: "Asia/Kolkata",
		},
		{
			name:     "lowercase pst",
			input:    "pst",
			expected: "America/Los_Angeles",
		},
		{
			name:     "mixed case EsT",
			input:    "EsT",
			expected: "America/New_York",
		},
		{
			name:     "IANA name unchanged",
			input:    "America/Chicago",
			expected: "America/Chicago",
		},
		{
			name:     "unknown abbreviation",
			input:    "XYZ",
			expected: "XYZ",
		},
		{
			name:     "GMT",
			input:    "GMT",
			expected: "Europe/London",
		},
		{
			name:     "JST",
			input:    "JST",
			expected: "Asia/Tokyo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTimeZone(tt.input)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2025, 1, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		showZone bool
		contains []string
	}{
		{
			name:     "without zone",
			showZone: false,
			contains: []string{"2025-01-15", "14:30:45"},
		},
		{
			name:     "with zone",
			showZone: true,
			contains: []string{"2025-01-15", "14:30:45", "UTC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(testTime, tt.showZone)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

func TestPrintTable(t *testing.T) {
	// This test just ensures the function doesn't panic
	headers := []string{"Col1", "Col2", "Col3"}
	rows := [][]string{
		{"a", "b", "c"},
		{"longer", "text", "here"},
		{"x", "y", "z"},
	}

	// Should not panic
	printTable(headers, rows)
}

func TestGroupZonesByRegion(t *testing.T) {
	zones := []string{
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"UTC",
	}

	grouped := groupZonesByRegion(zones)

	// Check that all regions exist
	expectedRegions := []string{"America", "Europe", "Asia", "Other"}
	for _, region := range expectedRegions {
		if _, ok := grouped[region]; !ok {
			t.Errorf("Expected region %q not found in grouped zones", region)
		}
	}

	// Check America region
	if len(grouped["America"]) != 2 {
		t.Errorf("Expected 2 zones in America region, got %d", len(grouped["America"]))
	}

	// Check Europe region
	if len(grouped["Europe"]) != 2 {
		t.Errorf("Expected 2 zones in Europe region, got %d", len(grouped["Europe"]))
	}

	// Check Asia region
	if len(grouped["Asia"]) != 1 {
		t.Errorf("Expected 1 zone in Asia region, got %d", len(grouped["Asia"]))
	}

	// Check Other region (UTC has no slash)
	if len(grouped["Other"]) != 1 {
		t.Errorf("Expected 1 zone in Other region, got %d", len(grouped["Other"]))
	}
}
