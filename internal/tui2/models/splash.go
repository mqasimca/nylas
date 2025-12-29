// Package models provides screen models for the TUI.
package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
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
	percent   float64
	startTime time.Time
}

// NewSplash creates a new splash screen.
func NewSplash(global *state.GlobalState) *SplashScreen {
	theme := styles.GetTheme(global.Theme)

	// Create progress bar with gradient from primary to accent
	prog := progress.New(
		progress.WithColors(theme.Primary, theme.Accent),
		progress.WithWidth(60),
		progress.WithoutPercentage(),
	)

	return &SplashScreen{
		global:    global,
		theme:     theme,
		progress:  prog,
		percent:   0.0,
		startTime: time.Now(),
	}
}

// Init implements tea.Model.
func (s *SplashScreen) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		s.progress.Init(),
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
	}

	return s, nil
}

// View implements tea.Model.
func (s *SplashScreen) View() tea.View {
	var sections []string

	// Calculate panel width first
	panelWidth := 60
	if s.global.WindowSize.Width < 70 {
		panelWidth = s.global.WindowSize.Width - 10
	}

	// ✨ GLOSSY Logo with shimmer effect
	logoText := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Bold(true).
		Render(nylasLogo())

	logo := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, logoText)
	sections = append(sections, logo)
	sections = append(sections, "")

	// Metallic subtitle with premium styling
	subtitle := styles.MetallicText("Terminal Interface", s.theme)
	subtitleText := lipgloss.NewStyle().
		Foreground(s.theme.Secondary).
		Italic(true).
		Render(subtitle)

	subtitleStyled := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, subtitleText)
	sections = append(sections, subtitleStyled)

	// Tagline
	taglineText := lipgloss.NewStyle().
		Foreground(s.theme.Accent).
		Render("Email & Calendar • Powered by Bubble Tea")

	tagline := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, taglineText)
	sections = append(sections, tagline)
	sections = append(sections, "")
	sections = append(sections, "")

	// Accent line separator
	separatorLine := styles.AccentLine(s.theme, 50, "━")
	separator := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, separatorLine)
	sections = append(sections, separator)
	sections = append(sections, "")

	// Progress bar with glow effect
	progressView := s.progress.View()

	// Add glow around progress bar
	progressGlowText := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Faint(true).
		Render(strings.Repeat("▔", 40))
	progressGlow := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, progressGlowText)

	progressCentered := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, progressView)

	progressBottomText := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Faint(true).
		Render(strings.Repeat("▁", 40))
	progressBottom := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, progressBottomText)

	sections = append(sections, progressGlow)
	sections = append(sections, progressCentered)
	sections = append(sections, progressBottom)
	sections = append(sections, "")

	// Loading text with animated dots and percentage
	dots := strings.Repeat(".", (int(s.percent*3)%3)+1)
	percentage := int(s.percent * 100)

	loadingText := lipgloss.NewStyle().
		Foreground(s.theme.Info).
		Bold(true).
		Render(fmt.Sprintf("Loading%s %d%%", dots, percentage))

	loading := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, loadingText)
	sections = append(sections, loading)
	sections = append(sections, "")

	// Premium hint with animation
	hintTextStr := "⚡ Press any key to skip ⚡"
	if int(s.percent*10)%2 == 0 {
		hintTextStr = "✨ Press any key to skip ✨"
	}

	hintText := lipgloss.NewStyle().
		Foreground(s.theme.Accent).
		Italic(true).
		Render(hintTextStr)

	hint := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, hintText)
	sections = append(sections, hint)
	sections = append(sections, "")

	// Version/credits
	creditsText := lipgloss.NewStyle().
		Foreground(s.theme.Dimmed.GetForeground()).
		Faint(true).
		Render("◆ Nylas CLI v2.0 ◆")

	credits := lipgloss.PlaceHorizontal(panelWidth-8, lipgloss.Center, creditsText)
	sections = append(sections, credits)

	// Join all sections (left-aligned within the panel)
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Premium border
	bordered := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(s.theme.Primary).
		BorderBackground(lipgloss.Color("#0a0a0a")).
		Background(lipgloss.Color("#0f0f0f")).
		Padding(3, 4).
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
