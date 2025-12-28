package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestParseSearchQuery_EmptyQuery(t *testing.T) {
	query := ParseSearchQuery("")

	if query.Raw != "" {
		t.Errorf("expected empty Raw, got %q", query.Raw)
	}
	if query.Text != "" {
		t.Errorf("expected empty Text, got %q", query.Text)
	}
	if !query.IsEmpty() {
		t.Error("expected IsEmpty() to return true for empty query")
	}
}

func TestParseSearchQuery_FreeText(t *testing.T) {
	query := ParseSearchQuery("hello world")

	if query.Raw != "hello world" {
		t.Errorf("expected Raw='hello world', got %q", query.Raw)
	}
	if query.Text != "hello world" {
		t.Errorf("expected Text='hello world', got %q", query.Text)
	}
	if query.IsEmpty() {
		t.Error("expected IsEmpty() to return false")
	}
}

func TestParseSearchQuery_FromOperator(t *testing.T) {
	query := ParseSearchQuery("from:john@example.com")

	if len(query.From) != 1 {
		t.Fatalf("expected 1 From value, got %d", len(query.From))
	}
	if query.From[0] != "john@example.com" {
		t.Errorf("expected From[0]='john@example.com', got %q", query.From[0])
	}
	if query.Text != "" {
		t.Errorf("expected empty Text, got %q", query.Text)
	}
}

func TestParseSearchQuery_ToOperator(t *testing.T) {
	query := ParseSearchQuery("to:jane@example.com")

	if len(query.To) != 1 {
		t.Fatalf("expected 1 To value, got %d", len(query.To))
	}
	if query.To[0] != "jane@example.com" {
		t.Errorf("expected To[0]='jane@example.com', got %q", query.To[0])
	}
}

func TestParseSearchQuery_SubjectOperator(t *testing.T) {
	query := ParseSearchQuery("subject:meeting")

	if len(query.Subject) != 1 {
		t.Fatalf("expected 1 Subject value, got %d", len(query.Subject))
	}
	if query.Subject[0] != "meeting" {
		t.Errorf("expected Subject[0]='meeting', got %q", query.Subject[0])
	}
}

func TestParseSearchQuery_QuotedValue(t *testing.T) {
	query := ParseSearchQuery(`subject:"team meeting"`)

	if len(query.Subject) != 1 {
		t.Fatalf("expected 1 Subject value, got %d", len(query.Subject))
	}
	if query.Subject[0] != "team meeting" {
		t.Errorf("expected Subject[0]='team meeting', got %q", query.Subject[0])
	}
}

func TestParseSearchQuery_MultipleOperators(t *testing.T) {
	query := ParseSearchQuery("from:john to:jane subject:hello")

	if len(query.From) != 1 || query.From[0] != "john" {
		t.Errorf("expected From=[john], got %v", query.From)
	}
	if len(query.To) != 1 || query.To[0] != "jane" {
		t.Errorf("expected To=[jane], got %v", query.To)
	}
	if len(query.Subject) != 1 || query.Subject[0] != "hello" {
		t.Errorf("expected Subject=[hello], got %v", query.Subject)
	}
}

func TestParseSearchQuery_IsOperator(t *testing.T) {
	query := ParseSearchQuery("is:unread is:starred")

	if len(query.Is) != 2 {
		t.Fatalf("expected 2 Is values, got %d", len(query.Is))
	}
	if query.Is[0] != "unread" {
		t.Errorf("expected Is[0]='unread', got %q", query.Is[0])
	}
	if query.Is[1] != "starred" {
		t.Errorf("expected Is[1]='starred', got %q", query.Is[1])
	}
}

func TestParseSearchQuery_HasOperator(t *testing.T) {
	query := ParseSearchQuery("has:attachment")

	if len(query.Has) != 1 {
		t.Fatalf("expected 1 Has value, got %d", len(query.Has))
	}
	if query.Has[0] != "attachment" {
		t.Errorf("expected Has[0]='attachment', got %q", query.Has[0])
	}
}

func TestParseSearchQuery_DateOperators(t *testing.T) {
	query := ParseSearchQuery("after:2025-01-01 before:2025-12-31")

	if query.After != "2025-01-01" {
		t.Errorf("expected After='2025-01-01', got %q", query.After)
	}
	if query.Before != "2025-12-31" {
		t.Errorf("expected Before='2025-12-31', got %q", query.Before)
	}
}

func TestParseSearchQuery_InOperator(t *testing.T) {
	query := ParseSearchQuery("in:inbox")

	if len(query.In) != 1 {
		t.Fatalf("expected 1 In value, got %d", len(query.In))
	}
	if query.In[0] != "inbox" {
		t.Errorf("expected In[0]='inbox', got %q", query.In[0])
	}
}

func TestParseSearchQuery_MixedWithFreeText(t *testing.T) {
	query := ParseSearchQuery("from:john meeting notes")

	if len(query.From) != 1 || query.From[0] != "john" {
		t.Errorf("expected From=[john], got %v", query.From)
	}
	if query.Text != "meeting notes" {
		t.Errorf("expected Text='meeting notes', got %q", query.Text)
	}
}

func TestParseSearchQuery_CaseInsensitiveOperators(t *testing.T) {
	query := ParseSearchQuery("FROM:john TO:jane SUBJECT:hello")

	if len(query.From) != 1 || query.From[0] != "john" {
		t.Errorf("expected From=[john], got %v", query.From)
	}
	if len(query.To) != 1 || query.To[0] != "jane" {
		t.Errorf("expected To=[jane], got %v", query.To)
	}
	if len(query.Subject) != 1 || query.Subject[0] != "hello" {
		t.Errorf("expected Subject=[hello], got %v", query.Subject)
	}
}

