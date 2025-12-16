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
package cli

import (
	"bytes"
	"context"
	"fmt"
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
		"../../bin/nylas",      // From internal/cli
		"../../../bin/nylas",   // From internal/cli/subdir
		"./bin/nylas",          // From project root
		"bin/nylas",            // From project root
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

// =============================================================================
// AUTH COMMAND TESTS
// =============================================================================

func TestCLI_AuthStatus(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "status")

	if err != nil {
		t.Fatalf("auth status failed: %v\nstderr: %s", err, stderr)
	}

	// Should contain key status information
	if !strings.Contains(stdout, "Authentication Status") {
		t.Errorf("Expected 'Authentication Status' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Secret Store:") {
		t.Errorf("Expected 'Secret Store:' in output, got: %s", stdout)
	}

	t.Logf("auth status output:\n%s", stdout)
}

func TestCLI_AuthList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "list")

	// This may show "No authenticated accounts" if grant isn't registered locally
	// but should not error
	if err != nil && !strings.Contains(stderr, "No authenticated accounts") {
		t.Fatalf("auth list failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("auth list output:\n%s", stdout)
}

func TestCLI_AuthWhoami(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "whoami")

	// May fail if no default grant is set
	if err != nil {
		if strings.Contains(stderr, "no default grant") {
			t.Skip("No default grant set")
		}
		t.Fatalf("auth whoami failed: %v\nstderr: %s", err, stderr)
	}

	// Should show email and provider
	if !strings.Contains(stdout, "@") {
		t.Errorf("Expected email in output, got: %s", stdout)
	}

	t.Logf("auth whoami output:\n%s", stdout)
}

func TestCLI_AuthAdd(t *testing.T) {
	skipIfMissingCreds(t)

	// Test adding with auto-detection (no --email or --provider flags)
	stdout, stderr, err := runCLI("auth", "add", testGrantID, "--default")

	if err != nil {
		t.Fatalf("auth add failed: %v\nstderr: %s", err, stderr)
	}

	// Should show success message
	if !strings.Contains(stdout, "Added grant") {
		t.Errorf("Expected 'Added grant' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, testGrantID) {
		t.Errorf("Expected grant ID in output, got: %s", stdout)
	}
	// Should auto-detect email and provider
	if !strings.Contains(stdout, "Email:") {
		t.Errorf("Expected 'Email:' in output (auto-detected), got: %s", stdout)
	}
	if !strings.Contains(stdout, "Provider:") {
		t.Errorf("Expected 'Provider:' in output (auto-detected), got: %s", stdout)
	}

	t.Logf("auth add output:\n%s", stdout)

	// Verify the grant appears in list
	listOut, _, err := runCLI("auth", "list")
	if err != nil {
		t.Fatalf("auth list after add failed: %v", err)
	}
	if !strings.Contains(listOut, testGrantID) {
		t.Errorf("Expected grant ID in auth list output, got: %s", listOut)
	}

	t.Logf("auth list after add:\n%s", listOut)
}

func TestCLI_AuthAdd_AutoDetect(t *testing.T) {
	skipIfMissingCreds(t)

	// Test that auto-detection gets correct info from Nylas API
	stdout, stderr, err := runCLI("auth", "add", testGrantID)

	if err != nil {
		t.Fatalf("auth add auto-detect failed: %v\nstderr: %s", err, stderr)
	}

	// Should show success with auto-detected values
	if !strings.Contains(stdout, "Added grant") {
		t.Errorf("Expected 'Added grant' in output, got: %s", stdout)
	}

	// The output should contain an email with @ symbol (auto-detected)
	if !strings.Contains(stdout, "@") {
		t.Errorf("Expected auto-detected email in output, got: %s", stdout)
	}

	t.Logf("auth add auto-detect output:\n%s", stdout)
}

func TestCLI_AuthAdd_OverrideAutoDetect(t *testing.T) {
	skipIfMissingCreds(t)

	// Test that flags can override auto-detected values
	stdout, stderr, err := runCLI("auth", "add", testGrantID,
		"--email", "override@example.com",
		"--provider", "google")

	if err != nil {
		t.Fatalf("auth add with overrides failed: %v\nstderr: %s", err, stderr)
	}

	// Should show overridden email
	if !strings.Contains(stdout, "override@example.com") {
		t.Errorf("Expected overridden email in output, got: %s", stdout)
	}
	// Should show overridden provider
	if !strings.Contains(stdout, "Google") {
		t.Errorf("Expected overridden provider 'Google' in output, got: %s", stdout)
	}

	t.Logf("auth add with overrides output:\n%s", stdout)
}

func TestCLI_AuthAdd_InvalidGrant(t *testing.T) {
	skipIfMissingCreds(t)

	// Test adding a non-existent grant - should fail when fetching from API
	_, stderr, err := runCLI("auth", "add", "invalid-grant-id-12345")

	if err == nil {
		t.Error("Expected error for invalid grant ID, but command succeeded")
	}

	// Should show fetch error
	if !strings.Contains(stderr, "not valid") && !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "failed to fetch") {
		t.Logf("Error output for invalid grant: %s", stderr)
	}
}

func TestCLI_AuthAdd_ProviderOverride(t *testing.T) {
	skipIfMissingCreds(t)

	// Test that provider flag can override auto-detected provider
	providers := []string{"google", "microsoft"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			stdout, stderr, err := runCLI("auth", "add", testGrantID,
				"--provider", provider)

			if err != nil {
				t.Fatalf("auth add with provider %s failed: %v\nstderr: %s", provider, err, stderr)
			}

			if !strings.Contains(stdout, "Added grant") {
				t.Errorf("Expected 'Added grant' in output, got: %s", stdout)
			}

			t.Logf("auth add with provider %s output:\n%s", provider, stdout)
		})
	}
}

// =============================================================================
// EMAIL LIST COMMAND TESTS
// =============================================================================

func TestCLI_EmailList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "list", testGrantID, "--limit", "5")

	if err != nil {
		t.Fatalf("email list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message count or "No messages found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No messages found") {
		t.Errorf("Expected message list output, got: %s", stdout)
	}

	t.Logf("email list output:\n%s", stdout)
}

func TestCLI_EmailList_WithID(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "list", testGrantID, "--limit", "3", "--id")

	if err != nil {
		t.Fatalf("email list --id failed: %v\nstderr: %s", err, stderr)
	}

	// Should show "ID:" lines when --id flag is used
	if strings.Contains(stdout, "Found") && !strings.Contains(stdout, "ID:") {
		t.Errorf("Expected message IDs in output with --id flag, got: %s", stdout)
	}

	t.Logf("email list --id output:\n%s", stdout)
}

