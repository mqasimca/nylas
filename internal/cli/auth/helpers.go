package auth

import (
	"github.com/mqasimca/nylas/internal/adapters/browser"
	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	nylasadapter "github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/adapters/oauth"
	authapp "github.com/mqasimca/nylas/internal/app/auth"
	"github.com/mqasimca/nylas/internal/ports"
)

// createDependencies creates all the common dependencies.
func createDependencies() (ports.ConfigStore, ports.SecretStore, ports.GrantStore, error) {
	configStore := config.NewDefaultFileStore()

	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return nil, nil, nil, err
	}

	grantStore := keyring.NewGrantStore(secretStore)

	return configStore, secretStore, grantStore, nil
}

// createConfigService creates the config service.
func createConfigService() (*authapp.ConfigService, ports.ConfigStore, ports.SecretStore, error) {
	configStore, secretStore, _, err := createDependencies()
	if err != nil {
		return nil, nil, nil, err
	}
	return authapp.NewConfigService(configStore, secretStore), configStore, secretStore, nil
}

// createGrantService creates the grant service.
func createGrantService() (*authapp.GrantService, *authapp.ConfigService, error) {
	configStore, secretStore, grantStore, err := createDependencies()
	if err != nil {
		return nil, nil, err
	}

	configSvc := authapp.NewConfigService(configStore, secretStore)

	// Create Nylas client
	client := nylasadapter.NewHTTPClient()

	// Set credentials if available
	cfg, _ := configStore.Load()
	client.SetRegion(cfg.Region)

	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)
	apiKey, _ := secretStore.Get(ports.KeyAPIKey)
	client.SetCredentials(clientID, clientSecret, apiKey)

	return authapp.NewGrantService(client, grantStore, configStore), configSvc, nil
}

// createAuthService creates the auth service.
func createAuthService() (*authapp.Service, *authapp.ConfigService, error) {
	configStore, secretStore, grantStore, err := createDependencies()
	if err != nil {
		return nil, nil, err
	}

	configSvc := authapp.NewConfigService(configStore, secretStore)

	// Create Nylas client
	client := nylasadapter.NewHTTPClient()

	// Set credentials if available
	cfg, _ := configStore.Load()
	client.SetRegion(cfg.Region)

	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)
	apiKey, _ := secretStore.Get(ports.KeyAPIKey)
	client.SetCredentials(clientID, clientSecret, apiKey)

	// Create OAuth server
	oauthServer := oauth.NewCallbackServer(cfg.CallbackPort)

	// Create browser
	browserAdapter := browser.NewDefaultBrowser()

	return authapp.NewService(client, grantStore, configStore, oauthServer, browserAdapter), configSvc, nil
}
