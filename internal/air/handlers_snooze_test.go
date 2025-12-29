package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleSnooze_List(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:      true,
		snoozedEmails: make(map[string]SnoozedEmail),
	}

	// Add a snoozed email
	server.snoozedEmails["test-123"] = SnoozedEmail{
		EmailID:     "test-123",
		SnoozeUntil: time.Now().Add(time.Hour).Unix(),
		CreatedAt:   time.Now().Unix(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/snooze", nil)
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := int(resp["count"].(float64))
	if count != 1 {
		t.Errorf("expected 1 snoozed email, got %d", count)
	}
}

func TestHandleSnooze_Create(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:      true,
		snoozedEmails: make(map[string]SnoozedEmail),
	}

	body, _ := json.Marshal(SnoozeRequest{
		EmailID:  "test-456",
		Duration: "2h",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/snooze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SnoozeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.EmailID != "test-456" {
		t.Errorf("expected email ID test-456, got %s", resp.EmailID)
	}

	// Verify snooze was stored
	if _, exists := server.snoozedEmails["test-456"]; !exists {
		t.Error("expected email to be in snoozed list")
	}
}

func TestHandleSnooze_CreateWithTimestamp(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:      true,
		snoozedEmails: make(map[string]SnoozedEmail),
	}

	futureTime := time.Now().Add(3 * time.Hour).Unix()
	body, _ := json.Marshal(SnoozeRequest{
		EmailID:     "test-789",
		SnoozeUntil: futureTime,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/snooze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SnoozeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.SnoozeUntil != futureTime {
		t.Errorf("expected snooze until %d, got %d", futureTime, resp.SnoozeUntil)
	}
}

func TestHandleSnooze_Delete(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:      true,
		snoozedEmails: make(map[string]SnoozedEmail),
	}

	// Add a snoozed email
	server.snoozedEmails["test-123"] = SnoozedEmail{
		EmailID:     "test-123",
		SnoozeUntil: time.Now().Add(time.Hour).Unix(),
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/snooze?email_id=test-123", nil)
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if _, exists := server.snoozedEmails["test-123"]; exists {
		t.Error("expected email to be removed from snoozed list")
	}
}

func TestHandleSnooze_PastTime(t *testing.T) {
	t.Parallel()
	server := &Server{
		demoMode:      true,
		snoozedEmails: make(map[string]SnoozedEmail),
	}

	pastTime := time.Now().Add(-time.Hour).Unix()
	body, _ := json.Marshal(SnoozeRequest{
		EmailID:     "test-123",
		SnoozeUntil: pastTime,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/snooze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for past snooze time, got %d", w.Code)
	}
}

func TestParseNaturalDuration(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		input   string
		wantErr bool
		checkFn func(int64) bool
	}{
		{"1h", false, func(ts int64) bool { return ts > now.Unix() && ts <= now.Add(2*time.Hour).Unix() }},
		{"2d", false, func(ts int64) bool { return ts > now.Add(24*time.Hour).Unix() }},
		{"30m", false, func(ts int64) bool { return ts > now.Unix() && ts <= now.Add(time.Hour).Unix() }},
		{"tomorrow", false, func(ts int64) bool { return ts > now.Unix() }},
		// "next week" returns next Monday 9 AM - could be < 24h away on Sunday
		{"next week", false, func(ts int64) bool {
			result := time.Unix(ts, 0)
			// Should be a Monday at 9 AM, in the future
			return result.After(now) && result.Weekday() == time.Monday && result.Hour() == 9
		}},
		{"weekend", false, func(ts int64) bool { return ts > now.Unix() }},
		{"later", false, func(ts int64) bool { return ts > now.Unix() }},
		{"9am", false, func(ts int64) bool { return ts > now.Unix() }},
		{"14:30", false, func(ts int64) bool { return ts > now.Unix() }},
		{"invalid_duration", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseNaturalDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNaturalDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil && !tt.checkFn(result) {
				t.Errorf("parseNaturalDuration(%q) = %d, time check failed", tt.input, result)
			}
		})
	}
}

func TestParseTimeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		wantHour int
		wantMin  int
		wantOK   bool
	}{
		{"9am", 9, 0, true},
		{"9pm", 21, 0, true},
		{"12pm", 12, 0, true},
		{"12am", 0, 0, true},
		{"14:30", 14, 30, true},
		{"2:30pm", 14, 30, true},
		{"9:00", 9, 0, true},
		{"25:00", 0, 0, false}, // Invalid hour
		{"12:60", 0, 0, false}, // Invalid minute
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			hour, min, ok := parseTimeString(tt.input)
			if ok != tt.wantOK {
				t.Errorf("parseTimeString(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if ok && (hour != tt.wantHour || min != tt.wantMin) {
				t.Errorf("parseTimeString(%q) = %d:%02d, want %d:%02d", tt.input, hour, min, tt.wantHour, tt.wantMin)
			}
		})
	}
}

func TestHandleSnooze_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := &Server{demoMode: true}

	req := httptest.NewRequest(http.MethodPut, "/api/snooze", nil)
	w := httptest.NewRecorder()

	server.handleSnooze(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// =============================================================================
// Frontend Filter Workflow Tests
// These tests simulate what the frontend JavaScript does to verify the API
// contracts are correct and the filter functionality works end-to-end.
// =============================================================================

// TestFilterWorkflow_VIPFilter tests the complete VIP filter workflow as the frontend uses it
