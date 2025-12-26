package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			grantSvc, configSvc, err := createGrantService()
			if err != nil {
				return err
			}

			status, err := configSvc.GetStatus()
			if err != nil {
				return err
			}

			// Get current grant info
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var grantInfo struct {
				ID       string `json:"id"`
				Email    string `json:"email"`
				Provider string `json:"provider"`
				Status   string `json:"status"`
			}

			grant, err := grantSvc.GetCurrentGrant(ctx)
			if err == nil {
				grantInfo.ID = grant.ID
				grantInfo.Email = grant.Email
				grantInfo.Provider = string(grant.Provider)
				grantInfo.Status = grant.Status
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				output := map[string]any{
					"configured":    status.IsConfigured,
					"region":        status.Region,
					"config_path":   status.ConfigPath,
					"secret_store":  status.SecretStore,
					"grant_count":   status.GrantCount,
					"default_grant": status.DefaultGrant,
				}
				if grantInfo.ID != "" {
					output["grant"] = grantInfo
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)

			bold.Println("Authentication Status")
			fmt.Println()

			if grantInfo.ID != "" {
				bold.Println("Current Account:")
				fmt.Printf("  Email: %s\n", grantInfo.Email)
				fmt.Printf("  Provider: %s\n", grantInfo.Provider)
				fmt.Printf("  Grant ID: %s\n", grantInfo.ID)
				if grantInfo.Status == "valid" {
					green.Printf("  Status: ✓ Valid\n")
				} else {
					yellow.Printf("  Status: %s\n", grantInfo.Status)
				}
				fmt.Println()
			}

			bold.Println("Configuration:")
			fmt.Printf("  Region: %s\n", status.Region)
			fmt.Printf("  Config Path: %s\n", status.ConfigPath)
			fmt.Printf("  Secret Store: %s\n", status.SecretStore)

			return nil
		},
	}
}
