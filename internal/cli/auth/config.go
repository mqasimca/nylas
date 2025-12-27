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
		region   string
		clientID string
		apiKey   string
		reset    bool
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure API credentials",
		Long: `Configure Nylas API credentials.

You can provide credentials via flags or interactively.
Get your credentials from https://dashboardv3.nylas.com

The CLI only requires your API Key - Client ID is auto-detected.`,
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
				// #nosec G104 -- color output errors are non-critical, best-effort display
				green.Println("✓ Configuration reset")
				return nil
			}

			reader := bufio.NewReader(os.Stdin)

			// Interactive mode if API key not provided
			if apiKey == "" {
				fmt.Println("Configure Nylas API Credentials")
				fmt.Println("Get your API key from: https://dashboardv3.nylas.com")
				fmt.Println()

				fmt.Print("API Key (hidden): ")
				apiKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read API key: %w", err)
				}
				fmt.Println()
				apiKey = strings.TrimSpace(string(apiKeyBytes))
			}

			if apiKey == "" {
				return fmt.Errorf("API key is required")
			}

			// Get region if not provided
			if region == "" {
				fmt.Print("Region [us/eu] (default: us): ")
				input, _ := reader.ReadString('\n')
				region = strings.TrimSpace(input)
				if region == "" {
					region = "us"
				}
			}

			// Auto-detect Client ID from API key if not provided
			if clientID == "" {
				fmt.Println()
				fmt.Println("Detecting applications...")

				client := nylasadapter.NewHTTPClient()
				client.SetRegion(region)
				client.SetCredentials("", "", apiKey) // Only API key needed for ListApplications

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				apps, err := client.ListApplications(ctx)
				cancel()

				if err != nil {
					yellow := color.New(color.FgYellow)
					// #nosec G104 -- color output errors are non-critical, best-effort display
					yellow.Printf("  Could not auto-detect Client ID: %v\n", err)
					fmt.Println()
					fmt.Print("Client ID (manual entry): ")
					input, _ := reader.ReadString('\n')
					clientID = strings.TrimSpace(input)
				} else if len(apps) == 0 {
					return fmt.Errorf("no applications found for this API key")
				} else if len(apps) == 1 {
					// Single app - auto-select
					app := apps[0]
					clientID = getAppClientID(app)
					green := color.New(color.FgGreen)
					green.Printf("  ✓ Found application: %s\n", getAppDisplayName(app))
				} else {
					// Multiple apps - let user choose
					fmt.Printf("  Found %d applications:\n\n", len(apps))
					for i, app := range apps {
						fmt.Printf("  [%d] %s\n", i+1, getAppDisplayName(app))
					}
					fmt.Println()
					fmt.Print("Select application (1-", len(apps), "): ")
					input, _ := reader.ReadString('\n')
					choice := strings.TrimSpace(input)

					var selected int
					if _, err := fmt.Sscanf(choice, "%d", &selected); err != nil || selected < 1 || selected > len(apps) {
						return fmt.Errorf("invalid selection: %s", choice)
					}

					app := apps[selected-1]
					clientID = getAppClientID(app)
					green := color.New(color.FgGreen)
					green.Printf("  ✓ Selected: %s\n", getAppDisplayName(app))
				}
			}

			if clientID == "" {
				return fmt.Errorf("client ID is required")
			}

			if err := configSvc.SetupConfig(region, clientID, "", apiKey); err != nil {
				return err
			}

			green := color.New(color.FgGreen)
			green.Println("✓ Configuration saved")

			// Auto-detect existing grants from Nylas API
			fmt.Println()
			fmt.Println("Checking for existing grants...")

			client := nylasadapter.NewHTTPClient()
			client.SetRegion(region)
			client.SetCredentials(clientID, "", apiKey)

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
	cmd.Flags().StringVar(&clientID, "client-id", "", "Nylas Client ID (auto-detected if not provided)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Nylas API Key")
	cmd.Flags().BoolVar(&reset, "reset", false, "Reset all configuration")

	return cmd
}

// getAppClientID returns the client ID for an application.
// It checks both ID and ApplicationID fields since the API may use either.
func getAppClientID(app domain.Application) string {
	if app.ApplicationID != "" {
		return app.ApplicationID
	}
	return app.ID
}

// getAppDisplayName returns a human-readable display name for an application.
func getAppDisplayName(app domain.Application) string {
	clientID := getAppClientID(app)
	env := app.Environment
	if env == "" {
		env = "production"
	}

	region := app.Region
	if region == "" {
		region = "us"
	}

	// Truncate client ID for display if too long
	displayID := clientID
	if len(displayID) > 20 {
		displayID = displayID[:17] + "..."
	}

	return fmt.Sprintf("%s (%s, %s)", displayID, env, region)
}
