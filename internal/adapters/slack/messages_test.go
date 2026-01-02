//go:build !integration
// +build !integration

package slack

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestMockClient_GetMessages_AllScenarios(t *testing.T) {
	tests := []struct {
		name        string
		params      *domain.SlackMessageQueryParams
		setupMock   func(*MockClient)
		wantLen     int
		wantErr     error
		wantHasMore bool
	}{
		{
			name:    "nil params returns default messages",
			params:  nil,
			wantLen: 2,
		},
		{
			name: "with channel ID",
			params: &domain.SlackMessageQueryParams{
				ChannelID: "C12345",
			},
			wantLen: 2,
		},
		{
			name: "with limit",
			params: &domain.SlackMessageQueryParams{
				ChannelID: "C12345",
				Limit:     5,
			},
			setupMock: func(m *MockClient) {
				m.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
					assert.Equal(t, 5, params.Limit)
					msgs := make([]domain.SlackMessage, 5)
					for i := 0; i < 5; i++ {
						msgs[i] = domain.SlackMessage{ID: "msg-" + string(rune('0'+i))}
					}
					return &domain.SlackMessageListResponse{
						Messages: msgs,
						HasMore:  true,
					}, nil
				}
			},
			wantLen:     5,
			wantHasMore: true,
		},
		{
			name: "with time range",
			params: &domain.SlackMessageQueryParams{
				ChannelID: "C12345",
				Oldest:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Newest:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			setupMock: func(m *MockClient) {
				m.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
					assert.False(t, params.Oldest.IsZero())
					assert.False(t, params.Newest.IsZero())
					return &domain.SlackMessageListResponse{
						Messages: []domain.SlackMessage{{ID: "time-filtered"}},
					}, nil
				}
			},
			wantLen: 1,
		},
		{
			name: "channel not found error",
			params: &domain.SlackMessageQueryParams{
				ChannelID: "C99999",
			},
			setupMock: func(m *MockClient) {
				m.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
					return nil, domain.ErrSlackChannelNotFound
				}
			},
			wantErr: domain.ErrSlackChannelNotFound,
		},
		{
			name:   "rate limit error",
			params: &domain.SlackMessageQueryParams{ChannelID: "C1"},
			setupMock: func(m *MockClient) {
				m.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
					return nil, domain.ErrSlackRateLimited
				}
			},
			wantErr: domain.ErrSlackRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			resp, err := mock.GetMessages(context.Background(), tt.params)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Messages, tt.wantLen)
			if tt.wantHasMore {
				assert.True(t, resp.HasMore)
			}
		})
	}
}

func TestMockClient_GetThreadReplies_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		threadTS  string
		limit     int
		setupMock func(*MockClient)
		wantLen   int
		wantErr   error
	}{
		{
			name:      "default replies",
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			limit:     10,
			wantLen:   2,
		},
		{
			name:      "custom replies",
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			limit:     5,
			setupMock: func(m *MockClient) {
				m.GetThreadRepliesFunc = func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
					return []domain.SlackMessage{
						{ID: "parent", Text: "Parent message"},
						{ID: "reply1", Text: "Reply 1", IsReply: true},
						{ID: "reply2", Text: "Reply 2", IsReply: true},
					}, nil
				}
			},
			wantLen: 3,
		},
		{
			name:      "thread not found",
			channelID: "C12345",
			threadTS:  "0000000000.000000",
			limit:     10,
			setupMock: func(m *MockClient) {
				m.GetThreadRepliesFunc = func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
					return nil, domain.ErrSlackMessageNotFound
				}
			},
			wantErr: domain.ErrSlackMessageNotFound,
		},
		{
			name:      "empty thread",
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			limit:     10,
			setupMock: func(m *MockClient) {
				m.GetThreadRepliesFunc = func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
					return []domain.SlackMessage{}, nil
				}
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			replies, err := mock.GetThreadReplies(context.Background(), tt.channelID, tt.threadTS, tt.limit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, replies)
				return
			}

			require.NoError(t, err)
			assert.Len(t, replies, tt.wantLen)
		})
	}
}

