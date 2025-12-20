package nylas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scheduler Configuration Tests

func TestHTTPClient_ListSchedulerConfigurations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/configurations", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":   "config-1",
					"name": "30 Minute Meeting",
					"slug": "30min",
				},
				{
					"id":   "config-2",
					"name": "1 Hour Meeting",
					"slug": "1hour",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	configs, err := client.ListSchedulerConfigurations(ctx)

	require.NoError(t, err)
	assert.Len(t, configs, 2)
	assert.Equal(t, "config-1", configs[0].ID)
	assert.Equal(t, "30 Minute Meeting", configs[0].Name)
	assert.Equal(t, "30min", configs[0].Slug)
}

func TestHTTPClient_GetSchedulerConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/configurations/config-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "config-123",
				"name": "Interview Meeting",
				"slug": "interview",
				"participants": []map[string]interface{}{
					{
						"email":        "interviewer@example.com",
						"name":         "Interviewer",
						"is_organizer": true,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	config, err := client.GetSchedulerConfiguration(ctx, "config-123")

	require.NoError(t, err)
	assert.Equal(t, "config-123", config.ID)
	assert.Equal(t, "Interview Meeting", config.Name)
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, "interviewer@example.com", config.Participants[0].Email)
}

func TestHTTPClient_CreateSchedulerConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/configurations", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Meeting Type", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "config-new",
				"name": "New Meeting Type",
				"slug": "new-meeting",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateSchedulerConfigurationRequest{
		Name: "New Meeting Type",
	}
	config, err := client.CreateSchedulerConfiguration(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "config-new", config.ID)
	assert.Equal(t, "New Meeting Type", config.Name)
}

func TestHTTPClient_UpdateSchedulerConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/configurations/config-456", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated Meeting", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "config-456",
				"name": "Updated Meeting",
				"slug": "updated",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.UpdateSchedulerConfigurationRequest{
		Name: strPtr("Updated Meeting"),
	}
	config, err := client.UpdateSchedulerConfiguration(ctx, "config-456", req)

	require.NoError(t, err)
	assert.Equal(t, "config-456", config.ID)
	assert.Equal(t, "Updated Meeting", config.Name)
}

func TestHTTPClient_DeleteSchedulerConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/configurations/config-delete", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteSchedulerConfiguration(ctx, "config-delete")

	require.NoError(t, err)
}

// Scheduler Session Tests

func TestHTTPClient_CreateSchedulerSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/sessions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "config-123", body["configuration_id"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session_id":       "session-abc",
				"configuration_id": "config-123",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateSchedulerSessionRequest{
		ConfigurationID: "config-123",
	}
	session, err := client.CreateSchedulerSession(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "session-abc", session.SessionID)
	assert.Equal(t, "config-123", session.ConfigurationID)
}

func TestHTTPClient_GetSchedulerSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/sessions/session-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session_id":       "session-123",
				"configuration_id": "config-456",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	session, err := client.GetSchedulerSession(ctx, "session-123")

	require.NoError(t, err)
	assert.Equal(t, "session-123", session.SessionID)
}

// Booking Tests

func TestHTTPClient_GetBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/bookings/booking-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"booking_id": "booking-123",
				"title":      "Interview with John Doe",
				"status":     "confirmed",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	booking, err := client.GetBooking(ctx, "booking-123")

	require.NoError(t, err)
	assert.Equal(t, "booking-123", booking.BookingID)
	assert.Equal(t, "Interview with John Doe", booking.Title)
	assert.Equal(t, "confirmed", booking.Status)
}

func TestHTTPClient_ListBookings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/bookings", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "config-123", r.URL.Query().Get("configuration_id"))

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"booking_id": "booking-1",
					"title":      "Meeting 1",
					"status":     "confirmed",
				},
				{
					"booking_id": "booking-2",
					"title":      "Meeting 2",
					"status":     "pending",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	bookings, err := client.ListBookings(ctx, "config-123")

	require.NoError(t, err)
	assert.Len(t, bookings, 2)
	assert.Equal(t, "booking-1", bookings[0].BookingID)
	assert.Equal(t, "confirmed", bookings[0].Status)
}

// Note: Bookings are created through the scheduler session flow, not directly

