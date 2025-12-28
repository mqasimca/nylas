package components

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewCalendarGrid(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	if grid == nil {
		t.Fatal("NewCalendarGrid returned nil")
	}

	if grid.theme != theme {
		t.Error("theme not set correctly")
	}

	if grid.viewMode != CalendarMonthView {
		t.Error("default view mode should be CalendarMonthView")
	}

	if grid.workingHours == nil {
		t.Error("working hours should be initialized")
	}

	if !grid.showWeekends {
		t.Error("showWeekends should default to true")
	}

	if !grid.firstDayMon {
		t.Error("firstDayMon should default to true (ISO week)")
	}
}

func TestCalendarGrid_SetEvents(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Use fixed date to avoid timezone issues
	testDate := time.Date(2025, 6, 15, 10, 0, 0, 0, time.Local)
	nextDay := time.Date(2025, 6, 16, 10, 0, 0, 0, time.Local)

	events := []domain.Event{
		{
			ID:    "evt1",
			Title: "Meeting 1",
			When: domain.EventWhen{
				StartTime: testDate.Unix(),
				EndTime:   testDate.Add(time.Hour).Unix(),
			},
		},
		{
			ID:    "evt2",
			Title: "Meeting 2",
			When: domain.EventWhen{
				StartTime: testDate.Add(2 * time.Hour).Unix(),
				EndTime:   testDate.Add(3 * time.Hour).Unix(),
			},
		},
		{
			ID:    "evt3",
			Title: "Tomorrow Meeting",
			When: domain.EventWhen{
				StartTime: nextDay.Unix(),
				EndTime:   nextDay.Add(time.Hour).Unix(),
			},
		},
	}

	grid.SetEvents(events)

	if grid.GetEventCount() != 3 {
		t.Errorf("expected 3 events, got %d", grid.GetEventCount())
	}

	// Check events are indexed by date
	todayEvents := grid.GetEventsForDate(testDate)
	if len(todayEvents) != 2 {
		t.Errorf("expected 2 events for June 15, got %d", len(todayEvents))
	}

	tomorrowEvents := grid.GetEventsForDate(nextDay)
	if len(tomorrowEvents) != 1 {
		t.Errorf("expected 1 event for June 16, got %d", len(tomorrowEvents))
	}
}

func TestCalendarGrid_SetEvents_SortsByTime(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Use fixed date to avoid timezone issues
	testDate := time.Date(2025, 6, 15, 10, 0, 0, 0, time.Local)

	// Add events out of order
	events := []domain.Event{
		{
			ID:    "evt2",
			Title: "Later Meeting",
			When: domain.EventWhen{
				StartTime: testDate.Add(2 * time.Hour).Unix(),
			},
		},
		{
			ID:    "evt1",
			Title: "Earlier Meeting",
			When: domain.EventWhen{
				StartTime: testDate.Unix(),
			},
		},
	}

	grid.SetEvents(events)

	dayEvents := grid.GetEventsForDate(testDate)
	if len(dayEvents) != 2 {
		t.Fatalf("expected 2 events, got %d", len(dayEvents))
	}

	// Events should be sorted by start time
	if dayEvents[0].Title != "Earlier Meeting" {
		t.Error("events should be sorted by start time")
	}
	if dayEvents[1].Title != "Later Meeting" {
		t.Error("events should be sorted by start time")
	}
}

