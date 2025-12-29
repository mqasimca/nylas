// Package models provides screen models for the TUI.
package models

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewSplash(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
	}

	splash := NewSplash(global)

	if splash == nil {
		t.Fatal("NewSplash returned nil")
	}

	if splash.global != global {
		t.Error("global state not set")
	}

	if splash.percent != 0.0 {
		t.Errorf("initial percent = %f, want 0.0", splash.percent)
	}
}

func TestSplash_ProgressOverTime(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
	}

	splash := NewSplash(global)

	// Simulate time passing
	splash.startTime = time.Now().Add(-1500 * time.Millisecond) // 1.5 seconds ago

	// Process a tick message
	msg := tickMsg(time.Now())
	updatedModel, _ := splash.Update(msg)
	updatedSplash, ok := updatedModel.(*SplashScreen)
	if !ok {
		t.Fatal("Update returned wrong type")
	}

	// Progress should be approximately 50% (1.5s / 3s)
	if updatedSplash.percent < 0.4 || updatedSplash.percent > 0.6 {
		t.Errorf("percent = %f, expected around 0.5", updatedSplash.percent)
	}
}

func TestSplash_TransitionsToDashboard(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
		WindowSize: struct{ Width, Height int }{
			Width:  100,
			Height: 40,
		},
	}

	splash := NewSplash(global)

	// Simulate completion (3+ seconds passed)
	splash.startTime = time.Now().Add(-3100 * time.Millisecond)

	msg := tickMsg(time.Now())
	_, cmd := splash.Update(msg)

	if cmd == nil {
		t.Fatal("expected navigation command, got nil")
	}

	// Execute the command to get the message
	result := cmd()
	if navMsg, ok := result.(NavigateMsg); ok {
		if navMsg.Screen != ScreenDashboard {
			t.Errorf("expected navigation to Dashboard, got %v", navMsg.Screen)
		}
	} else {
		t.Error("expected NavigateMsg")
	}
}

func TestSplash_HandlesWindowSizeMsg(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
	}

	splash := NewSplash(global)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	_, _ = splash.Update(msg)

	if global.WindowSize.Width != 120 {
		t.Errorf("width = %d, want 120", global.WindowSize.Width)
	}

	if global.WindowSize.Height != 40 {
		t.Errorf("height = %d, want 40", global.WindowSize.Height)
	}
}

func TestSplash_SkipWithKeyPress(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
		WindowSize: struct{ Width, Height int }{
			Width:  100,
			Height: 40,
		},
	}

	splash := NewSplash(global)

	// Press any key (e.g., Enter)
	msg := tea.KeyPressMsg{Text: "enter"}
	_, cmd := splash.Update(msg)

	if cmd == nil {
		t.Fatal("expected navigation command when key pressed, got nil")
	}

	// Execute the command to get the message
	result := cmd()
	if navMsg, ok := result.(NavigateMsg); ok {
		if navMsg.Screen != ScreenDashboard {
			t.Errorf("expected navigation to Dashboard, got %v", navMsg.Screen)
		}
	} else {
		t.Error("expected NavigateMsg when key pressed")
	}
}

// Glossy Enhancement Tests

func TestSplash_GlossyView_Renders(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
		WindowSize: struct{ Width, Height int }{
			Width:  100,
			Height: 40,
		},
	}

	splash := NewSplash(global)
	view := splash.View()

	// Should render without errors
	if view.Content == nil {
		t.Error("Splash view should have content")
	}

	// The actual glossy effects (sparkles, metallic text, progress glow, etc.)
	// are visual and verified through manual testing and the rendering pipeline
}

func TestSplash_GlossyVariousWindowSizes(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"large", 150, 50},
		{"medium", 100, 40},
		{"standard", 80, 24},
		{"small", 60, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global := &state.GlobalState{
				Theme: "k9s",
				Email: "test@example.com",
				WindowSize: struct{ Width, Height int }{
					Width:  tt.width,
					Height: tt.height,
				},
			}

			splash := NewSplash(global)
			view := splash.View()

			// Should render without errors at various sizes
			if view.Content == nil {
				t.Errorf("Splash should render at size %dx%d", tt.width, tt.height)
			}
		})
	}
}

func TestSplash_GlossyProgressAnimation(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
		WindowSize: struct{ Width, Height int }{
			Width:  100,
			Height: 40,
		},
	}

	splash := NewSplash(global)

	// Test at different progress points
	progressSteps := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, progress := range progressSteps {
		splash.percent = progress

		view := splash.View()

		// Should render at all progress levels
		if view.Content == nil {
			t.Errorf("Splash should render at progress %.2f", progress)
		}
	}
}

func TestSplash_GlossyMultipleThemes(t *testing.T) {
	themes := []string{"k9s", "catppuccin-mocha", "dracula", "nord", "tokyo-night"}

	for _, theme := range themes {
		t.Run(theme, func(t *testing.T) {
			global := &state.GlobalState{
				Theme: theme,
				Email: "test@example.com",
				WindowSize: struct{ Width, Height int }{
					Width:  100,
					Height: 40,
				},
			}

			splash := NewSplash(global)
			view := splash.View()

			// Should render with all themes
			if view.Content == nil {
				t.Errorf("Splash should render with %s theme", theme)
			}
		})
	}
}

func TestSplash_Init(t *testing.T) {
	global := &state.GlobalState{
		Theme: "k9s",
		Email: "test@example.com",
	}

	splash := NewSplash(global)
	cmd := splash.Init()

	// Should return a batch command (tick + progress init)
	if cmd == nil {
		t.Error("Splash.Init() should return a command")
	}
}
