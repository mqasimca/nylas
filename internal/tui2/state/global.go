// Package state provides state management for the TUI.
package state

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

// GlobalState holds shared application state.
type GlobalState struct {
	// Nylas integration
	Client     ports.NylasClient
	GrantStore ports.GrantStore
	GrantID    string
	Email      string
	Provider   string

	// UI state
	Theme      string
	WindowSize tea.WindowSizeMsg

	// Status
	StatusMessage string
	StatusLevel   int

	// Feature flags
	OfflineMode bool

	// Rate limiting
	RateLimiter *utils.RateLimiter
}

// NewGlobalState creates a new global state.
func NewGlobalState(client ports.NylasClient, grantStore ports.GrantStore, grantID, email, provider string) *GlobalState {
	return &GlobalState{
		Client:      client,
		GrantStore:  grantStore,
		GrantID:     grantID,
		Email:       email,
		Provider:    provider,
		Theme:       "k9s",
		WindowSize:  tea.WindowSizeMsg{Width: 80, Height: 24},
		RateLimiter: utils.NewRateLimiter(500 * time.Millisecond), // 500ms between API calls
	}
}

// SetWindowSize updates the window size.
func (g *GlobalState) SetWindowSize(width, height int) {
	g.WindowSize = tea.WindowSizeMsg{Width: width, Height: height}
}

// SetStatus sets the status message.
func (g *GlobalState) SetStatus(message string, level int) {
	g.StatusMessage = message
	g.StatusLevel = level
}

// ClearStatus clears the status message.
func (g *GlobalState) ClearStatus() {
	g.StatusMessage = ""
	g.StatusLevel = 0
}
