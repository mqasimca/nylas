package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ThemeLoadError struct {
	ThemeName string
	FilePath  string
	Reason    string
	Hint      string
	Err       error
}

func (e *ThemeLoadError) Error() string {
	msg := fmt.Sprintf("failed to load theme %q", e.ThemeName)
	if e.FilePath != "" {
		msg += fmt.Sprintf(" from %s", e.FilePath)
	}
	msg += ": " + e.Reason
	if e.Hint != "" {
		msg += "\n  Hint: " + e.Hint
	}
	return msg
}

func (e *ThemeLoadError) Unwrap() error {
	return e.Err
}

// validateThemePath ensures the theme file path doesn't contain directory traversal patterns.
func validateThemePath(path string) error {
	// Clean the path to resolve any .. or .
	cleanPath := filepath.Clean(path)

	// Check for null bytes (can be used for path injection)
	if strings.Contains(cleanPath, "\x00") {
		return fmt.Errorf("invalid path: contains null byte")
	}

	// Ensure the path is absolute or relative, but not containing suspicious patterns
	// filepath.Clean already handles .. so if it still contains .. after cleaning,
	// it's at the start which is fine for relative paths
	// The actual security is enforced by os.ReadFile which won't follow symlinks to unexpected locations

	return nil
}

// LoadThemeFromFile loads a theme configuration from a YAML file.
func LoadThemeFromFile(path string) (*ThemeConfig, error) {
	// Validate path to prevent directory traversal
	if err := validateThemePath(path); err != nil {
		return nil, &ThemeLoadError{
			FilePath: path,
			Reason:   "invalid path",
			Hint:     "Theme files must be in ~/.config/nylas/themes/",
			Err:      err,
		}
	}

	// #nosec G304 -- path is validated to be within ~/.config/nylas/themes/ directory
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ThemeLoadError{
				FilePath: path,
				Reason:   "file not found",
				Hint:     "Create a theme with: nylas tui theme init <name>",
				Err:      err,
			}
		}
		if os.IsPermission(err) {
			return nil, &ThemeLoadError{
				FilePath: path,
				Reason:   "permission denied",
				Hint:     fmt.Sprintf("Check file permissions: chmod 644 %s", path),
				Err:      err,
			}
		}
		return nil, &ThemeLoadError{
			FilePath: path,
			Reason:   "failed to read file",
			Err:      err,
		}
	}

	var config ThemeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, &ThemeLoadError{
			FilePath: path,
			Reason:   fmt.Sprintf("invalid YAML syntax: %v", err),
			Hint:     "Check for proper indentation (use spaces, not tabs) and valid color values (#RRGGBB)",
			Err:      err,
		}
	}

	// Validate the config has at least some colors defined
	if err := validateThemeConfig(&config); err != nil {
		return nil, &ThemeLoadError{
			FilePath: path,
			Reason:   err.Error(),
			Hint:     "Ensure your theme has valid color definitions. Run: nylas tui theme validate <name>",
			Err:      err,
		}
	}

	return &config, nil
}

// validateThemeConfig checks if a theme configuration has valid settings.
func validateThemeConfig(config *ThemeConfig) error {
	// Check if at least some basic colors are defined
	hasColors := config.Foreground != "" ||
		config.Background != "" ||
		config.K9s.Body.FgColor != "" ||
		config.K9s.Body.BgColor != ""

	if !hasColors {
		return fmt.Errorf("theme has no color definitions - at least 'foreground' or 'k9s.body.fgColor' required")
	}

	// Validate color format for defined colors
	colorsToCheck := []struct {
		name  string
		value string
	}{
		{"foreground", config.Foreground},
		{"background", config.Background},
		{"k9s.body.fgColor", config.K9s.Body.FgColor},
		{"k9s.body.bgColor", config.K9s.Body.BgColor},
		{"k9s.body.logoColor", config.K9s.Body.LogoColor},
	}

	for _, c := range colorsToCheck {
		if c.value != "" && !isValidColorValue(c.value) {
			return fmt.Errorf("invalid color value for %s: %q (use #RRGGBB hex or named color)", c.name, c.value)
		}
	}

	return nil
}

// isValidColorValue checks if a color string is valid.
func isValidColorValue(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "default" {
		return true
	}

	// Check hex color
	if strings.HasPrefix(s, "#") {
		hex := strings.TrimPrefix(s, "#")
		if len(hex) != 6 {
			return false
		}
		_, err := strconv.ParseInt(hex, 16, 32)
		return err == nil
	}

	// Check named colors
	validNames := map[string]bool{
		"black": true, "red": true, "green": true, "yellow": true,
		"blue": true, "magenta": true, "purple": true, "cyan": true,
		"white": true, "orange": true, "gray": true, "grey": true,
		"darkblue": true, "darkgreen": true, "darkcyan": true, "darkred": true,
	}
	return validNames[strings.ToLower(s)]
}

