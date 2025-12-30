package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

func TestHandleGetEmail_DemoMode_AllDemoEmails(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()
	// These IDs match demoEmails() function in handlers.go
	demoIDs := []string{"demo-email-001", "demo-email-002", "demo-email-003", "demo-email-004", "demo-email-005"}

	for _, id := range demoIDs {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/emails/"+id, nil)
			w := httptest.NewRecorder()

			server.handleGetEmail(w, req, id)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for %s, got %d", id, w.Code)
			}
		})
	}
}

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

// ================================
// EMAIL RESPONSE CONVERTER TESTS
// ================================

func TestEmailToResponse_Basic(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID:       "msg-123",
		ThreadID: "thread-456",
		Subject:  "Test Subject",
		Snippet:  "This is a snippet...",
		Body:     "<p>Full body content</p>",
		Unread:   true,
		Starred:  false,
		Folders:  []string{"INBOX"},
	}

	resp := emailToResponse(msg, false)

	if resp.ID != "msg-123" {
		t.Errorf("expected ID 'msg-123', got %s", resp.ID)
	}
	if resp.ThreadID != "thread-456" {
		t.Errorf("expected ThreadID 'thread-456', got %s", resp.ThreadID)
	}
	if resp.Subject != "Test Subject" {
		t.Errorf("expected Subject 'Test Subject', got %s", resp.Subject)
	}
	if resp.Snippet != "This is a snippet..." {
		t.Errorf("expected Snippet to match, got %s", resp.Snippet)
	}
	if resp.Body != "" {
		t.Error("expected Body to be empty when includeBody=false")
	}
	if !resp.Unread {
		t.Error("expected Unread to be true")
	}
	if resp.Starred {
		t.Error("expected Starred to be false")
	}
}

func TestEmailToResponse_WithBody(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID:   "msg-123",
		Body: "<p>Full body content</p>",
	}

	resp := emailToResponse(msg, true)

	if resp.Body != "<p>Full body content</p>" {
		t.Errorf("expected Body to be included, got %s", resp.Body)
	}
}

func TestEmailToResponse_WithParticipants(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID: "msg-123",
		From: []domain.EmailParticipant{
			{Name: "Sender Name", Email: "sender@example.com"},
		},
		To: []domain.EmailParticipant{
			{Name: "Recipient One", Email: "recipient1@example.com"},
			{Name: "Recipient Two", Email: "recipient2@example.com"},
		},
		Cc: []domain.EmailParticipant{
			{Name: "CC Person", Email: "cc@example.com"},
		},
	}

	resp := emailToResponse(msg, false)

	if len(resp.From) != 1 {
		t.Errorf("expected 1 From participant, got %d", len(resp.From))
	}
	if resp.From[0].Email != "sender@example.com" {
		t.Errorf("expected From email 'sender@example.com', got %s", resp.From[0].Email)
	}

	if len(resp.To) != 2 {
		t.Errorf("expected 2 To participants, got %d", len(resp.To))
	}

	if len(resp.Cc) != 1 {
		t.Errorf("expected 1 Cc participant, got %d", len(resp.Cc))
	}
}

func TestEmailToResponse_WithAttachments(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID: "msg-123",
		Attachments: []domain.Attachment{
			{ID: "att-1", Filename: "document.pdf", ContentType: "application/pdf", Size: 1024},
			{ID: "att-2", Filename: "image.png", ContentType: "image/png", Size: 2048},
		},
	}

	resp := emailToResponse(msg, false)

	if len(resp.Attachments) != 2 {
		t.Errorf("expected 2 attachments, got %d", len(resp.Attachments))
	}

	if resp.Attachments[0].Filename != "document.pdf" {
		t.Errorf("expected first attachment filename 'document.pdf', got %s", resp.Attachments[0].Filename)
	}
	if resp.Attachments[1].Size != 2048 {
		t.Errorf("expected second attachment size 2048, got %d", resp.Attachments[1].Size)
	}
}

func TestCachedEmailToResponse(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	cachedEmail := &cache.CachedEmail{
		ID:             "email-123",
		ThreadID:       "thread-456",
		FolderID:       "inbox",
		Subject:        "Test Email Subject",
		Snippet:        "This is a test snippet...",
		FromName:       "John Doe",
		FromEmail:      "john@example.com",
		To:             []string{"recipient@example.com"},
		Date:           testTime,
		Unread:         true,
		Starred:        false,
		HasAttachments: true,
	}

	resp := cachedEmailToResponse(cachedEmail)

	if resp.ID != "email-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "email-123")
	}
	if resp.ThreadID != "thread-456" {
		t.Errorf("ThreadID = %q, want %q", resp.ThreadID, "thread-456")
	}
	if resp.Subject != "Test Email Subject" {
		t.Errorf("Subject = %q, want %q", resp.Subject, "Test Email Subject")
	}
	if resp.Snippet != "This is a test snippet..." {
		t.Errorf("Snippet = %q, want %q", resp.Snippet, "This is a test snippet...")
	}
	if len(resp.From) != 1 || resp.From[0].Name != "John Doe" || resp.From[0].Email != "john@example.com" {
		t.Errorf("From = %+v, want [{John Doe john@example.com}]", resp.From)
	}
	if resp.Date != testTime.Unix() {
		t.Errorf("Date = %d, want %d", resp.Date, testTime.Unix())
	}
	if !resp.Unread {
		t.Error("Unread should be true")
	}
	if resp.Starred {
		t.Error("Starred should be false")
	}
	if len(resp.Folders) != 1 || resp.Folders[0] != "inbox" {
		t.Errorf("Folders = %v, want [inbox]", resp.Folders)
	}
}

func TestCachedEmailToResponse_EmptyFields(t *testing.T) {
	t.Parallel()

	cachedEmail := &cache.CachedEmail{
		ID:   "email-empty",
		Date: time.Time{},
	}

	resp := cachedEmailToResponse(cachedEmail)

	if resp.ID != "email-empty" {
		t.Errorf("ID = %q, want %q", resp.ID, "email-empty")
	}
	if resp.ThreadID != "" {
		t.Errorf("ThreadID should be empty, got %q", resp.ThreadID)
	}
	if len(resp.From) != 1 {
		t.Error("From should have one entry even with empty values")
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
