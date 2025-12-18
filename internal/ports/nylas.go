package ports

import (
	"context"
	"io"

	"github.com/mqasimca/nylas/internal/domain"
)

// NylasClient defines the interface for interacting with the Nylas API.
type NylasClient interface {
	// Auth operations
	BuildAuthURL(provider domain.Provider, redirectURI string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (*domain.Grant, error)
	ListGrants(ctx context.Context) ([]domain.Grant, error)
	GetGrant(ctx context.Context, grantID string) (*domain.Grant, error)
	RevokeGrant(ctx context.Context, grantID string) error

	// Message operations
	GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error)
	GetMessagesWithParams(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error)
	GetMessagesWithCursor(ctx context.Context, grantID string, params *domain.MessageQueryParams) (*domain.MessageListResponse, error)
	GetMessage(ctx context.Context, grantID, messageID string) (*domain.Message, error)
	SendMessage(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error)
	UpdateMessage(ctx context.Context, grantID, messageID string, req *domain.UpdateMessageRequest) (*domain.Message, error)
	DeleteMessage(ctx context.Context, grantID, messageID string) error

	// Scheduled message operations
	ListScheduledMessages(ctx context.Context, grantID string) ([]domain.ScheduledMessage, error)
	GetScheduledMessage(ctx context.Context, grantID, scheduleID string) (*domain.ScheduledMessage, error)
	CancelScheduledMessage(ctx context.Context, grantID, scheduleID string) error

	// Thread operations
	GetThreads(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error)
	GetThread(ctx context.Context, grantID, threadID string) (*domain.Thread, error)
	UpdateThread(ctx context.Context, grantID, threadID string, req *domain.UpdateMessageRequest) (*domain.Thread, error)
	DeleteThread(ctx context.Context, grantID, threadID string) error

	// Draft operations
	GetDrafts(ctx context.Context, grantID string, limit int) ([]domain.Draft, error)
	GetDraft(ctx context.Context, grantID, draftID string) (*domain.Draft, error)
	CreateDraft(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error)
	UpdateDraft(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error)
	DeleteDraft(ctx context.Context, grantID, draftID string) error
	SendDraft(ctx context.Context, grantID, draftID string) (*domain.Message, error)

	// Folder operations
	GetFolders(ctx context.Context, grantID string) ([]domain.Folder, error)
	GetFolder(ctx context.Context, grantID, folderID string) (*domain.Folder, error)
	CreateFolder(ctx context.Context, grantID string, req *domain.CreateFolderRequest) (*domain.Folder, error)
	UpdateFolder(ctx context.Context, grantID, folderID string, req *domain.UpdateFolderRequest) (*domain.Folder, error)
	DeleteFolder(ctx context.Context, grantID, folderID string) error

	// Attachment operations
	ListAttachments(ctx context.Context, grantID, messageID string) ([]domain.Attachment, error)
	GetAttachment(ctx context.Context, grantID, messageID, attachmentID string) (*domain.Attachment, error)
	DownloadAttachment(ctx context.Context, grantID, messageID, attachmentID string) (io.ReadCloser, error)

	// Calendar operations
	GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error)
	GetCalendar(ctx context.Context, grantID, calendarID string) (*domain.Calendar, error)
	CreateCalendar(ctx context.Context, grantID string, req *domain.CreateCalendarRequest) (*domain.Calendar, error)
	UpdateCalendar(ctx context.Context, grantID, calendarID string, req *domain.UpdateCalendarRequest) (*domain.Calendar, error)
	DeleteCalendar(ctx context.Context, grantID, calendarID string) error

	// Event operations
	GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error)
	GetEventsWithCursor(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) (*domain.EventListResponse, error)
	GetEvent(ctx context.Context, grantID, calendarID, eventID string) (*domain.Event, error)
	CreateEvent(ctx context.Context, grantID, calendarID string, req *domain.CreateEventRequest) (*domain.Event, error)
	UpdateEvent(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error)
	DeleteEvent(ctx context.Context, grantID, calendarID, eventID string) error
	SendRSVP(ctx context.Context, grantID, calendarID, eventID string, req *domain.SendRSVPRequest) error

	// Availability operations
	GetFreeBusy(ctx context.Context, grantID string, req *domain.FreeBusyRequest) (*domain.FreeBusyResponse, error)
	GetAvailability(ctx context.Context, req *domain.AvailabilityRequest) (*domain.AvailabilityResponse, error)

	// Contact operations
	GetContacts(ctx context.Context, grantID string, params *domain.ContactQueryParams) ([]domain.Contact, error)
	GetContactsWithCursor(ctx context.Context, grantID string, params *domain.ContactQueryParams) (*domain.ContactListResponse, error)
	GetContact(ctx context.Context, grantID, contactID string) (*domain.Contact, error)
	CreateContact(ctx context.Context, grantID string, req *domain.CreateContactRequest) (*domain.Contact, error)
	UpdateContact(ctx context.Context, grantID, contactID string, req *domain.UpdateContactRequest) (*domain.Contact, error)
	DeleteContact(ctx context.Context, grantID, contactID string) error
	GetContactGroups(ctx context.Context, grantID string) ([]domain.ContactGroup, error)
	GetContactGroup(ctx context.Context, grantID, groupID string) (*domain.ContactGroup, error)
	CreateContactGroup(ctx context.Context, grantID string, req *domain.CreateContactGroupRequest) (*domain.ContactGroup, error)
	UpdateContactGroup(ctx context.Context, grantID, groupID string, req *domain.UpdateContactGroupRequest) (*domain.ContactGroup, error)
	DeleteContactGroup(ctx context.Context, grantID, groupID string) error

	// Webhook operations (admin-level, uses API key)
	ListWebhooks(ctx context.Context) ([]domain.Webhook, error)
	GetWebhook(ctx context.Context, webhookID string) (*domain.Webhook, error)
	CreateWebhook(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error)
	UpdateWebhook(ctx context.Context, webhookID string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID string) error
	SendWebhookTestEvent(ctx context.Context, webhookURL string) error
	GetWebhookMockPayload(ctx context.Context, triggerType string) (map[string]interface{}, error)

	// Configuration
	SetRegion(region string)
	SetCredentials(clientID, clientSecret, apiKey string)
}

// OAuthServer defines the interface for the OAuth callback server.
type OAuthServer interface {
	// Start starts the server on the configured port.
	Start() error

	// Stop stops the server.
	Stop() error

	// WaitForCallback waits for the OAuth callback and returns the auth code.
	WaitForCallback(ctx context.Context) (string, error)

	// GetRedirectURI returns the redirect URI for OAuth.
	GetRedirectURI() string
}

// Browser defines the interface for opening URLs in the browser.
type Browser interface {
	// Open opens a URL in the default browser.
	Open(url string) error
}

// GrantStore defines the interface for storing grant information.
type GrantStore interface {
	// SaveGrant saves grant info to storage.
	SaveGrant(info domain.GrantInfo) error

	// GetGrant retrieves grant info by ID.
	GetGrant(grantID string) (*domain.GrantInfo, error)

	// GetGrantByEmail retrieves grant info by email.
	GetGrantByEmail(email string) (*domain.GrantInfo, error)

	// ListGrants returns all stored grants.
	ListGrants() ([]domain.GrantInfo, error)

	// DeleteGrant removes a grant from storage.
	DeleteGrant(grantID string) error

	// SetDefaultGrant sets the default grant ID.
	SetDefaultGrant(grantID string) error

	// GetDefaultGrant returns the default grant ID.
	GetDefaultGrant() (string, error)

	// ClearGrants removes all grants from storage.
	ClearGrants() error
}
