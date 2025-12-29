package models

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/components"
)

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
