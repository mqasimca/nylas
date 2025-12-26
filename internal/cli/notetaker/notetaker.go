package notetaker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

// NewNotetakerCmd creates the notetaker command and its subcommands.
func NewNotetakerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notetaker",
		Aliases: []string{"nt", "bot"},
		Short:   "Manage Nylas Notetaker bots",
		Long: `Manage Nylas Notetaker bots for meeting recording and transcription.

Notetaker bots can join video meetings (Zoom, Google Meet, Teams) to:
- Record the meeting
- Generate transcripts
- Provide meeting summaries

Use subcommands to create, list, show, delete notetakers and retrieve media.`,
		Example: `  # List all notetakers
  nylas notetaker list

  # Create a notetaker to join a meeting
  nylas notetaker create --meeting-link "https://zoom.us/j/123456789"

  # Show notetaker details
  nylas notetaker show <notetaker-id>

  # Get recording/transcript
  nylas notetaker media <notetaker-id>

  # Delete/cancel a notetaker
  nylas notetaker delete <notetaker-id>`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newMediaCmd())

	return cmd
}

// getClient creates and configures a Nylas client.
func getClient() (ports.NylasClient, error) {
	configStore := config.NewDefaultFileStore()
	cfg, _ := configStore.Load()

	// Check environment variables first (highest priority)
	apiKey := os.Getenv("NYLAS_API_KEY")
	clientID := os.Getenv("NYLAS_CLIENT_ID")
	clientSecret := os.Getenv("NYLAS_CLIENT_SECRET")

	// If API key not in env, try keyring/file store
	if apiKey == "" {
		secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
		if err == nil {
			apiKey, _ = secretStore.Get(ports.KeyAPIKey)
			if clientID == "" {
				clientID, _ = secretStore.Get(ports.KeyClientID)
			}
			if clientSecret == "" {
				clientSecret, _ = secretStore.Get(ports.KeyClientSecret)
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured. Set NYLAS_API_KEY environment variable or run 'nylas auth config'")
	}

	client := nylas.NewHTTPClient()
	client.SetRegion(cfg.Region)
	client.SetCredentials(clientID, clientSecret, apiKey)

	return client, nil
}

// getGrantID gets the grant ID from args or default.
// If the argument contains '@', it's treated as an email and looked up.
func getGrantID(args []string) (string, error) {
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return "", fmt.Errorf("couldn't access secret store: %w", err)
	}
	grantStore := keyring.NewGrantStore(secretStore)

	if len(args) > 0 {
		identifier := args[0]

		// If it looks like an email, try to find by email
		if strings.Contains(identifier, "@") {
			grant, err := grantStore.GetGrantByEmail(identifier)
			if err != nil {
				return "", fmt.Errorf("no grant found for email: %s", identifier)
			}
			return grant.ID, nil
		}

		// Otherwise treat as grant ID
		return identifier, nil
	}

	// Try to get default grant
	defaultGrant, err := grantStore.GetDefaultGrant()
	if err != nil {
		return "", fmt.Errorf("no grant ID provided and no default grant set. Use 'nylas auth list' to see available grants")
	}

	return defaultGrant, nil
}

// createContext creates a context with timeout.
func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
