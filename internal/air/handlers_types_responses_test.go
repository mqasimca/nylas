package air

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

func TestDraftToResponse_Basic(t *testing.T) {
	t.Parallel()

	draft := domain.Draft{
		ID:      "draft-123",
		Subject: "Draft Subject",
		Body:    "<p>Draft body</p>",
	}

	resp := draftToResponse(draft)

	if resp.ID != "draft-123" {
		t.Errorf("expected ID 'draft-123', got %s", resp.ID)
	}
	if resp.Subject != "Draft Subject" {
		t.Errorf("expected Subject 'Draft Subject', got %s", resp.Subject)
	}
	if resp.Body != "<p>Draft body</p>" {
		t.Errorf("expected Body to match, got %s", resp.Body)
	}
}

func TestDraftToResponse_WithRecipients(t *testing.T) {
	t.Parallel()

	draft := domain.Draft{
		ID: "draft-123",
		To: []domain.EmailParticipant{
			{Name: "To Person", Email: "to@example.com"},
		},
		Cc: []domain.EmailParticipant{
			{Name: "CC Person", Email: "cc@example.com"},
		},
		Bcc: []domain.EmailParticipant{
			{Name: "BCC Person", Email: "bcc@example.com"},
		},
	}

	resp := draftToResponse(draft)

	if len(resp.To) != 1 || resp.To[0].Email != "to@example.com" {
		t.Error("To recipients not converted correctly")
	}
	if len(resp.Cc) != 1 || resp.Cc[0].Email != "cc@example.com" {
		t.Error("Cc recipients not converted correctly")
	}
	if len(resp.Bcc) != 1 || resp.Bcc[0].Email != "bcc@example.com" {
		t.Error("Bcc recipients not converted correctly")
	}
}

func TestCalendarToResponse_Basic(t *testing.T) {
	t.Parallel()

	cal := domain.Calendar{
		ID:          "cal-123",
		Name:        "Work Calendar",
		Description: "Work events",
		Timezone:    "America/New_York",
		IsPrimary:   true,
		ReadOnly:    false,
		HexColor:    "#4285f4",
	}

	resp := calendarToResponse(cal)

	if resp.ID != "cal-123" {
		t.Errorf("expected ID 'cal-123', got %s", resp.ID)
	}
	if resp.Name != "Work Calendar" {
		t.Errorf("expected Name 'Work Calendar', got %s", resp.Name)
	}
	if resp.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got %s", resp.Timezone)
	}
	if !resp.IsPrimary {
		t.Error("expected IsPrimary to be true")
	}
	if resp.ReadOnly {
		t.Error("expected ReadOnly to be false")
	}
	if resp.HexColor != "#4285f4" {
		t.Errorf("expected HexColor '#4285f4', got %s", resp.HexColor)
	}
}

func TestContactToResponse_Basic(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID:          "contact-123",
		GivenName:   "John",
		Surname:     "Doe",
		Nickname:    "Johnny",
		CompanyName: "Acme Corp",
		JobTitle:    "Engineer",
		Birthday:    "1990-01-15",
		Notes:       "Test notes",
		PictureURL:  "https://example.com/photo.jpg",
		Source:      "google",
	}

	resp := contactToResponse(contact)

	if resp.ID != "contact-123" {
		t.Errorf("expected ID 'contact-123', got %s", resp.ID)
	}
	if resp.GivenName != "John" {
		t.Errorf("expected GivenName 'John', got %s", resp.GivenName)
	}
	if resp.Surname != "Doe" {
		t.Errorf("expected Surname 'Doe', got %s", resp.Surname)
	}
	if resp.CompanyName != "Acme Corp" {
		t.Errorf("expected CompanyName 'Acme Corp', got %s", resp.CompanyName)
	}
}

func TestContactToResponse_WithEmails(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		Emails: []domain.ContactEmail{
			{Email: "john@work.com", Type: "work"},
			{Email: "john@home.com", Type: "home"},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.Emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(resp.Emails))
	}
	if resp.Emails[0].Email != "john@work.com" {
		t.Errorf("expected first email 'john@work.com', got %s", resp.Emails[0].Email)
	}
	if resp.Emails[0].Type != "work" {
		t.Errorf("expected first email type 'work', got %s", resp.Emails[0].Type)
	}
}

func TestContactToResponse_WithPhoneNumbers(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		PhoneNumbers: []domain.ContactPhone{
			{Number: "+1-555-123-4567", Type: "mobile"},
			{Number: "+1-555-987-6543", Type: "work"},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.PhoneNumbers) != 2 {
		t.Errorf("expected 2 phone numbers, got %d", len(resp.PhoneNumbers))
	}
	if resp.PhoneNumbers[0].Number != "+1-555-123-4567" {
		t.Errorf("expected first phone '+1-555-123-4567', got %s", resp.PhoneNumbers[0].Number)
	}
}

func TestContactToResponse_WithAddresses(t *testing.T) {
	t.Parallel()

	contact := domain.Contact{
		ID: "contact-123",
		PhysicalAddresses: []domain.ContactAddress{
			{
				Type:          "work",
				StreetAddress: "123 Main St",
				City:          "San Francisco",
				State:         "CA",
				PostalCode:    "94102",
				Country:       "USA",
			},
		},
	}

	resp := contactToResponse(contact)

	if len(resp.Addresses) != 1 {
		t.Errorf("expected 1 address, got %d", len(resp.Addresses))
	}
	if resp.Addresses[0].City != "San Francisco" {
		t.Errorf("expected City 'San Francisco', got %s", resp.Addresses[0].City)
	}
}

