package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// MessageDetail displays a single message or thread in detail view.
type MessageDetail struct {
	global   *state.GlobalState
	theme    *styles.Theme
	viewport viewport.Model
	id       string              // Thread ID or message ID
	thread   *domain.Thread      // Thread data (if viewing thread)
	messages []*domain.Message   // All messages in thread
	message  *domain.Message     // Single message (for backward compatibility)
	loading  bool
	err      error
	ready    bool

	// Confirmation dialog
	pendingConfirmation *confirmationMsg
}

// NewMessageDetail creates a new message detail screen.
// The id parameter can be either a thread ID or message ID.
func NewMessageDetail(global *state.GlobalState, id string) *MessageDetail {
	theme := styles.GetTheme(global.Theme)

	return &MessageDetail{
		global:  global,
		theme:   theme,
		id:      id,
		loading: true,
	}
}

// Init implements tea.Model.
func (m *MessageDetail) Init() tea.Cmd {
	// Initialize viewport with current window size
	if m.global.WindowSize.Width > 0 && m.global.WindowSize.Height > 0 {
		m.viewport = viewport.New()
		m.viewport.SetWidth(m.global.WindowSize.Width)
		m.viewport.SetHeight(m.global.WindowSize.Height - 8)
		m.viewport.YPosition = 3
		m.ready = true
	}

	return tea.Batch(
		m.fetchMessage(),
	)
}

