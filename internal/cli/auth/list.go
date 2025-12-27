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

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all authenticated accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			grantSvc, _, err := createGrantService()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			grants, err := grantSvc.ListGrants(ctx)
			if err != nil {
				return err
			}

			if len(grants) == 0 {
				fmt.Println("No authenticated accounts")
				return nil
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(grants)
			}

			verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

			green := color.New(color.FgGreen)
			red := color.New(color.FgRed)
			yellow := color.New(color.FgYellow)
			dim := color.New(color.Faint)
			bold := color.New(color.Bold)

			// Print header
			_, _ = bold.Printf("  %-38s  %-24s  %-12s  %-12s  %s\n", "GRANT ID", "EMAIL", "PROVIDER", "STATUS", "DEFAULT")

			for _, g := range grants {
				// Print fixed-width columns first
				fmt.Printf("  %-38s  %-24s  %-12s  ",
					g.ID, g.Email, g.Provider.DisplayName())

				// Print status with color (fixed 12 char width)
				switch g.Status {
				case "valid":
					_, _ = green.Print("✓ valid     ")
				case "error":
					_, _ = red.Print("✗ error     ")
				case "revoked":
					_, _ = red.Print("✗ revoked   ")
				default:
					_, _ = yellow.Printf("%-12s", g.Status)
				}

				// Print default indicator
				fmt.Print("  ")
				if g.IsDefault {
					_, _ = green.Print("✓")
				}
				fmt.Println()

				// Show error details in verbose mode
				if verbose && g.Error != "" {
					_, _ = dim.Printf("    Error: %s\n", g.Error)
				}
			}

			return nil
		},
	}
}