func TestMockClient_SendMessage_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		request   *domain.SlackSendMessageRequest
		setupMock func(*MockClient)
		wantReply bool
		wantErr   error
	}{
		{
			name: "simple message",
			request: &domain.SlackSendMessageRequest{
				ChannelID: "C12345",
				Text:      "Hello, world!",
			},
			wantReply: false,
		},
		{
			name: "thread reply",
			request: &domain.SlackSendMessageRequest{
				ChannelID: "C12345",
				Text:      "This is a reply",
				ThreadTS:  "1234567890.123456",
			},
			wantReply: true,
		},
		{
			name: "broadcast reply",
			request: &domain.SlackSendMessageRequest{
				ChannelID: "C12345",
				Text:      "Important update",
				ThreadTS:  "1234567890.123456",
				Broadcast: true,
			},
			wantReply: true,
		},
		{
			name: "channel not found",
			request: &domain.SlackSendMessageRequest{
				ChannelID: "C99999",
				Text:      "Test",
			},
			setupMock: func(m *MockClient) {
				m.SendMessageFunc = func(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
					return nil, domain.ErrSlackChannelNotFound
				}
			},
			wantErr: domain.ErrSlackChannelNotFound,
		},
		{
			name: "permission denied",
			request: &domain.SlackSendMessageRequest{
				ChannelID: "C12345",
				Text:      "Forbidden",
			},
			setupMock: func(m *MockClient) {
				m.SendMessageFunc = func(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
					return nil, domain.ErrSlackPermissionDenied
				}
			},
			wantErr: domain.ErrSlackPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			msg, err := mock.SendMessage(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, msg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
			assert.Equal(t, tt.request.ChannelID, msg.ChannelID)
			assert.Equal(t, tt.request.Text, msg.Text)
			assert.Equal(t, tt.wantReply, msg.IsReply)
		})
	}
}

func TestMockClient_UpdateMessage_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		messageTS string
		newText   string
		setupMock func(*MockClient)
		wantErr   error
	}{
		{
			name:      "successful update",
			channelID: "C12345",
			messageTS: "1234567890.123456",
			newText:   "Updated message",
		},
		{
			name:      "message not found",
			channelID: "C12345",
			messageTS: "0000000000.000000",
			newText:   "Won't update",
			setupMock: func(m *MockClient) {
				m.UpdateMessageFunc = func(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
					return nil, domain.ErrSlackMessageNotFound
				}
			},
			wantErr: domain.ErrSlackMessageNotFound,
		},
		{
			name:      "permission denied",
			channelID: "C12345",
			messageTS: "1234567890.123456",
			newText:   "Can't edit",
			setupMock: func(m *MockClient) {
				m.UpdateMessageFunc = func(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
					return nil, domain.ErrSlackPermissionDenied
				}
			},
			wantErr: domain.ErrSlackPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			msg, err := mock.UpdateMessage(context.Background(), tt.channelID, tt.messageTS, tt.newText)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, msg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
			assert.True(t, msg.Edited)
			assert.Equal(t, tt.newText, msg.Text)
		})
	}
}

func TestMockClient_DeleteMessage_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		messageTS string
		setupMock func(*MockClient)
		wantErr   error
	}{
		{
			name:      "successful delete",
			channelID: "C12345",
			messageTS: "1234567890.123456",
		},
		{
			name:      "message not found",
			channelID: "C12345",
			messageTS: "0000000000.000000",
			setupMock: func(m *MockClient) {
				m.DeleteMessageFunc = func(ctx context.Context, channelID, messageTS string) error {
					return domain.ErrSlackMessageNotFound
				}
			},
			wantErr: domain.ErrSlackMessageNotFound,
		},
		{
			name:      "channel not found",
			channelID: "C99999",
			messageTS: "1234567890.123456",
			setupMock: func(m *MockClient) {
				m.DeleteMessageFunc = func(ctx context.Context, channelID, messageTS string) error {
					return domain.ErrSlackChannelNotFound
				}
			},
			wantErr: domain.ErrSlackChannelNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			err := mock.DeleteMessage(context.Background(), tt.channelID, tt.messageTS)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestMockClient_SearchMessages_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		limit     int
		setupMock func(*MockClient)
		wantLen   int
		wantErr   error
	}{
		{
			name:    "default search",
			query:   "test",
			limit:   10,
			wantLen: 1,
		},
		{
			name:  "custom search results",
			query: "important",
			limit: 20,
			setupMock: func(m *MockClient) {
				m.SearchMessagesFunc = func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
					return []domain.SlackMessage{
						{ID: "1", Text: "Important message 1"},
						{ID: "2", Text: "Important message 2"},
						{ID: "3", Text: "Important message 3"},
					}, nil
				}
			},
			wantLen: 3,
		},
		{
			name:  "no results",
			query: "nonexistent",
			limit: 10,
			setupMock: func(m *MockClient) {
				m.SearchMessagesFunc = func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
					return []domain.SlackMessage{}, nil
				}
			},
			wantLen: 0,
		},
		{
			name:  "search error",
			query: "error",
			limit: 10,
			setupMock: func(m *MockClient) {
				m.SearchMessagesFunc = func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
					return nil, domain.ErrSlackRateLimited
				}
			},
			wantErr: domain.ErrSlackRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			msgs, err := mock.SearchMessages(context.Background(), tt.query, tt.limit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, msgs)
				return
			}

			require.NoError(t, err)
			assert.Len(t, msgs, tt.wantLen)
		})
	}
}
