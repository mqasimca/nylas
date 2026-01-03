//go:build !integration

package air

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimeString_Extended(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHour int
		wantMin  int
		wantOK   bool
	}{
		// Simple hours
		{
			name:     "simple hour 9",
			input:    "9",
			wantHour: 9,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "24-hour format 14",
			input:    "14",
			wantHour: 14,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "hour with am",
			input:    "9am",
			wantHour: 9,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "hour with pm",
			input:    "3pm",
			wantHour: 15,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "12pm is noon",
			input:    "12pm",
			wantHour: 12,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "12am is midnight",
			input:    "12am",
			wantHour: 0,
			wantMin:  0,
			wantOK:   true,
		},

		// Hour:minute formats
		{
			name:     "hour:minute 24h",
			input:    "14:30",
			wantHour: 14,
			wantMin:  30,
			wantOK:   true,
		},
		{
			name:     "hour:minute with pm",
			input:    "3:30pm",
			wantHour: 15,
			wantMin:  30,
			wantOK:   true,
		},
		{
			name:     "hour:minute with am",
			input:    "9:15am",
			wantHour: 9,
			wantMin:  15,
			wantOK:   true,
		},
		{
			name:     "midnight with colon",
			input:    "0:00",
			wantHour: 0,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "end of day",
			input:    "23:59",
			wantHour: 23,
			wantMin:  59,
			wantOK:   true,
		},

		// Edge cases - uppercase
		{
			name:     "uppercase AM",
			input:    "9AM",
			wantHour: 9,
			wantMin:  0,
			wantOK:   true,
		},
		{
			name:     "uppercase PM",
			input:    "3PM",
			wantHour: 15,
			wantMin:  0,
			wantOK:   true,
		},

		// Invalid inputs
		{
			name:   "invalid - negative hour",
			input:  "-1",
			wantOK: false,
		},
		{
			name:   "invalid - hour too high",
			input:  "25",
			wantOK: false,
		},
		{
			name:   "invalid - minute too high",
			input:  "10:60",
			wantOK: false,
		},
		{
			name:   "invalid - not a number",
			input:  "abc",
			wantOK: false,
		},
		{
			name:   "invalid - too many colons",
			input:  "10:30:45",
			wantOK: false,
		},
		{
			name:   "invalid - empty string",
			input:  "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour, min, ok := parseTimeString(tt.input)

			assert.Equal(t, tt.wantOK, ok, "ok mismatch")
			if tt.wantOK {
				assert.Equal(t, tt.wantHour, hour, "hour mismatch")
				assert.Equal(t, tt.wantMin, min, "minute mismatch")
			}
		})
	}
}

func TestParseNaturalDuration_RelativeDurations_Extended(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		input         string
		expectedDelta time.Duration
		tolerance     time.Duration
	}{
		{
			name:          "1 hour",
			input:         "1h",
			expectedDelta: 1 * time.Hour,
			tolerance:     time.Second,
		},
		{
			name:          "5 hours",
			input:         "5h",
			expectedDelta: 5 * time.Hour,
			tolerance:     time.Second,
		},
		{
			name:          "1 day",
			input:         "1d",
			expectedDelta: 24 * time.Hour,
			tolerance:     time.Second,
		},
		{
			name:          "3 days",
			input:         "3d",
			expectedDelta: 72 * time.Hour,
			tolerance:     time.Second,
		},
		{
			name:          "1 week",
			input:         "1w",
			expectedDelta: 7 * 24 * time.Hour,
			tolerance:     time.Second,
		},
		{
			name:          "30 minutes",
			input:         "30m",
			expectedDelta: 30 * time.Minute,
			tolerance:     time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNaturalDuration(tt.input)

			require.NoError(t, err)

			expected := now.Add(tt.expectedDelta).Unix()
			diff := result - expected
			assert.True(t, diff >= -1 && diff <= 1,
				"Expected around %d, got %d (diff: %d)", expected, result, diff)
		})
	}
}

