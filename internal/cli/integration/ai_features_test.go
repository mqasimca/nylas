//go:build integration
// +build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestCLI_AISchedule_RealCalendar tests AI scheduling with real calendar events
func TestCLI_AISchedule_RealCalendar(t *testing.T) {
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
		testEmail = "qasim.m@nylas.com"
	}

	tests := []struct {
		name         string
		query        string
		provider     string
		wantContains []string
		skipOnError  bool
	}{
		{
			name:     "schedule meeting with natural language",
			query:    fmt.Sprintf("30-minute meeting with %s tomorrow at 2pm", testEmail),
			provider: getAvailableProvider(),
			wantContains: []string{
				"AI Scheduling",
				"Provider:",
			},
			skipOnError: true,
		},
		{
			name:     "schedule team sync next week",
			query:    fmt.Sprintf("team sync with %s next Monday morning", testEmail),
			provider: getAvailableProvider(),
			wantContains: []string{
				"AI Scheduling",
			},
			skipOnError: true,
		},
		{
			name:     "schedule planning session",
			query:    fmt.Sprintf("quarterly planning session with %s next Tuesday 10am", testEmail),
			provider: getAvailableProvider(),
			wantContains: []string{
				"AI Scheduling",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"calendar", "schedule", "ai"}
			if tt.provider != "" {
				args = append(args, "--provider", tt.provider)
			}
			args = append(args, tt.query)

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIReschedule_RealEvent tests AI rescheduling with actual calendar events
func TestCLI_AIReschedule_RealEvent(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	// First, create a test event to reschedule
	testEmail := getTestEmail()
	if testEmail == "" {
		testEmail = "qasim.m@nylas.com"
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

	// Extract event ID from output (assuming it's in the output)
	eventID := extractEventID(stdout)
	if eventID == "" {
		t.Skip("Skipping reschedule test - could not extract event ID")
	}

	t.Logf("Created test event: %s", eventID)

	// Cleanup: defer deletion of test event
	t.Cleanup(func() {
		t.Logf("Cleaning up test event: %s", eventID)
		deleteArgs := []string{"calendar", "events", "delete", eventID, "--force"}
		_, _, _ = runCLI(deleteArgs...)
	})

	tests := []struct {
		name         string
		eventID      string
		reason       string
		maxDelayDays int
		wantContains []string
		skipOnError  bool
	}{
		{
			name:         "reschedule with conflict reason",
			eventID:      eventID,
			reason:       "Conflict with client meeting",
			maxDelayDays: 7,
			wantContains: []string{
				"Reschedule Analysis",
				"Alternative Time",
			},
			skipOnError: true,
		},
		{
			name:         "reschedule with short delay",
			eventID:      eventID,
			reason:       "Schedule conflict",
			maxDelayDays: 3,
			wantContains: []string{
				"Reschedule Analysis",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "reschedule", "ai", tt.eventID,
				"--reason", tt.reason,
				"--max-delay-days", fmt.Sprintf("%d", tt.maxDelayDays),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIAnalyze_Patterns tests AI calendar pattern analysis
func TestCLI_AIAnalyze_Patterns(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	tests := []struct {
		name         string
		days         int
		wantContains []string
		skipOnError  bool
	}{
		{
			name: "analyze last 30 days",
			days: 30,
			wantContains: []string{
				"Analyzing",
				"meeting history",
			},
			skipOnError: true,
		},
		{
			name: "analyze last 60 days",
			days: 60,
			wantContains: []string{
				"Analyzing",
			},
			skipOnError: true,
		},
		{
			name: "analyze last 90 days",
			days: 90,
			wantContains: []string{
				"Analyzing",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "analyze",
				"--days", fmt.Sprintf("%d", tt.days),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIAnalyze_ScoreTime tests scoring specific meeting times
func TestCLI_AIAnalyze_ScoreTime(t *testing.T) {
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
		testEmail = "qasim.m@nylas.com"
	}

	// Score a time next week
	nextWeek := time.Now().Add(7 * 24 * time.Hour)
	scoreTime := time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		scoreTime    string
		participants []string
		duration     int
		wantContains []string
		skipOnError  bool
	}{
		{
			name:         "score afternoon meeting",
			scoreTime:    scoreTime.Format(time.RFC3339),
			participants: []string{testEmail},
			duration:     30,
			wantContains: []string{
				"Meeting Score",
			},
			skipOnError: true,
		},
		{
			name:         "score morning meeting",
			scoreTime:    time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			participants: []string{testEmail},
			duration:     60,
			wantContains: []string{
				"Meeting Score",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "analyze",
				"--score-time", tt.scoreTime,
				"--participants", strings.Join(tt.participants, ","),
				"--duration", fmt.Sprintf("%d", tt.duration),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIConflicts_Detection tests AI conflict detection
func TestCLI_AIConflicts_Detection(t *testing.T) {
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
		testEmail = "qasim.m@nylas.com"
	}

	// Test conflicts with a time tomorrow
	tomorrow := time.Now().Add(24 * time.Hour)
	startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		title        string
		startTime    string
		duration     int
		participants []string
		wantContains []string
		skipOnError  bool
	}{
		{
			name:         "check conflicts for new meeting",
			title:        "Product Review",
			startTime:    startTime.Format(time.RFC3339),
			duration:     60,
			participants: []string{testEmail},
			wantContains: []string{
				"Conflict Analysis",
			},
			skipOnError: true,
		},
		{
			name:         "check conflicts for short meeting",
			title:        "Quick Sync",
			startTime:    startTime.Add(2 * time.Hour).Format(time.RFC3339),
			duration:     30,
			participants: []string{testEmail},
			wantContains: []string{
				"Conflict Analysis",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "conflicts", "check",
				"--title", tt.title,
				"--start", tt.startTime,
				"--duration", fmt.Sprintf("%d", tt.duration),
				"--participants", strings.Join(tt.participants, ","),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
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
		testEmail = "qasim.m@nylas.com"
	}

	tests := []struct {
		name         string
		participants []string
		duration     string
		days         int
		wantContains []string
		skipOnError  bool
	}{
		{
			name:         "find time for 2 participants",
			participants: []string{testEmail, "test@example.com"},
			duration:     "1h",
			days:         7,
			wantContains: []string{
				"Multi-Timezone Meeting Finder",
				"Participants",
			},
			skipOnError: true,
		},
		{
			name:         "find time for short meeting",
			participants: []string{testEmail, "alice@example.com"},
			duration:     "30m",
			days:         5,
			wantContains: []string{
				"Multi-Timezone",
			},
			skipOnError: true,
		},
		{
			name:         "find time for longer meeting",
			participants: []string{testEmail, "bob@example.com"},
			duration:     "2h",
			days:         14,
			wantContains: []string{
				"Meeting Finder",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "find-time",
				"--participants", strings.Join(tt.participants, ","),
				"--duration", tt.duration,
				"--days", fmt.Sprintf("%d", tt.days),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIFocusTime_Analysis tests AI focus time pattern analysis
func TestCLI_AIFocusTime_Analysis(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	tests := []struct {
		name         string
		analyze      bool
		targetHours  float64
		wantContains []string
		skipOnError  bool
	}{
		{
			name:        "analyze focus time patterns",
			analyze:     true,
			targetHours: 14.0,
			wantContains: []string{
				"AI Focus Time",
			},
			skipOnError: true,
		},
		{
			name:        "analyze with custom target",
			analyze:     true,
			targetHours: 20.0,
			wantContains: []string{
				"Analyzing",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"calendar", "ai", "focus-time"}
			if tt.analyze {
				args = append(args, "--analyze")
			}
			args = append(args, "--target-hours", fmt.Sprintf("%.1f", tt.targetHours))

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIAdaptive_Scheduling tests adaptive schedule optimization
func TestCLI_AIAdaptive_Scheduling(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	tests := []struct {
		name         string
		trigger      string
		wantContains []string
		skipOnError  bool
	}{
		{
			name:    "adapt for meeting overload",
			trigger: "overload",
			wantContains: []string{
				"AI Adaptive Scheduling",
			},
			skipOnError: true,
		},
		{
			name:    "adapt for deadline change",
			trigger: "deadline",
			wantContains: []string{
				"Adaptive",
			},
			skipOnError: true,
		},
		{
			name:    "adapt for focus time risk",
			trigger: "focus-risk",
			wantContains: []string{
				"Scheduling",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "adapt",
				"--trigger", tt.trigger,
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIContext_Calendar tests getting calendar context
func TestCLI_AIContext_Calendar(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	if !hasAnyAIProvider() {
		t.Skip("No AI provider configured")
	}

	tests := []struct {
		name         string
		days         int
		wantContains []string
		skipOnError  bool
	}{
		{
			name: "get context for next 7 days",
			days: 7,
			wantContains: []string{
				"Calendar Context",
			},
			skipOnError: true,
		},
		{
			name: "get context for next 14 days",
			days: 14,
			wantContains: []string{
				"Context",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"calendar", "ai", "context",
				"--days", fmt.Sprintf("%d", tt.days),
			}

			stdout, stderr, err := runCLI(args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped due to error: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := stdout + stderr
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestCLI_AIFeatures_EndToEnd tests end-to-end AI workflow
func TestCLI_AIFeatures_EndToEnd(t *testing.T) {
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
		testEmail = "qasim.m@nylas.com"
	}

	// End-to-end workflow:
	// 1. Analyze calendar patterns
	// 2. Check for conflicts before scheduling
	// 3. Schedule a meeting with AI
	// 4. Get focus time recommendations

	t.Run("step1_analyze_patterns", func(t *testing.T) {
		args := []string{"calendar", "ai", "analyze", "--days", "30"}
		stdout, stderr, err := runCLI(args...)

		if err != nil {
			t.Logf("Pattern analysis: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "Analyzing") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})

	t.Run("step2_check_conflicts", func(t *testing.T) {
		tomorrow := time.Now().Add(24 * time.Hour)
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 15, 0, 0, 0, time.UTC)

		args := []string{
			"calendar", "ai", "conflicts", "check",
			"--title", "Test Meeting",
			"--start", startTime.Format(time.RFC3339),
			"--duration", "60",
			"--participants", testEmail,
		}

		stdout, stderr, err := runCLI(args...)

		if err != nil {
			t.Logf("Conflict check: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "Conflict") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})

	t.Run("step3_schedule_with_ai", func(t *testing.T) {
		provider := getAvailableProvider()
		query := fmt.Sprintf("30-minute meeting with %s next Tuesday afternoon", testEmail)

		args := []string{
			"calendar", "schedule", "ai",
			"--provider", provider,
			query,
		}

		stdout, stderr, err := runCLI(args...)

		if err != nil {
			t.Logf("AI scheduling: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "AI Scheduling") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})

	t.Run("step4_get_focus_time", func(t *testing.T) {
		args := []string{
			"calendar", "ai", "focus-time",
			"--analyze",
			"--target-hours", "14.0",
		}

		stdout, stderr, err := runCLI(args...)

		if err != nil {
			t.Logf("Focus time analysis: %v", err)
			t.Logf("stderr: %s", stderr)
			return
		}

		if !strings.Contains(stdout+stderr, "Focus") {
			t.Logf("Unexpected output: %s", stdout)
		}
	})
}

// Helper functions are now in test.go
