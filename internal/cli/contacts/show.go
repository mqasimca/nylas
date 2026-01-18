package contacts

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return common.NewShowCommand(common.ShowCommandConfig{
		Use:          "show <contact-id> [grant-id]",
		Aliases:      []string{"get", "read"},
		Short:        "Show contact details",
		Long:         "Display detailed information about a specific contact.",
		ResourceName: "contact",
		GetFunc: func(ctx context.Context, grantID, resourceID string) (interface{}, error) {
			client, err := common.GetNylasClient()
			if err != nil {
				return nil, err
			}
			return client.GetContact(ctx, grantID, resourceID)
		},
		DisplayFunc: func(resource interface{}) error {
			contact := resource.(*domain.Contact)

			// Name
			fmt.Printf("%s\n\n", common.BoldCyan.Sprint(contact.DisplayName()))

			// Work info
			if contact.CompanyName != "" || contact.JobTitle != "" {
				fmt.Printf("%s\n", common.Green.Sprint("Work"))
				if contact.JobTitle != "" {
					fmt.Printf("  Job Title: %s\n", contact.JobTitle)
				}
				if contact.CompanyName != "" {
					fmt.Printf("  Company: %s\n", contact.CompanyName)
				}
				if contact.ManagerName != "" {
					fmt.Printf("  Manager: %s\n", contact.ManagerName)
				}
				fmt.Println()
			}

			// Emails
			if len(contact.Emails) > 0 {
				fmt.Printf("%s\n", common.Green.Sprint("Email Addresses"))
				for _, e := range contact.Emails {
					typeStr := ""
					if e.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", e.Type)
					}
					fmt.Printf("  %s%s\n", e.Email, common.Dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Phone numbers
			if len(contact.PhoneNumbers) > 0 {
				fmt.Printf("%s\n", common.Green.Sprint("Phone Numbers"))
				for _, p := range contact.PhoneNumbers {
					typeStr := ""
					if p.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", p.Type)
					}
					fmt.Printf("  %s%s\n", p.Number, common.Dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Addresses
			if len(contact.PhysicalAddresses) > 0 {
				fmt.Printf("%s\n", common.Green.Sprint("Addresses"))
				for _, a := range contact.PhysicalAddresses {
					typeStr := ""
					if a.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", a.Type)
					}
					fmt.Printf("  %s\n", common.Dim.Sprint(typeStr))
					if a.StreetAddress != "" {
						fmt.Printf("    %s\n", a.StreetAddress)
					}
					cityState := ""
					if a.City != "" {
						cityState = a.City
					}
					if a.State != "" {
						if cityState != "" {
							cityState += ", "
						}
						cityState += a.State
					}
					if a.PostalCode != "" {
						if cityState != "" {
							cityState += " "
						}
						cityState += a.PostalCode
					}
					if cityState != "" {
						fmt.Printf("    %s\n", cityState)
					}
					if a.Country != "" {
						fmt.Printf("    %s\n", a.Country)
					}
				}
				fmt.Println()
			}

			// Web pages
			if len(contact.WebPages) > 0 {
				fmt.Printf("%s\n", common.Green.Sprint("Web Pages"))
				for _, w := range contact.WebPages {
					typeStr := ""
					if w.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", w.Type)
					}
					fmt.Printf("  %s%s\n", w.URL, common.Dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Personal info
			if contact.Birthday != "" || contact.Nickname != "" {
				fmt.Printf("%s\n", common.Green.Sprint("Personal"))
				if contact.Nickname != "" {
					fmt.Printf("  Nickname: %s\n", contact.Nickname)
				}
				if contact.Birthday != "" {
					fmt.Printf("  Birthday: %s\n", contact.Birthday)
				}
				fmt.Println()
			}

			// Notes
			if contact.Notes != "" {
				fmt.Printf("%s\n", common.Green.Sprint("Notes"))
				fmt.Printf("  %s\n\n", contact.Notes)
			}

			// Metadata
			fmt.Printf("%s\n", common.Green.Sprint("Details"))
			fmt.Printf("  ID: %s\n", common.Dim.Sprint(contact.ID))
			if contact.Source != "" {
				fmt.Printf("  Source: %s\n", contact.Source)
			}

			return nil
		},
		GetClient: common.GetNylasClient,
	})
}
