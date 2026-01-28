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
		ID:          msg.Timestamp,
		ChannelID:   channelID,
		UserID:      msg.User,
		Username:    msg.Username,
		Text:        msg.Text,
		Timestamp:   parseTimestamp(msg.Timestamp),
		ThreadTS:    msg.ThreadTimestamp,
		ReplyCount:  msg.ReplyCount,
		IsReply:     msg.ThreadTimestamp != "" && msg.ThreadTimestamp != msg.Timestamp,
		Edited:      msg.Edited != nil,
		Attachments: convertFiles(msg.Files),
		Reactions:   convertReactions(msg.Reactions),
	}
}

// convertFiles converts slack-go Files to domain.SlackAttachment slice.
func convertFiles(files []slack.File) []domain.SlackAttachment {
	if len(files) == 0 {
		return nil
	}

	result := make([]domain.SlackAttachment, len(files))
	for i, f := range files {
		result[i] = ConvertFile(f)
	}
	return result
}

// ConvertFile converts a single slack-go File to domain.SlackAttachment.
func ConvertFile(f slack.File) domain.SlackAttachment {
	return domain.SlackAttachment{
		ID:          f.ID,
		Name:        f.Name,
		Title:       f.Title,
		MimeType:    f.Mimetype,
		FileType:    f.Filetype,
		Size:        int64(f.Size),
		DownloadURL: f.URLPrivateDownload,
		Permalink:   f.Permalink,
		UserID:      f.User,
		Created:     int64(f.Created.Time().Unix()),
		ImageWidth:  f.OriginalW,
		ImageHeight: f.OriginalH,
		Thumb360:    f.Thumb360,
		Thumb480:    f.Thumb480,
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
	user := domain.SlackUser{
		ID:          u.ID,
		Name:        u.Name,
		RealName:    u.RealName,
		DisplayName: u.Profile.DisplayName,
		Title:       u.Profile.Title,
		Email:       u.Profile.Email,
		Phone:       u.Profile.Phone,
		Avatar:      u.Profile.Image72,
		IsBot:       u.IsBot,
		IsAdmin:     u.IsAdmin,
		Status:      u.Profile.StatusText,
		StatusEmoji: u.Profile.StatusEmoji,
		Timezone:    u.TZ,
	}

	// Extract custom profile fields (Department, Location, Start Date, etc.)
	if fields := u.Profile.Fields.ToMap(); len(fields) > 0 {
		user.CustomFields = make(map[string]string)
		for _, field := range fields {
			if field.Label != "" && field.Value != "" {
				user.CustomFields[field.Label] = field.Value
			}
		}
	}

	return user
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
