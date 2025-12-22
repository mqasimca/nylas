//go:build integration

// Package integration provides integration tests for all CLI commands.
// Run with: go test -tags=integration -v ./internal/cli/integration/...
//
// Required environment variables:
//   - NYLAS_API_KEY: Your Nylas API key
//   - NYLAS_GRANT_ID: A valid grant ID
//   - NYLAS_CLIENT_ID: Your Nylas client ID (optional)
//
// Optional environment variables:
//   - NYLAS_TEST_EMAIL: Email address for send tests (default: uses grant email)
//   - NYLAS_TEST_SEND_EMAIL: Set to "true" to enable send tests
//   - NYLAS_TEST_DELETE: Set to "true" to enable delete tests
//
// Test files are organized by feature:
//   - test.go: Common setup and helpers (this file)
//   - auth_test.go: Auth command tests
//   - email_test.go: Email command tests
//   - folders_test.go: Folder command tests
//   - threads_test.go: Thread command tests
//   - drafts_test.go: Draft command tests
//   - calendar_test.go: Calendar command tests
//   - contacts_test.go: Contact command tests
//   - webhooks_test.go: Webhook command tests
//   - misc_test.go: Help, error handling, workflow tests
package integration

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"gopkg.in/yaml.v3"
)

// Test configuration loaded from environment
var (
	testAPIKey   string
	testGrantID  string
	testClientID string
	testEmail    string
	testBinary   string
)

func init() {
	testAPIKey = os.Getenv("NYLAS_API_KEY")
	testGrantID = os.Getenv("NYLAS_GRANT_ID")
	testClientID = os.Getenv("NYLAS_CLIENT_ID")
	testEmail = os.Getenv("NYLAS_TEST_EMAIL")

	// Find the binary - try environment variable first, then common locations
	testBinary = os.Getenv("NYLAS_TEST_BINARY")
	if testBinary != "" {
		// If provided, try to make it absolute
		if !strings.HasPrefix(testBinary, "/") {
			if abs, err := exec.LookPath(testBinary); err == nil {
				testBinary = abs
			}
		}
		return
	}

	// Try to find binary relative to test directory
	candidates := []string{
		"../../bin/nylas",    // From internal/cli
		"../../../bin/nylas", // From internal/cli/subdir
		"./bin/nylas",        // From project root
		"bin/nylas",          // From project root
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			testBinary = c
			break
		}
	}
}

// rateLimitDelay adds a small delay between API calls to avoid rate limiting.
const rateLimitDelay = 500 * time.Millisecond

// skipIfMissingCreds skips the test if required credentials are missing.
// It also adds a rate limit delay to avoid hitting API rate limits.
func skipIfMissingCreds(t *testing.T) {
	// Add delay to avoid rate limiting between tests
	time.Sleep(rateLimitDelay)

	if testBinary == "" {
		t.Skip("CLI binary not found - run 'go build -o bin/nylas ./cmd/nylas' first")
	}
	if testAPIKey == "" {
		t.Skip("NYLAS_API_KEY not set")
	}
	if testGrantID == "" {
		t.Skip("NYLAS_GRANT_ID not set")
	}
}

