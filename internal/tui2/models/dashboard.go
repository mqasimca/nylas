// Package models provides screen models for the TUI.
package models

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

// ScreenType represents different screens in the application.
type ScreenType int

const (
	// ScreenSplash is the initial splash screen.
	ScreenSplash ScreenType = iota
	// ScreenDashboard is the main dashboard view.
	ScreenDashboard
	// ScreenMessages is the email list view.
	ScreenMessages
	// ScreenMessageDetail is the message detail view.
	ScreenMessageDetail
	// ScreenCompose is the compose email view.
	ScreenCompose
	// ScreenCalendar is the calendar view.
	ScreenCalendar
	// ScreenContacts is the contacts view.
	ScreenContacts
	// ScreenSettings is the settings view.
	ScreenSettings
	// ScreenHelp is the help view.
	ScreenHelp
	// ScreenDebug is the debug panel view.
	ScreenDebug
)

// NavigateMsg is sent to navigate to a new screen.
type NavigateMsg struct {
	Screen ScreenType
	Data   interface{}
}

// Dashboard is the main dashboard screen.
type Dashboard struct {
	global    *state.GlobalState
	theme     *styles.Theme
	statusBar *components.StatusBar
	footerBar *components.FooterBar
}

// NewDashboard creates a new dashboard screen.
func NewDashboard(global *state.GlobalState) *Dashboard {
	theme := styles.GetTheme(global.Theme)
	statusBar := components.NewStatusBar(theme, global.Email)
	footerBar := components.NewFooterBar(theme, global.Email)

	// Set default keybindings
	footerBar.SetBindings([]components.KeyBinding{
		{Key: "a", Description: "Air"},
		{Key: "c", Description: "Calendar"},
		{Key: "p", Description: "People"},
		{Key: "d", Description: "Debug"},
		{Key: "s", Description: "Settings"},
		{Key: "?", Description: "Help"},
		{Key: "Ctrl+C", Description: "Quit"},
	})

	return &Dashboard{
		global:    global,
		theme:     theme,
		statusBar: statusBar,
		footerBar: footerBar,
	}
}

// Init implements tea.Model.
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update component sizes
		d.statusBar.SetWidth(msg.Width)
		d.footerBar.SetWidth(msg.Width)
		d.global.SetWindowSize(msg.Width, msg.Height)

	case themeChangedMsg:
		// Theme was changed, show status message
		d.global.SetStatus(fmt.Sprintf("Theme changed to: %s", msg.theme), 0)
		return d, nil

	case tea.KeyMsg:
		// Use msg.String() for modifier combos (key.Text is empty for Ctrl+key)
		keyStr := msg.String()

		// Check ctrl+c (handled by app.go, but included for clarity)
		if keyStr == "ctrl+c" {
			return d, tea.Quit
		}

		// Check regular keys
		switch keyStr {
		case "a":
			// Navigate to Air (messages)
			return d, navigateToMessages()
		case "c":
			// Navigate to calendar
			return d, navigateToCalendar()
		case "p":
			// Navigate to contacts
			return d, navigateToContacts()
		case "d":
			// Navigate to debug panel
			return d, navigateToDebug()
		case "t":
			// Cycle through themes
			return d, d.cycleTheme()
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

func navigateToDebug() tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: ScreenDebug}
	}
}

// cycleTheme cycles to the next available theme.
func (d *Dashboard) cycleTheme() tea.Cmd {
	themes := styles.ListAvailableThemes()
	currentIndex := 0

	// Find current theme index
	for i, theme := range themes {
		if theme == d.global.Theme {
			currentIndex = i
			break
		}
	}

	// Move to next theme (wrap around)
	nextIndex := (currentIndex + 1) % len(themes)
	nextTheme := themes[nextIndex]

	// Update global theme
	d.global.Theme = nextTheme
	d.theme = styles.GetTheme(nextTheme)
	d.statusBar = components.NewStatusBar(d.theme, d.global.Email)
	d.footerBar = components.NewFooterBar(d.theme, d.global.Email)

	// Restore widths after creating new components
	d.statusBar.SetWidth(d.global.WindowSize.Width)
	d.footerBar.SetWidth(d.global.WindowSize.Width)

	// Restore keybindings
	d.footerBar.SetBindings([]components.KeyBinding{
		{Key: "a", Description: "Air"},
		{Key: "c", Description: "Calendar"},
		{Key: "p", Description: "People"},
		{Key: "d", Description: "Debug"},
		{Key: "s", Description: "Settings"},
		{Key: "?", Description: "Help"},
		{Key: "Ctrl+C", Description: "Quit"},
	})

	// Save theme to config file
	return func() tea.Msg {
		config, err := utils.LoadConfig()
		if err != nil {
			config = utils.DefaultConfig()
		}
		config.Theme = nextTheme
		_ = utils.SaveConfig(config) // Ignore error, not critical

		return themeChangedMsg{theme: nextTheme}
	}
}