func TestCachedEventToResponse(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2024, 1, 20, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)

	cachedEvent := &cache.CachedEvent{
		ID:          "event-123",
		CalendarID:  "cal-456",
		Title:       "Team Meeting",
		Description: "Weekly sync",
		Location:    "Conference Room A",
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      false,
		Status:      "confirmed",
		Busy:        true,
	}

	resp := cachedEventToResponse(cachedEvent)

	if resp.ID != "event-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "event-123")
	}
	if resp.CalendarID != "cal-456" {
		t.Errorf("CalendarID = %q, want %q", resp.CalendarID, "cal-456")
	}
	if resp.Title != "Team Meeting" {
		t.Errorf("Title = %q, want %q", resp.Title, "Team Meeting")
	}
	if resp.Description != "Weekly sync" {
		t.Errorf("Description = %q, want %q", resp.Description, "Weekly sync")
	}
	if resp.Location != "Conference Room A" {
		t.Errorf("Location = %q, want %q", resp.Location, "Conference Room A")
	}
	if resp.StartTime != startTime.Unix() {
		t.Errorf("StartTime = %d, want %d", resp.StartTime, startTime.Unix())
	}
	if resp.EndTime != endTime.Unix() {
		t.Errorf("EndTime = %d, want %d", resp.EndTime, endTime.Unix())
	}
	if resp.IsAllDay {
		t.Error("IsAllDay should be false")
	}
	if resp.Status != "confirmed" {
		t.Errorf("Status = %q, want %q", resp.Status, "confirmed")
	}
	if !resp.Busy {
		t.Error("Busy should be true")
	}
}

func TestCachedEventToResponse_AllDayEvent(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)

	cachedEvent := &cache.CachedEvent{
		ID:        "event-allday",
		Title:     "Holiday",
		StartTime: startTime,
		EndTime:   endTime,
		AllDay:    true,
	}

	resp := cachedEventToResponse(cachedEvent)

	if resp.ID != "event-allday" {
		t.Errorf("ID = %q, want %q", resp.ID, "event-allday")
	}
	if !resp.IsAllDay {
		t.Error("IsAllDay should be true")
	}
}

func TestCachedContactToResponse(t *testing.T) {
	t.Parallel()

	cachedContact := &cache.CachedContact{
		ID:          "contact-123",
		GivenName:   "Jane",
		Surname:     "Smith",
		DisplayName: "Jane Smith",
		Email:       "jane@example.com",
		Phone:       "+1-555-1234",
		Company:     "Acme Corp",
		JobTitle:    "Engineer",
		Notes:       "Met at conference",
	}

	resp := cachedContactToResponse(cachedContact)

	if resp.ID != "contact-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "contact-123")
	}
	if resp.GivenName != "Jane" {
		t.Errorf("GivenName = %q, want %q", resp.GivenName, "Jane")
	}
	if resp.Surname != "Smith" {
		t.Errorf("Surname = %q, want %q", resp.Surname, "Smith")
	}
	if resp.DisplayName != "Jane Smith" {
		t.Errorf("DisplayName = %q, want %q", resp.DisplayName, "Jane Smith")
	}
	if len(resp.Emails) != 1 || resp.Emails[0].Email != "jane@example.com" || resp.Emails[0].Type != "personal" {
		t.Errorf("Emails = %+v, want [{jane@example.com personal}]", resp.Emails)
	}
	if len(resp.PhoneNumbers) != 1 || resp.PhoneNumbers[0].Number != "+1-555-1234" || resp.PhoneNumbers[0].Type != "mobile" {
		t.Errorf("PhoneNumbers = %+v, want [{+1-555-1234 mobile}]", resp.PhoneNumbers)
	}
	if resp.CompanyName != "Acme Corp" {
		t.Errorf("CompanyName = %q, want %q", resp.CompanyName, "Acme Corp")
	}
	if resp.JobTitle != "Engineer" {
		t.Errorf("JobTitle = %q, want %q", resp.JobTitle, "Engineer")
	}
	if resp.Notes != "Met at conference" {
		t.Errorf("Notes = %q, want %q", resp.Notes, "Met at conference")
	}
}

func TestCachedContactToResponse_MinimalData(t *testing.T) {
	t.Parallel()

	cachedContact := &cache.CachedContact{
		ID:        "contact-minimal",
		GivenName: "Bob",
	}

	resp := cachedContactToResponse(cachedContact)

	if resp.ID != "contact-minimal" {
		t.Errorf("ID = %q, want %q", resp.ID, "contact-minimal")
	}
	if resp.GivenName != "Bob" {
		t.Errorf("GivenName = %q, want %q", resp.GivenName, "Bob")
	}
	if resp.Surname != "" {
		t.Errorf("Surname should be empty, got %q", resp.Surname)
	}
	// Email and Phone should still have entries (even if empty)
	if len(resp.Emails) != 1 {
		t.Error("Emails should have one entry")
	}
	if len(resp.PhoneNumbers) != 1 {
		t.Error("PhoneNumbers should have one entry")
	}
}