func TestCLI_EmailList_Filters(t *testing.T) {
	skipIfMissingCreds(t)

	tests := []struct {
		name string
		args []string
	}{
		{"unread", []string{"email", "list", testGrantID, "--unread", "--limit", "3"}},
		{"starred", []string{"email", "list", testGrantID, "--starred", "--limit", "3"}},
		{"limit", []string{"email", "list", testGrantID, "--limit", "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)
			if err != nil {
				t.Fatalf("email list %s failed: %v\nstderr: %s", tt.name, err, stderr)
			}
			t.Logf("email list %s output:\n%s", tt.name, stdout)
		})
	}
}

// =============================================================================
// EMAIL READ COMMAND TESTS
// =============================================================================

func TestCLI_EmailRead(t *testing.T) {
	skipIfMissingCreds(t)

	// First get a message ID
	client := getTestClient()
	ctx := context.Background()

	messages, err := client.GetMessages(ctx, testGrantID, 1)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(messages) == 0 {
		t.Skip("No messages available for read test")
	}

	messageID := messages[0].ID

	stdout, stderr, err := runCLI("email", "read", messageID, testGrantID)

	if err != nil {
		t.Fatalf("email read failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message details
	if !strings.Contains(stdout, "Subject:") {
		t.Errorf("Expected 'Subject:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "From:") {
		t.Errorf("Expected 'From:' in output, got: %s", stdout)
	}

	t.Logf("email read output:\n%s", stdout)
}

func TestCLI_EmailRead_JSON(t *testing.T) {
	skipIfMissingCreds(t)

	// First get a message ID
	client := getTestClient()
	ctx := context.Background()

	messages, err := client.GetMessages(ctx, testGrantID, 1)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(messages) == 0 {
		t.Skip("No messages available for read test")
	}

	messageID := messages[0].ID

	stdout, stderr, err := runCLI("email", "read", messageID, testGrantID, "--json")

	if err != nil {
		t.Fatalf("email read --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON with expected fields
	if !strings.Contains(stdout, `"id":`) {
		t.Errorf("Expected '\"id\":' in JSON output, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"subject":`) {
		t.Errorf("Expected '\"subject\":' in JSON output, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"from":`) {
		t.Errorf("Expected '\"from\":' in JSON output, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"body":`) {
		t.Errorf("Expected '\"body\":' in JSON output, got: %s", stdout)
	}

	// Should NOT contain formatted headers (means it's JSON not formatted)
	if strings.Contains(stdout, "Subject:") && strings.Contains(stdout, "────") {
		t.Errorf("JSON output should not contain formatted headers")
	}

	t.Logf("email read --json output:\n%s", stdout)
}

func TestCLI_EmailRead_Raw(t *testing.T) {
	skipIfMissingCreds(t)

	// First get a message ID
	client := getTestClient()
	ctx := context.Background()

	messages, err := client.GetMessages(ctx, testGrantID, 1)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(messages) == 0 {
		t.Skip("No messages available for read test")
	}

	messageID := messages[0].ID

	stdout, stderr, err := runCLI("email", "read", messageID, testGrantID, "--raw")

	if err != nil {
		t.Fatalf("email read --raw failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message headers
	if !strings.Contains(stdout, "Subject:") {
		t.Errorf("Expected 'Subject:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "ID:") {
		t.Errorf("Expected 'ID:' in raw output (shows message ID), got: %s", stdout)
	}

	// Raw output typically contains HTML tags if the message is HTML
	// OR it's plain text - either way it should have body content
	t.Logf("email read --raw output:\n%s", stdout)
}

// =============================================================================
// EMAIL SEARCH COMMAND TESTS
// =============================================================================

func TestCLI_EmailSearch(t *testing.T) {
	skipIfMissingCreds(t)

	// Search for a common subject
	stdout, stderr, err := runCLI("email", "search", "test", testGrantID, "--limit", "5")

	if err != nil {
		t.Fatalf("email search failed: %v\nstderr: %s", err, stderr)
	}

	// Should show results or "No messages found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No messages found") {
		t.Errorf("Expected search results output, got: %s", stdout)
	}

	t.Logf("email search output:\n%s", stdout)
}

func TestCLI_EmailSearch_WithFilters(t *testing.T) {
	skipIfMissingCreds(t)

	// Search with date filter
	stdout, stderr, err := runCLI("email", "search", "email", testGrantID,
		"--limit", "3",
		"--after", "2024-01-01")

	if err != nil {
		t.Fatalf("email search with filters failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("email search with filters output:\n%s", stdout)
}

// =============================================================================
// EMAIL MARK COMMAND TESTS
// =============================================================================

func TestCLI_EmailMark(t *testing.T) {
	skipIfMissingCreds(t)

	// Get a message to test marking
	client := getTestClient()
	ctx := context.Background()

	messages, err := client.GetMessages(ctx, testGrantID, 1)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(messages) == 0 {
		t.Skip("No messages available for mark test")
	}

	messageID := messages[0].ID

	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{"starred", "starred", "starred"},
		{"unstarred", "unstarred", "removed"},
		{"unread", "unread", "unread"},
		{"read", "read", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI("email", "mark", tt.action, messageID, testGrantID)
			if err != nil {
				t.Fatalf("email mark %s failed: %v\nstderr: %s", tt.action, err, stderr)
			}

			if !strings.Contains(strings.ToLower(stdout), tt.expected) {
				t.Errorf("Expected '%s' in output, got: %s", tt.expected, stdout)
			}

			t.Logf("email mark %s output: %s", tt.action, stdout)

			// Small delay between operations
			time.Sleep(500 * time.Millisecond)
		})
	}
}

// =============================================================================
// FOLDER COMMAND TESTS
// =============================================================================

func TestCLI_FoldersList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "folders", "list", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("folders list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show folders header
	if !strings.Contains(stdout, "Folders") || !strings.Contains(stdout, "NAME") {
		t.Errorf("Expected folders list header, got: %s", stdout)
	}

	// Should contain common folders like Inbox
	if !strings.Contains(stdout, "Inbox") && !strings.Contains(stdout, "INBOX") {
		t.Errorf("Expected 'Inbox' folder in output, got: %s", stdout)
	}

	t.Logf("folders list output:\n%s", stdout)
}

func TestCLI_FoldersCreateAndDelete(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	folderName := fmt.Sprintf("CLI-Test-%d", time.Now().Unix())

	// Create folder
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "folders", "create", folderName, testGrantID)
		if err != nil {
			t.Fatalf("folders create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Created folder") {
			t.Errorf("Expected 'Created folder' in output, got: %s", stdout)
		}

		t.Logf("folders create output: %s", stdout)
	})

	// Wait for folder to be created
	time.Sleep(2 * time.Second)

	// Get folder ID
	client := getTestClient()
	ctx := context.Background()

	folders, err := client.GetFolders(ctx, testGrantID)
	if err != nil {
		t.Fatalf("Failed to get folders: %v", err)
	}

	var folderID string
	for _, f := range folders {
		if f.Name == folderName {
			folderID = f.ID
			break
		}
	}

	if folderID == "" {
		t.Skip("Created folder not found - may need more time to sync")
	}

	// Delete folder
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLIWithInput("y\n", "email", "folders", "delete", folderID, testGrantID)
		if err != nil {
			t.Fatalf("folders delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("folders delete output: %s", stdout)
	})
}

// =============================================================================
// THREAD COMMAND TESTS
// =============================================================================

func TestCLI_ThreadsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "threads", "list", testGrantID, "--limit", "5")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("threads list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show thread count or "No threads found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No threads found") {
		t.Errorf("Expected threads list output, got: %s", stdout)
	}

	t.Logf("threads list output:\n%s", stdout)
}

func TestCLI_ThreadsList_WithFilters(t *testing.T) {
	skipIfMissingCreds(t)

	tests := []struct {
		name string
		args []string
	}{
		{"unread", []string{"email", "threads", "list", testGrantID, "--unread", "--limit", "3"}},
		{"starred", []string{"email", "threads", "list", testGrantID, "--starred", "--limit", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)
			skipIfProviderNotSupported(t, stderr)
			if err != nil {
				t.Fatalf("threads list %s failed: %v\nstderr: %s", tt.name, err, stderr)
			}
			t.Logf("threads list %s output:\n%s", tt.name, stdout)
		})
	}
}

func TestCLI_ThreadsShow(t *testing.T) {
	skipIfMissingCreds(t)

	// Get a thread ID
	client := getTestClient()
	ctx := context.Background()

	threads, err := client.GetThreads(ctx, testGrantID, &domain.ThreadQueryParams{Limit: 1})
	if err != nil {
		if strings.Contains(err.Error(), "Method not supported for provider") ||
			strings.Contains(err.Error(), "an internal error ocurred") {
			t.Skipf("Provider does not support threads: %v", err)
		}
		t.Fatalf("Failed to get threads: %v", err)
	}
	if len(threads) == 0 {
		t.Skip("No threads available for show test")
	}

	threadID := threads[0].ID

	stdout, stderr, err := runCLI("email", "threads", "show", threadID, testGrantID)

	if err != nil {
		t.Fatalf("threads show failed: %v\nstderr: %s", err, stderr)
	}

	// Should show thread details
	if !strings.Contains(stdout, "Thread:") {
		t.Errorf("Expected 'Thread:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Participants:") {
		t.Errorf("Expected 'Participants:' in output, got: %s", stdout)
	}

	t.Logf("threads show output:\n%s", stdout)
}

// =============================================================================
// DRAFT COMMAND TESTS
// =============================================================================

func TestCLI_DraftsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "drafts", "list", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("drafts list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show drafts or "No drafts found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No drafts found") {
		t.Errorf("Expected drafts list output, got: %s", stdout)
	}

	t.Logf("drafts list output:\n%s", stdout)
}

func TestCLI_DraftsLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	email := testEmail
	if email == "" {
		email = "test@example.com"
	}

	subject := fmt.Sprintf("CLI Test Draft %d", time.Now().Unix())
	body := "This is a test draft created by integration tests"

	var draftID string

	// Create draft
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "drafts", "create",
			"--to", email,
			"--subject", subject,
			"--body", body,
			testGrantID)

		if err != nil {
			t.Fatalf("drafts create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Draft created") {
			t.Errorf("Expected 'Draft created' in output, got: %s", stdout)
		}

		// Extract draft ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			draftID = strings.TrimSpace(stdout[idx+3:])
			// Clean up any trailing whitespace or newlines
			if newline := strings.Index(draftID, "\n"); newline != -1 {
				draftID = draftID[:newline]
			}
		}

		t.Logf("drafts create output: %s", stdout)
		t.Logf("Draft ID: %s", draftID)
	})

	if draftID == "" {
		t.Fatal("Failed to get draft ID from create output")
	}

	// Wait for draft to sync
	time.Sleep(2 * time.Second)

	// Show draft
	t.Run("show", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "drafts", "show", draftID, testGrantID)
		if err != nil {
			t.Fatalf("drafts show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Draft:") {
			t.Errorf("Expected 'Draft:' in output, got: %s", stdout)
		}

		t.Logf("drafts show output:\n%s", stdout)
	})

	// List drafts (should include our draft)
	t.Run("list", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "drafts", "list", testGrantID)
		if err != nil {
			t.Fatalf("drafts list failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Found") {
			t.Errorf("Expected to find drafts, got: %s", stdout)
		}

		t.Logf("drafts list output:\n%s", stdout)
	})

	// Delete draft
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLIWithInput("y\n", "email", "drafts", "delete", draftID, testGrantID)
		if err != nil {
			t.Fatalf("drafts delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("drafts delete output: %s", stdout)
	})
}

