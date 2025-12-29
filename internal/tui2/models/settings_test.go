// Package models provides screen models for the TUI.
package models

import (
	"testing"

	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewSettingsScreen(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	s := NewSettingsScreen(global)

	if s == nil {
		t.Fatal("NewSettingsScreen returned nil")
	}

	if s.global != global {
		t.Error("global state not set correctly")
	}

	if s.theme == nil {
		t.Error("theme not initialized")
	}

	// Check default values loaded from config
	if s.selectedTheme == "" {
		t.Error("selectedTheme should be set")
	}

	if s.cursor != 0 {
		t.Error("cursor should start at 0")
	}

	if s.saved {
		t.Error("should not be saved initially")
	}
}

func TestSettingsCycleTheme(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	s := NewSettingsScreen(global)
	initialTheme := s.selectedTheme

	// Cycle forward
	s.cycleTheme(1)
	if s.selectedTheme == initialTheme {
		t.Error("theme should change when cycling forward")
	}

	// Cycle backward
	s.cycleTheme(-1)
	if s.selectedTheme != initialTheme {
		t.Error("theme should return to initial when cycling backward")
	}

	// Cycle forward to end and wrap around
	for i := 0; i < 10; i++ {
		s.cycleTheme(1)
	}
	// Should have wrapped around, verify theme is valid
	if s.theme == nil {
		t.Error("theme should still be valid after wrapping")
	}
}

func TestSettingsToggleSetting(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	s := NewSettingsScreen(global)

	tests := []struct {
		name         string
		cursorPos    int
		checkField   func() bool
		initialValue bool
	}{
		{
			name:      "toggle animations",
			cursorPos: 1,
			checkField: func() bool {
				return s.enabledAnimations
			},
		},
		{
			name:      "toggle status bar",
			cursorPos: 3,
			checkField: func() bool {
				return s.showStatusBar
			},
		},
		{
			name:      "toggle footer",
			cursorPos: 4,
			checkField: func() bool {
				return s.showFooter
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.cursor = tt.cursorPos
			initialValue := tt.checkField()
			s.toggleSetting()
			newValue := tt.checkField()

			if newValue == initialValue {
				t.Errorf("setting should toggle from %v to %v", initialValue, !initialValue)
			}
		})
	}
}

func TestSettingsSplashDuration(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	s := NewSettingsScreen(global)
	s.cursor = 2 // Splash duration setting
	s.splashDuration = 3

	// Toggle should increment
	s.toggleSetting()
	if s.splashDuration != 4 {
		t.Errorf("splash duration should increment to 4, got %d", s.splashDuration)
	}

	// At max, should wrap to 1
	s.splashDuration = 10
	s.toggleSetting()
	if s.splashDuration != 1 {
		t.Errorf("splash duration should wrap to 1 when at max, got %d", s.splashDuration)
	}
}
