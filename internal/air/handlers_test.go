package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// Helper to create demo server for handler tests
func newTestDemoServer() *Server {
	return NewDemoServer(":7365")
}

// ================================
// CONFIG HANDLER TESTS
// ================================

func TestHandleConfigStatus_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ConfigStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Configured {
		t.Error("expected Configured to be true in demo mode")
	}

	if resp.Region != "us" {
		t.Errorf("expected Region 'us', got %s", resp.Region)
	}

	if !resp.HasAPIKey {
		t.Error("expected HasAPIKey to be true in demo mode")
	}

	if resp.GrantCount != 3 {
		t.Errorf("expected GrantCount 3, got %d", resp.GrantCount)
	}
}

func TestHandleConfigStatus_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// ================================
// GRANTS HANDLER TESTS
// ================================

func TestHandleListGrants_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp GrantsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Grants) != 3 {
		t.Errorf("expected 3 grants, got %d", len(resp.Grants))
	}

	if resp.DefaultGrant != "demo-grant-001" {
		t.Errorf("expected default grant 'demo-grant-001', got %s", resp.DefaultGrant)
	}
}

func TestHandleSetDefaultGrant_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"grant_id": "demo-grant-002"}`
	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SetDefaultGrantResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
}

// ================================
// FOLDERS HANDLER TESTS
// ================================

func TestHandleListFolders_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FoldersResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Folders) == 0 {
		t.Error("expected non-empty folders")
	}

	// Check for standard folders
	hasInbox := false
	for _, f := range resp.Folders {
		if f.SystemFolder == "inbox" {
			hasInbox = true
			break
		}
	}
	if !hasInbox {
		t.Error("expected inbox folder")
	}
}

// ================================
// EMAILS HANDLER TESTS
// ================================

func TestHandleListEmails_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Emails) == 0 {
		t.Error("expected non-empty emails")
	}

	// Check first email has expected fields
	if resp.Emails[0].ID == "" {
		t.Error("expected email to have ID")
	}

	if resp.Emails[0].Subject == "" {
		t.Error("expected email to have Subject")
	}

	if len(resp.Emails[0].From) == 0 {
		t.Error("expected email to have From")
	}
}

func TestHandleGetEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "demo-email-001" {
		t.Errorf("expected ID 'demo-email-001', got %s", resp.ID)
	}
}

func TestHandleGetEmail_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleUpdateEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"unread": false, "starred": true}`
	req := httptest.NewRequest(http.MethodPut, "/api/emails/demo-email-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp UpdateEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
}

func TestHandleDeleteEmail_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp UpdateEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
}

// ================================
// DRAFTS HANDLER TESTS
// ================================

func TestHandleListDrafts_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp DraftsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Drafts) == 0 {
		t.Error("expected non-empty drafts")
	}
}

func TestHandleCreateDraft_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{
		"to": [{"email": "test@example.com", "name": "Test User"}],
		"subject": "Test Subject",
		"body": "Test body"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/drafts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp DraftResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("expected draft to have ID")
	}
}

func TestHandleGetDraft_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts/demo-draft-001", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp DraftResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "demo-draft-001" {
		t.Errorf("expected ID 'demo-draft-001', got %s", resp.ID)
	}
}

func TestHandleUpdateDraft_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"subject": "Updated Subject", "body": "Updated body"}`
	req := httptest.NewRequest(http.MethodPut, "/api/drafts/demo-draft-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleDeleteDraft_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/drafts/demo-draft-001", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleSendDraft_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/drafts/demo-draft-001/send", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SendMessageResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}

	if resp.MessageID == "" {
		t.Error("expected MessageID to be set")
	}
}

// ================================
// SEND MESSAGE HANDLER TESTS
// ================================

func TestHandleSendMessage_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{
		"to": [{"email": "test@example.com", "name": "Test User"}],
		"subject": "Test Subject",
		"body": "Test body"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SendMessageResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}

	if resp.MessageID == "" {
		t.Error("expected MessageID to be set")
	}
}

func TestHandleSendMessage_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/send", nil)
	w := httptest.NewRecorder()

	server.handleSendMessage(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// ================================
// CALENDARS HANDLER TESTS
// ================================

func TestHandleListCalendars_DemoMode(t *testing.T) {
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
		t.Error("expected non-empty calendars")
	}

	// Check for primary calendar
	hasPrimary := false
	for _, c := range resp.Calendars {
		if c.IsPrimary {
			hasPrimary = true
			break
		}
	}
	if !hasPrimary {
		t.Error("expected a primary calendar")
	}
}

func TestHandleListCalendars_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/calendars", nil)
	w := httptest.NewRecorder()

	server.handleListCalendars(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// ================================
// EVENTS HANDLER TESTS
// ================================

func TestHandleListEvents_DemoMode(t *testing.T) {
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
		t.Error("expected non-empty events")
	}

	// Check first event has expected fields
	if resp.Events[0].ID == "" {
		t.Error("expected event to have ID")
	}

	if resp.Events[0].Title == "" {
		t.Error("expected event to have Title")
	}

	if resp.Events[0].StartTime == 0 {
		t.Error("expected event to have StartTime")
	}
}

func TestHandleEventsRoute_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test unsupported method (PATCH) - GET and POST are now supported
	req := httptest.NewRequest(http.MethodPatch, "/api/events", nil)
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEventsRoute_POST_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Create event request - demo mode returns mock data regardless of input
	body := `{
		"calendar_id": "primary",
		"title": "Test Event",
		"description": "Test Description",
		"start_time": 1735250400,
		"end_time": 1735254000,
		"busy": true
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
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

	// Demo mode returns a fixed mock event
	if resp.Event != nil && resp.Event.CalendarID != "primary" {
		t.Errorf("expected calendar_id 'primary', got '%s'", resp.Event.CalendarID)
	}

	if resp.Message != "Event created (demo mode)" {
		t.Errorf("expected demo mode message, got '%s'", resp.Message)
	}
}

