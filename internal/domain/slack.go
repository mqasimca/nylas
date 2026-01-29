// Package domain contains the core business logic and domain models.
package domain

import "time"

// SlackMessage represents a message in Slack.
type SlackMessage struct {
	ID          string            `json:"id"`                    // Message timestamp (ts) - unique identifier within the channel
	ChannelID   string            `json:"channel_id"`            // Channel where the message was posted
	UserID      string            `json:"user_id"`               // Slack user ID of the sender (e.g., "U1234567890")
	Username    string            `json:"username"`              // Display name of the sender (resolved from user profile)
	Text        string            `json:"text"`                  // Message content (may contain Slack markup)
	Timestamp   time.Time         `json:"timestamp"`             // When the message was sent
	ThreadTS    string            `json:"thread_ts,omitempty"`   // Thread timestamp - unique identifier for the parent message of a thread
	ReplyCount  int               `json:"reply_count,omitempty"` // Number of replies in the thread
	IsReply     bool              `json:"is_reply"`              // True if this message is a reply within a thread
	Edited      bool              `json:"edited"`                // True if the message has been edited
	Attachments []SlackAttachment `json:"attachments,omitempty"` // Files and media attached to the message
	Reactions   []SlackReaction   `json:"reactions,omitempty"`   // Emoji reactions on the message
}

// SlackChannel represents a Slack channel, DM, or group DM.
type SlackChannel struct {
	ID           string    `json:"id"`                      // Channel ID (e.g., "C1234567890" for public, "G..." for private)
	Name         string    `json:"name"`                    // Channel name without # prefix (empty for DMs)
	IsChannel    bool      `json:"is_channel"`              // True if this is a public channel
	IsGroup      bool      `json:"is_group"`                // True if this is a private channel (legacy)
	IsIM         bool      `json:"is_im"`                   // True if this is a direct message (1:1 conversation)
	IsMPIM       bool      `json:"is_mpim"`                 // Multi-Party Instant Message - a group DM with multiple users
	IsPrivate    bool      `json:"is_private"`              // True if channel is private (not visible to all workspace members)
	IsArchived   bool      `json:"is_archived"`             // True if channel has been archived
	IsMember     bool      `json:"is_member"`               // True if the authenticated user is a member
	IsShared     bool      `json:"is_shared"`               // True if shared with another workspace
	IsOrgShared  bool      `json:"is_org_shared"`           // True if shared across workspaces in the same Enterprise Grid org
	IsExtShared  bool      `json:"is_ext_shared"`           // True if shared with an external workspace (Slack Connect)
	Topic        string    `json:"topic,omitempty"`         // Current topic shown in channel header
	Purpose      string    `json:"purpose,omitempty"`       // Channel description shown in channel details
	MemberCount  int       `json:"member_count"`            // Number of members in the channel
	Created      time.Time `json:"created"`                 // When the channel was created
	LastActivity time.Time `json:"last_activity,omitempty"` // Timestamp of last message or activity
}

// SlackUser represents a Slack workspace member.
type SlackUser struct {
	ID           string            `json:"id"`                      // User ID (e.g., "U1234567890")
	Name         string            `json:"name"`                    // Username handle (e.g., "jsmith")
	RealName     string            `json:"real_name"`               // Full name from profile (e.g., "John Smith")
	DisplayName  string            `json:"display_name"`            // Custom display name set by user (may be empty)
	Title        string            `json:"title,omitempty"`         // Job title (e.g., "Software Engineer")
	Email        string            `json:"email,omitempty"`         // User's email address (requires users:read.email scope)
	Phone        string            `json:"phone,omitempty"`         // Phone number
	Avatar       string            `json:"avatar,omitempty"`        // URL to user's profile image (72x72 pixels)
	IsBot        bool              `json:"is_bot"`                  // True if this is a bot user
	IsAdmin      bool              `json:"is_admin"`                // True if user has admin privileges
	Status       string            `json:"status,omitempty"`        // Custom status text (e.g., "In a meeting")
	StatusEmoji  string            `json:"status_emoji,omitempty"`  // Status emoji (e.g., ":calendar:")
	Timezone     string            `json:"timezone,omitempty"`      // User's timezone in IANA format (e.g., "America/New_York")
	CustomFields map[string]string `json:"custom_fields,omitempty"` // Custom profile fields (label -> value)
}

// SlackAttachment represents a file attached to a message.
type SlackAttachment struct {
	ID          string `json:"id"`           // Unique file identifier (e.g., "F1234567890")
	Name        string `json:"name"`         // Original filename
	Title       string `json:"title"`        // Display title (may differ from name)
	MimeType    string `json:"mime_type"`    // MIME type (e.g., "image/png", "application/pdf")
	FileType    string `json:"file_type"`    // Slack file type (e.g., "png", "pdf")
	Size        int64  `json:"size"`         // File size in bytes
	DownloadURL string `json:"download_url"` // Private download URL (requires auth)
	Permalink   string `json:"permalink"`    // Permanent link to file in Slack
	UserID      string `json:"user_id"`      // User who uploaded the file
	Created     int64  `json:"created"`      // Unix timestamp when uploaded
	// Image-specific fields
	ImageWidth  int `json:"image_width,omitempty"`  // Original width in pixels
	ImageHeight int `json:"image_height,omitempty"` // Original height in pixels
	// Thumbnail URLs (for images)
	Thumb360 string `json:"thumb_360,omitempty"` // 360px thumbnail URL
	Thumb480 string `json:"thumb_480,omitempty"` // 480px thumbnail URL
}

