//go:build integration

package integration

import (
	"strings"
	"testing"
)

// =============================================================================
// SMART COMPOSE TESTS
// =============================================================================

func TestCLI_SmartComposeHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "smart-compose", "--help")

	if err != nil {
		t.Fatalf("email smart-compose --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show smart compose help
	if !strings.Contains(stdout, "AI-powered email drafts") || !strings.Contains(stdout, "--prompt") {
		t.Errorf("Expected smart compose help, got: %s", stdout)
	}

	t.Logf("email smart-compose --help output:\n%s", stdout)
}

func TestCLI_TrackingInfoHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("email", "tracking-info", "--help")

	if err != nil {
		t.Fatalf("email tracking-info --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show tracking info help
	if !strings.Contains(stdout, "email tracking") || !strings.Contains(stdout, "webhooks") {
		t.Errorf("Expected tracking info help, got: %s", stdout)
	}

	t.Logf("email tracking-info --help output:\n%s", stdout)
}

// Note: Actual Smart Compose requires Plus package and cannot be tested without proper subscription
// The following test is commented out as it would fail without the required subscription level

/*
func TestCLI_SmartCompose(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("email", "smart-compose", "--prompt", "Draft a brief thank you email")

	// Smart Compose requires Plus package, so we expect it to fail with a specific error
	// or succeed if the account has the proper subscription
	if err != nil {
		// Check if it's a subscription error
		if strings.Contains(stderr, "Plus package") || strings.Contains(stderr, "subscription") {
			t.Skip("Smart Compose requires Plus package subscription")
		}
		t.Fatalf("email smart-compose failed: %v\nstderr: %s", err, stderr)
	}

	// If successful, should contain AI-generated content
	if !strings.Contains(stdout, "AI-Generated Email") {
		t.Errorf("Expected AI-generated email output, got: %s", stdout)
	}

	t.Logf("email smart-compose output:\n%s", stdout)
}
*/
