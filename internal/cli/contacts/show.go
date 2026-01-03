package contacts

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <contact-id> [grant-id]",
		Aliases: []string{"get", "read"},
		Short:   "Show contact details",
		Long:    "Display detailed information about a specific contact.",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			var grantID string
			if len(args) > 1 {
				grantID = args[1]
			} else {
				grantID, err = getGrantID(nil)
				if err != nil {
					return err
				}
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			contact, err := client.GetContact(ctx, grantID, contactID)
			if err != nil {
				return fmt.Errorf("failed to get contact: %w", err)
			}

			cyan := color.New(color.FgCyan, color.Bold)
			green := color.New(color.FgGreen)
			dim := color.New(color.Faint)

			// Name
			fmt.Printf("%s\n\n", cyan.Sprint(contact.DisplayName()))

			// Work info
			if contact.CompanyName != "" || contact.JobTitle != "" {
				fmt.Printf("%s\n", green.Sprint("Work"))
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
				fmt.Printf("%s\n", green.Sprint("Email Addresses"))
				for _, e := range contact.Emails {
					typeStr := ""
					if e.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", e.Type)
					}
					fmt.Printf("  %s%s\n", e.Email, dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Phone numbers
			if len(contact.PhoneNumbers) > 0 {
				fmt.Printf("%s\n", green.Sprint("Phone Numbers"))
				for _, p := range contact.PhoneNumbers {
					typeStr := ""
					if p.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", p.Type)
					}
					fmt.Printf("  %s%s\n", p.Number, dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Addresses
			if len(contact.PhysicalAddresses) > 0 {
				fmt.Printf("%s\n", green.Sprint("Addresses"))
				for _, a := range contact.PhysicalAddresses {
					typeStr := ""
					if a.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", a.Type)
					}
					fmt.Printf("  %s\n", dim.Sprint(typeStr))
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
				fmt.Printf("%s\n", green.Sprint("Web Pages"))
				for _, w := range contact.WebPages {
					typeStr := ""
					if w.Type != "" {
						typeStr = fmt.Sprintf(" (%s)", w.Type)
					}
					fmt.Printf("  %s%s\n", w.URL, dim.Sprint(typeStr))
				}
				fmt.Println()
			}

			// Personal info
			if contact.Birthday != "" || contact.Nickname != "" {
				fmt.Printf("%s\n", green.Sprint("Personal"))
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
				fmt.Printf("%s\n", green.Sprint("Notes"))
				fmt.Printf("  %s\n\n", contact.Notes)
			}

			// Metadata
			fmt.Printf("%s\n", green.Sprint("Details"))
			fmt.Printf("  ID: %s\n", dim.Sprint(contact.ID))
			if contact.Source != "" {
				fmt.Printf("  Source: %s\n", contact.Source)
			}

			return nil
		},
	}

	return cmd
}
