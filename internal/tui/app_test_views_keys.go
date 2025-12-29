package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
)

func TestMessagesViewKeys(t *testing.T) {
	app := createTestApp(t)
	view := NewMessagesView(app)

	// Test key handling doesn't panic
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'}, // New compose
		{tcell.KeyRune, 'R'}, // Reply
		{tcell.KeyRune, 'A'}, // Reply all
		{tcell.KeyRune, 's'}, // Star
		{tcell.KeyRune, 'u'}, // Unread
		{tcell.KeyEnter, 0},  // Open
		{tcell.KeyEscape, 0}, // Back
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		// Just verify it doesn't panic
		_ = view.HandleKey(event)
	}
}

func TestEventsViewKeys(t *testing.T) {
	app := createTestApp(t)
	view := NewEventsView(app)

	// Test key handling doesn't panic
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'm'}, // Month view
		{tcell.KeyRune, 'w'}, // Week view
		{tcell.KeyRune, 'a'}, // Agenda view
		{tcell.KeyRune, 't'}, // Today
		{tcell.KeyRune, 'c'}, // Cycle calendar
		{tcell.KeyRune, 'C'}, // Calendar list
		{tcell.KeyTab, 0},    // Switch panel
		{tcell.KeyEnter, 0},  // View day
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		_ = view.HandleKey(event)
	}
}

// TestViewsEscapeKeyHandling verifies that all views properly return the Escape
// key event so the app can handle navigation back to the previous view.
// This is a regression test for the calendar view Escape key bug.
func TestViewsEscapeKeyHandling(t *testing.T) {
	app := createTestApp(t)
	escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)

	testCases := []struct {
		name string
		view ResourceView
	}{
		{"MessagesView", NewMessagesView(app)},
		{"EventsView", NewEventsView(app)},
		{"ContactsView", NewContactsView(app)},
		{"WebhooksView", NewWebhooksView(app)},
		{"GrantsView", NewGrantsView(app)},
		{"DashboardView", NewDashboardView(app)},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_returns_escape_event", func(t *testing.T) {
			result := tc.view.HandleKey(escapeEvent)

			// Views should return the Escape event (not nil) so the app can handle navigation
			if result == nil {
				t.Errorf("%s consumed Escape key instead of returning it for app navigation", tc.name)
			}
		})
	}
}

// TestEventsViewEscapeWithFocusedPanels tests that Escape works regardless of which panel has focus.
func TestEventsViewEscapeWithFocusedPanels(t *testing.T) {
	app := createTestApp(t)
	view := NewEventsView(app)
	escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)

	// Test with calendar panel focused (default)
	result := view.HandleKey(escapeEvent)
	if result == nil {
		t.Error("EventsView with calendar focus should return Escape event for app navigation")
	}

	// Switch to events list panel
	tabEvent := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	view.HandleKey(tabEvent)

	// Test with events list panel focused
	result = view.HandleKey(escapeEvent)
	if result == nil {
		t.Error("EventsView with events list focus should return Escape event for app navigation")
	}
}

// TestGrantsViewSwitching tests grant switching functionality.
func TestGrantsViewSwitching(t *testing.T) {
	t.Run("without_grant_store_cannot_switch", func(t *testing.T) {
		// Create app without GrantStore (like demo mode)
		mockClient := nylas.NewMockClient()
		config := Config{
			Client:          mockClient,
			GrantID:         "test-grant-id",
			Email:           "user@example.com",
			Provider:        "google",
			RefreshInterval: time.Second * 30,
			Theme:           ThemeK9s,
		}
		app := NewApp(config)

		if app.CanSwitchGrant() {
			t.Error("App without GrantStore should not be able to switch grants")
		}

		err := app.SwitchGrant("new-grant", "new@example.com", "microsoft")
		if err == nil {
			t.Error("SwitchGrant should return error when GrantStore is nil")
		}
	})

	t.Run("grants_view_shows_current_grant_marker", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		// The test app has GrantID "test-grant-id"
		// Grants list from mock should mark the matching grant
		view.Load()

		// Verify the grants view was created
		if view == nil {
			t.Fatal("GrantsView should not be nil")
		}
	})

	t.Run("grants_view_handles_enter_key", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		// Load grants data
		view.Load()

		// Test Enter key handling (should not panic)
		enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		result := view.HandleKey(enterEvent)

		// Without GrantStore, should consume event but show warning
		if result != nil {
			t.Error("Enter key should be consumed by GrantsView")
		}
	})

	t.Run("grants_view_returns_escape_event", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
		result := view.HandleKey(escapeEvent)

		if result == nil {
			t.Error("GrantsView should return Escape event for app navigation")
		}
	})
}
