package nylas

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
)

// MockClient is a mock implementation of NylasClient for testing.
type MockClient struct {
	// State
	Region       string
	ClientID     string
	ClientSecret string
	APIKey       string

	// Call tracking
	ExchangeCodeCalled      bool
	ListGrantsCalled        bool
	GetGrantCalled          bool
	RevokeGrantCalled       bool
	GetMessagesCalled       bool
	GetMessagesWithParamsCalled bool
	GetMessageCalled        bool
	SendMessageCalled       bool
	UpdateMessageCalled     bool
	DeleteMessageCalled     bool
	GetThreadsCalled        bool
	GetThreadCalled         bool
	UpdateThreadCalled      bool
	DeleteThreadCalled      bool
	GetDraftsCalled         bool
	GetDraftCalled          bool
	CreateDraftCalled       bool
	UpdateDraftCalled       bool
	DeleteDraftCalled       bool
	SendDraftCalled         bool
	GetFoldersCalled        bool
	GetFolderCalled         bool
	CreateFolderCalled      bool
	UpdateFolderCalled      bool
	DeleteFolderCalled      bool
	ListAttachmentsCalled    bool
	GetAttachmentCalled      bool
	DownloadAttachmentCalled bool
	LastGrantID              string
	LastMessageID           string
	LastThreadID            string
	LastDraftID             string
	LastFolderID            string
	LastAttachmentID        string

	// Custom functions
	ExchangeCodeFunc          func(ctx context.Context, code, redirectURI string) (*domain.Grant, error)
	ListGrantsFunc            func(ctx context.Context) ([]domain.Grant, error)
	GetGrantFunc              func(ctx context.Context, grantID string) (*domain.Grant, error)
	RevokeGrantFunc           func(ctx context.Context, grantID string) error
	GetMessagesFunc           func(ctx context.Context, grantID string, limit int) ([]domain.Message, error)
	GetMessagesWithParamsFunc func(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error)
	GetMessageFunc            func(ctx context.Context, grantID, messageID string) (*domain.Message, error)
	SendMessageFunc           func(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error)
	UpdateMessageFunc         func(ctx context.Context, grantID, messageID string, req *domain.UpdateMessageRequest) (*domain.Message, error)
	DeleteMessageFunc         func(ctx context.Context, grantID, messageID string) error
	GetThreadsFunc            func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error)
	GetThreadFunc             func(ctx context.Context, grantID, threadID string) (*domain.Thread, error)
	UpdateThreadFunc          func(ctx context.Context, grantID, threadID string, req *domain.UpdateMessageRequest) (*domain.Thread, error)
	DeleteThreadFunc          func(ctx context.Context, grantID, threadID string) error
	GetDraftsFunc             func(ctx context.Context, grantID string, limit int) ([]domain.Draft, error)
	GetDraftFunc              func(ctx context.Context, grantID, draftID string) (*domain.Draft, error)
	CreateDraftFunc           func(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error)
	UpdateDraftFunc           func(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error)
	DeleteDraftFunc           func(ctx context.Context, grantID, draftID string) error
	SendDraftFunc             func(ctx context.Context, grantID, draftID string) (*domain.Message, error)
	GetFoldersFunc            func(ctx context.Context, grantID string) ([]domain.Folder, error)
	GetFolderFunc             func(ctx context.Context, grantID, folderID string) (*domain.Folder, error)
	CreateFolderFunc          func(ctx context.Context, grantID string, req *domain.CreateFolderRequest) (*domain.Folder, error)
	UpdateFolderFunc          func(ctx context.Context, grantID, folderID string, req *domain.UpdateFolderRequest) (*domain.Folder, error)
	DeleteFolderFunc          func(ctx context.Context, grantID, folderID string) error
	ListAttachmentsFunc       func(ctx context.Context, grantID, messageID string) ([]domain.Attachment, error)
	GetAttachmentFunc         func(ctx context.Context, grantID, messageID, attachmentID string) (*domain.Attachment, error)
	DownloadAttachmentFunc    func(ctx context.Context, grantID, messageID, attachmentID string) (io.ReadCloser, error)
}

// NewMockClient creates a new MockClient.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// SetRegion sets the API region.
func (m *MockClient) SetRegion(region string) {
	m.Region = region
}

// SetCredentials sets the API credentials.
func (m *MockClient) SetCredentials(clientID, clientSecret, apiKey string) {
	m.ClientID = clientID
	m.ClientSecret = clientSecret
	m.APIKey = apiKey
}

