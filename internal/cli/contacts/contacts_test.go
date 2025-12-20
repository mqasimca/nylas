package contacts

import (
	"bytes"
	"testing"

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

func TestNewContactsCmd(t *testing.T) {
	cmd := NewContactsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "contacts", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "contact")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "contact")
	})

	t.Run("has_long_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "show", "create", "update", "delete", "groups", "search", "photo", "sync"}

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

	t.Run("has_limit_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("limit")
		assert.NotNil(t, flag)
		assert.Equal(t, "50", flag.DefValue)
	})

	t.Run("has_limit_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("n")
		assert.NotNil(t, flag)
		assert.Equal(t, "limit", flag.Name)
	})

	t.Run("has_email_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("email")
		assert.NotNil(t, flag)
	})

	t.Run("has_email_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("e")
		assert.NotNil(t, flag)
		assert.Equal(t, "email", flag.Name)
	})

	t.Run("has_source_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("source")
		assert.NotNil(t, flag)
	})

	t.Run("has_source_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("s")
		assert.NotNil(t, flag)
		assert.Equal(t, "source", flag.Name)
	})

	t.Run("has_id_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("id")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestShowCmd(t *testing.T) {
	cmd := newShowCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "show <contact-id> [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "get")
		assert.Contains(t, cmd.Aliases, "read")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Show")
	})

	t.Run("requires_args", func(t *testing.T) {
		assert.NotNil(t, cmd.Args)
	})
}

func TestCreateCmd(t *testing.T) {
	cmd := newCreateCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "create [grant-id]", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Create")
	})

	t.Run("has_first_name_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("first-name")
		assert.NotNil(t, flag)
	})

	t.Run("has_first_name_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("f")
		assert.NotNil(t, flag)
		assert.Equal(t, "first-name", flag.Name)
	})

	t.Run("has_last_name_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("last-name")
		assert.NotNil(t, flag)
	})

	t.Run("has_last_name_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("l")
		assert.NotNil(t, flag)
		assert.Equal(t, "last-name", flag.Name)
	})

	t.Run("has_email_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("email")
		assert.NotNil(t, flag)
	})

	t.Run("has_email_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("e")
		assert.NotNil(t, flag)
		assert.Equal(t, "email", flag.Name)
	})

	t.Run("has_phone_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("phone")
		assert.NotNil(t, flag)
	})

	t.Run("has_phone_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("p")
		assert.NotNil(t, flag)
		assert.Equal(t, "phone", flag.Name)
	})

	t.Run("has_company_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("company")
		assert.NotNil(t, flag)
	})

	t.Run("has_company_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("c")
		assert.NotNil(t, flag)
		assert.Equal(t, "company", flag.Name)
	})

	t.Run("has_job_title_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("job-title")
		assert.NotNil(t, flag)
	})

	t.Run("has_job_title_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("j")
		assert.NotNil(t, flag)
		assert.Equal(t, "job-title", flag.Name)
	})

	t.Run("has_notes_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("notes")
		assert.NotNil(t, flag)
	})

	t.Run("has_notes_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("n")
		assert.NotNil(t, flag)
		assert.Equal(t, "notes", flag.Name)
	})

	t.Run("has_long_description_with_examples", func(t *testing.T) {
		assert.Contains(t, cmd.Long, "Examples")
	})
}

func TestDeleteCmd(t *testing.T) {
	cmd := newDeleteCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "delete <contact-id> [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "rm")
		assert.Contains(t, cmd.Aliases, "remove")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Delete")
	})

	t.Run("has_force_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("force")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has_force_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("f")
		assert.NotNil(t, flag)
		assert.Equal(t, "force", flag.Name)
	})

	t.Run("requires_args", func(t *testing.T) {
		assert.NotNil(t, cmd.Args)
	})
}

func TestUpdateCmd(t *testing.T) {
	cmd := newUpdateCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "update <contact-id> [grant-id]", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Update")
	})

	t.Run("has_given_name_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("given-name")
		assert.NotNil(t, flag)
	})

	t.Run("has_surname_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("surname")
		assert.NotNil(t, flag)
	})

	t.Run("has_company_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("company")
		assert.NotNil(t, flag)
	})

	t.Run("has_job_title_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("job-title")
		assert.NotNil(t, flag)
	})

	t.Run("has_email_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("email")
		assert.NotNil(t, flag)
	})

	t.Run("has_phone_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("phone")
		assert.NotNil(t, flag)
	})

	t.Run("has_long_description_with_examples", func(t *testing.T) {
		assert.Contains(t, cmd.Long, "Examples")
	})
}

func TestGroupsCmd(t *testing.T) {
	cmd := newGroupsCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "groups", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "group")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "group")
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_required_subcommands", func(t *testing.T) {
		expectedCmds := []string{"list", "show", "create", "update", "delete"}

		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}

		for _, expected := range expectedCmds {
			assert.True(t, cmdMap[expected], "Missing expected subcommand: %s", expected)
		}
	})
}

func TestGroupsListCmd(t *testing.T) {
	cmd := newGroupsListCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "list [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "ls")
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
	})
}

func TestGroupsShowCmd(t *testing.T) {
	cmd := newGroupsShowCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "show <group-id> [grant-id]", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
	})
}

func TestGroupsCreateCmd(t *testing.T) {
	cmd := newGroupsCreateCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "create <name> [grant-id]", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Create")
	})

	t.Run("has_long_description_with_examples", func(t *testing.T) {
		assert.Contains(t, cmd.Long, "Examples")
	})
}

func TestGroupsUpdateCmd(t *testing.T) {
	cmd := newGroupsUpdateCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "update <group-id> [grant-id]", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Update")
	})

	t.Run("has_name_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("name")
		assert.NotNil(t, flag)
	})
}

