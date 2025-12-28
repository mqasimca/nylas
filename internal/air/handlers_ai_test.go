package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================================
// AI HANDLER TESTS
// ================================

func TestHandleAISummarize_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/ai/summarize", nil)
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAISummarize_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", nil)
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Invalid request body" {
		t.Errorf("expected error 'Invalid request body', got '%s'", resp.Error)
	}
}

func TestHandleAISummarize_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := bytes.NewBufferString("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", body)
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
		t.Error("expected Success to be false")
	}
}

func TestHandleAISummarize_EmptyPrompt(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Prompt is required" {
		t.Errorf("expected error 'Prompt is required', got '%s'", resp.Error)
	}
}

func TestHandleAISummarize_MissingClaudeCLI(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	reqBody := AIRequest{
		EmailID: "test-email-id",
		Prompt:  "Summarize this email: Hello world",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	// Should return 500 if claude CLI is not installed
	// The error message should mention Claude Code CLI
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("expected status 500 or 200, got %d", w.Code)
	}

	var resp AIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// If claude is not installed, we expect an error
	// If it is installed (dev machine), we expect success
	if w.Code == http.StatusInternalServerError {
		if resp.Success {
			t.Error("expected Success to be false for 500 response")
		}
		if resp.Error == "" {
			t.Error("expected non-empty error message")
		}
	}
}

func TestAIRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()

	req := AIRequest{
		EmailID: "email-123",
		Prompt:  "Summarize this email",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AIRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.EmailID != req.EmailID {
		t.Errorf("expected EmailID %s, got %s", req.EmailID, decoded.EmailID)
	}

	if decoded.Prompt != req.Prompt {
		t.Errorf("expected Prompt %s, got %s", req.Prompt, decoded.Prompt)
	}
}

func TestAIResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resp := AIResponse{
		Success: true,
		Summary: "This is a summary",
		Error:   "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("expected Success %v, got %v", resp.Success, decoded.Success)
	}

	if decoded.Summary != resp.Summary {
		t.Errorf("expected Summary %s, got %s", resp.Summary, decoded.Summary)
	}
}

func TestAIResponse_ErrorOmitsEmpty(t *testing.T) {
	t.Parallel()

	resp := AIResponse{
		Success: true,
		Summary: "Summary",
		Error:   "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Error field should be omitted when empty
	if bytes.Contains(data, []byte(`"error"`)) {
		t.Error("expected error field to be omitted when empty")
	}
}

func TestAIResponse_ErrorIncludedWhenPresent(t *testing.T) {
	t.Parallel()

	resp := AIResponse{
		Success: false,
		Summary: "",
		Error:   "Something went wrong",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Error field should be present
	if !bytes.Contains(data, []byte(`"error"`)) {
		t.Error("expected error field to be present when not empty")
	}
}

// ================================
// SMART REPLIES HANDLER TESTS
// ================================

func TestHandleAISmartReplies_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/ai/smart-replies", nil)
	w := httptest.NewRecorder()

	server.handleAISmartReplies(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAISmartReplies_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/smart-replies", nil)
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
		t.Error("expected Success to be false")
	}
}

func TestHandleAISmartReplies_MissingBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	reqBody := SmartReplyRequest{
		EmailID: "test-email-id",
		Subject: "Test Subject",
		From:    "sender@example.com",
		Body:    "", // Empty body
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Email body is required" {
		t.Errorf("expected error 'Email body is required', got '%s'", resp.Error)
	}
}

func TestSmartReplyRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()

	req := SmartReplyRequest{
		EmailID:   "email-123",
		Subject:   "Test Subject",
		From:      "sender@example.com",
		Body:      "Hello, this is a test email",
		ReplyType: "reply",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SmartReplyRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.EmailID != req.EmailID {
		t.Errorf("expected EmailID %s, got %s", req.EmailID, decoded.EmailID)
	}

	if decoded.Subject != req.Subject {
		t.Errorf("expected Subject %s, got %s", req.Subject, decoded.Subject)
	}

	if decoded.From != req.From {
		t.Errorf("expected From %s, got %s", req.From, decoded.From)
	}

	if decoded.Body != req.Body {
		t.Errorf("expected Body %s, got %s", req.Body, decoded.Body)
	}
}

func TestSmartReplyResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resp := SmartReplyResponse{
		Success: true,
		Replies: []string{"Reply 1", "Reply 2", "Reply 3"},
		Error:   "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SmartReplyResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("expected Success %v, got %v", resp.Success, decoded.Success)
	}

	if len(decoded.Replies) != len(resp.Replies) {
		t.Errorf("expected %d replies, got %d", len(resp.Replies), len(decoded.Replies))
	}
}

func TestParseRepliesFromText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "numbered list",
			input:    "1. Thanks for your email!\n2. I'll get back to you soon.\n3. Let me check on that.",
			expected: 3,
		},
		{
			name:     "plain lines",
			input:    "Thanks for reaching out\nI appreciate your message\nLooking forward to hearing from you",
			expected: 3,
		},
		{
			name:     "short lines ignored",
			input:    "Hi\nOk\nThanks for your detailed message about the project.",
			expected: 1,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRepliesFromText(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d replies, got %d: %v", tt.expected, len(result), result)
			}
		})
	}
}

// ================================
// ENHANCED SUMMARY HANDLER TESTS
// ================================

func TestHandleAIEnhancedSummary_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/ai/enhanced-summary", nil)
	w := httptest.NewRecorder()

	server.handleAIEnhancedSummary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAIEnhancedSummary_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhanced-summary", nil)
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
		t.Error("expected Success to be false")
	}
}

func TestHandleAIEnhancedSummary_MissingEmailBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	reqBody := EnhancedSummaryRequest{
		EmailID: "test-email-id",
		Subject: "Test Subject",
		From:    "sender@example.com",
		Body:    "", // Empty body
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Email body is required" {
		t.Errorf("expected error 'Email body is required', got '%s'", resp.Error)
	}
}

func TestEnhancedSummaryRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()

	req := EnhancedSummaryRequest{
		EmailID: "email-123",
		Subject: "Test Subject",
		From:    "sender@example.com",
		Body:    "Hello, this is a test email with some content.",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EnhancedSummaryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.EmailID != req.EmailID {
		t.Errorf("expected EmailID %s, got %s", req.EmailID, decoded.EmailID)
	}

	if decoded.Subject != req.Subject {
		t.Errorf("expected Subject %s, got %s", req.Subject, decoded.Subject)
	}
}

func TestEnhancedSummaryResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resp := EnhancedSummaryResponse{
		Success:     true,
		Summary:     "This is a summary of the email.",
		ActionItems: []string{"Review document", "Send follow-up"},
		Sentiment:   "positive",
		Category:    "task",
		Error:       "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EnhancedSummaryResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("expected Success %v, got %v", resp.Success, decoded.Success)
	}

	if decoded.Summary != resp.Summary {
		t.Errorf("expected Summary %s, got %s", resp.Summary, decoded.Summary)
	}

	if len(decoded.ActionItems) != len(resp.ActionItems) {
		t.Errorf("expected %d action items, got %d", len(resp.ActionItems), len(decoded.ActionItems))
	}

	if decoded.Sentiment != resp.Sentiment {
		t.Errorf("expected Sentiment %s, got %s", resp.Sentiment, decoded.Sentiment)
	}

	if decoded.Category != resp.Category {
		t.Errorf("expected Category %s, got %s", resp.Category, decoded.Category)
	}
}

func TestEnhancedSummaryResponse_ValidSentiments(t *testing.T) {
	t.Parallel()

	validSentiments := []string{"positive", "neutral", "negative", "urgent"}

	for _, sentiment := range validSentiments {
		resp := EnhancedSummaryResponse{
			Success:   true,
			Sentiment: sentiment,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal with sentiment %s: %v", sentiment, err)
		}

		var decoded EnhancedSummaryResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal with sentiment %s: %v", sentiment, err)
		}

		if decoded.Sentiment != sentiment {
			t.Errorf("expected Sentiment %s, got %s", sentiment, decoded.Sentiment)
		}
	}
}

func TestEnhancedSummaryResponse_ValidCategories(t *testing.T) {
	t.Parallel()

	validCategories := []string{"meeting", "task", "fyi", "question", "social"}

	for _, category := range validCategories {
		resp := EnhancedSummaryResponse{
			Success:  true,
			Category: category,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal with category %s: %v", category, err)
		}

		var decoded EnhancedSummaryResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal with category %s: %v", category, err)
		}

		if decoded.Category != category {
			t.Errorf("expected Category %s, got %s", category, decoded.Category)
		}
	}
}
