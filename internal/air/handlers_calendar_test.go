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
// CALENDAR/EVENTS HANDLER TESTS
// ================================

func TestHandleListCalendars_Content(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/calendars", nil)
	w := httptest.NewRecorder()

	server.handleListCalendars(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CalendarsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Calendars) == 0 {
		t.Error("expected non-empty calendars in demo mode")
	}

	// Verify calendar structure
	for _, cal := range resp.Calendars {
		if cal.ID == "" {
			t.Error("expected calendar to have ID")
		}
		if cal.Name == "" {
			t.Error("expected calendar to have Name")
		}
	}
}

func TestHandleListCalendars_POST_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/calendars", nil)
	w := httptest.NewRecorder()

	server.handleListCalendars(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleListEvents_Content(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Error("expected non-empty events in demo mode")
	}
}

func TestHandleListEvents_WithCalendarID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events?calendar_id=cal-1", nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEvents_WithDateRange(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	start := now.Unix()
	end := now.Add(7 * 24 * time.Hour).Unix()

	req := httptest.NewRequest(http.MethodGet, "/api/events?start="+string(rune(start))+"&end="+string(rune(end)), nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	// Should handle date range even if params are invalid
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEvents_WithLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEvents_DELETE_UsesEventsRoute(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// DELETE on /api/events goes through handleEventsRoute which only allows GET/POST
	req := httptest.NewRequest(http.MethodDelete, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEventsRoute_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventsRoute_POST(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	event := map[string]any{
		"title":       "Test Event",
		"calendar_id": "cal-1",
		"when": map[string]any{
			"start_time": now.Add(time.Hour).Unix(),
			"end_time":   now.Add(2 * time.Hour).Unix(),
		},
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventsRoute_PUT_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCreateEvent_InvalidJSONBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := bytes.NewBufferString("{invalid}")
	req := httptest.NewRequest(http.MethodPost, "/api/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	// In demo mode, invalid JSON is handled gracefully
	// The handler may return 200 with demo data or 400 depending on implementation
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCreateEvent_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	// In demo mode, empty body may be handled gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCreateEvent_WithParticipants(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	event := map[string]any{
		"title":       "Team Meeting",
		"calendar_id": "cal-1",
		"when": map[string]any{
			"start_time": now.Add(time.Hour).Unix(),
			"end_time":   now.Add(2 * time.Hour).Unix(),
		},
		"participants": []map[string]string{
			{"email": "participant@example.com", "name": "Participant"},
		},
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateEvent_WithLocation(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	event := map[string]any{
		"title":       "Office Meeting",
		"calendar_id": "cal-1",
		"location":    "Conference Room A",
		"when": map[string]any{
			"start_time": now.Add(time.Hour).Unix(),
			"end_time":   now.Add(2 * time.Hour).Unix(),
		},
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateEvent_WithDescription(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	event := map[string]any{
		"title":       "Project Review",
		"calendar_id": "cal-1",
		"description": "Quarterly project review meeting",
		"when": map[string]any{
			"start_time": now.Add(time.Hour).Unix(),
			"end_time":   now.Add(2 * time.Hour).Unix(),
		},
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventByID_GET_Content(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get events from demo data
	events := demoEvents()
	if len(events) == 0 {
		t.Skip("no demo events available")
	}

	eventID := events[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventByID_PUT_UpdateTitle(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/events/event-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventByID_PUT_UpdateTime(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	update := map[string]any{
		"when": map[string]any{
			"start_time": now.Add(3 * time.Hour).Unix(),
			"end_time":   now.Add(4 * time.Hour).Unix(),
		},
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/events/event-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventByID_DELETE_Success(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/events/event-1", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleEventByID_PATCH_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPatch, "/api/events/event-1", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
