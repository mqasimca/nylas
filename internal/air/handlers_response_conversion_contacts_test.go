//go:build !integration
// +build !integration

package air

import (
	"testing"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// TestContactToResponse tests contact domain to response conversion.
func TestContactToResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		contact   domain.Contact
		checkFunc func(resp ContactResponse) bool
		desc      string
	}{
		{
			name: "basic contact",
			contact: domain.Contact{
				ID:          "contact-001",
				GivenName:   "John",
				Surname:     "Doe",
				Nickname:    "Johnny",
				CompanyName: "Acme Corp",
				JobTitle:    "Engineer",
				Birthday:    "1990-01-15",
				Notes:       "Test notes",
				PictureURL:  "https://example.com/photo.jpg",
				Source:      "google",
			},
			checkFunc: func(resp ContactResponse) bool {
				return resp.ID == "contact-001" &&
					resp.GivenName == "John" &&
					resp.Surname == "Doe" &&
					resp.Nickname == "Johnny" &&
					resp.CompanyName == "Acme Corp" &&
					resp.JobTitle == "Engineer" &&
					resp.Birthday == "1990-01-15" &&
					resp.Notes == "Test notes" &&
					resp.PictureURL == "https://example.com/photo.jpg" &&
					resp.Source == "google"
			},
			desc: "all basic fields should be converted",
		},
		{
			name: "contact with emails",
			contact: domain.Contact{
				ID:        "contact-002",
				GivenName: "Jane",
				Emails: []domain.ContactEmail{
					{Email: "jane@work.com", Type: "work"},
					{Email: "jane@home.com", Type: "home"},
				},
			},
			checkFunc: func(resp ContactResponse) bool {
				return len(resp.Emails) == 2 &&
					resp.Emails[0].Email == "jane@work.com" &&
					resp.Emails[0].Type == "work" &&
					resp.Emails[1].Email == "jane@home.com"
			},
			desc: "emails should be converted",
		},
		{
			name: "contact with phone numbers",
			contact: domain.Contact{
				ID:        "contact-003",
				GivenName: "Bob",
				PhoneNumbers: []domain.ContactPhone{
					{Number: "+1-555-1234", Type: "mobile"},
					{Number: "+1-555-5678", Type: "work"},
				},
			},
			checkFunc: func(resp ContactResponse) bool {
				return len(resp.PhoneNumbers) == 2 &&
					resp.PhoneNumbers[0].Number == "+1-555-1234" &&
					resp.PhoneNumbers[0].Type == "mobile"
			},
			desc: "phone numbers should be converted",
		},
		{
			name: "contact with addresses",
			contact: domain.Contact{
				ID:        "contact-004",
				GivenName: "Alice",
				PhysicalAddresses: []domain.ContactAddress{
					{
						Type:          "home",
						StreetAddress: "123 Main St",
						City:          "San Francisco",
						State:         "CA",
						PostalCode:    "94105",
						Country:       "USA",
					},
				},
			},
			checkFunc: func(resp ContactResponse) bool {
				return len(resp.Addresses) == 1 &&
					resp.Addresses[0].Type == "home" &&
					resp.Addresses[0].StreetAddress == "123 Main St" &&
					resp.Addresses[0].City == "San Francisco" &&
					resp.Addresses[0].State == "CA"
			},
			desc: "addresses should be converted",
		},
		{
			name: "contact with all details",
			contact: domain.Contact{
				ID:          "contact-005",
				GivenName:   "Complete",
				Surname:     "Contact",
				CompanyName: "Full Corp",
				Emails: []domain.ContactEmail{
					{Email: "complete@example.com", Type: "work"},
				},
				PhoneNumbers: []domain.ContactPhone{
					{Number: "+1-555-0000", Type: "mobile"},
				},
				PhysicalAddresses: []domain.ContactAddress{
					{Type: "work", City: "NYC"},
				},
			},
			checkFunc: func(resp ContactResponse) bool {
				return resp.GivenName == "Complete" &&
					resp.Surname == "Contact" &&
					len(resp.Emails) == 1 &&
					len(resp.PhoneNumbers) == 1 &&
					len(resp.Addresses) == 1
			},
			desc: "all nested fields should be converted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := contactToResponse(tt.contact)

			if !tt.checkFunc(resp) {
				t.Errorf("contactToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}

// TestCachedContactToResponse_Extended tests cached contact conversion.
func TestCachedContactToResponse_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cached    *cache.CachedContact
		checkFunc func(resp ContactResponse) bool
		desc      string
	}{
		{
			name: "basic cached contact",
			cached: &cache.CachedContact{
				ID:          "cached-contact-001",
				GivenName:   "Cached",
				Surname:     "User",
				DisplayName: "Cached User",
				Email:       "cached@example.com",
				Phone:       "+1-555-1234",
				Company:     "Cache Corp",
				JobTitle:    "Developer",
				Notes:       "Cached notes",
			},
			checkFunc: func(resp ContactResponse) bool {
				return resp.ID == "cached-contact-001" &&
					resp.DisplayName == "Cached User" &&
					resp.CompanyName == "Cache Corp" &&
					len(resp.Emails) == 1 &&
					resp.Emails[0].Email == "cached@example.com" &&
					len(resp.PhoneNumbers) == 1 &&
					resp.PhoneNumbers[0].Number == "+1-555-1234"
			},
			desc: "all fields should be converted correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := cachedContactToResponse(tt.cached)

			if !tt.checkFunc(resp) {
				t.Errorf("cachedContactToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}

// TestCalendarToResponse tests calendar domain to response conversion.
func TestCalendarToResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		calendar  domain.Calendar
		checkFunc func(resp CalendarResponse) bool
		desc      string
	}{
		{
			name: "primary calendar",
			calendar: domain.Calendar{
				ID:          "cal-primary",
				Name:        "Personal",
				Description: "My personal calendar",
				Timezone:    "America/New_York",
				IsPrimary:   true,
				ReadOnly:    false,
				HexColor:    "#4285f4",
			},
			checkFunc: func(resp CalendarResponse) bool {
				return resp.ID == "cal-primary" &&
					resp.Name == "Personal" &&
					resp.Description == "My personal calendar" &&
					resp.Timezone == "America/New_York" &&
					resp.IsPrimary == true &&
					resp.ReadOnly == false &&
					resp.HexColor == "#4285f4"
			},
			desc: "all fields should be converted",
		},
		{
			name: "read-only calendar",
			calendar: domain.Calendar{
				ID:        "cal-holidays",
				Name:      "Holidays",
				ReadOnly:  true,
				IsPrimary: false,
			},
			checkFunc: func(resp CalendarResponse) bool {
				return resp.ReadOnly == true && resp.IsPrimary == false
			},
			desc: "read-only flag should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := calendarToResponse(tt.calendar)

			if !tt.checkFunc(resp) {
				t.Errorf("calendarToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}
