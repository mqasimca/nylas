// Package contacts provides contacts-related CLI commands.
package contacts

import (
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

// NewContactsCmd creates the contacts command group.
func NewContactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contacts",
		Aliases: []string{"contact"},
		Short:   "Manage contacts",
		Long: `Manage contacts from your connected accounts.

View contacts, create new contacts, update and delete contacts.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGroupsCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newPhotoCmd())
	cmd.AddCommand(newSyncCmd())

	return cmd
}

func getClient() (ports.NylasClient, error) {
	// Delegate to common.GetNylasClient() which handles caching internally
	return common.GetNylasClient()
}

// getGrantID gets the grant ID from args or default.
// Delegates to common.GetGrantID for consistent behavior across CLI commands.
func getGrantID(args []string) (string, error) {
	return common.GetGrantID(args)
}
