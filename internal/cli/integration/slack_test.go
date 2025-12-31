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
				"--type",
				"--limit",
			},
		},
		{
			name: "slack messages help",
			args: []string{"slack", "messages", "--help"},
			contains: []string{
				"messages",
				"--channel",
				"--limit",
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
				"--limit",
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

// =============================================================================
// SLACK CHANNELS TESTS
// =============================================================================

func TestSlack_ChannelsList(t *testing.T) {
	skipIfMissingSlackCreds(t)

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "list all channels",
			args:     []string{"slack", "channels"},
			contains: []string{}, // Just verify it runs
		},
		{
			name:     "list with limit",
			args:     []string{"slack", "channels", "--limit", "5"},
			contains: []string{},
		},
		{
			name:     "list public channels only",
			args:     []string{"slack", "channels", "--type", "public_channel"},
			contains: []string{},
		},
		{
			name:     "list with IDs",
			args:     []string{"slack", "channels", "--id"},
			contains: []string{"[C"}, // Channel IDs start with C
		},
		{
			name:     "exclude archived",
			args:     []string{"slack", "channels", "--exclude-archived"},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runSlackCLI(t, tt.args...)

			if err != nil {
				if strings.Contains(stderr, "not authenticated") {
					t.Skip("Not authenticated with Slack")
				}
				t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}

			t.Logf("Output:\n%s", stdout)
		})
	}
}

// =============================================================================
// SLACK MESSAGES TESTS
// =============================================================================

func TestSlack_MessagesList(t *testing.T) {
	skipIfMissingSlackCreds(t)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "list messages from channel",
			args:     []string{"slack", "messages", "--channel", slackUserChannel},
			contains: []string{}, // Just verify it runs
		},
		{
			name:     "list messages with limit",
			args:     []string{"slack", "messages", "--channel", slackUserChannel, "--limit", "5"},
			contains: []string{},
		},
		{
			name:     "list messages with IDs",
			args:     []string{"slack", "messages", "--channel", slackUserChannel, "--id"},
			contains: []string{}, // Should show message timestamps
		},
		{
			name:    "missing channel",
			args:    []string{"slack", "messages"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runSlackCLI(t, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				if strings.Contains(stderr, "not authenticated") {
					t.Skip("Not authenticated with Slack")
				}
				if strings.Contains(stderr, "channel not found") {
					t.Skipf("Channel %s not found", slackUserChannel)
				}
				t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}

			t.Logf("Output:\n%s", stdout)
		})
	}
}

// =============================================================================
// SLACK USERS TESTS
// =============================================================================

func TestSlack_UsersList(t *testing.T) {
	skipIfMissingSlackCreds(t)

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "list users",
			args:     []string{"slack", "users"},
			contains: []string{}, // Just verify it runs
		},
		{
			name:     "list users with limit",
			args:     []string{"slack", "users", "--limit", "10"},
			contains: []string{},
		},
		{
			name:     "list users with IDs",
			args:     []string{"slack", "users", "--id"},
			contains: []string{"[U"}, // User IDs start with U
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runSlackCLI(t, tt.args...)

			if err != nil {
				if strings.Contains(stderr, "not authenticated") {
					t.Skip("Not authenticated with Slack")
				}
				t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}

			t.Logf("Output:\n%s", stdout)
		})
	}
}

// =============================================================================
// SLACK SEARCH TESTS
// =============================================================================

func TestSlack_Search(t *testing.T) {
	skipIfMissingSlackCreds(t)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "search messages",
			args:     []string{"slack", "search", "--query", "test"},
			contains: []string{}, // Just verify it runs (may return no results)
		},
		{
			name:     "search with limit",
			args:     []string{"slack", "search", "--query", "hello", "--limit", "5"},
			contains: []string{},
		},
		{
			name:    "search missing query",
			args:    []string{"slack", "search"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runSlackCLI(t, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				if strings.Contains(stderr, "not authenticated") {
					t.Skip("Not authenticated with Slack")
				}
				// Search may fail with missing_scope if user token doesn't have search:read
				if strings.Contains(stderr, "missing_scope") {
					t.Skip("Token missing search:read scope")
				}
				t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}

			t.Logf("Output:\n%s", stdout)
		})
	}
}

