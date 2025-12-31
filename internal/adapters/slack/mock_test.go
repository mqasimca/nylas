//go:build !integration
// +build !integration

package slack

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewMockClient(t *testing.T) {
	mock := NewMockClient()

	assert.NotNil(t, mock)
	// All function fields should be nil initially
	assert.Nil(t, mock.TestAuthFunc)
	assert.Nil(t, mock.ListChannelsFunc)
	assert.Nil(t, mock.GetChannelFunc)
	assert.Nil(t, mock.GetMessagesFunc)
	assert.Nil(t, mock.GetThreadRepliesFunc)
	assert.Nil(t, mock.SendMessageFunc)
	assert.Nil(t, mock.UpdateMessageFunc)
	assert.Nil(t, mock.DeleteMessageFunc)
	assert.Nil(t, mock.ListUsersFunc)
	assert.Nil(t, mock.GetUserFunc)
	assert.Nil(t, mock.GetCurrentUserFunc)
	assert.Nil(t, mock.SearchMessagesFunc)
}

func TestMockClient_TestAuth(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		auth, err := mock.TestAuth(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "U123456", auth.UserID)
		assert.Equal(t, "Test Workspace", auth.TeamName)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.TestAuthFunc = func(ctx context.Context) (*domain.SlackAuth, error) {
			return &domain.SlackAuth{
				UserID:   "U99999",
				TeamID:   "T99999",
				TeamName: "Custom Team",
				UserName: "customuser",
			}, nil
		}

		auth, err := mock.TestAuth(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "U99999", auth.UserID)
		assert.Equal(t, "Custom Team", auth.TeamName)
	})

	t.Run("returns custom error", func(t *testing.T) {
		mock := NewMockClient()
		mock.TestAuthFunc = func(ctx context.Context) (*domain.SlackAuth, error) {
			return nil, domain.ErrSlackAuthFailed
		}

		auth, err := mock.TestAuth(context.Background())
		assert.ErrorIs(t, err, domain.ErrSlackAuthFailed)
		assert.Nil(t, auth)
	})
}

func TestMockClient_ListChannels(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		resp, err := mock.ListChannels(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, resp.Channels, 2)
		assert.Equal(t, "general", resp.Channels[0].Name)
	})

	t.Run("calls custom function with params", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
			// Verify params are passed through
			assert.Equal(t, 50, params.Limit)
			return &domain.SlackChannelListResponse{
				Channels: []domain.SlackChannel{
					{ID: "C1", Name: "custom-channel"},
				},
			}, nil
		}

		resp, err := mock.ListChannels(context.Background(), &domain.SlackChannelQueryParams{Limit: 50})
		require.NoError(t, err)
		assert.Len(t, resp.Channels, 1)
		assert.Equal(t, "custom-channel", resp.Channels[0].Name)
	})
}

func TestMockClient_GetChannel(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		ch, err := mock.GetChannel(context.Background(), "C12345")
		require.NoError(t, err)
		assert.Equal(t, "C12345", ch.ID)
		assert.Equal(t, "general", ch.Name)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
			if channelID == "C12345" {
				return &domain.SlackChannel{ID: "C12345", Name: "custom"}, nil
			}
			return nil, domain.ErrSlackChannelNotFound
		}

		ch, err := mock.GetChannel(context.Background(), "C12345")
		require.NoError(t, err)
		assert.Equal(t, "custom", ch.Name)

		ch, err = mock.GetChannel(context.Background(), "C99999")
		assert.ErrorIs(t, err, domain.ErrSlackChannelNotFound)
		assert.Nil(t, ch)
	})
}

func TestMockClient_GetMessages(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		resp, err := mock.GetMessages(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, resp.Messages, 2)
		assert.Equal(t, "Hello, world!", resp.Messages[0].Text)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
			return &domain.SlackMessageListResponse{
				Messages: []domain.SlackMessage{
					{ID: "custom-id", Text: "Custom message"},
				},
				HasMore: true,
			}, nil
		}

		resp, err := mock.GetMessages(context.Background(), &domain.SlackMessageQueryParams{ChannelID: "C1"})
		require.NoError(t, err)
		assert.Len(t, resp.Messages, 1)
		assert.True(t, resp.HasMore)
	})
}

func TestMockClient_GetThreadReplies(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		replies, err := mock.GetThreadReplies(context.Background(), "C1", "ts", 10)
		require.NoError(t, err)
		assert.Len(t, replies, 2)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetThreadRepliesFunc = func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
			return []domain.SlackMessage{
				{ID: "1", Text: "Reply 1"},
				{ID: "2", Text: "Reply 2"},
				{ID: "3", Text: "Reply 3"},
			}, nil
		}

		replies, err := mock.GetThreadReplies(context.Background(), "C1", "ts", 10)
		require.NoError(t, err)
		assert.Len(t, replies, 3)
	})
}

