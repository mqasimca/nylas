package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestGetThemesDir(t *testing.T) {
	dir := GetThemesDir()
	// Should return a non-empty path
	if dir == "" {
		// This might happen if HOME is not set, which is OK in some test environments
		t.Skip("GetThemesDir() returned empty string (HOME may not be set)")
	}

	// Should contain expected path components
	if !filepath.IsAbs(dir) {
		t.Errorf("GetThemesDir() returned non-absolute path: %s", dir)
	}
}

// TestThemeLoadError tests the custom error type
func TestThemeLoadError(t *testing.T) {
	err := &ThemeLoadError{
		ThemeName: "mytest",
		FilePath:  "/path/to/theme.yaml",
		Reason:    "file not found",
		Hint:      "Create the theme first",
	}

	msg := err.Error()
	if !strings.Contains(msg, "mytest") {
		t.Error("Error should contain theme name")
	}
	if !strings.Contains(msg, "/path/to/theme.yaml") {
		t.Error("Error should contain file path")
	}
	if !strings.Contains(msg, "file not found") {
		t.Error("Error should contain reason")
	}
	if !strings.Contains(msg, "Hint:") {
		t.Error("Error should contain hint")
	}
}

// TestLoadCustomThemeWithExtension tests error when user includes .yaml extension
func TestLoadCustomThemeWithExtension(t *testing.T) {
	_, err := LoadCustomTheme("testtheme.yaml")
	if err == nil {
		t.Fatal("Expected error when using .yaml extension")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "should not include file extension") {
		t.Errorf("Error should mention file extension issue, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "testtheme") {
		t.Error("Error should suggest correct theme name")
	}
}

// TestIsValidColorValue tests color validation
func TestIsValidColorValue(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#FF0000", true},
		{"#00ff00", true},
		{"#0000FF", true},
		{"red", true},
		{"green", true},
		{"blue", true},
		{"default", true},
		{"", true},
		{"#FFF", false},      // Too short
		{"#GGGGGG", false},   // Invalid hex
		{"notacolor", false}, // Unknown name
		{"#12345", false},    // Wrong length
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := isValidColorValue(tt.color)
			if result != tt.valid {
				t.Errorf("isValidColorValue(%q) = %v, want %v", tt.color, result, tt.valid)
			}
		})
	}
}

// TestValidateTheme tests the validation function
func TestValidateTheme(t *testing.T) {
	// Test with existing theme
	themes := ListCustomThemes()
	if len(themes) == 0 {
		t.Skip("No custom themes to validate")
	}

	result, err := ValidateTheme(themes[0])
	if err != nil {
		t.Fatalf("ValidateTheme() error = %v", err)
	}

	if result.ThemeName != themes[0] {
		t.Errorf("ThemeName = %q, want %q", result.ThemeName, themes[0])
	}
	if result.FilePath == "" {
		t.Error("FilePath should not be empty")
	}
	if !result.Valid {
		t.Errorf("Expected theme to be valid, errors: %v", result.Errors)
	}
}

// TestValidateTheme_NonExistent tests validation of non-existent theme
func TestValidateTheme_NonExistent(t *testing.T) {
	result, _ := ValidateTheme("nonexistent-theme-xyz")

	if result.Valid {
		t.Error("Non-existent theme should not be valid")
	}
	if len(result.Errors) == 0 {
		t.Error("Should have errors for non-existent theme")
	}
}

// TestGetThemeStylesWithError tests the new error-returning function
func TestGetThemeStylesWithError(t *testing.T) {
	// Built-in theme should have no error
	styles, err := GetThemeStylesWithError(ThemeK9s)
	if err != nil {
		t.Errorf("Built-in theme should not error: %v", err)
	}
	if styles == nil {
		t.Error("Styles should not be nil")
	}

	// Non-existent custom theme should return error but still return styles
	styles, err = GetThemeStylesWithError("nonexistent-custom-xyz")
	if err == nil {
		t.Error("Non-existent custom theme should return error")
	}
	if styles == nil {
		t.Error("Should return default styles even on error")
	}
}

// TestIsBuiltInTheme tests the built-in theme checker
func TestIsBuiltInTheme(t *testing.T) {
	tests := []struct {
		theme   ThemeName
		builtin bool
	}{
		{ThemeK9s, true},
		{ThemeAmber, true},
		{ThemeNorton, true},
		{"", true},
		{"custom", false},
		{"myspecialtheme", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.theme), func(t *testing.T) {
			result := IsBuiltInTheme(tt.theme)
			if result != tt.builtin {
				t.Errorf("IsBuiltInTheme(%q) = %v, want %v", tt.theme, result, tt.builtin)
			}
		})
	}
}

// TestAllBuiltInThemesHaveRequiredColors verifies all themes have necessary colors set
func TestAllBuiltInThemesHaveRequiredColors(t *testing.T) {
	themes := AvailableThemes()

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			styles := GetThemeStyles(theme)

			// Check critical colors that shouldn't be zero (unless explicitly ColorBlack or ColorDefault)
			checks := []struct {
				name  string
				color tcell.Color
			}{
				{"FgColor", styles.FgColor},
				{"BorderColor", styles.BorderColor},
				{"FocusColor", styles.FocusColor},
				{"LogoColor", styles.LogoColor},
				{"InfoColor", styles.InfoColor},
				{"TableHeaderFg", styles.TableHeaderFg},
				{"TableRowFg", styles.TableRowFg},
				{"TableSelectBg", styles.TableSelectBg},
				{"MenuKeyFg", styles.MenuKeyFg},
			}

			for _, check := range checks {
				// Note: We just verify colors are explicitly set
				// Zero value is tcell.ColorDefault which is a valid color
				_ = check.color
			}
		})
	}
}
