//go:build !integration
// +build !integration

package air

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// =============================================================================
// Response Conversion Tests
// =============================================================================

// TestEmailToResponse_Extended tests email domain to response conversion.
func TestEmailToResponse_Extended(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name        string
		message     domain.Message
		includeBody bool
		checkFunc   func(resp EmailResponse) bool
		desc        string
	}{
		{
			name: "basic message",
			message: domain.Message{
				ID:       "msg-001",
				ThreadID: "thread-001",
				Subject:  "Test Subject",
				Snippet:  "Test snippet...",
				Date:     now,
				Unread:   true,
				Starred:  false,
				Folders:  []string{"inbox"},
			},
			includeBody: false,
			checkFunc: func(resp EmailResponse) bool {
				return resp.ID == "msg-001" &&
					resp.ThreadID == "thread-001" &&
					resp.Subject == "Test Subject" &&
					resp.Snippet == "Test snippet..." &&
					resp.Unread == true &&
					resp.Starred == false &&
					resp.Body == ""
			},
			desc: "basic fields should match",
		},
		{
			name: "with body",
			message: domain.Message{
				ID:      "msg-002",
				Subject: "Subject",
				Body:    "<p>Email body content</p>",
			},
			includeBody: true,
			checkFunc: func(resp EmailResponse) bool {
				return resp.Body == "<p>Email body content</p>"
			},
			desc: "body should be included when requested",
		},
		{
			name: "body excluded",
			message: domain.Message{
				ID:      "msg-003",
				Subject: "Subject",
				Body:    "<p>Email body content</p>",
			},
			includeBody: false,
			checkFunc: func(resp EmailResponse) bool {
				return resp.Body == ""
			},
			desc: "body should be empty when not requested",
		},
		{
			name: "with participants",
			message: domain.Message{
				ID:      "msg-004",
				Subject: "Subject",
				From: []domain.EmailParticipant{
					{Name: "Sender", Email: "sender@example.com"},
				},
				To: []domain.EmailParticipant{
					{Name: "Recipient", Email: "recipient@example.com"},
				},
				Cc: []domain.EmailParticipant{
					{Name: "CC User", Email: "cc@example.com"},
				},
			},
			includeBody: false,
			checkFunc: func(resp EmailResponse) bool {
				return len(resp.From) == 1 && resp.From[0].Email == "sender@example.com" &&
					len(resp.To) == 1 && resp.To[0].Email == "recipient@example.com" &&
					len(resp.Cc) == 1 && resp.Cc[0].Email == "cc@example.com"
			},
			desc: "participants should be converted",
		},
		{
			name: "with attachments",
			message: domain.Message{
				ID:      "msg-005",
				Subject: "Subject",
				Attachments: []domain.Attachment{
					{ID: "att-001", Filename: "doc.pdf", ContentType: "application/pdf", Size: 1024},
					{ID: "att-002", Filename: "image.png", ContentType: "image/png", Size: 2048},
				},
			},
			includeBody: false,
			checkFunc: func(resp EmailResponse) bool {
				return len(resp.Attachments) == 2 &&
					resp.Attachments[0].Filename == "doc.pdf" &&
					resp.Attachments[1].Filename == "image.png"
			},
			desc: "attachments should be converted",
		},
		{
			name: "starred message",
			message: domain.Message{
				ID:      "msg-006",
				Subject: "Important",
				Starred: true,
			},
			includeBody: false,
			checkFunc: func(resp EmailResponse) bool {
				return resp.Starred == true
			},
			desc: "starred flag should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := emailToResponse(tt.message, tt.includeBody)

			if !tt.checkFunc(resp) {
				t.Errorf("emailToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}

// TestCachedEmailToResponse_Extended tests cached email conversion.
func TestCachedEmailToResponse_Extended(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		cached    *cache.CachedEmail
		checkFunc func(resp EmailResponse) bool
		desc      string
	}{
		{
			name: "basic cached email",
			cached: &cache.CachedEmail{
				ID:        "cached-001",
				ThreadID:  "thread-001",
				Subject:   "Cached Subject",
				Snippet:   "Cached snippet...",
				FromName:  "Sender Name",
				FromEmail: "sender@example.com",
				Date:      now,
				Unread:    true,
				Starred:   false,
				FolderID:  "inbox",
			},
			checkFunc: func(resp EmailResponse) bool {
				return resp.ID == "cached-001" &&
					resp.ThreadID == "thread-001" &&
					resp.Subject == "Cached Subject" &&
					len(resp.From) == 1 &&
					resp.From[0].Name == "Sender Name" &&
					resp.From[0].Email == "sender@example.com" &&
					resp.Unread == true &&
					len(resp.Folders) == 1 &&
					resp.Folders[0] == "inbox"
			},
			desc: "all fields should be converted correctly",
		},
		{
			name: "starred cached email",
			cached: &cache.CachedEmail{
				ID:       "cached-002",
				Subject:  "Starred",
				Starred:  true,
				FolderID: "starred",
			},
			checkFunc: func(resp EmailResponse) bool {
				return resp.Starred == true
			},
			desc: "starred should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := cachedEmailToResponse(tt.cached)

			if !tt.checkFunc(resp) {
				t.Errorf("cachedEmailToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}

// TestEventToResponse tests event domain to response conversion.
func TestEventToResponse(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		event     domain.Event
		checkFunc func(resp EventResponse) bool
		desc      string
	}{
		{
			name: "basic timed event",
			event: domain.Event{
				ID:          "evt-001",
				CalendarID:  "cal-001",
				Title:       "Meeting",
				Description: "Team sync",
				Location:    "Room A",
				When: domain.EventWhen{
					StartTime: now.Unix(),
					EndTime:   now.Add(1 * time.Hour).Unix(),
					Object:    "timespan",
				},
				Status: "confirmed",
				Busy:   true,
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.ID == "evt-001" &&
					resp.CalendarID == "cal-001" &&
					resp.Title == "Meeting" &&
					resp.Description == "Team sync" &&
					resp.Location == "Room A" &&
					resp.Status == "confirmed" &&
					resp.Busy == true &&
					resp.IsAllDay == false
			},
			desc: "basic event fields should match",
		},
		{
			name: "all-day event with date",
			event: domain.Event{
				ID:         "evt-002",
				CalendarID: "cal-001",
				Title:      "Holiday",
				When: domain.EventWhen{
					Date:   "2025-12-25",
					Object: "date",
				},
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.IsAllDay == true && resp.StartTime > 0
			},
			desc: "all-day event should have IsAllDay=true",
		},
		{
			name: "all-day event with date range",
			event: domain.Event{
				ID:         "evt-003",
				CalendarID: "cal-001",
				Title:      "Vacation",
				When: domain.EventWhen{
					StartDate: "2025-12-20",
					EndDate:   "2025-12-27",
					Object:    "datespan",
				},
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.IsAllDay == true &&
					resp.StartTime > 0 &&
					resp.EndTime > resp.StartTime
			},
			desc: "date range event should have proper start/end times",
		},
		{
			name: "event with participants",
			event: domain.Event{
				ID:    "evt-004",
				Title: "Team Meeting",
				Participants: []domain.Participant{
					{Person: domain.Person{Name: "Alice", Email: "alice@example.com"}, Status: "yes"},
					{Person: domain.Person{Name: "Bob", Email: "bob@example.com"}, Status: "maybe"},
				},
			},
			checkFunc: func(resp EventResponse) bool {
				return len(resp.Participants) == 2 &&
					resp.Participants[0].Name == "Alice" &&
					resp.Participants[0].Status == "yes" &&
					resp.Participants[1].Name == "Bob"
			},
			desc: "participants should be converted",
		},
		{
			name: "event with conferencing",
			event: domain.Event{
				ID:    "evt-005",
				Title: "Video Call",
				Conferencing: &domain.Conferencing{
					Provider: "Zoom",
					Details: &domain.ConferencingDetails{
						URL: "https://zoom.us/j/123456",
					},
				},
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.Conferencing != nil &&
					resp.Conferencing.Provider == "Zoom" &&
					resp.Conferencing.URL == "https://zoom.us/j/123456"
			},
			desc: "conferencing should be converted",
		},
		{
			name: "event without conferencing",
			event: domain.Event{
				ID:           "evt-006",
				Title:        "In-person Meeting",
				Conferencing: nil,
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.Conferencing == nil
			},
			desc: "nil conferencing should remain nil",
		},
		{
			name: "event with html link",
			event: domain.Event{
				ID:       "evt-007",
				Title:    "External Event",
				HtmlLink: "https://calendar.google.com/event/abc123",
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.HtmlLink == "https://calendar.google.com/event/abc123"
			},
			desc: "html link should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := eventToResponse(tt.event)

			if !tt.checkFunc(resp) {
				t.Errorf("eventToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}

// TestCachedEventToResponse_Extended tests cached event conversion.
func TestCachedEventToResponse_Extended(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		cached    *cache.CachedEvent
		checkFunc func(resp EventResponse) bool
		desc      string
	}{
		{
			name: "basic cached event",
			cached: &cache.CachedEvent{
				ID:          "cached-evt-001",
				CalendarID:  "cal-001",
				Title:       "Cached Meeting",
				Description: "Description",
				Location:    "Room B",
				StartTime:   now,
				EndTime:     now.Add(1 * time.Hour),
				AllDay:      false,
				Status:      "confirmed",
				Busy:        true,
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.ID == "cached-evt-001" &&
					resp.CalendarID == "cal-001" &&
					resp.Title == "Cached Meeting" &&
					resp.Location == "Room B" &&
					resp.IsAllDay == false &&
					resp.Busy == true
			},
			desc: "all fields should be converted correctly",
		},
		{
			name: "all-day cached event",
			cached: &cache.CachedEvent{
				ID:        "cached-evt-002",
				Title:     "All Day Event",
				AllDay:    true,
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
			},
			checkFunc: func(resp EventResponse) bool {
				return resp.IsAllDay == true
			},
			desc: "all-day flag should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := cachedEventToResponse(tt.cached)

			if !tt.checkFunc(resp) {
				t.Errorf("cachedEventToResponse() %s: got %+v", tt.desc, resp)
			}
		})
	}
}
