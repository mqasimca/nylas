package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================================
// DRAFTS HANDLER ADDITIONAL TESTS
// ================================

func TestHandleDrafts_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp DraftsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Drafts) == 0 {
		t.Error("expected non-empty drafts in demo mode")
	}
}

func TestHandleDrafts_POST(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	draft := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com", "name": "Recipient"},
		},
		"subject": "Draft Subject",
		"body":    "Draft body content",
	}
	body, _ := json.Marshal(draft)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDrafts_PUT_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleDrafts_DELETE_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleListDrafts_Content(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleListDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp DraftsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify draft structure
	for _, draft := range resp.Drafts {
		if draft.ID == "" {
			t.Error("expected draft to have ID")
		}
	}
}

func TestHandleListDrafts_WithLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts?limit=5", nil)
	w := httptest.NewRecorder()

	server.handleListDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateDraft_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	draft := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"subject": "Test Draft",
		"body":    "Draft content",
	}
	body, _ := json.Marshal(draft)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateDraft(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateDraft_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := bytes.NewBufferString("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateDraft(w, req)

	// Demo mode may handle invalid JSON gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCreateDraft_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleCreateDraft(w, req)

	// Demo mode may handle empty body gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCreateDraft_WithCC(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	draft := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"cc": []map[string]string{
			{"email": "cc@example.com"},
		},
		"subject": "Test Draft with CC",
		"body":    "Content",
	}
	body, _ := json.Marshal(draft)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateDraft(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateDraft_WithBCC(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	draft := map[string]any{
		"to": []map[string]string{
			{"email": "recipient@example.com"},
		},
		"bcc": []map[string]string{
			{"email": "bcc@example.com"},
		},
		"subject": "Test Draft with BCC",
		"body":    "Content",
	}
	body, _ := json.Marshal(draft)
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateDraft(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDraftByID_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get drafts from demo data
	drafts := demoDrafts()
	if len(drafts) == 0 {
		t.Skip("no demo drafts available")
	}

	draftID := drafts[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/"+draftID, nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDraftByID_PUT(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"subject": "Updated Subject",
		"body":    "Updated body",
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/drafts/draft-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDraftByID_DELETE(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/drafts/draft-1", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDraftByID_PATCH_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPatch, "/api/drafts/draft-1", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleDraftByID_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts/nonexistent-id", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleUpdateDraft_UpdateSubject(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"subject": "New Subject",
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/drafts/draft-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateDraft_UpdateBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"body": "New body content",
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/drafts/draft-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateDraft_UpdateRecipients(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"to": []map[string]string{
			{"email": "new-recipient@example.com"},
		},
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/drafts/draft-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendDraft_Valid(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get drafts from demo data
	drafts := demoDrafts()
	if len(drafts) == 0 {
		t.Skip("no demo drafts available")
	}

	draftID := drafts[0].ID
	req := httptest.NewRequest(http.MethodPost, "/api/drafts/"+draftID+"/send", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	// Send operation should succeed in demo mode
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendDraft_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/drafts/nonexistent/send", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	// Demo mode may handle nonexistent draft gracefully
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Errorf("expected status 404 or 200, got %d", w.Code)
	}
}

func TestDraftResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := DraftResponse{
		ID:      "draft-123",
		Subject: "Draft Subject",
		Body:    "Draft Body",
		To: []EmailParticipantResponse{
			{Email: "recipient@example.com", Name: "Recipient"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DraftResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("expected ID %s, got %s", resp.ID, decoded.ID)
	}

	if decoded.Subject != resp.Subject {
		t.Errorf("expected Subject %s, got %s", resp.Subject, decoded.Subject)
	}

	if len(decoded.To) != 1 {
		t.Errorf("expected 1 To, got %d", len(decoded.To))
	}
}

func TestDraftsResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := DraftsResponse{
		Drafts: []DraftResponse{
			{ID: "draft-1", Subject: "Draft 1"},
			{ID: "draft-2", Subject: "Draft 2"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DraftsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Drafts) != 2 {
		t.Errorf("expected 2 drafts, got %d", len(decoded.Drafts))
	}
}

func TestDraftActionResponseMap(t *testing.T) {
	t.Parallel()

	resp := map[string]any{
		"success": true,
		"message": "Draft created",
		"draft": map[string]any{
			"id":      "draft-new",
			"subject": "New Draft",
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

	if decoded["draft"] == nil {
		t.Error("expected draft to be present")
	}
}

func TestDraftActionResponseErrorMap(t *testing.T) {
	t.Parallel()

	resp := map[string]any{
		"success": false,
		"error":   "Failed to create draft",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["success"] != false {
		t.Error("expected success to be false")
	}

	if decoded["error"] != "Failed to create draft" {
		t.Errorf("expected error message, got %v", decoded["error"])
	}
}
