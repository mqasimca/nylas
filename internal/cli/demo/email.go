package demo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
)

var (
	boldWhite = color.New(color.FgWhite, color.Bold)
	cyan      = color.New(color.FgCyan)
	yellow    = color.New(color.FgYellow)
	green     = color.New(color.FgGreen)
	dim       = color.New(color.Faint)
)

// newDemoEmailCmd creates the demo email command with subcommands.
func newDemoEmailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "Explore email features with sample data",
		Long:  "Demo email commands showing sample messages, threads, and simulated operations.",
	}

	// Core commands
	cmd.AddCommand(newDemoEmailListCmd())
	cmd.AddCommand(newDemoEmailReadCmd())
	cmd.AddCommand(newDemoEmailSendCmd())
	cmd.AddCommand(newDemoEmailSearchCmd())
	cmd.AddCommand(newDemoEmailMarkCmd())
	cmd.AddCommand(newDemoEmailDeleteCmd())

	// Subcommand groups
	cmd.AddCommand(newDemoEmailFoldersCmd())
	cmd.AddCommand(newDemoEmailThreadsCmd())
	cmd.AddCommand(newDemoEmailDraftsCmd())
	cmd.AddCommand(newDemoEmailAttachmentsCmd())
	cmd.AddCommand(newDemoEmailScheduledCmd())
	cmd.AddCommand(newDemoEmailSmartComposeCmd())
	cmd.AddCommand(newDemoEmailAICmd())

	return cmd
}