// BuildAuthURL builds the OAuth authorization URL.
func (m *MockClient) BuildAuthURL(provider domain.Provider, redirectURI string) string {
	return "https://mock.nylas.com/auth?provider=" + string(provider)
}

// ExchangeCode exchanges an authorization code for tokens.
func (m *MockClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*domain.Grant, error) {
	m.ExchangeCodeCalled = true
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(ctx, code, redirectURI)
	}
	return &domain.Grant{
		ID:          "mock-grant-id",
		Email:       "test@example.com",
		Provider:    domain.ProviderGoogle,
		GrantStatus: "valid",
	}, nil
}

// ListGrants lists all grants.
func (m *MockClient) ListGrants(ctx context.Context) ([]domain.Grant, error) {
	m.ListGrantsCalled = true
	if m.ListGrantsFunc != nil {
		return m.ListGrantsFunc(ctx)
	}
	return []domain.Grant{}, nil
}

// GetGrant retrieves a specific grant.
func (m *MockClient) GetGrant(ctx context.Context, grantID string) (*domain.Grant, error) {
	m.GetGrantCalled = true
	m.LastGrantID = grantID
	if m.GetGrantFunc != nil {
		return m.GetGrantFunc(ctx, grantID)
	}
	return &domain.Grant{
		ID:          grantID,
		Email:       "test@example.com",
		Provider:    domain.ProviderGoogle,
		GrantStatus: "valid",
	}, nil
}

// RevokeGrant revokes a grant.
func (m *MockClient) RevokeGrant(ctx context.Context, grantID string) error {
	m.RevokeGrantCalled = true
	m.LastGrantID = grantID
	if m.RevokeGrantFunc != nil {
		return m.RevokeGrantFunc(ctx, grantID)
	}
	return nil
}

// GetMessages retrieves recent messages.
func (m *MockClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
	m.GetMessagesCalled = true
	m.LastGrantID = grantID
	if m.GetMessagesFunc != nil {
		return m.GetMessagesFunc(ctx, grantID, limit)
	}
	return []domain.Message{}, nil
}

// GetMessagesWithParams retrieves messages with query parameters.
func (m *MockClient) GetMessagesWithParams(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
	m.GetMessagesWithParamsCalled = true
	m.LastGrantID = grantID
	if m.GetMessagesWithParamsFunc != nil {
		return m.GetMessagesWithParamsFunc(ctx, grantID, params)
	}
	return []domain.Message{}, nil
}

// GetMessagesWithCursor retrieves messages with pagination cursor support.
func (m *MockClient) GetMessagesWithCursor(ctx context.Context, grantID string, params *domain.MessageQueryParams) (*domain.MessageListResponse, error) {
	m.GetMessagesWithParamsCalled = true
	m.LastGrantID = grantID
	if m.GetMessagesWithParamsFunc != nil {
		msgs, err := m.GetMessagesWithParamsFunc(ctx, grantID, params)
		return &domain.MessageListResponse{Data: msgs}, err
	}
	return &domain.MessageListResponse{Data: []domain.Message{}}, nil
}

// GetMessage retrieves a single message.
func (m *MockClient) GetMessage(ctx context.Context, grantID, messageID string) (*domain.Message, error) {
	m.GetMessageCalled = true
	m.LastGrantID = grantID
	m.LastMessageID = messageID
	if m.GetMessageFunc != nil {
		return m.GetMessageFunc(ctx, grantID, messageID)
	}
	return &domain.Message{
		ID:      messageID,
		GrantID: grantID,
		Subject: "Test Message",
		From:    []domain.EmailParticipant{{Email: "sender@example.com"}},
		Body:    "Test body",
	}, nil
}

// SendMessage sends an email.
func (m *MockClient) SendMessage(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error) {
	m.SendMessageCalled = true
	m.LastGrantID = grantID
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, grantID, req)
	}
	return &domain.Message{
		ID:      "sent-message-id",
		GrantID: grantID,
		Subject: req.Subject,
		To:      req.To,
		Body:    req.Body,
	}, nil
}

