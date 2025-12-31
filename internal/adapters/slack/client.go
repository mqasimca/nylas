// Package slack provides a Slack API client adapter.
package slack

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"golang.org/x/time/rate"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// Ensure Client implements ports.SlackClient.
var _ ports.SlackClient = (*Client)(nil)

// Client wraps the slack-go client with rate limiting.
type Client struct {
	api         *slack.Client
	userToken   string
	rateLimiter *rate.Limiter

	// Cache for user lookups (reduce API calls).
	userCache    map[string]*domain.SlackUser
	userCacheMu  sync.RWMutex
	userCacheTTL time.Duration
}

// ClientConfig holds configuration for the Slack client.
type ClientConfig struct {
	UserToken    string
	Debug        bool
	RateLimit    rate.Limit
	RateBurst    int
	UserCacheTTL time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		RateLimit:    rate.Limit(1),
		RateBurst:    1,
		UserCacheTTL: 5 * time.Minute,
		Debug:        false,
	}
}

// NewClient creates a new Slack client.
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.UserToken == "" {
		return nil, domain.ErrSlackNotConfigured
	}

	if config.RateLimit == 0 {
		config.RateLimit = DefaultConfig().RateLimit
	}
	if config.RateBurst == 0 {
		config.RateBurst = DefaultConfig().RateBurst
	}
	if config.UserCacheTTL == 0 {
		config.UserCacheTTL = DefaultConfig().UserCacheTTL
	}

	options := []slack.Option{}
	if config.Debug {
		options = append(options, slack.OptionDebug(true))
	}

	api := slack.New(config.UserToken, options...)

	return &Client{
		api:          api,
		userToken:    config.UserToken,
		rateLimiter:  rate.NewLimiter(config.RateLimit, config.RateBurst),
		userCache:    make(map[string]*domain.SlackUser),
		userCacheTTL: config.UserCacheTTL,
	}, nil
}

// waitForRateLimit blocks until rate limit allows.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrSlackRateLimited, err)
	}
	return nil
}

// handleSlackError converts Slack errors to domain errors.
func (c *Client) handleSlackError(err error) error {
	if err == nil {
		return nil
	}

	if rateLimitErr, ok := err.(*slack.RateLimitedError); ok {
		return fmt.Errorf("%w: retry after %v", domain.ErrSlackRateLimited, rateLimitErr.RetryAfter)
	}

	errStr := err.Error()
	switch errStr {
	case "channel_not_found":
		return domain.ErrSlackChannelNotFound
	case "message_not_found":
		return domain.ErrSlackMessageNotFound
	case "not_authed", "invalid_auth", "account_inactive", "token_revoked":
		return fmt.Errorf("%w: %s", domain.ErrSlackAuthFailed, errStr)
	case "missing_scope", "not_allowed_token_type":
		return fmt.Errorf("%w: %s", domain.ErrSlackPermissionDenied, errStr)
	}

	return fmt.Errorf("slack API error: %w", err)
}

// TestAuth validates the token and returns auth info.
func (c *Client) TestAuth(ctx context.Context) (*domain.SlackAuth, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	resp, err := c.api.AuthTestContext(ctx)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	return &domain.SlackAuth{
		UserID:   resp.UserID,
		TeamID:   resp.TeamID,
		TeamName: resp.Team,
		UserName: resp.User,
	}, nil
}

// getCachedUser returns a cached user if available and not expired.
func (c *Client) getCachedUser(userID string) (*domain.SlackUser, bool) {
	c.userCacheMu.RLock()
	defer c.userCacheMu.RUnlock()

	user, ok := c.userCache[userID]
	return user, ok
}

// setCachedUser stores a user in the cache.
func (c *Client) setCachedUser(user *domain.SlackUser) {
	c.userCacheMu.Lock()
	defer c.userCacheMu.Unlock()

	c.userCache[user.ID] = user
}
