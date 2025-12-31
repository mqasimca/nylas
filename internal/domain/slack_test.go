//go:build !integration
// +build !integration

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlackUser_BestDisplayName(t *testing.T) {
	tests := []struct {
		name string
		user SlackUser
		want string
	}{
		{
			name: "returns display name when set",
			user: SlackUser{
				ID:          "U12345",
				Name:        "johndoe",
				RealName:    "John Doe",
				DisplayName: "Johnny",
			},
			want: "Johnny",
		},
		{
			name: "returns real name when display name empty",
			user: SlackUser{
				ID:          "U12345",
				Name:        "johndoe",
				RealName:    "John Doe",
				DisplayName: "",
			},
			want: "John Doe",
		},
		{
			name: "returns username when both display and real name empty",
			user: SlackUser{
				ID:          "U12345",
				Name:        "johndoe",
				RealName:    "",
				DisplayName: "",
			},
			want: "johndoe",
		},
		{
			name: "returns empty when all names empty",
			user: SlackUser{
				ID:          "U12345",
				Name:        "",
				RealName:    "",
				DisplayName: "",
			},
			want: "",
		},
		{
			name: "handles whitespace in display name as non-empty",
			user: SlackUser{
				ID:          "U12345",
				Name:        "johndoe",
				RealName:    "John Doe",
				DisplayName: "  ",
			},
			want: "  ", // Whitespace is considered valid
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
		channel SlackChannel
		want    string
	}{
		{
			name: "returns DM for direct message",
			channel: SlackChannel{
				ID:   "D12345",
				Name: "",
				IsIM: true,
			},
			want: "DM",
		},
		{
			name: "returns Group DM for multi-party IM",
			channel: SlackChannel{
				ID:     "G12345",
				Name:   "mpdm-user1--user2-1",
				IsMPIM: true,
			},
			want: "Group DM",
		},
		{
			name: "returns #name for regular channel",
			channel: SlackChannel{
				ID:        "C12345",
				Name:      "general",
				IsChannel: true,
			},
			want: "#general",
		},
		{
			name: "returns ID when name is empty and not DM",
			channel: SlackChannel{
				ID:   "C12345",
				Name: "",
			},
			want: "C12345",
		},
		{
			name: "returns #name for private channel",
			channel: SlackChannel{
				ID:        "G12345",
				Name:      "secret-project",
				IsPrivate: true,
				IsChannel: true,
			},
			want: "#secret-project",
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
		name string
		msg  SlackMessage
		want bool
	}{
		{
			name: "returns true when ThreadTS is set",
			msg: SlackMessage{
				ID:       "1234567890.123456",
				ThreadTS: "1234567890.000000",
			},
			want: true,
		},
		{
			name: "returns false when ThreadTS is empty",
			msg: SlackMessage{
				ID:       "1234567890.123456",
				ThreadTS: "",
			},
			want: false,
		},
		{
			name: "returns true for thread parent",
			msg: SlackMessage{
				ID:         "1234567890.123456",
				ThreadTS:   "1234567890.123456", // Same as ID = parent
				ReplyCount: 5,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.IsThread()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlackChannel_ChannelType(t *testing.T) {
	tests := []struct {
		name    string
		channel SlackChannel
		want    string
	}{
		{
			name: "returns dm for direct message",
			channel: SlackChannel{
				ID:   "D12345",
				IsIM: true,
			},
			want: "dm",
		},
		{
			name: "returns group_dm for multi-party IM",
			channel: SlackChannel{
				ID:     "G12345",
				IsMPIM: true,
			},
			want: "group_dm",
		},
		{
			name: "returns private for private channel",
			channel: SlackChannel{
				ID:        "G12345",
				IsPrivate: true,
				IsChannel: true,
			},
			want: "private",
		},
		{
			name: "returns public for public channel",
			channel: SlackChannel{
				ID:        "C12345",
				IsChannel: true,
				IsPrivate: false,
			},
			want: "public",
		},
		{
			name: "prioritizes dm over other flags",
			channel: SlackChannel{
				ID:        "D12345",
				IsIM:      true,
				IsPrivate: true, // Should be ignored for DM
			},
			want: "dm",
		},
		{
			name: "prioritizes group_dm over private",
			channel: SlackChannel{
				ID:        "G12345",
				IsMPIM:    true,
				IsPrivate: true, // Should be ignored for MPIM
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
