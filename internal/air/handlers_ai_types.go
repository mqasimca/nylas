package air

// AIRequest represents an AI summarization request.
type AIRequest struct {
	EmailID string `json:"email_id"`
	Prompt  string `json:"prompt"`
}

// AIResponse represents the AI response.
type AIResponse struct {
	Success bool   `json:"success"`
	Summary string `json:"summary"`
	Error   string `json:"error,omitempty"`
}

// SmartReplyRequest represents a request for smart reply suggestions.
type SmartReplyRequest struct {
	EmailID   string `json:"email_id"`
	Subject   string `json:"subject"`
	From      string `json:"from"`
	Body      string `json:"body"`
	ReplyType string `json:"reply_type"` // "reply" or "reply_all"
}

// SmartReplyResponse represents the smart reply suggestions.
type SmartReplyResponse struct {
	Success bool     `json:"success"`
	Replies []string `json:"replies"`
	Error   string   `json:"error,omitempty"`
}

// EnhancedSummaryRequest represents an enhanced summary request.
type EnhancedSummaryRequest struct {
	EmailID string `json:"email_id"`
	Subject string `json:"subject"`
	From    string `json:"from"`
	Body    string `json:"body"`
}

// EnhancedSummaryResponse represents the enhanced summary with action items and sentiment.
type EnhancedSummaryResponse struct {
	Success     bool     `json:"success"`
	Summary     string   `json:"summary"`
	ActionItems []string `json:"action_items"`
	Sentiment   string   `json:"sentiment"` // "positive", "neutral", "negative", "urgent"
	Category    string   `json:"category"`  // "meeting", "task", "fyi", "question", "social"
	Error       string   `json:"error,omitempty"`
}

// AutoLabelRequest represents a request to auto-label an email.
type AutoLabelRequest struct {
	EmailID string `json:"email_id"`
	Subject string `json:"subject"`
	From    string `json:"from"`
	Body    string `json:"body"`
}

// AutoLabelResponse represents the auto-label response.
type AutoLabelResponse struct {
	Success  bool     `json:"success"`
	Labels   []string `json:"labels"`
	Category string   `json:"category"` // Primary category
	Priority string   `json:"priority"` // "high", "normal", "low"
	Error    string   `json:"error,omitempty"`
}

// ThreadSummaryRequest represents a request to summarize a thread.
type ThreadSummaryRequest struct {
	ThreadID string          `json:"thread_id"`
	Messages []ThreadMessage `json:"messages"`
}

// ThreadMessage represents a message in a thread for summarization.
type ThreadMessage struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Date    int64  `json:"date"`
}

// ThreadSummaryResponse represents the thread summary response.
type ThreadSummaryResponse struct {
	Success      bool     `json:"success"`
	Summary      string   `json:"summary"`
	KeyPoints    []string `json:"key_points"`
	ActionItems  []string `json:"action_items"`
	Participants []string `json:"participants"`
	Timeline     string   `json:"timeline"` // Brief timeline of the conversation
	NextSteps    string   `json:"next_steps,omitempty"`
	MessageCount int      `json:"message_count"`
	Error        string   `json:"error,omitempty"`
}
