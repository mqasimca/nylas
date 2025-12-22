// Package webhook provides webhook management CLI commands.
package webhook

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

var client ports.NylasClient

// NewWebhookCmd creates the webhook command group.
func NewWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook",
		Aliases: []string{"webhooks", "wh"},
		Short:   "Manage webhooks",
		Long: `Manage Nylas webhooks for event notifications.

Webhooks allow you to receive real-time notifications when events occur,
such as new messages, calendar events, or contact changes.

Note: Webhook management requires an API key (admin-level access).`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newTriggersCmd())
	cmd.AddCommand(newServerCmd())

	return cmd
}

func getClient() (ports.NylasClient, error) {
	if client != nil {
		return client, nil
	}

	// Use common client initialization which supports both keyring and env vars
	return common.GetNylasClient()
}

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
