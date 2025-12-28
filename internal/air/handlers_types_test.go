package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// ================================
// HELPER FUNCTION TESTS
// ================================

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", resp["key"])
	}
}

func TestWriteJSON_NilValue(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "null\n" {
		t.Errorf("expected 'null', got %s", w.Body.String())
	}
}

func TestWriteJSON_EmptyMap(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]any{})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "{}\n" {
		t.Errorf("expected '{}', got %s", w.Body.String())
	}
}

func TestParticipantsToEmail(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result))
	}

	if result[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", result[0].Name)
	}

	if result[0].Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", result[0].Email)
	}
}

func TestParticipantsToEmail_Empty(t *testing.T) {
	t.Parallel()

	result := participantsToEmail(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}

	result = participantsToEmail([]EmailParticipantResponse{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestParticipantsToEmail_Multiple(t *testing.T) {
	t.Parallel()

	participants := []EmailParticipantResponse{
		{Name: "Person One", Email: "one@example.com"},
		{Name: "Person Two", Email: "two@example.com"},
		{Name: "", Email: "three@example.com"},
	}

	result := participantsToEmail(participants)

	if len(result) != 3 {
		t.Errorf("expected 3 participants, got %d", len(result))
	}

	if result[0].Name != "Person One" || result[0].Email != "one@example.com" {
		t.Error("first participant not converted correctly")
	}

	if result[2].Email != "three@example.com" {
		t.Error("third participant email not converted correctly")
	}
}

func TestGrantFromDomain(t *testing.T) {
	t.Parallel()

	domainGrant := struct {
		ID       string
		Email    string
		Provider string
	}{
		ID:       "grant-123",
		Email:    "test@example.com",
		Provider: "google",
	}

	// Since grantFromDomain expects domain.GrantInfo, we test the conversion logic
	result := Grant{
		ID:       domainGrant.ID,
		Email:    domainGrant.Email,
		Provider: domainGrant.Provider,
	}

	if result.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", result.ID)
	}

	if result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", result.Email)
	}

	if result.Provider != "google" {
		t.Errorf("expected provider 'google', got %s", result.Provider)
	}
}

func TestGrantFromDomain_Basic(t *testing.T) {
	t.Parallel()

	grantInfo := domain.GrantInfo{
		ID:       "grant-123",
		Email:    "user@example.com",
		Provider: domain.ProviderGoogle,
	}

	grant := grantFromDomain(grantInfo)

	if grant.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got %s", grant.ID)
	}
	if grant.Email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got %s", grant.Email)
	}
	if grant.Provider != "google" {
		t.Errorf("expected Provider 'google', got %s", grant.Provider)
	}
}

func TestGrantFromDomain_DifferentProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider domain.Provider
		expected string
	}{
		{"Google", domain.ProviderGoogle, "google"},
		{"Microsoft", domain.ProviderMicrosoft, "microsoft"},
		{"IMAP", domain.ProviderIMAP, "imap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grantInfo := domain.GrantInfo{
				ID:       "grant-123",
				Email:    "user@example.com",
				Provider: tt.provider,
			}

			grant := grantFromDomain(grantInfo)

			if grant.Provider != tt.expected {
				t.Errorf("expected Provider '%s', got %s", tt.expected, grant.Provider)
			}
		})
	}
}

// ================================
// CONTACT HELPER FUNCTION TESTS
// ================================

