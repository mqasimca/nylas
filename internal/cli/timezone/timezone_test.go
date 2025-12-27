package timezone

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommand executes a command and captures output
func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	cmd.SetArgs(args)
	err = cmd.Execute()

	// #nosec G104 -- test cleanup, errors don't affect test validity
	wOut.Close()
	// #nosec G104 -- test cleanup, errors don't affect test validity
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	_, _ = bufOut.ReadFrom(rOut)
	_, _ = bufErr.ReadFrom(rErr)

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return bufOut.String(), bufErr.String(), err
}

func TestNewTimezoneCmd(t *testing.T) {
	cmd := NewTimezoneCmd()

	if cmd.Use != "timezone" {
		t.Errorf("Expected Use to be 'timezone', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Check that all subcommands are registered
	expectedCommands := []string{"convert", "find-meeting", "dst", "list", "info"}
	commands := cmd.Commands()

	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedCommands), len(commands))
	}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestConvertCmd_Help(t *testing.T) {
	cmd := newConvertCmd()

	if cmd.Use != "convert" {
		t.Errorf("Expected Use to be 'convert', got '%s'", cmd.Use)
	}

	// Check required flags
	fromFlag := cmd.Flags().Lookup("from")
	if fromFlag == nil {
		t.Error("Expected --from flag to exist")
	}

	toFlag := cmd.Flags().Lookup("to")
	if toFlag == nil {
		t.Error("Expected --to flag to exist")
	}

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("Expected --json flag to exist")
	}
}

func TestConvertCmd_MissingRequiredFlags(t *testing.T) {
	cmd := newConvertCmd()
	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when required flags are missing")
	}
}

func TestConvertCmd_ValidConversion(t *testing.T) {
	cmd := newConvertCmd()
	stdout, _, err := executeCommand(cmd, "--from", "UTC", "--to", "America/New_York")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Time Zone Conversion") {
		t.Error("Expected output to contain 'Time Zone Conversion'")
	}

	if !strings.Contains(stdout, "UTC") {
		t.Error("Expected output to contain 'UTC'")
	}

	if !strings.Contains(stdout, "America/New_York") {
		t.Error("Expected output to contain 'America/New_York'")
	}
}

func TestConvertCmd_WithAbbreviations(t *testing.T) {
	cmd := newConvertCmd()
	stdout, _, err := executeCommand(cmd, "--from", "PST", "--to", "EST")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that abbreviations were expanded
	if !strings.Contains(stdout, "America/Los_Angeles") {
		t.Error("Expected PST to be expanded to America/Los_Angeles")
	}

	if !strings.Contains(stdout, "America/New_York") {
		t.Error("Expected EST to be expanded to America/New_York")
	}
}

func TestConvertCmd_WithSpecificTime(t *testing.T) {
	cmd := newConvertCmd()
	testTime := "2025-01-01T12:00:00Z"

	stdout, _, err := executeCommand(cmd,
		"--from", "UTC",
		"--to", "America/Los_Angeles",
		"--time", testTime)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "2025-01-01") {
		t.Error("Expected output to contain the specified date")
	}
}

func TestConvertCmd_JSONOutput(t *testing.T) {
	cmd := newConvertCmd()
	stdout, _, err := executeCommand(cmd,
		"--from", "UTC",
		"--to", "America/New_York",
		"--json")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, `"from"`) {
		t.Error("Expected JSON output to contain 'from' field")
	}

	if !strings.Contains(stdout, `"to"`) {
		t.Error("Expected JSON output to contain 'to' field")
	}

	if !strings.Contains(stdout, `"zone"`) {
		t.Error("Expected JSON output to contain 'zone' field")
	}
}

func TestConvertCmd_InvalidTimeFormat(t *testing.T) {
	cmd := newConvertCmd()
	_, _, err := executeCommand(cmd,
		"--from", "UTC",
		"--to", "America/New_York",
		"--time", "invalid-time")

	if err == nil {
		t.Error("Expected error for invalid time format")
	}
}

func TestConvertCmd_InvalidTimeZone(t *testing.T) {
	cmd := newConvertCmd()
	_, _, err := executeCommand(cmd,
		"--from", "Invalid/Zone",
		"--to", "America/New_York")

	if err == nil {
		t.Error("Expected error for invalid time zone")
	}
}

func TestFindMeetingCmd_Help(t *testing.T) {
	cmd := newFindMeetingCmd()

	if cmd.Use != "find-meeting" {
		t.Errorf("Expected Use to be 'find-meeting', got '%s'", cmd.Use)
	}

	// Check required flags
	zonesFlag := cmd.Flags().Lookup("zones")
	if zonesFlag == nil {
		t.Error("Expected --zones flag to exist")
	}

	durationFlag := cmd.Flags().Lookup("duration")
	if durationFlag == nil {
		t.Error("Expected --duration flag to exist")
	}
}

