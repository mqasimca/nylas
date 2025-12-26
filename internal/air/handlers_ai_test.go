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
