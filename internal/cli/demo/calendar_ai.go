package demo

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
)

func newDemoCalendarAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered calendar features",
		Long:  "Demo AI-powered calendar analysis and suggestions.",
	}

	cmd.AddCommand(newDemoAIAnalyzeCmd())
	cmd.AddCommand(newDemoAIConflictsCmd())
	cmd.AddCommand(newDemoAIRescheduleCmd())
	cmd.AddCommand(newDemoAIFocusTimeCmd())
	cmd.AddCommand(newDemoAIAdaptCmd())

	return cmd
}

func newDemoAIAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "Analyze calendar patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Calendar Analysis"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = common.BoldWhite.Println("Weekly Calendar Analysis")
			fmt.Println()

			fmt.Println("📊 Meeting Statistics:")
			fmt.Printf("  Total meetings:        %s\n", common.Cyan.Sprint("23"))
			fmt.Printf("  Total meeting hours:   %s\n", common.Cyan.Sprint("18.5 hours"))
			fmt.Printf("  Average meeting length: %s\n", common.Cyan.Sprint("48 minutes"))
			fmt.Printf("  Focus time available:  %s\n", common.Yellow.Sprint("12 hours"))
			fmt.Println()

			fmt.Println("💡 AI Suggestions:")
			fmt.Printf("  • %s Consider batching 1:1s on Tuesdays\n", common.Green.Sprint("●"))
			fmt.Printf("  • %s Move recurring standup 30min later for focus time\n", common.Green.Sprint("●"))
			fmt.Printf("  • %s 3 meetings could be consolidated\n", common.Yellow.Sprint("●"))

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoAIConflictsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "conflicts",
		Short: "Detect scheduling conflicts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Conflict Detection"))
			fmt.Println()

			now := time.Now()

			fmt.Println("⚠️  Conflicts Found:")
			fmt.Println()
			fmt.Printf("  %s %s\n", common.Red.Sprint("●"), common.BoldWhite.Sprint("Double-booked: Project Review + Client Call"))
			_, _ = common.Dim.Printf("    %s at 2:00 PM - 3:00 PM\n", now.AddDate(0, 0, 2).Format("Mon, Jan 2"))
			fmt.Printf("    Suggestion: Move Project Review to 3:30 PM\n")
			fmt.Println()

			fmt.Printf("  %s %s\n", common.Yellow.Sprint("●"), common.BoldWhite.Sprint("Back-to-back: 4 meetings without break"))
			_, _ = common.Dim.Printf("    %s from 9:00 AM - 1:00 PM\n", now.AddDate(0, 0, 3).Format("Mon, Jan 2"))
			fmt.Printf("    Suggestion: Add 15-min buffer between meetings\n")

			return nil
		},
	}
}

func newDemoAIRescheduleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reschedule [event-id]",
		Short: "Get AI rescheduling suggestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-123"
			if len(args) > 0 {
				eventID = args[0]
			}

			fmt.Println()
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Reschedule Suggestions"))
			fmt.Println()
			fmt.Printf("Event: %s\n", eventID)
			fmt.Println()

			now := time.Now()
			fmt.Println("📅 Suggested Alternative Times:")
			fmt.Printf("  %s %s at 10:00 AM %s\n", common.Green.Sprint("1."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"), common.Green.Sprint("(Recommended)"))
			fmt.Printf("     Reason: All attendees available, minimal disruption\n")
			fmt.Println()
			fmt.Printf("  %s %s at 3:00 PM\n", common.Cyan.Sprint("2."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"))
			fmt.Printf("     Reason: Good focus time afterwards\n")

			return nil
		},
	}
}

func newDemoAIFocusTimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focus-time",
		Short: "Find focus time blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Focus Time Finder"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = common.BoldWhite.Println("Available Focus Time Blocks This Week")
			fmt.Println()

			now := time.Now()
			blocks := []struct {
				day      time.Time
				time     string
				duration string
			}{
				{now.AddDate(0, 0, 1), "8:00 AM - 10:00 AM", "2 hours"},
				{now.AddDate(0, 0, 1), "2:00 PM - 4:00 PM", "2 hours"},
				{now.AddDate(0, 0, 2), "9:00 AM - 12:00 PM", "3 hours"},
				{now.AddDate(0, 0, 4), "1:00 PM - 5:00 PM", "4 hours"},
			}

			for _, b := range blocks {
				fmt.Printf("  %s %s  %s %s\n",
					common.Green.Sprint("●"),
					b.day.Format("Mon, Jan 2"),
					common.Cyan.Sprint(b.time),
					common.Dim.Sprintf("(%s)", b.duration))
			}

			fmt.Println()
			fmt.Println("💡 Tip: Block these times to protect deep work sessions")
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoAIAdaptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "adapt",
		Short: "Get adaptive scheduling suggestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Adaptive Scheduling"))
			fmt.Println()

			fmt.Println("📊 Learning from your patterns...")
			fmt.Println()
			fmt.Println("Observations:")
			fmt.Printf("  • You're most productive between 9-11 AM\n")
			fmt.Printf("  • You prefer shorter meetings (< 30 min)\n")
			fmt.Printf("  • Thursdays have the most focus time\n")
			fmt.Println()
			fmt.Println("Recommendations:")
			fmt.Printf("  %s Schedule important work for morning blocks\n", common.Green.Sprint("●"))
			fmt.Printf("  %s Set default meeting duration to 25 minutes\n", common.Green.Sprint("●"))
			fmt.Printf("  %s Protect Thursday mornings for deep work\n", common.Green.Sprint("●"))

			return nil
		},
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// printDemoEvent prints a single event.
