//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AI config tests don't require API credentials since they're offline operations

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

func TestCLI_AIConfigGet_MissingArgs(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	_, _, err := runCLI("ai", "config", "get")

	if err == nil {
		t.Error("Expected error for missing key argument, got none")
	}
}
