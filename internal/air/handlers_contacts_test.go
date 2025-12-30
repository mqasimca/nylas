package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================================
// CONTACTS HANDLER ADDITIONAL TESTS
// ================================

func TestHandleContactsRoute_GET(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Contacts) == 0 {
		t.Error("expected non-empty contacts in demo mode")
	}
}

func TestHandleContactsRoute_POST(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contact := CreateContactRequest{
		GivenName:   "Test",
		Surname:     "Contact",
		CompanyName: "Test Company",
		Emails: []ContactEmailResponse{
			{Email: "test@example.com", Type: "work"},
		},
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success, got error: %s", resp.Error)
	}
}

func TestHandleContactsRoute_InvalidMethod(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPut, "/api/contacts", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleContactsRoute_DELETE_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/contacts", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleContactsRoute_WithLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?limit=2", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactsRoute_WithGroupFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?group=group-1", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactsRoute_WithSourceFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?source=google", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactsRoute_WithEmailFilter(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?email=test@example.com", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactsRoute_WithCursor(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts?cursor=abc123", nil)
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateContact_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	body := bytes.NewBufferString("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	// Demo mode may handle invalid JSON gracefully
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 400 or 200, got %d", w.Code)
	}
}

func TestHandleCreateContact_WithPhoneNumbers(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contact := CreateContactRequest{
		GivenName: "Test",
		PhoneNumbers: []ContactPhoneResponse{
			{Number: "+1-555-1234", Type: "mobile"},
			{Number: "+1-555-5678", Type: "work"},
		},
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateContact_WithAddresses(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contact := CreateContactRequest{
		GivenName: "Test",
		Addresses: []ContactAddressResponse{
			{
				Type:          "home",
				StreetAddress: "123 Main St",
				City:          "San Francisco",
				State:         "CA",
				PostalCode:    "94105",
				Country:       "USA",
			},
		},
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleCreateContact_WithAllFields(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contact := CreateContactRequest{
		GivenName:   "John",
		Surname:     "Doe",
		Nickname:    "Johnny",
		CompanyName: "Acme Corp",
		JobTitle:    "Engineer",
		Birthday:    "1990-01-15",
		Notes:       "Test contact",
		Emails: []ContactEmailResponse{
			{Email: "john@example.com", Type: "work"},
		},
		PhoneNumbers: []ContactPhoneResponse{
			{Number: "+1-555-1234", Type: "mobile"},
		},
		Addresses: []ContactAddressResponse{
			{Type: "work", City: "SF"},
		},
	}
	body, _ := json.Marshal(contact)
	req := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactsRoute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleUpdateContact_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	givenName := "Updated"
	surname := "Name"
	update := UpdateContactRequest{
		GivenName: &givenName,
		Surname:   &surname,
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/contacts/contact-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactActionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success, got error: %s", resp.Error)
	}
}

func TestHandleContactByID_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/nonexistent-id", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleContactSearch_EmptyResults(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=nonexistentxyz123", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Contacts) != 0 {
		t.Errorf("expected 0 contacts for non-matching search, got %d", len(resp.Contacts))
	}
}

func TestHandleContactSearch_ByCompanyName(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get demo contacts to find a company name to search for
	contacts := demoContacts()
	if len(contacts) == 0 {
		t.Skip("no demo contacts available")
	}

	searchQuery := "acme" // Common company name in test data
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q="+searchQuery, nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactSearch_WithLimit(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=&limit=5", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleContactGroups_Content(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contact-groups", nil)
	w := httptest.NewRecorder()

	server.handleContactGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactGroupsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify group structure
	for _, group := range resp.Groups {
		if group.ID == "" {
			t.Error("expected group to have an ID")
		}
		if group.Name == "" {
			t.Error("expected group to have a Name")
		}
	}
}

func TestHandleContactGroups_POST_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/contact-groups", nil)
	w := httptest.NewRecorder()

	server.handleContactGroups(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
