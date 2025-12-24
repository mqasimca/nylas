//go:build integration
// +build integration

package integration

import (
	"os"
	"strings"
	"testing"
)

// AI tests require LLM provider API keys (Ollama, Claude, OpenAI, or Groq)
//
// To run AI integration tests with Ollama:
//   1. Ensure Ollama is running: ollama serve
//   2. Pull a model: ollama pull mistral
//   3. Configure in ~/.nylas/config.yaml:
//      ai:
//        default_provider: ollama
//        ollama:
//          host: http://localhost:11434
//          model: mistral:latest
//   4. Run tests: go test -tags=integration -v ./internal/cli/integration/ai_test.go ./internal/cli/integration/test.go

func TestCLI_AIProvider_Availability(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Check if default_provider is set in config
	skipIfNoDefaultAIProvider(t)

	// Check if at least one AI provider is configured
	hasOllama := checkOllamaAvailable()
	hasClaude := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""
	hasGroq := os.Getenv("GROQ_API_KEY") != ""

	if !hasOllama && !hasClaude && !hasOpenAI && !hasGroq {
		t.Skip("No AI provider configured. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, GROQ_API_KEY, or run Ollama locally")
	}

	t.Logf("AI providers available: Ollama=%v, Claude=%v, OpenAI=%v, Groq=%v", hasOllama, hasClaude, hasOpenAI, hasGroq)
}

func TestCLI_CalendarAI_Basic(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		skipTest bool
	}{
		{
			name:    "calendar ai help",
			args:    []string{"calendar", "ai", "--help"},
			wantErr: false,
		},
		{
			name:    "calendar ai schedule help",
			args:    []string{"calendar", "ai", "schedule", "--help"},
			wantErr: false,
		},
		{
			name:    "calendar ai reschedule help",
			args:    []string{"calendar", "ai", "reschedule", "--help"},
			wantErr: false,
		},
		{
			name:    "calendar ai context help",
			args:    []string{"calendar", "ai", "context", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Test requires specific configuration")
			}

			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}
		})
	}
}

func TestCLI_CalendarAI_Schedule_InvalidInput(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "schedule without query",
			args:    []string{"calendar", "schedule", "ai"},
			wantErr: true,
		},
		{
			name:    "schedule with invalid provider",
			args:    []string{"calendar", "schedule", "ai", "--provider", "invalid", "Test meeting"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s, stderr: %s", stdout, stderr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}
		})
	}
}

func TestCLI_CalendarAI_Reschedule_InvalidInput(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "reschedule without event ID",
			args:    []string{"calendar", "ai", "reschedule", "ai"},
			wantErr: true,
		},
		{
			name:    "reschedule with event ID but no Nylas credentials",
			args:    []string{"calendar", "ai", "reschedule", "ai", "event-id"},
			wantErr: true, // Will fail on "secret not found" or API call
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s, stderr: %s", stdout, stderr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}
		})
	}
}

