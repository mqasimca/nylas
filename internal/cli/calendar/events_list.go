package calendar

import (
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newEventsListCmd() *cobra.Command {
	var (
		calendarID string
		limit      int
		days       int
		showAll    bool
		targetTZ   string
		showTZ     bool
	)

	cmd := &cobra.Command{
		Use:     "list [grant-id]",
		Aliases: []string{"ls"},
		Short:   "List calendar events",
		Long: `List events from the specified calendar or primary calendar.

Examples:
  # List events in your local timezone
  nylas calendar events list

  # List events converted to a specific timezone
  nylas calendar events list --timezone America/Los_Angeles

  # List events with timezone abbreviations shown
  nylas calendar events list --show-tz`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Auto-detect timezone if not specified
			if targetTZ == "" && cmd.Flags().Changed("timezone") {
				// User explicitly set --timezone="" to clear
				targetTZ = ""
			} else if targetTZ == "" {
				// Default to local timezone for conversion display
				targetTZ = getLocalTimeZone()
			}

			// Validate timezone if specified
			if targetTZ != "" {
				if err := validateTimeZone(targetTZ); err != nil {
					return err
				}
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			// If no calendar specified, try to get the primary calendar
			if calendarID == "" {
				calendars, err := client.GetCalendars(ctx, grantID)
				if err != nil {
					return common.WrapListError("calendars", err)
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
				return common.WrapListError("events", err)
			}

			if len(events) == 0 {
				fmt.Println("No events found.")
				return nil
			}

			fmt.Printf("Found %d event(s):\n\n", len(events))

			for _, event := range events {
				// Title with timezone badge (if showing timezone info)
				fmt.Printf("%s", common.Cyan.Sprint(event.Title))
				if showTZ && !event.When.IsAllDay() {
					// Get event's original timezone
					start := event.When.StartDateTime()
					originalTZ := start.Location().String()
					if originalTZ == "Local" {
						originalTZ = getLocalTimeZone()
					}

					// Add colored timezone badge
					badge := formatTimezoneBadge(originalTZ, true) // Use abbreviation
					fmt.Printf(" %s", common.Blue.Sprint(badge))
				}
				fmt.Println()

				// Time (with timezone conversion if requested)
				timeDisplay, err := formatEventTimeWithTZ(&event, targetTZ)
				if err != nil {
					fmt.Printf("  %s %s (timezone conversion error: %v)\n",
						common.Dim.Sprint("When:"),
						formatEventTime(event.When),
						err)
				} else {
					if timeDisplay.ShowConversion {
						// Show converted time prominently
						fmt.Printf("  %s %s", common.Dim.Sprint("When:"), timeDisplay.ConvertedTime)
						if showTZ {
							fmt.Printf(" %s", common.BoldBlue.Sprint(timeDisplay.ConvertedTimezone))
						}
						fmt.Println()
						// Show original time as reference
						fmt.Printf("       %s %s",
							common.Dim.Sprint("(Original:"),
							common.Dim.Sprint(timeDisplay.OriginalTime))
						if showTZ {
							fmt.Printf(" %s", common.Dim.Sprint(timeDisplay.OriginalTimezone))
						}
						fmt.Printf("%s\n", common.Dim.Sprint(")"))
					} else {
						// No conversion - show original time
						fmt.Printf("  %s %s", common.Dim.Sprint("When:"), timeDisplay.OriginalTime)
						if showTZ && timeDisplay.OriginalTimezone != "" {
							fmt.Printf(" %s", common.BoldBlue.Sprint(timeDisplay.OriginalTimezone))
						}
						fmt.Println()
					}
				}

				// Location
				if event.Location != "" {
					fmt.Printf("  %s %s\n", common.Dim.Sprint("Location:"), event.Location)
				}

				// Status
				statusColor := common.Green
				switch event.Status {
				case "cancelled":
					statusColor = common.Red
				case "tentative":
					statusColor = common.Yellow
				}
				if event.Status != "" {
					fmt.Printf("  %s %s\n", common.Dim.Sprint("Status:"), statusColor.Sprint(event.Status))
				}

				// Participants count
				if len(event.Participants) > 0 {
					fmt.Printf("  %s %d participant(s)\n", common.Dim.Sprint("Guests:"), len(event.Participants))
				}

				// ID
				fmt.Printf("  %s %s\n", common.Dim.Sprint("ID:"), common.Dim.Sprint(event.ID))
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&calendarID, "calendar", "c", "", "Calendar ID (defaults to primary)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of events to show")
	cmd.Flags().IntVarP(&days, "days", "d", 7, "Show events for the next N days (0 for no limit)")
	cmd.Flags().BoolVar(&showAll, "show-cancelled", false, "Include cancelled events")
	cmd.Flags().StringVar(&targetTZ, "timezone", "", "Display times in this timezone (e.g., America/Los_Angeles). Defaults to local timezone.")
	cmd.Flags().BoolVar(&showTZ, "show-tz", false, "Show timezone abbreviations (e.g., PST, EST)")

	return cmd
}
