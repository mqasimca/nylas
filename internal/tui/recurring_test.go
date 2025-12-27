package tui

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestIsRecurringEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    domain.Event
		expected bool
	}{
		{
			name:     "non-recurring event",
			event:    domain.Event{ID: "evt-1", Title: "Single Event"},
			expected: false,
		},
		{
			name: "event with recurrence",
			event: domain.Event{
				ID:         "evt-2",
				Title:      "Recurring Event",
				Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"},
			},
			expected: true,
		},
		{
			name: "event with master event ID",
			event: domain.Event{
				ID:            "evt-3",
				Title:         "Instance Event",
				MasterEventID: "master-123",
			},
			expected: true,
		},
		{
			name: "event with both",
			event: domain.Event{
				ID:            "evt-4",
				Title:         "Mixed Event",
				Recurrence:    []string{"RRULE:FREQ=DAILY"},
				MasterEventID: "master-456",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRecurringEvent(&tt.event)
			if result != tt.expected {
				t.Errorf("isRecurringEvent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatRecurrenceRule(t *testing.T) {
	tests := []struct {
		name     string
		rules    []string
		expected string
	}{
		{
			name:     "empty rules",
			rules:    []string{},
			expected: "",
		},
		{
			name:     "daily recurrence",
			rules:    []string{"RRULE:FREQ=DAILY"},
			expected: "Every day",
		},
		{
			name:     "daily recurrence every 2 days",
			rules:    []string{"FREQ=DAILY;INTERVAL=2"},
			expected: "Every 2 days",
		},
		{
			name:     "weekly recurrence",
			rules:    []string{"RRULE:FREQ=WEEKLY"},
			expected: "Every week",
		},
		{
			name:     "weekly on Monday and Wednesday",
			rules:    []string{"RRULE:FREQ=WEEKLY;BYDAY=MO,WE"},
			expected: "Every week on Mon, Wed",
		},
		{
			name:     "weekly every 2 weeks",
			rules:    []string{"FREQ=WEEKLY;INTERVAL=2"},
			expected: "Every 2 weeks",
		},
		{
			name:     "monthly recurrence",
			rules:    []string{"RRULE:FREQ=MONTHLY"},
			expected: "Every month",
		},
		{
			name:     "monthly every 3 months",
			rules:    []string{"FREQ=MONTHLY;INTERVAL=3"},
			expected: "Every 3 months",
		},
		{
			name:     "yearly recurrence",
			rules:    []string{"RRULE:FREQ=YEARLY"},
			expected: "Every year",
		},
		{
			name:     "with count",
			rules:    []string{"RRULE:FREQ=DAILY;COUNT=10"},
			expected: "Every day (10 times)",
		},
		{
			name:     "with until date",
			rules:    []string{"RRULE:FREQ=WEEKLY;UNTIL=20241231"},
			expected: "Every week until 2024-12-31",
		},
		{
			name:     "with until datetime",
			rules:    []string{"RRULE:FREQ=WEEKLY;UNTIL=20241231T235959Z"},
			expected: "Every week until 2024-12-31",
		},
		{
			name:     "skip EXDATE",
			rules:    []string{"EXDATE:20241225", "RRULE:FREQ=DAILY"},
			expected: "Every day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRecurrenceRule(tt.rules)
			if result != tt.expected {
				t.Errorf("formatRecurrenceRule(%v) = %q, want %q", tt.rules, result, tt.expected)
			}
		})
	}
}

func TestFormatDays(t *testing.T) {
	tests := []struct {
		byday    string
		expected string
	}{
		{"MO", "Mon"},
		{"TU", "Tue"},
		{"WE", "Wed"},
		{"TH", "Thu"},
		{"FR", "Fri"},
		{"SA", "Sat"},
		{"SU", "Sun"},
		{"MO,WE,FR", "Mon, Wed, Fri"},
		{"MO,TU,WE,TH,FR", "Mon, Tue, Wed, Thu, Fri"},
		{"1MO", "Mon"},  // First Monday
		{"2TU", "Tue"},  // Second Tuesday
		{"-1FR", "Fri"}, // Last Friday
	}

	for _, tt := range tests {
		t.Run(tt.byday, func(t *testing.T) {
			result := formatDays(tt.byday)
			if result != tt.expected {
				t.Errorf("formatDays(%q) = %q, want %q", tt.byday, result, tt.expected)
			}
		})
	}
}

func TestSplitRRuleParts(t *testing.T) {
	tests := []struct {
		rule     string
		expected []string
	}{
		{"FREQ=DAILY", []string{"FREQ=DAILY"}},
		{"FREQ=DAILY;INTERVAL=2", []string{"FREQ=DAILY", "INTERVAL=2"}},
		{"FREQ=WEEKLY;BYDAY=MO,WE;COUNT=10", []string{"FREQ=WEEKLY", "BYDAY=MO,WE", "COUNT=10"}},
		{"", []string{}}, // Empty string returns empty slice
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			result := splitRRuleParts(tt.rule)
			if len(result) != len(tt.expected) {
				t.Errorf("splitRRuleParts(%q) returned %d parts, want %d", tt.rule, len(result), len(tt.expected))
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("splitRRuleParts(%q)[%d] = %q, want %q", tt.rule, i, part, tt.expected[i])
				}
			}
		})
	}
}

