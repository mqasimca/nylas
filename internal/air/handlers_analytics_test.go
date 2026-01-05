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

func TestRecordEmailReceived(t *testing.T) {
	// Get initial value
	aStore.mu.RLock()
	initial := aStore.analytics.TotalReceived
	aStore.mu.RUnlock()

	// Record received email
	RecordEmailReceived()

	// Check increment
	aStore.mu.RLock()
	after := aStore.analytics.TotalReceived
	aStore.mu.RUnlock()

	assert.Equal(t, initial+1, after, "TotalReceived should increment by 1")
}

func TestRecordEmailSent(t *testing.T) {
	// Get initial value
	aStore.mu.RLock()
	initial := aStore.analytics.TotalSent
	aStore.mu.RUnlock()

	// Record sent email
	RecordEmailSent()

	// Check increment
	aStore.mu.RLock()
	after := aStore.analytics.TotalSent
	aStore.mu.RUnlock()

	assert.Equal(t, initial+1, after, "TotalSent should increment by 1")
}

func TestRecordInboxZero(t *testing.T) {
	// Reset analytics state for test
	aStore.mu.Lock()
	originalInboxZeroCount := aStore.analytics.InboxZeroCount
	originalCurrentStreak := aStore.analytics.CurrentStreak
	originalBestStreak := aStore.analytics.BestStreak

	// Set up known state
	aStore.analytics.InboxZeroCount = 5
	aStore.analytics.CurrentStreak = 3
	aStore.analytics.BestStreak = 5
	aStore.mu.Unlock()

	// Record inbox zero
	RecordInboxZero()

	// Check increments
	aStore.mu.RLock()
	inboxZeroCount := aStore.analytics.InboxZeroCount
	currentStreak := aStore.analytics.CurrentStreak
	bestStreak := aStore.analytics.BestStreak
	aStore.mu.RUnlock()

	assert.Equal(t, 6, inboxZeroCount, "InboxZeroCount should increment by 1")
	assert.Equal(t, 4, currentStreak, "CurrentStreak should increment by 1")
	assert.Equal(t, 5, bestStreak, "BestStreak should not change when current < best")

	// Test when current streak exceeds best streak
	aStore.mu.Lock()
	aStore.analytics.CurrentStreak = 5
	aStore.analytics.BestStreak = 5
	aStore.mu.Unlock()

	RecordInboxZero()

	aStore.mu.RLock()
	newBestStreak := aStore.analytics.BestStreak
	aStore.mu.RUnlock()

	assert.Equal(t, 6, newBestStreak, "BestStreak should update when current > best")

	// Restore original values
	aStore.mu.Lock()
	aStore.analytics.InboxZeroCount = originalInboxZeroCount
	aStore.analytics.CurrentStreak = originalCurrentStreak
	aStore.analytics.BestStreak = originalBestStreak
	aStore.mu.Unlock()
}

func TestRecordInboxZero_UpdatesBestStreak(t *testing.T) {
	// Set up state where current streak will exceed best
	aStore.mu.Lock()
	originalInboxZeroCount := aStore.analytics.InboxZeroCount
	originalCurrentStreak := aStore.analytics.CurrentStreak
	originalBestStreak := aStore.analytics.BestStreak

	aStore.analytics.CurrentStreak = 10
	aStore.analytics.BestStreak = 10
	aStore.mu.Unlock()

	// Record inbox zero
	RecordInboxZero()

	// Check that best streak is updated
	aStore.mu.RLock()
	newCurrentStreak := aStore.analytics.CurrentStreak
	newBestStreak := aStore.analytics.BestStreak
	aStore.mu.RUnlock()

	assert.Equal(t, 11, newCurrentStreak)
	assert.Equal(t, 11, newBestStreak)

	// Restore original values
	aStore.mu.Lock()
	aStore.analytics.InboxZeroCount = originalInboxZeroCount
	aStore.analytics.CurrentStreak = originalCurrentStreak
	aStore.analytics.BestStreak = originalBestStreak
	aStore.mu.Unlock()
}

