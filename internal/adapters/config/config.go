// Package config provides configuration file management.
package config

import (
	"os"
	"path/filepath"

	"github.com/mqasimca/nylas/internal/domain"
	"gopkg.in/yaml.v3"
)

// FileStore implements ConfigStore using a YAML file.
type FileStore struct {
	path string
}

// NewFileStore creates a new FileStore.
func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

// NewDefaultFileStore creates a FileStore at the default location.
func NewDefaultFileStore() *FileStore {
	return NewFileStore(DefaultConfigPath())
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nylas", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nylas", "config.yaml")
}

// DefaultConfigDir returns the default config directory.
func DefaultConfigDir() string {
	return filepath.Dir(DefaultConfigPath())
}

// Load loads the configuration from the file.
func (f *FileStore) Load() (*domain.Config, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.DefaultConfig(), nil
		}
		return nil, err
	}

	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Apply defaults for missing fields
	if config.Region == "" {
		config.Region = "us"
	}
	if config.CallbackPort == 0 {
		config.CallbackPort = 8080
	}
	if config.WatchInterval == 0 {
		config.WatchInterval = 10
	}

	// Apply API defaults
	if config.API == nil {
		config.API = &domain.APIConfig{
			BaseURL:    "https://api.us.nylas.com",
			Timeout:    "90s",
			RateLimit:  10,
			RetryCount: 3,
		}
	} else {
		if config.API.BaseURL == "" {
			config.API.BaseURL = "https://api.us.nylas.com"
		}
		if config.API.Timeout == "" {
			config.API.Timeout = "90s"
		}
		if config.API.RateLimit == 0 {
			config.API.RateLimit = 10
		}
		if config.API.RetryCount == 0 {
			config.API.RetryCount = 3
		}
	}

	// Apply Output defaults
	if config.Output == nil {
		config.Output = &domain.OutputConfig{
			Format:   "table",
			Color:    "auto",
			Timezone: "",
		}
	} else {
		if config.Output.Format == "" {
			config.Output.Format = "table"
		}
		if config.Output.Color == "" {
			config.Output.Color = "auto"
		}
	}

	return &config, nil
}

// Save saves the configuration to the file.
func (f *FileStore) Save(config *domain.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(f.path, data, 0600)
}

// Path returns the path to the config file.
func (f *FileStore) Path() string {
	return f.path
}

// Exists returns true if the config file exists.
func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.path)
	return err == nil
}