func TestSplitByComma(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"MO", []string{"MO"}},
		{"MO,WE", []string{"MO", "WE"}},
		{"MO,TU,WE,TH,FR", []string{"MO", "TU", "WE", "TH", "FR"}},
		{"", []string{}}, // Empty string returns empty slice
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitByComma(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitByComma(%q) returned %d parts, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("splitByComma(%q)[%d] = %q, want %q", tt.input, i, part, tt.expected[i])
				}
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{}, ", ", ""},
		{[]string{"one"}, ", ", "one"},
		{[]string{"one", "two"}, ", ", "one, two"},
		{[]string{"a", "b", "c"}, "-", "a-b-c"},
	}

	for _, tt := range tests {
		result := joinStrings(tt.strs, tt.sep)
		if result != tt.expected {
			t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.strs, tt.sep, result, tt.expected)
		}
	}
}

func TestIndexByte(t *testing.T) {
	tests := []struct {
		s        string
		c        byte
		expected int
	}{
		{"hello", 'e', 1},
		{"hello", 'l', 2},
		{"hello", 'x', -1},
		{"FREQ=DAILY", '=', 4},
		{"", 'x', -1},
	}

	for _, tt := range tests {
		result := indexByte(tt.s, tt.c)
		if result != tt.expected {
			t.Errorf("indexByte(%q, %q) = %d, want %d", tt.s, tt.c, result, tt.expected)
		}
	}
}

func TestFormatRecurrenceRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		rules    []string
		expected string
	}{
		{
			name:     "nil rules",
			rules:    nil,
			expected: "",
		},
		{
			name:     "yearly with interval",
			rules:    []string{"RRULE:FREQ=YEARLY;INTERVAL=2"},
			expected: "Every 2 years",
		},
		{
			name:     "unknown frequency returns raw",
			rules:    []string{"RRULE:FREQ=UNKNOWN"},
			expected: "FREQ=UNKNOWN",
		},
		{
			name:     "only EXDATE no RRULE",
			rules:    []string{"EXDATE:20241225"},
			expected: "",
		},
		{
			name:     "multiple rules take first RRULE",
			rules:    []string{"RRULE:FREQ=DAILY", "RRULE:FREQ=WEEKLY"},
			expected: "Every day",
		},
		{
			name:     "weekly with single day",
			rules:    []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"},
			expected: "Every week on Mon",
		},
		{
			name:     "monthly with interval and count",
			rules:    []string{"RRULE:FREQ=MONTHLY;INTERVAL=2;COUNT=6"},
			expected: "Every 2 months (6 times)",
		},
		{
			name:     "daily with until only",
			rules:    []string{"RRULE:FREQ=DAILY;UNTIL=20250101"},
			expected: "Every day until 2025-01-01",
		},
		{
			name:     "daily with count only",
			rules:    []string{"RRULE:FREQ=DAILY;COUNT=30"},
			expected: "Every day (30 times)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRecurrenceRule(tt.rules)
			if result != tt.expected {
				t.Errorf("formatRecurrenceRule(%v) = %q, want %q", tt.rules, result, tt.expected)
			}
		})
	}
}

func TestFormatDays_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		byday    string
		expected string
	}{
		{"empty string", "", ""},
		{"all days", "SU,MO,TU,WE,TH,FR,SA", "Sun, Mon, Tue, Wed, Thu, Fri, Sat"},
		{"unknown day code", "XX", "XX"},
		{"mixed valid and position", "1MO,3FR", "Mon, Fri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDays(tt.byday)
			if result != tt.expected {
				t.Errorf("formatDays(%q) = %q, want %q", tt.byday, result, tt.expected)
			}
		})
	}
}

func TestIsRecurringEvent_EmptyRecurrence(t *testing.T) {
	event := domain.Event{
		ID:         "evt-1",
		Title:      "Event",
		Recurrence: []string{}, // Empty slice, not nil
	}
	result := isRecurringEvent(&event)
	if result != false {
		t.Errorf("isRecurringEvent() with empty recurrence = %v, want false", result)
	}
}