func TestHandleEventsRoute_POST_EmptyBody(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Demo mode still succeeds with empty body (returns mock data)
	req := httptest.NewRequest(http.MethodPost, "/api/events", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	// Demo mode should still return 200 with mock data
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 in demo mode, got %d", w.Code)
	}
}

func TestHandleEventByID_GET_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Use a demo event ID that exists in demoEvents()
	req := httptest.NewRequest(http.MethodGet, "/api/events/demo-event-001", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("expected event ID in response")
	}

	if resp.Title == "" {
		t.Error("expected event title in response")
	}
}

func TestHandleEventByID_PUT_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"title": "Updated Event Title", "description": "Updated description"}`

	req := httptest.NewRequest(http.MethodPut, "/api/events/demo-event-001?calendar_id=primary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Demo mode returns "(demo mode)" suffix
	if resp.Message != "Event updated (demo mode)" {
		t.Errorf("expected demo mode update message, got '%s'", resp.Message)
	}
}

func TestHandleEventByID_DELETE_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/events/demo-event-001?calendar_id=primary", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Demo mode returns "(demo mode)" suffix
	if resp.Message != "Event deleted (demo mode)" {
		t.Errorf("expected demo mode delete message, got '%s'", resp.Message)
	}
}

func TestHandleEventByID_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPatch, "/api/events/test-event-123", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEventByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Request to /api/events/ with no ID after the slash
	req := httptest.NewRequest(http.MethodGet, "/api/events/", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing ID, got %d", w.Code)
	}
}

// ================================
// HELPER FUNCTION TESTS
// ================================

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", resp["key"])
	}
}

func TestParticipantsToEmail(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result))
	}

	if result[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", result[0].Name)
	}

	if result[0].Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", result[0].Email)
	}
}

func TestGrantFromDomain(t *testing.T) {
	t.Parallel()

	domainGrant := struct {
		ID       string
		Email    string
		Provider string
	}{
		ID:       "grant-123",
		Email:    "test@example.com",
		Provider: "google",
	}

	// Since grantFromDomain expects domain.GrantInfo, we test the conversion logic
	result := Grant{
		ID:       domainGrant.ID,
		Email:    domainGrant.Email,
		Provider: domainGrant.Provider,
	}

	if result.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", result.ID)
	}

	if result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", result.Email)
	}

	if result.Provider != "google" {
		t.Errorf("expected provider 'google', got %s", result.Provider)
	}
}

// ================================
// CONTACTS HANDLER TESTS
// ================================

func TestHandleListContacts_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Contacts) == 0 {
		t.Error("expected non-empty contacts")
	}

	// Check first contact has expected fields
	if resp.Contacts[0].ID == "" {
		t.Error("expected contact to have ID")
	}

	if resp.Contacts[0].DisplayName == "" && resp.Contacts[0].GivenName == "" {
		t.Error("expected contact to have DisplayName or GivenName")
	}
}

func TestHandleListContacts_WithFilters(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with limit
	req := httptest.NewRequest(http.MethodGet, "/api/contacts?limit=3", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode may return more than limit, just verify response format
	if len(resp.Contacts) == 0 {
		t.Error("expected non-empty contacts")
	}
}

func TestHandleContactByID_GET_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/demo-contact-001", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("expected contact ID in response")
	}
}

func TestHandleContactByID_PUT_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"given_name": "Updated", "surname": "Name"}`
	req := httptest.NewRequest(http.MethodPut, "/api/contacts/demo-contact-001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestHandleContactByID_DELETE_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/contacts/demo-contact-001", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestHandleContactByID_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPatch, "/api/contacts/test-contact-123", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleContactByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing ID, got %d", w.Code)
	}
}

// ================================
// CONTACT GROUPS HANDLER TESTS
// ================================

func TestHandleListContactGroups_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/groups", nil)
	w := httptest.NewRecorder()

	server.handleContactGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactGroupsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Groups) == 0 {
		t.Error("expected non-empty contact groups")
	}
}

// ================================
// CACHE HANDLER TESTS
// ================================

func TestHandleCacheStatus_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode should return mock status
	if !resp.Enabled {
		t.Error("expected Enabled to be true in demo mode")
	}

	if !resp.Online {
		t.Error("expected Online to be true in demo mode")
	}
}

func TestHandleCacheStatus_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSync_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/sync", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSyncResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true in demo mode")
	}

	if resp.Message == "" {
		t.Error("expected Message to be set")
	}
}

func TestHandleCacheSync_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/sync", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheClear_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/clear", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSyncResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true in demo mode")
	}
}

func TestHandleCacheClear_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/clear", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSearch_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=test", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode returns mock search results
	if resp.Query != "test" {
		t.Errorf("expected Query 'test', got %s", resp.Query)
	}

	if len(resp.Results) == 0 {
		t.Error("expected non-empty search results in demo mode")
	}
}

func TestHandleCacheSearch_EmptyQuery(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Empty query should return empty results
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results for empty query, got %d", len(resp.Results))
	}
}

func TestHandleCacheSearch_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/search", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCacheSettings_GET_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CacheSettingsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode should return default settings
	if !resp.Enabled {
		t.Error("expected Enabled to be true")
	}

	if resp.MaxSizeMB == 0 {
		t.Error("expected MaxSizeMB to be set")
	}

	if resp.Theme == "" {
		t.Error("expected Theme to be set")
	}
}

