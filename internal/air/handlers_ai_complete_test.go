package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseNaturalLanguageSearch(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected NLSearchResponse
	}{
		{
			name:  "from pattern",
			query: "emails from john",
			expected: NLSearchResponse{
				From: "john",
			},
		},
		{
			name:  "to pattern",
			query: "emails to sarah about project",
			expected: NLSearchResponse{
				To:       "sarah",
				Keywords: "project",
			},
		},
		{
			name:  "last week",
			query: "emails from last week",
			expected: NLSearchResponse{
				DateAfter: "7d",
			},
		},
		{
			name:  "yesterday",
			query: "messages from yesterday",
			expected: NLSearchResponse{
				DateAfter: "1d",
			},
		},
		{
			name:  "today",
			query: "emails from today",
			expected: NLSearchResponse{
				DateAfter: "0d",
			},
		},
		{
			name:  "this month",
			query: "all emails this month",
			expected: NLSearchResponse{
				DateAfter: "30d",
			},
		},
		{
			name:  "with attachment",
			query: "emails with attachment",
			expected: NLSearchResponse{
				HasAttach: true,
			},
		},
		{
			name:  "attached files",
			query: "messages with attached files",
			expected: NLSearchResponse{
				HasAttach: true,
				Keywords:  "files",
			},
		},
		{
			name:  "unread",
			query: "unread emails",
			expected: NLSearchResponse{
				IsUnread: true,
			},
		},
		{
			name:  "complex query",
			query: "unread emails from john last week about project",
			expected: NLSearchResponse{
				From:      "john",
				DateAfter: "7d",
				IsUnread:  true,
				Keywords:  "project",
			},
		},
		{
			name:  "keywords only",
			query: "invoice quarterly report",
			expected: NLSearchResponse{
				Keywords: "invoice quarterly report",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNaturalLanguageSearch(tt.query)

			if result.From != tt.expected.From {
				t.Errorf("From = %q, want %q", result.From, tt.expected.From)
			}
			if result.To != tt.expected.To {
				t.Errorf("To = %q, want %q", result.To, tt.expected.To)
			}
			if result.DateAfter != tt.expected.DateAfter {
				t.Errorf("DateAfter = %q, want %q", result.DateAfter, tt.expected.DateAfter)
			}
			if result.HasAttach != tt.expected.HasAttach {
				t.Errorf("HasAttach = %v, want %v", result.HasAttach, tt.expected.HasAttach)
			}
			if result.IsUnread != tt.expected.IsUnread {
				t.Errorf("IsUnread = %v, want %v", result.IsUnread, tt.expected.IsUnread)
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		query    string
		expected []string
	}{
		{
			query:    "project update meeting",
			expected: []string{"project", "update", "meeting"},
		},
		{
			query:    "the invoice from last week",
			expected: []string{"invoice"},
		},
		{
			query:    "emails about quarterly report",
			expected: []string{"quarterly", "report"},
		},
		{
			query:    "a an the and or",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := extractKeywords(tt.query)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d keywords, want %d", len(result), len(tt.expected))
				return
			}

			for i, kw := range result {
				if kw != tt.expected[i] {
					t.Errorf("keyword[%d] = %q, want %q", i, kw, tt.expected[i])
				}
			}
		})
	}
}

func TestHandleAICompleteEmptyText(t *testing.T) {
	server := &Server{}

	body := CompleteRequest{Text: "", MaxLength: 100}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/complete", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIComplete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result CompleteResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Suggestion != "" {
		t.Errorf("expected empty suggestion for empty text")
	}
}

func TestHandleAICompleteInvalidBody(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/ai/complete", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIComplete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleNLSearch(t *testing.T) {
	server := &Server{}

	body := NLSearchRequest{Query: "emails from john last week"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/nl-search", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleNLSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result NLSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.From != "john" {
		t.Errorf("expected From 'john', got %q", result.From)
	}

	if result.DateAfter != "7d" {
		t.Errorf("expected DateAfter '7d', got %q", result.DateAfter)
	}
}

func TestHandleNLSearchEmptyQuery(t *testing.T) {
	server := &Server{}

	body := NLSearchRequest{Query: ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/nl-search", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleNLSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleNLSearchInvalidBody(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/ai/nl-search", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleNLSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestBuildCompletionPrompt(t *testing.T) {
	text := "Hello, I wanted to follow up on"
	maxLen := 50

	prompt := buildCompletionPrompt(text, maxLen)

	if prompt == "" {
		t.Error("expected non-empty prompt")
	}

	if len(prompt) < len(text) {
		t.Error("prompt should include the input text")
	}
}
