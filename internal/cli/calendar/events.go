package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"ev", "event"},
		Short:   "Manage calendar events",
		Long:    "List, view, create, and delete calendar events.",
	}

	cmd.AddCommand(newEventsListCmd())
	cmd.AddCommand(newEventsShowCmd())
	cmd.AddCommand(newEventsCreateCmd())
	cmd.AddCommand(newEventsUpdateCmd())
	cmd.AddCommand(newEventsDeleteCmd())
	cmd.AddCommand(newEventsRSVPCmd())

	return cmd
}

func newEventsListCmd() *cobra.Command {
	var (
		calendarID string
		limit      int
		days       int
		showAll    bool
	)

	cmd := &cobra.Command{
		Use:     "list [grant-id]",
		Aliases: []string{"ls"},
		Short:   "List calendar events",
		Long:    "List events from the specified calendar or primary calendar.",
		Args:    cobra.MaximumNArgs(1),
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

			// If no calendar specified, try to get the primary calendar
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary {
						calendarID = cal.ID
						break
					}
				}
				if calendarID == "" && len(calendars) > 0 {
					calendarID = calendars[0].ID
				}
				if calendarID == "" {
					return common.NewUserError(
						"no calendars found",
						"Connect a calendar account with: nylas auth login",
					)
				}
			}

			params := &domain.EventQueryParams{
				Limit:   limit,
				OrderBy: "start", // Sort by start time ascending
			}

			// Set time range if days specified
			if days > 0 {
				now := time.Now()
				params.Start = now.Unix()
				params.End = now.AddDate(0, 0, days).Unix()
			}

			if showAll {
				params.ShowCancelled = true
			}

			events, err := client.GetEvents(ctx, grantID, calendarID, params)
			if err != nil {
				return fmt.Errorf("failed to get events: %w", err)
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			dim := color.New(color.Faint)

			fmt.Printf("Found %d event(s):\n\n", len(events))

			for _, event := range events {
				// Title
				fmt.Printf("%s\n", cyan.Sprint(event.Title))

				// Time
				when := formatEventTime(event.When)
				fmt.Printf("  %s %s\n", dim.Sprint("When:"), when)

				// Location
				if event.Location != "" {
					fmt.Printf("  %s %s\n", dim.Sprint("Location:"), event.Location)
				}

				// Status
				statusColor := green
				if event.Status == "cancelled" {
					statusColor = color.New(color.FgRed)
				} else if event.Status == "tentative" {
					statusColor = yellow
				}
				if event.Status != "" {
					fmt.Printf("  %s %s\n", dim.Sprint("Status:"), statusColor.Sprint(event.Status))
				}

				// Participants count
				if len(event.Participants) > 0 {
					fmt.Printf("  %s %d participant(s)\n", dim.Sprint("Guests:"), len(event.Participants))
				}

				// ID
				fmt.Printf("  %s %s\n", dim.Sprint("ID:"), dim.Sprint(event.ID))
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of events to show")
	cmd.Flags().IntVarP(&days, "days", "d", 7, "Show events for the next N days (0 for no limit)")
	cmd.Flags().BoolVar(&showAll, "show-cancelled", false, "Include cancelled events")

	return cmd
}

func newEventsShowCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:     "show <event-id> [grant-id]",
		Aliases: []string{"read", "get"},
		Short:   "Show event details",
		Long:    "Display detailed information about a specific event.",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]

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

			// Get calendar ID if not specified
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary {
						calendarID = cal.ID
						break
					}
				}
				if calendarID == "" && len(calendars) > 0 {
					calendarID = calendars[0].ID
				}
			}

			event, err := client.GetEvent(ctx, grantID, calendarID, eventID)
			if err != nil {
				return fmt.Errorf("failed to get event: %w", err)
			}

			cyan := color.New(color.FgCyan, color.Bold)
			green := color.New(color.FgGreen)
			dim := color.New(color.Faint)

			// Title
			fmt.Printf("%s\n\n", cyan.Sprint(event.Title))

			// Time
			fmt.Printf("%s\n", green.Sprint("When"))
			fmt.Printf("  %s\n\n", formatEventTime(event.When))

			// Location
			if event.Location != "" {
				fmt.Printf("%s\n", green.Sprint("Location"))
				fmt.Printf("  %s\n\n", event.Location)
			}

			// Description
			if event.Description != "" {
				fmt.Printf("%s\n", green.Sprint("Description"))
				fmt.Printf("  %s\n\n", event.Description)
			}

			// Organizer
			if event.Organizer != nil {
				fmt.Printf("%s\n", green.Sprint("Organizer"))
				if event.Organizer.Name != "" {
					fmt.Printf("  %s <%s>\n\n", event.Organizer.Name, event.Organizer.Email)
				} else {
					fmt.Printf("  %s\n\n", event.Organizer.Email)
				}
			}

			// Participants
			if len(event.Participants) > 0 {
				fmt.Printf("%s\n", green.Sprint("Participants"))
				for _, p := range event.Participants {
					status := formatParticipantStatus(p.Status)
					if p.Name != "" {
						fmt.Printf("  %s <%s> %s\n", p.Name, p.Email, status)
					} else {
						fmt.Printf("  %s %s\n", p.Email, status)
					}
				}
				fmt.Println()
			}

			// Conferencing
			if event.Conferencing != nil && event.Conferencing.Details != nil {
				fmt.Printf("%s\n", green.Sprint("Video Conference"))
				if event.Conferencing.Provider != "" {
					fmt.Printf("  Provider: %s\n", event.Conferencing.Provider)
				}
				if event.Conferencing.Details.URL != "" {
					fmt.Printf("  URL: %s\n", event.Conferencing.Details.URL)
				}
				fmt.Println()
			}

			// Metadata
			fmt.Printf("%s\n", green.Sprint("Details"))
			fmt.Printf("  Status: %s\n", event.Status)
			fmt.Printf("  Busy: %v\n", event.Busy)
			if event.Visibility != "" {
				fmt.Printf("  Visibility: %s\n", event.Visibility)
			}
			fmt.Printf("  ID: %s\n", dim.Sprint(event.ID))
			fmt.Printf("  Calendar: %s\n", dim.Sprint(event.CalendarID))

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")

	return cmd
}

