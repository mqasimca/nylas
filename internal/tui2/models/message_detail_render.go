package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

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
			fmt.Fprintf(content, "  📎 %s (%s)\n", att.Filename, size)
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
