package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// ConflictResolver detects and resolves scheduling conflicts.
type ConflictResolver struct {
	calendarClient CalendarClient
	scorer         *MeetingScorer
	learner        *PatternLearner
}

// NewConflictResolver creates a new conflict resolver.
func NewConflictResolver(client CalendarClient, patterns *domain.MeetingPattern) *ConflictResolver {
	return &ConflictResolver{
		calendarClient: client,
		scorer:         NewMeetingScorer(patterns),
		learner:        NewPatternLearner(client),
	}
}

// DetectConflicts analyzes a proposed event for conflicts.
func (cr *ConflictResolver) DetectConflicts(ctx context.Context, grantID string, proposed *domain.Event, patterns *domain.MeetingPattern) (*domain.ConflictAnalysis, error) {
	// Get existing events around the proposed time
	startTime := time.Unix(proposed.When.StartTime, 0)
	endTime := time.Unix(proposed.When.EndTime, 0)

	// Expand search window to detect soft conflicts
	searchStart := startTime.Add(-2 * time.Hour)
	searchEnd := endTime.Add(2 * time.Hour)

	// Fetch calendars and events
	calendars, err := cr.calendarClient.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendars: %w", err)
	}

	var allEvents []domain.Event
	params := &domain.EventQueryParams{
		Start: searchStart.Unix(),
		End:   searchEnd.Unix(),
	}

	for _, calendar := range calendars {
		events, err := cr.calendarClient.GetEvents(ctx, grantID, calendar.ID, params)
		if err != nil {
			continue
		}
		allEvents = append(allEvents, events...)
	}

	// Detect conflicts
	hardConflicts := cr.detectHardConflicts(proposed, allEvents)
	softConflicts := cr.detectSoftConflicts(proposed, allEvents, patterns)

	canProceed := len(hardConflicts) == 0

	// Generate recommendations
	recommendations := cr.generateConflictRecommendations(hardConflicts, softConflicts)

	// Generate alternative times if there are conflicts
	var alternatives []domain.RescheduleOption
	if len(hardConflicts) > 0 || len(softConflicts) > 2 {
		alternatives = cr.suggestAlternatives(ctx, grantID, proposed, allEvents, patterns)
	}

	// Generate AI recommendation
	aiRec := cr.generateAIRecommendation(hardConflicts, softConflicts, alternatives)

	return &domain.ConflictAnalysis{
		ProposedEvent:    proposed,
		HardConflicts:    hardConflicts,
		SoftConflicts:    softConflicts,
		TotalConflicts:   len(hardConflicts) + len(softConflicts),
		CanProceed:       canProceed,
		Recommendations:  recommendations,
		AlternativeTimes: alternatives,
		AIRecommendation: aiRec,
	}, nil
}

// detectHardConflicts finds overlapping meetings.
func (cr *ConflictResolver) detectHardConflicts(proposed *domain.Event, existing []domain.Event) []domain.Conflict {
	var conflicts []domain.Conflict

	proposedStart := time.Unix(proposed.When.StartTime, 0)
	proposedEnd := time.Unix(proposed.When.EndTime, 0)

	for _, event := range existing {
		eventStart := time.Unix(event.When.StartTime, 0)
		eventEnd := time.Unix(event.When.EndTime, 0)

		// Check for overlap
		if proposedStart.Before(eventEnd) && proposedEnd.After(eventStart) {
			severity := domain.SeverityCritical
			if event.Status == "tentative" {
				severity = domain.SeverityHigh
			}

			conflicts = append(conflicts, domain.Conflict{
				ID:               fmt.Sprintf("hard_%s", event.ID),
				Type:             domain.ConflictTypeHard,
				Severity:         severity,
				ProposedEvent:    proposed,
				ConflictingEvent: &event,
				Description:      fmt.Sprintf("Overlaps with '%s'", event.Title),
				Impact:           "Cannot attend both meetings simultaneously",
				Suggestion:       "Reschedule one of the meetings",
				CanAutoResolve:   false,
			})
		}
	}

	return conflicts
}

