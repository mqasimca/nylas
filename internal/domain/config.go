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

	// Working hours settings
	WorkingHours *WorkingHoursConfig `yaml:"working_hours,omitempty"`

	// AI settings
	AI *AIConfig `yaml:"ai,omitempty"`
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

// WorkingHoursConfig represents working hours configuration.
type WorkingHoursConfig struct {
	Default   *DaySchedule `yaml:"default,omitempty"`
	Monday    *DaySchedule `yaml:"monday,omitempty"`
	Tuesday   *DaySchedule `yaml:"tuesday,omitempty"`
	Wednesday *DaySchedule `yaml:"wednesday,omitempty"`
	Thursday  *DaySchedule `yaml:"thursday,omitempty"`
	Friday    *DaySchedule `yaml:"friday,omitempty"`
	Saturday  *DaySchedule `yaml:"saturday,omitempty"`
	Sunday    *DaySchedule `yaml:"sunday,omitempty"`
	Weekend   *DaySchedule `yaml:"weekend,omitempty"` // Applies to Sat/Sun if specific days not set
}

// DaySchedule represents working hours for a specific day.
type DaySchedule struct {
	Enabled bool         `yaml:"enabled"`          // Whether working hours apply
	Start   string       `yaml:"start,omitempty"`  // Start time (HH:MM format)
	End     string       `yaml:"end,omitempty"`    // End time (HH:MM format)
	Breaks  []BreakBlock `yaml:"breaks,omitempty"` // Break periods (lunch, coffee, etc.)
}

// BreakBlock represents a break period within working hours.
type BreakBlock struct {
	Name  string `yaml:"name"`           // Break name (e.g., "Lunch", "Coffee Break")
	Start string `yaml:"start"`          // Start time (HH:MM format)
	End   string `yaml:"end"`            // End time (HH:MM format)
	Type  string `yaml:"type,omitempty"` // Optional type: "lunch", "coffee", "custom"
}

// GetScheduleForDay returns the schedule for a given weekday.
// Checks day-specific, weekend, then default in order of precedence.
func (w *WorkingHoursConfig) GetScheduleForDay(weekday string) *DaySchedule {
	if w == nil {
		return DefaultWorkingHours()
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

	return DefaultWorkingHours()
}

// DefaultWorkingHours returns standard 9-5 working hours.
func DefaultWorkingHours() *DaySchedule {
	return &DaySchedule{
		Enabled: true,
		Start:   "09:00",
		End:     "17:00",
	}
}