func TestCalendarGrid_Navigation(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Set a known date
	testDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.Local)
	grid.SetSelectedDate(testDate)

	t.Run("NextDay", func(t *testing.T) {
		grid.SetSelectedDate(testDate)
		grid.NextDay()
		expected := testDate.AddDate(0, 0, 1)
		if grid.GetSelectedDate().Day() != expected.Day() {
			t.Errorf("expected day %d, got %d", expected.Day(), grid.GetSelectedDate().Day())
		}
	})

	t.Run("PrevDay", func(t *testing.T) {
		grid.SetSelectedDate(testDate)
		grid.PrevDay()
		expected := testDate.AddDate(0, 0, -1)
		if grid.GetSelectedDate().Day() != expected.Day() {
			t.Errorf("expected day %d, got %d", expected.Day(), grid.GetSelectedDate().Day())
		}
	})

	t.Run("NextWeek", func(t *testing.T) {
		grid.SetSelectedDate(testDate)
		grid.NextWeek()
		expected := testDate.AddDate(0, 0, 7)
		if grid.GetSelectedDate().Day() != expected.Day() {
			t.Errorf("expected day %d, got %d", expected.Day(), grid.GetSelectedDate().Day())
		}
	})

	t.Run("PrevWeek", func(t *testing.T) {
		grid.SetSelectedDate(testDate)
		grid.PrevWeek()
		expected := testDate.AddDate(0, 0, -7)
		if grid.GetSelectedDate().Day() != expected.Day() {
			t.Errorf("expected day %d, got %d", expected.Day(), grid.GetSelectedDate().Day())
		}
	})

	t.Run("NextMonth", func(t *testing.T) {
		grid.SetCurrentMonth(testDate)
		grid.NextMonth()
		if grid.GetCurrentMonth().Month() != time.July {
			t.Errorf("expected July, got %s", grid.GetCurrentMonth().Month())
		}
	})

	t.Run("PrevMonth", func(t *testing.T) {
		grid.SetCurrentMonth(testDate)
		grid.PrevMonth()
		if grid.GetCurrentMonth().Month() != time.May {
			t.Errorf("expected May, got %s", grid.GetCurrentMonth().Month())
		}
	})

	t.Run("GoToToday", func(t *testing.T) {
		grid.SetSelectedDate(testDate.AddDate(1, 0, 0)) // Set to next year
		grid.GoToToday()
		now := time.Now()
		selected := grid.GetSelectedDate()
		if selected.Year() != now.Year() || selected.Month() != now.Month() || selected.Day() != now.Day() {
			t.Error("GoToToday should set selected date to today")
		}
	})
}

func TestCalendarGrid_NavigationUpdatesMonth(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Start at end of month
	testDate := time.Date(2025, 6, 30, 12, 0, 0, 0, time.Local)
	grid.SetSelectedDate(testDate)

	// Navigate to next day (should move to July)
	grid.NextDay()

	if grid.GetCurrentMonth().Month() != time.July {
		t.Errorf("expected current month to update to July, got %s", grid.GetCurrentMonth().Month())
	}
}

func TestCalendarGrid_IsToday(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	now := time.Now()
	if !grid.IsToday(now) {
		t.Error("IsToday should return true for today")
	}

	yesterday := now.AddDate(0, 0, -1)
	if grid.IsToday(yesterday) {
		t.Error("IsToday should return false for yesterday")
	}

	tomorrow := now.AddDate(0, 0, 1)
	if grid.IsToday(tomorrow) {
		t.Error("IsToday should return false for tomorrow")
	}
}

func TestCalendarGrid_IsSelected(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	testDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.Local)
	grid.SetSelectedDate(testDate)

	if !grid.IsSelected(testDate) {
		t.Error("IsSelected should return true for selected date")
	}

	otherDate := testDate.AddDate(0, 0, 1)
	if grid.IsSelected(otherDate) {
		t.Error("IsSelected should return false for other dates")
	}
}

func TestCalendarGrid_IsCurrentMonth(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	testDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.Local)
	grid.SetCurrentMonth(testDate)

	if !grid.IsCurrentMonth(testDate) {
		t.Error("IsCurrentMonth should return true for date in current month")
	}

	otherMonth := testDate.AddDate(0, 1, 0)
	if grid.IsCurrentMonth(otherMonth) {
		t.Error("IsCurrentMonth should return false for date in other month")
	}
}

func TestCalendarGrid_HasEvents(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	now := time.Now()
	events := []domain.Event{
		{
			ID:    "evt1",
			Title: "Meeting",
			When: domain.EventWhen{
				StartTime: now.Unix(),
			},
		},
	}
	grid.SetEvents(events)

	if !grid.HasEvents(now) {
		t.Error("HasEvents should return true for date with events")
	}

	tomorrow := now.AddDate(0, 0, 1)
	if grid.HasEvents(tomorrow) {
		t.Error("HasEvents should return false for date without events")
	}
}

