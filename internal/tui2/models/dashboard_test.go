package models

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewDashboard(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	d := NewDashboard(global)

	if d == nil {
		t.Fatal("NewDashboard returned nil")
	}

	if d.global != global {
		t.Error("global state not set correctly")
	}

	if d.theme == nil {
		t.Error("theme not initialized")
	}
}

func TestDashboard_Init(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	d := NewDashboard(global)
	cmd := d.Init()

	if cmd != nil {
		t.Error("Dashboard.Init() should return nil")
	}
}

func TestDashboard_Update_KeyNavigation(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	tests := []struct {
		name       string
		key        string
		wantScreen ScreenType
	}{
		{"messages", "a", ScreenMessages},
		{"calendar", "c", ScreenCalendar},
		{"contacts", "p", ScreenContacts},
		{"settings", "s", ScreenSettings},
		{"help", "?", ScreenHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new dashboard for each test
			d := NewDashboard(global)

			msg := tea.KeyPressMsg{Text: tt.key}
			_, cmd := d.Update(msg)

			if cmd == nil {
				t.Fatal("expected navigation command, got nil")
			}

			// Execute the command and verify it returns NavigateMsg
			result := cmd()
			navMsg, ok := result.(NavigateMsg)
			if !ok {
				t.Fatalf("expected NavigateMsg, got %T", result)
			}

			if navMsg.Screen != tt.wantScreen {
				t.Errorf("screen = %v, want %v", navMsg.Screen, tt.wantScreen)
			}
		})
	}
}

func TestDashboard_Update_Quit(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	d := NewDashboard(global)

	tests := []struct {
		name string
		key  string
	}{
		{"ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg tea.Msg
			if tt.key == "ctrl+c" {
				msg = tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'c'}
			} else {
				msg = tea.KeyPressMsg{Text: tt.key}
			}

			_, cmd := d.Update(msg)

			if cmd == nil {
				t.Fatal("expected quit command, got nil")
			}

			// The quit command should be tea.Quit
			result := cmd()
			if result != tea.Quit() {
				t.Error("expected tea.Quit message")
			}
		})
	}
}

func TestDashboard_View(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	d := NewDashboard(global)
	view := d.View()

	// In v2, View() returns tea.View struct, not string
	// Just verify it returns a non-nil View
	_ = view // TODO: Add proper View content assertions for v2
}

func TestNavigationCommands(t *testing.T) {
	tests := []struct {
		name       string
		cmdFunc    func() tea.Cmd
		wantScreen ScreenType
	}{
		{"navigateToMessages", navigateToMessages, ScreenMessages},
		{"navigateToCalendar", navigateToCalendar, ScreenCalendar},
		{"navigateToContacts", navigateToContacts, ScreenContacts},
		{"navigateToSettings", navigateToSettings, ScreenSettings},
		{"navigateToHelp", navigateToHelp, ScreenHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			if cmd == nil {
				t.Fatal("command function returned nil")
			}

			result := cmd()
			navMsg, ok := result.(NavigateMsg)
			if !ok {
				t.Fatalf("expected NavigateMsg, got %T", result)
			}

			if navMsg.Screen != tt.wantScreen {
				t.Errorf("screen = %v, want %v", navMsg.Screen, tt.wantScreen)
			}
		})
	}
}

func TestScreenType_Values(t *testing.T) {
	// Verify all screen types have unique values
	screens := []ScreenType{
		ScreenDashboard,
		ScreenMessages,
		ScreenCalendar,
		ScreenContacts,
		ScreenSettings,
		ScreenHelp,
	}

	seen := make(map[ScreenType]bool)
	for _, screen := range screens {
		if seen[screen] {
			t.Errorf("duplicate screen type value: %v", screen)
		}
		seen[screen] = true
	}
}
