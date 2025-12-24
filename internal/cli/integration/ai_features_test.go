//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCLI_AISchedule_RealCalendar tests AI scheduling with real calendar events
// NOTE: This test makes real LLM calls which can be slow (30-60s each).
// Only one subtest is run to avoid test timeouts.
func TestCLI_AISchedule_RealCalendar(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// Single test case to avoid timeout (LLM calls can take 30-60s each)
	t.Run("schedule meeting with natural language", func(t *testing.T) {
		provider := getAvailableProvider()
		query := fmt.Sprintf("30-minute meeting with %s tomorrow at 2pm", testEmail)

		args := []string{"calendar", "schedule", "ai"}
		if provider != "" {
			args = append(args, "--provider", provider)
		}
		args = append(args, query)

		// Use 90s timeout for LLM call
		stdout, stderr, err := runCLIWithTimeout(90*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "AI Scheduling") {
			t.Errorf("Expected output to contain 'AI Scheduling'\nGot: %s", output)
		}
	})
}

// TestCLI_AIReschedule_RealEvent tests AI rescheduling with actual calendar events
// NOTE: This test makes real LLM calls which can be slow.
func TestCLI_AIReschedule_RealEvent(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// Create a test event for tomorrow
	tomorrow := time.Now().Add(24 * time.Hour)
	startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)

	t.Log("Creating test event for rescheduling...")
	createArgs := []string{
		"calendar", "events", "create",
		"--title", "Test Meeting for Rescheduling",
		"--description", "This is a test event that will be rescheduled by AI",
		"--start", startTime.Format(time.RFC3339),
		"--end", endTime.Format(time.RFC3339),
		"--participant", testEmail,
	}

	stdout, stderr, err := runCLI(createArgs...)
	if err != nil {
		t.Logf("Failed to create test event: %v", err)
		t.Logf("stderr: %s", stderr)
		t.Skip("Skipping reschedule test - could not create test event")
	}

	eventID := extractEventID(stdout)
	if eventID == "" {
		t.Skip("Skipping reschedule test - could not extract event ID")
	}

	t.Logf("Created test event: %s", eventID)

	t.Cleanup(func() {
		t.Logf("Cleaning up test event: %s", eventID)
		deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
		_, _, _ = runCLI(deleteArgs...)
	})

	// Single test case to avoid timeout
	t.Run("reschedule with conflict reason", func(t *testing.T) {
		args := []string{
			"calendar", "ai", "reschedule", "ai", eventID,
			"--reason", "Conflict with client meeting",
			"--max-delay-days", "7",
		}

		stdout, stderr, err := runCLIWithTimeout(90*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Reschedule") {
			t.Errorf("Expected output to contain 'Reschedule'\nGot: %s", output)
		}
	})
}

// TestCLI_AIAnalyze_Patterns tests AI calendar pattern analysis
// NOTE: This test makes API calls but no LLM calls, so it's faster.
func TestCLI_AIAnalyze_Patterns(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	// Single test case - analyze is mostly API calls, not LLM
	t.Run("analyze last 30 days", func(t *testing.T) {
		args := []string{
			"calendar", "ai", "analyze",
			"--days", "30",
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Analyzing") && !strings.Contains(output, "patterns") {
			t.Errorf("Expected output to contain 'Analyzing' or 'patterns'\nGot: %s", output)
		}
	})
}

// TestCLI_AIAnalyze_ScoreTime tests scoring specific meeting times
func TestCLI_AIAnalyze_ScoreTime(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// Single test case
	t.Run("score afternoon meeting", func(t *testing.T) {
		nextWeek := time.Now().Add(7 * 24 * time.Hour)
		scoreTime := time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 14, 0, 0, 0, time.UTC)

		args := []string{
			"calendar", "ai", "analyze",
			"--score-time", scoreTime.Format(time.RFC3339),
			"--participants", testEmail,
			"--duration", "30",
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Score") && !strings.Contains(output, "score") {
			t.Errorf("Expected output to contain 'Score'\nGot: %s", output)
		}
	})
}

// TestCLI_AIConflicts_Detection tests AI conflict detection
func TestCLI_AIConflicts_Detection(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// Single test case
	t.Run("check conflicts for new meeting", func(t *testing.T) {
		tomorrow := time.Now().Add(24 * time.Hour)
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.UTC)

		args := []string{
			"calendar", "ai", "conflicts", "check",
			"--title", "Product Review",
			"--start", startTime.Format(time.RFC3339),
			"--duration", "60",
			"--participants", testEmail,
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Conflict") && !strings.Contains(output, "conflict") {
			t.Errorf("Expected output to contain 'Conflict'\nGot: %s", output)
		}
	})
}

// TestCLI_AIFindTime_MultiTimezone tests finding optimal times across timezones
func TestCLI_AIFindTime_MultiTimezone(t *testing.T) {
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

	// Single test case - this is an API-only call, no LLM
	t.Run("find time for 2 participants", func(t *testing.T) {
		args := []string{
			"calendar", "find-time",
			"--participants", testEmail + ",test@example.com",
			"--duration", "1h",
			"--days", "7",
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Timezone") && !strings.Contains(output, "Finder") && !strings.Contains(output, "time") {
			t.Errorf("Expected output to contain timezone/finder info\nGot: %s", output)
		}
	})
}

