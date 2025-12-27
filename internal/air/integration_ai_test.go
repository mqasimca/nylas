//go:build integration
// +build integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// AI INTEGRATION TESTS
// ================================

func TestIntegration_AISummarize(t *testing.T) {
	server := testServer(t)

	// Test with a simple prompt
	reqBody := AIRequest{
		EmailID: "test-email-id",
		Prompt:  "Say 'test successful' in exactly those words",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	// If claude CLI is installed, we should get a response
	// If not, we should get an error about CLI not found
	if w.Code == http.StatusOK {
		var resp AIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		t.Logf("AI response: %s", resp.Summary)
	} else if w.Code == http.StatusInternalServerError {
		var resp AIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if strings.Contains(resp.Error, "Claude Code CLI not found") {
			t.Skip("Claude Code CLI not installed, skipping AI test")
		}

		t.Logf("AI error: %s", resp.Error)
	} else {
		t.Fatalf("unexpected status %d: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AISummarize_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/summarize", nil)
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AISummarize_EmptyPrompt(t *testing.T) {
	server := testServer(t)

	reqBody := AIRequest{
		EmailID: "test-email-id",
		Prompt:  "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp AIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for empty prompt")
	}

	if resp.Error != "Prompt is required" {
		t.Errorf("expected 'Prompt is required', got: %s", resp.Error)
	}
}

// =============================================================================
// Phase 7: Middleware Integration Tests
// =============================================================================

// TestIntegration_Middleware_Compression verifies gzip compression works with full server.
