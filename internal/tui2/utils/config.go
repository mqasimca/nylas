// Package utils provides utility functions for TUI.
package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TUIConfig represents TUI configuration.
type TUIConfig struct {
	Theme             string `json:"theme"`
	AnimationsEnabled bool   `json:"animations_enabled"`
	SplashDurationSec int    `json:"splash_duration_sec"`
	ShowStatusBar     bool   `json:"show_status_bar"`
	ShowFooter        bool   `json:"show_footer"`
}

// DefaultConfig returns default configuration.
func DefaultConfig() *TUIConfig {
	return &TUIConfig{
		Theme:             "k9s",
		AnimationsEnabled: true,
		SplashDurationSec: 3,
		ShowStatusBar:     true,
		ShowFooter:        true,
	}
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".config", "nylas")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "tui.json"), nil
}

// LoadConfig loads configuration from file.
func LoadConfig() (*TUIConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var config TUIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	return &config, nil
}

// SaveConfig saves configuration to file.
func SaveConfig(config *TUIConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
