package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
	"github.com/mqasimca/nylas/internal/tui2/utils"
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
func (m *MessageList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Handle search dialog messages
	case components.SearchDialogSubmitMsg:
		m.searchMode = SearchModeActive
		m.searchDialog = nil
		m.search.SetValue(msg.Query)
		m.searchQuery = components.ParseSearchQuery(msg.Query)

		// Store current threads for filtering
		if len(m.allThreads) == 0 {
			m.allThreads = m.threads
		}

		// If query has API-searchable operators, do API search
		if m.shouldUseAPISearch(m.searchQuery) {
			m.loading = true
			m.global.SetStatus("Searching...", 0)
			return m, tea.Batch(m.spinner.Tick, m.searchMessagesAPI(m.searchQuery))
		}

		// Otherwise, apply client-side filter
		m.applyClientFilter(m.searchQuery)
		m.global.SetStatus(fmt.Sprintf("Found %d results", len(m.threads)), 0)
		return m, nil

	case components.SearchDialogCancelMsg:
		m.searchMode = SearchModeOff
		m.searchDialog = nil
		return m, nil

	case tea.KeyMsg:
		// Use msg.String() for all key matching (v2 pattern)
		key := msg.Key()
		keyStr := msg.String()

		// When in advanced search dialog mode, delegate to dialog
		if m.searchMode == SearchModeAdvanced && m.searchDialog != nil {
			var cmd tea.Cmd
			m.searchDialog, cmd = m.searchDialog.Update(msg)
			return m, cmd
		}

		// When in search input mode, handle search-specific keys first
		if m.searchMode == SearchModeInput {
			switch key.Code {
			case tea.KeyEsc:
				// Cancel search and return to normal mode
				m.searchMode = SearchModeOff
				m.search.Reset()
				m.search.Blur()
				// Restore original threads
				if len(m.allThreads) > 0 {
					m.threads = m.allThreads
					m.updateThreadTable()
				}
				m.global.SetStatus("Search cancelled", 0)
				return m, nil

			case tea.KeyEnter:
				// Submit search
				m.searchMode = SearchModeActive
				m.search.Blur()
				m.searchQuery = m.search.Query()

				// If query has API-searchable operators, do API search
				if m.shouldUseAPISearch(m.searchQuery) {
					m.loading = true
					m.global.SetStatus("Searching...", 0)
					return m, tea.Batch(m.spinner.Tick, m.searchMessagesAPI(m.searchQuery))
				}

				// Otherwise, apply client-side filter
				m.applyClientFilter(m.searchQuery)
				m.global.SetStatus(fmt.Sprintf("Found %d results", len(m.threads)), 0)
				return m, nil

			default:
				// Pass to search input
				m.search, cmd = m.search.Update(msg)
				// Real-time client-side filtering as user types
				m.searchQuery = m.search.Query()
				if !m.searchQuery.IsEmpty() {
					m.applyClientFilter(m.searchQuery)
				} else if len(m.allThreads) > 0 {
					m.threads = m.allThreads
					m.updateThreadTable()
				}
				return m, cmd
			}
		}

		// Handle Esc key
		if key.Code == tea.KeyEsc {
			// If search is active, clear it first
			if m.searchMode == SearchModeActive {
				m.searchMode = SearchModeOff
				m.search.Reset()
				if len(m.allThreads) > 0 {
					m.threads = m.allThreads
					m.updateThreadTable()
				}
				m.global.SetStatus("Search cleared", 0)
				return m, nil
			}
			// Go back to dashboard
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Handle ctrl+c (handled by app.go)
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle ctrl+r (refresh)
		if keyStr == "ctrl+r" {
			// Clear search when refreshing
			m.searchMode = SearchModeOff
			m.search.Reset()
			m.allThreads = nil

			// Refresh messages (respecting current folder filter)
			m.loading = true
			if m.selectedFolderID != "" {
				return m, tea.Batch(m.spinner.Tick, m.fetchMessagesForFolder(m.selectedFolderID))
			}
			return m, tea.Batch(m.spinner.Tick, m.fetchMessages())
		}

		// Handle '/' to activate search
		if keyStr == "/" {
			m.searchMode = SearchModeInput
			// Store current threads for filtering
			if len(m.allThreads) == 0 {
				m.allThreads = m.threads
			}
			return m, m.search.Focus()
		}

		// Handle '?' to open advanced search dialog
		if keyStr == "?" {
			m.searchMode = SearchModeAdvanced
			m.searchDialog = components.NewSearchDialog(m.theme)
			m.searchDialog.SetSize(m.global.WindowSize.Width, m.global.WindowSize.Height)
			// Store current threads for filtering
			if len(m.allThreads) == 0 {
				m.allThreads = m.threads
			}
			// If there's an existing search query, populate the dialog
			if m.searchQuery != nil && !m.searchQuery.IsEmpty() {
				m.searchDialog.SetQuery(m.search.Value())
			}
			return m, m.searchDialog.Init()
		}

		// Handle tab/shift+tab using keyStr
		if keyStr == "shift+tab" {
			// Focus previous pane
			m.layout.FocusPrevious()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil
		}

		// Handle tab
		if keyStr == "tab" {
			// Focus next pane
			m.layout.FocusNext()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil
		}

		// Handle Enter key
		if key.Code == tea.KeyEnter {
			// Handle enter based on focused pane
			switch m.layout.GetFocused() {
			case components.FolderPane:
				// Folder selected - reload messages for this folder
				selectedItem := m.layout.SelectedFolder()
				if folderItem, ok := selectedItem.(components.FolderItem); ok {
					// Skip special "show all" item
					if folderItem.Folder.ID == "_show_all" {
						// TODO: Load all folders
						return m, nil
					}

					// Set selected folder and reload messages
					m.selectedFolderID = folderItem.Folder.ID
					m.loading = true
					// Clear search when changing folders
					m.searchMode = SearchModeOff
					m.search.Reset()
					m.allThreads = nil
					m.global.SetStatus(fmt.Sprintf("Loading messages from %s...", folderItem.Folder.Name), 0)
					return m, tea.Batch(m.spinner.Tick, m.fetchMessagesForFolder(folderItem.Folder.ID))
				}
				return m, nil

			case components.MessagePane:
				// Thread selected - navigate to thread detail view
				idx := m.layout.SelectedMessageIndex()
				if idx >= 0 && idx < len(m.threads) {
					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: ScreenMessageDetail,
							Data:   m.threads[idx].ID, // Pass thread ID
						}
					}
				}
				return m, nil
			}
		}

		// Handle text keys using keyStr
		switch keyStr {
		case "h":
			// Focus previous pane
			m.layout.FocusPrevious()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil

		case "l":
			// Focus next pane
			m.layout.FocusNext()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil

		case "c":
			// Compose new message
			return m, func() tea.Msg {
				return NavigateMsg{
					Screen: ScreenCompose,
					Data:   ComposeData{Mode: ComposeModeNew},
				}
			}

		case "r":
			// Reply to latest message in selected thread (only when message pane is focused)
			if m.layout.GetFocused() == components.MessagePane {
				idx := m.layout.SelectedMessageIndex()
				if idx >= 0 && idx < len(m.threads) {
					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: ScreenCompose,
							Data: ComposeData{
								Mode:    ComposeModeReply,
								Message: &m.threads[idx].LatestDraftOrMessage,
							},
						}
					}
				}
			}
			return m, nil

		case "a":
			// Reply all to latest message in selected thread (only when message pane is focused)
			if m.layout.GetFocused() == components.MessagePane {
				idx := m.layout.SelectedMessageIndex()
				if idx >= 0 && idx < len(m.threads) {
					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: ScreenCompose,
							Data: ComposeData{
								Mode:    ComposeModeReplyAll,
								Message: &m.threads[idx].LatestDraftOrMessage,
							},
						}
					}
				}
			}
			return m, nil

		case "f":
			// Forward latest message in selected thread (only when message pane is focused)
			if m.layout.GetFocused() == components.MessagePane {
				idx := m.layout.SelectedMessageIndex()
				if idx >= 0 && idx < len(m.threads) {
					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: ScreenCompose,
							Data: ComposeData{
								Mode:    ComposeModeForward,
								Message: &m.threads[idx].LatestDraftOrMessage,
							},
						}
					}
				}
			}
			return m, nil

		}

	case tea.WindowSizeMsg:
		// Update layout size
		m.global.SetWindowSize(msg.Width, msg.Height)
		// Reserve space for header (2 lines), search (1 line if active), and footer (2 lines)
		reservedLines := 4
		if m.searchMode == SearchModeInput || m.searchMode == SearchModeActive {
			reservedLines = 5 // Extra line for search bar
		}
		layoutHeight := msg.Height - reservedLines
		if layoutHeight < 10 {
			layoutHeight = 10 // Minimum height
		}
		m.layout.SetSize(msg.Width, layoutHeight)
		m.search.SetWidth(msg.Width)
		// Resize search dialog if open
		if m.searchDialog != nil {
			m.searchDialog.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case threadsLoadedMsg:
		m.threads = msg.threads
		m.loading = false
		m.updateThreadTable()
		return m, nil

	case foldersLoadedMsg:
		m.foldersLoaded = true
		m.loadingFolders = false
		m.layout.SetFolders(msg.folders)
		m.global.SetStatus(fmt.Sprintf("Loaded %d folders", len(msg.folders)), 0)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		m.loadingFolders = false
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update layout (delegates to focused pane)
	cmd = m.layout.Update(msg)
	cmds = append(cmds, cmd)

	// Update preview when cursor moves in message pane
	if m.layout.GetFocused() == components.MessagePane && !m.loading && len(m.threads) > 0 {
		idx := m.layout.SelectedMessageIndex()
		if idx >= 0 && idx < len(m.threads) {
			m.showThreadPreview(m.threads[idx].ID)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m *MessageList) View() tea.View {
	// Show advanced search dialog if open
	if m.searchMode == SearchModeAdvanced && m.searchDialog != nil {
		return tea.NewView(m.searchDialog.View())
	}

	if m.err != nil {
		return tea.NewView(m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to go back", m.err)))
	}

	// Build header
	header := m.theme.Title.Render("Messages") + " " +
		m.theme.Subtitle.Render(fmt.Sprintf("(%s)", m.global.Email))

	// Show loading spinner if loading
	if m.loading {
		header += " " + m.spinner.View()
	}

	// Show search status if active
	if m.searchMode == SearchModeActive && !m.searchQuery.IsEmpty() {
		searchInfo := fmt.Sprintf(" [Searching: %s - %d results]", m.search.Value(), len(m.threads))
		searchStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary).Italic(true)
		header += searchStyle.Render(searchInfo)
	}

	// Build search bar (only shown when in search mode)
	var searchBar string
	if m.searchMode != SearchModeOff {
		searchBar = m.search.ViewInline() + "\n"
	}

	// Build help text
	var help string
	switch m.searchMode {
	case SearchModeInput:
		help = m.theme.Help.Render("Enter: search  Esc: cancel  ?: advanced  | from: to: subject: is:unread has:attachment")
	case SearchModeActive:
		help = m.theme.Help.Render("/: search  ?: advanced  Esc: clear  c: compose  Tab: switch pane  Ctrl+R: refresh")
	default:
		help = m.theme.Help.Render("/: search  ?: advanced  c: compose  r: reply  a: reply all  f: forward  Tab: pane  Ctrl+R: refresh  esc: back")
	}

	// Build layout
	layoutView := m.layout.View()

	// Join all sections with single newlines to maximize space
	return tea.NewView(header + "\n" + searchBar + layoutView + "\n" + help)
}

// fetchMessages fetches the message list.
func (m *MessageList) fetchMessages() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch threads instead of individual messages for proper conversation grouping
		params := &domain.ThreadQueryParams{
			Limit: 50,
		}
		threads, err := m.global.Client.GetThreads(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return threadsLoadedMsg{threads}
	}
}

// fetchMessagesForFolder fetches threads filtered by folder ID.
func (m *MessageList) fetchMessagesForFolder(folderID string) tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch threads filtered by folder
		params := &domain.ThreadQueryParams{
			In:    []string{folderID},
			Limit: 50,
		}

		threads, err := m.global.Client.GetThreads(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return threadsLoadedMsg{threads}
	}
}