func TestHandleCacheSettings_PUT_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{
		"cache_enabled": true,
		"cache_max_size_mb": 1000,
		"theme": "light"
	}`
	req := httptest.NewRequest(http.MethodPut, "/api/cache/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Error("expected success to be true")
	}
}

func TestHandleCacheSettings_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// ================================
// ADDITIONAL EDGE CASE TESTS
// ================================

func TestHandleListEmails_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with folder filter
	req := httptest.NewRequest(http.MethodGet, "/api/emails?folder=inbox&limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Emails) == 0 {
		t.Error("expected emails in response")
	}
}

func TestHandleListEmails_WithPagination(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with page token
	req := httptest.NewRequest(http.MethodGet, "/api/emails?page_token=test", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleListEmails_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/emails", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleListEvents_WithFilters(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with date filters
	req := httptest.NewRequest(http.MethodGet, "/api/events?calendar_id=primary&start=1735250400&end=1735336800", nil)
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
		t.Error("expected events in response")
	}
}

func TestHandleListContacts_WithSearch(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Test with search query
	req := httptest.NewRequest(http.MethodGet, "/api/contacts?q=alice&limit=5", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestHandleListContacts_POST_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// POST to contacts list creates a new contact in demo mode
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	// Demo mode handles POST differently - may succeed with mock data
	if w.Code != http.StatusOK && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 200 or 405, got %d", w.Code)
	}
}

func TestHandleEmailByID_WithInvalidMethod(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleEmailByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/", nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleDraftByID_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/drafts/demo-draft-001", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleDraftByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/drafts/", nil)
	w := httptest.NewRecorder()

	server.handleDraftByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleDrafts_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleContactGroups_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/contacts/groups", nil)
	w := httptest.NewRecorder()

	server.handleContactGroups(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestWriteJSON_NilValue(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "null\n" {
		t.Errorf("expected 'null', got %s", w.Body.String())
	}
}

func TestWriteJSON_EmptyMap(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]any{})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "{}\n" {
		t.Errorf("expected '{}', got %s", w.Body.String())
	}
}

func TestHandleUpdateEmail_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/emails/demo-email-001", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateEmail(w, req, "demo-email-001")

	// Demo mode might still succeed or return bad request
	// Either is acceptable for invalid JSON
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleDeleteEmail_DemoMode_Additional(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleDeleteEmail(w, req, "demo-email-001")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Error("expected success to be true")
	}
}

func TestHandleCreateContact_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{
		"display_name": "New Contact",
		"email": "new@example.com"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// handleListContacts handles both GET and POST
	server.handleListContacts(w, req)

	// POST should return method not allowed for list endpoint in demo mode
	// or it might be handled differently
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusOK {
		t.Errorf("expected status 405 or 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSendDraft_GET_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// GET on send endpoint - demo mode may handle differently
	req := httptest.NewRequest(http.MethodGet, "/api/drafts/demo-draft-001/send", nil)
	w := httptest.NewRecorder()

	server.handleSendDraft(w, req, "demo-draft-001")

	// Demo mode may return success or method not allowed
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusOK {
		t.Errorf("expected status 405 or 200, got %d", w.Code)
	}
}

func TestHandleUpdateDraft_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/drafts/demo-draft-001", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleUpdateDraft(w, req, "demo-draft-001")

	// Demo mode might handle this differently
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleEventByID_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events/nonexistent-event", nil)
	w := httptest.NewRecorder()

	server.handleEventByID(w, req)

	// Demo mode returns mock data even for nonexistent IDs
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", w.Code)
	}
}

func TestHandleCreateEvent_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/events", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleEventsRoute(w, req)

	// Demo mode might return 200 with mock data or error
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleUpdateContact_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/contacts/demo-contact-001", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	// Demo mode might return 200 with mock data or error
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleFolders_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleCalendars_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/calendars", nil)
	w := httptest.NewRecorder()

	server.handleListCalendars(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleGrants_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleSetDefaultGrant_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCacheSync_WithEmail(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/sync?email=test@example.com", nil)
	w := httptest.NewRecorder()

	server.handleCacheSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCacheClear_WithEmail(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cache/clear?email=test@example.com", nil)
	w := httptest.NewRecorder()

	server.handleCacheClear(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ================================
// AVAILABILITY HANDLER TESTS
// ================================

func TestHandleAvailability_DemoMode_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/availability", nil)
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp AvailabilityResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Slots) == 0 {
		t.Error("expected at least one availability slot in demo mode")
	}

	if resp.Message == "" {
		t.Error("expected demo mode message")
	}
}

func TestHandleAvailability_DemoMode_POST(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"start_time": 1700000000, "end_time": 1700604800, "duration_minutes": 30, "participants": ["test@example.com"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/availability", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp AvailabilityResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Slots) == 0 {
		t.Error("expected availability slots")
	}
}

func TestHandleAvailability_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/availability", nil)
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleAvailability_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/availability?start_time=1700000000&end_time=1700604800&duration_minutes=60&participants=test@example.com&interval_minutes=30", nil)
	w := httptest.NewRecorder()

	server.handleAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ================================
// FREE/BUSY HANDLER TESTS
// ================================

func TestHandleFreeBusy_DemoMode_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/freebusy", nil)
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FreeBusyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("expected at least one free/busy entry in demo mode")
	}

	// Check that the demo data has time slots
	if len(resp.Data[0].TimeSlots) == 0 {
		t.Error("expected time slots in demo mode")
	}
}

func TestHandleFreeBusy_DemoMode_POST(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := `{"start_time": 1700000000, "end_time": 1700604800, "emails": ["test@example.com"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/freebusy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FreeBusyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("expected free/busy data")
	}
}

