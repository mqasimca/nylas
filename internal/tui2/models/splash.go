// Package models provides screen models for the TUI.
package models

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

const splashDuration = 3 * time.Second

// tickMsg is sent every 50ms to update the progress bar.
type tickMsg time.Time

// SplashScreen is the initial splash screen with animated progress.
type SplashScreen struct {
	global    *state.GlobalState
	theme     *styles.Theme
	progress  progress.Model
	spinner   spinner.Model
	percent   float64
	startTime time.Time
}

// NewSplash creates a new splash screen.
func NewSplash(global *state.GlobalState) *SplashScreen {
	theme := styles.GetTheme(global.Theme)

	// Create clean progress bar with gradient
	prog := progress.New(
		progress.WithColors(theme.Primary, theme.Accent),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Create spinner for activity indication
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &SplashScreen{
		global:    global,
		theme:     theme,
		progress:  prog,
		spinner:   sp,
		percent:   0.0,
		startTime: time.Now(),
	}
}

// Init implements tea.Model.
func (s *SplashScreen) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		s.progress.Init(),
		s.spinner.Tick,
	)
}

// Update implements tea.Model.
func (s *SplashScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any key press skips the splash screen
		return s, func() tea.Msg {
			return NavigateMsg{Screen: ScreenDashboard}
		}

	case tea.WindowSizeMsg:
		// Update global window size
		s.global.SetWindowSize(msg.Width, msg.Height)

	case tickMsg:
		// Calculate progress based on elapsed time
		elapsed := time.Since(s.startTime)
		s.percent = min(elapsed.Seconds()/splashDuration.Seconds(), 1.0)

		if s.percent >= 1.0 {
			// Transition to dashboard
			return s, func() tea.Msg {
				return NavigateMsg{Screen: ScreenDashboard}
			}
		}

		// Smooth animation using SetPercent
		cmd := s.progress.SetPercent(s.percent)
		return s, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:
		updatedProgress, cmd := s.progress.Update(msg)
		s.progress = updatedProgress
		return s, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}

// View implements tea.Model.
func (s *SplashScreen) View() tea.View {
	// Calculate panel width
	panelWidth := 50
	if s.global.WindowSize.Width < 60 {
		panelWidth = s.global.WindowSize.Width - 10
	}
	contentWidth := panelWidth - 6

	var sections []string

	// Clean logo
	logoText := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Bold(true).
		Render(nylasLogo())
	logo := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, logoText)
	sections = append(sections, logo)

	// Subtitle - clean and simple
	subtitleText := lipgloss.NewStyle().
		Foreground(s.theme.Secondary).
		Render("Terminal Interface")
	subtitle := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, subtitleText)
	sections = append(sections, subtitle)
	sections = append(sections, "")

	// Simple separator
	sepStyle := lipgloss.NewStyle().Foreground(s.theme.Dimmed.GetForeground())
	separator := sepStyle.Render(styles.RepeatChar('─', contentWidth))
	sections = append(sections, separator)
	sections = append(sections, "")

	// Progress bar - clean, no glow effects
	progressView := s.progress.View()
	progressCentered := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, progressView)
	sections = append(sections, progressCentered)
	sections = append(sections, "")

	// Loading status with spinner and percentage
	percentage := int(s.percent * 100)
	spinnerView := s.spinner.View()

	statusStyle := lipgloss.NewStyle().Foreground(s.theme.Foreground)
	loadingText := statusStyle.Render(fmt.Sprintf("Initializing... %d%%", percentage))

	// Combine spinner and loading text
	loadingLine := lipgloss.JoinHorizontal(lipgloss.Left, spinnerView, " ", loadingText)
	loadingCentered := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, loadingLine)
	sections = append(sections, loadingCentered)
	sections = append(sections, "")

	// Skip hint - subtle
	hintStyle := lipgloss.NewStyle().
		Foreground(s.theme.Dimmed.GetForeground()).
		Italic(true)
	hintText := hintStyle.Render("Press any key to skip")
	hint := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, hintText)
	sections = append(sections, hint)

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Clean rounded border
	bordered := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Primary).
		Background(lipgloss.Color("#0d0d0d")).
		Padding(2, 3).
		Width(panelWidth).
		Render(content)

	// Center in terminal
	centered := lipgloss.Place(
		s.global.WindowSize.Width,
		s.global.WindowSize.Height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)

	return tea.NewView(centered)
}

// tickCmd returns a command that sends a tick message every 50ms.
func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// nylasLogo returns ASCII art for the Nylas logo.
func nylasLogo() string {
	return `
███╗   ██╗██╗   ██╗██╗      █████╗ ███████╗
████╗  ██║╚██╗ ██╔╝██║     ██╔══██╗██╔════╝
██╔██╗ ██║ ╚████╔╝ ██║     ███████║███████╗
██║╚██╗██║  ╚██╔╝  ██║     ██╔══██║╚════██║
██║ ╚████║   ██║   ███████╗██║  ██║███████║
╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚═╝  ╚═╝╚══════╝
`
}
