//go:build integration

package integration

import (
	"strings"
	"testing"
)

// =============================================================================
// SCHEDULER COMMAND TESTS
// =============================================================================

func TestCLI_SchedulerHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("scheduler", "--help")

	if err != nil {
		t.Fatalf("scheduler --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show scheduler subcommands
	if !strings.Contains(stdout, "configurations") || !strings.Contains(stdout, "bookings") {
		t.Errorf("Expected scheduler subcommands in help, got: %s", stdout)
	}

	t.Logf("scheduler --help output:\n%s", stdout)
}

// Configurations Tests

func TestCLI_SchedulerConfigurationsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("scheduler", "configurations", "--help")

	if err != nil {
		t.Fatalf("scheduler configurations --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show configuration subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "create") {
		t.Errorf("Expected configuration subcommands in help, got: %s", stdout)
	}

	t.Logf("scheduler configurations --help output:\n%s", stdout)
}

func TestCLI_SchedulerConfigurationsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "configurations", "list")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("scheduler configurations list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show configurations list or "No scheduler configurations found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No scheduler configurations found") {
		t.Errorf("Expected configurations list output, got: %s", stdout)
	}

	t.Logf("scheduler configurations list output:\n%s", stdout)
}

func TestCLI_SchedulerConfigurationsListJSON(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "configurations", "list", "--json")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("scheduler configurations list --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should output JSON (array)
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "[") {
		t.Errorf("Expected JSON array output, got: %s", stdout)
	}

	t.Logf("scheduler configurations list --json output:\n%s", stdout)
}

// Sessions Tests

func TestCLI_SchedulerSessionsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("scheduler", "sessions", "--help")

	if err != nil {
		t.Fatalf("scheduler sessions --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show session subcommands
	if !strings.Contains(stdout, "create") || !strings.Contains(stdout, "show") {
		t.Errorf("Expected session subcommands in help, got: %s", stdout)
	}

	t.Logf("scheduler sessions --help output:\n%s", stdout)
}

// Bookings Tests

func TestCLI_SchedulerBookingsHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("scheduler", "bookings", "--help")

	if err != nil {
		t.Fatalf("scheduler bookings --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show booking subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "confirm") {
		t.Errorf("Expected booking subcommands in help, got: %s", stdout)
	}

	t.Logf("scheduler bookings --help output:\n%s", stdout)
}

func TestCLI_SchedulerBookingsList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "bookings", "list")
	skipIfProviderNotSupported(t, stderr)

	// Skip if bookings endpoint isn't available in this Nylas API version
	if err != nil && strings.Contains(stderr, "Unrecognized request URL") {
		t.Skip("Scheduler bookings endpoint not available in this Nylas API version")
	}

	if err != nil {
		t.Fatalf("scheduler bookings list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show bookings list or "No bookings found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No bookings found") {
		t.Errorf("Expected bookings list output, got: %s", stdout)
	}

	t.Logf("scheduler bookings list output:\n%s", stdout)
}

func TestCLI_SchedulerBookingsListJSON(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "bookings", "list", "--json")
	skipIfProviderNotSupported(t, stderr)

	// Skip if bookings endpoint isn't available in this Nylas API version
	if err != nil && strings.Contains(stderr, "Unrecognized request URL") {
		t.Skip("Scheduler bookings endpoint not available in this Nylas API version")
	}

	if err != nil {
		t.Fatalf("scheduler bookings list --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should output JSON (array)
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "[") {
		t.Errorf("Expected JSON array output, got: %s", stdout)
	}

	t.Logf("scheduler bookings list --json output:\n%s", stdout)
}

// Pages Tests

func TestCLI_SchedulerPagesHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("scheduler", "pages", "--help")

	if err != nil {
		t.Fatalf("scheduler pages --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show page subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "create") {
		t.Errorf("Expected page subcommands in help, got: %s", stdout)
	}

	t.Logf("scheduler pages --help output:\n%s", stdout)
}

func TestCLI_SchedulerPagesList(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "pages", "list")
	skipIfProviderNotSupported(t, stderr)

	// Skip if pages endpoint isn't available in this Nylas API version
	if err != nil && strings.Contains(stderr, "Unrecognized request URL") {
		t.Skip("Scheduler pages endpoint not available in this Nylas API version")
	}

	if err != nil {
		t.Fatalf("scheduler pages list failed: %v\nstderr: %s", err, stderr)
	}

	// Should show pages list or "No scheduler pages found"
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No scheduler pages found") {
		t.Errorf("Expected pages list output, got: %s", stdout)
	}

	t.Logf("scheduler pages list output:\n%s", stdout)
}

func TestCLI_SchedulerPagesListJSON(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("scheduler", "pages", "list", "--json")
	skipIfProviderNotSupported(t, stderr)

	// Skip if pages endpoint isn't available in this Nylas API version
	if err != nil && strings.Contains(stderr, "Unrecognized request URL") {
		t.Skip("Scheduler pages endpoint not available in this Nylas API version")
	}

	if err != nil {
		t.Fatalf("scheduler pages list --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should output JSON (array)
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "[") {
		t.Errorf("Expected JSON array output, got: %s", stdout)
	}

	t.Logf("scheduler pages list --json output:\n%s", stdout)
}