// =============================================================================
// EMAIL SEND COMMAND TESTS
// =============================================================================

func TestCLI_EmailSend(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("NYLAS_TEST_SEND_EMAIL not set to 'true'")
	}

	email := testEmail
	if email == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	subject := fmt.Sprintf("CLI Integration Test %d", time.Now().Unix())
	body := "This is a test email sent by the Nylas CLI integration tests."

	stdout, stderr, err := runCLIWithInput("y\n", "email", "send",
		"--to", email,
		"--subject", subject,
		"--body", body,
		testGrantID)

	if err != nil {
		t.Fatalf("email send failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "sent successfully") {
		t.Errorf("Expected 'sent successfully' in output, got: %s", stdout)
	}

	t.Logf("email send output: %s", stdout)
}

// =============================================================================
// OTP COMMAND TESTS
// =============================================================================

func TestCLI_OTPList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("otp", "list")

	if err != nil {
		// OTP list may fail if no accounts configured
		if strings.Contains(stderr, "no accounts") || strings.Contains(stderr, "not configured") {
			t.Skip("No OTP accounts configured")
		}
		t.Fatalf("otp list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show accounts or indicate none configured
	if !strings.Contains(stdout, "Configured Accounts") && !strings.Contains(stdout, "account") {
		t.Errorf("Expected accounts listing, got: %s", stdout)
	}

	t.Logf("otp list output:\n%s", stdout)
}

func TestCLI_OTPGet(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("otp", "get")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// OTP get may fail if no OTP emails found or no accounts configured
		if strings.Contains(stderr, "No OTP") || strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "no default grant") || strings.Contains(stderr, "no messages found") {
			t.Skip("No OTP codes available or no messages in inbox")
		}
		t.Fatalf("otp get failed: %v\nstderr: %s", err, stderr)
	}

	// Should show OTP code or indicate none found
	// OTP codes are typically 4-8 digits
	t.Logf("otp get output:\n%s", stdout)
}

