package models

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// SearchMode represents the current search mode.
type SearchMode int

const (
	// SearchModeOff means search is not active.
	SearchModeOff SearchMode = iota
	// SearchModeInput means user is typing a search query.
	SearchModeInput
	// SearchModeActive means search results are being displayed.
	SearchModeActive
	// SearchModeAdvanced means the advanced search dialog is open.
	SearchModeAdvanced
)

// BackMsg is sent to go back to the previous screen.
type BackMsg struct{}

// MessageList is the three-pane email list screen.
type MessageList struct {
	global *state.GlobalState
	theme  *styles.Theme

	layout       *components.ThreePaneLayout
	spinner      spinner.Model
	search       *components.Search
	searchDialog *components.SearchDialog

	threads          []domain.Thread
	allThreads       []domain.Thread // All threads before filtering (for client-side search)
	foldersLoaded    bool
	loadingFolders   bool
	selectedFolderID string // Currently selected folder for filtering

	// Search state
	searchMode  SearchMode
	searchQuery *components.SearchQuery

	loading bool
	err     error
}

// NewMessageList creates a new message list screen.
func NewMessageList(global *state.GlobalState) *MessageList {
	theme := styles.GetTheme(global.Theme)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Create three-pane layout
	layout := components.NewThreePaneLayout(theme)

	// Create search component
	search := components.NewSearch(theme)

	// Initialize layout with current window size if available
	if global.WindowSize.Width > 0 && global.WindowSize.Height > 0 {
		layoutHeight := global.WindowSize.Height - 5 // Reserve extra line for search
		if layoutHeight < 10 {
			layoutHeight = 10
		}
		layout.SetSize(global.WindowSize.Width, layoutHeight)
		search.SetWidth(global.WindowSize.Width)
	}

	return &MessageList{
		global:      global,
		theme:       theme,
		layout:      layout,
		spinner:     s,
		search:      search,
		searchMode:  SearchModeOff,
		searchQuery: &components.SearchQuery{},
		loading:     true,
	}
}

// Init implements tea.Model.
func (m *MessageList) Init() tea.Cmd {
	// Only fetch messages initially to avoid rate limiting
	// Folders can be fetched later if needed
	return tea.Batch(
		m.spinner.Tick,
		m.fetchMessages(),
	)
}

// Update implements tea.Model.
