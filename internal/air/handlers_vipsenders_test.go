package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleVIPSenders_Get(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/inbox/vip", nil)
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := resp["vip_senders"]; !ok {
		t.Error("expected vip_senders in response")
	}
}

func TestHandleVIPSenders_Add(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(map[string]string{"email": "boss@company.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/inbox/vip", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify VIP was added
	config := server.getOrCreateSplitInboxConfig()
	found := false
	for _, vip := range config.VIPSenders {
		if vip == "boss@company.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected VIP sender to be added")
	}
}

func TestHandleVIPSenders_Remove(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	// First add a VIP
	server.addVIPSender("boss@company.com")

	// Then remove
	req := httptest.NewRequest(http.MethodDelete, "/api/inbox/vip?email=boss@company.com", nil)
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify VIP was removed
	config := server.getOrCreateSplitInboxConfig()
	for _, vip := range config.VIPSenders {
		if vip == "boss@company.com" {
			t.Error("expected VIP sender to be removed")
		}
	}
}

func TestCategorizeEmail_VIPPriority(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	// Add VIP sender
	server.addVIPSender("boss@company.com")

	// Categorize email from VIP (should override newsletter pattern)
	category, _ := server.categorizeEmail("newsletter@boss@company.com", "Newsletter", nil)
	if category != CategoryVIP {
		t.Errorf("expected VIP category for VIP sender, got %s", category)
	}
}

func TestMatchesRule(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	tests := []struct {
		name    string
		rule    CategoryRule
		from    string
		subject string
		want    bool
	}{
		{
			name:    "sender match",
			rule:    CategoryRule{Type: "sender", Pattern: "newsletter"},
			from:    "newsletter@example.com",
			subject: "",
			want:    true,
		},
		{
			name:    "subject match",
			rule:    CategoryRule{Type: "subject", Pattern: "urgent"},
			from:    "",
			subject: "urgent: please respond", // lowercase as passed by categorizeEmail
			want:    true,
		},
		{
			name:    "domain match",
			rule:    CategoryRule{Type: "domain", Pattern: "@company.com"},
			from:    "user@company.com",
			subject: "",
			want:    true,
		},
		{
			name:    "regex match",
			rule:    CategoryRule{Type: "sender", Pattern: "^no-?reply@", IsRegex: true},
			from:    "noreply@example.com",
			subject: "",
			want:    true,
		},
		{
			name:    "no match",
			rule:    CategoryRule{Type: "sender", Pattern: "xyz123"},
			from:    "user@example.com",
			subject: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.matchesRule(tt.rule, tt.from, tt.subject, nil)
			if got != tt.want {
				t.Errorf("matchesRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleVIPSenders_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPut, "/api/inbox/vip", nil)
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// =============================================================================
// Snooze Tests
// =============================================================================
