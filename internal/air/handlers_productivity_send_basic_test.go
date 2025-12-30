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
