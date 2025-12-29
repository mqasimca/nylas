// Package components provides reusable UI components.
package components

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// LoadingOverlay shows a loading spinner with a message.
type LoadingOverlay struct {
	spinner spinner.Model
	message string
	theme   *styles.Theme
	width   int
	height  int
}

// NewLoadingOverlay creates a new loading overlay.
func NewLoadingOverlay(theme *styles.Theme, message string) *LoadingOverlay {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &LoadingOverlay{
		spinner: s,
		message: message,
		theme:   theme,
	}
}

// SetSize sets the overlay size.
func (l *LoadingOverlay) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// SetMessage sets the loading message.
func (l *LoadingOverlay) SetMessage(message string) {
	l.message = message
}

// Init implements tea.Model.
func (l *LoadingOverlay) Init() tea.Cmd {
	return l.spinner.Tick
}

// Update implements tea.Model.
func (l *LoadingOverlay) Update(msg tea.Msg) (*LoadingOverlay, tea.Cmd) {
	var cmd tea.Cmd
	l.spinner, cmd = l.spinner.Update(msg)
	return l, cmd
}

// View renders the loading overlay.
func (l *LoadingOverlay) View() string {
	if l.width == 0 || l.height == 0 {
		return ""
	}

	// Create loading content
	spinnerView := l.spinner.View()
	messageStyle := lipgloss.NewStyle().
		Foreground(l.theme.Foreground).
		MarginTop(1)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		spinnerView,
		messageStyle.Render(l.message),
	)

	// Center in screen
	centered := lipgloss.Place(
		l.width,
		l.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	// Semi-transparent background effect
	overlay := lipgloss.NewStyle().
		Background(l.theme.Background).
		Render(centered)

	return overlay
}
