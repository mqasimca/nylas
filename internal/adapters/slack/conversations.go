package slack

import (
	"context"
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

// ResolveChannelByName finds a channel by name.
func (c *Client) ResolveChannelByName(ctx context.Context, name string) (*domain.SlackChannel, error) {
	name = strings.TrimPrefix(name, "#")
	name = strings.ToLower(name)

	cursor := ""
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
