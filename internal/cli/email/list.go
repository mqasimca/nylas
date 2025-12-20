package email

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var limit int
	var unread bool
	var starred bool
	var from string
	var folder string
	var showID bool
	var all bool
	var allFolders bool
	var maxItems int
	var metadataPair string

	cmd := &cobra.Command{
		Use:   "list [grant-id]",
		Short: "List recent emails",
		Long: `List recent emails from your inbox. Use grant-id or the default account.

By default, only shows messages from INBOX. Use --folder to specify a different
folder, or --all-folders to show messages from all folders.

Use --all to fetch all messages (paginated automatically).
Use --max to limit total messages when using --all.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			params := &domain.MessageQueryParams{
				Limit: limit,
			}

			if cmd.Flags().Changed("unread") {
				params.Unread = &unread
			}
			if cmd.Flags().Changed("starred") {
				params.Starred = &starred
			}
			if from != "" {
				params.From = from
			}
			if metadataPair != "" {
				params.MetadataPair = metadataPair
			}

			// Default to INBOX unless --all-folders is set or specific folder is provided
			if folder != "" {
				params.In = []string{folder}
			} else if !allFolders {
				params.In = []string{"INBOX"}
			}

			var messages []domain.Message

			if all {
				// Use pagination to fetch all messages
				pageSize := 50 // Optimal page size for API
				if limit > 0 && limit < pageSize {
					pageSize = limit
				}
				params.Limit = pageSize

				fetcher := func(ctx context.Context, cursor string) (common.PageResult[domain.Message], error) {
					params.PageToken = cursor
					resp, err := client.GetMessagesWithCursor(ctx, grantID, params)
					if err != nil {
						return common.PageResult[domain.Message]{}, err
					}
					return common.PageResult[domain.Message]{
						Data:       resp.Data,
						NextCursor: resp.Pagination.NextCursor,
					}, nil
				}

				config := common.DefaultPaginationConfig()
				config.PageSize = pageSize
				config.MaxItems = maxItems

				messages, err = common.FetchAllPages(ctx, config, fetcher)
				if err != nil {
					return fmt.Errorf("failed to fetch messages: %w", err)
				}
			} else {
				// Standard single-page fetch
				messages, err = client.GetMessagesWithParams(ctx, grantID, params)
				if err != nil {
					return fmt.Errorf("failed to get messages: %w", err)
				}
			}

			if len(messages) == 0 {
				fmt.Println("No messages found.")
				return nil
			}

			fmt.Printf("Found %d messages:\n\n", len(messages))
			for i, msg := range messages {
				printMessageSummaryWithID(msg, i+1, showID)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of messages to fetch (per page with --all)")
	cmd.Flags().BoolVarP(&unread, "unread", "u", false, "Only show unread messages")
	cmd.Flags().BoolVarP(&starred, "starred", "s", false, "Only show starred messages")
	cmd.Flags().StringVarP(&from, "from", "f", "", "Filter by sender email")
	cmd.Flags().StringVar(&folder, "folder", "", "Filter by folder (e.g., INBOX, SENT, TRASH, or folder ID)")
	cmd.Flags().BoolVar(&allFolders, "all-folders", false, "Show messages from all folders (default: INBOX only)")
	cmd.Flags().BoolVar(&showID, "id", false, "Show message IDs")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Fetch all messages (paginated)")
	cmd.Flags().IntVar(&maxItems, "max", 0, "Maximum messages to fetch with --all (0=unlimited)")
	cmd.Flags().StringVar(&metadataPair, "metadata", "", "Filter by metadata (format: key:value, only key1-key5 supported)")

	return cmd
}
