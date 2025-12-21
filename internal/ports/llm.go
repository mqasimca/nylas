package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// LLMProvider defines the interface for LLM providers.
type LLMProvider interface {
	// Chat sends a chat completion request
	Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)

	// ChatWithTools sends a chat request with function calling
	ChatWithTools(ctx context.Context, req *domain.ChatRequest, tools []domain.Tool) (*domain.ChatResponse, error)

	// StreamChat streams chat responses
	StreamChat(ctx context.Context, req *domain.ChatRequest, callback func(chunk string) error) error

	// Name returns the provider name
	Name() string

	// IsAvailable checks if provider is configured and accessible
	IsAvailable(ctx context.Context) bool
}

// LLMRouter defines the interface for routing between multiple LLM providers.
type LLMRouter interface {
	// GetProvider returns the specified provider or default if empty
	GetProvider(name string) (LLMProvider, error)

	// Chat sends a chat request using the default provider with fallback
	Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)

	// ChatWithProvider sends a chat request using a specific provider
	ChatWithProvider(ctx context.Context, provider string, req *domain.ChatRequest) (*domain.ChatResponse, error)

	// ListProviders returns available provider names
	ListProviders() []string
}
