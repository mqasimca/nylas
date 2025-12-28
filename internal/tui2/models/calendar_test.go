package models

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewCalendarScreen(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	if screen == nil {
		t.Fatal("NewCalendarScreen returned nil")
	}

	if screen.global != global {
		t.Error("global state not set correctly")
	}

	if screen.theme == nil {
		t.Error("theme should be initialized")
	}

	if screen.calendarGrid == nil {
		t.Error("calendar grid should be initialized")
	}

	if !screen.loading {
		t.Error("loading should be true initially")
	}
}

func TestCalendarScreen_Init(t *testing.T) {
	mockClient := &nylas.MockClient{
		GetCalendarsFunc: func(_ context.Context, _ string) ([]domain.Calendar, error) {
			return []domain.Calendar{
				{ID: "cal1", Name: "Primary", IsPrimary: true},
				{ID: "cal2", Name: "Work", IsPrimary: false},
			}, nil
		},
	}

	global := state.NewGlobalState(
		mockClient,
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	cmd := screen.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestCalendarScreen_Update_KeyNavigation(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = false // Disable loading for navigation tests

	tests := []struct {
		name           string
		key            tea.KeyPressMsg
		expectViewMode components.CalendarViewMode
	}{
		{
			name:           "m key switches to month view",
			key:            tea.KeyPressMsg{Text: "m"},
			expectViewMode: components.CalendarMonthView,
		},
		{
			name:           "w key switches to week view",
			key:            tea.KeyPressMsg{Text: "w"},
			expectViewMode: components.CalendarWeekView,
		},
		{
			name:           "g key switches to agenda view",
			key:            tea.KeyPressMsg{Text: "g"},
			expectViewMode: components.CalendarAgendaView,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen.calendarGrid.SetViewMode(components.CalendarMonthView) // Reset
			_, _ = screen.Update(tt.key)

			if screen.calendarGrid.GetViewMode() != tt.expectViewMode {
				t.Errorf("expected view mode %v, got %v", tt.expectViewMode, screen.calendarGrid.GetViewMode())
			}
		})
	}
}

func TestCalendarScreen_Update_EscapeGoesBack(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Create escape key message
	escKey := tea.KeyPressMsg{Code: tea.KeyEsc}

	_, cmd := screen.Update(escKey)

	if cmd == nil {
		t.Error("Escape should return a command")
	}

	// Execute command and check for BackMsg
	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Errorf("Escape should return BackMsg, got %T", msg)
	}
}

func TestCalendarScreen_Update_CtrlCQuits(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Create ctrl+c key message
	ctrlC := tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'c'}

	_, cmd := screen.Update(ctrlC)

	if cmd == nil {
		t.Error("Ctrl+C should return a command")
	}
}

func TestCalendarScreen_Update_WindowSizeMsg(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Send window size message
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	_, _ = screen.Update(sizeMsg)

	if screen.width != 120 {
		t.Errorf("expected width 120, got %d", screen.width)
	}

	if screen.height != 40 {
		t.Errorf("expected height 40, got %d", screen.height)
	}
}

