package demo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

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
			fmt.Println(common.Dim.Sprint("⏰ Demo Mode - Scheduled Messages"))
			fmt.Println()
			fmt.Printf("Found %d scheduled messages:\n\n", len(scheduled))

			for _, s := range scheduled {
				sendTime := time.Unix(s.CloseTime, 0)
				fmt.Printf("  ⏰ %s\n", common.BoldWhite.Sprint(s.ScheduleID))
				fmt.Printf("     Status: %s\n", s.Status)
				fmt.Printf("     Sends at: %s\n", sendTime.Format("Jan 2, 2006 3:04 PM"))
				fmt.Println()
			}

			fmt.Println(common.Dim.Sprint("To manage scheduled messages: nylas auth login"))

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
			fmt.Println(common.Dim.Sprint("⏰ Demo Mode - Cancel Scheduled Message (Simulated)"))
			fmt.Println()
			_, _ = common.Green.Printf("✓ Scheduled message '%s' would be cancelled\n", scheduleID)
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To manage scheduled messages: nylas auth login"))

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
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Smart Compose"))
			fmt.Printf("Prompt: %s\n\n", common.BoldWhite.Sprint(prompt))
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println(suggestion.Suggestion)
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To use AI compose with your emails: nylas auth login"))

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
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Email Summary"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			_, _ = common.BoldWhite.Println("Summary:")
			fmt.Println("This email discusses the Q4 planning meeting action items.")
			fmt.Println("Key points:")
			fmt.Println("  • Review Q4 roadmap by Friday")
			fmt.Println("  • Submit budget proposals")
			fmt.Println("  • Schedule 1:1s with new team members")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To use AI features with your emails: nylas auth login"))

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
			fmt.Println(common.Dim.Sprint("🤖 Demo Mode - AI Extract Key Info"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 60))
			_, _ = common.BoldWhite.Println("Extracted Information:")
			fmt.Println("  Action Items: 3")
			fmt.Println("  Deadlines: Friday (Q4 roadmap review)")
			fmt.Println("  People Mentioned: Sarah Chen, team members")
			fmt.Println("  Sentiment: Professional, positive")
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To use AI features with your emails: nylas auth login"))

			return nil
		},
	}
}
