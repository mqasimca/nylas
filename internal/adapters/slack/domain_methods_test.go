//go:build !integration
// +build !integration

package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestSlackUser_BestDisplayName(t *testing.T) {
	tests := []struct {
		name string
		user domain.SlackUser
		want string
	}{
		{
			name: "returns display name when set",
			user: domain.SlackUser{
				Name:        "jsmith",
				RealName:    "John Smith",
				DisplayName: "Johnny",
			},
			want: "Johnny",
		},
		{
			name: "returns real name when display name empty",
			user: domain.SlackUser{
				Name:        "jsmith",
				RealName:    "John Smith",
				DisplayName: "",
			},
			want: "John Smith",
		},
		{
			name: "returns name when both empty",
			user: domain.SlackUser{
				Name:        "jsmith",
				RealName:    "",
				DisplayName: "",
			},
			want: "jsmith",
		},
		{
			name: "returns empty when all empty",
			user: domain.SlackUser{
				Name:        "",
				RealName:    "",
				DisplayName: "",
			},
			want: "",
		},
		{
			name: "prefers display name over real name",
			user: domain.SlackUser{
				Name:        "username",
				RealName:    "Real Name Here",
				DisplayName: "Preferred Display",
			},
			want: "Preferred Display",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.BestDisplayName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackChannel_ChannelDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		channel domain.SlackChannel
		want    string
	}{
		{
			name: "DM returns 'DM'",
			channel: domain.SlackChannel{
				ID:   "D12345",
				Name: "",
				IsIM: true,
			},
			want: "DM",
		},
		{
			name: "MPIM returns 'Group DM'",
			channel: domain.SlackChannel{
				ID:     "G12345",
				Name:   "mpdm-user1--user2--user3-1",
				IsMPIM: true,
			},
			want: "Group DM",
		},
		{
			name: "channel with name returns #name",
			channel: domain.SlackChannel{
				ID:        "C12345",
				Name:      "general",
				IsChannel: true,
			},
			want: "#general",
		},
		{
			name: "channel without name returns ID",
			channel: domain.SlackChannel{
				ID:        "C12345",
				Name:      "",
				IsChannel: true,
			},
			want: "C12345",
		},
		{
			name: "private channel with name returns #name",
			channel: domain.SlackChannel{
				ID:        "G12345",
				Name:      "private-team",
				IsPrivate: true,
			},
			want: "#private-team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.ChannelDisplayName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackMessage_IsThread(t *testing.T) {
	tests := []struct {
		name    string
		message domain.SlackMessage
		want    bool
	}{
		{
			name: "message with thread_ts is in thread",
			message: domain.SlackMessage{
				ID:       "1234567890.123456",
				ThreadTS: "1234567890.000000",
			},
			want: true,
		},
		{
			name: "message without thread_ts is not in thread",
			message: domain.SlackMessage{
				ID:       "1234567890.123456",
				ThreadTS: "",
			},
			want: false,
		},
		{
			name: "thread parent (same ts)",
			message: domain.SlackMessage{
				ID:       "1234567890.123456",
				ThreadTS: "1234567890.123456",
			},
			want: true,
		},
		{
			name: "reply in thread",
			message: domain.SlackMessage{
				ID:       "1234567891.123456",
				ThreadTS: "1234567890.123456",
				IsReply:  true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.message.IsThread()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackChannel_ChannelType(t *testing.T) {
	tests := []struct {
		name    string
		channel domain.SlackChannel
		want    string
	}{
		{
			name: "direct message",
			channel: domain.SlackChannel{
				IsIM: true,
			},
			want: "dm",
		},
		{
			name: "group DM",
			channel: domain.SlackChannel{
				IsMPIM: true,
			},
			want: "group_dm",
		},
		{
			name: "private channel",
			channel: domain.SlackChannel{
				IsPrivate: true,
			},
			want: "private",
		},
		{
			name: "public channel",
			channel: domain.SlackChannel{
				IsChannel: true,
			},
			want: "public",
		},
		{
			name:    "public channel by default",
			channel: domain.SlackChannel{},
			want:    "public",
		},
		{
			name: "DM takes precedence",
			channel: domain.SlackChannel{
				IsIM:      true,
				IsPrivate: true,
			},
			want: "dm",
		},
		{
			name: "MPIM takes precedence over private",
			channel: domain.SlackChannel{
				IsMPIM:    true,
				IsPrivate: true,
			},
			want: "group_dm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.ChannelType()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackMessage_FieldsPreserved(t *testing.T) {
	msg := domain.SlackMessage{
		ID:         "1234567890.123456",
		ChannelID:  "C12345",
		UserID:     "U12345",
		Username:   "testuser",
		Text:       "Hello, world!",
		ThreadTS:   "1234567890.000000",
		ReplyCount: 5,
		IsReply:    true,
		Edited:     true,
		Attachments: []domain.SlackAttachment{
			{ID: "F1", Name: "file.txt"},
		},
		Reactions: []domain.SlackReaction{
			{Name: "thumbsup", Count: 3, Users: []string{"U1", "U2", "U3"}},
		},
	}

	assert.Equal(t, "1234567890.123456", msg.ID)
	assert.Equal(t, "C12345", msg.ChannelID)
	assert.Equal(t, "U12345", msg.UserID)
	assert.Equal(t, "testuser", msg.Username)
	assert.Equal(t, "Hello, world!", msg.Text)
	assert.Equal(t, "1234567890.000000", msg.ThreadTS)
	assert.Equal(t, 5, msg.ReplyCount)
	assert.True(t, msg.IsReply)
	assert.True(t, msg.Edited)
	assert.Len(t, msg.Attachments, 1)
	assert.Len(t, msg.Reactions, 1)
}

func TestSlackChannel_FieldsPreserved(t *testing.T) {
	ch := domain.SlackChannel{
		ID:          "C12345",
		Name:        "general",
		IsChannel:   true,
		IsGroup:     false,
		IsIM:        false,
		IsMPIM:      false,
		IsPrivate:   false,
		IsArchived:  false,
		IsMember:    true,
		IsShared:    true,
		IsOrgShared: true,
		IsExtShared: false,
		Topic:       "General discussion",
		Purpose:     "Company-wide announcements",
		MemberCount: 100,
	}

	assert.Equal(t, "C12345", ch.ID)
	assert.Equal(t, "general", ch.Name)
	assert.True(t, ch.IsChannel)
	assert.False(t, ch.IsGroup)
	assert.False(t, ch.IsIM)
	assert.False(t, ch.IsMPIM)
	assert.False(t, ch.IsPrivate)
	assert.False(t, ch.IsArchived)
	assert.True(t, ch.IsMember)
	assert.True(t, ch.IsShared)
	assert.True(t, ch.IsOrgShared)
	assert.False(t, ch.IsExtShared)
	assert.Equal(t, "General discussion", ch.Topic)
	assert.Equal(t, "Company-wide announcements", ch.Purpose)
	assert.Equal(t, 100, ch.MemberCount)
}

func TestSlackUser_FieldsPreserved(t *testing.T) {
	user := domain.SlackUser{
		ID:          "U12345",
		Name:        "jsmith",
		RealName:    "John Smith",
		DisplayName: "Johnny",
		Email:       "john@example.com",
		Avatar:      "https://example.com/avatar.png",
		IsBot:       false,
		IsAdmin:     true,
		Status:      "In a meeting",
		Timezone:    "America/New_York",
	}

	assert.Equal(t, "U12345", user.ID)
	assert.Equal(t, "jsmith", user.Name)
	assert.Equal(t, "John Smith", user.RealName)
	assert.Equal(t, "Johnny", user.DisplayName)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "https://example.com/avatar.png", user.Avatar)
	assert.False(t, user.IsBot)
	assert.True(t, user.IsAdmin)
	assert.Equal(t, "In a meeting", user.Status)
	assert.Equal(t, "America/New_York", user.Timezone)
}

func TestSlackReaction_FieldsPreserved(t *testing.T) {
	reaction := domain.SlackReaction{
		Name:  "heart",
		Count: 5,
		Users: []string{"U1", "U2", "U3", "U4", "U5"},
	}

	assert.Equal(t, "heart", reaction.Name)
	assert.Equal(t, 5, reaction.Count)
	assert.Len(t, reaction.Users, 5)
}

func TestSlackAttachment_FieldsPreserved(t *testing.T) {
	attachment := domain.SlackAttachment{
		ID:          "F12345",
		Name:        "document.pdf",
		Title:       "Important Document",
		MimeType:    "application/pdf",
		FileType:    "pdf",
		Size:        1024,
		DownloadURL: "https://files.slack.com/download/doc.pdf",
		Permalink:   "https://workspace.slack.com/files/doc.pdf",
		UserID:      "U12345",
		Created:     1703980800,
		ImageWidth:  0,
		ImageHeight: 0,
		Thumb360:    "",
		Thumb480:    "",
	}

	assert.Equal(t, "F12345", attachment.ID)
	assert.Equal(t, "document.pdf", attachment.Name)
	assert.Equal(t, "Important Document", attachment.Title)
	assert.Equal(t, "application/pdf", attachment.MimeType)
	assert.Equal(t, "pdf", attachment.FileType)
	assert.Equal(t, int64(1024), attachment.Size)
	assert.Equal(t, "https://files.slack.com/download/doc.pdf", attachment.DownloadURL)
	assert.Equal(t, "https://workspace.slack.com/files/doc.pdf", attachment.Permalink)
	assert.Equal(t, "U12345", attachment.UserID)
	assert.Equal(t, int64(1703980800), attachment.Created)
}

func TestSlackAuth_FieldsPreserved(t *testing.T) {
	auth := domain.SlackAuth{
		UserID:    "U12345",
		TeamID:    "T12345",
		TeamName:  "Test Workspace",
		UserName:  "testuser",
		UserEmail: "test@example.com",
	}

	assert.Equal(t, "U12345", auth.UserID)
	assert.Equal(t, "T12345", auth.TeamID)
	assert.Equal(t, "Test Workspace", auth.TeamName)
	assert.Equal(t, "testuser", auth.UserName)
	assert.Equal(t, "test@example.com", auth.UserEmail)
}