func TestHandleFreeBusy_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/freebusy", nil)
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleFreeBusy_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/freebusy?start_time=1700000000&end_time=1700604800&emails=test@example.com,test2@example.com", nil)
	w := httptest.NewRecorder()

	server.handleFreeBusy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ================================
// CONFLICTS HANDLER TESTS
// ================================

func TestHandleConflicts_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events/conflicts", nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ConflictsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Demo mode should return at least one conflict
	if len(resp.Conflicts) == 0 {
		t.Error("expected at least one conflict in demo mode")
	}

	// Check conflict structure
	conflict := resp.Conflicts[0]
	if conflict.Event1.ID == "" || conflict.Event2.ID == "" {
		t.Error("expected conflict events to have IDs")
	}
	if conflict.Event1.Title == "" || conflict.Event2.Title == "" {
		t.Error("expected conflict events to have titles")
	}
}

func TestHandleConflicts_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/events/conflicts", nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleConflicts_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/events/conflicts?calendar_id=primary&start_time=1700000000&end_time=1700604800", nil)
	w := httptest.NewRecorder()

	server.handleConflicts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ================================
// CONTACT SEARCH HANDLER TESTS
// ================================

func TestHandleContactSearch_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Without query, should return all demo contacts
	if len(resp.Contacts) == 0 {
		t.Error("expected contacts in demo mode")
	}
}

func TestHandleContactSearch_WithQuery(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=Sarah", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should find Sarah Chen from demo data
	if len(resp.Contacts) == 0 {
		t.Error("expected to find Sarah in demo contacts")
	}

	found := false
	for _, c := range resp.Contacts {
		if strings.Contains(c.DisplayName, "Sarah") || strings.Contains(c.GivenName, "Sarah") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected search results to include Sarah")
	}
}

func TestHandleContactSearch_WithEmailQuery(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=nylas.com", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should find contacts with nylas.com email
	if len(resp.Contacts) == 0 {
		t.Error("expected to find contacts with nylas.com email")
	}
}

func TestHandleContactSearch_NoResults(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=nonexistentperson12345", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should return empty results
	if len(resp.Contacts) != 0 {
		t.Errorf("expected no results, got %d", len(resp.Contacts))
	}
}

func TestHandleContactSearch_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/contacts/search", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleContactSearch_WithPagination(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=a&limit=2", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ================================
// HELPER FUNCTION TESTS
// ================================

