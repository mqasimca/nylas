//go:build integration
// +build integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// EMAILS INTEGRATION TESTS
// ================================

func TestIntegration_ListEmails(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=5", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d emails (has_more: %v)", len(resp.Emails), resp.HasMore)

	if len(resp.Emails) == 0 {
		t.Log("Warning: no emails found in account")
		return
	}

	// Verify email structure
	first := resp.Emails[0]
	if first.ID == "" {
		t.Error("expected email to have ID")
	}
	if len(first.From) == 0 {
		t.Error("expected email to have From")
	}
	if first.Date == 0 {
		t.Error("expected email to have Date")
	}

	t.Logf("First email: %s from %s", first.Subject, first.From[0].Email)
}

func TestIntegration_ListEmails_WithFilters(t *testing.T) {
	server := testServer(t)

	// Test unread filter
	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=5&unread=true", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d unread emails", len(resp.Emails))

	// All returned emails should be unread
	for _, email := range resp.Emails {
		if !email.Unread {
			t.Errorf("expected email %s to be unread", email.ID)
		}
	}
}

func TestIntegration_GetEmail(t *testing.T) {
	server := testServer(t)

	// First get a list of emails to get a valid ID
	listReq := httptest.NewRequest(http.MethodGet, "/api/emails?limit=1", nil)
	listW := httptest.NewRecorder()
	server.handleListEmails(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Skipf("Skipping: cannot list emails: %s", listW.Body.String())
	}

	var listResp EmailsResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(listResp.Emails) == 0 {
		t.Skip("Skipping: no emails in account to test")
	}

	emailID := listResp.Emails[0].ID

	// Now get the specific email
	req := httptest.NewRequest(http.MethodGet, "/api/emails/"+emailID, nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != emailID {
		t.Errorf("expected ID %s, got %s", emailID, resp.ID)
	}

	// Full email should have body
	if resp.Body == "" {
		t.Log("Warning: email body is empty")
	}

	t.Logf("Got email: %s (body length: %d)", resp.Subject, len(resp.Body))
}

// ================================
// DRAFTS INTEGRATION TESTS
// ================================

func TestIntegration_ListDrafts(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp DraftsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d drafts", len(resp.Drafts))

	// Drafts are optional, so just verify the response structure
	for _, draft := range resp.Drafts {
		if draft.ID == "" {
			t.Error("expected draft to have ID")
		}
		t.Logf("Draft: %s", draft.Subject)
	}
}

// ================================