func TestParseNaturalDuration_SpecialKeywords_Extended(t *testing.T) {
	now := time.Now()

	t.Run("later today", func(t *testing.T) {
		result, err := parseNaturalDuration("later today")
		require.NoError(t, err)

		// Should be 4 hours from now or 5 PM, whichever is first
		fourHoursLater := now.Add(4 * time.Hour).Unix()
		fivePM := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location()).Unix()

		// Result should be one of these two
		assert.True(t, result == fourHoursLater || result == fivePM || (result >= fivePM-1 && result <= fivePM+1),
			"Expected around %d or %d, got %d", fourHoursLater, fivePM, result)
	})

	t.Run("later", func(t *testing.T) {
		result, err := parseNaturalDuration("later")
		require.NoError(t, err)

		// Same as "later today"
		fourHoursLater := now.Add(4 * time.Hour).Unix()
		fivePM := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location()).Unix()

		assert.True(t, result == fourHoursLater || result == fivePM || (result >= fivePM-1 && result <= fivePM+1),
			"Expected around %d or %d, got %d", fourHoursLater, fivePM, result)
	})

	t.Run("tonight", func(t *testing.T) {
		result, err := parseNaturalDuration("tonight")
		require.NoError(t, err)

		// Should be 8 PM today or tomorrow if past 8 PM
		tonight := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		if tonight.Before(now) {
			tonight = tonight.Add(24 * time.Hour)
		}

		diff := result - tonight.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", tonight.Unix(), result, diff)
	})

	t.Run("tomorrow basic", func(t *testing.T) {
		result, err := parseNaturalDuration("tomorrow")
		require.NoError(t, err)

		// Should be 9 AM tomorrow
		tomorrow9AM := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())

		diff := result - tomorrow9AM.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", tomorrow9AM.Unix(), result, diff)
	})

	t.Run("tomorrow with time", func(t *testing.T) {
		result, err := parseNaturalDuration("tomorrow 2pm")
		require.NoError(t, err)

		// Should be 2 PM tomorrow
		tomorrow2PM := time.Date(now.Year(), now.Month(), now.Day()+1, 14, 0, 0, 0, now.Location())

		diff := result - tomorrow2PM.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", tomorrow2PM.Unix(), result, diff)
	})

	t.Run("tomorrow at time", func(t *testing.T) {
		result, err := parseNaturalDuration("tomorrow at 3:30pm")
		require.NoError(t, err)

		// Should be 3:30 PM tomorrow
		tomorrowTime := time.Date(now.Year(), now.Month(), now.Day()+1, 15, 30, 0, 0, now.Location())

		diff := result - tomorrowTime.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", tomorrowTime.Unix(), result, diff)
	})

	t.Run("next week", func(t *testing.T) {
		result, err := parseNaturalDuration("next week")
		require.NoError(t, err)

		// Should be Monday 9 AM
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextMonday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 9, 0, 0, 0, now.Location())

		diff := result - nextMonday.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", nextMonday.Unix(), result, diff)
	})

	t.Run("monday", func(t *testing.T) {
		result, err := parseNaturalDuration("monday")
		require.NoError(t, err)

		// Should be Monday 9 AM (same as next week)
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextMonday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 9, 0, 0, 0, now.Location())

		diff := result - nextMonday.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", nextMonday.Unix(), result, diff)
	})

	t.Run("weekend", func(t *testing.T) {
		result, err := parseNaturalDuration("weekend")
		require.NoError(t, err)

		// Should be Saturday 10 AM
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		saturday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSaturday, 10, 0, 0, 0, now.Location())

		diff := result - saturday.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", saturday.Unix(), result, diff)
	})

	t.Run("this weekend", func(t *testing.T) {
		result, err := parseNaturalDuration("this weekend")
		require.NoError(t, err)

		// Should be Saturday 10 AM
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		saturday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSaturday, 10, 0, 0, 0, now.Location())

		diff := result - saturday.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", saturday.Unix(), result, diff)
	})

	t.Run("saturday", func(t *testing.T) {
		result, err := parseNaturalDuration("saturday")
		require.NoError(t, err)

		// Should be Saturday 10 AM
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		saturday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSaturday, 10, 0, 0, 0, now.Location())

		diff := result - saturday.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", saturday.Unix(), result, diff)
	})
}

func TestParseNaturalDuration_SpecificTimes_Extended(t *testing.T) {
	now := time.Now()

	t.Run("specific time 9am", func(t *testing.T) {
		result, err := parseNaturalDuration("9am")
		require.NoError(t, err)

		// Should be 9 AM today or tomorrow
		target := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}

		diff := result - target.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", target.Unix(), result, diff)
	})

	t.Run("specific time 14:30", func(t *testing.T) {
		result, err := parseNaturalDuration("14:30")
		require.NoError(t, err)

		// Should be 2:30 PM today or tomorrow
		target := time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, now.Location())
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}

		diff := result - target.Unix()
		assert.True(t, diff >= -1 && diff <= 1,
			"Expected around %d, got %d (diff: %d)", target.Unix(), result, diff)
	})
}

func TestParseNaturalDuration_Errors_Extended(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid input",
			input: "invalid",
		},
		{
			name:  "random text",
			input: "some random text",
		},
		{
			name:  "partial match",
			input: "in 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseNaturalDuration(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestParseNaturalDuration_CaseInsensitive_Extended(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input string
	}{
		{"uppercase TOMORROW", "TOMORROW"},
		{"mixed case Tomorrow", "Tomorrow"},
		{"uppercase LATER", "LATER"},
		{"uppercase TONIGHT", "TONIGHT"},
		{"uppercase 1H", "1H"},
		{"uppercase 1D", "1D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNaturalDuration(tt.input)
			require.NoError(t, err)
			assert.True(t, result > now.Unix(), "Expected future timestamp")
		})
	}
}
