// messages.go provides message operations for Slack channels.

package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/mqasimca/nylas/internal/domain"
)

// GetMessages retrieves messages from a channel.
func (c *Client) GetMessages(ctx context.Context, params *domain.SlackMessageQueryParams) (*domain.SlackMessageListResponse, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if params == nil || params.ChannelID == "" {
		return nil, fmt.Errorf("%w: channel_id is required", domain.ErrSlackChannelNotFound)
	}

	apiParams := &slack.GetConversationHistoryParameters{
		ChannelID: params.ChannelID,
		Limit:     params.Limit,
		Cursor:    params.Cursor,
		Inclusive: params.Inclusive,
	}

	if !params.Oldest.IsZero() {
		apiParams.Oldest = formatTimestamp(params.Oldest)
	}
	if !params.Newest.IsZero() {
		apiParams.Latest = formatTimestamp(params.Newest)
	}

	if apiParams.Limit == 0 {
		apiParams.Limit = 15
	}

	resp, err := c.api.GetConversationHistoryContext(ctx, apiParams)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	messages := make([]domain.SlackMessage, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, convertMessage(msg, params.ChannelID))
	}

	return &domain.SlackMessageListResponse{
		Messages:   messages,
		HasMore:    resp.HasMore,
		NextCursor: resp.ResponseMetaData.NextCursor,
	}, nil
}

// GetThreadReplies retrieves replies in a thread.
func (c *Client) GetThreadReplies(ctx context.Context, channelID, threadTS string, limit int) ([]domain.SlackMessage, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = 15
	}

	msgs, _, _, err := c.api.GetConversationRepliesContext(ctx, &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
		Limit:     limit,
	})
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	replies := make([]domain.SlackMessage, 0, len(msgs))
	for _, msg := range msgs {
		replies = append(replies, convertMessage(msg, channelID))
	}

	return replies, nil
}

// SendMessage sends a new message.
func (c *Client) SendMessage(ctx context.Context, req *domain.SlackSendMessageRequest) (*domain.SlackMessage, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if req == nil || req.ChannelID == "" {
		return nil, fmt.Errorf("%w: channel_id is required", domain.ErrSlackChannelNotFound)
	}

	options := []slack.MsgOption{
		slack.MsgOptionText(req.Text, false),
	}

	if req.ThreadTS != "" {
		options = append(options, slack.MsgOptionTS(req.ThreadTS))
		if req.Broadcast {
			options = append(options, slack.MsgOptionBroadcast())
		}
	}

	channelID, timestamp, _, err := c.api.SendMessageContext(ctx, req.ChannelID, options...)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	return &domain.SlackMessage{
		ID:        timestamp,
		ChannelID: channelID,
		Text:      req.Text,
		Timestamp: parseTimestamp(timestamp),
		ThreadTS:  req.ThreadTS,
		IsReply:   req.ThreadTS != "",
	}, nil
}

// UpdateMessage edits an existing message.
func (c *Client) UpdateMessage(ctx context.Context, channelID, messageTS, newText string) (*domain.SlackMessage, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	_, timestamp, _, err := c.api.UpdateMessageContext(ctx, channelID,
		messageTS,
		slack.MsgOptionText(newText, false),
	)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	return &domain.SlackMessage{
		ID:        timestamp,
		ChannelID: channelID,
		Text:      newText,
		Timestamp: parseTimestamp(timestamp),
		Edited:    true,
	}, nil
}

// DeleteMessage removes a message.
func (c *Client) DeleteMessage(ctx context.Context, channelID, messageTS string) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return err
	}

	_, _, err := c.api.DeleteMessageContext(ctx, channelID, messageTS)
	return c.handleSlackError(err)
}

// SearchMessages searches for messages.
func (c *Client) SearchMessages(ctx context.Context, query string, limit int) ([]domain.SlackMessage, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = 20
	}

	params := slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         limit,
		Highlight:     false,
	}

	resp, err := c.api.SearchMessagesContext(ctx, query, params)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	messages := make([]domain.SlackMessage, 0, len(resp.Matches))
	for _, match := range resp.Matches {
		messages = append(messages, convertSearchMatch(match))
	}

	return messages, nil
}
