// users.go provides user listing commands for Slack workspaces.

package slack

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
)

// newUsersCmd creates the users command for managing workspace users.
func newUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user", "members"},
		Short:   "Manage workspace users",
		Long:    `Commands for listing and managing Slack workspace users.`,
	}

	cmd.AddCommand(newUserListCmd())

	return cmd
}

// newUserListCmd creates the list subcommand for listing workspace users.
func newUserListCmd() *cobra.Command {
	var (
		limit  int
		showID bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List workspace users",
		Long: `List members of your Slack workspace.

Examples:
  # List users
  nylas slack users list

  # Show user IDs
  nylas slack users list --id

  # Limit results
  nylas slack users list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			resp, err := client.ListUsers(ctx, limit, "")
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			if len(resp.Users) == 0 {
				fmt.Println("No users found")
				return nil
			}

			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)
			yellow := color.New(color.FgYellow)

			for _, u := range resp.Users {
				name := u.BestDisplayName()
				fmt.Print(cyan.Sprint(name))

				if u.Name != "" && u.Name != name {
					_, _ = dim.Printf(" (@%s)", u.Name)
				}

				if showID {
					_, _ = dim.Printf(" [%s]", u.ID)
				}

				if u.IsBot {
					_, _ = yellow.Print(" [bot]")
				}
				if u.IsAdmin {
					_, _ = yellow.Print(" [admin]")
				}

				fmt.Println()

				if u.Status != "" {
					_, _ = dim.Printf("  %s\n", u.Status)
				}
			}

			if resp.NextCursor != "" {
				dim := color.New(color.Faint)
				_, _ = dim.Printf("\n(more users available)\n")
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 100, "Maximum number of users to return")
	cmd.Flags().BoolVar(&showID, "id", false, "Show user IDs")

	return cmd
}
