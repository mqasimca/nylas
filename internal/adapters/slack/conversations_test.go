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

func TestMockClient_ListChannels_AllScenarios(t *testing.T) {
	tests := []struct {
		name       string
		params     *domain.SlackChannelQueryParams
		setupMock  func(*MockClient)
		wantLen    int
		wantCursor string
		wantErr    error
	}{
		{
			name:    "nil params returns defaults",
			params:  nil,
			wantLen: 2,
		},
		{
			name: "public channels only",
			params: &domain.SlackChannelQueryParams{
				Types: []string{"public_channel"},
			},
			wantLen: 2,
		},
		{
			name: "with limit",
			params: &domain.SlackChannelQueryParams{
				Limit: 50,
			},
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					assert.Equal(t, 50, params.Limit)
					channels := make([]domain.SlackChannel, 50)
					for i := 0; i < 50; i++ {
						channels[i] = domain.SlackChannel{ID: "C" + string(rune('A'+i%26))}
					}
					return &domain.SlackChannelListResponse{
						Channels:   channels,
						NextCursor: "next-page",
					}, nil
				}
			},
			wantLen:    50,
			wantCursor: "next-page",
		},
		{
			name: "exclude archived",
			params: &domain.SlackChannelQueryParams{
				ExcludeArchived: true,
			},
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					assert.True(t, params.ExcludeArchived)
					return &domain.SlackChannelListResponse{
						Channels: []domain.SlackChannel{
							{ID: "C1", Name: "active-channel", IsArchived: false},
						},
					}, nil
				}
			},
			wantLen: 1,
		},
		{
			name: "with cursor pagination",
			params: &domain.SlackChannelQueryParams{
				Cursor: "previous-cursor",
			},
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					assert.Equal(t, "previous-cursor", params.Cursor)
					return &domain.SlackChannelListResponse{
						Channels: []domain.SlackChannel{
							{ID: "C100", Name: "channel-page-2"},
						},
						NextCursor: "",
					}, nil
				}
			},
			wantLen:    1,
			wantCursor: "",
		},
		{
			name: "with team ID",
			params: &domain.SlackChannelQueryParams{
				TeamID: "T12345",
			},
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					assert.Equal(t, "T12345", params.TeamID)
					return &domain.SlackChannelListResponse{
						Channels: []domain.SlackChannel{
							{ID: "C1", Name: "team-channel"},
						},
					}, nil
				}
			},
			wantLen: 1,
		},
		{
			name:   "rate limited",
			params: nil,
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					return nil, domain.ErrSlackRateLimited
				}
			},
			wantErr: domain.ErrSlackRateLimited,
		},
		{
			name:   "auth failed",
			params: nil,
			setupMock: func(m *MockClient) {
				m.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
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

			resp, err := mock.ListChannels(context.Background(), tt.params)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Channels, tt.wantLen)
			assert.Equal(t, tt.wantCursor, resp.NextCursor)
		})
	}
}

func TestMockClient_GetChannel_AllScenarios(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		setupMock func(*MockClient)
		wantName  string
		wantErr   error
	}{
		{
			name:      "existing channel",
			channelID: "C12345",
			wantName:  "general",
		},
		{
			name:      "custom channel",
			channelID: "C99999",
			setupMock: func(m *MockClient) {
				m.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
					return &domain.SlackChannel{
						ID:        channelID,
						Name:      "custom-channel",
						IsChannel: true,
						IsMember:  true,
					}, nil
				}
			},
			wantName: "custom-channel",
		},
		{
			name:      "channel not found",
			channelID: "C_INVALID",
			setupMock: func(m *MockClient) {
				m.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
					return nil, domain.ErrSlackChannelNotFound
				}
			},
			wantErr: domain.ErrSlackChannelNotFound,
		},
		{
			name:      "private channel",
			channelID: "G12345",
			setupMock: func(m *MockClient) {
				m.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
					return &domain.SlackChannel{
						ID:        channelID,
						Name:      "private-channel",
						IsPrivate: true,
					}, nil
				}
			},
			wantName: "private-channel",
		},
		{
			name:      "direct message",
			channelID: "D12345",
			setupMock: func(m *MockClient) {
				m.GetChannelFunc = func(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
					return &domain.SlackChannel{
						ID:   channelID,
						Name: "",
						IsIM: true,
					}, nil
				}
			},
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			ch, err := mock.GetChannel(context.Background(), tt.channelID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, ch)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ch)
			assert.Equal(t, tt.channelID, ch.ID)
			assert.Equal(t, tt.wantName, ch.Name)
		})
	}
}

