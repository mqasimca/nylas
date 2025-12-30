package demo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
)

func newDemoEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Manage calendar events",
		Long:  "Demo commands for managing calendar events.",
	}

	cmd.AddCommand(newDemoEventsListCmd())
	cmd.AddCommand(newDemoEventsShowCmd())
	cmd.AddCommand(newDemoEventsCreateCmd())
	cmd.AddCommand(newDemoEventsUpdateCmd())
	cmd.AddCommand(newDemoEventsDeleteCmd())

	return cmd
}

func newDemoEventsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List events",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			events, err := client.GetEvents(ctx, "demo-grant", "primary", nil)
			if err != nil {
				return fmt.Errorf("failed to get demo events: %w", err)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Events List"))
			fmt.Println()
			fmt.Printf("Found %d events:\n\n", len(events))

			for _, event := range events {
				printDemoEvent(event, false)
			}

			return nil
		},
	}
}

func newDemoEventsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [event-id]",
		Short: "Show event details",
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-001"
			if len(args) > 0 {
				eventID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Event Details"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = boldWhite.Println("Team Standup Meeting")
			fmt.Printf("  ID:          %s\n", eventID)
			fmt.Printf("  Calendar:    Work Calendar\n")
			fmt.Printf("  Start:       Tomorrow at 9:00 AM\n")
			fmt.Printf("  End:         Tomorrow at 9:30 AM\n")
			fmt.Printf("  Status:      %s\n", green.Sprint("confirmed"))
			fmt.Printf("  Location:    Zoom Meeting\n")
			fmt.Printf("  Organizer:   demo@example.com\n")
			fmt.Printf("  Attendees:   3 participants\n")
			fmt.Printf("  Recurring:   Weekly on weekdays\n")
			fmt.Println(strings.Repeat("─", 50))

			return nil
		},
	}
}

func newDemoEventsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create an event",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			_, _ = green.Println("✓ Event would be created (demo mode)")
			_, _ = dim.Printf("  Event ID: evt-demo-%d\n", time.Now().Unix())
			return nil
		},
	}
}

func newDemoEventsUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [event-id]",
		Short: "Update an event",
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-123"
			if len(args) > 0 {
				eventID = args[0]
			}
			fmt.Println()
			_, _ = green.Printf("✓ Event %s would be updated (demo mode)\n", eventID)
			return nil
		},
	}
}

func newDemoEventsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [event-id]",
		Short: "Delete an event",
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-123"
			if len(args) > 0 {
				eventID = args[0]
			}
			fmt.Println()
			_, _ = green.Printf("✓ Event %s would be deleted (demo mode)\n", eventID)
			return nil
		},
	}
}

// ============================================================================
// AVAILABILITY & SCHEDULING COMMANDS
// ============================================================================

// newDemoAvailabilityCmd shows sample availability.
