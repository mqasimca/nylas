package slack

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

func newSendCmd() *cobra.Command {
	var (
		channelID   string
		channelName string
		text        string
		noConfirm   bool
	)

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message to a channel",
		Long: `Send a message to a Slack channel as yourself.

The message will appear with your name and profile picture,
exactly as if you typed it in Slack.

Examples:
  # Send to channel
  nylas slack send --channel general --text "Hello team!"

  # Send without confirmation
  nylas slack send --channel general --text "Quick update" --yes`,
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
				return common.NewUserError("channel is required", "Use --channel or --channel-id")
			}

			if text == "" {
				return common.NewUserError("message text is required", "Use --text")
			}

			if !noConfirm {
				cyan := color.New(color.FgCyan)
				fmt.Printf("Channel: %s\n", cyan.Sprint(channelName))
				fmt.Printf("Message: %s\n\n", text)
				fmt.Print("Send this message? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(strings.ToLower(confirm))

				if confirm != "y" && confirm != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			msg, err := client.SendMessage(ctx, &domain.SlackSendMessageRequest{
				ChannelID: resolvedChannelID,
				Text:      text,
			})
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Message sent! ID: %s\n", msg.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&channelName, "channel", "c", "", "Channel name")
	cmd.Flags().StringVar(&channelID, "channel-id", "", "Channel ID")
	cmd.Flags().StringVarP(&text, "text", "t", "", "Message text")
	cmd.Flags().BoolVarP(&noConfirm, "yes", "y", false, "Skip confirmation")

	return cmd
}

func newReplyCmd() *cobra.Command {
	var (
		channelID   string
		channelName string
		threadTS    string
		text        string
		broadcast   bool
		noConfirm   bool
	)

	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to a thread",
		Long: `Reply to a Slack thread as yourself.

Use the message timestamp/ID from 'nylas slack messages --id' as the thread ID.

Examples:
  # Reply to a thread
  nylas slack reply --channel general --thread 1234567890.123456 --text "Got it!"

  # Reply and also post to channel
  nylas slack reply --channel general --thread 1234567890.123456 --text "Update" --broadcast`,
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
				return common.NewUserError("channel is required", "")
			}
			if threadTS == "" {
				return common.NewUserError(
					"thread timestamp is required",
					"Use --thread with the message ID from 'nylas slack messages --id'",
				)
			}
			if text == "" {
				return common.NewUserError("message text is required", "Use --text")
			}

			if !noConfirm {
				cyan := color.New(color.FgCyan)
				fmt.Printf("Channel: %s\n", cyan.Sprint(channelName))
				fmt.Printf("Thread:  %s\n", threadTS)
				fmt.Printf("Reply:   %s\n", text)
				if broadcast {
					fmt.Println("(Also posting to channel)")
				}
				fmt.Print("\nSend this reply? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(strings.ToLower(confirm))

				if confirm != "y" && confirm != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			msg, err := client.SendMessage(ctx, &domain.SlackSendMessageRequest{
				ChannelID: resolvedChannelID,
				Text:      text,
				ThreadTS:  threadTS,
				Broadcast: broadcast,
			})
			if err != nil {
				return fmt.Errorf("failed to send reply: %w", err)
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Reply sent! ID: %s\n", msg.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&channelName, "channel", "c", "", "Channel name")
	cmd.Flags().StringVar(&channelID, "channel-id", "", "Channel ID")
	cmd.Flags().StringVar(&threadTS, "thread", "", "Thread timestamp to reply to")
	cmd.Flags().StringVarP(&text, "text", "t", "", "Reply text")
	cmd.Flags().BoolVar(&broadcast, "broadcast", false, "Also post to channel")
	cmd.Flags().BoolVarP(&noConfirm, "yes", "y", false, "Skip confirmation")

	_ = cmd.MarkFlagRequired("thread")

	return cmd
}
