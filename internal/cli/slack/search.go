// search.go provides message search functionality for Slack workspaces.

package slack

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
)

// newSearchCmd creates the search command for searching messages.
func newSearchCmd() *cobra.Command {
	var (
		query  string
		limit  int
		showID bool
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search messages",
		Long: `Search for messages in your Slack workspace.

Uses Slack's search syntax. Examples:
  - "from:@alice" - messages from a user
  - "in:#general" - messages in a channel
  - "has:link" - messages with links
  - "before:2024-01-01" - messages before a date

Examples:
  # Search for messages
  nylas slack search --query "project update"

  # Search with Slack modifiers
  nylas slack search --query "from:@alice in:#general"

  # Limit results
  nylas slack search --query "important" --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				return common.NewUserError("search query is required", "Use --query")
			}

			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			messages, err := client.SearchMessages(ctx, query, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(messages) == 0 {
				fmt.Printf("No messages found for: %s\n", query)
				return nil
			}

			cyan := color.New(color.FgCyan)
			_, _ = cyan.Printf("Found %d messages:\n\n", len(messages))

			for _, msg := range messages {
				printMessage(msg, showID, false)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Search query")
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum number of results")
	cmd.Flags().BoolVar(&showID, "id", false, "Show message IDs")

	_ = cmd.MarkFlagRequired("query")

	return cmd
}
