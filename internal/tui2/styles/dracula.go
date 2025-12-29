// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// DraculaTheme returns a Dracula color scheme theme.
func DraculaTheme() *Theme {
	t := &Theme{
		Name:       "dracula",
		Foreground: lipgloss.Color("#F8F8F2"),
		Background: lipgloss.Color("#282A36"),
		Primary:    lipgloss.Color("#BD93F9"), // Purple
		Secondary:  lipgloss.Color("#FF79C6"), // Pink
		Accent:     lipgloss.Color("#8BE9FD"), // Cyan
		Success:    lipgloss.Color("#50FA7B"), // Green
		Warning:    lipgloss.Color("#F1FA8C"), // Yellow
		Error:      lipgloss.Color("#FF5555"), // Red
		Info:       lipgloss.Color("#8BE9FD"), // Cyan
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
		Foreground(lipgloss.Color("#6272A4"))

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