// UpdateMessage updates message properties.
func (m *MockClient) UpdateMessage(ctx context.Context, grantID, messageID string, req *domain.UpdateMessageRequest) (*domain.Message, error) {
	m.UpdateMessageCalled = true
	m.LastGrantID = grantID
	m.LastMessageID = messageID
	if m.UpdateMessageFunc != nil {
		return m.UpdateMessageFunc(ctx, grantID, messageID, req)
	}
	msg := &domain.Message{
		ID:      messageID,
		GrantID: grantID,
		Subject: "Updated Message",
	}
	if req.Unread != nil {
		msg.Unread = *req.Unread
	}
	if req.Starred != nil {
		msg.Starred = *req.Starred
	}
	return msg, nil
}

// DeleteMessage deletes a message.
func (m *MockClient) DeleteMessage(ctx context.Context, grantID, messageID string) error {
	m.DeleteMessageCalled = true
	m.LastGrantID = grantID
	m.LastMessageID = messageID
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, grantID, messageID)
	}
	return nil
}

// ListScheduledMessages retrieves scheduled messages.
func (m *MockClient) ListScheduledMessages(ctx context.Context, grantID string) ([]domain.ScheduledMessage, error) {
	m.LastGrantID = grantID
	return []domain.ScheduledMessage{
		{ScheduleID: "schedule-1", Status: "pending", CloseTime: 1700000000},
		{ScheduleID: "schedule-2", Status: "scheduled", CloseTime: 1700100000},
	}, nil
}

// GetScheduledMessage retrieves a specific scheduled message.
func (m *MockClient) GetScheduledMessage(ctx context.Context, grantID, scheduleID string) (*domain.ScheduledMessage, error) {
	m.LastGrantID = grantID
	return &domain.ScheduledMessage{
		ScheduleID: scheduleID,
		Status:     "pending",
		CloseTime:  1700000000,
	}, nil
}

// CancelScheduledMessage cancels a scheduled message.
func (m *MockClient) CancelScheduledMessage(ctx context.Context, grantID, scheduleID string) error {
	m.LastGrantID = grantID
	return nil
}

// GetThreads retrieves threads.
func (m *MockClient) GetThreads(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
	m.GetThreadsCalled = true
	m.LastGrantID = grantID
	if m.GetThreadsFunc != nil {
		return m.GetThreadsFunc(ctx, grantID, params)
	}
	return []domain.Thread{}, nil
}

// GetThread retrieves a single thread.
func (m *MockClient) GetThread(ctx context.Context, grantID, threadID string) (*domain.Thread, error) {
	m.GetThreadCalled = true
	m.LastGrantID = grantID
	m.LastThreadID = threadID
	if m.GetThreadFunc != nil {
		return m.GetThreadFunc(ctx, grantID, threadID)
	}
	return &domain.Thread{
		ID:      threadID,
		GrantID: grantID,
		Subject: "Test Thread",
	}, nil
}

// UpdateThread updates thread properties.
func (m *MockClient) UpdateThread(ctx context.Context, grantID, threadID string, req *domain.UpdateMessageRequest) (*domain.Thread, error) {
	m.UpdateThreadCalled = true
	m.LastGrantID = grantID
	m.LastThreadID = threadID
	if m.UpdateThreadFunc != nil {
		return m.UpdateThreadFunc(ctx, grantID, threadID, req)
	}
	thread := &domain.Thread{
		ID:      threadID,
		GrantID: grantID,
		Subject: "Updated Thread",
	}
	if req.Unread != nil {
		thread.Unread = *req.Unread
	}
	if req.Starred != nil {
		thread.Starred = *req.Starred
	}
	return thread, nil
}

// DeleteThread deletes a thread.
func (m *MockClient) DeleteThread(ctx context.Context, grantID, threadID string) error {
	m.DeleteThreadCalled = true
	m.LastGrantID = grantID
	m.LastThreadID = threadID
	if m.DeleteThreadFunc != nil {
		return m.DeleteThreadFunc(ctx, grantID, threadID)
	}
	return nil
}

// GetDrafts retrieves drafts.
func (m *MockClient) GetDrafts(ctx context.Context, grantID string, limit int) ([]domain.Draft, error) {
	m.GetDraftsCalled = true
	m.LastGrantID = grantID
	if m.GetDraftsFunc != nil {
		return m.GetDraftsFunc(ctx, grantID, limit)
	}
	return []domain.Draft{}, nil
}

// GetDraft retrieves a single draft.
func (m *MockClient) GetDraft(ctx context.Context, grantID, draftID string) (*domain.Draft, error) {
	m.GetDraftCalled = true
	m.LastGrantID = grantID
	m.LastDraftID = draftID
	if m.GetDraftFunc != nil {
		return m.GetDraftFunc(ctx, grantID, draftID)
	}
	return &domain.Draft{
		ID:      draftID,
		GrantID: grantID,
		Subject: "Test Draft",
	}, nil
}

