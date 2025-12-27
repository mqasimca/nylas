package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// OpenAIClient implements LLMProvider for OpenAI.
type OpenAIClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(config *domain.OpenAIConfig) *OpenAIClient {
	if config == nil {
		config = &domain.OpenAIConfig{
			Model: "gpt-4-turbo",
		}
	}

	apiKey := expandEnvVar(config.APIKey)
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &OpenAIClient{
		apiKey: apiKey,
		model:  config.Model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the provider name.
func (c *OpenAIClient) Name() string {
	return "openai"
}

// IsAvailable checks if OpenAI API key is configured.
func (c *OpenAIClient) IsAvailable(ctx context.Context) bool {
	return c.apiKey != ""
}

// Chat sends a chat completion request.
func (c *OpenAIClient) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	return c.ChatWithTools(ctx, req, nil)
}

// ChatWithTools sends a chat request with function calling.
func (c *OpenAIClient) ChatWithTools(ctx context.Context, req *domain.ChatRequest, tools []domain.Tool) (*domain.ChatResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("openai API key not configured")
	}

	// Prepare OpenAI request
	openaiReq := map[string]any{
		"model":    c.getModel(req.Model),
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

	// Send request
	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
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

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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

func (c *OpenAIClient) getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return c.model
}

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
