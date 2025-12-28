package cache

import (
	"testing"
	"time"
)

// ================================
// SEARCH QUERY PARSING TESTS
// ================================

func TestParseSearchQuery_Basic(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("hello world")

	if query.Text != "hello world" {
		t.Errorf("expected Text 'hello world', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_FromOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("from:john@example.com important email")

	if query.From != "john@example.com" {
		t.Errorf("expected From 'john@example.com', got '%s'", query.From)
	}
	if query.Text != "important email" {
		t.Errorf("expected Text 'important email', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_ToOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("to:recipient@test.com")

	if query.To != "recipient@test.com" {
		t.Errorf("expected To 'recipient@test.com', got '%s'", query.To)
	}
}

func TestParseSearchQuery_SubjectOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("subject:meeting notes")

	if query.Subject != "meeting" {
		t.Errorf("expected Subject 'meeting', got '%s'", query.Subject)
	}
}

func TestParseSearchQuery_SubjectWithQuotes(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery(`subject:"meeting notes" from:john@test.com`)

	if query.Subject != "meeting notes" {
		t.Errorf("expected Subject 'meeting notes', got '%s'", query.Subject)
	}
	if query.From != "john@test.com" {
		t.Errorf("expected From 'john@test.com', got '%s'", query.From)
	}
}

func TestParseSearchQuery_HasAttachment(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("has:attachment")

	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
}

func TestParseSearchQuery_HasAttachments(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("has:attachments")

	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true for 'attachments'")
	}
}

func TestParseSearchQuery_IsUnread(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:unread")

	if query.IsUnread == nil || !*query.IsUnread {
		t.Error("expected IsUnread to be true")
	}
}

func TestParseSearchQuery_IsRead(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:read")

	if query.IsUnread == nil || *query.IsUnread {
		t.Error("expected IsUnread to be false (read)")
	}
}

func TestParseSearchQuery_IsStarred(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:starred")

	if query.IsStarred == nil || !*query.IsStarred {
		t.Error("expected IsStarred to be true")
	}
}

func TestParseSearchQuery_InFolder(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("in:INBOX")

	if query.In != "INBOX" {
		t.Errorf("expected In 'INBOX', got '%s'", query.In)
	}
}

func TestParseSearchQuery_DateAfter(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:2024-01-15")

	if query.After.IsZero() {
		t.Error("expected After date to be set")
	}
	if query.After.Year() != 2024 || query.After.Month() != 1 || query.After.Day() != 15 {
		t.Errorf("expected After date 2024-01-15, got %v", query.After)
	}
}

func TestParseSearchQuery_DateBefore(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("before:2024-12-31")

	if query.Before.IsZero() {
		t.Error("expected Before date to be set")
	}
	if query.Before.Year() != 2024 || query.Before.Month() != 12 || query.Before.Day() != 31 {
		t.Errorf("expected Before date 2024-12-31, got %v", query.Before)
	}
}

func TestParseSearchQuery_RelativeDateToday(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:today")

	if query.After.IsZero() {
		t.Error("expected After date to be set for 'today'")
	}
	// Should be today's date at midnight
	now := time.Now()
	if query.After.Year() != now.Year() || query.After.Month() != now.Month() || query.After.Day() != now.Day() {
		t.Errorf("expected After date to be today, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateYesterday(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:yesterday")

	if query.After.IsZero() {
		t.Error("expected After date to be set for 'yesterday'")
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	if query.After.Day() != yesterday.Day() {
		t.Errorf("expected After date to be yesterday, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateDays(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:7d")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '7d'")
	}
	expected := time.Now().AddDate(0, 0, -7)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~7 days ago, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateWeeks(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:2w")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '2w'")
	}
	expected := time.Now().AddDate(0, 0, -14)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~14 days ago, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateMonths(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:3m")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '3m'")
	}
	expected := time.Now().AddDate(0, -3, 0)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~3 months ago, got %v", query.After)
	}
}

func TestParseSearchQuery_MultipleOperators(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("from:sender@test.com to:recipient@test.com is:unread has:attachment important")

	if query.From != "sender@test.com" {
		t.Errorf("expected From 'sender@test.com', got '%s'", query.From)
	}
	if query.To != "recipient@test.com" {
		t.Errorf("expected To 'recipient@test.com', got '%s'", query.To)
	}
	if query.IsUnread == nil || !*query.IsUnread {
		t.Error("expected IsUnread to be true")
	}
	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
	if query.Text != "important" {
		t.Errorf("expected Text 'important', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_EmptyQuery(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("")

	if query.Text != "" {
		t.Errorf("expected empty Text, got '%s'", query.Text)
	}
	if query.From != "" {
		t.Errorf("expected empty From, got '%s'", query.From)
	}
}

func TestParseSearchQuery_DateFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string // YYYY-MM-DD
	}{
		{"ISO format", "after:2024-06-15", "2024-06-15"},
		{"Slash format", "after:2024/06/15", "2024-06-15"},
		{"US format", "after:06/15/2024", "2024-06-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseSearchQuery(tt.input)
			if query.After.IsZero() {
				t.Errorf("expected After date to be set for %s", tt.input)
				return
			}
			got := query.After.Format("2006-01-02")
			if got != tt.expected {
				t.Errorf("expected date %s, got %s", tt.expected, got)
			}
		})
	}
}
