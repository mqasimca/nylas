package contacts

import (
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	// Get client reference for delete function
	client, _ := getClient()

	return common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <contact-id> [grant-id]",
		Aliases:      []string{"rm", "remove"},
		Short:        "Delete a contact",
		Long:         "Delete a contact by its ID.",
		ResourceName: "contact",
		DeleteFunc:   client.DeleteContact,
		GetClient:    getClient,
	})
}
