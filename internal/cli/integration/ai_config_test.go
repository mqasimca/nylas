//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

func TestCLI_AIHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("ai", "--help")

	if err != nil {
		t.Fatalf("ai --help failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"Manage AI/LLM settings",
		"config",
		"AI configuration",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIConfigHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	stdout, stderr, err := runCLI("ai", "config", "--help")

	if err != nil {
		t.Fatalf("ai config --help failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"AI configuration",
		"show",
		"list",
		"get",
		"set",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
		}
	}
}

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

func TestCLI_AIConfigGet(t *testing.T) {
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
    model: llama3.1:8b
  claude:
    model: claude-3-5-sonnet-20241022
  groq:
    model: mixtral-8x7b-32768
  openrouter:
    model: anthropic/claude-3.5-sonnet
  fallback:
    enabled: true
    providers:
      - ollama
      - claude
      - openai
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name: "get default_provider",
			key:  "default_provider",
			want: "ollama",
		},
		{
			name: "get ollama.host",
			key:  "ollama.host",
			want: "http://localhost:11434",
		},
		{
			name: "get ollama.model",
			key:  "ollama.model",
			want: "llama3.1:8b",
		},
		{
			name: "get claude.model",
			key:  "claude.model",
			want: "claude-3-5-sonnet-20241022",
		},
		{
			name: "get fallback.enabled",
			key:  "fallback.enabled",
			want: "true",
		},
		{
			name: "get fallback.providers",
			key:  "fallback.providers",
			want: "ollama,claude,openai",
		},
		{
			name:    "get non-existent key",
			key:     "openai.model",
			wantErr: true,
		},
		{
			name: "get groq.model",
			key:  "groq.model",
			want: "mixtral-8x7b-32768",
		},
		{
			name: "get openrouter.model",
			key:  "openrouter.model",
			want: "anthropic/claude-3.5-sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI("ai", "config", "get", tt.key, "--config", configPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			output := strings.TrimSpace(stdout)
			if output != tt.want {
				t.Errorf("got %q, want %q", output, tt.want)
			}
		})
	}
}

func TestCLI_AIConfigSet(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		key      string
		value    string
		wantErr  bool
		validate func(t *testing.T, configPath string)
	}{
		{
			name:  "set default_provider",
			key:   "default_provider",
			value: "claude",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "default_provider", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "claude" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "claude")
				}
			},
		},
		{
			name:  "set ollama.host",
			key:   "ollama.host",
			value: "http://remote-ollama:11434",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "ollama.host", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "http://remote-ollama:11434" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "http://remote-ollama:11434")
				}
			},
		},
		{
			name:  "set ollama.model",
			key:   "ollama.model",
			value: "mistral:latest",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "ollama.model", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "mistral:latest" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "mistral:latest")
				}
			},
		},
		{
			name:  "set claude.model",
			key:   "claude.model",
			value: "claude-3-5-sonnet-20241022",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "claude.model", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "claude-3-5-sonnet-20241022" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "claude-3-5-sonnet-20241022")
				}
			},
		},
		{
			name:  "set openai.model",
			key:   "openai.model",
			value: "gpt-4o",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "openai.model", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "gpt-4o" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "gpt-4o")
				}
			},
		},
		{
			name:  "set groq.model",
			key:   "groq.model",
			value: "mixtral-8x7b-32768",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "groq.model", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "mixtral-8x7b-32768" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "mixtral-8x7b-32768")
				}
			},
		},
		{
			name:  "set fallback.enabled true",
			key:   "fallback.enabled",
			value: "true",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "fallback.enabled", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "true" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "true")
				}
			},
		},
		{
			name:  "set fallback.providers",
			key:   "fallback.providers",
			value: "ollama,claude,openai,groq",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "fallback.providers", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "ollama,claude,openai,groq" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "ollama,claude,openai,groq")
				}
			},
		},
		{
			name:  "set openrouter.api_key",
			key:   "openrouter.api_key",
			value: "${OPENROUTER_API_KEY}",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "openrouter.api_key", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "${OPENROUTER_API_KEY}" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "${OPENROUTER_API_KEY}")
				}
			},
		},
		{
			name:  "set openrouter.model",
			key:   "openrouter.model",
			value: "anthropic/claude-3.5-sonnet",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "openrouter.model", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "anthropic/claude-3.5-sonnet" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "anthropic/claude-3.5-sonnet")
				}
			},
		},
		{
			name:  "set default_provider to openrouter",
			key:   "default_provider",
			value: "openrouter",
			validate: func(t *testing.T, configPath string) {
				stdout, _, err := runCLI("ai", "config", "get", "default_provider", "--config", configPath)
				if err != nil {
					t.Fatalf("failed to verify: %v", err)
				}
				if strings.TrimSpace(stdout) != "openrouter" {
					t.Errorf("got %q, want %q", strings.TrimSpace(stdout), "openrouter")
				}
			},
		},
		{
			name:    "set invalid provider",
			key:     "default_provider",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "set unknown key",
			key:     "unknown.key",
			value:   "value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Initialize with basic config
			configContent := `region: us
callback_port: 8080
`
			if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			stdout, stderr, err := runCLI("ai", "config", "set", tt.key, tt.value, "--config", configPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			// Verify success message
			expectedMsg := "✓ Set " + tt.key + " = " + tt.value
			if !strings.Contains(stdout, expectedMsg) {
				t.Errorf("Expected output to contain %q\nGot: %s", expectedMsg, stdout)
			}

			// Verify configuration was saved
			if !strings.Contains(stdout, "Configuration saved to:") {
				t.Errorf("Expected output to contain save confirmation\nGot: %s", stdout)
			}

			// Run custom validation if provided
			if tt.validate != nil {
				tt.validate(t, configPath)
			}
		})
	}
}

