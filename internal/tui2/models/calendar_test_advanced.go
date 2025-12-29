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
