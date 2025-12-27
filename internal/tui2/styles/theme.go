// Package styles provides Lip Gloss styling for the TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Theme holds all style definitions.
type Theme struct {
	Name string

	// Base colors
	Foreground lipgloss.Color
	Background lipgloss.Color
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// Component styles
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	Status         lipgloss.Style
	Border         lipgloss.Style
	Selected       lipgloss.Style
	Dimmed         lipgloss.Style
	Help           lipgloss.Style
	Error_         lipgloss.Style
	Success_       lipgloss.Style
	KeyBinding     lipgloss.Style
	KeyDescription lipgloss.Style
}

// DefaultTheme returns the default k9s-style theme.
func DefaultTheme() *Theme {
	t := &Theme{
		Name:       "k9s",
		Foreground: lipgloss.Color("#FFFFFF"),
		Background: lipgloss.Color("#000000"),
		Primary:    lipgloss.Color("#00D9FF"),
		Secondary:  lipgloss.Color("#FF9500"),
		Accent:     lipgloss.Color("#00FF9F"),
		Success:    lipgloss.Color("#00FF00"),
		Warning:    lipgloss.Color("#FFFF00"),
		Error:      lipgloss.Color("#FF0000"),
		Info:       lipgloss.Color("#00D9FF"),
	}

	// Build component styles
	t.Title = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Padding(0, 1)

	t.Status = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(t.Primary).
		Padding(0, 1)

	t.Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	t.Selected = lipgloss.NewStyle().
		Foreground(t.Background).
		Background(t.Primary).
		Bold(true)

	t.Dimmed = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	t.Help = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Padding(0, 1)

	t.Error_ = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true)

	t.Success_ = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	t.KeyBinding = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	t.KeyDescription = lipgloss.NewStyle().
		Foreground(t.Dimmed.GetForeground())

	return t
}

// GetTheme returns a theme by name.
func GetTheme(name string) *Theme {
	switch name {
	case "k9s":
		return DefaultTheme()
	default:
		return DefaultTheme()
	}
}
