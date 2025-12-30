// Package components provides reusable UI components.
package components

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// StatusBar displays app status and information.
type StatusBar struct {
	width       int
	theme       *styles.Theme
	appName     string
	version     string
	email       string
	online      bool
	unreadCount int
	eventCount  int
	currentTime time.Time
}

// NewStatusBar creates a new status bar.
func NewStatusBar(theme *styles.Theme, email string) *StatusBar {
	return &StatusBar{
		theme:       theme,
		appName:     "Nylas CLI",
		version:     "v2.0",
		email:       email,
		online:      true,
		unreadCount: 0,
		eventCount:  0,
		currentTime: time.Now(),
	}
}

// SetWidth sets the width of the status bar.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetOnline sets the online status.
func (s *StatusBar) SetOnline(online bool) {
	s.online = online
}

// SetUnreadCount sets the unread message count.
func (s *StatusBar) SetUnreadCount(count int) {
	s.unreadCount = count
}

// SetEventCount sets the today's event count.
func (s *StatusBar) SetEventCount(count int) {
	s.eventCount = count
}

// Update updates the current time.
func (s *StatusBar) Update() {
	s.currentTime = time.Now()
}

// View renders the status bar with clean k9s-style design.
func (s *StatusBar) View() string {
	if s.width == 0 {
		return ""
	}

	sep := lipgloss.NewStyle().Foreground(s.theme.Dimmed.GetForeground()).Render(" │ ")

	// Right section (k9s style): email | provider | grant_id | time | status
	emailStyle := lipgloss.NewStyle().Foreground(s.theme.Foreground)
	email := emailStyle.Render(s.email)

	providerBadge := lipgloss.NewStyle().
		Foreground(s.theme.Background).
		Background(s.theme.Secondary).
		Padding(0, 1).
		Render("microsoft") // TODO: get from global state

	grantStyle := lipgloss.NewStyle().Foreground(s.theme.Dimmed.GetForeground())
	grantID := grantStyle.Render("46c724e7-1eaa-4c75-810c-8f8a7ead6009") // TODO: get from global state

	timeStyle := lipgloss.NewStyle().Foreground(s.theme.Foreground)
	timeStr := timeStyle.Render(s.currentTime.Format("15:04:05"))

	// Refresh indicator
	refreshStyle := lipgloss.NewStyle().Foreground(s.theme.Dimmed.GetForeground())
	refresh := refreshStyle.Render("<1s>")

	// Live status
	statusDot := "●"
	statusColor := s.theme.Success
	statusText := "Live"
	if !s.online {
		statusColor = s.theme.Error
		statusText = "Offline"
	}
	status := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(fmt.Sprintf("%s %s", statusDot, statusText))

	right := lipgloss.JoinHorizontal(lipgloss.Left,
		email, sep, providerBadge, sep, grantID, "   ", timeStr, " ", refresh, " ", status,
	)

	// Calculate spacing
	rightWidth := lipgloss.Width(right)

	if rightWidth >= s.width {
		// Minimal version
		return lipgloss.NewStyle().
			Width(s.width).
			MaxWidth(s.width).
			Render(right)
	}

	// Right-align the content
	spacer := lipgloss.NewStyle().Width(s.width - rightWidth).Render("")
	content := lipgloss.JoinHorizontal(lipgloss.Left, spacer, right)

	return lipgloss.NewStyle().
		Width(s.width).
		MaxWidth(s.width).
		Render(content)
}

// FooterBar displays keybindings and user info.
type FooterBar struct {
	width    int
	theme    *styles.Theme
	email    string
	bindings []KeyBinding
}

// KeyBinding represents a keyboard shortcut.
type KeyBinding struct {
	Key         string
	Description string
}

// NewFooterBar creates a new footer bar.
func NewFooterBar(theme *styles.Theme, email string) *FooterBar {
	return &FooterBar{
		theme:    theme,
		email:    email,
		bindings: []KeyBinding{},
	}
}

// SetWidth sets the width of the footer bar.
func (f *FooterBar) SetWidth(width int) {
	f.width = width
}

// SetBindings sets the keyboard bindings to display.
func (f *FooterBar) SetBindings(bindings []KeyBinding) {
	f.bindings = bindings
}

// View renders the footer bar.
func (f *FooterBar) View() string {
	if f.width == 0 {
		return ""
	}

	// Simple k9s-style footer: :command | ?:help | ^c:quit
	sep := lipgloss.NewStyle().Foreground(f.theme.Dimmed.GetForeground()).Render(" │ ")

	cmdStyle := lipgloss.NewStyle().Foreground(f.theme.Warning)
	descStyle := lipgloss.NewStyle().Foreground(f.theme.Foreground)

	items := []string{
		cmdStyle.Render(":") + descStyle.Render("command"),
		cmdStyle.Render("?:") + descStyle.Render("help"),
		cmdStyle.Render("^c:") + descStyle.Render("quit"),
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left, items[0], sep, items[1], sep, items[2])

	return lipgloss.NewStyle().
		Width(f.width).
		MaxWidth(f.width).
		Render(content)
}
