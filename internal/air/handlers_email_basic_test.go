package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ================================
// EMAILS HANDLER TESTS
// ================================

func TestHandleListEmails_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Emails) == 0 {
		t.Error("expected non-empty emails")
	}

	// Check first email has expected fields
	if resp.Emails[0].ID == "" {
		t.Error("expected email to have ID")
	}

	if resp.Emails[0].Subject == "" {
		t.Error("expected email to have Subject")
	}

	if len(resp.Emails[0].From) == 0 {
		t.Error("expected email to have From")
	}
}

func TestHandleGetEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "demo-email-001" {
		t.Errorf("expected ID 'demo-email-001', got %s", resp.ID)
	}
}

func TestHandleGetEmail_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"unread": false, "starred": true}`
	req := httptest.NewRequest(http.MethodPut, "/api/emails/demo-email-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp UpdateEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
}

func TestHandleDeleteEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp UpdateEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
}

func TestHandleListEmails_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with folder filter
	req := httptest.NewRequest(http.MethodGet, "/api/emails?folder=inbox&limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Emails) == 0 {
		t.Error("expected emails in response")
	}
}

func TestHandleListEmails_WithPagination(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with page token
	req := httptest.NewRequest(http.MethodGet, "/api/emails?page_token=test", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/emails", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEmailByID_WithInvalidMethod(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEmailByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/emails/demo-email-001", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateEmail(w, req, "demo-email-001")

	// Demo mode might still succeed or return bad request
	// Either is acceptable for invalid JSON
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleDeleteEmail_DemoMode_Additional(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleDeleteEmail(w, req, "demo-email-001")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Error("expected success to be true")
	}
}

func TestHandleGetEmail_DemoMode_Found(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get a known demo email ID (matches demoEmails() function)
	req := httptest.NewRequest(http.MethodGet, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleGetEmail(w, req, "demo-email-001")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "demo-email-001" {
		t.Errorf("expected ID 'demo-email-001', got %s", resp.ID)
	}
}

func TestHandleGetEmail_DemoMode_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/nonexistent-id", nil)
	w := httptest.NewRecorder()

	server.handleGetEmail(w, req, "nonexistent-id")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
