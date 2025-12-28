package air

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

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

// handleAISummarize handles POST /api/ai/summarize requests.
func (s *Server) handleAISummarize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Prompt is required",
		})
		return
	}

	// Run claude -p with the prompt
	summary, err := runClaudeCommand(req.Prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, AIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, AIResponse{
		Success: true,
		Summary: summary,
	})
}

// handleAISmartReplies handles POST /api/ai/smart-replies requests.
func (s *Server) handleAISmartReplies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SmartReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SmartReplyResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Body == "" {
		writeJSON(w, http.StatusBadRequest, SmartReplyResponse{
			Success: false,
			Error:   "Email body is required",
		})
		return
	}

	// Truncate body for prompt
	body := req.Body
	if len(body) > 2000 {
		body = body[:2000] + "..."
	}

	// Build prompt for smart replies
	prompt := fmt.Sprintf(`Generate exactly 3 short, professional email reply suggestions for this email. Each reply should be 1-2 sentences max. Return ONLY a JSON array of 3 strings, nothing else.

From: %s
Subject: %s

%s

Return format: ["Reply 1", "Reply 2", "Reply 3"]`, req.From, req.Subject, body)

	result, err := runClaudeCommand(prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SmartReplyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Parse the JSON array from the result
	var replies []string
	// Try to extract JSON array from response
	start := strings.Index(result, "[")
	end := strings.LastIndex(result, "]")
	if start != -1 && end != -1 && end > start {
		jsonStr := result[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &replies); err != nil {
			// Fallback: split by newlines if JSON parsing fails
			replies = parseRepliesFromText(result)
		}
	} else {
		replies = parseRepliesFromText(result)
	}

	// Ensure we have exactly 3 replies
	for len(replies) < 3 {
		replies = append(replies, "Thanks for your email!")
	}
	if len(replies) > 3 {
		replies = replies[:3]
	}

	writeJSON(w, http.StatusOK, SmartReplyResponse{
		Success: true,
		Replies: replies,
	})
}

// parseRepliesFromText extracts reply suggestions from plain text.
func parseRepliesFromText(text string) []string {
	var replies []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove numbering like "1.", "2.", etc.
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' && line[1] == '.' {
			line = strings.TrimSpace(line[2:])
		}
		// Remove quotes
		line = strings.Trim(line, `"'`)
		if len(line) > 10 && len(line) < 200 {
			replies = append(replies, line)
		}
	}
	return replies
}

