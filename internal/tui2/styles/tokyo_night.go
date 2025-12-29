// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// TokyoNightTheme returns a Tokyo Night theme.
func TokyoNightTheme() *Theme {
	t := &Theme{
		Name:       "tokyo_night",
		Foreground: lipgloss.Color("#C0CAF5"), // Foreground
		Background: lipgloss.Color("#1A1B26"), // Background
		Primary:    lipgloss.Color("#7AA2F7"), // Blue
		Secondary:  lipgloss.Color("#BB9AF7"), // Purple
		Accent:     lipgloss.Color("#7DCFFF"), // Cyan
		Success:    lipgloss.Color("#9ECE6A"), // Green
		Warning:    lipgloss.Color("#E0AF68"), // Yellow
		Error:      lipgloss.Color("#F7768E"), // Red
		Info:       lipgloss.Color("#7DCFFF"), // Cyan
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
		Foreground(lipgloss.Color("#565F89"))

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
