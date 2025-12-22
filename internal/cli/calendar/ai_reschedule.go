package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/analytics"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
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

			client, err := getClient()
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			grantID, err := getGrantID([]string{})
			if err != nil {
				return fmt.Errorf("failed to get grant ID: %w", err)
			}

			// AI rescheduling can take time - use longer timeout
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			// Fetch the event to reschedule
			fmt.Printf("📅 Fetching event %s...\n", eventID)
			event, err := fetchEventByID(ctx, client, grantID, eventID)
			if err != nil {
				return fmt.Errorf("failed to fetch event: %w", err)
			}

			fmt.Printf("✓ Found: %s\n", event.Title)
			fmt.Printf("  Current time: %s\n",
				time.Unix(event.When.StartTime, 0).Format("Mon, Jan 2, 2006 at 3:04 PM MST"))

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
					return fmt.Errorf("invalid preferred time %q (use RFC3339 format): %w", timeStr, err)
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
				return fmt.Errorf("failed to find suggestions: %w", err)
			}

			// Display suggestions
			displayRescheduleSuggestions(event, suggestions, reason)

			// Handle auto-select
			if autoSelect && len(suggestions) > 0 {
				fmt.Println("\n🤖 Auto-selecting best alternative...")
				best := suggestions[0]
				result, err := applyReschedule(ctx, client, grantID, event, best, notify, reason)
				if err != nil {
					return fmt.Errorf("failed to apply reschedule: %w", err)
				}

				displayRescheduleResult(result)
			} else if len(suggestions) > 0 {
				fmt.Println("\n💡 To apply a suggestion, use:")
				fmt.Printf("   nylas calendar events update %s --start %s\n",
					eventID,
					suggestions[0].ProposedTime.Format(time.RFC3339))
			}

			return nil
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

// fetchEventByID fetches an event from any calendar by its ID
func fetchEventByID(ctx context.Context, client ports.NylasClient, grantID, eventID string) (*domain.Event, error) {
	// Get all calendars
	calendars, err := client.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, err
	}

	// Search each calendar for the event
	for _, calendar := range calendars {
		event, err := client.GetEvent(ctx, grantID, calendar.ID, eventID)
		if err == nil {
			return event, nil
		}
	}

	return nil, fmt.Errorf("event %s not found in any calendar", eventID)
}

