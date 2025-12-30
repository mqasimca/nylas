package calendar

import (
	"testing"
	"time"
)

func TestGetLocalTimeZone(t *testing.T) {
	tz := getLocalTimeZone()

	if tz == "" {
		t.Error("Expected non-empty timezone, got empty string")
	}

	// Verify it's a valid timezone
	_, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("getLocalTimeZone() returned invalid timezone %q: %v", tz, err)
	}
}

func TestValidateTimeZone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{
			name:    "valid IANA timezone",
			tz:      "America/New_York",
			wantErr: false,
		},
		{
			name:    "valid UTC",
			tz:      "UTC",
			wantErr: false,
		},
		{
			name:    "valid Europe timezone",
			tz:      "Europe/London",
			wantErr: false,
		},
		{
			name:    "valid Asia timezone",
			tz:      "Asia/Tokyo",
			wantErr: false,
		},
		{
			name:    "invalid timezone",
			tz:      "Invalid/Timezone",
			wantErr: true,
		},
		{
			name:    "timezone abbreviation (not IANA)",
			tz:      "PST",
			wantErr: true,
		},
		{
			name:    "empty timezone",
			tz:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeZone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTimeZone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestGetSystemTimeZone(t *testing.T) {
	tz := getSystemTimeZone()

	// Should return a valid IANA timezone or UTC
	if tz == "" {
		t.Error("Expected non-empty timezone")
	}

	// Verify it's loadable
	_, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("getSystemTimeZone() returned invalid timezone %q: %v", tz, err)
	}
}