func TestHandleGetAnalyticsDashboard(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/dashboard", nil)
	w := httptest.NewRecorder()

	s.handleGetAnalyticsDashboard(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var analytics EmailAnalytics
	err := json.NewDecoder(resp.Body).Decode(&analytics)
	require.NoError(t, err)

	// Check that we got valid analytics data
	assert.GreaterOrEqual(t, analytics.TotalReceived, 0)
	assert.GreaterOrEqual(t, analytics.TotalSent, 0)
}

func TestHandleGetAnalyticsTrends(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name          string
		period        string
		expectedField string
	}{
		{
			name:          "default period (week)",
			period:        "",
			expectedField: "week",
		},
		{
			name:          "week period",
			period:        "week",
			expectedField: "week",
		},
		{
			name:          "month period",
			period:        "month",
			expectedField: "month",
		},
		{
			name:          "quarter period",
			period:        "quarter",
			expectedField: "quarter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/analytics/trends"
			if tt.period != "" {
				url += "?period=" + tt.period
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			s.handleGetAnalyticsTrends(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			var result map[string]any
			err := json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedField, result["period"])
			assert.NotNil(t, result["weeklyTrend"])
			assert.NotNil(t, result["hourlyVolume"])
			assert.NotNil(t, result["dailyVolume"])
		})
	}
}

func TestHandleGetFocusTimeSuggestions(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/focus-time", nil)
	w := httptest.NewRecorder()

	s.handleGetFocusTimeSuggestions(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var suggestions []FocusTimeSuggestion
	err := json.NewDecoder(resp.Body).Decode(&suggestions)
	require.NoError(t, err)

	// Should have at least the evening suggestion which is always added
	assert.NotEmpty(t, suggestions)

	// Check that evening suggestion is present
	hasEvening := false
	for _, s := range suggestions {
		if s.StartHour == 18 && s.EndHour == 20 {
			hasEvening = true
			assert.Equal(t, "Weekdays", s.Day)
			assert.Equal(t, 90, s.Score)
		}
	}
	assert.True(t, hasEvening, "Should have evening focus suggestion")
}

func TestHandleGetProductivityStats(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/productivity", nil)
	w := httptest.NewRecorder()

	s.handleGetProductivityStats(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var result map[string]any
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Check all expected fields are present
	assert.Contains(t, result, "responseRate")
	assert.Contains(t, result, "avgResponseTime")
	assert.Contains(t, result, "inboxZeroCount")
	assert.Contains(t, result, "currentStreak")
	assert.Contains(t, result, "bestStreak")
	assert.Contains(t, result, "focusTimeHours")
	assert.Contains(t, result, "emailsProcessed")
}

func TestFocusTimeSuggestion_Fields(t *testing.T) {
	suggestion := FocusTimeSuggestion{
		StartHour: 6,
		EndHour:   9,
		Day:       "Weekdays",
		Reason:    "Low email volume",
		Score:     85,
	}

	assert.Equal(t, 6, suggestion.StartHour)
	assert.Equal(t, 9, suggestion.EndHour)
	assert.Equal(t, "Weekdays", suggestion.Day)
	assert.Equal(t, "Low email volume", suggestion.Reason)
	assert.Equal(t, 85, suggestion.Score)
}

func TestEmailAnalytics_Fields(t *testing.T) {
	analytics := EmailAnalytics{
		TotalReceived:   100,
		TotalSent:       50,
		TotalArchived:   80,
		TotalDeleted:    20,
		AvgResponseTime: 2.5,
		ResponseRate:    75.0,
		BusiestHour:     10,
		BusiestDay:      "Tuesday",
		InboxZeroCount:  5,
		CurrentStreak:   3,
		BestStreak:      7,
		FocusTimeHours:  10.5,
	}

	assert.Equal(t, 100, analytics.TotalReceived)
	assert.Equal(t, 50, analytics.TotalSent)
	assert.Equal(t, 80, analytics.TotalArchived)
	assert.Equal(t, 20, analytics.TotalDeleted)
	assert.Equal(t, 2.5, analytics.AvgResponseTime)
	assert.Equal(t, 75.0, analytics.ResponseRate)
	assert.Equal(t, 10, analytics.BusiestHour)
	assert.Equal(t, "Tuesday", analytics.BusiestDay)
	assert.Equal(t, 5, analytics.InboxZeroCount)
	assert.Equal(t, 3, analytics.CurrentStreak)
	assert.Equal(t, 7, analytics.BestStreak)
	assert.Equal(t, 10.5, analytics.FocusTimeHours)
}

func TestSenderStats_Fields(t *testing.T) {
	stats := SenderStats{
		Email:    "test@example.com",
		Name:     "Test User",
		Count:    42,
		AvgReply: 1.5,
	}

	assert.Equal(t, "test@example.com", stats.Email)
	assert.Equal(t, "Test User", stats.Name)
	assert.Equal(t, 42, stats.Count)
	assert.Equal(t, 1.5, stats.AvgReply)
}

func TestDayVolume_Fields(t *testing.T) {
	volume := DayVolume{
		Date:     "2024-01-15",
		Received: 45,
		Sent:     12,
	}

	assert.Equal(t, "2024-01-15", volume.Date)
	assert.Equal(t, 45, volume.Received)
	assert.Equal(t, 12, volume.Sent)
}
