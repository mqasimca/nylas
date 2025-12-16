package otp

import (
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

	// Set credentials
	cfg, _ := configStore.Load()
	client.SetRegion(cfg.Region)

	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)
	apiKey, _ := secretStore.Get(ports.KeyAPIKey)
	client.SetCredentials(clientID, clientSecret, apiKey)

	return otpapp.NewService(client, grantStore, configStore), nil
}
