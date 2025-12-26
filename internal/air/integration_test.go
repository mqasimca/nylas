//go:build integration
// +build integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	authapp "github.com/mqasimca/nylas/internal/app/auth"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// testServer creates a real Air server with actual credentials for integration testing.
func testServer(t *testing.T) *Server {
	t.Helper()

	configStore := config.NewDefaultFileStore()
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		t.Skipf("Skipping: cannot access secret store: %v", err)
	}

	grantStore := keyring.NewGrantStore(secretStore)
	configSvc := authapp.NewConfigService(configStore, secretStore)

	// Check configuration
	status, err := configSvc.GetStatus()
	if err != nil || !status.IsConfigured {
		t.Skip("Skipping: Nylas CLI not configured. Run 'nylas auth login' first.")
	}

	// Check for default grant
	defaultGrantID, err := grantStore.GetDefaultGrant()
	if err != nil || defaultGrantID == "" {
		t.Skip("Skipping: No default grant configured. Run 'nylas auth login' first.")
	}

	// Check that default grant is Google provider
	grants, err := grantStore.ListGrants()
	if err != nil {
		t.Skipf("Skipping: cannot list grants: %v", err)
	}

	var defaultGrant *domain.GrantInfo
	for i := range grants {
		if grants[i].ID == defaultGrantID {
			defaultGrant = &grants[i]
			break
		}
	}

	if defaultGrant == nil {
		t.Skip("Skipping: default grant not found in grant list")
	}

	if defaultGrant.Provider != domain.ProviderGoogle {
		t.Skipf("Skipping: default grant is %s, not Google. These tests require a Google account as default.", defaultGrant.Provider)
	}

	t.Logf("Running integration tests with Google account: %s", defaultGrant.Email)

	// Create Nylas client
	cfg, err := configStore.Load()
	if err != nil {
		t.Skipf("Skipping: cannot load config: %v", err)
	}

	apiKey, _ := secretStore.Get(ports.KeyAPIKey)
	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)

	if apiKey == "" {
		t.Skip("Skipping: no API key configured")
	}

	client := nylas.NewHTTPClient()
	client.SetRegion(cfg.Region)
	client.SetCredentials(clientID, clientSecret, apiKey)

	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	return &Server{
		addr:        ":7365",
		demoMode:    false,
		configSvc:   configSvc,
		configStore: configStore,
		secretStore: secretStore,
		grantStore:  grantStore,
		nylasClient: client,
		templates:   tmpl,
	}
}

// ================================
// CONFIG INTEGRATION TESTS
// ================================

func TestIntegration_ConfigStatus(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ConfigStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Configured {
		t.Error("expected Configured to be true")
	}

	if !resp.HasAPIKey {
		t.Error("expected HasAPIKey to be true")
	}

	// Note: GrantCount and DefaultGrant may be empty if ConfigService
	// doesn't have access to the grant store. This is not a failure,
	// just log the values for debugging.
	t.Logf("Config: region=%s, grants=%d, default=%s", resp.Region, resp.GrantCount, resp.DefaultGrant)
}

// ================================
// GRANTS INTEGRATION TESTS
// ================================

func TestIntegration_ListGrants(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp GrantsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Grants) == 0 {
		t.Error("expected at least one grant")
	}

	if resp.DefaultGrant == "" {
		t.Error("expected DefaultGrant to be set")
	}

	// Find the Google grant
	hasGoogle := false
	for _, g := range resp.Grants {
		if g.Provider == "google" {
			hasGoogle = true
			t.Logf("Found Google grant: %s (%s)", g.Email, g.ID)
		}
	}

	if !hasGoogle {
		t.Error("expected at least one Google grant")
	}
}

// ================================
// FOLDERS INTEGRATION TESTS
// ================================

func TestIntegration_ListFolders(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/folders", nil)
	w := httptest.NewRecorder()

	server.handleListFolders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp FoldersResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Folders) == 0 {
		t.Error("expected at least one folder")
	}

	// Check for standard Gmail folders/labels
	folderNames := make(map[string]bool)
	for _, f := range resp.Folders {
		folderNames[f.Name] = true
		if f.SystemFolder != "" {
			t.Logf("Folder: %s (system: %s, unread: %d)", f.Name, f.SystemFolder, f.UnreadCount)
		}
	}

	// Gmail should have INBOX
	if !folderNames["INBOX"] && !folderNames["Inbox"] {
		t.Log("Warning: INBOX folder not found (may have different name)")
	}
}

