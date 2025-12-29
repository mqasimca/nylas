// Package models provides screen models for the TUI.
package models

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestHelp_KeyboardShortcuts(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		expectQuit bool
		expectBack bool
	}{
		{"esc navigates back", "esc", false, true},
		{"q navigates back", "q", false, true},
		{"ctrl+c quits", "ctrl+c", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
			global.Theme = "k9s"
			global.SetWindowSize(100, 50)
			h := NewHelpScreen(global)

			msg := tea.KeyPressMsg{Text: tt.key}
			_, cmd := h.Update(msg)

			if tt.expectQuit {
				// Should return quit command
				if cmd == nil {
					t.Fatal("expected quit command, got nil")
				}
			}

			if tt.expectBack {
				// Should return back message
				if cmd == nil {
					t.Fatal("expected back command, got nil")
				}
				result := cmd()
				if _, ok := result.(BackMsg); !ok {
					t.Errorf("expected BackMsg, got %T", result)
				}
			}
		})
	}
}

func TestHelp_ViewRendersCorrectly(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"
	global.SetWindowSize(100, 50)
	h := NewHelpScreen(global)

	view := h.View()

	if view.Content == nil {
		t.Error("View should have content")
	}
}
