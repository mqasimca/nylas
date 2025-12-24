package domain

import "fmt"

// ChatMessage represents a chat message for AI/LLM interactions.
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant, tool
	Content string `json:"content"` // Message content
	Name    string `json:"name,omitempty"`
}

// ChatRequest represents a request to an LLM provider.
type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse represents a response from an LLM provider.
type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     TokenUsage `json:"usage"`
	Model     string     `json:"model,omitempty"`
	Provider  string     `json:"provider,omitempty"`
}

// Tool represents a function/tool available to the LLM.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall represents a tool invocation from the LLM.
type ToolCall struct {
	ID        string         `json:"id"`
	Function  string         `json:"function"`
	Arguments map[string]any `json:"arguments"`
}

// TokenUsage represents token usage statistics.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIConfig represents AI/LLM configuration.
type AIConfig struct {
	DefaultProvider string            `yaml:"default_provider"` // ollama, claude, openai, groq
	Fallback        *AIFallbackConfig `yaml:"fallback,omitempty"`
	Privacy         *PrivacyConfig    `yaml:"privacy,omitempty"`
	Features        *FeaturesConfig   `yaml:"features,omitempty"`
	Ollama          *OllamaConfig     `yaml:"ollama,omitempty"`
	Claude          *ClaudeConfig     `yaml:"claude,omitempty"`
	OpenAI          *OpenAIConfig     `yaml:"openai,omitempty"`
	Groq            *GroqConfig       `yaml:"groq,omitempty"`
	OpenRouter      *OpenRouterConfig `yaml:"openrouter,omitempty"`
}

// AIFallbackConfig represents fallback configuration.
type AIFallbackConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Providers []string `yaml:"providers"` // Try in order
}

// IsConfigured returns true if the AI config has at least one provider configured.
func (c *AIConfig) IsConfigured() bool {
	if c == nil {
		return false
	}
	return c.Ollama != nil || c.Claude != nil || c.OpenAI != nil ||
		c.Groq != nil || c.OpenRouter != nil
}

// ValidateForProvider validates that the required fields are set for the given provider.
func (c *AIConfig) ValidateForProvider(provider string) error {
	if c == nil {
		return fmt.Errorf("AI configuration is nil")
	}

	switch provider {
	case "ollama":
		if c.Ollama == nil {
			return fmt.Errorf("ollama configuration not found in config.yaml")
		}
		if c.Ollama.Host == "" {
			return fmt.Errorf("ollama.host is required")
		}
		if c.Ollama.Model == "" {
			return fmt.Errorf("ollama.model is required")
		}
	case "claude":
		if c.Claude == nil {
			return fmt.Errorf("claude configuration not found in config.yaml")
		}
		if c.Claude.Model == "" {
			return fmt.Errorf("claude.model is required")
		}
	case "openai":
		if c.OpenAI == nil {
			return fmt.Errorf("openai configuration not found in config.yaml")
		}
		if c.OpenAI.Model == "" {
			return fmt.Errorf("openai.model is required")
		}
	case "groq":
		if c.Groq == nil {
			return fmt.Errorf("groq configuration not found in config.yaml")
		}
		if c.Groq.Model == "" {
			return fmt.Errorf("groq.model is required")
		}
	case "openrouter":
		if c.OpenRouter == nil {
			return fmt.Errorf("openrouter configuration not found in config.yaml")
		}
		if c.OpenRouter.Model == "" {
			return fmt.Errorf("openrouter.model is required")
		}
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	return nil
}

// DefaultAIConfig returns a default AI configuration for first-time setup.
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		DefaultProvider: "ollama",
		Ollama: &OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "mistral:latest",
		},
	}
}

// OllamaConfig represents Ollama-specific configuration.
type OllamaConfig struct {
	Host  string `yaml:"host"`  // e.g., http://localhost:11434
	Model string `yaml:"model"` // e.g., mistral:latest
}

// ClaudeConfig represents Claude/Anthropic-specific configuration.
type ClaudeConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // Can use ${ENV_VAR}
	Model  string `yaml:"model"`             // e.g., claude-3-5-sonnet-20241022
}

// OpenAIConfig represents OpenAI-specific configuration.
type OpenAIConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // Can use ${ENV_VAR}
	Model  string `yaml:"model"`             // e.g., gpt-4-turbo
}

// GroqConfig represents Groq-specific configuration.
type GroqConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // Can use ${ENV_VAR}
	Model  string `yaml:"model"`             // e.g., mixtral-8x7b-32768
}

