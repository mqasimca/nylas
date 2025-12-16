package contacts

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "groups [grant-id]",
		Aliases: []string{"group"},
		Short:   "List contact groups",
		Long:    "List all contact groups for the specified grant or default account.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			groups, err := client.GetContactGroups(ctx, grantID)
			if err != nil {
				return fmt.Errorf("failed to get contact groups: %w", err)
			}

			if len(groups) == 0 {
				fmt.Println("No contact groups found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			fmt.Printf("Found %d contact group(s):\n\n", len(groups))

			table := common.NewTable("NAME", "ID", "PATH")
			for _, group := range groups {
				table.AddRow(
					cyan.Sprint(group.Name),
					dim.Sprint(group.ID),
					group.Path,
				)
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