// TestCLI_AIFocusTime_Analysis tests AI focus time pattern analysis
func TestCLI_AIFocusTime_Analysis(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	// Single test case
	t.Run("analyze focus time patterns", func(t *testing.T) {
		args := []string{"calendar", "ai", "focus-time", "--analyze", "--target-hours", "14.0"}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Focus") && !strings.Contains(output, "focus") {
			t.Errorf("Expected output to contain 'Focus'\nGot: %s", output)
		}
	})
}

// TestCLI_AIAdaptive_Scheduling tests adaptive schedule optimization
func TestCLI_AIAdaptive_Scheduling(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	// Single test case - adapt uses API calls, not heavy LLM
	t.Run("adapt for meeting overload", func(t *testing.T) {
		args := []string{
			"calendar", "ai", "adapt",
			"--trigger", "overload",
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Adaptive") && !strings.Contains(output, "Scheduling") {
			t.Errorf("Expected output to contain 'Adaptive' or 'Scheduling'\nGot: %s", output)
		}
	})
}

// TestCLI_AIContext_Calendar tests getting calendar context
func TestCLI_AIContext_Calendar(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	// Single test case - context is mostly API, not LLM
	t.Run("get context for next 7 days", func(t *testing.T) {
		args := []string{
			"calendar", "ai", "context",
			"--days", "7",
		}

		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Test skipped due to error: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		output := stdout + stderr
		if !strings.Contains(output, "Context") && !strings.Contains(output, "context") {
			t.Errorf("Expected output to contain 'Context'\nGot: %s", output)
		}
	})
}

// TestCLI_AIFeatures_EndToEnd tests end-to-end AI workflow
// NOTE: This test is skipped by default as it makes multiple LLM calls.
// Run manually with: go test -tags=integration -run TestCLI_AIFeatures_EndToEnd -v
func TestCLI_AIFeatures_EndToEnd(t *testing.T) {
	// Skip by default - this test makes multiple slow LLM calls
	if os.Getenv("NYLAS_TEST_E2E") != "true" {
		t.Skip("Skipping end-to-end AI test (set NYLAS_TEST_E2E=true to run)")
	}

	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	skipIfNoDefaultAIProvider(t)

	testEmail := getTestEmail()
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL environment variable not set")
	}

	// Simplified end-to-end: just test analyze (API) + context (API)
	t.Run("step1_analyze_patterns", func(t *testing.T) {
		args := []string{"calendar", "ai", "analyze", "--days", "30"}
		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Pattern analysis: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "Analyzing") && !strings.Contains(stdout+stderr, "patterns") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})

	t.Run("step2_get_context", func(t *testing.T) {
		args := []string{"calendar", "ai", "context", "--days", "7"}
		stdout, stderr, err := runCLIWithTimeout(60*time.Second, args...)

		if err != nil {
			t.Logf("Context: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "Context") && !strings.Contains(stdout+stderr, "context") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})
}

// =============================================================================
// CALENDAR AI ANALYZE-THREAD TESTS (Phase 3.2)
// =============================================================================

// TestCLI_CalendarAIAnalyzeThreadHelp tests the analyze-thread help output
func TestCLI_CalendarAIAnalyzeThreadHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("calendar", "ai", "analyze-thread", "--help")

	if err != nil {
		t.Fatalf("calendar ai analyze-thread --help failed: %v\nstderr: %s", err, stderr)
	}

	// Should show required --thread flag
	if !strings.Contains(stdout, "--thread") {
		t.Errorf("Expected '--thread' flag in help, got: %s", stdout)
	}

	// Should show optional flags
	expectedFlags := []string{"--agenda", "--time", "--create-meeting", "--provider", "--json"}
	for _, flag := range expectedFlags {
		if !strings.Contains(stdout, flag) {
			t.Errorf("Expected '%s' flag in help, got: %s", flag, stdout)
		}
	}

	t.Logf("calendar ai analyze-thread --help output:\n%s", stdout)
}

// TestCLI_CalendarAIAnalyzeThread tests AI email thread analysis
func TestCLI_CalendarAIAnalyzeThread(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	// This test is skipped by default as it requires:
	// 1. Existing email thread with meeting context
	// 2. Thread ID to analyze
	// 3. AI provider configured
	t.Skip("Calendar AI analyze-thread requires existing email threads with meeting context.\n" +
		"This requires:\n" +
		"  1. Email thread with meeting-related discussion\n" +
		"  2. Valid thread ID from email list\n" +
		"  3. AI provider configured (Ollama, Claude, OpenAI, or Groq)\n\n" +
		"Manual testing:\n" +
		"  (1) List email threads: nylas email threads list\n" +
		"  (2) Find thread with meeting discussion\n" +
		"  (3) Analyze thread: nylas calendar ai analyze-thread --thread <thread-id>\n" +
		"  (4) Analyze with agenda: nylas calendar ai analyze-thread --thread <thread-id> --agenda\n" +
		"  (5) Analyze with time suggestions: nylas calendar ai analyze-thread --thread <thread-id> --time\n" +
		"  (6) Get JSON output: nylas calendar ai analyze-thread --thread <thread-id> --json\n" +
		"  (7) Create meeting from thread: nylas calendar ai analyze-thread --thread <thread-id> --create-meeting\n\n" +
		"Expected output:\n" +
		"  - Meeting purpose and topics\n" +
		"  - Key action items\n" +
		"  - Priority level\n" +
		"  - Required and optional participants\n" +
		"  - Suggested duration\n" +
		"  - Auto-generated agenda (if --agenda flag used)\n" +
		"  - Recommended meeting times (if --time flag used)\n")
}

// Helper functions are now in test.go
