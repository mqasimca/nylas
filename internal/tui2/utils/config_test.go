package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Theme", config.Theme, "k9s"},
		{"AnimationsEnabled", config.AnimationsEnabled, true},
		{"SplashDurationSec", config.SplashDurationSec, 3},
		{"ShowStatusBar", config.ShowStatusBar, true},
		{"ShowFooter", config.ShowFooter, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".config", "nylas", "tui.json")
	if configPath != expectedPath {
		t.Errorf("GetConfigPath() = %q, want %q", configPath, expectedPath)
	}

	// Verify config directory was created
	configDir := filepath.Join(tmpDir, ".config", "nylas")
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("config path is not a directory")
	}

	// Verify permissions
	if info.Mode().Perm() != 0755 {
		t.Errorf("config directory permissions = %o, want 0755", info.Mode().Perm())
	}
}

func TestLoadConfig_FileDoesNotExist(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	config, err := LoadConfig()

	// Should not error, should return defaults
	if err != nil {
		t.Errorf("LoadConfig() with non-existent file should not error, got: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	// Should be default config
	if config.Theme != "k9s" {
		t.Errorf("expected default theme 'k9s', got %q", config.Theme)
	}
	if !config.AnimationsEnabled {
		t.Error("expected default AnimationsEnabled true")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	// Create custom config
	customConfig := &TUIConfig{
		Theme:             "custom-theme",
		AnimationsEnabled: false,
		SplashDurationSec: 5,
		ShowStatusBar:     false,
		ShowFooter:        false,
	}

	// Save config
	err := SaveConfig(customConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify loaded config matches saved config
	if loadedConfig.Theme != customConfig.Theme {
		t.Errorf("Theme = %q, want %q", loadedConfig.Theme, customConfig.Theme)
	}
	if loadedConfig.AnimationsEnabled != customConfig.AnimationsEnabled {
		t.Errorf("AnimationsEnabled = %v, want %v", loadedConfig.AnimationsEnabled, customConfig.AnimationsEnabled)
	}
	if loadedConfig.SplashDurationSec != customConfig.SplashDurationSec {
		t.Errorf("SplashDurationSec = %d, want %d", loadedConfig.SplashDurationSec, customConfig.SplashDurationSec)
	}
	if loadedConfig.ShowStatusBar != customConfig.ShowStatusBar {
		t.Errorf("ShowStatusBar = %v, want %v", loadedConfig.ShowStatusBar, customConfig.ShowStatusBar)
	}
	if loadedConfig.ShowFooter != customConfig.ShowFooter {
		t.Errorf("ShowFooter = %v, want %v", loadedConfig.ShowFooter, customConfig.ShowFooter)
	}
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	config := DefaultConfig()

	// Save config (directory doesn't exist yet)
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file was created
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestSaveConfig_JSONFormatting(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	config := DefaultConfig()

	// Save config
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Read raw file content
	configPath, _ := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Verify it's valid JSON
	var parsed TUIConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("config file is not valid JSON: %v", err)
	}

	// Verify it's pretty-printed (should contain newlines and spaces)
	content := string(data)
	if len(content) < 50 {
		t.Error("config file seems too short, might not be pretty-printed")
	}
	// Pretty-printed JSON should have newlines
	hasNewlines := false
	for _, c := range content {
		if c == '\n' {
			hasNewlines = true
			break
		}
	}
	if !hasNewlines {
		t.Error("config file should be pretty-printed with newlines")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "nylas")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write invalid JSON
	configPath := filepath.Join(configDir, "tui.json")
	if err := os.WriteFile(configPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	// Load config should return default config on error
	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() with invalid JSON should return error")
	}

	// Should still return default config
	if config == nil {
		t.Fatal("expected non-nil config even with error")
	}
	if config.Theme != "k9s" {
		t.Errorf("expected default theme on error, got %q", config.Theme)
	}
}

func TestLoadConfig_FilePermissionError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	// Create config directory and file
	configDir := filepath.Join(tmpDir, ".config", "nylas")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "tui.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0000); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	defer func() { _ = os.Chmod(configPath, 0644) }() // Cleanup

	// Load config should handle permission error gracefully
	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() with unreadable file should return error")
	}

	// Should return default config
	if config == nil {
		t.Fatal("expected non-nil config even with error")
	}
}

func TestTUIConfig_JSONSerialization(t *testing.T) {
	original := &TUIConfig{
		Theme:             "nord",
		AnimationsEnabled: false,
		SplashDurationSec: 10,
		ShowStatusBar:     false,
		ShowFooter:        true,
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Deserialize from JSON
	var parsed TUIConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Verify all fields match
	if parsed.Theme != original.Theme {
		t.Errorf("Theme = %q, want %q", parsed.Theme, original.Theme)
	}
	if parsed.AnimationsEnabled != original.AnimationsEnabled {
		t.Errorf("AnimationsEnabled = %v, want %v", parsed.AnimationsEnabled, original.AnimationsEnabled)
	}
	if parsed.SplashDurationSec != original.SplashDurationSec {
		t.Errorf("SplashDurationSec = %d, want %d", parsed.SplashDurationSec, original.SplashDurationSec)
	}
	if parsed.ShowStatusBar != original.ShowStatusBar {
		t.Errorf("ShowStatusBar = %v, want %v", parsed.ShowStatusBar, original.ShowStatusBar)
	}
	if parsed.ShowFooter != original.ShowFooter {
		t.Errorf("ShowFooter = %v, want %v", parsed.ShowFooter, original.ShowFooter)
	}
}

func TestConfigWorkflow(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("HOME", tmpDir)

	// Step 1: Load config (doesn't exist yet)
	config1, err := LoadConfig()
	if err != nil {
		t.Fatalf("first LoadConfig() error = %v", err)
	}
	if config1.Theme != "k9s" {
		t.Error("first load should return defaults")
	}

	// Step 2: Modify and save
	config1.Theme = "custom"
	config1.AnimationsEnabled = false
	if err := SaveConfig(config1); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Step 3: Load again
	config2, err := LoadConfig()
	if err != nil {
		t.Fatalf("second LoadConfig() error = %v", err)
	}
	if config2.Theme != "custom" {
		t.Error("second load should return saved config")
	}
	if config2.AnimationsEnabled {
		t.Error("second load should have AnimationsEnabled=false")
	}

	// Step 4: Modify and save again
	config2.SplashDurationSec = 7
	if err := SaveConfig(config2); err != nil {
		t.Fatalf("second SaveConfig() error = %v", err)
	}

	// Step 5: Load final time
	config3, err := LoadConfig()
	if err != nil {
		t.Fatalf("third LoadConfig() error = %v", err)
	}
	if config3.SplashDurationSec != 7 {
		t.Error("third load should have SplashDurationSec=7")
	}
	if config3.Theme != "custom" {
		t.Error("third load should preserve Theme")
	}
}
