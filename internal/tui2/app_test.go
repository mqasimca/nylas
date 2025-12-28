package tui2

import (
	"testing"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewApp(t *testing.T) {
	client := nylas.NewHTTPClient()
	cfg := Config{
		Client:   client,
		GrantID:  "test-grant",
		Email:    "test@example.com",
		Provider: "google",
		Theme:    "k9s",
	}

	app := NewApp(cfg)

	if app == nil {
		t.Fatal("NewApp returned nil")
	}

	if app.global == nil {
		t.Fatal("App.global is nil")
	}

	if app.global.GrantID != "test-grant" {
		t.Errorf("Expected GrantID to be 'test-grant', got %s", app.global.GrantID)
	}

	if app.global.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", app.global.Email)
	}

	if len(app.stack) != 1 {
		t.Errorf("Expected stack length to be 1, got %d", len(app.stack))
	}
}

func TestGlobalState(t *testing.T) {
	client := nylas.NewHTTPClient()
	global := state.NewGlobalState(client, nil, "test-grant", "test@example.com", "google")

	if global.GrantID != "test-grant" {
		t.Errorf("Expected GrantID to be 'test-grant', got %s", global.GrantID)
	}

	if global.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", global.Email)
	}

	if global.Provider != "google" {
		t.Errorf("Expected Provider to be 'google', got %s", global.Provider)
	}

	// Test SetWindowSize
	global.SetWindowSize(100, 50)
	if global.WindowSize.Width != 100 || global.WindowSize.Height != 50 {
		t.Errorf("Expected WindowSize to be 100x50, got %dx%d", global.WindowSize.Width, global.WindowSize.Height)
	}

	// Test SetStatus
	global.SetStatus("test message", 1)
	if global.StatusMessage != "test message" {
		t.Errorf("Expected StatusMessage to be 'test message', got %s", global.StatusMessage)
	}
	if global.StatusLevel != 1 {
		t.Errorf("Expected StatusLevel to be 1, got %d", global.StatusLevel)
	}

	// Test ClearStatus
	global.ClearStatus()
	if global.StatusMessage != "" {
		t.Errorf("Expected StatusMessage to be empty, got %s", global.StatusMessage)
	}
	if global.StatusLevel != 0 {
		t.Errorf("Expected StatusLevel to be 0, got %d", global.StatusLevel)
	}
}

func TestScreenTypeString(t *testing.T) {
	tests := []struct {
		screen   ScreenType
		expected string
	}{
		{ScreenDashboard, "Dashboard"},
		{ScreenMessages, "Messages"},
		{ScreenMessageView, "Message"},
		{ScreenCompose, "Compose"},
		{ScreenCalendar, "Calendar"},
		{ScreenEventForm, "Event"},
		{ScreenContacts, "Contacts"},
		{ScreenSettings, "Settings"},
		{ScreenHelp, "Help"},
	}

	for _, tt := range tests {
		if got := tt.screen.String(); got != tt.expected {
			t.Errorf("ScreenType.String() = %v, want %v", got, tt.expected)
		}
	}
}
