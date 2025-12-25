package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/domain"
)

func TestGetConfigValue(t *testing.T) {
	tests := []struct {
		name      string
		ai        *domain.AIConfig
		key       string
		want      string
		wantError bool
	}{
		{
			name: "get default_provider",
			ai: &domain.AIConfig{
				DefaultProvider: "ollama",
			},
			key:       "default_provider",
			want:      "ollama",
			wantError: false,
		},
		{
			name: "get ollama.host",
			ai: &domain.AIConfig{
				Ollama: &domain.OllamaConfig{
					Host:  "http://localhost:11434",
					Model: "llama3.1:8b",
				},
			},
			key:       "ollama.host",
			want:      "http://localhost:11434",
			wantError: false,
		},
		{
			name: "get ollama.model",
			ai: &domain.AIConfig{
				Ollama: &domain.OllamaConfig{
					Host:  "http://localhost:11434",
					Model: "mistral:latest",
				},
			},
			key:       "ollama.model",
			want:      "mistral:latest",
			wantError: false,
		},
		{
			name: "get claude.model",
			ai: &domain.AIConfig{
				Claude: &domain.ClaudeConfig{
					Model: "claude-3-5-sonnet-20241022",
				},
			},
			key:       "claude.model",
			want:      "claude-3-5-sonnet-20241022",
			wantError: false,
		},
		{
			name: "get openai.model",
			ai: &domain.AIConfig{
				OpenAI: &domain.OpenAIConfig{
					Model: "gpt-4o",
				},
			},
			key:       "openai.model",
			want:      "gpt-4o",
			wantError: false,
		},
		{
			name: "get groq.model",
			ai: &domain.AIConfig{
				Groq: &domain.GroqConfig{
					Model: "mixtral-8x7b-32768",
				},
			},
			key:       "groq.model",
			want:      "mixtral-8x7b-32768",
			wantError: false,
		},
		{
			name: "get fallback.enabled",
			ai: &domain.AIConfig{
				Fallback: &domain.AIFallbackConfig{
					Enabled:   true,
					Providers: []string{"ollama", "claude"},
				},
			},
			key:       "fallback.enabled",
			want:      "true",
			wantError: false,
		},
		{
			name: "get fallback.providers",
			ai: &domain.AIConfig{
				Fallback: &domain.AIFallbackConfig{
					Enabled:   true,
					Providers: []string{"ollama", "claude", "openai"},
				},
			},
			key:       "fallback.providers",
			want:      "ollama,claude,openai",
			wantError: false,
		},
		{
			name: "get non-existent provider",
			ai: &domain.AIConfig{
				DefaultProvider: "ollama",
			},
			key:       "claude.model",
			wantError: true,
		},
		{
			name: "get unknown key",
			ai: &domain.AIConfig{
				DefaultProvider: "ollama",
			},
			key:       "invalid.key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigValue(tt.ai, tt.key)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	tests := []struct {
		name      string
		ai        *domain.AIConfig
		key       string
		value     string
		wantError bool
		validate  func(*testing.T, *domain.AIConfig)
	}{
		{
			name:  "set default_provider to ollama",
			ai:    &domain.AIConfig{},
			key:   "default_provider",
			value: "ollama",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.DefaultProvider != "ollama" {
					t.Errorf("DefaultProvider = %q, want %q", ai.DefaultProvider, "ollama")
				}
			},
		},
		{
			name:  "set default_provider to claude",
			ai:    &domain.AIConfig{},
			key:   "default_provider",
			value: "claude",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.DefaultProvider != "claude" {
					t.Errorf("DefaultProvider = %q, want %q", ai.DefaultProvider, "claude")
				}
			},
		},
		{
			name:      "set default_provider to invalid",
			ai:        &domain.AIConfig{},
			key:       "default_provider",
			value:     "invalid",
			wantError: true,
		},
		{
			name:  "set ollama.host",
			ai:    &domain.AIConfig{},
			key:   "ollama.host",
			value: "http://localhost:11434",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Ollama == nil {
					t.Fatal("Ollama config is nil")
				}
				if ai.Ollama.Host != "http://localhost:11434" {
					t.Errorf("Ollama.Host = %q, want %q", ai.Ollama.Host, "http://localhost:11434")
				}
			},
		},
		{
			name:  "set ollama.model",
			ai:    &domain.AIConfig{},
			key:   "ollama.model",
			value: "llama3.1:8b",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Ollama == nil {
					t.Fatal("Ollama config is nil")
				}
				if ai.Ollama.Model != "llama3.1:8b" {
					t.Errorf("Ollama.Model = %q, want %q", ai.Ollama.Model, "llama3.1:8b")
				}
			},
		},
		{
			name:  "set claude.model",
			ai:    &domain.AIConfig{},
			key:   "claude.model",
			value: "claude-3-5-sonnet-20241022",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Claude == nil {
					t.Fatal("Claude config is nil")
				}
				if ai.Claude.Model != "claude-3-5-sonnet-20241022" {
					t.Errorf("Claude.Model = %q, want %q", ai.Claude.Model, "claude-3-5-sonnet-20241022")
				}
			},
		},
		{
			name:  "set claude.api_key",
			ai:    &domain.AIConfig{},
			key:   "claude.api_key",
			value: "${ANTHROPIC_API_KEY}",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Claude == nil {
					t.Fatal("Claude config is nil")
				}
				if ai.Claude.APIKey != "${ANTHROPIC_API_KEY}" {
					t.Errorf("Claude.APIKey = %q, want %q", ai.Claude.APIKey, "${ANTHROPIC_API_KEY}")
				}
			},
		},
		{
			name:  "set openai.model",
			ai:    &domain.AIConfig{},
			key:   "openai.model",
			value: "gpt-4o",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.OpenAI == nil {
					t.Fatal("OpenAI config is nil")
				}
				if ai.OpenAI.Model != "gpt-4o" {
					t.Errorf("OpenAI.Model = %q, want %q", ai.OpenAI.Model, "gpt-4o")
				}
			},
		},
		{
			name:  "set groq.model",
			ai:    &domain.AIConfig{},
			key:   "groq.model",
			value: "mixtral-8x7b-32768",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Groq == nil {
					t.Fatal("Groq config is nil")
				}
				if ai.Groq.Model != "mixtral-8x7b-32768" {
					t.Errorf("Groq.Model = %q, want %q", ai.Groq.Model, "mixtral-8x7b-32768")
				}
			},
		},
		{
			name:  "set fallback.enabled true",
			ai:    &domain.AIConfig{},
			key:   "fallback.enabled",
			value: "true",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Fallback == nil {
					t.Fatal("Fallback config is nil")
				}
				if !ai.Fallback.Enabled {
					t.Error("Fallback.Enabled = false, want true")
				}
			},
		},
		{
			name:  "set fallback.enabled false",
			ai:    &domain.AIConfig{},
			key:   "fallback.enabled",
			value: "false",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Fallback == nil {
					t.Fatal("Fallback config is nil")
				}
				if ai.Fallback.Enabled {
					t.Error("Fallback.Enabled = true, want false")
				}
			},
		},
		{
			name:  "set fallback.providers",
			ai:    &domain.AIConfig{},
			key:   "fallback.providers",
			value: "ollama,claude,openai",
			validate: func(t *testing.T, ai *domain.AIConfig) {
				if ai.Fallback == nil {
					t.Fatal("Fallback config is nil")
				}
				want := []string{"ollama", "claude", "openai"}
				if len(ai.Fallback.Providers) != len(want) {
					t.Errorf("Fallback.Providers length = %d, want %d", len(ai.Fallback.Providers), len(want))
					return
				}
				for i, p := range ai.Fallback.Providers {
					if p != want[i] {
						t.Errorf("Fallback.Providers[%d] = %q, want %q", i, p, want[i])
					}
				}
			},
		},
		{
			name:      "set unknown key",
			ai:        &domain.AIConfig{},
			key:       "unknown.key",
			value:     "value",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setConfigValue(tt.ai, tt.key, tt.value)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tt.ai)
			}
		})
	}
}

