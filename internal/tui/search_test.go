package tui

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestParseSearchQuery_FreeText(t *testing.T) {
	sq := ParseSearchQuery("hello world")

	if sq.FreeText != "hello world" {
		t.Errorf("FreeText = %q, want %q", sq.FreeText, "hello world")
	}

	if sq.From != "" || sq.To != "" || sq.Subject != "" {
		t.Error("operators should be empty for free text only")
	}
}

func TestParseSearchQuery_FromOperator(t *testing.T) {
	sq := ParseSearchQuery("from:test@example.com")

	if sq.From != "test@example.com" {
		t.Errorf("From = %q, want %q", sq.From, "test@example.com")
	}

	if sq.FreeText != "" {
		t.Errorf("FreeText = %q, want empty", sq.FreeText)
	}
}

func TestParseSearchQuery_ToOperator(t *testing.T) {
	sq := ParseSearchQuery("to:recipient@example.com")

	if sq.To != "recipient@example.com" {
		t.Errorf("To = %q, want %q", sq.To, "recipient@example.com")
	}
}

func TestParseSearchQuery_SubjectOperator(t *testing.T) {
	sq := ParseSearchQuery("subject:meeting")

	if sq.Subject != "meeting" {
		t.Errorf("Subject = %q, want %q", sq.Subject, "meeting")
	}
}

func TestParseSearchQuery_HasAttachment(t *testing.T) {
	sq := ParseSearchQuery("has:attachment")

	if !sq.HasAttachment {
		t.Error("HasAttachment should be true")
	}
}

func TestParseSearchQuery_IsUnread(t *testing.T) {
	sq := ParseSearchQuery("is:unread")

	if sq.IsUnread == nil || !*sq.IsUnread {
		t.Error("IsUnread should be true")
	}

	sq = ParseSearchQuery("is:read")

	if sq.IsUnread == nil || *sq.IsUnread {
		t.Error("IsUnread should be false for is:read")
	}
}

func TestParseSearchQuery_IsStarred(t *testing.T) {
	sq := ParseSearchQuery("is:starred")

	if sq.IsStarred == nil || !*sq.IsStarred {
		t.Error("IsStarred should be true")
	}
}

func TestParseSearchQuery_CombinedOperators(t *testing.T) {
	sq := ParseSearchQuery("from:sender@example.com subject:meeting hello world")

	if sq.From != "sender@example.com" {
		t.Errorf("From = %q, want %q", sq.From, "sender@example.com")
	}

	if sq.Subject != "meeting" {
		t.Errorf("Subject = %q, want %q", sq.Subject, "meeting")
	}

	if sq.FreeText != "hello world" {
		t.Errorf("FreeText = %q, want %q", sq.FreeText, "hello world")
	}
}

func TestSplitSearchParts_Quotes(t *testing.T) {
	parts := splitSearchParts("subject:\"hello world\" from:test")

	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2", len(parts))
	}

	if parts[0] != "subject:hello world" {
		t.Errorf("parts[0] = %q, want %q", parts[0], "subject:hello world")
	}

	if parts[1] != "from:test" {
		t.Errorf("parts[1] = %q, want %q", parts[1], "from:test")
	}
}

func TestSearchQuery_MatchesThread(t *testing.T) {
	thread := &domain.Thread{
		Subject:        "Weekly Team Meeting",
		Snippet:        "Let's discuss project updates",
		HasAttachments: true,
		Unread:         true,
		Starred:        false,
		Participants: []domain.EmailParticipant{
			{Email: "sender@example.com", Name: "Sender Name"},
			{Email: "recipient@example.com", Name: "Recipient Name"},
		},
	}

	tests := []struct {
		name    string
		query   string
		matches bool
	}{
		{"free text match subject", "meeting", true},
		{"free text match snippet", "project", true},
		{"free text no match", "vacation", false},
		{"from operator match", "from:sender@example.com", true},
		{"from operator partial match", "from:sender", true},
		{"from operator no match", "from:other@example.com", false},
		{"subject match", "subject:team", true},
		{"subject no match", "subject:vacation", false},
		{"has attachment match", "has:attachment", true},
		{"is unread match", "is:unread", true},
		{"is read no match", "is:read", false},
		{"is starred no match", "is:starred", false},
		{"combined match", "from:sender subject:meeting", true},
		{"combined partial fail", "from:sender subject:vacation", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := ParseSearchQuery(tt.query)
			result := sq.MatchesThread(thread)
			if result != tt.matches {
				t.Errorf("MatchesThread(%q) = %v, want %v", tt.query, result, tt.matches)
			}
		})
	}
}

func TestSearchQuery_MatchesDraft(t *testing.T) {
	draft := &domain.Draft{
		Subject: "Draft Email",
		Body:    "This is the draft body content",
		To: []domain.EmailParticipant{
			{Email: "recipient@example.com", Name: "Recipient"},
		},
		Attachments: []domain.Attachment{
			{ID: "att-1", Filename: "file.pdf"},
		},
	}

	tests := []struct {
		name    string
		query   string
		matches bool
	}{
		{"free text match subject", "draft", true},
		{"free text match body", "content", true},
		{"free text no match", "vacation", false},
		{"to operator match", "to:recipient@example.com", true},
		{"to operator partial match", "to:recipient", true},
		{"to operator no match", "to:other@example.com", false},
		{"subject match", "subject:email", true},
		{"subject no match", "subject:vacation", false},
		{"has attachment match", "has:attachment", true},
		{"combined match", "to:recipient subject:draft", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := ParseSearchQuery(tt.query)
			result := sq.MatchesDraft(draft)
			if result != tt.matches {
				t.Errorf("MatchesDraft(%q) = %v, want %v", tt.query, result, tt.matches)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s       string
		substr  string
		matches bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "lo wo", true},
		{"Hello World", "foo", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := containsIgnoreCase(tt.s, tt.substr)
		if result != tt.matches {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.matches)
		}
	}
}

func TestGetSearchHints(t *testing.T) {
	hints := GetSearchHints()

	if hints == "" {
		t.Error("GetSearchHints should not return empty string")
	}

	// Check for key operators
	expectedOperators := []string{"from:", "to:", "subject:", "has:attachment"}
	for _, op := range expectedOperators {
		if !containsIgnoreCase(hints, op) {
			t.Errorf("GetSearchHints() should contain %q", op)
		}
	}
}

func TestSearchQuery_EmptyQuery(t *testing.T) {
	sq := ParseSearchQuery("")

	if sq.From != "" || sq.To != "" || sq.Subject != "" || sq.FreeText != "" {
		t.Error("empty query should result in empty SearchQuery")
	}

	if sq.HasAttachment || sq.IsUnread != nil || sq.IsStarred != nil {
		t.Error("flags should not be set for empty query")
	}
}

func TestSearchQuery_MatchesThread_NoAttachments(t *testing.T) {
	thread := &domain.Thread{
		Subject:        "No attachments email",
		HasAttachments: false,
	}

	sq := ParseSearchQuery("has:attachment")
	if sq.MatchesThread(thread) {
		t.Error("thread without attachments should not match has:attachment")
	}
}
