package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

func (f *FocusOptimizer) CreateProtectedBlocks(ctx context.Context, grantID string, blocks []domain.FocusTimeBlock, settings *domain.FocusTimeSettings) ([]*domain.ProtectedBlock, error) {
	// Get user's primary calendar
	calendars, err := f.calendarClient.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("get calendars: %w", err)
	}

	if len(calendars) == 0 {
		return nil, fmt.Errorf("no calendars found")
	}

	// Use primary calendar (first calendar)
	calendarID := calendars[0].ID

	var protectedBlocks []*domain.ProtectedBlock

	for _, block := range blocks {
		// Calculate next occurrence of this day/time
		startTime := f.nextOccurrence(block.DayOfWeek, block.StartTime)
		endTime := f.nextOccurrence(block.DayOfWeek, block.EndTime)

		// Create calendar event for focus time
		eventReq := &domain.CreateEventRequest{
			Title:       "Focus Time",
			Description: block.Reason,
			When: domain.EventWhen{
				StartTime: startTime.Unix(),
				EndTime:   endTime.Unix(),
			},
			Busy: true, // Show as busy
		}

		event, err := f.nylasClient.CreateEvent(ctx, grantID, calendarID, eventReq)
		if err != nil {
			return nil, fmt.Errorf("create calendar event: %w", err)
		}

		// Create protected block record
		protectedBlock := &domain.ProtectedBlock{
			ID:                fmt.Sprintf("focus_%d", time.Now().UnixNano()),
			CalendarEventID:   event.ID,
			StartTime:         startTime,
			EndTime:           endTime,
			Duration:          block.Duration,
			IsRecurring:       true,
			RecurrencePattern: "weekly",
			Priority:          domain.PriorityHigh,
			Reason:            block.Reason,
			AllowOverride:     settings.AllowUrgentOverride,
			ProtectionRules: domain.FocusProtectionRule{
				AutoDecline:          settings.AutoDecline,
				SuggestAlternatives:  true,
				AllowCriticalMeeting: settings.AllowUrgentOverride,
				RequireApproval:      settings.RequireApproval,
				DeclineMessage:       "This time is blocked for focus work. Alternative times are available.",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		protectedBlocks = append(protectedBlocks, protectedBlock)
	}

	return protectedBlocks, nil
}

// nextOccurrence finds the next occurrence of a specific day and time.
func (f *FocusOptimizer) nextOccurrence(dayOfWeek, timeStr string) time.Time {
	now := time.Now()

	// Parse target time
	targetTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return now
	}

	// Map day of week to time.Weekday
	dayMap := map[string]time.Weekday{
		"Sunday":    time.Sunday,
		"Monday":    time.Monday,
		"Tuesday":   time.Tuesday,
		"Wednesday": time.Wednesday,
		"Thursday":  time.Thursday,
		"Friday":    time.Friday,
		"Saturday":  time.Saturday,
	}

	targetDay, ok := dayMap[dayOfWeek]
	if !ok {
		return now
	}

	// Find next occurrence of target day
	daysUntil := (int(targetDay) - int(now.Weekday()) + 7) % 7
	if daysUntil == 0 && now.Hour() > targetTime.Hour() {
		daysUntil = 7 // If today but time passed, go to next week
	}

	nextDate := now.AddDate(0, 0, daysUntil)

	// Combine date with time
	result := time.Date(
		nextDate.Year(), nextDate.Month(), nextDate.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		now.Location(),
	)

	return result
}

// AdaptSchedule performs real-time adaptive schedule optimization.
func (f *FocusOptimizer) AdaptSchedule(ctx context.Context, grantID string, trigger domain.AdaptiveTrigger) (*domain.AdaptiveScheduleChange, error) {
	// Get upcoming events
	events, err := f.getUpcomingEvents(ctx, grantID, 14) // Next 2 weeks
	if err != nil {
		return nil, fmt.Errorf("get upcoming events: %w", err)
	}

	// Detect what needs to change based on trigger
	changes := f.detectRequiredChanges(ctx, grantID, events, trigger)

	// Calculate impact
	impact := f.calculateAdaptiveImpact(changes)

	// Create adaptive schedule change record
	change := &domain.AdaptiveScheduleChange{
		ID:             fmt.Sprintf("adapt_%d", time.Now().UnixNano()),
		Timestamp:      time.Now(),
		Trigger:        trigger,
		ChangeType:     f.determineChangeType(changes),
		AffectedEvents: f.extractEventIDs(changes),
		Changes:        changes,
		Reason:         f.explainAdaptiveReason(trigger, impact),
		Impact:         impact,
		UserApproval:   domain.ApprovalPending,
		AutoApplied:    false, // Require user approval
		Confidence:     f.calculateAdaptiveConfidence(changes),
	}

	return change, nil
}

// getUpcomingEvents gets events for the next N days.
func (f *FocusOptimizer) getUpcomingEvents(ctx context.Context, grantID string, days int) ([]domain.Event, error) {
	start := time.Now()
	end := start.AddDate(0, 0, days)

	// Get all calendars
	calendars, err := f.calendarClient.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("get calendars: %w", err)
	}

	// Fetch events from all calendars
	var allEvents []domain.Event
	params := &domain.EventQueryParams{
		Start: start.Unix(),
		End:   end.Unix(),
	}

	for _, calendar := range calendars {
		events, err := f.calendarClient.GetEvents(ctx, grantID, calendar.ID, params)
		if err != nil {
			continue
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// detectRequiredChanges detects what schedule changes are needed.
func (f *FocusOptimizer) detectRequiredChanges(ctx context.Context, grantID string, events []domain.Event, trigger domain.AdaptiveTrigger) []domain.ScheduleModification {
	var modifications []domain.ScheduleModification

	switch trigger {
	case domain.TriggerMeetingOverload:
		// Find low-priority meetings to reschedule or decline
		for _, event := range events {
			if f.isLowPriorityMeeting(&event) {
				modifications = append(modifications, domain.ScheduleModification{
					EventID:     event.ID,
					Action:      "reschedule",
					Description: "Move low-priority meeting to reduce meeting overload",
				})
			}
		}

	case domain.TriggerFocusTimeAtRisk:
		// Protect more focus time by moving meetings
		for _, event := range events {
			if f.conflictsWithFocusTime(&event) && f.canReschedule(&event) {
				modifications = append(modifications, domain.ScheduleModification{
					EventID:     event.ID,
					Action:      "reschedule",
					Description: "Move meeting to protect focus time",
				})
			}
		}

	case domain.TriggerDeadlineChange:
		// Increase focus time for urgent deadline
		modifications = append(modifications, domain.ScheduleModification{
			Action:      "protect",
			Description: "Add additional focus blocks due to deadline pressure",
		})
	}

	return modifications
}
