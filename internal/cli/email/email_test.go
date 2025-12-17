package email

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

func TestNewEmailCmd(t *testing.T) {
	cmd := NewEmailCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "email", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "read", "send", "search", "mark", "delete", "folders", "threads", "drafts"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestListCommand(t *testing.T) {
	cmd := newListCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "list [grant-id]", cmd.Use)
	})

	t.Run("has_limit_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("limit")
		assert.NotNil(t, flag)
		assert.Equal(t, "10", flag.DefValue)
	})

	t.Run("has_unread_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("unread")
		assert.NotNil(t, flag)
	})

	t.Run("has_starred_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("starred")
		assert.NotNil(t, flag)
	})

	t.Run("has_from_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("from")
		assert.NotNil(t, flag)
	})
}

func TestReadCommand(t *testing.T) {
	cmd := newReadCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "read <message-id> [grant-id]", cmd.Use)
	})

	t.Run("has_show_alias", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "show")
	})

	t.Run("has_mark_read_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("mark-read")
		assert.NotNil(t, flag)
	})

	t.Run("has_raw_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("raw")
		assert.NotNil(t, flag)
	})
}

func TestSendCommand(t *testing.T) {
	cmd := newSendCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "send [grant-id]", cmd.Use)
	})

	t.Run("has_to_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("to")
		assert.NotNil(t, flag)
	})

	t.Run("has_subject_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("subject")
		assert.NotNil(t, flag)
	})

	t.Run("has_body_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("body")
		assert.NotNil(t, flag)
	})

	t.Run("has_cc_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("cc")
		assert.NotNil(t, flag)
	})

	t.Run("has_bcc_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("bcc")
		assert.NotNil(t, flag)
	})
}

func TestSearchCommand(t *testing.T) {
	cmd := newSearchCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "search <query> [grant-id]", cmd.Use)
	})

	t.Run("has_limit_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("limit")
		assert.NotNil(t, flag)
		assert.Equal(t, "20", flag.DefValue)
	})

	t.Run("has_from_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("from")
		assert.NotNil(t, flag)
	})

	t.Run("has_after_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("after")
		assert.NotNil(t, flag)
	})

	t.Run("has_before_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("before")
		assert.NotNil(t, flag)
	})
}

func TestMarkCommand(t *testing.T) {
	cmd := newMarkCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "mark", cmd.Use)
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.Len(t, subcommands, 4) // read, unread, starred, unstarred
	})

	t.Run("has_read_subcommand", func(t *testing.T) {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == "read" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestFoldersCommand(t *testing.T) {
	cmd := newFoldersCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "folders", cmd.Use)
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.GreaterOrEqual(t, len(subcommands), 3) // list, create, delete
	})
}

func TestFoldersListCommand(t *testing.T) {
	cmd := newFoldersListCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "list [grant-id]", cmd.Use)
	})

	t.Run("has_id_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("id")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "folders")
	})
}

