package auth

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke current authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			authSvc, _, err := createAuthService()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := authSvc.Logout(ctx); err != nil {
				return err
			}

			green := color.New(color.FgGreen)
			green.Println("✓ Successfully logged out")

			return nil
		},
	}
}
