package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
					fmt.Print("API Key: ")
					input, _ := reader.ReadString('\n')
					apiKey = strings.TrimSpace(input)
				}

				if clientSecret == "" {
					fmt.Print("Client Secret (optional, press Enter to skip): ")
					input, _ := reader.ReadString('\n')
					clientSecret = strings.TrimSpace(input)
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
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  nylas auth login    Authenticate with your email provider")

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