func TestFindMeetingCmd_MissingZones(t *testing.T) {
	cmd := newFindMeetingCmd()
	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when --zones flag is missing")
	}
}

func TestFindMeetingCmd_ValidRequest(t *testing.T) {
	cmd := newFindMeetingCmd()
	stdout, _, err := executeCommand(cmd,
		"--zones", "America/New_York,Europe/London")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Meeting Time Finder") {
		t.Error("Expected output to contain 'Meeting Time Finder'")
	}

	if !strings.Contains(stdout, "Time Zones:") {
		t.Error("Expected output to contain 'Time Zones:'")
	}
}

func TestFindMeetingCmd_WithAllOptions(t *testing.T) {
	cmd := newFindMeetingCmd()
	stdout, _, err := executeCommand(cmd,
		"--zones", "PST,EST,IST",
		"--duration", "30m",
		"--start-hour", "10:00",
		"--end-hour", "16:00",
		"--exclude-weekends")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "30m") {
		t.Error("Expected output to contain duration")
	}

	if !strings.Contains(stdout, "10:00 - 16:00") {
		t.Error("Expected output to contain working hours")
	}

	if !strings.Contains(stdout, "Weekends") {
		t.Error("Expected output to mention weekend exclusion")
	}
}

func TestFindMeetingCmd_InvalidDuration(t *testing.T) {
	cmd := newFindMeetingCmd()
	_, _, err := executeCommand(cmd,
		"--zones", "UTC",
		"--duration", "invalid")

	if err == nil {
		t.Error("Expected error for invalid duration format")
	}
}

func TestDSTCmd_Help(t *testing.T) {
	cmd := newDSTCmd()

	if cmd.Use != "dst" {
		t.Errorf("Expected Use to be 'dst', got '%s'", cmd.Use)
	}

	zoneFlag := cmd.Flags().Lookup("zone")
	if zoneFlag == nil {
		t.Error("Expected --zone flag to exist")
	}

	yearFlag := cmd.Flags().Lookup("year")
	if yearFlag == nil {
		t.Error("Expected --year flag to exist")
	}
}

func TestDSTCmd_WithDSTZone(t *testing.T) {
	cmd := newDSTCmd()
	stdout, _, err := executeCommand(cmd,
		"--zone", "America/New_York",
		"--year", "2026")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "DST Transitions") {
		t.Error("Expected output to contain 'DST Transitions'")
	}

	if !strings.Contains(stdout, "America/New_York") {
		t.Error("Expected output to contain zone name")
	}
}

func TestDSTCmd_WithNonDSTZone(t *testing.T) {
	cmd := newDSTCmd()
	stdout, _, err := executeCommand(cmd,
		"--zone", "America/Phoenix",
		"--year", "2026")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "No DST transitions found") {
		t.Error("Expected output to indicate no DST transitions")
	}
}

func TestDSTCmd_JSONOutput(t *testing.T) {
	cmd := newDSTCmd()
	stdout, _, err := executeCommand(cmd,
		"--zone", "America/New_York",
		"--year", "2026",
		"--json")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, `"zone"`) {
		t.Error("Expected JSON output to contain 'zone' field")
	}

	if !strings.Contains(stdout, `"year"`) {
		t.Error("Expected JSON output to contain 'year' field")
	}
}

func TestListCmd_Help(t *testing.T) {
	cmd := newListCmd()

	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got '%s'", cmd.Use)
	}

	filterFlag := cmd.Flags().Lookup("filter")
	if filterFlag == nil {
		t.Error("Expected --filter flag to exist")
	}
}

func TestListCmd_AllZones(t *testing.T) {
	cmd := newListCmd()
	stdout, _, err := executeCommand(cmd)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "IANA Time Zones") {
		t.Error("Expected output to contain 'IANA Time Zones'")
	}

	if !strings.Contains(stdout, "Total:") {
		t.Error("Expected output to contain total count")
	}
}

func TestListCmd_WithFilter(t *testing.T) {
	cmd := newListCmd()
	stdout, _, err := executeCommand(cmd, "--filter", "America")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "America") {
		t.Error("Expected output to contain 'America'")
	}

	if !strings.Contains(stdout, "filtered") {
		t.Error("Expected output to indicate filtering")
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	cmd := newListCmd()
	stdout, _, err := executeCommand(cmd, "--json")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, `"zones"`) {
		t.Error("Expected JSON output to contain 'zones' field")
	}

	if !strings.Contains(stdout, `"count"`) {
		t.Error("Expected JSON output to contain 'count' field")
	}
}

func TestInfoCmd_Help(t *testing.T) {
	cmd := newInfoCmd()

	if cmd.Use != "info [ZONE]" {
		t.Errorf("Expected Use to be 'info [ZONE]', got '%s'", cmd.Use)
	}

	zoneFlag := cmd.Flags().Lookup("zone")
	if zoneFlag == nil {
		t.Error("Expected --zone flag to exist")
	}
}