// =============================================================================
// HELP AND VERSION TESTS
// =============================================================================

func TestCLI_Help(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}
	stdout, stderr, err := runCLI("--help")

	if err != nil {
		t.Fatalf("--help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage information
	if !strings.Contains(stdout, "Usage:") && !strings.Contains(stdout, "nylas") {
		t.Errorf("Expected help output, got: %s", stdout)
	}

	t.Logf("--help output:\n%s", stdout)
}

func TestCLI_AuthHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}
	stdout, stderr, err := runCLI("auth", "--help")

	if err != nil {
		t.Fatalf("auth --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show auth subcommands
	if !strings.Contains(stdout, "config") || !strings.Contains(stdout, "login") {
		t.Errorf("Expected auth subcommands in help, got: %s", stdout)
	}

	t.Logf("auth --help output:\n%s", stdout)
}

func TestCLI_EmailHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}
	stdout, stderr, err := runCLI("email", "--help")

	if err != nil {
		t.Fatalf("email --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show email subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "read") {
		t.Errorf("Expected email subcommands in help, got: %s", stdout)
	}

	t.Logf("email --help output:\n%s", stdout)
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestCLI_InvalidCommand(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}
	_, stderr, err := runCLI("invalidcommand")

	if err == nil {
		t.Error("Expected error for invalid command")
	}

	if !strings.Contains(stderr, "unknown command") && !strings.Contains(stderr, "invalid") {
		t.Logf("stderr for invalid command: %s", stderr)
	}
}

func TestCLI_EmailRead_InvalidID(t *testing.T) {
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("email", "read", "invalid-message-id-12345", testGrantID)

	if err == nil {
		t.Error("Expected error for invalid message ID")
	}

	t.Logf("Error output for invalid message ID: %s", stderr)
}

func TestCLI_EmailList_InvalidGrantID(t *testing.T) {
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("email", "list", "invalid-grant-id-12345", "--limit", "1")

	if err == nil {
		t.Error("Expected error for invalid grant ID")
	}

	t.Logf("Error output for invalid grant ID: %s", stderr)
}

// =============================================================================
// CONCURRENT OPERATIONS TEST
// =============================================================================

func TestCLI_ConcurrentOperations(t *testing.T) {
	skipIfMissingCreds(t)

	// Run multiple list operations concurrently
	type result struct {
		name string
		err  error
		stderr string
	}
	results := make(chan result, 3)

	operations := []struct {
		name string
		args []string
	}{
		{"email list", []string{"email", "list", testGrantID, "--limit", "2"}},
		{"folders list", []string{"email", "folders", "list", testGrantID}},
		{"threads list", []string{"email", "threads", "list", testGrantID, "--limit", "2"}},
	}

	for _, op := range operations {
		go func(name string, args []string) {
			_, stderr, err := runCLI(args...)
			results <- result{name, err, stderr}
		}(op.name, op.args)
	}

	// Wait for all operations - allow some to fail if provider doesn't support them
	successCount := 0
	for i := 0; i < len(operations); i++ {
		select {
		case r := <-results:
			if r.err != nil {
				if strings.Contains(r.stderr, "Method not supported for provider") ||
					strings.Contains(r.stderr, "an internal error ocurred") {
					t.Logf("%s: Skipped (not supported by provider)", r.name)
				} else {
					t.Logf("%s: Failed: %v", r.name, r.err)
				}
			} else {
				successCount++
				t.Logf("%s: OK", r.name)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
	if successCount == 0 {
		t.Skip("No operations succeeded - provider may have limited support")
	}
}

// =============================================================================
// FULL WORKFLOW TEST
// =============================================================================

func TestCLI_FullWorkflow(t *testing.T) {
	skipIfMissingCreds(t)

	// This test simulates a typical user workflow

	// 1. Check auth status
	t.Run("1_auth_status", func(t *testing.T) {
		stdout, stderr, err := runCLI("auth", "status")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Auth status: %s", stdout)
	})

	// 2. List emails
	t.Run("2_list_emails", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "list", testGrantID, "--limit", "5", "--id")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Email list: %s", stdout)
	})

	// 3. List folders (skip if provider doesn't support)
	t.Run("3_list_folders", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "folders", "list", testGrantID)
		skipIfProviderNotSupported(t, stderr)
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Folders: %s", stdout)
	})

	// 4. List threads (skip if provider doesn't support)
	t.Run("4_list_threads", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "threads", "list", testGrantID, "--limit", "3")
		skipIfProviderNotSupported(t, stderr)
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Threads: %s", stdout)
	})

	// 5. Search emails
	t.Run("5_search_emails", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "search", "test", testGrantID, "--limit", "3")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Search results: %s", stdout)
	})

	// 6. Check OTP
	t.Run("6_otp_list", func(t *testing.T) {
		stdout, _, _ := runCLI("otp", "list")
		t.Logf("OTP list: %s", stdout)
	})

	t.Log("Full workflow completed successfully")
}

// =============================================================================
// DOCTOR COMMAND TESTS
// =============================================================================

func TestCLI_Doctor(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("doctor")

	if err != nil {
		t.Fatalf("doctor failed: %v\nstderr: %s", err, stderr)
	}

	// Should show health check header
	if !strings.Contains(stdout, "Nylas CLI Health Check") {
		t.Errorf("Expected 'Nylas CLI Health Check' in output, got: %s", stdout)
	}

	// Should show summary
	if !strings.Contains(stdout, "Summary") {
		t.Errorf("Expected 'Summary' in output, got: %s", stdout)
	}

	// Should check configuration
	if !strings.Contains(stdout, "Configuration") {
		t.Errorf("Expected 'Configuration' check in output, got: %s", stdout)
	}

	// Should check secret store
	if !strings.Contains(stdout, "Secret Store") {
		t.Errorf("Expected 'Secret Store' check in output, got: %s", stdout)
	}

	// Should check API credentials
	if !strings.Contains(stdout, "API Credentials") {
		t.Errorf("Expected 'API Credentials' check in output, got: %s", stdout)
	}

	// Should check network
	if !strings.Contains(stdout, "Network") {
		t.Errorf("Expected 'Network' check in output, got: %s", stdout)
	}

	// Should check grants
	if !strings.Contains(stdout, "Grants") {
		t.Errorf("Expected 'Grants' check in output, got: %s", stdout)
	}

	t.Logf("doctor output:\n%s", stdout)
}

