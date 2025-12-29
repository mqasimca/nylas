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
// Response Body Validation Tests
// =============================================================================

func TestHandleExecCommand_ResponseStructure(t *testing.T) {
	t.Parallel()

	server := &Server{}

	// Test blocked command response structure
	body, _ := json.Marshal(ExecRequest{Command: "rm -rf /"})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response has expected fields
	if resp.Error == "" {
		t.Error("Expected non-empty error field for blocked command")
	}
	if resp.Output != "" {
		t.Errorf("Expected empty output for blocked command, got: %s", resp.Output)
	}
}

func TestHandleExecCommand_EmptyCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(ExecRequest{Command: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for empty command, got %d", http.StatusForbidden, w.Code)
	}

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message for empty command")
	}
}

func TestHandleExecCommand_WhitespaceOnlyCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(ExecRequest{Command: "   \t\n  "})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for whitespace command, got %d", http.StatusForbidden, w.Code)
	}
}

func TestConfigStatusResponse_JSONStructure(t *testing.T) {
	t.Parallel()

	// Test the response structure directly (handler requires configService)
	resp := ConfigStatusResponse{
		Configured:   true,
		ClientID:     "client-123",
		Region:       "us",
		HasAPIKey:    true,
		GrantCount:   2,
		DefaultGrant: "grant-abc",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ConfigStatusResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Configured != true {
		t.Error("Expected Configured to be true")
	}
	if decoded.GrantCount != 2 {
		t.Errorf("Expected GrantCount 2, got %d", decoded.GrantCount)
	}
	if decoded.Region != "us" {
		t.Errorf("Expected Region 'us', got %q", decoded.Region)
	}
}

func TestConfigStatusResponse_Unconfigured(t *testing.T) {
	t.Parallel()

	resp := ConfigStatusResponse{
		Configured: false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify JSON has expected structure
	if !strings.Contains(string(data), `"configured":false`) {
		t.Errorf("Expected configured:false, got: %s", string(data))
	}
}

func TestGrantsResponse_JSONStructure(t *testing.T) {
	t.Parallel()

	// Test the response structure directly (handler requires grantStore)
	resp := GrantsResponse{
		Grants:       []Grant{{ID: "test", Email: "test@example.com", Provider: "google"}},
		DefaultGrant: "test",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded GrantsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Grants) != 1 {
		t.Errorf("Expected 1 grant, got %d", len(decoded.Grants))
	}
	if decoded.DefaultGrant != "test" {
		t.Errorf("Expected default_grant 'test', got %q", decoded.DefaultGrant)
	}
}

func TestGrantsResponse_EmptyGrants(t *testing.T) {
	t.Parallel()

	resp := GrantsResponse{
		Grants:       []Grant{},
		DefaultGrant: "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify it produces valid empty array JSON
	if !strings.Contains(string(data), `"grants":[]`) {
		t.Errorf("Expected empty grants array, got: %s", string(data))
	}
}

// =============================================================================
