//go:build integration

package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

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

func TestCLI_ContactsListWithID(t *testing.T) {
	skipIfMissingCreds(t)

	stdout, stderr, err := runCLI("contacts", "list", testGrantID, "--id")
	skipIfProviderNotSupported(t, stderr)

	if err != nil {
		t.Fatalf("contacts list --id failed: %v\nstderr: %s", err, stderr)
	}

	// Should show contacts list with IDs or "No contacts found"
	if strings.Contains(stdout, "Found") {
		// If contacts are found, the ID column should be present
		if !strings.Contains(stdout, "ID") {
			t.Errorf("Expected 'ID' column in output with --id flag, got: %s", stdout)
		}
	}

	t.Logf("contacts list --id output:\n%s", stdout)
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
