//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

func TestCLI_AIConfigShow(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		setup    func(*testing.T) string // Returns config file path
		cleanup  func(string)
		wantErr  bool
		contains []string
	}{
		{
			name: "show with existing config",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: llama3.1:8b
`
				if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
				return configPath
			},
			contains: []string{
				"AI Configuration:",
				"default_provider: ollama",
				"ollama:",
				"host: http://localhost:11434",
				"model: llama3.1:8b",
			},
		},
		{
			name: "show with openrouter config",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `region: us
callback_port: 8080
ai:
  default_provider: openrouter
  openrouter:
    api_key: ${OPENROUTER_API_KEY}
    model: anthropic/claude-3.5-sonnet
`
				if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
				return configPath
			},
			contains: []string{
				"AI Configuration:",
				"default_provider: openrouter",
				"openrouter:",
				"api_key: ${OPENROUTER_API_KEY}",
				"model: anthropic/claude-3.5-sonnet",
			},
		},
		{
			name: "show with no AI config",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				configContent := `region: us
callback_port: 8080
`
				if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
				return configPath
			},
			contains: []string{
				"No AI configuration found",
				"To configure AI, use:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setup(t)
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup(configPath)
				}
			}()

			stdout, stderr, err := runCLI("ai", "config", "show", "--config", configPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_AIConfigList(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  ollama:
    host: http://localhost:11434
    model: mistral:latest
  claude:
    model: claude-3-5-sonnet-20241022
  fallback:
    enabled: true
    providers:
      - ollama
      - claude
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("ai", "config", "list", "--config", configPath)

	if err != nil {
		t.Fatalf("ai config list failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"AI Configuration:",
		"default_provider: ollama",
		"Ollama:",
		"host: http://localhost:11434",
		"model: mistral:latest",
		"Claude:",
		"model: claude-3-5-sonnet-20241022",
		"Fallback:",
		"enabled: true",
		"providers: [ollama, claude]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIConfigShow_WithPrivacyAndFeatures(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with privacy and features
	configContent := `region: us
callback_port: 8080
ai:
  default_provider: ollama
  privacy:
    allow_cloud_ai: false
    data_retention: 90
    local_storage_only: true
  features:
    natural_language_scheduling: true
    predictive_scheduling: true
    focus_time_protection: true
    conflict_resolution: true
    email_context_analysis: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	stdout, stderr, err := runCLI("ai", "config", "show", "--config", configPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	expectedStrings := []string{
		"AI Configuration:",
		"default_provider: ollama",
		"privacy:",
		"allow_cloud_ai: false",
		"data_retention: 90",
		"local_storage_only: true",
		"features:",
		"natural_language_scheduling: true",
		"predictive_scheduling: true",
		"focus_time_protection: true",
		"conflict_resolution: true",
		"email_context_analysis: false",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}
