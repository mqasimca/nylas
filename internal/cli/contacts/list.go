package contacts

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
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
			// Check if we should use structured output (JSON/YAML/quiet)
			if common.IsJSON(cmd) {
				_, err := common.WithClient(args, func(ctx context.Context, client ports.NylasClient, grantID string) (struct{}, error) {
					params := &domain.ContactQueryParams{
						Limit:  limit,
						Email:  email,
						Source: source,
					}

					contacts, err := client.GetContacts(ctx, grantID, params)
					if err != nil {
						return struct{}{}, common.WrapListError("contacts", err)
					}

					out := common.GetOutputWriter(cmd)
					return struct{}{}, out.Write(contacts)
				})
				return err
			}

			// Traditional table output
			client, err := common.GetNylasClient()
			if err != nil {
				return err
			}

			grantID, err := common.GetGrantID(args)
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
				return common.WrapListError("contacts", err)
			}

			if len(contacts) == 0 {
				common.PrintEmptyState("contacts")
				return nil
			}

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
						common.Dim.Sprint(contact.ID),
						common.Cyan.Sprint(name),
						email,
						phone,
						common.Dim.Sprint(company),
					)
				} else {
					table.AddRow(
						common.Cyan.Sprint(name),
						email,
						phone,
						common.Dim.Sprint(company),
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
