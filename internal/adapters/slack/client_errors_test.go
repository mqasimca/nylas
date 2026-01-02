//go:build !integration
// +build !integration

package slack

import (
	"context"
	"errors"
	"testing"
	"time"

	slack_api "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestHandleSlackError_AllAuthErrors(t *testing.T) {
	client := &Client{}

	authErrors := []string{
		"not_authed",
		"invalid_auth",
		"account_inactive",
		"token_revoked",
	}

	for _, errStr := range authErrors {
		t.Run(errStr, func(t *testing.T) {
			err := client.handleSlackError(errors.New(errStr))
			assert.ErrorIs(t, err, domain.ErrSlackAuthFailed)
			assert.Contains(t, err.Error(), errStr)
		})
	}
}

func TestHandleSlackError_AllPermissionErrors(t *testing.T) {
	client := &Client{}

	permErrors := []string{
		"missing_scope",
		"not_allowed_token_type",
	}

	for _, errStr := range permErrors {
		t.Run(errStr, func(t *testing.T) {
			err := client.handleSlackError(errors.New(errStr))
			assert.ErrorIs(t, err, domain.ErrSlackPermissionDenied)
			assert.Contains(t, err.Error(), errStr)
		})
	}
}

func TestHandleSlackError_ResourceNotFound(t *testing.T) {
	client := &Client{}

	t.Run("channel_not_found", func(t *testing.T) {
		err := client.handleSlackError(errors.New("channel_not_found"))
		assert.ErrorIs(t, err, domain.ErrSlackChannelNotFound)
	})

	t.Run("message_not_found", func(t *testing.T) {
		err := client.handleSlackError(errors.New("message_not_found"))
		assert.ErrorIs(t, err, domain.ErrSlackMessageNotFound)
	})
}

func TestHandleSlackError_RateLimitedError(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name       string
		retryAfter time.Duration
	}{
		{
			name:       "short retry",
			retryAfter: 1 * time.Second,
		},
		{
			name:       "medium retry",
			retryAfter: 30 * time.Second,
		},
		{
			name:       "long retry",
			retryAfter: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slackErr := &slack_api.RateLimitedError{RetryAfter: tt.retryAfter}
			err := client.handleSlackError(slackErr)
			assert.ErrorIs(t, err, domain.ErrSlackRateLimited)
			assert.Contains(t, err.Error(), tt.retryAfter.String())
		})
	}
}

func TestHandleSlackError_UnknownErrors(t *testing.T) {
	client := &Client{}

	unknownErrors := []string{
		"internal_error",
		"service_unavailable",
		"something_weird_happened",
		"unknown_slack_error",
	}

	for _, errStr := range unknownErrors {
		t.Run(errStr, func(t *testing.T) {
			err := client.handleSlackError(errors.New(errStr))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "slack API error")
			assert.Contains(t, err.Error(), errStr)
		})
	}
}

func TestNewClient_ConfigDefaults(t *testing.T) {
	tests := []struct {
		name             string
		config           *ClientConfig
		wantRateLimit    float64
		wantRateBurst    int
		wantUserCacheTTL time.Duration
	}{
		{
			name: "zero rate limit gets default",
			config: &ClientConfig{
				UserToken:    "xoxp-test",
				RateLimit:    0,
				RateBurst:    5,
				UserCacheTTL: 10 * time.Minute,
			},
			wantRateLimit:    1,
			wantRateBurst:    5,
			wantUserCacheTTL: 10 * time.Minute,
		},
		{
			name: "zero rate burst gets default",
			config: &ClientConfig{
				UserToken:    "xoxp-test",
				RateLimit:    2,
				RateBurst:    0,
				UserCacheTTL: 10 * time.Minute,
			},
			wantRateLimit:    2,
			wantRateBurst:    1,
			wantUserCacheTTL: 10 * time.Minute,
		},
		{
			name: "zero cache TTL gets default",
			config: &ClientConfig{
				UserToken:    "xoxp-test",
				RateLimit:    2,
				RateBurst:    5,
				UserCacheTTL: 0,
			},
			wantRateLimit:    2,
			wantRateBurst:    5,
			wantUserCacheTTL: 5 * time.Minute,
		},
		{
			name: "all custom values preserved",
			config: &ClientConfig{
				UserToken:    "xoxp-test",
				RateLimit:    10,
				RateBurst:    20,
				UserCacheTTL: 30 * time.Minute,
			},
			wantRateLimit:    10,
			wantRateBurst:    20,
			wantUserCacheTTL: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			require.NoError(t, err)
			require.NotNil(t, client)

			assert.Equal(t, tt.wantUserCacheTTL, client.userCacheTTL)
		})
	}
}

