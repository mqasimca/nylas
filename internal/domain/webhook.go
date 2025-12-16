package domain

import "time"

// Webhook represents a Nylas webhook subscription.
type Webhook struct {
	ID                         string    `json:"id"`
	Description                string    `json:"description,omitempty"`
	TriggerTypes               []string  `json:"trigger_types"`
	WebhookURL                 string    `json:"webhook_url"`
	WebhookSecret              string    `json:"webhook_secret,omitempty"`
	Status                     string    `json:"status"` // active, inactive, failing
	NotificationEmailAddresses []string  `json:"notification_email_addresses,omitempty"`
	StatusUpdatedAt            time.Time `json:"status_updated_at,omitempty"`
	CreatedAt                  time.Time `json:"created_at,omitempty"`
	UpdatedAt                  time.Time `json:"updated_at,omitempty"`
}

// CreateWebhookRequest for creating a new webhook.
type CreateWebhookRequest struct {
	TriggerTypes               []string `json:"trigger_types"`
	WebhookURL                 string   `json:"webhook_url"`
	Description                string   `json:"description,omitempty"`
	NotificationEmailAddresses []string `json:"notification_email_addresses,omitempty"`
}

// UpdateWebhookRequest for updating a webhook.
type UpdateWebhookRequest struct {
	TriggerTypes               []string `json:"trigger_types,omitempty"`
	WebhookURL                 string   `json:"webhook_url,omitempty"`
	Description                string   `json:"description,omitempty"`
	NotificationEmailAddresses []string `json:"notification_email_addresses,omitempty"`
	Status                     string   `json:"status,omitempty"` // active, inactive
}

// WebhookTestRequest for sending a test webhook event.
type WebhookTestRequest struct {
	WebhookURL string `json:"webhook_url"`
}

// WebhookMockPayloadRequest for getting a mock payload.
type WebhookMockPayloadRequest struct {
	TriggerType string `json:"trigger_type"`
}

// WebhookListResponse represents a paginated webhook list.
type WebhookListResponse struct {
	Data       []Webhook  `json:"data"`
	Pagination Pagination `json:"pagination,omitempty"`
}

// Common webhook trigger types.
const (
	// Grant triggers
	TriggerGrantCreated = "grant.created"
	TriggerGrantDeleted = "grant.deleted"
	TriggerGrantExpired = "grant.expired"
	TriggerGrantUpdated = "grant.updated"

	// Message triggers
	TriggerMessageCreated          = "message.created"
	TriggerMessageUpdated          = "message.updated"
	TriggerMessageOpenedTruncated  = "message.opened.truncated"
	TriggerMessageLinkClickedMeta  = "message.link_clicked.metadata"

	// Thread triggers
	TriggerThreadReplied = "thread.replied"

	// Event triggers
	TriggerEventCreated = "event.created"
	TriggerEventUpdated = "event.updated"
	TriggerEventDeleted = "event.deleted"

	// Contact triggers
	TriggerContactCreated = "contact.created"
	TriggerContactUpdated = "contact.updated"
	TriggerContactDeleted = "contact.deleted"

	// Calendar triggers
	TriggerCalendarCreated = "calendar.created"
	TriggerCalendarUpdated = "calendar.updated"
	TriggerCalendarDeleted = "calendar.deleted"

	// Folder triggers
	TriggerFolderCreated = "folder.created"
	TriggerFolderUpdated = "folder.updated"
	TriggerFolderDeleted = "folder.deleted"
)

// AllTriggerTypes returns all available trigger types.
func AllTriggerTypes() []string {
	return []string{
		TriggerGrantCreated,
		TriggerGrantDeleted,
		TriggerGrantExpired,
		TriggerGrantUpdated,
		TriggerMessageCreated,
		TriggerMessageUpdated,
		TriggerThreadReplied,
		TriggerEventCreated,
		TriggerEventUpdated,
		TriggerEventDeleted,
		TriggerContactCreated,
		TriggerContactUpdated,
		TriggerContactDeleted,
		TriggerCalendarCreated,
		TriggerCalendarUpdated,
		TriggerCalendarDeleted,
		TriggerFolderCreated,
		TriggerFolderUpdated,
		TriggerFolderDeleted,
	}
}

// TriggerTypeCategories returns trigger types grouped by category.
func TriggerTypeCategories() map[string][]string {
	return map[string][]string{
		"grant": {
			TriggerGrantCreated,
			TriggerGrantDeleted,
			TriggerGrantExpired,
			TriggerGrantUpdated,
		},
		"message": {
			TriggerMessageCreated,
			TriggerMessageUpdated,
		},
		"thread": {
			TriggerThreadReplied,
		},
		"event": {
			TriggerEventCreated,
			TriggerEventUpdated,
			TriggerEventDeleted,
		},
		"contact": {
			TriggerContactCreated,
			TriggerContactUpdated,
			TriggerContactDeleted,
		},
		"calendar": {
			TriggerCalendarCreated,
			TriggerCalendarUpdated,
			TriggerCalendarDeleted,
		},
		"folder": {
			TriggerFolderCreated,
			TriggerFolderUpdated,
			TriggerFolderDeleted,
		},
	}
}
