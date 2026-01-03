package calendar

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [grant-id]",
		Aliases: []string{"ls"},
		Short:   "List calendars",
		Long:    "List all calendars for the specified grant or default account.",
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

			ctx, cancel := common.CreateContext()
			defer cancel()

			calendars, err := client.GetCalendars(ctx, grantID)
			if err != nil {
				return fmt.Errorf("failed to get calendars: %w", err)
			}

			if len(calendars) == 0 {
				fmt.Println("No calendars found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			dim := color.New(color.Faint)

			fmt.Printf("Found %d calendar(s):\n\n", len(calendars))

			table := common.NewTable("NAME", "ID", "PRIMARY", "READ-ONLY")
			for _, cal := range calendars {
				primary := ""
				if cal.IsPrimary {
					primary = green.Sprint("Yes")
				}
				readOnly := ""
				if cal.ReadOnly {
					readOnly = dim.Sprint("Yes")
				}
				table.AddRow(cyan.Sprint(cal.Name), cal.ID, primary, readOnly)
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
