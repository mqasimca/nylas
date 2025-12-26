package otp

import (
	"os"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	nylasadapter "github.com/mqasimca/nylas/internal/adapters/nylas"
	otpapp "github.com/mqasimca/nylas/internal/app/otp"
	"github.com/mqasimca/nylas/internal/ports"
)

// createOTPService creates the OTP service.
func createOTPService() (*otpapp.Service, error) {
	configStore := config.NewDefaultFileStore()

	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return nil, err
	}

	grantStore := keyring.NewGrantStore(secretStore)

	// Create Nylas client
	client := nylasadapter.NewHTTPClient()

	// Set credentials - check env vars first
	cfg, _ := configStore.Load()
	client.SetRegion(cfg.Region)

	apiKey := os.Getenv("NYLAS_API_KEY")
	clientID := os.Getenv("NYLAS_CLIENT_ID")
	clientSecret := os.Getenv("NYLAS_CLIENT_SECRET")

	// If API key not in env, try keyring/file store
	if apiKey == "" {
		apiKey, _ = secretStore.Get(ports.KeyAPIKey)
		if clientID == "" {
			clientID, _ = secretStore.Get(ports.KeyClientID)
		}
		if clientSecret == "" {
			clientSecret, _ = secretStore.Get(ports.KeyClientSecret)
		}
	}

	client.SetCredentials(clientID, clientSecret, apiKey)

	return otpapp.NewService(client, grantStore, configStore), nil
}
