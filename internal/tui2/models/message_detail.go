package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// MessageDetail displays a single message in detail view.
type MessageDetail struct {
	global   *state.GlobalState
	theme    *styles.Theme
	viewport viewport.Model
	message  *domain.Message
	loading  bool
	err      error
	ready    bool
}

// NewMessageDetail creates a new message detail screen.
func NewMessageDetail(global *state.GlobalState, messageID string) *MessageDetail {
	theme := styles.GetTheme(global.Theme)

	return &MessageDetail{
		global:  global,
		theme:   theme,
		message: &domain.Message{ID: messageID},
		loading: true,
	}
}

// Init implements tea.Model.
func (m *MessageDetail) Init() tea.Cmd {
	// Initialize viewport with current window size
	if m.global.WindowSize.Width > 0 && m.global.WindowSize.Height > 0 {
		m.viewport = viewport.New(m.global.WindowSize.Width, m.global.WindowSize.Height-8)
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
		switch msg.String() {
		case "esc":
			// Go back to message list
			return m, func() tea.Msg { return BackMsg{} }

		case "ctrl+c":
			return m, tea.Quit

		case "r":
			// Reply (placeholder)
			m.global.SetStatus("Reply not yet implemented", 1)
			return m, nil

		case "f":
			// Forward (placeholder)
			m.global.SetStatus("Forward not yet implemented", 1)
			return m, nil

		case "d":
			// Download attachments (placeholder)
			if len(m.message.Attachments) > 0 {
				m.global.SetStatus("Download not yet implemented", 1)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-8) // -8 for header and footer
			m.viewport.YPosition = 3
			m.ready = true
			m.updateViewport()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8
			m.updateViewport()
		}
		return m, nil

	case messageLoadedMsg:
		m.message = msg.message
		m.loading = false
		m.updateViewport()
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
func (m *MessageDetail) View() string {
	if m.err != nil {
		return m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress 'esc' to go back", m.err))
	}

	if m.loading {
		return m.theme.Title.Render("Loading message...")
	}

	if !m.ready {
		return "Initializing..."
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

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// fetchMessage fetches the full message details.
func (m *MessageDetail) fetchMessage() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get the message using the Nylas client
		message, err := m.global.Client.GetMessage(ctx, m.global.GrantID, m.message.ID)
		if err != nil {
			return errMsg{err}
		}

		return messageLoadedMsg{message}
	}
}

// updateViewport updates the viewport content with the message details.
func (m *MessageDetail) updateViewport() {
	if m.message == nil {
		return
	}

	var content strings.Builder

	// Subject
	content.WriteString(m.theme.Title.Render("Subject: "))
	if m.message.Subject != "" {
		content.WriteString(m.message.Subject)
	} else {
		content.WriteString("(no subject)")
	}
	content.WriteString("\n\n")

	// From
	content.WriteString(m.theme.KeyBinding.Render("From: "))
	if len(m.message.From) > 0 {
		from := formatParticipant(m.message.From[0])
		content.WriteString(from)
	} else {
		content.WriteString("Unknown")
	}
	content.WriteString("\n")

	// To
	if len(m.message.To) > 0 {
		content.WriteString(m.theme.KeyBinding.Render("To: "))
		toList := make([]string, len(m.message.To))
		for i, to := range m.message.To {
			toList[i] = formatParticipant(to)
		}
		content.WriteString(strings.Join(toList, ", "))
		content.WriteString("\n")
	}

	// Cc
	if len(m.message.Cc) > 0 {
		content.WriteString(m.theme.KeyBinding.Render("Cc: "))
		ccList := make([]string, len(m.message.Cc))
		for i, cc := range m.message.Cc {
			ccList[i] = formatParticipant(cc)
		}
		content.WriteString(strings.Join(ccList, ", "))
		content.WriteString("\n")
	}

	// Date
	content.WriteString(m.theme.KeyBinding.Render("Date: "))
	content.WriteString(m.message.Date.Format("Mon, Jan 2, 2006 at 3:04 PM MST"))
	content.WriteString("\n")

	// Attachments
	if len(m.message.Attachments) > 0 {
		content.WriteString("\n")
		content.WriteString(m.theme.KeyBinding.Render(fmt.Sprintf("Attachments (%d):", len(m.message.Attachments))))
		content.WriteString("\n")
		for _, att := range m.message.Attachments {
			size := formatSize(att.Size)
			content.WriteString(fmt.Sprintf("  📎 %s (%s)\n", att.Filename, size))
		}
	}

	// Separator
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", 80))
	content.WriteString("\n\n")

	// Body
	if m.message.Body != "" {
		// Strip HTML tags for basic display (simple implementation)
		body := stripHTML(m.message.Body)
		content.WriteString(body)
	} else {
		content.WriteString(m.theme.Dimmed.Render("(no content)"))
	}

	m.viewport.SetContent(content.String())
}

// buildHelpText builds the help text for the footer.
func (m *MessageDetail) buildHelpText() string {
	helps := []string{
		"↑/↓: scroll",
		"r: reply",
		"f: forward",
	}

	if len(m.message.Attachments) > 0 {
		helps = append(helps, "d: download")
	}

	helps = append(helps, "esc: back", "Ctrl+C: quit")

	return m.theme.Help.Render(strings.Join(helps, "  "))
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
