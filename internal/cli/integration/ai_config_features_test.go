//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

func TestCLI_AIConfigPrivacy(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Initialize with basic config
	configContent := `region: us
callback_port: 8080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		value    string
		validate func(*testing.T, string)
	}{
		{
			name:  "set privacy.allow_cloud_ai to false",
			key:   "privacy.allow_cloud_ai",
			value: "false",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "privacy.allow_cloud_ai", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "false" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "false")
				}
			},
		},
		{
			name:  "set privacy.data_retention to 90",
			key:   "privacy.data_retention",
			value: "90",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "privacy.data_retention", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "90" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "90")
				}
			},
		},
		{
			name:  "set privacy.local_storage_only to true",
			key:   "privacy.local_storage_only",
			value: "true",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "privacy.local_storage_only", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "true" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI("ai", "config", "set", tt.key, tt.value, "--config", configPath)

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			// Verify success message
			expectedMsg := "✓ Set " + tt.key + " = " + tt.value
			if !strings.Contains(stdout, expectedMsg) {
				t.Errorf("Expected output to contain %q\nGot: %s", expectedMsg, stdout)
			}

			// Run custom validation
			if tt.validate != nil {
				tt.validate(t, configPath)
			}
		})
	}
}

// Feature Toggles Tests

func TestCLI_AIConfigFeatures(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Initialize with basic config
	configContent := `region: us
callback_port: 8080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		value    string
		validate func(*testing.T, string)
	}{
		{
			name:  "set features.natural_language_scheduling to true",
			key:   "features.natural_language_scheduling",
			value: "true",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "features.natural_language_scheduling", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "true" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "true")
				}
			},
		},
		{
			name:  "set features.predictive_scheduling to false",
			key:   "features.predictive_scheduling",
			value: "false",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "features.predictive_scheduling", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "false" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "false")
				}
			},
		},
		{
			name:  "set features.focus_time_protection to true",
			key:   "features.focus_time_protection",
			value: "true",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "features.focus_time_protection", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "true" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "true")
				}
			},
		},
		{
			name:  "set features.conflict_resolution to true",
			key:   "features.conflict_resolution",
			value: "true",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "features.conflict_resolution", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "true" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "true")
				}
			},
		},
		{
			name:  "set features.email_context_analysis to false",
			key:   "features.email_context_analysis",
			value: "false",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "features.email_context_analysis", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "false" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI("ai", "config", "set", tt.key, tt.value, "--config", configPath)

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			// Verify success message
			expectedMsg := "✓ Set " + tt.key + " = " + tt.value
			if !strings.Contains(stdout, expectedMsg) {
				t.Errorf("Expected output to contain %q\nGot: %s", expectedMsg, stdout)
			}

			// Run custom validation
			if tt.validate != nil {
				tt.validate(t, configPath)
			}
		})
	}
}

// Data Management Commands Tests