func TestGroupsDeleteCmd(t *testing.T) {
	cmd := newGroupsDeleteCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "delete <group-id> [grant-id]", cmd.Use)
	})

	t.Run("has_aliases", func(t *testing.T) {
		assert.Contains(t, cmd.Aliases, "rm")
		assert.Contains(t, cmd.Aliases, "remove")
	})

	t.Run("has_force_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("force")
		assert.NotNil(t, flag)
	})
}

func TestContactsCommandHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "--help")

	assert.NoError(t, err)

	expectedStrings := []string{
		"contacts",
		"list",
		"show",
		"create",
		"update",
		"delete",
		"groups",
	}

	for _, expected := range expectedStrings {
		assert.Contains(t, stdout, expected, "Help output should contain %q", expected)
	}
}

func TestContactsListHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "list", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "list")
	assert.Contains(t, stdout, "--limit")
	assert.Contains(t, stdout, "--email")
	assert.Contains(t, stdout, "--source")
	assert.Contains(t, stdout, "--id")
}

func TestContactsCreateHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "create", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "create")
	assert.Contains(t, stdout, "--first-name")
	assert.Contains(t, stdout, "--last-name")
	assert.Contains(t, stdout, "--email")
	assert.Contains(t, stdout, "--phone")
	assert.Contains(t, stdout, "--company")
}

func TestContactsDeleteHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "delete", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "delete")
	assert.Contains(t, stdout, "--force")
}

func TestContactsUpdateHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "update", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "update")
	assert.Contains(t, stdout, "--given-name")
	assert.Contains(t, stdout, "--surname")
	assert.Contains(t, stdout, "--company")
	assert.Contains(t, stdout, "--email")
}

func TestContactsGroupsHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "groups", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "groups")
	assert.Contains(t, stdout, "list")
	assert.Contains(t, stdout, "show")
	assert.Contains(t, stdout, "create")
	assert.Contains(t, stdout, "update")
	assert.Contains(t, stdout, "delete")
}

func TestContactsGroupsCreateHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "groups", "create", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "create")
	assert.Contains(t, stdout, "<name>")
}

func TestContactsGroupsUpdateHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "groups", "update", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "update")
	assert.Contains(t, stdout, "--name")
}

func TestContactsGroupsDeleteHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "groups", "delete", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "delete")
	assert.Contains(t, stdout, "--force")
}

func TestSearchCmd(t *testing.T) {
	cmd := newSearchCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "search", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Search")
	})

	t.Run("has_company_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("company")
		assert.NotNil(t, flag)
	})

	t.Run("has_email_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("email")
		assert.NotNil(t, flag)
	})

	t.Run("has_phone_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("phone")
		assert.NotNil(t, flag)
	})

	t.Run("has_source_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("source")
		assert.NotNil(t, flag)
	})

	t.Run("has_group_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("group")
		assert.NotNil(t, flag)
	})

	t.Run("has_has_email_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("has-email")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has_limit_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("limit")
		assert.NotNil(t, flag)
		assert.Equal(t, "50", flag.DefValue)
	})

	t.Run("has_json_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("json")
		assert.NotNil(t, flag)
	})
}

func TestPhotoCmd(t *testing.T) {
	cmd := newPhotoCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "photo", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "photo")
	})

	t.Run("has_subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("has_download_subcommand", func(t *testing.T) {
		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}
		assert.True(t, cmdMap["download"], "Missing download subcommand")
	})

	t.Run("has_info_subcommand", func(t *testing.T) {
		cmdMap := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			cmdMap[sub.Name()] = true
		}
		assert.True(t, cmdMap["info"], "Missing info subcommand")
	})
}

func TestPhotoDownloadCmd(t *testing.T) {
	cmd := newPhotoDownloadCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "download <contact-id>", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "Download")
	})

	t.Run("has_output_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("output")
		assert.NotNil(t, flag)
	})

	t.Run("has_output_shorthand", func(t *testing.T) {
		flag := cmd.Flags().ShorthandLookup("o")
		assert.NotNil(t, flag)
		assert.Equal(t, "output", flag.Name)
	})

	t.Run("has_json_flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("json")
		assert.NotNil(t, flag)
	})

	t.Run("requires_args", func(t *testing.T) {
		assert.NotNil(t, cmd.Args)
	})
}

func TestPhotoInfoCmd(t *testing.T) {
	cmd := newPhotoInfoCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "info", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has_long_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Long)
		assert.Contains(t, cmd.Long, "Profile Picture")
	})
}

func TestSyncCmd(t *testing.T) {
	cmd := newSyncCmd()

	t.Run("command_name", func(t *testing.T) {
		assert.Equal(t, "sync", cmd.Use)
	})

	t.Run("has_short_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Short)
		assert.Contains(t, cmd.Short, "sync")
	})

	t.Run("has_long_description", func(t *testing.T) {
		assert.NotEmpty(t, cmd.Long)
		assert.Contains(t, cmd.Long, "Synchronization")
	})
}

func TestContactsSearchHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "search", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "search")
	assert.Contains(t, stdout, "--company")
	assert.Contains(t, stdout, "--email")
	assert.Contains(t, stdout, "--has-email")
}

func TestContactsPhotoHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "photo", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "photo")
	assert.Contains(t, stdout, "download")
	assert.Contains(t, stdout, "info")
}

func TestContactsSyncHelp(t *testing.T) {
	cmd := NewContactsCmd()
	stdout, _, err := executeCommand(cmd, "sync", "--help")

	assert.NoError(t, err)
	assert.Contains(t, stdout, "sync")
}
