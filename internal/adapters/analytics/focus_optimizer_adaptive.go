package analytics

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
)

// Helper functions for adaptive scheduling
func (f *FocusOptimizer) isLowPriorityMeeting(event *domain.Event) bool {
	// Check if meeting is low priority based on patterns
	// Simplified for now
	return len(event.Participants) <= 2
}

func (f *FocusOptimizer) conflictsWithFocusTime(event *domain.Event) bool {
	// Check if event overlaps with recommended focus time
	// Simplified for now
	return false
}

func (f *FocusOptimizer) canReschedule(event *domain.Event) bool {
	// Check if event can be rescheduled
	// Simplified for now
	return !event.ReadOnly
}

func (f *FocusOptimizer) determineChangeType(modifications []domain.ScheduleModification) domain.AdaptiveChangeType {
	if len(modifications) == 0 {
		return domain.ChangeTypeProtectBlock
	}

	// Return most common change type
	for _, mod := range modifications {
		switch mod.Action {
		case "reschedule":
			return domain.ChangeTypeRescheduleMeeting
		case "shorten":
			return domain.ChangeTypeShortenMeeting
		case "decline":
			return domain.ChangeTypeDeclineMeeting
		}
	}

	return domain.ChangeTypeProtectBlock
}

func (f *FocusOptimizer) extractEventIDs(modifications []domain.ScheduleModification) []string {
	var ids []string
	for _, mod := range modifications {
		if mod.EventID != "" {
			ids = append(ids, mod.EventID)
		}
	}
	return ids
}

func (f *FocusOptimizer) calculateAdaptiveImpact(modifications []domain.ScheduleModification) domain.AdaptiveImpact {
	impact := domain.AdaptiveImpact{
		FocusTimeGained:      2.0, // Hours
		MeetingsRescheduled:  0,
		MeetingsDeclined:     0,
		DurationSaved:        0,
		ConflictsResolved:    0,
		ParticipantsAffected: 0,
		PredictedBenefit:     "Improved focus time availability",
	}

	for _, mod := range modifications {
		switch mod.Action {
		case "reschedule":
			impact.MeetingsRescheduled++
		case "decline":
			impact.MeetingsDeclined++
		case "shorten":
			impact.DurationSaved += mod.OldDuration - mod.NewDuration
		}
	}

	return impact
}

func (f *FocusOptimizer) explainAdaptiveReason(trigger domain.AdaptiveTrigger, impact domain.AdaptiveImpact) string {
	switch trigger {
	case domain.TriggerMeetingOverload:
		return fmt.Sprintf("Meeting load increased: reducing by rescheduling %d meetings", impact.MeetingsRescheduled)
	case domain.TriggerFocusTimeAtRisk:
		return fmt.Sprintf("Focus time at risk: protecting %.1f additional hours", impact.FocusTimeGained)
	case domain.TriggerDeadlineChange:
		return "Urgent deadline detected: increasing focus time priority"
	default:
		return "Schedule optimization recommended"
	}
}

func (f *FocusOptimizer) calculateAdaptiveConfidence(modifications []domain.ScheduleModification) float64 {
	if len(modifications) == 0 {
		return 50.0
	}

	// Higher confidence for more modifications (more data)
	confidence := 60.0 + float64(min(len(modifications), 10))*3.0

	if confidence > 95.0 {
		confidence = 95.0
	}

	return confidence
}

// OptimizeMeetingDuration analyzes meetings and recommends duration optimizations.
func (f *FocusOptimizer) OptimizeMeetingDuration(ctx context.Context, grantID string, calendarID string, eventID string) (*domain.DurationOptimization, error) {
	// Get event details
	event, err := f.nylasClient.GetEvent(ctx, grantID, calendarID, eventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	// Get historical data for similar meetings
	analysis, err := f.patternLearner.AnalyzeHistory(ctx, grantID, 90)
	if err != nil {
		return nil, fmt.Errorf("analyze history: %w", err)
	}

	if analysis.Patterns == nil {
		// No patterns available, can't optimize
		return nil, fmt.Errorf("not enough historical data for duration optimization")
	}

	patterns := analysis.Patterns

	// Calculate current duration
	currentDuration := int((event.When.EndDateTime().Sub(event.When.StartDateTime())).Minutes())

	// Get historical duration data
	historicalData := patterns.Duration.Overall

	// Recommend optimized duration (typically shorter based on actual usage)
	recommendedDuration := historicalData.AverageActual

	// Apply common optimization: if scheduled for 60 min but avg actual is ~45, recommend 45
	if currentDuration == 60 && historicalData.AverageActual < 50 {
		recommendedDuration = 45
	} else if currentDuration == 30 && historicalData.AverageActual < 25 {
		recommendedDuration = 25
	}

	timeSavings := currentDuration - recommendedDuration
	if timeSavings < 0 {
		timeSavings = 0
	}

	optimization := &domain.DurationOptimization{
		EventID:             eventID,
		CurrentDuration:     currentDuration,
		RecommendedDuration: recommendedDuration,
		HistoricalData:      historicalData,
		TimeSavings:         timeSavings,
		Confidence:          f.calculateDurationConfidence(historicalData),
		Reason:              fmt.Sprintf("Historical data shows meetings average %d minutes", historicalData.AverageActual),
		Recommendation:      fmt.Sprintf("Reduce from %d to %d minutes to save %d minutes", currentDuration, recommendedDuration, timeSavings),
	}

	return optimization, nil
}

// calculateDurationConfidence calculates confidence in duration recommendations.
func (f *FocusOptimizer) calculateDurationConfidence(stats domain.DurationStats) float64 {
	// Higher confidence if low variance (consistent meeting lengths)
	if stats.Variance < 10.0 {
		return 90.0
	} else if stats.Variance < 20.0 {
		return 75.0
	} else if stats.Variance < 30.0 {
		return 60.0
	}
	return 50.0
}
