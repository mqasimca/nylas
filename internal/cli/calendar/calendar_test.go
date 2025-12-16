package calendar

import (
	"bytes"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// executeCommand executes a command and captures its output.
func executeCommand(root *cobra.Command, args ...string) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	err := root.Execute()

	return stdout.String(), stderr.String(), err
}

func TestNewCalendarCmd(t *testing.T) {
	cmd := NewCalendarCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "calendar", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "cal")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "calendar")
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "events", "availability"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestListCmd(t *testing.T) {
	cmd := newListCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "list [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "ls")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "List")
	})
}

func TestEventsCmd(t *testing.T) {
	cmd := newEventsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "events", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "ev")
		assert.Contains(t, cmd.Aliases, "event")
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "show", "create", "delete"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestEventsListCmd(t *testing.T) {
	cmd := newEventsListCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "list [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "ls")
	})

	t.Run("has_calendar_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("calendar")
		assert.NotNil(t, flag)
	})

	t.Run("has_calendar_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("c")
		assert.NotNil(t, flag)
		assert.Equal(t, "calendar", flag.Name)
	})

	t.Run("has_limit_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("limit")
		assert.NotNil(t, flag)
		assert.Equal(t, "10", flag.DefValue)
	})

	t.Run("has_days_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("days")
		assert.NotNil(t, flag)
		assert.Equal(t, "7", flag.DefValue)
	})

	t.Run("has_show_cancelled_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("show-cancelled")
		assert.NotNil(t, flag)
	})
}

func TestEventsShowCmd(t *testing.T) {
	cmd := newEventsShowCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "show <event-id> [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "read")
		assert.Contains(t, cmd.Aliases, "get")
	})

	t.Run("has_calendar_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("calendar")
		assert.NotNil(t, flag)
	})
}

func TestEventsCreateCmd(t *testing.T) {
	cmd := newEventsCreateCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "create [grant-id]", cmd.Use)
	})

	t.Run("has_title_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("title")
		assert.NotNil(t, flag)
	})

	t.Run("has_title_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("t")
		assert.NotNil(t, flag)
		assert.Equal(t, "title", flag.Name)
	})

	t.Run("has_description_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("description")
		assert.NotNil(t, flag)
	})

	t.Run("has_location_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("location")
		assert.NotNil(t, flag)
	})

	t.Run("has_start_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("start")
		assert.NotNil(t, flag)
	})

	t.Run("has_end_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("end")
		assert.NotNil(t, flag)
	})

	t.Run("has_all_day_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("all-day")
		assert.NotNil(t, flag)
	})

	t.Run("has_participant_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("participant")
		assert.NotNil(t, flag)
	})

	t.Run("has_busy_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("busy")
		assert.NotNil(t, flag)
		assert.Equal(t, "true", flag.DefValue)
	})

	t.Run("has_calendar_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("calendar")
		assert.NotNil(t, flag)
	})
}

func TestEventsDeleteCmd(t *testing.T) {
	cmd := newEventsDeleteCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "delete <event-id> [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "rm")
		assert.Contains(t, cmd.Aliases, "remove")
	})

	t.Run("has_force_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("force")
		assert.NotNil(t, flag)
	})

	t.Run("has_calendar_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("calendar")
		assert.NotNil(t, flag)
	})
}

func TestAvailabilityCmd(t *testing.T) {
	cmd := newAvailabilityCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "availability", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "avail")
		assert.Contains(t, cmd.Aliases, "freebusy")
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"check", "find"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestFreeBusyCmd(t *testing.T) {
	cmd := newFreeBusyCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "check [grant-id]", cmd.Use)
	})

	t.Run("has_emails_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("emails")
		assert.NotNil(t, flag)
	})

	t.Run("has_start_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("start")
		assert.NotNil(t, flag)
	})

	t.Run("has_end_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("end")
		assert.NotNil(t, flag)
	})

	t.Run("has_duration_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("duration")
		assert.NotNil(t, flag)
	})

	t.Run("has_format_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("format")
		assert.NotNil(t, flag)
		assert.Equal(t, "text", flag.DefValue)
	})

	t.Run("has_examples", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Example)
		assert.Contains(t, cmd.Example, "availability check")
	})
}

func TestFindSlotsCmd(t *testing.T) {
	cmd := newFindSlotsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "find", cmd.Use)
	})

	t.Run("has_participants_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("participants")
		assert.NotNil(t, flag)
	})

	t.Run("has_start_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("start")
		assert.NotNil(t, flag)
	})

	t.Run("has_end_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("end")
		assert.NotNil(t, flag)
	})

	t.Run("has_duration_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("duration")
		assert.NotNil(t, flag)
		assert.Equal(t, "30", flag.DefValue)
	})

	t.Run("has_interval_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("interval")
		assert.NotNil(t, flag)
		assert.Equal(t, "15", flag.DefValue)
	})

	t.Run("has_format_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("format")
		assert.NotNil(t, flag)
		assert.Equal(t, "text", flag.DefValue)
	})

	t.Run("has_examples", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Example)
		assert.Contains(t, cmd.Example, "availability find")
	})
}

