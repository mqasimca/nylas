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
	global.SetWindowSize(100, 40)

	d := NewDashboard(global)
	view := d.View()

	// In v2, View() returns tea.View struct
	// Verify it has content
	if view.Content == nil {
		t.Error("View should have content")
	}
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

// Glossy Enhancement Tests

func TestDashboard_GlossyView_Renders(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")
	global.SetWindowSize(150, 50)

	d := NewDashboard(global)
	view := d.View()

	// Should render without errors
	if view.Content == nil {
		t.Error("Dashboard view should have content")
	}

	// Test that the view renders with a large window
	// The actual glossy effects (sparkles, emojis, etc.) are visual
	// and verified through manual testing and the rendering pipeline
}

func TestDashboard_GlossyVariousWindowSizes(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"large", 150, 50},
		{"medium", 100, 40},
		{"standard", 80, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global.SetWindowSize(tt.width, tt.height)
			d := NewDashboard(global)
			view := d.View()

			// Should render without errors at various sizes
			if view.Content == nil {
				t.Errorf("Dashboard should render at size %dx%d", tt.width, tt.height)
			}
		})
	}
}

func TestDashboard_SmallWindow(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")
	global.SetWindowSize(40, 20) // Very small window

	d := NewDashboard(global)
	view := d.View()

	// Should still render without panic
	if view.Content == nil {
		t.Error("Dashboard should render even with small window")
	}
}

func TestDashboard_CycleTheme(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")
	global.Theme = "k9s"

	d := NewDashboard(global)

	// Press 't' to cycle theme
	msg := tea.KeyPressMsg{Text: "t"}
	_, cmd := d.Update(msg)

	if cmd == nil {
		t.Fatal("expected theme change command, got nil")
	}

	// Execute command
	result := cmd()

	// Should return themeChangedMsg
	themeMsg, ok := result.(themeChangedMsg)
	if !ok {
		t.Fatalf("expected themeChangedMsg, got %T", result)
	}

	// Theme should have changed
	if themeMsg.theme == "k9s" {
		t.Error("theme should have changed from k9s")
	}
}

func TestDashboard_WindowSizeUpdate(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	d := NewDashboard(global)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 200, Height: 60}
	_, cmd := d.Update(msg)

	// Should not return a command for window resize
	if cmd != nil {
		t.Error("window size update should not return command")
	}

	// Global window size should be updated
	if global.WindowSize.Width != 200 || global.WindowSize.Height != 60 {
		t.Error("window size should be updated in global state")
	}

	// Components should handle the resize properly
	// We can verify by rendering the view
	view := d.View()
	if view.Content == nil {
		t.Error("Dashboard should render after window resize")
	}
}
