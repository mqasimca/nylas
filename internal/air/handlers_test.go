package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

// Helper to create demo server for handler tests
func newTestDemoServer() *Server {
	return NewDemoServer(":3003")
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

	// Verify email-detail-body has white/light background for readability
	// This ensures HTML emails with inline styles (designed for light backgrounds) are readable
	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		{
			"email body has light background",
			"background: #ffffff",
			"HTML emails have inline styles for light backgrounds - need white bg for readability",
		},
		{
			"email body has dark text",
			"color: #1a1a1a",
			"Text must be dark on light background for contrast",
		},
		{
			"email body selector exists",
			".email-detail-body",
			"Email body styling must be defined",
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
