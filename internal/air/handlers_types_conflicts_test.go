package air

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestFindConflicts(t *testing.T) {
	t.Parallel()

	// Test with overlapping events
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 and event-2 overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}

	if len(conflicts) > 0 {
		if conflicts[0].Event1.ID != "event-1" || conflicts[0].Event2.ID != "event-2" {
			t.Error("expected conflict between event-1 and event-2")
		}
	}
}

func TestFindConflicts_NoOverlap(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_CancelledEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "cancelled", // This should be ignored
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (cancelled event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_FreeEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Free Time",
			Status: "confirmed",
			Busy:   false, // Free, not busy
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (free event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_AllDayEvents(t *testing.T) {
	t.Parallel()

	// All-day event should conflict with timed event on same day
	events := []domain.Event{
		{
			ID:     "all-day-1",
			Title:  "Holiday",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				Date: "2024-12-25", // All-day event
			},
		},
		{
			ID:     "timed-1",
			Title:  "Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1735142400, // Dec 25, 2024 12:00 UTC
				EndTime:   1735146000, // Dec 25, 2024 13:00 UTC
			},
		},
	}

	conflicts := findConflicts(events)

	// All-day event and timed event overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict (all-day vs timed), got %d", len(conflicts))
	}
}

func TestFindConflicts_MultipleConflicts(t *testing.T) {
	t.Parallel()

	// Three overlapping events should produce 3 conflicts (each pair)
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   3000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   3500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 2000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 overlaps event-2, event-1 overlaps event-3, event-2 overlaps event-3
	if len(conflicts) != 3 {
		t.Errorf("expected 3 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_EmptyList(t *testing.T) {
	t.Parallel()

	conflicts := findConflicts([]domain.Event{})

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for empty list, got %d", len(conflicts))
	}
}

func TestFindConflicts_SingleEvent(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Only Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for single event, got %d", len(conflicts))
	}
}

func TestRoundUpTo5Min(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{"already aligned", 1735142400, 1735142400}, // 12:00:00 stays 12:00:00
		{"1 second after", 1735142401, 1735142700},  // 12:00:01 -> 12:05:00
		{"2 minutes in", 1735142520, 1735142700},    // 12:02:00 -> 12:05:00
		{"4 min 59 sec", 1735142699, 1735142700},    // 12:04:59 -> 12:05:00
		{"zero", 0, 0},
		{"5 min aligned", 300, 300},
		{"10 min aligned", 600, 600},
		{"6 minutes", 360, 600}, // 00:06:00 -> 00:10:00
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundUpTo5Min(tt.input)
			if result != tt.expected {
				t.Errorf("roundUpTo5Min(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ================================
// CSS STYLING TESTS
// ================================
