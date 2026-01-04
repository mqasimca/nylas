package otp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/cli/common"
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
				common.PrintEmptyStateWithHint("accounts", "Run 'nylas auth login' to add an account")
				return nil
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(accounts)
			}

			_, _ = common.BoldCyan.Println("Configured Accounts")
			fmt.Println()

			// Print header
			_, _ = common.Bold.Printf("  %-3s  %-28s  %-12s\n", "#", "EMAIL", "PROVIDER")

			for i, acc := range accounts {
				fmt.Printf("  %-3d  %-28s  %-12s\n",
					i+1, acc.Email, acc.Provider.DisplayName())
			}

			fmt.Printf("\n  %d account(s) configured\n", len(accounts))

			return nil
		},
	}
}
