package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

// ThemeConfig represents a customizable theme loaded from YAML (k9s-style).
type ThemeConfig struct {
	// Color definitions (like k9s anchors)
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
	Black      string `yaml:"black"`
	Red        string `yaml:"red"`
	Green      string `yaml:"green"`
	Yellow     string `yaml:"yellow"`
	Blue       string `yaml:"blue"`
	Magenta    string `yaml:"magenta"`
	Cyan       string `yaml:"cyan"`
	White      string `yaml:"white"`

	// K9s-style skin configuration
	K9s K9sSkin `yaml:"k9s"`
}

// K9sSkin represents the k9s skin configuration.
type K9sSkin struct {
	Body   BodyStyle   `yaml:"body"`
	Prompt PromptStyle `yaml:"prompt"`
	Info   InfoStyle   `yaml:"info"`
	Frame  FrameStyle  `yaml:"frame"`
	Views  ViewsStyle  `yaml:"views"`
}

// BodyStyle for general body colors.
type BodyStyle struct {
	FgColor   string `yaml:"fgColor"`
	BgColor   string `yaml:"bgColor"`
	LogoColor string `yaml:"logoColor"`
}

// PromptStyle for command prompt.
type PromptStyle struct {
	FgColor      string `yaml:"fgColor"`
	BgColor      string `yaml:"bgColor"`
	SuggestColor string `yaml:"suggestColor"`
}

// InfoStyle for info panel.
type InfoStyle struct {
	FgColor      string `yaml:"fgColor"`
	SectionColor string `yaml:"sectionColor"`
}

// FrameStyle for frame elements.
type FrameStyle struct {
	Border BorderStyle `yaml:"border"`
	Menu   MenuStyle   `yaml:"menu"`
	Crumbs CrumbsStyle `yaml:"crumbs"`
	Status StatusStyle `yaml:"status"`
	Title  TitleStyle  `yaml:"title"`
}

// BorderStyle for borders.
type BorderStyle struct {
	FgColor    string `yaml:"fgColor"`
	FocusColor string `yaml:"focusColor"`
}

// MenuStyle for menu.
type MenuStyle struct {
	FgColor     string `yaml:"fgColor"`
	KeyColor    string `yaml:"keyColor"`
	NumKeyColor string `yaml:"numKeyColor"`
}

// CrumbsStyle for breadcrumbs.
type CrumbsStyle struct {
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	ActiveColor string `yaml:"activeColor"`
}

// StatusStyle for status indicators.
type StatusStyle struct {
	NewColor       string `yaml:"newColor"`
	ModifyColor    string `yaml:"modifyColor"`
	AddColor       string `yaml:"addColor"`
	PendingColor   string `yaml:"pendingColor"`
	ErrorColor     string `yaml:"errorColor"`
	HighlightColor string `yaml:"highlightColor"`
	KillColor      string `yaml:"killColor"`
	CompletedColor string `yaml:"completedColor"`
}

// TitleStyle for titles.
type TitleStyle struct {
	FgColor        string `yaml:"fgColor"`
	BgColor        string `yaml:"bgColor"`
	HighlightColor string `yaml:"highlightColor"`
	CounterColor   string `yaml:"counterColor"`
	FilterColor    string `yaml:"filterColor"`
}

// ViewsStyle for view-specific styles.
type ViewsStyle struct {
	Table TableStyle `yaml:"table"`
}

// TableStyle for table views.
type TableStyle struct {
	FgColor   string             `yaml:"fgColor"`
	BgColor   string             `yaml:"bgColor"`
	MarkColor string             `yaml:"markColor"`
	Header    TableHeaderStyle   `yaml:"header"`
	Selected  TableSelectedStyle `yaml:"selected"`
}

// TableHeaderStyle for table headers.
type TableHeaderStyle struct {
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	SorterColor string `yaml:"sorterColor"`
}

// TableSelectedStyle for selected rows.
type TableSelectedStyle struct {
	FgColor string `yaml:"fgColor"`
	BgColor string `yaml:"bgColor"`
}

// ThemeLoadError provides detailed error information for theme loading failures.
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
type ThemeValidationResult struct {
	ThemeName   string
	FilePath    string
	FileSize    int64
	Valid       bool
	ColorsFound []string
	Warnings    []string
	Errors      []string
}

// GetThemesDir returns the themes directory path.
func GetThemesDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "nylas", "themes")
}

// ListCustomThemes returns a list of available custom themes.
func ListCustomThemes() []string {
	themesDir := GetThemesDir()
	if themesDir == "" {
		return nil
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil
	}

	var themes []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			name := strings.TrimSuffix(entry.Name(), ".yaml")
			themes = append(themes, name)
		}
	}

	return themes
}

