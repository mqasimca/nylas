//go:build integration
// +build integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// CALENDARS INTEGRATION TESTS
// ================================

func TestIntegration_ListCalendars(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/calendars", nil)
	w := httptest.NewRecorder()

	server.handleListCalendars(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CalendarsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Calendars) == 0 {
		t.Error("expected at least one calendar")
	}

	// Check for primary calendar
	hasPrimary := false
	for _, c := range resp.Calendars {
		if c.IsPrimary {
			hasPrimary = true
			t.Logf("Primary calendar: %s (%s)", c.Name, c.ID)
		} else {
			t.Logf("Calendar: %s (read_only: %v)", c.Name, c.ReadOnly)
		}
	}

	if !hasPrimary {
		t.Error("expected a primary calendar")
	}
}

// ================================
// EVENTS INTEGRATION TESTS
// ================================

func TestIntegration_ListEvents(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/events?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EventsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d events (has_more: %v)", len(resp.Events), resp.HasMore)

	for _, event := range resp.Events {
		if event.ID == "" {
			t.Error("expected event to have ID")
		}
		if event.StartTime == 0 {
			t.Error("expected event to have StartTime")
		}

		startTime := time.Unix(event.StartTime, 0)
		t.Logf("Event: %s @ %s", event.Title, startTime.Format("2006-01-02 15:04"))
	}
}

func TestIntegration_ListEvents_WithDateRange(t *testing.T) {
	server := testServer(t)

	// Get events for the current week
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday())).Truncate(24 * time.Hour)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	req := httptest.NewRequest(http.MethodGet,
		"/api/events?limit=20&start="+formatInt64(startOfWeek.Unix())+"&end="+formatInt64(endOfWeek.Unix()), nil)
	w := httptest.NewRecorder()

	server.handleListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EventsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d events for week %s - %s",
		len(resp.Events),
		startOfWeek.Format("2006-01-02"),
		endOfWeek.Format("2006-01-02"))
}

