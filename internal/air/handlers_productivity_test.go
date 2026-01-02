//go:build !integration
// +build !integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =============================================================================
// Productivity Feature Tests (Snooze, Scheduled Send, Templates)
// =============================================================================

// TestHandleScheduledSend_GET tests listing scheduled sends.
func TestHandleScheduledSend_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/scheduled", nil)
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleScheduledSend_POST_Create tests creating a scheduled send.
func TestHandleScheduledSend_POST_Create(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	futureTime := time.Now().Add(2 * time.Hour).Unix()
	body := map[string]any{
		"to":      []map[string]string{{"email": "test@example.com", "name": "Test User"}},
		"subject": "Test Scheduled Email",
		"body":    "<p>Test body</p>",
		"send_at": futureTime,
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("expected status 200 or 201, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTemplates_GET tests listing email templates.
func TestHandleTemplates_GET(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	// Add test templates
	server.emailTemplates["tmpl-001"] = EmailTemplate{
		ID:        "tmpl-001",
		Name:      "Test Template",
		Subject:   "Test Subject",
		Body:      "<p>Test body</p>",
		CreatedAt: time.Now().Unix(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleTemplates_POST_Create tests creating a template.
func TestHandleTemplates_POST_Create(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	template := map[string]string{
		"name":    "New Template",
		"subject": "Template Subject",
		"body":    "<p>Template body content</p>",
	}
	body, _ := json.Marshal(template)

	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("expected status 200 or 201, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTemplateByID_GET tests retrieving a template by ID.
func TestHandleTemplateByID_GET(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	server.emailTemplates["tmpl-001"] = EmailTemplate{
		ID:      "tmpl-001",
		Name:    "Test",
		Subject: "Subject",
		Body:    "Body",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/tmpl-001", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleTemplateByID_DELETE tests deleting a template.
func TestHandleTemplateByID_DELETE(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	server.emailTemplates["tmpl-001"] = EmailTemplate{
		ID:   "tmpl-001",
		Name: "To Delete",
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/templates/tmpl-001", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify deletion
	if _, exists := server.emailTemplates["tmpl-001"]; exists {
		t.Error("expected template to be deleted")
	}
}

// TestHandleUndoSend_GET tests getting undo send config.
func TestHandleUndoSend_GET(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode: true,
		undoSendConfig: &UndoSendConfig{
			Enabled:        true,
			GracePeriodSec: 10,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/undo-send", nil)
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["enabled"] != true {
		t.Error("expected enabled to be true")
	}
}

// TestHandleUndoSend_PUT tests updating undo send config.
func TestHandleUndoSend_PUT(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode: true,
		undoSendConfig: &UndoSendConfig{
			Enabled:        false,
			GracePeriodSec: 5,
		},
	}

	update := map[string]any{
		"enabled":      true,
		"grace_period": 15,
	}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest(http.MethodPut, "/api/undo-send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandlePendingSends tests listing pending sends.
func TestHandlePendingSends(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	// Add a pending send
	server.pendingSends["send-001"] = PendingSend{
		ID:     "send-001",
		SendAt: time.Now().Add(10 * time.Second).Unix(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pending-sends", nil)
	w := httptest.NewRecorder()

	server.handlePendingSends(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestSplitInboxConfig tests split inbox configuration.
func TestSplitInboxConfig(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode: true,
		splitInboxConfig: &SplitInboxConfig{
			Enabled: true,
			VIPSenders: []string{
				"ceo@company.com",
				"important@example.com",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/inbox/split", nil)
	w := httptest.NewRecorder()

	server.handleSplitInbox(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	config, ok := resp["config"].(map[string]any)
	if !ok {
		t.Fatal("expected config in response")
	}
	if config["enabled"] != true {
		t.Error("expected enabled to be true")
	}
}

// TestHandleVIPSenders_GET tests getting VIP senders.
func TestHandleVIPSenders_GET(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode: true,
		splitInboxConfig: &SplitInboxConfig{
			Enabled: true,
			VIPSenders: []string{
				"vip1@example.com",
				"vip2@example.com",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/inbox/vip", nil)
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleVIPSenders_POST tests adding a VIP sender.
func TestHandleVIPSenders_POST(t *testing.T) {
	t.Parallel()

	server := &Server{
		demoMode: true,
		splitInboxConfig: &SplitInboxConfig{
			Enabled:    true,
			VIPSenders: []string{},
		},
	}

	body := map[string]string{
		"email": "newvip@example.com",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/inbox/vip", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestSnoozedEmail_Types tests snoozed email type structure.
func TestSnoozedEmail_Types(t *testing.T) {
	t.Parallel()

	now := time.Now()
	snooze := SnoozedEmail{
		EmailID:     "email-001",
		SnoozeUntil: now.Add(2 * time.Hour).Unix(),
		CreatedAt:   now.Unix(),
	}

	if snooze.EmailID != "email-001" {
		t.Errorf("expected EmailID email-001, got %s", snooze.EmailID)
	}

	if snooze.SnoozeUntil <= now.Unix() {
		t.Error("expected SnoozeUntil to be in the future")
	}
}

// TestPendingSend_Types tests pending send type structure.
func TestPendingSend_Types(t *testing.T) {
	t.Parallel()

	now := time.Now()
	pending := PendingSend{
		ID:     "pending-001",
		SendAt: now.Add(10 * time.Second).Unix(),
	}

	if pending.ID != "pending-001" {
		t.Errorf("expected ID pending-001, got %s", pending.ID)
	}

	if pending.SendAt <= now.Unix() {
		t.Error("expected SendAt to be in the future")
	}
}

// TestEmailTemplate_Types tests email template type structure.
func TestEmailTemplate_Types(t *testing.T) {
	t.Parallel()

	now := time.Now()
	template := EmailTemplate{
		ID:        "tmpl-001",
		Name:      "Welcome Email",
		Subject:   "Welcome to our platform",
		Body:      "<h1>Welcome!</h1><p>Thank you for joining.</p>",
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
	}

	if template.ID != "tmpl-001" {
		t.Errorf("expected ID tmpl-001, got %s", template.ID)
	}

	if template.Name != "Welcome Email" {
		t.Errorf("expected Name 'Welcome Email', got %s", template.Name)
	}

	if template.Subject != "Welcome to our platform" {
		t.Error("expected correct subject")
	}
}

// TestUndoSendConfig_Types tests undo send config type structure.
func TestUndoSendConfig_Types(t *testing.T) {
	t.Parallel()

	config := UndoSendConfig{
		Enabled:        true,
		GracePeriodSec: 10,
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}

	if config.GracePeriodSec != 10 {
		t.Errorf("expected GracePeriodSec 10, got %d", config.GracePeriodSec)
	}
}

// TestSplitInboxConfig_Types tests split inbox config type structure.
func TestSplitInboxConfig_Types(t *testing.T) {
	t.Parallel()

	config := SplitInboxConfig{
		Enabled: true,
		VIPSenders: []string{
			"ceo@company.com",
			"manager@company.com",
		},
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}

	if len(config.VIPSenders) != 2 {
		t.Errorf("expected 2 VIP senders, got %d", len(config.VIPSenders))
	}
}
