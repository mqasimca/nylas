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

// newDemoNotetakerCmd creates the demo notetaker command with subcommands.
func newDemoNotetakerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notetaker",
		Short: "Explore AI notetaker features with sample data",
		Long:  "Demo notetaker commands showing sample meeting recordings and transcripts.",
	}

	cmd.AddCommand(newDemoNotetakerListCmd())
	cmd.AddCommand(newDemoNotetakerShowCmd())
	cmd.AddCommand(newDemoNotetakerCreateCmd())
	cmd.AddCommand(newDemoNotetakerDeleteCmd())
	cmd.AddCommand(newDemoNotetakerMediaCmd())

	return cmd
}

// ============================================================================
// LIST COMMAND
// ============================================================================

// newDemoNotetakerListCmd lists sample notetakers.
func newDemoNotetakerListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sample notetakers",
		Long:  "Display a list of sample AI notetaker sessions.",
		Example: `  # List sample notetakers
  nylas demo notetaker list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			notetakers, err := client.ListNotetakers(ctx, "demo-grant", nil)
			if err != nil {
				return fmt.Errorf("failed to get demo notetakers: %w", err)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Sample AI Notetakers"))
			fmt.Println(dim.Sprint("These are sample notetaker sessions for demonstration purposes."))
			fmt.Println()
			fmt.Printf("Found %d notetakers:\n\n", len(notetakers))

			for _, nt := range notetakers {
				printDemoNotetaker(nt)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To use AI notetakers on your meetings: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// ============================================================================
// SHOW COMMAND
// ============================================================================

// newDemoNotetakerShowCmd shows a sample notetaker.
func newDemoNotetakerShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [notetaker-id]",
		Aliases: []string{"read"},
		Short:   "Show a sample notetaker session",
		Long:    "Display a sample notetaker session with details.",
		Example: `  # Show first sample notetaker
  nylas demo notetaker show

  # Show specific notetaker
  nylas demo notetaker show notetaker-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			notetakerID := "notetaker-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			nt, err := client.GetNotetaker(ctx, "demo-grant", notetakerID)
			if err != nil {
				return fmt.Errorf("failed to get demo notetaker: %w", err)
			}

			// Get media info
			media, _ := client.GetNotetakerMedia(ctx, "demo-grant", notetakerID)

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Sample Notetaker Session"))
			fmt.Println()
			printDemoNotetakerFull(*nt, media)

			fmt.Println(dim.Sprint("To use AI notetakers on your meetings: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// ============================================================================
// CREATE COMMAND
// ============================================================================

// newDemoNotetakerCreateCmd simulates creating a notetaker.
func newDemoNotetakerCreateCmd() *cobra.Command {
	var meetingLink string
	var name string
	var joinAt string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Simulate creating a notetaker",
		Long: `Simulate creating an AI notetaker to join a meeting.

No actual notetaker is created - this is just a demonstration of the command flow.`,
		Example: `  # Create notetaker for a Zoom meeting
  nylas demo notetaker create --meeting-link "https://zoom.us/j/123456789"

  # Create with custom name
  nylas demo notetaker create --meeting-link "https://meet.google.com/abc-defg-hij" --name "Project Review Bot"

  # Schedule notetaker for later
  nylas demo notetaker create --meeting-link "https://zoom.us/j/123" --join-at "2024-01-15 10:00"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if meetingLink == "" {
				meetingLink = "https://zoom.us/j/123456789"
			}
			if name == "" {
				name = "Nylas Notetaker"
			}

			// Detect meeting provider
			var provider string
			if strings.Contains(meetingLink, "zoom.us") {
				provider = "Zoom"
			} else if strings.Contains(meetingLink, "meet.google.com") {
				provider = "Google Meet"
			} else if strings.Contains(meetingLink, "teams.microsoft.com") {
				provider = "Microsoft Teams"
			} else {
				provider = "Unknown"
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Simulated Notetaker Creation"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Printf("Name:         %s\n", name)
			fmt.Printf("Meeting Link: %s\n", meetingLink)
			fmt.Printf("Provider:     %s\n", provider)
			if joinAt != "" {
				fmt.Printf("Join At:      %s\n", joinAt)
			} else {
				fmt.Printf("Join At:      %s\n", green.Sprint("Immediately"))
			}
			fmt.Println()
			fmt.Println("Settings:")
			fmt.Printf("  Recording:     %s\n", green.Sprint("Enabled"))
			fmt.Printf("  Transcription: %s\n", green.Sprint("Enabled"))
			fmt.Printf("  Summary:       %s\n", green.Sprint("Enabled"))
			fmt.Printf("  Action Items:  %s\n", green.Sprint("Enabled"))
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			_, _ = green.Println("✓ Notetaker would be created (demo mode - no actual notetaker created)")
			_, _ = dim.Printf("  Notetaker ID: notetaker-demo-%d\n", time.Now().Unix())
			fmt.Println()
			fmt.Println(dim.Sprint("To create real notetakers, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&meetingLink, "meeting-link", "", "Video conference meeting link")
	cmd.Flags().StringVar(&name, "name", "", "Display name for the notetaker bot")
	cmd.Flags().StringVar(&joinAt, "join-at", "", "When to join the meeting (optional)")

	return cmd
}

// ============================================================================
// DELETE COMMAND
// ============================================================================

// newDemoNotetakerDeleteCmd simulates deleting/cancelling a notetaker.
func newDemoNotetakerDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [notetaker-id]",
		Short: "Simulate deleting/cancelling a notetaker",
		Long:  "Simulate deleting or cancelling an AI notetaker session.",
		Example: `  # Delete/cancel a notetaker
  nylas demo notetaker delete notetaker-001

  # Force delete without confirmation
  nylas demo notetaker delete notetaker-001 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-123"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Simulated Notetaker Deletion"))
			fmt.Println()

			if !force {
				_, _ = yellow.Println("⚠ Would prompt for confirmation in real mode")
			}

			fmt.Printf("Notetaker ID: %s\n", notetakerID)
			fmt.Println()
			_, _ = green.Println("✓ Notetaker would be cancelled/deleted (demo mode - no actual deletion)")
			fmt.Println("  If the notetaker was in a meeting, it would leave immediately")
			fmt.Println()
			fmt.Println(dim.Sprint("To manage real notetakers, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

// ============================================================================
// MEDIA COMMAND
// ============================================================================

// newDemoNotetakerMediaCmd creates the media subcommand group.
func newDemoNotetakerMediaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Access notetaker media (recordings, transcripts)",
		Long:  "Demo commands for accessing notetaker recordings and transcripts.",
	}

	cmd.AddCommand(newDemoMediaShowCmd())
	cmd.AddCommand(newDemoMediaDownloadCmd())
	cmd.AddCommand(newDemoMediaTranscriptCmd())
	cmd.AddCommand(newDemoMediaSummaryCmd())
	cmd.AddCommand(newDemoMediaActionItemsCmd())

	return cmd
}

func newDemoMediaShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [notetaker-id]",
		Short: "Show available media for a notetaker",
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Notetaker Media"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Printf("Notetaker: %s\n", notetakerID)
			fmt.Println()

			fmt.Println("📁 Available Media:")
			fmt.Println()
			fmt.Printf("  %s Recording\n", green.Sprint("●"))
			fmt.Printf("    Format:   MP4\n")
			fmt.Printf("    Size:     245.6 MB\n")
			fmt.Printf("    Duration: 45:32\n")
			_, _ = dim.Printf("    URL:      https://media.example.com/recordings/%s.mp4\n", notetakerID)
			fmt.Println()

			fmt.Printf("  %s Transcript\n", green.Sprint("●"))
			fmt.Printf("    Format:   VTT\n")
			fmt.Printf("    Size:     128.4 KB\n")
			_, _ = dim.Printf("    URL:      https://media.example.com/transcripts/%s.vtt\n", notetakerID)
			fmt.Println()

			fmt.Printf("  %s Summary\n", green.Sprint("●"))
			fmt.Printf("    Format:   JSON\n")
			fmt.Printf("    Size:     4.2 KB\n")
			fmt.Println()

			fmt.Printf("  %s Action Items\n", green.Sprint("●"))
			fmt.Printf("    Count:    5 items\n")

			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoMediaDownloadCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "download [notetaker-id]",
		Short: "Download notetaker recording",
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			if output == "" {
				output = fmt.Sprintf("%s-recording.mp4", notetakerID)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Download Recording"))
			fmt.Println()
			fmt.Printf("Notetaker: %s\n", notetakerID)
			fmt.Printf("Output:    %s\n", output)
			fmt.Println()
			_, _ = green.Println("✓ Recording would be downloaded (demo mode)")
			fmt.Printf("  Size: 245.6 MB\n")
			fmt.Printf("  Duration: 45:32\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")

	return cmd
}

func newDemoMediaTranscriptCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "transcript [notetaker-id]",
		Short: "Get meeting transcript",
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			if format == "" {
				format = "text"
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - Meeting Transcript"))
			fmt.Println()
			fmt.Printf("Notetaker: %s\n", notetakerID)
			fmt.Printf("Format:    %s\n", format)
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()

			// Sample transcript
			fmt.Println("[00:00:05] John: Good morning everyone, let's get started with the standup.")
			fmt.Println()
			fmt.Println("[00:00:12] Sarah: Sure! Yesterday I finished the authentication module.")
			fmt.Println("                 Today I'm moving on to the API integration.")
			fmt.Println()
			fmt.Println("[00:00:28] Mike: I'm still working on the database optimization.")
			fmt.Println("                Should be done by end of day.")
			fmt.Println()
			fmt.Println("[00:00:45] John: Great progress team! Any blockers?")
			fmt.Println()
			_, _ = dim.Println("... (transcript continues)")
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "Output format (text, vtt, json)")

	return cmd
}

func newDemoMediaSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary [notetaker-id]",
		Short: "Get AI-generated meeting summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Meeting Summary"))
			fmt.Println()
			fmt.Printf("Notetaker: %s\n", notetakerID)
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Println("Meeting Summary")
			fmt.Println()

			fmt.Println("📋 Overview:")
			fmt.Println("   The team conducted their daily standup meeting to discuss")
			fmt.Println("   progress on the current sprint. All team members provided")
			fmt.Println("   updates on their tasks and no major blockers were identified.")
			fmt.Println()

			fmt.Println("👥 Participants:")
			fmt.Println("   • John (Meeting Host)")
			fmt.Println("   • Sarah")
			fmt.Println("   • Mike")
			fmt.Println()

			fmt.Println("📝 Key Points:")
			fmt.Println("   • Authentication module completed")
			fmt.Println("   • API integration work starting today")
			fmt.Println("   • Database optimization in progress")
			fmt.Println("   • Sprint on track for completion")
			fmt.Println()

			fmt.Println("⏱️  Duration: 15 minutes")
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoMediaActionItemsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "action-items [notetaker-id]",
		Short: "Get AI-extracted action items",
		RunE: func(cmd *cobra.Command, args []string) error {
			notetakerID := "notetaker-demo-001"
			if len(args) > 0 {
				notetakerID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Action Items"))
			fmt.Println()
			fmt.Printf("Notetaker: %s\n", notetakerID)
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Println("Action Items")
			fmt.Println()

			items := []struct {
				assignee string
				task     string
				due      string
			}{
				{"Sarah", "Complete API integration for user endpoints", "Today"},
				{"Mike", "Finish database optimization", "End of day"},
				{"John", "Review Sarah's authentication PR", "Tomorrow"},
				{"Sarah", "Write unit tests for auth module", "Tomorrow"},
				{"Mike", "Document database schema changes", "This week"},
			}

			for i, item := range items {
				fmt.Printf("  %s %s\n", cyan.Sprintf("%d.", i+1), boldWhite.Sprint(item.task))
				fmt.Printf("     Assignee: %s\n", item.assignee)
				fmt.Printf("     Due:      %s\n", item.due)
				fmt.Println()
			}

			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// printDemoNotetaker prints a notetaker summary.
func printDemoNotetaker(nt domain.Notetaker) {
	// State icon and color
	var stateIcon string
	var stateColor *color.Color

	switch nt.State {
	case domain.NotetakerStateComplete:
		stateIcon = "✓"
		stateColor = green
	case domain.NotetakerStateAttending:
		stateIcon = "●"
		stateColor = cyan
	case domain.NotetakerStateScheduled:
		stateIcon = "○"
		stateColor = yellow
	default:
		stateIcon = "?"
		stateColor = dim
	}

	title := nt.MeetingTitle
	if title == "" {
		title = "Untitled Meeting"
	}

	fmt.Printf("  %s %s\n", stateColor.Sprint(stateIcon), boldWhite.Sprint(title))
	fmt.Printf("    State: %s\n", stateColor.Sprint(string(nt.State)))
	fmt.Printf("    Link:  %s\n", dim.Sprint(nt.MeetingLink))

	if !nt.JoinTime.IsZero() {
		fmt.Printf("    Join:  %s\n", nt.JoinTime.Format("Jan 2, 2006 3:04 PM"))
	}

	_, _ = dim.Printf("    ID:    %s\n", nt.ID)
	fmt.Println()
}

// printDemoNotetakerFull prints full notetaker details.
func printDemoNotetakerFull(nt domain.Notetaker, media *domain.MediaData) {
	title := nt.MeetingTitle
	if title == "" {
		title = "Untitled Meeting"
	}

	// State icon and color
	var stateColor *color.Color
	switch nt.State {
	case domain.NotetakerStateComplete:
		stateColor = green
	case domain.NotetakerStateAttending:
		stateColor = cyan
	case domain.NotetakerStateScheduled:
		stateColor = yellow
	default:
		stateColor = dim
	}

	fmt.Println("────────────────────────────────────────────────────")
	_, _ = boldWhite.Printf("Meeting: %s\n", title)
	fmt.Printf("State:   %s\n", stateColor.Sprint(string(nt.State)))
	fmt.Printf("Link:    %s\n", nt.MeetingLink)
	fmt.Printf("ID:      %s\n", nt.ID)

	if !nt.JoinTime.IsZero() {
		fmt.Printf("Join:    %s\n", nt.JoinTime.Format("Jan 2, 2006 3:04 PM"))
	}

	fmt.Printf("Created: %s\n", nt.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", nt.UpdatedAt.Format(time.RFC3339))

	if media != nil {
		fmt.Println("\n📁 Media Files:")

		if media.Recording != nil {
			fmt.Printf("  Recording:  %s\n", media.Recording.ContentType)
			fmt.Printf("              Size: %s\n", formatDemoSize(media.Recording.Size))
			_, _ = dim.Printf("              URL: %s\n", media.Recording.URL)
		}

		if media.Transcript != nil {
			fmt.Printf("  Transcript: %s\n", media.Transcript.ContentType)
			fmt.Printf("              Size: %s\n", formatDemoSize(media.Transcript.Size))
			_, _ = dim.Printf("              URL: %s\n", media.Transcript.URL)
		}
	}

	fmt.Println("────────────────────────────────────────────────────")
	fmt.Println()
}

// formatDemoSize formats a file size in bytes to a human-readable string.
func formatDemoSize(bytes int64) string {
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
