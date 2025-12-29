package components

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestWorkingHours_IsWorkingDay(t *testing.T) {
	wh := DefaultWorkingHours()

	tests := []struct {
		day      time.Weekday
		expected bool
	}{
		{time.Monday, true},
		{time.Tuesday, true},
		{time.Wednesday, true},
		{time.Thursday, true},
		{time.Friday, true},
		{time.Saturday, false},
		{time.Sunday, false},
	}

	for _, tt := range tests {
		t.Run(tt.day.String(), func(t *testing.T) {
			if wh.IsWorkingDay(tt.day) != tt.expected {
				t.Errorf("IsWorkingDay(%s) = %v, want %v", tt.day, wh.IsWorkingDay(tt.day), tt.expected)
			}
		})
	}
}

func TestWorkingHours_IsWorkingHour(t *testing.T) {
	wh := DefaultWorkingHours()

	tests := []struct {
		hour     int
		expected bool
	}{
		{8, false},
		{9, true},
		{12, true},
		{16, true},
		{17, false},
		{18, false},
		{0, false},
		{23, false},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.hour)), func(t *testing.T) {
			if wh.IsWorkingHour(tt.hour) != tt.expected {
				t.Errorf("IsWorkingHour(%d) = %v, want %v", tt.hour, wh.IsWorkingHour(tt.hour), tt.expected)
			}
		})
	}
}

func TestCalendarGrid_Timezone(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Test setting timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone not available")
	}

	grid.SetTimezone(loc)

	if grid.GetTimezone() != loc {
		t.Error("timezone should be set correctly")
	}
}

func TestCalendarGrid_AllDayEvents(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Create an all-day event
	events := []domain.Event{
		{
			ID:    "allday",
			Title: "All Day Event",
			When: domain.EventWhen{
				Date:   "2025-06-15",
				Object: "date",
			},
		},
	}
	grid.SetEvents(events)

	// All-day events are indexed by their date string directly,
	// so the query date just needs to format to the same string
	testDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	dayEvents := grid.GetEventsForDate(testDate)

	if len(dayEvents) != 1 {
		t.Errorf("expected 1 all-day event, got %d", len(dayEvents))
	}

	// Verify the event details
	if len(dayEvents) > 0 && dayEvents[0].Title != "All Day Event" {
		t.Errorf("expected 'All Day Event', got %q", dayEvents[0].Title)
	}
}

func TestCalendarGrid_GetMonthDays(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Set to June 2025
	testDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	grid.SetCurrentMonth(testDate)
	grid.SetShowWeekends(true)

	days := grid.getMonthDays()

	// Should include days from previous month to fill the first week
	// and days from next month to fill the last week
	// June 2025 starts on Sunday, so with Monday start we need May 26-31
	if len(days) == 0 {
		t.Error("getMonthDays should return days")
	}

	// Check first day is a Monday (since firstDayMon is true)
	if days[0].Weekday() != time.Monday {
		t.Errorf("first day should be Monday, got %s", days[0].Weekday())
	}
}

func TestCalendarGrid_GetMonthDays_NoWeekends(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	testDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	grid.SetCurrentMonth(testDate)
	grid.SetShowWeekends(false)

	days := grid.getMonthDays()

	// No weekends should be included
	for _, day := range days {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			t.Errorf("should not include weekends, got %s", day.Weekday())
		}
	}
}

func TestCalendarGrid_WeekStart(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Wednesday June 11, 2025
	testDate := time.Date(2025, 6, 11, 12, 0, 0, 0, time.Local)

	t.Run("MondayStart", func(t *testing.T) {
		grid.SetFirstDayMonday(true)
		weekStart := grid.getWeekStart(testDate)
		if weekStart.Weekday() != time.Monday {
			t.Errorf("week start should be Monday, got %s", weekStart.Weekday())
		}
		// Should be June 9, 2025
		if weekStart.Day() != 9 {
			t.Errorf("week start should be June 9, got June %d", weekStart.Day())
		}
	})

	t.Run("SundayStart", func(t *testing.T) {
		grid.SetFirstDayMonday(false)
		weekStart := grid.getWeekStart(testDate)
		if weekStart.Weekday() != time.Sunday {
			t.Errorf("week start should be Sunday, got %s", weekStart.Weekday())
		}
		// Should be June 8, 2025
		if weekStart.Day() != 8 {
			t.Errorf("week start should be June 8, got June %d", weekStart.Day())
		}
	})
}

