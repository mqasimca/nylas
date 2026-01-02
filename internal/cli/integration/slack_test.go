//go:build integration

package integration

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Slack integration tests require:
//   - SLACK_USER_TOKEN: Slack User OAuth token (xoxp-...)
//   - SLACK_USER_CHANNEL: Slack channel name for testing (without #)
//
// Run with: go test -tags=integration -v ./internal/cli/integration/... -run TestSlack

var (
	slackUserToken   string
	slackUserChannel string
)

func init() {
	slackUserToken = os.Getenv("SLACK_USER_TOKEN")
	slackUserChannel = os.Getenv("SLACK_USER_CHANNEL")
}

// skipIfMissingSlackCreds skips the test if Slack credentials are not set.
func skipIfMissingSlackCreds(t *testing.T) {
	t.Helper()

	if testBinary == "" {
		t.Skip("CLI binary not found - run 'go build -o bin/nylas ./cmd/nylas' first")
	}
	if slackUserToken == "" {
		t.Skip("SLACK_USER_TOKEN not set")
	}
	if slackUserChannel == "" {
		t.Skip("SLACK_USER_CHANNEL not set")
	}
}

// runSlackCLI executes a Slack CLI command with the token set.
func runSlackCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	// Set environment for Slack
	origToken := os.Getenv("SLACK_USER_TOKEN")
	os.Setenv("SLACK_USER_TOKEN", slackUserToken)
	defer func() {
		if origToken == "" {
			os.Unsetenv("SLACK_USER_TOKEN")
		} else {
			os.Setenv("SLACK_USER_TOKEN", origToken)
		}
	}()

	return runCLIWithTimeout(30*time.Second, args...)
}

// =============================================================================
// SLACK HELP TESTS
// =============================================================================

func TestSlack_Help(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "slack main help",
			args: []string{"slack", "--help"},
			contains: []string{
				"Slack",
				"auth",
				"channels",
				"messages",
				"send",
				"reply",
				"users",
				"search",
			},
		},
		{
			name: "slack auth help",
			args: []string{"slack", "auth", "--help"},
			contains: []string{
				"auth",
				"set",
				"status",
				"remove",
			},
		},
		{
			name: "slack channels help",
			args: []string{"slack", "channels", "--help"},
			contains: []string{
				"channels",
				"list",
				"info",
			},
		},
		{
			name: "slack channels list help",
			args: []string{"slack", "channels", "list", "--help"},
			contains: []string{
				"list",
				"--type",
				"--limit",
				"--all-workspace",
				"--created-after",
			},
		},
		{
			name: "slack messages help",
			args: []string{"slack", "messages", "--help"},
			contains: []string{
				"messages",
				"list",
			},
		},
		{
			name: "slack messages list help",
			args: []string{"slack", "messages", "list", "--help"},
			contains: []string{
				"list",
				"--channel",
				"--limit",
				"--all",
				"--expand-threads",
			},
		},
		{
			name: "slack send help",
			args: []string{"slack", "send", "--help"},
			contains: []string{
				"send",
				"--channel",
				"--text",
			},
		},
		{
			name: "slack reply help",
			args: []string{"slack", "reply", "--help"},
			contains: []string{
				"reply",
				"--thread",
				"--text",
			},
		},
		{
			name: "slack users help",
			args: []string{"slack", "users", "--help"},
			contains: []string{
				"users",
				"list",
			},
		},
		{
			name: "slack users list help",
			args: []string{"slack", "users", "list", "--help"},
			contains: []string{
				"list",
				"--limit",
				"--id",
			},
		},
		{
			name: "slack search help",
			args: []string{"slack", "search", "--help"},
			contains: []string{
				"search",
				"--query",
				"--limit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, _ := runCLI(tt.args...)

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

// =============================================================================
// SLACK AUTH TESTS
// =============================================================================

func TestSlack_AuthStatus(t *testing.T) {
	skipIfMissingSlackCreds(t)

	stdout, stderr, err := runSlackCLI(t, "slack", "auth", "status")

	if err != nil {
		// Auth status may fail if token is invalid
		if strings.Contains(stderr, "not authenticated") {
			t.Skipf("Not authenticated with Slack: %s", stderr)
		}
		t.Fatalf("slack auth status failed: %v\nstderr: %s", err, stderr)
	}

	// Should show authentication status
	if !strings.Contains(stdout, "Authenticated") &&
		!strings.Contains(stdout, "User:") &&
		!strings.Contains(stdout, "Team:") {
		t.Errorf("Expected auth status output, got: %s", stdout)
	}

	t.Logf("Auth status:\n%s", stdout)
}
