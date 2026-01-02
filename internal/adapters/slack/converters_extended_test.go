//go:build !integration
// +build !integration

package slack

import (
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

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
