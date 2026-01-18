package notetaker

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return common.NewDeleteCommand(common.DeleteCommandConfig{
		Use:          "delete <notetaker-id> [grant-id]",
		Aliases:      []string{"rm", "cancel"},
		Short:        "Delete or cancel a notetaker",
		Long:         "Delete a notetaker. If the notetaker is scheduled or active, this will cancel it.\n\nThis action cannot be undone. Once deleted, any recordings or transcripts that haven't been saved will be lost.",
		ResourceName: "notetaker",
		DeleteFunc: func(ctx context.Context, grantID, resourceID string) error {
			client, err := common.GetNylasClient()
			if err != nil {
				return err
			}
			return client.DeleteNotetaker(ctx, grantID, resourceID)
		},
		GetClient: common.GetNylasClient,
		ShowDetailsFunc: func(ctx context.Context, grantID, resourceID string) (string, error) {
			client, err := common.GetNylasClient()
			if err != nil {
				return "", err
			}
			notetaker, err := client.GetNotetaker(ctx, grantID, resourceID)
			if err != nil {
				return "", err
			}
			details := fmt.Sprintf("Delete notetaker %s?", resourceID)
			if notetaker.MeetingTitle != "" {
				details += fmt.Sprintf("\n  Title: %s", notetaker.MeetingTitle)
			}
			details += fmt.Sprintf("\n  State: %s", formatState(notetaker.State))
			details += "\n\nThis action cannot be undone."
			return details, nil
		},
	})
}
