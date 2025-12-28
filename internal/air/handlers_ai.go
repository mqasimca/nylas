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
