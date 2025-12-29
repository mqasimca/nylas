package ai

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewOllamaClient(t *testing.T) {
	t.Run("nil config returns nil", func(t *testing.T) {
		client := NewOllamaClient(nil)
		if client != nil {
			t.Error("expected nil client for nil config")
		}
	})

	t.Run("custom config", func(t *testing.T) {
		config := &domain.OllamaConfig{
			Host:  "http://custom:8080",
			Model: "llama2",
		}
		client := NewOllamaClient(config)

		if client == nil {
			t.Fatal("expected non-nil client")
		}

		if client.baseURL != "http://custom:8080" {
			t.Errorf("baseURL = %q, want %q", client.baseURL, "http://custom:8080")
		}

		if client.model != "llama2" {
			t.Errorf("model = %q, want %q", client.model, "llama2")
		}

		if client.client == nil {
			t.Error("HTTP client is nil")
		}
	})
}

func TestOllamaClient_Name(t *testing.T) {
	client := NewOllamaClient(&domain.OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "mistral:latest",
	})
	if name := client.Name(); name != "ollama" {
		t.Errorf("Name() = %q, want %q", name, "ollama")
	}
}

func TestOllamaClient_GetModel(t *testing.T) {
	client := NewOllamaClient(&domain.OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "mistral:latest",
	})

	tests := []struct {
		name         string
		requestModel string
		want         string
	}{
		{
			name:         "use request model",
			requestModel: "llama2",
			want:         "llama2",
		},
		{
			name:         "use default model",
			requestModel: "",
			want:         "mistral:latest",
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

func TestOllamaClient_ConvertMessages(t *testing.T) {
	client := NewOllamaClient(&domain.OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "mistral:latest",
	})

	messages := []domain.ChatMessage{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	converted := client.convertMessages(messages)

	if len(converted) != len(messages) {
		t.Errorf("converted messages count = %d, want %d", len(converted), len(messages))
	}

	for i, msg := range converted {
		if msg["role"] != messages[i].Role {
			t.Errorf("message[%d] role = %q, want %q", i, msg["role"], messages[i].Role)
		}
		if msg["content"] != messages[i].Content {
			t.Errorf("message[%d] content = %q, want %q", i, msg["content"], messages[i].Content)
		}
	}
}

func TestOllamaClient_ConvertTools(t *testing.T) {
	client := NewOllamaClient(&domain.OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "mistral:latest",
	})

	tools := []domain.Tool{
		{
			Name:        "get_weather",
			Description: "Get current weather",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City name",
					},
				},
			},
		},
	}

	converted := client.convertTools(tools)

	if len(converted) != len(tools) {
		t.Errorf("converted tools count = %d, want %d", len(converted), len(tools))
	}

	if converted[0]["type"] != "function" {
		t.Errorf("tool type = %v, want %q", converted[0]["type"], "function")
	}

	fn, ok := converted[0]["function"].(map[string]any)
	if !ok {
		t.Fatal("function field is not a map")
	}

	if fn["name"] != tools[0].Name {
		t.Errorf("function name = %v, want %q", fn["name"], tools[0].Name)
	}

	if fn["description"] != tools[0].Description {
		t.Errorf("function description = %v, want %q", fn["description"], tools[0].Description)
	}
}

func TestOllamaClient_IsAvailable(t *testing.T) {
	client := NewOllamaClient(&domain.OllamaConfig{
		Host:  "http://localhost:11434",
		Model: "mistral:latest",
	})

	ctx := context.Background()

	// This will attempt to connect to localhost:11434
	// In unit tests, this will likely fail unless Ollama is running
	// We're just testing that the method doesn't panic
	_ = client.IsAvailable(ctx)
}
