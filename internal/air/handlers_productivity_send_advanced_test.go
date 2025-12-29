//go:build integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