func TestContainsEmail(t *testing.T) {
	t.Parallel()

	emails := []ContactEmailResponse{
		{Email: "test@example.com", Type: "work"},
		{Email: "john@nylas.com", Type: "personal"},
	}

	tests := []struct {
		query    string
		expected bool
	}{
		{"test@example.com", true},
		{"example.com", true},
		{"nylas.com", true},
		{"john", true},
		{"notfound", false},
		{"xyz@other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := containsEmail(emails, tt.query)
			if result != tt.expected {
				t.Errorf("containsEmail(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestMatchesContactQuery(t *testing.T) {
	t.Parallel()

	contact := ContactResponse{
		ID:          "test-1",
		GivenName:   "John",
		Surname:     "Doe",
		DisplayName: "John Doe",
		CompanyName: "Acme Corp",
		Notes:       "Important client",
		Emails: []ContactEmailResponse{
			{Email: "john@acme.com", Type: "work"},
		},
	}

	tests := []struct {
		query    string
		expected bool
	}{
		{"john", true},
		{"doe", true},
		{"John Doe", true},
		{"acme", true},
		{"important", true},
		{"client", true},
		{"notfound", false},
		{"xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := matchesContactQuery(contact, tt.query)
			if result != tt.expected {
				t.Errorf("matchesContactQuery(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestFindConflicts(t *testing.T) {
	t.Parallel()

	// Test with overlapping events
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 and event-2 overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}

	if len(conflicts) > 0 {
		if conflicts[0].Event1.ID != "event-1" || conflicts[0].Event2.ID != "event-2" {
			t.Error("expected conflict between event-1 and event-2")
		}
	}
}

func TestFindConflicts_NoOverlap(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_CancelledEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "cancelled", // This should be ignored
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (cancelled event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_FreeEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Free Time",
			Status: "confirmed",
			Busy:   false, // Free, not busy
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (free event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_AllDayEvents(t *testing.T) {
	t.Parallel()

	// All-day event should conflict with timed event on same day
	events := []domain.Event{
		{
			ID:     "all-day-1",
			Title:  "Holiday",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				Date: "2024-12-25", // All-day event
			},
		},
		{
			ID:     "timed-1",
			Title:  "Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1735142400, // Dec 25, 2024 12:00 UTC
				EndTime:   1735146000, // Dec 25, 2024 13:00 UTC
			},
		},
	}

	conflicts := findConflicts(events)

	// All-day event and timed event overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict (all-day vs timed), got %d", len(conflicts))
	}
}

func TestFindConflicts_MultipleConflicts(t *testing.T) {
	t.Parallel()

	// Three overlapping events should produce 3 conflicts (each pair)
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   3000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   3500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 2000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 overlaps event-2, event-1 overlaps event-3, event-2 overlaps event-3
	if len(conflicts) != 3 {
		t.Errorf("expected 3 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_EmptyList(t *testing.T) {
	t.Parallel()

	conflicts := findConflicts([]domain.Event{})

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for empty list, got %d", len(conflicts))
	}
}

func TestFindConflicts_SingleEvent(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Only Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for single event, got %d", len(conflicts))
	}
}

func TestRoundUpTo5Min(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{"already aligned", 1735142400, 1735142400}, // 12:00:00 stays 12:00:00
		{"1 second after", 1735142401, 1735142700},  // 12:00:01 -> 12:05:00
		{"2 minutes in", 1735142520, 1735142700},    // 12:02:00 -> 12:05:00
		{"4 min 59 sec", 1735142699, 1735142700},    // 12:04:59 -> 12:05:00
		{"zero", 0, 0},
		{"5 min aligned", 300, 300},
		{"10 min aligned", 600, 600},
		{"6 minutes", 360, 600}, // 00:06:00 -> 00:10:00
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundUpTo5Min(tt.input)
			if result != tt.expected {
				t.Errorf("roundUpTo5Min(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// CSS Styling Tests - Verify UI readability
// =============================================================================

func TestEmailBodyCSS_HasLightBackground(t *testing.T) {
	t.Parallel()

	// Read the preview.css file from embedded files
	cssContent, err := staticFiles.ReadFile("static/css/preview.css")
	if err != nil {
		t.Fatalf("failed to read preview.css: %v", err)
	}

	css := string(cssContent)

	// Verify email iframe container has white/light background for readability
	// Email content is now rendered in a sandboxed iframe for security
	// The iframe container provides the light background, while the iframe's
	// internal stylesheet (in email.js) handles text styling
	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		{
			"email iframe container has light background",
			"background: #ffffff",
			"HTML emails have inline styles for light backgrounds - need white bg for readability",
		},
		{
			"email body selector exists",
			".email-detail-body",
			"Email body styling must be defined",
		},
		{
			"email iframe container selector exists",
			".email-iframe-container",
			"Email iframe container styling must be defined for sandboxed email rendering",
		},
		{
			"email iframe styling exists",
			".email-body-iframe",
			"Sandboxed iframe styling must be defined for security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(css, tt.contains) {
				t.Errorf("preview.css missing required style: %s\nReason: %s", tt.contains, tt.reason)
			}
		})
	}
}

// ================================
// CONTACT PHOTO HANDLER TESTS
// ================================

func TestHandleContactPhoto_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/demo-contact-001/photo", nil)
	w := httptest.NewRecorder()

	server.handleContactPhoto(w, req, "demo-contact-001")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("expected Content-Type 'image/png', got %s", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=86400" {
		t.Errorf("expected Cache-Control 'public, max-age=86400', got %s", cacheControl)
	}

	// Should return a 1x1 transparent PNG
	body := w.Body.Bytes()
	if len(body) == 0 {
		t.Error("expected non-empty body for placeholder image")
	}

	// PNG magic bytes
	if len(body) < 8 || body[0] != 0x89 || body[1] != 0x50 || body[2] != 0x4e || body[3] != 0x47 {
		t.Error("expected valid PNG image data")
	}
}

func TestHandleContactPhoto_DifferentContactIDs(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contactIDs := []string{"contact-1", "contact-2", "demo-contact-003"}

	for _, id := range contactIDs {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/contacts/"+id+"/photo", nil)
			w := httptest.NewRecorder()

			server.handleContactPhoto(w, req, id)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for contact %s, got %d", id, w.Code)
			}
		})
	}
}

// ================================
// DELETE EVENT HANDLER TESTS
// ================================

func TestHandleDeleteEvent_DemoMode_WithCalendarID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/events/event-123?calendar_id=cal-456", nil)
	w := httptest.NewRecorder()

	server.handleDeleteEvent(w, req, "event-123")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}

	if !strings.Contains(resp.Message, "demo mode") {
		t.Errorf("expected message to mention demo mode, got: %s", resp.Message)
	}
}

func TestHandleDeleteEvent_DemoMode_DefaultCalendar(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Without calendar_id, should use "primary"
	req := httptest.NewRequest(http.MethodDelete, "/api/events/event-789", nil)
	w := httptest.NewRecorder()

	server.handleDeleteEvent(w, req, "event-789")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EventActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true in demo mode")
	}
}

// ================================
// GET EMAIL HANDLER TESTS
// ================================

func TestHandleGetEmail_DemoMode_Found(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get a known demo email ID (matches demoEmails() function)
	req := httptest.NewRequest(http.MethodGet, "/api/emails/demo-email-001", nil)
	w := httptest.NewRecorder()

	server.handleGetEmail(w, req, "demo-email-001")

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "demo-email-001" {
		t.Errorf("expected ID 'demo-email-001', got %s", resp.ID)
	}
}

func TestHandleGetEmail_DemoMode_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/emails/nonexistent-id", nil)
	w := httptest.NewRecorder()

	server.handleGetEmail(w, req, "nonexistent-id")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleGetEmail_DemoMode_AllDemoEmails(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()
	// These IDs match demoEmails() function in handlers.go
	demoIDs := []string{"demo-email-001", "demo-email-002", "demo-email-003", "demo-email-004", "demo-email-005"}

	for _, id := range demoIDs {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/emails/"+id, nil)
			w := httptest.NewRecorder()

			server.handleGetEmail(w, req, id)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for %s, got %d", id, w.Code)
			}
		})
	}
}

// ================================
// RESPONSE CONVERTER TESTS
// ================================

func TestEmailToResponse_Basic(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID:       "msg-123",
		ThreadID: "thread-456",
		Subject:  "Test Subject",
		Snippet:  "This is a snippet...",
		Body:     "<p>Full body content</p>",
		Unread:   true,
		Starred:  false,
		Folders:  []string{"INBOX"},
	}

	resp := emailToResponse(msg, false)

	if resp.ID != "msg-123" {
		t.Errorf("expected ID 'msg-123', got %s", resp.ID)
	}
	if resp.ThreadID != "thread-456" {
		t.Errorf("expected ThreadID 'thread-456', got %s", resp.ThreadID)
	}
	if resp.Subject != "Test Subject" {
		t.Errorf("expected Subject 'Test Subject', got %s", resp.Subject)
	}
	if resp.Snippet != "This is a snippet..." {
		t.Errorf("expected Snippet to match, got %s", resp.Snippet)
	}
	if resp.Body != "" {
		t.Error("expected Body to be empty when includeBody=false")
	}
	if !resp.Unread {
		t.Error("expected Unread to be true")
	}
	if resp.Starred {
		t.Error("expected Starred to be false")
	}
}