// Update implements tea.Model.
func (m *MessageDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Use msg.String() for all key matching (v2 pattern)
		key := msg.Key()
		keyStr := msg.String()

		// Handle confirmation dialog first
		if m.pendingConfirmation != nil {
			switch keyStr {
			case "y":
				cmd := m.pendingConfirmation.onConfirm
				m.pendingConfirmation = nil
				return m, cmd
			case "n":
				m.pendingConfirmation = nil
				m.global.SetStatus("Cancelled", 0)
				return m, nil
			}
			// Also check for Esc key
			if key.Code == tea.KeyEsc {
				m.pendingConfirmation = nil
				m.global.SetStatus("Cancelled", 0)
				return m, nil
			}
			return m, nil
		}

		// Check Esc key
		if key.Code == tea.KeyEsc {
			// Go back to message list
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Check ctrl+c (handled by app.go)
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		switch keyStr {

		case "r":
			// Reply
			msg := m.getActiveMessage()
			if msg != nil {
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ScreenCompose,
						Data: ComposeData{
							Mode:    ComposeModeReply,
							Message: msg,
						},
					}
				}
			}
			return m, nil

		case "a":
			// Reply All
			msg := m.getActiveMessage()
			if msg != nil {
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ScreenCompose,
						Data: ComposeData{
							Mode:    ComposeModeReplyAll,
							Message: msg,
						},
					}
				}
			}
			return m, nil

		case "f":
			// Forward
			msg := m.getActiveMessage()
			if msg != nil {
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ScreenCompose,
						Data: ComposeData{
							Mode:    ComposeModeForward,
							Message: msg,
						},
					}
				}
			}
			return m, nil

		case "s":
			// Star/Unstar toggle
			if m.thread != nil || m.message != nil {
				return m, m.toggleStar()
			}
			return m, nil

		case "u":
			// Mark read/unread toggle
			if m.thread != nil || m.message != nil {
				return m, m.toggleUnread()
			}
			return m, nil

		case "d":
			// Delete with confirmation
			subject := ""
			if m.thread != nil {
				subject = m.thread.Subject
			} else if m.message != nil {
				subject = m.message.Subject
			}
			if subject == "" {
				subject = "(no subject)"
			}

			if m.thread != nil || m.message != nil {
				itemType := "message"
				if m.thread != nil {
					itemType = fmt.Sprintf("thread (%d messages)", len(m.messages))
				}
				return m, func() tea.Msg {
					return confirmationMsg{
						message:   fmt.Sprintf("Delete %s '%s'? (y/n)", itemType, subject),
						onConfirm: m.deleteMessage(),
					}
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		if !m.ready {
			m.viewport = viewport.New()
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - 8) // -8 for header and footer
			m.viewport.YPosition = 3
			m.ready = true
			m.updateViewport()
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - 8)
			m.updateViewport()
		}
		return m, nil

	case messageLoadedMsg:
		m.message = msg.message
		m.loading = false
		m.updateViewport()
		return m, nil

	case threadLoadedMsg:
		m.thread = msg.thread
		m.messages = msg.messages
		m.loading = false
		m.updateViewport()
		return m, nil

	case messageUpdatedMsg:
		m.message = msg.message
		m.updateViewport()
		m.global.SetStatus(fmt.Sprintf("Message %s", msg.action), 0)
		return m, nil

	case messageDeletedMsg:
		m.global.SetStatus("Message deleted", 0)
		return m, func() tea.Msg { return BackMsg{} }

	case messageActionErrorMsg:
		m.global.SetStatus(fmt.Sprintf("Failed to %s: %v", msg.action, msg.err), 1)
		return m, nil

	case confirmationMsg:
		m.global.SetStatus(msg.message, 0)
		m.pendingConfirmation = &msg
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	}

	// Update viewport
	if m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m *MessageDetail) View() tea.View {
	if m.err != nil {
		return tea.NewView(m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress 'esc' to go back", m.err)))
	}

	if m.loading {
		return tea.NewView(m.theme.Title.Render("Loading message..."))
	}

	if !m.ready {
		return tea.NewView("Initializing...")
	}

	// Build header
	var sections []string

	// Title bar
	title := m.theme.Title.Render("Message Detail")
	sections = append(sections, title)
	sections = append(sections, "")

	// Viewport content
	sections = append(sections, m.viewport.View())
	sections = append(sections, "")

	// Help text
	help := m.buildHelpText()
	sections = append(sections, help)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// fetchMessage fetches the full message or thread details.
// It tries to fetch as a thread first, then falls back to a single message.
func (m *MessageDetail) fetchMessage() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Try fetching as a thread first (since MessageList now passes thread IDs)
		thread, err := m.global.Client.GetThread(ctx, m.global.GrantID, m.id)
		if err == nil {
			// Successfully fetched as thread
			// Now fetch all messages in the thread
			var messages []*domain.Message
			for _, msgID := range thread.MessageIDs {
				msg, err := m.global.Client.GetMessage(ctx, m.global.GrantID, msgID)
				if err != nil {
					// Log error but continue with other messages
					continue
				}
				messages = append(messages, msg)
			}
			return threadLoadedMsg{thread: thread, messages: messages}
		}

		// Fall back to fetching as a single message
		message, err := m.global.Client.GetMessage(ctx, m.global.GrantID, m.id)
		if err != nil {
			return errMsg{err}
		}

		return messageLoadedMsg{message}
	}
}

// updateViewport updates the viewport content with the message or thread details.
func (m *MessageDetail) updateViewport() {
	// Handle thread view (multiple messages)
	if m.thread != nil && len(m.messages) > 0 {
		m.renderThreadView()
		return
	}

	// Handle single message view
	if m.message == nil {
		return
	}

	var content strings.Builder
	m.renderMessage(&content, m.message, false)
	m.viewport.SetContent(content.String())
}

// renderThreadView renders all messages in a thread conversation.
func (m *MessageDetail) renderThreadView() {
	var content strings.Builder

	// Thread header
	subject := m.thread.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	msgCount := len(m.messages)
	content.WriteString(m.theme.Title.Render(fmt.Sprintf("Thread: %s (%d messages)", subject, msgCount)))
	content.WriteString("\n\n")

	// Render each message in the thread
	for i, msg := range m.messages {
		// Message separator (except for first message)
		if i > 0 {
			content.WriteString("\n")
			content.WriteString(m.theme.Dimmed.Render(strings.Repeat("─", 80)))
			content.WriteString("\n\n")
		}

		// Message header with number
		content.WriteString(m.theme.Subtitle.Render(fmt.Sprintf("Message %d of %d", i+1, msgCount)))
		content.WriteString("\n\n")

		// Render the message
		m.renderMessage(&content, msg, true)
	}

	m.viewport.SetContent(content.String())
}

