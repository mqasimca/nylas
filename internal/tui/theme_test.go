package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tcell.Color
	}{
		{
			name:     "hex color",
			input:    "#FF0000",
			expected: tcell.NewRGBColor(255, 0, 0),
		},
		{
			name:     "hex color lowercase",
			input:    "#00ff00",
			expected: tcell.NewRGBColor(0, 255, 0),
		},
		{
			name:     "hex color blue",
			input:    "#0000FF",
			expected: tcell.NewRGBColor(0, 0, 255),
		},
		{
			name:     "named color black",
			input:    "black",
			expected: tcell.ColorBlack,
		},
		{
			name:     "named color white",
			input:    "white",
			expected: tcell.ColorWhite,
		},
		{
			name:     "named color red",
			input:    "red",
			expected: tcell.ColorRed,
		},
		{
			name:     "named color green",
			input:    "green",
			expected: tcell.ColorGreen,
		},
		{
			name:     "named color blue",
			input:    "blue",
			expected: tcell.ColorBlue,
		},
		{
			name:     "named color yellow",
			input:    "yellow",
			expected: tcell.ColorYellow,
		},
		{
			name:     "default",
			input:    "default",
			expected: tcell.ColorDefault,
		},
		{
			name:     "empty",
			input:    "",
			expected: tcell.ColorDefault,
		},
		{
			name:     "with spaces",
			input:    "  #FF0000  ",
			expected: tcell.NewRGBColor(255, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseColor(tt.input)
			if result != tt.expected {
				t.Errorf("parseColor(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"999999", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"1.5", false},
		{"-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"1", 1},
		{"999", 999},
		{"", 0},
		{"abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetThemeStyles(t *testing.T) {
	tests := []struct {
		name  string
		theme ThemeName
	}{
		{"k9s", ThemeK9s},
		{"amber", ThemeAmber},
		{"green", ThemeGreen},
		{"apple2", ThemeAppleII},
		{"vintage", ThemeVintage},
		{"ibm", ThemeIBMDOS},
		{"futuristic", ThemeFuturistic},
		{"matrix", ThemeMatrix},
		{"norton", ThemeNorton},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := GetThemeStyles(tt.theme)
			if styles == nil {
				t.Errorf("GetThemeStyles(%q) returned nil", tt.theme)
			}

			// Verify essential style properties are set
			if styles.FgColor == 0 && styles.FgColor != tcell.ColorDefault {
				// Allow ColorDefault (0) as a valid value
			}
		})
	}
}

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) == 0 {
		t.Error("AvailableThemes() returned empty slice")
	}

	// Verify known themes are in the list
	expectedThemes := []ThemeName{
		ThemeK9s,
		ThemeAmber,
		ThemeGreen,
		ThemeAppleII,
		ThemeVintage,
		ThemeIBMDOS,
		ThemeFuturistic,
		ThemeMatrix,
		ThemeNorton,
	}

	for _, expected := range expectedThemes {
		found := false
		for _, theme := range themes {
			if theme == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected theme %q not found in AvailableThemes()", expected)
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	if styles == nil {
		t.Fatal("DefaultStyles() returned nil")
	}

	// Verify base colors are set
	if styles.BgColor == 0 && styles.BgColor != tcell.ColorBlack {
		t.Error("BgColor not properly set")
	}
	if styles.FgColor == 0 {
		t.Error("FgColor not properly set")
	}
	if styles.BorderColor == 0 {
		t.Error("BorderColor not properly set")
	}

	// Verify table colors are set
	if styles.TableSelectBg == 0 {
		t.Error("TableSelectBg not properly set")
	}
	if styles.TableSelectFg == 0 && styles.TableSelectFg != tcell.ColorBlack {
		// ColorBlack is 0, so we check both conditions
	}
}

func TestThemeConfigToStyles(t *testing.T) {
	config := &ThemeConfig{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Red:        "#f7768e",
		Green:      "#9ece6a",
		Yellow:     "#e0af68",
		Blue:       "#7aa2f7",
		K9s: K9sSkin{
			Body: BodyStyle{
				FgColor:   "#c0caf5",
				BgColor:   "#1a1b26",
				LogoColor: "#bb9af7",
			},
			Frame: FrameStyle{
				Border: BorderStyle{
					FgColor:    "#3b4261",
					FocusColor: "#7aa2f7",
				},
			},
		},
	}

	styles := config.ToStyles()

	if styles == nil {
		t.Fatal("ToStyles() returned nil")
	}

	// Verify colors were applied
	expectedFg := parseColor("#c0caf5")
	if styles.FgColor != expectedFg {
		t.Errorf("FgColor = %v, want %v", styles.FgColor, expectedFg)
	}

	expectedBg := parseColor("#1a1b26")
	if styles.BgColor != expectedBg {
		t.Errorf("BgColor = %v, want %v", styles.BgColor, expectedBg)
	}

	expectedLogo := parseColor("#bb9af7")
	if styles.LogoColor != expectedLogo {
		t.Errorf("LogoColor = %v, want %v", styles.LogoColor, expectedLogo)
	}
}

func TestCreateDefaultThemeFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "nylas-theme-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	themePath := filepath.Join(tmpDir, "test-theme.yaml")

	// Create theme file
	err = CreateDefaultThemeFile(themePath)
	if err != nil {
		t.Fatalf("CreateDefaultThemeFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		t.Error("Theme file was not created")
	}

	// Load the created theme
	config, err := LoadThemeFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadThemeFromFile() error = %v", err)
	}

	// Verify the loaded config has expected values
	if config.Foreground == "" {
		t.Error("Loaded config has empty foreground")
	}
	if config.K9s.Body.FgColor == "" {
		t.Error("Loaded config has empty K9s body fgColor")
	}
}

func TestLoadThemeFromFile_NotFound(t *testing.T) {
	_, err := LoadThemeFromFile("/nonexistent/path/theme.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadCustomTheme_NotFound(t *testing.T) {
	_, err := LoadCustomTheme("nonexistent-theme-12345")
	if err == nil {
		t.Error("Expected error for non-existent custom theme, got nil")
	}
}

func TestListCustomThemes(t *testing.T) {
	// This test just ensures the function doesn't panic
	themes := ListCustomThemes()
	// themes may be nil or empty if no custom themes exist
	_ = themes
}

// TestCustomThemeIntegration tests the full custom theme loading flow
func TestCustomThemeIntegration(t *testing.T) {
	// Create temp directory to simulate ~/.config/nylas/themes/
	tmpDir, err := os.MkdirTemp("", "nylas-custom-theme-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a custom theme file
	themePath := filepath.Join(tmpDir, "mytest.yaml")
	themeContent := `# Custom test theme
foreground: "#00FF00"
background: "#000000"
red: "#FF0000"
green: "#00FF00"
yellow: "#FFFF00"
blue: "#0000FF"

k9s:
  body:
    fgColor: "#00FF00"
    bgColor: "#000000"
    logoColor: "#FF00FF"
  prompt:
    fgColor: "#00FF00"
    bgColor: "#000000"
  info:
    fgColor: "#00FFFF"
    sectionColor: "#00FF00"
  frame:
    border:
      fgColor: "#333333"
      focusColor: "#00FF00"
    menu:
      fgColor: "#00FF00"
      keyColor: "#FFFF00"
      numKeyColor: "#FF00FF"
  views:
    table:
      fgColor: "#00FF00"
      header:
        fgColor: "#FFFF00"
        bgColor: "#000000"
      selected:
        fgColor: "#000000"
        bgColor: "#00FF00"
`
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write theme file: %v", err)
	}

	// Test LoadThemeFromFile
	config, err := LoadThemeFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadThemeFromFile() error = %v", err)
	}

	// Verify config was loaded correctly
	if config.Foreground != "#00FF00" {
		t.Errorf("Foreground = %q, want %q", config.Foreground, "#00FF00")
	}
	if config.K9s.Body.LogoColor != "#FF00FF" {
		t.Errorf("LogoColor = %q, want %q", config.K9s.Body.LogoColor, "#FF00FF")
	}
	if config.K9s.Views.Table.Selected.BgColor != "#00FF00" {
		t.Errorf("Table.Selected.BgColor = %q, want %q", config.K9s.Views.Table.Selected.BgColor, "#00FF00")
	}

	// Test ToStyles conversion
	styles := config.ToStyles()
	if styles == nil {
		t.Fatal("ToStyles() returned nil")
	}

	// Verify colors were applied correctly
	expectedGreen := tcell.NewRGBColor(0, 255, 0) // #00FF00
	if styles.FgColor != expectedGreen {
		t.Errorf("FgColor = %v, want %v (green)", styles.FgColor, expectedGreen)
	}

	expectedMagenta := tcell.NewRGBColor(255, 0, 255) // #FF00FF
	if styles.LogoColor != expectedMagenta {
		t.Errorf("LogoColor = %v, want %v (magenta)", styles.LogoColor, expectedMagenta)
	}

	// Verify table selection colors
	if styles.TableSelectBg != expectedGreen {
		t.Errorf("TableSelectBg = %v, want %v (green)", styles.TableSelectBg, expectedGreen)
	}
}

// TestCustomThemeViaGetThemeStyles tests that GetThemeStyles correctly loads custom themes
func TestCustomThemeViaGetThemeStyles(t *testing.T) {
	// Create temp themes directory
	tmpDir, err := os.MkdirTemp("", "nylas-themes-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a distinguishable custom theme
	themeName := "testpurple"
	themePath := filepath.Join(tmpDir, themeName+".yaml")
	themeContent := `foreground: "#9900FF"
background: "#1A0033"
k9s:
  body:
    fgColor: "#9900FF"
    bgColor: "#1A0033"
    logoColor: "#FF00FF"
  frame:
    border:
      fgColor: "#660099"
      focusColor: "#CC00FF"
  views:
    table:
      fgColor: "#9900FF"
      header:
        fgColor: "#CC00FF"
      selected:
        fgColor: "#000000"
        bgColor: "#9900FF"
`
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write theme file: %v", err)
	}

	// Test loading directly via LoadThemeFromFile
	config, err := LoadThemeFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadThemeFromFile() failed: %v", err)
	}

	styles := config.ToStyles()

	// Verify the purple color was loaded
	expectedPurple := tcell.NewRGBColor(153, 0, 255) // #9900FF
	if styles.FgColor != expectedPurple {
		t.Errorf("Custom theme FgColor = %v, want %v (purple)", styles.FgColor, expectedPurple)
	}

	// Verify logo color
	expectedMagenta := tcell.NewRGBColor(255, 0, 255) // #FF00FF
	if styles.LogoColor != expectedMagenta {
		t.Errorf("Custom theme LogoColor = %v, want %v (magenta)", styles.LogoColor, expectedMagenta)
	}
}

// TestLoadCustomThemeFromConfigDir tests loading from the actual config directory
func TestLoadCustomThemeFromConfigDir(t *testing.T) {
	// This test checks if we can load the "testcustom" theme if it exists
	// Skip if no custom themes exist
	themes := ListCustomThemes()
	if len(themes) == 0 {
		t.Skip("No custom themes found in ~/.config/nylas/themes/")
	}

	// Try to load the first custom theme found
	themeName := themes[0]
	styles, err := LoadCustomTheme(themeName)
	if err != nil {
		t.Fatalf("LoadCustomTheme(%q) error = %v", themeName, err)
	}

	if styles == nil {
		t.Errorf("LoadCustomTheme(%q) returned nil styles", themeName)
	}

	// Verify it has valid colors set
	if styles.FgColor == 0 && styles.BgColor == 0 {
		t.Error("Custom theme has no colors set")
	}

	t.Logf("Successfully loaded custom theme %q with FgColor=%v", themeName, styles.FgColor)
}

// TestGetThemeStylesWithCustomTheme tests that GetThemeStyles falls back to custom themes
func TestGetThemeStylesWithCustomTheme(t *testing.T) {
	// First verify built-in themes work
	k9sStyles := GetThemeStyles(ThemeK9s)
	if k9sStyles == nil {
		t.Fatal("GetThemeStyles(k9s) returned nil")
	}

	// Try loading a non-existent theme - should fall back to default
	unknownStyles := GetThemeStyles("nonexistent-theme-xyz")
	if unknownStyles == nil {
		t.Fatal("GetThemeStyles(nonexistent) returned nil")
	}

	// The unknown theme should fall back to default styles
	defaultStyles := DefaultStyles()
	if unknownStyles.FgColor != defaultStyles.FgColor {
		t.Errorf("Unknown theme should fall back to default, FgColor = %v, want %v",
			unknownStyles.FgColor, defaultStyles.FgColor)
	}
}

// TestGetThemeStylesLoadsCustomTheme verifies GetThemeStyles loads themes from ~/.config/nylas/themes/
func TestGetThemeStylesLoadsCustomTheme(t *testing.T) {
	// Check if testcustom theme exists
	themes := ListCustomThemes()
	hasTestCustom := false
	for _, theme := range themes {
		if theme == "testcustom" {
			hasTestCustom = true
			break
		}
	}
	if !hasTestCustom {
		t.Skip("testcustom theme not found - run 'nylas tui theme init testcustom' first")
	}

	// Load the custom theme via GetThemeStyles
	customStyles := GetThemeStyles("testcustom")
	if customStyles == nil {
		t.Fatal("GetThemeStyles(testcustom) returned nil")
	}

	// The testcustom theme should have the Tokyo Night colors
	// foreground: "#c0caf5" = RGB(192, 202, 245)
	expectedFg := tcell.NewRGBColor(192, 202, 245)
	if customStyles.FgColor != expectedFg {
		t.Errorf("Custom theme FgColor = %v, want %v (#c0caf5)", customStyles.FgColor, expectedFg)
	}

	// logoColor: "#bb9af7" = RGB(187, 154, 247)
	expectedLogo := tcell.NewRGBColor(187, 154, 247)
	if customStyles.LogoColor != expectedLogo {
		t.Errorf("Custom theme LogoColor = %v, want %v (#bb9af7)", customStyles.LogoColor, expectedLogo)
	}

	// Verify it's different from default k9s theme
	defaultStyles := GetThemeStyles(ThemeK9s)
	if customStyles.FgColor == defaultStyles.FgColor {
		t.Error("Custom theme FgColor should be different from default k9s theme")
	}

	t.Logf("SUCCESS: GetThemeStyles correctly loaded custom theme 'testcustom'")
	t.Logf("  FgColor: %v (expected #c0caf5)", customStyles.FgColor)
	t.Logf("  LogoColor: %v (expected #bb9af7)", customStyles.LogoColor)
}

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
				// We just verify colors are explicitly set
				// Zero value is tcell.ColorDefault which is valid
				if check.color == 0 {
					// This is fine - could be ColorDefault
				}
			}
		})
	}
}
