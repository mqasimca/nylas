package slack

import (
	"context"
	"os"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	slackadapter "github.com/mqasimca/nylas/internal/adapters/slack"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

const (
	slackTokenKey  = "slack_user_token"
	slackAuthKey   = "slack_auth_info"
	defaultTimeout = 30 * time.Second
)

// storeSlackToken stores the Slack token in the keyring.
func storeSlackToken(token string) error {
	store, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return err
	}
	return store.Set(slackTokenKey, token)
}

// getSlackToken retrieves the Slack token from environment or keyring.
func getSlackToken() (string, error) {
	if token := os.Getenv("SLACK_USER_TOKEN"); token != "" {
		return token, nil
	}

	store, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return "", err
	}

	token, err := store.Get(slackTokenKey)
	if err != nil {
		return "", domain.ErrSlackNotConfigured
	}

	return token, nil
}

// removeSlackToken removes the Slack token from the keyring.
func removeSlackToken() error {
	store, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return err
	}
	return store.Delete(slackTokenKey)
}

// getSlackClientFromKeyring creates a client using stored credentials.
func getSlackClientFromKeyring() (ports.SlackClient, error) {
	token, err := getSlackToken()
	if err != nil {
		return nil, err
	}
	return getSlackClient(token)
}

// getSlackClient creates a new Slack client with the given token.
func getSlackClient(token string) (ports.SlackClient, error) {
	config := slackadapter.DefaultConfig()
	config.UserToken = token
	config.Debug = os.Getenv("SLACK_DEBUG") == "true"
	return slackadapter.NewClient(config)
}

// createContext creates a context with default timeout.
func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// resolveChannelName resolves a channel name to its ID.
func resolveChannelName(ctx context.Context, client ports.SlackClient, name string) (string, error) {
	slackClient, ok := client.(*slackadapter.Client)
	if !ok {
		resp, err := client.ListChannels(ctx, &domain.SlackChannelQueryParams{
			Types:           []string{"public_channel", "private_channel"},
			ExcludeArchived: true,
			Limit:           200,
		})
		if err != nil {
			return "", err
		}

		for _, ch := range resp.Channels {
			if ch.Name == name {
				return ch.ID, nil
			}
		}
		return "", domain.ErrSlackChannelNotFound
	}

	ch, err := slackClient.ResolveChannelByName(ctx, name)
	if err != nil {
		return "", err
	}
	return ch.ID, nil
}
