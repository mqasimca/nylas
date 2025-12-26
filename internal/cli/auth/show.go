package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [grant-id]",
		Short: "Show detailed grant information",
		Long: `Show detailed information about a grant.

If no grant-id is specified, shows the current/default grant.

Information includes:
  - Grant ID and email
  - Provider (Google, Microsoft, etc.)
  - Grant status
  - Scopes/permissions
  - Creation and update timestamps`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configStore := config.NewDefaultFileStore()
			cfg, err := configStore.Load()
			if err != nil {
				cfg = &domain.Config{Region: "us"}
			}

			// Check environment variables first (highest priority)
			apiKey, clientID, clientSecret := getCredentialsFromEnv()

			// If API key not in env, try keyring/file store
			if apiKey == "" {
				secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
				if err != nil {
					return fmt.Errorf("not authenticated: run 'nylas auth config' first")
				}

				apiKey, err = secretStore.Get(ports.KeyAPIKey)
				if err != nil {
					return fmt.Errorf("not authenticated: run 'nylas auth config' first")
				}

				if clientID == "" {
					clientID, _ = secretStore.Get(ports.KeyClientID)
				}
				if clientSecret == "" {
					clientSecret, _ = secretStore.Get(ports.KeyClientSecret)
				}
			}

			client := nylas.NewHTTPClient()
			client.SetRegion(cfg.Region)
			client.SetCredentials(clientID, clientSecret, apiKey)

			// Get grant ID
			var grantID string
			var grantStore ports.GrantStore
			if len(args) > 0 {
				grantID = args[0]
			} else {
				// Need to access local grant store for default grant
				secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
				if err != nil {
					return fmt.Errorf("no default grant set: specify a grant ID or run 'nylas auth login'")
				}
				grantStore = keyring.NewGrantStore(secretStore)
				grantID, err = grantStore.GetDefaultGrant()
				if err != nil {
					return fmt.Errorf("no default grant set: specify a grant ID or run 'nylas auth login'")
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			grant, err := client.GetGrant(ctx, grantID)
			if err != nil {
				return fmt.Errorf("failed to get grant details: %w", err)
			}

			// Display grant information
			boldWhite := color.New(color.Bold, color.FgWhite)
			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			dim := color.New(color.Faint)

			fmt.Println("════════════════════════════════════════════════════════════")
			boldWhite.Printf("Grant Details\n")
			fmt.Println("════════════════════════════════════════════════════════════")

			fmt.Printf("\n")
			fmt.Printf("Grant ID:    %s\n", cyan.Sprint(grant.ID))
			fmt.Printf("Email:       %s\n", grant.Email)
			fmt.Printf("Provider:    %s\n", formatProvider(string(grant.Provider)))

			// Status with color
			statusColor := green
			statusIcon := "✓"
			if grant.GrantStatus != "valid" {
				statusColor = yellow
				statusIcon = "⚠"
			}
			fmt.Printf("Status:      %s %s\n", statusIcon, statusColor.Sprint(grant.GrantStatus))

			// Timestamps
			if !grant.CreatedAt.Time.IsZero() {
				fmt.Printf("\nCreated:     %s\n", grant.CreatedAt.Time.Format("Jan 2, 2006 3:04 PM"))
			}
			if !grant.UpdatedAt.Time.IsZero() {
				fmt.Printf("Updated:     %s\n", grant.UpdatedAt.Time.Format("Jan 2, 2006 3:04 PM"))
			}

			// Scopes
			if len(grant.Scope) > 0 {
				fmt.Printf("\nScopes:\n")
				for _, scope := range grant.Scope {
					fmt.Printf("  %s %s\n", dim.Sprint("•"), scope)
				}
			}

			// Check if this is the default grant (only if we have access to grant store)
			if grantStore != nil {
				defaultGrant, _ := grantStore.GetDefaultGrant()
				if defaultGrant == grant.ID {
					fmt.Printf("\n%s This is the default grant\n", green.Sprint("★"))
				}
			}

			return nil
		},
	}
}

func formatProvider(provider string) string {
	switch strings.ToLower(provider) {
	case "google":
		return "Google"
	case "microsoft":
		return "Microsoft 365"
	case "imap":
		return "IMAP"
	case "ews":
		return "Exchange (EWS)"
	case "yahoo":
		return "Yahoo"
	case "icloud":
		return "iCloud"
	case "zoom":
		return "Zoom"
	default:
		return provider
	}
}
