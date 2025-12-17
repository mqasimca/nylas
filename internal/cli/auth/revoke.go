package auth

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <grant-id>",
		Short: "Revoke a specific grant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			grantID := args[0]

			authSvc, _, err := createAuthService()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := authSvc.LogoutGrant(ctx, grantID); err != nil {
				return err
			}

			green := color.New(color.FgGreen)
			green.Printf("✓ Grant %s revoked\n", grantID)

			return nil
		},
	}
}
