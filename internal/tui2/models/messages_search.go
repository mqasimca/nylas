package models

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
)

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
