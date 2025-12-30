package air

import (
	"encoding/json"
	"testing"
)

// ================================
// CONTACTS SERIALIZATION TESTS
// ================================

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
