//go:build !integration
// +build !integration

package air

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// =============================================================================
// Contacts CRUD Handler Tests
// =============================================================================

// TestHandleGetContact_DemoMode_Found tests retrieving existing contact.
func TestHandleGetContact_DemoMode_Found(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	contacts := demoContacts()
	if len(contacts) == 0 {
		t.Skip("no demo contacts available")
	}

	contactID := contacts[0].ID
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/"+contactID, nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != contactID {
		t.Errorf("expected ID %s, got %s", contactID, resp.ID)
	}
}

// TestHandleGetContact_DemoMode_NotFound tests non-existent contact.
func TestHandleGetContact_DemoMode_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/nonexistent-contact-id", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestHandleUpdateContact_DemoMode_WithFields tests contact update with various fields.
func TestHandleUpdateContact_DemoMode_WithFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		update  UpdateContactRequest
		wantOK  bool
		wantMsg string
	}{
		{
			name: "update name",
			update: func() UpdateContactRequest {
				name := "Updated"
				surname := "Name"
				return UpdateContactRequest{GivenName: &name, Surname: &surname}
			}(),
			wantOK: true,
		},
		{
			name: "update company",
			update: func() UpdateContactRequest {
				company := "New Company"
				return UpdateContactRequest{CompanyName: &company}
			}(),
			wantOK: true,
		},
		{
			name: "update with emails",
			update: UpdateContactRequest{
				Emails: []ContactEmailResponse{
					{Email: "new@example.com", Type: "work"},
				},
			},
			wantOK: true,
		},
		{
			name: "update with phones",
			update: UpdateContactRequest{
				PhoneNumbers: []ContactPhoneResponse{
					{Number: "+1-555-9999", Type: "mobile"},
				},
			},
			wantOK: true,
		},
		{
			name: "update with addresses",
			update: UpdateContactRequest{
				Addresses: []ContactAddressResponse{
					{Type: "home", City: "New York", State: "NY"},
				},
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTestDemoServer()

			body, _ := json.Marshal(tt.update)
			req := httptest.NewRequest(http.MethodPut, "/api/contacts/demo-contact-001", bytes.NewBuffer(body))
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

			if resp.Success != tt.wantOK {
				t.Errorf("expected success=%v, got %v", tt.wantOK, resp.Success)
			}
		})
	}
}

// TestHandleDeleteContact_DemoMode tests contact deletion in demo mode.
func TestHandleDeleteContact_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/contacts/demo-contact-001", nil)
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
		t.Error("expected success to be true")
	}

	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

// TestHandleContactByID_InvalidMethod tests unsupported methods.
func TestHandleContactByID_InvalidMethod(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	methods := []string{http.MethodPatch, http.MethodOptions, http.MethodHead}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/contacts/demo-contact-001", nil)
			w := httptest.NewRecorder()

			server.handleContactByID(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleContactByID_MissingID tests missing contact ID error.
func TestHandleContactByID_MissingID(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandleContactPhoto_DemoMode tests contact photo in demo mode.
func TestHandleContactPhoto_DemoMode(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/demo-contact-001/photo", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check content type is image
	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", contentType)
	}

	// Check cache header
	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("expected Cache-Control header")
	}

	// Body should have PNG data
	if w.Body.Len() == 0 {
		t.Error("expected non-empty body for photo")
	}
}

// TestHandleContactPhoto_InvalidMethod tests photo endpoint with wrong method.
func TestHandleContactPhoto_InvalidMethod(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/contacts/demo-contact-001/photo", nil)
	w := httptest.NewRecorder()

	server.handleContactByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// TestDemoContacts_Coverage tests demo contacts data.
func TestDemoContacts_Coverage(t *testing.T) {
	t.Parallel()

	contacts := demoContacts()

	if len(contacts) == 0 {
		t.Error("expected non-empty demo contacts")
	}

	hasEmails := false
	hasPhones := false
	hasCompany := false
	hasAddress := false

	for _, c := range contacts {
		if c.ID == "" {
			t.Error("expected contact to have ID")
		}
		if c.DisplayName == "" {
			t.Error("expected contact to have DisplayName")
		}
		if len(c.Emails) > 0 {
			hasEmails = true
		}
		if len(c.PhoneNumbers) > 0 {
			hasPhones = true
		}
		if c.CompanyName != "" {
			hasCompany = true
		}
		if len(c.Addresses) > 0 {
			hasAddress = true
		}
	}

	if !hasEmails {
		t.Error("expected at least one contact with emails")
	}
	if !hasPhones {
		t.Error("expected at least one contact with phone numbers")
	}
	if !hasCompany {
		t.Error("expected at least one contact with company")
	}
	if !hasAddress {
		t.Error("expected at least one contact with address")
	}
}

// TestDemoContactGroups_Coverage tests demo contact groups data.
func TestDemoContactGroups_Coverage(t *testing.T) {
	t.Parallel()

	groups := demoContactGroups()

	if len(groups) == 0 {
		t.Error("expected non-empty demo contact groups")
	}

	for _, g := range groups {
		if g.ID == "" {
			t.Error("expected group to have ID")
		}
		if g.Name == "" {
			t.Error("expected group to have Name")
		}
		if g.Path == "" {
			t.Error("expected group to have Path")
		}
	}
}

// TestHandleContactSearch_ByName tests search by contact name.
func TestHandleContactSearch_ByName(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	// Get a name from demo contacts
	contacts := demoContacts()
	if len(contacts) == 0 {
		t.Skip("no demo contacts available")
	}

	searchName := contacts[0].GivenName
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q="+searchName, nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ContactsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should have at least one result
	if len(resp.Contacts) == 0 {
		t.Error("expected at least one contact in search results")
	}
}

// TestHandleContactSearch_ByEmail tests search by email.
func TestHandleContactSearch_ByEmail(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodGet, "/api/contacts/search?q=@example.com", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestHandleContactSearch_POST_NotAllowed tests search with wrong method.
func TestHandleContactSearch_POST_NotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestDemoServer()

	req := httptest.NewRequest(http.MethodPost, "/api/contacts/search?q=test", nil)
	w := httptest.NewRecorder()

	server.handleContactSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// TestHandleListContacts_ResponseStructure tests contacts list response.
func TestHandleListContacts_ResponseStructure(t *testing.T) {
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

	// Verify structure
	for _, contact := range resp.Contacts {
		if contact.ID == "" {
			t.Error("expected contact to have ID")
		}
		if contact.DisplayName == "" {
			t.Error("expected contact to have DisplayName")
		}
	}
}