// renderMessage renders a single message to the content builder.
func (m *MessageDetail) renderMessage(content *strings.Builder, msg *domain.Message, compact bool) {
	// From
	content.WriteString(m.theme.KeyBinding.Render("From: "))
	if len(msg.From) > 0 {
		from := formatParticipant(msg.From[0])
		content.WriteString(from)
	} else {
		content.WriteString("Unknown")
	}
	content.WriteString("\n")

	// To
	if len(msg.To) > 0 {
		content.WriteString(m.theme.KeyBinding.Render("To: "))
		toList := make([]string, len(msg.To))
		for i, to := range msg.To {
			toList[i] = formatParticipant(to)
		}
		content.WriteString(strings.Join(toList, ", "))
		content.WriteString("\n")
	}

	// Cc
	if len(msg.Cc) > 0 {
		content.WriteString(m.theme.KeyBinding.Render("Cc: "))
		ccList := make([]string, len(msg.Cc))
		for i, cc := range msg.Cc {
			ccList[i] = formatParticipant(cc)
		}
		content.WriteString(strings.Join(ccList, ", "))
		content.WriteString("\n")
	}

	// Date
	content.WriteString(m.theme.KeyBinding.Render("Date: "))
	content.WriteString(msg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM MST"))
	content.WriteString("\n")

	// Attachments
	if len(msg.Attachments) > 0 {
		content.WriteString("\n")
		content.WriteString(m.theme.KeyBinding.Render(fmt.Sprintf("Attachments (%d):", len(msg.Attachments))))
		content.WriteString("\n")
		for _, att := range msg.Attachments {
			size := formatSize(att.Size)
			content.WriteString(fmt.Sprintf("  📎 %s (%s)\n", att.Filename, size))
		}
	}

	// Body
	content.WriteString("\n")
	if msg.Body != "" {
		// Strip HTML tags for basic display
		body := stripHTML(msg.Body)
		content.WriteString(body)
	} else {
		content.WriteString(m.theme.Dimmed.Render("(no content)"))
	}
	content.WriteString("\n")
}

// getActiveMessage returns the message to use for compose actions.
// For threads, returns the latest message. For single messages, returns that message.
func (m *MessageDetail) getActiveMessage() *domain.Message {
	// Thread view: use latest message (last in chronological order)
	if m.thread != nil && len(m.messages) > 0 {
		return m.messages[len(m.messages)-1]
	}

	// Single message view
	return m.message
}

// buildHelpText builds the help text for the footer.
func (m *MessageDetail) buildHelpText() string {
	helps := []string{
		"↑/↓: scroll",
		"r: reply",
		"a: reply all",
		"f: forward",
		"s: star",
		"u: mark read/unread",
		"d: delete",
		"esc: back",
	}

	return m.theme.Help.Render(strings.Join(helps, "  "))
}

// toggleStar toggles the starred status of the message or thread.
func (m *MessageDetail) toggleStar() tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Thread view: update thread
		if m.thread != nil {
			newStarred := !m.thread.Starred

			updated, err := m.global.Client.UpdateThread(
				ctx,
				m.global.GrantID,
				m.thread.ID,
				&domain.UpdateMessageRequest{Starred: &newStarred},
			)
			if err != nil {
				return messageActionErrorMsg{err: err, action: "star thread"}
			}

			action := "starred"
			if !newStarred {
				action = "unstarred"
			}

			// Update local state
			m.thread = updated
			return messageUpdatedMsg{message: &updated.LatestDraftOrMessage, action: action}
		}

		// Single message view: update message
		newStarred := !m.message.Starred

		updated, err := m.global.Client.UpdateMessage(
			ctx,
			m.global.GrantID,
			m.message.ID,
			&domain.UpdateMessageRequest{Starred: &newStarred},
		)
		if err != nil {
			return messageActionErrorMsg{err: err, action: "star"}
		}

		action := "starred"
		if !newStarred {
			action = "unstarred"
		}

		return messageUpdatedMsg{message: updated, action: action}
	}
}

