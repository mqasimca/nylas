package demo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
)

// newDemoContactsCmd creates the demo contacts command with subcommands.
func newDemoContactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Explore contacts features with sample data",
		Long:  "Demo contacts commands showing sample contacts and simulated operations.",
	}

	// Basic CRUD
	cmd.AddCommand(newDemoContactsListCmd())
	cmd.AddCommand(newDemoContactsShowCmd())
	cmd.AddCommand(newDemoContactsCreateCmd())
	cmd.AddCommand(newDemoContactsUpdateCmd())
	cmd.AddCommand(newDemoContactsDeleteCmd())

	// Search
	cmd.AddCommand(newDemoContactsSearchCmd())

	// Groups
	cmd.AddCommand(newDemoContactsGroupsCmd())

	// Photo
	cmd.AddCommand(newDemoContactsPhotoCmd())

	// Sync
	cmd.AddCommand(newDemoContactsSyncCmd())

	return cmd
}

// ============================================================================
// BASIC CRUD COMMANDS
// ============================================================================

// newDemoContactsListCmd lists sample contacts.
func newDemoContactsListCmd() *cobra.Command {
	var limit int
	var showID bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sample contacts",
		Long:  "Display a list of realistic sample contacts.",
		Example: `  # List sample contacts
  nylas demo contacts list

  # List with IDs shown
  nylas demo contacts list --id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			contacts, err := client.GetContacts(ctx, "demo-grant", nil)
			if err != nil {
				return fmt.Errorf("failed to get demo contacts: %w", err)
			}

			if limit > 0 && limit < len(contacts) {
				contacts = contacts[:limit]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Sample Contacts"))
			fmt.Println(dim.Sprint("These are sample contacts for demonstration purposes."))
			fmt.Println()
			fmt.Printf("Found %d contacts:\n\n", len(contacts))

			for _, contact := range contacts {
				printDemoContact(contact, showID)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To connect your real contacts: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of contacts to show")
	cmd.Flags().BoolVar(&showID, "id", false, "Show contact IDs")

	return cmd
}

// newDemoContactsShowCmd shows a sample contact.
func newDemoContactsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [contact-id]",
		Aliases: []string{"read"},
		Short:   "Show a sample contact",
		Long:    "Display a sample contact to see the full contact format.",
		Example: `  # Show first sample contact
  nylas demo contacts show

  # Show specific contact
  nylas demo contacts show contact-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			contactID := "contact-001"
			if len(args) > 0 {
				contactID = args[0]
			}

			contact, err := client.GetContact(ctx, "demo-grant", contactID)
			if err != nil {
				return fmt.Errorf("failed to get demo contact: %w", err)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Sample Contact"))
			fmt.Println()
			printDemoContactFull(*contact)

			fmt.Println(dim.Sprint("To connect your real contacts: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// newDemoContactsCreateCmd simulates creating a contact.
func newDemoContactsCreateCmd() *cobra.Command {
	var firstName string
	var lastName string
	var email string
	var phone string
	var company string
	var jobTitle string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Simulate creating a contact",
		Long:  "Simulate creating a contact to see how the create command works.",
		Example: `  # Create a basic contact
  nylas demo contacts create --first-name "John" --last-name "Doe" --email "john@example.com"

  # Create with company info
  nylas demo contacts create --first-name "Jane" --last-name "Smith" --email "jane@company.com" --company "Acme Inc" --title "Engineer"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if firstName == "" {
				firstName = "Demo"
			}
			if lastName == "" {
				lastName = "Contact"
			}
			if email == "" {
				email = "demo@example.com"
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Simulated Contact Creation"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Printf("Name:     %s %s\n", firstName, lastName)
			fmt.Printf("Email:    %s\n", email)
			if phone != "" {
				fmt.Printf("Phone:    %s\n", phone)
			}
			if company != "" {
				fmt.Printf("Company:  %s\n", company)
			}
			if jobTitle != "" {
				fmt.Printf("Title:    %s\n", jobTitle)
			}
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			_, _ = green.Println("✓ Contact would be created (demo mode - no actual contact created)")
			_, _ = dim.Printf("  Contact ID: contact-demo-%d\n", time.Now().Unix())
			fmt.Println()
			fmt.Println(dim.Sprint("To create real contacts, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	cmd.Flags().StringVar(&jobTitle, "title", "", "Job title")

	return cmd
}

// newDemoContactsUpdateCmd simulates updating a contact.
func newDemoContactsUpdateCmd() *cobra.Command {
	var email string
	var phone string
	var company string
	var jobTitle string

	cmd := &cobra.Command{
		Use:   "update [contact-id]",
		Short: "Simulate updating a contact",
		Long:  "Simulate updating a contact to see how the update command works.",
		Example: `  # Update contact email
  nylas demo contacts update contact-001 --email "newemail@example.com"

  # Update company info
  nylas demo contacts update contact-001 --company "New Company" --title "Senior Engineer"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := "contact-demo-123"
			if len(args) > 0 {
				contactID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Simulated Contact Update"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = dim.Printf("Contact ID: %s\n", contactID)
			fmt.Println()
			_, _ = boldWhite.Println("Changes:")
			if email != "" {
				fmt.Printf("  Email:   %s\n", email)
			}
			if phone != "" {
				fmt.Printf("  Phone:   %s\n", phone)
			}
			if company != "" {
				fmt.Printf("  Company: %s\n", company)
			}
			if jobTitle != "" {
				fmt.Printf("  Title:   %s\n", jobTitle)
			}
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			_, _ = green.Println("✓ Contact would be updated (demo mode - no actual changes made)")
			fmt.Println()
			fmt.Println(dim.Sprint("To update real contacts, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "New email address")
	cmd.Flags().StringVar(&phone, "phone", "", "New phone number")
	cmd.Flags().StringVar(&company, "company", "", "New company name")
	cmd.Flags().StringVar(&jobTitle, "title", "", "New job title")

	return cmd
}

// newDemoContactsDeleteCmd simulates deleting a contact.
func newDemoContactsDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [contact-id]",
		Short: "Simulate deleting a contact",
		Long:  "Simulate deleting a contact to see how the delete command works.",
		Example: `  # Delete a contact
  nylas demo contacts delete contact-001

  # Force delete without confirmation
  nylas demo contacts delete contact-001 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := "contact-demo-123"
			if len(args) > 0 {
				contactID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Simulated Contact Deletion"))
			fmt.Println()

			if !force {
				_, _ = yellow.Println("⚠ Would prompt for confirmation in real mode")
			}

			fmt.Printf("Contact ID: %s\n", contactID)
			fmt.Println()
			_, _ = green.Println("✓ Contact would be deleted (demo mode - no actual deletion)")
			fmt.Println()
			fmt.Println(dim.Sprint("To delete real contacts, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

// ============================================================================
// SEARCH COMMAND
// ============================================================================

// newDemoContactsSearchCmd simulates searching contacts.
func newDemoContactsSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search sample contacts",
		Long:  "Search through sample contacts by name, email, or company.",
		Example: `  # Search by name
  nylas demo contacts search "John"

  # Search by email domain
  nylas demo contacts search "@acme.com"

  # Search by company
  nylas demo contacts search "Acme"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := "John"
			if len(args) > 0 {
				query = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👤 Demo Mode - Contact Search"))
			fmt.Println()
			fmt.Printf("Search query: %s\n\n", cyan.Sprint(query))

			// Sample search results
			results := []struct {
				name    string
				email   string
				company string
			}{
				{"John Smith", "john.smith@acme.com", "Acme Corp"},
				{"Johnny Appleseed", "johnny@example.com", "Example Inc"},
				{"Sarah Johnson", "sarah.johnson@acme.com", "Acme Corp"},
			}

			fmt.Printf("Found %d contacts:\n\n", len(results))

			for _, r := range results {
				fmt.Printf("  %s %s\n", "👤", boldWhite.Sprint(r.name))
				fmt.Printf("    📧 %s\n", r.email)
				fmt.Printf("    💼 %s\n", dim.Sprint(r.company))
				fmt.Println()
			}

			fmt.Println(dim.Sprint("To search your real contacts: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// ============================================================================
// GROUPS COMMAND
// ============================================================================

// newDemoContactsGroupsCmd creates the groups subcommand.
func newDemoContactsGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage contact groups",
		Long:  "Demo commands for managing contact groups.",
	}

	cmd.AddCommand(newDemoGroupsListCmd())
	cmd.AddCommand(newDemoGroupsShowCmd())
	cmd.AddCommand(newDemoGroupsCreateCmd())
	cmd.AddCommand(newDemoGroupsDeleteCmd())
	cmd.AddCommand(newDemoGroupsAddMemberCmd())
	cmd.AddCommand(newDemoGroupsRemoveMemberCmd())

	return cmd
}

func newDemoGroupsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List contact groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("👥 Demo Mode - Contact Groups"))
			fmt.Println()

			groups := []struct {
				name  string
				count int
			}{
				{"Work Colleagues", 25},
				{"Friends & Family", 42},
				{"Clients", 18},
				{"Vendors", 12},
				{"Newsletter Subscribers", 156},
			}

			for _, g := range groups {
				fmt.Printf("  %s %s %s\n", cyan.Sprint("●"), boldWhite.Sprint(g.name), dim.Sprintf("(%d)", g.count))
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To manage your real contact groups: nylas auth login"))

			return nil
		},
	}
}

func newDemoGroupsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [group-id]",
		Short: "Show group details",
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := "Work Colleagues"
			if len(args) > 0 {
				groupName = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("👥 Demo Mode - Contact Group Details"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Printf("Group: %s\n", groupName)
			fmt.Printf("  Members:     25\n")
			fmt.Printf("  Created:     Jan 15, 2024\n")
			fmt.Printf("  Description: Team members and work contacts\n")
			fmt.Println()
			fmt.Println("Members:")
			fmt.Printf("  • John Smith (john@example.com)\n")
			fmt.Printf("  • Jane Doe (jane@example.com)\n")
			fmt.Printf("  • Bob Wilson (bob@example.com)\n")
			_, _ = dim.Printf("  ... and 22 more\n")
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoGroupsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a contact group",
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := "New Group"
			if len(args) > 0 {
				groupName = args[0]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Group '%s' would be created (demo mode)\n", groupName)
			_, _ = dim.Printf("  Group ID: group-demo-%d\n", time.Now().Unix())

			return nil
		},
	}
}

func newDemoGroupsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [group-id]",
		Short: "Delete a contact group",
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := "group-demo-123"
			if len(args) > 0 {
				groupID = args[0]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Group %s would be deleted (demo mode)\n", groupID)

			return nil
		},
	}
}

func newDemoGroupsAddMemberCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-member [group-id] [contact-id]",
		Short: "Add a contact to a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := "group-demo-123"
			contactID := "contact-demo-456"
			if len(args) > 0 {
				groupID = args[0]
			}
			if len(args) > 1 {
				contactID = args[1]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Contact %s would be added to group %s (demo mode)\n", contactID, groupID)

			return nil
		},
	}
}

func newDemoGroupsRemoveMemberCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-member [group-id] [contact-id]",
		Short: "Remove a contact from a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := "group-demo-123"
			contactID := "contact-demo-456"
			if len(args) > 0 {
				groupID = args[0]
			}
			if len(args) > 1 {
				contactID = args[1]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Contact %s would be removed from group %s (demo mode)\n", contactID, groupID)

			return nil
		},
	}
}

