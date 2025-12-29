package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
				// #nosec G115 -- & 0xFF masks to 0-255 range, no overflow possible
				r := int32((val >> 16) & 0xFF) // #nosec G115
				g := int32((val >> 8) & 0xFF)  // #nosec G115
				b := int32(val & 0xFF)         // #nosec G115
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