func TestHTTPClient_ConfirmBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/bookings/booking-123", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"booking_id": "booking-123",
				"status":     "confirmed",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.ConfirmBookingRequest{}
	booking, err := client.ConfirmBooking(ctx, "booking-123", req)

	require.NoError(t, err)
	assert.Equal(t, "booking-123", booking.BookingID)
	assert.Equal(t, "confirmed", booking.Status)
}

func TestHTTPClient_RescheduleBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/bookings/booking-456/reschedule", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"booking_id": "booking-456",
				"status":     "confirmed",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.ConfirmBookingRequest{}
	booking, err := client.RescheduleBooking(ctx, "booking-456", req)

	require.NoError(t, err)
	assert.Equal(t, "booking-456", booking.BookingID)
}

func TestHTTPClient_CancelBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/bookings/booking-cancel", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "cancelled", r.URL.Query().Get("reason"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.CancelBooking(ctx, "booking-cancel", "cancelled")

	require.NoError(t, err)
}

// Scheduler Page Tests

func TestHTTPClient_GetSchedulerPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/pages/page-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "page-123",
				"name": "Scheduling Page",
				"slug": "schedule-me",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	page, err := client.GetSchedulerPage(ctx, "page-123")

	require.NoError(t, err)
	assert.Equal(t, "page-123", page.ID)
	assert.Equal(t, "Scheduling Page", page.Name)
	assert.Equal(t, "schedule-me", page.Slug)
}

func TestHTTPClient_ListSchedulerPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/pages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":   "page-1",
					"name": "Page 1",
					"slug": "page-1",
				},
				{
					"id":   "page-2",
					"name": "Page 2",
					"slug": "page-2",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	pages, err := client.ListSchedulerPages(ctx)

	require.NoError(t, err)
	assert.Len(t, pages, 2)
	assert.Equal(t, "page-1", pages[0].ID)
}

func TestHTTPClient_CreateSchedulerPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/pages", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Page", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "page-new",
				"name": "New Page",
				"slug": "new-page",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateSchedulerPageRequest{
		Name: "New Page",
	}
	page, err := client.CreateSchedulerPage(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "page-new", page.ID)
	assert.Equal(t, "New Page", page.Name)
}

func TestHTTPClient_UpdateSchedulerPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/pages/page-789", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Updated Page", body["name"])

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "page-789",
				"name": "Updated Page",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode( // Test helper, encode error not actionable
			response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.UpdateSchedulerPageRequest{
		Name: strPtr("Updated Page"),
	}
	page, err := client.UpdateSchedulerPage(ctx, "page-789", req)

	require.NoError(t, err)
	assert.Equal(t, "page-789", page.ID)
	assert.Equal(t, "Updated Page", page.Name)
}

func TestHTTPClient_DeleteSchedulerPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/scheduling/pages/page-delete", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteSchedulerPage(ctx, "page-delete")

	require.NoError(t, err)
}

// Mock Client Tests

