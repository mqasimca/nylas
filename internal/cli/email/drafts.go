package email

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newDraftsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drafts",
		Short: "Manage email drafts",
		Long:  "List, create, edit, and send draft emails.",
	}

	cmd.AddCommand(newDraftsListCmd())
	cmd.AddCommand(newDraftsCreateCmd())
	cmd.AddCommand(newDraftsShowCmd())
	cmd.AddCommand(newDraftsSendCmd())
	cmd.AddCommand(newDraftsDeleteCmd())

	return cmd
}

func newDraftsListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list [grant-id]",
		Short: "List drafts",
		Args:  cobra.MaximumNArgs(1),
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

			drafts, err := client.GetDrafts(ctx, grantID, limit)
			if err != nil {
				return fmt.Errorf("failed to get drafts: %w", err)
			}

			if len(drafts) == 0 {
				fmt.Println("No drafts found.")
				return nil
			}

			fmt.Printf("Found %d drafts:\n\n", len(drafts))
			fmt.Printf("%-15s %-25s %-35s %s\n", "ID", "TO", "SUBJECT", "UPDATED")
			fmt.Println("--------------------------------------------------------------------------------")

			for _, d := range drafts {
				toStr := ""
				if len(d.To) > 0 {
					toStr = formatContacts(d.To)
				}
				if len(toStr) > 23 {
					toStr = toStr[:20] + "..."
				}

				subj := d.Subject
				if subj == "" {
					subj = "(no subject)"
				}
				if len(subj) > 33 {
					subj = subj[:30] + "..."
				}

				// Show first 12 chars of ID
				idShort := d.ID
				if len(idShort) > 12 {
					idShort = idShort[:12] + "..."
				}

				dateStr := formatTimeAgo(d.UpdatedAt)

				fmt.Printf("%-15s %-25s %-35s %s\n", idShort, toStr, subj, dim.Sprint(dateStr))
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of drafts to fetch")

	return cmd
}

func newDraftsCreateCmd() *cobra.Command {
	var to []string
	var cc []string
	var subject string
	var body string
	var replyTo string

	cmd := &cobra.Command{
		Use:   "create [grant-id]",
		Short: "Create a new draft",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			// Interactive mode if nothing provided
			if len(to) == 0 && subject == "" && body == "" {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("To (comma-separated, optional): ")
				input, _ := reader.ReadString('\n')
				to = parseEmails(strings.TrimSpace(input))

				fmt.Print("Subject: ")
				subject, _ = reader.ReadString('\n')
				subject = strings.TrimSpace(subject)

				fmt.Println("Body (end with a line containing only '.'):")
				var bodyLines []string
				for {
					line, _ := reader.ReadString('\n')
					line = strings.TrimSuffix(line, "\n")
					if line == "." {
						break
					}
					bodyLines = append(bodyLines, line)
				}
				body = strings.Join(bodyLines, "\n")
			}

			ctx, cancel := createContext()
			defer cancel()

			req := &domain.CreateDraftRequest{
				Subject:      subject,
				Body:         body,
				To:           parseContacts(to),
				ReplyToMsgID: replyTo,
			}

			if len(cc) > 0 {
				req.Cc = parseContacts(cc)
			}

			draft, err := client.CreateDraft(ctx, grantID, req)
			if err != nil {
				return fmt.Errorf("failed to create draft: %w", err)
			}

			printSuccess("Draft created! ID: %s", draft.ID)
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&to, "to", "t", nil, "Recipient email addresses")
	cmd.Flags().StringSliceVar(&cc, "cc", nil, "CC email addresses")
	cmd.Flags().StringVarP(&subject, "subject", "s", "", "Email subject")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Email body")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Message ID to reply to")

	return cmd
}

func newDraftsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <draft-id> [grant-id]",
		Short: "Show draft details",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			draftID := args[0]

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

			ctx, cancel := createContext()
			defer cancel()

			draft, err := client.GetDraft(ctx, grantID, draftID)
			if err != nil {
				return fmt.Errorf("failed to get draft: %w", err)
			}

			fmt.Println("════════════════════════════════════════════════════════════")
			boldWhite.Printf("Draft: %s\n", draft.Subject)
			fmt.Println("════════════════════════════════════════════════════════════")

			fmt.Printf("ID:      %s\n", draft.ID)
			if len(draft.To) > 0 {
				fmt.Printf("To:      %s\n", formatContacts(draft.To))
			}
			if len(draft.Cc) > 0 {
				fmt.Printf("Cc:      %s\n", formatContacts(draft.Cc))
			}
			fmt.Printf("Updated: %s\n", draft.UpdatedAt.Format("Jan 2, 2006 3:04 PM"))

			if draft.Body != "" {
				fmt.Println("\nBody:")
				fmt.Println("────────────────────────────────────────────────────────────")
				fmt.Println(stripHTML(draft.Body))
			}

			return nil
		},
	}
}

func newDraftsSendCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "send <draft-id> [grant-id]",
		Short: "Send a draft",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			draftID := args[0]

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

			// Get draft info first
			ctx, cancel := createContext()
			defer cancel()

			draft, err := client.GetDraft(ctx, grantID, draftID)
			if err != nil {
				return fmt.Errorf("failed to get draft: %w", err)
			}

			// Confirmation
			if !force {
				fmt.Println("Send this draft?")
				fmt.Printf("  To:      %s\n", formatContacts(draft.To))
				fmt.Printf("  Subject: %s\n", draft.Subject)
				fmt.Print("\n[y/N]: ")

				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" && confirm != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			msg, err := client.SendDraft(ctx, grantID, draftID)
			if err != nil {
				return fmt.Errorf("failed to send draft: %w", err)
			}

			printSuccess("Draft sent! Message ID: %s", msg.ID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newDraftsDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <draft-id> [grant-id]",
		Short: "Delete a draft",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			draftID := args[0]

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

			if !force {
				fmt.Printf("Delete draft %s? [y/N]: ", draftID)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" && confirm != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			err = client.DeleteDraft(ctx, grantID, draftID)
			if err != nil {
				return fmt.Errorf("failed to delete draft: %w", err)
			}

			printSuccess("Draft deleted")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
