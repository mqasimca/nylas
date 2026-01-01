// Package ports defines the interfaces for external adapters.
package ports

import (
	"context"
	"io"

	"github.com/mqasimca/nylas/internal/domain"
)

// SlackClient defines the contract for Slack API operations.
// Implementations handle authentication, rate limiting, and Slack API error responses.
type SlackClient interface {
	// TestAuth validates the workspace token and returns authentication details.
	// Use this to verify token validity before other operations.
	TestAuth(ctx context.Context) (*domain.SlackAuth, error)

	// ListChannels returns all accessible channels in the workspace.
	// Use params to filter by type, exclude archived, or paginate results.
	ListChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error)

	// GetChannel returns a single channel by ID.
	// Returns an error if the channel doesn't exist or is not accessible.
	GetChannel(ctx context.Context, channelID string) (*domain.SlackChannel, error)

	// ListMyChannels returns only channels the current user is a member of.
	// Faster than ListChannels for large workspaces.
	ListMyChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error)

	// GetMessages returns messages from a channel in reverse chronological order.
	// Use params to specify time range, limit, and pagination cursor.
	GetMessages(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error)

	// GetThreadReplies returns replies in a message thread.
	// threadTS is the timestamp of the parent message. Replies are in chronological order.
	GetThreadReplies(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error)

	// SendMessage posts a new message to a channel.
	// Returns the sent message with its timestamp ID.
	SendMessage(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error)

	// UpdateMessage edits an existing message.
	// messageTS is the timestamp ID of the message to edit.
	UpdateMessage(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error)

	// DeleteMessage removes a message from a channel.
	// Only the message author or workspace admins can delete messages.
	DeleteMessage(ctx context.Context, channelID, messageTS string) error

	// ListUsers returns workspace members, excluding deleted/deactivated users.
	// Use limit and cursor for pagination through large workspaces.
	ListUsers(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error)

	// GetUser returns a single user by their Slack user ID.
	// Returns an error if the user doesn't exist.
	GetUser(ctx context.Context, userID string) (*domain.SlackUser, error)

	// GetCurrentUser returns the user associated with the current token.
	GetCurrentUser(ctx context.Context) (*domain.SlackUser, error)

	// SearchMessages searches for messages matching a query string.
	// Query supports Slack search syntax (e.g., "from:@user", "in:#channel").
	SearchMessages(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error)

	// ListFiles returns files uploaded to a channel or workspace.
	// Use params to filter by channel, user, or file type.
	ListFiles(ctx context.Context, params *domain.SlackFileQueryParams) (*domain.SlackFileListResponse, error)

	// GetFileInfo returns metadata for a single file by its ID.
	// Returns an error if the file doesn't exist or is not accessible.
	GetFileInfo(ctx context.Context, fileID string) (*domain.SlackAttachment, error)

	// DownloadFile downloads file content from a private download URL.
	// The caller must close the returned ReadCloser when done.
	DownloadFile(ctx context.Context, downloadURL string) (io.ReadCloser, error)
}
