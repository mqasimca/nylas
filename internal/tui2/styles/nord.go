// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// NordTheme returns a Nord color scheme theme.
func NordTheme() *Theme {
	t := &Theme{
		Name:       "nord",
		Foreground: lipgloss.Color("#ECEFF4"), // Snow Storm
		Background: lipgloss.Color("#2E3440"), // Polar Night
		Primary:    lipgloss.Color("#88C0D0"), // Frost (cyan)
		Secondary:  lipgloss.Color("#81A1C1"), // Frost (blue)
		Accent:     lipgloss.Color("#A3BE8C"), // Aurora (green)
		Success:    lipgloss.Color("#A3BE8C"), // Aurora (green)
		Warning:    lipgloss.Color("#EBCB8B"), // Aurora (yellow)
		Error:      lipgloss.Color("#BF616A"), // Aurora (red)
		Info:       lipgloss.Color("#88C0D0"), // Frost (cyan)
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
		Foreground(lipgloss.Color("#4C566A"))

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
