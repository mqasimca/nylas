package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================================
// EMAIL HANDLER ADDITIONAL TESTS
// ================================

func TestHandleListEmails_FolderFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?folder=INBOX", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_UnreadFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?unread=true", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_StarredFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?starred=true", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_CategoryFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?category=primary", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_LimitParam(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_InvalidLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with invalid limit (should use default)
	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=abc", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_ExcessiveLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with limit > 200 (should cap at 200)
	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=500", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_CombinedFilters(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails?folder=INBOX&unread=true&limit=25", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEmailByID_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get an email from demo data
	emails := demoEmails()
	if len(emails) == 0 {
		t.Skip("no demo emails available")
	}

	emailID := emails[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/emails/"+emailID, nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEmailByID_PUT(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"unread":  false,
		"starred": true,
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/emails/email-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEmailByID_DELETE(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/emails/email-1", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEmailByID_PATCH_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPatch, "/api/emails/email-1", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_MarkRead(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"unread": false,
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/emails/email-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_Star(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"starred": true,
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/emails/email-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_MoveToFolder(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"folders": []string{"folder-archive"},
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/emails/email-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/emails/email-1", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	// Should handle empty body gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	msg := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com", "name": "Recipient"},
		},
		"subject": "Test Subject",
		"body":    "Test body content",
	}
	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := bytes.NewBufferString("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	// Demo mode may handle invalid JSON gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", nil)
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	// Demo mode may handle empty body gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_WithCC(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	msg := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"cc": []map[string]string{
			{"email": "cc@example.com"},
		},
		"subject": "Test",
		"body":    "Body",
	}
	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_WithBCC(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	msg := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"bcc": []map[string]string{
			{"email": "bcc@example.com"},
		},
		"subject": "Test",
		"body":    "Body",
	}
	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendMessage_WithReplyTo(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	msg := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"subject":             "Re: Original Subject",
		"body":                "Reply content",
		"reply_to_message_id": "original-message-id",
	}
	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "/api/messages/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEmailResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := EmailResponse{
		ID:      "email-123",
		Subject: "Test Subject",
		Body:    "Test Body",
		From: []EmailParticipantResponse{
			{Email: "sender@example.com", Name: "Sender"},
		},
		To: []EmailParticipantResponse{
			{Email: "recipient@example.com", Name: "Recipient"},
		},
		Unread:  true,
		Starred: false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EmailResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("expected ID %s, got %s", resp.ID, decoded.ID)
	}

	if decoded.Subject != resp.Subject {
		t.Errorf("expected Subject %s, got %s", resp.Subject, decoded.Subject)
	}

	if len(decoded.From) != 1 {
		t.Errorf("expected 1 From, got %d", len(decoded.From))
	}

	if len(decoded.To) != 1 {
		t.Errorf("expected 1 To, got %d", len(decoded.To))
	}
}

func TestEmailActionResponseMap(t *testing.T) {
	t.Parallel()

	resp := map[string]any{
		"success": true,
		"message": "Email sent successfully",
		"email": map[string]any{
			"id":      "email-new",
			"subject": "Test",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["success"] != true {
		t.Error("expected success to be true")
	}

	if decoded["email"] == nil {
		t.Error("expected email to be present")
	}
}

func TestHandleListFolders_DemoModeContent(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FoldersResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify folders have required fields
	for _, folder := range resp.Folders {
		if folder.ID == "" {
			t.Error("expected folder to have ID")
		}
		if folder.Name == "" {
			t.Error("expected folder to have Name")
		}
	}
}

func TestHandleListFolders_POST_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