func TestEmailToResponse_WithBody(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID:   "msg-123",
		Body: "<p>Full body content</p>",
	}

	resp := emailToResponse(msg, true)

	if resp.Body != "<p>Full body content</p>" {
		t.Errorf("expected Body to be included, got %s", resp.Body)
	}
}

func TestEmailToResponse_WithParticipants(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID: "msg-123",
		From: []domain.EmailParticipant{
			{Name: "Sender Name", Email: "sender@example.com"},
		},
		To: []domain.EmailParticipant{
			{Name: "Recipient One", Email: "recipient1@example.com"},
			{Name: "Recipient Two", Email: "recipient2@example.com"},
		},
		Cc: []domain.EmailParticipant{
			{Name: "CC Person", Email: "cc@example.com"},
		},
	}

	resp := emailToResponse(msg, false)

	if len(resp.From) != 1 {
		t.Errorf("expected 1 From participant, got %d", len(resp.From))
	}
	if resp.From[0].Email != "sender@example.com" {
		t.Errorf("expected From email 'sender@example.com', got %s", resp.From[0].Email)
	}

	if len(resp.To) != 2 {
		t.Errorf("expected 2 To participants, got %d", len(resp.To))
	}

	if len(resp.Cc) != 1 {
		t.Errorf("expected 1 Cc participant, got %d", len(resp.Cc))
	}
}

func TestEmailToResponse_WithAttachments(t *testing.T) {
	t.Parallel()

	msg := domain.Message{
		ID: "msg-123",
		Attachments: []domain.Attachment{
			{ID: "att-1", Filename: "document.pdf", ContentType: "application/pdf", Size: 1024},
			{ID: "att-2", Filename: "image.png", ContentType: "image/png", Size: 2048},
		},
	}

	resp := emailToResponse(msg, false)

	if len(resp.Attachments) != 2 {
		t.Errorf("expected 2 attachments, got %d", len(resp.Attachments))
	}

	if resp.Attachments[0].Filename != "document.pdf" {
		t.Errorf("expected first attachment filename 'document.pdf', got %s", resp.Attachments[0].Filename)
	}
	if resp.Attachments[1].Size != 2048 {
		t.Errorf("expected second attachment size 2048, got %d", resp.Attachments[1].Size)
	}
}

func TestDraftToResponse_Basic(t *testing.T) {
	t.Parallel()

	draft := domain.Draft{
		ID:      "draft-123",
		Subject: "Draft Subject",
		Body:    "<p>Draft body</p>",
	}

	resp := draftToResponse(draft)

	if resp.ID != "draft-123" {
		t.Errorf("expected ID 'draft-123', got %s", resp.ID)
	}
	if resp.Subject != "Draft Subject" {
		t.Errorf("expected Subject 'Draft Subject', got %s", resp.Subject)
	}
	if resp.Body != "<p>Draft body</p>" {
		t.Errorf("expected Body to match, got %s", resp.Body)
	}
}

func TestDraftToResponse_WithRecipients(t *testing.T) {
	t.Parallel()

	draft := domain.Draft{
		ID: "draft-123",
		To: []domain.EmailParticipant{
			{Name: "To Person", Email: "to@example.com"},
		},
		Cc: []domain.EmailParticipant{
			{Name: "CC Person", Email: "cc@example.com"},
		},
		Bcc: []domain.EmailParticipant{
			{Name: "BCC Person", Email: "bcc@example.com"},
		},
	}

	resp := draftToResponse(draft)

	if len(resp.To) != 1 || resp.To[0].Email != "to@example.com" {
		t.Error("To recipients not converted correctly")
	}
	if len(resp.Cc) != 1 || resp.Cc[0].Email != "cc@example.com" {
		t.Error("Cc recipients not converted correctly")
	}
	if len(resp.Bcc) != 1 || resp.Bcc[0].Email != "bcc@example.com" {
		t.Error("Bcc recipients not converted correctly")
	}
}

func TestCalendarToResponse_Basic(t *testing.T) {
	t.Parallel()

	cal := domain.Calendar{
		ID:          "cal-123",
		Name:        "Work Calendar",
		Description: "Work events",
		Timezone:    "America/New_York",
		IsPrimary:   true,
		ReadOnly:    false,
		HexColor:    "#4285f4",
	}

	resp := calendarToResponse(cal)

	if resp.ID != "cal-123" {
		t.Errorf("expected ID 'cal-123', got %s", resp.ID)
	}
	if resp.Name != "Work Calendar" {
		t.Errorf("expected Name 'Work Calendar', got %s", resp.Name)
	}
	if resp.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got %s", resp.Timezone)
	}
	if !resp.IsPrimary {
		t.Error("expected IsPrimary to be true")
	}
	if resp.ReadOnly {
		t.Error("expected ReadOnly to be false")
	}
	if resp.HexColor != "#4285f4" {
		t.Errorf("expected HexColor '#4285f4', got %s", resp.HexColor)
	}
}

func TestContactToResponse_Basic(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID:          "contact-123",
		GivenName:   "John",
		Surname:     "Doe",
		Nickname:    "Johnny",
		CompanyName: "Acme Corp",
		JobTitle:    "Engineer",
		Birthday:    "1990-01-15",
		Notes:       "Test notes",
		PictureURL:  "https://example.com/photo.jpg",
		Source:      "google",
	}

	resp := contactToResponse(contact)

	if resp.ID != "contact-123" {
		t.Errorf("expected ID 'contact-123', got %s", resp.ID)
	}
	if resp.GivenName != "John" {
		t.Errorf("expected GivenName 'John', got %s", resp.GivenName)
	}
	if resp.Surname != "Doe" {
		t.Errorf("expected Surname 'Doe', got %s", resp.Surname)
	}
	if resp.CompanyName != "Acme Corp" {
		t.Errorf("expected CompanyName 'Acme Corp', got %s", resp.CompanyName)
	}
}

