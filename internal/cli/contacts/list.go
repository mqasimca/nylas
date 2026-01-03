package contacts

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		limit  int
		email  string
		source string
		showID bool
	)

	cmd := &cobra.Command{
		Use:     "list [grant-id]",
		Aliases: []string{"ls"},
		Short:   "List contacts",
		Long:    "List all contacts for the specified grant or default account.",
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

			params := &domain.ContactQueryParams{
				Limit:  limit,
				Email:  email,
				Source: source,
			}

			contacts, err := client.GetContacts(ctx, grantID, params)
			if err != nil {
				return fmt.Errorf("failed to get contacts: %w", err)
			}

			if len(contacts) == 0 {
				fmt.Println("No contacts found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			fmt.Printf("Found %d contact(s):\n\n", len(contacts))

			var table *common.Table
			if showID {
				table = common.NewTable("ID", "NAME", "EMAIL", "PHONE", "COMPANY")
			} else {
				table = common.NewTable("NAME", "EMAIL", "PHONE", "COMPANY")
			}
			for _, contact := range contacts {
				name := contact.DisplayName()
				email := contact.PrimaryEmail()
				phone := contact.PrimaryPhone()
				company := contact.CompanyName
				if contact.JobTitle != "" && company != "" {
					company = fmt.Sprintf("%s - %s", contact.JobTitle, company)
				} else if contact.JobTitle != "" {
					company = contact.JobTitle
				}

				if showID {
					table.AddRow(
						dim.Sprint(contact.ID),
						cyan.Sprint(name),
						email,
						phone,
						dim.Sprint(company),
					)
				} else {
					table.AddRow(
						cyan.Sprint(name),
						email,
						phone,
						dim.Sprint(company),
					)
				}
			}
			table.Render()

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Maximum number of contacts to show")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Filter by email address")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Filter by source (address_book, inbox, domain)")
	cmd.Flags().BoolVar(&showID, "id", false, "Show contact IDs")

	return cmd
}