func TestMockClient_SendMessage(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		msg, err := mock.SendMessage(context.Background(), &domain.SlackSendMessageRequest{
			ChannelID: "C1",
			Text:      "Hello",
		})
		require.NoError(t, err)
		assert.Equal(t, "C1", msg.ChannelID)
		assert.Equal(t, "Hello", msg.Text)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.SendMessageFunc = func(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
			return &domain.SlackMessage{
				ID:        "custom-ts",
				ChannelID: req.ChannelID,
				Text:      req.Text,
			}, nil
		}

		msg, err := mock.SendMessage(context.Background(), &domain.SlackSendMessageRequest{
			ChannelID: "C1",
			Text:      "Hello world",
		})
		require.NoError(t, err)
		assert.Equal(t, "custom-ts", msg.ID)
	})

	t.Run("handles thread reply", func(t *testing.T) {
		mock := NewMockClient()
		msg, err := mock.SendMessage(context.Background(), &domain.SlackSendMessageRequest{
			ChannelID: "C1",
			Text:      "Reply",
			ThreadTS:  "1234567890.123456",
		})
		require.NoError(t, err)
		assert.True(t, msg.IsReply)
		assert.Equal(t, "1234567890.123456", msg.ThreadTS)
	})
}

func TestMockClient_UpdateMessage(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		msg, err := mock.UpdateMessage(context.Background(), "C1", "ts", "new text")
		require.NoError(t, err)
		assert.True(t, msg.Edited)
		assert.Equal(t, "new text", msg.Text)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.UpdateMessageFunc = func(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
			return &domain.SlackMessage{
				ID:        messageTS,
				ChannelID: channelID,
				Text:      newText,
				Edited:    true,
			}, nil
		}

		msg, err := mock.UpdateMessage(context.Background(), "C1", "ts", "updated")
		require.NoError(t, err)
		assert.True(t, msg.Edited)
	})
}

func TestMockClient_DeleteMessage(t *testing.T) {
	t.Run("returns nil when func not set", func(t *testing.T) {
		mock := NewMockClient()
		err := mock.DeleteMessage(context.Background(), "C1", "ts")
		assert.NoError(t, err)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteMessageFunc = func(ctx context.Context, channelID, messageTS string) error {
			return nil
		}

		err := mock.DeleteMessage(context.Background(), "C1", "ts")
		assert.NoError(t, err)
	})

	t.Run("returns custom error", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeleteMessageFunc = func(ctx context.Context, channelID, messageTS string) error {
			return domain.ErrSlackMessageNotFound
		}

		err := mock.DeleteMessage(context.Background(), "C1", "ts")
		assert.ErrorIs(t, err, domain.ErrSlackMessageNotFound)
	})
}

func TestMockClient_ListUsers(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		resp, err := mock.ListUsers(context.Background(), 100, "")
		require.NoError(t, err)
		assert.Len(t, resp.Users, 2)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
			return &domain.SlackUserListResponse{
				Users: []domain.SlackUser{
					{ID: "U1", Name: "alice"},
					{ID: "U2", Name: "bob"},
					{ID: "U3", Name: "charlie"},
				},
			}, nil
		}

		resp, err := mock.ListUsers(context.Background(), 100, "")
		require.NoError(t, err)
		assert.Len(t, resp.Users, 3)
	})
}

func TestMockClient_GetUser(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		user, err := mock.GetUser(context.Background(), "U12345")
		require.NoError(t, err)
		assert.Equal(t, "U12345", user.ID)
		assert.Equal(t, "testuser", user.Name)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
			return &domain.SlackUser{ID: userID, Name: "customuser"}, nil
		}

		user, err := mock.GetUser(context.Background(), "U12345")
		require.NoError(t, err)
		assert.Equal(t, "customuser", user.Name)
	})
}

func TestMockClient_GetCurrentUser(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		user, err := mock.GetCurrentUser(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "U123456", user.ID)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetCurrentUserFunc = func(ctx context.Context) (*domain.SlackUser, error) {
			return &domain.SlackUser{ID: "U99999", Name: "me"}, nil
		}

		user, err := mock.GetCurrentUser(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "me", user.Name)
	})
}

