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

// newDemoCalendarCmd creates the demo calendar command with subcommands.
func newDemoCalendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Explore calendar features with sample data",
		Long:  "Demo calendar commands showing sample events and simulated operations.",
	}

	// Calendar management
	cmd.AddCommand(newDemoCalendarsListCmd())
	cmd.AddCommand(newDemoCalendarShowCmd())

	// Event management
	cmd.AddCommand(newDemoCalendarListCmd())
	cmd.AddCommand(newDemoCalendarCreateCmd())
	cmd.AddCommand(newDemoCalendarUpdateCmd())
	cmd.AddCommand(newDemoCalendarDeleteCmd())

	// Events subcommand group
	cmd.AddCommand(newDemoEventsCmd())

	// Availability & Scheduling
	cmd.AddCommand(newDemoAvailabilityCmd())
	cmd.AddCommand(newDemoFindTimeCmd())
	cmd.AddCommand(newDemoScheduleCmd())

	// Advanced features
	cmd.AddCommand(newDemoRecurringCmd())
	cmd.AddCommand(newDemoVirtualCmd())

	// AI features
	cmd.AddCommand(newDemoCalendarAICmd())

	return cmd
}

// ============================================================================
// CALENDAR MANAGEMENT COMMANDS
// ============================================================================

// newDemoCalendarsListCmd lists sample calendars.
func newDemoCalendarsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendars",
		Short: "List sample calendars",
		Long:  "Display a list of sample calendars.",
		Example: `  # List sample calendars
  nylas demo calendar calendars`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			calendars, err := client.GetCalendars(ctx, "demo-grant")
			if err != nil {
				return fmt.Errorf("failed to get demo calendars: %w", err)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Sample Calendars"))
			fmt.Println()

			for _, cal := range calendars {
				primary := ""
				if cal.IsPrimary {
					primary = green.Sprint(" (primary)")
				}
				fmt.Printf("  %s %s%s\n", cal.HexColor, boldWhite.Sprint(cal.Name), primary)
				// #nosec G104 -- color output errors are non-critical, best-effort display
				dim.Printf("    ID: %s\n", cal.ID)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To connect your real calendar: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// newDemoCalendarShowCmd shows a sample calendar.
func newDemoCalendarShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [calendar-id]",
		Short: "Show sample calendar details",
		Long:  "Display details of a sample calendar.",
		Example: `  # Show primary calendar
  nylas demo calendar show primary

  # Show specific calendar
  nylas demo calendar show cal-work-123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			calID := "primary"
			if len(args) > 0 {
				calID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Calendar Details"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			// Show sample calendar details
			// #nosec G104 -- color output errors are non-critical, best-effort display
			boldWhite.Println("Work Calendar")
			fmt.Printf("  ID:          %s\n", calID)
			fmt.Printf("  Owner:       demo@example.com\n")
			fmt.Printf("  Timezone:    America/New_York\n")
			fmt.Printf("  Color:       %s\n", cyan.Sprint("●"))
			fmt.Printf("  Primary:     %s\n", green.Sprint("Yes"))
			fmt.Printf("  Read-only:   No\n")
			fmt.Printf("  Description: Work meetings and appointments\n")

			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			fmt.Println(dim.Sprint("To view your real calendars: nylas auth login"))

			return nil
		},
	}

	return cmd
}

// ============================================================================
// EVENT MANAGEMENT COMMANDS
// ============================================================================

// newDemoCalendarListCmd lists sample calendar events.
func newDemoCalendarListCmd() *cobra.Command {
	var limit int
	var showID bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sample calendar events",
		Long:  "Display a list of realistic sample calendar events.",
		Example: `  # List sample events
  nylas demo calendar list

  # List with IDs shown
  nylas demo calendar list --id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := nylas.NewDemoClient()
			ctx := context.Background()

			events, err := client.GetEvents(ctx, "demo-grant", "primary", nil)
			if err != nil {
				return fmt.Errorf("failed to get demo events: %w", err)
			}

			if limit > 0 && limit < len(events) {
				events = events[:limit]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Sample Events"))
			fmt.Println(dim.Sprint("These are sample events for demonstration purposes."))
			fmt.Println()
			fmt.Printf("Found %d events:\n\n", len(events))

			for _, event := range events {
				printDemoEvent(event, showID)
			}

			fmt.Println()
			fmt.Println(dim.Sprint("To connect your real calendar: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of events to show")
	cmd.Flags().BoolVar(&showID, "id", false, "Show event IDs")

	return cmd
}

// newDemoCalendarCreateCmd simulates creating a calendar event.
func newDemoCalendarCreateCmd() *cobra.Command {
	var title string
	var startTime string
	var duration int
	var location string
	var attendees []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Simulate creating a calendar event",
		Long: `Simulate creating a calendar event to see how the create command works.

No actual event is created - this is just a demonstration of the command flow.`,
		Example: `  # Simulate creating an event
  nylas demo calendar create --title "Team Meeting" --start "2024-01-15 10:00" --duration 60

  # With location and attendees
  nylas demo calendar create --title "Lunch" --start "tomorrow 12:00" --location "Downtown Cafe" --attendee "john@example.com"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				title = "Demo Meeting"
			}
			if startTime == "" {
				startTime = time.Now().Add(1 * time.Hour).Format("Jan 2, 2006 3:04 PM")
			}
			if duration == 0 {
				duration = 30
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Simulated Event Creation"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			// #nosec G104 -- color output errors are non-critical, best-effort display
			boldWhite.Printf("Title:    %s\n", title)
			fmt.Printf("Start:    %s\n", startTime)
			fmt.Printf("Duration: %d minutes\n", duration)
			if location != "" {
				fmt.Printf("Location: %s\n", location)
			}
			if len(attendees) > 0 {
				fmt.Printf("Attendees: %s\n", strings.Join(attendees, ", "))
			}
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			green.Println("✓ Event would be created (demo mode - no actual event created)")
			dim.Printf("  Event ID: evt-demo-%d\n", time.Now().Unix())
			fmt.Println()
			fmt.Println(dim.Sprint("To create real events, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Event title")
	cmd.Flags().StringVar(&startTime, "start", "", "Start time")
	cmd.Flags().IntVar(&duration, "duration", 30, "Duration in minutes")
	cmd.Flags().StringVar(&location, "location", "", "Event location")
	cmd.Flags().StringArrayVar(&attendees, "attendee", nil, "Attendee email (can be repeated)")

	return cmd
}

// newDemoCalendarUpdateCmd simulates updating a calendar event.
func newDemoCalendarUpdateCmd() *cobra.Command {
	var title string
	var startTime string
	var location string

	cmd := &cobra.Command{
		Use:   "update [event-id]",
		Short: "Simulate updating a calendar event",
		Long:  "Simulate updating a calendar event to see how the update command works.",
		Example: `  # Update event title
  nylas demo calendar update evt-123 --title "Updated Meeting"

  # Update location
  nylas demo calendar update evt-123 --location "Conference Room B"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-123"
			if len(args) > 0 {
				eventID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Simulated Event Update"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			dim.Printf("Event ID: %s\n", eventID)
			fmt.Println()
			boldWhite.Println("Changes:")
			if title != "" {
				fmt.Printf("  Title:    %s\n", title)
			}
			if startTime != "" {
				fmt.Printf("  Start:    %s\n", startTime)
			}
			if location != "" {
				fmt.Printf("  Location: %s\n", location)
			}
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			green.Println("✓ Event would be updated (demo mode - no actual changes made)")
			fmt.Println()
			fmt.Println(dim.Sprint("To update real events, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New event title")
	cmd.Flags().StringVar(&startTime, "start", "", "New start time")
	cmd.Flags().StringVar(&location, "location", "", "New event location")

	return cmd
}

// newDemoCalendarDeleteCmd simulates deleting a calendar event.
func newDemoCalendarDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [event-id]",
		Short: "Simulate deleting a calendar event",
		Long:  "Simulate deleting a calendar event to see how the delete command works.",
		Example: `  # Delete an event
  nylas demo calendar delete evt-123

  # Force delete without confirmation
  nylas demo calendar delete evt-123 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := "evt-demo-123"
			if len(args) > 0 {
				eventID = args[0]
			}

			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Simulated Event Deletion"))
			fmt.Println()

			if !force {
				yellow.Println("⚠ Would prompt for confirmation in real mode")
			}

			fmt.Printf("Event ID: %s\n", eventID)
			fmt.Println()
			green.Println("✓ Event would be deleted (demo mode - no actual deletion)")
			fmt.Println()
			fmt.Println(dim.Sprint("To delete real events, connect your account: nylas auth login"))

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

// ============================================================================
// EVENTS SUBCOMMAND GROUP
// ============================================================================

// newDemoEventsCmd creates the events subcommand group.
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
			boldWhite.Println("Team Standup Meeting")
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
			green.Println("✓ Event would be created (demo mode)")
			dim.Printf("  Event ID: evt-demo-%d\n", time.Now().Unix())
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
			green.Printf("✓ Event %s would be updated (demo mode)\n", eventID)
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
			green.Printf("✓ Event %s would be deleted (demo mode)\n", eventID)
			return nil
		},
	}
}

// ============================================================================
// AVAILABILITY & SCHEDULING COMMANDS
// ============================================================================

// newDemoAvailabilityCmd shows sample availability.
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
			fmt.Println(dim.Sprint("📅 Demo Mode - Sample Availability"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Println("Available Time Slots (Next 7 Days)")
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
				fmt.Printf("  %s %s  %s\n", green.Sprint("●"), slot.day, cyan.Sprint(slot.time))
			}

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			fmt.Println(dim.Sprint("To check your real availability: nylas auth login"))

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
			fmt.Println(dim.Sprint("📅 Demo Mode - Find Meeting Time"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Println("Best Available Times for All Participants")
			fmt.Println()

			now := time.Now()
			fmt.Printf("  %s %s at 10:00 AM %s\n", green.Sprint("1."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"), green.Sprint("(Best match)"))
			fmt.Printf("  %s %s at 2:00 PM\n", cyan.Sprint("2."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"))
			fmt.Printf("  %s %s at 11:00 AM\n", cyan.Sprint("3."), now.AddDate(0, 0, 2).Format("Mon, Jan 2"))

			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			fmt.Println(dim.Sprint("To find real meeting times: nylas auth login"))

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
			fmt.Println(dim.Sprint("📅 Demo Mode - Smart Scheduling"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))

			boldWhite.Println("Scheduling: Project Review")
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
			green.Println("✓ Meeting would be scheduled (demo mode)")
			fmt.Println()
			fmt.Println(dim.Sprint("To schedule real meetings: nylas auth login"))

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
			fmt.Println(dim.Sprint("📅 Demo Mode - Create Recurring Event"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Println("Weekly Team Standup")
			fmt.Println()
			fmt.Printf("  Pattern:     Every Monday, Wednesday, Friday\n")
			fmt.Printf("  Time:        9:00 AM - 9:30 AM\n")
			fmt.Printf("  Ends:        After 52 occurrences\n")
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			green.Println("✓ Recurring event would be created (demo mode)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List recurring events",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(dim.Sprint("📅 Demo Mode - Recurring Events"))
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
				fmt.Printf("  %s %s\n", green.Sprint("●"), boldWhite.Sprint(e.title))
				dim.Printf("    %s\n", e.pattern)
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
			fmt.Println(dim.Sprint("📅 Demo Mode - Virtual Calendars"))
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
				fmt.Printf("  %s %s\n", cyan.Sprint("●"), boldWhite.Sprint(cal.name))
				dim.Printf("    Source: %s\n", cal.source)
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
			green.Println("✓ Virtual calendar would be created (demo mode)")
			return nil
		},
	})

	return cmd
}

// ============================================================================
// AI FEATURES
// ============================================================================

// newDemoCalendarAICmd creates the AI subcommand group for calendar.
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
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Calendar Analysis"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Println("Weekly Calendar Analysis")
			fmt.Println()

			fmt.Println("📊 Meeting Statistics:")
			fmt.Printf("  Total meetings:        %s\n", cyan.Sprint("23"))
			fmt.Printf("  Total meeting hours:   %s\n", cyan.Sprint("18.5 hours"))
			fmt.Printf("  Average meeting length: %s\n", cyan.Sprint("48 minutes"))
			fmt.Printf("  Focus time available:  %s\n", yellow.Sprint("12 hours"))
			fmt.Println()

			fmt.Println("💡 AI Suggestions:")
			fmt.Printf("  • %s Consider batching 1:1s on Tuesdays\n", green.Sprint("●"))
			fmt.Printf("  • %s Move recurring standup 30min later for focus time\n", green.Sprint("●"))
			fmt.Printf("  • %s 3 meetings could be consolidated\n", yellow.Sprint("●"))

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
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Conflict Detection"))
			fmt.Println()

			now := time.Now()

			fmt.Println("⚠️  Conflicts Found:")
			fmt.Println()
			fmt.Printf("  %s %s\n", color.New(color.FgRed).Sprint("●"), boldWhite.Sprint("Double-booked: Project Review + Client Call"))
			dim.Printf("    %s at 2:00 PM - 3:00 PM\n", now.AddDate(0, 0, 2).Format("Mon, Jan 2"))
			fmt.Printf("    Suggestion: Move Project Review to 3:30 PM\n")
			fmt.Println()

			fmt.Printf("  %s %s\n", yellow.Sprint("●"), boldWhite.Sprint("Back-to-back: 4 meetings without break"))
			dim.Printf("    %s from 9:00 AM - 1:00 PM\n", now.AddDate(0, 0, 3).Format("Mon, Jan 2"))
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
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Reschedule Suggestions"))
			fmt.Println()
			fmt.Printf("Event: %s\n", eventID)
			fmt.Println()

			now := time.Now()
			fmt.Println("📅 Suggested Alternative Times:")
			fmt.Printf("  %s %s at 10:00 AM %s\n", green.Sprint("1."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"), green.Sprint("(Recommended)"))
			fmt.Printf("     Reason: All attendees available, minimal disruption\n")
			fmt.Println()
			fmt.Printf("  %s %s at 3:00 PM\n", cyan.Sprint("2."), now.AddDate(0, 0, 1).Format("Mon, Jan 2"))
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
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Focus Time Finder"))
			fmt.Println()
			fmt.Println(strings.Repeat("─", 50))
			boldWhite.Println("Available Focus Time Blocks This Week")
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
					green.Sprint("●"),
					b.day.Format("Mon, Jan 2"),
					cyan.Sprint(b.time),
					dim.Sprintf("(%s)", b.duration))
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
			fmt.Println(dim.Sprint("🤖 Demo Mode - AI Adaptive Scheduling"))
			fmt.Println()

			fmt.Println("📊 Learning from your patterns...")
			fmt.Println()
			fmt.Println("Observations:")
			fmt.Printf("  • You're most productive between 9-11 AM\n")
			fmt.Printf("  • You prefer shorter meetings (< 30 min)\n")
			fmt.Printf("  • Thursdays have the most focus time\n")
			fmt.Println()
			fmt.Println("Recommendations:")
			fmt.Printf("  %s Schedule important work for morning blocks\n", green.Sprint("●"))
			fmt.Printf("  %s Set default meeting duration to 25 minutes\n", green.Sprint("●"))
			fmt.Printf("  %s Protect Thursday mornings for deep work\n", green.Sprint("●"))

			return nil
		},
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// printDemoEvent prints a single event.
func printDemoEvent(event domain.Event, showID bool) {
	startTime := time.Unix(event.When.StartTime, 0)
	endTime := time.Unix(event.When.EndTime, 0)

	// Format time range
	var timeStr string
	if startTime.Day() == endTime.Day() {
		timeStr = fmt.Sprintf("%s - %s",
			startTime.Format("Jan 2, 3:04 PM"),
			endTime.Format("3:04 PM"))
	} else {
		timeStr = fmt.Sprintf("%s - %s",
			startTime.Format("Jan 2, 3:04 PM"),
			endTime.Format("Jan 2, 3:04 PM"))
	}

	// Status indicator
	statusColor := green
	if event.Status == "cancelled" {
		statusColor = color.New(color.FgRed)
	}

	fmt.Printf("  %s %s\n", statusColor.Sprint("●"), boldWhite.Sprint(event.Title))
	fmt.Printf("    %s\n", dim.Sprint(timeStr))

	if event.Location != "" {
		fmt.Printf("    📍 %s\n", event.Location)
	}

	if event.Conferencing != nil && event.Conferencing.Details != nil && event.Conferencing.Details.URL != "" {
		fmt.Printf("    🔗 %s\n", dim.Sprint(event.Conferencing.Details.URL))
	}

	if len(event.Participants) > 0 {
		names := make([]string, 0, len(event.Participants))
		for _, p := range event.Participants {
			if p.Name != "" {
				names = append(names, p.Name)
			} else if p.Email != "" {
				names = append(names, p.Email)
			}
		}
		if len(names) > 0 {
			fmt.Printf("    👥 %s\n", strings.Join(names, ", "))
		}
	}

	if showID {
		dim.Printf("    ID: %s\n", event.ID)
	}

	fmt.Println()
}
