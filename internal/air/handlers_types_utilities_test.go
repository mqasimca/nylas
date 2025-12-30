package air

import (
	"testing"
)

func TestContainsEmail(t *testing.T) {
	t.Parallel()

	emails := []ContactEmailResponse{
		{Email: "test@example.com", Type: "work"},
		{Email: "john@demo.org", Type: "personal"},
	}

	tests := []struct {
		query    string
		expected bool
	}{
		{"test@example.com", true},
		{"example.com", true},
		{"demo.org", true},
		{"john", true},
		{"notfound", false},
		{"xyz@other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := containsEmail(emails, tt.query)
			if result != tt.expected {
				t.Errorf("containsEmail(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestMatchesContactQuery(t *testing.T) {
	t.Parallel()

	contact := ContactResponse{
		ID:          "test-1",
		GivenName:   "John",
		Surname:     "Doe",
		DisplayName: "John Doe",
		CompanyName: "Acme Corp",
		Notes:       "Important client",
		Emails: []ContactEmailResponse{
			{Email: "john@acme.com", Type: "work"},
		},
	}

	tests := []struct {
		query    string
		expected bool
	}{
		{"john", true},
		{"doe", true},
		{"John Doe", true},
		{"acme", true},
		{"important", true},
		{"client", true},
		{"notfound", false},
		{"xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := matchesContactQuery(contact, tt.query)
			if result != tt.expected {
				t.Errorf("matchesContactQuery(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

// ================================
// CONFLICT DETECTION TESTS
// ================================