func TestParseTimeInput(t *testing.T) {
	t.Run("parses_tomorrow", func(t *testing.T) {
		result, err := parseTimeInput("tomorrow")
		assert.NoError(t, err)
		expected := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, expected.Month(), result.Month())
	})

	t.Run("parses_tomorrow_with_time", func(t *testing.T) {
		result, err := parseTimeInput("tomorrow 9am")
		assert.NoError(t, err)
		expected := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, 9, result.Hour())
	})

	t.Run("parses_today", func(t *testing.T) {
		result, err := parseTimeInput("today")
		assert.NoError(t, err)
		now := time.Now()
		assert.Equal(t, now.Day(), result.Day())
		assert.Equal(t, now.Month(), result.Month())
	})

	t.Run("parses_iso_datetime", func(t *testing.T) {
		result, err := parseTimeInput("2024-01-15 14:30")
		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("parses_time_only", func(t *testing.T) {
		result, err := parseTimeInput("15:00")
		assert.NoError(t, err)
		assert.Equal(t, 15, result.Hour())
		assert.Equal(t, 0, result.Minute())
	})

	t.Run("returns_error_for_invalid_input", func(t *testing.T) {
		_, err := parseTimeInput("invalid time string xyz")
		assert.Error(t, err)
	})
}

func TestParseDuration(t *testing.T) {
	t.Run("parses_hours", func(t *testing.T) {
		result, err := parseDuration("8h")
		assert.NoError(t, err)
		assert.Equal(t, 8*time.Hour, result)
	})

	t.Run("parses_days", func(t *testing.T) {
		result, err := parseDuration("7d")
		assert.NoError(t, err)
		assert.Equal(t, 7*24*time.Hour, result)
	})

	t.Run("parses_minutes", func(t *testing.T) {
		result, err := parseDuration("30m")
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Minute, result)
	})

	t.Run("returns_error_for_invalid", func(t *testing.T) {
		_, err := parseDuration("invalid")
		assert.Error(t, err)
	})
}

func TestParseEventTime(t *testing.T) {
	t.Run("parses_all_day_event", func(t *testing.T) {
		when, err := parseEventTime("2024-01-15", "", true)
		assert.NoError(t, err)
		assert.Equal(t, "date", when.Object)
		assert.Equal(t, "2024-01-15", when.Date)
	})

	t.Run("parses_timed_event", func(t *testing.T) {
		when, err := parseEventTime("2024-01-15 14:00", "2024-01-15 15:00", false)
		assert.NoError(t, err)
		assert.Equal(t, "timespan", when.Object)
		assert.NotZero(t, when.StartTime)
		assert.NotZero(t, when.EndTime)
	})

	t.Run("defaults_end_to_one_hour", func(t *testing.T) {
		when, err := parseEventTime("2024-01-15 14:00", "", false)
		assert.NoError(t, err)
		assert.Equal(t, "timespan", when.Object)
		// End should be 1 hour after start
		assert.Equal(t, when.StartTime+3600, when.EndTime)
	})

	t.Run("parses_date_range", func(t *testing.T) {
		when, err := parseEventTime("2024-01-15", "2024-01-17", true)
		assert.NoError(t, err)
		assert.Equal(t, "datespan", when.Object)
		assert.Equal(t, "2024-01-15", when.StartDate)
		assert.Equal(t, "2024-01-17", when.EndDate)
	})

	t.Run("returns_error_for_invalid_start", func(t *testing.T) {
		_, err := parseEventTime("invalid", "", false)
		assert.Error(t, err)
	})
}

func TestFormatEventTime(t *testing.T) {
	t.Run("formats_all_day_event", func(t *testing.T) {
		when := domain.EventWhen{
			Object: "date",
			Date:   "2024-01-15",
		}
		result := formatEventTime(when)
		assert.Contains(t, result, "Jan 15, 2024")
		assert.Contains(t, result, "all day")
	})

	t.Run("formats_timed_event_same_day", func(t *testing.T) {
		start := time.Date(2024, 1, 15, 14, 0, 0, 0, time.Local)
		end := time.Date(2024, 1, 15, 15, 0, 0, 0, time.Local)
		when := domain.EventWhen{
			Object:    "timespan",
			StartTime: start.Unix(),
			EndTime:   end.Unix(),
		}
		result := formatEventTime(when)
		assert.Contains(t, result, "Jan 15, 2024")
		assert.Contains(t, result, "2:00 PM")
		assert.Contains(t, result, "3:00 PM")
	})
}

func TestFormatParticipantStatus(t *testing.T) {
	t.Run("formats_yes", func(t *testing.T) {
		result := formatParticipantStatus("yes")
		assert.Contains(t, result, "accepted")
	})

	t.Run("formats_no", func(t *testing.T) {
		result := formatParticipantStatus("no")
		assert.Contains(t, result, "declined")
	})

	t.Run("formats_maybe", func(t *testing.T) {
		result := formatParticipantStatus("maybe")
		assert.Contains(t, result, "tentative")
	})

	t.Run("formats_noreply", func(t *testing.T) {
		result := formatParticipantStatus("noreply")
		assert.Contains(t, result, "pending")
	})

	t.Run("empty_for_unknown", func(t *testing.T) {
		result := formatParticipantStatus("unknown")
		assert.Empty(t, result)
	})
}

func TestCalendarCommandHelp(t *testing.T) {
	cmd := NewCalendarCmd()
	stdout, _, err := executeCommand(cmd, "--help")

	assert.NoError(t, err)

	expectedStrings := []string{
		"calendar",
		"list",
		"events",
		"availability",
	}

	for _, expected := range expectedStrings {
		assert.Contains(t, stdout, expected, "Help output should contain %q", expected)
	}
}

func TestCalendarEventsHelp(t *testing.T) {
	cmd := NewCalendarCmd()
	stdout, _, err := executeCommand(cmd, "events", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "events")
	assert.Contains(t, stdout, "list")
	assert.Contains(t, stdout, "show")
	assert.Contains(t, stdout, "create")
	assert.Contains(t, stdout, "delete")
}

func TestCalendarAvailabilityHelp(t *testing.T) {
	cmd := NewCalendarCmd()
	stdout, _, err := executeCommand(cmd, "availability", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "availability")
	assert.Contains(t, stdout, "check")
	assert.Contains(t, stdout, "find")
}
