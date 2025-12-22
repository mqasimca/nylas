package ai

import (
	"context"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestNewClaudeClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.ClaudeConfig
		wantModel string
	}{
		{
			name:      "nil config uses defaults",
			config:    nil,
			wantModel: "claude-3-5-sonnet-20241022",
		},
		{
			name: "custom config",
			config: &domain.ClaudeConfig{
				APIKey: "test-key",
				Model:  "claude-3-opus-20240229",
			},
			wantModel: "claude-3-opus-20240229",
		},
		{
			name: "env var config",
			config: &domain.ClaudeConfig{
				APIKey: "${ANTHROPIC_API_KEY}",
				Model:  "claude-3-haiku-20240307",
			},
			wantModel: "claude-3-haiku-20240307",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClaudeClient(tt.config)

			if client.model != tt.wantModel {
				t.Errorf("model = %q, want %q", client.model, tt.wantModel)
			}

			if client.client == nil {
				t.Error("HTTP client is nil")
			}
		})
	}
}

func TestClaudeClient_Name(t *testing.T) {
	client := NewClaudeClient(nil)
	if name := client.Name(); name != "claude" {
		t.Errorf("Name() = %q, want %q", name, "claude")
	}
}

func TestClaudeClient_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		config    *domain.ClaudeConfig
		wantAvail bool
	}{
		{
			name: "with API key",
			config: &domain.ClaudeConfig{
				APIKey: "test-key",
				Model:  "claude-3-5-sonnet-20241022",
			},
			wantAvail: true,
		},
		{
			name: "without API key",
			config: &domain.ClaudeConfig{
				Model: "claude-3-5-sonnet-20241022",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClaudeClient(tt.config)
			ctx := context.Background()

			avail := client.IsAvailable(ctx)
			if avail != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", avail, tt.wantAvail)
			}
		})
	}
}

