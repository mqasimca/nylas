//go:build !integration
// +build !integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =============================================================================
// Events Handler Additional Tests
// =============================================================================

// TestHandleEventByID_MissingID tests missing event ID error.
func TestHandleEventByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events/", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandleEventByID_DELETE_DemoMode tests event deletion in demo mode.
func TestHandleEventByID_DELETE_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/events/demo-event-001?calendar_id=primary", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

// TestHandleGetEvent_DemoMode_Found tests retrieving existing event.
func TestHandleGetEvent_DemoMode_Found(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	events := demoEvents()
	if len(events) == 0 {
		t.Skip("no demo events available")
	}

	eventID := events[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"?calendar_id=primary", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != eventID {
		t.Errorf("expected ID %s, got %s", eventID, resp.ID)
	}
}

// TestHandleGetEvent_DemoMode_NotFound tests non-existent event.
func TestHandleGetEvent_DemoMode_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events/nonexistent-event-id", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestHandleUpdateEvent_DemoMode_Success tests event update in demo mode.
func TestHandleUpdateEvent_DemoMode_Success(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	update := map[string]any{
		"title":       "Updated Event Title",
		"description": "Updated description",
	}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest(http.MethodPut, "/api/events/demo-event-001?calendar_id=primary", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	if resp.Event == nil {
		t.Error("expected event in response")
	}
}

// TestHandleCreateEvent_DemoMode_Success tests event creation in demo mode.
func TestHandleCreateEvent_DemoMode_Success(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	event := CreateEventRequest{
		CalendarID:  "primary",
		Title:       "New Test Event",
		Description: "Test event description",
		Location:    "Test Location",
		StartTime:   now.Add(1 * time.Hour).Unix(),
		EndTime:     now.Add(2 * time.Hour).Unix(),
		Timezone:    "America/New_York",
		IsAllDay:    false,
		Busy:        true,
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success, got error: %s", resp.Error)
	}

	if resp.Event == nil {
		t.Error("expected event in response")
	}
}

// TestHandleCreateEvent_AllDayEvent tests all-day event creation.
func TestHandleCreateEvent_AllDayEvent(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	event := CreateEventRequest{
		CalendarID: "primary",
		Title:      "All Day Event",
		StartTime:  tomorrow.Unix(),
		EndTime:    tomorrow.Add(24 * time.Hour).Unix(),
		IsAllDay:   true,
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

// TestHandleListEvents_WithPagination tests events list with cursor.
func TestHandleListEvents_WithPagination(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events?cursor=abc123&limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleListEvents_StructuredResponse tests events response structure.
func TestHandleListEvents_StructuredResponse(t *testing.T) {
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

	// Verify response structure
	for _, event := range resp.Events {
		if event.ID == "" {
			t.Error("expected event to have ID")
		}
		if event.Title == "" {
			t.Error("expected event to have Title")
		}
	}
}

// TestDemoEvents_Coverage tests demo events data.
func TestDemoEvents_Coverage(t *testing.T) {
	t.Parallel()

	events := demoEvents()

	if len(events) == 0 {
		t.Error("expected non-empty demo events")
	}

	// Check for variety of event types
	hasAllDay := false
	hasConferencing := false
	hasParticipants := false
	hasLocation := false

	for _, e := range events {
		if e.IsAllDay {
			hasAllDay = true
		}
		if e.Conferencing != nil {
			hasConferencing = true
		}
		if len(e.Participants) > 0 {
			hasParticipants = true
		}
		if e.Location != "" {
			hasLocation = true
		}
	}

	if !hasAllDay {
		t.Error("expected at least one all-day event")
	}
	if !hasConferencing {
		t.Error("expected at least one event with conferencing")
	}
	if !hasParticipants {
		t.Error("expected at least one event with participants")
	}
	if !hasLocation {
		t.Error("expected at least one event with location")
	}
}

// TestDemoCalendars_Coverage tests demo calendars data.
func TestDemoCalendars_Coverage(t *testing.T) {
	t.Parallel()

	calendars := demoCalendars()

	if len(calendars) == 0 {
		t.Error("expected non-empty demo calendars")
	}

	hasPrimary := false
	hasReadOnly := false
	hasColor := false

	for _, c := range calendars {
		if c.ID == "" {
			t.Error("expected calendar to have ID")
		}
		if c.Name == "" {
			t.Error("expected calendar to have Name")
		}
		if c.IsPrimary {
			hasPrimary = true
		}
		if c.ReadOnly {
			hasReadOnly = true
		}
		if c.HexColor != "" {
			hasColor = true
		}
	}

	if !hasPrimary {
		t.Error("expected at least one primary calendar")
	}
	if !hasReadOnly {
		t.Error("expected at least one read-only calendar")
	}
	if !hasColor {
		t.Error("expected at least one calendar with color")
	}
}