func newEventsCreateCmd() *cobra.Command {
	var (
		calendarID   string
		title        string
		description  string
		location     string
		startTime    string
		endTime      string
		allDay       bool
		participants []string
		busy         bool
	)

	cmd := &cobra.Command{
		Use:   "create [grant-id]",
		Short: "Create a new event",
		Long: `Create a new calendar event.

Examples:
  # Create a simple event
  nylas calendar events create --title "Meeting" --start "2024-01-15 14:00" --end "2024-01-15 15:00"

  # Create an all-day event
  nylas calendar events create --title "Vacation" --start "2024-01-15" --all-day

  # Create event with participants
  nylas calendar events create --title "Team Sync" --start "2024-01-15 10:00" --end "2024-01-15 11:00" \
    --participant "alice@example.com" --participant "bob@example.com"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return common.NewUserError(
					"title is required",
					"Use --title to specify event title",
				)
			}
			if startTime == "" {
				return common.NewUserError(
					"start time is required",
					"Use --start to specify start time (e.g., '2024-01-15 14:00' or '2024-01-15' for all-day)",
				)
			}

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

			// Get calendar ID if not specified
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary && !cal.ReadOnly {
						calendarID = cal.ID
						break
					}
				}
				// Fallback to any writable calendar
				if calendarID == "" {
					for _, cal := range calendars {
						if !cal.ReadOnly {
							calendarID = cal.ID
							break
						}
					}
				}
				if calendarID == "" {
					return common.NewUserError(
						"no writable calendar found",
						"Specify a calendar with --calendar",
					)
				}
			}

			// Parse times
			when, err := parseEventTime(startTime, endTime, allDay)
			if err != nil {
				return err
			}

			req := &domain.CreateEventRequest{
				Title:       title,
				Description: description,
				Location:    location,
				When:        *when,
				Busy:        busy,
			}

			// Add participants
			for _, email := range participants {
				req.Participants = append(req.Participants, domain.Participant{
					Email: email,
				})
			}

			spinner := common.NewSpinner("Creating event...")
			spinner.Start()

			event, err := client.CreateEvent(ctx, grantID, calendarID, req)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to create event: %w", err)
			}

			green := color.New(color.FgGreen)
			fmt.Printf("%s Event created successfully!\n\n", green.Sprint("✓"))
			fmt.Printf("Title: %s\n", event.Title)
			fmt.Printf("When: %s\n", formatEventTime(event.When))
			fmt.Printf("ID: %s\n", event.ID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().StringVarP(&title, "title", "t", "", "Event title (required)")
	cmd.Flags().StringVarP(&description, "description", "D", "", "Event description")
	cmd.Flags().StringVarP(&location, "location", "l", "", "Event location")
	cmd.Flags().StringVarP(&startTime, "start", "s", "", "Start time (e.g., '2024-01-15 14:00' or '2024-01-15')")
	cmd.Flags().StringVarP(&endTime, "end", "e", "", "End time (defaults to 1 hour after start)")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "Create an all-day event")
	cmd.Flags().StringArrayVarP(&participants, "participant", "p", nil, "Add participant email (can be used multiple times)")
	cmd.Flags().BoolVar(&busy, "busy", true, "Mark time as busy")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("start")

	return cmd
}

func newEventsDeleteCmd() *cobra.Command {
	var (
		calendarID string
		force      bool
	)

	cmd := &cobra.Command{
		Use:     "delete <event-id> [grant-id]",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete an event",
		Long:    "Delete a calendar event by its ID.",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]

			if !force {
				fmt.Printf("Are you sure you want to delete event %s? [y/N] ", eventID)
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

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

			// Get calendar ID if not specified
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary {
						calendarID = cal.ID
						break
					}
				}
				if calendarID == "" && len(calendars) > 0 {
					calendarID = calendars[0].ID
				}
			}

			spinner := common.NewSpinner("Deleting event...")
			spinner.Start()

			err = client.DeleteEvent(ctx, grantID, calendarID, eventID)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to delete event: %w", err)
			}

			green := color.New(color.FgGreen)
			fmt.Printf("%s Event deleted successfully.\n", green.Sprint("✓"))

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func newEventsUpdateCmd() *cobra.Command {
	var (
		calendarID   string
		title        string
		description  string
		location     string
		startTime    string
		endTime      string
		allDay       bool
		participants []string
		busy         bool
		visibility   string
	)

	cmd := &cobra.Command{
		Use:   "update <event-id> [grant-id]",
		Short: "Update an existing event",
		Long: `Update a calendar event.

Examples:
  # Update event title
  nylas calendar events update <event-id> --title "New Title"

  # Update event time
  nylas calendar events update <event-id> --start "2024-01-15 14:00" --end "2024-01-15 15:00"

  # Update location and description
  nylas calendar events update <event-id> --location "Conference Room A" --description "Weekly sync"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]

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

			// Get calendar ID if not specified
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary {
						calendarID = cal.ID
						break
					}
				}
				if calendarID == "" && len(calendars) > 0 {
					calendarID = calendars[0].ID
				}
			}

			req := &domain.UpdateEventRequest{}

			if cmd.Flags().Changed("title") {
				req.Title = &title
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}
			if cmd.Flags().Changed("location") {
				req.Location = &location
			}
			if cmd.Flags().Changed("busy") {
				req.Busy = &busy
			}
			if cmd.Flags().Changed("visibility") {
				req.Visibility = &visibility
			}

			// Handle time changes
			if cmd.Flags().Changed("start") {
				when, err := parseEventTime(startTime, endTime, allDay)
				if err != nil {
					return err
				}
				req.When = when
			}

			// Handle participants
			if len(participants) > 0 {
				for _, email := range participants {
					req.Participants = append(req.Participants, domain.Participant{
						Email: email,
					})
				}
			}

			spinner := common.NewSpinner("Updating event...")
			spinner.Start()

			event, err := client.UpdateEvent(ctx, grantID, calendarID, eventID, req)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to update event: %w", err)
			}

			green := color.New(color.FgGreen)
			fmt.Printf("%s Event updated successfully!\n\n", green.Sprint("✓"))
			fmt.Printf("Title: %s\n", event.Title)
			fmt.Printf("When: %s\n", formatEventTime(event.When))
			fmt.Printf("ID: %s\n", event.ID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().StringVarP(&title, "title", "t", "", "Event title")
	cmd.Flags().StringVarP(&description, "description", "D", "", "Event description")
	cmd.Flags().StringVarP(&location, "location", "l", "", "Event location")
	cmd.Flags().StringVarP(&startTime, "start", "s", "", "Start time (e.g., '2024-01-15 14:00')")
	cmd.Flags().StringVarP(&endTime, "end", "e", "", "End time")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "Set as all-day event")
	cmd.Flags().StringArrayVarP(&participants, "participant", "p", nil, "Set participant emails (replaces existing)")
	cmd.Flags().BoolVar(&busy, "busy", true, "Mark time as busy")
	cmd.Flags().StringVar(&visibility, "visibility", "", "Event visibility (public, private, default)")

	return cmd
}

