// Package tui2 provides a Bubble Tea-based terminal user interface for Nylas.
package tui2

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/mqasimca/nylas/internal/tui2/models"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

// Config holds the configuration for the TUI application.
type Config struct {
	Client     ports.NylasClient
	GrantStore ports.GrantStore
	GrantID    string
	Email      string
	Provider   string
	Theme      string
}

// App is the root application model using Model Stack pattern.
type App struct {
	stack  []tea.Model        // Screen stack
	global *state.GlobalState // Shared state
}

// NewApp creates a new TUI application.
func NewApp(cfg Config) *App {
	// Create global state
	global := state.NewGlobalState(
		cfg.Client,
		cfg.GrantStore,
		cfg.GrantID,
		cfg.Email,
		cfg.Provider,
	)

	if cfg.Theme != "" {
		global.Theme = cfg.Theme
	}

	app := &App{
		global: global,
	}

	// Push initial screen (Dashboard)
	dashboard := models.NewDashboard(global)
	app.stack = []tea.Model{dashboard}

	return app
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	if len(a.stack) > 0 {
		return a.stack[len(a.stack)-1].Init()
	}
	return nil
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keyboard shortcuts
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		// Update global window size
		a.global.SetWindowSize(msg.Width, msg.Height)

	case models.NavigateMsg:
		// Navigate to new screen
		return a, a.navigate(msg)

	case models.BackMsg:
		// Go back to previous screen
		return a, a.back()

	case QuitMsg:
		// Quit application
		return a, tea.Quit

	case StatusMsg:
		// Update status
		a.global.SetStatus(msg.Message, int(msg.Level))
	}

	// Forward message to current screen
	if len(a.stack) > 0 {
		current := a.stack[len(a.stack)-1]
		updated, cmd := current.Update(msg)
		a.stack[len(a.stack)-1] = updated
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model.
func (a *App) View() string {
	if len(a.stack) > 0 {
		return a.stack[len(a.stack)-1].View()
	}
	return "Loading..."
}

// navigate pushes a new screen onto the stack.
func (a *App) navigate(msg models.NavigateMsg) tea.Cmd {
	var screen tea.Model

	switch msg.Screen {
	case models.ScreenDashboard:
		screen = models.NewDashboard(a.global)
	case models.ScreenMessages:
		screen = models.NewMessageList(a.global)
	case models.ScreenMessageDetail:
		// Extract message ID from Data
		messageID, ok := msg.Data.(string)
		if !ok {
			a.global.SetStatus("Invalid message ID", int(StatusError))
			return nil
		}
		screen = models.NewMessageDetail(a.global, messageID)
	case models.ScreenCalendar:
		// TODO: Implement calendar screen
		a.global.SetStatus("Calendar screen not yet implemented", int(StatusWarning))
		return nil
	case models.ScreenContacts:
		// TODO: Implement contacts screen
		a.global.SetStatus("Contacts screen not yet implemented", int(StatusWarning))
		return nil
	case models.ScreenSettings:
		// TODO: Implement settings screen
		a.global.SetStatus("Settings screen not yet implemented", int(StatusWarning))
		return nil
	case models.ScreenHelp:
		// TODO: Implement help screen
		a.global.SetStatus("Help screen not yet implemented", int(StatusWarning))
		return nil
	default:
		return nil
	}

	// Push screen onto stack
	a.stack = append(a.stack, screen)

	// Initialize the new screen
	return screen.Init()
}

// back pops the current screen from the stack.
func (a *App) back() tea.Cmd {
	if len(a.stack) > 1 {
		a.stack = a.stack[:len(a.stack)-1]
	}
	return nil
}

// Run starts the TUI application.
func Run(cfg Config) error {
	app := NewApp(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