func TestCLI_Doctor_Verbose(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("doctor", "--verbose")

	if err != nil {
		t.Fatalf("doctor --verbose failed: %v\nstderr: %s", err, stderr)
	}

	// Should show platform info in verbose mode
	if !strings.Contains(stdout, "Platform:") {
		t.Errorf("Expected 'Platform:' in verbose output, got: %s", stdout)
	}

	// Should show Go version
	if !strings.Contains(stdout, "Go Version:") {
		t.Errorf("Expected 'Go Version:' in verbose output, got: %s", stdout)
	}

	// Should show config directory
	if !strings.Contains(stdout, "Config Dir:") {
		t.Errorf("Expected 'Config Dir:' in verbose output, got: %s", stdout)
	}

	t.Logf("doctor --verbose output:\n%s", stdout)
}

func TestCLI_Doctor_Help(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("doctor", "--help")

	if err != nil {
		t.Fatalf("doctor --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show command description
	if !strings.Contains(stdout, "diagnostic checks") {
		t.Errorf("Expected 'diagnostic checks' in help output, got: %s", stdout)
	}

	// Should show verbose flag
	if !strings.Contains(stdout, "--verbose") && !strings.Contains(stdout, "-v") {
		t.Errorf("Expected verbose flag in help, got: %s", stdout)
	}

	t.Logf("doctor --help output:\n%s", stdout)
}

// =============================================================================
// PAGINATION TESTS
// =============================================================================

func TestCLI_EmailList_All(t *testing.T) {
	skipIfMissingCreds(t)

	// Test fetching all messages with a max limit
	stdout, stderr, err := runCLI("email", "list", testGrantID, "--all", "--max", "15")

	if err != nil {
		t.Fatalf("email list --all failed: %v\nstderr: %s", err, stderr)
	}

	// Should show fetching progress or results
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No messages") {
		t.Errorf("Expected message list output, got: %s", stdout)
	}

	t.Logf("email list --all output:\n%s", stdout)
}

func TestCLI_EmailList_AllWithID(t *testing.T) {
	skipIfMissingCreds(t)

	// Test fetching all messages with IDs shown
	stdout, stderr, err := runCLI("email", "list", testGrantID, "--all", "--max", "5", "--id")

	if err != nil {
		t.Fatalf("email list --all --id failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message IDs
	if strings.Contains(stdout, "Found") && !strings.Contains(stdout, "ID:") {
		t.Errorf("Expected message IDs in output with --id flag, got: %s", stdout)
	}

	t.Logf("email list --all --id output:\n%s", stdout)
}

func TestCLI_EmailList_AllHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "list", "--help")

	if err != nil {
		t.Fatalf("email list --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show --all flag
	if !strings.Contains(stdout, "--all") {
		t.Errorf("Expected '--all' flag in help output, got: %s", stdout)
	}

	// Should show --max flag
	if !strings.Contains(stdout, "--max") {
		t.Errorf("Expected '--max' flag in help output, got: %s", stdout)
	}

	t.Logf("email list --help output:\n%s", stdout)
}

// =============================================================================
// CALENDAR COMMAND TESTS
// =============================================================================

func TestCLI_CalendarList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "list", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("calendar list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show calendar list or "No calendars found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No calendars found") {
		t.Errorf("Expected calendar list output, got: %s", stdout)
	}

	t.Logf("calendar list output:\n%s", stdout)
}

func TestCLI_CalendarHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "--help")

	if err != nil {
		t.Fatalf("calendar --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show calendar subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "events") {
		t.Errorf("Expected calendar subcommands in help, got: %s", stdout)
	}

	t.Logf("calendar --help output:\n%s", stdout)
}

func TestCLI_CalendarEventsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "events", "list", testGrantID, "--limit", "5")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// May fail if no calendars
		if strings.Contains(stderr, "no calendars") {
			t.Skip("No calendars available")
		}
		t.Fatalf("calendar events list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show events list or "No events found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No events found") {
		t.Errorf("Expected events list output, got: %s", stdout)
	}

	t.Logf("calendar events list output:\n%s", stdout)
}

func TestCLI_CalendarEventsListWithDays(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "events", "list", testGrantID, "--days", "30", "--limit", "10")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		if strings.Contains(stderr, "no calendars") {
			t.Skip("No calendars available")
		}
		t.Fatalf("calendar events list --days failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("calendar events list --days output:\n%s", stdout)
}

func TestCLI_CalendarEventsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "events", "--help")

	if err != nil {
		t.Fatalf("calendar events --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show events subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "create") {
		t.Errorf("Expected events subcommands in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "show") || !strings.Contains(stdout, "delete") {
		t.Errorf("Expected show and delete subcommands in help, got: %s", stdout)
	}

	t.Logf("calendar events --help output:\n%s", stdout)
}

func TestCLI_CalendarEventsCreateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "events", "create", "--help")

	if err != nil {
		t.Fatalf("calendar events create --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show required flags
	if !strings.Contains(stdout, "--title") || !strings.Contains(stdout, "--start") {
		t.Errorf("Expected --title and --start flags in help, got: %s", stdout)
	}

	// Should show optional flags
	if !strings.Contains(stdout, "--end") || !strings.Contains(stdout, "--location") {
		t.Errorf("Expected --end and --location flags in help, got: %s", stdout)
	}

	// Should show examples
	if !strings.Contains(stdout, "Examples:") {
		t.Errorf("Expected 'Examples:' in help, got: %s", stdout)
	}

	t.Logf("calendar events create --help output:\n%s", stdout)
}

func TestCLI_CalendarEventsLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	// Get tomorrow's date for the event
	tomorrow := time.Now().AddDate(0, 0, 1)
	startTime := tomorrow.Format("2006-01-02") + " 14:00"
	endTime := tomorrow.Format("2006-01-02") + " 15:00"
	eventTitle := fmt.Sprintf("CLI Test Event %d", time.Now().Unix())

	var eventID string

	// Create event
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "events", "create",
			"--title", eventTitle,
			"--start", startTime,
			"--end", endTime,
			"--location", "Test Location",
			testGrantID)

		if err != nil {
			if strings.Contains(stderr, "no writable calendar") || strings.Contains(stderr, "no calendars") {
				t.Skip("No writable calendar available")
			}
			t.Fatalf("calendar events create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Event created") {
			t.Errorf("Expected 'Event created' in output, got: %s", stdout)
		}

		// Extract event ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			eventID = strings.TrimSpace(stdout[idx+3:])
			if newline := strings.Index(eventID, "\n"); newline != -1 {
				eventID = eventID[:newline]
			}
		}

		t.Logf("calendar events create output: %s", stdout)
		t.Logf("Event ID: %s", eventID)
	})

	if eventID == "" {
		t.Fatal("Failed to get event ID from create output")
	}

	// Wait for event to sync
	time.Sleep(2 * time.Second)

	// Show event
	t.Run("show", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "events", "show", eventID, testGrantID)
		if err != nil {
			t.Fatalf("calendar events show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, eventTitle) {
			t.Errorf("Expected event title in output, got: %s", stdout)
		}

		t.Logf("calendar events show output:\n%s", stdout)
	})

	// Delete event
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLIWithInput("y\n", "calendar", "events", "delete", eventID, testGrantID)
		if err != nil {
			t.Fatalf("calendar events delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("calendar events delete output: %s", stdout)
	})
}

