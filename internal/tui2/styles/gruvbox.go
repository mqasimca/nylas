// Package styles provides Lip Gloss styling for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// GruvboxTheme returns a Gruvbox dark theme.
func GruvboxTheme() *Theme {
	t := &Theme{
		Name:       "gruvbox",
		Foreground: lipgloss.Color("#EBDBB2"), // Light1
		Background: lipgloss.Color("#282828"), // Dark0
		Primary:    lipgloss.Color("#FE8019"), // Bright orange
		Secondary:  lipgloss.Color("#FABD2F"), // Bright yellow
		Accent:     lipgloss.Color("#83A598"), // Bright blue
		Success:    lipgloss.Color("#B8BB26"), // Bright green
		Warning:    lipgloss.Color("#FABD2F"), // Bright yellow
		Error:      lipgloss.Color("#FB4934"), // Bright red
		Info:       lipgloss.Color("#83A598"), // Bright blue
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
		Foreground(lipgloss.Color("#928374"))

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
