package ai

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.OpenAIConfig
		wantModel string
	}{
		{
			name:      "nil config uses defaults",
			config:    nil,
			wantModel: "gpt-4-turbo",
		},
		{
			name: "custom config",
			config: &domain.OpenAIConfig{
				APIKey: "test-key",
				Model:  "gpt-3.5-turbo",
			},
			wantModel: "gpt-3.5-turbo",
		},
		{
			name: "env var config",
			config: &domain.OpenAIConfig{
				APIKey: "$OPENAI_API_KEY",
				Model:  "gpt-4",
			},
			wantModel: "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.config)

			if client.model != tt.wantModel {
				t.Errorf("model = %q, want %q", client.model, tt.wantModel)
			}

			if client.client == nil {
				t.Error("HTTP client is nil")
			}
		})
	}
}

func TestOpenAIClient_Name(t *testing.T) {
	client := NewOpenAIClient(nil)
	if name := client.Name(); name != "openai" {
		t.Errorf("Name() = %q, want %q", name, "openai")
	}
}

func TestOpenAIClient_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.OpenAIConfig
		wantAvail bool
	}{
		{
			name: "with API key",
			config: &domain.OpenAIConfig{
				APIKey: "test-key",
				Model:  "gpt-4-turbo",
			},
			wantAvail: true,
		},
		{
			name: "without API key",
			config: &domain.OpenAIConfig{
				Model: "gpt-4-turbo",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.config)
			ctx := context.Background()

			avail := client.IsAvailable(ctx)
			if avail != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", avail, tt.wantAvail)
			}
		})
	}
}

func TestOpenAIClient_GetModel(t *testing.T) {
	client := NewOpenAIClient(&domain.OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-4-turbo",
	})

	tests := []struct {
		name         string
		requestModel string
		want         string
	}{
		{
			name:         "use request model",
			requestModel: "gpt-3.5-turbo",
			want:         "gpt-3.5-turbo",
		},
		{
			name:         "use default model",
			requestModel: "",
			want:         "gpt-4-turbo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.GetModel(tt.requestModel)
			if got != tt.want {
				t.Errorf("GetModel() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Note: ConvertMessages and ConvertTools tests moved to base_client_test.go
// since these are now shared functions in base_client.go

func TestOpenAIClient_ChatWithTools_NoAPIKey(t *testing.T) {
	client := NewOpenAIClient(&domain.OpenAIConfig{
		Model: "gpt-4-turbo",
		// No API key
	})

	ctx := context.Background()
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := client.ChatWithTools(ctx, req, nil)
	if err == nil {
		t.Error("expected error when API key not configured, got nil")
	}
}