func TestThreadsCommand(t *testing.T) {
	cmd := newThreadsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "threads", cmd.Use)
	})

	t.Run("has_list_subcommand", func(t *testing.T) {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == "list" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestDraftsCommand(t *testing.T) {
	cmd := newDraftsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "drafts", cmd.Use)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "create", "show", "send", "delete"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("formatTimeAgo", func(t *testing.T) {
		tests := []struct {
			duration time.Duration
			expected string
		}{
			{30 * time.Second, "just now"},
			{1 * time.Minute, "1 minute ago"},
			{5 * time.Minute, "5 minutes ago"},
			{1 * time.Hour, "1 hour ago"},
			{5 * time.Hour, "5 hours ago"},
			{24 * time.Hour, "yesterday"},
			{48 * time.Hour, "2 days ago"},
		}

		for _, tt := range tests {
			past := time.Now().Add(-tt.duration)
			got := formatTimeAgo(past)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("truncate", func(t *testing.T) {
		tests := []struct {
			input    string
			maxLen   int
			expected string
		}{
			{"hello", 10, "hello"},
			{"hello world", 8, "hello..."},
			{"short", 5, "short"},
		}

		for _, tt := range tests {
			got := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("formatContact", func(t *testing.T) {
		tests := []struct {
			contact  domain.EmailParticipant
			expected string
		}{
			{domain.EmailParticipant{Name: "John", Email: "john@example.com"}, "John"},
			{domain.EmailParticipant{Name: "", Email: "jane@example.com"}, "jane@example.com"},
		}

		for _, tt := range tests {
			got := formatContact(tt.contact)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("formatContacts", func(t *testing.T) {
		contacts := []domain.EmailParticipant{
			{Name: "John", Email: "john@example.com"},
			{Name: "", Email: "jane@example.com"},
		}
		got := formatContacts(contacts)
		assert.Equal(t, "John, jane@example.com", got)
	})

	t.Run("stripHTML", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"<p>Hello</p>", "Hello"},
			{"<html><body><h1>Title</h1><p>Content</p></body></html>", "Title\n\nContent"},
			{"Plain text", "Plain text"},
			{"Line 1<br>Line 2", "Line 1\nLine 2"},
			{"<div>Block 1</div><div>Block 2</div>", "Block 1\n\nBlock 2"},
			{"Text with &nbsp; entities &amp; &lt;tags&gt;", "Text with \u00a0 entities & <tags>"},
			{"<style>body{color:red}</style>Content", "Content"},
		}

		for _, tt := range tests {
			got := stripHTML(tt.input)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("formatSize", func(t *testing.T) {
		tests := []struct {
			bytes    int64
			expected string
		}{
			{500, "500 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{1048576, "1.0 MB"},
		}

		for _, tt := range tests {
			got := formatSize(tt.bytes)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("parseEmails", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{"a@b.com, c@d.com", []string{"a@b.com", "c@d.com"}},
			{"single@test.com", []string{"single@test.com"}},
			{"", nil},
			{"  spaced@test.com  ,  other@test.com  ", []string{"spaced@test.com", "other@test.com"}},
		}

		for _, tt := range tests {
			got := parseEmails(tt.input)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("parseContacts", func(t *testing.T) {
		tests := []struct {
			input    []string
			expected []domain.EmailParticipant
		}{
			{
				[]string{"test@example.com"},
				[]domain.EmailParticipant{{Email: "test@example.com"}},
			},
			{
				[]string{"John Doe <john@example.com>"},
				[]domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
			},
			{
				[]string{"plain@test.com", "Named <named@test.com>"},
				[]domain.EmailParticipant{
					{Email: "plain@test.com"},
					{Name: "Named", Email: "named@test.com"},
				},
			},
		}

		for _, tt := range tests {
			got := parseContacts(tt.input)
			assert.Equal(t, tt.expected, got)
		}
	})

	t.Run("parseDate", func(t *testing.T) {
		date, err := parseDate("2024-01-15")
		assert.NoError(t, err)
		assert.Equal(t, 2024, date.Year())
		assert.Equal(t, time.January, date.Month())
		assert.Equal(t, 15, date.Day())

		_, err = parseDate("invalid")
		assert.Error(t, err)
	})
}

func TestSendCommandScheduleFlags(t *testing.T) {
	cmd := newSendCmd()

	t.Run("has_schedule_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("schedule")
		assert.NotNil(t, flag)
	})

	t.Run("has_yes_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("yes")
		assert.NotNil(t, flag)
	})

	t.Run("has_yes_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("y")
		assert.NotNil(t, flag)
		assert.Equal(t, "yes", flag.Name)
	})
}

func TestParseScheduleTime(t *testing.T) {
	t.Run("parses_unix_timestamp", func(t *testing.T) {
		// A valid Unix timestamp (Jan 15, 2024)
		result, err := parseScheduleTime("1705320600")
		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
	})

	t.Run("parses_duration_minutes", func(t *testing.T) {
		now := time.Now()
		result, err := parseScheduleTime("30m")
		assert.NoError(t, err)
		// Should be approximately 30 minutes from now
		diff := result.Sub(now)
		assert.True(t, diff >= 29*time.Minute && diff <= 31*time.Minute)
	})

	t.Run("parses_duration_hours", func(t *testing.T) {
		now := time.Now()
		result, err := parseScheduleTime("2h")
		assert.NoError(t, err)
		// Should be approximately 2 hours from now
		diff := result.Sub(now)
		assert.True(t, diff >= 119*time.Minute && diff <= 121*time.Minute)
	})

	t.Run("parses_duration_days", func(t *testing.T) {
		now := time.Now()
		result, err := parseScheduleTime("1d")
		assert.NoError(t, err)
		// Should be approximately 1 day from now
		diff := result.Sub(now)
		assert.True(t, diff >= 23*time.Hour && diff <= 25*time.Hour)
	})

	t.Run("parses_tomorrow", func(t *testing.T) {
		result, err := parseScheduleTime("tomorrow")
		assert.NoError(t, err)
		expected := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, 9, result.Hour()) // Default to 9am
	})

	t.Run("parses_tomorrow_with_time", func(t *testing.T) {
		result, err := parseScheduleTime("tomorrow 2pm")
		assert.NoError(t, err)
		expected := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, 14, result.Hour())
	})

	t.Run("parses_datetime_format", func(t *testing.T) {
		// Use a future date to avoid "in the past" error
		futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02") + " 14:30"
		result, err := parseScheduleTime(futureDate)
		assert.NoError(t, err)
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("returns_error_for_invalid_format", func(t *testing.T) {
		_, err := parseScheduleTime("invalid time format xyz")
		assert.Error(t, err)
	})

	t.Run("returns_error_for_today_without_time", func(t *testing.T) {
		_, err := parseScheduleTime("today")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "specify a time")
	})
}

func TestParseTimeOfDay(t *testing.T) {
	t.Run("parses_24_hour_format", func(t *testing.T) {
		result, err := parseTimeOfDay("14:30")
		assert.NoError(t, err)
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("parses_12_hour_format_pm", func(t *testing.T) {
		result, err := parseTimeOfDay("2:30pm")
		assert.NoError(t, err)
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("parses_12_hour_format_am", func(t *testing.T) {
		result, err := parseTimeOfDay("9am")
		assert.NoError(t, err)
		assert.Equal(t, 9, result.Hour())
	})

	t.Run("parses_12_hour_format_with_space", func(t *testing.T) {
		result, err := parseTimeOfDay("3 pm")
		assert.NoError(t, err)
		assert.Equal(t, 15, result.Hour())
	})

	t.Run("returns_error_for_invalid_format", func(t *testing.T) {
		_, err := parseTimeOfDay("invalid")
		assert.Error(t, err)
	})
}

func TestEmailCommandHelp(t *testing.T) {
	cmd := NewEmailCmd()
	stdout, _, err := executeCommand(cmd, "--help")

	assert.NoError(t, err)

	expectedStrings := []string{
		"email",
		"list",
		"read",
		"send",
		"search",
		"folders",
		"threads",
		"drafts",
	}

	for _, expected := range expectedStrings {
		assert.Contains(t, stdout, expected, "Help output should contain %q", expected)
	}
}

func TestEmailSendHelp(t *testing.T) {
	cmd := NewEmailCmd()
	stdout, _, err := executeCommand(cmd, "send", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "send")
	assert.Contains(t, stdout, "--to")
	assert.Contains(t, stdout, "--subject")
	assert.Contains(t, stdout, "--body")
	assert.Contains(t, stdout, "--schedule")
	assert.Contains(t, stdout, "--yes")
}
