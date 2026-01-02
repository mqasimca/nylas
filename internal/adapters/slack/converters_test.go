//go:build !integration
// +build !integration

package slack

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name      string
		msg       slack.Message
		channelID string
		want      struct {
			id        string
			channelID string
			userID    string
			text      string
			isReply   bool
			edited    bool
		}
	}{
		{
			name: "basic message",
			msg: slack.Message{
				Msg: slack.Msg{
					Timestamp: "1234567890.123456",
					User:      "U12345",
					Username:  "testuser",
					Text:      "Hello world",
				},
			},
			channelID: "C12345",
			want: struct {
				id        string
				channelID string
				userID    string
				text      string
				isReply   bool
				edited    bool
			}{
				id:        "1234567890.123456",
				channelID: "C12345",
				userID:    "U12345",
				text:      "Hello world",
				isReply:   false,
				edited:    false,
			},
		},
		{
			name: "threaded reply",
			msg: slack.Message{
				Msg: slack.Msg{
					Timestamp:       "1234567891.123456",
					ThreadTimestamp: "1234567890.123456",
					User:            "U12345",
					Text:            "Reply text",
				},
			},
			channelID: "C12345",
			want: struct {
				id        string
				channelID string
				userID    string
				text      string
				isReply   bool
				edited    bool
			}{
				id:        "1234567891.123456",
				channelID: "C12345",
				userID:    "U12345",
				text:      "Reply text",
				isReply:   true,
				edited:    false,
			},
		},
		{
			name: "edited message",
			msg: slack.Message{
				Msg: slack.Msg{
					Timestamp: "1234567890.123456",
					User:      "U12345",
					Text:      "Edited text",
					Edited: &slack.Edited{
						User:      "U12345",
						Timestamp: "1234567891.123456",
					},
				},
			},
			channelID: "C12345",
			want: struct {
				id        string
				channelID string
				userID    string
				text      string
				isReply   bool
				edited    bool
			}{
				id:        "1234567890.123456",
				channelID: "C12345",
				userID:    "U12345",
				text:      "Edited text",
				isReply:   false,
				edited:    true,
			},
		},
		{
			name: "thread parent (not a reply)",
			msg: slack.Message{
				Msg: slack.Msg{
					Timestamp:       "1234567890.123456",
					ThreadTimestamp: "1234567890.123456", // Same as timestamp = parent
					User:            "U12345",
					Text:            "Parent message",
					ReplyCount:      5,
				},
			},
			channelID: "C12345",
			want: struct {
				id        string
				channelID string
				userID    string
				text      string
				isReply   bool
				edited    bool
			}{
				id:        "1234567890.123456",
				channelID: "C12345",
				userID:    "U12345",
				text:      "Parent message",
				isReply:   false,
				edited:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessage(tt.msg, tt.channelID)

			assert.Equal(t, tt.want.id, got.ID)
			assert.Equal(t, tt.want.channelID, got.ChannelID)
			assert.Equal(t, tt.want.userID, got.UserID)
			assert.Equal(t, tt.want.text, got.Text)
			assert.Equal(t, tt.want.isReply, got.IsReply)
			assert.Equal(t, tt.want.edited, got.Edited)
		})
	}
}

func TestConvertChannel(t *testing.T) {
	tests := []struct {
		name string
		ch   slack.Channel
		want struct {
			id         string
			name       string
			isChannel  bool
			isPrivate  bool
			isIM       bool
			isMPIM     bool
			isArchived bool
		}
	}{
		{
			name: "public channel",
			ch: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID: "C12345",
					},
					Name:    "general",
					Topic:   slack.Topic{Value: "General discussion"},
					Purpose: slack.Purpose{Value: "Company-wide announcements"},
				},
				IsChannel: true,
			},
			want: struct {
				id         string
				name       string
				isChannel  bool
				isPrivate  bool
				isIM       bool
				isMPIM     bool
				isArchived bool
			}{
				id:         "C12345",
				name:       "general",
				isChannel:  true,
				isPrivate:  false,
				isIM:       false,
				isMPIM:     false,
				isArchived: false,
			},
		},
		{
			name: "private channel",
			ch: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "G12345",
						IsPrivate: true,
					},
					Name: "secret-project",
				},
				IsChannel: true,
			},
			want: struct {
				id         string
				name       string
				isChannel  bool
				isPrivate  bool
				isIM       bool
				isMPIM     bool
				isArchived bool
			}{
				id:         "G12345",
				name:       "secret-project",
				isChannel:  true,
				isPrivate:  true,
				isIM:       false,
				isMPIM:     false,
				isArchived: false,
			},
		},
		{
			name: "direct message",
			ch: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:   "D12345",
						IsIM: true,
					},
				},
			},
			want: struct {
				id         string
				name       string
				isChannel  bool
				isPrivate  bool
				isIM       bool
				isMPIM     bool
				isArchived bool
			}{
				id:         "D12345",
				name:       "",
				isChannel:  false,
				isPrivate:  false,
				isIM:       true,
				isMPIM:     false,
				isArchived: false,
			},
		},
		{
			name: "multi-party IM",
			ch: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:     "G12345",
						IsMpIM: true,
					},
					Name: "mpdm-user1--user2--user3-1",
				},
			},
			want: struct {
				id         string
				name       string
				isChannel  bool
				isPrivate  bool
				isIM       bool
				isMPIM     bool
				isArchived bool
			}{
				id:         "G12345",
				name:       "mpdm-user1--user2--user3-1",
				isChannel:  false,
				isPrivate:  false,
				isIM:       false,
				isMPIM:     true,
				isArchived: false,
			},
		},
		{
			name: "archived channel",
			ch: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID: "C12345",
					},
					Name:       "old-project",
					IsArchived: true,
				},
				IsChannel: true,
			},
			want: struct {
				id         string
				name       string
				isChannel  bool
				isPrivate  bool
				isIM       bool
				isMPIM     bool
				isArchived bool
			}{
				id:         "C12345",
				name:       "old-project",
				isChannel:  true,
				isPrivate:  false,
				isIM:       false,
				isMPIM:     false,
				isArchived: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertChannel(tt.ch)

			assert.Equal(t, tt.want.id, got.ID)
			assert.Equal(t, tt.want.name, got.Name)
			assert.Equal(t, tt.want.isChannel, got.IsChannel)
			assert.Equal(t, tt.want.isPrivate, got.IsPrivate)
			assert.Equal(t, tt.want.isIM, got.IsIM)
			assert.Equal(t, tt.want.isMPIM, got.IsMPIM)
			assert.Equal(t, tt.want.isArchived, got.IsArchived)
		})
	}
}
