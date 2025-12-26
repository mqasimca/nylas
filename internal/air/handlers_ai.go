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

// runClaudeCommand runs the claude CLI with the given prompt.
func runClaudeCommand(prompt string) (string, error) {
	// Create context with timeout (30 seconds for AI response)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find claude binary
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("Claude Code CLI not found. Please install it from https://claude.ai/code")
	}

	// Create command: echo "prompt" | claude -p
	cmd := exec.CommandContext(ctx, claudePath, "-p")
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Claude Code timed out after 30 seconds")
		}
		// Return stderr if available
		if stderr.Len() > 0 {
			return "", fmt.Errorf("Claude Code error: %s", stderr.String())
		}
		return "", fmt.Errorf("Claude Code error: %v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}
