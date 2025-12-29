package air

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// Natural Language Duration Parser
// =============================================================================

// parseNaturalDuration parses natural language duration into Unix timestamp.
func parseNaturalDuration(input string) (int64, error) {
	now := time.Now()
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle relative durations: "1h", "2d", "30m", "1w"
	if matched, _ := regexp.MatchString(`^\d+[hdwm]$`, input); matched {
		num, _ := strconv.Atoi(input[:len(input)-1])
		unit := input[len(input)-1]
		switch unit {
		case 'h':
			return now.Add(time.Duration(num) * time.Hour).Unix(), nil
		case 'd':
			return now.Add(time.Duration(num) * 24 * time.Hour).Unix(), nil
		case 'w':
			return now.Add(time.Duration(num) * 7 * 24 * time.Hour).Unix(), nil
		case 'm':
			return now.Add(time.Duration(num) * time.Minute).Unix(), nil
		}
	}

	// Handle "later today" (4 hours or 5 PM, whichever is first)
	if input == "later" || input == "later today" {
		later := now.Add(4 * time.Hour)
		fivePM := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())
		if fivePM.After(now) && fivePM.Before(later) {
			return fivePM.Unix(), nil
		}
		return later.Unix(), nil
	}

	// Handle "tonight" (8 PM today)
	if input == "tonight" {
		tonight := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		if tonight.Before(now) {
			tonight = tonight.Add(24 * time.Hour)
		}
		return tonight.Unix(), nil
	}

	// Handle "tomorrow" (9 AM tomorrow)
	if strings.HasPrefix(input, "tomorrow") {
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())

		// Check for time specification: "tomorrow 2pm", "tomorrow at 3:30"
		parts := strings.Fields(input)
		if len(parts) > 1 {
			timeStr := parts[len(parts)-1]
			if strings.HasPrefix(parts[1], "at") && len(parts) > 2 {
				timeStr = parts[2]
			}
			if hour, min, ok := parseTimeString(timeStr); ok {
				tomorrow = time.Date(now.Year(), now.Month(), now.Day()+1, hour, min, 0, 0, now.Location())
			}
		}
		return tomorrow.Unix(), nil
	}

	// Handle "next week" (Monday 9 AM)
	if input == "next week" || input == "monday" {
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextMonday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 9, 0, 0, 0, now.Location())
		return nextMonday.Unix(), nil
	}

	// Handle "this weekend" (Saturday 10 AM)
	if input == "weekend" || input == "this weekend" || input == "saturday" {
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		saturday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSaturday, 10, 0, 0, 0, now.Location())
		return saturday.Unix(), nil
	}

	// Handle specific times: "9am", "14:30", "3:30pm"
	if hour, min, ok := parseTimeString(input); ok {
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}
		return target.Unix(), nil
	}

	return 0, &parseError{input: input}
}

// parseTimeString parses time strings like "9am", "14:30", "3:30pm".
func parseTimeString(s string) (hour, min int, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))

	isPM := strings.HasSuffix(s, "pm")
	isAM := strings.HasSuffix(s, "am")
	s = strings.TrimSuffix(strings.TrimSuffix(s, "pm"), "am")

	parts := strings.Split(s, ":")
	if len(parts) == 1 {
		// Just hour: "9", "14"
		h, err := strconv.Atoi(parts[0])
		if err != nil || h < 0 || h > 23 {
			return 0, 0, false
		}
		hour = h
		min = 0
	} else if len(parts) == 2 {
		// Hour:min: "9:30", "14:00"
		h, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
			return 0, 0, false
		}
		hour = h
		min = m
	} else {
		return 0, 0, false
	}

	// Handle AM/PM
	if isPM && hour < 12 {
		hour += 12
	} else if isAM && hour == 12 {
		hour = 0
	}

	return hour, min, true
}
