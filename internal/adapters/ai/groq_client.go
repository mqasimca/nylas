package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
)

// GroqClient implements LLMProvider for Groq.
type GroqClient struct {
	*BaseClient
}

// NewGroqClient creates a new Groq client.
func NewGroqClient(config *domain.GroqConfig) *GroqClient {
	if config == nil {
		config = &domain.GroqConfig{
			Model: "mixtral-8x7b-32768",
		}
	}

	apiKey := GetAPIKeyFromEnv(config.APIKey, "GROQ_API_KEY")

	return &GroqClient{
		BaseClient: NewBaseClient(
			apiKey,
			config.Model,
			"https://api.groq.com/openai/v1",
			0, // Use default timeout
		),
	}
}

// Name returns the provider name.
func (c *GroqClient) Name() string {
	return "groq"
}

// IsAvailable checks if Groq API key is configured.
func (c *GroqClient) IsAvailable(ctx context.Context) bool {
	return c.IsConfigured()
}

// Chat sends a chat completion request.
func (c *GroqClient) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	return c.ChatWithTools(ctx, req, nil)
}

// ChatWithTools sends a chat request with function calling.
func (c *GroqClient) ChatWithTools(ctx context.Context, req *domain.ChatRequest, tools []domain.Tool) (*domain.ChatResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("groq API key not configured")
	}

	// Prepare Groq request (OpenAI-compatible format)
	groqReq := map[string]any{
		"model":    c.GetModel(req.Model),
		"messages": ConvertMessagesToMaps(req.Messages),
	}

	if req.MaxTokens > 0 {
		groqReq["max_tokens"] = req.MaxTokens
	}

	if req.Temperature > 0 {
		groqReq["temperature"] = req.Temperature
	}

	// Tools support
	if len(tools) > 0 {
		groqReq["tools"] = ConvertToolsOpenAIFormat(tools)
		groqReq["tool_choice"] = "auto"
	}

	// Send request using base client
	var groqResp struct {
		Choices []struct {
			Message struct {
				Role      string `json:"role"`
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	if err := c.DoJSONRequestAndDecode(ctx, "POST", "/chat/completions", groqReq, headers, &groqResp); err != nil {
		return nil, err
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Groq")
	}

	response := &domain.ChatResponse{
		Content:  groqResp.Choices[0].Message.Content,
		Model:    groqResp.Model,
		Provider: "groq",
		Usage: domain.TokenUsage{
			PromptTokens:     groqResp.Usage.PromptTokens,
			CompletionTokens: groqResp.Usage.CompletionTokens,
			TotalTokens:      groqResp.Usage.TotalTokens,
		},
	}

	// Convert tool calls if present
	for _, tc := range groqResp.Choices[0].Message.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
			response.ToolCalls = append(response.ToolCalls, domain.ToolCall{
				ID:        tc.ID,
				Function:  tc.Function.Name,
				Arguments: args,
			})
		}
	}

	return response, nil
}

// StreamChat streams chat responses.
func (c *GroqClient) StreamChat(ctx context.Context, req *domain.ChatRequest, callback func(chunk string) error) error {
	return FallbackStreamChat(ctx, req, c.Chat, callback)
}
