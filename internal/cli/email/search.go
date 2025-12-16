package email

import (
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var limit int
	var from string
	var to string
	var subject string
	var after string
	var before string
	var hasAttachment bool

	cmd := &cobra.Command{
		Use:   "search <query> [grant-id]",
		Short: "Search emails",
		Long:  "Search for emails matching a query string or filters.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			var grantID string
			if len(args) > 1 {
				grantID = args[1]
			} else {
				grantID, err = getGrantID(nil)
				if err != nil {
					return err
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			params := &domain.MessageQueryParams{
				Limit: limit,
			}

			// Use query as subject search since Nylas v3 doesn't support full-text 'q'
			// If --subject flag is also provided, it takes precedence
			if subject != "" {
				params.Subject = subject
			} else {
				params.Subject = query
			}

			if from != "" {
				params.From = from
			}
			if to != "" {
				params.To = to
			}
			if cmd.Flags().Changed("has-attachment") {
				params.HasAttachment = &hasAttachment
			}

			// Parse date filters
			if after != "" {
				t, err := parseDate(after)
				if err != nil {
					return fmt.Errorf("invalid 'after' date: %w", err)
				}
				params.ReceivedAfter = t.Unix()
			}
			if before != "" {
				t, err := parseDate(before)
				if err != nil {
					return fmt.Errorf("invalid 'before' date: %w", err)
				}
				params.ReceivedBefore = t.Unix()
			}

			messages, err := client.GetMessagesWithParams(ctx, grantID, params)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(messages) == 0 {
				fmt.Println("No messages found matching your search.")
				return nil
			}

			fmt.Printf("Found %d messages:\n\n", len(messages))
			for i, msg := range messages {
				printMessageSummary(msg, i+1)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum number of results")
	cmd.Flags().StringVar(&from, "from", "", "Filter by sender")
	cmd.Flags().StringVar(&to, "to", "", "Filter by recipient")
	cmd.Flags().StringVar(&subject, "subject", "", "Filter by subject")
	cmd.Flags().StringVar(&after, "after", "", "Messages after date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&before, "before", "", "Messages before date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&hasAttachment, "has-attachment", false, "Only messages with attachments")

	return cmd
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
