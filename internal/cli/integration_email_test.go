//go:build integration

package cli

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

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

func TestCLI_EmailShow(t *testing.T) {
	skipIfMissingCreds(t)

	// Test the 'show' alias for 'read' command
	client := getTestClient()
	ctx := context.Background()

	messages, err := client.GetMessages(ctx, testGrantID, 1)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	if len(messages) == 0 {
		t.Skip("No messages available for show test")
	}

	messageID := messages[0].ID

	// Use 'show' alias instead of 'read'
	stdout, stderr, err := runCLI("email", "show", messageID, testGrantID)

	if err != nil {
		t.Fatalf("email show (alias) failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message details (same output as 'read')
	if !strings.Contains(stdout, "Subject:") {
		t.Errorf("Expected 'Subject:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "From:") {
		t.Errorf("Expected 'From:' in output, got: %s", stdout)
	}

	t.Logf("email show (alias) output:\n%s", stdout)
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
// EMAIL SEND COMMAND TESTS
// =============================================================================

func TestCLI_EmailSend(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}
	skipIfMissingCreds(t)

	email := testEmail
	if email == "" {
		email = "test@example.com"
	}

	stdout, stderr, err := runCLI("email", "send",
		"--to", email,
		"--subject", "CLI Integration Test",
		"--body", "This is a test email from the CLI integration tests.",
		testGrantID)

	if err != nil {
		t.Fatalf("email send failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "sent") && !strings.Contains(stdout, "Message") {
		t.Errorf("Expected send confirmation in output, got: %s", stdout)
	}

	t.Logf("email send output:\n%s", stdout)
}

func TestCLI_EmailHelp(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "--help")

	if err != nil {
		t.Fatalf("email --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show email subcommands
	if !strings.Contains(stdout, "list") {
		t.Errorf("Expected 'list' in email help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "read") {
		t.Errorf("Expected 'read' in email help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "send") {
		t.Errorf("Expected 'send' in email help, got: %s", stdout)
	}

	t.Logf("email help output:\n%s", stdout)
}

func TestCLI_EmailRead_InvalidID(t *testing.T) {
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("email", "read", "invalid-message-id", testGrantID)

	if err == nil {
		t.Error("Expected error for invalid message ID, but command succeeded")
	}

	t.Logf("email read invalid ID error: %s", stderr)
}

func TestCLI_EmailList_InvalidGrantID(t *testing.T) {
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("email", "list", "invalid-grant-id", "--limit", "1")

	if err == nil {
		t.Error("Expected error for invalid grant ID, but command succeeded")
	}

	t.Logf("email list invalid grant error: %s", stderr)
}

// =============================================================================
// EMAIL LIST ALL COMMAND TESTS
// =============================================================================

func TestCLI_EmailList_All(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "list", "all", "--limit", "5")

	if err != nil {
		// Skip if auth fails for "all" command (requires different auth setup)
		if strings.Contains(stderr, "Bearer token invalid") || strings.Contains(stderr, "unauthorized") {
			t.Skip("email list all requires different auth setup")
		}
		t.Fatalf("email list all failed: %v\nstderr: %s", err, stderr)
	}

	// Should show message count or "No messages found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No messages found") {
		t.Errorf("Expected message list output, got: %s", stdout)
	}

	t.Logf("email list all output:\n%s", stdout)
}

func TestCLI_EmailList_AllWithID(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "list", "all", "--limit", "3", "--id")

	if err != nil {
		// Skip if auth fails for "all" command (requires different auth setup)
		if strings.Contains(stderr, "Bearer token invalid") || strings.Contains(stderr, "unauthorized") {
			t.Skip("email list all requires different auth setup")
		}
		t.Fatalf("email list all --id failed: %v\nstderr: %s", err, stderr)
	}

	// Should show "ID:" lines when --id flag is used (if messages exist)
	if strings.Contains(stdout, "Found") && !strings.Contains(stdout, "ID:") {
		t.Errorf("Expected message IDs in output with --id flag, got: %s", stdout)
	}

	t.Logf("email list all --id output:\n%s", stdout)
}

func TestCLI_EmailList_AllHelp(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "list", "all", "--help")

	if err != nil {
		t.Fatalf("email list all --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show help for the all subcommand
	if !strings.Contains(stdout, "all") {
		t.Errorf("Expected 'all' in help output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--limit") {
		t.Errorf("Expected '--limit' flag in help output, got: %s", stdout)
	}

	t.Logf("email list all help output:\n%s", stdout)
}

// =============================================================================
// SCHEDULED SEND TESTS
// =============================================================================

func TestCLI_EmailSendHelp_Schedule(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "send", "--help")

	if err != nil {
		t.Fatalf("email send --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show schedule options in help
	if !strings.Contains(stdout, "--schedule") {
		t.Errorf("Expected '--schedule' in send help, got: %s", stdout)
	}

	t.Logf("email send help output:\n%s", stdout)
}

func TestCLI_EmailSend_ScheduleFlag(t *testing.T) {
	// Test that schedule flag is recognized (without actually sending)
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Just verify the flag is accepted by checking help
	stdout, stderr, err := runCLI("email", "send", "--help")

	if err != nil {
		t.Fatalf("email send --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show schedule flag with duration examples
	if !strings.Contains(stdout, "2h") && !strings.Contains(stdout, "tomorrow") {
		t.Errorf("Expected schedule duration examples in help, got: %s", stdout)
	}

	t.Logf("email send help shows schedule options")
}

func TestCLI_EmailSend_Scheduled(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}
	skipIfMissingCreds(t)

	email := testEmail
	if email == "" {
		email = "test@example.com"
	}

	// Schedule for 1 hour from now using duration format
	stdout, stderr, err := runCLI("email", "send",
		"--to", email,
		"--subject", "Scheduled Email Test",
		"--body", "This is a scheduled test email from CLI integration tests.",
		"--schedule", "1h",
		testGrantID)

	if err != nil {
		t.Fatalf("email send scheduled failed: %v\nstderr: %s", err, stderr)
	}

	// Should show scheduled confirmation
	if !strings.Contains(stdout, "scheduled") && !strings.Contains(stdout, "Scheduled") && !strings.Contains(stdout, "Message") {
		t.Errorf("Expected scheduled confirmation in output, got: %s", stdout)
	}

	t.Logf("email send scheduled output:\n%s", stdout)
}

// =============================================================================
// ADVANCED SEARCH COMMAND TESTS (Phase 3)
// =============================================================================

func TestCLI_EmailSearchHelp_AdvancedFlags(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "search", "--help")

	if err != nil {
		t.Fatalf("email search --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show advanced search flags
	if !strings.Contains(stdout, "--unread") {
		t.Errorf("Expected '--unread' flag in search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--starred") {
		t.Errorf("Expected '--starred' flag in search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--in") {
		t.Errorf("Expected '--in' flag in search help, got: %s", stdout)
	}

	t.Logf("email search help output:\n%s", stdout)
}

func TestCLI_EmailSearch_AdvancedFilters(t *testing.T) {
	skipIfMissingCreds(t)

	tests := []struct {
		name string
		args []string
	}{
		{"unread", []string{"email", "search", "test", testGrantID, "--unread", "--limit", "3"}},
		{"starred", []string{"email", "search", "test", testGrantID, "--starred", "--limit", "3"}},
		{"folder", []string{"email", "search", "test", testGrantID, "--in", "INBOX", "--limit", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)
			if err != nil {
				t.Fatalf("email search %s failed: %v\nstderr: %s", tt.name, err, stderr)
			}
			t.Logf("email search %s output:\n%s", tt.name, stdout)
		})
	}
}

// =============================================================================
// THREAD SEARCH COMMAND TESTS (Phase 3)
// =============================================================================

func TestCLI_ThreadsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "threads", "--help")

	if err != nil {
		t.Fatalf("email threads --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show threads subcommands including search
	if !strings.Contains(stdout, "search") {
		t.Errorf("Expected 'search' subcommand in threads help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "list") {
		t.Errorf("Expected 'list' subcommand in threads help, got: %s", stdout)
	}

	t.Logf("email threads help output:\n%s", stdout)
}

func TestCLI_ThreadsSearchHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "threads", "search", "--help")

	if err != nil {
		t.Fatalf("email threads search --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show search flags
	if !strings.Contains(stdout, "--from") {
		t.Errorf("Expected '--from' flag in threads search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--to") {
		t.Errorf("Expected '--to' flag in threads search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--subject") {
		t.Errorf("Expected '--subject' flag in threads search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--has-attachment") {
		t.Errorf("Expected '--has-attachment' flag in threads search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--unread") {
		t.Errorf("Expected '--unread' flag in threads search help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--starred") {
		t.Errorf("Expected '--starred' flag in threads search help, got: %s", stdout)
	}

	t.Logf("email threads search help output:\n%s", stdout)
}

func TestCLI_ThreadsSearch(t *testing.T) {
	skipIfMissingCreds(t)

	// Thread search uses filters (no full-text query), so we search by subject
	stdout, stderr, err := runCLI("email", "threads", "search", testGrantID, "--subject", "test", "--limit", "3")

	if err != nil {
		t.Fatalf("email threads search failed: %v\nstderr: %s", err, stderr)
	}

	// Should show results or "No threads found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No threads found") {
		t.Errorf("Expected search results output, got: %s", stdout)
	}

	t.Logf("email threads search output:\n%s", stdout)
}

func TestCLI_ThreadsSearch_WithFilters(t *testing.T) {
	skipIfMissingCreds(t)

	tests := []struct {
		name string
		args []string
	}{
		{"with-from", []string{"email", "threads", "search", testGrantID, "--from", "test@example.com", "--limit", "3"}},
		{"with-subject", []string{"email", "threads", "search", testGrantID, "--subject", "test", "--limit", "3"}},
		{"unread", []string{"email", "threads", "search", testGrantID, "--unread", "--limit", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)
			if err != nil {
				t.Fatalf("threads search %s failed: %v\nstderr: %s", tt.name, err, stderr)
			}
			t.Logf("threads search %s output:\n%s", tt.name, stdout)
		})
	}
}