func TestMockClient_SearchMessages(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		msgs, err := mock.SearchMessages(context.Background(), "test query", 10)
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Contains(t, msgs[0].Text, "test query")
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.SearchMessagesFunc = func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
			if query == "important" {
				return []domain.SlackMessage{
					{ID: "1", Text: "Important message"},
				}, nil
			}
			return nil, nil
		}

		msgs, err := mock.SearchMessages(context.Background(), "important", 10)
		require.NoError(t, err)
		assert.Len(t, msgs, 1)

		msgs, err = mock.SearchMessages(context.Background(), "nothing", 10)
		require.NoError(t, err)
		assert.Nil(t, msgs)
	})
}

func TestMockClient_ImplementsInterface(t *testing.T) {
	// This test verifies at compile time that MockClient implements the interface
	mock := NewMockClient()

	ctx := context.Background()

	// Call all methods to verify interface implementation with default behavior
	_, err := mock.TestAuth(ctx)
	assert.NoError(t, err)

	_, err = mock.ListChannels(ctx, nil)
	assert.NoError(t, err)

	_, err = mock.GetChannel(ctx, "C1")
	assert.NoError(t, err)

	_, err = mock.GetMessages(ctx, nil)
	assert.NoError(t, err)

	_, err = mock.GetThreadReplies(ctx, "C1", "ts", 10)
	assert.NoError(t, err)

	_, err = mock.SendMessage(ctx, &domain.SlackSendMessageRequest{ChannelID: "C1", Text: "test"})
	assert.NoError(t, err)

	_, err = mock.UpdateMessage(ctx, "C1", "ts", "text")
	assert.NoError(t, err)

	err = mock.DeleteMessage(ctx, "C1", "ts")
	assert.NoError(t, err)

	_, err = mock.ListUsers(ctx, 100, "")
	assert.NoError(t, err)

	_, err = mock.GetUser(ctx, "U1")
	assert.NoError(t, err)

	_, err = mock.GetCurrentUser(ctx)
	assert.NoError(t, err)

	_, err = mock.SearchMessages(ctx, "query", 10)
	assert.NoError(t, err)
}

func TestMockClient_ErrorScenarios(t *testing.T) {
	mock := NewMockClient()

	// Set up all functions to return errors
	testErr := errors.New("test error")

	mock.TestAuthFunc = func(ctx context.Context) (*domain.SlackAuth, error) {
		return nil, testErr
	}
	mock.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
		return nil, testErr
	}
	mock.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
		return nil, testErr
	}
	mock.GetMessagesFunc = func(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
		return nil, testErr
	}
	mock.GetThreadRepliesFunc = func(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
		return nil, testErr
	}
	mock.SendMessageFunc = func(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
		return nil, testErr
	}
	mock.UpdateMessageFunc = func(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
		return nil, testErr
	}
	mock.DeleteMessageFunc = func(ctx context.Context, channelID, messageTS string) error {
		return testErr
	}
	mock.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
		return nil, testErr
	}
	mock.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
		return nil, testErr
	}
	mock.GetCurrentUserFunc = func(ctx context.Context) (*domain.SlackUser, error) {
		return nil, testErr
	}
	mock.SearchMessagesFunc = func(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
		return nil, testErr
	}

	ctx := context.Background()

	_, err := mock.TestAuth(ctx)
	assert.ErrorIs(t, err, testErr)

	_, err = mock.ListChannels(ctx, nil)
	assert.ErrorIs(t, err, testErr)

	_, err = mock.GetChannel(ctx, "C1")
	assert.ErrorIs(t, err, testErr)

	_, err = mock.GetMessages(ctx, nil)
	assert.ErrorIs(t, err, testErr)

	_, err = mock.GetThreadReplies(ctx, "C1", "ts", 10)
	assert.ErrorIs(t, err, testErr)

	_, err = mock.SendMessage(ctx, &domain.SlackSendMessageRequest{ChannelID: "C1", Text: "test"})
	assert.ErrorIs(t, err, testErr)

	_, err = mock.UpdateMessage(ctx, "C1", "ts", "text")
	assert.ErrorIs(t, err, testErr)

	err = mock.DeleteMessage(ctx, "C1", "ts")
	assert.ErrorIs(t, err, testErr)

	_, err = mock.ListUsers(ctx, 100, "")
	assert.ErrorIs(t, err, testErr)

	_, err = mock.GetUser(ctx, "U1")
	assert.ErrorIs(t, err, testErr)

	_, err = mock.GetCurrentUser(ctx)
	assert.ErrorIs(t, err, testErr)

	_, err = mock.SearchMessages(ctx, "query", 10)
	assert.ErrorIs(t, err, testErr)
}
