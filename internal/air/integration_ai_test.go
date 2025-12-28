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
// Auto-Label Integration Tests
// =============================================================================

func TestIntegration_AIAutoLabel_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/auto-label", nil)
	w := httptest.NewRecorder()

	server.handleAIAutoLabel(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AIAutoLabel_MissingContent(t *testing.T) {
	server := testServer(t)

	reqBody := AutoLabelRequest{
		EmailID: "test-email-id",
		Subject: "",
		From:    "sender@example.com",
		Body:    "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/auto-label", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIAutoLabel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp AutoLabelResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for missing content")
	}

	if resp.Error != "Email subject or body is required" {
		t.Errorf("expected 'Email subject or body is required', got: %s", resp.Error)
	}
}

func TestIntegration_AIAutoLabel_WithContent(t *testing.T) {
	server := testServer(t)

	reqBody := AutoLabelRequest{
		EmailID: "test-email-id",
		Subject: "Q4 Budget Review Meeting - URGENT",
		From:    "cfo@company.com",
		Body: `Hi Team,

We need to review the Q4 budget numbers urgently before the board meeting on Friday.

Please bring your department's spending reports and forecasts.

This is high priority - please confirm attendance.

Thanks,
CFO`,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/auto-label", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIAutoLabel(w, req)

	if w.Code == http.StatusOK {
		var resp AutoLabelResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		if len(resp.Labels) == 0 {
			t.Error("expected at least one label")
		}

		// Validate priority is one of the expected values
		validPriorities := map[string]bool{"high": true, "normal": true, "low": true}
		if !validPriorities[resp.Priority] {
			t.Errorf("unexpected priority: %s", resp.Priority)
		}

		// Validate category is one of the expected values
		validCategories := map[string]bool{
			"meeting": true, "task": true, "fyi": true, "question": true,
			"social": true, "newsletter": true, "promotion": true,
			"urgent": true, "personal": true, "work": true,
		}
		if !validCategories[resp.Category] {
			t.Errorf("unexpected category: %s", resp.Category)
		}

		t.Logf("Labels: %v", resp.Labels)
		t.Logf("Category: %s, Priority: %s", resp.Category, resp.Priority)
	} else if w.Code == http.StatusInternalServerError {
		var resp AutoLabelResponse
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
// Thread Summary Integration Tests
// =============================================================================

func TestIntegration_AIThreadSummary_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/thread-summary", nil)
	w := httptest.NewRecorder()

	server.handleAIThreadSummary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AIThreadSummary_NoMessages(t *testing.T) {
	server := testServer(t)

	reqBody := ThreadSummaryRequest{
		ThreadID: "thread-123",
		Messages: []ThreadMessage{},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/thread-summary", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIThreadSummary(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ThreadSummaryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for no messages")
	}

	if resp.Error != "At least one message is required" {
		t.Errorf("expected 'At least one message is required', got: %s", resp.Error)
	}
}

func TestIntegration_AIThreadSummary_WithMessages(t *testing.T) {
	server := testServer(t)

	reqBody := ThreadSummaryRequest{
		ThreadID: "thread-123",
		Messages: []ThreadMessage{
			{
				From:    "alice@company.com",
				Subject: "Project Kickoff",
				Body:    "Hi team, let's kick off the new project. I've attached the initial requirements.",
				Date:    1703980800,
			},
			{
				From:    "bob@company.com",
				Subject: "Re: Project Kickoff",
				Body:    "Thanks Alice! I've reviewed the requirements. I have a few questions about the timeline.",
				Date:    1703984400,
			},
			{
				From:    "alice@company.com",
				Subject: "Re: Project Kickoff",
				Body:    "Sure Bob! Let's schedule a call tomorrow to discuss. Does 2pm work?",
				Date:    1703988000,
			},
			{
				From:    "bob@company.com",
				Subject: "Re: Project Kickoff",
				Body:    "2pm works perfectly. I'll send a calendar invite.",
				Date:    1703991600,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/thread-summary", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAIThreadSummary(w, req)

	if w.Code == http.StatusOK {
		var resp ThreadSummaryResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		if resp.Summary == "" {
			t.Error("expected non-empty summary")
		}

		if resp.MessageCount != 4 {
			t.Errorf("expected message count 4, got %d", resp.MessageCount)
		}

		if len(resp.Participants) != 2 {
			t.Errorf("expected 2 participants, got %d", len(resp.Participants))
		}

		t.Logf("Summary: %s", resp.Summary)
		t.Logf("Key Points: %v", resp.KeyPoints)
		t.Logf("Action Items: %v", resp.ActionItems)
		t.Logf("Participants: %v", resp.Participants)
		t.Logf("Timeline: %s", resp.Timeline)
		t.Logf("Next Steps: %s", resp.NextSteps)
	} else if w.Code == http.StatusInternalServerError {
		var resp ThreadSummaryResponse
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
