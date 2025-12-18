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
// CALENDAR CRUD COMMAND TESTS
// =============================================================================

func TestCLI_CalendarShowHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "show", "--help")

	if err != nil {
		t.Fatalf("calendar show --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage
	if !strings.Contains(stdout, "<calendar-id>") {
		t.Errorf("Expected '<calendar-id>' in help, got: %s", stdout)
	}

	t.Logf("calendar show --help output:\n%s", stdout)
}

func TestCLI_CalendarCreateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "create", "--help")

	if err != nil {
		t.Fatalf("calendar create --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage and flags
	if !strings.Contains(stdout, "<name>") {
		t.Errorf("Expected '<name>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--description") || !strings.Contains(stdout, "--timezone") {
		t.Errorf("Expected --description and --timezone flags in help, got: %s", stdout)
	}

	t.Logf("calendar create --help output:\n%s", stdout)
}

func TestCLI_CalendarUpdateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "update", "--help")

	if err != nil {
		t.Fatalf("calendar update --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage and flags
	if !strings.Contains(stdout, "<calendar-id>") {
		t.Errorf("Expected '<calendar-id>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--name") || !strings.Contains(stdout, "--color") {
		t.Errorf("Expected --name and --color flags in help, got: %s", stdout)
	}

	t.Logf("calendar update --help output:\n%s", stdout)
}

func TestCLI_CalendarDeleteHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "delete", "--help")

	if err != nil {
		t.Fatalf("calendar delete --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage and flags
	if !strings.Contains(stdout, "<calendar-id>") {
		t.Errorf("Expected '<calendar-id>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--force") {
		t.Errorf("Expected --force flag in help, got: %s", stdout)
	}

	t.Logf("calendar delete --help output:\n%s", stdout)
}

func TestCLI_CalendarCRUDLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	calendarName := fmt.Sprintf("CLI Test Calendar %d", time.Now().Unix())
	var calendarID string

	// Create calendar
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "create", calendarName,
			"--description", "Test calendar created by CLI",
			"--timezone", "America/New_York",
			testGrantID)

		if err != nil {
			// Some providers may not support calendar creation
			if strings.Contains(stderr, "not supported") || strings.Contains(stderr, "read-only") {
				t.Skip("Calendar creation not supported by provider")
			}
			t.Fatalf("calendar create failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Created calendar") {
			t.Errorf("Expected 'Created calendar' in output, got: %s", stdout)
		}

		// Extract calendar ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			calendarID = strings.TrimSpace(stdout[idx+3:])
			if paren := strings.Index(calendarID, ")"); paren != -1 {
				calendarID = calendarID[:paren]
			}
			if newline := strings.Index(calendarID, "\n"); newline != -1 {
				calendarID = calendarID[:newline]
			}
		}

		t.Logf("calendar create output: %s", stdout)
		t.Logf("Calendar ID: %s", calendarID)
	})

	if calendarID == "" {
		t.Fatal("Failed to get calendar ID from create output")
	}

	// Wait for calendar to sync
	time.Sleep(2 * time.Second)

	// Show calendar
	t.Run("show", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "show", calendarID, testGrantID)
		if err != nil {
			t.Fatalf("calendar show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, calendarName) {
			t.Errorf("Expected calendar name in output, got: %s", stdout)
		}

		t.Logf("calendar show output:\n%s", stdout)
	})

	// Update calendar
	t.Run("update", func(t *testing.T) {
		newName := calendarName + " Updated"
		stdout, stderr, err := runCLI("calendar", "update", calendarID,
			"--name", newName,
			"--description", "Updated description",
			testGrantID)

		if err != nil {
			// Some providers may not support calendar updates
			if strings.Contains(stderr, "not supported") {
				t.Skip("Calendar update not supported by provider")
			}
			t.Fatalf("calendar update failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Updated calendar") {
			t.Errorf("Expected 'Updated calendar' in output, got: %s", stdout)
		}

		t.Logf("calendar update output: %s", stdout)
	})

	// Delete calendar
	t.Run("delete", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "delete", calendarID, "--force", testGrantID)
		if err != nil {
			// Some providers may not support calendar deletion
			if strings.Contains(stderr, "not supported") {
				t.Skip("Calendar deletion not supported by provider")
			}
			t.Fatalf("calendar delete failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "deleted") {
			t.Errorf("Expected 'deleted' in output, got: %s", stdout)
		}

		t.Logf("calendar delete output: %s", stdout)
	})
}

