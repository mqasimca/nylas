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
// Scheduled Send Tests
// =============================================================================

func TestHandleScheduledSend_List_DemoMode(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/scheduled", nil)
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	scheduled := resp["scheduled"].([]any)
	if len(scheduled) < 1 {
		t.Error("expected at least one demo scheduled message")
	}
}

func TestHandleScheduledSend_Create_DemoMode(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(ScheduledSendRequest{
		To:            []EmailParticipantResponse{{Email: "test@example.com", Name: "Test"}},
		Subject:       "Test Subject",
		Body:          "Test body",
		SendAtNatural: "tomorrow 9am",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ScheduledSendResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.ScheduleID == "" {
		t.Error("expected schedule ID to be set")
	}
}

func TestHandleScheduledSend_CreateWithTimestamp(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	futureTime := time.Now().Add(2 * time.Hour).Unix()
	body, _ := json.Marshal(ScheduledSendRequest{
		To:      []EmailParticipantResponse{{Email: "test@example.com"}},
		Subject: "Test",
		Body:    "Test",
		SendAt:  futureTime,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleScheduledSend_NoTime(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(ScheduledSendRequest{
		To:      []EmailParticipantResponse{{Email: "test@example.com"}},
		Subject: "Test",
		Body:    "Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing send time, got %d", w.Code)
	}
}

func TestHandleScheduledSend_TooSoon(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	// Send time less than 1 minute in future
	tooSoon := time.Now().Add(30 * time.Second).Unix()
	body, _ := json.Marshal(ScheduledSendRequest{
		To:      []EmailParticipantResponse{{Email: "test@example.com"}},
		Subject: "Test",
		Body:    "Test",
		SendAt:  tooSoon,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for send time too soon, got %d", w.Code)
	}
}

func TestHandleScheduledSend_Cancel_DemoMode(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodDelete, "/api/scheduled?schedule_id=test-123", nil)
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleScheduledSend_CancelNoID(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodDelete, "/api/scheduled", nil)
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing schedule ID, got %d", w.Code)
	}
}

func TestHandleScheduledSend_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPut, "/api/scheduled", nil)
	w := httptest.NewRecorder()

	server.handleScheduledSend(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// =============================================================================
// Undo Send Tests
// =============================================================================

func TestHandleUndoSend_GetConfig(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/undo-send", nil)
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var config UndoSendConfig
	if err := json.NewDecoder(w.Body).Decode(&config); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !config.Enabled {
		t.Error("expected undo send to be enabled by default")
	}
	if config.GracePeriodSec != 10 {
		t.Errorf("expected default grace period of 10, got %d", config.GracePeriodSec)
	}
}

func TestHandleUndoSend_UpdateConfig(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(UndoSendConfig{
		Enabled:        true,
		GracePeriodSec: 30,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/undo-send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify config was updated
	config := server.getOrCreateUndoSendConfig()
	if config.GracePeriodSec != 30 {
		t.Errorf("expected grace period of 30, got %d", config.GracePeriodSec)
	}
}

func TestHandleUndoSend_ValidateGracePeriod(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	// Test minimum bound (should be 5)
	body, _ := json.Marshal(UndoSendConfig{GracePeriodSec: 2})
	req := httptest.NewRequest(http.MethodPut, "/api/undo-send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	config := server.getOrCreateUndoSendConfig()
	if config.GracePeriodSec < 5 {
		t.Errorf("grace period should be at least 5, got %d", config.GracePeriodSec)
	}

	// Test maximum bound (should be 60)
	body, _ = json.Marshal(UndoSendConfig{GracePeriodSec: 120})
	req = httptest.NewRequest(http.MethodPut, "/api/undo-send", bytes.NewReader(body))
	w = httptest.NewRecorder()

	server.handleUndoSend(w, req)

	config = server.getOrCreateUndoSendConfig()
	if config.GracePeriodSec > 60 {
		t.Errorf("grace period should be at most 60, got %d", config.GracePeriodSec)
	}
}

func TestHandleUndoSend_UndoMessage(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	// Add a pending send
	server.pendingSends["msg-123"] = PendingSend{
		ID:      "msg-123",
		Subject: "Test",
		SendAt:  time.Now().Add(time.Minute).Unix(),
	}

	body, _ := json.Marshal(map[string]string{"message_id": "msg-123"})
	req := httptest.NewRequest(http.MethodPost, "/api/undo-send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp UndoSendResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify message was cancelled
	if !server.pendingSends["msg-123"].Cancelled {
		t.Error("expected message to be marked as cancelled")
	}
}

func TestHandleUndoSend_ExpiredGracePeriod(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	// Add a pending send with expired grace period
	server.pendingSends["msg-456"] = PendingSend{
		ID:      "msg-456",
		Subject: "Test",
		SendAt:  time.Now().Add(-time.Minute).Unix(), // Already expired
	}

	body, _ := json.Marshal(map[string]string{"message_id": "msg-456"})
	req := httptest.NewRequest(http.MethodPost, "/api/undo-send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for expired grace period, got %d", w.Code)
	}
}

func TestHandleUndoSend_NotFound(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	body, _ := json.Marshal(map[string]string{"message_id": "nonexistent"})
	req := httptest.NewRequest(http.MethodPost, "/api/undo-send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleUndoSend(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for nonexistent message, got %d", w.Code)
	}
}

func TestHandlePendingSends(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:     true,
		pendingSends: make(map[string]PendingSend),
	}

	// Add some pending sends
	server.pendingSends["msg-1"] = PendingSend{
		ID:     "msg-1",
		SendAt: time.Now().Add(time.Minute).Unix(),
	}
	server.pendingSends["msg-2"] = PendingSend{
		ID:        "msg-2",
		SendAt:    time.Now().Add(time.Minute).Unix(),
		Cancelled: true, // Should not appear
	}
	server.pendingSends["msg-3"] = PendingSend{
		ID:     "msg-3",
		SendAt: time.Now().Add(-time.Minute).Unix(), // Expired, should not appear
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pending-sends", nil)
	w := httptest.NewRecorder()

	server.handlePendingSends(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := int(resp["count"].(float64))
	if count != 1 {
		t.Errorf("expected 1 pending send (non-cancelled, non-expired), got %d", count)
	}
}

func TestHandlePendingSends_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPost, "/api/pending-sends", nil)
	w := httptest.NewRecorder()

	server.handlePendingSends(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// =============================================================================
// Email Templates Tests
// =============================================================================

func TestHandleTemplates_List(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TemplateListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should return default templates
	if len(resp.Templates) < 3 {
		t.Error("expected at least 3 default templates")
	}
}

func TestHandleTemplates_Create(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	body, _ := json.Marshal(EmailTemplate{
		Name:     "My Template",
		Subject:  "Hello {{name}}",
		Body:     "Hi {{name}}, this is a test for {{company}}.",
		Shortcut: "/mytemplate",
		Category: "greeting",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var template EmailTemplate
	if err := json.NewDecoder(w.Body).Decode(&template); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if template.ID == "" {
		t.Error("expected template ID to be generated")
	}
	if len(template.Variables) != 2 {
		t.Errorf("expected 2 variables (name, company), got %d", len(template.Variables))
	}
}

func TestHandleTemplates_CreateNoName(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(EmailTemplate{Body: "Test body"})
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing name, got %d", w.Code)
	}
}

func TestHandleTemplates_CreateNoBody(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	body, _ := json.Marshal(EmailTemplate{Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing body, got %d", w.Code)
	}
}

func TestHandleTemplateByID_Get(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	// Add a template
	server.emailTemplates["tmpl-123"] = EmailTemplate{
		ID:   "tmpl-123",
		Name: "Test Template",
		Body: "Test body",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/tmpl-123", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var template EmailTemplate
	if err := json.NewDecoder(w.Body).Decode(&template); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if template.Name != "Test Template" {
		t.Errorf("expected name 'Test Template', got '%s'", template.Name)
	}
}

func TestHandleTemplateByID_GetDefault(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	// Get a default template
	req := httptest.NewRequest(http.MethodGet, "/api/templates/default-thanks", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleTemplateByID_NotFound(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleTemplateByID_Update(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	// Add a template
	server.emailTemplates["tmpl-123"] = EmailTemplate{
		ID:   "tmpl-123",
		Name: "Original",
		Body: "Original body",
	}

	body, _ := json.Marshal(EmailTemplate{Name: "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/api/templates/tmpl-123", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify update
	if server.emailTemplates["tmpl-123"].Name != "Updated" {
		t.Error("expected template name to be updated")
	}
}

func TestHandleTemplateByID_Delete(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	// Add a template
	server.emailTemplates["tmpl-123"] = EmailTemplate{ID: "tmpl-123"}

	req := httptest.NewRequest(http.MethodDelete, "/api/templates/tmpl-123", nil)
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if _, exists := server.emailTemplates["tmpl-123"]; exists {
		t.Error("expected template to be deleted")
	}
}

func TestHandleTemplateByID_Expand(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:       true,
		emailTemplates: make(map[string]EmailTemplate),
	}

	// Add a template with variables
	server.emailTemplates["tmpl-123"] = EmailTemplate{
		ID:        "tmpl-123",
		Name:      "Test",
		Subject:   "Hello {{name}}",
		Body:      "Hi {{name}}, welcome to {{company}}!",
		Variables: []string{"name", "company"},
	}

	body, _ := json.Marshal(map[string]any{
		"variables": map[string]string{
			"name":    "Alice",
			"company": "Acme Inc",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/templates/tmpl-123/expand", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["subject"] != "Hello Alice" {
		t.Errorf("expected subject 'Hello Alice', got '%s'", resp["subject"])
	}
	if resp["body"] != "Hi Alice, welcome to Acme Inc!" {
		t.Errorf("expected expanded body, got '%s'", resp["body"])
	}
}

func TestExtractTemplateVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text     string
		expected []string
	}{
		{"Hello {{name}}", []string{"name"}},
		{"{{greeting}}, {{name}}!", []string{"greeting", "name"}},
		{"No variables here", []string{}},
		{"{{name}} and {{name}} again", []string{"name"}}, // Deduplication
		{"{{a}} {{b}} {{c}}", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			vars := extractTemplateVariables(tt.text)
			if len(vars) != len(tt.expected) {
				t.Errorf("expected %d variables, got %d: %v", len(tt.expected), len(vars), vars)
			}
		})
	}
}

func TestDefaultTemplates(t *testing.T) {
	t.Parallel()

	templates := defaultTemplates()

	if len(templates) < 3 {
		t.Errorf("expected at least 3 default templates, got %d", len(templates))
	}

	// Check that all templates have required fields
	for _, tmpl := range templates {
		if tmpl.ID == "" {
			t.Error("template missing ID")
		}
		if tmpl.Name == "" {
			t.Error("template missing name")
		}
		if tmpl.Body == "" {
			t.Error("template missing body")
		}
		if tmpl.CreatedAt == 0 {
			t.Error("template missing created_at")
		}
	}
}

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodDelete, "/api/templates", nil)
	w := httptest.NewRecorder()

	server.handleTemplates(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
