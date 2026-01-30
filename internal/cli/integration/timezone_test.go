//go:build integration

package integration

import (
	"strings"
	"testing"
)

// Timezone tests don't require API credentials since they're offline operations

func TestCLI_TimezoneConvert(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found - run 'go build -o bin/nylas ./cmd/nylas' first")
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "convert UTC to EST",
			args: []string{"timezone", "convert", "--from", "UTC", "--to", "America/New_York"},
			contains: []string{
				"Time Zone Conversion",
				"UTC",
				"America/New_York",
				"Offset:",
			},
		},
		{
			name: "convert with abbreviations",
			args: []string{"timezone", "convert", "--from", "PST", "--to", "EST"},
			contains: []string{
				"America/Los_Angeles",
				"America/New_York",
			},
		},
		{
			name: "convert with specific time",
			args: []string{"timezone", "convert", "--from", "UTC", "--to", "Asia/Tokyo", "--time", "2025-01-01T12:00:00Z"},
			contains: []string{
				"2025-01-01",
				"UTC",
				"Asia/Tokyo",
			},
		},
		{
			name: "convert JSON output",
			args: []string{"timezone", "convert", "--from", "UTC", "--to", "EST", "--json"},
			contains: []string{
				`"from"`,
				`"to"`,
				`"zone"`,
				`"time"`,
			},
		},
		{
			name:    "missing required from flag",
			args:    []string{"timezone", "convert", "--to", "EST"},
			wantErr: true,
		},
		{
			name:    "missing required to flag",
			args:    []string{"timezone", "convert", "--from", "PST"},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			args:    []string{"timezone", "convert", "--from", "UTC", "--to", "EST", "--time", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s, stderr: %s", stdout, stderr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_TimezoneFindMeeting(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "find meeting basic",
			args: []string{"timezone", "find-meeting", "--zones", "America/New_York,Europe/London"},
			contains: []string{
				"Meeting Time Finder",
				"Time Zones:",
				"Duration:",
			},
		},
		{
			name: "find meeting with all options",
			args: []string{"timezone", "find-meeting", "--zones", "PST,EST,IST", "--duration", "30m", "--start-hour", "10:00", "--end-hour", "16:00", "--exclude-weekends"},
			contains: []string{
				"30m",
				"10:00 - 16:00",
			},
		},
		{
			name: "find meeting JSON output",
			args: []string{"timezone", "find-meeting", "--zones", "UTC,EST", "--json"},
			contains: []string{
				`"time_zones"`,
			},
		},
		{
			name:    "missing zones flag",
			args:    []string{"timezone", "find-meeting"},
			wantErr: true,
		},
		{
			name:    "invalid duration",
			args:    []string{"timezone", "find-meeting", "--zones", "UTC", "--duration", "invalid"},
			wantErr: true,
		},
		{
			name:    "invalid working hours",
			args:    []string{"timezone", "find-meeting", "--zones", "UTC", "--start-hour", "25:00"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_TimezoneDST(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "DST for zone with transitions",
			args: []string{"timezone", "dst", "--zone", "America/New_York", "--year", "2026"},
			contains: []string{
				"DST Transitions",
				"America/New_York",
				"2026",
			},
		},
		{
			name: "DST for zone without transitions",
			args: []string{"timezone", "dst", "--zone", "America/Phoenix", "--year", "2026"},
			contains: []string{
				"No DST transitions found",
			},
		},
		{
			name: "DST with abbreviation",
			args: []string{"timezone", "dst", "--zone", "PST"},
			contains: []string{
				"America/Los_Angeles",
			},
		},
		{
			name: "DST JSON output",
			args: []string{"timezone", "dst", "--zone", "EST", "--json"},
			contains: []string{
				`"zone"`,
				`"year"`,
			},
		},
		{
			name:    "missing zone flag",
			args:    []string{"timezone", "dst"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_TimezoneList(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "list all zones",
			args: []string{"timezone", "list"},
			contains: []string{
				"IANA Time Zones",
				"Total:",
			},
		},
		{
			name: "list with filter",
			args: []string{"timezone", "list", "--filter", "America"},
			contains: []string{
				"America",
				"filtered",
			},
		},
		{
			name: "list JSON output",
			args: []string{"timezone", "list", "--json"},
			contains: []string{
				`"zones"`,
				`"count"`,
			},
		},
		{
			name: "list with non-matching filter",
			args: []string{"timezone", "list", "--filter", "NonExistentZone12345"},
			contains: []string{
				"No zones found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_TimezoneInfo(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name: "info with positional arg",
			args: []string{"timezone", "info", "America/New_York"},
			contains: []string{
				"Time Zone Information",
				"America/New_York",
				"Abbreviation:",
				"UTC Offset:",
			},
		},
		{
			name: "info with flag",
			args: []string{"timezone", "info", "--zone", "Europe/London"},
			contains: []string{
				"Europe/London",
			},
		},
		{
			name: "info with abbreviation",
			args: []string{"timezone", "info", "PST"},
			contains: []string{
				"America/Los_Angeles",
				"expanded from 'PST'",
			},
		},
		{
			name: "info with specific time",
			args: []string{"timezone", "info", "--zone", "UTC", "--time", "2025-01-01T12:00:00Z"},
			contains: []string{
				"UTC",
				"2025-01-01",
			},
		},
		{
			name: "info JSON output",
			args: []string{"timezone", "info", "--zone", "UTC", "--json"},
			contains: []string{
				`"zone"`,
				`"abbreviation"`,
				`"is_dst"`,
			},
		},
		{
			name:    "missing zone",
			args:    []string{"timezone", "info"},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			args:    []string{"timezone", "info", "--zone", "UTC", "--time", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCLI(tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. stdout: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}

func TestCLI_TimezoneHelp(t *testing.T) {
	if testBinary == "" {
		t.Skip("CLI binary not found")
	}

	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "timezone main help",
			args: []string{"timezone", "--help"},
			contains: []string{
				"time zones",
				"convert",
				"find-meeting",
				"dst",
				"list",
				"info",
			},
		},
		{
			name: "convert help",
			args: []string{"timezone", "convert", "--help"},
			contains: []string{
				"Convert",
				"--from",
				"--to",
			},
		},
		{
			name: "find-meeting help",
			args: []string{"timezone", "find-meeting", "--help"},
			contains: []string{
				"time slots",
				"--zones",
				"--duration",
			},
		},
		{
			name: "dst help",
			args: []string{"timezone", "dst", "--help"},
			contains: []string{
				"DST",
				"--zone",
				"--year",
			},
		},
		{
			name: "list help",
			args: []string{"timezone", "list", "--help"},
			contains: []string{
				"IANA time zones",
				"--filter",
			},
		},
		{
			name: "info help",
			args: []string{"timezone", "info", "--help"},
			contains: []string{
				"information",
				"--zone",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, _ := runCLI(tt.args...)

			for _, expected := range tt.contains {
				if !strings.Contains(stdout, expected) {
					t.Errorf("Expected help to contain %q\nGot: %s", expected, stdout)
				}
			}
		})
	}
}
