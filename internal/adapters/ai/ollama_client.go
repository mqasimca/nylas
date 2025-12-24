package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// OllamaClient implements LLMProvider for Ollama.
type OllamaClient struct {
	host   string
	model  string
	client *http.Client
}

// NewOllamaClient creates a new Ollama client.
// Returns nil if config is nil - callers should validate config before calling.
func NewOllamaClient(config *domain.OllamaConfig) *OllamaClient {
	if config == nil {
		return nil
	}

	return &OllamaClient{
		host:  config.Host,
		model: config.Model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the provider name.
func (c *OllamaClient) Name() string {
	return "ollama"
}

// IsAvailable checks if Ollama is accessible.
func (c *OllamaClient) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Chat sends a chat completion request.
func (c *OllamaClient) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	return c.ChatWithTools(ctx, req, nil)
}

// ChatWithTools sends a chat request with function calling.
// Note: Ollama's tool support may vary by model.
func (c *OllamaClient) ChatWithTools(ctx context.Context, req *domain.ChatRequest, tools []domain.Tool) (*domain.ChatResponse, error) {
	// Prepare Ollama request
	ollamaReq := map[string]any{
		"model":    c.getModel(req.Model),
		"messages": c.convertMessages(req.Messages),
		"stream":   false,
	}

	if req.MaxTokens > 0 {
		ollamaReq["options"] = map[string]any{
			"num_predict": req.MaxTokens,
		}
	}

	if req.Temperature > 0 {
		if options, ok := ollamaReq["options"].(map[string]any); ok {
			options["temperature"] = req.Temperature
		} else {
			ollamaReq["options"] = map[string]any{
				"temperature": req.Temperature,
			}
		}
	}

	// Tools support (if model supports it)
	if len(tools) > 0 {
		ollamaReq["tools"] = c.convertTools(tools)
	}

	// Send request
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var ollamaResp struct {
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string         `json:"name"`
					Arguments map[string]any `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		Model string `json:"model"`
		Done  bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	response := &domain.ChatResponse{
		Content:  ollamaResp.Message.Content,
		Model:    ollamaResp.Model,
		Provider: "ollama",
		Usage: domain.TokenUsage{
			// Ollama doesn't always provide token counts
			TotalTokens: 0,
		},
	}

	// Convert tool calls if present
	for _, tc := range ollamaResp.Message.ToolCalls {
		response.ToolCalls = append(response.ToolCalls, domain.ToolCall{
			Function:  tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return response, nil
}

// StreamChat streams chat responses.
func (c *OllamaClient) StreamChat(ctx context.Context, req *domain.ChatRequest, callback func(chunk string) error) error {
	// Prepare Ollama request
	ollamaReq := map[string]any{
		"model":    c.getModel(req.Model),
		"messages": c.convertMessages(req.Messages),
		"stream":   true,
	}

	if req.Temperature > 0 {
		ollamaReq["options"] = map[string]any{
			"temperature": req.Temperature,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Stream response
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode chunk: %w", err)
		}

		if chunk.Message.Content != "" {
			if err := callback(chunk.Message.Content); err != nil {
				return err
			}
		}

		if chunk.Done {
			break
		}
	}

	return nil
}

// Helper methods

func (c *OllamaClient) getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return c.model
}

func (c *OllamaClient) convertMessages(messages []domain.ChatMessage) []map[string]string {
	result := make([]map[string]string, len(messages))
	for i, msg := range messages {
		result[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return result
}

func (c *OllamaClient) convertTools(tools []domain.Tool) []map[string]any {
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
