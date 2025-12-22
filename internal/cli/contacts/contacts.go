// Package contacts provides contacts-related CLI commands.
package contacts

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

var client ports.NylasClient

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
	if client != nil {
		return client, nil
	}

	c, err := common.GetNylasClient()
	if err != nil {
		return nil, err
	}

	client = c
	return client, nil
}

func getGrantID(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return "", err
	}

	grantStore := keyring.NewGrantStore(secretStore)
	return grantStore.GetDefaultGrant()
}

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