func TestSearchQuery_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		query    *SearchQuery
		expected bool
	}{
		{
			name:     "empty query",
			query:    &SearchQuery{},
			expected: true,
		},
		{
			name:     "with text",
			query:    &SearchQuery{Text: "hello"},
			expected: false,
		},
		{
			name:     "with from",
			query:    &SearchQuery{From: []string{"john"}},
			expected: false,
		},
		{
			name:     "with after",
			query:    &SearchQuery{After: "2025-01-01"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestSearchQuery_ToNativeQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    *SearchQuery
		expected string
	}{
		{
			name:     "empty query",
			query:    &SearchQuery{},
			expected: "",
		},
		{
			name:     "text only",
			query:    &SearchQuery{Text: "hello world"},
			expected: "hello world",
		},
		{
			name:     "from only",
			query:    &SearchQuery{From: []string{"john@example.com"}},
			expected: "from:john@example.com",
		},
		{
			name:     "multiple operators",
			query:    &SearchQuery{From: []string{"john"}, To: []string{"jane"}, Subject: []string{"meeting"}},
			expected: "from:john to:jane subject:meeting",
		},
		{
			name:     "with dates",
			query:    &SearchQuery{After: "2025-01-01", Before: "2025-12-31"},
			expected: "after:2025-01-01 before:2025-12-31",
		},
		{
			name:     "operators with text",
			query:    &SearchQuery{From: []string{"john"}, Text: "meeting notes"},
			expected: "from:john meeting notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.ToNativeQuery(); got != tt.expected {
				t.Errorf("ToNativeQuery() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestSearchQuery_HighlightMatches(t *testing.T) {
	// Create a simple highlight style for testing
	highlightStyle := lipgloss.NewStyle().Bold(true)

	tests := []struct {
		name       string
		query      *SearchQuery
		text       string
		field      string
		wantChange bool // Whether the text should change (have highlighting)
	}{
		{
			name:       "empty query does not highlight",
			query:      &SearchQuery{},
			text:       "Hello World",
			field:      "subject",
			wantChange: false,
		},
		{
			name:       "empty text returns empty",
			query:      &SearchQuery{Text: "hello"},
			text:       "",
			field:      "subject",
			wantChange: false,
		},
		{
			name:       "free text matches in subject",
			query:      &SearchQuery{Text: "hello"},
			text:       "Hello World",
			field:      "subject",
			wantChange: true,
		},
		{
			name:       "from operator matches from field",
			query:      &SearchQuery{From: []string{"john"}},
			text:       "John Doe",
			field:      "from",
			wantChange: true,
		},
		{
			name:       "from operator does not match subject field directly",
			query:      &SearchQuery{From: []string{"john"}},
			text:       "Meeting notes",
			field:      "subject",
			wantChange: false,
		},
		{
			name:       "subject operator matches subject field",
			query:      &SearchQuery{Subject: []string{"meeting"}},
			text:       "Meeting Notes",
			field:      "subject",
			wantChange: true,
		},
		{
			name:       "multiple words in free text",
			query:      &SearchQuery{Text: "hello world"},
			text:       "Hello beautiful World",
			field:      "subject",
			wantChange: true,
		},
		{
			name:       "case insensitive matching",
			query:      &SearchQuery{Text: "HELLO"},
			text:       "hello world",
			field:      "subject",
			wantChange: true,
		},
		{
			name:       "to operator matches to field",
			query:      &SearchQuery{To: []string{"jane"}},
			text:       "Jane Smith",
			field:      "to",
			wantChange: true,
		},
		{
			name:       "free text also matches from field",
			query:      &SearchQuery{Text: "john"},
			text:       "John Doe",
			field:      "from",
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.query.HighlightMatches(tt.text, tt.field, highlightStyle)

			if tt.wantChange {
				if result == tt.text {
					t.Errorf("expected text to be highlighted, but got unchanged: %q", result)
				}
				// Verify the result contains ANSI escape sequences (from lipgloss styling)
				if !strings.Contains(result, "\x1b[") {
					t.Errorf("expected ANSI escape sequences in result, got: %q", result)
				}
			} else {
				if result != tt.text {
					t.Errorf("expected text unchanged, got %q, want %q", result, tt.text)
				}
			}
		})
	}
}

func TestSearchQuery_HighlightMatches_SpecialCharacters(t *testing.T) {
	highlightStyle := lipgloss.NewStyle().Bold(true)

	// Test that regex special characters are properly escaped
	query := &SearchQuery{Text: "test.email@example.com"}
	result := query.HighlightMatches("test.email@example.com", "subject", highlightStyle)

	// The text should be highlighted (changed)
	if result == "test.email@example.com" {
		t.Error("expected text to be highlighted with special characters")
	}
}

func TestSearchQuery_HighlightMatches_MultipleMatches(t *testing.T) {
	highlightStyle := lipgloss.NewStyle().Bold(true)

	query := &SearchQuery{Text: "test"}
	result := query.HighlightMatches("test one test two test", "subject", highlightStyle)

	// Should have multiple highlights (count ANSI sequences)
	count := strings.Count(result, "\x1b[")
	// Each match has opening and closing sequences, so at least 6 for 3 matches
	if count < 6 {
		t.Errorf("expected multiple highlights (at least 6 escape sequences), got %d in: %q", count, result)
	}
}
