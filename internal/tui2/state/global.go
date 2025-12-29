// Package state provides state management for the TUI.
package state

import (
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

// GlobalState holds shared application state.
type GlobalState struct {
	mu sync.RWMutex

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
	g.mu.Lock()
	defer g.mu.Unlock()
	g.WindowSize = tea.WindowSizeMsg{Width: width, Height: height}
}

// SetStatus sets the status message.
func (g *GlobalState) SetStatus(message string, level int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.StatusMessage = message
	g.StatusLevel = level
}

// ClearStatus clears the status message.
func (g *GlobalState) ClearStatus() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.StatusMessage = ""
	g.StatusLevel = 0
}

// GetStatus returns the current status message and level.
func (g *GlobalState) GetStatus() (string, int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.StatusMessage, g.StatusLevel
}

// GetWindowSize returns the current window size.
func (g *GlobalState) GetWindowSize() tea.WindowSizeMsg {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.WindowSize
}
