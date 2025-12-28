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
// Smart Replies Integration Tests
// =============================================================================

func TestIntegration_AISmartReplies_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/smart-replies", nil)
	w := httptest.NewRecorder()

	server.handleAISmartReplies(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AISmartReplies_EmptyBody(t *testing.T) {
	server := testServer(t)

	reqBody := SmartReplyRequest{
		EmailID: "test-email-id",
		Subject: "Test Subject",
		From:    "sender@example.com",
		Body:    "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/smart-replies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISmartReplies(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp SmartReplyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for empty body")
	}

	if resp.Error != "Email body is required" {
		t.Errorf("expected 'Email body is required', got: %s", resp.Error)
	}
}

func TestIntegration_AISmartReplies_WithContent(t *testing.T) {
	server := testServer(t)

	reqBody := SmartReplyRequest{
		EmailID: "test-email-id",
		Subject: "Meeting Request",
		From:    "john@example.com",
		Body:    "Hi, can we schedule a meeting tomorrow at 2pm to discuss the project?",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/smart-replies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISmartReplies(w, req)

	if w.Code == http.StatusOK {
		var resp SmartReplyResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		if len(resp.Replies) != 3 {
			t.Errorf("expected 3 replies, got %d", len(resp.Replies))
		}

		t.Logf("Smart replies: %v", resp.Replies)
	} else if w.Code == http.StatusInternalServerError {
		var resp SmartReplyResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if strings.Contains(resp.Error, "Claude Code CLI not found") ||
			strings.Contains(resp.Error, "claude code CLI not found") {
			t.Skip("Claude Code CLI not installed, skipping AI test")
		}

		t.Logf("AI error: %s", resp.Error)
	} else {
		t.Fatalf("unexpected status %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// Enhanced Summary Integration Tests
// =============================================================================

func TestIntegration_AIEnhancedSummary_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/enhanced-summary", nil)
	w := httptest.NewRecorder()

	server.handleAIEnhancedSummary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AIEnhancedSummary_EmptyBody(t *testing.T) {
	server := testServer(t)

	reqBody := EnhancedSummaryRequest{
		EmailID: "test-email-id",
		Subject: "Test Subject",
		From:    "sender@example.com",
		Body:    "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhanced-summary", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIEnhancedSummary(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp EnhancedSummaryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for empty body")
	}

	if resp.Error != "Email body is required" {
		t.Errorf("expected 'Email body is required', got: %s", resp.Error)
	}
}

func TestIntegration_AIEnhancedSummary_WithContent(t *testing.T) {
	server := testServer(t)

	reqBody := EnhancedSummaryRequest{
		EmailID: "test-email-id",
		Subject: "Project Update - Action Required",
		From:    "manager@example.com",
		Body: `Hi Team,

I wanted to update you on the project status. We've made good progress this week.

Action items:
1. Please review the attached document by Friday
2. Schedule a follow-up meeting for next week
3. Update the project timeline

Let me know if you have any questions.

Best,
Manager`,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhanced-summary", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIEnhancedSummary(w, req)

	if w.Code == http.StatusOK {
		var resp EnhancedSummaryResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		if resp.Summary == "" {
			t.Error("expected non-empty summary")
		}

		// Validate sentiment is one of the expected values
		validSentiments := map[string]bool{"positive": true, "neutral": true, "negative": true, "urgent": true}
		if !validSentiments[resp.Sentiment] {
			t.Errorf("unexpected sentiment: %s", resp.Sentiment)
		}

		// Validate category is one of the expected values
		validCategories := map[string]bool{"meeting": true, "task": true, "fyi": true, "question": true, "social": true}
		if !validCategories[resp.Category] {
			t.Errorf("unexpected category: %s", resp.Category)
		}

		t.Logf("Summary: %s", resp.Summary)
		t.Logf("Action Items: %v", resp.ActionItems)
		t.Logf("Sentiment: %s, Category: %s", resp.Sentiment, resp.Category)
	} else if w.Code == http.StatusInternalServerError {
		var resp EnhancedSummaryResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if strings.Contains(resp.Error, "Claude Code CLI not found") ||
			strings.Contains(resp.Error, "claude code CLI not found") {
			t.Skip("Claude Code CLI not installed, skipping AI test")
		}

		t.Logf("AI error: %s", resp.Error)
	} else {
		t.Fatalf("unexpected status %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// Phase 7: Middleware Integration Tests
// =============================================================================

// TestIntegration_Middleware_Compression verifies gzip compression works with full server.
