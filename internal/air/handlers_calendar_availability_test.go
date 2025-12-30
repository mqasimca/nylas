package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ================================
// AVAILABILITY AND FREE-BUSY TESTS
// ================================

func TestHandleAvailability_POST_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	availability := map[string]any{
		"participants": []map[string]string{
			{"email": "user@example.com"},
		},
		"start_time":       now.Unix(),
		"end_time":         now.Add(7 * 24 * time.Hour).Unix(),
		"duration_minutes": 30,
	}
	body, _ := json.Marshal(availability)
	req := httptest.NewRequest(http.MethodPost, "/api/availability", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleAvailability_GET_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	url := "/api/availability?email=user@example.com&start=" +
		string(rune(now.Unix())) + "&end=" + string(rune(now.Add(24*time.Hour).Unix()))
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleFreeBusy_POST_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	freeBusy := map[string]any{
		"emails":     []string{"user@example.com"},
		"start_time": now.Unix(),
		"end_time":   now.Add(7 * 24 * time.Hour).Unix(),
	}
	body, _ := json.Marshal(freeBusy)
	req := httptest.NewRequest(http.MethodPost, "/api/free-busy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleConflicts_ValidRequest(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	req := httptest.NewRequest(http.MethodGet,
		"/api/conflicts?start="+string(rune(now.Unix()))+"&end="+string(rune(now.Add(24*time.Hour).Unix())),
		nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleConflicts_POST_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// ================================
// RESPONSE SERIALIZATION TESTS
// ================================

func TestCalendarResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := CalendarResponse{
		ID:          "cal-123",
		Name:        "Work Calendar",
		Description: "My work calendar",
		IsPrimary:   true,
		ReadOnly:    false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CalendarResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("expected ID %s, got %s", resp.ID, decoded.ID)
	}

	if decoded.Name != resp.Name {
		t.Errorf("expected Name %s, got %s", resp.Name, decoded.Name)
	}

	if !decoded.IsPrimary {
		t.Error("expected IsPrimary to be true")
	}
}

func TestEventResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := EventResponse{
		ID:          "event-123",
		Title:       "Team Meeting",
		Description: "Weekly team sync",
		Location:    "Room A",
		CalendarID:  "cal-1",
		Status:      "confirmed",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EventResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("expected ID %s, got %s", resp.ID, decoded.ID)
	}

	if decoded.Title != resp.Title {
		t.Errorf("expected Title %s, got %s", resp.Title, decoded.Title)
	}
}

func TestEventActionResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := EventActionResponse{
		Success: true,
		Message: "Event created",
		Event: &EventResponse{
			ID:    "event-new",
			Title: "New Event",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EventActionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("expected Success to be true")
	}

	if decoded.Event == nil {
		t.Error("expected Event to be present")
	}
}