// OpenRouterConfig represents OpenRouter-specific configuration.
type OpenRouterConfig struct {
	APIKey string `yaml:"api_key,omitempty"` // Can use ${ENV_VAR}
	Model  string `yaml:"model"`             // e.g., anthropic/claude-3.5-sonnet
}

// PrivacyConfig represents privacy settings for AI features.
type PrivacyConfig struct {
	AllowCloudAI     bool `yaml:"allow_cloud_ai"`     // Require explicit opt-in for cloud AI
	DataRetention    int  `yaml:"data_retention"`     // Days to keep learned patterns (0 = disabled)
	LocalStorageOnly bool `yaml:"local_storage_only"` // Only use local storage, no cloud
}

// FeaturesConfig represents feature toggles for AI capabilities.
type FeaturesConfig struct {
	NaturalLanguageScheduling bool `yaml:"natural_language_scheduling"` // Enable natural language scheduling
	PredictiveScheduling      bool `yaml:"predictive_scheduling"`       // Enable predictive scheduling
	FocusTimeProtection       bool `yaml:"focus_time_protection"`       // Enable focus time protection
	ConflictResolution        bool `yaml:"conflict_resolution"`         // Enable conflict resolution
	EmailContextAnalysis      bool `yaml:"email_context_analysis"`      // Enable email context analysis
}

// EmailThreadAnalysis represents the AI analysis of an email thread.
type EmailThreadAnalysis struct {
	ThreadID          string                 `json:"thread_id"`
	Subject           string                 `json:"subject"`
	MessageCount      int                    `json:"message_count"`
	ParticipantCount  int                    `json:"participant_count"`
	Purpose           string                 `json:"purpose"`            // Primary meeting purpose
	Topics            []string               `json:"topics"`             // Key topics discussed
	Priority          MeetingPriority        `json:"priority"`           // Detected priority level
	SuggestedDuration int                    `json:"suggested_duration"` // In minutes
	Participants      []ParticipantInfo      `json:"participants"`
	Agenda            *MeetingAgenda         `json:"agenda,omitempty"`
	BestMeetingTime   *MeetingTimeSuggestion `json:"best_meeting_time,omitempty"`
	UrgencyIndicators []string               `json:"urgency_indicators,omitempty"`
}

// ParticipantInfo represents a participant with their involvement level.
type ParticipantInfo struct {
	Email         string           `json:"email"`
	Name          string           `json:"name,omitempty"`
	Required      bool             `json:"required"`
	Involvement   InvolvementLevel `json:"involvement"`
	MentionCount  int              `json:"mention_count"`
	MessageCount  int              `json:"message_count"`
	LastMessageAt string           `json:"last_message_at,omitempty"`
}

// InvolvementLevel represents how involved a participant is in the thread.
type InvolvementLevel string

const (
	InvolvementHigh   InvolvementLevel = "high"   // Active contributor, decision maker
	InvolvementMedium InvolvementLevel = "medium" // Regular participant
	InvolvementLow    InvolvementLevel = "low"    // Minimal involvement, FYI
)

// MeetingAgenda represents an auto-generated meeting agenda.
type MeetingAgenda struct {
	Title    string       `json:"title"`
	Duration int          `json:"duration"` // In minutes
	Items    []AgendaItem `json:"items"`
	Notes    []string     `json:"notes,omitempty"` // Additional context
}

// AgendaItem represents a single agenda item.
type AgendaItem struct {
	Title       string `json:"title"`
	Duration    int    `json:"duration"` // In minutes
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"` // Quote from email thread
	Owner       string `json:"owner,omitempty"`  // Who should lead this item
	Decision    bool   `json:"decision"`         // Does this require a decision?
}

// MeetingTimeSuggestion represents a suggested meeting time based on thread analysis.
type MeetingTimeSuggestion struct {
	Time      string `json:"time"`      // ISO 8601 format
	Timezone  string `json:"timezone"`  // IANA timezone ID
	Score     int    `json:"score"`     // 0-100
	Reasoning string `json:"reasoning"` // Why this time was chosen
}

// EmailAnalysisRequest represents a request to analyze an email thread.
type EmailAnalysisRequest struct {
	ThreadID      string `json:"thread_id"`
	IncludeAgenda bool   `json:"include_agenda"`
	IncludeTime   bool   `json:"include_time"`
}
