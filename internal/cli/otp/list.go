package otp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			otpSvc, err := createOTPService()
			if err != nil {
				return err
			}

			accounts, err := otpSvc.ListAccounts()
			if err != nil {
				return err
			}

			if len(accounts) == 0 {
				fmt.Println("No accounts configured")
				fmt.Println("Run 'nylas auth login' to add an account")
				return nil
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(accounts)
			}

			cyan := color.New(color.FgCyan, color.Bold)
			bold := color.New(color.Bold)

			_, _ = cyan.Println("Configured Accounts")
			fmt.Println()

			// Print header
			_, _ = bold.Printf("  %-3s  %-28s  %-12s\n", "#", "EMAIL", "PROVIDER")

			for i, acc := range accounts {
				fmt.Printf("  %-3d  %-28s  %-12s\n",
					i+1, acc.Email, acc.Provider.DisplayName())
			}

			fmt.Printf("\n  %d account(s) configured\n", len(accounts))

			return nil
		},
	}
}
