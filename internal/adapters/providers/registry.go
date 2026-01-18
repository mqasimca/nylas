// Package providers implements a provider registry pattern for extensible multi-provider support.
package providers

import (
	"fmt"
	"sync"

	"github.com/mqasimca/nylas/internal/ports"
)

// ProviderFactory is a function that creates a new provider client
type ProviderFactory func(config ProviderConfig) (ports.NylasClient, error)

// ProviderConfig contains configuration for initializing a provider
type ProviderConfig struct {
	APIKey       string
	ClientID     string
	ClientSecret string
	BaseURL      string
	Region       string
}

// Registry maintains a registry of provider factories
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

var defaultRegistry = &Registry{
	factories: make(map[string]ProviderFactory),
}

// Register registers a provider factory with the default registry
// Providers should call this in their init() function
func Register(name string, factory ProviderFactory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[name] = factory
}

// Get returns a provider factory by name
func Get(name string) (ProviderFactory, error) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	factory, ok := defaultRegistry.factories[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", name)
	}
	return factory, nil
}

// List returns all registered provider names
func List() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	names := make([]string, 0, len(defaultRegistry.factories))
	for name := range defaultRegistry.factories {
		names = append(names, name)
	}
	return names
}

// NewClient creates a new provider client by name
func NewClient(provider string, config ProviderConfig) (ports.NylasClient, error) {
	factory, err := Get(provider)
	if err != nil {
		return nil, err
	}
	return factory(config)
}