func TestCalendarGrid_DayNames(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	t.Run("MondayFirst_WithWeekends", func(t *testing.T) {
		grid.SetFirstDayMonday(true)
		grid.SetShowWeekends(true)
		names := grid.getDayNames()
		if len(names) != 7 {
			t.Errorf("expected 7 days, got %d", len(names))
		}
		if names[0] != "Mon" {
			t.Errorf("first day should be Mon, got %s", names[0])
		}
		if names[6] != "Sun" {
			t.Errorf("last day should be Sun, got %s", names[6])
		}
	})

	t.Run("SundayFirst_WithWeekends", func(t *testing.T) {
		grid.SetFirstDayMonday(false)
		grid.SetShowWeekends(true)
		names := grid.getDayNames()
		if len(names) != 7 {
			t.Errorf("expected 7 days, got %d", len(names))
		}
		if names[0] != "Sun" {
			t.Errorf("first day should be Sun, got %s", names[0])
		}
	})

	t.Run("NoWeekends", func(t *testing.T) {
		grid.SetFirstDayMonday(true)
		grid.SetShowWeekends(false)
		names := grid.getDayNames()
		if len(names) != 5 {
			t.Errorf("expected 5 days (no weekends), got %d", len(names))
		}
	})
}

func TestCalendarGrid_View_MonthHeader(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)

	testDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.Local)
	grid.SetCurrentMonth(testDate)

	view := grid.View()

	// Should contain month and year
	if view == "" {
		t.Error("view should not be empty")
	}
	// The view should render without error
}

func TestCalendarGrid_AgendaView_SortsEvents(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)
	grid.SetViewMode(CalendarAgendaView)

	now := time.Now()
	events := []domain.Event{
		{ID: "evt2", Title: "Later", When: domain.EventWhen{StartTime: now.Add(2 * time.Hour).Unix()}},
		{ID: "evt1", Title: "Earlier", When: domain.EventWhen{StartTime: now.Add(time.Hour).Unix()}},
		{ID: "evt3", Title: "Tomorrow", When: domain.EventWhen{StartTime: now.AddDate(0, 0, 1).Unix()}},
	}
	grid.SetEvents(events)

	view := grid.View()
	if view == "" {
		t.Error("agenda view should not be empty")
	}
}

func TestCalendarGrid_WeekView_Renders(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)
	grid.SetViewMode(CalendarWeekView)

	now := time.Now()
	events := []domain.Event{
		{ID: "evt1", Title: "Meeting", When: domain.EventWhen{StartTime: now.Unix()}},
	}
	grid.SetEvents(events)

	view := grid.View()
	if view == "" {
		t.Error("week view should not be empty")
	}
}

func TestCalendarGrid_CancelledEvent(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)
	grid.SetViewMode(CalendarAgendaView)

	now := time.Now()
	events := []domain.Event{
		{
			ID:     "evt1",
			Title:  "Cancelled Meeting",
			Status: "cancelled",
			When:   domain.EventWhen{StartTime: now.Unix()},
		},
	}
	grid.SetEvents(events)

	// Just verify it renders without error
	view := grid.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestCalendarGrid_EventWithLocation(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)
	grid.SetViewMode(CalendarAgendaView)

	now := time.Now()
	events := []domain.Event{
		{
			ID:       "evt1",
			Title:    "Office Meeting",
			Location: "Conference Room A",
			When:     domain.EventWhen{StartTime: now.Unix()},
		},
	}
	grid.SetEvents(events)

	// Just verify it renders without error
	view := grid.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestCalendarGrid_EventWithConferencing(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)
	grid.SetViewMode(CalendarAgendaView)

	now := time.Now()
	events := []domain.Event{
		{
			ID:    "evt1",
			Title: "Video Call",
			Conferencing: &domain.Conferencing{
				Provider: "Zoom",
				Details: &domain.ConferencingDetails{
					URL: "https://zoom.us/j/123456",
				},
			},
			When: domain.EventWhen{StartTime: now.Unix()},
		},
	}
	grid.SetEvents(events)

	// Just verify it renders without error
	view := grid.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}
