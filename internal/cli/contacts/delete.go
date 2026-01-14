package contacts

import (
	"context"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <contact-id> [grant-id]",
		Aliases:      []string{"rm", "remove"},
		Short:        "Delete a contact",
		Long:         "Delete a contact by its ID.",
		ResourceName: "contact",
		DeleteFunc: func(ctx context.Context, grantID, resourceID string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return client.DeleteContact(ctx, grantID, resourceID)
		},
		GetClient: getClient,
	})
}