// fetchFolders fetches the folder list (lazy loaded).
func (m *MessageList) fetchFolders() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		folders, err := m.global.Client.GetFolders(ctx, m.global.GrantID)
		if err != nil {
			return errMsg{err}
		}

		// Filter to show only important folders
		importantFolders := utils.FilterImportantFolders(folders)

		// Sort by importance
		sortedFolders := utils.SortFoldersByImportance(importantFolders)

		// Convert to list items with icons
		items := make([]list.Item, len(sortedFolders))
		for i, folder := range sortedFolders {
			items[i] = components.FolderItem{Folder: folder}
		}

		// Add "Show all folders..." option at the end if we filtered some
		if len(folders) > len(sortedFolders) {
			items = append(items, components.FolderItem{
				Folder: domain.Folder{
					Name:       fmt.Sprintf("📋 Show all %d folders...", len(folders)),
					ID:         "_show_all",
					TotalCount: len(folders) - len(sortedFolders),
				},
			})
		}

		return foldersLoadedMsg{folders: items}
	}
}

// updateMessageTable updates the message table in the layout.
func (m *MessageList) updateThreadTable() {
	// Create highlight style for search matches
	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FFFF00")).
		Bold(true)

	rows := make([]table.Row, len(m.threads))
	for i, thread := range m.threads {
		// Get participants (from the latest message)
		from := "Unknown"
		if len(thread.Participants) > 0 {
			if thread.Participants[0].Name != "" {
				from = thread.Participants[0].Name
			} else {
				from = thread.Participants[0].Email
			}
		}

		subject := thread.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		// Add message count if more than 1 message in thread
		msgCount := len(thread.MessageIDs)
		if msgCount > 1 {
			subject = fmt.Sprintf("%s (%d)", subject, msgCount)
		}

		// Apply search highlighting when search is active
		if m.searchMode == SearchModeActive && m.searchQuery != nil && !m.searchQuery.IsEmpty() {
			from = m.searchQuery.HighlightMatches(from, "from", highlightStyle)
			subject = m.searchQuery.HighlightMatches(subject, "subject", highlightStyle)
		}

		// Format date using latest message received date
		date := formatDate(thread.LatestMessageRecvDate)

		rows[i] = table.Row{
			truncate(from, 20),
			truncate(subject, 40),
			date,
		}
	}

	m.layout.SetMessages(rows)

	// Auto-preview the first thread
	if len(m.threads) > 0 {
		m.showThreadPreview(m.threads[0].ID)
	}
}

