// Package config provides configuration management for go-nylas plugins.
// Plugins can load configuration from environment variables or config files.
package config

import (
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/domain"
)

// Config represents the Nylas CLI configuration accessible to plugins.
// Plugins receive configuration via environment variables set by the core CLI.
type Config struct {
	// API Configuration
	APIKey   string
	GrantID  string
	Region   string
	ClientID string

	// Application Configuration
	CallbackPort int
	DefaultGrant string

	// Working hours (if configured)
	WorkingHours *WorkingHoursConfig

	// Internal config reference (not exposed to plugins directly)
	internal *domain.Config
}

// WorkingHoursConfig represents working hours configuration.
type WorkingHoursConfig struct {
	Default   *DaySchedule
	Monday    *DaySchedule
	Tuesday   *DaySchedule
	Wednesday *DaySchedule
	Thursday  *DaySchedule
	Friday    *DaySchedule
	Saturday  *DaySchedule
	Sunday    *DaySchedule
	Weekend   *DaySchedule
}

// DaySchedule represents working hours for a specific day.
type DaySchedule struct {
	Enabled bool
	Start   string // HH:MM format
	End     string // HH:MM format
	Breaks  []BreakBlock
}

// BreakBlock represents a break period within working hours.
type BreakBlock struct {
	Name  string
	Start string
	End   string
	Type  string
}

// LoadFromEnv loads configuration from environment variables.
// The core CLI sets these environment variables when executing plugins:
//   - NYLAS_API_KEY: API key for authentication
//   - NYLAS_GRANT_ID: Grant ID for the current user
//   - NYLAS_REGION: API region (us, eu, etc.)
//   - NYLAS_CLIENT_ID: OAuth client ID
//   - NYLAS_CLI_VERSION: Core CLI version
func LoadFromEnv() (*Config, error) {
	apiKey := os.Getenv("NYLAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("NYLAS_API_KEY environment variable not set")
	}

	grantID := os.Getenv("NYLAS_GRANT_ID")
	if grantID == "" {
		return nil, fmt.Errorf("NYLAS_GRANT_ID environment variable not set")
	}

	region := os.Getenv("NYLAS_REGION")
	if region == "" {
		region = "us" // Default to US region
	}

	clientID := os.Getenv("NYLAS_CLIENT_ID")

	return &Config{
		APIKey:   apiKey,
		GrantID:  grantID,
		Region:   region,
		ClientID: clientID,
	}, nil
}

// LoadFromFile loads configuration from the standard Nylas CLI config file.
// This is useful for plugins that run independently of the core CLI.
func LoadFromFile() (*Config, error) {
	store := config.NewDefaultFileStore()
	internalCfg, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get API key from keyring
	apiKey := os.Getenv("NYLAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("NYLAS_API_KEY not set (required for authentication)")
	}

	// Convert internal config to public config
	cfg := &Config{
		APIKey:       apiKey,
		GrantID:      internalCfg.DefaultGrant,
		Region:       internalCfg.Region,
		CallbackPort: internalCfg.CallbackPort,
		DefaultGrant: internalCfg.DefaultGrant,
		internal:     internalCfg,
	}

	// Convert working hours if present
	if internalCfg.WorkingHours != nil {
		cfg.WorkingHours = convertWorkingHours(internalCfg.WorkingHours)
	}

	return cfg, nil
}

// GetGrantID returns the grant ID for the current configuration.
func (c *Config) GetGrantID() string {
	return c.GrantID
}

// GetAPIKey returns the API key for authentication.
func (c *Config) GetAPIKey() string {
	return c.APIKey
}

// GetRegion returns the API region.
func (c *Config) GetRegion() string {
	return c.Region
}

// GetClientID returns the OAuth client ID.
func (c *Config) GetClientID() string {
	return c.ClientID
}

// GetCallbackPort returns the OAuth callback port.
func (c *Config) GetCallbackPort() int {
	return c.CallbackPort
}

// GetScheduleForDay returns the working hours schedule for a given weekday.
func (w *WorkingHoursConfig) GetScheduleForDay(weekday string) *DaySchedule {
	if w == nil {
		return defaultWorkingHours()
	}

	// Check day-specific schedule first
	var daySchedule *DaySchedule
	switch weekday {
	case "monday":
		daySchedule = w.Monday
	case "tuesday":
		daySchedule = w.Tuesday
	case "wednesday":
		daySchedule = w.Wednesday
	case "thursday":
		daySchedule = w.Thursday
	case "friday":
		daySchedule = w.Friday
	case "saturday":
		daySchedule = w.Saturday
	case "sunday":
		daySchedule = w.Sunday
	}

	if daySchedule != nil {
		return daySchedule
	}

	// Check weekend schedule for Sat/Sun
	if (weekday == "saturday" || weekday == "sunday") && w.Weekend != nil {
		return w.Weekend
	}

	// Fall back to default
	if w.Default != nil {
		return w.Default
	}

	return defaultWorkingHours()
}

// Helper functions

func convertWorkingHours(internal *domain.WorkingHoursConfig) *WorkingHoursConfig {
	if internal == nil {
		return nil
	}

	return &WorkingHoursConfig{
		Default:   convertDaySchedule(internal.Default),
		Monday:    convertDaySchedule(internal.Monday),
		Tuesday:   convertDaySchedule(internal.Tuesday),
		Wednesday: convertDaySchedule(internal.Wednesday),
		Thursday:  convertDaySchedule(internal.Thursday),
		Friday:    convertDaySchedule(internal.Friday),
		Saturday:  convertDaySchedule(internal.Saturday),
		Sunday:    convertDaySchedule(internal.Sunday),
		Weekend:   convertDaySchedule(internal.Weekend),
	}
}

func convertDaySchedule(internal *domain.DaySchedule) *DaySchedule {
	if internal == nil {
		return nil
	}

	breaks := make([]BreakBlock, len(internal.Breaks))
	for i, b := range internal.Breaks {
		breaks[i] = BreakBlock{
			Name:  b.Name,
			Start: b.Start,
			End:   b.End,
			Type:  b.Type,
		}
	}

	return &DaySchedule{
		Enabled: internal.Enabled,
		Start:   internal.Start,
		End:     internal.End,
		Breaks:  breaks,
	}
}

func defaultWorkingHours() *DaySchedule {
	return &DaySchedule{
		Enabled: true,
		Start:   "09:00",
		End:     "17:00",
	}
}