// CreateDraft creates a new draft.
func (m *MockClient) CreateDraft(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	m.CreateDraftCalled = true
	m.LastGrantID = grantID
	if m.CreateDraftFunc != nil {
		return m.CreateDraftFunc(ctx, grantID, req)
	}

	// Convert request attachments to response attachments (with generated IDs)
	var attachments []domain.Attachment
	for i, a := range req.Attachments {
		attachments = append(attachments, domain.Attachment{
			ID:          fmt.Sprintf("attach-%d", i+1),
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
		})
	}

	return &domain.Draft{
		ID:          "new-draft-id",
		GrantID:     grantID,
		Subject:     req.Subject,
		Body:        req.Body,
		To:          req.To,
		Attachments: attachments,
	}, nil
}

// UpdateDraft updates an existing draft.
func (m *MockClient) UpdateDraft(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	m.UpdateDraftCalled = true
	m.LastGrantID = grantID
	m.LastDraftID = draftID
	if m.UpdateDraftFunc != nil {
		return m.UpdateDraftFunc(ctx, grantID, draftID, req)
	}

	// Convert request attachments to response attachments
	var attachments []domain.Attachment
	for i, a := range req.Attachments {
		attachments = append(attachments, domain.Attachment{
			ID:          fmt.Sprintf("attach-%d", i+1),
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
		})
	}

	return &domain.Draft{
		ID:          draftID,
		GrantID:     grantID,
		Subject:     req.Subject,
		Body:        req.Body,
		To:          req.To,
		Attachments: attachments,
	}, nil
}

// DeleteDraft deletes a draft.
func (m *MockClient) DeleteDraft(ctx context.Context, grantID, draftID string) error {
	m.DeleteDraftCalled = true
	m.LastGrantID = grantID
	m.LastDraftID = draftID
	if m.DeleteDraftFunc != nil {
		return m.DeleteDraftFunc(ctx, grantID, draftID)
	}
	return nil
}

// SendDraft sends a draft.
func (m *MockClient) SendDraft(ctx context.Context, grantID, draftID string) (*domain.Message, error) {
	m.SendDraftCalled = true
	m.LastGrantID = grantID
	m.LastDraftID = draftID
	if m.SendDraftFunc != nil {
		return m.SendDraftFunc(ctx, grantID, draftID)
	}
	return &domain.Message{
		ID:      "sent-from-draft-id",
		GrantID: grantID,
		Subject: "Sent Draft",
	}, nil
}

// GetFolders retrieves all folders.
func (m *MockClient) GetFolders(ctx context.Context, grantID string) ([]domain.Folder, error) {
	m.GetFoldersCalled = true
	m.LastGrantID = grantID
	if m.GetFoldersFunc != nil {
		return m.GetFoldersFunc(ctx, grantID)
	}
	return []domain.Folder{
		{ID: "inbox", Name: "Inbox", SystemFolder: "inbox"},
		{ID: "sent", Name: "Sent", SystemFolder: "sent"},
		{ID: "drafts", Name: "Drafts", SystemFolder: "drafts"},
	}, nil
}

// GetFolder retrieves a single folder.
func (m *MockClient) GetFolder(ctx context.Context, grantID, folderID string) (*domain.Folder, error) {
	m.GetFolderCalled = true
	m.LastGrantID = grantID
	m.LastFolderID = folderID
	if m.GetFolderFunc != nil {
		return m.GetFolderFunc(ctx, grantID, folderID)
	}
	return &domain.Folder{
		ID:      folderID,
		GrantID: grantID,
		Name:    "Test Folder",
	}, nil
}

// CreateFolder creates a new folder.
func (m *MockClient) CreateFolder(ctx context.Context, grantID string, req *domain.CreateFolderRequest) (*domain.Folder, error) {
	m.CreateFolderCalled = true
	m.LastGrantID = grantID
	if m.CreateFolderFunc != nil {
		return m.CreateFolderFunc(ctx, grantID, req)
	}
	return &domain.Folder{
		ID:      "new-folder-id",
		GrantID: grantID,
		Name:    req.Name,
	}, nil
}