func TestConfigSetAndGet_Integration(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	store := config.NewFileStore(configPath)
	cfg := domain.DefaultConfig()

	// Initialize AI config
	cfg.AI = &domain.AIConfig{
		DefaultProvider: "ollama",
		Ollama: &domain.OllamaConfig{
			Host:  "http://localhost:11434",
			Model: "llama3.1:8b",
		},
	}

	// Save config
	if err := store.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load and verify
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.AI == nil {
		t.Fatal("AI config is nil after load")
	}

	if loaded.AI.DefaultProvider != "ollama" {
		t.Errorf("DefaultProvider = %q, want %q", loaded.AI.DefaultProvider, "ollama")
	}

	if loaded.AI.Ollama == nil {
		t.Fatal("Ollama config is nil")
	}

	if loaded.AI.Ollama.Host != "http://localhost:11434" {
		t.Errorf("Ollama.Host = %q, want %q", loaded.AI.Ollama.Host, "http://localhost:11434")
	}

	if loaded.AI.Ollama.Model != "llama3.1:8b" {
		t.Errorf("Ollama.Model = %q, want %q", loaded.AI.Ollama.Model, "llama3.1:8b")
	}

	// Modify and save
	_ = setConfigValue(loaded.AI, "claude.model", "claude-3-5-sonnet-20241022")
	if err := store.Save(loaded); err != nil {
		t.Fatalf("failed to save modified config: %v", err)
	}

	// Load again and verify Claude was added
	loaded2, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load config again: %v", err)
	}

	if loaded2.AI.Claude == nil {
		t.Fatal("Claude config is nil after modification")
	}

	if loaded2.AI.Claude.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Claude.Model = %q, want %q", loaded2.AI.Claude.Model, "claude-3-5-sonnet-20241022")
	}

	// Verify original Ollama config is still there
	if loaded2.AI.Ollama == nil {
		t.Fatal("Ollama config was lost after modification")
	}

	if loaded2.AI.Ollama.Model != "llama3.1:8b" {
		t.Errorf("Ollama.Model changed unexpectedly to %q", loaded2.AI.Ollama.Model)
	}
}