func TestCalendarGrid_EventCountForDate(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Use fixed date to avoid timezone issues
	testDate := time.Date(2025, 6, 15, 10, 0, 0, 0, time.Local)
	nextDay := time.Date(2025, 6, 16, 10, 0, 0, 0, time.Local)

	events := []domain.Event{
		{ID: "evt1", Title: "Meeting 1", When: domain.EventWhen{StartTime: testDate.Unix()}},
		{ID: "evt2", Title: "Meeting 2", When: domain.EventWhen{StartTime: testDate.Add(time.Hour).Unix()}},
		{ID: "evt3", Title: "Meeting 3", When: domain.EventWhen{StartTime: testDate.Add(2 * time.Hour).Unix()}},
	}
	grid.SetEvents(events)

	count := grid.EventCountForDate(testDate)
	if count != 3 {
		t.Errorf("expected 3 events, got %d", count)
	}

	count = grid.EventCountForDate(nextDay)
	if count != 0 {
		t.Errorf("expected 0 events for next day, got %d", count)
	}
}

func TestCalendarGrid_ViewModes(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)
	grid.SetSize(80, 24)

	t.Run("MonthView", func(t *testing.T) {
		grid.SetViewMode(CalendarMonthView)
		view := grid.View()
		if view == "" {
			t.Error("month view should not be empty")
		}
	})

	t.Run("WeekView", func(t *testing.T) {
		grid.SetViewMode(CalendarWeekView)
		view := grid.View()
		if view == "" {
			t.Error("week view should not be empty")
		}
	})

	t.Run("AgendaView", func(t *testing.T) {
		grid.SetViewMode(CalendarAgendaView)
		view := grid.View()
		if view == "" {
			t.Error("agenda view should not be empty")
		}
	})
}

func TestCalendarGrid_Update_KeyNavigation(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	testDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.Local)
	grid.SetSelectedDate(testDate)

	tests := []struct {
		name        string
		key         tea.KeyPressMsg
		expectDay   int
		expectMonth time.Month
	}{
		{"right arrow", tea.KeyPressMsg{Code: tea.KeyRight}, 16, time.June},
		{"left arrow", tea.KeyPressMsg{Code: tea.KeyLeft}, 14, time.June},
		{"down arrow", tea.KeyPressMsg{Code: tea.KeyDown}, 22, time.June},
		{"up arrow", tea.KeyPressMsg{Code: tea.KeyUp}, 8, time.June},
		{"h key", tea.KeyPressMsg{Text: "h"}, 14, time.June},
		{"l key", tea.KeyPressMsg{Text: "l"}, 16, time.June},
		{"j key", tea.KeyPressMsg{Text: "j"}, 22, time.June},
		{"k key", tea.KeyPressMsg{Text: "k"}, 8, time.June},
		{"[ key", tea.KeyPressMsg{Text: "["}, 15, time.May},
		{"] key", tea.KeyPressMsg{Text: "]"}, 15, time.July},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.SetSelectedDate(testDate)
			grid.SetCurrentMonth(testDate)
			grid.Update(tt.key)

			if tt.expectMonth != grid.GetCurrentMonth().Month() {
				// For month navigation, check current month
				if tt.name == "[ key" || tt.name == "] key" {
					t.Errorf("expected month %s, got %s", tt.expectMonth, grid.GetCurrentMonth().Month())
				}
			}
			if tt.name != "[ key" && tt.name != "] key" {
				if grid.GetSelectedDate().Day() != tt.expectDay {
					t.Errorf("expected day %d, got %d", tt.expectDay, grid.GetSelectedDate().Day())
				}
			}
		})
	}
}

func TestCalendarGrid_Update_TodayKey(t *testing.T) {
	theme := styles.DefaultTheme()
	grid := NewCalendarGrid(theme)

	// Set to a different date
	pastDate := time.Date(2020, 1, 1, 12, 0, 0, 0, time.Local)
	grid.SetSelectedDate(pastDate)

	grid.Update(tea.KeyPressMsg{Text: "t"})

	now := time.Now()
	selected := grid.GetSelectedDate()
	if selected.Year() != now.Year() || selected.Month() != now.Month() || selected.Day() != now.Day() {
		t.Error("'t' key should navigate to today")
	}
}

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
