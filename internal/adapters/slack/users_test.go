//go:build !integration
// +build !integration

package slack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestMockClient_ListUsers_AllScenarios(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		cursor     string
		setupMock  func(*MockClient)
		wantLen    int
		wantCursor string
		wantErr    error
	}{
		{
			name:    "default returns users",
			limit:   100,
			cursor:  "",
			wantLen: 2,
		},
		{
			name:   "with custom limit",
			limit:  5,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					assert.Equal(t, 5, limit)
					users := make([]domain.SlackUser, 5)
					for i := 0; i < 5; i++ {
						users[i] = domain.SlackUser{ID: "U" + string(rune('A'+i))}
					}
					return &domain.SlackUserListResponse{
						Users:      users,
						NextCursor: "more",
					}, nil
				}
			},
			wantLen:    5,
			wantCursor: "more",
		},
		{
			name:   "with pagination cursor",
			limit:  100,
			cursor: "dXNlcjpVMDYxTkZUVDI=",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					assert.Equal(t, "dXNlcjpVMDYxTkZUVDI=", cursor)
					return &domain.SlackUserListResponse{
						Users: []domain.SlackUser{
							{ID: "U100", Name: "user100"},
						},
					}, nil
				}
			},
			wantLen: 1,
		},
		{
			name:   "includes bots",
			limit:  100,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					return &domain.SlackUserListResponse{
						Users: []domain.SlackUser{
							{ID: "U1", Name: "human", IsBot: false},
							{ID: "B1", Name: "slackbot", IsBot: true},
						},
					}, nil
				}
			},
			wantLen: 2,
		},
		{
			name:   "includes admins",
			limit:  100,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					return &domain.SlackUserListResponse{
						Users: []domain.SlackUser{
							{ID: "U1", Name: "regular", IsAdmin: false},
							{ID: "U2", Name: "admin", IsAdmin: true},
						},
					}, nil
				}
			},
			wantLen: 2,
		},
		{
			name:   "rate limited",
			limit:  100,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					return nil, domain.ErrSlackRateLimited
				}
			},
			wantErr: domain.ErrSlackRateLimited,
		},
		{
			name:   "auth failed",
			limit:  100,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					return nil, domain.ErrSlackAuthFailed
				}
			},
			wantErr: domain.ErrSlackAuthFailed,
		},
		{
			name:   "empty workspace",
			limit:  100,
			cursor: "",
			setupMock: func(m *MockClient) {
				m.ListUsersFunc = func(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
					return &domain.SlackUserListResponse{
						Users: []domain.SlackUser{},
					}, nil
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

			resp, err := mock.ListUsers(context.Background(), tt.limit, tt.cursor)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Users, tt.wantLen)
			assert.Equal(t, tt.wantCursor, resp.NextCursor)
		})
	}
}

func TestMockClient_GetUser_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupMock func(*MockClient)
		wantName  string
		wantErr   error
	}{
		{
			name:     "existing user",
			userID:   "U12345",
			wantName: "testuser",
		},
		{
			name:   "custom user",
			userID: "U99999",
			setupMock: func(m *MockClient) {
				m.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
					return &domain.SlackUser{
						ID:          userID,
						Name:        "customuser",
						RealName:    "Custom User",
						DisplayName: "Custom",
						Email:       "custom@example.com",
					}, nil
				}
			},
			wantName: "customuser",
		},
		{
			name:   "bot user",
			userID: "B12345",
			setupMock: func(m *MockClient) {
				m.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
					return &domain.SlackUser{
						ID:    userID,
						Name:  "slackbot",
						IsBot: true,
					}, nil
				}
			},
			wantName: "slackbot",
		},
		{
			name:   "user not found",
			userID: "U_INVALID",
			setupMock: func(m *MockClient) {
				m.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
					return nil, domain.ErrSlackAuthFailed
				}
			},
			wantErr: domain.ErrSlackAuthFailed,
		},
		{
			name:   "rate limited",
			userID: "U12345",
			setupMock: func(m *MockClient) {
				m.GetUserFunc = func(ctx context.Context, userID string) (*domain.SlackUser, error) {
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

			user, err := mock.GetUser(context.Background(), tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)
			assert.Equal(t, tt.userID, user.ID)
			assert.Equal(t, tt.wantName, user.Name)
		})
	}
}

