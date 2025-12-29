//go:build integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutoLabelResponse_ErrorOmitsEmpty(t *testing.T) {
	t.Parallel()

	resp := AutoLabelResponse{
		Success:  true,
		Labels:   []string{"inbox"},
		Category: "fyi",
		Priority: "normal",
		Error:    "",
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

func TestAutoLabelResponse_ValidPriorities(t *testing.T) {
	t.Parallel()

	validPriorities := []string{"high", "normal", "low"}

	for _, priority := range validPriorities {
		resp := AutoLabelResponse{
			Success:  true,
			Priority: priority,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal with priority %s: %v", priority, err)
		}

		var decoded AutoLabelResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal with priority %s: %v", priority, err)
		}

		if decoded.Priority != priority {
			t.Errorf("expected Priority %s, got %s", priority, decoded.Priority)
		}
	}
}

func TestAutoLabelResponse_ValidCategories(t *testing.T) {
	t.Parallel()

	validCategories := []string{"meeting", "task", "fyi", "question", "social", "newsletter", "promotion", "urgent", "personal", "work"}

	for _, category := range validCategories {
		resp := AutoLabelResponse{
			Success:  true,
			Category: category,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal with category %s: %v", category, err)
		}

		var decoded AutoLabelResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal with category %s: %v", category, err)
		}

		if decoded.Category != category {
			t.Errorf("expected Category %s, got %s", category, decoded.Category)
		}
	}
}

// ================================
// THREAD SUMMARY HANDLER TESTS
// ================================

func TestHandleAIThreadSummary_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/ai/thread-summary", nil)
	w := httptest.NewRecorder()

	server.handleAIThreadSummary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAIThreadSummary_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/thread-summary", nil)
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "Invalid request body" {
		t.Errorf("expected error 'Invalid request body', got '%s'", resp.Error)
	}
}

func TestHandleAIThreadSummary_NoMessages(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	reqBody := ThreadSummaryRequest{
		ThreadID: "thread-123",
		Messages: []ThreadMessage{}, // Empty messages
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
		t.Error("expected Success to be false")
	}

	if resp.Error != "At least one message is required" {
		t.Errorf("expected error 'At least one message is required', got '%s'", resp.Error)
	}
}

func TestThreadSummaryRequest_JSONMarshaling(t *testing.T) {
	t.Parallel()

	req := ThreadSummaryRequest{
		ThreadID: "thread-123",
		Messages: []ThreadMessage{
			{
				From:    "alice@example.com",
				Subject: "Project Update",
				Body:    "Here's the update on the project.",
				Date:    1703980800,
			},
			{
				From:    "bob@example.com",
				Subject: "Re: Project Update",
				Body:    "Thanks for the update!",
				Date:    1703984400,
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ThreadSummaryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ThreadID != req.ThreadID {
		t.Errorf("expected ThreadID %s, got %s", req.ThreadID, decoded.ThreadID)
	}

	if len(decoded.Messages) != len(req.Messages) {
		t.Errorf("expected %d messages, got %d", len(req.Messages), len(decoded.Messages))
	}

	if decoded.Messages[0].From != req.Messages[0].From {
		t.Errorf("expected first message From %s, got %s", req.Messages[0].From, decoded.Messages[0].From)
	}
}

func TestThreadMessage_JSONMarshaling(t *testing.T) {
	t.Parallel()

	msg := ThreadMessage{
		From:    "sender@example.com",
		Subject: "Test Subject",
		Body:    "This is the message body",
		Date:    1703980800,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ThreadMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.From != msg.From {
		t.Errorf("expected From %s, got %s", msg.From, decoded.From)
	}

	if decoded.Subject != msg.Subject {
		t.Errorf("expected Subject %s, got %s", msg.Subject, decoded.Subject)
	}

	if decoded.Body != msg.Body {
		t.Errorf("expected Body %s, got %s", msg.Body, decoded.Body)
	}

	if decoded.Date != msg.Date {
		t.Errorf("expected Date %d, got %d", msg.Date, decoded.Date)
	}
}

func TestThreadSummaryResponse_JSONMarshaling(t *testing.T) {
	t.Parallel()

	resp := ThreadSummaryResponse{
		Success:      true,
		Summary:      "Discussion about project timeline and deliverables.",
		KeyPoints:    []string{"Deadline is Friday", "Budget approved", "Team assigned"},
		ActionItems:  []string{"Send report by EOD", "Schedule follow-up meeting"},
		Participants: []string{"alice@example.com", "bob@example.com", "charlie@example.com"},
		Timeline:     "Started Monday, updates shared Wednesday, decision made Friday",
		NextSteps:    "Await final approval from stakeholders",
		MessageCount: 5,
		Error:        "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ThreadSummaryResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("expected Success %v, got %v", resp.Success, decoded.Success)
	}

	if decoded.Summary != resp.Summary {
		t.Errorf("expected Summary %s, got %s", resp.Summary, decoded.Summary)
	}

	if len(decoded.KeyPoints) != len(resp.KeyPoints) {
		t.Errorf("expected %d key points, got %d", len(resp.KeyPoints), len(decoded.KeyPoints))
	}

	if len(decoded.ActionItems) != len(resp.ActionItems) {
		t.Errorf("expected %d action items, got %d", len(resp.ActionItems), len(decoded.ActionItems))
	}

	if len(decoded.Participants) != len(resp.Participants) {
		t.Errorf("expected %d participants, got %d", len(resp.Participants), len(decoded.Participants))
	}

	if decoded.Timeline != resp.Timeline {
		t.Errorf("expected Timeline %s, got %s", resp.Timeline, decoded.Timeline)
	}

	if decoded.NextSteps != resp.NextSteps {
		t.Errorf("expected NextSteps %s, got %s", resp.NextSteps, decoded.NextSteps)
	}

	if decoded.MessageCount != resp.MessageCount {
		t.Errorf("expected MessageCount %d, got %d", resp.MessageCount, decoded.MessageCount)
	}
}

func TestThreadSummaryResponse_ErrorOmitsEmpty(t *testing.T) {
	t.Parallel()

	resp := ThreadSummaryResponse{
		Success:      true,
		Summary:      "Summary",
		KeyPoints:    []string{},
		ActionItems:  []string{},
		Participants: []string{},
		MessageCount: 1,
		Error:        "",
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

func TestThreadSummaryResponse_NextStepsOmitsEmpty(t *testing.T) {
	t.Parallel()

	resp := ThreadSummaryResponse{
		Success:      true,
		Summary:      "Summary",
		KeyPoints:    []string{},
		ActionItems:  []string{},
		Participants: []string{},
		NextSteps:    "",
		MessageCount: 1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// NextSteps field should be omitted when empty
	if bytes.Contains(data, []byte(`"next_steps"`)) {
		t.Error("expected next_steps field to be omitted when empty")
	}
}