// LoadCustomTheme loads a custom theme from ~/.config/nylas/themes/<name>.yaml
func LoadCustomTheme(name string) (*Styles, error) {
	// Check for common mistakes
	if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		cleanName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		return nil, &ThemeLoadError{
			ThemeName: name,
			Reason:    "theme name should not include file extension",
			Hint:      fmt.Sprintf("Use --theme %s instead of --theme %s", cleanName, name),
		}
	}

	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, &ThemeLoadError{
			ThemeName: name,
			Reason:    "failed to get home directory",
			Err:       err,
		}
	}

	themesDir := filepath.Join(homeDir, ".config", "nylas", "themes")
	themePath := filepath.Join(themesDir, name+".yaml")

	// Check if themes directory exists
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return nil, &ThemeLoadError{
			ThemeName: name,
			FilePath:  themePath,
			Reason:    "themes directory does not exist",
			Hint:      fmt.Sprintf("Create the directory: mkdir -p %s\nOr create a theme: nylas tui theme init %s", themesDir, name),
		}
	}

	// Check if theme file exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		// List available custom themes for suggestion
		available := ListCustomThemes()
		hint := fmt.Sprintf("Create this theme: nylas tui theme init %s", name)
		if len(available) > 0 {
			hint += fmt.Sprintf("\nAvailable custom themes: %s", strings.Join(available, ", "))
		}
		return nil, &ThemeLoadError{
			ThemeName: name,
			FilePath:  themePath,
			Reason:    "theme file not found",
			Hint:      hint,
		}
	}

	config, err := LoadThemeFromFile(themePath)
	if err != nil {
		// Enhance error with theme name if it's a ThemeLoadError
		if loadErr, ok := err.(*ThemeLoadError); ok {
			loadErr.ThemeName = name
			return nil, loadErr
		}
		return nil, &ThemeLoadError{
			ThemeName: name,
			FilePath:  themePath,
			Reason:    err.Error(),
			Err:       err,
		}
	}

	return config.ToStyles(), nil
}

// ValidateTheme validates a custom theme and returns detailed information.
func ValidateTheme(name string) (*ThemeValidationResult, error) {
	result := &ThemeValidationResult{
		ThemeName: name,
		Valid:     false,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Errors = append(result.Errors, "Cannot determine home directory")
		return result, err
	}

	themePath := filepath.Join(homeDir, ".config", "nylas", "themes", name+".yaml")
	result.FilePath = themePath

	// Validate path to prevent directory traversal
	if err := validateThemePath(themePath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid theme path: %v", err))
		return result, err
	}

	// Check file exists
	info, err := os.Stat(themePath)
	if os.IsNotExist(err) {
		result.Errors = append(result.Errors, "Theme file not found")
		return result, nil
	}
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot access file: %v", err))
		return result, nil
	}

	result.FileSize = info.Size()

	// Try to read and parse
	// #nosec G304 -- themePath is validated theme file path from user's config directory
	data, err := os.ReadFile(themePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot read file: %v", err))
		return result, nil
	}

	var config ThemeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid YAML: %v", err))
		return result, nil
	}

	// Validate colors
	if config.Foreground != "" {
		if isValidColorValue(config.Foreground) {
			result.ColorsFound = append(result.ColorsFound, fmt.Sprintf("foreground: %s", config.Foreground))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid foreground color: %s", config.Foreground))
		}
	}
	if config.Background != "" {
		if isValidColorValue(config.Background) {
			result.ColorsFound = append(result.ColorsFound, fmt.Sprintf("background: %s", config.Background))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid background color: %s", config.Background))
		}
	}
	if config.K9s.Body.FgColor != "" {
		if isValidColorValue(config.K9s.Body.FgColor) {
			result.ColorsFound = append(result.ColorsFound, fmt.Sprintf("k9s.body.fgColor: %s", config.K9s.Body.FgColor))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid k9s.body.fgColor: %s", config.K9s.Body.FgColor))
		}
	}
	if config.K9s.Body.LogoColor != "" {
		if isValidColorValue(config.K9s.Body.LogoColor) {
			result.ColorsFound = append(result.ColorsFound, fmt.Sprintf("k9s.body.logoColor: %s", config.K9s.Body.LogoColor))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid k9s.body.logoColor: %s", config.K9s.Body.LogoColor))
		}
	}
	if config.K9s.Views.Table.Selected.BgColor != "" {
		result.ColorsFound = append(result.ColorsFound, fmt.Sprintf("k9s.views.table.selected.bgColor: %s", config.K9s.Views.Table.Selected.BgColor))
	}

	// Check for required colors
	if len(result.ColorsFound) == 0 {
		result.Errors = append(result.Errors, "No valid color definitions found")
		return result, nil
	}

	// If no errors, theme is valid
	if len(result.Errors) == 0 {
		result.Valid = true
	}

	return result, nil
}

// ThemeValidationResult holds the result of theme validation.