// detectSoftConflicts finds potential issues like back-to-back meetings.
func (cr *ConflictResolver) detectSoftConflicts(proposed *domain.Event, existing []domain.Event, patterns *domain.MeetingPattern) []domain.Conflict {
	var conflicts []domain.Conflict

	proposedStart := time.Unix(proposed.When.StartTime, 0)
	proposedEnd := time.Unix(proposed.When.EndTime, 0)

	// Check for back-to-back meetings (no buffer time)
	for _, event := range existing {
		eventStart := time.Unix(event.When.StartTime, 0)
		eventEnd := time.Unix(event.When.EndTime, 0)

		// Meeting ends exactly when proposed starts (or vice versa)
		if eventEnd.Equal(proposedStart) || proposedEnd.Equal(eventStart) {
			conflicts = append(conflicts, domain.Conflict{
				ID:               fmt.Sprintf("soft_b2b_%s", event.ID),
				Type:             domain.ConflictTypeSoftBackToBack,
				Severity:         domain.SeverityMedium,
				ProposedEvent:    proposed,
				ConflictingEvent: &event,
				Description:      fmt.Sprintf("Back-to-back with '%s'", event.Title),
				Impact:           "No buffer time for breaks or overruns",
				Suggestion:       "Add 15-minute buffer between meetings",
				CanAutoResolve:   true,
			})
		}

		// Very close meetings (< 15 min gap)
		gap := eventStart.Sub(proposedEnd)
		if gap > 0 && gap < 15*time.Minute {
			conflicts = append(conflicts, domain.Conflict{
				ID:               fmt.Sprintf("soft_close_%s", event.ID),
				Type:             domain.ConflictTypeSoftBackToBack,
				Severity:         domain.SeverityLow,
				ProposedEvent:    proposed,
				ConflictingEvent: &event,
				Description:      fmt.Sprintf("Only %d min gap before '%s'", int(gap.Minutes()), event.Title),
				Impact:           "Minimal buffer time",
				Suggestion:       "Consider adding more buffer time",
				CanAutoResolve:   true,
			})
		}
	}

	// Check if interrupting focus time
	if patterns != nil && len(patterns.Productivity.FocusBlocks) > 0 {
		for _, block := range patterns.Productivity.FocusBlocks {
			if cr.isInFocusBlock(proposedStart, block) {
				conflicts = append(conflicts, domain.Conflict{
					ID:             fmt.Sprintf("soft_focus_%s_%s", block.DayOfWeek, block.StartTime),
					Type:           domain.ConflictTypeSoftFocusTime,
					Severity:       domain.SeverityHigh,
					ProposedEvent:  proposed,
					Description:    fmt.Sprintf("Interrupts focus time (%s %s-%s)", block.DayOfWeek, block.StartTime, block.EndTime),
					Impact:         "Reduces productivity during peak focus hours",
					Suggestion:     "Schedule outside of focus time blocks",
					CanAutoResolve: true,
				})
			}
		}
	}

	// Check meeting overload (too many meetings in a day)
	meetingsOnDay := cr.countMeetingsOnDay(proposedStart, existing)
	if meetingsOnDay >= 6 {
		conflicts = append(conflicts, domain.Conflict{
			ID:             fmt.Sprintf("soft_overload_%s", proposedStart.Format("2006-01-02")),
			Type:           domain.ConflictTypeSoftOverload,
			Severity:       domain.SeverityMedium,
			ProposedEvent:  proposed,
			Description:    fmt.Sprintf("Already have %d meetings this day", meetingsOnDay),
			Impact:         "Meeting fatigue and reduced productivity",
			Suggestion:     "Consider spreading meetings across more days",
			CanAutoResolve: true,
		})
	}

	return conflicts
}

// isInFocusBlock checks if a time falls within a focus time block.
func (cr *ConflictResolver) isInFocusBlock(t time.Time, block domain.TimeBlock) bool {
	if t.Weekday().String() != block.DayOfWeek {
		return false
	}

	hour := t.Hour()
	startHour := parseHour(block.StartTime)
	endHour := parseHour(block.EndTime)

	return hour >= startHour && hour < endHour
}

// countMeetingsOnDay counts how many meetings are on the same day.
func (cr *ConflictResolver) countMeetingsOnDay(day time.Time, events []domain.Event) int {
	count := 0
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	for _, event := range events {
		eventTime := time.Unix(event.When.StartTime, 0)
		if eventTime.After(dayStart) && eventTime.Before(dayEnd) {
			count++
		}
	}

	return count
}

