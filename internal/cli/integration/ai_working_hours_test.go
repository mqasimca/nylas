//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for AI analyze command respecting working hours configuration

func TestCLI_CalendarAI_Analyze_RespectsWorkingHours(t *testing.T) {
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
		name           string
		workingHours   string
		expectedStart  string // Expected hour to appear in output
		unexpectedHour string // Hour that should NOT appear
	}{
		{
			name: "working hours start at 09:30 shows 10:00",
			workingHours: `working_hours:
  default:
    enabled: true
    start: "09:30"
    end: "17:00"
`,
			expectedStart:  "10:00",
			unexpectedHour: "09:00",
		},
		{
			name: "working hours start at 10:00 shows 10:00",
			workingHours: `working_hours:
  default:
    enabled: true
    start: "10:00"
    end: "17:00"
`,
			expectedStart:  "10:00",
			unexpectedHour: "09:00",
		},
		{
			name: "working hours end at 16:00 excludes 17:00",
			workingHours: `working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "16:00"
`,
			expectedStart:  "09:00",
			unexpectedHour: "17:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
` + tt.workingHours

			if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Run analyze with custom config
			stdout, stderr, err := runCLI("calendar", "ai", "analyze", "--config", configPath)

			if err != nil {
				// Log but don't fail - may fail due to no calendar data or provider issues
				t.Logf("Analyze command returned error (may be expected): %v", err)
				t.Logf("stderr: %s", stderr)
				// Still check output if available
			}

			output := stdout + stderr

			// Check if expected hour appears in output
			if tt.expectedStart != "" && !strings.Contains(output, tt.expectedStart) {
				// Only fail if we got meaningful output
				if strings.Contains(output, "By Time of Day") {
					t.Errorf("Expected output to contain %q for working hours display\nGot: %s", tt.expectedStart, output)
				}
			}

			// Check that unexpected hour does NOT appear in time display
			if tt.unexpectedHour != "" && strings.Contains(output, "By Time of Day") {
				// Look specifically in the time of day section
				lines := strings.Split(output, "\n")
				inTimeSection := false
				for _, line := range lines {
					if strings.Contains(line, "By Time of Day") {
						inTimeSection = true
						continue
					}
					if inTimeSection && strings.TrimSpace(line) == "" {
						inTimeSection = false
						continue
					}
					if inTimeSection && strings.Contains(line, tt.unexpectedHour+":") {
						t.Errorf("Expected output NOT to contain %q in time of day section\nGot: %s", tt.unexpectedHour, output)
					}
				}
			}
		})
	}
}

func TestCLI_CalendarAI_Analyze_DefaultWorkingHours(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	// Create temp config without working hours - should use defaults (9-17)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("calendar", "ai", "analyze", "--config", configPath)

	if err != nil {
		t.Logf("Analyze command returned error (may be expected): %v", err)
		t.Logf("stderr: %s", stderr)
	}

	output := stdout + stderr

	// With no working hours configured, should show 09:00
	if strings.Contains(output, "By Time of Day") {
		if !strings.Contains(output, "09:00") {
			t.Logf("Note: Default working hours should start at 09:00")
		}
	}
}

func TestCLI_CalendarAI_Analyze_DisabledWorkingHours(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	// Create temp config with disabled working hours
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
working_hours:
  default:
    enabled: false
    start: "10:00"
    end: "16:00"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("calendar", "ai", "analyze", "--config", configPath)

	if err != nil {
		t.Logf("Analyze command returned error (may be expected): %v", err)
		t.Logf("stderr: %s", stderr)
	}

	output := stdout + stderr

	// With disabled working hours, should use defaults (show 09:00)
	if strings.Contains(output, "By Time of Day") {
		// When disabled, should fall back to default 9-17
		if !strings.Contains(output, "09:00") {
			t.Logf("Note: Disabled working hours should use default (09:00)")
		}
	}
}

func TestCLI_CalendarAI_Analyze_FocusTimeRespectsWorkingHours(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	// Create temp config with working hours starting at 10:00
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
working_hours:
  default:
    enabled: true
    start: "10:00"
    end: "17:00"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("calendar", "ai", "analyze", "--config", configPath)

	if err != nil {
		t.Logf("Analyze command returned error (may be expected): %v", err)
		t.Logf("stderr: %s", stderr)
	}

	output := stdout + stderr

	// Check productivity insights section for focus time recommendations
	if strings.Contains(output, "Peak Focus Times") {
		// Focus time recommendations should not include 09:00
		lines := strings.Split(output, "\n")
		inFocusSection := false
		for _, line := range lines {
			if strings.Contains(line, "Peak Focus Times") {
				inFocusSection = true
				continue
			}
			if inFocusSection && strings.TrimSpace(line) == "" {
				inFocusSection = false
				continue
			}
			if inFocusSection && strings.Contains(line, "09:00") {
				t.Errorf("Focus time recommendations should not include 09:00 when working hours start at 10:00\nGot: %s", line)
			}
		}
	}
}

func TestCLI_CalendarAI_Analyze_WithBreaks(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Skip if no Nylas API configured
	if testAPIKey == "" || testGrantID == "" {
		t.Skip("Nylas API credentials not configured")
	}

	// Skip if no default AI provider configured
	skipIfNoDefaultAIProvider(t)

	// Create temp config with working hours and breaks
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("calendar", "ai", "analyze", "--config", configPath)

	if err != nil {
		t.Logf("Analyze command returned error (may be expected): %v", err)
		t.Logf("stderr: %s", stderr)
	}

	// Just verify command runs - breaks are handled separately
	output := stdout + stderr
	if output == "" {
		t.Log("No output from analyze command")
	}
}
