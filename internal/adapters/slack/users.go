// users.go provides user management operations for Slack workspaces.

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

	// Use paginated API to respect limit - fetch only needed pages.
	// Slack API enforces max 200 users per request.
	pageSize := min(limit, 200)

	options := []slack.GetUsersOption{
		slack.GetUsersOptionLimit(pageSize),
	}

	pager := c.api.GetUsersPaginated(options...)
	result := make([]domain.SlackUser, 0, limit)
	hasMore := false

	for {
		var err error
		pager, err = pager.Next(ctx)
		if err != nil {
			// pager.Done() returns true when pagination is complete (not an error)
			if pager.Done(err) {
				break
			}
			return nil, c.handleSlackError(err)
		}

		for _, u := range pager.Users {
			// Skip deactivated users - they can't be messaged or mentioned
			if u.Deleted {
				continue
			}
			result = append(result, convertUser(u))
			if len(result) >= limit {
				hasMore = true
				break
			}
		}

		if len(result) >= limit {
			break
		}
	}

	// Note: We use a placeholder cursor value since the paginated API doesn't expose
	// the actual cursor. This signals to callers that more results exist, but they
	// should increase the limit rather than paginate with this cursor.
	nextCursor := ""
	if hasMore {
		nextCursor = "more"
	}

	return &domain.SlackUserListResponse{
		Users:      result,
		NextCursor: nextCursor,
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

// GetUsersForMessages enriches messages with usernames by looking up user IDs.
// It modifies the messages slice in-place, updating Username fields for messages
// where UserID is present but Username is empty.
func (c *Client) GetUsersForMessages(ctx context.Context, messages []domain.SlackMessage) error {
	// Collect unique user IDs to minimize API calls (one call per user, not per message)
	userIDs := make(map[string]bool)
	for _, msg := range messages {
		if msg.UserID != "" && msg.Username == "" {
			userIDs[msg.UserID] = true
		}
	}

	for userID := range userIDs {
		user, err := c.GetUser(ctx, userID)
		if err != nil {
			// Continue on error - partial enrichment is better than failing entirely.
			// Common errors: deleted users, rate limits, or permission issues.
			continue
		}

		// Apply username to all messages from this user
		for i := range messages {
			if messages[i].UserID == userID && messages[i].Username == "" {
				messages[i].Username = user.BestDisplayName()
			}
		}
	}

	return nil
}
