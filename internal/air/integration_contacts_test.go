//go:build integration
// +build integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
