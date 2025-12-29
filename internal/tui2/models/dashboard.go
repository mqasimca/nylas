// Package models provides screen models for the TUI.
package models

import (
	"fmt"
	"image/color"

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

	// Status bar with enhanced styling
	statusBarView := d.statusBar.View()

	// Build dashboard layout
	var sections []string

	// ✨ GLOSSY WELCOME with metallic effect
	panelWidth := 70
	if d.global.WindowSize.Width < 80 {
		panelWidth = d.global.WindowSize.Width - 10
	}

	welcomeTitleText := lipgloss.NewStyle().
		Foreground(d.theme.Primary).
		Bold(true).
		Render("✨ NYLAS CLI ✨")

	welcomeTitle := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, welcomeTitleText)

	welcomeSubtitleText := lipgloss.NewStyle().
		Foreground(d.theme.Secondary).
		Italic(true).
		Render("✦ Premium Terminal Interface ✦")

	welcomeSubtitle := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, welcomeSubtitleText)

	sections = append(sections, welcomeTitle)
	sections = append(sections, welcomeSubtitle)
	sections = append(sections, "")

	// Accent line separator
	separatorLine := styles.AccentLine(d.theme, 60, "─")
	separator := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, separatorLine)
	sections = append(sections, separator)
	sections = append(sections, "")

	// Account info in a glass panel style
	accountInfoText := lipgloss.NewStyle().
		Foreground(d.theme.Dimmed.GetForeground()).
		Render(fmt.Sprintf("🔐 %s  •  %s", d.global.Email, d.global.Provider))
	accountInfo := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, accountInfoText)
	sections = append(sections, accountInfo)
	sections = append(sections, "")

	// Quick Actions with GLOSSY cards
	actionsTitleText := lipgloss.NewStyle().
		Foreground(d.theme.Primary).
		Bold(true).
		Underline(true).
		Render("⚡ QUICK ACTIONS ⚡")

	actionsTitle := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, actionsTitleText)
	sections = append(sections, actionsTitle)
	sections = append(sections, "")

	actions := []struct {
		key   string
		icon  string
		desc  string
		color color.Color
	}{
		{"a", "📧", "Air - View your inbox", d.theme.Primary},
		{"c", "📅", "Calendar - Manage events", d.theme.Secondary},
		{"p", "👥", "Contacts - Manage contacts", d.theme.Accent},
		{"d", "🐛", "Debug - System diagnostics", lipgloss.Color("#FF6B9D")},
		{"s", "⚙️ ", "Settings - Configure app", d.theme.Warning},
		{"?", "❓", "Help - Keyboard shortcuts", d.theme.Info},
	}

	// Add each action item individually
	for _, action := range actions {
		// Glossy key badge
		keyBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(action.color).
			Bold(true).
			Padding(0, 1).
			Render(action.key)

		// Action description with icon
		descText := fmt.Sprintf("%s  %s", action.icon, action.desc)

		// Create action line with badge and description
		line := fmt.Sprintf("   %s   %s", keyBadge, descText)

		sections = append(sections, line)
	}
	sections = append(sections, "")
	sections = append(sections, "")

	// Theme info with premium styling
	themeInfoText := lipgloss.NewStyle().
		Foreground(d.theme.Dimmed.GetForeground()).
		Italic(true).
		Render(fmt.Sprintf("◆ Theme: %s ◆ Press 't' to cycle ◆", d.theme.Name))
	themeInfo := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, themeInfoText)
	sections = append(sections, themeInfo)

	// Join all sections (left-aligned within the panel)
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Premium double border with shadow
	bordered := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(d.theme.Primary).
		BorderBackground(lipgloss.Color("#0a0a0a")).
		Background(lipgloss.Color("#0f0f0f")).
		Padding(2, 4).
		Width(panelWidth).
		Render(content)

	// Center content
	centered := lipgloss.Place(
		d.global.WindowSize.Width,
		d.global.WindowSize.Height-4, // Leave room for status bar and footer
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)

	// Footer bar
	footerBarView := d.footerBar.View()

	// Combine all with spacing
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBarView,
		centered,
		footerBarView,
	)

	return tea.NewView(fullView)
}
