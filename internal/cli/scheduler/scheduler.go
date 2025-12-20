// Package scheduler provides scheduler-related CLI commands.
package scheduler

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

// NewSchedulerCmd creates the scheduler command group.
func NewSchedulerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scheduler",
		Aliases: []string{"sched"},
		Short:   "Manage Nylas Scheduler",
		Long: `Manage Nylas Scheduler configurations, sessions, bookings, and pages.

The Nylas Scheduler allows you to create meeting booking workflows,
manage availability, and handle scheduling sessions.`,
	}

	cmd.AddCommand(newConfigurationsCmd())
	cmd.AddCommand(newSessionsCmd())
	cmd.AddCommand(newBookingsCmd())
	cmd.AddCommand(newPagesCmd())

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
