// Package models provides screen models for the TUI.
package models

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewContactsScreen(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	c := NewContactsScreen(global)

	if c == nil {
		t.Fatal("NewContactsScreen returned nil")
	}

	if c.global != global {
		t.Error("global state not set correctly")
	}

	if c.theme == nil {
		t.Error("theme not initialized")
	}

	if !c.loading {
		t.Error("should be loading initially")
	}
}

func TestContactItem(t *testing.T) {
	tests := []struct {
		name        string
		contact     domain.Contact
		wantTitle   string
		wantDescLen int
	}{
		{
			name: "contact with name",
			contact: domain.Contact{
				GivenName: "John",
				Surname:   "Doe",
				Emails:    []domain.ContactEmail{{Email: "john@example.com"}},
			},
			wantTitle:   "John Doe",
			wantDescLen: 16, // email length
		},
		{
			name: "contact with email only",
			contact: domain.Contact{
				Emails: []domain.ContactEmail{{Email: "test@example.com"}},
			},
			wantTitle:   "test@example.com",
			wantDescLen: 16, // email length
		},
		{
			name: "contact with company",
			contact: domain.Contact{
				GivenName:   "Jane",
				Surname:     "Smith",
				Emails:      []domain.ContactEmail{{Email: "jane@company.com"}},
				CompanyName: "Acme Corp",
			},
			wantTitle:   "Jane Smith",
			wantDescLen: 30, // email + " • " + company
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := contactItem{contact: tt.contact}

			title := item.Title()
			if title != tt.wantTitle {
				t.Errorf("Title() = %q, want %q", title, tt.wantTitle)
			}

			desc := item.Description()
			if len(desc) != tt.wantDescLen {
				t.Errorf("Description() length = %d, want %d", len(desc), tt.wantDescLen)
			}

			filterValue := item.FilterValue()
			if filterValue == "" {
				t.Error("FilterValue() should not be empty")
			}
		})
	}
}

func TestContactItemEdgeCases(t *testing.T) {
	t.Run("no name or email", func(t *testing.T) {
		item := contactItem{contact: domain.Contact{}}
		title := item.Title()
		if title != "Unknown Contact" {
			t.Errorf("Title() = %q, want %q", title, "Unknown Contact")
		}
	})

	t.Run("with phone number", func(t *testing.T) {
		item := contactItem{
			contact: domain.Contact{
				GivenName: "Test",
				Emails:    []domain.ContactEmail{{Email: "test@example.com"}},
				PhoneNumbers: []domain.ContactPhone{
					{Number: "+1234567890"},
				},
			},
		}
		desc := item.Description()
		if desc == "" {
			t.Error("Description() should not be empty with phone number")
		}
	})
}