func TestCLI_AIConfigSetAndGet_MultipleValues(t *testing.T) {
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

	// Set multiple values
	settings := map[string]string{
		"default_provider":   "ollama",
		"ollama.host":        "http://localhost:11434",
		"ollama.model":       "llama3.1:8b",
		"claude.model":       "claude-3-5-sonnet-20241022",
		"fallback.enabled":   "true",
		"fallback.providers": "ollama,claude",
	}

	for key, value := range settings {
		_, stderr, err := runCLI("ai", "config", "set", key, value, "--config", configPath)
		if err != nil {
			t.Fatalf("Failed to set %s=%s: %v\nstderr: %s", key, value, err, stderr)
		}
	}

	// Verify all values
	for key, expectedValue := range settings {
		stdout, stderr, err := runCLI("ai", "config", "get", key, "--config", configPath)
		if err != nil {
			t.Fatalf("Failed to get %s: %v\nstderr: %s", key, err, stderr)
		}

		got := strings.TrimSpace(stdout)
		if got != expectedValue {
			t.Errorf("For key %s: got %q, want %q", key, got, expectedValue)
		}
	}

	// Verify list shows all values
	stdout, stderr, err := runCLI("ai", "config", "list", "--config", configPath)
	if err != nil {
		t.Fatalf("ai config list failed: %v\nstderr: %s", err, stderr)
	}

	expectedInList := []string{
		"default_provider: ollama",
		"host: http://localhost:11434",
		"model: llama3.1:8b",
		"model: claude-3-5-sonnet-20241022",
		"enabled: true",
		"providers: [ollama, claude]",
	}

	for _, expected := range expectedInList {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected list output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIConfigGet_MissingArgs(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	_, _, err := runCLI("ai", "config", "get")

	if err == nil {
		t.Error("Expected error for missing key argument, got none")
	}
}

func TestCLI_AIConfigSet_MissingArgs(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing both key and value",
			args: []string{"ai", "config", "set"},
		},
		{
			name: "missing value",
			args: []string{"ai", "config", "set", "default_provider"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := runCLI(tt.args...)

			if err == nil {
				t.Error("Expected error for missing arguments, got none")
			}
		})
	}
}

// Privacy Settings Tests

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

func TestCLI_AIUsageCommand(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Test usage command (should work even with no data)
	stdout, stderr, err := runCLI("ai", "usage")

	if err != nil {
		t.Fatalf("ai usage failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"AI Usage for",
		"Total Requests:",
		"Ollama:",
		"Claude:",
		"OpenAI:",
		"Groq:",
		"OpenRouter:",
		"Total Tokens:",
		"Estimated Cost:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

func TestCLI_AIUsageJSON(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Test usage command with JSON output
	stdout, stderr, err := runCLI("ai", "usage", "--json")

	if err != nil {
		t.Fatalf("ai usage --json failed: %v\nstderr: %s", err, stderr)
	}

	// Should be valid JSON
	if !strings.Contains(stdout, "{") || !strings.Contains(stdout, "}") {
		t.Errorf("Expected JSON output, got: %s", stdout)
	}

	// Should contain expected fields
	expectedFields := []string{
		"month",
		"total_requests",
		"estimated_cost",
	}

	for _, field := range expectedFields {
		if !strings.Contains(stdout, field) {
			t.Errorf("Expected JSON to contain field %q\nGot: %s", field, stdout)
		}
	}
}

func TestCLI_AISetBudget(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Set a budget
	stdout, stderr, err := runCLI("ai", "set-budget", "--monthly", "50")

	if err != nil {
		t.Fatalf("ai set-budget failed: %v\nstderr: %s", err, stderr)
	}

	expectedStrings := []string{
		"✓ Monthly budget set to $50.00",
		"Alert threshold: 80%",
		"Budget applies to:",
		"Claude",
		"OpenAI",
		"Groq",
		"OpenRouter",
		"Ollama (local) usage is free",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}

	// Verify with show-budget
	stdout, stderr, err = runCLI("ai", "show-budget")

	if err != nil {
		t.Fatalf("ai show-budget failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Monthly Limit:       $50.00") {
		t.Errorf("Expected show-budget to show $50.00 limit\nGot: %s", stdout)
	}
}

func TestCLI_AIClearData(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	// Clear data with --force flag (no confirmation prompt)
	stdout, stderr, err := runCLI("ai", "clear-data", "--force")

	if err != nil {
		t.Fatalf("ai clear-data failed: %v\nstderr: %s", err, stderr)
	}

	// Should succeed (even if no data exists)
	expectedStrings := []string{
		"AI data cleared",
		"Learned scheduling patterns",
		"Usage statistics",
		"Cached responses",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
		}
	}
}

// Privacy and Features Config Show Test

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
