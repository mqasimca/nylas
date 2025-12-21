package calendar

import (
	"github.com/spf13/cobra"
)

// newAICmd creates the AI command group for calendar intelligence features.
func newAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ai",
		Aliases: []string{"intelligence"},
		Short:   "AI-powered calendar intelligence",
		Long: `AI-powered calendar intelligence features.

Use machine learning to analyze your calendar patterns and optimize your schedule:
- Email thread analysis: Extract meeting context from email threads
- Focus time protection: Automatically block time for deep work
- Adaptive scheduling: Real-time schedule optimization
- Conflict resolution: Smart conflict detection and resolution
- Meeting analysis: Learn from historical patterns`,
		Example: `  # Analyze email thread for meeting context
  nylas calendar ai analyze-thread --thread thread_abc123

  # Analyze productivity patterns
  nylas calendar ai analyze

  # Enable focus time protection
  nylas calendar ai focus-time --enable

  # Detect and resolve conflicts
  nylas calendar ai conflicts

  # Adaptive schedule optimization
  nylas calendar ai adapt

  # Smart reschedule suggestions
  nylas calendar ai reschedule <event-id>`,
	}

	// Add AI subcommands
	cmd.AddCommand(newAnalyzeThreadCmd())
	cmd.AddCommand(newAnalyzeCmd())
	cmd.AddCommand(newConflictsCmd())
	cmd.AddCommand(newRescheduleCmd())
	cmd.AddCommand(newFocusTimeCmd())
	cmd.AddCommand(newAdaptCmd())

	return cmd
}
