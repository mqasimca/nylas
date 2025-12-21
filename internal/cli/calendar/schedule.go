package calendar

import (
	"github.com/spf13/cobra"
)

// newScheduleCmd creates the schedule command group.
func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Schedule meetings with AI assistance",
		Long: `Schedule meetings using AI-powered natural language processing.

The AI will understand your request, analyze participant timezones,
check availability, and suggest optimal meeting times.`,
	}

	cmd.AddCommand(newAIScheduleCmd())

	return cmd
}