// generateConflictRecommendations creates actionable recommendations.
func (cr *ConflictResolver) generateConflictRecommendations(hard, soft []domain.Conflict) []string {
	var recommendations []string

	if len(hard) > 0 {
		recommendations = append(recommendations, "⚠️ Hard conflicts detected - must reschedule")
		for _, conflict := range hard {
			recommendations = append(recommendations, fmt.Sprintf("  • %s", conflict.Suggestion))
		}
	}

	if len(soft) > 2 {
		recommendations = append(recommendations, "⚠️ Multiple soft conflicts detected:")
		focusConflicts := 0
		b2bConflicts := 0

		for _, conflict := range soft {
			switch conflict.Type {
			case domain.ConflictTypeSoftFocusTime:
				focusConflicts++
			case domain.ConflictTypeSoftBackToBack:
				b2bConflicts++
			}
		}

		if focusConflicts > 0 {
			recommendations = append(recommendations, "  • Consider protecting your focus time")
		}
		if b2bConflicts > 1 {
			recommendations = append(recommendations, "  • Add buffer time between meetings")
		}
	}

	if len(hard) == 0 && len(soft) == 0 {
		recommendations = append(recommendations, "✓ No conflicts detected - good time for this meeting")
	}

	return recommendations
}

// suggestAlternatives finds better times for the meeting.
func (cr *ConflictResolver) suggestAlternatives(ctx context.Context, grantID string, proposed *domain.Event, existing []domain.Event, patterns *domain.MeetingPattern) []domain.RescheduleOption {
	var options []domain.RescheduleOption

	duration := time.Unix(proposed.When.EndTime, 0).Sub(time.Unix(proposed.When.StartTime, 0))
	startTime := time.Unix(proposed.When.StartTime, 0)

	// Try same day, later times
	for i := 1; i <= 4; i++ {
		altTime := startTime.Add(time.Duration(i) * time.Hour)
		option := cr.evaluateAlternative(altTime, duration, proposed, existing, patterns)
		if option != nil && option.Score > 50 {
			options = append(options, *option)
		}
	}

	// Try next day, same time
	nextDay := startTime.AddDate(0, 0, 1)
	option := cr.evaluateAlternative(nextDay, duration, proposed, existing, patterns)
	if option != nil && option.Score > 50 {
		options = append(options, *option)
	}

	// Try best time from patterns
	if patterns != nil && cr.scorer != nil {
		bestTime := cr.findBestTimeFromPatterns(startTime, duration, patterns)
		if bestTime != nil {
			option := cr.evaluateAlternative(*bestTime, duration, proposed, existing, patterns)
			if option != nil && option.Score > 50 {
				options = append(options, *option)
			}
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(options)-1; i++ {
		for j := i + 1; j < len(options); j++ {
			if options[j].Score > options[i].Score {
				options[i], options[j] = options[j], options[i]
			}
		}
	}

	// Return top 3
	if len(options) > 3 {
		options = options[:3]
	}

	return options
}

// evaluateAlternative scores an alternative time slot.
func (cr *ConflictResolver) evaluateAlternative(startTime time.Time, duration time.Duration, proposed *domain.Event, existing []domain.Event, patterns *domain.MeetingPattern) *domain.RescheduleOption {
	endTime := startTime.Add(duration)

	// Create temporary event for conflict checking
	tempEvent := &domain.Event{
		When: domain.EventWhen{
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
		},
	}

	// Check for conflicts
	hardConflicts := cr.detectHardConflicts(tempEvent, existing)
	if len(hardConflicts) > 0 {
		return nil // Skip times with hard conflicts
	}

	softConflicts := cr.detectSoftConflicts(tempEvent, existing, patterns)

	// Score using meeting scorer
	score := 70 // Base score
	if cr.scorer != nil && patterns != nil {
		meetingScore := cr.scorer.ScoreMeetingTime(startTime, []string{}, int(duration.Minutes()))
		score = meetingScore.Score
	}

	// Adjust score based on conflicts
	score -= len(softConflicts) * 10
	if score < 0 {
		score = 0
	}

	// Build pros and cons
	pros := []string{}
	cons := []string{}

	if len(softConflicts) == 0 {
		pros = append(pros, "No conflicts detected")
	}

	dayOfWeek := startTime.Weekday().String()
	if patterns != nil {
		if rate, exists := patterns.Acceptance.ByDayOfWeek[dayOfWeek]; exists && rate > 0.8 {
			pros = append(pros, fmt.Sprintf("High acceptance rate on %ss (%.0f%%)", dayOfWeek, rate*100))
		}
	}

	if len(softConflicts) > 0 {
		cons = append(cons, fmt.Sprintf("%d soft conflict(s)", len(softConflicts)))
	}

	// Calculate time difference
	originalTime := time.Unix(proposed.When.StartTime, 0)
	daysDiff := int(startTime.Sub(originalTime).Hours() / 24)
	if daysDiff > 0 {
		cons = append(cons, fmt.Sprintf("%d day delay", daysDiff))
	}

	return &domain.RescheduleOption{
		ProposedTime:     startTime,
		EndTime:          endTime,
		Score:            score,
		Confidence:       float64(score),
		Pros:             pros,
		Cons:             cons,
		Conflicts:        softConflicts,
		ParticipantMatch: 1.0, // Assume all available for now
		AIInsight:        cr.generateOptionInsight(score, softConflicts, daysDiff),
	}
}

// findBestTimeFromPatterns finds the best time based on learned patterns.
func (cr *ConflictResolver) findBestTimeFromPatterns(around time.Time, duration time.Duration, patterns *domain.MeetingPattern) *time.Time {
	if patterns == nil || len(patterns.Acceptance.ByDayOfWeek) == 0 {
		return nil
	}

	// Find best day
	bestDay := ""
	bestRate := 0.0
	for day, rate := range patterns.Acceptance.ByDayOfWeek {
		if rate > bestRate {
			bestRate = rate
			bestDay = day
		}
	}

	// Find best hour
	bestHour := 14 // Default to 2 PM
	if len(patterns.Acceptance.ByTimeOfDay) > 0 {
		bestHourRate := 0.0
		for hourStr, rate := range patterns.Acceptance.ByTimeOfDay {
			if rate > bestHourRate {
				bestHourRate = rate
				bestHour = parseHour(hourStr)
			}
		}
	}

	// Find next occurrence of best day
	daysUntil := cr.daysUntilWeekday(around, bestDay)
	targetDate := around.AddDate(0, 0, daysUntil)
	targetTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), bestHour, 0, 0, 0, around.Location())

	return &targetTime
}

