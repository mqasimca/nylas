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
	id       string            // Thread ID or message ID
	thread   *domain.Thread    // Thread data (if viewing thread)
	messages []*domain.Message // All messages in thread
	message  *domain.Message   // Single message (for backward compatibility)
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
