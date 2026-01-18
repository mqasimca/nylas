package inbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

func newMessagesCmd() *cobra.Command {
	var (
		limit      int
		unread     bool
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "messages [inbox-id]",
		Short: "List messages in an inbound inbox",
		Long: `List messages received at an inbound inbox.

Examples:
  # List messages for an inbox
  nylas inbound messages abc123

  # List only unread messages
  nylas inbound messages abc123 --unread

  # Limit to 5 messages
  nylas inbound messages abc123 --limit 5

  # Output as JSON
  nylas inbound messages abc123 --json

  # Use environment variable for inbox ID
  export NYLAS_INBOUND_GRANT_ID=abc123
  nylas inbound messages`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMessages(args, limit, unread, jsonOutput)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of messages to show")
	cmd.Flags().BoolVarP(&unread, "unread", "u", false, "Show only unread messages")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func runMessages(args []string, limit int, unread bool, jsonOutput bool) error {
	inboxID, err := getInboxID(args)
	if err != nil {
		printError("%v", err)
		return err
	}

	_, err = common.WithClientNoGrant(func(ctx context.Context, client ports.NylasClient) (struct{}, error) {
		// Build query params
		params := &domain.MessageQueryParams{
			Limit: limit,
		}
		if unread {
			unreadVal := true
			params.Unread = &unreadVal
		}

		messages, err := client.GetInboundMessages(ctx, inboxID, params)
		if err != nil {
			return struct{}{}, common.WrapListError("messages", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(messages, "", "  ")
			fmt.Println(string(data))
			return struct{}{}, nil
		}

		if len(messages) == 0 {
			if unread {
				common.PrintEmptyState("unread messages")
			} else {
				common.PrintEmptyStateWithHint("messages", "Send an email to the inbox address to receive messages here")
			}
			return struct{}{}, nil
		}

		// Count unread
		unreadCount := 0
		for _, msg := range messages {
			if msg.Unread {
				unreadCount++
			}
		}

		if unread {
			_, _ = common.BoldWhite.Printf("Unread Messages (%d)\n\n", len(messages))
		} else {
			_, _ = common.BoldWhite.Printf("Messages (%d total, %d unread)\n\n", len(messages), unreadCount)
		}

		for i, msg := range messages {
			printInboundMessageSummary(msg, i)
		}

		fmt.Println()
		_, _ = common.Dim.Println("Use 'nylas email read <message-id> [inbox-id]' to view full message")

		return struct{}{}, nil
	})

	return err
}
