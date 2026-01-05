package email

import (
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	client, _ := getClient()

	return common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <message-id> [grant-id]",
		Short:        "Delete an email message",
		Long:         "Delete an email message (moves to trash).",
		ResourceName: "message",
		DeleteFunc:   client.DeleteMessage,
		GetClient:    getClient,
	})
}
