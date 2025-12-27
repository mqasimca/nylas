package auth

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <email-or-grant-id>",
		Short: "Switch active grant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			grantSvc, _, err := createGrantService()
			if err != nil {
				return err
			}

			identifier := args[0]

			// Try as email first
			if strings.Contains(identifier, "@") {
				if err := grantSvc.SwitchGrantByEmail(identifier); err != nil {
					return err
				}
			} else {
				if err := grantSvc.SwitchGrant(identifier); err != nil {
					return err
				}
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Switched to %s\n", identifier)

			return nil
		},
	}
}