// showThreadPreview displays a thread preview in the preview pane.
func (m *MessageList) showThreadPreview(threadID string) {
	// Find thread by ID
	var thread *domain.Thread
	for i := range m.threads {
		if m.threads[i].ID == threadID {
			thread = &m.threads[i]
			break
		}
	}

	if thread == nil {
		m.layout.SetPreview("Thread not found")
		return
	}

	// Build preview content
	var preview strings.Builder

	// Header with message count
	msgCount := len(thread.MessageIDs)
	header := thread.Subject
	if msgCount > 1 {
		header = fmt.Sprintf("%s (%d messages)", header, msgCount)
	}
	preview.WriteString(m.theme.Title.Render("Subject: ") + header + "\n\n")

	// Participants (thread-level field, always populated)
	if len(thread.Participants) > 0 {
		participants := make([]string, 0, len(thread.Participants))
		for _, p := range thread.Participants {
			if p.Name != "" {
				participants = append(participants, fmt.Sprintf("%s <%s>", p.Name, p.Email))
			} else {
				participants = append(participants, p.Email)
			}
		}
		preview.WriteString(m.theme.KeyBinding.Render("Participants: ") + strings.Join(participants, ", ") + "\n")
	}

	// Date - use LatestMessageRecvDate from thread (always populated)
	var displayDate time.Time
	if !thread.LatestMessageRecvDate.IsZero() {
		displayDate = thread.LatestMessageRecvDate
	} else if !thread.LatestMessageSentDate.IsZero() {
		displayDate = thread.LatestMessageSentDate
	}

	if !displayDate.IsZero() {
		preview.WriteString(m.theme.KeyBinding.Render("Date: ") + displayDate.Format("Mon Jan 2, 2006 at 3:04 PM") + "\n")
	}

	preview.WriteString("\n" + strings.Repeat("─", 50) + "\n\n")

	// Body - use snippet from thread
	content := thread.Snippet

	if content != "" {
		preview.WriteString(content)
	} else {
		preview.WriteString(m.theme.Dimmed.Render("(no content available - press Enter to view full message)"))
	}

	m.layout.SetPreview(preview.String())
}