// handleAIEnhancedSummary handles POST /api/ai/enhanced-summary requests.
func (s *Server) handleAIEnhancedSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnhancedSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, EnhancedSummaryResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Body == "" {
		writeJSON(w, http.StatusBadRequest, EnhancedSummaryResponse{
			Success: false,
			Error:   "Email body is required",
		})
		return
	}

	// Truncate body for prompt
	body := req.Body
	if len(body) > 3000 {
		body = body[:3000] + "..."
	}

	// Build prompt for enhanced summary
	prompt := fmt.Sprintf(`Analyze this email and provide a structured response in JSON format.

From: %s
Subject: %s

%s

Return ONLY valid JSON in this exact format:
{
  "summary": "2-3 sentence summary of the email",
  "action_items": ["action 1", "action 2"],
  "sentiment": "positive|neutral|negative|urgent",
  "category": "meeting|task|fyi|question|social"
}

Rules:
- action_items: List specific tasks or requests. Empty array if none.
- sentiment: Choose ONE based on tone and urgency
- category: Choose the PRIMARY purpose of the email`, req.From, req.Subject, body)

	result, err := runClaudeCommand(prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EnhancedSummaryResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Parse JSON response
	var parsed struct {
		Summary     string   `json:"summary"`
		ActionItems []string `json:"action_items"`
		Sentiment   string   `json:"sentiment"`
		Category    string   `json:"category"`
	}

	// Try to extract JSON from response
	start := strings.Index(result, "{")
	end := strings.LastIndex(result, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr := result[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			// Fallback to basic summary
			writeJSON(w, http.StatusOK, EnhancedSummaryResponse{
				Success:     true,
				Summary:     result,
				ActionItems: []string{},
				Sentiment:   "neutral",
				Category:    "fyi",
			})
			return
		}
	} else {
		// No JSON found, use raw result as summary
		writeJSON(w, http.StatusOK, EnhancedSummaryResponse{
			Success:     true,
			Summary:     result,
			ActionItems: []string{},
			Sentiment:   "neutral",
			Category:    "fyi",
		})
		return
	}

	// Validate sentiment
	validSentiments := map[string]bool{"positive": true, "neutral": true, "negative": true, "urgent": true}
	if !validSentiments[parsed.Sentiment] {
		parsed.Sentiment = "neutral"
	}

	// Validate category
	validCategories := map[string]bool{"meeting": true, "task": true, "fyi": true, "question": true, "social": true}
	if !validCategories[parsed.Category] {
		parsed.Category = "fyi"
	}

	writeJSON(w, http.StatusOK, EnhancedSummaryResponse{
		Success:     true,
		Summary:     parsed.Summary,
		ActionItems: parsed.ActionItems,
		Sentiment:   parsed.Sentiment,
		Category:    parsed.Category,
	})
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

// handleAIAutoLabel handles POST /api/ai/auto-label requests.
func (s *Server) handleAIAutoLabel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AutoLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, AutoLabelResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Body == "" && req.Subject == "" {
		writeJSON(w, http.StatusBadRequest, AutoLabelResponse{
			Success: false,
			Error:   "Email subject or body is required",
		})
		return
	}

	// Truncate body for prompt
	body := req.Body
	if len(body) > 2000 {
		body = body[:2000] + "..."
	}

	// Build prompt for auto-labeling
	prompt := fmt.Sprintf(`Analyze this email and suggest appropriate labels. Return ONLY valid JSON.

From: %s
Subject: %s

%s

Return format:
{
  "labels": ["label1", "label2"],
  "category": "one of: meeting|task|fyi|question|social|newsletter|promotion|urgent|personal|work",
  "priority": "one of: high|normal|low"
}

Rules:
- labels: 1-4 relevant labels (e.g., "finance", "project-x", "team", "client")
- category: Choose the PRIMARY purpose
- priority: Based on urgency and sender importance`, req.From, req.Subject, body)

	result, err := runClaudeCommand(prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, AutoLabelResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Parse JSON response
	var parsed struct {
		Labels   []string `json:"labels"`
		Category string   `json:"category"`
		Priority string   `json:"priority"`
	}

	// Try to extract JSON from response
	start := strings.Index(result, "{")
	end := strings.LastIndex(result, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr := result[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			// Fallback to defaults
			writeJSON(w, http.StatusOK, AutoLabelResponse{
				Success:  true,
				Labels:   []string{"inbox"},
				Category: "fyi",
				Priority: "normal",
			})
			return
		}
	} else {
		writeJSON(w, http.StatusOK, AutoLabelResponse{
			Success:  true,
			Labels:   []string{"inbox"},
			Category: "fyi",
			Priority: "normal",
		})
		return
	}

	// Validate category
	validCategories := map[string]bool{
		"meeting": true, "task": true, "fyi": true, "question": true,
		"social": true, "newsletter": true, "promotion": true,
		"urgent": true, "personal": true, "work": true,
	}
	if !validCategories[parsed.Category] {
		parsed.Category = "fyi"
	}

	// Validate priority
	validPriorities := map[string]bool{"high": true, "normal": true, "low": true}
	if !validPriorities[parsed.Priority] {
		parsed.Priority = "normal"
	}

	// Ensure at least one label
	if len(parsed.Labels) == 0 {
		parsed.Labels = []string{"inbox"}
	}

	writeJSON(w, http.StatusOK, AutoLabelResponse{
		Success:  true,
		Labels:   parsed.Labels,
		Category: parsed.Category,
		Priority: parsed.Priority,
	})
}

// handleAIThreadSummary handles POST /api/ai/thread-summary requests.
func (s *Server) handleAIThreadSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ThreadSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ThreadSummaryResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, ThreadSummaryResponse{
			Success: false,
			Error:   "At least one message is required",
		})
		return
	}

	// Build conversation text from messages
	var conversationBuilder strings.Builder
	participants := make(map[string]bool)

	for i, msg := range req.Messages {
		// Track participants
		if msg.From != "" {
			participants[msg.From] = true
		}

		// Truncate individual message bodies
		body := msg.Body
		if len(body) > 1000 {
			body = body[:1000] + "..."
		}

		conversationBuilder.WriteString(fmt.Sprintf("--- Message %d ---\n", i+1))
		conversationBuilder.WriteString(fmt.Sprintf("From: %s\n", msg.From))
		if msg.Subject != "" {
			conversationBuilder.WriteString(fmt.Sprintf("Subject: %s\n", msg.Subject))
		}
		conversationBuilder.WriteString(body)
		conversationBuilder.WriteString("\n\n")

		// Limit total conversation length
		if conversationBuilder.Len() > 6000 {
			conversationBuilder.WriteString("... (additional messages truncated)")
			break
		}
	}

	// Get participant list
	participantList := make([]string, 0, len(participants))
	for p := range participants {
		participantList = append(participantList, p)
	}

	// Build prompt for thread summary
	prompt := fmt.Sprintf(`Summarize this email thread conversation. Return ONLY valid JSON.

%s

Return format:
{
  "summary": "2-4 sentence overall summary of the thread",
  "key_points": ["point 1", "point 2", "point 3"],
  "action_items": ["action 1", "action 2"],
  "timeline": "Brief timeline description (e.g., 'Started Monday with request, followed up Wednesday, resolved Friday')",
  "next_steps": "What needs to happen next, if anything"
}

Rules:
- summary: Capture the main topic and outcome of the conversation
- key_points: 2-5 most important points discussed
- action_items: Specific tasks mentioned (empty array if none)
- timeline: Brief description of how the conversation evolved
- next_steps: Clear next action if any, or empty string`, conversationBuilder.String())

	result, err := runClaudeCommand(prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ThreadSummaryResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Parse JSON response
	var parsed struct {
		Summary     string   `json:"summary"`
		KeyPoints   []string `json:"key_points"`
		ActionItems []string `json:"action_items"`
		Timeline    string   `json:"timeline"`
		NextSteps   string   `json:"next_steps"`
	}

	// Try to extract JSON from response
	start := strings.Index(result, "{")
	end := strings.LastIndex(result, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr := result[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			// Fallback to raw result
			writeJSON(w, http.StatusOK, ThreadSummaryResponse{
				Success:      true,
				Summary:      result,
				KeyPoints:    []string{},
				ActionItems:  []string{},
				Participants: participantList,
				Timeline:     "",
				MessageCount: len(req.Messages),
			})
			return
		}
	} else {
		writeJSON(w, http.StatusOK, ThreadSummaryResponse{
			Success:      true,
			Summary:      result,
			KeyPoints:    []string{},
			ActionItems:  []string{},
			Participants: participantList,
			Timeline:     "",
			MessageCount: len(req.Messages),
		})
		return
	}

	writeJSON(w, http.StatusOK, ThreadSummaryResponse{
		Success:      true,
		Summary:      parsed.Summary,
		KeyPoints:    parsed.KeyPoints,
		ActionItems:  parsed.ActionItems,
		Participants: participantList,
		Timeline:     parsed.Timeline,
		NextSteps:    parsed.NextSteps,
		MessageCount: len(req.Messages),
	})
}

// runClaudeCommand runs the claude CLI with the given prompt.
func runClaudeCommand(prompt string) (string, error) {
	// Create context with timeout (30 seconds for AI response)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find claude binary
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude code CLI not found: please install it from https://claude.ai/code")
	}

	// Create command: echo "prompt" | claude -p
	// #nosec G204 -- claudePath verified via exec.LookPath from system PATH, user prompt only in stdin (not in command path)
	cmd := exec.CommandContext(ctx, claudePath, "-p")
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude code timed out after 30 seconds")
		}
		// Return stderr if available
		if stderr.Len() > 0 {
			return "", fmt.Errorf("claude code error: %s", stderr.String())
		}
		return "", fmt.Errorf("claude code error: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}
