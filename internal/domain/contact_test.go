package domain

import (
	"testing"
)

// =============================================================================
// Contact Tests
// =============================================================================

func TestContact_DisplayName(t *testing.T) {
	tests := []struct {
		name    string
		contact Contact
		want    string
	}{
		{
			name: "returns full name when both given and surname present",
			contact: Contact{
				GivenName: "John",
				Surname:   "Doe",
			},
			want: "John Doe",
		},
		{
			name: "returns given name only when surname empty",
			contact: Contact{
				GivenName: "John",
			},
			want: "John",
		},
		{
			name: "returns surname only when given name empty",
			contact: Contact{
				Surname: "Doe",
			},
			want: "Doe",
		},
		{
			name: "returns nickname when no name available",
			contact: Contact{
				Nickname: "Johnny",
			},
			want: "Johnny",
		},
		{
			name: "returns first email when no name or nickname",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "john@example.com", Type: "work"},
				},
			},
			want: "john@example.com",
		},
		{
			name: "returns Unknown when no identifiers available",
			contact: Contact{
				ID: "contact-123",
			},
			want: "Unknown",
		},
		{
			name: "prefers full name over nickname",
			contact: Contact{
				GivenName: "John",
				Surname:   "Doe",
				Nickname:  "Johnny",
			},
			want: "John Doe",
		},
		{
			name: "prefers given name over nickname",
			contact: Contact{
				GivenName: "John",
				Nickname:  "Johnny",
			},
			want: "John",
		},
		{
			name: "prefers surname over nickname",
			contact: Contact{
				Surname:  "Doe",
				Nickname: "Johnny",
			},
			want: "Doe",
		},
		{
			name: "prefers nickname over email",
			contact: Contact{
				Nickname: "Johnny",
				Emails: []ContactEmail{
					{Email: "john@example.com"},
				},
			},
			want: "Johnny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.contact.DisplayName()
			if got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContact_PrimaryEmail(t *testing.T) {
	tests := []struct {
		name    string
		contact Contact
		want    string
	}{
		{
			name: "returns primary type email first",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "other@example.com", Type: "other"},
					{Email: "primary@example.com", Type: "primary"},
				},
			},
			want: "primary@example.com",
		},
		{
			name: "returns work email when no primary",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "other@example.com", Type: "other"},
					{Email: "work@example.com", Type: "work"},
				},
			},
			want: "work@example.com",
		},
		{
			name: "returns home email when no primary or work",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "other@example.com", Type: "other"},
					{Email: "home@example.com", Type: "home"},
				},
			},
			want: "home@example.com",
		},
		{
			name: "returns first email when no preferred type",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "first@example.com", Type: "other"},
					{Email: "second@example.com", Type: "school"},
				},
			},
			want: "first@example.com",
		},
		{
			name: "returns empty string when no emails",
			contact: Contact{
				ID: "contact-123",
			},
			want: "",
		},
		{
			name: "returns first matching preferred type in slice order",
			contact: Contact{
				Emails: []ContactEmail{
					{Email: "work@example.com", Type: "work"},
					{Email: "primary@example.com", Type: "primary"},
				},
			},
			want: "work@example.com", // Returns first match in slice order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.contact.PrimaryEmail()
			if got != tt.want {
				t.Errorf("PrimaryEmail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContact_PrimaryPhone(t *testing.T) {
	tests := []struct {
		name    string
		contact Contact
		want    string
	}{
		{
			name: "returns mobile phone first",
			contact: Contact{
				PhoneNumbers: []ContactPhone{
					{Number: "+1-555-111-1111", Type: "other"},
					{Number: "+1-555-222-2222", Type: "mobile"},
				},
			},
			want: "+1-555-222-2222",
		},
		{
			name: "returns work phone when no mobile",
			contact: Contact{
				PhoneNumbers: []ContactPhone{
					{Number: "+1-555-111-1111", Type: "other"},
					{Number: "+1-555-333-3333", Type: "work"},
				},
			},
			want: "+1-555-333-3333",
		},
		{
			name: "returns home phone when no mobile or work",
			contact: Contact{
				PhoneNumbers: []ContactPhone{
					{Number: "+1-555-111-1111", Type: "pager"},
					{Number: "+1-555-444-4444", Type: "home"},
				},
			},
			want: "+1-555-444-4444",
		},
		{
			name: "returns first phone when no preferred type",
			contact: Contact{
				PhoneNumbers: []ContactPhone{
					{Number: "+1-555-111-1111", Type: "pager"},
					{Number: "+1-555-222-2222", Type: "business_fax"},
				},
			},
			want: "+1-555-111-1111",
		},
		{
			name: "returns empty string when no phones",
			contact: Contact{
				ID: "contact-123",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.contact.PrimaryPhone()
			if got != tt.want {
				t.Errorf("PrimaryPhone() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// ContactEmail Tests
// =============================================================================

func TestContactEmail_Creation(t *testing.T) {
	tests := []struct {
		name  string
		email ContactEmail
	}{
		{
			name: "creates work email",
			email: ContactEmail{
				Email: "work@example.com",
				Type:  "work",
			},
		},
		{
			name: "creates home email",
			email: ContactEmail{
				Email: "home@example.com",
				Type:  "home",
			},
		},
		{
			name: "creates email without type",
			email: ContactEmail{
				Email: "noType@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.email.Email == "" {
				t.Error("ContactEmail.Email should not be empty")
			}
		})
	}
}

// =============================================================================
// ContactPhone Tests
// =============================================================================

func TestContactPhone_Creation(t *testing.T) {
	tests := []struct {
		name  string
		phone ContactPhone
	}{
		{
			name: "creates mobile phone",
			phone: ContactPhone{
				Number: "+1-555-123-4567",
				Type:   "mobile",
			},
		},
		{
			name: "creates work phone",
			phone: ContactPhone{
				Number: "+1-555-987-6543",
				Type:   "work",
			},
		},
		{
			name: "creates phone without type",
			phone: ContactPhone{
				Number: "+1-555-000-0000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.phone.Number == "" {
				t.Error("ContactPhone.Number should not be empty")
			}
		})
	}
}

// =============================================================================
// ContactAddress Tests
// =============================================================================

func TestContactAddress_Creation(t *testing.T) {
	addr := ContactAddress{
		Type:          "work",
		StreetAddress: "123 Main St",
		City:          "San Francisco",
		State:         "CA",
		PostalCode:    "94105",
		Country:       "USA",
	}

	if addr.Type != "work" {
		t.Errorf("ContactAddress.Type = %q, want %q", addr.Type, "work")
	}
	if addr.City != "San Francisco" {
		t.Errorf("ContactAddress.City = %q, want %q", addr.City, "San Francisco")
	}
	if addr.State != "CA" {
		t.Errorf("ContactAddress.State = %q, want %q", addr.State, "CA")
	}
	if addr.Country != "USA" {
		t.Errorf("ContactAddress.Country = %q, want %q", addr.Country, "USA")
	}
}

// =============================================================================
// ContactGroup Tests
// =============================================================================

func TestContactGroup_Creation(t *testing.T) {
	group := ContactGroup{
		ID:      "group-123",
		GrantID: "grant-456",
		Name:    "Work Contacts",
		Path:    "Work/Colleagues",
	}

	if group.ID != "group-123" {
		t.Errorf("ContactGroup.ID = %q, want %q", group.ID, "group-123")
	}
	if group.Name != "Work Contacts" {
		t.Errorf("ContactGroup.Name = %q, want %q", group.Name, "Work Contacts")
	}
	if group.Path != "Work/Colleagues" {
		t.Errorf("ContactGroup.Path = %q, want %q", group.Path, "Work/Colleagues")
	}
}

// =============================================================================
// ContactQueryParams Tests
// =============================================================================

func TestContactQueryParams_Creation(t *testing.T) {
	params := ContactQueryParams{
		Limit:          50,
		PageToken:      "token-123",
		Email:          "john@example.com",
		PhoneNumber:    "+1-555-123-4567",
		Source:         "address_book",
		Group:          "group-123",
		Recurse:        true,
		ProfilePicture: true,
	}

	if params.Limit != 50 {
		t.Errorf("ContactQueryParams.Limit = %d, want 50", params.Limit)
	}
	if params.Source != "address_book" {
		t.Errorf("ContactQueryParams.Source = %q, want %q", params.Source, "address_book")
	}
	if !params.Recurse {
		t.Error("ContactQueryParams.Recurse should be true")
	}
	if !params.ProfilePicture {
		t.Error("ContactQueryParams.ProfilePicture should be true")
	}
}

// =============================================================================
// CreateContactRequest Tests
// =============================================================================

func TestCreateContactRequest_Creation(t *testing.T) {
	req := CreateContactRequest{
		GivenName:   "John",
		MiddleName:  "William",
		Surname:     "Doe",
		Suffix:      "Jr.",
		Nickname:    "Johnny",
		Birthday:    "1990-01-15",
		CompanyName: "Acme Corp",
		JobTitle:    "Engineer",
		ManagerName: "Jane Manager",
		Notes:       "Met at conference",
		Emails: []ContactEmail{
			{Email: "john@example.com", Type: "work"},
		},
		PhoneNumbers: []ContactPhone{
			{Number: "+1-555-123-4567", Type: "mobile"},
		},
		WebPages: []ContactWebPage{
			{URL: "https://johndoe.com", Type: "profile"},
		},
		IMAddresses: []ContactIM{
			{IMAddress: "johndoe", Type: "skype"},
		},
		PhysicalAddresses: []ContactAddress{
			{Type: "work", City: "San Francisco"},
		},
		Groups: []ContactGroupInfo{
			{ID: "group-123"},
		},
	}

	if req.GivenName != "John" {
		t.Errorf("CreateContactRequest.GivenName = %q, want %q", req.GivenName, "John")
	}
	if req.Surname != "Doe" {
		t.Errorf("CreateContactRequest.Surname = %q, want %q", req.Surname, "Doe")
	}
	if len(req.Emails) != 1 {
		t.Errorf("CreateContactRequest.Emails length = %d, want 1", len(req.Emails))
	}
	if len(req.PhoneNumbers) != 1 {
		t.Errorf("CreateContactRequest.PhoneNumbers length = %d, want 1", len(req.PhoneNumbers))
	}
}

// =============================================================================
// UpdateContactRequest Tests
// =============================================================================

func TestUpdateContactRequest_Creation(t *testing.T) {
	givenName := "John"
	surname := "Smith"

	req := UpdateContactRequest{
		GivenName: &givenName,
		Surname:   &surname,
		Emails: []ContactEmail{
			{Email: "john.smith@example.com", Type: "work"},
		},
	}

	if req.GivenName == nil || *req.GivenName != "John" {
		t.Errorf("UpdateContactRequest.GivenName = %v, want %q", req.GivenName, "John")
	}
	if req.Surname == nil || *req.Surname != "Smith" {
		t.Errorf("UpdateContactRequest.Surname = %v, want %q", req.Surname, "Smith")
	}
	if len(req.Emails) != 1 {
		t.Errorf("UpdateContactRequest.Emails length = %d, want 1", len(req.Emails))
	}
}

// =============================================================================
// ContactWebPage Tests
// =============================================================================

func TestContactWebPage_Creation(t *testing.T) {
	tests := []struct {
		name    string
		webPage ContactWebPage
	}{
		{
			name: "creates profile webpage",
			webPage: ContactWebPage{
				URL:  "https://linkedin.com/in/johndoe",
				Type: "profile",
			},
		},
		{
			name: "creates work webpage",
			webPage: ContactWebPage{
				URL:  "https://acme.com",
				Type: "work",
			},
		},
		{
			name: "creates blog webpage",
			webPage: ContactWebPage{
				URL:  "https://johndoe.blog",
				Type: "blog",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.webPage.URL == "" {
				t.Error("ContactWebPage.URL should not be empty")
			}
		})
	}
}

// =============================================================================
// ContactIM Tests
// =============================================================================

func TestContactIM_Creation(t *testing.T) {
	tests := []struct {
		name string
		im   ContactIM
	}{
		{
			name: "creates skype IM",
			im: ContactIM{
				IMAddress: "johndoe.skype",
				Type:      "skype",
			},
		},
		{
			name: "creates slack IM",
			im: ContactIM{
				IMAddress: "@johndoe",
				Type:      "other",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.im.IMAddress == "" {
				t.Error("ContactIM.IMAddress should not be empty")
			}
		})
	}
}

// =============================================================================
// ContactGroupInfo Tests
// =============================================================================

func TestContactGroupInfo_Creation(t *testing.T) {
	info := ContactGroupInfo{
		ID: "group-456",
	}

	if info.ID != "group-456" {
		t.Errorf("ContactGroupInfo.ID = %q, want %q", info.ID, "group-456")
	}
}

// =============================================================================
// ContactListResponse Tests
// =============================================================================

func TestContactListResponse_Creation(t *testing.T) {
	resp := ContactListResponse{
		Data: []Contact{
			{ID: "contact-1", GivenName: "John"},
			{ID: "contact-2", GivenName: "Jane"},
		},
		Pagination: Pagination{
			NextCursor: "cursor-123",
			HasMore:    true,
		},
	}

	if len(resp.Data) != 2 {
		t.Errorf("ContactListResponse.Data length = %d, want 2", len(resp.Data))
	}
	if !resp.Pagination.HasMore {
		t.Error("ContactListResponse.Pagination.HasMore should be true")
	}
	if resp.Pagination.NextCursor != "cursor-123" {
		t.Errorf("Pagination.NextCursor = %q, want %q", resp.Pagination.NextCursor, "cursor-123")
	}
}

// =============================================================================
// CreateContactGroupRequest Tests
// =============================================================================

func TestCreateContactGroupRequest_Creation(t *testing.T) {
	req := CreateContactGroupRequest{
		Name: "VIP Customers",
	}

	if req.Name != "VIP Customers" {
		t.Errorf("CreateContactGroupRequest.Name = %q, want %q", req.Name, "VIP Customers")
	}
}

// =============================================================================
// UpdateContactGroupRequest Tests
// =============================================================================

func TestUpdateContactGroupRequest_Creation(t *testing.T) {
	name := "Updated Group Name"
	req := UpdateContactGroupRequest{
		Name: &name,
	}

	if req.Name == nil || *req.Name != "Updated Group Name" {
		t.Errorf("UpdateContactGroupRequest.Name = %v, want %q", req.Name, "Updated Group Name")
	}
}