// UpdateFolder updates an existing folder.
func (m *MockClient) UpdateFolder(ctx context.Context, grantID, folderID string, req *domain.UpdateFolderRequest) (*domain.Folder, error) {
	m.UpdateFolderCalled = true
	m.LastGrantID = grantID
	m.LastFolderID = folderID
	if m.UpdateFolderFunc != nil {
		return m.UpdateFolderFunc(ctx, grantID, folderID, req)
	}
	return &domain.Folder{
		ID:      folderID,
		GrantID: grantID,
		Name:    req.Name,
	}, nil
}

// DeleteFolder deletes a folder.
func (m *MockClient) DeleteFolder(ctx context.Context, grantID, folderID string) error {
	m.DeleteFolderCalled = true
	m.LastGrantID = grantID
	m.LastFolderID = folderID
	if m.DeleteFolderFunc != nil {
		return m.DeleteFolderFunc(ctx, grantID, folderID)
	}
	return nil
}

// ListAttachments retrieves all attachments for a message.
func (m *MockClient) ListAttachments(ctx context.Context, grantID, messageID string) ([]domain.Attachment, error) {
	m.ListAttachmentsCalled = true
	m.LastGrantID = grantID
	m.LastMessageID = messageID
	if m.ListAttachmentsFunc != nil {
		return m.ListAttachmentsFunc(ctx, grantID, messageID)
	}
	return []domain.Attachment{
		{
			ID:          "attach-1",
			GrantID:     grantID,
			Filename:    "test.pdf",
			ContentType: "application/pdf",
			Size:        1024,
		},
		{
			ID:          "attach-2",
			GrantID:     grantID,
			Filename:    "image.png",
			ContentType: "image/png",
			Size:        2048,
		},
	}, nil
}

// GetAttachment retrieves attachment metadata.
func (m *MockClient) GetAttachment(ctx context.Context, grantID, messageID, attachmentID string) (*domain.Attachment, error) {
	m.GetAttachmentCalled = true
	m.LastGrantID = grantID
	m.LastAttachmentID = attachmentID
	if m.GetAttachmentFunc != nil {
		return m.GetAttachmentFunc(ctx, grantID, messageID, attachmentID)
	}
	return &domain.Attachment{
		ID:          attachmentID,
		GrantID:     grantID,
		Filename:    "test.pdf",
		ContentType: "application/pdf",
		Size:        1024,
	}, nil
}

// DownloadAttachment downloads attachment content.
func (m *MockClient) DownloadAttachment(ctx context.Context, grantID, messageID, attachmentID string) (io.ReadCloser, error) {
	m.DownloadAttachmentCalled = true
	m.LastGrantID = grantID
	m.LastAttachmentID = attachmentID
	if m.DownloadAttachmentFunc != nil {
		return m.DownloadAttachmentFunc(ctx, grantID, messageID, attachmentID)
	}
	return io.NopCloser(strings.NewReader("mock attachment content")), nil
}

// GetCalendars retrieves all calendars.
func (m *MockClient) GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error) {
	return []domain.Calendar{
		{ID: "primary", Name: "Primary Calendar", IsPrimary: true},
	}, nil
}

// GetCalendar retrieves a single calendar.
func (m *MockClient) GetCalendar(ctx context.Context, grantID, calendarID string) (*domain.Calendar, error) {
	return &domain.Calendar{
		ID:        calendarID,
		Name:      "Test Calendar",
		IsPrimary: calendarID == "primary",
	}, nil
}

// CreateCalendar creates a new calendar.
func (m *MockClient) CreateCalendar(ctx context.Context, grantID string, req *domain.CreateCalendarRequest) (*domain.Calendar, error) {
	return &domain.Calendar{
		ID:          "new-calendar-id",
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Timezone:    req.Timezone,
	}, nil
}

// UpdateCalendar updates an existing calendar.
func (m *MockClient) UpdateCalendar(ctx context.Context, grantID, calendarID string, req *domain.UpdateCalendarRequest) (*domain.Calendar, error) {
	cal := &domain.Calendar{ID: calendarID}
	if req.Name != nil {
		cal.Name = *req.Name
	}
	if req.Description != nil {
		cal.Description = *req.Description
	}
	if req.Location != nil {
		cal.Location = *req.Location
	}
	if req.Timezone != nil {
		cal.Timezone = *req.Timezone
	}
	if req.HexColor != nil {
		cal.HexColor = *req.HexColor
	}
	return cal, nil
}

// DeleteCalendar deletes a calendar.
func (m *MockClient) DeleteCalendar(ctx context.Context, grantID, calendarID string) error {
	return nil
}

