// Package admin provides admin-related CLI commands.
package admin

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

var client ports.NylasClient

// NewAdminCmd creates the admin command group.
func NewAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administration commands (requires API key)",
		Long: `Administration commands for managing applications, connectors, credentials, and grants.

These commands require API key authentication and are used for managing
the Nylas platform at an organizational level.`,
	}

	cmd.AddCommand(newApplicationsCmd())
	cmd.AddCommand(newConnectorsCmd())
	cmd.AddCommand(newCredentialsCmd())
	cmd.AddCommand(newGrantsCmd())

	return cmd
}

func getClient() (ports.NylasClient, error) {
	if client != nil {
		return client, nil
	}

	configStore := config.NewDefaultFileStore()
	cfg, err := configStore.Load()
	if err != nil {
		cfg = &domain.Config{Region: "us"}
	}

	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return nil, err
	}

	apiKey, err := secretStore.Get(ports.KeyAPIKey)
	if err != nil {
		return nil, err
	}

	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)

	c := nylas.NewHTTPClient()
	c.SetRegion(cfg.Region)
	c.SetCredentials(clientID, clientSecret, apiKey)

	return c, nil
}

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