// newDemoEmailListCmd lists sample emails.
func newDemoEmailListCmd() *cobra.Command {
	var limit int
	var showID bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sample emails",
		Long:  "Display a list of realistic sample emails to explore the CLI output format.",
		Example: `  # List sample emails
  nylas demo email list

  # List with IDs shown
  nylas demo email list --id

  # Limit to 5 emails
  nylas demo email list --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			messages, err := client.GetMessages(ctx, "demo-grant", limit)
			if err != nil {
				return fmt.Errorf("failed to get demo messages: %w", err)
			}

			if limit > 0 && limit < len(messages) {
				messages = messages[:limit]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Sample Emails"))
			fmt.Println(dim.Sprint("These are sample emails for demonstration purposes."))
			fmt.Println()
			fmt.Printf("Found %d messages:\n\n", len(messages))

			for i, msg := range messages {
				printDemoMessageSummary(msg, i+1, showID)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To connect your real email: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of messages to show")
	cmd.Flags().BoolVar(&showID, "id", false, "Show message IDs")

	return cmd
}

// newDemoEmailReadCmd reads a sample email.
func newDemoEmailReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [message-id]",
		Short: "Read a sample email",
		Long:  "Display a sample email to see the full message format.",
		Example: `  # Read first sample email
  nylas demo email read

  # Read specific message
  nylas demo email read msg-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			messageID := "msg-001"
			if len(args) > 0 {
				messageID = args[0]
			}

			msg, err := client.GetMessage(ctx, "demo-grant", messageID)
			if err != nil {
				return fmt.Errorf("failed to get demo message: %w", err)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Sample Email"))
			fmt.Println()
			printDemoMessage(*msg)

			fmt.Println(dim.Sprint("To connect your real email: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// newDemoEmailSendCmd simulates sending an email.
func newDemoEmailSendCmd() *cobra.Command {
	var to string
	var subject string
	var body string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Simulate sending an email",
		Long: `Simulate sending an email to see how the send command works.

No actual email is sent - this is just a demonstration of the command flow.`,
		Example: `  # Simulate sending an email
  nylas demo email send --to user@example.com --subject "Hello" --body "Test message"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if to == "" {
				to = "recipient@example.com"
			}
			if subject == "" {
				subject = "Demo Email Subject"
			}
			if body == "" {
				body = "This is a demo email body.\n\nNo actual email was sent."
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Simulated Send"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Printf("To:      %s\n", to)
			boldWhite.Printf("Subject: %s\n", subject)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println(body)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			green.Println("✓ Email would be sent (demo mode - no actual email sent)")
			fmt.Println()
			fmt.Println(dim.Sprint("To send real emails, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient email address")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&body, "body", "", "Email body")

	return cmd
}

// printDemoMessageSummary prints a single-line message summary.
func printDemoMessageSummary(msg domain.Message, index int, showID bool) {
	status := " "
	if msg.Unread {
		status = cyan.Sprint("●")
	}

	star := " "
	if msg.Starred {
		star = yellow.Sprint("★")
	}

	from := formatDemoContacts(msg.From)
	if len(from) > 20 {
		from = from[:17] + "..."
	}

	subject := msg.Subject
	if len(subject) > 40 {
		subject = subject[:37] + "..."
	}

	dateStr := formatDemoTimeAgo(msg.Date)
	if len(dateStr) > 12 {
		dateStr = msg.Date.Format("Jan 2")
	}

	if showID {
		fmt.Printf("%s %s %-20s %-40s %s\n", status, star, from, subject, dim.Sprint(dateStr))
		dim.Printf("      ID: %s\n", msg.ID)
	} else {
		fmt.Printf("%s %s %-20s %-40s %s\n", status, star, from, subject, dim.Sprint(dateStr))
	}
}

// printDemoMessage prints a full message.
func printDemoMessage(msg domain.Message) {
	status := ""
	if msg.Unread {
		status += cyan.Sprint("●") + " "
	}
	if msg.Starred {
		status += yellow.Sprint("★") + " "
	}

	fmt.Println(strings.Repeat("─", 60))
	boldWhite.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("From:    %s\n", formatDemoContacts(msg.From))
	if len(msg.To) > 0 {
		fmt.Printf("To:      %s\n", formatDemoContacts(msg.To))
	}
	fmt.Printf("Date:    %s (%s)\n", msg.Date.Format("Jan 2, 2006 3:04 PM"), formatDemoTimeAgo(msg.Date))
	if status != "" {
		fmt.Printf("Status:  %s\n", status)
	}
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(msg.Body)
	fmt.Println()
}

// formatDemoContacts formats multiple contacts for display.
func formatDemoContacts(contacts []domain.EmailParticipant) string {
	names := make([]string, len(contacts))
	for i, c := range contacts {
		if c.Name != "" {
			names[i] = c.Name
		} else {
			names[i] = c.Email
		}
	}
	return strings.Join(names, ", ")
}

// formatDemoTimeAgo formats a time as a relative string.
func formatDemoTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 48*time.Hour {
		return "yesterday"
	}
	days := int(diff.Hours() / 24)
	return fmt.Sprintf("%d days ago", days)
}

// newDemoEmailSearchCmd searches sample emails.
func newDemoEmailSearchCmd() *cobra.Command {
	var query string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search sample emails",
		Long:  "Search through sample emails to see how search works.",
		Example: `  # Search for emails
  nylas demo email search --query "meeting"
  nylas demo email search -q "project"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			messages, _ := client.GetMessages(ctx, "demo-grant", 10)

			// Filter by query if provided
			if query != "" {
				var filtered []domain.Message
				for _, msg := range messages {
					if strings.Contains(strings.ToLower(msg.Subject), strings.ToLower(query)) ||
						strings.Contains(strings.ToLower(msg.Body), strings.ToLower(query)) {
						filtered = append(filtered, msg)
					}
				}
				messages = filtered
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🔍 Demo Mode - Email Search"))
			if query != "" {
				fmt.Printf("Searching for: %s\n", boldWhite.Sprint(query))
			}
			fmt.Println()
			fmt.Printf("Found %d messages:\n\n", len(messages))

			for i, msg := range messages {
				printDemoMessageSummary(msg, i+1, false)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To search your real emails: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Search query")

	return cmd
}

// newDemoEmailMarkCmd marks sample emails.
func newDemoEmailMarkCmd() *cobra.Command {
	var read, unread, starred, unstarred bool

	cmd := &cobra.Command{
		Use:   "mark [message-id]",
		Short: "Mark sample emails (simulated)",
		Long:  "Simulate marking emails as read/unread/starred.",
		Example: `  # Mark as read
  nylas demo email mark msg-001 --read

  # Mark as starred
  nylas demo email mark msg-001 --starred`,
		RunE: func(cmd *cobra.Command, args []string) error {
			messageID := "msg-001"
			if len(args) > 0 {
				messageID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Mark Email (Simulated)"))
			fmt.Println()

			if read {
				green.Printf("✓ Message %s would be marked as read\n", messageID)
			}
			if unread {
				green.Printf("✓ Message %s would be marked as unread\n", messageID)
			}
			if starred {
				green.Printf("✓ Message %s would be starred\n", messageID)
			}
			if unstarred {
				green.Printf("✓ Message %s would be unstarred\n", messageID)
			}

			if !read && !unread && !starred && !unstarred {
				fmt.Println("No action specified. Use --read, --unread, --starred, or --unstarred")
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To manage your real emails: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().BoolVar(&read, "read", false, "Mark as read")
	cmd.Flags().BoolVar(&unread, "unread", false, "Mark as unread")
	cmd.Flags().BoolVar(&starred, "starred", false, "Mark as starred")
	cmd.Flags().BoolVar(&unstarred, "unstarred", false, "Remove star")

	return cmd
}

// newDemoEmailDeleteCmd deletes sample emails (simulated).
func newDemoEmailDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [message-id]",
		Short: "Delete sample email (simulated)",
		Long:  "Simulate deleting an email.",
		Example: `  # Delete an email
  nylas demo email delete msg-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			messageID := "msg-001"
			if len(args) > 0 {
				messageID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Delete Email (Simulated)"))
			fmt.Println()
			green.Printf("✓ Message %s would be deleted\n", messageID)
			fmt.Println()
			fmt.Println(dim.Sprint("To manage your real emails: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// newDemoEmailFoldersCmd manages sample folders.
func newDemoEmailFoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage sample email folders",
		Long:  "Demo folder commands showing sample folders.",
	}

	cmd.AddCommand(newDemoEmailFoldersListCmd())
	cmd.AddCommand(newDemoEmailFoldersCreateCmd())
	cmd.AddCommand(newDemoEmailFoldersDeleteCmd())

	return cmd
}

func newDemoEmailFoldersListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List sample folders",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			folders, _ := client.GetFolders(ctx, "demo-grant")

			fmt.Println()
			fmt.Println(dim.Sprint("📁 Demo Mode - Sample Folders"))
			fmt.Println()

			for _, f := range folders {
				system := ""
				if f.SystemFolder != "" {
					system = dim.Sprintf(" (%s)", f.SystemFolder)
				}
				fmt.Printf("  📁 %-15s %s%s\n", f.Name, dim.Sprintf("%d items", f.TotalCount), system)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To manage your real folders: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailFoldersCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a folder (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = "New Folder"
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📁 Demo Mode - Create Folder (Simulated)"))
			fmt.Println()
			green.Printf("✓ Folder '%s' would be created\n", name)
			fmt.Println()
			fmt.Println(dim.Sprint("To create real folders: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Folder name")

	return cmd
}

func newDemoEmailFoldersDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [folder-id]",
		Short: "Delete a folder (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			folderID := "work"
			if len(args) > 0 {
				folderID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📁 Demo Mode - Delete Folder (Simulated)"))
			fmt.Println()
			green.Printf("✓ Folder '%s' would be deleted\n", folderID)
			fmt.Println()
			fmt.Println(dim.Sprint("To manage real folders: nylas auth login"))

			return nil
		},
	}
}

// newDemoEmailThreadsCmd manages sample threads.
func newDemoEmailThreadsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threads",
		Short: "Manage sample email threads",
		Long:  "Demo thread commands showing sample conversations.",
	}

	cmd.AddCommand(newDemoEmailThreadsListCmd())
	cmd.AddCommand(newDemoEmailThreadsReadCmd())

	return cmd
}

func newDemoEmailThreadsListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sample threads",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			threads, _ := client.GetThreads(ctx, "demo-grant", nil)

			if limit > 0 && limit < len(threads) {
				threads = threads[:limit]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Sample Threads"))
			fmt.Println()
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

				subject := t.Subject
				if len(subject) > 40 {
					subject = subject[:37] + "..."
				}

				fmt.Printf("%s %s %-40s %s\n", status, star, subject, dim.Sprintf("%d messages", len(t.MessageIDs)))
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To view your real threads: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of threads to show")

	return cmd
}

func newDemoEmailThreadsReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read [thread-id]",
		Short: "Read a sample thread",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			threadID := "thread-001"
			if len(args) > 0 {
				threadID = args[0]
			}

			thread, _ := client.GetThread(ctx, "demo-grant", threadID)

			fmt.Println()
			fmt.Println(dim.Sprint("📧 Demo Mode - Sample Thread"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			boldWhite.Printf("Subject: %s\n", thread.Subject)
			fmt.Printf("Messages: %d\n", len(thread.MessageIDs))
			fmt.Printf("Participants: %s\n", formatDemoParticipants(thread.Participants))
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(dim.Sprint("To view your real threads: nylas auth login"))

			return nil
		},
	}
}

func formatDemoParticipants(participants []domain.EmailParticipant) string {
	names := make([]string, len(participants))
	for i, p := range participants {
		if p.Name != "" {
			names[i] = p.Name
		} else {
			names[i] = p.Email
		}
	}
	return strings.Join(names, ", ")
}

// newDemoEmailDraftsCmd manages sample drafts.
func newDemoEmailDraftsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drafts",
		Short: "Manage sample email drafts",
		Long:  "Demo draft commands showing sample drafts.",
	}

	cmd.AddCommand(newDemoEmailDraftsListCmd())
	cmd.AddCommand(newDemoEmailDraftsCreateCmd())
	cmd.AddCommand(newDemoEmailDraftsDeleteCmd())
	cmd.AddCommand(newDemoEmailDraftsSendCmd())

	return cmd
}

func newDemoEmailDraftsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List sample drafts",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			drafts, _ := client.GetDrafts(ctx, "demo-grant", 10)

			fmt.Println()
			fmt.Println(dim.Sprint("📝 Demo Mode - Sample Drafts"))
			fmt.Println()
			fmt.Printf("Found %d drafts:\n\n", len(drafts))

			for _, d := range drafts {
				to := ""
				if len(d.To) > 0 {
					to = d.To[0].Email
				}
				fmt.Printf("  📝 %s\n", boldWhite.Sprint(d.Subject))
				fmt.Printf("     To: %s\n", to)
				dim.Printf("     ID: %s\n", d.ID)
				fmt.Println()
			}

			fmt.Println(dim.Sprint("To manage your real drafts: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailDraftsCreateCmd() *cobra.Command {
	var to, subject, body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("📝 Demo Mode - Create Draft (Simulated)"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("To:      %s\n", to)
			fmt.Printf("Subject: %s\n", subject)
			fmt.Println(strings.Repeat("─", 50))
			if body != "" {
				fmt.Println(body)
				fmt.Println(strings.Repeat("─", 50))
			}
			fmt.Println()
			green.Println("✓ Draft would be created with ID: draft-demo-new")
			fmt.Println()
			fmt.Println(dim.Sprint("To create real drafts: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "recipient@example.com", "Recipient email")
	cmd.Flags().StringVar(&subject, "subject", "Draft Subject", "Email subject")
	cmd.Flags().StringVar(&body, "body", "", "Email body")

	return cmd
}

func newDemoEmailDraftsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [draft-id]",
		Short: "Delete a draft (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			draftID := "draft-001"
			if len(args) > 0 {
				draftID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📝 Demo Mode - Delete Draft (Simulated)"))
			fmt.Println()
			green.Printf("✓ Draft '%s' would be deleted\n", draftID)
			fmt.Println()
			fmt.Println(dim.Sprint("To manage real drafts: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailDraftsSendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send [draft-id]",
		Short: "Send a draft (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			draftID := "draft-001"
			if len(args) > 0 {
				draftID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📝 Demo Mode - Send Draft (Simulated)"))
			fmt.Println()
			green.Printf("✓ Draft '%s' would be sent\n", draftID)
			fmt.Println()
			fmt.Println(dim.Sprint("To send real drafts: nylas auth login"))

			return nil
		},
	}
}

// newDemoEmailAttachmentsCmd manages sample attachments.
func newDemoEmailAttachmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachments",
		Short: "Manage sample email attachments",
		Long:  "Demo attachment commands.",
	}

	cmd.AddCommand(newDemoEmailAttachmentsListCmd())
	cmd.AddCommand(newDemoEmailAttachmentsDownloadCmd())

	return cmd
}

func newDemoEmailAttachmentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [message-id]",
		Short: "List attachments for a message",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			messageID := "msg-001"
			if len(args) > 0 {
				messageID = args[0]
			}

			attachments, _ := client.ListAttachments(ctx, "demo-grant", messageID)

			fmt.Println()
			fmt.Println(dim.Sprint("📎 Demo Mode - Sample Attachments"))
			fmt.Printf("Message: %s\n\n", messageID)

			for _, a := range attachments {
				fmt.Printf("  📎 %s\n", boldWhite.Sprint(a.Filename))
				fmt.Printf("     Type: %s\n", a.ContentType)
				fmt.Printf("     Size: %s\n", formatDemoBytes(a.Size))
				dim.Printf("     ID: %s\n", a.ID)
				fmt.Println()
			}

			fmt.Println(dim.Sprint("To view real attachments: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailAttachmentsDownloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "download [attachment-id]",
		Short: "Download an attachment (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			attachmentID := "attach-001"
			if len(args) > 0 {
				attachmentID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📎 Demo Mode - Download Attachment (Simulated)"))
			fmt.Println()
			green.Printf("✓ Attachment '%s' would be downloaded to current directory\n", attachmentID)
			fmt.Println()
			fmt.Println(dim.Sprint("To download real attachments: nylas auth login"))

			return nil
		},
	}
}

func formatDemoBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// newDemoEmailScheduledCmd manages scheduled messages.
func newDemoEmailScheduledCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduled",
		Short: "Manage scheduled messages",
		Long:  "Demo scheduled message commands.",
	}

	cmd.AddCommand(newDemoEmailScheduledListCmd())
	cmd.AddCommand(newDemoEmailScheduledCancelCmd())

	return cmd
}

func newDemoEmailScheduledListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			scheduled, _ := client.ListScheduledMessages(ctx, "demo-grant")

			fmt.Println()
			fmt.Println(dim.Sprint("⏰ Demo Mode - Scheduled Messages"))
			fmt.Println()
			fmt.Printf("Found %d scheduled messages:\n\n", len(scheduled))

			for _, s := range scheduled {
				sendTime := time.Unix(s.CloseTime, 0)
				fmt.Printf("  ⏰ %s\n", boldWhite.Sprint(s.ScheduleID))
				fmt.Printf("     Status: %s\n", s.Status)
				fmt.Printf("     Sends at: %s\n", sendTime.Format("Jan 2, 2006 3:04 PM"))
				fmt.Println()
			}

			fmt.Println(dim.Sprint("To manage scheduled messages: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailScheduledCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel [schedule-id]",
		Short: "Cancel a scheduled message (simulated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleID := "schedule-001"
			if len(args) > 0 {
				scheduleID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("⏰ Demo Mode - Cancel Scheduled Message (Simulated)"))
			fmt.Println()
			green.Printf("✓ Scheduled message '%s' would be cancelled\n", scheduleID)
			fmt.Println()
			fmt.Println(dim.Sprint("To manage scheduled messages: nylas auth login"))

			return nil
		},
	}
}

// newDemoEmailSmartComposeCmd provides AI smart compose demo.
func newDemoEmailSmartComposeCmd() *cobra.Command {
	var prompt string

	cmd := &cobra.Command{
		Use:   "smart-compose",
		Short: "AI-powered email composition (demo)",
		Long:  "Demo AI smart compose generating sample email drafts.",
		Example: `  # Generate an email draft
  nylas demo email smart-compose --prompt "Thank the team for their hard work"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			if prompt == "" {
				prompt = "write a follow-up email"
			}

			suggestion, _ := client.SmartCompose(ctx, "demo-grant", &domain.SmartComposeRequest{Prompt: prompt})

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Smart Compose"))
			fmt.Printf("Prompt: %s\n\n", boldWhite.Sprint(prompt))
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println(suggestion.Suggestion)
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(dim.Sprint("To use AI compose with your emails: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt for AI composition")

	return cmd
}

// newDemoEmailAICmd provides AI email features demo.
func newDemoEmailAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered email features (demo)",
		Long:  "Demo AI email intelligence features.",
	}

	cmd.AddCommand(newDemoEmailAISummarizeCmd())
	cmd.AddCommand(newDemoEmailAIExtractCmd())

	return cmd
}

func newDemoEmailAISummarizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summarize [message-id]",
		Short: "Summarize an email with AI (demo)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Email Summary"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			boldWhite.Println("Summary:")
			fmt.Println("This email discusses the Q4 planning meeting action items.")
			fmt.Println("Key points:")
			fmt.Println("  • Review Q4 roadmap by Friday")
			fmt.Println("  • Submit budget proposals")
			fmt.Println("  • Schedule 1:1s with new team members")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(dim.Sprint("To use AI features with your emails: nylas auth login"))

			return nil
		},
	}
}

func newDemoEmailAIExtractCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extract [message-id]",
		Short: "Extract key info from email with AI (demo)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Extract Key Info"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			boldWhite.Println("Extracted Information:")
			fmt.Println("  Action Items: 3")
			fmt.Println("  Deadlines: Friday (Q4 roadmap review)")
			fmt.Println("  People Mentioned: Sarah Chen, team members")
			fmt.Println("  Sentiment: Professional, positive")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(dim.Sprint("To use AI features with your emails: nylas auth login"))

			return nil
		},
	}
}