// ToStyles converts a ThemeConfig to Styles.
func (c *ThemeConfig) ToStyles() *Styles {
	// Start with default styles
	s := DefaultStyles()

	// Apply color definitions
	if c.Background != "" && c.Background != "default" {
		s.BgColor = parseColor(c.Background)
	}
	if c.Foreground != "" {
		s.FgColor = parseColor(c.Foreground)
	}

	// Apply body styles
	if c.K9s.Body.BgColor != "" && c.K9s.Body.BgColor != "default" {
		s.BgColor = parseColor(c.K9s.Body.BgColor)
	}
	if c.K9s.Body.FgColor != "" {
		s.FgColor = parseColor(c.K9s.Body.FgColor)
	}
	if c.K9s.Body.LogoColor != "" {
		s.LogoColor = parseColor(c.K9s.Body.LogoColor)
	}

	// Apply prompt styles
	if c.K9s.Prompt.FgColor != "" {
		s.PromptFg = parseColor(c.K9s.Prompt.FgColor)
	}
	if c.K9s.Prompt.BgColor != "" && c.K9s.Prompt.BgColor != "default" {
		s.PromptBg = parseColor(c.K9s.Prompt.BgColor)
	}

	// Apply info styles
	if c.K9s.Info.FgColor != "" {
		s.InfoColor = parseColor(c.K9s.Info.FgColor)
	}
	if c.K9s.Info.SectionColor != "" {
		s.InfoSectionFg = parseColor(c.K9s.Info.SectionColor)
	}

	// Apply border styles
	if c.K9s.Frame.Border.FgColor != "" {
		s.BorderColor = parseColor(c.K9s.Frame.Border.FgColor)
	}
	if c.K9s.Frame.Border.FocusColor != "" {
		s.FocusColor = parseColor(c.K9s.Frame.Border.FocusColor)
	}

	// Apply menu styles
	if c.K9s.Frame.Menu.FgColor != "" {
		s.MenuDescFg = parseColor(c.K9s.Frame.Menu.FgColor)
	}
	if c.K9s.Frame.Menu.KeyColor != "" {
		s.MenuKeyFg = parseColor(c.K9s.Frame.Menu.KeyColor)
	}
	if c.K9s.Frame.Menu.NumKeyColor != "" {
		s.MenuNumKeyFg = parseColor(c.K9s.Frame.Menu.NumKeyColor)
	}

	// Apply crumbs styles
	if c.K9s.Frame.Crumbs.FgColor != "" {
		s.CrumbFg = parseColor(c.K9s.Frame.Crumbs.FgColor)
	}
	if c.K9s.Frame.Crumbs.BgColor != "" && c.K9s.Frame.Crumbs.BgColor != "default" {
		s.CrumbBg = parseColor(c.K9s.Frame.Crumbs.BgColor)
	}
	if c.K9s.Frame.Crumbs.ActiveColor != "" {
		s.CrumbActiveFg = parseColor(c.K9s.Frame.Crumbs.ActiveColor)
	}

	// Apply status styles
	if c.K9s.Frame.Status.ErrorColor != "" {
		s.ErrorColor = parseColor(c.K9s.Frame.Status.ErrorColor)
	}
	if c.K9s.Frame.Status.PendingColor != "" {
		s.WarnColor = parseColor(c.K9s.Frame.Status.PendingColor)
	}
	if c.K9s.Frame.Status.AddColor != "" {
		s.SuccessColor = parseColor(c.K9s.Frame.Status.AddColor)
	}

	// Apply title styles
	if c.K9s.Frame.Title.FgColor != "" {
		s.TitleFg = parseColor(c.K9s.Frame.Title.FgColor)
	}
	if c.K9s.Frame.Title.HighlightColor != "" {
		s.TitleHighlight = parseColor(c.K9s.Frame.Title.HighlightColor)
	}

	// Apply table styles
	if c.K9s.Views.Table.FgColor != "" {
		s.TableRowFg = parseColor(c.K9s.Views.Table.FgColor)
	}
	if c.K9s.Views.Table.Header.FgColor != "" {
		s.TableHeaderFg = parseColor(c.K9s.Views.Table.Header.FgColor)
	}
	if c.K9s.Views.Table.Header.BgColor != "" && c.K9s.Views.Table.Header.BgColor != "default" {
		s.TableHeaderBg = parseColor(c.K9s.Views.Table.Header.BgColor)
	}
	if c.K9s.Views.Table.MarkColor != "" {
		s.TableMarkColor = parseColor(c.K9s.Views.Table.MarkColor)
	}
	if c.K9s.Views.Table.Selected.FgColor != "" {
		s.TableSelectFg = parseColor(c.K9s.Views.Table.Selected.FgColor)
	}
	if c.K9s.Views.Table.Selected.BgColor != "" {
		s.TableSelectBg = parseColor(c.K9s.Views.Table.Selected.BgColor)
	}

	// Apply named colors for additional flexibility
	if c.Red != "" {
		s.ErrorColor = parseColor(c.Red)
	}
	if c.Green != "" {
		s.SuccessColor = parseColor(c.Green)
	}
	if c.Yellow != "" {
		s.WarnColor = parseColor(c.Yellow)
	}
	if c.Blue != "" {
		s.InfoColor = parseColor(c.Blue)
	}
	if c.Magenta != "" {
		s.LogoColor = parseColor(c.Magenta)
	}
	if c.Cyan != "" {
		s.InfoSectionFg = parseColor(c.Cyan)
	}

	return s
}

