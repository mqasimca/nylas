package air

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

// CompleteRequest represents a smart compose request
type CompleteRequest struct {
	Text      string `json:"text"`
	MaxLength int    `json:"maxLength"`
	Context   string `json:"context,omitempty"`
}

// CompleteResponse represents a smart compose response
type CompleteResponse struct {
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
}

// handleAIComplete handles smart compose autocomplete requests
func (s *Server) handleAIComplete(w http.ResponseWriter, r *http.Request) {
	var req CompleteRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		w.Header().Set("Content-Type", "application/json")
		resp := CompleteResponse{Suggestion: "", Confidence: 0}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode", http.StatusInternalServerError)
		}
		return
	}

	if req.MaxLength == 0 {
		req.MaxLength = 100
	}

	suggestion := getAICompletion(req.Text, req.MaxLength)

	w.Header().Set("Content-Type", "application/json")
	resp := CompleteResponse{
		Suggestion: suggestion,
		Confidence: 0.8,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getAICompletion gets completion from Claude via CLI
func getAICompletion(text string, maxLen int) string {
	prompt := buildCompletionPrompt(text, maxLen)

	//nolint:gosec // G204: Command is hardcoded "claude" binary, prompt is user-controlled but safe for CLI arg
	cmd := exec.Command("claude", "-p", prompt)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	suggestion := strings.TrimSpace(string(output))

	// Limit length
	if len(suggestion) > maxLen {
		// Try to break at word boundary
		if idx := strings.LastIndex(suggestion[:maxLen], " "); idx > 0 {
			suggestion = suggestion[:idx]
		} else {
			suggestion = suggestion[:maxLen]
		}
	}

	return suggestion
}

// buildCompletionPrompt creates prompt for autocomplete
func buildCompletionPrompt(text string, maxLen int) string {
	return strings.Join([]string{
		"You are an email autocomplete assistant.",
		"Complete the following email text naturally.",
		"Only provide the completion, not the original text.",
		"Keep it concise and professional.",
		"Maximum " + string(rune(maxLen)) + " characters.",
		"",
		"Text to complete:",
		text,
		"",
		"Completion:",
	}, "\n")
}

// NLSearchRequest represents a natural language search request
type NLSearchRequest struct {
	Query string `json:"query"`
}

// NLSearchResponse represents parsed search parameters
type NLSearchResponse struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Subject    string `json:"subject,omitempty"`
	DateAfter  string `json:"dateAfter,omitempty"`
	DateBefore string `json:"dateBefore,omitempty"`
	HasAttach  bool   `json:"hasAttachment,omitempty"`
	IsUnread   bool   `json:"isUnread,omitempty"`
	Keywords   string `json:"keywords,omitempty"`
}

// handleNLSearch handles natural language search queries
func (s *Server) handleNLSearch(w http.ResponseWriter, r *http.Request) {
	var req NLSearchRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query required", http.StatusBadRequest)
		return
	}

	result := parseNaturalLanguageSearch(req.Query)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// parseNaturalLanguageSearch converts NL query to search params
func parseNaturalLanguageSearch(query string) NLSearchResponse {
	result := NLSearchResponse{}
	queryLower := strings.ToLower(query)

	// Parse time-based patterns FIRST (before "from" to avoid conflicts)
	if strings.Contains(queryLower, "last week") {
		result.DateAfter = "7d"
	} else if strings.Contains(queryLower, "yesterday") {
		result.DateAfter = "1d"
	} else if strings.Contains(queryLower, "today") {
		result.DateAfter = "0d"
	} else if strings.Contains(queryLower, "this month") {
		result.DateAfter = "30d"
	}

	// Parse "from X" patterns (skip time words)
	timeWords := map[string]bool{"last": true, "yesterday": true, "today": true, "this": true}
	if strings.Contains(queryLower, "from ") && !strings.Contains(queryLower, "from last") &&
		!strings.Contains(queryLower, "from yesterday") && !strings.Contains(queryLower, "from today") {
		parts := strings.SplitN(queryLower, "from ", 2)
		if len(parts) > 1 {
			words := strings.Fields(parts[1])
			if len(words) > 0 && !timeWords[words[0]] {
				result.From = words[0]
			}
		}
	}

	// Parse "to X" patterns
	if strings.Contains(queryLower, "to ") {
		parts := strings.SplitN(queryLower, "to ", 2)
		if len(parts) > 1 {
			words := strings.Fields(parts[1])
			if len(words) > 0 {
				result.To = words[0]
			}
		}
	}

	// Parse attachment pattern
	if strings.Contains(queryLower, "attachment") ||
		strings.Contains(queryLower, "attached") {
		result.HasAttach = true
	}

	// Parse unread pattern
	if strings.Contains(queryLower, "unread") {
		result.IsUnread = true
	}

	// Extract remaining keywords
	keywords := extractKeywords(queryLower)
	if len(keywords) > 0 {
		result.Keywords = strings.Join(keywords, " ")
	}

	return result
}

// extractKeywords extracts search keywords from query
func extractKeywords(query string) []string {
	stopWords := map[string]bool{
		"from": true, "to": true, "about": true, "with": true,
		"the": true, "a": true, "an": true, "and": true,
		"or": true, "in": true, "on": true, "at": true,
		"last": true, "week": true, "month": true, "yesterday": true,
		"today": true, "emails": true, "email": true, "messages": true,
	}

	words := strings.Fields(query)
	keywords := []string{}

	for _, word := range words {
		word = strings.Trim(word, ".,!?")
		if !stopWords[word] && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}
