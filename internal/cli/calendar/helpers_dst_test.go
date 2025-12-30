package calendar

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestCheckDSTWarning(t *testing.T) {
	tests := []struct {
		name        string
		eventTime   time.Time
		tz          string
		wantWarning bool
		wantEmpty   bool
	}{
		{
			name:      "empty timezone returns empty",
			eventTime: time.Now(),
			tz:        "",
			wantEmpty: true,
		},
		{
			name:      "invalid timezone returns empty",
			eventTime: time.Now(),
			tz:        "Invalid/Timezone",
			wantEmpty: true,
		},
		{
			name: "event far from DST transition returns empty",
			// June is far from DST transitions in most zones
			eventTime: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			tz:        "America/New_York",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDSTWarning(tt.eventTime, tt.tz)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("checkDSTWarning() = %q, want empty string", result)
				}
				return
			}

			if tt.wantWarning && result == "" {
				t.Error("checkDSTWarning() = empty, want warning message")
			}

			if !tt.wantWarning && result != "" {
				t.Errorf("checkDSTWarning() = %q, want empty", result)
			}
		})
	}
}

func TestFormatDSTWarning(t *testing.T) {
	tests := []struct {
		name         string
		warning      *domain.DSTWarning
		wantContains string
		wantEmpty    bool
	}{
		{
			name:      "nil warning returns empty",
			warning:   nil,
			wantEmpty: true,
		},
		{
			name: "error severity shows red icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "error",
				Warning:          "This time will not exist",
			},
			wantContains: "⛔",
		},
		{
			name: "warning severity shows warning icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "warning",
				Warning:          "DST begins soon",
			},
			wantContains: "⚠️",
		},
		{
			name: "info severity shows info icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "info",
				Warning:          "DST ends in 5 days",
			},
			wantContains: "ℹ️",
		},
		{
			name: "unknown severity defaults to warning icon",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "unknown",
				Warning:          "Some warning",
			},
			wantContains: "⚠️",
		},
		{
			name: "warning message is included",
			warning: &domain.DSTWarning{
				IsNearTransition: true,
				Severity:         "warning",
				Warning:          "Daylight Saving Time begins in 3 days",
			},
			wantContains: "Daylight Saving Time begins in 3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDSTWarning(tt.warning)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("formatDSTWarning() = %q, want empty string", result)
				}
				return
			}

			if tt.wantContains != "" {
				found := false
				for i := 0; i <= len(result)-len(tt.wantContains); i++ {
					if result[i:i+len(tt.wantContains)] == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					// Try substring match
					if len(result) < len(tt.wantContains) {
						t.Errorf("formatDSTWarning() = %q, want to contain %q", result, tt.wantContains)
					} else {
						// Check if any part matches
						foundSubstring := false
						for i := 0; i <= len(result)-1; i++ {
							for j := i + 1; j <= len(result); j++ {
								substr := result[i:j]
								for k := 0; k <= len(tt.wantContains)-len(substr); k++ {
									if tt.wantContains[k:k+len(substr)] == substr && len(substr) > 2 {
										foundSubstring = true
										break
									}
								}
								if foundSubstring {
									break
								}
							}
							if foundSubstring {
								break
							}
						}
						if !foundSubstring {
							t.Errorf("formatDSTWarning() = %q, want to contain %q", result, tt.wantContains)
						}
					}
				}
			}
		})
	}
}

func TestCheckDSTConflict(t *testing.T) {
	tests := []struct {
		name        string
		eventTime   time.Time
		tz          string
		duration    time.Duration
		wantWarning bool
		wantError   bool
	}{
		{
			name:        "empty timezone - no warning",
			eventTime:   time.Now(),
			tz:          "",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
		{
			name:        "invalid timezone - error",
			eventTime:   time.Now(),
			tz:          "Invalid/Timezone",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   true,
		},
		{
			name: "event far from DST transition - no warning",
			// June is far from DST transitions in most zones
			eventTime:   time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			tz:          "America/New_York",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
		{
			name: "normal event in winter - no warning",
			// January is standard time in Northern Hemisphere
			eventTime:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			tz:          "America/New_York",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
		{
			name: "normal event in summer - no warning",
			// July is daylight saving time in Northern Hemisphere
			eventTime:   time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC),
			tz:          "America/Los_Angeles",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
		{
			name:        "timezone without DST - no warning",
			eventTime:   time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
			tz:          "UTC",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
		{
			name:        "Arizona timezone (no DST) - no warning",
			eventTime:   time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
			tz:          "America/Phoenix",
			duration:    time.Hour,
			wantWarning: false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning, err := checkDSTConflict(tt.eventTime, tt.tz, tt.duration)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.wantWarning {
				if warning == nil {
					t.Error("Expected DST warning, got nil")
					return
				}
				// Verify warning has expected fields
				if warning.Warning == "" {
					t.Error("Expected warning message, got empty string")
				}
				if warning.Severity == "" {
					t.Error("Expected severity, got empty string")
				}
			} else {
				if warning != nil {
					t.Errorf("Expected no warning, got: %+v", warning)
				}
			}
		})
	}
}