func TestNewConfigCmd(t *testing.T) {
	cmd := newConfigCmd()

	if cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check subcommands
	expectedCommands := []string{"show", "list", "get", "set"}
	commands := cmd.Commands()

	if len(commands) != len(expectedCommands) {
		t.Errorf("got %d subcommands, want %d", len(commands), len(expectedCommands))
	}

	for _, expected := range expectedCommands {
		found := false
		for _, c := range commands {
			if c.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestNewAICmd(t *testing.T) {
	cmd := NewAICmd()

	if cmd.Use != "ai" {
		t.Errorf("Use = %q, want %q", cmd.Use, "ai")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check that config subcommand is registered
	commands := cmd.Commands()
	found := false
	for _, c := range commands {
		if c.Name() == "config" {
			found = true
			break
		}
	}

	if !found {
		t.Error("config subcommand not found")
	}
}

func TestConfigFile_NotExists(t *testing.T) {
	// Test behavior when config file doesn't exist
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.yaml")

	store := config.NewFileStore(configPath)

	// Load should return default config
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// AI should be nil for default config
	if cfg.AI != nil {
		t.Error("AI config should be nil for default config")
	}

	// Initialize and save
	cfg.AI = &domain.AIConfig{
		DefaultProvider: "ollama",
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "typical_api_key",
			key:  "sk-proj-abcdefghijklmnopqrstuvwxyz123456",
			want: "sk-proj-***...***3456",
		},
		{
			name: "short_key_masked_completely",
			key:  "short",
			want: "***",
		},
		{
			name: "exactly_12_chars",
			key:  "123456789012",
			want: "***",
		},
		{
			name: "13_chars_shows_partial",
			key:  "1234567890123",
			want: "12345678***...***0123",
		},
		{
			name: "openai_style_key",
			key:  "sk-1234567890abcdefghijklmnop",
			want: "sk-12345***...***mnop",
		},
		{
			name: "anthropic_style_key",
			key:  "sk-ant-api03-abcdefghijklmnop",
			want: "sk-ant-a***...***mnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.key)
			if got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