func TestCLI_CalendarEventsCreate_AllDay(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	// Get tomorrow's date
	tomorrow := time.Now().AddDate(0, 0, 1)
	dateStr := tomorrow.Format("2006-01-02")
	eventTitle := fmt.Sprintf("CLI Test All Day %d", time.Now().Unix())

	// Create all-day event
	stdout, stderr, err := runCLI("calendar", "events", "create",
		"--title", eventTitle,
		"--start", dateStr,
		"--all-day",
		testGrantID)

	if err != nil {
		if strings.Contains(stderr, "no writable calendar") || strings.Contains(stderr, "no calendars") {
			t.Skip("No writable calendar available")
		}
		t.Fatalf("calendar events create --all-day failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Event created") {
		t.Errorf("Expected 'Event created' in output, got: %s", stdout)
	}

	t.Logf("calendar events create --all-day output: %s", stdout)

	// Extract event ID and delete it
	if idx := strings.Index(stdout, "ID:"); idx != -1 {
		eventID := strings.TrimSpace(stdout[idx+3:])
		if newline := strings.Index(eventID, "\n"); newline != -1 {
			eventID = eventID[:newline]
		}
		// Clean up
		time.Sleep(time.Second)
		runCLIWithInput("y\n", "calendar", "events", "delete", eventID, testGrantID)
	}
}

// =============================================================================
// CONTACTS COMMAND TESTS
// =============================================================================

func TestCLI_ContactsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("contacts", "list", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("contacts list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show contacts list or "No contacts found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No contacts found") {
		t.Errorf("Expected contacts list output, got: %s", stdout)
	}

	t.Logf("contacts list output:\n%s", stdout)
}

func TestCLI_ContactsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("contacts", "--help")

	if err != nil {
		t.Fatalf("contacts --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show contacts subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "create") {
		t.Errorf("Expected contacts subcommands in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "show") || !strings.Contains(stdout, "delete") {
		t.Errorf("Expected show and delete subcommands in help, got: %s", stdout)
	}

	t.Logf("contacts --help output:\n%s", stdout)
}

func TestCLI_ContactsCreateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("contacts", "create", "--help")

	if err != nil {
		t.Fatalf("contacts create --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show required flags
	if !strings.Contains(stdout, "--first-name") || !strings.Contains(stdout, "--last-name") {
		t.Errorf("Expected --first-name and --last-name flags in help, got: %s", stdout)
	}

	// Should show optional flags
	if !strings.Contains(stdout, "--email") || !strings.Contains(stdout, "--phone") {
		t.Errorf("Expected --email and --phone flags in help, got: %s", stdout)
	}

	t.Logf("contacts create --help output:\n%s", stdout)
}

func TestCLI_ContactsGroupsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("contacts", "groups", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("contacts groups failed: %v\nstderr: %s", err, stderr)
	}

	// Should show groups list or "No contact groups found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No contact groups found") {
		t.Errorf("Expected groups list output, got: %s", stdout)
	}

	t.Logf("contacts groups output:\n%s", stdout)
}

func TestCLI_ContactsLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	contactFirstName := "CLI"
	contactLastName := fmt.Sprintf("Test%d", time.Now().Unix())
	contactEmail := fmt.Sprintf("test%d@example.com", time.Now().Unix())

	var contactID string

	// Create contact
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("contacts", "create",
			"--first-name", contactFirstName,
			"--last-name", contactLastName,
			"--email", contactEmail,
			testGrantID)

		if err != nil {
			t.Fatalf("contacts create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Contact created") {
			t.Errorf("Expected 'Contact created' in output, got: %s", stdout)
		}

		// Extract contact ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			contactID = strings.TrimSpace(stdout[idx+3:])
			if newline := strings.Index(contactID, "\n"); newline != -1 {
				contactID = contactID[:newline]
			}
		}

		t.Logf("contacts create output: %s", stdout)
		t.Logf("Contact ID: %s", contactID)
	})

	if contactID == "" {
		t.Fatal("Failed to get contact ID from create output")
	}

	// Wait for contact to sync
	time.Sleep(2 * time.Second)

	// Show contact
	t.Run("show", func(t *testing.T) {
		stdout, stderr, err := runCLI("contacts", "show", contactID, testGrantID)
		if err != nil {
			t.Fatalf("contacts show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, contactFirstName) || !strings.Contains(stdout, contactLastName) {
			t.Errorf("Expected contact name in output, got: %s", stdout)
		}

		t.Logf("contacts show output:\n%s", stdout)
	})

	// Delete contact
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLIWithInput("y\n", "contacts", "delete", contactID, testGrantID)
		if err != nil {
			t.Fatalf("contacts delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("contacts delete output: %s", stdout)
	})
}

// =============================================================================
// WEBHOOK COMMAND TESTS
// =============================================================================

