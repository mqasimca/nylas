package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	messages         []domain.Message
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
		layout.SetSize(global.WindowSize.Width, global.WindowSize.Height-4)
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
		// Handle global shortcuts first
		switch msg.String() {
		case "esc":
			// Go back to dashboard
			return m, func() tea.Msg { return BackMsg{} }

		case "ctrl+c":
			return m, tea.Quit

		case "h", "shift+tab":
			// Focus previous pane
			m.layout.FocusPrevious()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil

		case "l", "tab":
			// Focus next pane
			m.layout.FocusNext()
			// Lazy load folders when focusing on folder pane
			if m.layout.GetFocused() == components.FolderPane && !m.foldersLoaded && !m.loadingFolders {
				m.loadingFolders = true
				m.global.SetStatus("Loading folders...", 0)
				return m, m.fetchFolders()
			}
			return m, nil

		case "r":
			// Refresh messages (respecting current folder filter)
			m.loading = true
			if m.selectedFolderID != "" {
				return m, tea.Batch(m.spinner.Tick, m.fetchMessagesForFolder(m.selectedFolderID))
			}
			return m, tea.Batch(m.spinner.Tick, m.fetchMessages())

		case "enter":
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
				// Message selected - navigate to detail view
				idx := m.layout.SelectedMessageIndex()
				if idx >= 0 && idx < len(m.messages) {
					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: ScreenMessageDetail,
							Data:   m.messages[idx].ID,
						}
					}
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		// Update layout size
		m.global.SetWindowSize(msg.Width, msg.Height)
		// Pass full height minus header (2 lines: text + newline) and footer (2 lines: newline + text)
		// The layout's SetSize will handle borders and titles
		m.layout.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case messagesLoadedMsg:
		m.messages = msg.messages
		m.loading = false
		m.updateMessageTable()
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
	if m.layout.GetFocused() == components.MessagePane && !m.loading && len(m.messages) > 0 {
		idx := m.layout.SelectedMessageIndex()
		if idx >= 0 && idx < len(m.messages) {
			m.showMessagePreview(m.messages[idx].ID)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m *MessageList) View() string {
	if m.err != nil {
		return m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to go back", m.err))
	}

	// Build header
	header := m.theme.Title.Render("Messages") + " " +
		m.theme.Subtitle.Render(fmt.Sprintf("(%s)", m.global.Email))

	// Show loading spinner if loading
	if m.loading {
		header += " " + m.spinner.View()
	}

	// Build help text
	help := m.theme.Help.Render("Tab/h/l: switch pane  Enter: select  r: refresh  esc: back  Ctrl+C: quit")

	// Build layout
	layoutView := m.layout.View()

	// Join all sections with single newlines to maximize space
	return header + "\n" + layoutView + "\n" + help
}

// fetchMessages fetches the message list.
func (m *MessageList) fetchMessages() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		messages, err := m.global.Client.GetMessages(ctx, m.global.GrantID, 20)
		if err != nil {
			return errMsg{err}
		}
		return messagesLoadedMsg{messages}
	}
}

// fetchMessagesForFolder fetches messages filtered by folder ID.
func (m *MessageList) fetchMessagesForFolder(folderID string) tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Use GetMessagesWithParams to filter by folder
		params := &domain.MessageQueryParams{
			In:    []string{folderID},
			Limit: 50,
		}

		messages, err := m.global.Client.GetMessagesWithParams(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return messagesLoadedMsg{messages}
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
func (m *MessageList) updateMessageTable() {
	rows := make([]table.Row, len(m.messages))
	for i, msg := range m.messages {
		from := "Unknown"
		if len(msg.From) > 0 {
			if msg.From[0].Name != "" {
				from = msg.From[0].Name
			} else {
				from = msg.From[0].Email
			}
		}

		subject := msg.Subject
		if subject == "" {
			subject = "(no subject)"
		}

		// Format date
		date := formatDate(msg.Date)

		rows[i] = table.Row{
			truncate(from, 20),
			truncate(subject, 40),
			date,
		}
	}

	m.layout.SetMessages(rows)

	// Auto-preview the first message
	if len(m.messages) > 0 {
		m.showMessagePreview(m.messages[0].ID)
	}
}

// showMessagePreview displays a message in the preview pane.
func (m *MessageList) showMessagePreview(msgID string) {
	// Find message by ID
	var msg *domain.Message
	for i := range m.messages {
		if m.messages[i].ID == msgID {
			msg = &m.messages[i]
			break
		}
	}

	if msg == nil {
		m.layout.SetPreview("Message not found")
		return
	}

	// Build preview content
	var preview strings.Builder

	// Header
	preview.WriteString(m.theme.Title.Render("Subject: ") + msg.Subject + "\n\n")

	// From
	from := "Unknown"
	if len(msg.From) > 0 {
		if msg.From[0].Name != "" {
			from = fmt.Sprintf("%s <%s>", msg.From[0].Name, msg.From[0].Email)
		} else {
			from = msg.From[0].Email
		}
	}
	preview.WriteString(m.theme.KeyBinding.Render("From: ") + from + "\n")

	// To
	if len(msg.To) > 0 {
		toList := make([]string, len(msg.To))
		for i, to := range msg.To {
			if to.Name != "" {
				toList[i] = fmt.Sprintf("%s <%s>", to.Name, to.Email)
			} else {
				toList[i] = to.Email
			}
		}
		preview.WriteString(m.theme.KeyBinding.Render("To: ") + strings.Join(toList, ", ") + "\n")
	}

	// Date
	preview.WriteString(m.theme.KeyBinding.Render("Date: ") + msg.Date.Format("Mon Jan 2, 2006 at 3:04 PM") + "\n")

	preview.WriteString("\n" + strings.Repeat("─", 50) + "\n\n")

	// Body - use snippet from message list, or body if available
	content := msg.Snippet
	if content == "" && msg.Body != "" {
		content = msg.Body
	}

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