func TestInfoCmd_WithPositionalArg(t *testing.T) {
	cmd := newInfoCmd()
	stdout, _, err := executeCommand(cmd, "America/New_York")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Time Zone Information") {
		t.Error("Expected output to contain 'Time Zone Information'")
	}

	if !strings.Contains(stdout, "America/New_York") {
		t.Error("Expected output to contain zone name")
	}

	if !strings.Contains(stdout, "Abbreviation:") {
		t.Error("Expected output to contain abbreviation")
	}

	if !strings.Contains(stdout, "UTC Offset:") {
		t.Error("Expected output to contain UTC offset")
	}
}

func TestInfoCmd_WithFlag(t *testing.T) {
	cmd := newInfoCmd()
	stdout, _, err := executeCommand(cmd, "--zone", "Europe/London")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Europe/London") {
		t.Error("Expected output to contain zone name")
	}
}

func TestInfoCmd_WithAbbreviation(t *testing.T) {
	cmd := newInfoCmd()
	stdout, _, err := executeCommand(cmd, "PST")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "America/Los_Angeles") {
		t.Error("Expected PST to be expanded to America/Los_Angeles")
	}

	if !strings.Contains(stdout, "expanded from 'PST'") {
		t.Error("Expected output to show abbreviation expansion")
	}
}

func TestInfoCmd_MissingZone(t *testing.T) {
	cmd := newInfoCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Error("Expected error when zone is not provided")
	}
}

func TestInfoCmd_JSONOutput(t *testing.T) {
	cmd := newInfoCmd()
	stdout, _, err := executeCommand(cmd, "--zone", "UTC", "--json")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, `"zone"`) {
		t.Error("Expected JSON output to contain 'zone' field")
	}

	if !strings.Contains(stdout, `"abbreviation"`) {
		t.Error("Expected JSON output to contain 'abbreviation' field")
	}

	if !strings.Contains(stdout, `"is_dst"`) {
		t.Error("Expected JSON output to contain 'is_dst' field")
	}
}

func TestConvertCmd_SameTimezone(t *testing.T) {
	cmd := newConvertCmd()
	stdout, _, err := executeCommand(cmd, "--from", "UTC", "--to", "UTC")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "same offset") {
		t.Error("Expected output to indicate same offset")
	}
}

func TestFindMeetingCmd_InvalidDateFormat(t *testing.T) {
	cmd := newFindMeetingCmd()
	_, _, err := executeCommand(cmd,
		"--zones", "UTC",
		"--start-date", "invalid-date")

	if err == nil {
		t.Error("Expected error for invalid start date format")
	}
}

func TestFindMeetingCmd_InvalidEndDate(t *testing.T) {
	cmd := newFindMeetingCmd()
	_, _, err := executeCommand(cmd,
		"--zones", "UTC",
		"--end-date", "not-a-date")

	if err == nil {
		t.Error("Expected error for invalid end date format")
	}
}

func TestFindMeetingCmd_NoZones(t *testing.T) {
	cmd := newFindMeetingCmd()
	_, _, err := executeCommand(cmd,
		"--zones", "",
		"--duration", "1h")

	if err == nil {
		t.Error("Expected error when no zones provided")
	}
}

func TestFindMeetingCmd_InvalidWorkingHours(t *testing.T) {
	cmd := newFindMeetingCmd()
	_, _, err := executeCommand(cmd,
		"--zones", "UTC",
		"--start-hour", "invalid",
		"--end-hour", "17:00")

	if err == nil {
		t.Error("Expected error for invalid working hours")
	}
}

func TestDSTCmd_InvalidYear(t *testing.T) {
	cmd := newDSTCmd()
	_, _, err := executeCommand(cmd,
		"--zone", "America/New_York",
		"--year", "0")

	// Service should handle gracefully even with year 0
	// This test verifies the command doesn't panic
	_ = err
}

func TestListCmd_EmptyFilter(t *testing.T) {
	cmd := newListCmd()
	stdout, _, err := executeCommand(cmd, "--filter", "NonExistentZone12345")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "No time zones found") {
		t.Error("Expected output to indicate no zones found")
	}
}

func TestInfoCmd_InvalidTimeFormat(t *testing.T) {
	cmd := newInfoCmd()
	_, _, err := executeCommand(cmd,
		"--zone", "UTC",
		"--time", "invalid-time")

	if err == nil {
		t.Error("Expected error for invalid time format")
	}
}

func TestInfoCmd_InvalidTimezone(t *testing.T) {
	cmd := newInfoCmd()
	_, _, err := executeCommand(cmd, "Invalid/Timezone")

	if err == nil {
		t.Error("Expected error for invalid timezone")
	}
}
