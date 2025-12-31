package slack

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/slack"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

func newMessagesCmd() *cobra.Command {
	var (
		channelID   string
		channelName string
		limit       int
		showID      bool
		threadTS    string
	)

	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"msg", "msgs"},
		Short:   "Read messages from a channel",
		Long: `Read messages from a Slack channel, DM, or group DM.

Examples:
  # Read from channel by name
  nylas slack messages --channel general

  # Read from channel by ID
  nylas slack messages --channel-id C1234567890

  # Read thread replies
  nylas slack messages --channel general --thread 1234567890.123456

  # Limit results
  nylas slack messages --channel general --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			ctx, cancel := createContext()
			defer cancel()

			resolvedChannelID := channelID
			if channelName != "" && channelID == "" {
				resolvedChannelID, err = resolveChannelName(ctx, client, channelName)
				if err != nil {
					return fmt.Errorf("channel not found: %s", channelName)
				}
			}

			if resolvedChannelID == "" {
				return common.NewUserError(
					"channel is required",
					"Use --channel NAME or --channel-id ID",
				)
			}

			if threadTS != "" {
				return showThreadReplies(ctx, client, resolvedChannelID, threadTS, limit, showID)
			}

			resp, err := client.GetMessages(ctx, &domain.SlackMessageQueryParams{
				ChannelID: resolvedChannelID,
				Limit:     limit,
			})
			if err != nil {
				return fmt.Errorf("failed to get messages: %w", err)
			}

			if len(resp.Messages) == 0 {
				fmt.Println("No messages found")
				return nil
			}

			if slackClient, ok := client.(*slack.Client); ok {
				_ = slackClient.GetUsersForMessages(ctx, resp.Messages)
			}

			for i := len(resp.Messages) - 1; i >= 0; i-- {
				printMessage(resp.Messages[i], showID)
			}

			if resp.HasMore {
				dim := color.New(color.Faint)
				_, _ = dim.Printf("\n(more messages available, use --limit to fetch more)\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&channelName, "channel", "c", "", "Channel name (without #)")
	cmd.Flags().StringVar(&channelID, "channel-id", "", "Channel ID")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of messages to fetch")
	cmd.Flags().BoolVar(&showID, "id", false, "Show message timestamps/IDs")
	cmd.Flags().StringVar(&threadTS, "thread", "", "Thread timestamp to show replies")

	return cmd
}

func showThreadReplies(ctx context.Context, client ports.SlackClient, channelID, threadTS string, limit int, showID bool) error {
	replies, err := client.GetThreadReplies(ctx, channelID, threadTS, limit)
	if err != nil {
		return fmt.Errorf("failed to get thread replies: %w", err)
	}

	if len(replies) == 0 {
		fmt.Println("No replies found")
		return nil
	}

	cyan := color.New(color.FgCyan)
	_, _ = cyan.Printf("Thread with %d messages:\n\n", len(replies))

	for _, msg := range replies {
		printMessage(msg, showID)
	}

	return nil
}

func printMessage(msg domain.SlackMessage, showID bool) {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)
	yellow := color.New(color.FgYellow)

	username := msg.Username
	if username == "" {
		username = msg.UserID
	}

	fmt.Print(cyan.Sprint(username))
	fmt.Print("  ")
	_, _ = dim.Print(msg.Timestamp.Format("Jan 2 15:04"))

	if showID {
		_, _ = dim.Printf(" [%s]", msg.ID)
	}

	if msg.IsReply {
		_, _ = yellow.Print(" (reply)")
	}
	if msg.Edited {
		_, _ = dim.Print(" (edited)")
	}
	fmt.Println()

	fmt.Println(msg.Text)

	if msg.ReplyCount > 0 {
		_, _ = dim.Printf("  └─ %d replies\n", msg.ReplyCount)
	}

	if len(msg.Reactions) > 0 {
		fmt.Print("  ")
		for _, r := range msg.Reactions {
			fmt.Printf(":%s: %d  ", r.Name, r.Count)
		}
		fmt.Println()
	}

	fmt.Println()
}
