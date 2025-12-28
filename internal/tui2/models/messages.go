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

// BackMsg is sent to go back to the previous screen.
type BackMsg struct{}

// MessageList is the three-pane email list screen.
type MessageList struct {
	global *state.GlobalState
	theme  *styles.Theme

	layout  *components.ThreePaneLayout
	spinner spinner.Model

	threads          []domain.Thread
	foldersLoaded    bool
	loadingFolders   bool
	selectedFolderID string // Currently selected folder for filtering

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

	// Initialize layout with current window size if available
	if global.WindowSize.Width > 0 && global.WindowSize.Height > 0 {
		layoutHeight := global.WindowSize.Height - 4
		if layoutHeight < 10 {
			layoutHeight = 10
		}
		layout.SetSize(global.WindowSize.Width, layoutHeight)
	}

	return &MessageList{
		global:  global,
		theme:   theme,
		layout:  layout,
		spinner: s,
		loading: true,
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
	case tea.KeyMsg:
		// Use msg.String() for all key matching (v2 pattern)
		key := msg.Key()
		keyStr := msg.String()

		// Handle Esc key
		if key.Code == tea.KeyEsc {
			// Go back to dashboard
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Handle ctrl+c (handled by app.go)
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle ctrl+r (refresh)
		if keyStr == "ctrl+r" {
			// Refresh messages (respecting current folder filter)
			m.loading = true
			if m.selectedFolderID != "" {
				return m, tea.Batch(m.spinner.Tick, m.fetchMessagesForFolder(m.selectedFolderID))
			}
			return m, tea.Batch(m.spinner.Tick, m.fetchMessages())
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
		// Reserve space for header (2 lines) and footer (2 lines) = 4 lines
		// Layout height: terminal height - 4 lines for header/footer
		layoutHeight := msg.Height - 4
		if layoutHeight < 10 {
			layoutHeight = 10 // Minimum height
		}
		m.layout.SetSize(msg.Width, layoutHeight)
		return m, nil

	case threadsLoadedMsg:
		m.threads = msg.threads
		m.loading = false
		m.updateThreadTable()
		return m, nil

	case messagesLoadedMsg:
		// Keep for backward compatibility if needed
		m.loading = false
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

	// Build help text
	help := m.theme.Help.Render("c: compose  r: reply  a: reply all  f: forward  Tab: switch pane  Ctrl+R: refresh  esc: back")

	// Build layout (in v2 this returns tea.View, not string, so we need a way to get its string content)
	// For now, keeping the old pattern - this might need adjustment
	layoutView := m.layout.View()

	// Join all sections with single newlines to maximize space
	return tea.NewView(header + "\n" + layoutView + "\n" + help)
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

type messagesLoadedMsg struct {
	messages []domain.Message
}

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