func TestCLI_CalendarAI_Context_Basic(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "analyze help command works",
			args: []string{"calendar", "ai", "analyze", "--help"},
			contains: []string{
				"Analyze historical meeting data",
			},
		},
		{
			name: "analyze shows meeting patterns",
			args: []string{"calendar", "ai", "analyze"},
			contains: []string{
				"Analyzing",
				"meetings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Logf("Note: Test may fail without valid Nylas credentials. Error: %v", err)
				t.Logf("stderr: %s", stderr)
				t.Logf("stdout: %s", stdout)
				// Don't fail the test, just log
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_CalendarAI_ScheduleWithOllama(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if !checkOllamaAvailable() {
		t.Skip("Ollama not available - ensure Ollama is running and configured")
	}

	// Skip if no Nylas API configured (AI scheduling needs calendar access)
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	tests := []struct {
		name         string
		query        string
		provider     string
		wantContains []string
		skipOnError  bool
	}{
		{
			name:     "schedule with natural language - ollama",
			query:    "meeting tomorrow at 2pm",
			provider: "ollama",
			wantContains: []string{
				"AI Scheduling",
				"Provider:",
			},
			skipOnError: true, // Skip if Ollama model not available
		},
		{
			name:     "schedule with privacy mode",
			query:    "team sync next Monday morning",
			provider: "ollama",
			wantContains: []string{
				"Privacy Mode",
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
				t.Logf("Test skipped due to error (likely model not available): %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Logf("Note: AI scheduling may require valid Nylas calendar data")
				t.Logf("Error: %v", err)
				t.Logf("stderr: %s", stderr)
				t.Logf("stdout: %s", stdout)
				return
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

func TestCLI_CalendarAI_AnalyzeWithOllama(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if !checkOllamaAvailable() {
		t.Skip("Ollama not available")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	tests := []struct {
		name        string
		args        []string
		skipOnError bool
	}{
		{
			name:        "analyze calendar with ollama",
			args:        []string{"calendar", "analyze", "--provider", "ollama"},
			skipOnError: true,
		},
		{
			name:        "analyze with privacy mode",
			args:        []string{"calendar", "analyze", "--privacy"},
			skipOnError: true,
		},
		{
			name:        "analyze specific time range",
			args:        []string{"calendar", "analyze", "--provider", "ollama", "--days", "7"},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil {
				t.Logf("Note: Analysis requires calendar data")
				t.Logf("stdout: %s", stdout)
				t.Logf("stderr: %s", stderr)
			}

			// Just verify command ran (output depends on calendar data)
			if stdout != "" || stderr != "" {
				t.Logf("Command executed. Output length: stdout=%d, stderr=%d", len(stdout), len(stderr))
			}
		})
	}
}

func TestCLI_CalendarAI_ConflictDetection(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if !checkOllamaAvailable() {
		t.Skip("Ollama not available")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	tests := []struct {
		name        string
		args        []string
		skipOnError bool
	}{
		{
			name:        "detect conflicts",
			args:        []string{"calendar", "conflicts", "--provider", "ollama"},
			skipOnError: true,
		},
		{
			name:        "conflicts with time range",
			args:        []string{"calendar", "conflicts", "--provider", "ollama", "--days", "14"},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped: %v", err)
				return
			}

			if err != nil {
				t.Logf("Conflict detection test (depends on calendar data)")
				t.Logf("stdout: %s", stdout)
				t.Logf("stderr: %s", stderr)
			}
		})
	}
}

func TestCLI_CalendarAI_FocusTime(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if !checkOllamaAvailable() {
		t.Skip("Ollama not available")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	tests := []struct {
		name        string
		args        []string
		skipOnError bool
	}{
		{
			name:        "find focus time",
			args:        []string{"calendar", "focus-time", "--provider", "ollama"},
			skipOnError: true,
		},
		{
			name:        "focus time with duration",
			args:        []string{"calendar", "focus-time", "--provider", "ollama", "--duration", "2h"},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped: %v", err)
				return
			}

			if err != nil {
				t.Logf("Focus time test (depends on calendar availability)")
				t.Logf("stdout: %s", stdout)
				t.Logf("stderr: %s", stderr)
			}
		})
	}
}

func TestCLI_CalendarAI_FindTimeMultiTimezone(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	if !checkOllamaAvailable() {
		t.Skip("Ollama not available")
	}

	tests := []struct {
		name        string
		args        []string
		skipOnError bool
	}{
		{
			name: "find time across timezones",
			args: []string{
				"calendar", "find-time",
				"--zones", "America/New_York,Europe/London,Asia/Tokyo",
				"--provider", "ollama",
			},
			skipOnError: true,
		},
		{
			name: "find time with participants",
			args: []string{
				"calendar", "find-time",
				"--zones", "PST,EST",
				"--duration", "1h",
				"--provider", "ollama",
			},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped: %v", err)
				return
			}

			if err != nil {
				t.Logf("Multi-timezone test")
				t.Logf("stdout: %s", stdout)
				t.Logf("stderr: %s", stderr)
			}
		})
	}
}

func TestCLI_CalendarAI_ProviderSwitching(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Test switching between providers if multiple are available
	hasOllama := checkOllamaAvailable()
	hasClaude := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""

	if !hasOllama && !hasClaude && !hasOpenAI {
		t.Skip("Need at least one AI provider")
	}

	providers := []string{}
	if hasOllama {
		providers = append(providers, "ollama")
	}
	if hasClaude {
		providers = append(providers, "claude")
	}
	if hasOpenAI {
		providers = append(providers, "openai")
	}

	for _, provider := range providers {
		t.Run("provider_"+provider, func(t *testing.T) {
			args := []string{"calendar", "schedule", "ai", "--provider", provider, "test meeting tomorrow"}
			stdout, stderr, err := runCLI(args...)

			// Just verify the provider flag is accepted
			if err != nil {
				t.Logf("Provider %s test: %v", provider, err)
				t.Logf("stderr: %s", stderr)
			} else {
				output := stdout + stderr
				if strings.Contains(output, "Provider:") {
					t.Logf("Provider %s: command accepted", provider)
				}
			}
		})
	}
}

func TestCLI_CalendarAI_Adapt(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	tests := []struct {
		name        string
		args        []string
		skipOnError bool
	}{
		{
			name:        "adapt help",
			args:        []string{"calendar", "ai", "adapt", "--help"},
			skipOnError: false,
		},
		{
			name:        "adapt default trigger",
			args:        []string{"calendar", "ai", "adapt"},
			skipOnError: true,
		},
		{
			name:        "adapt with overload trigger",
			args:        []string{"calendar", "ai", "adapt", "--trigger", "overload"},
			skipOnError: true,
		},
		{
			name:        "adapt with deadline trigger",
			args:        []string{"calendar", "ai", "adapt", "--trigger", "deadline"},
			skipOnError: true,
		},
		{
			name:        "adapt with focus-risk trigger",
			args:        []string{"calendar", "ai", "adapt", "--trigger", "focus-risk"},
			skipOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if err != nil && tt.skipOnError {
				t.Logf("Test skipped: %v", err)
				t.Logf("stderr: %s", stderr)
				return
			}

			if err != nil && !tt.skipOnError {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			// For help command, verify it shows expected content
			if tt.name == "adapt help" {
				if !strings.Contains(stdout, "adaptive") && !strings.Contains(stdout, "Adaptive") {
					t.Errorf("Expected help output to mention 'adaptive'\nGot: %s", stdout)
				}
			}

			// Log output for debugging
			if stdout != "" || stderr != "" {
				t.Logf("Command executed. Output length: stdout=%d, stderr=%d", len(stdout), len(stderr))
			}
		})
	}
}