// themeChangedMsg is sent when the theme is changed.
type themeChangedMsg struct {
	theme string
}

// View implements tea.Model.
func (d *Dashboard) View() tea.View {
	// Update status bar
	d.statusBar.Update()
	d.statusBar.SetOnline(true)
	d.statusBar.SetUnreadCount(0)
	d.statusBar.SetEventCount(0)

	// Status bar at top
	statusBarView := d.statusBar.View()

	var lines []string

	// App name and current view badge (k9s style)
	appName := lipgloss.NewStyle().
		Foreground(d.theme.Primary).
		Bold(true).
		Render("NYLAS")
	viewBadge := lipgloss.NewStyle().
		Foreground(d.theme.Background).
		Background(d.theme.Primary).
		Padding(0, 1).
		Render(":dashboard")
	lines = append(lines, fmt.Sprintf("%s\n%s", appName, viewBadge))
	lines = append(lines, "")

	// Quick Navigation header
	navHeader := lipgloss.NewStyle().
		Foreground(d.theme.Primary).
		Bold(true).
		Render("Quick Navigation")
	lines = append(lines, navHeader)
	lines = append(lines, "")

	// Navigation items - vim style commands with key badges and Nerd Font icons
	navItems := []struct {
		key   string
		icon  string
		cmd   string
		label string
		desc  string
	}{
		{"a", "\uf0e0", ":a", "Air", "Email messages"},       // nf-fa-envelope
		{"e", "\uf073", ":e", "Events", "Calendar events"},   // nf-fa-calendar
		{"c", "\uf0c0", ":c", "Contacts", "Contacts"},        // nf-fa-users
		{"w", "\uf0e8", ":w", "Webhooks", "Webhooks"},        // nf-fa-sitemap
		{"d", "\uf188", ":d", "Debug", "Debug panel"},        // nf-fa-bug
		{"s", "\uf013", ":s", "Settings", "Settings"},        // nf-fa-cog
		{"?", "\uf059", ":?", "Help", "Help & shortcuts"},    // nf-fa-question_circle
	}

	// Key badge style (nice button)
	keyBadgeStyle := lipgloss.NewStyle().
		Foreground(d.theme.Background).
		Background(d.theme.Primary).
		Bold(true).
		Padding(0, 1)

	iconStyle := lipgloss.NewStyle().Foreground(d.theme.Accent).Width(2)
	cmdStyle := lipgloss.NewStyle().Foreground(d.theme.Warning)
	labelStyle := lipgloss.NewStyle().Foreground(d.theme.Secondary).Width(12)
	descStyle := lipgloss.NewStyle().Foreground(d.theme.Info)

	for _, item := range navItems {
		keyBadge := keyBadgeStyle.Render(item.key)
		icon := iconStyle.Render(item.icon)
		cmd := cmdStyle.Render(fmt.Sprintf("%-6s", item.cmd))
		label := labelStyle.Render(item.label)
		desc := descStyle.Render(item.desc)
		lines = append(lines, fmt.Sprintf("  %s %s %s%s%s", keyBadge, icon, cmd, label, desc))
	}
	lines = append(lines, "")

	// Command mode hint
	hintStyle := lipgloss.NewStyle().Foreground(d.theme.Primary)
	lines = append(lines, hintStyle.Render("Press : to enter command mode"))

	// Join content
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Calculate available height for content
	contentHeight := d.global.WindowSize.Height - 3 // status bar + footer

	// Pad content to fill available space (push footer to bottom)
	contentView := lipgloss.NewStyle().
		Width(d.global.WindowSize.Width).
		Height(contentHeight).
		Padding(0, 1).
		Render(content)

	// Footer bar
	footerBarView := d.footerBar.View()

	// Combine all layers
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBarView,
		contentView,
		footerBarView,
	)

	return tea.NewView(fullView)
}
