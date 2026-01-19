package cache

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SearchQuery represents a parsed search query with operators.
type SearchQuery struct {
	// Text is the free-text search portion
	Text string
	// Operators are field-specific filters
	From          string
	To            string
	Subject       string
	HasAttachment *bool
	IsUnread      *bool
	IsStarred     *bool
	After         time.Time
	Before        time.Time
	In            string // Folder ID
}

// operatorRegex matches search operators like "from:john@example.com"
var operatorRegex = regexp.MustCompile(`(\w+):("[^"]+"|[^\s]+)`)

// ParseSearchQuery parses a search string into structured query.
// Supports operators: from:, to:, subject:, has:attachment, is:unread, is:starred, after:, before:, in:
func ParseSearchQuery(query string) *SearchQuery {
	sq := &SearchQuery{}
	remaining := query

	// Extract operators
	matches := operatorRegex.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		operator := strings.ToLower(match[1])
		value := strings.Trim(match[2], `"`)

		switch operator {
		case "from":
			sq.From = value
		case "to":
			sq.To = value
		case "subject":
			sq.Subject = value
		case "has":
			if strings.EqualFold(value, "attachment") || strings.EqualFold(value, "attachments") {
				t := true
				sq.HasAttachment = &t
			}
		case "is":
			switch strings.ToLower(value) {
			case "unread":
				t := true
				sq.IsUnread = &t
			case "read":
				f := false
				sq.IsUnread = &f
			case "starred":
				t := true
				sq.IsStarred = &t
			}
		case "after":
			if t, err := parseDate(value); err == nil {
				sq.After = t
			}
		case "before":
			if t, err := parseDate(value); err == nil {
				sq.Before = t
			}
		case "in":
			sq.In = value
		}

		// Remove matched operator from remaining text
		remaining = strings.Replace(remaining, match[0], "", 1)
	}

	// Clean up remaining text
	sq.Text = strings.TrimSpace(remaining)

	return sq
}

// parseDate parses common date formats.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Handle relative dates
	s = strings.ToLower(s)
	now := time.Now()

	switch s {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location()), nil
	case "week", "thisweek", "this-week":
		weekday := int(now.Weekday())
		return now.AddDate(0, 0, -weekday), nil
	case "month", "thismonth", "this-month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	}

	// Handle "Xd", "Xw", "Xm" ago
	if strings.HasSuffix(s, "d") || strings.HasSuffix(s, "w") || strings.HasSuffix(s, "m") {
		var num int
		var unit string
		if _, err := fmt.Sscanf(s, "%d%s", &num, &unit); err == nil {
			switch unit {
			case "d":
				return now.AddDate(0, 0, -num), nil
			case "w":
				return now.AddDate(0, 0, -num*7), nil
			case "m":
				return now.AddDate(0, -num, 0), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("unknown date format: %s", s)
}

// SearchEmails performs an advanced search with operator support.
func (s *EmailStore) SearchAdvanced(query string, limit int) ([]*CachedEmail, error) {
	sq := ParseSearchQuery(query)
	return s.SearchWithQuery(sq, limit)
}

// SearchWithQuery performs a search using a parsed query.
func (s *EmailStore) SearchWithQuery(sq *SearchQuery, limit int) ([]*CachedEmail, error) {
	if limit <= 0 {
		limit = 50
	}

	var conditions []string
	var args []any

	// FTS search for text
	if sq.Text != "" {
		conditions = append(conditions, "e.rowid IN (SELECT rowid FROM emails_fts WHERE emails_fts MATCH ?)")
		args = append(args, sq.Text)
	}

	// Subject search (can use FTS or LIKE)
	if sq.Subject != "" {
		conditions = append(conditions, "e.subject LIKE ?")
		args = append(args, "%"+sq.Subject+"%")
	}

	// From filter
	if sq.From != "" {
		conditions = append(conditions, "(e.from_email LIKE ? OR e.from_name LIKE ?)")
		args = append(args, "%"+sq.From+"%", "%"+sq.From+"%")
	}

	// To filter
	if sq.To != "" {
		conditions = append(conditions, "e.to_json LIKE ?")
		args = append(args, "%"+sq.To+"%")
	}

	// Has attachment filter
	if sq.HasAttachment != nil && *sq.HasAttachment {
		conditions = append(conditions, "e.has_attachments = 1")
	}

	// Unread filter
	if sq.IsUnread != nil {
		if *sq.IsUnread {
			conditions = append(conditions, "e.unread = 1")
		} else {
			conditions = append(conditions, "e.unread = 0")
		}
	}

	// Starred filter
	if sq.IsStarred != nil && *sq.IsStarred {
		conditions = append(conditions, "e.starred = 1")
	}

	// Date filters
	if !sq.After.IsZero() {
		conditions = append(conditions, "e.date >= ?")
		args = append(args, sq.After.Unix())
	}
	if !sq.Before.IsZero() {
		conditions = append(conditions, "e.date < ?")
		args = append(args, sq.Before.Unix())
	}

	// Folder filter
	if sq.In != "" {
		conditions = append(conditions, "e.folder_id = ?")
		args = append(args, sq.In)
	}

	// Build query
	baseQuery := `
		SELECT e.id, e.thread_id, e.folder_id, e.subject, e.snippet,
			e.from_name, e.from_email, e.to_json, e.cc_json, e.bcc_json,
			e.date, e.unread, e.starred, e.has_attachments,
			e.body_html, e.body_text, e.cached_at
		FROM emails e
	`

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY e.date DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search emails: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Pre-allocate slice with expected capacity
	emails := make([]*CachedEmail, 0, limit)
	for rows.Next() {
		email, err := scanEmailGeneric(rows)
		if err != nil {
			return nil, fmt.Errorf("scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, rows.Err()
}

// UnifiedSearch searches across emails, events, and contacts.
type UnifiedSearchResult struct {
	Type     string // "email", "event", "contact"
	ID       string
	Title    string
	Subtitle string
	Date     time.Time
}

// UnifiedSearch performs search across all data types.
func UnifiedSearch(db *sql.DB, query string, limit int) ([]*UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	perType := limit / 3
	if perType < 5 {
		perType = 5
	}

	var results []*UnifiedSearchResult

	// Search emails
	emailStore := NewEmailStore(db)
	emails, err := emailStore.Search(query, perType)
	if err == nil {
		for _, e := range emails {
			results = append(results, &UnifiedSearchResult{
				Type:     "email",
				ID:       e.ID,
				Title:    e.Subject,
				Subtitle: e.FromName + " <" + e.FromEmail + ">",
				Date:     e.Date,
			})
		}
	}

	// Search events
	eventStore := NewEventStore(db)
	events, err := eventStore.Search(query, perType)
	if err == nil {
		for _, e := range events {
			results = append(results, &UnifiedSearchResult{
				Type:     "event",
				ID:       e.ID,
				Title:    e.Title,
				Subtitle: e.Location,
				Date:     e.StartTime,
			})
		}
	}

	// Search contacts
	contactStore := NewContactStore(db)
	contacts, err := contactStore.Search(query, perType)
	if err == nil {
		for _, c := range contacts {
			name := c.DisplayName
			if name == "" {
				name = c.GivenName + " " + c.Surname
			}
			results = append(results, &UnifiedSearchResult{
				Type:     "contact",
				ID:       c.ID,
				Title:    strings.TrimSpace(name),
				Subtitle: c.Email,
				Date:     c.CachedAt,
			})
		}
	}

	return results, nil
}
