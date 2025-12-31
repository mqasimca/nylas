//go:build !integration
// +build !integration

package slack

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, float64(1), float64(config.RateLimit))
	assert.Equal(t, 1, config.RateBurst)
	assert.Equal(t, 5*time.Minute, config.UserCacheTTL)
	assert.False(t, config.Debug)
	assert.Empty(t, config.UserToken)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *ClientConfig
		wantErr error
	}{
		{
			name:    "nil config returns error",
			config:  nil,
			wantErr: domain.ErrSlackNotConfigured,
		},
		{
			name:    "empty token returns error",
			config:  &ClientConfig{},
			wantErr: domain.ErrSlackNotConfigured,
		},
		{
			name: "valid config creates client",
			config: &ClientConfig{
				UserToken: "xoxp-test-token",
			},
			wantErr: nil,
		},
		{
			name: "config with defaults applied",
			config: &ClientConfig{
				UserToken: "xoxp-test-token",
			},
			wantErr: nil,
		},
		{
			name: "config with custom values",
			config: &ClientConfig{
				UserToken:    "xoxp-test-token",
				Debug:        true,
				RateLimit:    2,
				RateBurst:    5,
				UserCacheTTL: 10 * time.Minute,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.api)
				assert.NotNil(t, client.rateLimiter)
				assert.NotNil(t, client.userCache)
			}
		})
	}
}

func TestNewClient_AppliesDefaults(t *testing.T) {
	config := &ClientConfig{
		UserToken: "xoxp-test-token",
		// Leave other fields at zero values
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	// Defaults should be applied
	assert.Equal(t, DefaultConfig().UserCacheTTL, client.userCacheTTL)
}

func TestHandleSlackError(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name        string
		err         error
		wantErrType error
		wantNil     bool
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:        "channel_not_found",
			err:         errors.New("channel_not_found"),
			wantErrType: domain.ErrSlackChannelNotFound,
		},
		{
			name:        "message_not_found",
			err:         errors.New("message_not_found"),
			wantErrType: domain.ErrSlackMessageNotFound,
		},
		{
			name:        "not_authed",
			err:         errors.New("not_authed"),
			wantErrType: domain.ErrSlackAuthFailed,
		},
		{
			name:        "invalid_auth",
			err:         errors.New("invalid_auth"),
			wantErrType: domain.ErrSlackAuthFailed,
		},
		{
			name:        "account_inactive",
			err:         errors.New("account_inactive"),
			wantErrType: domain.ErrSlackAuthFailed,
		},
		{
			name:        "token_revoked",
			err:         errors.New("token_revoked"),
			wantErrType: domain.ErrSlackAuthFailed,
		},
		{
			name:        "missing_scope",
			err:         errors.New("missing_scope"),
			wantErrType: domain.ErrSlackPermissionDenied,
		},
		{
			name:        "not_allowed_token_type",
			err:         errors.New("not_allowed_token_type"),
			wantErrType: domain.ErrSlackPermissionDenied,
		},
		{
			name:        "rate limited error",
			err:         &slack.RateLimitedError{RetryAfter: 30 * time.Second},
			wantErrType: domain.ErrSlackRateLimited,
		},
		{
			name:        "unknown error wrapped",
			err:         errors.New("some_unknown_error"),
			wantErrType: nil, // Will be wrapped as generic slack API error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.handleSlackError(tt.err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if tt.wantErrType != nil {
				assert.ErrorIs(t, got, tt.wantErrType)
			} else {
				assert.Contains(t, got.Error(), "slack API error")
			}
		})
	}
}

func TestUserCache(t *testing.T) {
	client := &Client{
		userCache:    make(map[string]*domain.SlackUser),
		userCacheTTL: 5 * time.Minute,
	}

	t.Run("get from empty cache returns not found", func(t *testing.T) {
		user, ok := client.getCachedUser("U12345")
		assert.False(t, ok)
		assert.Nil(t, user)
	})

	t.Run("set and get user", func(t *testing.T) {
		testUser := &domain.SlackUser{
			ID:       "U12345",
			Name:     "testuser",
			RealName: "Test User",
		}

		client.setCachedUser(testUser)

		user, ok := client.getCachedUser("U12345")
		assert.True(t, ok)
		assert.Equal(t, testUser, user)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		user, ok := client.getCachedUser("U99999")
		assert.False(t, ok)
		assert.Nil(t, user)
	})

	t.Run("overwrite cached user", func(t *testing.T) {
		testUser := &domain.SlackUser{
			ID:       "U12345",
			Name:     "updated",
			RealName: "Updated User",
		}

		client.setCachedUser(testUser)

		user, ok := client.getCachedUser("U12345")
		assert.True(t, ok)
		assert.Equal(t, "updated", user.Name)
	})
}

func TestWaitForRateLimit_ContextCanceled(t *testing.T) {
	client, err := NewClient(&ClientConfig{
		UserToken: "xoxp-test",
		RateLimit: 0.01, // Very slow rate to force waiting
		RateBurst: 1,
	})
	require.NoError(t, err)

	// Use up the burst
	ctx := context.Background()
	err = client.waitForRateLimit(ctx)
	require.NoError(t, err)

	// Now cancel the context while waiting
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.waitForRateLimit(cancelCtx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSlackRateLimited)
}
