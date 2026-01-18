package email

import (
	"context"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <message-id> [grant-id]",
		Short:        "Delete an email message",
		Long:         "Delete an email message (moves to trash).",
		ResourceName: "message",
		DeleteFunc: func(ctx context.Context, grantID, resourceID string) error {
			client, err := common.GetNylasClient()
			if err != nil {
				return err
			}
			return client.DeleteMessage(ctx, grantID, resourceID)
		},
		GetClient: common.GetNylasClient,
	})
}