func newEventsRSVPCmd() *cobra.Command {
	var (
		calendarID string
		comment    string
	)

	cmd := &cobra.Command{
		Use:   "rsvp <event-id> <status> [grant-id]",
		Short: "RSVP to an event invitation",
		Long: `Respond to an event invitation with your RSVP status.

Status options:
  - yes    Accept the invitation
  - no     Decline the invitation
  - maybe  Tentatively accept

Examples:
  # Accept an event invitation
  nylas calendar events rsvp <event-id> yes

  # Decline with a comment
  nylas calendar events rsvp <event-id> no --comment "I have a conflict"

  # Tentatively accept
  nylas calendar events rsvp <event-id> maybe`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]
			status := strings.ToLower(args[1])

			// Validate status
			if status != "yes" && status != "no" && status != "maybe" {
				return common.NewUserError(
					"invalid RSVP status",
					"Status must be 'yes', 'no', or 'maybe'",
				)
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			var grantID string
			if len(args) > 2 {
				grantID = args[2]
			} else {
				grantID, err = getGrantID(nil)
				if err != nil {
					return err
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			// Get calendar ID if not specified
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return fmt.Errorf("failed to get calendars: %w", err)
				}
				for _, cal := range calendars {
					if cal.IsPrimary {
						calendarID = cal.ID
						break
					}
				}
				if calendarID == "" && len(calendars) > 0 {
					calendarID = calendars[0].ID
				}
			}

			req := &domain.SendRSVPRequest{
				Status:  status,
				Comment: comment,
			}

			spinner := common.NewSpinner("Sending RSVP...")
			spinner.Start()

			err = client.SendRSVP(ctx, grantID, calendarID, eventID, req)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to send RSVP: %w", err)
			}

			green := color.New(color.FgGreen)
			statusText := map[string]string{
				"yes":   "accepted",
				"no":    "declined",
				"maybe": "tentatively accepted",
			}
			fmt.Printf("%s RSVP sent! You have %s the invitation.\n", green.Sprint("✓"), statusText[status])

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional comment with your RSVP")

	return cmd
}

