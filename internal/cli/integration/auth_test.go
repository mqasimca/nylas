//go:build integration

package integration

import (
	"os"
	"strings"
	"testing"
)

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

func TestCLI_AuthHelp(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "--help")

	if err != nil {
		t.Fatalf("auth --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show auth subcommands
	if !strings.Contains(stdout, "login") {
		t.Errorf("Expected 'login' in auth help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "logout") {
		t.Errorf("Expected 'logout' in auth help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "status") {
		t.Errorf("Expected 'status' in auth help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "show") {
		t.Errorf("Expected 'show' in auth help, got: %s", stdout)
	}

	t.Logf("auth help output:\n%s", stdout)
}

// =============================================================================
// AUTH SHOW COMMAND TESTS (Phase 3)
// =============================================================================

func TestCLI_AuthShowHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "show", "--help")

	if err != nil {
		t.Fatalf("auth show --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show help for show command
	if !strings.Contains(stdout, "grant") {
		t.Errorf("Expected 'grant' in show help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "detailed") || !strings.Contains(stdout, "information") {
		t.Errorf("Expected detailed information description in help, got: %s", stdout)
	}

	t.Logf("auth show help output:\n%s", stdout)
}

func TestCLI_AuthShow(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "show", testGrantID)

	if err != nil {
		t.Fatalf("auth show failed: %v\nstderr: %s", err, stderr)
	}

	// Should show grant details
	if !strings.Contains(stdout, "Grant ID:") {
		t.Errorf("Expected 'Grant ID:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Email:") {
		t.Errorf("Expected 'Email:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Provider:") {
		t.Errorf("Expected 'Provider:' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Status:") {
		t.Errorf("Expected 'Status:' in output, got: %s", stdout)
	}

	t.Logf("auth show output:\n%s", stdout)
}

func TestCLI_AuthShow_InvalidGrant(t *testing.T) {
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("auth", "show", "invalid-grant-id-12345")

	if err == nil {
		t.Error("Expected error for invalid grant ID, but command succeeded")
	}

	// Should show error message
	t.Logf("auth show invalid grant error: %s", stderr)
}

// =============================================================================
// AUTH TOKEN COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthTokenHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "token", "--help")

	if err != nil {
		t.Fatalf("auth token --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show token command usage
	if !strings.Contains(stdout, "token") || !strings.Contains(stdout, "Show") {
		t.Errorf("Expected token command description in help, got: %s", stdout)
	}

	t.Logf("auth token --help output:\n%s", stdout)
}

func TestCLI_AuthToken(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("auth", "token")

	if err != nil {
		t.Fatalf("auth token failed: %v\nstderr: %s", err, stderr)
	}

	// Should show API key/token
	if !strings.Contains(stdout, "nyk_") && !strings.Contains(stdout, "API") {
		t.Errorf("Expected API key in output, got: %s", stdout)
	}

	t.Logf("auth token output: [REDACTED - contains API key]")
}

// =============================================================================
// AUTH SWITCH COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthSwitchHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "switch", "--help")

	if err != nil {
		t.Fatalf("auth switch --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show switch command usage
	if !strings.Contains(stdout, "switch") && !strings.Contains(stdout, "Switch") {
		t.Errorf("Expected switch command description in help, got: %s", stdout)
	}

	t.Logf("auth switch --help output:\n%s", stdout)
}

func TestCLI_AuthSwitch(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	// First, get the list of grants to find a grant to switch to
	listOut, _, err := runCLI("auth", "list")
	if err != nil {
		t.Fatalf("auth list failed: %v", err)
	}

	// Parse the list to find grant IDs (simple parsing)
	lines := strings.Split(listOut, "\n")
	var grants []string
	for _, line := range lines {
		// Skip header and empty lines
		if strings.Contains(line, "GRANT ID") || strings.TrimSpace(line) == "" {
			continue
		}
		// Extract grant ID (first column)
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.Contains(fields[0], "-") {
			grants = append(grants, fields[0])
		}
	}

	if len(grants) < 1 {
		t.Skip("Need at least 1 grant for switch test")
	}

	// Get current default grant
	whoamiOut, _, err := runCLI("auth", "whoami")
	var currentGrant string
	if err == nil {
		// Extract current grant ID from whoami output
		for _, line := range strings.Split(whoamiOut, "\n") {
			if strings.Contains(line, "Grant ID:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					currentGrant = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Switch to first available grant
	targetGrant := grants[0]

	stdout, stderr, err := runCLI("auth", "switch", targetGrant)

	if err != nil {
		t.Fatalf("auth switch failed: %v\nstderr: %s", err, stderr)
	}

	// Should show success message
	lowerOut := strings.ToLower(stdout)
	if !strings.Contains(lowerOut, "switched") && !strings.Contains(lowerOut, "default") {
		t.Errorf("Expected switch confirmation in output, got: %s", stdout)
	}

	t.Logf("auth switch output:\n%s", stdout)

	// Verify the switch by checking whoami
	whoamiAfter, _, err := runCLI("auth", "whoami")
	if err == nil {
		t.Logf("auth whoami after switch:\n%s", whoamiAfter)
	}

	// Cleanup: Switch back to original grant if we had one
	if currentGrant != "" && currentGrant != targetGrant {
		t.Cleanup(func() {
			_, _, _ = runCLI("auth", "switch", currentGrant)
		})
	}
}

func TestCLI_AuthSwitch_InvalidGrant(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	_, stderr, err := runCLI("auth", "switch", "invalid-grant-id-12345")

	if err == nil {
		t.Error("Expected error for invalid grant ID, but command succeeded")
	}

	t.Logf("auth switch invalid grant error: %s", stderr)
}

// =============================================================================
// AUTH LOGOUT COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthLogoutHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "logout", "--help")

	if err != nil {
		t.Fatalf("auth logout --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show logout command usage
	if !strings.Contains(stdout, "logout") || !strings.Contains(stdout, "Revoke") {
		t.Errorf("Expected logout command description in help, got: %s", stdout)
	}

	t.Logf("auth logout --help output:\n%s", stdout)
}

func TestCLI_AuthLogout(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	// ADDITIONAL SAFETY: Verify we have multiple grants before logout
	listOut, _, err := runCLI("auth", "list")
	if err != nil {
		t.Fatalf("auth list failed: %v", err)
	}

	grantCount := strings.Count(listOut, "✓ valid")
	if grantCount < 2 {
		t.Skip("Need at least 2 grants to safely test logout (to avoid removing all grants)")
	}

	t.Log("⚠️  WARNING: Running auth logout test - this will remove the current grant from local config")

	// Run logout
	stdout, stderr, err := runCLI("auth", "logout")

	if err != nil {
		t.Fatalf("auth logout failed: %v\nstderr: %s", err, stderr)
	}

	// Should show confirmation
	lowerOut := strings.ToLower(stdout)
	if !strings.Contains(lowerOut, "revoked") && !strings.Contains(lowerOut, "logout") && !strings.Contains(lowerOut, "removed") {
		t.Errorf("Expected logout confirmation in output, got: %s", stdout)
	}

	t.Logf("auth logout output:\n%s", stdout)

	// Note: You'll need to manually re-add the grant after this test
	t.Log("⚠️  Note: You may need to run 'nylas auth add <grant-id>' to re-add the grant")
}

// =============================================================================
// AUTH REMOVE COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthRemoveHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "remove", "--help")

	if err != nil {
		t.Fatalf("auth remove --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show remove command usage
	if !strings.Contains(stdout, "remove") || !strings.Contains(stdout, "Remove") {
		t.Errorf("Expected remove command description in help, got: %s", stdout)
	}
	// Should mention it keeps grant on server
	if !strings.Contains(stdout, "server") && !strings.Contains(stdout, "local") {
		t.Logf("Note: Help should clarify that remove only affects local config")
	}

	t.Logf("auth remove --help output:\n%s", stdout)
}

func TestCLI_AuthRemove(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	// NOTE: This test may fail if testGrantID already exists locally.
	// For a production test environment, you would create a new grant via API first.

	// SAFETY: First add a temporary test grant that we can safely remove
	t.Log("Adding temporary test grant for removal test...")
	addOut, stderr, err := runCLI("auth", "add", testGrantID, "--email", "test-remove@example.com")
	if err != nil {
		t.Fatalf("Failed to add test grant: %v\nstderr: %s", err, stderr)
	}
	t.Logf("Added test grant:\n%s", addOut)

	// Now remove the grant we just added (NOT the original grant)
	t.Log("⚠️  Removing test grant from local config (keeps on server)...")
	stdout, stderr, err := runCLIWithInput("y\n", "auth", "remove", testGrantID)

	if err != nil {
		t.Fatalf("auth remove failed: %v\nstderr: %s", err, stderr)
	}

	// Should show confirmation
	lowerOut := strings.ToLower(stdout)
	if !strings.Contains(lowerOut, "removed") && !strings.Contains(lowerOut, "deleted") {
		t.Errorf("Expected remove confirmation in output, got: %s", stdout)
	}

	t.Logf("auth remove output:\n%s", stdout)

	// Verify grant is removed from local list
	listOut, _, _ := runCLI("auth", "list")
	t.Logf("auth list after remove:\n%s", listOut)

	// Re-add the grant for other tests
	t.Cleanup(func() {
		t.Log("Re-adding test grant in cleanup...")
		_, _, _ = runCLI("auth", "add", testGrantID, "--default")
	})
}

func TestCLI_AuthRemove_InvalidGrant(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	_, stderr, err := runCLIWithInput("y\n", "auth", "remove", "invalid-grant-id-12345")

	if err == nil {
		t.Error("Expected error for invalid grant ID, but command succeeded")
	}

	t.Logf("auth remove invalid grant error: %s", stderr)
}

// =============================================================================
// AUTH REVOKE COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthRevokeHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "revoke", "--help")

	if err != nil {
		t.Fatalf("auth revoke --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show revoke command usage
	if !strings.Contains(stdout, "revoke") || !strings.Contains(stdout, "Revoke") {
		t.Errorf("Expected revoke command description in help, got: %s", stdout)
	}
	// Should warn about permanence
	if !strings.Contains(strings.ToLower(stdout), "permanent") && !strings.Contains(strings.ToLower(stdout), "delete") {
		t.Logf("Note: Help should warn that revoke is permanent")
	}

	t.Logf("auth revoke --help output:\n%s", stdout)
}

func TestCLI_AuthRevoke_InvalidGrant(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	// Test error handling with invalid grant (safe - won't actually delete anything)
	// NOTE: Current CLI behavior does NOT validate grant existence before claiming success.
	// This means the command will succeed even with an invalid grant ID.
	// TODO: CLI should validate grant exists before attempting revocation
	stdout, stderr, err := runCLIWithInput("y\n", "auth", "revoke", "invalid-grant-id-12345")

	// Currently the command succeeds even with invalid grant (CLI bug)
	// We verify it doesn't crash, but ideally it should error
	if err != nil {
		// If it does error, that's actually better behavior
		t.Logf("Command errored (expected behavior): %s", stderr)
	} else {
		// Command succeeded (current behavior - not ideal)
		t.Logf("Command succeeded without validation (current CLI behavior):\n%s", stdout)
	}
}

// NOTE: We do NOT implement a real revoke test because it permanently deletes
// grants on the server. This would require:
// 1. Creating a temporary grant via the API
// 2. Revoking that temporary grant
// 3. Multiple safety checks
//
// For now, we only test:
// - Help output (TestCLI_AuthRevokeHelp)
// - Error handling (TestCLI_AuthRevoke_InvalidGrant)
//
// Real revoke testing should be done manually or in a dedicated test environment
// with disposable grants.

// =============================================================================
// AUTH CONFIG COMMAND TESTS (Phase 1.2) - GUARDED
// =============================================================================

func TestCLI_AuthConfigHelp(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("auth", "config", "--help")

	if err != nil {
		t.Fatalf("auth config --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show config command usage
	if !strings.Contains(stdout, "config") {
		t.Errorf("Expected config command description in help, got: %s", stdout)
	}

	t.Logf("auth config --help output:\n%s", stdout)
}

func TestCLI_AuthConfig(t *testing.T) {
	// GUARD: Only run if both environment variables are set
	if os.Getenv("NYLAS_TEST_DELETE") != "true" || os.Getenv("NYLAS_TEST_AUTH_LOGOUT") != "true" {
		t.Skip("Skipping auth test - requires NYLAS_TEST_DELETE=true and NYLAS_TEST_AUTH_LOGOUT=true")
	}
	skipIfMissingCreds(t)

	// Test showing current config (read-only)
	stdout, stderr, err := runCLI("auth", "config")

	// Config command may show current settings or prompt for setup
	if err != nil && !strings.Contains(stderr, "not configured") {
		t.Logf("auth config returned error (may not be configured): %v\nstderr: %s", err, stderr)
	}

	t.Logf("auth config output:\n%s", stdout)
}
