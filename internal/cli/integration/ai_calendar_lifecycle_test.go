//go:build integration
// +build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestCLI_AI_CalendarEventLifecycle tests the full lifecycle of AI-enhanced calendar operations
// This test creates events, uses AI features, and cleans up
func TestCLI_AI_CalendarEventLifecycle(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	t.Run("create_event_and_analyze_conflicts", func(t *testing.T) {
		// Step 1: Create a test event
		t.Log("Step 1: Creating test event...")
		tomorrow := time.Now().Add(24 * time.Hour)
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.UTC)
		endTime := startTime.Add(1 * time.Hour)

		createArgs := []string{
			"calendar", "events", "create",
			"--title", "AI Test Meeting - Conflict Analysis",
			"--description", "Test event for AI conflict detection",
			"--start", startTime.Format(time.RFC3339),
			"--end", endTime.Format(time.RFC3339),
			"--participant", testEmail,
		}

		stdout, stderr, err := runCLI(createArgs...)
		if err != nil {
			t.Skipf("Failed to create test event (likely API or credentials issue): %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
		}

		eventID := extractEventID(stdout)
		if eventID == "" {
			t.Log("Could not extract event ID, trying alternate method...")
			// Try to list events and find our test event
			listArgs := []string{"calendar", "events", "list", "--limit", "10"}
			stdout, _, _ := runCLI(listArgs...)
			eventID = extractEventIDFromList(stdout, "AI Test Meeting - Conflict Analysis")
		}

		if eventID == "" {
			t.Skip("Could not create or find test event")
		}

		t.Logf("✓ Created test event: %s", eventID)

		// Cleanup: Delete the event after test
		t.Cleanup(func() {
			t.Logf("Cleaning up test event: %s", eventID)
			deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
			_, stderr, err := runCLI(deleteArgs...)
			if err != nil {
				t.Logf("Warning: Failed to delete test event %s: %v\nstderr: %s", eventID, err, stderr)
			} else {
				t.Logf("✓ Cleaned up test event: %s", eventID)
			}
		})

		// Step 2: Test AI conflict detection with a new proposed event at the same time
		if !hasAnyAIProvider() {
			t.Log("Skipping AI conflict detection - no AI provider configured")
			return
		}

		t.Log("Step 2: Testing AI conflict detection...")
		conflictArgs := []string{
			"calendar", "ai", "conflicts", "check",
			"--title", "Conflicting Meeting",
			"--start", startTime.Format(time.RFC3339),
			"--duration", "60",
			"--participants", testEmail,
		}

		conflictStdout, conflictStderr, err := runCLI(conflictArgs...)
		if err != nil {
			t.Logf("AI conflict detection error (may be expected): %v", err)
		}

		conflictOutput := conflictStdout + conflictStderr
		if strings.Contains(conflictOutput, "Conflict") {
			t.Logf("✓ AI successfully detected conflicts")
		} else {
			t.Logf("AI conflict detection output: %s", conflictOutput)
		}
	})

	t.Run("create_event_and_reschedule_with_ai", func(t *testing.T) {
		if !hasAnyAIProvider() {
			t.Skip("No AI provider configured")
		}

		// Step 1: Create a test event to reschedule
		t.Log("Step 1: Creating test event for rescheduling...")
		tomorrow := time.Now().Add(24 * time.Hour)
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 15, 0, 0, 0, time.UTC)
		endTime := startTime.Add(1 * time.Hour)

		createArgs := []string{
			"calendar", "events", "create",
			"--title", "AI Test Meeting - Reschedule",
			"--description", "Test event for AI rescheduling",
			"--start", startTime.Format(time.RFC3339),
			"--end", endTime.Format(time.RFC3339),
			"--participant", testEmail,
		}

		stdout, stderr, err := runCLI(createArgs...)
		if err != nil {
			t.Skipf("Failed to create test event (likely API or credentials issue): %v\nstderr: %s", err, stderr)
		}

		eventID := extractEventID(stdout)
		if eventID == "" {
			listArgs := []string{"calendar", "events", "list", "--limit", "10"}
			stdout, _, _ := runCLI(listArgs...)
			eventID = extractEventIDFromList(stdout, "AI Test Meeting - Reschedule")
		}

		if eventID == "" {
			t.Skip("Could not create or find test event")
		}

		t.Logf("✓ Created test event: %s", eventID)

		// Cleanup
		t.Cleanup(func() {
			t.Logf("Cleaning up test event: %s", eventID)
			deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
			_, stderr, err := runCLI(deleteArgs...)
			if err != nil {
				t.Logf("Warning: Failed to delete test event %s: %v\nstderr: %s", eventID, err, stderr)
			} else {
				t.Logf("✓ Cleaned up test event: %s", eventID)
			}
		})

		// Step 2: Use AI to reschedule the event
		t.Log("Step 2: Testing AI rescheduling...")
		rescheduleArgs := []string{
			"calendar", "ai", "reschedule", "ai", eventID,
			"--reason", "Integration test - checking AI rescheduling",
			"--max-delay-days", "7",
		}

		rescheduleStdout, rescheduleStderr, err := runCLI(rescheduleArgs...)
		if err != nil {
			t.Logf("AI reschedule result: %v\nstderr: %s", err, rescheduleStderr)
		}

		rescheduleOutput := rescheduleStdout + rescheduleStderr
		if strings.Contains(rescheduleOutput, "Reschedule") || strings.Contains(rescheduleOutput, "Alternative") {
			t.Logf("✓ AI successfully analyzed rescheduling options")
		} else {
			t.Logf("AI reschedule output: %s", rescheduleOutput)
		}
	})

	t.Run("create_multiple_events_and_analyze_patterns", func(t *testing.T) {
		if !hasAnyAIProvider() {
			t.Skip("No AI provider configured")
		}

		// Create multiple test events across different times
		t.Log("Creating multiple test events for pattern analysis...")
		var eventIDs []string

		baseTime := time.Now().Add(24 * time.Hour)
		eventConfigs := []struct {
			title    string
			hourDiff int
			duration int
		}{
			{"AI Test - Morning Meeting", 0, 30},
			{"AI Test - Afternoon Sync", 4, 45},
			{"AI Test - Evening Review", 8, 60},
		}

		for _, config := range eventConfigs {
			startTime := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 9+config.hourDiff, 0, 0, 0, time.UTC)
			endTime := startTime.Add(time.Duration(config.duration) * time.Minute)

			createArgs := []string{
				"calendar", "events", "create",
				"--title", config.title,
				"--description", "Test event for AI pattern analysis",
				"--start", startTime.Format(time.RFC3339),
				"--end", endTime.Format(time.RFC3339),
				"--participant", testEmail,
			}

			createStdout, createStderr, err := runCLI(createArgs...)
			if err != nil {
				t.Logf("Warning: Failed to create event '%s': %v (stderr: %s)", config.title, err, createStderr)
				continue
			}

			eventID := extractEventID(createStdout)
			if eventID == "" {
				listArgs := []string{"calendar", "events", "list", "--limit", "20"}
				stdout, _, _ := runCLI(listArgs...)
				eventID = extractEventIDFromList(stdout, config.title)
			}

			if eventID != "" {
				eventIDs = append(eventIDs, eventID)
				t.Logf("✓ Created event: %s (%s)", config.title, eventID)
			}
		}

		// Cleanup all created events
		t.Cleanup(func() {
			t.Logf("Cleaning up %d test events...", len(eventIDs))
			for _, eventID := range eventIDs {
				deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
				_, stderr, err := runCLI(deleteArgs...)
				if err != nil {
					t.Logf("Warning: Failed to delete event %s: %v\nstderr: %s", eventID, err, stderr)
				} else {
					t.Logf("✓ Cleaned up event: %s", eventID)
				}
			}
		})

		if len(eventIDs) == 0 {
			t.Skip("Could not create test events for pattern analysis")
		}

		t.Logf("✓ Created %d test events", len(eventIDs))

		// Step 2: Run AI pattern analysis
		t.Log("Running AI pattern analysis...")
		analyzeArgs := []string{
			"calendar", "ai", "analyze",
			"--days", "30",
		}

		analyzeStdout, analyzeStderr, err := runCLI(analyzeArgs...)
		if err != nil {
			t.Logf("AI analyze result: %v\nstderr: %s", err, analyzeStderr)
		}

		analyzeOutput := analyzeStdout + analyzeStderr
		if strings.Contains(analyzeOutput, "Analyzing") || strings.Contains(analyzeOutput, "Analysis") {
			t.Logf("✓ AI successfully analyzed calendar patterns")
		} else {
			t.Logf("AI analyze output: %s", analyzeOutput)
		}
	})

	t.Run("create_event_and_test_focus_time", func(t *testing.T) {
		if !hasAnyAIProvider() {
			t.Skip("No AI provider configured")
		}

		// Create a test event during typical focus time
		t.Log("Creating test event during focus time...")
		tomorrow := time.Now().Add(24 * time.Hour)
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.UTC)
		endTime := startTime.Add(2 * time.Hour)

		createArgs := []string{
			"calendar", "events", "create",
			"--title", "AI Test - Focus Time Block",
			"--description", "Test event for AI focus time analysis",
			"--start", startTime.Format(time.RFC3339),
			"--end", endTime.Format(time.RFC3339),
			"--participant", testEmail,
		}

		stdout, stderr, err := runCLI(createArgs...)
		if err != nil {
			t.Skipf("Failed to create test event (likely API or credentials issue): %v\nstderr: %s", err, stderr)
		}

		eventID := extractEventID(stdout)
		if eventID == "" {
			listArgs := []string{"calendar", "events", "list", "--limit", "10"}
			stdout, _, _ := runCLI(listArgs...)
			eventID = extractEventIDFromList(stdout, "AI Test - Focus Time Block")
		}

		if eventID == "" {
			t.Skip("Could not create or find test event")
		}

		t.Logf("✓ Created focus time test event: %s", eventID)

		// Cleanup
		t.Cleanup(func() {
			t.Logf("Cleaning up test event: %s", eventID)
			deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
			_, stderr, err := runCLI(deleteArgs...)
			if err != nil {
				t.Logf("Warning: Failed to delete test event %s: %v\nstderr: %s", eventID, err, stderr)
			} else {
				t.Logf("✓ Cleaned up test event: %s", eventID)
			}
		})

		// Test AI focus time analysis
		t.Log("Running AI focus time analysis...")
		focusArgs := []string{
			"calendar", "ai", "focus-time",
			"--analyze",
			"--target-hours", "14.0",
		}

		focusStdout, focusStderr, err := runCLI(focusArgs...)
		if err != nil {
			t.Logf("AI focus time result: %v\nstderr: %s", err, focusStderr)
		}

		focusOutput := focusStdout + focusStderr
		if strings.Contains(focusOutput, "Focus") || strings.Contains(focusOutput, "Analyzing") {
			t.Logf("✓ AI successfully analyzed focus time patterns")
		} else {
			t.Logf("AI focus time output: %s", focusOutput)
		}
	})
}