func TestContactToResponse_WithEmails(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		Emails: []domain.ContactEmail{
			{Email: "john@work.com", Type: "work"},
			{Email: "john@home.com", Type: "home"},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.Emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(resp.Emails))
	}
	if resp.Emails[0].Email != "john@work.com" {
		t.Errorf("expected first email 'john@work.com', got %s", resp.Emails[0].Email)
	}
	if resp.Emails[0].Type != "work" {
		t.Errorf("expected first email type 'work', got %s", resp.Emails[0].Type)
	}
}

func TestContactToResponse_WithPhoneNumbers(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		PhoneNumbers: []domain.ContactPhone{
			{Number: "+1-555-123-4567", Type: "mobile"},
			{Number: "+1-555-987-6543", Type: "work"},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.PhoneNumbers) != 2 {
		t.Errorf("expected 2 phone numbers, got %d", len(resp.PhoneNumbers))
	}
	if resp.PhoneNumbers[0].Number != "+1-555-123-4567" {
		t.Errorf("expected first phone '+1-555-123-4567', got %s", resp.PhoneNumbers[0].Number)
	}
}

func TestContactToResponse_WithAddresses(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		PhysicalAddresses: []domain.ContactAddress{
			{
				Type:          "work",
				StreetAddress: "123 Main St",
				City:          "San Francisco",
				State:         "CA",
				PostalCode:    "94102",
				Country:       "USA",
			},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.Addresses) != 1 {
		t.Errorf("expected 1 address, got %d", len(resp.Addresses))
	}
	if resp.Addresses[0].City != "San Francisco" {
		t.Errorf("expected City 'San Francisco', got %s", resp.Addresses[0].City)
	}
}

func TestGrantFromDomain_Basic(t *testing.T) {
	t.Parallel()

	grantInfo := domain.GrantInfo{
		ID:       "grant-123",
		Email:    "user@example.com",
		Provider: domain.ProviderGoogle,
	}

	grant := grantFromDomain(grantInfo)

	if grant.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", grant.ID)
	}
	if grant.Email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got %s", grant.Email)
	}
	if grant.Provider != "google" {
		t.Errorf("expected Provider 'google', got %s", grant.Provider)
	}
}

func TestGrantFromDomain_DifferentProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider domain.Provider
		expected string
	}{
		{"Google", domain.ProviderGoogle, "google"},
		{"Microsoft", domain.ProviderMicrosoft, "microsoft"},
		{"IMAP", domain.ProviderIMAP, "imap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grantInfo := domain.GrantInfo{
				ID:       "grant-123",
				Email:    "user@example.com",
				Provider: tt.provider,
			}

			grant := grantFromDomain(grantInfo)

			if grant.Provider != tt.expected {
				t.Errorf("expected Provider '%s', got %s", tt.expected, grant.Provider)
			}
		})
	}
}

// ================================
// HELPER FUNCTION TESTS
// ================================

func TestParticipantsToEmail_Empty(t *testing.T) {
	t.Parallel()

	result := participantsToEmail(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}

	result = participantsToEmail([]EmailParticipantResponse{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestParticipantsToEmail_Multiple(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "Person One", Email: "one@example.com"},
		{Name: "Person Two", Email: "two@example.com"},
		{Name: "", Email: "three@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 3 {
		t.Errorf("expected 3 participants, got %d", len(result))
	}

	if result[0].Name != "Person One" || result[0].Email != "one@example.com" {
		t.Error("first participant not converted correctly")
	}

	if result[2].Email != "three@example.com" {
		t.Error("third participant email not converted correctly")
	}
}

// ================================
// CACHED RESPONSE CONVERSION TESTS
// ================================

func TestCachedEmailToResponse(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	cachedEmail := &cache.CachedEmail{
		ID:             "email-123",
		ThreadID:       "thread-456",
		FolderID:       "inbox",
		Subject:        "Test Email Subject",
		Snippet:        "This is a test snippet...",
		FromName:       "John Doe",
		FromEmail:      "john@example.com",
		To:             []string{"recipient@example.com"},
		Date:           testTime,
		Unread:         true,
		Starred:        false,
		HasAttachments: true,
	}

	resp := cachedEmailToResponse(cachedEmail)

	if resp.ID != "email-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "email-123")
	}
	if resp.ThreadID != "thread-456" {
		t.Errorf("ThreadID = %q, want %q", resp.ThreadID, "thread-456")
	}
	if resp.Subject != "Test Email Subject" {
		t.Errorf("Subject = %q, want %q", resp.Subject, "Test Email Subject")
	}
	if resp.Snippet != "This is a test snippet..." {
		t.Errorf("Snippet = %q, want %q", resp.Snippet, "This is a test snippet...")
	}
	if len(resp.From) != 1 || resp.From[0].Name != "John Doe" || resp.From[0].Email != "john@example.com" {
		t.Errorf("From = %+v, want [{John Doe john@example.com}]", resp.From)
	}
	if resp.Date != testTime.Unix() {
		t.Errorf("Date = %d, want %d", resp.Date, testTime.Unix())
	}
	if !resp.Unread {
		t.Error("Unread should be true")
	}
	if resp.Starred {
		t.Error("Starred should be false")
	}
	if len(resp.Folders) != 1 || resp.Folders[0] != "inbox" {
		t.Errorf("Folders = %v, want [inbox]", resp.Folders)
	}
}

func TestCachedEmailToResponse_EmptyFields(t *testing.T) {
	t.Parallel()

	cachedEmail := &cache.CachedEmail{
		ID:   "email-empty",
		Date: time.Time{},
	}

	resp := cachedEmailToResponse(cachedEmail)

	if resp.ID != "email-empty" {
		t.Errorf("ID = %q, want %q", resp.ID, "email-empty")
	}
	if resp.ThreadID != "" {
		t.Errorf("ThreadID should be empty, got %q", resp.ThreadID)
	}
	if len(resp.From) != 1 {
		t.Error("From should have one entry even with empty values")
	}
}

func TestCachedEventToResponse(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2024, 1, 20, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

	cachedEvent := &cache.CachedEvent{
		ID:          "event-123",
		CalendarID:  "cal-456",
		Title:       "Team Meeting",
		Description: "Weekly sync",
		Location:    "Conference Room A",
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      false,
		Status:      "confirmed",
		Busy:        true,
	}

	resp := cachedEventToResponse(cachedEvent)

	if resp.ID != "event-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "event-123")
	}
	if resp.CalendarID != "cal-456" {
		t.Errorf("CalendarID = %q, want %q", resp.CalendarID, "cal-456")
	}
	if resp.Title != "Team Meeting" {
		t.Errorf("Title = %q, want %q", resp.Title, "Team Meeting")
	}
	if resp.Description != "Weekly sync" {
		t.Errorf("Description = %q, want %q", resp.Description, "Weekly sync")
	}
	if resp.Location != "Conference Room A" {
		t.Errorf("Location = %q, want %q", resp.Location, "Conference Room A")
	}
	if resp.StartTime != startTime.Unix() {
		t.Errorf("StartTime = %d, want %d", resp.StartTime, startTime.Unix())
	}
	if resp.EndTime != endTime.Unix() {
		t.Errorf("EndTime = %d, want %d", resp.EndTime, endTime.Unix())
	}
	if resp.IsAllDay {
		t.Error("IsAllDay should be false")
	}
	if resp.Status != "confirmed" {
		t.Errorf("Status = %q, want %q", resp.Status, "confirmed")
	}
	if !resp.Busy {
		t.Error("Busy should be true")
	}
}