func TestMockClient_SchedulerOperations(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	// Test ListSchedulerConfigurations
	configs, err := mock.ListSchedulerConfigurations(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, configs)

	// Test GetSchedulerConfiguration
	config, err := mock.GetSchedulerConfiguration(ctx, "config-123")
	require.NoError(t, err)
	assert.Equal(t, "config-123", config.ID)

	// Test CreateSchedulerConfiguration
	createReq := &domain.CreateSchedulerConfigurationRequest{Name: "Test Config"}
	created, err := mock.CreateSchedulerConfiguration(ctx, createReq)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)

	// Test UpdateSchedulerConfiguration
	updateReq := &domain.UpdateSchedulerConfigurationRequest{Name: strPtr("Updated")}
	updated, err := mock.UpdateSchedulerConfiguration(ctx, "config-456", updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)

	// Test DeleteSchedulerConfiguration
	err = mock.DeleteSchedulerConfiguration(ctx, "config-789")
	require.NoError(t, err)

	// Test CreateSchedulerSession
	sessionReq := &domain.CreateSchedulerSessionRequest{ConfigurationID: "config-123"}
	session, err := mock.CreateSchedulerSession(ctx, sessionReq)
	require.NoError(t, err)
	assert.NotEmpty(t, session.SessionID)

	// Test GetSchedulerSession
	getSession, err := mock.GetSchedulerSession(ctx, "session-123")
	require.NoError(t, err)
	assert.Equal(t, "session-123", getSession.SessionID)

	// Test GetBooking
	booking, err := mock.GetBooking(ctx, "booking-123")
	require.NoError(t, err)
	assert.Equal(t, "booking-123", booking.BookingID)

	// Test ListBookings
	bookings, err := mock.ListBookings(ctx, "config-123")
	require.NoError(t, err)
	assert.NotEmpty(t, bookings)

	// Test ConfirmBooking
	confirmReq := &domain.ConfirmBookingRequest{}
	confirmed, err := mock.ConfirmBooking(ctx, "booking-123", confirmReq)
	require.NoError(t, err)
	assert.Equal(t, "confirmed", confirmed.Status)

	// Test RescheduleBooking
	rescheduleReq := &domain.ConfirmBookingRequest{}
	rescheduled, err := mock.RescheduleBooking(ctx, "booking-456", rescheduleReq)
	require.NoError(t, err)
	assert.NotEmpty(t, rescheduled.BookingID)

	// Test CancelBooking
	err = mock.CancelBooking(ctx, "booking-789", "User cancelled")
	require.NoError(t, err)

	// Test GetSchedulerPage
	page, err := mock.GetSchedulerPage(ctx, "page-123")
	require.NoError(t, err)
	assert.Equal(t, "page-123", page.ID)

	// Test ListSchedulerPages
	pages, err := mock.ListSchedulerPages(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, pages)

	// Test CreateSchedulerPage
	pageReq := &domain.CreateSchedulerPageRequest{Name: "Test Page"}
	newPage, err := mock.CreateSchedulerPage(ctx, pageReq)
	require.NoError(t, err)
	assert.NotEmpty(t, newPage.ID)

	// Test UpdateSchedulerPage
	updatePageReq := &domain.UpdateSchedulerPageRequest{Name: strPtr("Updated Page")}
	updatedPage, err := mock.UpdateSchedulerPage(ctx, "page-456", updatePageReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Page", updatedPage.Name)

	// Test DeleteSchedulerPage
	err = mock.DeleteSchedulerPage(ctx, "page-789")
	require.NoError(t, err)
}

// Demo Client Tests

func TestDemoClient_SchedulerOperations(t *testing.T) {
	ctx := context.Background()
	demo := nylas.NewDemoClient()

	// Test ListSchedulerConfigurations
	configs, err := demo.ListSchedulerConfigurations(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, configs)

	// Test GetSchedulerConfiguration
	config, err := demo.GetSchedulerConfiguration(ctx, "demo-config")
	require.NoError(t, err)
	assert.NotEmpty(t, config.ID)

	// Test CreateSchedulerConfiguration
	createReq := &domain.CreateSchedulerConfigurationRequest{Name: "Demo Config"}
	created, err := demo.CreateSchedulerConfiguration(ctx, createReq)
	require.NoError(t, err)
	assert.Equal(t, "Demo Config", created.Name)

	// Test UpdateSchedulerConfiguration
	updateReq := &domain.UpdateSchedulerConfigurationRequest{Name: strPtr("Demo Updated")}
	updated, err := demo.UpdateSchedulerConfiguration(ctx, "demo-config", updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Demo Updated", updated.Name)

	// Test DeleteSchedulerConfiguration
	err = demo.DeleteSchedulerConfiguration(ctx, "demo-config")
	require.NoError(t, err)

	// Test sessions, bookings, and pages
	session, err := demo.CreateSchedulerSession(ctx, &domain.CreateSchedulerSessionRequest{ConfigurationID: "demo-config"})
	require.NoError(t, err)
	assert.NotEmpty(t, session.SessionID)

	getSession, err := demo.GetSchedulerSession(ctx, "demo-session")
	require.NoError(t, err)
	assert.NotEmpty(t, getSession.SessionID)

	bookings, err := demo.ListBookings(ctx, "demo-config")
	require.NoError(t, err)
	assert.NotEmpty(t, bookings)

	pages, err := demo.ListSchedulerPages(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, pages)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
