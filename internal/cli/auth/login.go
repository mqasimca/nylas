package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/domain"
)

func newLoginCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with an email provider",
		Long: `Authenticate with an email provider via OAuth.

Supported providers:
  google     Google/Gmail
  microsoft  Microsoft/Outlook`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate provider
			p, err := domain.ParseProvider(provider)
			if err != nil {
				return fmt.Errorf("invalid provider: %s (use 'google' or 'microsoft')", provider)
			}

			// Check if configured
			configSvc, _, _, err := createConfigService()
			if err != nil {
				return err
			}

			if !configSvc.IsConfigured() {
				return fmt.Errorf("nylas not configured - run 'nylas auth config' first")
			}

			// Create auth service
			authSvc, _, err := createAuthService()
			if err != nil {
				return err
			}

			fmt.Println("Opening browser for authentication...")
			fmt.Println("Complete the sign-in process in your browser.")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			grant, err := authSvc.Login(ctx, p)
			if err != nil {
				return err
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("\n✓ Successfully authenticated!\n")
			fmt.Printf("  Email:    %s\n", grant.Email)
			fmt.Printf("  Provider: %s\n", grant.Provider.DisplayName())
			fmt.Printf("  Grant ID: %s\n", grant.ID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "google", "Email provider (google, microsoft)")

	return cmd
}
