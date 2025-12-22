package ai

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewRouter(t *testing.T) {
	tests := []struct {
		name          string
		config        *domain.AIConfig
		wantDefault   string
		wantProviders int
		wantFallback  int
	}{
		{
			name:          "nil config",
			config:        nil,
			wantDefault:   "",
			wantProviders: 0,
			wantFallback:  0,
		},
		{
			name: "single provider",
			config: &domain.AIConfig{
				DefaultProvider: "ollama",
				Ollama: &domain.OllamaConfig{
					Host:  "http://localhost:11434",
					Model: "mistral:latest",
				},
			},
			wantDefault:   "ollama",
			wantProviders: 1,
			wantFallback:  1, // Default only
		},
		{
			name: "multiple providers with fallback",
			config: &domain.AIConfig{
				DefaultProvider: "claude",
				Ollama: &domain.OllamaConfig{
					Host:  "http://localhost:11434",
					Model: "mistral:latest",
				},
				Claude: &domain.ClaudeConfig{
					APIKey: "test-key",
					Model:  "claude-3-5-sonnet-20241022",
				},
				OpenAI: &domain.OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-4-turbo",
				},
				Fallback: &domain.AIFallbackConfig{
					Enabled:   true,
					Providers: []string{"claude", "openai", "ollama"},
				},
			},
			wantDefault:   "claude",
			wantProviders: 3,
			wantFallback:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(tt.config)

			if router.defaultProvider != tt.wantDefault {
				t.Errorf("defaultProvider = %q, want %q", router.defaultProvider, tt.wantDefault)
			}

			if len(router.providers) != tt.wantProviders {
				t.Errorf("providers count = %d, want %d", len(router.providers), tt.wantProviders)
			}

			if len(router.fallbackChain) != tt.wantFallback {
				t.Errorf("fallbackChain count = %d, want %d", len(router.fallbackChain), tt.wantFallback)
			}
		})
	}
}

func TestRouter_GetProvider(t *testing.T) {
	config := &domain.AIConfig{
		DefaultProvider: "ollama",
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
		Claude: &domain.ClaudeConfig{
			APIKey: "test-key",
			Model:  "claude-3-5-sonnet-20241022",
		},
	}

	router := NewRouter(config)

	tests := []struct {
		name         string
		providerName string
		wantErr      bool
		wantName     string
	}{
		{
			name:         "get default provider",
			providerName: "",
			wantErr:      false,
			wantName:     "ollama",
		},
		{
			name:         "get specific provider",
			providerName: "claude",
			wantErr:      false,
			wantName:     "claude",
		},
		{
			name:         "get non-existent provider",
			providerName: "unknown",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := router.GetProvider(tt.providerName)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider.Name() != tt.wantName {
				t.Errorf("provider name = %q, want %q", provider.Name(), tt.wantName)
			}
		})
	}
}

func TestRouter_GetProvider_NoDefault(t *testing.T) {
	// Router with no default provider
	router := NewRouter(&domain.AIConfig{
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
	})

	// Should error when requesting default (empty string)
	_, err := router.GetProvider("")
	if err == nil {
		t.Error("expected error when no default provider configured")
	}
}

func TestRouter_ListProviders(t *testing.T) {
	config := &domain.AIConfig{
		DefaultProvider: "ollama",
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
		Claude: &domain.ClaudeConfig{
			APIKey: "test-key",
			Model:  "claude-3-5-sonnet-20241022",
		},
		OpenAI: &domain.OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4-turbo",
		},
	}

	router := NewRouter(config)
	providers := router.ListProviders()

	if len(providers) != 3 {
		t.Errorf("providers count = %d, want 3", len(providers))
	}

	// Check that expected providers are in the list
	expected := map[string]bool{
		"ollama": false,
		"claude": false,
		"openai": false,
	}

	for _, name := range providers {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("provider %q not in list", name)
		}
	}
}

func TestRouter_ChatWithProvider(t *testing.T) {
	config := &domain.AIConfig{
		DefaultProvider: "ollama",
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
	}

	router := NewRouter(config)
	ctx := context.Background()

	// Test with non-existent provider
	_, err := router.ChatWithProvider(ctx, "unknown", &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	})

	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestRouter_Chat(t *testing.T) {
	// Test with fallback disabled - router should fail fast if provider unavailable
	config := &domain.AIConfig{
		DefaultProvider: "ollama",
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
		Fallback: &domain.AIFallbackConfig{
			Enabled:   false,
			Providers: []string{},
		},
	}

	router := NewRouter(config)
	ctx := context.Background()

	// Test Chat without provider (should use default)
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// This will attempt to connect to localhost:11434, which will fail in test
	// We're testing that the method routes to the correct provider
	_, err := router.Chat(ctx, req)
	// Error is expected since we don't have a real Ollama server running
	// and fallback is disabled
	if err == nil {
		t.Error("expected error when Ollama server not available, got nil")
	}
}

func TestRouter_Chat_NoDefaultProvider(t *testing.T) {
	// Router with no default provider
	router := NewRouter(nil)
	ctx := context.Background()

	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := router.Chat(ctx, req)
	if err == nil {
		t.Error("expected error when no default provider configured")
	}
}
