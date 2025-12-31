// mock.go provides mock implementations of the Slack client for testing.

package slack

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// Ensure MockClient implements ports.SlackClient.
var _ ports.SlackClient = (*MockClient)(nil)

// MockClient is a mock implementation of SlackClient for testing.
type MockClient struct {
	TestAuthFunc         func(ctx context.Context) (*domain.SlackAuth, error)
	ListChannelsFunc     func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error)
	ListMyChannelsFunc   func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error)
	GetChannelFunc       func(ctx context.Context, channelID string) (*domain.SlackChannel, error)
	GetMessagesFunc      func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error)
	GetThreadRepliesFunc func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error)
	SendMessageFunc      func(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error)
	UpdateMessageFunc    func(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error)
	DeleteMessageFunc    func(ctx context.Context, channelID, messageTS string) error
	ListUsersFunc        func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error)
	GetUserFunc          func(ctx context.Context, userID string) (*domain.SlackUser, error)
	GetCurrentUserFunc   func(ctx context.Context) (*domain.SlackUser, error)
	SearchMessagesFunc   func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error)
}

// NewMockClient creates a new mock client with default implementations.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// TestAuth validates the token and returns auth info.
func (m *MockClient) TestAuth(ctx context.Context) (*domain.SlackAuth, error) {
	if m.TestAuthFunc != nil {
		return m.TestAuthFunc(ctx)
	}
	return &domain.SlackAuth{
		UserID:   "U123456",
		TeamID:   "T123456",
		TeamName: "Test Workspace",
		UserName: "testuser",
	}, nil
}

// ListChannels returns all accessible channels.
func (m *MockClient) ListChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
	if m.ListChannelsFunc != nil {
		return m.ListChannelsFunc(ctx, params)
	}
	return &domain.SlackChannelListResponse{
		Channels: []domain.SlackChannel{
			{ID: "C123456", Name: "general", IsChannel: true},
			{ID: "C234567", Name: "random", IsChannel: true},
		},
	}, nil
}

// GetChannel returns a single channel by ID.
func (m *MockClient) GetChannel(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
	if m.GetChannelFunc != nil {
		return m.GetChannelFunc(ctx, channelID)
	}
	return &domain.SlackChannel{
		ID:        channelID,
		Name:      "general",
		IsChannel: true,
	}, nil
}

// ListMyChannels returns only channels the user is a member of.
func (m *MockClient) ListMyChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
	if m.ListMyChannelsFunc != nil {
		return m.ListMyChannelsFunc(ctx, params)
	}
	return &domain.SlackChannelListResponse{
		Channels: []domain.SlackChannel{
			{ID: "C123456", Name: "general", IsChannel: true, IsMember: true},
			{ID: "C234567", Name: "random", IsChannel: true, IsMember: true},
		},
	}, nil
}

// GetMessages returns messages from a channel.
func (m *MockClient) GetMessages(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
	if m.GetMessagesFunc != nil {
		return m.GetMessagesFunc(ctx, params)
	}
	return &domain.SlackMessageListResponse{
		Messages: []domain.SlackMessage{
			{ID: "1234567890.123456", Text: "Hello, world!", Username: "testuser"},
			{ID: "1234567890.123457", Text: "How are you?", Username: "otheruser"},
		},
	}, nil
}

// GetThreadReplies returns replies in a thread.
func (m *MockClient) GetThreadReplies(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
	if m.GetThreadRepliesFunc != nil {
		return m.GetThreadRepliesFunc(ctx, channelID, threadTS, limit)
	}
	return []domain.SlackMessage{
		{ID: threadTS, Text: "Original message", Username: "testuser"},
		{ID: "1234567890.123458", Text: "Reply 1", Username: "otheruser", ThreadTS: threadTS, IsReply: true},
	}, nil
}

// SendMessage sends a new message.
func (m *MockClient) SendMessage(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, req)
	}
	return &domain.SlackMessage{
		ID:        "1234567890.999999",
		ChannelID: req.ChannelID,
		Text:      req.Text,
		ThreadTS:  req.ThreadTS,
		IsReply:   req.ThreadTS != "",
	}, nil
}

// UpdateMessage edits an existing message.
func (m *MockClient) UpdateMessage(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
	if m.UpdateMessageFunc != nil {
		return m.UpdateMessageFunc(ctx, channelID, messageTS, newText)
	}
	return &domain.SlackMessage{
		ID:        messageTS,
		ChannelID: channelID,
		Text:      newText,
		Edited:    true,
	}, nil
}

// DeleteMessage removes a message.
func (m *MockClient) DeleteMessage(ctx context.Context, channelID, messageTS string) error {
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, channelID, messageTS)
	}
	return nil
}

// ListUsers returns workspace members.
func (m *MockClient) ListUsers(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, cursor)
	}
	return &domain.SlackUserListResponse{
		Users: []domain.SlackUser{
			{ID: "U123456", Name: "testuser", RealName: "Test User"},
			{ID: "U234567", Name: "otheruser", RealName: "Other User"},
		},
	}, nil
}

// GetUser returns a single user by ID.
func (m *MockClient) GetUser(ctx context.Context, userID string) (*domain.SlackUser, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return &domain.SlackUser{
		ID:       userID,
		Name:     "testuser",
		RealName: "Test User",
	}, nil
}

// GetCurrentUser returns the authenticated user.
func (m *MockClient) GetCurrentUser(ctx context.Context) (*domain.SlackUser, error) {
	if m.GetCurrentUserFunc != nil {
		return m.GetCurrentUserFunc(ctx)
	}
	return &domain.SlackUser{
		ID:       "U123456",
		Name:     "testuser",
		RealName: "Test User",
	}, nil
}

// SearchMessages searches for messages.
func (m *MockClient) SearchMessages(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
	if m.SearchMessagesFunc != nil {
		return m.SearchMessagesFunc(ctx, query, limit)
	}
	return []domain.SlackMessage{
		{ID: "1234567890.123456", Text: "Message matching: " + query, Username: "testuser"},
	}, nil
}
