package timezone

import (
	"github.com/spf13/cobra"
)

// NewTimezoneCmd creates the timezone command.
func NewTimezoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timezone",
		Short: "Time zone conversion and meeting scheduling utilities",
		Long: `Utilities for working with time zones, converting times, finding meeting slots,
and managing DST transitions. All operations work offline without API access.

Examples:
  # Convert current time from PST to IST
  nylas timezone convert --from America/Los_Angeles --to Asia/Kolkata

  # Find meeting time across multiple zones
  nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo" --duration 1h

  # Check DST transitions for a zone
  nylas timezone dst --zone America/New_York --year 2026

  # List all available time zones
  nylas timezone list

  # Get detailed info about a time zone
  nylas timezone info America/Los_Angeles`,
	}

	// Add subcommands
	cmd.AddCommand(newConvertCmd())
	cmd.AddCommand(newFindMeetingCmd())
	cmd.AddCommand(newDSTCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())

	return cmd
}
