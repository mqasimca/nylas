// conversations.go provides channel and conversation operations for Slack.

package slack

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"

	"github.com/mqasimca/nylas/internal/domain"
)

// ListChannels returns all accessible channels.
func (c *Client) ListChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if params == nil {
		params = &domain.SlackChannelQueryParams{}
	}

	types := params.Types
	if len(types) == 0 {
		types = []string{"public_channel", "private_channel", "mpim", "im"}
	}

	limit := params.Limit
	if limit == 0 {
		limit = 100
	}

	apiParams := &slack.GetConversationsParameters{
		Types:           types,
		ExcludeArchived: params.ExcludeArchived,
		Limit:           limit,
		Cursor:          params.Cursor,
		TeamID:          params.TeamID,
	}

	channels, nextCursor, err := c.api.GetConversationsContext(ctx, apiParams)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	result := make([]domain.SlackChannel, 0, len(channels))
	for _, ch := range channels {
		result = append(result, convertChannel(ch))
	}

	return &domain.SlackChannelListResponse{
		Channels:   result,
		NextCursor: nextCursor,
	}, nil
}

// GetChannel returns a single channel by ID.
func (c *Client) GetChannel(ctx context.Context, channelID string) (*domain.SlackChannel, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	ch, err := c.api.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	channel := convertChannel(*ch)
	return &channel, nil
}

// ListMyChannels returns only channels the current user is a member of.
// This is much faster than ListChannels for workspaces with many channels.
func (c *Client) ListMyChannels(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if params == nil {
		params = &domain.SlackChannelQueryParams{}
	}

	types := params.Types
	if len(types) == 0 {
		types = []string{"public_channel", "private_channel"}
	}

	limit := params.Limit
	if limit == 0 {
		limit = 200
	}

	apiParams := &slack.GetConversationsForUserParameters{
		Types:           types,
		ExcludeArchived: params.ExcludeArchived,
		Limit:           limit,
		Cursor:          params.Cursor,
		TeamID:          params.TeamID,
	}

	channels, nextCursor, err := c.api.GetConversationsForUserContext(ctx, apiParams)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	result := make([]domain.SlackChannel, 0, len(channels))
	for _, ch := range channels {
		result = append(result, convertChannel(ch))
	}

	return &domain.SlackChannelListResponse{
		Channels:   result,
		NextCursor: nextCursor,
	}, nil
}

// ResolveChannelByName finds a channel by name (case-insensitive).
// Note: This function iterates through all channels until a match is found,
// which may be slow for large workspaces. Consider using channel IDs directly
// when possible.
func (c *Client) ResolveChannelByName(ctx context.Context, name string) (*domain.SlackChannel, error) {
	name = strings.TrimPrefix(name, "#")
	name = strings.ToLower(name)

	cursor := ""
	channelsSearched := 0
	warnedLargeWorkspace := false
	for {
		resp, err := c.ListChannels(ctx, &domain.SlackChannelQueryParams{
			Types:           []string{"public_channel", "private_channel"},
			ExcludeArchived: true,
			Limit:           200,
			Cursor:          cursor,
		})
		if err != nil {
			return nil, err
		}

		channelsSearched += len(resp.Channels)

		// Warn once when searching through many channels
		if channelsSearched > 500 && !warnedLargeWorkspace {
			warnedLargeWorkspace = true
			fmt.Fprintf(os.Stderr, "Warning: Large workspace detected (%d+ channels). Consider using channel ID instead of name for better performance.\n", channelsSearched)
		}

		for _, ch := range resp.Channels {
			if strings.ToLower(ch.Name) == name {
				return &ch, nil
			}
		}

		if resp.NextCursor == "" {
			break
		}
		cursor = resp.NextCursor
	}

	return nil, domain.ErrSlackChannelNotFound
}