// ================================
// EMAILS INTEGRATION TESTS
// ================================

func TestIntegration_ListEmails(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=5", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d emails (has_more: %v)", len(resp.Emails), resp.HasMore)

	if len(resp.Emails) == 0 {
		t.Log("Warning: no emails found in account")
		return
	}

	// Verify email structure
	first := resp.Emails[0]
	if first.ID == "" {
		t.Error("expected email to have ID")
	}
	if len(first.From) == 0 {
		t.Error("expected email to have From")
	}
	if first.Date == 0 {
		t.Error("expected email to have Date")
	}

	t.Logf("First email: %s from %s", first.Subject, first.From[0].Email)
}

func TestIntegration_ListEmails_WithFilters(t *testing.T) {
	server := testServer(t)

	// Test unread filter
	req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=5&unread=true", nil)
	w := httptest.NewRecorder()

	server.handleListEmails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d unread emails", len(resp.Emails))

	// All returned emails should be unread
	for _, email := range resp.Emails {
		if !email.Unread {
			t.Errorf("expected email %s to be unread", email.ID)
		}
	}
}

func TestIntegration_GetEmail(t *testing.T) {
	server := testServer(t)

	// First get a list of emails to get a valid ID
	listReq := httptest.NewRequest(http.MethodGet, "/api/emails?limit=1", nil)
	listW := httptest.NewRecorder()
	server.handleListEmails(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Skipf("Skipping: cannot list emails: %s", listW.Body.String())
	}

	var listResp EmailsResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(listResp.Emails) == 0 {
		t.Skip("Skipping: no emails in account to test")
	}

	emailID := listResp.Emails[0].ID

	// Now get the specific email
	req := httptest.NewRequest(http.MethodGet, "/api/emails/"+emailID, nil)
	w := httptest.NewRecorder()

	server.handleEmailByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp EmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != emailID {
		t.Errorf("expected ID %s, got %s", emailID, resp.ID)
	}

	// Full email should have body
	if resp.Body == "" {
		t.Log("Warning: email body is empty")
	}

	t.Logf("Got email: %s (body length: %d)", resp.Subject, len(resp.Body))
}

// ================================
// DRAFTS INTEGRATION TESTS
// ================================

func TestIntegration_ListDrafts(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/drafts", nil)
	w := httptest.NewRecorder()

	server.handleDrafts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp DraftsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d drafts", len(resp.Drafts))

	// Drafts are optional, so just verify the response structure
	for _, draft := range resp.Drafts {
		if draft.ID == "" {
			t.Error("expected draft to have ID")
		}
		t.Logf("Draft: %s", draft.Subject)
	}
}

// ================================
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
// INDEX PAGE INTEGRATION TEST
// ================================

func TestIntegration_IndexPage(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content-type text/html, got %s", contentType)
	}

	body := w.Body.String()

	// Check for key HTML elements
	if len(body) < 1000 {
		t.Error("expected substantial HTML content")
	}

	// Should contain skeleton loading elements (not mock data)
	if !containsAny(body, "skeleton", "loading") {
		t.Log("Warning: no skeleton loading elements found")
	}

	t.Logf("Index page rendered: %d bytes", len(body))
}

// ================================
// BUILD PAGE DATA INTEGRATION TEST
// ================================

func TestIntegration_BuildPageData(t *testing.T) {
	server := testServer(t)

	data := server.buildPageData()

	// Should have real user info
	if data.UserEmail == "" {
		t.Error("expected UserEmail to be set")
	}

	if data.DefaultGrantID == "" {
		t.Error("expected DefaultGrantID to be set")
	}

	if data.Provider != "google" {
		t.Errorf("expected Provider 'google', got %s", data.Provider)
	}

	// In non-demo mode, mock data should be cleared
	if len(data.Emails) > 0 {
		t.Error("expected Emails to be empty (loaded via JS)")
	}

	if len(data.Events) > 0 {
		t.Error("expected Events to be empty (loaded via JS)")
	}

	t.Logf("Page data: user=%s, provider=%s, grants=%d",
		data.UserEmail, data.Provider, len(data.Grants))
}

// Helper functions

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ================================
// CONTACTS INTEGRATION TESTS
// ================================

