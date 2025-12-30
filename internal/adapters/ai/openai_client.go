package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
)

// OpenAIClient implements LLMProvider for OpenAI.
type OpenAIClient struct {
	*BaseClient
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(config *domain.OpenAIConfig) *OpenAIClient {
	if config == nil {
		config = &domain.OpenAIConfig{
			Model: "gpt-4-turbo",
		}
	}

	apiKey := GetAPIKeyFromEnv(config.APIKey, "OPENAI_API_KEY")

	return &OpenAIClient{
		BaseClient: NewBaseClient(
			apiKey,
			config.Model,
			"https://api.openai.com/v1",
			0, // Use default timeout
		),
	}
}

// Name returns the provider name.
func (c *OpenAIClient) Name() string {
	return "openai"
}

// IsAvailable checks if OpenAI API key is configured.
func (c *OpenAIClient) IsAvailable(ctx context.Context) bool {
	return c.IsConfigured()
}

// Chat sends a chat completion request.
func (c *OpenAIClient) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	return c.ChatWithTools(ctx, req, nil)
}

// ChatWithTools sends a chat request with function calling.
func (c *OpenAIClient) ChatWithTools(ctx context.Context, req *domain.ChatRequest, tools []domain.Tool) (*domain.ChatResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("openai API key not configured")
	}

	// Prepare OpenAI request
	openaiReq := map[string]any{
		"model":    c.GetModel(req.Model),
		"messages": c.convertMessages(req.Messages),
	}

	if req.MaxTokens > 0 {
		openaiReq["max_tokens"] = req.MaxTokens
	}

	if req.Temperature > 0 {
		openaiReq["temperature"] = req.Temperature
	}

	// Tools support
	if len(tools) > 0 {
		openaiReq["tools"] = c.convertTools(tools)
		openaiReq["tool_choice"] = "auto"
	}

	// Send request using base client
	var openaiResp struct {
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

	if err := c.DoJSONRequestAndDecode(ctx, "POST", "/chat/completions", openaiReq, headers, &openaiResp); err != nil {
		return nil, err
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	response := &domain.ChatResponse{
		Content:  openaiResp.Choices[0].Message.Content,
		Model:    openaiResp.Model,
		Provider: "openai",
		Usage: domain.TokenUsage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}

	// Convert tool calls if present
	for _, tc := range openaiResp.Choices[0].Message.ToolCalls {
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
func (c *OpenAIClient) StreamChat(ctx context.Context, req *domain.ChatRequest, callback func(chunk string) error) error {
	// Simplified streaming implementation
	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}
	return callback(resp.Content)
}

// Helper methods

func (c *OpenAIClient) convertMessages(messages []domain.ChatMessage) []map[string]string {
	result := make([]map[string]string, len(messages))
	for i, msg := range messages {
		result[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return result
}

func (c *OpenAIClient) convertTools(tools []domain.Tool) []map[string]any {
	result := make([]map[string]any, len(tools))
	for i, tool := range tools {
		result[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	return result
}