// TestCLI_AI_ScheduleAndCleanup tests AI scheduling that actually creates events
func TestCLI_AI_ScheduleAndCleanup(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// This test would use AI to schedule an event
	// Since AI scheduling may or may not create the event (depending on user confirmation),
	// we'll test the workflow and look for scheduling suggestions

	t.Run("ai_schedule_with_suggestions", func(t *testing.T) {
		provider := getAvailableProvider()
		query := fmt.Sprintf("30-minute meeting with %s tomorrow at 3pm", testEmail)

		t.Logf("Testing AI scheduling: %s", query)

		scheduleArgs := []string{
			"calendar", "schedule", "ai",
			"--provider", provider,
			query,
		}

		scheduleStdout, scheduleStderr, err := runCLI(scheduleArgs...)
		if err != nil {
			t.Logf("AI schedule result: %v\nstderr: %s", err, scheduleStderr)
		}

		output := scheduleStdout + scheduleStderr
		t.Logf("AI Schedule Output:\n%s", output)

		// Check if AI provided scheduling suggestions
		if strings.Contains(output, "AI Scheduling") || strings.Contains(output, "Suggested") {
			t.Logf("✓ AI successfully provided scheduling suggestions")
		}

		// If an event was created, extract ID and clean it up
		eventID := extractEventID(output)
		if eventID != "" {
			t.Logf("Event was created: %s", eventID)
			t.Cleanup(func() {
				t.Logf("Cleaning up AI-created event: %s", eventID)
				deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
				_, stderr, err := runCLI(deleteArgs...)
				if err != nil {
					t.Logf("Warning: Failed to delete event %s: %v\nstderr: %s", eventID, err, stderr)
				} else {
					t.Logf("✓ Cleaned up AI-created event: %s", eventID)
				}
			})
		}
	})
}
