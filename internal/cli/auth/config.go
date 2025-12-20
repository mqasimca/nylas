package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	nylasadapter "github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
)

func newConfigCmd() *cobra.Command {
	var (
		region       string
		clientID     string
		clientSecret string
		apiKey       string
		reset        bool
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure API credentials",
		Long: `Configure Nylas API credentials.

You can provide credentials via flags or interactively.
Get your credentials from https://dashboard.nylas.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configSvc, _, _, err := createConfigService()
			if err != nil {
				return err
			}

			if reset {
				if err := configSvc.ResetConfig(); err != nil {
					return err
				}
				green := color.New(color.FgGreen)
				green.Println("✓ Configuration reset")
				return nil
			}

			// Interactive mode if credentials not provided
			if clientID == "" || apiKey == "" {
				reader := bufio.NewReader(os.Stdin)

				fmt.Println("Configure Nylas API Credentials")
				fmt.Println("Get your credentials from: https://dashboard.nylas.com")
				fmt.Println()

				if clientID == "" {
					fmt.Print("Client ID: ")
					input, _ := reader.ReadString('\n')
					clientID = strings.TrimSpace(input)
				}

				if apiKey == "" {
					fmt.Print("API Key (hidden): ")
					// Use password masking for API key
					apiKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						return fmt.Errorf("failed to read API key: %w", err)
					}
					fmt.Println() // Add newline after hidden input
					apiKey = strings.TrimSpace(string(apiKeyBytes))
				}

				if clientSecret == "" {
					fmt.Print("Client Secret (optional, hidden - press Enter to skip): ")
					// Use password masking for client secret
					secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						return fmt.Errorf("failed to read client secret: %w", err)
					}
					fmt.Println() // Add newline after hidden input
					clientSecret = strings.TrimSpace(string(secretBytes))
				}

				if region == "" {
					fmt.Print("Region [us/eu] (default: us): ")
					input, _ := reader.ReadString('\n')
					region = strings.TrimSpace(input)
					if region == "" {
						region = "us"
					}
				}
			}

			if clientID == "" {
				return fmt.Errorf("client ID is required")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required")
			}

			if err := configSvc.SetupConfig(region, clientID, clientSecret, apiKey); err != nil {
				return err
			}

			green := color.New(color.FgGreen)
			green.Println("✓ Configuration saved")

			// Auto-detect existing grants from Nylas API
			fmt.Println()
			fmt.Println("Checking for existing grants...")

			client := nylasadapter.NewHTTPClient()
			client.SetRegion(region)
			client.SetCredentials(clientID, clientSecret, apiKey)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			grants, err := client.ListGrants(ctx)
			if err != nil {
				yellow := color.New(color.FgYellow)
				yellow.Printf("  Could not fetch grants: %v\n", err)
				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  nylas auth login    Authenticate with your email provider")
				return nil
			}

			if len(grants) == 0 {
				fmt.Println("  No existing grants found")
				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  nylas auth login    Authenticate with your email provider")
				return nil
			}

			// Get grant store to save grants locally
			grantStore, err := createGrantStore()
			if err != nil {
				yellow := color.New(color.FgYellow)
				yellow.Printf("  Could not save grants locally: %v\n", err)
				return nil
			}

			// Add all valid grants, first one becomes default
			addedCount := 0
			for i, grant := range grants {
				if !grant.IsValid() {
					continue
				}

				grantInfo := domain.GrantInfo{
					ID:       grant.ID,
					Email:    grant.Email,
					Provider: grant.Provider,
				}

				if err := grantStore.SaveGrant(grantInfo); err != nil {
					continue
				}

				// Set first grant as default
				if addedCount == 0 {
					_ = grantStore.SetDefaultGrant(grant.ID)
				}

				addedCount++
				if i == 0 {
					green.Printf("  ✓ Added %s (%s) [default]\n", grant.Email, grant.Provider.DisplayName())
				} else {
					green.Printf("  ✓ Added %s (%s)\n", grant.Email, grant.Provider.DisplayName())
				}
			}

			if addedCount > 0 {
				fmt.Println()
				fmt.Printf("Added %d grant(s). Run 'nylas auth list' to see all accounts.\n", addedCount)
			} else {
				fmt.Println("  No valid grants found")
				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  nylas auth login    Authenticate with your email provider")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&region, "region", "r", "us", "API region (us or eu)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Nylas Client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Nylas Client Secret")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Nylas API Key")
	cmd.Flags().BoolVar(&reset, "reset", false, "Reset all configuration")

	return cmd
}
