package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/mqasimca/nylas/internal/tui"
)

// NewTUICmd creates the tui command.
func NewTUICmd() *cobra.Command {
	var refreshInterval int

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive terminal UI",
		Long: `Launch a k9s-style terminal interface for managing your Nylas email.

The TUI provides:
  - Real-time email list with auto-refresh
  - Keyboard-driven navigation (vim-style: j/k)
  - Read, star, and manage messages
  - Resource views for messages, events, contacts, webhooks, grants

Navigation:
  ↑/k, ↓/j    Move up/down
  g/G         Go to top/bottom
  enter       Open/select
  esc         Go back
  :           Command mode
  /           Filter
  ?           Help
  q           Quit`,
		Example: `  # Launch TUI
  nylas tui

  # Launch with custom refresh interval
  nylas tui --refresh 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(time.Duration(refreshInterval) * time.Second)
		},
	}

	cmd.Flags().IntVar(&refreshInterval, "refresh", 3, "Refresh interval in seconds")

	return cmd
}

func runTUI(refreshInterval time.Duration) error {
	// Load config
	configStore := config.NewDefaultFileStore()
	cfg, err := configStore.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize secret store
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return fmt.Errorf("failed to initialize secret store: %w", err)
	}

	// Get API key
	apiKey, err := secretStore.Get(ports.KeyAPIKey)
	if err != nil {
		return fmt.Errorf("API key not configured. Run 'nylas auth config' first")
	}

	// Get credentials
	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)

	// Create Nylas client
	client := nylas.NewHTTPClient()
	client.SetRegion(cfg.Region)
	client.SetCredentials(clientID, clientSecret, apiKey)

	// Get default grant
	grantStore := keyring.NewGrantStore(secretStore)
	grantID, err := grantStore.GetDefaultGrant()
	if err != nil {
		return fmt.Errorf("no default grant set. Run 'nylas auth login' first")
	}

	// Get grant email for display
	grantInfo, err := grantStore.GetGrant(grantID)
	if err != nil {
		return fmt.Errorf("failed to get grant info: %w", err)
	}

	// Create TUI app (k9s-style using tview)
	app := tui.NewApp(tui.Config{
		Client:          client,
		GrantID:         grantID,
		Email:           grantInfo.Email,
		Provider:        string(grantInfo.Provider),
		RefreshInterval: refreshInterval,
	})

	// Run the application
	return app.Run()
}