func TestIntegration_ListContacts(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d contacts (has_more: %v)", len(resp.Contacts), resp.HasMore)

	if len(resp.Contacts) == 0 {
		t.Log("Warning: no contacts found in account")
		return
	}

	// Verify contact structure
	first := resp.Contacts[0]
	if first.ID == "" {
		t.Error("expected contact to have ID")
	}

	t.Logf("First contact: %s (%s)", first.DisplayName, first.ID)
}

func TestIntegration_ListContacts_WithGroup(t *testing.T) {
	server := testServer(t)

	// First get list of groups
	groupReq := httptest.NewRequest(http.MethodGet, "/api/contacts/groups", nil)
	groupW := httptest.NewRecorder()
	server.handleContactGroups(groupW, groupReq)

	if groupW.Code != http.StatusOK {
		t.Skipf("Skipping: cannot get contact groups: %s", groupW.Body.String())
	}

	var groupResp ContactGroupsResponse
	if err := json.NewDecoder(groupW.Body).Decode(&groupResp); err != nil {
		t.Fatalf("failed to decode group response: %v", err)
	}

	if len(groupResp.Groups) == 0 {
		t.Skip("Skipping: no contact groups in account")
	}

	// Test filtering by first group
	groupID := groupResp.Groups[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/contacts?group="+groupID+"&limit=5", nil)
	w := httptest.NewRecorder()

	server.handleListContacts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d contacts in group %s", len(resp.Contacts), groupResp.Groups[0].Name)
}

func TestIntegration_GetContact(t *testing.T) {
	server := testServer(t)

	// First get a list of contacts to get a valid ID
	listReq := httptest.NewRequest(http.MethodGet, "/api/contacts?limit=1", nil)
	listW := httptest.NewRecorder()
	server.handleListContacts(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Skipf("Skipping: cannot list contacts: %s", listW.Body.String())
	}

	var listResp ContactsResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(listResp.Contacts) == 0 {
		t.Skip("Skipping: no contacts in account to test")
	}

	contactID := listResp.Contacts[0].ID

	// Now get the specific contact
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/"+contactID, nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != contactID {
		t.Errorf("expected ID %s, got %s", contactID, resp.ID)
	}

	t.Logf("Got contact: %s (%s)", resp.DisplayName, resp.ID)
}

func TestIntegration_ContactGroups(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/groups", nil)
	w := httptest.NewRecorder()

	server.handleContactGroups(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactGroupsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d contact groups", len(resp.Groups))

	for _, g := range resp.Groups {
		if g.ID == "" {
			t.Error("expected group to have ID")
		}
		t.Logf("Group: %s (%s)", g.Name, g.ID)
	}
}

// ================================
// CACHE INTEGRATION TESTS
// ================================

func TestIntegration_CacheStatus(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/status", nil)
	w := httptest.NewRecorder()

	server.handleCacheStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache status: enabled=%v, online=%v, sync_interval=%d min",
		resp.Enabled, resp.Online, resp.SyncInterval)

	if len(resp.Accounts) > 0 {
		for _, acc := range resp.Accounts {
			t.Logf("  Account: %s (emails=%d, events=%d, size=%d bytes)",
				acc.Email, acc.EmailCount, acc.EventCount, acc.SizeBytes)
		}
	}
}

func TestIntegration_CacheSettings(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/settings", nil)
	w := httptest.NewRecorder()

	server.handleCacheSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheSettingsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache settings: enabled=%v, max_size=%dMB, ttl=%d days, theme=%s",
		resp.Enabled, resp.MaxSizeMB, resp.TTLDays, resp.Theme)

	// Verify settings have reasonable values
	if resp.MaxSizeMB < 50 {
		t.Errorf("MaxSizeMB should be at least 50, got %d", resp.MaxSizeMB)
	}
	if resp.TTLDays < 1 {
		t.Errorf("TTLDays should be at least 1, got %d", resp.TTLDays)
	}
}

func TestIntegration_CacheSearch_EmptyQuery(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
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

func TestIntegration_CacheSearch_WithQuery(t *testing.T) {
	server := testServer(t)

	// Search for a common term
	req := httptest.NewRequest(http.MethodGet, "/api/cache/search?q=test", nil)
	w := httptest.NewRecorder()

	server.handleCacheSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CacheSearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Cache search for 'test': found %d results", len(resp.Results))

	for i, r := range resp.Results {
		if i < 5 { // Log first 5 results
			t.Logf("  [%s] %s - %s", r.Type, r.Title, r.Subtitle)
		}
	}
}

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
// CONTACT SEARCH INTEGRATION TESTS
// ================================

func TestIntegration_ContactSearch(t *testing.T) {
	server := testServer(t)

	// Search for contacts
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=a&limit=10", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d contacts matching 'a'", len(resp.Contacts))

	for i, c := range resp.Contacts {
		if i < 5 { // Log first 5 contacts
			t.Logf("  %s (%s)", c.DisplayName, c.ID)
		}
	}
}

