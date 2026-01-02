// messages.go provides message listing and reading commands for Slack channels.

package slack

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/slack"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// newMessagesCmd creates the messages command for managing Slack messages.
func newMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"msg", "msgs"},
		Short:   "Manage Slack messages",
		Long:    `Commands for reading and managing Slack messages.`,
	}

	cmd.AddCommand(newMessageListCmd())

	return cmd
}

// newMessageListCmd creates the list subcommand for reading channel messages.
func newMessageListCmd() *cobra.Command {
	var (
		channelID     string
		channelName   string
		limit         int
		showID        bool
		threadTS      string
		fetchAll      bool
		expandThreads bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Read messages from a channel",
		Long: `Read messages from a Slack channel, DM, or group DM.

Examples:
  # Read from channel by name
  nylas slack messages list --channel general

  # Read from channel by ID
  nylas slack messages list --channel-id C1234567890

  # Read thread replies
  nylas slack messages list --channel general --thread 1234567890.123456

  # Limit results
  nylas slack messages list --channel general --limit 100

  # Fetch ALL messages (paginate through entire history)
  nylas slack messages list --channel general --all

  # Expand all threads inline (show thread replies under parent messages)
  nylas slack messages list --channel general --expand-threads`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			// Use longer timeout when fetching all messages
			var ctx context.Context
			var cancel context.CancelFunc
			if fetchAll || expandThreads {
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
			} else {
				ctx, cancel = createContext()
			}
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

			var allMessages []domain.SlackMessage
			cursor := ""
			pageSize := limit
			if fetchAll {
				pageSize = 1000 // Max per request when fetching all
			}

			for {
				resp, err := client.GetMessages(ctx, &domain.SlackMessageQueryParams{
					ChannelID: resolvedChannelID,
					Limit:     pageSize,
					Cursor:    cursor,
				})
				if err != nil {
					return fmt.Errorf("failed to get messages: %w", err)
				}

				allMessages = append(allMessages, resp.Messages...)

				// If not fetching all, or no more pages, stop
				if !fetchAll || resp.NextCursor == "" {
					if !fetchAll && resp.HasMore {
						dim := color.New(color.Faint)
						_, _ = dim.Printf("(fetched %d messages, more available - use --all to fetch all)\n\n", len(allMessages))
					}
					break
				}

				cursor = resp.NextCursor
				dim := color.New(color.Faint)
				_, _ = dim.Printf("Fetched %d messages...\n", len(allMessages))
			}

			if len(allMessages) == 0 {
				fmt.Println("No messages found")
				return nil
			}

			slackClient, isSlackClient := client.(*slack.Client)
			if isSlackClient {
				if err := slackClient.GetUsersForMessages(ctx, allMessages); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve all usernames: %v\n", err)
				}
			}

			// If expanding threads, fetch replies for messages with threads
			threadReplies := make(map[string][]domain.SlackMessage)
			if expandThreads {
				dim := color.New(color.Faint)
				var allReplies []domain.SlackMessage
				for _, msg := range allMessages {
					if msg.ReplyCount > 0 {
						replies, threadErr := client.GetThreadReplies(ctx, resolvedChannelID, msg.ID, 100)
						if threadErr == nil && len(replies) > 1 {
							// Skip first reply (it's the parent message)
							threadReplies[msg.ID] = replies[1:]
							allReplies = append(allReplies, replies[1:]...)
						}
					}
				}
				// Resolve usernames for thread replies
				if isSlackClient && len(allReplies) > 0 {
					_ = slackClient.GetUsersForMessages(ctx, allReplies)
					// Update the threadReplies map with resolved usernames
					for msgID, replies := range threadReplies {
						for i := range replies {
							for _, resolved := range allReplies {
								if replies[i].ID == resolved.ID {
									replies[i].Username = resolved.Username
									break
								}
							}
						}
						threadReplies[msgID] = replies
					}
				}
				if len(threadReplies) > 0 {
					_, _ = dim.Printf("Expanded %d threads\n\n", len(threadReplies))
				}
			}

			// Print in chronological order (oldest first)
			for i := len(allMessages) - 1; i >= 0; i-- {
				msg := allMessages[i]
				printMessage(msg, showID, expandThreads)

				// Print thread replies if expanded
				if replies, ok := threadReplies[msg.ID]; ok {
					for _, reply := range replies {
						printThreadReply(reply, showID)
					}
				}
			}

			dim := color.New(color.Faint)
			_, _ = dim.Printf("\nTotal: %d messages\n", len(allMessages))

			return nil
		},
	}

	cmd.Flags().StringVarP(&channelName, "channel", "c", "", "Channel name (without #)")
	cmd.Flags().StringVar(&channelID, "channel-id", "", "Channel ID")
	cmd.Flags().IntVarP(&limit, "limit", "l", 500, "Number of messages to fetch (max 1000 per page)")
	cmd.Flags().BoolVar(&showID, "id", false, "Show message timestamps/IDs")
	cmd.Flags().StringVar(&threadTS, "thread", "", "Thread timestamp to show replies")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "Fetch all messages (paginate through entire history)")
	cmd.Flags().BoolVar(&expandThreads, "expand-threads", false, "Expand thread replies inline")

	return cmd
}

// showThreadReplies fetches and displays replies within a thread.
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
		printMessage(msg, showID, false)
	}

	return nil
}

// printMessage formats and prints a single Slack message to stdout.
func printMessage(msg domain.SlackMessage, showID bool, hideReplies bool) {
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

	// Only show reply count indicator when not expanding threads
	if msg.ReplyCount > 0 && !hideReplies {
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

// printThreadReply formats and prints a thread reply with indentation.
func printThreadReply(msg domain.SlackMessage, showID bool) {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	username := msg.Username
	if username == "" {
		username = msg.UserID
	}

	fmt.Print("    ↳ ")
	fmt.Print(cyan.Sprint(username))
	fmt.Print("  ")
	_, _ = dim.Print(msg.Timestamp.Format("Jan 2 15:04"))

	if showID {
		_, _ = dim.Printf(" [%s]", msg.ID)
	}
	if msg.Edited {
		_, _ = dim.Print(" (edited)")
	}
	fmt.Println()

	// Indent the message text
	fmt.Printf("      %s\n", msg.Text)

	if len(msg.Reactions) > 0 {
		fmt.Print("      ")
		for _, r := range msg.Reactions {
			fmt.Printf(":%s: %d  ", r.Name, r.Count)
		}
		fmt.Println()
	}

	fmt.Println()
}
