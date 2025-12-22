package common

import (
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// GetNylasClient creates a Nylas API client with credentials from keyring or environment variables.
// It checks credentials in this order:
// 1. System keyring (if available)
// 2. Encrypted file store (if keyring unavailable)
// 3. Environment variables (NYLAS_API_KEY, NYLAS_CLIENT_ID, NYLAS_CLIENT_SECRET)
//
// This allows the CLI to work in multiple environments:
// - Local development (keyring)
// - CI/CD pipelines (environment variables)
// - Docker containers (environment variables)
// - Integration tests (environment variables with NYLAS_DISABLE_KEYRING=true)
func GetNylasClient() (ports.NylasClient, error) {
	// Load configuration
	configStore := config.NewDefaultFileStore()
	cfg, err := configStore.Load()
	if err != nil {
		cfg = &domain.Config{Region: "us"}
	}

	var apiKey, clientID, clientSecret string

	// Try to get credentials from keyring/file store
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err == nil {
		// Successfully created secret store, try to get credentials
		apiKey, _ = secretStore.Get(ports.KeyAPIKey)
		clientID, _ = secretStore.Get(ports.KeyClientID)
		clientSecret, _ = secretStore.Get(ports.KeyClientSecret)
	}

	// Fall back to environment variables if not found in keyring
	if apiKey == "" {
		apiKey = os.Getenv("NYLAS_API_KEY")
	}
	if clientID == "" {
		clientID = os.Getenv("NYLAS_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("NYLAS_CLIENT_SECRET")
	}

	// Validate that we have at least the API key
	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured. Set NYLAS_API_KEY environment variable or run 'nylas auth config'")
	}

	// Create and configure the HTTP client
	c := nylas.NewHTTPClient()
	c.SetRegion(cfg.Region)
	c.SetCredentials(clientID, clientSecret, apiKey)

	return c, nil
}

// GetGrantID returns the grant ID from arguments, keyring, or environment variable.
// It checks in this order:
// 1. Command line argument (if provided)
// 2. Stored default grant (from keyring/file)
// 3. Environment variable (NYLAS_GRANT_ID)
func GetGrantID(args []string) (string, error) {
	// If provided as argument, use it
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	var grantID string

	// Try to get from keyring/file store
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err == nil {
		grantStore := keyring.NewGrantStore(secretStore)
		grantID, _ = grantStore.GetDefaultGrant()
	}

	// Fall back to environment variable
	if grantID == "" {
		grantID = os.Getenv("NYLAS_GRANT_ID")
	}

	if grantID == "" {
		return "", fmt.Errorf("no grant ID provided. Specify grant ID as argument or set NYLAS_GRANT_ID environment variable")
	}

	return grantID, nil
}
