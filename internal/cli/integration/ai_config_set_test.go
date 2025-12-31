//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

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
