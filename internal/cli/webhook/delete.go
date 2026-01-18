package webhook

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <webhook-id>",
		Aliases:      []string{"rm", "remove"},
		Short:        "Delete a webhook",
		Long:         "Delete a webhook by ID.\n\nThis permanently removes the webhook and stops all event notifications.",
		ResourceName: "webhook",
		DeleteFuncNoGrant: func(ctx context.Context, resourceID string) error {
			client, err := common.GetNylasClient()
			if err != nil {
				return err
			}
			return client.DeleteWebhook(ctx, resourceID)
		},
		GetClient: common.GetNylasClient,
		GetDetailsFunc: func(ctx context.Context, resourceID string) (string, error) {
			client, err := common.GetNylasClient()
			if err != nil {
				return "", err
			}
			webhook, err := client.GetWebhook(ctx, resourceID)
			if err != nil {
				return "", err
			}
			details := "Webhook to delete:\n"
			details += fmt.Sprintf("  ID:  %s\n", webhook.ID)
			details += fmt.Sprintf("  URL: %s\n", webhook.WebhookURL)
			if webhook.Description != "" {
				details += fmt.Sprintf("  Description: %s\n", webhook.Description)
			}
			details += fmt.Sprintf("  Triggers: %v", webhook.TriggerTypes)
			return details, nil
		},
	})

	// Add examples
	cmd.Example = `  # Delete a webhook (with confirmation)
  nylas webhook delete webhook-abc123

  # Delete without confirmation
  nylas webhook delete webhook-abc123 --force`

	return cmd
}
