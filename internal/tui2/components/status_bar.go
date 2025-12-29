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

// View renders the status bar with glossy styling.
func (s *StatusBar) View() string {
	if s.width == 0 {
		return ""
	}

	// ✨ Left section: App info with metallic effect
	appInfo := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Bold(true).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Render(fmt.Sprintf("✨ %s %s", s.appName, s.version))

	// Middle section: Stats with badge styling
	inboxBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(s.theme.Info).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("%d", s.unreadCount))

	inbox := lipgloss.NewStyle().
		Foreground(s.theme.Foreground).
		Render(fmt.Sprintf("📥 %s", inboxBadge))

	eventsBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(s.theme.Warning).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("%d", s.eventCount))

	events := lipgloss.NewStyle().
		Foreground(s.theme.Foreground).
		Render(fmt.Sprintf("📅 %s", eventsBadge))

	// Right section: Status with glow and time
	statusDot := "●"
	statusColor := s.theme.Success
	statusText := "ONLINE"
	if !s.online {
		statusDot = "●"
		statusColor = s.theme.Error
		statusText = "OFFLINE"
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Render(fmt.Sprintf("%s %s", statusDot, statusText))

	clock := lipgloss.NewStyle().
		Foreground(s.theme.Accent).
		Background(lipgloss.Color("#1a1a1a")).
		Bold(true).
		Padding(0, 1).
		Render(s.currentTime.Format("⏰ 15:04:05"))

	// Glossy separator
	sep := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Render(" ◆ ")

	// Combine sections
	left := lipgloss.JoinHorizontal(lipgloss.Left, appInfo)
	middle := lipgloss.JoinHorizontal(lipgloss.Left, inbox, sep, events)
	right := lipgloss.JoinHorizontal(lipgloss.Left, statusStyle, sep, clock)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	middleWidth := lipgloss.Width(middle)
	rightWidth := lipgloss.Width(right)

	totalContent := leftWidth + middleWidth + rightWidth
	if totalContent >= s.width {
		// Not enough space, show minimal version
		return lipgloss.NewStyle().
			Width(s.width).
			Background(lipgloss.Color("#0a0a0a")).
			Foreground(s.theme.Foreground).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(s.theme.Primary).
			Render(appInfo)
	}

	// Calculate gaps
	remainingSpace := s.width - totalContent
	gap1 := remainingSpace / 2
	gap2 := remainingSpace - gap1

	spacer1 := lipgloss.NewStyle().Width(gap1).Render("")
	spacer2 := lipgloss.NewStyle().Width(gap2).Render("")

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		spacer1,
		middle,
		spacer2,
		right,
	)

	// Premium background with gradient effect and border
	return lipgloss.NewStyle().
		Width(s.width).
		Background(lipgloss.Color("#0a0a0a")).
		Foreground(s.theme.Foreground).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(s.theme.Primary).
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

	// Build bindings display with glossy badges
	var bindingStrs []string
	for i, b := range f.bindings {
		// Alternate colors for visual variety
		badgeColor := f.theme.Primary
		if i%2 == 1 {
			badgeColor = f.theme.Accent
		}

		// Glossy key badge
		key := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(badgeColor).
			Bold(true).
			Padding(0, 1).
			Render(b.Key)

		// Description
		desc := lipgloss.NewStyle().
			Foreground(f.theme.Foreground).
			Render(b.Description)

		// Separator
		sep := lipgloss.NewStyle().
			Foreground(f.theme.Dimmed.GetForeground()).
			Render(" • ")

		bindingStrs = append(bindingStrs, fmt.Sprintf("%s %s%s", key, desc, sep))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left, bindingStrs...)

	// User email with premium badge on the right
	emailBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(f.theme.Secondary).
		Bold(true).
		Padding(0, 1).
		Render("👤")

	emailText := lipgloss.NewStyle().
		Foreground(f.theme.Foreground).
		Render(fmt.Sprintf(" %s", f.email))

	right := lipgloss.JoinHorizontal(lipgloss.Left, emailBadge, emailText)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	if leftWidth+rightWidth >= f.width {
		// Not enough space - show minimal version
		return lipgloss.NewStyle().
			Width(f.width).
			Background(lipgloss.Color("#0a0a0a")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(f.theme.Primary).
			Render(left)
	}

	gap := f.width - leftWidth - rightWidth
	spacer := lipgloss.NewStyle().Width(gap).Render("")

	content := lipgloss.JoinHorizontal(lipgloss.Left, left, spacer, right)

	// Premium footer with top border and dark background
	return lipgloss.NewStyle().
		Width(f.width).
		Background(lipgloss.Color("#0a0a0a")).
		Foreground(f.theme.Foreground).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(f.theme.Primary).
		Render(content)
}
