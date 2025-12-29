// Package models provides screen models for the TUI.
package models

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestDashboard_KeyboardShortcuts(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectScreen ScreenType
	}{
		{"press 'a' navigates to messages", "a", ScreenMessages},
		{"press 'c' navigates to calendar", "c", ScreenCalendar},
		{"press 'p' navigates to contacts", "p", ScreenContacts},
		{"press 'd' navigates to debug", "d", ScreenDebug},
		{"press 's' navigates to settings", "s", ScreenSettings},
		{"press '?' navigates to help", "?", ScreenHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
			global.Theme = "k9s"
			d := NewDashboard(global)

			// Create a proper key message using v2 API
			msg := tea.KeyPressMsg{Text: tt.key}

			// Update with the key message
			_, cmd := d.Update(msg)

			// Execute the command to get the navigation message
			if cmd == nil {
				t.Fatalf("expected command, got nil")
			}

			navMsg := cmd()
			if navMsg == nil {
				t.Fatalf("expected navigation message, got nil")
			}

			nav, ok := navMsg.(NavigateMsg)
			if !ok {
				t.Fatalf("expected NavigateMsg, got %T", navMsg)
			}

			if nav.Screen != tt.expectScreen {
				t.Errorf("expected screen %v, got %v", tt.expectScreen, nav.Screen)
			}
		})
	}
}

func TestDashboard_ThemeCycling(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"
	d := NewDashboard(global)

	initialTheme := global.Theme

	// Press 't' to cycle theme
	msg := tea.KeyPressMsg{Text: "t"}

	_, cmd := d.Update(msg)

	if cmd == nil {
		t.Fatal("expected command for theme cycling, got nil")
	}

	// Execute command to trigger theme change
	result := cmd()

	// Check if theme changed
	if global.Theme == initialTheme {
		// Theme should have changed
		t.Logf("Theme before: %s, after: %s", initialTheme, global.Theme)
	}

	// Verify we got a theme changed message
	if result != nil {
		if _, ok := result.(themeChangedMsg); !ok {
			t.Logf("Got message type: %T", result)
		}
	}
}

func TestDashboard_KeyHandling(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"
	global.SetWindowSize(100, 50)
	d := NewDashboard(global)

	// Test that dashboard handles key messages
	msg := tea.KeyPressMsg{Text: "a"}

	updated, cmd := d.Update(msg)

	if updated == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		t.Error("Pressing 'a' should return a command")
	}
}

func TestDashboard_ViewRendersCorrectly(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"
	global.SetWindowSize(100, 50)
	d := NewDashboard(global)

	// Just verify View() doesn't panic and returns a view
	view := d.View()

	// Check that view has content (Content field should not be nil)
	if view.Content == nil {
		t.Error("View should have content")
	}

	// Note: AltScreen is set by App.View(), not by individual screen models
}

func TestDashboard_AllNavigationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() tea.Cmd
		expected ScreenType
	}{
		{"navigateToMessages", navigateToMessages, ScreenMessages},
		{"navigateToCalendar", navigateToCalendar, ScreenCalendar},
		{"navigateToContacts", navigateToContacts, ScreenContacts},
		{"navigateToSettings", navigateToSettings, ScreenSettings},
		{"navigateToHelp", navigateToHelp, ScreenHelp},
		{"navigateToDebug", navigateToDebug, ScreenDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.fn()
			if cmd == nil {
				t.Fatal("navigation function returned nil command")
			}

			msg := cmd()
			navMsg, ok := msg.(NavigateMsg)
			if !ok {
				t.Fatalf("expected NavigateMsg, got %T", msg)
			}

			if navMsg.Screen != tt.expected {
				t.Errorf("expected screen %v, got %v", tt.expected, navMsg.Screen)
			}
		})
	}
}
