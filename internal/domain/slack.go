// Package domain contains the core business logic and domain models.
package domain

import "time"

// SlackMessage represents a message in Slack.
type SlackMessage struct {
	ID          string            `json:"id"`
	ChannelID   string            `json:"channel_id"`
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Text        string            `json:"text"`
	Timestamp   time.Time         `json:"timestamp"`
	ThreadTS    string            `json:"thread_ts,omitempty"`
	ReplyCount  int               `json:"reply_count,omitempty"`
	IsReply     bool              `json:"is_reply"`
	Edited      bool              `json:"edited"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Reactions   []SlackReaction   `json:"reactions,omitempty"`
}

// SlackChannel represents a Slack channel, DM, or group DM.
type SlackChannel struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	IsChannel    bool      `json:"is_channel"`
	IsGroup      bool      `json:"is_group"`
	IsIM         bool      `json:"is_im"`
	IsMPIM       bool      `json:"is_mpim"`
	IsPrivate    bool      `json:"is_private"`
	IsArchived   bool      `json:"is_archived"`
	Topic        string    `json:"topic,omitempty"`
	Purpose      string    `json:"purpose,omitempty"`
	MemberCount  int       `json:"member_count"`
	Created      time.Time `json:"created"`
	LastActivity time.Time `json:"last_activity,omitempty"`
}

// SlackUser represents a Slack workspace member.
type SlackUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RealName    string `json:"real_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	IsBot       bool   `json:"is_bot"`
	IsAdmin     bool   `json:"is_admin"`
	Status      string `json:"status,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

// SlackAttachment represents a message attachment.
type SlackAttachment struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// SlackReaction represents an emoji reaction.
type SlackReaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// SlackMessageQueryParams for filtering messages.
type SlackMessageQueryParams struct {
	ChannelID string    `json:"channel_id"`
	Limit     int       `json:"limit,omitempty"`
	Cursor    string    `json:"cursor,omitempty"`
	Oldest    time.Time `json:"oldest,omitempty"`
	Newest    time.Time `json:"newest,omitempty"`
	Inclusive bool      `json:"inclusive,omitempty"`
}

// SlackSendMessageRequest for sending messages.
type SlackSendMessageRequest struct {
	ChannelID string `json:"channel_id"`
	Text      string `json:"text"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	Broadcast bool   `json:"reply_broadcast,omitempty"`
}

// SlackChannelQueryParams for filtering channels.
type SlackChannelQueryParams struct {
	Types           []string `json:"types,omitempty"`
	ExcludeArchived bool     `json:"exclude_archived,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	Cursor          string   `json:"cursor,omitempty"`
}

// SlackMessageListResponse with pagination.
type SlackMessageListResponse struct {
	Messages   []SlackMessage `json:"messages"`
	HasMore    bool           `json:"has_more"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// SlackChannelListResponse with pagination.
type SlackChannelListResponse struct {
	Channels   []SlackChannel `json:"channels"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// SlackUserListResponse with pagination.
type SlackUserListResponse struct {
	Users      []SlackUser `json:"users"`
	NextCursor string      `json:"next_cursor,omitempty"`
}

// SlackAuth stores authentication info.
type SlackAuth struct {
	UserID    string `json:"user_id"`
	TeamID    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email,omitempty"`
}

// BestDisplayName returns the best display name for a user.
func (u SlackUser) BestDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.RealName != "" {
		return u.RealName
	}
	return u.Name
}

// ChannelDisplayName returns the best display name for a channel.
func (c SlackChannel) ChannelDisplayName() string {
	if c.IsIM {
		return "DM"
	}
	if c.IsMPIM {
		return "Group DM"
	}
	if c.Name != "" {
		return "#" + c.Name
	}
	return c.ID
}

// IsThread returns true if the message is part of a thread.
func (m SlackMessage) IsThread() bool {
	return m.ThreadTS != ""
}

// ChannelType returns the type of channel as a string.
func (c SlackChannel) ChannelType() string {
	switch {
	case c.IsIM:
		return "dm"
	case c.IsMPIM:
		return "group_dm"
	case c.IsPrivate:
		return "private"
	default:
		return "public"
	}
}