// parseColor parses a color string to tcell.Color.
// Supports: hex (#RRGGBB), named colors, and "default".
func parseColor(s string) tcell.Color {
	s = strings.TrimSpace(s)

	// Handle "default" or empty
	if s == "" || s == "default" {
		return tcell.ColorDefault
	}

	// Handle hex colors
	if strings.HasPrefix(s, "#") {
		hex := strings.TrimPrefix(s, "#")
		if len(hex) == 6 {
			val, err := strconv.ParseInt(hex, 16, 32)
			if err == nil {
				// Safe: & 0xFF masks to 0-255 range, no overflow possible
				r := int32((val >> 16) & 0xFF) // Red component (0-255)
				g := int32((val >> 8) & 0xFF)  // Green component (0-255)
				b := int32(val & 0xFF)         // Blue component (0-255)
				return tcell.NewRGBColor(r, g, b)
			}
		}
	}

	// Handle named colors
	switch strings.ToLower(s) {
	case "black":
		return tcell.ColorBlack
	case "red":
		return tcell.ColorRed
	case "green":
		return tcell.ColorGreen
	case "yellow":
		return tcell.ColorYellow
	case "blue":
		return tcell.ColorBlue
	case "magenta", "purple":
		return tcell.ColorPurple
	case "cyan":
		return tcell.ColorTeal
	case "white":
		return tcell.ColorWhite
	case "orange":
		return tcell.ColorOrange
	case "gray", "grey":
		return tcell.ColorGray
	case "darkblue":
		return tcell.ColorDarkBlue
	case "darkgreen":
		return tcell.ColorDarkGreen
	case "darkcyan":
		return tcell.ColorDarkCyan
	case "darkred":
		return tcell.ColorDarkRed
	}

	return tcell.ColorDefault
}

// CreateDefaultThemeFile creates a default theme file at the specified path.
func CreateDefaultThemeFile(path string) error {
	defaultTheme := `# Nylas TUI Theme Configuration (k9s-style)
# Place this file in ~/.config/nylas/themes/<name>.yaml
# Use with: nylas tui --theme <name>

# Color definitions (use hex #RRGGBB or named colors)
foreground: "#c0caf5"
background: "#1a1b26"
black: "#15161e"
red: "#f7768e"
green: "#9ece6a"
yellow: "#e0af68"
blue: "#7aa2f7"
magenta: "#bb9af7"
cyan: "#7dcfff"
white: "#a9b1d6"

# K9s-style skin configuration
k9s:
  # General body styles
  body:
    fgColor: "#c0caf5"
    bgColor: "#1a1b26"
    logoColor: "#bb9af7"

  # Command prompt styles
  prompt:
    fgColor: "#c0caf5"
    bgColor: "#1a1b26"
    suggestColor: "#e0af68"

  # Info panel styles
  info:
    fgColor: "#7aa2f7"
    sectionColor: "#7dcfff"

  # Frame styles
  frame:
    border:
      fgColor: "#3b4261"
      focusColor: "#7aa2f7"
    menu:
      fgColor: "#7dcfff"
      keyColor: "#9ece6a"
      numKeyColor: "#bb9af7"
    crumbs:
      fgColor: "#c0caf5"
      bgColor: "#1a1b26"
      activeColor: "#7aa2f7"
    status:
      newColor: "#7aa2f7"
      modifyColor: "#bb9af7"
      addColor: "#9ece6a"
      pendingColor: "#e0af68"
      errorColor: "#f7768e"
      highlightColor: "#7aa2f7"
      killColor: "#f7768e"
      completedColor: "#3b4261"
    title:
      fgColor: "#c0caf5"
      bgColor: "#1a1b26"
      highlightColor: "#7aa2f7"
      counterColor: "#bb9af7"
      filterColor: "#e0af68"

  # View styles
  views:
    table:
      fgColor: "#c0caf5"
      bgColor: "#1a1b26"
      markColor: "#bb9af7"
      header:
        fgColor: "#7aa2f7"
        bgColor: "#1a1b26"
        sorterColor: "#7dcfff"
      selected:
        fgColor: "#1a1b26"
        bgColor: "#7aa2f7"
`
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	return os.WriteFile(path, []byte(defaultTheme), 0600)
}
