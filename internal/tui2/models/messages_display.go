package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

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
