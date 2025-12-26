//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

// TestCLI_EmailAI_AnalyzeHelp tests the email ai analyze help command.
func TestCLI_EmailAI_AnalyzeHelp(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "ai", "analyze", "--help")
	if err != nil {
		t.Fatalf("email ai analyze --help failed: %v\nstderr: %s", err, stderr)
	}

	// Check for expected help content
	expectedStrings := []string{
		"Analyze recent emails",
		"--limit",
		"--provider",
		"--unread",
		"--folder",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}

	t.Logf("email ai analyze --help output:\n%s", stdout)
}

// TestCLI_EmailAI_Help tests the email ai command group help.
func TestCLI_EmailAI_Help(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "ai", "--help")
	if err != nil {
		t.Fatalf("email ai --help failed: %v\nstderr: %s", err, stderr)
	}

	// Check for expected help content
	expectedStrings := []string{
		"AI-powered email intelligence",
		"analyze",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}

	t.Logf("email ai --help output:\n%s", stdout)
}

// TestCLI_EmailAI_AnalyzeBasic tests basic email ai analyze functionality.
func TestCLI_EmailAI_AnalyzeBasic(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfNoDefaultAIProvider(t)
	t.Parallel()

	// Run with rate limiting since it makes API calls
	stdout, stderr, err := runCLIWithRateLimit(t, "email", "ai", "analyze", "--limit", "5")

	if err != nil {
		// Check if it's an AI configuration error (acceptable)
		if strings.Contains(stderr, "AI is not configured") ||
			strings.Contains(stderr, "not configured") {
			t.Skip("AI not configured, skipping test")
		}
		t.Fatalf("email ai analyze failed: %v\nstderr: %s", err, stderr)
	}

	// Check for expected output patterns
	expectedPatterns := []string{
		"Fetching",
		"emails",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(stdout, pattern) {
			t.Errorf("Expected output to contain %q\nGot: %s", pattern, stdout)
		}
	}

	t.Logf("email ai analyze output:\n%s", stdout)
}

// TestCLI_EmailAI_AnalyzeWithProvider tests email ai analyze with specific provider.
func TestCLI_EmailAI_AnalyzeWithProvider(t *testing.T) {
	skipIfMissingCreds(t)
	t.Parallel()

	// Test with ollama if available
	if !checkOllamaAvailable() {
		t.Skip("Ollama not available, skipping provider test")
	}

	stdout, stderr, err := runCLIWithRateLimit(t, "email", "ai", "analyze", "--limit", "3", "--provider", "ollama")

	if err != nil {
		if strings.Contains(stderr, "AI is not configured") ||
			strings.Contains(stderr, "not configured") ||
			strings.Contains(stderr, "provider") ||
			strings.Contains(stderr, "no default grant") ||
			strings.Contains(stderr, "no grant ID") {
			t.Skip("AI/Ollama not configured or no default grant, skipping test")
		}
		t.Fatalf("email ai analyze --provider ollama failed: %v\nstderr: %s", err, stderr)
	}

	// Check that ollama was used
	if !strings.Contains(stdout, "ollama") && !strings.Contains(stdout, "Provider:") {
		t.Logf("Note: Could not verify ollama was used. Output: %s", stdout)
	}

	t.Logf("email ai analyze --provider ollama output:\n%s", stdout)
}

// TestCLI_EmailAI_AnalyzeUnread tests email ai analyze with --unread flag.
func TestCLI_EmailAI_AnalyzeUnread(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfNoDefaultAIProvider(t)
	t.Parallel()

	stdout, stderr, err := runCLIWithRateLimit(t, "email", "ai", "analyze", "--limit", "5", "--unread")

	if err != nil {
		if strings.Contains(stderr, "AI is not configured") {
			t.Skip("AI not configured, skipping test")
		}
		// It's okay if no unread emails exist
		if strings.Contains(stdout, "No emails found") {
			t.Log("No unread emails found, test passed")
			return
		}
		t.Fatalf("email ai analyze --unread failed: %v\nstderr: %s", err, stderr)
	}

	t.Logf("email ai analyze --unread output:\n%s", stdout)
}

// TestCLI_EmailAI_AnalyzeFolder tests email ai analyze with --folder flag.
func TestCLI_EmailAI_AnalyzeFolder(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfNoDefaultAIProvider(t)
	t.Parallel()

	stdout, stderr, err := runCLIWithRateLimit(t, "email", "ai", "analyze", "--limit", "3", "--folder", "INBOX")

	if err != nil {
		if strings.Contains(stderr, "AI is not configured") {
			t.Skip("AI not configured, skipping test")
		}
		t.Fatalf("email ai analyze --folder INBOX failed: %v\nstderr: %s", err, stderr)
	}

	// Should fetch from INBOX
	if !strings.Contains(stdout, "Fetching") {
		t.Errorf("Expected 'Fetching' in output, got: %s", stdout)
	}

	t.Logf("email ai analyze --folder INBOX output:\n%s", stdout)
}

// TestCLI_EmailAI_AnalyzeOutputFormat tests the output format of email ai analyze.
func TestCLI_EmailAI_AnalyzeOutputFormat(t *testing.T) {
	skipIfMissingCreds(t)
	skipIfNoDefaultAIProvider(t)
	t.Parallel()

	stdout, stderr, err := runCLIWithRateLimit(t, "email", "ai", "analyze", "--limit", "5")

	if err != nil {
		if strings.Contains(stderr, "AI is not configured") {
			t.Skip("AI not configured, skipping test")
		}
		if strings.Contains(stdout, "No emails found") {
			t.Log("No emails found, skipping format check")
			return
		}
		t.Fatalf("email ai analyze failed: %v\nstderr: %s", err, stderr)
	}

	// Check for expected output sections (when emails are found and analyzed)
	// These may not all appear depending on AI response, so we just check basic structure
	if strings.Contains(stdout, "Email Analysis") {
		// Good - the header is present
		t.Log("Output format includes 'Email Analysis' header")
	}

	if strings.Contains(stdout, "Provider:") {
		// Good - provider info is shown
		t.Log("Output format includes provider information")
	}

	t.Logf("email ai analyze output format:\n%s", stdout)
}
