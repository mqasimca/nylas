// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// CyberpunkTheme returns a neon cyberpunk theme with pink, cyan, and yellow.
func CyberpunkTheme() *Theme {
	t := &Theme{
		Name:       "cyberpunk",
		Foreground: lipgloss.Color("#FFFFFF"),
		Background: lipgloss.Color("#0A0E27"),
		Primary:    lipgloss.Color("#FF6B9D"), // Neon pink
		Secondary:  lipgloss.Color("#00D9FF"), // Cyan
		Accent:     lipgloss.Color("#FFED4E"), // Neon yellow
		Success:    lipgloss.Color("#39FF14"), // Neon green
		Warning:    lipgloss.Color("#FF9500"), // Orange
		Error:      lipgloss.Color("#FF006E"), // Hot pink
		Info:       lipgloss.Color("#00D9FF"), // Cyan
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
		Foreground(lipgloss.Color("#666B7A"))

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
