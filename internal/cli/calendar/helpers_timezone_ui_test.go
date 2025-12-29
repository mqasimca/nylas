package calendar

import (
	"testing"
)

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
