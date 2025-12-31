// converters.go provides type conversion utilities between slack-go and domain types.

package slack

import (
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"

	"github.com/mqasimca/nylas/internal/domain"
)

// convertMessage converts slack-go Message to domain.SlackMessage.
func convertMessage(msg slack.Message, channelID string) domain.SlackMessage {
	return domain.SlackMessage{
		ID:         msg.Timestamp,
		ChannelID:  channelID,
		UserID:     msg.User,
		Username:   msg.Username,
		Text:       msg.Text,
		Timestamp:  parseTimestamp(msg.Timestamp),
		ThreadTS:   msg.ThreadTimestamp,
		ReplyCount: msg.ReplyCount,
		IsReply:    msg.ThreadTimestamp != "" && msg.ThreadTimestamp != msg.Timestamp,
		Edited:     msg.Edited != nil,
		Reactions:  convertReactions(msg.Reactions),
	}
}

// convertChannel converts slack-go Channel to domain.SlackChannel.
func convertChannel(ch slack.Channel) domain.SlackChannel {
	return domain.SlackChannel{
		ID:          ch.ID,
		Name:        ch.Name,
		IsChannel:   ch.IsChannel,
		IsGroup:     ch.IsGroup,
		IsIM:        ch.IsIM,
		IsMPIM:      ch.IsMpIM,
		IsPrivate:   ch.IsPrivate,
		IsArchived:  ch.IsArchived,
		IsMember:    ch.IsMember,
		IsShared:    ch.IsShared,
		IsOrgShared: ch.IsOrgShared,
		IsExtShared: ch.IsExtShared,
		Topic:       ch.Topic.Value,
		Purpose:     ch.Purpose.Value,
		MemberCount: ch.NumMembers,
		Created:     time.Unix(int64(ch.Created), 0),
	}
}

// convertUser converts slack-go User to domain.SlackUser.
func convertUser(u slack.User) domain.SlackUser {
	return domain.SlackUser{
		ID:          u.ID,
		Name:        u.Name,
		RealName:    u.RealName,
		DisplayName: u.Profile.DisplayName,
		Email:       u.Profile.Email,
		Avatar:      u.Profile.Image72,
		IsBot:       u.IsBot,
		IsAdmin:     u.IsAdmin,
		Status:      u.Profile.StatusText,
		Timezone:    u.TZ,
	}
}

// convertReactions converts slack-go reactions to domain type.
func convertReactions(reactions []slack.ItemReaction) []domain.SlackReaction {
	if len(reactions) == 0 {
		return nil
	}

	result := make([]domain.SlackReaction, len(reactions))
	for i, r := range reactions {
		result[i] = domain.SlackReaction{
			Name:  r.Name,
			Count: r.Count,
			Users: r.Users,
		}
	}
	return result
}

// convertSearchMatch converts search result to domain message.
func convertSearchMatch(match slack.SearchMessage) domain.SlackMessage {
	return domain.SlackMessage{
		ID:        match.Timestamp,
		ChannelID: match.Channel.ID,
		UserID:    match.User,
		Username:  match.Username,
		Text:      match.Text,
		Timestamp: parseTimestamp(match.Timestamp),
	}
}

// parseTimestamp converts Slack timestamp string to time.Time.
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}

	parts := strings.Split(ts, ".")
	if len(parts) == 0 {
		return time.Time{}
	}

	sec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.Unix(sec, 0)
}

// formatTimestamp converts time.Time to Slack timestamp string.
func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return strconv.FormatInt(t.Unix(), 10) + ".000000"
}
