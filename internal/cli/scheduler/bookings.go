package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newBookingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bookings",
		Aliases: []string{"booking"},
		Short:   "Manage scheduler bookings",
		Long:    "Manage scheduler bookings (scheduled meetings).",
	}

	cmd.AddCommand(newBookingListCmd())
	cmd.AddCommand(newBookingShowCmd())
	cmd.AddCommand(newBookingConfirmCmd())
	cmd.AddCommand(newBookingRescheduleCmd())
	cmd.AddCommand(newBookingCancelCmd())

	return cmd
}

func newBookingListCmd() *cobra.Command {
	var (
		configID   string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List scheduler bookings",
		Long:    "List all scheduler bookings, optionally filtered by configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			bookings, err := client.ListBookings(ctx, configID)
			if err != nil {
				return fmt.Errorf("failed to list bookings: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(bookings)
			}

			if len(bookings) == 0 {
				fmt.Println("No bookings found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)

			fmt.Printf("Found %d booking(s):\n\n", len(bookings))

			table := common.NewTable("TITLE", "ID", "START TIME", "STATUS")
			for _, b := range bookings {
				status := b.Status
				switch b.Status {
				case "confirmed":
					status = green.Sprint(status)
				case "pending":
					status = yellow.Sprint(status)
				case "cancelled":
					status = color.New(color.Faint).Sprint(status)
				}

				startTime := b.StartTime.Format("2006-01-02 15:04")
				table.AddRow(cyan.Sprint(b.Title), b.BookingID, startTime, status)
			}
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringVar(&configID, "config-id", "", "Filter by configuration ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newBookingShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <booking-id>",
		Short: "Show booking details",
		Long:  "Show detailed information about a specific booking.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			booking, err := client.GetBooking(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get booking: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(booking)
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			bold := color.New(color.Bold)

			_, _ = bold.Printf("Booking: %s\n", booking.Title)
			fmt.Printf("  ID: %s\n", cyan.Sprint(booking.BookingID))
			fmt.Printf("  Status: %s\n", getStatusColor(booking.Status).Sprint(booking.Status))
			fmt.Printf("  Start: %s\n", booking.StartTime.Format(time.RFC1123))
			fmt.Printf("  End: %s\n", booking.EndTime.Format(time.RFC1123))

			if booking.EventID != "" {
				fmt.Printf("  Event ID: %s\n", booking.EventID)
			}

			if len(booking.Participants) > 0 {
				fmt.Printf("\nParticipants (%d):\n", len(booking.Participants))
				for i, p := range booking.Participants {
					fmt.Printf("  %d. %s <%s>", i+1, p.Name, p.Email)
					if p.Status == "yes" {
						fmt.Printf(" %s", green.Sprint("✓"))
					}
					fmt.Println()
				}
			}

			if booking.Conferencing != nil && booking.Conferencing.URL != "" {
				fmt.Printf("\nConferencing:\n")
				fmt.Printf("  URL: %s\n", cyan.Sprint(booking.Conferencing.URL))
				if booking.Conferencing.MeetingCode != "" {
					fmt.Printf("  Meeting Code: %s\n", booking.Conferencing.MeetingCode)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newBookingConfirmCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "confirm <booking-id>",
		Short: "Confirm a booking",
		Long:  "Confirm a pending booking.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := &domain.ConfirmBookingRequest{
				Status: "confirmed",
				Reason: reason,
			}

			ctx, cancel := createContext()
			defer cancel()

			booking, err := client.ConfirmBooking(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("failed to confirm booking: %w", err)
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Confirmed booking: %s\n", booking.BookingID)
			fmt.Printf("  Status: %s\n", booking.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "Reason for confirmation")

	return cmd
}

func newBookingRescheduleCmd() *cobra.Command {
	var (
		startTime int64
		endTime   int64
		timezone  string
		reason    string
	)

	cmd := &cobra.Command{
		Use:   "reschedule <booking-id>",
		Short: "Reschedule a booking",
		Long: `Reschedule an existing booking to a new time.

You must provide the new start and end times as Unix timestamps.`,
		Example: `  # Reschedule to a new time
  nylas scheduler bookings reschedule abc123 --start-time 1704067200 --end-time 1704070800

  # Reschedule with timezone
  nylas scheduler bookings reschedule abc123 --start-time 1704067200 --end-time 1704070800 --timezone "America/New_York"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if startTime == 0 || endTime == 0 {
				return fmt.Errorf("both --start-time and --end-time are required")
			}

			if endTime <= startTime {
				return fmt.Errorf("end-time must be after start-time")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			req := &domain.RescheduleBookingRequest{
				StartTime: startTime,
				EndTime:   endTime,
				Timezone:  timezone,
				Reason:    reason,
			}

			ctx, cancel := createContext()
			defer cancel()

			booking, err := client.RescheduleBooking(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("failed to reschedule booking: %w", err)
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Rescheduled booking: %s\n", booking.BookingID)
			fmt.Printf("  New start: %s\n", booking.StartTime.Format(time.RFC1123))
			fmt.Printf("  New end: %s\n", booking.EndTime.Format(time.RFC1123))

			return nil
		},
	}

	cmd.Flags().Int64Var(&startTime, "start-time", 0, "New start time (Unix timestamp, required)")
	cmd.Flags().Int64Var(&endTime, "end-time", 0, "New end time (Unix timestamp, required)")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone for the booking (e.g., America/New_York)")
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for rescheduling")

	return cmd
}

func newBookingCancelCmd() *cobra.Command {
	var (
		reason string
		yes    bool
	)

	cmd := &cobra.Command{
		Use:   "cancel <booking-id>",
		Short: "Cancel a booking",
		Long:  "Cancel a scheduled booking.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Printf("Are you sure you want to cancel booking %s? (y/N): ", args[0])
				var confirm string
				_, _ = fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			if err := client.CancelBooking(ctx, args[0], reason); err != nil {
				return fmt.Errorf("failed to cancel booking: %w", err)
			}

			green := color.New(color.FgGreen)
			_, _ = green.Printf("✓ Cancelled booking: %s\n", args[0])

			return nil
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "Cancellation reason")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func getStatusColor(status string) *color.Color {
	switch status {
	case "confirmed":
		return color.New(color.FgGreen)
	case "pending":
		return color.New(color.FgYellow)
	case "cancelled":
		return color.New(color.Faint)
	default:
		return color.New(color.Reset)
	}
}