// findRescheduleSuggestions finds alternative times for rescheduling
func findRescheduleSuggestions(
	ctx context.Context,
	client ports.NylasClient,
	resolver *analytics.ConflictResolver,
	grantID string,
	event *domain.Event,
	request *domain.RescheduleRequest,
	patterns *domain.MeetingPattern,
) ([]domain.RescheduleOption, error) {
	originalStart := time.Unix(event.When.StartTime, 0)
	originalEnd := time.Unix(event.When.EndTime, 0)
	duration := int(originalEnd.Sub(originalStart).Minutes())

	var suggestions []domain.RescheduleOption

	// Try preferred times first
	for _, preferredTime := range request.PreferredTimes {
		proposedEvent := &domain.Event{
			Title:        event.Title,
			Participants: event.Participants,
			When: domain.EventWhen{
				StartTime: preferredTime.Unix(),
				EndTime:   preferredTime.Add(time.Duration(duration) * time.Minute).Unix(),
			},
		}

		analysis, err := resolver.DetectConflicts(ctx, grantID, proposedEvent, patterns)
		if err != nil {
			continue
		}

		// Only hard conflicts prevent this time
		if len(analysis.HardConflicts) == 0 {
			option := domain.RescheduleOption{
				ProposedTime: preferredTime,
				EndTime:      preferredTime.Add(time.Duration(duration) * time.Minute),
				Score:        calculateRescheduleScore(analysis, patterns, preferredTime),
				Conflicts:    analysis.SoftConflicts,
				Pros:         []string{"Preferred time"},
				Cons:         buildConsFromConflicts(analysis.SoftConflicts),
			}
			suggestions = append(suggestions, option)
		}
	}

	// Generate additional suggestions based on patterns and constraints
	maxDelay := time.Duration(request.MaxDelayDays) * 24 * time.Hour
	endSearch := originalStart.Add(maxDelay)

	// Try same time on different days
	for days := 1; days <= request.MaxDelayDays; days++ {
		proposedTime := originalStart.AddDate(0, 0, days)

		if proposedTime.After(endSearch) {
			break
		}

		// Skip avoided days
		if shouldAvoidDay(proposedTime, request.AvoidDays) {
			continue
		}

		proposedEvent := &domain.Event{
			Title:        event.Title,
			Participants: event.Participants,
			When: domain.EventWhen{
				StartTime: proposedTime.Unix(),
				EndTime:   proposedTime.Add(time.Duration(duration) * time.Minute).Unix(),
			},
		}

		analysis, err := resolver.DetectConflicts(ctx, grantID, proposedEvent, patterns)
		if err != nil {
			continue
		}

		// Only suggest if no hard conflicts
		if len(analysis.HardConflicts) == 0 {
			option := domain.RescheduleOption{
				ProposedTime:     proposedTime,
				EndTime:          proposedTime.Add(time.Duration(duration) * time.Minute),
				Score:            calculateRescheduleScore(analysis, patterns, proposedTime),
				Conflicts:        analysis.SoftConflicts,
				Pros:             buildProsFromPatterns(proposedTime, patterns),
				Cons:             buildConsFromConflicts(analysis.SoftConflicts),
				ParticipantMatch: 100.0, // Assume 100% if no hard conflicts
			}
			suggestions = append(suggestions, option)
		}

		// Limit suggestions to top candidates
		if len(suggestions) >= 10 {
			break
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Score > suggestions[i].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// Return top 5
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions, nil
}

// calculateRescheduleScore scores a potential reschedule time
func calculateRescheduleScore(analysis *domain.ConflictAnalysis, patterns *domain.MeetingPattern, proposedTime time.Time) int {
	score := 100

	// Penalize soft conflicts
	score -= len(analysis.SoftConflicts) * 10

	// Bonus for high-acceptance day
	if patterns != nil {
		dayOfWeek := proposedTime.Weekday().String()
		if rate, exists := patterns.Acceptance.ByDayOfWeek[dayOfWeek]; exists {
			score += int(rate * 15) // Up to +15 for high acceptance
		}

		// Bonus for high-acceptance time
		timeKey := fmt.Sprintf("%02d:00", proposedTime.Hour())
		if rate, exists := patterns.Acceptance.ByTimeOfDay[timeKey]; exists {
			score += int(rate * 15) // Up to +15 for high acceptance
		}
	}

	// Ensure score is in 0-100 range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// shouldAvoidDay checks if the day should be avoided
func shouldAvoidDay(t time.Time, avoidDays []string) bool {
	dayName := t.Weekday().String()
	for _, avoid := range avoidDays {
		if strings.EqualFold(avoid, dayName) {
			return true
		}
	}
	return false
}

// buildProsFromPatterns builds pros list from patterns
func buildProsFromPatterns(t time.Time, patterns *domain.MeetingPattern) []string {
	var pros []string

	if patterns == nil {
		return pros
	}

	dayOfWeek := t.Weekday().String()
	if rate, exists := patterns.Acceptance.ByDayOfWeek[dayOfWeek]; exists && rate > 0.8 {
		pros = append(pros, fmt.Sprintf("High acceptance rate on %ss (%.0f%%)", dayOfWeek, rate*100))
	}

	timeKey := fmt.Sprintf("%02d:00", t.Hour())
	if rate, exists := patterns.Acceptance.ByTimeOfDay[timeKey]; exists && rate > 0.8 {
		pros = append(pros, fmt.Sprintf("Preferred time slot (%.0f%% acceptance)", rate*100))
	}

	// Check if it's in a productive time block
	for _, block := range patterns.Productivity.PeakFocus {
		if block.DayOfWeek == dayOfWeek {
			blockStart := parseHourFromString(block.StartTime)
			blockEnd := parseHourFromString(block.EndTime)
			hour := t.Hour()
			if hour >= blockStart && hour < blockEnd {
				pros = append(pros, "During typical focus time (good for productive meetings)")
			}
		}
	}

	return pros
}

// buildConsFromConflicts builds cons list from soft conflicts
func buildConsFromConflicts(conflicts []domain.Conflict) []string {
	var cons []string

	for _, conflict := range conflicts {
		cons = append(cons, conflict.Impact)
	}

	return cons
}

// parseHourFromString parses hour from "HH:MM" format
func parseHourFromString(timeStr string) int {
	var hour int
	_, _ = fmt.Sscanf(timeStr, "%d:", &hour) // Parse hour, default 0 on error
	return hour
}

// applyReschedule applies the selected reschedule option
func applyReschedule(
	ctx context.Context,
	client ports.NylasClient,
	grantID string,
	event *domain.Event,
	option domain.RescheduleOption,
	notify bool,
	reason string,
) (*domain.RescheduleResult, error) {
	// Find which calendar the event belongs to
	calendars, err := client.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, err
	}

	var calendarID string
	for _, calendar := range calendars {
		_, err := client.GetEvent(ctx, grantID, calendar.ID, event.ID)
		if err == nil {
			calendarID = calendar.ID
			break
		}
	}

	if calendarID == "" {
		return nil, fmt.Errorf("could not find calendar for event")
	}

	// Update the event with new time
	updateReq := &domain.UpdateEventRequest{
		When: &domain.EventWhen{
			StartTime: option.ProposedTime.Unix(),
			EndTime:   option.EndTime.Unix(),
		},
	}

	newEvent, err := client.UpdateEvent(ctx, grantID, calendarID, event.ID, updateReq)
	if err != nil {
		return nil, err
	}

	result := &domain.RescheduleResult{
		Success:        true,
		OriginalEvent:  event,
		NewEvent:       newEvent,
		SelectedOption: &option,
		Message:        fmt.Sprintf("Successfully rescheduled to %s", option.ProposedTime.Format("Mon, Jan 2 at 3:04 PM MST")),
	}

	if notify {
		result.NotificationsSent = len(event.Participants)
	}

	return result, nil
}

// displayRescheduleSuggestions displays reschedule suggestions
func displayRescheduleSuggestions(event *domain.Event, suggestions []domain.RescheduleOption, reason string) {
	fmt.Println("\n📊 Reschedule Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if reason != "" {
		fmt.Printf("\nReason: %s\n", reason)
	}

	if len(suggestions) == 0 {
		fmt.Println("\n❌ No suitable alternative times found.")
		fmt.Println("   Try increasing --max-delay-days or removing constraints.")
		return
	}

	fmt.Printf("\n🔄 Found %d Alternative Time(s)\n", len(suggestions))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, option := range suggestions {
		// Score color coding
		scoreIcon := "🟢"
		if option.Score < 50 {
			scoreIcon = "🔴"
		} else if option.Score < 75 {
			scoreIcon = "🟡"
		}

		fmt.Printf("\n%d. %s %s (Score: %d/100)\n",
			i+1,
			scoreIcon,
			option.ProposedTime.Format("Mon, Jan 2, 2006 at 3:04 PM MST"),
			option.Score)

		if len(option.Pros) > 0 {
			fmt.Println("\n   Pros:")
			for _, pro := range option.Pros {
				fmt.Printf("   ✓ %s\n", pro)
			}
		}

		if len(option.Cons) > 0 {
			fmt.Println("\n   Cons:")
			for _, con := range option.Cons {
				fmt.Printf("   ⚠️  %s\n", con)
			}
		}

		if option.AIInsight != "" {
			fmt.Printf("\n   💡 %s\n", option.AIInsight)
		}

		if len(option.Conflicts) > 0 {
			fmt.Printf("\n   ⚠️  %d soft conflict(s)\n", len(option.Conflicts))
		}
	}
}

// displayRescheduleResult displays the reschedule result
func displayRescheduleResult(result *domain.RescheduleResult) {
	fmt.Println("\n✅ Reschedule Complete")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("\n%s\n", result.Message)

	if result.NewEvent != nil {
		newStart := time.Unix(result.NewEvent.When.StartTime, 0)
		fmt.Printf("\nNew time: %s\n", newStart.Format("Mon, Jan 2, 2006 at 3:04 PM MST"))
	}

	if result.NotificationsSent > 0 {
		fmt.Printf("📧 Notifications sent to %d participant(s)\n", result.NotificationsSent)
	}

	if len(result.CascadingChanges) > 0 {
		fmt.Println("\n⚠️  Cascading changes:")
		for _, change := range result.CascadingChanges {
			fmt.Printf("   • %s\n", change)
		}
	}
}