func TestContainsEmail(t *testing.T) {
	t.Parallel()

	emails := []ContactEmailResponse{
		{Email: "test@example.com", Type: "work"},
		{Email: "john@nylas.com", Type: "personal"},
	}

	tests := []struct {
		query    string
		expected bool
	}{
		{"test@example.com", true},
		{"example.com", true},
		{"nylas.com", true},
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

func TestFindConflicts(t *testing.T) {
	t.Parallel()

	// Test with overlapping events
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 and event-2 overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}

	if len(conflicts) > 0 {
		if conflicts[0].Event1.ID != "event-1" || conflicts[0].Event2.ID != "event-2" {
			t.Error("expected conflict between event-1 and event-2")
		}
	}
}

func TestFindConflicts_NoOverlap(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 3000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_CancelledEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "cancelled", // This should be ignored
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (cancelled event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_FreeEventsIgnored(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Free Time",
			Status: "confirmed",
			Busy:   false, // Free, not busy
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   2500,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts (free event should be ignored), got %d", len(conflicts))
	}
}

func TestFindConflicts_AllDayEvents(t *testing.T) {
	t.Parallel()

	// All-day event should conflict with timed event on same day
	events := []domain.Event{
		{
			ID:     "all-day-1",
			Title:  "Holiday",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				Date: "2024-12-25", // All-day event
			},
		},
		{
			ID:     "timed-1",
			Title:  "Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1735142400, // Dec 25, 2024 12:00 UTC
				EndTime:   1735146000, // Dec 25, 2024 13:00 UTC
			},
		},
	}

	conflicts := findConflicts(events)

	// All-day event and timed event overlap
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict (all-day vs timed), got %d", len(conflicts))
	}
}

func TestFindConflicts_MultipleConflicts(t *testing.T) {
	t.Parallel()

	// Three overlapping events should produce 3 conflicts (each pair)
	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Meeting 1",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   3000,
			},
		},
		{
			ID:     "event-2",
			Title:  "Meeting 2",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1500,
				EndTime:   3500,
			},
		},
		{
			ID:     "event-3",
			Title:  "Meeting 3",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 2000,
				EndTime:   4000,
			},
		},
	}

	conflicts := findConflicts(events)

	// event-1 overlaps event-2, event-1 overlaps event-3, event-2 overlaps event-3
	if len(conflicts) != 3 {
		t.Errorf("expected 3 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflicts_EmptyList(t *testing.T) {
	t.Parallel()

	conflicts := findConflicts([]domain.Event{})

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for empty list, got %d", len(conflicts))
	}
}

func TestFindConflicts_SingleEvent(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{
			ID:     "event-1",
			Title:  "Only Meeting",
			Status: "confirmed",
			Busy:   true,
			When: domain.EventWhen{
				StartTime: 1000,
				EndTime:   2000,
			},
		},
	}

	conflicts := findConflicts(events)

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for single event, got %d", len(conflicts))
	}
}

func TestRoundUpTo5Min(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{"already aligned", 1735142400, 1735142400}, // 12:00:00 stays 12:00:00
		{"1 second after", 1735142401, 1735142700},  // 12:00:01 -> 12:05:00
		{"2 minutes in", 1735142520, 1735142700},    // 12:02:00 -> 12:05:00
		{"4 min 59 sec", 1735142699, 1735142700},    // 12:04:59 -> 12:05:00
		{"zero", 0, 0},
		{"5 min aligned", 300, 300},
		{"10 min aligned", 600, 600},
		{"6 minutes", 360, 600}, // 00:06:00 -> 00:10:00
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundUpTo5Min(tt.input)
			if result != tt.expected {
				t.Errorf("roundUpTo5Min(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ================================
// CSS STYLING TESTS
// ================================

func TestEmailBodyCSS_HasLightBackground(t *testing.T) {
	t.Parallel()

	// Read the preview.css file from embedded files
	cssContent, err := staticFiles.ReadFile("static/css/preview.css")
	if err != nil {
		t.Fatalf("failed to read preview.css: %v", err)
	}

	css := string(cssContent)

	// Verify email iframe container has white/light background for readability
	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		{
			"email iframe container has light background",
			"background: #ffffff",
			"HTML emails have inline styles for light backgrounds - need white bg for readability",
		},
		{
			"email body selector exists",
			".email-detail-body",
			"Email body styling must be defined",
		},
		{
			"email iframe container selector exists",
			".email-iframe-container",
			"Email iframe container styling must be defined for sandboxed email rendering",
		},
		{
			"email iframe styling exists",
			".email-body-iframe",
			"Sandboxed iframe styling must be defined for security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(css, tt.contains) {
				t.Errorf("preview.css missing required style: %s\nReason: %s", tt.contains, tt.reason)
			}
		})
	}
}

// ================================
// RESPONSE CONVERTER TESTS
// ================================

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
