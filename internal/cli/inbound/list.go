package inbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all inbound inboxes",
		Long: `List all inbound inboxes for your Nylas application.

Inbound inboxes are managed email addresses that can receive emails without
requiring OAuth authentication.

Examples:
  # List all inbound inboxes
  nylas inbound list

  # List inboxes as JSON
  nylas inbound list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func runList(jsonOutput bool) error {
	_, err := common.WithClientNoGrant(func(ctx context.Context, client ports.NylasClient) (struct{}, error) {
		inboxes, err := client.ListInboundInboxes(ctx)
		if err != nil {
			return struct{}{}, common.WrapListError("inboxes", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(inboxes, "", "  ")
			fmt.Println(string(data))
			return struct{}{}, nil
		}

		if len(inboxes) == 0 {
			common.PrintEmptyStateWithHint("inboxes", "Create one with: nylas inbound create <email-prefix>")
			return struct{}{}, nil
		}

		_, _ = common.BoldWhite.Printf("Inbound Inboxes (%d)\n\n", len(inboxes))

		for i, inbox := range inboxes {
			printInboxSummary(inbox, i)
		}

		fmt.Println()
		_, _ = common.Dim.Println("Use 'nylas inbound messages [inbox-id]' to view messages")

		return struct{}{}, nil
	})
	return err
}
