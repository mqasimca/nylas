// Package ports defines the interfaces for external adapters.
package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// SlackClient defines the contract for Slack operations.
type SlackClient interface {
	// TestAuth validates the token and returns user/team info.
	TestAuth(ctx context.Context) (*domain.SlackAuth, error)

	// ListChannels returns all accessible channels.
	ListChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error)

	// GetChannel returns a single channel by ID.
	GetChannel(ctx context.Context, channelID string) (*domain.SlackChannel, error)

	// GetMessages returns messages from a channel.
	GetMessages(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error)

	// GetThreadReplies returns replies in a thread.
	GetThreadReplies(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error)

	// SendMessage sends a new message to a channel.
	SendMessage(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error)

	// UpdateMessage edits an existing message.
	UpdateMessage(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error)

	// DeleteMessage removes a message.
	DeleteMessage(ctx context.Context, channelID, messageTS string) error

	// ListUsers returns workspace members.
	ListUsers(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error)

	// GetUser returns a single user by ID.
	GetUser(ctx context.Context, userID string) (*domain.SlackUser, error)

	// GetCurrentUser returns the authenticated user.
	GetCurrentUser(ctx context.Context) (*domain.SlackUser, error)

	// SearchMessages searches for messages matching a query.
	SearchMessages(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error)
}