// GetEvents retrieves events.
func (m *MockClient) GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	return []domain.Event{}, nil
}

// GetEventsWithCursor retrieves events with pagination.
func (m *MockClient) GetEventsWithCursor(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) (*domain.EventListResponse, error) {
	return &domain.EventListResponse{Data: []domain.Event{}}, nil
}

// GetEvent retrieves a single event.
func (m *MockClient) GetEvent(ctx context.Context, grantID, calendarID, eventID string) (*domain.Event, error) {
	return &domain.Event{
		ID:         eventID,
		CalendarID: calendarID,
		Title:      "Test Event",
	}, nil
}

// CreateEvent creates a new event.
func (m *MockClient) CreateEvent(ctx context.Context, grantID, calendarID string, req *domain.CreateEventRequest) (*domain.Event, error) {
	return &domain.Event{
		ID:         "new-event-id",
		CalendarID: calendarID,
		Title:      req.Title,
	}, nil
}

// UpdateEvent updates an existing event.
func (m *MockClient) UpdateEvent(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {
	event := &domain.Event{
		ID:         eventID,
		CalendarID: calendarID,
	}
	if req.Title != nil {
		event.Title = *req.Title
	}
	return event, nil
}

// DeleteEvent deletes an event.
func (m *MockClient) DeleteEvent(ctx context.Context, grantID, calendarID, eventID string) error {
	return nil
}

// SendRSVP sends an RSVP response to an event invitation.
func (m *MockClient) SendRSVP(ctx context.Context, grantID, calendarID, eventID string, req *domain.SendRSVPRequest) error {
	return nil
}

// GetFreeBusy retrieves free/busy information.
func (m *MockClient) GetFreeBusy(ctx context.Context, grantID string, req *domain.FreeBusyRequest) (*domain.FreeBusyResponse, error) {
	now := req.StartTime
	result := &domain.FreeBusyResponse{
		Data: make([]domain.FreeBusyCalendar, len(req.Emails)),
	}
	for i, email := range req.Emails {
		result.Data[i] = domain.FreeBusyCalendar{
			Email: email,
			TimeSlots: []domain.TimeSlot{
				{
					StartTime: now + 3600,  // 1 hour from start
					EndTime:   now + 7200,  // 2 hours from start
					Status:    "busy",
				},
			},
		}
	}
	return result, nil
}

// GetAvailability finds available meeting times.
func (m *MockClient) GetAvailability(ctx context.Context, req *domain.AvailabilityRequest) (*domain.AvailabilityResponse, error) {
	duration := int64(req.DurationMinutes * 60)
	result := &domain.AvailabilityResponse{
		Data: domain.AvailabilityData{
			TimeSlots: []domain.AvailableSlot{
				{
					StartTime: req.StartTime + 7200,
					EndTime:   req.StartTime + 7200 + duration,
				},
				{
					StartTime: req.StartTime + 14400,
					EndTime:   req.StartTime + 14400 + duration,
				},
			},
		},
	}
	return result, nil
}

// GetContacts retrieves contacts.
func (m *MockClient) GetContacts(ctx context.Context, grantID string, params *domain.ContactQueryParams) ([]domain.Contact, error) {
	return []domain.Contact{
		{
			ID:        "contact-1",
			GivenName: "John",
			Surname:   "Doe",
			Emails:    []domain.ContactEmail{{Email: "john@example.com", Type: "work"}},
		},
	}, nil
}

// GetContactsWithCursor retrieves contacts with pagination.
func (m *MockClient) GetContactsWithCursor(ctx context.Context, grantID string, params *domain.ContactQueryParams) (*domain.ContactListResponse, error) {
	return &domain.ContactListResponse{
		Data: []domain.Contact{
			{
				ID:        "contact-1",
				GivenName: "John",
				Surname:   "Doe",
				Emails:    []domain.ContactEmail{{Email: "john@example.com", Type: "work"}},
			},
		},
	}, nil
}

// GetContact retrieves a single contact.
func (m *MockClient) GetContact(ctx context.Context, grantID, contactID string) (*domain.Contact, error) {
	return &domain.Contact{
		ID:        contactID,
		GivenName: "John",
		Surname:   "Doe",
		Emails:    []domain.ContactEmail{{Email: "john@example.com", Type: "work"}},
	}, nil
}

// CreateContact creates a new contact.
func (m *MockClient) CreateContact(ctx context.Context, grantID string, req *domain.CreateContactRequest) (*domain.Contact, error) {
	return &domain.Contact{
		ID:        "new-contact-id",
		GivenName: req.GivenName,
		Surname:   req.Surname,
		Emails:    req.Emails,
	}, nil
}

// UpdateContact updates an existing contact.
func (m *MockClient) UpdateContact(ctx context.Context, grantID, contactID string, req *domain.UpdateContactRequest) (*domain.Contact, error) {
	contact := &domain.Contact{ID: contactID}
	if req.GivenName != nil {
		contact.GivenName = *req.GivenName
	}
	if req.Surname != nil {
		contact.Surname = *req.Surname
	}
	contact.Emails = req.Emails
	return contact, nil
}

// DeleteContact deletes a contact.
func (m *MockClient) DeleteContact(ctx context.Context, grantID, contactID string) error {
	return nil
}

// GetContactGroups retrieves contact groups.
func (m *MockClient) GetContactGroups(ctx context.Context, grantID string) ([]domain.ContactGroup, error) {
	return []domain.ContactGroup{
		{ID: "group-1", Name: "Contacts"},
	}, nil
}

// GetContactGroup retrieves a single contact group.
func (m *MockClient) GetContactGroup(ctx context.Context, grantID, groupID string) (*domain.ContactGroup, error) {
	return &domain.ContactGroup{
		ID:      groupID,
		GrantID: grantID,
		Name:    "Test Group",
	}, nil
}

// CreateContactGroup creates a new contact group.
func (m *MockClient) CreateContactGroup(ctx context.Context, grantID string, req *domain.CreateContactGroupRequest) (*domain.ContactGroup, error) {
	return &domain.ContactGroup{
		ID:      "new-group-id",
		GrantID: grantID,
		Name:    req.Name,
	}, nil
}

// UpdateContactGroup updates an existing contact group.
func (m *MockClient) UpdateContactGroup(ctx context.Context, grantID, groupID string, req *domain.UpdateContactGroupRequest) (*domain.ContactGroup, error) {
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

// DeleteContactGroup deletes a contact group.
func (m *MockClient) DeleteContactGroup(ctx context.Context, grantID, groupID string) error {
	return nil
}

// ListWebhooks lists all webhooks.
func (m *MockClient) ListWebhooks(ctx context.Context) ([]domain.Webhook, error) {
	return []domain.Webhook{
		{
			ID:           "webhook-1",
			Description:  "Test Webhook",
			TriggerTypes: []string{domain.TriggerMessageCreated},
			WebhookURL:   "https://example.com/webhook",
			Status:       "active",
		},
	}, nil
}

// GetWebhook retrieves a single webhook.
func (m *MockClient) GetWebhook(ctx context.Context, webhookID string) (*domain.Webhook, error) {
	return &domain.Webhook{
		ID:           webhookID,
		Description:  "Test Webhook",
		TriggerTypes: []string{domain.TriggerMessageCreated},
		WebhookURL:   "https://example.com/webhook",
		Status:       "active",
	}, nil
}

// CreateWebhook creates a new webhook.
func (m *MockClient) CreateWebhook(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	return &domain.Webhook{
		ID:            "new-webhook-id",
		Description:   req.Description,
		TriggerTypes:  req.TriggerTypes,
		WebhookURL:    req.WebhookURL,
		WebhookSecret: "mock-secret-12345",
		Status:        "active",
	}, nil
}

// UpdateWebhook updates an existing webhook.
func (m *MockClient) UpdateWebhook(ctx context.Context, webhookID string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	webhook := &domain.Webhook{
		ID:     webhookID,
		Status: "active",
	}
	if req.Description != "" {
		webhook.Description = req.Description
	}
	if req.WebhookURL != "" {
		webhook.WebhookURL = req.WebhookURL
	}
	if len(req.TriggerTypes) > 0 {
		webhook.TriggerTypes = req.TriggerTypes
	}
	if req.Status != "" {
		webhook.Status = req.Status
	}
	return webhook, nil
}

// DeleteWebhook deletes a webhook.
func (m *MockClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	return nil
}

// SendWebhookTestEvent sends a test event to a webhook URL.
func (m *MockClient) SendWebhookTestEvent(ctx context.Context, webhookURL string) error {
	return nil
}

// GetWebhookMockPayload returns a mock payload for a trigger type.
func (m *MockClient) GetWebhookMockPayload(ctx context.Context, triggerType string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"specversion": "1.0",
		"type":        triggerType,
		"source":      "/nylas/test",
		"id":          "mock-event-id",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id": "mock-object-id",
			},
		},
	}, nil
}

