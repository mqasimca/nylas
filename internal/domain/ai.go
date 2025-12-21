package domain

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