// =============================================================================
// CALENDAR EVENT UPDATE/RSVP COMMAND TESTS
// =============================================================================

func TestCLI_CalendarEventsUpdateHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "events", "update", "--help")

	if err != nil {
		t.Fatalf("calendar events update --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage and flags
	if !strings.Contains(stdout, "<event-id>") {
		t.Errorf("Expected '<event-id>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--title") || !strings.Contains(stdout, "--location") {
		t.Errorf("Expected --title and --location flags in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--visibility") {
		t.Errorf("Expected --visibility flag in help, got: %s", stdout)
	}

	t.Logf("calendar events update --help output:\n%s", stdout)
}

func TestCLI_CalendarEventsRSVPHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "events", "rsvp", "--help")

	if err != nil {
		t.Fatalf("calendar events rsvp --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show usage and status options
	if !strings.Contains(stdout, "<event-id>") {
		t.Errorf("Expected '<event-id>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "<status>") {
		t.Errorf("Expected '<status>' in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "yes") || !strings.Contains(stdout, "no") || !strings.Contains(stdout, "maybe") {
		t.Errorf("Expected RSVP status options (yes, no, maybe) in help, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--comment") {
		t.Errorf("Expected --comment flag in help, got: %s", stdout)
	}

	t.Logf("calendar events rsvp --help output:\n%s", stdout)
}

func TestCLI_CalendarEventsUpdateLifecycle(t *testing.T) {
	skipIfMissingCreds(t)

	if os.Getenv("NYLAS_TEST_DELETE") != "true" {
		t.Skip("NYLAS_TEST_DELETE not set to 'true'")
	}

	// Get tomorrow's date for the event
	tomorrow := time.Now().AddDate(0, 0, 1)
	startTime := tomorrow.Format("2006-01-02") + " 14:00"
	endTime := tomorrow.Format("2006-01-02") + " 15:00"
	eventTitle := fmt.Sprintf("CLI Update Test %d", time.Now().Unix())

	var eventID string

	// Create event first
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "events", "create",
			"--title", eventTitle,
			"--start", startTime,
			"--end", endTime,
			"--location", "Original Location",
			testGrantID)

		if err != nil {
			if strings.Contains(stderr, "no writable calendar") || strings.Contains(stderr, "no calendars") {
				t.Skip("No writable calendar available")
			}
			t.Fatalf("calendar events create failed: %v\nstderr: %s", err, stderr)
		}

		// Extract event ID from output
		if idx := strings.Index(stdout, "ID:"); idx != -1 {
			eventID = strings.TrimSpace(stdout[idx+3:])
			if newline := strings.Index(eventID, "\n"); newline != -1 {
				eventID = eventID[:newline]
			}
		}

		t.Logf("Event created with ID: %s", eventID)
	})

	if eventID == "" {
		t.Fatal("Failed to get event ID from create output")
	}

	// Wait for event to sync
	time.Sleep(2 * time.Second)

	// Update event
	t.Run("update", func(t *testing.T) {
		newTitle := eventTitle + " Updated"
		stdout, stderr, err := runCLI("calendar", "events", "update", eventID,
			"--title", newTitle,
			"--location", "Updated Location",
			"--description", "Updated description",
			testGrantID)

		if err != nil {
			t.Fatalf("calendar events update failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Event updated") {
			t.Errorf("Expected 'Event updated' in output, got: %s", stdout)
		}

		t.Logf("calendar events update output: %s", stdout)
	})

	// Verify update by showing the event
	t.Run("verify", func(t *testing.T) {
		stdout, stderr, err := runCLI("calendar", "events", "show", eventID, testGrantID)
		if err != nil {
			t.Fatalf("calendar events show failed: %v\nstderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Updated") {
			t.Errorf("Expected updated title in output, got: %s", stdout)
		}

		t.Logf("calendar events show (updated) output:\n%s", stdout)
	})

	// Clean up
	t.Run("cleanup", func(t *testing.T) {
		runCLIWithInput("y\n", "calendar", "events", "delete", eventID, testGrantID)
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
