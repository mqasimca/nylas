package email

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newThreadsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threads",
		Short: "Manage email threads/conversations",
		Long:  "List and manage email threads (conversations).",
	}

	cmd.AddCommand(newThreadsListCmd())
	cmd.AddCommand(newThreadsShowCmd())

	return cmd
}

func newThreadsListCmd() *cobra.Command {
	var limit int
	var unread bool
	var starred bool
	var subject string

	cmd := &cobra.Command{
		Use:   "list [grant-id]",
		Short: "List email threads",
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

			params := &domain.ThreadQueryParams{
				Limit: limit,
			}

			if cmd.Flags().Changed("unread") {
				params.Unread = &unread
			}
			if cmd.Flags().Changed("starred") {
				params.Starred = &starred
			}
			if subject != "" {
				params.Subject = subject
			}

			threads, err := client.GetThreads(ctx, grantID, params)
			if err != nil {
				return fmt.Errorf("failed to get threads: %w", err)
			}

			if len(threads) == 0 {
				fmt.Println("No threads found.")
				return nil
			}

			fmt.Printf("Found %d threads:\n\n", len(threads))

			for _, t := range threads {
				status := " "
				if t.Unread {
					status = cyan.Sprint("●")
				}

				star := " "
				if t.Starred {
					star = yellow.Sprint("★")
				}

				attach := " "
				if t.HasAttachments {
					attach = "📎"
				}

				// Format participants
				participants := formatContacts(t.Participants)
				if len(participants) > 25 {
					participants = participants[:22] + "..."
				}

				subj := t.Subject
				if len(subj) > 35 {
					subj = subj[:32] + "..."
				}

				msgCount := fmt.Sprintf("(%d)", len(t.MessageIDs))
				dateStr := formatTimeAgo(t.LatestMessageRecvDate)

				fmt.Printf("%s %s %s %-25s %-35s %-5s %s\n",
					status, star, attach, participants, subj, msgCount, dim.Sprint(dateStr))
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of threads to fetch")
	cmd.Flags().BoolVarP(&unread, "unread", "u", false, "Only show unread threads")
	cmd.Flags().BoolVarP(&starred, "starred", "s", false, "Only show starred threads")
	cmd.Flags().StringVar(&subject, "subject", "", "Filter by subject")

	return cmd
}

func newThreadsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <thread-id> [grant-id]",
		Short: "Show thread details",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			threadID := args[0]

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

			thread, err := client.GetThread(ctx, grantID, threadID)
			if err != nil {
				return fmt.Errorf("failed to get thread: %w", err)
			}

			// Print thread details
			fmt.Println("════════════════════════════════════════════════════════════")
			boldWhite.Printf("Thread: %s\n", thread.Subject)
			fmt.Println("════════════════════════════════════════════════════════════")

			fmt.Printf("Participants: %s\n", formatContacts(thread.Participants))
			fmt.Printf("Messages:     %d\n", len(thread.MessageIDs))
			if len(thread.DraftIDs) > 0 {
				fmt.Printf("Drafts:       %d\n", len(thread.DraftIDs))
			}

			status := []string{}
			if thread.Unread {
				status = append(status, cyan.Sprint("unread"))
			}
			if thread.Starred {
				status = append(status, yellow.Sprint("starred"))
			}
			if thread.HasAttachments {
				status = append(status, "has attachments")
			}
			if len(status) > 0 {
				fmt.Printf("Status:       %s\n", formatContacts(nil))
			}

			fmt.Printf("\nFirst message: %s\n", thread.EarliestMessageDate.Format("Jan 2, 2006 3:04 PM"))
			fmt.Printf("Latest:        %s\n", thread.LatestMessageRecvDate.Format("Jan 2, 2006 3:04 PM"))

			fmt.Println("\nSnippet:")
			fmt.Println(thread.Snippet)

			fmt.Println("\nMessage IDs:")
			for i, msgID := range thread.MessageIDs {
				fmt.Printf("  %d. %s\n", i+1, msgID)
			}

			return nil
		},
	}
}