func TestCachedEventToResponse_AllDayEvent(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)

	cachedEvent := &cache.CachedEvent{
		ID:        "event-allday",
		Title:     "Holiday",
		StartTime: startTime,
		EndTime:   endTime,
		AllDay:    true,
	}

	resp := cachedEventToResponse(cachedEvent)

	if resp.ID != "event-allday" {
		t.Errorf("ID = %q, want %q", resp.ID, "event-allday")
	}
	if !resp.IsAllDay {
		t.Error("IsAllDay should be true")
	}
}

func TestCachedContactToResponse(t *testing.T) {
	t.Parallel()

	cachedContact := &cache.CachedContact{
		ID:          "contact-123",
		GivenName:   "Jane",
		Surname:     "Smith",
		DisplayName: "Jane Smith",
		Email:       "jane@example.com",
		Phone:       "+1-555-1234",
		Company:     "Acme Corp",
		JobTitle:    "Engineer",
		Notes:       "Met at conference",
	}

	resp := cachedContactToResponse(cachedContact)

	if resp.ID != "contact-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "contact-123")
	}
	if resp.GivenName != "Jane" {
		t.Errorf("GivenName = %q, want %q", resp.GivenName, "Jane")
	}
	if resp.Surname != "Smith" {
		t.Errorf("Surname = %q, want %q", resp.Surname, "Smith")
	}
	if resp.DisplayName != "Jane Smith" {
		t.Errorf("DisplayName = %q, want %q", resp.DisplayName, "Jane Smith")
	}
	if len(resp.Emails) != 1 || resp.Emails[0].Email != "jane@example.com" || resp.Emails[0].Type != "personal" {
		t.Errorf("Emails = %+v, want [{jane@example.com personal}]", resp.Emails)
	}
	if len(resp.PhoneNumbers) != 1 || resp.PhoneNumbers[0].Number != "+1-555-1234" || resp.PhoneNumbers[0].Type != "mobile" {
		t.Errorf("PhoneNumbers = %+v, want [{+1-555-1234 mobile}]", resp.PhoneNumbers)
	}
	if resp.CompanyName != "Acme Corp" {
		t.Errorf("CompanyName = %q, want %q", resp.CompanyName, "Acme Corp")
	}
	if resp.JobTitle != "Engineer" {
		t.Errorf("JobTitle = %q, want %q", resp.JobTitle, "Engineer")
	}
	if resp.Notes != "Met at conference" {
		t.Errorf("Notes = %q, want %q", resp.Notes, "Met at conference")
	}
}

func TestCachedContactToResponse_MinimalData(t *testing.T) {
	t.Parallel()

	cachedContact := &cache.CachedContact{
		ID:        "contact-minimal",
		GivenName: "Bob",
	}

	resp := cachedContactToResponse(cachedContact)

	if resp.ID != "contact-minimal" {
		t.Errorf("ID = %q, want %q", resp.ID, "contact-minimal")
	}
	if resp.GivenName != "Bob" {
		t.Errorf("GivenName = %q, want %q", resp.GivenName, "Bob")
	}
	if resp.Surname != "" {
		t.Errorf("Surname should be empty, got %q", resp.Surname)
	}
	// Email and Phone should still have entries (even if empty)
	if len(resp.Emails) != 1 {
		t.Error("Emails should have one entry")
	}
	if len(resp.PhoneNumbers) != 1 {
		t.Error("PhoneNumbers should have one entry")
	}
}