// SlackFileQueryParams defines filters for listing files in a workspace.
type SlackFileQueryParams struct {
	ChannelID string   `json:"channel_id,omitempty"` // Filter by channel
	UserID    string   `json:"user_id,omitempty"`    // Filter by uploader
	Types     []string `json:"types,omitempty"`      // File types: images, pdfs, docs, etc.
	Limit     int      `json:"limit,omitempty"`      // Max files to return (default: 20)
	Cursor    string   `json:"cursor,omitempty"`     // Pagination cursor
}

// SlackFileListResponse represents a paginated list of files.
type SlackFileListResponse struct {
	Files      []SlackAttachment `json:"files"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

// SlackReaction represents an emoji reaction.
type SlackReaction struct {
	Name  string   `json:"name"`  // Emoji name without colons (e.g., "thumbsup", "heart")
	Count int      `json:"count"` // Number of users who added this reaction
	Users []string `json:"users"` // User IDs who reacted with this emoji
}

// SlackMessageQueryParams defines filters for querying messages in a channel.
type SlackMessageQueryParams struct {
	ChannelID string    `json:"channel_id"`          // Required: channel to fetch messages from
	Limit     int       `json:"limit,omitempty"`     // Max messages to return (default: 100, max: 1000)
	Cursor    string    `json:"cursor,omitempty"`    // Pagination cursor from previous response
	Oldest    time.Time `json:"oldest,omitempty"`    // Only return messages after this time
	Newest    time.Time `json:"newest,omitempty"`    // Only return messages before this time
	Inclusive bool      `json:"inclusive,omitempty"` // Include messages with exact Oldest/Newest timestamps
}

// SlackSendMessageRequest represents the parameters for sending a message to a channel.
type SlackSendMessageRequest struct {
	ChannelID string `json:"channel_id"`                // Required: channel to post the message to
	Text      string `json:"text"`                      // Required: message content (supports Slack markup)
	ThreadTS  string `json:"thread_ts,omitempty"`       // Thread timestamp to reply to (creates threaded reply)
	Broadcast bool   `json:"reply_broadcast,omitempty"` // Also post thread reply to channel (requires ThreadTS)
}

// SlackChannelQueryParams defines filters for listing channels in a workspace.
type SlackChannelQueryParams struct {
	// Types: public_channel, private_channel, mpim (group DM), im (direct message)
	Types           []string `json:"types,omitempty"`            // Channel types to include (defaults to all)
	ExcludeArchived bool     `json:"exclude_archived,omitempty"` // Skip archived channels
	Limit           int      `json:"limit,omitempty"`            // Max channels to return (default: 100, max: 1000)
	Cursor          string   `json:"cursor,omitempty"`           // Pagination cursor from previous response
	TeamID          string   `json:"team_id,omitempty"`          // Required for Enterprise Grid workspaces
}

// SlackMessageListResponse represents a paginated list of messages.
type SlackMessageListResponse struct {
	Messages   []SlackMessage `json:"messages"`              // Messages in reverse chronological order
	HasMore    bool           `json:"has_more"`              // True if more messages exist beyond this page
	NextCursor string         `json:"next_cursor,omitempty"` // Cursor for fetching the next page
}

// SlackChannelListResponse represents a paginated list of channels.
type SlackChannelListResponse struct {
	Channels   []SlackChannel `json:"channels"`              // Channels matching the query
	NextCursor string         `json:"next_cursor,omitempty"` // Cursor for fetching the next page
}

// SlackUserListResponse represents a paginated list of users.
type SlackUserListResponse struct {
	Users      []SlackUser `json:"users"`                 // Workspace members (excludes deleted users)
	NextCursor string      `json:"next_cursor,omitempty"` // Cursor for fetching the next page
}

// SlackAuth represents authentication information for a Slack workspace connection.
type SlackAuth struct {
	UserID    string `json:"user_id"`              // Authenticated user's ID
	TeamID    string `json:"team_id"`              // Workspace ID (e.g., "T1234567890")
	TeamName  string `json:"team_name"`            // Workspace display name
	UserName  string `json:"user_name"`            // Authenticated user's username
	UserEmail string `json:"user_email,omitempty"` // User's email (if available)
}

// BestDisplayName returns the best display name for a user.
// Priority: DisplayName > RealName > Name (username handle).
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
// Returns "DM" for direct messages, "Group DM" for MPIMs, or "#name" for channels.
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
// A message is in a thread if it has a ThreadTS value (either parent or reply).
func (m SlackMessage) IsThread() bool {
	return m.ThreadTS != ""
}

// ChannelType returns the type of channel as a string.
// Returns one of: "dm", "group_dm", "private", or "public".
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
