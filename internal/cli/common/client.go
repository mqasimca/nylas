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

// GetNylasClient creates a Nylas API client with credentials from environment variables or keyring.
// It checks credentials in this order:
// 1. Environment variables (NYLAS_API_KEY, NYLAS_CLIENT_ID, NYLAS_CLIENT_SECRET) - highest priority
// 2. System keyring (if available and env vars not set)
// 3. Encrypted file store (if keyring unavailable)
//
// This allows the CLI to work in multiple environments:
// - CI/CD pipelines (environment variables)
// - Docker containers (environment variables)
// - Integration tests (environment variables with NYLAS_DISABLE_KEYRING=true)
// - Local development (keyring)
func GetNylasClient() (ports.NylasClient, error) {
	// Load configuration
	configStore := config.NewDefaultFileStore()
	cfg, err := configStore.Load()
	if err != nil {
		cfg = &domain.Config{Region: "us"}
	}

	// First, check environment variables (highest priority)
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

// GetAPIKey returns the API key from environment variable or keyring.
// It checks in this order:
// 1. Environment variable (NYLAS_API_KEY) - highest priority
// 2. System keyring (if available)
// 3. Encrypted file store (if keyring unavailable)
func GetAPIKey() (string, error) {
	// First check environment variable (highest priority)
	apiKey := os.Getenv("NYLAS_API_KEY")

	// If not in env, try keyring/file store
	if apiKey == "" {
		secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
		if err == nil {
			apiKey, _ = secretStore.Get(ports.KeyAPIKey)
		}
	}

	if apiKey == "" {
		return "", fmt.Errorf("API key not configured. Set NYLAS_API_KEY environment variable or run 'nylas auth config'")
	}

	return apiKey, nil
}

// GetGrantID returns the grant ID from arguments, environment variable, or keyring.
// It checks in this order:
// 1. Command line argument (if provided)
// 2. Environment variable (NYLAS_GRANT_ID) - highest priority after arg
// 3. Stored default grant (from keyring/file)
func GetGrantID(args []string) (string, error) {
	// If provided as argument, use it
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	// Check environment variable first (highest priority after arg)
	grantID := os.Getenv("NYLAS_GRANT_ID")

	// If not in env, try keyring/file store
	if grantID == "" {
		secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
		if err == nil {
			grantStore := keyring.NewGrantStore(secretStore)
			grantID, _ = grantStore.GetDefaultGrant()
		}
	}

	if grantID == "" {
		return "", fmt.Errorf("no grant ID provided. Specify grant ID as argument or set NYLAS_GRANT_ID environment variable")
	}

	return grantID, nil
}
