// Package client provides Nylas API clients for plugins.
package client

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/pkg/go-nylas/common"
	"github.com/mqasimca/nylas/pkg/go-nylas/config"
)

// EmailClient provides email operations for plugins.
type EmailClient struct {
	adapter *nylas.HTTPClient
	grantID string
	config  *config.Config
}

// NewEmailClient creates a new email client for the given configuration.
func NewEmailClient(cfg *config.Config) (*EmailClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.GetAPIKey() == "" {
		return nil, fmt.Errorf("API key not configured")
	}
	if cfg.GetGrantID() == "" {
		return nil, fmt.Errorf("grant ID not configured")
	}

	// Create internal HTTP client
	adapter := nylas.NewHTTPClient()
	adapter.SetRegion(cfg.GetRegion())
	adapter.SetCredentials(cfg.GetClientID(), "", cfg.GetAPIKey())

	return &EmailClient{
		adapter: adapter,
		grantID: cfg.GetGrantID(),
		config:  cfg,
	}, nil
}

// Message represents an email message.
type Message struct {
	ID          string
	ThreadID    string
	Subject     string
	From        []common.Participant
	To          []common.Participant
	Cc          []common.Participant
	Bcc         []common.Participant
	ReplyTo     []common.Participant
	Body        string
	Snippet     string
	Date        int64
	Unread      bool
	Starred     bool
	Folders     []string
	Attachments []common.Attachment
}

// ListOptions contains options for listing messages.
type ListOptions struct {
	Limit     int
	Offset    int
	PageToken string
	Subject   string
	From      string
	To        string
	Unread    *bool
}

// SendOptions contains options for sending messages.
type SendOptions struct {
	Subject     string
	To          []common.Participant
	Cc          []common.Participant
	Bcc         []common.Participant
	ReplyTo     []common.Participant
	Body        string
	Attachments []common.Attachment
}

// List retrieves a list of messages.
func (c *EmailClient) List(ctx context.Context, opts *ListOptions) ([]*Message, error) {
	if opts == nil {
		opts = &ListOptions{Limit: 10}
	}

	// Convert to internal params
	params := &domain.MessageQueryParams{
		Limit:     opts.Limit,
		Offset:    opts.Offset,
		PageToken: opts.PageToken,
		Subject:   opts.Subject,
		From:      opts.From,
		To:        opts.To,
		Unread:    opts.Unread,
	}

	// Call internal adapter
	internalMessages, err := c.adapter.GetMessagesWithParams(ctx, c.grantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// Convert to public format
	messages := make([]*Message, len(internalMessages))
	for i, m := range internalMessages {
		messages[i] = convertMessage(&m)
	}

	return messages, nil
}

// Get retrieves a single message by ID.
func (c *EmailClient) Get(ctx context.Context, messageID string) (*Message, error) {
	internalMessage, err := c.adapter.GetMessage(ctx, c.grantID, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return convertMessage(internalMessage), nil
}

// Send sends an email message.
func (c *EmailClient) Send(ctx context.Context, opts *SendOptions) (*Message, error) {
	if opts == nil {
		return nil, fmt.Errorf("send options cannot be nil")
	}

	// Convert to internal format
	msg := &domain.SendMessageRequest{
		Subject: opts.Subject,
		Body:    opts.Body,
	}

	// Convert participants
	msg.To = make([]domain.Participant, len(opts.To))
	for i, p := range opts.To {
		msg.To[i] = domain.Participant{Name: p.Name, Email: p.Email}
	}

	msg.Cc = make([]domain.Participant, len(opts.Cc))
	for i, p := range opts.Cc {
		msg.Cc[i] = domain.Participant{Name: p.Name, Email: p.Email}
	}

	msg.Bcc = make([]domain.Participant, len(opts.Bcc))
	for i, p := range opts.Bcc {
		msg.Bcc[i] = domain.Participant{Name: p.Name, Email: p.Email}
	}

	// Send via internal adapter
	sentMessage, err := c.adapter.SendMessage(ctx, c.grantID, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return convertMessage(sentMessage), nil
}

// Delete deletes a message.
func (c *EmailClient) Delete(ctx context.Context, messageID string) error {
	err := c.adapter.DeleteMessage(ctx, c.grantID, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// Update updates message properties (read status, starred, etc).
func (c *EmailClient) Update(ctx context.Context, messageID string, unread *bool, starred *bool) (*Message, error) {
	update := &domain.MessageUpdate{
		Unread:  unread,
		Starred: starred,
	}

	updatedMessage, err := c.adapter.UpdateMessage(ctx, c.grantID, messageID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return convertMessage(updatedMessage), nil
}

// Helper functions

func convertMessage(internal *domain.Message) *Message {
	if internal == nil {
		return nil
	}

	msg := &Message{
		ID:       internal.ID,
		ThreadID: internal.ThreadID,
		Subject:  internal.Subject,
		Body:     internal.Body,
		Snippet:  internal.Snippet,
		Date:     internal.Date,
		Unread:   internal.Unread,
		Starred:  internal.Starred,
		Folders:  internal.Folders,
	}

	// Convert participants
	msg.From = make([]common.Participant, len(internal.From))
	for i, p := range internal.From {
		msg.From[i] = common.Participant{Name: p.Name, Email: p.Email}
	}

	msg.To = make([]common.Participant, len(internal.To))
	for i, p := range internal.To {
		msg.To[i] = common.Participant{Name: p.Name, Email: p.Email}
	}

	msg.Cc = make([]common.Participant, len(internal.Cc))
	for i, p := range internal.Cc {
		msg.Cc[i] = common.Participant{Name: p.Name, Email: p.Email}
	}

	msg.Bcc = make([]common.Participant, len(internal.Bcc))
	for i, p := range internal.Bcc {
		msg.Bcc[i] = common.Participant{Name: p.Name, Email: p.Email}
	}

	msg.ReplyTo = make([]common.Participant, len(internal.ReplyTo))
	for i, p := range internal.ReplyTo {
		msg.ReplyTo[i] = common.Participant{Name: p.Name, Email: p.Email}
	}

	// Convert attachments
	msg.Attachments = make([]common.Attachment, len(internal.Attachments))
	for i, a := range internal.Attachments {
		msg.Attachments[i] = common.Attachment{
			ID:          a.ID,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        int(a.Size),
			ContentID:   a.ContentID,
		}
	}

	return msg
}