// runCLI executes a CLI command and returns stdout, stderr, and error
func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command(testBinary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Build environment with all necessary variables
	env := []string{
		"NYLAS_API_KEY=" + testAPIKey,
		"NYLAS_GRANT_ID=" + testGrantID,
		"NYLAS_DISABLE_KEYRING=true", // Disable keyring during tests to avoid macOS prompts
	}

	// Pass through AI provider credentials if set
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		env = append(env, "OPENAI_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		env = append(env, "GROQ_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		env = append(env, "OPENROUTER_API_KEY="+apiKey)
	}
	if ollamaHost := os.Getenv("OLLAMA_HOST"); ollamaHost != "" {
		env = append(env, "OLLAMA_HOST="+ollamaHost)
	}

	// Set environment for the CLI
	cmd.Env = append(os.Environ(), env...)

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// runCLIWithInput executes a CLI command with stdin input
func runCLIWithInput(input string, args ...string) (string, string, error) {
	cmd := exec.Command(testBinary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(input)

	// Build environment with all necessary variables
	env := []string{
		"NYLAS_API_KEY=" + testAPIKey,
		"NYLAS_GRANT_ID=" + testGrantID,
		"NYLAS_DISABLE_KEYRING=true", // Disable keyring during tests to avoid macOS prompts
	}

	// Pass through AI provider credentials if set
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		env = append(env, "OPENAI_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		env = append(env, "GROQ_API_KEY="+apiKey)
	}
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		env = append(env, "OPENROUTER_API_KEY="+apiKey)
	}
	if ollamaHost := os.Getenv("OLLAMA_HOST"); ollamaHost != "" {
		env = append(env, "OLLAMA_HOST="+ollamaHost)
	}

	cmd.Env = append(os.Environ(), env...)

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// getTestClient creates a test API client
func getTestClient() *nylas.HTTPClient {
	client := nylas.NewHTTPClient()
	client.SetCredentials(testClientID, "", testAPIKey)
	return client
}

// skipIfProviderNotSupported checks if the stderr indicates the provider doesn't support
// the operation and skips the test if so.
func skipIfProviderNotSupported(t *testing.T, stderr string) {
	t.Helper()
	// Various error messages that indicate provider limitation
	if strings.Contains(stderr, "Method not supported for provider") ||
		strings.Contains(stderr, "an internal error ocurred") || // Nylas API typo
		strings.Contains(stderr, "an internal error occurred") {
		t.Skipf("Provider does not support this operation: %s", strings.TrimSpace(stderr))
	}
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// hasAnyAIProvider checks if any AI provider is configured
func hasAnyAIProvider() bool {
	return checkOllamaAvailable() ||
		os.Getenv("ANTHROPIC_API_KEY") != "" ||
		os.Getenv("OPENAI_API_KEY") != "" ||
		os.Getenv("GROQ_API_KEY") != ""
}

// checkOllamaAvailable checks if Ollama is running
func checkOllamaAvailable() bool {
	// Check if Ollama is running by making a request to its API
	client := &http.Client{Timeout: 2 * time.Second}

	// Try common Ollama locations
	hosts := []string{
		"http://localhost:11434",
		"http://192.168.1.100:11434",
		"http://linux.local:11434",
	}

	for _, host := range hosts {
		resp, err := client.Get(host + "/api/tags")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
	}

	return false
}

// getTestEmail returns the test email from environment or default
func getTestEmail() string {
	if testEmail != "" {
		return testEmail
	}
	return getEnvOrDefault("NYLAS_TEST_EMAIL", "")
}

// extractEventID extracts event ID from CLI output
func extractEventID(output string) string {
	// Look for event ID patterns in output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for "ID: <id>" or "Event ID: <id>"
		if strings.Contains(line, "Event ID:") || strings.Contains(line, "ID:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if (part == "ID:" || part == "Event" && i+1 < len(parts) && parts[i+1] == "ID:") && i+1 < len(parts) {
					// Next field should be the ID
					nextIdx := i + 1
					if part == "Event" {
						nextIdx = i + 2
					}
					if nextIdx < len(parts) {
						return parts[nextIdx]
					}
				}
			}
		}
		// Also try to match event_* or cal_event_* patterns
		if strings.Contains(line, "event_") || strings.Contains(line, "cal_event_") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "event_") || strings.HasPrefix(part, "cal_event_") {
					// Clean up any trailing punctuation
					id := strings.TrimRight(part, ".,;:\"'")
					return id
				}
			}
		}
	}
	return ""
}

// extractEventIDFromList extracts event ID from list output by finding title
func extractEventIDFromList(output, title string) string {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, title) {
			// Look for ID in the same line or nearby lines
			if strings.Contains(line, "ID:") {
				parts := strings.Split(line, "ID:")
				if len(parts) > 1 {
					idPart := strings.TrimSpace(parts[1])
					fields := strings.Fields(idPart)
					if len(fields) > 0 {
						return fields[0]
					}
				}
			}
			// Check previous lines for ID
			for j := i - 1; j >= 0 && j >= i-3; j-- {
				if strings.Contains(lines[j], "ID:") {
					parts := strings.Split(lines[j], "ID:")
					if len(parts) > 1 {
						idPart := strings.TrimSpace(parts[1])
						fields := strings.Fields(idPart)
						if len(fields) > 0 {
							return fields[0]
						}
					}
				}
			}
		}
	}
	return ""
}

// getAvailableProvider returns the first available AI provider
func getAvailableProvider() string {
	if checkOllamaAvailable() {
		return "ollama"
	}
	if getEnvOrDefault("ANTHROPIC_API_KEY", "") != "" {
		return "claude"
	}
	if getEnvOrDefault("OPENAI_API_KEY", "") != "" {
		return "openai"
	}
	if getEnvOrDefault("GROQ_API_KEY", "") != "" {
		return "groq"
	}
	return "" // Return empty string when no provider is available
}

// getAIConfigFromUserConfig reads the AI configuration from ~/.config/nylas/config.yaml
func getAIConfigFromUserConfig() map[string]interface{} {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configPath := filepath.Join(home, ".config", "nylas", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil
	}

	aiConfig, ok := config["ai"].(map[string]interface{})
	if !ok {
		return nil
	}

	return aiConfig
}

// Ensure imports are used (these are used in other test files)
var (
	_ = context.Background
	_ = domain.ProviderGoogle
)
