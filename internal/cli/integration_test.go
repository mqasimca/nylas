//go:build integration

// Package cli provides integration tests for all CLI commands.
// Run with: go test -tags=integration -v ./internal/cli/...
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
// Test files are organized by command:
//   - integration_test.go: Common setup and helpers (this file)
//   - integration_auth_test.go: Auth command tests
//   - integration_email_test.go: Email command tests
//   - integration_folders_test.go: Folder command tests
//   - integration_threads_test.go: Thread command tests
//   - integration_drafts_test.go: Draft command tests
//   - integration_calendar_test.go: Calendar command tests
//   - integration_contacts_test.go: Contact command tests
//   - integration_webhooks_test.go: Webhook command tests
//   - integration_misc_test.go: Help, error handling, workflow tests
package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
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

	// Set environment for the CLI
	cmd.Env = append(os.Environ(),
		"NYLAS_API_KEY="+testAPIKey,
		"NYLAS_GRANT_ID="+testGrantID,
	)

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

	cmd.Env = append(os.Environ(),
		"NYLAS_API_KEY="+testAPIKey,
		"NYLAS_GRANT_ID="+testGrantID,
	)

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

// Ensure imports are used (these are used in other test files)
var (
	_ = context.Background
	_ = domain.ProviderGoogle
)