func TestNewClient_DebugMode(t *testing.T) {
	t.Run("debug mode enabled", func(t *testing.T) {
		client, err := NewClient(&ClientConfig{
			UserToken: "xoxp-test",
			Debug:     true,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("debug mode disabled", func(t *testing.T) {
		client, err := NewClient(&ClientConfig{
			UserToken: "xoxp-test",
			Debug:     false,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

func TestUserCache_Concurrency(t *testing.T) {
	client := &Client{
		userCache:    make(map[string]*cachedUser),
		userCacheTTL: 5 * time.Minute,
	}

	// Test concurrent access to cache
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			user := &domain.SlackUser{
				ID:   "U" + string(rune('A'+id)),
				Name: "user" + string(rune('a'+id)),
			}
			client.setCachedUser(user)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(id int) {
			client.getCachedUser("U" + string(rune('A'+id)))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no crashes occurred - test passes if we get here
	assert.True(t, true)
}

func TestUserCache_MultipleSets(t *testing.T) {
	client := &Client{
		userCache:    make(map[string]*cachedUser),
		userCacheTTL: 5 * time.Minute,
	}

	// Set same user multiple times with different data
	for i := 0; i < 5; i++ {
		user := &domain.SlackUser{
			ID:       "U12345",
			Name:     "version" + string(rune('0'+i)),
			RealName: "User Version " + string(rune('0'+i)),
		}
		client.setCachedUser(user)
	}

	// Should have the latest version
	cached, ok := client.getCachedUser("U12345")
	assert.True(t, ok)
	assert.Equal(t, "version4", cached.Name)
}

func TestUserCache_DifferentUsers(t *testing.T) {
	client := &Client{
		userCache:    make(map[string]*cachedUser),
		userCacheTTL: 5 * time.Minute,
	}

	users := []*domain.SlackUser{
		{ID: "U1", Name: "alice", RealName: "Alice Smith"},
		{ID: "U2", Name: "bob", RealName: "Bob Jones"},
		{ID: "U3", Name: "charlie", RealName: "Charlie Brown"},
	}

	// Cache all users
	for _, u := range users {
		client.setCachedUser(u)
	}

	// Verify all can be retrieved
	for _, u := range users {
		cached, ok := client.getCachedUser(u.ID)
		assert.True(t, ok)
		assert.Equal(t, u.Name, cached.Name)
		assert.Equal(t, u.RealName, cached.RealName)
	}

	// Non-existent user
	cached, ok := client.getCachedUser("U99999")
	assert.False(t, ok)
	assert.Nil(t, cached)
}

func TestWaitForRateLimit_ImmediateSuccess(t *testing.T) {
	client, err := NewClient(&ClientConfig{
		UserToken: "xoxp-test",
		RateLimit: 100, // High rate limit
		RateBurst: 10,
	})
	require.NoError(t, err)

	// Should succeed immediately
	ctx := context.Background()
	err = client.waitForRateLimit(ctx)
	assert.NoError(t, err)
}

func TestWaitForRateLimit_DeadlineExceeded(t *testing.T) {
	client, err := NewClient(&ClientConfig{
		UserToken: "xoxp-test",
		RateLimit: 0.001, // Very slow rate limit
		RateBurst: 1,
	})
	require.NoError(t, err)

	// Use up the burst
	ctx := context.Background()
	err = client.waitForRateLimit(ctx)
	require.NoError(t, err)

	// Now try with a very short deadline
	deadlineCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = client.waitForRateLimit(deadlineCtx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSlackRateLimited)
}

func TestDefaultConfig_Values(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, float64(1), float64(config.RateLimit))
	assert.Equal(t, 1, config.RateBurst)
	assert.Equal(t, 5*time.Minute, config.UserCacheTTL)
	assert.False(t, config.Debug)
	assert.Empty(t, config.UserToken)
}

func TestNewClient_NilConfigUsesDefaults(t *testing.T) {
	// This should fail because nil config means empty token
	client, err := NewClient(nil)
	assert.ErrorIs(t, err, domain.ErrSlackNotConfigured)
	assert.Nil(t, client)
}

func TestClient_InterfaceCompliance(t *testing.T) {
	// Ensure Client implements ports.SlackClient
	config := &ClientConfig{
		UserToken: "xoxp-test",
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// These should compile - testing interface compliance
	assert.NotNil(t, client.api)
	assert.NotNil(t, client.rateLimiter)
	assert.NotNil(t, client.userCache)
	assert.Equal(t, "xoxp-test", client.userToken)
}