// toggleUnread toggles the unread status of the message or thread.
func (m *MessageDetail) toggleUnread() tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Thread view: update thread
		if m.thread != nil {
			newUnread := !m.thread.Unread

			updated, err := m.global.Client.UpdateThread(
				ctx,
				m.global.GrantID,
				m.thread.ID,
				&domain.UpdateMessageRequest{Unread: &newUnread},
			)
			if err != nil {
				return messageActionErrorMsg{err: err, action: "mark thread read/unread"}
			}

			action := "marked as unread"
			if !newUnread {
				action = "marked as read"
			}

			// Update local state
			m.thread = updated
			return messageUpdatedMsg{message: &updated.LatestDraftOrMessage, action: action}
		}

		// Single message view: update message
		newUnread := !m.message.Unread

		updated, err := m.global.Client.UpdateMessage(
			ctx,
			m.global.GrantID,
			m.message.ID,
			&domain.UpdateMessageRequest{Unread: &newUnread},
		)
		if err != nil {
			return messageActionErrorMsg{err: err, action: "mark read/unread"}
		}

		action := "marked as unread"
		if !newUnread {
			action = "marked as read"
		}

		return messageUpdatedMsg{message: updated, action: action}
	}
}

// deleteMessage deletes the message or thread.
func (m *MessageDetail) deleteMessage() tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Thread view: delete thread
		if m.thread != nil {
			err := m.global.Client.DeleteThread(ctx, m.global.GrantID, m.thread.ID)
			if err != nil {
				return messageActionErrorMsg{err: err, action: "delete thread"}
			}

			return messageDeletedMsg{messageID: m.thread.ID}
		}

		// Single message view: delete message
		err := m.global.Client.DeleteMessage(ctx, m.global.GrantID, m.message.ID)
		if err != nil {
			return messageActionErrorMsg{err: err, action: "delete"}
		}

		return messageDeletedMsg{messageID: m.message.ID}
	}
}

// formatParticipant formats an email participant for display.
func formatParticipant(p domain.EmailParticipant) string {
	if p.Name != "" {
		return fmt.Sprintf("%s <%s>", p.Name, p.Email)
	}
	return p.Email
}

// formatSize formats a byte size for display.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// stripHTML removes HTML tags from a string (basic implementation).
func stripHTML(html string) string {
	// Very basic HTML stripping - replace common tags with formatting
	s := html

	// Replace common block elements with line breaks
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</div>", "\n")

	// Remove all remaining HTML tags
	inTag := false
	var result strings.Builder
	for _, ch := range s {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}

	// Clean up excessive newlines
	cleaned := result.String()
	cleaned = strings.ReplaceAll(cleaned, "\n\n\n", "\n\n")

	// Decode common HTML entities
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")

	return strings.TrimSpace(cleaned)
}

// messageLoadedMsg is sent when a message is loaded.
type messageLoadedMsg struct {
	message *domain.Message
}

// threadLoadedMsg is sent when a thread is loaded.
type threadLoadedMsg struct {
	thread   *domain.Thread
	messages []*domain.Message
}

// messageUpdatedMsg is sent when a message is updated.
type messageUpdatedMsg struct {
	message *domain.Message
	action  string
}

// messageDeletedMsg is sent when a message is deleted.
type messageDeletedMsg struct {
	messageID string
}

// messageActionErrorMsg is sent when a message action fails.
type messageActionErrorMsg struct {
	err    error
	action string
}

// confirmationMsg is sent when confirmation is needed.
type confirmationMsg struct {
	message   string
	onConfirm tea.Cmd
}
