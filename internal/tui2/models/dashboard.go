// Package models provides screen models for the TUI.
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// ScreenType represents different screens in the application.
type ScreenType int

const (
	// ScreenDashboard is the main dashboard view.
	ScreenDashboard ScreenType = iota
	// ScreenMessages is the email list view.
	ScreenMessages
	// ScreenMessageDetail is the message detail view.
	ScreenMessageDetail
	// ScreenCalendar is the calendar view.
	ScreenCalendar
	// ScreenContacts is the contacts view.
	ScreenContacts
	// ScreenSettings is the settings view.
	ScreenSettings
	// ScreenHelp is the help view.
	ScreenHelp
)

// NavigateMsg is sent to navigate to a new screen.
type NavigateMsg struct {
	Screen ScreenType
	Data   interface{}
}

// Dashboard is the main dashboard screen.
type Dashboard struct {
	global *state.GlobalState
	theme  *styles.Theme
}

// NewDashboard creates a new dashboard screen.
func NewDashboard(global *state.GlobalState) *Dashboard {
	return &Dashboard{
		global: global,
		theme:  styles.GetTheme(global.Theme),
	}
}

// Init implements tea.Model.
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return d, tea.Quit
		case "m":
			// Navigate to messages
			return d, navigateToMessages()
		case "c":
			// Navigate to calendar
			return d, navigateToCalendar()
		case "p":
			// Navigate to contacts
			return d, navigateToContacts()
		case "s":
			// Navigate to settings
			return d, navigateToSettings()
		case "?":
			// Navigate to help
			return d, navigateToHelp()
		}
	}

	return d, nil
}

// Navigation commands
func navigateToMessages() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenMessages}
	}
}

func navigateToCalendar() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenCalendar}
	}
}

func navigateToContacts() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenContacts}
	}
}

func navigateToSettings() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenSettings}
	}
}

func navigateToHelp() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenHelp}
	}
}

// View implements tea.Model.
func (d *Dashboard) View() string {
	// Build dashboard layout
	var sections []string

	// Header
	header := d.theme.Title.Render("Nylas CLI - Dashboard")
	accountInfo := d.theme.Subtitle.Render(fmt.Sprintf("Account: %s (%s)", d.global.Email, d.global.Provider))
	sections = append(sections, header)
	sections = append(sections, accountInfo)
	sections = append(sections, "")

	// Welcome message
	welcome := lipgloss.NewStyle().
		Foreground(d.theme.Foreground).
		Padding(1, 2).
		Render("Welcome to Nylas CLI Terminal UI (Bubble Tea Edition)")
	sections = append(sections, welcome)
	sections = append(sections, "")

	// Quick Actions
	actionsTitle := d.theme.Title.Render("Quick Actions:")
	sections = append(sections, actionsTitle)
	sections = append(sections, "")

	actions := []struct {
		key  string
		desc string
	}{
		{"m", "Messages - View your inbox"},
		{"c", "Calendar - Manage events"},
		{"p", "Contacts - Manage contacts"},
		{"s", "Settings - Configure preferences"},
		{"?", "Help - Show keyboard shortcuts"},
		{"Ctrl+C", "Quit - Exit application"},
	}

	for _, action := range actions {
		keyStyle := d.theme.KeyBinding.Render(fmt.Sprintf("[%s]", action.key))
		descStyle := d.theme.KeyDescription.Render(action.desc)
		line := fmt.Sprintf("  %s  %s", keyStyle, descStyle)
		sections = append(sections, line)
	}

	sections = append(sections, "")

	// Help text at bottom
	help := d.theme.Help.Render("Press any key listed above to navigate")
	sections = append(sections, help)

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add border
	bordered := d.theme.Border.Render(content)

	return bordered
}
