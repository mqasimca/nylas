package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/domain"
)

// Executor handles plugin execution.
type Executor struct {
	config  *domain.Config
	version string
}

// NewExecutor creates a new plugin executor.
func NewExecutor(cfg *domain.Config, version string) *Executor {
	return &Executor{
		config:  cfg,
		version: version,
	}
}

// Execute runs a plugin with the given arguments.
// It sets up the environment with necessary configuration and credentials,
// then executes the plugin, passing through stdin/stdout/stderr.
func (e *Executor) Execute(plugin *Plugin, args []string) error {
	// Create command
	cmd := exec.Command(plugin.Path, args...)

	// Pass through stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up environment
	env, err := e.prepareEnvironment()
	if err != nil {
		return fmt.Errorf("failed to prepare environment: %w", err)
	}
	cmd.Env = env

	// Execute plugin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}

// prepareEnvironment prepares the environment variables for plugin execution.
// This includes passing configuration and credentials to the plugin.
func (e *Executor) prepareEnvironment() ([]string, error) {
	// Start with current environment
	env := os.Environ()

	// Add CLI version
	env = append(env, fmt.Sprintf("NYLAS_CLI_VERSION=%s", e.version))

	// Add region
	env = append(env, fmt.Sprintf("NYLAS_REGION=%s", e.config.Region))

	// Get API key from keyring
	apiKey, err := getAPIKey()
	if err == nil && apiKey != "" {
		env = append(env, fmt.Sprintf("NYLAS_API_KEY=%s", apiKey))
	}

	// Get client ID from keyring
	clientID, err := getClientID()
	if err == nil && clientID != "" {
		env = append(env, fmt.Sprintf("NYLAS_CLIENT_ID=%s", clientID))
	}

	// Add grant ID (default grant)
	if e.config.DefaultGrant != "" {
		env = append(env, fmt.Sprintf("NYLAS_GRANT_ID=%s", e.config.DefaultGrant))
	}

	// Add callback port
	env = append(env, fmt.Sprintf("NYLAS_CALLBACK_PORT=%d", e.config.CallbackPort))

	return env, nil
}

// getAPIKey retrieves the API key from the keyring.
func getAPIKey() (string, error) {
	ring := keyring.NewSystemKeyring()
	return ring.Get("api_key")
}

// getClientID retrieves the client ID from the keyring.
func getClientID() (string, error) {
	ring := keyring.NewSystemKeyring()
	return ring.Get("client_id")
}

// LoadConfig loads the CLI configuration.
func LoadConfig() (*domain.Config, error) {
	store := config.NewDefaultFileStore()
	return store.Load()
}