// daysUntilWeekday calculates days until a specific weekday.
func (cr *ConflictResolver) daysUntilWeekday(from time.Time, targetDay string) int {
	target := dayNameToWeekday(targetDay)
	current := from.Weekday()

	days := int(target) - int(current)
	if days <= 0 {
		days += 7
	}

	return days
}

// generateAIRecommendation creates an AI recommendation message.
func (cr *ConflictResolver) generateAIRecommendation(hard, soft []domain.Conflict, alternatives []domain.RescheduleOption) string {
	if len(hard) > 0 {
		if len(alternatives) > 0 {
			return fmt.Sprintf("❌ Cannot proceed due to %d hard conflict(s). Recommend rescheduling to alternative time slot (Score: %d/100)", len(hard), alternatives[0].Score)
		}
		return fmt.Sprintf("❌ Cannot proceed due to %d hard conflict(s). Manual rescheduling required", len(hard))
	}

	if len(soft) > 2 {
		if len(alternatives) > 0 {
			return fmt.Sprintf("⚠️ Proceeding not recommended due to %d soft conflicts. Consider alternative time (Score: %d/100)", len(soft), alternatives[0].Score)
		}
		return fmt.Sprintf("⚠️ Proceeding possible but not ideal (%d soft conflicts)", len(soft))
	}

	if len(soft) > 0 {
		return fmt.Sprintf("✓ Can proceed with %d minor soft conflict(s)", len(soft))
	}

	return "✓ Excellent time - no conflicts detected"
}

// generateOptionInsight creates an insight for a reschedule option.
func (cr *ConflictResolver) generateOptionInsight(score int, conflicts []domain.Conflict, daysDiff int) string {
	if score >= 90 {
		return "Excellent alternative with minimal disruption"
	}
	if score >= 75 {
		if daysDiff == 0 {
			return "Same day alternative - minimal delay"
		}
		return "Good alternative with acceptable trade-offs"
	}
	if score >= 60 {
		return "Acceptable but consider other options"
	}
	return "Suboptimal - many conflicts remain"
}