// Helper functions

func formatEventTime(when domain.EventWhen) string {
	if when.IsAllDay() {
		start := when.StartDateTime()
		end := when.EndDateTime()
		if start.Equal(end) || end.IsZero() {
			return start.Format("Mon, Jan 2, 2006") + " (all day)"
		}
		return fmt.Sprintf("%s - %s (all day)",
			start.Format("Mon, Jan 2, 2006"),
			end.Format("Mon, Jan 2, 2006"))
	}

	start := when.StartDateTime()
	end := when.EndDateTime()

	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		// Same day
		return fmt.Sprintf("%s, %s - %s",
			start.Format("Mon, Jan 2, 2006"),
			start.Format("3:04 PM"),
			end.Format("3:04 PM"))
	}

	return fmt.Sprintf("%s - %s",
		start.Format("Mon, Jan 2, 2006 3:04 PM"),
		end.Format("Mon, Jan 2, 2006 3:04 PM"))
}

func formatParticipantStatus(status string) string {
	switch status {
	case "yes":
		return color.GreenString("✓ accepted")
	case "no":
		return color.RedString("✗ declined")
	case "maybe":
		return color.YellowString("? tentative")
	case "noreply":
		return color.New(color.Faint).Sprint("pending")
	default:
		return ""
	}
}

func parseEventTime(startStr, endStr string, allDay bool) (*domain.EventWhen, error) {
	when := &domain.EventWhen{}

	// Try parsing as date first (YYYY-MM-DD)
	if allDay || len(startStr) <= 10 {
		startDate, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			when.Object = "date"
			when.Date = startDate.Format("2006-01-02")
			if endStr != "" {
				endDate, err := time.Parse("2006-01-02", endStr)
				if err != nil {
					return nil, fmt.Errorf("invalid end date format: %s (use YYYY-MM-DD)", endStr)
				}
				if !endDate.Equal(startDate) {
					when.Object = "datespan"
					when.StartDate = when.Date
					when.Date = ""
					when.EndDate = endDate.Format("2006-01-02")
				}
			}
			return when, nil
		}
	}

	// Try parsing as datetime
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	var startTime time.Time
	var parsed bool
	for _, format := range formats {
		t, err := time.ParseInLocation(format, startStr, time.Local)
		if err == nil {
			startTime = t
			parsed = true
			break
		}
	}
	if !parsed {
		return nil, fmt.Errorf("invalid start time format: %s (use 'YYYY-MM-DD HH:MM' or 'YYYY-MM-DD')", startStr)
	}

	when.Object = "timespan"
	when.StartTime = startTime.Unix()

	if endStr != "" {
		var endTime time.Time
		for _, format := range formats {
			t, err := time.ParseInLocation(format, endStr, time.Local)
			if err == nil {
				endTime = t
				break
			}
		}
		if endTime.IsZero() {
			return nil, fmt.Errorf("invalid end time format: %s", endStr)
		}
		when.EndTime = endTime.Unix()
	} else {
		// Default to 1 hour duration
		when.EndTime = startTime.Add(time.Hour).Unix()
	}

	return when, nil
}