// ============================================================================
// PHOTO COMMAND
// ============================================================================

// newDemoContactsPhotoCmd creates the photo subcommand.
func newDemoContactsPhotoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "photo",
		Short: "Manage contact photos",
		Long:  "Demo commands for managing contact photos.",
	}

	cmd.AddCommand(newDemoPhotoGetCmd())
	cmd.AddCommand(newDemoPhotoSetCmd())
	cmd.AddCommand(newDemoPhotoRemoveCmd())

	return cmd
}

func newDemoPhotoGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [contact-id]",
		Short: "Get contact photo",
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := "contact-demo-123"
			if len(args) > 0 {
				contactID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📷 Demo Mode - Contact Photo"))
			fmt.Println()
			fmt.Printf("Contact ID: %s\n", contactID)
			fmt.Printf("Photo URL:  https://example.com/photos/%s.jpg\n", contactID)
			fmt.Printf("Size:       128x128\n")
			fmt.Printf("Format:     JPEG\n")

			return nil
		},
	}
}

func newDemoPhotoSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [contact-id] [photo-path]",
		Short: "Set contact photo",
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := "contact-demo-123"
			photoPath := "photo.jpg"
			if len(args) > 0 {
				contactID = args[0]
			}
			if len(args) > 1 {
				photoPath = args[1]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Photo '%s' would be set for contact %s (demo mode)\n", photoPath, contactID)

			return nil
		},
	}
}

func newDemoPhotoRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [contact-id]",
		Short: "Remove contact photo",
		RunE: func(cmd *cobra.Command, args []string) error {
			contactID := "contact-demo-123"
			if len(args) > 0 {
				contactID = args[0]
			}

			fmt.Println()
			_, _ = green.Printf("✓ Photo would be removed from contact %s (demo mode)\n", contactID)

			return nil
		},
	}
}

// ============================================================================
// SYNC COMMAND
// ============================================================================

// newDemoContactsSyncCmd simulates syncing contacts.
func newDemoContactsSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync contacts",
		Long:  "Demo contact synchronization features.",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("🔄 Demo Mode - Contact Sync Status"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("  Last sync:     %s\n", green.Sprint("2 minutes ago"))
			fmt.Printf("  Total contacts: 247\n")
			fmt.Printf("  Added:         5\n")
			fmt.Printf("  Updated:       12\n")
			fmt.Printf("  Deleted:       2\n")
			fmt.Printf("  Sync status:   %s\n", green.Sprint("Up to date"))
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "now",
		Short: "Trigger sync now",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("🔄 Demo Mode - Contact Sync"))
			fmt.Println()
			fmt.Println("Syncing contacts...")
			fmt.Println()
			_, _ = green.Println("✓ Sync would be triggered (demo mode)")
			fmt.Printf("  Estimated time: 30 seconds\n")

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "export",
		Short: "Export contacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("📤 Demo Mode - Export Contacts"))
			fmt.Println()
			fmt.Println("Available export formats:")
			fmt.Printf("  • CSV  - Comma-separated values\n")
			fmt.Printf("  • VCF  - vCard format\n")
			fmt.Printf("  • JSON - JSON format\n")
			fmt.Println()
			_, _ = green.Println("✓ Export would be generated (demo mode)")

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "import",
		Short: "Import contacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("📥 Demo Mode - Import Contacts"))
			fmt.Println()
			fmt.Println("Supported import formats:")
			fmt.Printf("  • CSV  - Comma-separated values\n")
			fmt.Printf("  • VCF  - vCard format\n")
			fmt.Printf("  • JSON - JSON format\n")
			fmt.Println()
			_, _ = green.Println("✓ Import would be processed (demo mode)")

			return nil
		},
	})

	return cmd
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// printDemoContact prints a contact summary.
func printDemoContact(contact domain.Contact, showID bool) {
	name := fmt.Sprintf("%s %s", contact.GivenName, contact.Surname)
	name = strings.TrimSpace(name)

	fmt.Printf("  %s %s\n", "👤", boldWhite.Sprint(name))

	if len(contact.Emails) > 0 {
		fmt.Printf("    📧 %s\n", contact.Emails[0].Email)
	}

	if contact.CompanyName != "" || contact.JobTitle != "" {
		company := contact.CompanyName
		if contact.JobTitle != "" {
			if company != "" {
				company = contact.JobTitle + " at " + company
			} else {
				company = contact.JobTitle
			}
		}
		fmt.Printf("    💼 %s\n", dim.Sprint(company))
	}

	if len(contact.PhoneNumbers) > 0 {
		fmt.Printf("    📱 %s\n", contact.PhoneNumbers[0].Number)
	}

	if showID {
		_, _ = dim.Printf("    ID: %s\n", contact.ID)
	}

	fmt.Println()
}

// printDemoContactFull prints a full contact.
func printDemoContactFull(contact domain.Contact) {
	name := fmt.Sprintf("%s %s", contact.GivenName, contact.Surname)
	name = strings.TrimSpace(name)

	fmt.Println(strings.Repeat("─", 50))
	_, _ = boldWhite.Printf("Name: %s\n", name)

	if contact.CompanyName != "" {
		fmt.Printf("Company: %s\n", contact.CompanyName)
	}
	if contact.JobTitle != "" {
		fmt.Printf("Title: %s\n", contact.JobTitle)
	}

	if len(contact.Emails) > 0 {
		fmt.Println("\nEmails:")
		for _, e := range contact.Emails {
			emailType := e.Type
			if emailType == "" {
				emailType = "email"
			}
			fmt.Printf("  %s: %s\n", emailType, e.Email)
		}
	}

	if len(contact.PhoneNumbers) > 0 {
		fmt.Println("\nPhone Numbers:")
		for _, p := range contact.PhoneNumbers {
			phoneType := p.Type
			if phoneType == "" {
				phoneType = "phone"
			}
			fmt.Printf("  %s: %s\n", phoneType, p.Number)
		}
	}

	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
}