func TestIntegration_CreateUpdateDeleteEvent(t *testing.T) {
	server := testServer(t)

	// Step 1: Create an event
	createBody := `{
		"calendar_id": "primary",
		"title": "Air Integration Test Event",
		"description": "Test event created by integration tests",
		"start_time": ` + formatInt64(time.Now().Add(24*time.Hour).Unix()) + `,
		"end_time": ` + formatInt64(time.Now().Add(25*time.Hour).Unix()) + `,
		"busy": true
	}`

	createReq := httptest.NewRequest(http.MethodPost, "/api/events", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	server.handleEventsRoute(createW, createReq)

	if createW.Code != http.StatusOK {
		t.Fatalf("create event: expected status 200, got %d: %s", createW.Code, createW.Body.String())
	}

	var createResp EventActionResponse
	if err := json.NewDecoder(createW.Body).Decode(&createResp); err != nil {
		t.Fatalf("create event: failed to decode response: %v", err)
	}

	if !createResp.Success {
		t.Fatal("create event: expected success to be true")
	}

	if createResp.Event == nil {
		t.Fatal("create event: expected event in response")
	}

	eventID := createResp.Event.ID
	calendarID := createResp.Event.CalendarID
	t.Logf("Created event: %s (calendar: %s)", eventID, calendarID)

	// Cleanup: delete the event at the end of the test
	defer func() {
		deleteReq := httptest.NewRequest(http.MethodDelete,
			"/api/events/"+eventID+"?calendar_id="+calendarID, nil)
		deleteW := httptest.NewRecorder()
		server.handleEventByID(deleteW, deleteReq)
		t.Logf("Cleanup: deleted event %s (status: %d)", eventID, deleteW.Code)
	}()

	// Step 2: Update the event
	updateBody := `{
		"title": "Updated Air Integration Test Event",
		"description": "Updated description"
	}`

	updateReq := httptest.NewRequest(http.MethodPut,
		"/api/events/"+eventID+"?calendar_id="+calendarID,
		strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()

	server.handleEventByID(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("update event: expected status 200, got %d: %s", updateW.Code, updateW.Body.String())
	}

	var updateResp EventActionResponse
	if err := json.NewDecoder(updateW.Body).Decode(&updateResp); err != nil {
		t.Fatalf("update event: failed to decode response: %v", err)
	}

	if !updateResp.Success {
		t.Fatal("update event: expected success to be true")
	}

	t.Logf("Updated event: %s", eventID)

	// Step 3: Get the event to verify update
	getReq := httptest.NewRequest(http.MethodGet,
		"/api/events/"+eventID+"?calendar_id="+calendarID, nil)
	getW := httptest.NewRecorder()

	server.handleEventByID(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("get event: expected status 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var getResp EventResponse
	if err := json.NewDecoder(getW.Body).Decode(&getResp); err != nil {
		t.Fatalf("get event: failed to decode response: %v", err)
	}

	if getResp.Title != "Updated Air Integration Test Event" {
		t.Errorf("get event: expected updated title, got '%s'", getResp.Title)
	}

	t.Logf("Verified event update: title=%s", getResp.Title)
}

func TestIntegration_EventByID_NotFound(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/events/nonexistent-event-12345?calendar_id=primary", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	// Should return 404 Not Found
	if w.Code != http.StatusNotFound {
		t.Logf("Note: got status %d for non-existent event (may vary by provider)", w.Code)
	}
}

// ================================
// ================================
// AVAILABILITY INTEGRATION TESTS
// ================================

func TestIntegration_Availability(t *testing.T) {
	server := testServer(t)

	// Get availability for next week
	now := time.Now()
	startTime := now.Unix()
	endTime := now.Add(7 * 24 * time.Hour).Unix()

	// Get current user email
	grants, _ := server.grantStore.ListGrants()
	defaultID, _ := server.grantStore.GetDefaultGrant()
	var email string
	for _, g := range grants {
		if g.ID == defaultID {
			email = g.Email
			break
		}
	}

	if email == "" {
		t.Skip("Skipping: no default grant email found")
	}

	body := `{
		"start_time": ` + formatInt64(startTime) + `,
		"end_time": ` + formatInt64(endTime) + `,
		"duration_minutes": 30,
		"participants": ["` + email + `"]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/availability", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvailabilityResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d available slots", len(resp.Slots))

	for i, slot := range resp.Slots {
		if i < 5 { // Log first 5 slots
			start := time.Unix(slot.StartTime, 0)
			end := time.Unix(slot.EndTime, 0)
			t.Logf("  Slot: %s - %s", start.Format("2006-01-02 15:04"), end.Format("15:04"))
		}
	}
}

func TestIntegration_Availability_GET(t *testing.T) {
	server := testServer(t)

	// Get availability using query params
	now := time.Now()
	startTime := now.Unix()
	endTime := now.Add(7 * 24 * time.Hour).Unix()

	req := httptest.NewRequest(http.MethodGet,
		"/api/availability?start_time="+formatInt64(startTime)+
			"&end_time="+formatInt64(endTime)+
			"&duration_minutes=60", nil)
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvailabilityResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d available 60-minute slots", len(resp.Slots))
}

// ================================
// FREE/BUSY INTEGRATION TESTS
// ================================

func TestIntegration_FreeBusy(t *testing.T) {
	server := testServer(t)

	// Get free/busy for next week
	now := time.Now()
	startTime := now.Unix()
	endTime := now.Add(7 * 24 * time.Hour).Unix()

	// Get current user email
	grants, _ := server.grantStore.ListGrants()
	defaultID, _ := server.grantStore.GetDefaultGrant()
	var email string
	for _, g := range grants {
		if g.ID == defaultID {
			email = g.Email
			break
		}
	}

	if email == "" {
		t.Skip("Skipping: no default grant email found")
	}

	body := `{
		"start_time": ` + formatInt64(startTime) + `,
		"end_time": ` + formatInt64(endTime) + `,
		"emails": ["` + email + `"]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/freebusy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp FreeBusyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Got free/busy data for %d calendars", len(resp.Data))

	for _, cal := range resp.Data {
		t.Logf("  %s: %d busy slots", cal.Email, len(cal.TimeSlots))
		for i, slot := range cal.TimeSlots {
			if i < 3 { // Log first 3 slots
				start := time.Unix(slot.StartTime, 0)
				end := time.Unix(slot.EndTime, 0)
				t.Logf("    %s: %s - %s", slot.Status, start.Format("2006-01-02 15:04"), end.Format("15:04"))
			}
		}
	}
}

func TestIntegration_FreeBusy_GET(t *testing.T) {
	server := testServer(t)

	// Get free/busy using query params
	now := time.Now()
	startTime := now.Unix()
	endTime := now.Add(7 * 24 * time.Hour).Unix()

	req := httptest.NewRequest(http.MethodGet,
		"/api/freebusy?start_time="+formatInt64(startTime)+
			"&end_time="+formatInt64(endTime), nil)
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp FreeBusyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Got free/busy data via GET for %d calendars", len(resp.Data))
}

// ================================
// CONFLICTS INTEGRATION TESTS
// ================================

func TestIntegration_Conflicts(t *testing.T) {
	server := testServer(t)

	// Check for conflicts in current week
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday())).Truncate(24 * time.Hour)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	req := httptest.NewRequest(http.MethodGet,
		"/api/events/conflicts?start_time="+formatInt64(startOfWeek.Unix())+
			"&end_time="+formatInt64(endOfWeek.Unix()), nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ConflictsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d conflicts for week %s - %s",
		len(resp.Conflicts),
		startOfWeek.Format("2006-01-02"),
		endOfWeek.Format("2006-01-02"))

	for _, conflict := range resp.Conflicts {
		t.Logf("  Conflict: '%s' overlaps with '%s'",
			conflict.Event1.Title, conflict.Event2.Title)
	}
}

func TestIntegration_Conflicts_NextMonth(t *testing.T) {
	server := testServer(t)

	// Check for conflicts in next month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	req := httptest.NewRequest(http.MethodGet,
		"/api/events/conflicts?calendar_id=primary&start_time="+formatInt64(startOfMonth.Unix())+
			"&end_time="+formatInt64(endOfMonth.Unix()), nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ConflictsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d conflicts for month %s", len(resp.Conflicts), startOfMonth.Format("2006-01"))
}

// ================================
