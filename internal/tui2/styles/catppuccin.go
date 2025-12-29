// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// CatppuccinTheme returns a Catppuccin Mocha theme.
func CatppuccinTheme() *Theme {
	t := &Theme{
		Name:       "catppuccin",
		Foreground: lipgloss.Color("#CDD6F4"), // Text
		Background: lipgloss.Color("#1E1E2E"), // Base
		Primary:    lipgloss.Color("#CBA6F7"), // Mauve
		Secondary:  lipgloss.Color("#F5C2E7"), // Pink
		Accent:     lipgloss.Color("#94E2D5"), // Teal
		Success:    lipgloss.Color("#A6E3A1"), // Green
		Warning:    lipgloss.Color("#F9E2AF"), // Yellow
		Error:      lipgloss.Color("#F38BA8"), // Red
		Info:       lipgloss.Color("#89B4FA"), // Blue
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
		Foreground(t.Background).
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
		Foreground(lipgloss.Color("#6C7086"))

	t.Help = lipgloss.NewStyle().
		Foreground(t.Accent).
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