// ListNotetakers lists all notetakers for a grant.
func (m *MockClient) ListNotetakers(ctx context.Context, grantID string, params *domain.NotetakerQueryParams) ([]domain.Notetaker, error) {
	return []domain.Notetaker{
		{
			ID:           "notetaker-1",
			State:        domain.NotetakerStateComplete,
			MeetingLink:  "https://zoom.us/j/123456789",
			MeetingTitle: "Test Meeting",
		},
	}, nil
}

// GetNotetaker retrieves a single notetaker.
func (m *MockClient) GetNotetaker(ctx context.Context, grantID, notetakerID string) (*domain.Notetaker, error) {
	return &domain.Notetaker{
		ID:           notetakerID,
		State:        domain.NotetakerStateComplete,
		MeetingLink:  "https://zoom.us/j/123456789",
		MeetingTitle: "Test Meeting",
		MeetingInfo: &domain.MeetingInfo{
			Provider: "zoom",
		},
	}, nil
}

// CreateNotetaker creates a new notetaker.
func (m *MockClient) CreateNotetaker(ctx context.Context, grantID string, req *domain.CreateNotetakerRequest) (*domain.Notetaker, error) {
	return &domain.Notetaker{
		ID:          "new-notetaker-id",
		State:       domain.NotetakerStateScheduled,
		MeetingLink: req.MeetingLink,
		BotConfig:   req.BotConfig,
	}, nil
}

