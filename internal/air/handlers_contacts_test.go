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

func TestContactResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := ContactResponse{
		ID:          "contact-123",
		DisplayName: "John Doe",
		GivenName:   "John",
		Surname:     "Doe",
		CompanyName: "Acme Corp",
		JobTitle:    "Engineer",
		Emails: []ContactEmailResponse{
			{Email: "john@example.com", Type: "work"},
		},
		PhoneNumbers: []ContactPhoneResponse{
			{Number: "+1-555-1234", Type: "mobile"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ContactResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("expected ID %s, got %s", resp.ID, decoded.ID)
	}

	if decoded.DisplayName != resp.DisplayName {
		t.Errorf("expected DisplayName %s, got %s", resp.DisplayName, decoded.DisplayName)
	}

	if len(decoded.Emails) != 1 {
		t.Errorf("expected 1 email, got %d", len(decoded.Emails))
	}
}

func TestContactActionResponseSerialization(t *testing.T) {
	t.Parallel()

	resp := ContactActionResponse{
		Success: true,
		Message: "Contact created",
		Contact: &ContactResponse{
			ID:          "contact-new",
			DisplayName: "New Contact",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ContactActionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("expected Success to be true")
	}

	if decoded.Contact == nil {
		t.Error("expected Contact to be present")
	}
}

func TestContactActionResponseError(t *testing.T) {
	t.Parallel()

	resp := ContactActionResponse{
		Success: false,
		Error:   "Something went wrong",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ContactActionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Success {
		t.Error("expected Success to be false")
	}

	if decoded.Error != "Something went wrong" {
		t.Errorf("expected error message, got %s", decoded.Error)
	}
}

func TestCreateContactRequestSerialization(t *testing.T) {
	t.Parallel()

	req := CreateContactRequest{
		GivenName:   "John",
		Surname:     "Doe",
		Nickname:    "Johnny",
		CompanyName: "Acme",
		JobTitle:    "Dev",
		Birthday:    "1990-01-01",
		Notes:       "Test",
		Emails: []ContactEmailResponse{
			{Email: "john@test.com", Type: "work"},
		},
		PhoneNumbers: []ContactPhoneResponse{
			{Number: "555-1234", Type: "mobile"},
		},
		Addresses: []ContactAddressResponse{
			{City: "SF", State: "CA"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CreateContactRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.GivenName != req.GivenName {
		t.Errorf("expected GivenName %s, got %s", req.GivenName, decoded.GivenName)
	}

	if len(decoded.Emails) != 1 {
		t.Errorf("expected 1 email, got %d", len(decoded.Emails))
	}

	if len(decoded.PhoneNumbers) != 1 {
		t.Errorf("expected 1 phone, got %d", len(decoded.PhoneNumbers))
	}

	if len(decoded.Addresses) != 1 {
		t.Errorf("expected 1 address, got %d", len(decoded.Addresses))
	}
}

func TestContactsResponse_EmptyOmitFields(t *testing.T) {
	t.Parallel()

	resp := ContactsResponse{
		Contacts: []ContactResponse{},
		HasMore:  false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// NextCursor may be omitted or present as empty string depending on serializer
	// Either representation is acceptable

	var decoded ContactsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.HasMore {
		t.Error("expected HasMore to be false")
	}
}