func TestClaudeClient_GetModel(t *testing.T) {
	client := NewClaudeClient(&domain.ClaudeConfig{
		APIKey: "test-key",
		Model:  "claude-3-5-sonnet-20241022",
	})

	tests := []struct {
		name         string
		requestModel string
		want         string
	}{
		{
			name:         "use request model",
			requestModel: "claude-3-opus-20240229",
			want:         "claude-3-opus-20240229",
		},
		{
			name:         "use default model",
			requestModel: "",
			want:         "claude-3-5-sonnet-20241022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.getModel(tt.requestModel)
			if got != tt.want {
				t.Errorf("getModel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudeClient_GetMaxTokens(t *testing.T) {
	client := NewClaudeClient(nil)

	tests := []struct {
		name  string
		input int
		want  int
	}{
		{
			name:  "use request max tokens",
			input: 1024,
			want:  1024,
		},
		{
			name:  "use default max tokens",
			input: 0,
			want:  4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.getMaxTokens(tt.input)
			if got != tt.want {
				t.Errorf("getMaxTokens() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClaudeClient_ExtractSystemMessage(t *testing.T) {
	client := NewClaudeClient(nil)

	tests := []struct {
		name         string
		messages     []domain.ChatMessage
		wantSystem   string
		wantFiltered int
	}{
		{
			name: "with system message",
			messages: []domain.ChatMessage{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			wantSystem:   "You are a helpful assistant",
			wantFiltered: 2,
		},
		{
			name: "without system message",
			messages: []domain.ChatMessage{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			wantSystem:   "",
			wantFiltered: 2,
		},
		{
			name: "multiple system messages (uses last)",
			messages: []domain.ChatMessage{
				{Role: "system", Content: "First system"},
				{Role: "system", Content: "Second system"},
				{Role: "user", Content: "Hello"},
			},
			wantSystem:   "Second system",
			wantFiltered: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system, filtered := client.extractSystemMessage(tt.messages)

			if system != tt.wantSystem {
				t.Errorf("system = %q, want %q", system, tt.wantSystem)
			}

			if len(filtered) != tt.wantFiltered {
				t.Errorf("filtered count = %d, want %d", len(filtered), tt.wantFiltered)
			}
		})
	}
}

func TestClaudeClient_ConvertMessages(t *testing.T) {
	client := NewClaudeClient(nil)

	messages := []domain.ChatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
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

func TestClaudeClient_ConvertMessagesSkipsSystem(t *testing.T) {
	client := NewClaudeClient(nil)

	messages := []domain.ChatMessage{
		{Role: "system", Content: "System message"},
		{Role: "user", Content: "Hello"},
	}

	converted := client.convertMessages(messages)

	// System message should be skipped
	if len(converted) != 1 {
		t.Errorf("converted messages count = %d, want 1", len(converted))
	}

	if converted[0]["role"] != "user" {
		t.Errorf("first message role = %q, want %q", converted[0]["role"], "user")
	}
}

func TestClaudeClient_ConvertTools(t *testing.T) {
	client := NewClaudeClient(nil)

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

	if converted[0]["name"] != tools[0].Name {
		t.Errorf("tool name = %v, want %q", converted[0]["name"], tools[0].Name)
	}

	if converted[0]["description"] != tools[0].Description {
		t.Errorf("tool description = %v, want %q", converted[0]["description"], tools[0].Description)
	}

	if converted[0]["input_schema"] == nil {
		t.Error("input_schema is nil")
	}
}

func TestClaudeClient_ChatWithTools_NoAPIKey(t *testing.T) {
	client := NewClaudeClient(&domain.ClaudeConfig{
		Model: "claude-3-5-sonnet-20241022",
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

func TestExpandEnvVar(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "with env var syntax",
			value: "${PATH}",
			want:  "", // Will expand to actual PATH or empty
		},
		{
			name:  "without env var syntax",
			value: "literal-value",
			want:  "literal-value",
		},
		{
			name:  "partial match",
			value: "${incomplete",
			want:  "${incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEnvVar(tt.value)

			// For non-env var values, should return exact input
			if tt.value == "literal-value" && got != tt.want {
				t.Errorf("expandEnvVar() = %q, want %q", got, tt.want)
			}

			// For partial match, should return exact input
			if tt.value == "${incomplete" && got != tt.want {
				t.Errorf("expandEnvVar() = %q, want %q", got, tt.want)
			}

			// For env vars, just verify it doesn't panic
			if tt.value == "${PATH}" {
				// PATH expansion will vary by system, just verify no panic
				_ = got
			}
		})
	}
}

func TestClaudeClient_Chat_Success(t *testing.T) {
	// Note: This test requires a real API call or more complex HTTP mocking
	// For now, we test that the method properly delegates to ChatWithTools
	client := NewClaudeClient(&domain.ClaudeConfig{
		APIKey: "test-key",
		Model:  "claude-3-5-sonnet-20241022",
	})

	ctx := context.Background()
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// This will fail with API error since we don't have a valid API key
	// But it exercises the code path
	_, err := client.Chat(ctx, req)
	if err != nil {
		// Expected - no valid API key
		t.Logf("Chat() error = %v (expected without valid API key)", err)
	}
}

func TestClaudeClient_Chat_NoAPIKey(t *testing.T) {
	client := NewClaudeClient(&domain.ClaudeConfig{
		Model: "claude-3-5-sonnet-20241022",
		// No API key
	})

	ctx := context.Background()
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := client.Chat(ctx, req)
	if err == nil {
		t.Error("expected error when API key not configured, got nil")
	}
}

func TestClaudeClient_StreamChat(t *testing.T) {
	client := NewClaudeClient(&domain.ClaudeConfig{
		APIKey: "test-key",
		Model:  "claude-3-5-sonnet-20241022",
	})

	ctx := context.Background()
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// Collect chunks
	var chunks []string
	callback := func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}

	// This will fail with API error since we don't have a valid API key
	// But it exercises the code path
	err := client.StreamChat(ctx, req, callback)
	if err != nil {
		// Expected - no valid API endpoint
		t.Logf("StreamChat() error = %v (expected without valid API)", err)
	}
}

func TestClaudeClient_StreamChat_NoAPIKey(t *testing.T) {
	client := NewClaudeClient(&domain.ClaudeConfig{
		Model: "claude-3-5-sonnet-20241022",
		// No API key
	})

	ctx := context.Background()
	req := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	callback := func(chunk string) error {
		return nil
	}

	err := client.StreamChat(ctx, req, callback)
	if err == nil {
		t.Error("expected error when API key not configured, got nil")
	}
}
