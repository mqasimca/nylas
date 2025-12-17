package domain

// Config represents the application configuration.
// Note: client_id is stored in keystore, not config file.
type Config struct {
	Region       string      `yaml:"region"`
	CallbackPort int         `yaml:"callback_port"`
	DefaultGrant string      `yaml:"default_grant"`
	Grants       []GrantInfo `yaml:"grants"`

	// OTP-specific settings
	CopyToClipboard bool `yaml:"copy_to_clipboard"`
	WatchInterval   int  `yaml:"watch_interval"`

	// TUI settings
	TUITheme string `yaml:"tui_theme,omitempty"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Region:          "us",
		CallbackPort:    8080,
		CopyToClipboard: true,
		WatchInterval:   10,
	}
}

// ConfigStatus represents the current configuration status.
type ConfigStatus struct {
	IsConfigured    bool   `json:"configured"`
	Region          string `json:"region"`
	ClientID        string `json:"client_id,omitempty"`
	HasAPIKey       bool   `json:"has_api_key"`
	HasClientSecret bool   `json:"has_client_secret"`
	SecretStore     string `json:"secret_store"`
	ConfigPath      string `json:"config_path"`
	GrantCount      int    `json:"grant_count"`
	DefaultGrant    string `json:"default_grant,omitempty"`
}
