//go:build integration

package integration

import (
	"strings"
	"testing"
)

// =============================================================================
// OTP WATCH COMMAND TESTS
// =============================================================================

func TestCLI_OTPWatchHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("otp", "watch", "--help")

	if err != nil {
		t.Fatalf("otp watch --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show watch command description
	if !strings.Contains(stdout, "watch") && !strings.Contains(stdout, "Watch") {
		t.Errorf("Expected 'watch' in help output, got: %s", stdout)
	}

	// Should show interval flag
	if !strings.Contains(stdout, "--interval") {
		t.Errorf("Expected '--interval' flag in help, got: %s", stdout)
	}

	t.Logf("otp watch --help output:\n%s", stdout)
}

func TestCLI_OTPWatch_InvalidInterval(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	// Test with invalid interval
	_, stderr, err := runCLI("otp", "watch", "--interval", "-1")

	if err == nil {
		t.Error("Expected error for negative interval, but command succeeded")
	}

	t.Logf("otp watch invalid interval error: %s", stderr)
}

func TestCLI_OTPWatch(t *testing.T) {
	skipIfMissingCreds(t)

	t.Skip("otp watch is a long-running command that requires manual testing. " +
		"Automated testing would require: (1) background process management, " +
		"(2) sending test OTP emails, (3) process cleanup. " +
		"Help and error handling tests provide sufficient coverage.")

	// NOTE: To manually test otp watch:
	// 1. Run: nylas otp watch --interval 5
	// 2. Send OTP email to configured account
	// 3. Verify OTP is detected and displayed
	// 4. Press Ctrl+C to stop
}

// =============================================================================
// OTP MESSAGES COMMAND TESTS
// =============================================================================

func TestCLI_OTPMessagesHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("otp", "messages", "--help")

	if err != nil {
		t.Fatalf("otp messages --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show messages command description
	if !strings.Contains(stdout, "messages") && !strings.Contains(stdout, "Messages") {
		t.Errorf("Expected 'messages' in help output, got: %s", stdout)
	}

	// Should show limit flag
	if !strings.Contains(stdout, "--limit") {
		t.Errorf("Expected '--limit' flag in help, got: %s", stdout)
	}

	t.Logf("otp messages --help output:\n%s", stdout)
}

func TestCLI_OTPMessages(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	stdout, stderr, err := runCLI("otp", "messages", testGrantID, "--limit", "5")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// Messages may fail if no OTP messages found
		if strings.Contains(stderr, "no messages") || strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "no default grant") {
			t.Skip("No OTP messages available")
		}
		t.Fatalf("otp messages failed: %v\nstderr: %s", err, stderr)
	}

	// Should show messages or indicate none found
	if !strings.Contains(stdout, "OTP") && !strings.Contains(stdout, "messages") &&
		!strings.Contains(stdout, "No messages") && !strings.Contains(stdout, "Found") {
		t.Errorf("Expected OTP messages output, got: %s", stdout)
	}

	t.Logf("otp messages output:\n%s", stdout)
}

func TestCLI_OTPMessagesWithLimit(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	stdout, stderr, err := runCLI("otp", "messages", testGrantID, "--limit", "3")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		if strings.Contains(stderr, "no messages") || strings.Contains(stderr, "not found") {
			t.Skip("No OTP messages available")
		}
		t.Fatalf("otp messages --limit failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("otp messages --limit output:\n%s", stdout)
}

func TestCLI_OTPMessagesJSON(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	stdout, stderr, err := runCLI("otp", "messages", testGrantID, "--limit", "3", "--json")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		if strings.Contains(stderr, "no messages") || strings.Contains(stderr, "not found") {
			t.Skip("No OTP messages available")
		}
		t.Fatalf("otp messages --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON (starts with [ or {) or show "No messages found"
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && trimmed[0] != '[' && trimmed[0] != '{' &&
		!strings.Contains(stdout, "No messages") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	t.Logf("otp messages --json output:\n%s", stdout)
}

// =============================================================================
// OTP GET COMMAND ENHANCED TESTS
// =============================================================================

func TestCLI_OTPGetHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("otp", "get", "--help")

	if err != nil {
		t.Fatalf("otp get --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show get command description
	if !strings.Contains(stdout, "get") && !strings.Contains(stdout, "Get") {
		t.Errorf("Expected 'get' in help output, got: %s", stdout)
	}

	t.Logf("otp get --help output:\n%s", stdout)
}

func TestCLI_OTPGetWithGrantID(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	stdout, stderr, err := runCLI("otp", "get", testGrantID)
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		// OTP get may fail if no OTP codes found
		if strings.Contains(stderr, "No OTP") || strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "no messages found") || strings.Contains(stderr, "no OTP found") {
			t.Skip("No OTP codes available")
		}
		t.Fatalf("otp get with grant ID failed: %v\nstderr: %s", err, stderr)
	}

	// Should show OTP code or indicate none found
	t.Logf("otp get with grant ID output:\n%s", stdout)
}

// =============================================================================
// OTP LIST COMMAND ENHANCED TESTS
// =============================================================================

func TestCLI_OTPListHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("otp", "list", "--help")

	if err != nil {
		t.Fatalf("otp list --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show list command description
	if !strings.Contains(stdout, "list") && !strings.Contains(stdout, "List") {
		t.Errorf("Expected 'list' in help output, got: %s", stdout)
	}

	t.Logf("otp list --help output:\n%s", stdout)
}

// NOTE: TestCLI_OTPList already exists in misc_test.go

func TestCLI_OTPListJSON(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfKeyringDisabled(t)

	stdout, stderr, err := runCLI("otp", "list", "--json")

	if err != nil {
		// OTP list may fail if no accounts configured
		if strings.Contains(stderr, "no accounts") || strings.Contains(stderr, "not configured") {
			t.Skip("No OTP accounts configured")
		}
		t.Fatalf("otp list --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON (starts with [ or {) or show message
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && trimmed[0] != '[' && trimmed[0] != '{' &&
		!strings.Contains(stdout, "accounts") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	t.Logf("otp list --json output:\n%s", stdout)
}

// =============================================================================
// OTP HELP COMMAND TESTS
// =============================================================================

func TestCLI_OTPHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("otp", "--help")

	if err != nil {
		t.Fatalf("otp --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show OTP subcommands
	if !strings.Contains(stdout, "list") {
		t.Errorf("Expected 'list' subcommand in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "get") {
		t.Errorf("Expected 'get' subcommand in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "watch") {
		t.Errorf("Expected 'watch' subcommand in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "messages") {
		t.Errorf("Expected 'messages' subcommand in help, got: %s", stdout)
	}

	t.Logf("otp --help output:\n%s", stdout)
}
