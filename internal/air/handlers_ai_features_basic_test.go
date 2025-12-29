package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// ================================
// AUTO-LABEL HANDLER TESTS
// ================================

func TestHandleAIAutoLabel_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/ai/auto-label", nil)
	w := httptest.NewRecorder()

	server.handleAIAutoLabel(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAIAutoLabel_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/auto-label", nil)
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Invalid request body" {
		t.Errorf("expected error 'Invalid request body', got '%s'", resp.Error)
	}
}

func TestHandleAIAutoLabel_MissingContent(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	reqBody := AutoLabelRequest{
		EmailID: "test-email-id",
		Subject: "", // Empty subject
		Body:    "", // Empty body
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Email subject or body is required" {
		t.Errorf("expected error 'Email subject or body is required', got '%s'", resp.Error)
	}
}

func TestAutoLabelRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()

	req := AutoLabelRequest{
		EmailID: "email-123",
		Subject: "Meeting Tomorrow",
		From:    "boss@company.com",
		Body:    "Let's discuss the project status",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AutoLabelRequest
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

func TestAutoLabelResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resp := AutoLabelResponse{
		Success:  true,
		Labels:   []string{"work", "project-x", "urgent"},
		Category: "task",
		Priority: "high",
		Error:    "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AutoLabelResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("expected Success %v, got %v", resp.Success, decoded.Success)
	}

	if len(decoded.Labels) != len(resp.Labels) {
		t.Errorf("expected %d labels, got %d", len(resp.Labels), len(decoded.Labels))
	}

	if decoded.Category != resp.Category {
		t.Errorf("expected Category %s, got %s", resp.Category, decoded.Category)
	}

	if decoded.Priority != resp.Priority {
		t.Errorf("expected Priority %s, got %s", resp.Priority, decoded.Priority)
	}
}
