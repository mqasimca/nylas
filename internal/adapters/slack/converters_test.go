//go:build !integration
// +build !integration

package slack

import (
	"testing"
	"time"

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

func TestConvertUser(t *testing.T) {
	tests := []struct {
		name string
		user slack.User
		want struct {
			id          string
			name        string
			realName    string
			displayName string
			email       string
			isBot       bool
			isAdmin     bool
		}
	}{
		{
			name: "regular user",
			user: slack.User{
				ID:       "U12345",
				Name:     "johndoe",
				RealName: "John Doe",
				Profile: slack.UserProfile{
					DisplayName: "Johnny",
					Email:       "john@example.com",
					Image72:     "https://example.com/avatar.png",
					StatusText:  "Working from home",
				},
				TZ:      "America/New_York",
				IsAdmin: false,
				IsBot:   false,
			},
			want: struct {
				id          string
				name        string
				realName    string
				displayName string
				email       string
				isBot       bool
				isAdmin     bool
			}{
				id:          "U12345",
				name:        "johndoe",
				realName:    "John Doe",
				displayName: "Johnny",
				email:       "john@example.com",
				isBot:       false,
				isAdmin:     false,
			},
		},
		{
			name: "admin user",
			user: slack.User{
				ID:       "U99999",
				Name:     "admin",
				RealName: "Admin User",
				Profile: slack.UserProfile{
					DisplayName: "Admin",
					Email:       "admin@example.com",
				},
				IsAdmin: true,
				IsBot:   false,
			},
			want: struct {
				id          string
				name        string
				realName    string
				displayName string
				email       string
				isBot       bool
				isAdmin     bool
			}{
				id:          "U99999",
				name:        "admin",
				realName:    "Admin User",
				displayName: "Admin",
				email:       "admin@example.com",
				isBot:       false,
				isAdmin:     true,
			},
		},
		{
			name: "bot user",
			user: slack.User{
				ID:       "B12345",
				Name:     "slackbot",
				RealName: "Slack Bot",
				Profile: slack.UserProfile{
					DisplayName: "SlackBot",
				},
				IsBot:   true,
				IsAdmin: false,
			},
			want: struct {
				id          string
				name        string
				realName    string
				displayName string
				email       string
				isBot       bool
				isAdmin     bool
			}{
				id:          "B12345",
				name:        "slackbot",
				realName:    "Slack Bot",
				displayName: "SlackBot",
				email:       "",
				isBot:       true,
				isAdmin:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUser(tt.user)

			assert.Equal(t, tt.want.id, got.ID)
			assert.Equal(t, tt.want.name, got.Name)
			assert.Equal(t, tt.want.realName, got.RealName)
			assert.Equal(t, tt.want.displayName, got.DisplayName)
			assert.Equal(t, tt.want.email, got.Email)
			assert.Equal(t, tt.want.isBot, got.IsBot)
			assert.Equal(t, tt.want.isAdmin, got.IsAdmin)
		})
	}
}

func TestConvertReactions(t *testing.T) {
	tests := []struct {
		name      string
		reactions []slack.ItemReaction
		wantLen   int
	}{
		{
			name:      "empty reactions",
			reactions: nil,
			wantLen:   0,
		},
		{
			name:      "empty slice",
			reactions: []slack.ItemReaction{},
			wantLen:   0,
		},
		{
			name: "single reaction",
			reactions: []slack.ItemReaction{
				{Name: "thumbsup", Count: 3, Users: []string{"U1", "U2", "U3"}},
			},
			wantLen: 1,
		},
		{
			name: "multiple reactions",
			reactions: []slack.ItemReaction{
				{Name: "thumbsup", Count: 3, Users: []string{"U1", "U2", "U3"}},
				{Name: "heart", Count: 1, Users: []string{"U4"}},
				{Name: "fire", Count: 2, Users: []string{"U5", "U6"}},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertReactions(tt.reactions)

			if tt.wantLen == 0 {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, tt.wantLen)
				for i, r := range tt.reactions {
					assert.Equal(t, r.Name, got[i].Name)
					assert.Equal(t, r.Count, got[i].Count)
					assert.Equal(t, r.Users, got[i].Users)
				}
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name string
		ts   string
		want time.Time
	}{
		{
			name: "valid timestamp",
			ts:   "1234567890.123456",
			want: time.Unix(1234567890, 0),
		},
		{
			name: "timestamp without decimal",
			ts:   "1234567890",
			want: time.Unix(1234567890, 0),
		},
		{
			name: "empty timestamp",
			ts:   "",
			want: time.Time{},
		},
		{
			name: "invalid timestamp",
			ts:   "invalid",
			want: time.Time{},
		},
		{
			name: "recent timestamp",
			ts:   "1703980800.000000",
			want: time.Unix(1703980800, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimestamp(tt.ts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "valid time",
			t:    time.Unix(1234567890, 0),
			want: "1234567890.000000",
		},
		{
			name: "zero time",
			t:    time.Time{},
			want: "",
		},
		{
			name: "recent time",
			t:    time.Unix(1703980800, 0),
			want: "1703980800.000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimestamp(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertSearchMatch(t *testing.T) {
	match := slack.SearchMessage{
		Timestamp: "1234567890.123456",
		User:      "U12345",
		Username:  "testuser",
		Text:      "Search result text",
		Channel: slack.CtxChannel{
			ID:   "C12345",
			Name: "general",
		},
	}

	got := convertSearchMatch(match)

	assert.Equal(t, "1234567890.123456", got.ID)
	assert.Equal(t, "C12345", got.ChannelID)
	assert.Equal(t, "U12345", got.UserID)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "Search result text", got.Text)
	assert.Equal(t, time.Unix(1234567890, 0), got.Timestamp)
}