func TestMockClient_GetCurrentUser_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockClient)
		wantID    string
		wantName  string
		wantErr   error
	}{
		{
			name:     "default current user",
			wantID:   "U123456",
			wantName: "testuser",
		},
		{
			name: "custom current user",
			setupMock: func(m *MockClient) {
				m.GetCurrentUserFunc = func(ctx context.Context) (*domain.SlackUser, error) {
					return &domain.SlackUser{
						ID:       "U99999",
						Name:     "me",
						RealName: "Current User",
						IsAdmin:  true,
					}, nil
				}
			},
			wantID:   "U99999",
			wantName: "me",
		},
		{
			name: "auth failed",
			setupMock: func(m *MockClient) {
				m.GetCurrentUserFunc = func(ctx context.Context) (*domain.SlackUser, error) {
					return nil, domain.ErrSlackAuthFailed
				}
			},
			wantErr: domain.ErrSlackAuthFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			user, err := mock.GetCurrentUser(context.Background())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)
			assert.Equal(t, tt.wantID, user.ID)
			assert.Equal(t, tt.wantName, user.Name)
		})
	}
}

func TestSlackUser_BestDisplayName_AllCases(t *testing.T) {
	tests := []struct {
		name string
		user domain.SlackUser
		want string
	}{
		{
			name: "all names set - prefers display name",
			user: domain.SlackUser{
				Name:        "handle",
				RealName:    "Real Name",
				DisplayName: "Display Name",
			},
			want: "Display Name",
		},
		{
			name: "no display name - uses real name",
			user: domain.SlackUser{
				Name:        "handle",
				RealName:    "Real Name",
				DisplayName: "",
			},
			want: "Real Name",
		},
		{
			name: "only handle - uses handle",
			user: domain.SlackUser{
				Name:        "handle",
				RealName:    "",
				DisplayName: "",
			},
			want: "handle",
		},
		{
			name: "whitespace display name treated as set",
			user: domain.SlackUser{
				Name:        "handle",
				RealName:    "Real Name",
				DisplayName: " ",
			},
			want: " ",
		},
		{
			name: "all empty",
			user: domain.SlackUser{
				Name:        "",
				RealName:    "",
				DisplayName: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.BestDisplayName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackUser_Fields(t *testing.T) {
	user := domain.SlackUser{
		ID:          "U12345",
		Name:        "jsmith",
		RealName:    "John Smith",
		DisplayName: "Johnny",
		Email:       "john@example.com",
		Avatar:      "https://example.com/avatar.png",
		IsBot:       false,
		IsAdmin:     true,
		Status:      "In a meeting",
		Timezone:    "America/New_York",
	}

	assert.Equal(t, "U12345", user.ID)
	assert.Equal(t, "jsmith", user.Name)
	assert.Equal(t, "John Smith", user.RealName)
	assert.Equal(t, "Johnny", user.DisplayName)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "https://example.com/avatar.png", user.Avatar)
	assert.False(t, user.IsBot)
	assert.True(t, user.IsAdmin)
	assert.Equal(t, "In a meeting", user.Status)
	assert.Equal(t, "America/New_York", user.Timezone)
}

func TestSlackUser_BotUser(t *testing.T) {
	bot := domain.SlackUser{
		ID:    "B12345",
		Name:  "slackbot",
		IsBot: true,
	}

	assert.True(t, bot.IsBot)
	assert.False(t, bot.IsAdmin)
	assert.Equal(t, "slackbot", bot.BestDisplayName())
}

func TestSlackUser_AdminUser(t *testing.T) {
	admin := domain.SlackUser{
		ID:      "U12345",
		Name:    "admin",
		IsAdmin: true,
		IsBot:   false,
	}

	assert.True(t, admin.IsAdmin)
	assert.False(t, admin.IsBot)
}
