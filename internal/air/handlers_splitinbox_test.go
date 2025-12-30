package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSplitInbox_Get(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/inbox/split", nil)
	w := httptest.NewRecorder()

	server.handleSplitInbox(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SplitInboxResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Config.Enabled {
		t.Error("expected split inbox to be enabled by default")
	}
	if len(resp.Config.Categories) == 0 {
		t.Error("expected default categories to be set")
	}
	if len(resp.Categories) == 0 {
		t.Error("expected category counts to be returned")
	}
}

func TestHandleSplitInbox_Put(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	config := SplitInboxConfig{
		Enabled:    true,
		Categories: []InboxCategory{CategoryPrimary, CategoryVIP},
		VIPSenders: []string{"boss@company.com"},
	}
	body, _ := json.Marshal(config)

	req := httptest.NewRequest(http.MethodPut, "/api/inbox/split", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleSplitInbox(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != true {
		t.Error("expected success to be true")
	}
}

func TestHandleSplitInbox_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodDelete, "/api/inbox/split", nil)
	w := httptest.NewRecorder()

	server.handleSplitInbox(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCategorizeEmail(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	testCases := []struct {
		name    string
		from    string
		subject string
		wantCat InboxCategory
	}{
		{"newsletter", "newsletter@example.com", "Weekly Update", CategoryNewsletters},
		{"social", "notifications@linkedin.com", "New connection", CategorySocial},
		{"updates", "receipt@stripe.com", "Payment received", CategoryUpdates},
		{"promotions", "deals@store.com", "50% off sale", CategoryPromotions},
		{"primary", "friend@gmail.com", "Hey there!", CategoryPrimary},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"email_id": "test-123",
				"from":     tc.from,
				"subject":  tc.subject,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/inbox/categorize", bytes.NewReader(body))
			w := httptest.NewRecorder()

			server.handleCategorizeEmail(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			var resp CategorizedEmail
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Category != tc.wantCat {
				t.Errorf("expected category %s, got %s", tc.wantCat, resp.Category)
			}
		})
	}
}

func TestHandleCategorizeEmail_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/inbox/categorize", nil)
	w := httptest.NewRecorder()

	server.handleCategorizeEmail(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
