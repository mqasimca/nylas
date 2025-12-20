package nylas

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// DemoClient is a client that returns realistic demo data for screenshots and demos.
// It implements the ports.NylasClient interface without requiring any credentials.
type DemoClient struct{}

// NewDemoClient creates a new DemoClient for demo mode.
func NewDemoClient() *DemoClient {
	return &DemoClient{}
}

// SetRegion is a no-op for demo client.
func (d *DemoClient) SetRegion(region string) {}

// SetCredentials is a no-op for demo client.
func (d *DemoClient) SetCredentials(clientID, clientSecret, apiKey string) {}

// BuildAuthURL returns a mock auth URL.
func (d *DemoClient) BuildAuthURL(provider domain.Provider, redirectURI string) string {
	return "https://demo.nylas.com/auth"
}

// ExchangeCode returns a mock grant.
func (d *DemoClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*domain.Grant, error) {
	return &domain.Grant{
		ID:          "demo-grant-id",
		Email:       "demo@example.com",
		Provider:    domain.ProviderGoogle,
		GrantStatus: "valid",
	}, nil
}

// ListGrants returns demo grants.
func (d *DemoClient) ListGrants(ctx context.Context) ([]domain.Grant, error) {
	return []domain.Grant{
		{
			ID:          "demo-grant-001",
			Email:       "demo@example.com",
			Provider:    domain.ProviderGoogle,
			GrantStatus: "valid",
			CreatedAt:   domain.UnixTime{Time: time.Now().Add(-30 * 24 * time.Hour)},
		},
		{
			ID:          "demo-grant-002",
			Email:       "work@company.com",
			Provider:    domain.ProviderMicrosoft,
			GrantStatus: "valid",
			CreatedAt:   domain.UnixTime{Time: time.Now().Add(-7 * 24 * time.Hour)},
		},
	}, nil
}

// GetGrant returns a demo grant.
func (d *DemoClient) GetGrant(ctx context.Context, grantID string) (*domain.Grant, error) {
	return &domain.Grant{
		ID:          grantID,
		Email:       "demo@example.com",
		Provider:    domain.ProviderGoogle,
		GrantStatus: "valid",
	}, nil
}

// RevokeGrant is a no-op for demo client.
func (d *DemoClient) RevokeGrant(ctx context.Context, grantID string) error {
	return nil
}

// GetMessages returns demo messages.
func (d *DemoClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
	return d.getDemoMessages(), nil
}

// GetMessagesWithParams returns demo messages.
func (d *DemoClient) GetMessagesWithParams(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
	return d.getDemoMessages(), nil
}

// GetMessagesWithCursor returns demo messages with pagination.
func (d *DemoClient) GetMessagesWithCursor(ctx context.Context, grantID string, params *domain.MessageQueryParams) (*domain.MessageListResponse, error) {
	return &domain.MessageListResponse{
		Data: d.getDemoMessages(),
	}, nil
}