func TestCalendarScreen_Update_CalendarsLoadedMsg(t *testing.T) {
	mockClient := &nylas.MockClient{
		GetEventsFunc: func(_ context.Context, _ string, _ string, _ *domain.EventQueryParams) ([]domain.Event, error) {
			return []domain.Event{}, nil
		},
	}

	global := state.NewGlobalState(
		mockClient,
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Send calendars loaded message
	msg := calendarsLoadedMsg{
		calendars: []domain.Calendar{
			{ID: "cal1", Name: "Primary", IsPrimary: true},
			{ID: "cal2", Name: "Work", IsPrimary: false},
		},
	}

	_, cmd := screen.Update(msg)

	if screen.loading {
		t.Error("loading should be false after calendars loaded")
	}

	if !screen.calendarsLoaded {
		t.Error("calendarsLoaded should be true")
	}

	if len(screen.calendars) != 2 {
		t.Errorf("expected 2 calendars, got %d", len(screen.calendars))
	}

	if screen.selectedCalendar == nil {
		t.Error("primary calendar should be auto-selected")
	} else if screen.selectedCalendar.ID != "cal1" {
		t.Error("primary calendar should be selected")
	}

	// Should trigger event fetch
	if cmd == nil {
		t.Error("should return command to fetch events")
	}
}

func TestCalendarScreen_Update_EventsLoadedMsg(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loadingEvents = true

	now := time.Now()
	msg := eventsLoadedMsg{
		events: []domain.Event{
			{ID: "evt1", Title: "Meeting 1", When: domain.EventWhen{StartTime: now.Unix()}},
			{ID: "evt2", Title: "Meeting 2", When: domain.EventWhen{StartTime: now.Add(time.Hour).Unix()}},
		},
	}

	_, _ = screen.Update(msg)

	if screen.loadingEvents {
		t.Error("loadingEvents should be false after events loaded")
	}

	if len(screen.events) != 2 {
		t.Errorf("expected 2 events, got %d", len(screen.events))
	}
}

func TestCalendarScreen_Update_ErrorMsg(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = true

	testErr := errMsg{err: errors.New("test error")}
	_, _ = screen.Update(testErr)

	if screen.err == nil {
		t.Error("err should be set after error message")
	}

	if screen.loading {
		t.Error("loading should be false after error")
	}
}

func TestCalendarScreen_Update_TodayKey(t *testing.T) {
	mockClient := &nylas.MockClient{
		GetEventsFunc: func(_ context.Context, _ string, _ string, _ *domain.EventQueryParams) ([]domain.Event, error) {
			return []domain.Event{}, nil
		},
	}

	global := state.NewGlobalState(
		mockClient,
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = false
	screen.selectedCalendar = &domain.Calendar{ID: "cal1", Name: "Primary"}

	// Set to a different date
	futureDate := time.Now().AddDate(0, 3, 0)
	screen.calendarGrid.SetSelectedDate(futureDate)

	// Press 't' to go to today
	tKey := tea.KeyPressMsg{Text: "t"}
	_, _ = screen.Update(tKey)

	now := time.Now()
	selected := screen.calendarGrid.GetSelectedDate()
	if selected.Year() != now.Year() || selected.Month() != now.Month() || selected.Day() != now.Day() {
		t.Error("'t' key should navigate to today")
	}
}

func TestCalendarScreen_View(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = false
	screen.selectedCalendar = &domain.Calendar{ID: "cal1", Name: "Primary"}

	view := screen.View()
	// In v2, View is a struct - just verify it can be created
	_ = view
}

func TestCalendarScreen_View_WithError(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.err = errors.New("calendar not found")

	view := screen.View()
	// In v2, View is a struct - just verify it can be created
	_ = view
}

func TestCalendarScreen_View_Loading(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = true

	view := screen.View()
	// In v2, View is a struct - just verify it can be created
	_ = view
}

func TestCalendarScreen_ViewModes(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = false

	modes := []components.CalendarViewMode{
		components.CalendarMonthView,
		components.CalendarWeekView,
		components.CalendarAgendaView,
	}

	for _, mode := range modes {
		t.Run(viewModeName(mode), func(t *testing.T) {
			screen.calendarGrid.SetViewMode(mode)
			view := screen.View()
			// In v2, View is a struct - just verify it can be created
			_ = view
		})
	}
}

func TestCalendarScreen_RenderEventSummary(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	now := time.Now()

	tests := []struct {
		name  string
		event domain.Event
	}{
		{
			name: "basic event",
			event: domain.Event{
				ID:    "evt1",
				Title: "Team Meeting",
				When:  domain.EventWhen{StartTime: now.Unix(), EndTime: now.Add(time.Hour).Unix()},
			},
		},
		{
			name: "all-day event",
			event: domain.Event{
				ID:    "evt2",
				Title: "Holiday",
				When:  domain.EventWhen{Date: now.Format("2006-01-02"), Object: "date"},
			},
		},
		{
			name: "cancelled event",
			event: domain.Event{
				ID:     "evt3",
				Title:  "Cancelled Meeting",
				Status: "cancelled",
				When:   domain.EventWhen{StartTime: now.Unix()},
			},
		},
		{
			name: "event with location",
			event: domain.Event{
				ID:       "evt4",
				Title:    "Office Meeting",
				Location: "Conference Room A",
				When:     domain.EventWhen{StartTime: now.Unix()},
			},
		},
		{
			name: "event with conferencing",
			event: domain.Event{
				ID:    "evt5",
				Title: "Video Call",
				Conferencing: &domain.Conferencing{
					Provider: "Zoom",
					Details:  &domain.ConferencingDetails{URL: "https://zoom.us/j/123"},
				},
				When: domain.EventWhen{StartTime: now.Unix()},
			},
		},
		{
			name: "event with participants",
			event: domain.Event{
				ID:    "evt6",
				Title: "Team Sync",
				Participants: []domain.Participant{
					{Email: "user1@example.com"},
					{Email: "user2@example.com"},
					{Email: "user3@example.com"},
				},
				When: domain.EventWhen{StartTime: now.Unix()},
			},
		},
		{
			name: "free event",
			event: domain.Event{
				ID:    "evt7",
				Title: "Focus Time",
				Busy:  false,
				When:  domain.EventWhen{StartTime: now.Unix()},
			},
		},
		{
			name: "event with no title",
			event: domain.Event{
				ID:   "evt8",
				When: domain.EventWhen{StartTime: now.Unix()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := screen.renderEventSummary(tt.event)
			if result == "" {
				t.Error("renderEventSummary should not return empty string")
			}
		})
	}
}

func TestCalendarItem(t *testing.T) {
	item := calendarItem{
		calendar: domain.Calendar{
			ID:          "cal1",
			Name:        "Primary",
			Description: "My primary calendar",
			IsPrimary:   true,
		},
	}

	title := item.Title()
	if title == "" {
		t.Error("Title should not be empty")
	}
	if title != "Primary ★" {
		t.Errorf("expected 'Primary ★', got %q", title)
	}

	desc := item.Description()
	if desc != "My primary calendar" {
		t.Errorf("expected 'My primary calendar', got %q", desc)
	}

	filter := item.FilterValue()
	if filter != "Primary" {
		t.Errorf("expected 'Primary', got %q", filter)
	}
}

func TestCalendarItem_NonPrimary(t *testing.T) {
	item := calendarItem{
		calendar: domain.Calendar{
			ID:        "cal2",
			Name:      "Work",
			IsPrimary: false,
		},
	}

	title := item.Title()
	if title != "Work" {
		t.Errorf("expected 'Work' without star, got %q", title)
	}
}

func TestCalendarScreen_AutoSelectPrimaryCalendar(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Calendars with primary not first
	msg := calendarsLoadedMsg{
		calendars: []domain.Calendar{
			{ID: "cal2", Name: "Work", IsPrimary: false},
			{ID: "cal1", Name: "Primary", IsPrimary: true},
			{ID: "cal3", Name: "Personal", IsPrimary: false},
		},
	}

	_, _ = screen.Update(msg)

	if screen.selectedCalendar == nil {
		t.Fatal("a calendar should be selected")
	}

	if screen.selectedCalendar.ID != "cal1" {
		t.Errorf("primary calendar should be selected, got %s", screen.selectedCalendar.ID)
	}
}

func TestCalendarScreen_FallbackToFirstCalendar(t *testing.T) {
	global := state.NewGlobalState(
		&nylas.MockClient{},
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)

	// Calendars with no primary
	msg := calendarsLoadedMsg{
		calendars: []domain.Calendar{
			{ID: "cal1", Name: "Work", IsPrimary: false},
			{ID: "cal2", Name: "Personal", IsPrimary: false},
		},
	}

	_, _ = screen.Update(msg)

	if screen.selectedCalendar == nil {
		t.Fatal("a calendar should be selected")
	}

	if screen.selectedCalendar.ID != "cal1" {
		t.Errorf("first calendar should be selected when no primary, got %s", screen.selectedCalendar.ID)
	}
}

func TestCalendarScreen_Update_RefreshKey(t *testing.T) {
	mockClient := &nylas.MockClient{
		GetCalendarsFunc: func(_ context.Context, _ string) ([]domain.Calendar, error) {
			return []domain.Calendar{
				{ID: "cal1", Name: "Primary", IsPrimary: true},
			}, nil
		},
	}

	global := state.NewGlobalState(
		mockClient,
		nil,
		"test-grant-id",
		"test@example.com",
		"google",
	)

	screen := NewCalendarScreen(global)
	screen.loading = false

	// Press Ctrl+R to refresh
	ctrlR := tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'r'}
	_, cmd := screen.Update(ctrlR)

	if !screen.loading {
		t.Error("loading should be true after refresh")
	}

	if cmd == nil {
		t.Error("refresh should return a command")
	}
}

// Helper function for view mode names
func viewModeName(mode components.CalendarViewMode) string {
	switch mode {
	case components.CalendarMonthView:
		return "MonthView"
	case components.CalendarWeekView:
		return "WeekView"
	case components.CalendarAgendaView:
		return "AgendaView"
	default:
		return "Unknown"
	}
}
