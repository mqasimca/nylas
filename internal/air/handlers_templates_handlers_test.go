//go:build !integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestServerWithTemplates() *Server {
	s := &Server{
		templatesMu:    sync.RWMutex{},
		emailTemplates: make(map[string]EmailTemplate),
	}
	return s
}

func TestHandleTemplates_List(t *testing.T) {
	s := createTestServerWithTemplates()

	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TemplateListResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Should return default templates when empty
	assert.NotEmpty(t, result.Templates)
}

func TestHandleTemplates_ListWithCategory(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add templates with different categories
	s.emailTemplates["tmpl-1"] = EmailTemplate{
		ID:       "tmpl-1",
		Name:     "Greeting 1",
		Body:     "Hello",
		Category: "greeting",
	}
	s.emailTemplates["tmpl-2"] = EmailTemplate{
		ID:       "tmpl-2",
		Name:     "Closing 1",
		Body:     "Bye",
		Category: "closing",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates?category=greeting", nil)
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TemplateListResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Should only return greeting templates
	for _, tmpl := range result.Templates {
		assert.Equal(t, "greeting", tmpl.Category)
	}
}

func TestHandleTemplates_Create(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"name": "New Template", "body": "Hello {{name}}", "category": "greeting"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created EmailTemplate
	err := json.NewDecoder(resp.Body).Decode(&created)
	require.NoError(t, err)

	assert.Equal(t, "New Template", created.Name)
	assert.Equal(t, "Hello {{name}}", created.Body)
	assert.Equal(t, "greeting", created.Category)
	assert.Contains(t, created.Variables, "name")
	assert.NotEmpty(t, created.ID)
	assert.True(t, created.CreatedAt > 0)
}

func TestHandleTemplates_CreateMissingName(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"body": "Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleTemplates_CreateMissingBody(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"name": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	s := createTestServerWithTemplates()

	req := httptest.NewRequest(http.MethodDelete, "/api/templates", nil)
	w := httptest.NewRecorder()

	s.handleTemplates(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleTemplateByID_Get(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add a template
	s.emailTemplates["tmpl-test"] = EmailTemplate{
		ID:   "tmpl-test",
		Name: "Test Template",
		Body: "Test body",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/tmpl-test", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var template EmailTemplate
	err := json.NewDecoder(resp.Body).Decode(&template)
	require.NoError(t, err)

	assert.Equal(t, "tmpl-test", template.ID)
	assert.Equal(t, "Test Template", template.Name)
}

func TestHandleTemplateByID_GetDefault(t *testing.T) {
	s := createTestServerWithTemplates()

	// Get a default template (no custom templates added)
	req := httptest.NewRequest(http.MethodGet, "/api/templates/default-thanks", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var template EmailTemplate
	err := json.NewDecoder(resp.Body).Decode(&template)
	require.NoError(t, err)

	assert.Equal(t, "default-thanks", template.ID)
}

func TestHandleTemplateByID_NotFound(t *testing.T) {
	s := createTestServerWithTemplates()

	req := httptest.NewRequest(http.MethodGet, "/api/templates/nonexistent", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleTemplateByID_Delete(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add a template
	s.emailTemplates["tmpl-delete"] = EmailTemplate{
		ID:   "tmpl-delete",
		Name: "To Delete",
		Body: "Test body",
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/templates/tmpl-delete", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deleted
	_, exists := s.emailTemplates["tmpl-delete"]
	assert.False(t, exists)
}

func TestHandleTemplateByID_Update(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add a template
	s.emailTemplates["tmpl-update"] = EmailTemplate{
		ID:        "tmpl-update",
		Name:      "Original Name",
		Body:      "Original body",
		CreatedAt: 1704067200,
	}

	body := `{"name": "Updated Name", "body": "Updated body with {{variable}}"}`
	req := httptest.NewRequest(http.MethodPut, "/api/templates/tmpl-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated EmailTemplate
	err := json.NewDecoder(resp.Body).Decode(&updated)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated body with {{variable}}", updated.Body)
	assert.Contains(t, updated.Variables, "variable")
	assert.True(t, updated.UpdatedAt > 0)
}

func TestHandleTemplateByID_UpdateNotFound(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"name": "Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/templates/nonexistent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleTemplateByID_MissingID(t *testing.T) {
	s := createTestServerWithTemplates()

	req := httptest.NewRequest(http.MethodGet, "/api/templates/", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleTemplateByID_Expand(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add a template
	s.emailTemplates["tmpl-expand"] = EmailTemplate{
		ID:        "tmpl-expand",
		Name:      "Test",
		Subject:   "Hello {{name}}",
		Body:      "Dear {{name}}, welcome to {{company}}!",
		Variables: []string{"name", "company"},
	}

	body := `{"variables": {"name": "John", "company": "Acme"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates/tmpl-expand/expand", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Hello John", result["subject"])
	assert.Equal(t, "Dear John, welcome to Acme!", result["body"])
}

func TestHandleTemplateByID_ExpandNotFound(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"variables": {}}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates/nonexistent/expand", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleTemplateByID_ExpandDefaultTemplate(t *testing.T) {
	s := createTestServerWithTemplates()

	body := `{"variables": {"name": "John", "topic": "meeting"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/templates/default-followup/expand", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	subject := result["subject"].(string)
	resultBody := result["body"].(string)

	assert.Contains(t, subject, "meeting")
	assert.Contains(t, resultBody, "John")
	assert.Contains(t, resultBody, "meeting")
}

func TestHandleTemplateByID_ExpandMethodNotAllowed(t *testing.T) {
	s := createTestServerWithTemplates()

	// Add a template
	s.emailTemplates["tmpl-expand"] = EmailTemplate{
		ID:   "tmpl-expand",
		Name: "Test",
		Body: "Test",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/templates/tmpl-expand/expand", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleTemplateByID_MethodNotAllowed(t *testing.T) {
	s := createTestServerWithTemplates()

	req := httptest.NewRequest(http.MethodPatch, "/api/templates/some-id", nil)
	w := httptest.NewRecorder()

	s.handleTemplateByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