func (d *DemoClient) getDemoMessages() []domain.Message {
	now := time.Now()
	return []domain.Message{
		{
			ID:       "msg-001",
			Subject:  "Q4 Planning Meeting - Action Items",
			From:     []domain.EmailParticipant{{Name: "Sarah Chen", Email: "sarah.chen@company.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-15 * time.Minute),
			Unread:   true,
			Starred:  true,
			Snippet:  "Hi team, here are the action items from today's planning meeting...",
			Body:     "Hi team,\n\nHere are the action items from today's planning meeting:\n\n1. Review Q4 roadmap by Friday\n2. Submit budget proposals\n3. Schedule 1:1s with new team members\n\nBest,\nSarah",
			ThreadID: "thread-001",
		},
		{
			ID:       "msg-002",
			Subject:  "[GitHub] Pull request #247: Add dark mode support",
			From:     []domain.EmailParticipant{{Name: "GitHub", Email: "noreply@github.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-2 * time.Hour),
			Unread:   true,
			Starred:  false,
			Snippet:  "alex-dev requested your review on: Add dark mode support for the dashboard...",
			Body:     "alex-dev requested your review on:\n\nAdd dark mode support for the dashboard\n\n+156 -23 lines changed\n\nView pull request: https://github.com/example/repo/pull/247",
			ThreadID: "thread-002",
		},
		{
			ID:       "msg-003",
			Subject:  "Your AWS bill for December 2024",
			From:     []domain.EmailParticipant{{Name: "Amazon Web Services", Email: "billing@aws.amazon.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-5 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "Your AWS charges for December 2024 are $127.43. View your detailed bill...",
			Body:     "Hello,\n\nYour AWS charges for December 2024 are $127.43.\n\nView your detailed bill in the AWS Billing Console.\n\nThank you for using Amazon Web Services.",
			ThreadID: "thread-003",
		},
		{
			ID:       "msg-004",
			Subject:  "Re: Lunch tomorrow?",
			From:     []domain.EmailParticipant{{Name: "Mike Johnson", Email: "mike.j@gmail.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-1 * 24 * time.Hour),
			Unread:   false,
			Starred:  true,
			Snippet:  "Sounds great! How about that new Italian place on 5th? I heard they have...",
			Body:     "Sounds great! How about that new Italian place on 5th? I heard they have amazing pasta.\n\nLet's meet at 12:30?\n\n- Mike",
			ThreadID: "thread-004",
		},
		{
			ID:       "msg-005",
			Subject:  "Weekly Newsletter: Top Tech Stories",
			From:     []domain.EmailParticipant{{Name: "TechCrunch", Email: "newsletter@techcrunch.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-1*24*time.Hour - 3*time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "This week's top stories: AI breakthroughs, startup funding rounds, and more...",
			Body:     "This week's top stories:\n\n1. OpenAI announces new model\n2. Startup raises $50M Series B\n3. Apple's latest patent reveals AR glasses plans\n\nRead more at techcrunch.com",
			ThreadID: "thread-005",
		},
		{
			ID:       "msg-006",
			Subject:  "Your package has been delivered",
			From:     []domain.EmailParticipant{{Name: "Amazon", Email: "ship-confirm@amazon.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-2 * 24 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "Your package was delivered. It was handed directly to a resident...",
			Body:     "Your package was delivered.\n\nDelivered: December 15, 2024, 2:34 PM\nLeft with: Resident\n\nTrack your package at amazon.com/orders",
			ThreadID: "thread-006",
		},
		{
			ID:       "msg-007",
			Subject:  "Invitation: Team Standup @ Daily 9:00 AM",
			From:     []domain.EmailParticipant{{Name: "Google Calendar", Email: "calendar-notification@google.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-3 * 24 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "You've been invited to a recurring event: Team Standup...",
			Body:     "You've been invited to a recurring event.\n\nTeam Standup\nDaily at 9:00 AM - 9:15 AM\n\nJoin with Google Meet: meet.google.com/abc-defg-hij",
			ThreadID: "thread-007",
		},
		{
			ID:       "msg-008",
			Subject:  "Your Spotify Wrapped 2024 is here!",
			From:     []domain.EmailParticipant{{Name: "Spotify", Email: "no-reply@spotify.com"}},
			To:       []domain.EmailParticipant{{Name: "Demo User", Email: "demo@example.com"}},
			Date:     now.Add(-5 * 24 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "See your year in music. You listened to 47,832 minutes of music this year...",
			Body:     "Your 2024 Wrapped is here!\n\nYou listened to 47,832 minutes of music this year.\nYour top genre: Electronic\nYour top artist: Daft Punk\n\nSee your full Wrapped at spotify.com/wrapped",
			ThreadID: "thread-008",
		},
	}
}

// GetMessage returns a demo message.
func (d *DemoClient) GetMessage(ctx context.Context, grantID, messageID string) (*domain.Message, error) {
	messages := d.getDemoMessages()
	for _, msg := range messages {
		if msg.ID == messageID {
			return &msg, nil
		}
	}
	return &messages[0], nil
}

// SendMessage simulates sending a message.
func (d *DemoClient) SendMessage(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error) {
	msg := &domain.Message{
		ID:   "sent-demo-msg",
		Date: time.Now(),
	}
	if req != nil {
		msg.Subject = req.Subject
		msg.To = req.To
		msg.Body = req.Body
	}
	return msg, nil
}

// UpdateMessage simulates updating a message.
func (d *DemoClient) UpdateMessage(ctx context.Context, grantID, messageID string, req *domain.UpdateMessageRequest) (*domain.Message, error) {
	msg := &domain.Message{ID: messageID, Subject: "Updated Message"}
	if req.Unread != nil {
		msg.Unread = *req.Unread
	}
	if req.Starred != nil {
		msg.Starred = *req.Starred
	}
	return msg, nil
}

// DeleteMessage simulates deleting a message.
func (d *DemoClient) DeleteMessage(ctx context.Context, grantID, messageID string) error {
	return nil
}

// GetThreads returns demo threads.
func (d *DemoClient) GetThreads(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
	return d.getDemoThreads(), nil
}

func (d *DemoClient) getDemoThreads() []domain.Thread {
	now := time.Now()
	return []domain.Thread{
		{
			ID:                    "thread-001",
			Subject:               "Q4 Planning Meeting - Action Items",
			Unread:                true,
			Starred:               true,
			Snippet:               "Hi team, here are the action items from today's planning meeting...",
			LatestMessageRecvDate: now.Add(-15 * time.Minute),
			EarliestMessageDate:   now.Add(-2 * time.Hour),
			MessageIDs:            []string{"msg-001"},
			Participants: []domain.EmailParticipant{
				{Name: "Sarah Chen", Email: "sarah.chen@company.com"},
				{Name: "Demo User", Email: "demo@example.com"},
			},
			HasAttachments: true,
		},
		{
			ID:                    "thread-002",
			Subject:               "[GitHub] Pull request #247: Add dark mode support",
			Unread:                true,
			Starred:               false,
			Snippet:               "alex-dev requested your review on: Add dark mode support for the dashboard...",
			LatestMessageRecvDate: now.Add(-2 * time.Hour),
			EarliestMessageDate:   now.Add(-3 * time.Hour),
			MessageIDs:            []string{"msg-002", "msg-002b"},
			Participants: []domain.EmailParticipant{
				{Name: "GitHub", Email: "noreply@github.com"},
			},
		},
		{
			ID:                    "thread-003",
			Subject:               "Your AWS bill for December 2024",
			Unread:                false,
			Starred:               false,
			Snippet:               "Your AWS charges for December 2024 are $127.43. View your detailed bill...",
			LatestMessageRecvDate: now.Add(-5 * time.Hour),
			EarliestMessageDate:   now.Add(-5 * time.Hour),
			MessageIDs:            []string{"msg-003"},
			Participants: []domain.EmailParticipant{
				{Name: "Amazon Web Services", Email: "billing@aws.amazon.com"},
			},
		},
		{
			ID:                    "thread-004",
			Subject:               "Re: Lunch tomorrow?",
			Unread:                false,
			Starred:               true,
			Snippet:               "Sounds great! How about that new Italian place on 5th? I heard they have...",
			LatestMessageRecvDate: now.Add(-1 * 24 * time.Hour),
			EarliestMessageDate:   now.Add(-2 * 24 * time.Hour),
			MessageIDs:            []string{"msg-004", "msg-004b", "msg-004c"},
			Participants: []domain.EmailParticipant{
				{Name: "Mike Johnson", Email: "mike.j@gmail.com"},
				{Name: "Demo User", Email: "demo@example.com"},
			},
		},
		{
			ID:                    "thread-005",
			Subject:               "Weekly Newsletter: Top Tech Stories",
			Unread:                false,
			Starred:               false,
			Snippet:               "This week's top stories: AI breakthroughs, startup funding rounds, and more...",
			LatestMessageRecvDate: now.Add(-1*24*time.Hour - 3*time.Hour),
			EarliestMessageDate:   now.Add(-1*24*time.Hour - 3*time.Hour),
			MessageIDs:            []string{"msg-005"},
			Participants: []domain.EmailParticipant{
				{Name: "TechCrunch", Email: "newsletter@techcrunch.com"},
			},
		},
		{
			ID:                    "thread-006",
			Subject:               "Your package has been delivered",
			Unread:                false,
			Starred:               false,
			Snippet:               "Your package was delivered. It was handed directly to a resident...",
			LatestMessageRecvDate: now.Add(-2 * 24 * time.Hour),
			EarliestMessageDate:   now.Add(-3 * 24 * time.Hour),
			MessageIDs:            []string{"msg-006", "msg-006b"},
			Participants: []domain.EmailParticipant{
				{Name: "Amazon", Email: "ship-confirm@amazon.com"},
			},
		},
		{
			ID:                    "thread-007",
			Subject:               "Invitation: Team Standup @ Daily 9:00 AM",
			Unread:                false,
			Starred:               false,
			Snippet:               "You've been invited to a recurring event: Team Standup...",
			LatestMessageRecvDate: now.Add(-3 * 24 * time.Hour),
			EarliestMessageDate:   now.Add(-3 * 24 * time.Hour),
			MessageIDs:            []string{"msg-007"},
			Participants: []domain.EmailParticipant{
				{Name: "Google Calendar", Email: "calendar-notification@google.com"},
			},
		},
		{
			ID:                    "thread-008",
			Subject:               "Your Spotify Wrapped 2024 is here!",
			Unread:                false,
			Starred:               false,
			Snippet:               "See your year in music. You listened to 47,832 minutes of music this year...",
			LatestMessageRecvDate: now.Add(-5 * 24 * time.Hour),
			EarliestMessageDate:   now.Add(-5 * 24 * time.Hour),
			MessageIDs:            []string{"msg-008"},
			Participants: []domain.EmailParticipant{
				{Name: "Spotify", Email: "no-reply@spotify.com"},
			},
		},
	}
}

// GetThread returns a demo thread.
func (d *DemoClient) GetThread(ctx context.Context, grantID, threadID string) (*domain.Thread, error) {
	threads := d.getDemoThreads()
	for _, t := range threads {
		if t.ID == threadID {
			return &t, nil
		}
	}
	return &threads[0], nil
}

// UpdateThread simulates updating a thread.
func (d *DemoClient) UpdateThread(ctx context.Context, grantID, threadID string, req *domain.UpdateMessageRequest) (*domain.Thread, error) {
	thread := &domain.Thread{ID: threadID, Subject: "Updated Thread"}
	if req.Unread != nil {
		thread.Unread = *req.Unread
	}
	if req.Starred != nil {
		thread.Starred = *req.Starred
	}
	return thread, nil
}

// DeleteThread simulates deleting a thread.
func (d *DemoClient) DeleteThread(ctx context.Context, grantID, threadID string) error {
	return nil
}

// GetDrafts returns demo drafts.
func (d *DemoClient) GetDrafts(ctx context.Context, grantID string, limit int) ([]domain.Draft, error) {
	return []domain.Draft{
		{
			ID:      "draft-001",
			Subject: "Re: Project proposal",
			To:      []domain.EmailParticipant{{Email: "client@company.com"}},
			Body:    "Thank you for the proposal...",
		},
	}, nil
}

// GetDraft returns a demo draft.
func (d *DemoClient) GetDraft(ctx context.Context, grantID, draftID string) (*domain.Draft, error) {
	return &domain.Draft{
		ID:      draftID,
		Subject: "Re: Project proposal",
		Body:    "Thank you for the proposal...",
	}, nil
}

// CreateDraft simulates creating a draft.
func (d *DemoClient) CreateDraft(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	return &domain.Draft{ID: "new-draft", Subject: req.Subject, Body: req.Body, To: req.To}, nil
}

// UpdateDraft simulates updating a draft.
func (d *DemoClient) UpdateDraft(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	return &domain.Draft{ID: draftID, Subject: req.Subject, Body: req.Body, To: req.To}, nil
}

// DeleteDraft simulates deleting a draft.
func (d *DemoClient) DeleteDraft(ctx context.Context, grantID, draftID string) error {
	return nil
}

// SendDraft simulates sending a draft.
func (d *DemoClient) SendDraft(ctx context.Context, grantID, draftID string) (*domain.Message, error) {
	return &domain.Message{ID: "sent-from-draft", Subject: "Sent Draft"}, nil
}

// GetFolders returns demo folders.
func (d *DemoClient) GetFolders(ctx context.Context, grantID string) ([]domain.Folder, error) {
	return []domain.Folder{
		{ID: "inbox", Name: "Inbox", SystemFolder: "inbox", TotalCount: 1247},
		{ID: "sent", Name: "Sent", SystemFolder: "sent", TotalCount: 532},
		{ID: "drafts", Name: "Drafts", SystemFolder: "drafts", TotalCount: 3},
		{ID: "trash", Name: "Trash", SystemFolder: "trash", TotalCount: 45},
		{ID: "work", Name: "Work", TotalCount: 89},
		{ID: "personal", Name: "Personal", TotalCount: 156},
	}, nil
}

// GetFolder returns a demo folder.
func (d *DemoClient) GetFolder(ctx context.Context, grantID, folderID string) (*domain.Folder, error) {
	return &domain.Folder{ID: folderID, Name: "Demo Folder"}, nil
}

// CreateFolder simulates creating a folder.
func (d *DemoClient) CreateFolder(ctx context.Context, grantID string, req *domain.CreateFolderRequest) (*domain.Folder, error) {
	return &domain.Folder{ID: "new-folder", Name: req.Name}, nil
}

// UpdateFolder simulates updating a folder.
func (d *DemoClient) UpdateFolder(ctx context.Context, grantID, folderID string, req *domain.UpdateFolderRequest) (*domain.Folder, error) {
	return &domain.Folder{ID: folderID, Name: req.Name}, nil
}

// DeleteFolder simulates deleting a folder.
func (d *DemoClient) DeleteFolder(ctx context.Context, grantID, folderID string) error {
	return nil
}

// ListAttachments returns demo attachments for a message.
func (d *DemoClient) ListAttachments(ctx context.Context, grantID, messageID string) ([]domain.Attachment, error) {
	return []domain.Attachment{
		{
			ID:          "attach-001",
			Filename:    "quarterly-report.pdf",
			ContentType: "application/pdf",
			Size:        245760,
		},
		{
			ID:          "attach-002",
			Filename:    "presentation.pptx",
			ContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			Size:        1048576,
		},
	}, nil
}

// GetAttachment returns demo attachment metadata.
func (d *DemoClient) GetAttachment(ctx context.Context, grantID, messageID, attachmentID string) (*domain.Attachment, error) {
	return &domain.Attachment{
		ID:          attachmentID,
		Filename:    "quarterly-report.pdf",
		ContentType: "application/pdf",
		Size:        245760,
	}, nil
}

// DownloadAttachment returns mock attachment content.
func (d *DemoClient) DownloadAttachment(ctx context.Context, grantID, messageID, attachmentID string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("demo attachment content")), nil
}

// GetCalendars returns demo calendars.
func (d *DemoClient) GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error) {
	return []domain.Calendar{
		{ID: "primary", Name: "Personal", IsPrimary: true, HexColor: "#4285F4"},
		{ID: "work", Name: "Work", IsPrimary: false, HexColor: "#0F9D58"},
		{ID: "family", Name: "Family", IsPrimary: false, HexColor: "#DB4437"},
	}, nil
}

// GetCalendar returns a demo calendar.
func (d *DemoClient) GetCalendar(ctx context.Context, grantID, calendarID string) (*domain.Calendar, error) {
	return &domain.Calendar{ID: calendarID, Name: "Demo Calendar", IsPrimary: true}, nil
}

// CreateCalendar simulates creating a calendar.
func (d *DemoClient) CreateCalendar(ctx context.Context, grantID string, req *domain.CreateCalendarRequest) (*domain.Calendar, error) {
	return &domain.Calendar{
		ID:          "new-demo-calendar",
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Timezone:    req.Timezone,
	}, nil
}

// UpdateCalendar simulates updating a calendar.
func (d *DemoClient) UpdateCalendar(ctx context.Context, grantID, calendarID string, req *domain.UpdateCalendarRequest) (*domain.Calendar, error) {
	cal := &domain.Calendar{ID: calendarID}
	if req.Name != nil {
		cal.Name = *req.Name
	}
	if req.Description != nil {
		cal.Description = *req.Description
	}
	return cal, nil
}

// DeleteCalendar simulates deleting a calendar.
func (d *DemoClient) DeleteCalendar(ctx context.Context, grantID, calendarID string) error {
	return nil
}

// GetEvents returns demo events.
func (d *DemoClient) GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	return d.getDemoEvents(), nil
}

// GetEventsWithCursor returns demo events with pagination.
func (d *DemoClient) GetEventsWithCursor(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) (*domain.EventListResponse, error) {
	return &domain.EventListResponse{Data: d.getDemoEvents()}, nil
}

func (d *DemoClient) getDemoEvents() []domain.Event {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return []domain.Event{
		{
			ID:         "event-001",
			CalendarID: "primary",
			Title:      "Team Standup",
			When: domain.EventWhen{
				StartTime: today.Add(9 * time.Hour).Unix(),
				EndTime:   today.Add(9*time.Hour + 15*time.Minute).Unix(),
			},
			Status:   "confirmed",
			Location: "Conference Room A",
			Participants: []domain.Participant{
				{Name: "Sarah Chen", Email: "sarah@company.com", Status: "yes"},
				{Name: "Mike Johnson", Email: "mike@company.com", Status: "yes"},
				{Name: "Demo User", Email: "demo@example.com", Status: "yes"},
			},
		},
		{
			ID:         "event-002",
			CalendarID: "primary",
			Title:      "1:1 with Manager",
			When: domain.EventWhen{
				StartTime: today.Add(11 * time.Hour).Unix(),
				EndTime:   today.Add(11*time.Hour + 30*time.Minute).Unix(),
			},
			Status:       "confirmed",
			Location:     "Google Meet",
			Conferencing: &domain.Conferencing{Provider: "Google Meet", Details: &domain.ConferencingDetails{URL: "https://meet.google.com/abc-defg-hij"}},
		},
		{
			ID:         "event-003",
			CalendarID: "primary",
			Title:      "Lunch Break",
			When: domain.EventWhen{
				StartTime: today.Add(12 * time.Hour).Unix(),
				EndTime:   today.Add(13 * time.Hour).Unix(),
			},
			Status: "confirmed",
		},
		{
			ID:         "event-004",
			CalendarID: "work",
			Title:      "Project Review",
			When: domain.EventWhen{
				StartTime: today.Add(14 * time.Hour).Unix(),
				EndTime:   today.Add(15 * time.Hour).Unix(),
			},
			Status:      "confirmed",
			Location:    "Main Office - Room 302",
			Description: "Quarterly project review with stakeholders",
			Participants: []domain.Participant{
				{Name: "Product Team", Email: "product@company.com", Status: "yes"},
				{Name: "Engineering", Email: "eng@company.com", Status: "maybe"},
			},
		},
		{
			ID:         "event-005",
			CalendarID: "primary",
			Title:      "Dentist Appointment",
			When: domain.EventWhen{
				StartTime: today.Add(24*time.Hour + 10*time.Hour).Unix(),
				EndTime:   today.Add(24*time.Hour + 11*time.Hour).Unix(),
			},
			Status:   "confirmed",
			Location: "123 Health St, Suite 400",
		},
		{
			ID:         "event-006",
			CalendarID: "family",
			Title:      "Birthday Party - Mom",
			When: domain.EventWhen{
				StartTime: today.Add(3*24*time.Hour + 18*time.Hour).Unix(),
				EndTime:   today.Add(3*24*time.Hour + 21*time.Hour).Unix(),
			},
			Status:   "confirmed",
			Location: "Family Home",
		},
		{
			ID:         "event-007",
			CalendarID: "primary",
			Title:      "Gym Session",
			When: domain.EventWhen{
				StartTime: today.Add(24*time.Hour + 7*time.Hour).Unix(),
				EndTime:   today.Add(24*time.Hour + 8*time.Hour).Unix(),
			},
			Status:   "confirmed",
			Location: "Downtown Fitness",
		},
	}
}

// GetEvent returns a demo event.
func (d *DemoClient) GetEvent(ctx context.Context, grantID, calendarID, eventID string) (*domain.Event, error) {
	events := d.getDemoEvents()
	for _, event := range events {
		if event.ID == eventID {
			return &event, nil
		}
	}
	return &events[0], nil
}

// CreateEvent simulates creating an event.
func (d *DemoClient) CreateEvent(ctx context.Context, grantID, calendarID string, req *domain.CreateEventRequest) (*domain.Event, error) {
	return &domain.Event{ID: "new-event", CalendarID: calendarID, Title: req.Title}, nil
}

// UpdateEvent simulates updating an event.
func (d *DemoClient) UpdateEvent(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {
	event := &domain.Event{ID: eventID, CalendarID: calendarID}
	if req.Title != nil {
		event.Title = *req.Title
	}
	return event, nil
}

// DeleteEvent simulates deleting an event.
func (d *DemoClient) DeleteEvent(ctx context.Context, grantID, calendarID, eventID string) error {
	return nil
}

// SendRSVP simulates sending an RSVP response.
func (d *DemoClient) SendRSVP(ctx context.Context, grantID, calendarID, eventID string, req *domain.SendRSVPRequest) error {
	return nil
}

// GetFreeBusy returns demo free/busy information.
func (d *DemoClient) GetFreeBusy(ctx context.Context, grantID string, req *domain.FreeBusyRequest) (*domain.FreeBusyResponse, error) {
	result := &domain.FreeBusyResponse{
		Data: make([]domain.FreeBusyCalendar, len(req.Emails)),
	}
	for i, email := range req.Emails {
		result.Data[i] = domain.FreeBusyCalendar{
			Email: email,
			TimeSlots: []domain.TimeSlot{
				{StartTime: req.StartTime + 3600, EndTime: req.StartTime + 7200, Status: "busy"},
			},
		}
	}
	return result, nil
}

// GetAvailability returns demo availability slots.
func (d *DemoClient) GetAvailability(ctx context.Context, req *domain.AvailabilityRequest) (*domain.AvailabilityResponse, error) {
	duration := int64(req.DurationMinutes * 60)
	return &domain.AvailabilityResponse{
		Data: domain.AvailabilityData{
			TimeSlots: []domain.AvailableSlot{
				{StartTime: req.StartTime + 7200, EndTime: req.StartTime + 7200 + duration},
				{StartTime: req.StartTime + 14400, EndTime: req.StartTime + 14400 + duration},
			},
		},
	}, nil
}

// GetContacts returns demo contacts.
func (d *DemoClient) GetContacts(ctx context.Context, grantID string, params *domain.ContactQueryParams) ([]domain.Contact, error) {
	return d.getDemoContacts(), nil
}

// GetContactsWithCursor returns demo contacts with pagination.
func (d *DemoClient) GetContactsWithCursor(ctx context.Context, grantID string, params *domain.ContactQueryParams) (*domain.ContactListResponse, error) {
	return &domain.ContactListResponse{Data: d.getDemoContacts()}, nil
}

func (d *DemoClient) getDemoContacts() []domain.Contact {
	return []domain.Contact{
		{
			ID:           "contact-001",
			GivenName:    "Sarah",
			Surname:      "Chen",
			Emails:       []domain.ContactEmail{{Email: "sarah.chen@company.com", Type: "work"}},
			PhoneNumbers: []domain.ContactPhone{{Number: "+1-555-0101", Type: "mobile"}},
			CompanyName:  "Acme Corp",
			JobTitle:     "Engineering Manager",
		},
		{
			ID:           "contact-002",
			GivenName:    "Mike",
			Surname:      "Johnson",
			Emails:       []domain.ContactEmail{{Email: "mike.j@gmail.com", Type: "personal"}},
			PhoneNumbers: []domain.ContactPhone{{Number: "+1-555-0102", Type: "mobile"}},
		},
		{
			ID:           "contact-003",
			GivenName:    "Emily",
			Surname:      "Williams",
			Emails:       []domain.ContactEmail{{Email: "emily.w@startup.io", Type: "work"}},
			PhoneNumbers: []domain.ContactPhone{{Number: "+1-555-0103", Type: "work"}},
			CompanyName:  "TechStart Inc",
			JobTitle:     "CEO",
		},
		{
			ID:          "contact-004",
			GivenName:   "Alex",
			Surname:     "Kumar",
			Emails:      []domain.ContactEmail{{Email: "alex.kumar@dev.com", Type: "work"}},
			CompanyName: "DevOps Solutions",
			JobTitle:    "Senior Developer",
		},
		{
			ID:           "contact-005",
			GivenName:    "Jessica",
			Surname:      "Martinez",
			Emails:       []domain.ContactEmail{{Email: "jess.m@design.co", Type: "work"}},
			PhoneNumbers: []domain.ContactPhone{{Number: "+1-555-0105", Type: "mobile"}},
			CompanyName:  "Creative Design Co",
			JobTitle:     "Lead Designer",
		},
		{
			ID:          "contact-006",
			GivenName:   "David",
			Surname:     "Brown",
			Emails:      []domain.ContactEmail{{Email: "david.b@consulting.com", Type: "work"}},
			CompanyName: "Brown Consulting",
			JobTitle:    "Consultant",
		},
	}
}

// GetContact returns a demo contact.
func (d *DemoClient) GetContact(ctx context.Context, grantID, contactID string) (*domain.Contact, error) {
	contacts := d.getDemoContacts()
	for _, contact := range contacts {
		if contact.ID == contactID {
			return &contact, nil
		}
	}
	return &contacts[0], nil
}

// GetContactWithPicture returns a demo contact with optional profile picture.
func (d *DemoClient) GetContactWithPicture(ctx context.Context, grantID, contactID string, includePicture bool) (*domain.Contact, error) {
	contact, err := d.GetContact(ctx, grantID, contactID)
	if err != nil {
		return nil, err
	}
	if includePicture {
		// Demo base64-encoded 1x1 pixel image
		contact.Picture = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	}
	return contact, nil
}

// CreateContact simulates creating a contact.
func (d *DemoClient) CreateContact(ctx context.Context, grantID string, req *domain.CreateContactRequest) (*domain.Contact, error) {
	return &domain.Contact{ID: "new-contact", GivenName: req.GivenName, Surname: req.Surname}, nil
}

// UpdateContact simulates updating a contact.
func (d *DemoClient) UpdateContact(ctx context.Context, grantID, contactID string, req *domain.UpdateContactRequest) (*domain.Contact, error) {
	contact := &domain.Contact{ID: contactID}
	if req.GivenName != nil {
		contact.GivenName = *req.GivenName
	}
	if req.Surname != nil {
		contact.Surname = *req.Surname
	}
	return contact, nil
}

// DeleteContact simulates deleting a contact.
func (d *DemoClient) DeleteContact(ctx context.Context, grantID, contactID string) error {
	return nil
}

// GetContactGroups returns demo contact groups.
func (d *DemoClient) GetContactGroups(ctx context.Context, grantID string) ([]domain.ContactGroup, error) {
	return []domain.ContactGroup{
		{ID: "group-001", Name: "Coworkers"},
		{ID: "group-002", Name: "Friends"},
		{ID: "group-003", Name: "Family"},
		{ID: "group-004", Name: "VIP"},
	}, nil
}

// GetContactGroup returns a demo contact group.
func (d *DemoClient) GetContactGroup(ctx context.Context, grantID, groupID string) (*domain.ContactGroup, error) {
	return &domain.ContactGroup{
		ID:      groupID,
		GrantID: grantID,
		Name:    "Demo Group",
	}, nil
}

// CreateContactGroup creates a demo contact group.
func (d *DemoClient) CreateContactGroup(ctx context.Context, grantID string, req *domain.CreateContactGroupRequest) (*domain.ContactGroup, error) {
	return &domain.ContactGroup{
		ID:      "group-new",
		GrantID: grantID,
		Name:    req.Name,
	}, nil
}

// UpdateContactGroup updates a demo contact group.
func (d *DemoClient) UpdateContactGroup(ctx context.Context, grantID, groupID string, req *domain.UpdateContactGroupRequest) (*domain.ContactGroup, error) {
	name := "Updated Group"
	if req.Name != nil {
		name = *req.Name
	}
	return &domain.ContactGroup{
		ID:      groupID,
		GrantID: grantID,
		Name:    name,
	}, nil
}

// DeleteContactGroup deletes a demo contact group.
func (d *DemoClient) DeleteContactGroup(ctx context.Context, grantID, groupID string) error {
	return nil
}

// ListWebhooks returns demo webhooks.
func (d *DemoClient) ListWebhooks(ctx context.Context) ([]domain.Webhook, error) {
	return []domain.Webhook{
		{
			ID:           "webhook-001",
			Description:  "Message notifications",
			TriggerTypes: []string{domain.TriggerMessageCreated, domain.TriggerMessageUpdated},
			WebhookURL:   "https://api.myapp.com/webhooks/nylas",
			Status:       "active",
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:           "webhook-002",
			Description:  "Calendar sync",
			TriggerTypes: []string{domain.TriggerEventCreated, domain.TriggerEventUpdated},
			WebhookURL:   "https://api.myapp.com/calendar/sync",
			Status:       "active",
			CreatedAt:    time.Now().Add(-14 * 24 * time.Hour),
		},
		{
			ID:           "webhook-003",
			Description:  "Contact updates (paused)",
			TriggerTypes: []string{domain.TriggerContactCreated},
			WebhookURL:   "https://api.myapp.com/contacts",
			Status:       "inactive",
			CreatedAt:    time.Now().Add(-7 * 24 * time.Hour),
		},
	}, nil
}

// GetWebhook returns a demo webhook.
func (d *DemoClient) GetWebhook(ctx context.Context, webhookID string) (*domain.Webhook, error) {
	webhooks, _ := d.ListWebhooks(ctx)
	for _, webhook := range webhooks {
		if webhook.ID == webhookID {
			return &webhook, nil
		}
	}
	return &webhooks[0], nil
}

// CreateWebhook simulates creating a webhook.
func (d *DemoClient) CreateWebhook(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	return &domain.Webhook{
		ID:            "new-webhook",
		Description:   req.Description,
		TriggerTypes:  req.TriggerTypes,
		WebhookURL:    req.WebhookURL,
		WebhookSecret: "wh_secret_demo_12345",
		Status:        "active",
	}, nil
}

// UpdateWebhook simulates updating a webhook.
func (d *DemoClient) UpdateWebhook(ctx context.Context, webhookID string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	return &domain.Webhook{ID: webhookID, Description: req.Description, Status: req.Status}, nil
}

// DeleteWebhook simulates deleting a webhook.
func (d *DemoClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	return nil
}

// SendWebhookTestEvent simulates sending a test event.
func (d *DemoClient) SendWebhookTestEvent(ctx context.Context, webhookURL string) error {
	return nil
}

// GetWebhookMockPayload returns a mock payload for a trigger type.
func (d *DemoClient) GetWebhookMockPayload(ctx context.Context, triggerType string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"specversion": "1.0",
		"type":        triggerType,
		"source":      "/nylas/demo",
		"id":          "demo-event-id",
		"data":        map[string]interface{}{"object": map[string]interface{}{"id": "demo-object-id"}},
	}, nil
}

// ListScheduledMessages returns demo scheduled messages.
func (d *DemoClient) ListScheduledMessages(ctx context.Context, grantID string) ([]domain.ScheduledMessage, error) {
	now := time.Now()
	return []domain.ScheduledMessage{
		{
			ScheduleID: "schedule-001",
			Status:     "scheduled",
			CloseTime:  now.Add(1 * time.Hour).Unix(),
		},
		{
			ScheduleID: "schedule-002",
			Status:     "scheduled",
			CloseTime:  now.Add(24 * time.Hour).Unix(),
		},
	}, nil
}

// GetScheduledMessage returns a demo scheduled message.
func (d *DemoClient) GetScheduledMessage(ctx context.Context, grantID, scheduleID string) (*domain.ScheduledMessage, error) {
	return &domain.ScheduledMessage{
		ScheduleID: scheduleID,
		Status:     "scheduled",
		CloseTime:  time.Now().Add(1 * time.Hour).Unix(),
	}, nil
}

// CancelScheduledMessage simulates canceling a scheduled message.
func (d *DemoClient) CancelScheduledMessage(ctx context.Context, grantID, scheduleID string) error {
	return nil
}

// SmartCompose generates an AI-powered email draft based on a prompt.
func (d *DemoClient) SmartCompose(ctx context.Context, grantID string, req *domain.SmartComposeRequest) (*domain.SmartComposeSuggestion, error) {
	// Generate realistic demo response based on prompt
	suggestion := "Dear Colleague,\n\nThank you for reaching out. "
	if req != nil && req.Prompt != "" {
		suggestion += "I understand you'd like to " + req.Prompt + ". "
	}
	suggestion += "I've reviewed your request and wanted to follow up with some thoughts.\n\n"
	suggestion += "Based on our previous discussions, I believe we can move forward with this initiative. "
	suggestion += "I'll coordinate with the team and get back to you with a detailed plan by the end of the week.\n\n"
	suggestion += "Please let me know if you have any questions or need clarification on any points.\n\n"
	suggestion += "Best regards"

	return &domain.SmartComposeSuggestion{
		Suggestion: suggestion,
	}, nil
}

// SmartComposeReply generates an AI-powered reply to a specific message.
func (d *DemoClient) SmartComposeReply(ctx context.Context, grantID, messageID string, req *domain.SmartComposeRequest) (*domain.SmartComposeSuggestion, error) {
	// Generate realistic demo reply based on prompt
	suggestion := "Hi there,\n\nThank you for your message. "
	if req != nil && req.Prompt != "" {
		suggestion += req.Prompt + "\n\n"
	}
	suggestion += "I've taken a look at what you sent and wanted to respond quickly. "
	suggestion += "Your points are well-taken, and I agree with your assessment.\n\n"
	suggestion += "I'll review the details more thoroughly and provide a comprehensive response shortly. "
	suggestion += "In the meantime, please don't hesitate to reach out if you have any urgent concerns.\n\n"
	suggestion += "Thanks again for bringing this to my attention.\n\n"
	suggestion += "Best"

	return &domain.SmartComposeSuggestion{
		Suggestion: suggestion,
	}, nil
}

// ListNotetakers returns demo notetakers.
func (d *DemoClient) ListNotetakers(ctx context.Context, grantID string, params *domain.NotetakerQueryParams) ([]domain.Notetaker, error) {
	now := time.Now()
	return []domain.Notetaker{
		{
			ID:           "notetaker-001",
			State:        domain.NotetakerStateComplete,
			MeetingLink:  "https://zoom.us/j/123456789",
			MeetingTitle: "Q4 Planning Meeting",
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
		},
		{
			ID:           "notetaker-002",
			State:        domain.NotetakerStateAttending,
			MeetingLink:  "https://meet.google.com/abc-defg-hij",
			MeetingTitle: "Weekly Standup",
			CreatedAt:    now.Add(-30 * time.Minute),
			UpdatedAt:    now.Add(-5 * time.Minute),
		},
		{
			ID:           "notetaker-003",
			State:        domain.NotetakerStateScheduled,
			MeetingLink:  "https://teams.microsoft.com/l/meetup-join/xyz",
			MeetingTitle: "Client Demo",
			JoinTime:     now.Add(2 * time.Hour),
			CreatedAt:    now.Add(-24 * time.Hour),
			UpdatedAt:    now.Add(-24 * time.Hour),
		},
	}, nil
}

// GetNotetaker returns a demo notetaker.
func (d *DemoClient) GetNotetaker(ctx context.Context, grantID, notetakerID string) (*domain.Notetaker, error) {
	notetakers, _ := d.ListNotetakers(ctx, grantID, nil)
	for _, nt := range notetakers {
		if nt.ID == notetakerID {
			return &nt, nil
		}
	}
	return &notetakers[0], nil
}

// CreateNotetaker simulates creating a notetaker.
func (d *DemoClient) CreateNotetaker(ctx context.Context, grantID string, req *domain.CreateNotetakerRequest) (*domain.Notetaker, error) {
	now := time.Now()
	nt := &domain.Notetaker{
		ID:          "new-notetaker",
		State:       domain.NotetakerStateScheduled,
		MeetingLink: req.MeetingLink,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.JoinTime > 0 {
		nt.JoinTime = time.Unix(req.JoinTime, 0)
	}
	if req.BotConfig != nil {
		nt.BotConfig = req.BotConfig
	}
	return nt, nil
}

// DeleteNotetaker simulates deleting a notetaker.
func (d *DemoClient) DeleteNotetaker(ctx context.Context, grantID, notetakerID string) error {
	return nil
}

// GetNotetakerMedia returns demo notetaker media.
func (d *DemoClient) GetNotetakerMedia(ctx context.Context, grantID, notetakerID string) (*domain.MediaData, error) {
	return &domain.MediaData{
		Recording: &domain.MediaFile{
			URL:         "https://storage.nylas.com/recordings/demo-recording.mp4",
			ContentType: "video/mp4",
			Size:        125829120, // 120 MB
			ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
		},
		Transcript: &domain.MediaFile{
			URL:         "https://storage.nylas.com/transcripts/demo-transcript.json",
			ContentType: "application/json",
			Size:        51200, // 50 KB
			ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
		},
	}, nil
}

// ListInboundInboxes returns demo inbound inboxes.
func (d *DemoClient) ListInboundInboxes(ctx context.Context) ([]domain.InboundInbox, error) {
	now := time.Now()
	return []domain.InboundInbox{
		{
			ID:          "inbox-demo-001",
			Email:       "support@demo-app.nylas.email",
			GrantStatus: "valid",
			CreatedAt:   domain.UnixTime{Time: now.Add(-30 * 24 * time.Hour)},
			UpdatedAt:   domain.UnixTime{Time: now.Add(-1 * time.Hour)},
		},
		{
			ID:          "inbox-demo-002",
			Email:       "sales@demo-app.nylas.email",
			GrantStatus: "valid",
			CreatedAt:   domain.UnixTime{Time: now.Add(-14 * 24 * time.Hour)},
			UpdatedAt:   domain.UnixTime{Time: now.Add(-2 * time.Hour)},
		},
		{
			ID:          "inbox-demo-003",
			Email:       "info@demo-app.nylas.email",
			GrantStatus: "valid",
			CreatedAt:   domain.UnixTime{Time: now.Add(-7 * 24 * time.Hour)},
			UpdatedAt:   domain.UnixTime{Time: now.Add(-30 * time.Minute)},
		},
	}, nil
}

// GetInboundInbox returns a demo inbound inbox.
func (d *DemoClient) GetInboundInbox(ctx context.Context, grantID string) (*domain.InboundInbox, error) {
	inboxes, _ := d.ListInboundInboxes(ctx)
	for _, inbox := range inboxes {
		if inbox.ID == grantID {
			return &inbox, nil
		}
	}
	return &inboxes[0], nil
}

// CreateInboundInbox simulates creating an inbound inbox.
func (d *DemoClient) CreateInboundInbox(ctx context.Context, email string) (*domain.InboundInbox, error) {
	now := time.Now()
	return &domain.InboundInbox{
		ID:          "inbox-new",
		Email:       email + "@demo-app.nylas.email",
		GrantStatus: "valid",
		CreatedAt:   domain.UnixTime{Time: now},
		UpdatedAt:   domain.UnixTime{Time: now},
	}, nil
}

// DeleteInboundInbox simulates deleting an inbound inbox.
func (d *DemoClient) DeleteInboundInbox(ctx context.Context, grantID string) error {
	return nil
}

// GetInboundMessages returns demo inbound messages.
func (d *DemoClient) GetInboundMessages(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.InboundMessage, error) {
	now := time.Now()
	return []domain.InboundMessage{
		{
			ID:       "inbound-001",
			GrantID:  grantID,
			Subject:  "New Lead: Enterprise Plan Inquiry",
			From:     []domain.EmailParticipant{{Name: "John Smith", Email: "john.smith@bigcorp.com"}},
			To:       []domain.EmailParticipant{{Name: "Sales", Email: "sales@demo-app.nylas.email"}},
			Date:     now.Add(-10 * time.Minute),
			Unread:   true,
			Starred:  true,
			Snippet:  "Hi, I'm interested in learning more about your enterprise plan...",
			Body:     "Hi,\n\nI'm interested in learning more about your enterprise plan. Our company has about 500 employees and we're looking for a solution that can scale with our growth.\n\nCan we schedule a call this week?\n\nBest,\nJohn Smith\nVP of Engineering\nBigCorp Inc.",
			ThreadID: "inbound-thread-001",
		},
		{
			ID:       "inbound-002",
			GrantID:  grantID,
			Subject:  "Support Request: Integration Help",
			From:     []domain.EmailParticipant{{Name: "Sarah Johnson", Email: "sarah@startup.io"}},
			To:       []domain.EmailParticipant{{Name: "Support", Email: "support@demo-app.nylas.email"}},
			Date:     now.Add(-1 * time.Hour),
			Unread:   true,
			Starred:  false,
			Snippet:  "We're having trouble connecting our calendar integration...",
			Body:     "Hello,\n\nWe're having trouble connecting our calendar integration. The OAuth flow completes but we're not seeing any events sync.\n\nCan you help troubleshoot?\n\nThanks,\nSarah",
			ThreadID: "inbound-thread-002",
		},
		{
			ID:       "inbound-003",
			GrantID:  grantID,
			Subject:  "Partnership Opportunity",
			From:     []domain.EmailParticipant{{Name: "Mike Chen", Email: "mike@partner-company.com"}},
			To:       []domain.EmailParticipant{{Name: "Info", Email: "info@demo-app.nylas.email"}},
			Date:     now.Add(-3 * time.Hour),
			Unread:   false,
			Starred:  true,
			Snippet:  "We're a SaaS company looking for email integration partners...",
			Body:     "Hi there,\n\nWe're a SaaS company serving the healthcare industry and we're looking for email integration partners.\n\nWould love to explore a potential partnership.\n\nBest,\nMike Chen\nBusiness Development",
			ThreadID: "inbound-thread-003",
		},
		{
			ID:       "inbound-004",
			GrantID:  grantID,
			Subject:  "Billing Question",
			From:     []domain.EmailParticipant{{Name: "Lisa Park", Email: "lisa@customer.com"}},
			To:       []domain.EmailParticipant{{Name: "Support", Email: "support@demo-app.nylas.email"}},
			Date:     now.Add(-1 * 24 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "I have a question about my latest invoice...",
			Body:     "Hi,\n\nI have a question about my latest invoice. It seems like I was charged for 15 seats but we only have 10 active users.\n\nCan you look into this?\n\nThanks,\nLisa",
			ThreadID: "inbound-thread-004",
		},
		{
			ID:       "inbound-005",
			GrantID:  grantID,
			Subject:  "Feature Request: Dark Mode",
			From:     []domain.EmailParticipant{{Name: "Alex Rivera", Email: "alex@user.org"}},
			To:       []domain.EmailParticipant{{Name: "Info", Email: "info@demo-app.nylas.email"}},
			Date:     now.Add(-2 * 24 * time.Hour),
			Unread:   false,
			Starred:  false,
			Snippet:  "Would love to see dark mode support in the dashboard...",
			Body:     "Hello,\n\nI'm a happy user of your product but I work late hours and would really appreciate dark mode support.\n\nIs this on your roadmap?\n\nThanks,\nAlex",
			ThreadID: "inbound-thread-005",
		},
	}, nil
}

// Scheduler Demo Implementations

func (d *DemoClient) ListSchedulerConfigurations(ctx context.Context) ([]domain.SchedulerConfiguration, error) {
	return []domain.SchedulerConfiguration{
		{ID: "config-demo-1", Name: "30 Minute Meeting", Slug: "30min-demo"},
		{ID: "config-demo-2", Name: "1 Hour Meeting", Slug: "1hour-demo"},
	}, nil
}

func (d *DemoClient) GetSchedulerConfiguration(ctx context.Context, configID string) (*domain.SchedulerConfiguration, error) {
	return &domain.SchedulerConfiguration{
		ID:   configID,
		Name: "30 Minute Meeting",
		Slug: "30min-demo",
	}, nil
}

func (d *DemoClient) CreateSchedulerConfiguration(ctx context.Context, req *domain.CreateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error) {
	return &domain.SchedulerConfiguration{
		ID:   "config-demo-new",
		Name: req.Name,
		Slug: req.Slug,
	}, nil
}

func (d *DemoClient) UpdateSchedulerConfiguration(ctx context.Context, configID string, req *domain.UpdateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error) {
	name := "Updated Configuration"
	if req.Name != nil {
		name = *req.Name
	}
	return &domain.SchedulerConfiguration{
		ID:   configID,
		Name: name,
	}, nil
}

func (d *DemoClient) DeleteSchedulerConfiguration(ctx context.Context, configID string) error {
	return nil
}

func (d *DemoClient) CreateSchedulerSession(ctx context.Context, req *domain.CreateSchedulerSessionRequest) (*domain.SchedulerSession, error) {
	return &domain.SchedulerSession{
		SessionID:       "session-demo-123",
		ConfigurationID: req.ConfigurationID,
		BookingURL:      "https://schedule.nylas.com/demo-session",
	}, nil
}

func (d *DemoClient) GetSchedulerSession(ctx context.Context, sessionID string) (*domain.SchedulerSession, error) {
	return &domain.SchedulerSession{
		SessionID:       sessionID,
		ConfigurationID: "config-demo-1",
		BookingURL:      "https://schedule.nylas.com/demo-session",
	}, nil
}

func (d *DemoClient) ListBookings(ctx context.Context, configID string) ([]domain.Booking, error) {
	return []domain.Booking{
		{
			BookingID: "booking-demo-1",
			Title:     "Demo Meeting",
			Status:    "confirmed",
		},
	}, nil
}

func (d *DemoClient) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return &domain.Booking{
		BookingID: bookingID,
		Title:     "Demo Meeting",
		Status:    "confirmed",
	}, nil
}

func (d *DemoClient) ConfirmBooking(ctx context.Context, bookingID string, req *domain.ConfirmBookingRequest) (*domain.Booking, error) {
	return &domain.Booking{
		BookingID: bookingID,
		Status:    req.Status,
	}, nil
}

func (d *DemoClient) RescheduleBooking(ctx context.Context, bookingID string, req *domain.ConfirmBookingRequest) (*domain.Booking, error) {
	return &domain.Booking{
		BookingID: bookingID,
		Status:    "confirmed",
	}, nil
}

func (d *DemoClient) CancelBooking(ctx context.Context, bookingID string, reason string) error {
	return nil
}

func (d *DemoClient) ListSchedulerPages(ctx context.Context) ([]domain.SchedulerPage, error) {
	return []domain.SchedulerPage{
		{ID: "page-demo-1", Name: "Demo Booking Page", Slug: "book-demo"},
	}, nil
}

func (d *DemoClient) GetSchedulerPage(ctx context.Context, pageID string) (*domain.SchedulerPage, error) {
	return &domain.SchedulerPage{
		ID:   pageID,
		Name: "Demo Booking Page",
		Slug: "book-demo",
	}, nil
}

func (d *DemoClient) CreateSchedulerPage(ctx context.Context, req *domain.CreateSchedulerPageRequest) (*domain.SchedulerPage, error) {
	return &domain.SchedulerPage{
		ID:   "page-demo-new",
		Name: req.Name,
		Slug: req.Slug,
	}, nil
}

func (d *DemoClient) UpdateSchedulerPage(ctx context.Context, pageID string, req *domain.UpdateSchedulerPageRequest) (*domain.SchedulerPage, error) {
	name := "Updated Page"
	if req.Name != nil {
		name = *req.Name
	}
	return &domain.SchedulerPage{
		ID:   pageID,
		Name: name,
	}, nil
}

func (d *DemoClient) DeleteSchedulerPage(ctx context.Context, pageID string) error {
	return nil
}

// Admin Demo Implementations

func (d *DemoClient) ListApplications(ctx context.Context) ([]domain.Application, error) {
	return []domain.Application{
		{ID: "app-demo-1", ApplicationID: "app-id-demo-1", Region: "us"},
	}, nil
}

func (d *DemoClient) GetApplication(ctx context.Context, appID string) (*domain.Application, error) {
	return &domain.Application{
		ID:            appID,
		ApplicationID: appID,
		Region:        "us",
	}, nil
}

func (d *DemoClient) CreateApplication(ctx context.Context, req *domain.CreateApplicationRequest) (*domain.Application, error) {
	return &domain.Application{
		ID:            "app-demo-new",
		ApplicationID: "app-id-demo-new",
		Region:        req.Region,
	}, nil
}

func (d *DemoClient) UpdateApplication(ctx context.Context, appID string, req *domain.UpdateApplicationRequest) (*domain.Application, error) {
	return &domain.Application{
		ID:            appID,
		ApplicationID: appID,
		Region:        "us",
	}, nil
}

func (d *DemoClient) DeleteApplication(ctx context.Context, appID string) error {
	return nil
}

func (d *DemoClient) ListConnectors(ctx context.Context) ([]domain.Connector, error) {
	return []domain.Connector{
		{ID: "conn-demo-1", Name: "Google Demo Connector", Provider: "google"},
		{ID: "conn-demo-2", Name: "Microsoft Demo Connector", Provider: "microsoft"},
	}, nil
}

func (d *DemoClient) GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error) {
	return &domain.Connector{
		ID:       connectorID,
		Name:     "Google Demo Connector",
		Provider: "google",
	}, nil
}

func (d *DemoClient) CreateConnector(ctx context.Context, req *domain.CreateConnectorRequest) (*domain.Connector, error) {
	return &domain.Connector{
		ID:       "conn-demo-new",
		Name:     req.Name,
		Provider: req.Provider,
	}, nil
}

func (d *DemoClient) UpdateConnector(ctx context.Context, connectorID string, req *domain.UpdateConnectorRequest) (*domain.Connector, error) {
	name := "Updated Connector"
	if req.Name != nil {
		name = *req.Name
	}
	return &domain.Connector{
		ID:       connectorID,
		Name:     name,
		Provider: "google",
	}, nil
}

func (d *DemoClient) DeleteConnector(ctx context.Context, connectorID string) error {
	return nil
}

func (d *DemoClient) ListCredentials(ctx context.Context, connectorID string) ([]domain.ConnectorCredential, error) {
	return []domain.ConnectorCredential{
		{ID: "cred-demo-1", Name: "OAuth Demo Credential", CredentialType: "oauth"},
	}, nil
}

func (d *DemoClient) GetCredential(ctx context.Context, credentialID string) (*domain.ConnectorCredential, error) {
	return &domain.ConnectorCredential{
		ID:             credentialID,
		Name:           "OAuth Demo Credential",
		CredentialType: "oauth",
	}, nil
}

func (d *DemoClient) CreateCredential(ctx context.Context, connectorID string, req *domain.CreateCredentialRequest) (*domain.ConnectorCredential, error) {
	return &domain.ConnectorCredential{
		ID:             "cred-demo-new",
		Name:           req.Name,
		CredentialType: req.CredentialType,
	}, nil
}

func (d *DemoClient) UpdateCredential(ctx context.Context, credentialID string, req *domain.UpdateCredentialRequest) (*domain.ConnectorCredential, error) {
	name := "Updated Credential"
	if req.Name != nil {
		name = *req.Name
	}
	return &domain.ConnectorCredential{
		ID:             credentialID,
		Name:           name,
		CredentialType: "oauth",
	}, nil
}

func (d *DemoClient) DeleteCredential(ctx context.Context, credentialID string) error {
	return nil
}

func (d *DemoClient) ListAllGrants(ctx context.Context, params *domain.GrantsQueryParams) ([]domain.Grant, error) {
	return []domain.Grant{
		{
			ID:          "grant-demo-1",
			Provider:    "google",
			Email:       "demo1@example.com",
			GrantStatus: "valid",
		},
		{
			ID:          "grant-demo-2",
			Provider:    "microsoft",
			Email:       "demo2@example.com",
			GrantStatus: "valid",
		},
	}, nil
}

func (d *DemoClient) GetGrantStats(ctx context.Context) (*domain.GrantStats, error) {
	return &domain.GrantStats{
		Total:      10,
		Valid:      8,
		Invalid:    2,
		ByProvider: map[string]int{"google": 6, "microsoft": 4},
		ByStatus:   map[string]int{"valid": 8, "invalid": 2},
	}, nil
}

// Virtual Calendar Demo Implementations

func (d *DemoClient) CreateVirtualCalendarGrant(ctx context.Context, email string) (*domain.VirtualCalendarGrant, error) {
	return &domain.VirtualCalendarGrant{
		ID:          "vcal-demo-1",
		Provider:    "virtual-calendar",
		Email:       email,
		GrantStatus: "valid",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
	}, nil
}

func (d *DemoClient) ListVirtualCalendarGrants(ctx context.Context) ([]domain.VirtualCalendarGrant, error) {
	return []domain.VirtualCalendarGrant{
		{
			ID:          "vcal-demo-1",
			Provider:    "virtual-calendar",
			Email:       "conference-room-a@demo.com",
			GrantStatus: "valid",
			CreatedAt:   1704067200,
			UpdatedAt:   1704067200,
		},
		{
			ID:          "vcal-demo-2",
			Provider:    "virtual-calendar",
			Email:       "conference-room-b@demo.com",
			GrantStatus: "valid",
			CreatedAt:   1704153600,
			UpdatedAt:   1704153600,
		},
		{
			ID:          "vcal-demo-3",
			Provider:    "virtual-calendar",
			Email:       "resource-projector@demo.com",
			GrantStatus: "valid",
			CreatedAt:   1704240000,
			UpdatedAt:   1704240000,
		},
	}, nil
}

func (d *DemoClient) GetVirtualCalendarGrant(ctx context.Context, grantID string) (*domain.VirtualCalendarGrant, error) {
	return &domain.VirtualCalendarGrant{
		ID:          grantID,
		Provider:    "virtual-calendar",
		Email:       "conference-room-a@demo.com",
		GrantStatus: "valid",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
	}, nil
}

func (d *DemoClient) DeleteVirtualCalendarGrant(ctx context.Context, grantID string) error {
	return nil
}

// Recurring Event Demo Implementations

func (d *DemoClient) GetRecurringEventInstances(ctx context.Context, grantID, calendarID, masterEventID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	now := time.Now()
	return []domain.Event{
		{
			ID:            "event-demo-instance-1",
			GrantID:       grantID,
			CalendarID:    calendarID,
			Title:         "Weekly Demo Team Meeting",
			Description:   "Recurring team sync - Instance 1",
			MasterEventID: masterEventID,
			When: domain.EventWhen{
				StartTime: now.Unix(),
				EndTime:   now.Add(1 * time.Hour).Unix(),
				Object:    "timespan",
			},
			Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"},
			Status:     "confirmed",
			Busy:       true,
		},
		{
			ID:            "event-demo-instance-2",
			GrantID:       grantID,
			CalendarID:    calendarID,
			Title:         "Weekly Demo Team Meeting",
			Description:   "Recurring team sync - Instance 2",
			MasterEventID: masterEventID,
			When: domain.EventWhen{
				StartTime: now.Add(7 * 24 * time.Hour).Unix(),
				EndTime:   now.Add(7*24*time.Hour + 1*time.Hour).Unix(),
				Object:    "timespan",
			},
			Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"},
			Status:     "confirmed",
			Busy:       true,
		},
		{
			ID:            "event-demo-instance-3",
			GrantID:       grantID,
			CalendarID:    calendarID,
			Title:         "Weekly Demo Team Meeting",
			Description:   "Recurring team sync - Instance 3",
			MasterEventID: masterEventID,
			When: domain.EventWhen{
				StartTime: now.Add(14 * 24 * time.Hour).Unix(),
				EndTime:   now.Add(14*24*time.Hour + 1*time.Hour).Unix(),
				Object:    "timespan",
			},
			Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"},
			Status:     "confirmed",
			Busy:       true,
		},
	}, nil
}

func (d *DemoClient) UpdateRecurringEventInstance(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {
	now := time.Now()
	title := "Updated Demo Recurring Event Instance"
	if req.Title != nil {
		title = *req.Title
	}
	return &domain.Event{
		ID:         eventID,
		GrantID:    grantID,
		CalendarID: calendarID,
		Title:      title,
		When: domain.EventWhen{
			StartTime: now.Unix(),
			EndTime:   now.Add(1 * time.Hour).Unix(),
			Object:    "timespan",
		},
		Status: "confirmed",
		Busy:   true,
	}, nil
}

func (d *DemoClient) DeleteRecurringEventInstance(ctx context.Context, grantID, calendarID, eventID string) error {
	return nil
}