// =============================================================================
// SLACK SEND TESTS (Read-only by default)
// =============================================================================

func TestSlack_Send_DryRun(t *testing.T) {
	skipIfMissingSlackCreds(t)

	// This test verifies the send command validates inputs without actually sending
	// We expect an error because we're not confirming (no --yes flag and no stdin)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "send requires channel",
			args:    []string{"slack", "send", "--text", "test"},
			wantErr: true,
		},
		{
			name:    "send requires text",
			args:    []string{"slack", "send", "--channel", slackUserChannel},
			wantErr: true,
		},
		{
			name:    "reply requires thread",
			args:    []string{"slack", "reply", "--channel", slackUserChannel, "--text", "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := runSlackCLI(t, tt.args...)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none. stderr: %s", stderr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nstderr: %s", err, stderr)
			}
		})
	}
}

// TestSlack_SendMessage actually sends a message. Only runs if SLACK_TEST_SEND=true.
func TestSlack_SendMessage(t *testing.T) {
	skipIfMissingSlackCreds(t)

	if os.Getenv("SLACK_TEST_SEND") != "true" {
		t.Skip("SLACK_TEST_SEND not set to 'true' - skipping actual send test")
	}

	testMessage := "Integration test message from nylas CLI at " + time.Now().Format(time.RFC3339)

	stdout, stderr, err := runSlackCLI(t,
		"slack", "send",
		"--channel", slackUserChannel,
		"--text", testMessage,
		"--yes", // Skip confirmation
	)

	if err != nil {
		if strings.Contains(stderr, "not authenticated") {
			t.Skip("Not authenticated with Slack")
		}
		if strings.Contains(stderr, "channel not found") {
			t.Skipf("Channel %s not found", slackUserChannel)
		}
		t.Fatalf("Send failed: %v\nstderr: %s", err, stderr)
	}

	// Should confirm message was sent
	if !strings.Contains(stdout, "Message sent") && !strings.Contains(stdout, "ID:") {
		t.Errorf("Expected send confirmation, got: %s", stdout)
	}

	t.Logf("Send output:\n%s", stdout)
}

// =============================================================================
// SLACK WORKFLOW TEST
// =============================================================================

func TestSlack_Workflow(t *testing.T) {
	skipIfMissingSlackCreds(t)

	// Test a typical workflow: auth status -> list channels -> read messages

	t.Run("auth_status", func(t *testing.T) {
		stdout, stderr, err := runSlackCLI(t, "slack", "auth", "status")
		if err != nil {
			if strings.Contains(stderr, "not authenticated") {
				t.Skip("Not authenticated with Slack")
			}
			t.Fatalf("Auth status failed: %v", err)
		}
		t.Logf("Auth: %s", strings.TrimSpace(stdout))
	})

	t.Run("list_channels", func(t *testing.T) {
		stdout, stderr, err := runSlackCLI(t, "slack", "channels", "--limit", "5")
		if err != nil {
			t.Fatalf("List channels failed: %v\nstderr: %s", err, stderr)
		}

		// Verify test channel exists
		if !strings.Contains(stdout, slackUserChannel) {
			t.Logf("Warning: Test channel %s not found in first 5 channels", slackUserChannel)
		}
		t.Logf("Channels: %d lines", len(strings.Split(stdout, "\n")))
	})

	t.Run("read_messages", func(t *testing.T) {
		stdout, stderr, err := runSlackCLI(t, "slack", "messages", "--channel", slackUserChannel, "--limit", "3")
		if err != nil {
			if strings.Contains(stderr, "channel not found") {
				t.Skipf("Channel %s not found", slackUserChannel)
			}
			t.Fatalf("Read messages failed: %v\nstderr: %s", err, stderr)
		}

		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		t.Logf("Messages: %d lines of output", len(lines))
	})

	t.Run("list_users", func(t *testing.T) {
		stdout, stderr, err := runSlackCLI(t, "slack", "users", "--limit", "5")
		if err != nil {
			t.Fatalf("List users failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Users: %d lines", len(strings.Split(stdout, "\n")))
	})
}