// Message types

type threadsLoadedMsg struct {
	threads []domain.Thread
}

type foldersLoadedMsg struct {
	folders []list.Item
}

type errMsg struct {
	err error
}

// formatDate formats a date for display.
func formatDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < 1*time.Minute:
		return "just now"
	case diff < 1*time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}

// truncate truncates a string to a maximum length.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// shouldUseAPISearch returns true if the query should use API search.
// API search is used when there are operators that require server-side filtering.
func (m *MessageList) shouldUseAPISearch(query *components.SearchQuery) bool {
	// Use API search for date filters (can't do client-side)
	if query.After != "" || query.Before != "" {
		return true
	}
	// Use API search for is: operators (unread, starred)
	if len(query.Is) > 0 {
		return true
	}
	// Use API search for has: operators (attachment)
	if len(query.Has) > 0 {
		return true
	}
	// Use API search for in: operators (folder)
	if len(query.In) > 0 {
		return true
	}
	return false
}

// searchMessagesAPI performs an API search using the native query.
func (m *MessageList) searchMessagesAPI(query *components.SearchQuery) tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build native search query
		nativeQuery := query.ToNativeQuery()

		// Fetch threads with search query
		params := &domain.ThreadQueryParams{
			Limit:       50,
			SearchQuery: nativeQuery,
		}

		// Apply folder filter if selected
		if m.selectedFolderID != "" {
			params.In = []string{m.selectedFolderID}
		}

		threads, err := m.global.Client.GetThreads(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return threadsLoadedMsg{threads}
	}
}

