package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/analytics"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

func newRescheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reschedule",
		Short: "AI-powered meeting rescheduling",
		Long: `Intelligently reschedule meetings with AI-powered conflict detection and alternative time suggestions.

Analyzes your calendar patterns to suggest optimal alternative times that work for all participants.`,
	}

	cmd.AddCommand(newAIRescheduleCmd())

	return cmd
}

func newAIRescheduleCmd() *cobra.Command {
	var (
		reason         string
		preferredTimes []string
		maxDelayDays   int
		notify         bool
		autoSelect     bool
		mustInclude    []string
		avoidDays      []string
	)

	cmd := &cobra.Command{
		Use:   "ai <event-id>",
		Short: "AI-powered rescheduling with smart suggestions",
		Example: `  # Get AI suggestions for rescheduling
  nylas calendar reschedule ai event_abc123 \
    --reason "Conflict with client meeting"

  # Reschedule with constraints
  nylas calendar reschedule ai event_abc123 \
    --max-delay-days 7 \
    --avoid-days Friday \
    --must-include john@company.com

  # Auto-select best time and notify participants
  nylas calendar reschedule ai event_abc123 \
    --reason "Calendar conflict" \
    --auto-select \
    --notify`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]

			_, err := common.WithClient([]string{}, func(ctx context.Context, client ports.NylasClient, grantID string) (struct{}, error) {
				// Fetch the event to reschedule
				fmt.Printf("📅 Fetching event %s...\n", eventID)
				event, err := fetchEventByID(ctx, client, grantID, eventID)
				if err != nil {
					return struct{}{}, common.WrapFetchError("event", err)
				}

				fmt.Printf("✓ Found: %s\n", event.Title)
				fmt.Printf("  Current time: %s\n",
					time.Unix(event.When.StartTime, 0).Format(common.DisplayWeekdayDateTime))

				// Analyze patterns
				fmt.Println("\n🔍 Analyzing your calendar patterns...")
				learner := analytics.NewPatternLearner(client)
				analysis, err := learner.AnalyzeHistory(ctx, grantID, 90)
				if err != nil {
					fmt.Printf("⚠️  Could not analyze patterns: %v\n", err)
				}

				var patterns *domain.MeetingPattern
				if analysis != nil && analysis.Patterns != nil {
					patterns = analysis.Patterns
				}

				// Create conflict resolver
				resolver := analytics.NewConflictResolver(client, patterns)

				// Parse preferred times if provided
				var preferredTimeParsed []time.Time
				for _, timeStr := range preferredTimes {
					t, err := time.Parse(time.RFC3339, timeStr)
					if err != nil {
						return struct{}{}, common.NewUserError(fmt.Sprintf("invalid preferred time %q", timeStr), "use RFC3339 format")
					}
					preferredTimeParsed = append(preferredTimeParsed, t)
				}

				// Build reschedule request
				request := &domain.RescheduleRequest{
					EventID:            eventID,
					Reason:             reason,
					PreferredTimes:     preferredTimeParsed,
					MustInclude:        mustInclude,
					AvoidDays:          avoidDays,
					MaxDelayDays:       maxDelayDays,
					NotifyParticipants: notify,
				}

				// Get reschedule suggestions
				fmt.Println("\n⚙️  Finding optimal alternative times...")
				suggestions, err := findRescheduleSuggestions(ctx, client, resolver, grantID, event, request, patterns)
				if err != nil {
					return struct{}{}, common.WrapGetError("reschedule suggestions", err)
				}

				// Display suggestions
				displayRescheduleSuggestions(event, suggestions, reason)

				// Handle auto-select
				if autoSelect && len(suggestions) > 0 {
					fmt.Println("\n🤖 Auto-selecting best alternative...")
					best := suggestions[0]
					result, err := applyReschedule(ctx, client, grantID, event, best, notify, reason)
					if err != nil {
						return struct{}{}, common.WrapUpdateError("reschedule", err)
					}

					displayRescheduleResult(result)
				} else if len(suggestions) > 0 {
					fmt.Println("\n💡 To apply a suggestion, use:")
					fmt.Printf("   nylas calendar events update %s --start %s\n",
						eventID,
						suggestions[0].ProposedTime.Format(time.RFC3339))
				}

				return struct{}{}, nil
			})
			return err
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "Reason for rescheduling")
	cmd.Flags().StringSliceVar(&preferredTimes, "preferred-times", nil, "Preferred alternative times (RFC3339 format)")
	cmd.Flags().IntVar(&maxDelayDays, "max-delay-days", 14, "Maximum days to delay the meeting")
	cmd.Flags().BoolVar(&notify, "notify", false, "Send notification to participants")
	cmd.Flags().BoolVar(&autoSelect, "auto-select", false, "Automatically select and apply best alternative")
	cmd.Flags().StringSliceVar(&mustInclude, "must-include", nil, "Participant emails that must be available")
	cmd.Flags().StringSliceVar(&avoidDays, "avoid-days", nil, "Days to avoid (e.g., Friday, Monday)")

	return cmd
}