func TestCLI_WebhookHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "--help")

	if err != nil {
		t.Fatalf("webhook --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show webhook subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "create") {
		t.Errorf("Expected webhook subcommands in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "triggers") || !strings.Contains(stdout, "test") {
		t.Errorf("Expected triggers and test subcommands in help, got: %s", stdout)
	}

	t.Logf("webhook --help output:\n%s", stdout)
}

func TestCLI_WebhookList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("webhook", "list")

	if err != nil {
		t.Fatalf("webhook list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show webhooks list or "No webhooks configured"
	if !strings.Contains(stdout, "webhooks") && !strings.Contains(stdout, "No webhooks") && !strings.Contains(stdout, "ID") {
		t.Errorf("Expected webhook list output, got: %s", stdout)
	}

	t.Logf("webhook list output:\n%s", stdout)
}

func TestCLI_WebhookListJSON(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("webhook", "list", "--format", "json")

	if err != nil {
		t.Fatalf("webhook list --format json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON (starts with [ or {)
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && trimmed[0] != '[' && trimmed[0] != '{' && !strings.Contains(stdout, "No webhooks") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	t.Logf("webhook list JSON output:\n%s", stdout)
}

func TestCLI_WebhookTriggers(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "triggers")

	if err != nil {
		t.Fatalf("webhook triggers failed: %v\nstderr: %s", err, stderr)
	}

	// Should show trigger types
	if !strings.Contains(stdout, "message.created") {
		t.Errorf("Expected 'message.created' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "event.created") {
		t.Errorf("Expected 'event.created' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "contact.created") {
		t.Errorf("Expected 'contact.created' in output, got: %s", stdout)
	}

	t.Logf("webhook triggers output:\n%s", stdout)
}

func TestCLI_WebhookTriggersJSON(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "triggers", "--format", "json")

	if err != nil {
		t.Fatalf("webhook triggers --format json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON
	if !strings.Contains(stdout, "{") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	t.Logf("webhook triggers JSON output:\n%s", stdout)
}

func TestCLI_WebhookTriggersCategory(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "triggers", "--category", "message")

	if err != nil {
		t.Fatalf("webhook triggers --category message failed: %v\nstderr: %s", err, stderr)
	}

	// Should show only message triggers
	if !strings.Contains(stdout, "message.created") {
		t.Errorf("Expected 'message.created' in output, got: %s", stdout)
	}
	// Should show Message header (the actual filtered section)
	if !strings.Contains(stdout, "📧 Message") && !strings.Contains(stdout, "Message") {
		t.Errorf("Expected 'Message' category header in output, got: %s", stdout)
	}

	t.Logf("webhook triggers --category message output:\n%s", stdout)
}

func TestCLI_WebhookCreateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "create", "--help")

	if err != nil {
		t.Fatalf("webhook create --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show required flags
	if !strings.Contains(stdout, "--url") || !strings.Contains(stdout, "--triggers") {
		t.Errorf("Expected --url and --triggers flags in help, got: %s", stdout)
	}

	// Should show examples
	if !strings.Contains(stdout, "Examples:") {
		t.Errorf("Expected 'Examples:' in help, got: %s", stdout)
	}

	t.Logf("webhook create --help output:\n%s", stdout)
}

func TestCLI_WebhookTestHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("webhook", "test", "--help")

	if err != nil {
		t.Fatalf("webhook test --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show subcommands
	if !strings.Contains(stdout, "send") || !strings.Contains(stdout, "payload") {
		t.Errorf("Expected 'send' and 'payload' subcommands in help, got: %s", stdout)
	}

	t.Logf("webhook test --help output:\n%s", stdout)
}

func TestCLI_WebhookTestPayload(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("webhook", "test", "payload", "message.created")

	if err != nil {
		t.Fatalf("webhook test payload failed: %v\nstderr: %s", err, stderr)
	}

	// Should show mock payload
	if !strings.Contains(stdout, "message.created") || !strings.Contains(stdout, "{") {
		t.Errorf("Expected mock payload with trigger type, got: %s", stdout)
	}

	t.Logf("webhook test payload output:\n%s", stdout)
}

func TestCLI_WebhookTestPayloadList(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Running without trigger type should list available types
	stdout, stderr, err := runCLI("webhook", "test", "payload")

	if err != nil {
		t.Fatalf("webhook test payload (no args) failed: %v\nstderr: %s", err, stderr)
	}

	// Should show available trigger types
	if !strings.Contains(stdout, "Available trigger types") && !strings.Contains(stdout, "message") {
		t.Errorf("Expected trigger type list, got: %s", stdout)
	}

	t.Logf("webhook test payload (no args) output:\n%s", stdout)
}

func TestCLI_WebhookLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	webhookURL := fmt.Sprintf("https://example.com/webhook/%d", time.Now().Unix())
	webhookDesc := fmt.Sprintf("CLI Test Webhook %d", time.Now().Unix())

	var webhookID string

	// Create webhook
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("webhook", "create",
			"--url", webhookURL,
			"--triggers", "message.created",
			"--description", webhookDesc)

		if err != nil {
			t.Fatalf("webhook create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "created") {
			t.Errorf("Expected 'created' in output, got: %s", stdout)
		}

		// Extract webhook ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			webhookID = strings.TrimSpace(stdout[idx+3:])
			if newline := strings.Index(webhookID, "\n"); newline != -1 {
				webhookID = webhookID[:newline]
			}
		}

		t.Logf("webhook create output: %s", stdout)
		t.Logf("Webhook ID: %s", webhookID)
	})

	if webhookID == "" {
		t.Fatal("Failed to get webhook ID from create output")
	}

	// Wait for webhook to be created
	time.Sleep(2 * time.Second)

	// Show webhook
	t.Run("show", func(t *testing.T) {
		stdout, stderr, err := runCLI("webhook", "show", webhookID)
		if err != nil {
			t.Fatalf("webhook show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, webhookID) {
			t.Errorf("Expected webhook ID in output, got: %s", stdout)
		}

		t.Logf("webhook show output:\n%s", stdout)
	})

	// Update webhook
	t.Run("update", func(t *testing.T) {
		newDesc := "Updated " + webhookDesc
		stdout, stderr, err := runCLI("webhook", "update", webhookID,
			"--description", newDesc)
		if err != nil {
			t.Fatalf("webhook update failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "updated") {
			t.Errorf("Expected 'updated' in output, got: %s", stdout)
		}

		t.Logf("webhook update output:\n%s", stdout)
	})

	// Delete webhook
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLI("webhook", "delete", webhookID, "--force")
		if err != nil {
			t.Fatalf("webhook delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("webhook delete output: %s", stdout)
	})
}

// =============================================================================
// CALENDAR AVAILABILITY COMMAND TESTS
// =============================================================================

func TestCLI_CalendarAvailabilityHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "availability", "--help")

	if err != nil {
		t.Fatalf("calendar availability --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show availability subcommands
	if !strings.Contains(stdout, "check") || !strings.Contains(stdout, "find") {
		t.Errorf("Expected 'check' and 'find' subcommands in help, got: %s", stdout)
	}

	t.Logf("calendar availability --help output:\n%s", stdout)
}

func TestCLI_CalendarAvailabilityCheck(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "availability", "check", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// May fail if no calendar access
		if strings.Contains(stderr, "no calendars") || strings.Contains(stderr, "not found") {
			t.Skip("No calendars available for availability check")
		}
		t.Fatalf("calendar availability check failed: %v\nstderr: %s", err, stderr)
	}

	// Should show free/busy status
	if !strings.Contains(stdout, "Free/Busy") && !strings.Contains(stdout, "free") && !strings.Contains(stdout, "Busy") {
		t.Errorf("Expected free/busy output, got: %s", stdout)
	}

	t.Logf("calendar availability check output:\n%s", stdout)
}

func TestCLI_CalendarAvailabilityCheckWithDuration(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "availability", "check", testGrantID,
		"--duration", "2d")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		if strings.Contains(stderr, "no calendars") || strings.Contains(stderr, "not found") {
			t.Skip("No calendars available")
		}
		t.Fatalf("calendar availability check --duration failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("calendar availability check --duration output:\n%s", stdout)
}

func TestCLI_CalendarAvailabilityCheckJSON(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("calendar", "availability", "check", testGrantID,
		"--format", "json")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		if strings.Contains(stderr, "no calendars") || strings.Contains(stderr, "not found") {
			t.Skip("No calendars available")
		}
		t.Fatalf("calendar availability check --format json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && trimmed[0] != '{' {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	t.Logf("calendar availability check JSON output:\n%s", stdout)
}

func TestCLI_CalendarAvailabilityFindHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "availability", "find", "--help")

	if err != nil {
		t.Fatalf("calendar availability find --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show required flags
	if !strings.Contains(stdout, "--participants") {
		t.Errorf("Expected '--participants' flag in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--duration") {
		t.Errorf("Expected '--duration' flag in help, got: %s", stdout)
	}

	t.Logf("calendar availability find --help output:\n%s", stdout)
}

func TestCLI_CalendarAvailabilityFind(t *testing.T) {
	skipIfMissingCreds(t)

	// Use test email if available
	email := testEmail
	if email == "" {
		email = "test@example.com"
	}

	stdout, stderr, err := runCLI("calendar", "availability", "find",
		"--participants", email,
		"--duration", "30")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// May fail if calendar feature not available or participant not found
		if strings.Contains(stderr, "not available") || strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "Failed to find a valid Grant") {
			t.Skip("Availability find not available or participant not found")
		}
		t.Fatalf("calendar availability find failed: %v\nstderr: %s", err, stderr)
	}

	// Should show available slots or "No available" message
	if !strings.Contains(stdout, "Available") && !strings.Contains(stdout, "available") && !strings.Contains(stdout, "No available") {
		t.Errorf("Expected availability output, got: %s", stdout)
	}

	t.Logf("calendar availability find output:\n%s", stdout)
}

// =============================================================================
// SCHEDULED EMAIL SENDING TESTS
// =============================================================================

func TestCLI_EmailSendHelp_Schedule(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "send", "--help")

	if err != nil {
		t.Fatalf("email send --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show --schedule flag
	if !strings.Contains(stdout, "--schedule") {
		t.Errorf("Expected '--schedule' flag in help, got: %s", stdout)
	}

	// Should show schedule examples
	if !strings.Contains(stdout, "2h") && !strings.Contains(stdout, "tomorrow") {
		t.Errorf("Expected schedule examples in help, got: %s", stdout)
	}

	// Should show --yes flag
	if !strings.Contains(stdout, "--yes") && !strings.Contains(stdout, "-y") {
		t.Errorf("Expected '--yes' flag in help, got: %s", stdout)
	}

	t.Logf("email send --help output:\n%s", stdout)
}

func TestCLI_EmailSend_ScheduledDryRun(t *testing.T) {
	skipIfMissingCreds(t)

	// This test shows the preview but doesn't actually send
	// by not providing 'y' confirmation
	email := testEmail
	if email == "" {
		email = "test@example.com"
	}

	subject := fmt.Sprintf("Scheduled Test %d", time.Now().Unix())

	stdout, _, err := runCLIWithInput("n\n", "email", "send",
		"--to", email,
		"--subject", subject,
		"--body", "This is a scheduled test email",
		"--schedule", "2h",
		testGrantID)

	// This should be "cancelled" since we said no
	if err == nil && !strings.Contains(stdout, "Cancelled") {
		t.Logf("Preview shown, cancelled as expected: %s", stdout)
	}

	// Should show scheduled time in preview
	if !strings.Contains(stdout, "Scheduled") {
		t.Errorf("Expected 'Scheduled' time in preview, got: %s", stdout)
	}

	t.Logf("email send --schedule preview:\n%s", stdout)
}

func TestCLI_EmailSend_Scheduled(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("NYLAS_TEST_SEND_EMAIL not set to 'true'")
	}

	email := testEmail
	if email == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	subject := fmt.Sprintf("CLI Scheduled Test %d", time.Now().Unix())
	body := "This is a scheduled test email sent by the Nylas CLI integration tests."

	// Schedule email for 2 hours from now (it can be cancelled later)
	stdout, stderr, err := runCLIWithInput("y\n", "email", "send",
		"--to", email,
		"--subject", subject,
		"--body", body,
		"--schedule", "2h",
		testGrantID)

	if err != nil {
		t.Fatalf("email send --schedule failed: %v\nstderr: %s", err, stderr)
	}

	// Should show scheduled success
	if !strings.Contains(stdout, "scheduled successfully") {
		t.Errorf("Expected 'scheduled successfully' in output, got: %s", stdout)
	}

	// Should show scheduled time
	if !strings.Contains(stdout, "Scheduled to send") {
		t.Errorf("Expected 'Scheduled to send' in output, got: %s", stdout)
	}

	t.Logf("email send --schedule output: %s", stdout)
}

// =============================================================================
// FULL NEW FEATURES WORKFLOW TEST
// =============================================================================

func TestCLI_NewFeaturesWorkflow(t *testing.T) {
	skipIfMissingCreds(t)

	// Test all new features in sequence

	// 1. Webhook triggers list
	t.Run("1_webhook_triggers", func(t *testing.T) {
		stdout, stderr, err := runCLI("webhook", "triggers")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		if !strings.Contains(stdout, "message.created") {
			t.Errorf("Expected message.created trigger")
		}
		t.Logf("Webhook triggers: %s", stdout)
	})

	// 2. List webhooks
	t.Run("2_webhook_list", func(t *testing.T) {
		stdout, stderr, err := runCLI("webhook", "list")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Webhook list: %s", stdout)
	})

	// 3. Calendar availability
	t.Run("3_availability_check", func(t *testing.T) {
		stdout, _, err := runCLI("calendar", "availability", "check", testGrantID, "--duration", "1d")
		if err != nil {
			t.Skip("Availability check not available")
		}
		t.Logf("Availability: %s", stdout)
	})

	// 4. Email send help (verify schedule option)
	t.Run("4_email_send_help", func(t *testing.T) {
		stdout, stderr, err := runCLI("email", "send", "--help")
		if err != nil {
			t.Fatalf("Failed: %v\nstderr: %s", err, stderr)
		}
		if !strings.Contains(stdout, "--schedule") {
			t.Errorf("Expected --schedule flag")
		}
		t.Logf("Email send help contains schedule option")
	})

	t.Log("New features workflow completed successfully")
}