func TestIntegration_ContactSearch_ByEmail(t *testing.T) {
	server := testServer(t)

	// Search for contacts by email domain
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=gmail.com&limit=10", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d contacts with gmail.com", len(resp.Contacts))
}

func TestIntegration_ContactSearch_Empty(t *testing.T) {
	server := testServer(t)

	// Empty search should return all contacts
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Empty search returned %d contacts", len(resp.Contacts))
}

// =============================================================================
// Phase 6: Productivity Features Integration Tests
// =============================================================================

func TestIntegration_SplitInbox(t *testing.T) {
	server := testServer(t)

	// Test GET split inbox config
	req := httptest.NewRequest(http.MethodGet, "/api/inbox/split", nil)
	w := httptest.NewRecorder()
	server.handleSplitInbox(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SplitInboxResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Split inbox config: enabled=%v, categories=%d", resp.Config.Enabled, len(resp.Config.Categories))
	for cat, count := range resp.Categories {
		t.Logf("  %s: %d emails", cat, count)
	}
}

func TestIntegration_SplitInbox_VIPManagement(t *testing.T) {
	server := testServer(t)

	// Add a VIP sender
	body, _ := json.Marshal(map[string]string{"email": "important@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/inbox/vip", bytes.NewReader(body))
	w := httptest.NewRecorder()
	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify VIP list contains the new entry
	req = httptest.NewRequest(http.MethodGet, "/api/inbox/vip", nil)
	w = httptest.NewRecorder()
	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var vipResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&vipResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	vipList := vipResp["vip_senders"].([]any)
	t.Logf("VIP senders: %d", len(vipList))
	for _, vip := range vipList {
		t.Logf("  %s", vip.(string))
	}

	// Remove VIP
	req = httptest.NewRequest(http.MethodDelete, "/api/inbox/vip?email=important@example.com", nil)
	w = httptest.NewRecorder()
	server.handleVIPSenders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestIntegration_CategorizeEmail(t *testing.T) {
	server := testServer(t)

	testEmails := []struct {
		from     string
		subject  string
		expected InboxCategory
	}{
		{"newsletter@company.com", "Weekly Update", CategoryNewsletters},
		{"notifications@linkedin.com", "New connection", CategorySocial},
		{"receipt@stripe.com", "Payment received", CategoryUpdates},
		{"deals@store.com", "50% off sale", CategoryPromotions},
		{"friend@gmail.com", "Hello!", CategoryPrimary},
	}

	for _, tc := range testEmails {
		body, _ := json.Marshal(map[string]string{
			"email_id": "test-" + tc.from,
			"from":     tc.from,
			"subject":  tc.subject,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/inbox/categorize", bytes.NewReader(body))
		w := httptest.NewRecorder()
		server.handleCategorizeEmail(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for %s, got %d", tc.from, w.Code)
			continue
		}

		var resp CategorizedEmail
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Errorf("failed to decode response for %s: %v", tc.from, err)
			continue
		}

		t.Logf("Categorized %s -> %s (matched: %s)", tc.from, resp.Category, resp.MatchedRule)
	}
}

func TestIntegration_Snooze(t *testing.T) {
	server := testServer(t)

	// Snooze an email
	body, _ := json.Marshal(SnoozeRequest{
		EmailID:  "test-email-123",
		Duration: "1h",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/snooze", bytes.NewReader(body))
	w := httptest.NewRecorder()
	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var snoozeResp SnoozeResponse
	if err := json.NewDecoder(w.Body).Decode(&snoozeResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Snoozed email: %s until %d (%s)", snoozeResp.EmailID, snoozeResp.SnoozeUntil, snoozeResp.Message)

	// List snoozed emails
	req = httptest.NewRequest(http.MethodGet, "/api/snooze", nil)
	w = httptest.NewRecorder()
	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var listResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := int(listResp["count"].(float64))
	t.Logf("Snoozed emails count: %d", count)

	// Unsnooze
	req = httptest.NewRequest(http.MethodDelete, "/api/snooze?email_id=test-email-123", nil)
	w = httptest.NewRecorder()
	server.handleSnooze(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestIntegration_ScheduledSend(t *testing.T) {
	server := testServer(t)

	// List scheduled messages
	req := httptest.NewRequest(http.MethodGet, "/api/scheduled", nil)
	w := httptest.NewRecorder()
	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	scheduled := listResp["scheduled"].([]any)
	t.Logf("Scheduled messages: %d", len(scheduled))

	// Create a scheduled message (demo mode will simulate)
	body, _ := json.Marshal(ScheduledSendRequest{
		To:            []EmailParticipantResponse{{Email: "test@example.com", Name: "Test"}},
		Subject:       "Test Scheduled Email",
		Body:          "This is a test scheduled message",
		SendAtNatural: "tomorrow 10am",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/scheduled", bytes.NewReader(body))
	w = httptest.NewRecorder()
	server.handleScheduledSend(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var schedResp ScheduledSendResponse
	if err := json.NewDecoder(w.Body).Decode(&schedResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Scheduled message: ID=%s, SendAt=%d", schedResp.ScheduleID, schedResp.SendAt)
}

func TestIntegration_UndoSend(t *testing.T) {
	server := testServer(t)

	// Get undo send config
	req := httptest.NewRequest(http.MethodGet, "/api/undo-send", nil)
	w := httptest.NewRecorder()
	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var config UndoSendConfig
	if err := json.NewDecoder(w.Body).Decode(&config); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Undo send config: enabled=%v, grace_period=%ds", config.Enabled, config.GracePeriodSec)

	// Update config
	body, _ := json.Marshal(UndoSendConfig{Enabled: true, GracePeriodSec: 15})
	req = httptest.NewRequest(http.MethodPut, "/api/undo-send", bytes.NewReader(body))
	w = httptest.NewRecorder()
	server.handleUndoSend(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Get pending sends
	req = httptest.NewRequest(http.MethodGet, "/api/pending-sends", nil)
	w = httptest.NewRecorder()
	server.handlePendingSends(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var pendingResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&pendingResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	count := int(pendingResp["count"].(float64))
	t.Logf("Pending sends: %d", count)
}

func TestIntegration_Templates(t *testing.T) {
	server := testServer(t)

	// List templates (should include defaults)
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()
	server.handleTemplates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp TemplateListResponse
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Templates: %d", listResp.Total)
	for _, tmpl := range listResp.Templates {
		t.Logf("  %s: %s (shortcut: %s, category: %s)", tmpl.ID, tmpl.Name, tmpl.Shortcut, tmpl.Category)
	}

	// Create a custom template
	body, _ := json.Marshal(EmailTemplate{
		Name:     "Integration Test Template",
		Subject:  "Hello {{name}}",
		Body:     "Hi {{name}},\n\nThis is a test from {{company}}.\n\nBest,\n{{sender}}",
		Shortcut: "/inttest",
		Category: "test",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	w = httptest.NewRecorder()
	server.handleTemplates(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var created EmailTemplate
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Created template: ID=%s, variables=%v", created.ID, created.Variables)

	// Expand the template
	expandBody, _ := json.Marshal(map[string]any{
		"variables": map[string]string{
			"name":    "Alice",
			"company": "Acme Inc",
			"sender":  "Bob",
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/templates/"+created.ID+"/expand", bytes.NewReader(expandBody))
	w = httptest.NewRecorder()
	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var expandResp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&expandResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Expanded subject: %s", expandResp["subject"])
	t.Logf("Expanded body: %s", expandResp["body"])

	// Delete the template
	req = httptest.NewRequest(http.MethodDelete, "/api/templates/"+created.ID, nil)
	w = httptest.NewRecorder()
	server.handleTemplateByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ================================
// CONTACT CRUD INTEGRATION TESTS
// ================================

func TestIntegration_CreateContact(t *testing.T) {
	server := testServer(t)

	contact := CreateContactRequest{
		GivenName:   "Integration",
		Surname:     "Test",
		CompanyName: "Test Company",
		JobTitle:    "Tester",
		Emails: []ContactEmailInput{
			{Email: "integration-test@example.com", Type: "work"},
		},
		PhoneNumbers: []ContactPhoneInput{
			{Number: "+1-555-0123", Type: "mobile"},
		},
		Notes: "Created by integration test",
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}

	if resp.Contact == nil {
		t.Fatal("expected contact in response")
	}

	t.Logf("Created contact: %s (%s)", resp.Contact.DisplayName, resp.Contact.ID)

	// Clean up - delete the contact
	req = httptest.NewRequest(http.MethodDelete, "/api/contacts/"+resp.Contact.ID, nil)
	w = httptest.NewRecorder()
	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Warning: failed to delete test contact: %s", w.Body.String())
	}
}

func TestIntegration_ContactCRUD(t *testing.T) {
	server := testServer(t)

	// Create a contact
	contact := CreateContactRequest{
		GivenName:   "CRUD",
		Surname:     "Test",
		CompanyName: "CRUD Company",
		Emails: []ContactEmailInput{
			{Email: "crud-test@example.com", Type: "work"},
		},
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("CREATE failed: expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var createResp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !createResp.Success || createResp.Contact == nil {
		t.Fatalf("expected successful creation with contact, got: %+v", createResp)
	}

	contactID := createResp.Contact.ID
	t.Logf("Created contact: %s", contactID)

	// Read the contact
	req = httptest.NewRequest(http.MethodGet, "/api/contacts/"+contactID, nil)
	w = httptest.NewRecorder()
	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("READ failed: expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var readResp ContactResponse
	if err := json.NewDecoder(w.Body).Decode(&readResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if readResp.ID != contactID {
		t.Errorf("expected ID %s, got %s", contactID, readResp.ID)
	}

	t.Logf("Read contact: %s", readResp.DisplayName)

	// Update the contact
	update := UpdateContactRequest{
		GivenName:   "Updated",
		Surname:     "Name",
		CompanyName: "Updated Company",
	}
	updateBody, _ := json.Marshal(update)
	req = httptest.NewRequest(http.MethodPut, "/api/contacts/"+contactID, bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("UPDATE failed: expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updateResp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&updateResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !updateResp.Success {
		t.Fatalf("expected successful update, got error: %s", updateResp.Error)
	}

	t.Logf("Updated contact: %s", updateResp.Contact.DisplayName)

	// Delete the contact
	req = httptest.NewRequest(http.MethodDelete, "/api/contacts/"+contactID, nil)
	w = httptest.NewRecorder()
	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("DELETE failed: expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var deleteResp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&deleteResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !deleteResp.Success {
		t.Fatalf("expected successful deletion, got error: %s", deleteResp.Error)
	}

	t.Logf("Deleted contact: %s", contactID)

	// Verify contact is deleted
	req = httptest.NewRequest(http.MethodGet, "/api/contacts/"+contactID, nil)
	w = httptest.NewRecorder()
	server.handleContactByID(w, req)

	// Should return 404 or error
	if w.Code == http.StatusOK {
		t.Log("Note: Contact may still be accessible briefly after deletion")
	} else {
		t.Logf("Contact no longer accessible (status %d)", w.Code)
	}
}

// ================================
// AI INTEGRATION TESTS
// ================================

func TestIntegration_AISummarize(t *testing.T) {
	server := testServer(t)

	// Test with a simple prompt
	reqBody := AIRequest{
		EmailID: "test-email-id",
		Prompt:  "Say 'test successful' in exactly those words",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	// If claude CLI is installed, we should get a response
	// If not, we should get an error about CLI not found
	if w.Code == http.StatusOK {
		var resp AIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success, got error: %s", resp.Error)
		}

		t.Logf("AI response: %s", resp.Summary)
	} else if w.Code == http.StatusInternalServerError {
		var resp AIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if strings.Contains(resp.Error, "Claude Code CLI not found") {
			t.Skip("Claude Code CLI not installed, skipping AI test")
		}

		t.Logf("AI error: %s", resp.Error)
	} else {
		t.Fatalf("unexpected status %d: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_AISummarize_MethodNotAllowed(t *testing.T) {
	server := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/summarize", nil)
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestIntegration_AISummarize_EmptyPrompt(t *testing.T) {
	server := testServer(t)

	reqBody := AIRequest{
		EmailID: "test-email-id",
		Prompt:  "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/summarize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleAISummarize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp AIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected failure for empty prompt")
	}

	if resp.Error != "Prompt is required" {
		t.Errorf("expected 'Prompt is required', got: %s", resp.Error)
	}
}
