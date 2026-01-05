//go:build !integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test handler types
func TestAvailabilityRequest_Fields(t *testing.T) {
	req := AvailabilityRequest{
		StartTime:       1704067200,
		EndTime:         1704153600,
		DurationMinutes: 30,
		Participants:    []string{"user1@example.com", "user2@example.com"},
		IntervalMinutes: 15,
	}

	assert.Equal(t, int64(1704067200), req.StartTime)
	assert.Equal(t, int64(1704153600), req.EndTime)
	assert.Equal(t, 30, req.DurationMinutes)
	assert.Equal(t, 15, req.IntervalMinutes)
	assert.Len(t, req.Participants, 2)
}

func TestAvailabilityResponse_Fields(t *testing.T) {
	resp := AvailabilityResponse{
		Slots: []AvailableSlotResponse{
			{StartTime: 1704067200, EndTime: 1704070800, Emails: []string{"test@example.com"}},
		},
		Message: "Test message",
	}

	assert.Len(t, resp.Slots, 1)
	assert.Equal(t, "Test message", resp.Message)
}

func TestAvailableSlotResponse_Fields(t *testing.T) {
	slot := AvailableSlotResponse{
		StartTime: 1704067200,
		EndTime:   1704070800,
		Emails:    []string{"user1@example.com", "user2@example.com"},
	}

	assert.Equal(t, int64(1704067200), slot.StartTime)
	assert.Equal(t, int64(1704070800), slot.EndTime)
	assert.Len(t, slot.Emails, 2)
}

func TestFreeBusyRequest_Fields(t *testing.T) {
	req := FreeBusyRequest{
		StartTime: 1704067200,
		EndTime:   1704153600,
		Emails:    []string{"user@example.com"},
	}

	assert.Equal(t, int64(1704067200), req.StartTime)
	assert.Equal(t, int64(1704153600), req.EndTime)
	assert.Len(t, req.Emails, 1)
}

func TestFreeBusyResponse_Fields(t *testing.T) {
	resp := FreeBusyResponse{
		Data: []FreeBusyCalendarResponse{
			{
				Email: "user@example.com",
				TimeSlots: []TimeSlotResponse{
					{StartTime: 1704067200, EndTime: 1704070800, Status: "busy"},
				},
			},
		},
	}

	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "user@example.com", resp.Data[0].Email)
	assert.Len(t, resp.Data[0].TimeSlots, 1)
	assert.Equal(t, "busy", resp.Data[0].TimeSlots[0].Status)
}

func TestFreeBusyCalendarResponse_Fields(t *testing.T) {
	calResp := FreeBusyCalendarResponse{
		Email: "test@example.com",
		TimeSlots: []TimeSlotResponse{
			{StartTime: 1704067200, EndTime: 1704070800, Status: "busy"},
			{StartTime: 1704074400, EndTime: 1704078000, Status: "free"},
		},
	}

	assert.Equal(t, "test@example.com", calResp.Email)
	assert.Len(t, calResp.TimeSlots, 2)
}

func TestTimeSlotResponse_Fields(t *testing.T) {
	slot := TimeSlotResponse{
		StartTime: 1704067200,
		EndTime:   1704070800,
		Status:    "free",
	}

	assert.Equal(t, int64(1704067200), slot.StartTime)
	assert.Equal(t, int64(1704070800), slot.EndTime)
	assert.Equal(t, "free", slot.Status)
}

func TestConflictsResponse_Fields(t *testing.T) {
	resp := ConflictsResponse{
		Conflicts: []EventConflict{
			{
				Event1: EventResponse{ID: "e1", Title: "Event 1"},
				Event2: EventResponse{ID: "e2", Title: "Event 2"},
			},
		},
		HasMore: false,
	}

	assert.Len(t, resp.Conflicts, 1)
	assert.False(t, resp.HasMore)
}

func TestEventConflict_Fields(t *testing.T) {
	conflict := EventConflict{
		Event1: EventResponse{ID: "e1", Title: "Meeting 1"},
		Event2: EventResponse{ID: "e2", Title: "Meeting 2"},
	}

	assert.Equal(t, "e1", conflict.Event1.ID)
	assert.Equal(t, "e2", conflict.Event2.ID)
}

// Demo mode handler tests
func TestHandleAvailability_DemoMode(t *testing.T) {
	s := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/availability", nil)
	w := httptest.NewRecorder()

	s.handleAvailability(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result AvailabilityResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Slots)
	assert.Contains(t, result.Message, "Demo mode")
}

func TestHandleAvailability_DemoMode_POST(t *testing.T) {
	s := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPost, "/api/availability", nil)
	w := httptest.NewRecorder()

	s.handleAvailability(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result AvailabilityResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Slots)
}

func TestHandleAvailability_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodDelete, "/api/availability", nil)
	w := httptest.NewRecorder()

	s.handleAvailability(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleFreeBusy_DemoMode(t *testing.T) {
	s := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/calendars/freebusy", nil)
	w := httptest.NewRecorder()

	s.handleFreeBusy(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result FreeBusyResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Data)
	assert.Equal(t, "demo@example.com", result.Data[0].Email)
}

func TestHandleFreeBusy_DemoMode_POST(t *testing.T) {
	s := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPost, "/api/calendars/freebusy", nil)
	w := httptest.NewRecorder()

	s.handleFreeBusy(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandleFreeBusy_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodDelete, "/api/calendars/freebusy", nil)
	w := httptest.NewRecorder()

	s.handleFreeBusy(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleConflicts_DemoMode(t *testing.T) {
	s := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodGet, "/api/calendar/conflicts", nil)
	w := httptest.NewRecorder()

	s.handleConflicts(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result ConflictsResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Conflicts)
	assert.False(t, result.HasMore)
}

func TestHandleConflicts_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/calendar/conflicts", nil)
	w := httptest.NewRecorder()

	s.handleConflicts(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleAvailability_NotConfigured(t *testing.T) {
	s := &Server{demoMode: false, nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/availability", nil)
	w := httptest.NewRecorder()

	s.handleAvailability(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandleFreeBusy_NotConfigured(t *testing.T) {
	s := &Server{demoMode: false, nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/calendars/freebusy", nil)
	w := httptest.NewRecorder()

	s.handleFreeBusy(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandleConflicts_NotConfigured(t *testing.T) {
	s := &Server{demoMode: false, nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/calendar/conflicts", nil)
	w := httptest.NewRecorder()

	s.handleConflicts(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
