package demo

import (
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newDemoAvailabilityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "availability",
		Short: "Check sample availability",
		Long:  "Display sample availability slots for scheduling.",
		Example: `  # Check availability for next week
  nylas demo calendar availability

  # Check availability for specific email
  nylas demo calendar availability --email john@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Sample Availability"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = common.BoldWhite.Println("Available Time Slots (Next 7 Days)")
			fmt.Println()

			// Sample availability slots
			now := time.Now()
			slots := []struct {
				day  string
				time string
			}{
				{now.AddDate(0, 0, 1).Format("Mon, Jan 2"), "9:00 AM - 10:00 AM"},
				{now.AddDate(0, 0, 1).Format("Mon, Jan 2"), "2:00 PM - 4:00 PM"},
				{now.AddDate(0, 0, 2).Format("Mon, Jan 2"), "10:00 AM - 12:00 PM"},
				{now.AddDate(0, 0, 2).Format("Mon, Jan 2"), "3:00 PM - 5:00 PM"},
				{now.AddDate(0, 0, 3).Format("Mon, Jan 2"), "9:00 AM - 11:00 AM"},
			}

			for _, slot := range slots {
				fmt.Printf("  %s %s  %s\n", common.Green.Sprint("●"), slot.day, common.Cyan.Sprint(slot.time))
			}

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To check your real availability: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().String("email", "", "Email to check availability for")
	cmd.Flags().String("duration", "30m", "Meeting duration")

	return cmd
}

// newDemoFindTimeCmd simulates finding a meeting time.
func newDemoFindTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-time",
		Short: "Find available meeting times",
		Long:  "Find times when all participants are available.",
		Example: `  # Find time for multiple participants
  nylas demo calendar find-time --attendee alice@example.com --attendee bob@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Find Meeting Time"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = common.BoldWhite.Println("Best Available Times for All Participants")
			fmt.Println()

			now := time.Now()
			fmt.Printf("  %s %s at 10:00 AM %s\n", common.Green.Sprint("1."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"), common.Green.Sprint("(Best match)"))
			fmt.Printf("  %s %s at 2:00 PM\n", common.Cyan.Sprint("2."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"))
			fmt.Printf("  %s %s at 11:00 AM\n", common.Cyan.Sprint("3."), now.AddDate(0, 0, 2).Format("Mon, Jan 2"))

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To find real meeting times: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringArray("attendee", nil, "Attendee email")
	cmd.Flags().String("duration", "30m", "Meeting duration")

	return cmd
}

// newDemoScheduleCmd simulates scheduling a meeting.
func newDemoScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Schedule a meeting",
		Long:  "Smart meeting scheduling with availability checking.",
		Example: `  # Schedule with participants
  nylas demo calendar schedule --title "Project Review" --attendee team@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Smart Scheduling"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			_, _ = common.BoldWhite.Println("Scheduling: Project Review")
			fmt.Println()
			fmt.Println("Checking availability for all participants...")
			fmt.Println()

			now := time.Now().Add(24 * time.Hour)
			fmt.Printf("  Best time found: %s at 10:00 AM\n", now.Format("Mon, Jan 2"))
			fmt.Printf("  Duration: 30 minutes\n")
			fmt.Printf("  All 3 participants available\n")

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			_, _ = common.Green.Println("✓ Meeting would be scheduled (demo mode)")
			fmt.Println()
			fmt.Println(common.Dim.Sprint("To schedule real meetings: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().String("title", "", "Meeting title")
	cmd.Flags().StringArray("attendee", nil, "Attendee email")
	cmd.Flags().String("duration", "30m", "Meeting duration")

	return cmd
}

// ============================================================================
// ADVANCED FEATURES
// ============================================================================

// newDemoRecurringCmd shows recurring event features.
func newDemoRecurringCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recurring",
		Short: "Manage recurring events",
		Long:  "Demo recurring event creation and management.",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a recurring event",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Create Recurring Event"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			_, _ = common.BoldWhite.Println("Weekly Team Standup")
			fmt.Println()
			fmt.Printf("  Pattern:     Every Monday, Wednesday, Friday\n")
			fmt.Printf("  Time:        9:00 AM - 9:30 AM\n")
			fmt.Printf("  Ends:        After 52 occurrences\n")
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			_, _ = common.Green.Println("✓ Recurring event would be created (demo mode)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List recurring events",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Recurring Events"))
			fmt.Println()

			events := []struct {
				title   string
				pattern string
			}{
				{"Team Standup", "Weekly on Mon, Wed, Fri"},
				{"1:1 with Manager", "Bi-weekly on Tuesdays"},
				{"Sprint Review", "Every 2 weeks on Fridays"},
				{"Monthly All-Hands", "Monthly on 1st Monday"},
			}

			for _, e := range events {
				fmt.Printf("  %s %s\n", common.Green.Sprint("●"), common.BoldWhite.Sprint(e.title))
				_, _ = common.Dim.Printf("    %s\n", e.pattern)
				fmt.Println()
			}

			return nil
		},
	})

	return cmd
}

// newDemoVirtualCmd shows virtual calendar features.
func newDemoVirtualCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "virtual",
		Short: "Manage virtual calendars",
		Long:  "Demo virtual calendar features for external calendar integration.",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List virtual calendars",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(common.Dim.Sprint("📅 Demo Mode - Virtual Calendars"))
			fmt.Println()

			calendars := []struct {
				name   string
				source string
			}{
				{"Team Shared Calendar", "Google Workspace"},
				{"Project Deadlines", "External ICS"},
				{"Holiday Calendar", "Public Calendar"},
			}

			for _, cal := range calendars {
				fmt.Printf("  %s %s\n", common.Cyan.Sprint("●"), common.BoldWhite.Sprint(cal.name))
				_, _ = common.Dim.Printf("    Source: %s\n", cal.source)
				fmt.Println()
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create virtual calendar",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			_, _ = common.Green.Println("✓ Virtual calendar would be created (demo mode)")
			return nil
		},
	})

	return cmd
}

// ============================================================================
// AI FEATURES
// ============================================================================

// newDemoCalendarAICmd creates the AI subcommand group for calendar.