// applyClientFilter filters threads client-side based on the search query.
func (m *MessageList) applyClientFilter(query *components.SearchQuery) {
	if query.IsEmpty() || len(m.allThreads) == 0 {
		m.threads = m.allThreads
		m.updateThreadTable()
		return
	}

	var filtered []domain.Thread

	for _, thread := range m.allThreads {
		if m.threadMatchesQuery(thread, query) {
			filtered = append(filtered, thread)
		}
	}

	m.threads = filtered
	m.updateThreadTable()
}

// threadMatchesQuery checks if a thread matches the search query.
func (m *MessageList) threadMatchesQuery(thread domain.Thread, query *components.SearchQuery) bool {
	// Check free text (matches subject, snippet, or participants)
	if query.Text != "" {
		text := strings.ToLower(query.Text)
		matched := false

		// Check subject
		if strings.Contains(strings.ToLower(thread.Subject), text) {
			matched = true
		}

		// Check snippet
		if !matched && strings.Contains(strings.ToLower(thread.Snippet), text) {
			matched = true
		}

		// Check participants
		if !matched {
			for _, p := range thread.Participants {
				if strings.Contains(strings.ToLower(p.Email), text) ||
					strings.Contains(strings.ToLower(p.Name), text) {
					matched = true
					break
				}
			}
		}

		if !matched {
			return false
		}
	}

	// Check from: operator
	if len(query.From) > 0 {
		matched := false
		for _, from := range query.From {
			fromLower := strings.ToLower(from)
			for _, p := range thread.Participants {
				if strings.Contains(strings.ToLower(p.Email), fromLower) ||
					strings.Contains(strings.ToLower(p.Name), fromLower) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check to: operator (for threads, check all participants)
	if len(query.To) > 0 {
		matched := false
		for _, to := range query.To {
			toLower := strings.ToLower(to)
			for _, p := range thread.Participants {
				if strings.Contains(strings.ToLower(p.Email), toLower) ||
					strings.Contains(strings.ToLower(p.Name), toLower) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check subject: operator
	if len(query.Subject) > 0 {
		matched := false
		subjectLower := strings.ToLower(thread.Subject)
		for _, subj := range query.Subject {
			if strings.Contains(subjectLower, strings.ToLower(subj)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}
