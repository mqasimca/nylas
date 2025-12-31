package slack

import (
	"context"

	"github.com/slack-go/slack"

	"github.com/mqasimca/nylas/internal/domain"
)

// ListUsers returns workspace members.
func (c *Client) ListUsers(ctx context.Context, limit int, cursor string) (*domain.SlackUserListResponse, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = 100
	}

	options := []slack.GetUsersOption{
		slack.GetUsersOptionLimit(limit),
	}

	users, err := c.api.GetUsersContext(ctx, options...)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	result := make([]domain.SlackUser, 0, len(users))
	for _, u := range users {
		if u.Deleted {
			continue
		}
		result = append(result, convertUser(u))
	}

	return &domain.SlackUserListResponse{
		Users:      result,
		NextCursor: "", // Pagination handled via GetUsersPaginated if needed
	}, nil
}

// GetUser returns a single user by ID.
func (c *Client) GetUser(ctx context.Context, userID string) (*domain.SlackUser, error) {
	if cached, ok := c.getCachedUser(userID); ok {
		return cached, nil
	}

	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	u, err := c.api.GetUserInfoContext(ctx, userID)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	user := convertUser(*u)
	c.setCachedUser(&user)

	return &user, nil
}

// GetCurrentUser returns the authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*domain.SlackUser, error) {
	auth, err := c.TestAuth(ctx)
	if err != nil {
		return nil, err
	}

	return c.GetUser(ctx, auth.UserID)
}

// GetUsersForMessages enriches messages with usernames.
func (c *Client) GetUsersForMessages(ctx context.Context, messages []domain.SlackMessage) error {
	userIDs := make(map[string]bool)
	for _, msg := range messages {
		if msg.UserID != "" && msg.Username == "" {
			userIDs[msg.UserID] = true
		}
	}

	for userID := range userIDs {
		user, err := c.GetUser(ctx, userID)
		if err != nil {
			continue
		}

		for i := range messages {
			if messages[i].UserID == userID && messages[i].Username == "" {
				messages[i].Username = user.BestDisplayName()
			}
		}
	}

	return nil
}
