// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantName  string
	}{
		{"default k9s", "k9s", "k9s"},
		{"empty defaults to k9s", "", "k9s"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"nord", "nord", "nord"},
		{"dracula", "dracula", "dracula"},
		{"catppuccin", "catppuccin", "catppuccin"},
		{"gruvbox", "gruvbox", "gruvbox"},
		{"tokyo_night", "tokyo_night", "tokyo_night"},
		{"invalid defaults to k9s", "invalid", "k9s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.themeName)
			if theme == nil {
				t.Fatal("GetTheme returned nil")
			}
			if theme.Name != tt.wantName {
				t.Errorf("GetTheme(%q) = %q, want %q", tt.themeName, theme.Name, tt.wantName)
			}
		})
	}
}

func TestListAvailableThemes(t *testing.T) {
	themes := ListAvailableThemes()

	if len(themes) != 7 {
		t.Errorf("ListAvailableThemes() returned %d themes, want 7", len(themes))
	}

	// Check all expected themes are present
	expectedThemes := map[string]bool{
		"k9s":         true,
		"cyberpunk":   true,
		"nord":        true,
		"dracula":     true,
		"catppuccin":  true,
		"gruvbox":     true,
		"tokyo_night": true,
	}

	for _, theme := range themes {
		if !expectedThemes[theme] {
			t.Errorf("Unexpected theme in list: %s", theme)
		}
		delete(expectedThemes, theme)
	}

	if len(expectedThemes) > 0 {
		t.Errorf("Missing themes: %v", expectedThemes)
	}
}

func TestThemeHasAllStyles(t *testing.T) {
	themes := []string{"k9s", "cyberpunk", "nord", "dracula", "catppuccin", "gruvbox", "tokyo_night"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			theme := GetTheme(themeName)

			// Check all required colors are set
			if theme.Foreground == nil {
				t.Error("Foreground color is nil")
			}
			if theme.Background == nil {
				t.Error("Background color is nil")
			}
			if theme.Primary == nil {
				t.Error("Primary color is nil")
			}
			if theme.Secondary == nil {
				t.Error("Secondary color is nil")
			}
			if theme.Accent == nil {
				t.Error("Accent color is nil")
			}
			if theme.Success == nil {
				t.Error("Success color is nil")
			}
			if theme.Warning == nil {
				t.Error("Warning color is nil")
			}
			if theme.Error == nil {
				t.Error("Error color is nil")
			}
			if theme.Info == nil {
				t.Error("Info color is nil")
			}

			// Check component styles exist
			if theme.Title.GetForeground() == nil {
				t.Error("Title style foreground is nil")
			}
			if theme.Subtitle.GetForeground() == nil {
				t.Error("Subtitle style foreground is nil")
			}
			// Border style exists (it's a struct, not a pointer)
			_ = theme.Border.GetBorderStyle()
		})
	}
}

func TestGradientBorder(t *testing.T) {
	theme := GetTheme("k9s")
	style := GradientBorder(theme.Primary, theme.Accent)

	// Border style exists (it's a struct, not a pointer)
	_ = style.GetBorderStyle()
}