func TestMockClient_ListMyChannels_AllScenarios(t *testing.T) {
	tests := []struct {
		name       string
		params     *domain.SlackChannelQueryParams
		setupMock  func(*MockClient)
		wantLen    int
		wantMember bool
		wantErr    error
	}{
		{
			name:       "default returns member channels",
			params:     nil,
			wantLen:    2,
			wantMember: true,
		},
		{
			name: "with types filter",
			params: &domain.SlackChannelQueryParams{
				Types: []string{"public_channel"},
			},
			setupMock: func(m *MockClient) {
				m.ListMyChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					assert.Equal(t, []string{"public_channel"}, params.Types)
					return &domain.SlackChannelListResponse{
						Channels: []domain.SlackChannel{
							{ID: "C1", Name: "public", IsMember: true, IsChannel: true},
						},
					}, nil
				}
			},
			wantLen:    1,
			wantMember: true,
		},
		{
			name: "empty result",
			params: &domain.SlackChannelQueryParams{
				Types: []string{"im"},
			},
			setupMock: func(m *MockClient) {
				m.ListMyChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
					return &domain.SlackChannelListResponse{
						Channels: []domain.SlackChannel{},
					}, nil
				}
			},
			wantLen: 0,
		},
		{
			name:   "rate limited",
			params: nil,
			setupMock: func(m *MockClient) {
				m.ListMyChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
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

			resp, err := mock.ListMyChannels(context.Background(), tt.params)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Channels, tt.wantLen)
			if tt.wantLen > 0 && tt.wantMember {
				for _, ch := range resp.Channels {
					assert.True(t, ch.IsMember)
				}
			}
		})
	}
}

func TestMockClient_ChannelTypes(t *testing.T) {
	mock := NewMockClient()

	mock.ListChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
		return &domain.SlackChannelListResponse{
			Channels: []domain.SlackChannel{
				{ID: "C1", Name: "public", IsChannel: true},
				{ID: "G1", Name: "private", IsPrivate: true},
				{ID: "D1", Name: "", IsIM: true},
				{ID: "G2", Name: "group-dm", IsMPIM: true},
			},
		}, nil
	}

	resp, err := mock.ListChannels(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, resp.Channels, 4)

	// Verify channel types
	assert.True(t, resp.Channels[0].IsChannel)
	assert.False(t, resp.Channels[0].IsPrivate)

	assert.True(t, resp.Channels[1].IsPrivate)
	assert.False(t, resp.Channels[1].IsIM)

	assert.True(t, resp.Channels[2].IsIM)
	assert.False(t, resp.Channels[2].IsMPIM)

	assert.True(t, resp.Channels[3].IsMPIM)
	assert.False(t, resp.Channels[3].IsChannel)
}

func TestChannelDisplayName_Variations(t *testing.T) {
	tests := []struct {
		name    string
		channel domain.SlackChannel
		want    string
	}{
		{
			name: "public channel",
			channel: domain.SlackChannel{
				ID:        "C12345",
				Name:      "general",
				IsChannel: true,
			},
			want: "#general",
		},
		{
			name: "private channel",
			channel: domain.SlackChannel{
				ID:        "G12345",
				Name:      "secret",
				IsPrivate: true,
			},
			want: "#secret",
		},
		{
			name: "direct message",
			channel: domain.SlackChannel{
				ID:   "D12345",
				IsIM: true,
			},
			want: "DM",
		},
		{
			name: "group direct message",
			channel: domain.SlackChannel{
				ID:     "G12345",
				Name:   "mpdm-user1--user2--user3",
				IsMPIM: true,
			},
			want: "Group DM",
		},
		{
			name: "channel with empty name",
			channel: domain.SlackChannel{
				ID: "C12345",
			},
			want: "C12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.ChannelDisplayName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChannelType_Variations(t *testing.T) {
	tests := []struct {
		name    string
		channel domain.SlackChannel
		want    string
	}{
		{
			name:    "default is public",
			channel: domain.SlackChannel{},
			want:    "public",
		},
		{
			name:    "public channel",
			channel: domain.SlackChannel{IsChannel: true},
			want:    "public",
		},
		{
			name:    "private channel",
			channel: domain.SlackChannel{IsPrivate: true},
			want:    "private",
		},
		{
			name:    "direct message",
			channel: domain.SlackChannel{IsIM: true},
			want:    "dm",
		},
		{
			name:    "group DM",
			channel: domain.SlackChannel{IsMPIM: true},
			want:    "group_dm",
		},
		{
			name: "DM priority over private",
			channel: domain.SlackChannel{
				IsIM:      true,
				IsPrivate: true,
			},
			want: "dm",
		},
		{
			name: "MPIM priority over private",
			channel: domain.SlackChannel{
				IsMPIM:    true,
				IsPrivate: true,
			},
			want: "group_dm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.ChannelType()
			assert.Equal(t, tt.want, got)
		})
	}
}
