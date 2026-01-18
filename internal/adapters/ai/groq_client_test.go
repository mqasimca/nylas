package ai

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewGroqClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.GroqConfig
		wantModel string
	}{
		{
			name:      "nil config uses defaults",
			config:    nil,
			wantModel: "mixtral-8x7b-32768",
		},
		{
			name: "custom config",
			config: &domain.GroqConfig{
				APIKey: "test-key",
				Model:  "llama2-70b-4096",
			},
			wantModel: "llama2-70b-4096",
		},
		{
			name: "env var config",
			config: &domain.GroqConfig{
				APIKey: "$GROQ_API_KEY",
				Model:  "gemma-7b-it",
			},
			wantModel: "gemma-7b-it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGroqClient(tt.config)

			if client.model != tt.wantModel {
				t.Errorf("model = %q, want %q", client.model, tt.wantModel)
			}

			if client.client == nil {
				t.Error("HTTP client is nil")
			}
		})
	}
}

func TestGroqClient_Name(t *testing.T) {
	client := NewGroqClient(nil)
	if name := client.Name(); name != "groq" {
		t.Errorf("Name() = %q, want %q", name, "groq")
	}
}

func TestGroqClient_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.GroqConfig
		wantAvail bool
	}{
		{
			name: "with API key",
			config: &domain.GroqConfig{
				APIKey: "test-key",
				Model:  "mixtral-8x7b-32768",
			},
			wantAvail: true,
		},
		{
			name: "without API key",
			config: &domain.GroqConfig{
				Model: "mixtral-8x7b-32768",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGroqClient(tt.config)
			ctx := context.Background()

			avail := client.IsAvailable(ctx)
			if avail != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", avail, tt.wantAvail)
			}
		})
	}
}

func TestGroqClient_GetModel(t *testing.T) {
	client := NewGroqClient(&domain.GroqConfig{
		APIKey: "test-key",
		Model:  "mixtral-8x7b-32768",
	})

	tests := []struct {
		name         string
		requestModel string
		want         string
	}{
		{
			name:         "use request model",
			requestModel: "llama2-70b-4096",
			want:         "llama2-70b-4096",
		},
		{
			name:         "use default model",
			requestModel: "",
			want:         "mixtral-8x7b-32768",
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

func TestGroqClient_ChatWithTools_NoAPIKey(t *testing.T) {
	client := NewGroqClient(&domain.GroqConfig{
		Model: "mixtral-8x7b-32768",
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
