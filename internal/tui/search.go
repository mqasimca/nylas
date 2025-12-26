package tui

import (
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
)

// SearchQuery represents a parsed search query with operators.
type SearchQuery struct {
	From          string // from: operator
	To            string // to: operator
	Subject       string // subject: operator
	HasAttachment bool   // has:attachment
	IsUnread      *bool  // is:unread or is:read
	IsStarred     *bool  // is:starred
	FreeText      string // Everything else
}

// ParseSearchQuery parses a search string into operators and free text.
// Supported operators:
//   - from:email@example.com
//   - to:email@example.com
//   - subject:keyword
//   - has:attachment
//   - is:unread / is:read
//   - is:starred
func ParseSearchQuery(query string) *SearchQuery {
	sq := &SearchQuery{}

	// Split by spaces, but be careful with quoted strings
	parts := splitSearchParts(query)
	var freeTextParts []string

	for _, part := range parts {
		lower := strings.ToLower(part)

		switch {
		case strings.HasPrefix(lower, "from:"):
			sq.From = strings.TrimPrefix(part, "from:")
			sq.From = strings.TrimPrefix(sq.From, "From:")
		case strings.HasPrefix(lower, "to:"):
			sq.To = strings.TrimPrefix(part, "to:")
			sq.To = strings.TrimPrefix(sq.To, "To:")
		case strings.HasPrefix(lower, "subject:"):
			sq.Subject = strings.TrimPrefix(part, "subject:")
			sq.Subject = strings.TrimPrefix(sq.Subject, "Subject:")
		case lower == "has:attachment":
			sq.HasAttachment = true
		case lower == "is:unread":
			unread := true
			sq.IsUnread = &unread
		case lower == "is:read":
			read := false
			sq.IsUnread = &read
		case lower == "is:starred":
			starred := true
			sq.IsStarred = &starred
		default:
			freeTextParts = append(freeTextParts, part)
		}
	}

	sq.FreeText = strings.Join(freeTextParts, " ")
	return sq
}

// splitSearchParts splits a search query by spaces, respecting quoted strings.
func splitSearchParts(query string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range query {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = 0
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// MatchesThread checks if a thread matches the search query.
func (sq *SearchQuery) MatchesThread(thread *domain.Thread) bool {
	// Check from: operator
	if sq.From != "" {
		found := false
		for _, p := range thread.Participants {
			if containsIgnoreCase(p.Email, sq.From) || containsIgnoreCase(p.Name, sq.From) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check to: operator (participants for threads)
	if sq.To != "" {
		found := false
		for _, p := range thread.Participants {
			if containsIgnoreCase(p.Email, sq.To) || containsIgnoreCase(p.Name, sq.To) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check subject: operator
	if sq.Subject != "" && !containsIgnoreCase(thread.Subject, sq.Subject) {
		return false
	}

	// Check has:attachment - for threads, check if HasAttachments field exists or snippet mentions attachments
	if sq.HasAttachment && !thread.HasAttachments {
		return false
	}

	// Check is:unread
	if sq.IsUnread != nil {
		if *sq.IsUnread && !thread.Unread {
			return false
		}
		if !*sq.IsUnread && thread.Unread {
			return false
		}
	}

	// Check is:starred
	if sq.IsStarred != nil {
		if *sq.IsStarred && !thread.Starred {
			return false
		}
		if !*sq.IsStarred && thread.Starred {
			return false
		}
	}

	// Check free text - search in subject and snippet
	if sq.FreeText != "" {
		if !containsIgnoreCase(thread.Subject, sq.FreeText) &&
			!containsIgnoreCase(thread.Snippet, sq.FreeText) {
			// Also check participants
			found := false
			for _, p := range thread.Participants {
				if containsIgnoreCase(p.Email, sq.FreeText) || containsIgnoreCase(p.Name, sq.FreeText) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// MatchesDraft checks if a draft matches the search query.
func (sq *SearchQuery) MatchesDraft(draft *domain.Draft) bool {
	// Check to: operator
	if sq.To != "" {
		found := false
		for _, p := range draft.To {
			if containsIgnoreCase(p.Email, sq.To) || containsIgnoreCase(p.Name, sq.To) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check subject: operator
	if sq.Subject != "" && !containsIgnoreCase(draft.Subject, sq.Subject) {
		return false
	}

	// Check has:attachment
	if sq.HasAttachment && len(draft.Attachments) == 0 {
		return false
	}

	// Check free text - search in subject and body
	if sq.FreeText != "" {
		if !containsIgnoreCase(draft.Subject, sq.FreeText) &&
			!containsIgnoreCase(draft.Body, sq.FreeText) {
			// Also check recipients
			found := false
			for _, p := range draft.To {
				if containsIgnoreCase(p.Email, sq.FreeText) || containsIgnoreCase(p.Name, sq.FreeText) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// GetSearchHints returns hints about available search operators.
func GetSearchHints() string {
	return "Search: from:, to:, subject:, has:attachment, is:unread, is:starred"
}
