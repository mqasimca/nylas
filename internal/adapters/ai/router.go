package ai

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// Router implements LLMRouter for routing between multiple LLM providers.
type Router struct {
	providers       map[string]ports.LLMProvider
	defaultProvider string
	fallbackChain   []string
}

// NewRouter creates a new LLM router from configuration.
func NewRouter(config *domain.AIConfig) *Router {
	if config == nil {
		return &Router{
			providers:       make(map[string]ports.LLMProvider),
			defaultProvider: "",
			fallbackChain:   []string{},
		}
	}

	router := &Router{
		providers:       make(map[string]ports.LLMProvider),
		defaultProvider: config.DefaultProvider,
		fallbackChain:   []string{},
	}

	// Initialize providers based on config
	if config.Ollama != nil {
		router.providers["ollama"] = NewOllamaClient(config.Ollama)
	}

	if config.Claude != nil {
		router.providers["claude"] = NewClaudeClient(config.Claude)
	}

	if config.OpenAI != nil {
		router.providers["openai"] = NewOpenAIClient(config.OpenAI)
	}

	if config.Groq != nil {
		router.providers["groq"] = NewGroqClient(config.Groq)
	}

	// Set up fallback chain
	if config.Fallback != nil && config.Fallback.Enabled {
		router.fallbackChain = config.Fallback.Providers
	} else if router.defaultProvider != "" {
		// If no fallback configured, use only default provider
		router.fallbackChain = []string{router.defaultProvider}
	}

	return router
}

// GetProvider returns the specified provider or default if empty.
func (r *Router) GetProvider(name string) (ports.LLMProvider, error) {
	if name == "" {
		name = r.defaultProvider
	}

	if name == "" {
		return nil, fmt.Errorf("no provider specified and no default provider configured")
	}

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not configured", name)
	}

	return provider, nil
}

// Chat sends a chat request using the default provider with fallback.
func (r *Router) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	// If no fallback chain, use default provider
	if len(r.fallbackChain) == 0 {
		provider, err := r.GetProvider("")
		if err != nil {
			return nil, err
		}
		return provider.Chat(ctx, req)
	}

	// Try each provider in fallback chain
	var lastErr error
	for _, providerName := range r.fallbackChain {
		provider, exists := r.providers[providerName]
		if !exists {
			lastErr = fmt.Errorf("provider %q not configured", providerName)
			continue
		}

		// Check if provider is available
		if !provider.IsAvailable(ctx) {
			lastErr = fmt.Errorf("provider %q not available", providerName)
			continue
		}

		// Try to send request
		resp, err := provider.Chat(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("provider %q failed: %w", providerName, err)
			continue
		}

		// Success!
		return resp, nil
	}

	// All providers failed
	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}
	return nil, fmt.Errorf("no providers available")
}

// ChatWithProvider sends a chat request using a specific provider.
func (r *Router) ChatWithProvider(ctx context.Context, providerName string, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	provider, err := r.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	if !provider.IsAvailable(ctx) {
		return nil, fmt.Errorf("provider %q not available", providerName)
	}

	return provider.Chat(ctx, req)
}

// ListProviders returns available provider names.
func (r *Router) ListProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