// DeleteNotetaker deletes a notetaker.
func (m *MockClient) DeleteNotetaker(ctx context.Context, grantID, notetakerID string) error {
	return nil
}

// GetNotetakerMedia retrieves notetaker media.
func (m *MockClient) GetNotetakerMedia(ctx context.Context, grantID, notetakerID string) (*domain.MediaData, error) {
	return &domain.MediaData{
		Recording: &domain.MediaFile{
			URL:         "https://storage.nylas.com/recording.mp4",
			ContentType: "video/mp4",
			Size:        1024000,
			ExpiresAt:   1700000000,
		},
		Transcript: &domain.MediaFile{
			URL:         "https://storage.nylas.com/transcript.txt",
			ContentType: "text/plain",
			Size:        4096,
			ExpiresAt:   1700000000,
		},
	}, nil
}

// ListInboundInboxes lists all inbound inboxes.
func (m *MockClient) ListInboundInboxes(ctx context.Context) ([]domain.InboundInbox, error) {
	return []domain.InboundInbox{
		{
			ID:          "inbox-1",
			Email:       "support@app.nylas.email",
			GrantStatus: "valid",
		},
		{
			ID:          "inbox-2",
			Email:       "info@app.nylas.email",
			GrantStatus: "valid",
		},
	}, nil
}

// GetInboundInbox retrieves a specific inbound inbox.
func (m *MockClient) GetInboundInbox(ctx context.Context, grantID string) (*domain.InboundInbox, error) {
	m.LastGrantID = grantID
	return &domain.InboundInbox{
		ID:          grantID,
		Email:       "support@app.nylas.email",
		GrantStatus: "valid",
	}, nil
}

// CreateInboundInbox creates a new inbound inbox.
func (m *MockClient) CreateInboundInbox(ctx context.Context, email string) (*domain.InboundInbox, error) {
	return &domain.InboundInbox{
		ID:          "new-inbox-id",
		Email:       email + "@app.nylas.email",
		GrantStatus: "valid",
	}, nil
}

// DeleteInboundInbox deletes an inbound inbox.
func (m *MockClient) DeleteInboundInbox(ctx context.Context, grantID string) error {
	m.LastGrantID = grantID
	return nil
}

// GetInboundMessages retrieves messages for an inbound inbox.
func (m *MockClient) GetInboundMessages(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.InboundMessage, error) {
	m.LastGrantID = grantID
	return []domain.InboundMessage{
		{
			ID:      "inbound-msg-1",
			GrantID: grantID,
			Subject: "New Lead Submission",
			From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
			Snippet: "Hi, I'm interested in your services...",
			Unread:  true,
		},
		{
			ID:      "inbound-msg-2",
			GrantID: grantID,
			Subject: "Support Request #12345",
			From:    []domain.EmailParticipant{{Name: "Jane Smith", Email: "jane@example.com"}},
			Snippet: "I need help with my account...",
			Unread:  false,
		},
	}, nil
}
