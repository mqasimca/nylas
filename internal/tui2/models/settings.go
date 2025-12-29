// Package models provides screen models for the TUI.
package models

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

// SettingsScreen shows application settings.
type SettingsScreen struct {
	global            *state.GlobalState
	theme             *styles.Theme
	selectedTheme     string
	enabledAnimations bool
	splashDuration    int
	showStatusBar     bool
	showFooter        bool
	cursor            int
	saved             bool
	statusMessage     string
}

// NewSettingsScreen creates a new settings screen.
func NewSettingsScreen(global *state.GlobalState) *SettingsScreen {
	theme := styles.GetTheme(global.Theme)

	// Load current config
	config, _ := utils.LoadConfig()

	s := &SettingsScreen{
		global:            global,
		theme:             theme,
		selectedTheme:     config.Theme,
		enabledAnimations: config.AnimationsEnabled,
		splashDuration:    config.SplashDurationSec,
		showStatusBar:     config.ShowStatusBar,
		showFooter:        config.ShowFooter,
		cursor:            0,
		saved:             false,
	}

	return s
}

// Init implements tea.Model.
func (s *SettingsScreen) Init() tea.Cmd {
	// Huh uses v1 bubbletea, so we can't return its Cmd directly
	// Just return nil and let form initialize itself
	return nil
}

// Update implements tea.Model.
func (s *SettingsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.global.SetWindowSize(msg.Width, msg.Height)

	case settingsSavedMsg:
		if msg.err != nil {
			s.statusMessage = fmt.Sprintf("Error saving: %v", msg.err)
		} else if msg.success {
			s.statusMessage = "✓ Settings saved successfully!"
			s.saved = true
			// Apply theme change immediately
			s.global.Theme = s.selectedTheme
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return s, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return s, tea.Quit
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < 4 { // 5 settings total
				s.cursor++
			}
		case "enter", " ":
			s.toggleSetting()
		case "ctrl+s", "s":
			// Save settings
			return s, s.saveSettings()
		case "left", "h":
			if s.cursor == 0 {
				// Cycle themes backward
				s.cycleTheme(-1)
			}
		case "right", "l":
			if s.cursor == 0 {
				// Cycle themes forward
				s.cycleTheme(1)
			}
		case "+":
			if s.cursor == 2 && s.splashDuration < 10 {
				s.splashDuration++
			}
		case "-":
			if s.cursor == 2 && s.splashDuration > 1 {
				s.splashDuration--
			}
		}
	}

	return s, nil
}

// cycleTheme cycles through available themes.
func (s *SettingsScreen) cycleTheme(direction int) {
	themes := styles.ListAvailableThemes()
	currentIndex := 0
	for i, theme := range themes {
		if theme == s.selectedTheme {
			currentIndex = i
			break
		}
	}

	newIndex := currentIndex + direction
	if newIndex < 0 {
		newIndex = len(themes) - 1
	} else if newIndex >= len(themes) {
		newIndex = 0
	}

	s.selectedTheme = themes[newIndex]
	// Update theme immediately for preview
	s.theme = styles.GetTheme(s.selectedTheme)
}

// toggleSetting toggles the current setting.
func (s *SettingsScreen) toggleSetting() {
	switch s.cursor {
	case 0:
		// Theme - cycle forward
		s.cycleTheme(1)
	case 1:
		s.enabledAnimations = !s.enabledAnimations
	case 2:
		// Splash duration - increment
		if s.splashDuration < 10 {
			s.splashDuration++
		} else {
			s.splashDuration = 1
		}
	case 3:
		s.showStatusBar = !s.showStatusBar
	case 4:
		s.showFooter = !s.showFooter
	}
}

// saveSettings saves settings to config file.
func (s *SettingsScreen) saveSettings() tea.Cmd {
	return func() tea.Msg {
		config := &utils.TUIConfig{
			Theme:             s.selectedTheme,
			AnimationsEnabled: s.enabledAnimations,
			SplashDurationSec: s.splashDuration,
			ShowStatusBar:     s.showStatusBar,
			ShowFooter:        s.showFooter,
		}

		if err := utils.SaveConfig(config); err != nil {
			return settingsSavedMsg{err: err}
		}

		return settingsSavedMsg{success: true}
	}
}

// settingsSavedMsg is sent when settings are saved.
type settingsSavedMsg struct {
	success bool
	err     error
}

// View implements tea.Model.
func (s *SettingsScreen) View() tea.View {
	var sections []string

	// Header
	header := s.theme.Title.Render("Settings")
	sections = append(sections, header)
	sections = append(sections, "")

	// Settings options
	settings := []struct {
		label string
		value string
	}{
		{"Theme", s.selectedTheme + " (← →: cycle)"},
		{"Animations", boolToYesNo(s.enabledAnimations)},
		{"Splash Duration", fmt.Sprintf("%d seconds (+/-: adjust)", s.splashDuration)},
		{"Status Bar", boolToYesNo(s.showStatusBar)},
		{"Footer", boolToYesNo(s.showFooter)},
	}

	for i, setting := range settings {
		cursor := "  "
		if s.cursor == i {
			cursor = "❯ "
		}

		labelStyle := lipgloss.NewStyle().Foreground(s.theme.Foreground)
		valueStyle := lipgloss.NewStyle().Foreground(s.theme.Accent).Bold(true)

		if s.cursor == i {
			labelStyle = labelStyle.Background(s.theme.Primary).Foreground(s.theme.Background)
			valueStyle = valueStyle.Background(s.theme.Primary).Foreground(s.theme.Background)
		}

		label := labelStyle.Render(fmt.Sprintf("%-20s", setting.label))
		value := valueStyle.Render(setting.value)

		line := cursor + label + " " + value
		sections = append(sections, line)
	}

	sections = append(sections, "")

	// Status message
	if s.statusMessage != "" {
		status := lipgloss.NewStyle().
			Foreground(s.theme.Success).
			Render(s.statusMessage)
		sections = append(sections, status)
		sections = append(sections, "")
	}

	// Help
	help := s.theme.Help.Render("↑/↓: Navigate • Enter/Space: Toggle • Ctrl+S: Save • Esc: Cancel")
	sections = append(sections, help)

	// Join sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add border
	bordered := s.theme.Border.Render(content)

	return tea.NewView(bordered)
}

// boolToYesNo converts bool to Yes/No string.
func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
