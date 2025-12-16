// Package auth provides the auth subcommands.
package auth

import (
	"github.com/spf13/cobra"
)

// NewAuthCmd creates the auth command group.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long: `Manage Nylas API authentication.

Commands:
  login     Authenticate with an email provider via OAuth
  logout    Revoke the current authentication
  status    Show current authentication status
  whoami    Show current user info
  list      List all authenticated accounts
  switch    Switch between authenticated accounts
  add       Manually add an existing grant
  config    Configure API credentials`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newWhoamiCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSwitchCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newRevokeCmd())
	cmd.AddCommand(newTokenCmd())

	return cmd
}
