package contacts

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestDeduplicateContacts_EmptyList(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	req := &domain.DeduplicationRequest{
		Contacts: []domain.Contact{},
	}

	result, err := service.DeduplicateContacts(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OriginalCount != 0 {
		t.Errorf("expected OriginalCount=0, got %d", result.OriginalCount)
	}
	if result.DeduplicatedCount != 0 {
		t.Errorf("expected DeduplicatedCount=0, got %d", result.DeduplicatedCount)
	}
	if len(result.DuplicateGroups) != 0 {
		t.Errorf("expected no duplicate groups, got %d", len(result.DuplicateGroups))
	}
}

func TestDeduplicateContacts_NoDuplicates(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	contacts := []domain.Contact{
		{ID: "1", GivenName: "Alice", Surname: "Smith", Emails: []domain.ContactEmail{{Email: "alice@example.com"}}},
		{ID: "2", GivenName: "Bob", Surname: "Jones", Emails: []domain.ContactEmail{{Email: "bob@example.com"}}},
	}

	req := &domain.DeduplicationRequest{
		Contacts: contacts,
	}

	result, err := service.DeduplicateContacts(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OriginalCount != 2 {
		t.Errorf("expected OriginalCount=2, got %d", result.OriginalCount)
	}
}

func TestMergeContacts_EmptyList(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	_, err := service.MergeContacts(ctx, []domain.Contact{}, "prefer_first")
	if err == nil {
		t.Error("expected error for empty contact list")
	}
	if err != nil && err.Error() != "no contacts to merge" {
		t.Errorf("expected 'no contacts to merge' error, got: %v", err)
	}
}

func TestMergeContacts_SingleContact(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	contact := domain.Contact{
		ID:        "1",
		GivenName: "Alice",
		Surname:   "Smith",
		Emails:    []domain.ContactEmail{{Email: "alice@example.com"}},
	}

	result, err := service.MergeContacts(ctx, []domain.Contact{contact}, "prefer_first")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "1" {
		t.Errorf("expected ID=1, got %s", result.ID)
	}
	if result.GivenName != "Alice" {
		t.Errorf("expected GivenName=Alice, got %s", result.GivenName)
	}
}

func TestParseVCard_NotImplemented(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	_, err := service.ParseVCard(ctx, "BEGIN:VCARD\nEND:VCARD")
	if err == nil {
		t.Error("expected error for unimplemented vCard parsing")
	}
}

func TestExportVCard_NotImplemented(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	contact := domain.Contact{
		ID:        "1",
		GivenName: "Alice",
		Surname:   "Smith",
	}

	_, err := service.ExportVCard(ctx, []domain.Contact{contact})
	if err == nil {
		t.Error("expected error for unimplemented vCard export")
	}
}

func TestMapVCardFields_NotImplemented(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	contact := domain.Contact{
		ID:        "1",
		GivenName: "Alice",
	}

	_, err := service.MapVCardFields(ctx, "outlook", "google", &contact)
	if err == nil {
		t.Error("expected error for unimplemented field mapping")
	}
}

func TestHasCommonEmail(t *testing.T) {
	service := NewService()

	tests := []struct {
		name string
		c1   domain.Contact
		c2   domain.Contact
		want bool
	}{
		{
			name: "same email",
			c1:   domain.Contact{Emails: []domain.ContactEmail{{Email: "test@example.com"}}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{{Email: "test@example.com"}}},
			want: true,
		},
		{
			name: "case insensitive match",
			c1:   domain.Contact{Emails: []domain.ContactEmail{{Email: "Test@Example.com"}}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{{Email: "test@example.com"}}},
			want: true,
		},
		{
			name: "different emails",
			c1:   domain.Contact{Emails: []domain.ContactEmail{{Email: "alice@example.com"}}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{{Email: "bob@example.com"}}},
			want: false,
		},
		{
			name: "multiple emails with one match",
			c1:   domain.Contact{Emails: []domain.ContactEmail{{Email: "alice@example.com"}, {Email: "shared@example.com"}}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{{Email: "bob@example.com"}, {Email: "shared@example.com"}}},
			want: true,
		},
		{
			name: "no emails",
			c1:   domain.Contact{Emails: []domain.ContactEmail{}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{}},
			want: false,
		},
		{
			name: "one has no emails",
			c1:   domain.Contact{Emails: []domain.ContactEmail{{Email: "test@example.com"}}},
			c2:   domain.Contact{Emails: []domain.ContactEmail{}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasCommonEmail(tt.c1, tt.c2)
			if got != tt.want {
				t.Errorf("hasCommonEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCommonPhone(t *testing.T) {
	service := NewService()

	tests := []struct {
		name string
		c1   domain.Contact
		c2   domain.Contact
		want bool
	}{
		{
			name: "same phone",
			c1:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}}},
			c2:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}}},
			want: true,
		},
		{
			name: "different phones",
			c1:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}}},
			c2:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+0987654321"}}},
			want: false,
		},
		{
			name: "multiple phones with one match",
			c1:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+1111111111"}, {Number: "+1234567890"}}},
			c2:   domain.Contact{PhoneNumbers: []domain.ContactPhone{{Number: "+2222222222"}, {Number: "+1234567890"}}},
			want: true,
		},
		{
			name: "no phones",
			c1:   domain.Contact{PhoneNumbers: []domain.ContactPhone{}},
			c2:   domain.Contact{PhoneNumbers: []domain.ContactPhone{}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasCommonPhone(tt.c1, tt.c2)
			if got != tt.want {
				t.Errorf("hasCommonPhone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSimilarName(t *testing.T) {
	service := NewService()

	tests := []struct {
		name string
		c1   domain.Contact
		c2   domain.Contact
		want bool
	}{
		{
			name: "exact match",
			c1:   domain.Contact{GivenName: "John", Surname: "Doe"},
			c2:   domain.Contact{GivenName: "John", Surname: "Doe"},
			want: true,
		},
		{
			name: "case insensitive match",
			c1:   domain.Contact{GivenName: "john", Surname: "doe"},
			c2:   domain.Contact{GivenName: "John", Surname: "Doe"},
			want: true,
		},
		{
			name: "different names",
			c1:   domain.Contact{GivenName: "John", Surname: "Doe"},
			c2:   domain.Contact{GivenName: "Jane", Surname: "Smith"},
			want: false,
		},
		{
			name: "same given name, different surname",
			c1:   domain.Contact{GivenName: "John", Surname: "Doe"},
			c2:   domain.Contact{GivenName: "John", Surname: "Smith"},
			want: false,
		},
		{
			name: "empty names",
			c1:   domain.Contact{GivenName: "", Surname: ""},
			c2:   domain.Contact{GivenName: "", Surname: ""},
			want: true, // Empty names match (both produce " " after concatenation)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.hasSimilarName(tt.c1, tt.c2)
			if got != tt.want {
				t.Errorf("hasSimilarName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateCompleteness(t *testing.T) {
	service := NewService()

	tests := []struct {
		name    string
		contact domain.Contact
		want    int
	}{
		{
			name:    "empty contact",
			contact: domain.Contact{},
			want:    0,
		},
		{
			name:    "only name",
			contact: domain.Contact{GivenName: "John"},
			want:    1,
		},
		{
			name:    "name and surname",
			contact: domain.Contact{GivenName: "John", Surname: "Doe"},
			want:    2,
		},
		{
			name:    "full name and email",
			contact: domain.Contact{GivenName: "John", Surname: "Doe", Emails: []domain.ContactEmail{{Email: "john@example.com"}}},
			want:    3,
		},
		{
			name:    "full name, email, and phone",
			contact: domain.Contact{GivenName: "John", Surname: "Doe", Emails: []domain.ContactEmail{{Email: "john@example.com"}}, PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}}},
			want:    4,
		},
		{
			name: "all fields",
			contact: domain.Contact{
				GivenName:    "John",
				Surname:      "Doe",
				Emails:       []domain.ContactEmail{{Email: "john@example.com"}},
				PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}},
				JobTitle:     "Engineer",
				CompanyName:  "Acme Corp",
			},
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.calculateCompleteness(tt.contact)
			if got != tt.want {
				t.Errorf("calculateCompleteness() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeContacts_MultipleContacts(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	contacts := []domain.Contact{
		{
			ID:        "1",
			GivenName: "John",
			Surname:   "Doe",
			Emails:    []domain.ContactEmail{{Email: "john@example.com"}},
		},
		{
			ID:           "2",
			GivenName:    "John",
			Surname:      "Doe",
			PhoneNumbers: []domain.ContactPhone{{Number: "+1234567890"}},
		},
		{
			ID:          "3",
			GivenName:   "John",
			Surname:     "Doe",
			CompanyName: "Acme Corp",
		},
	}

	result, err := service.MergeContacts(ctx, contacts, "most_complete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should select most complete contact or merge fields
	if result.GivenName != "John" || result.Surname != "Doe" {
		t.Errorf("unexpected name: %s %s", result.GivenName, result.Surname)
	}

	// At least one of these should be present after merge
	hasData := len(result.Emails) > 0 || len(result.PhoneNumbers) > 0 || result.CompanyName != ""
	if !hasData {
		t.Error("merged contact should have some data from source contacts")
	}
}
