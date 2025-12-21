package analytics

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// FocusOptimizer provides AI-powered focus time protection and optimization.
type FocusOptimizer struct {
	calendarClient CalendarClient
	nylasClient    ports.NylasClient
	patternLearner *PatternLearner
}

// NewFocusOptimizer creates a new focus time optimizer.
func NewFocusOptimizer(nylasClient ports.NylasClient) *FocusOptimizer {
	return &FocusOptimizer{
		calendarClient: nylasClient, // NylasClient implements CalendarClient
		nylasClient:    nylasClient,
		patternLearner: NewPatternLearner(nylasClient),
	}
}

// AnalyzeFocusTimePatterns analyzes productivity patterns and recommends focus time blocks.
func (f *FocusOptimizer) AnalyzeFocusTimePatterns(ctx context.Context, grantID string, settings *domain.FocusTimeSettings) (*domain.FocusTimeAnalysis, error) {
	// Analyze calendar history to learn patterns
	analysis, err := f.patternLearner.AnalyzeHistory(ctx, grantID, 90) // Last 90 days
	if err != nil {
		return nil, fmt.Errorf("analyze history: %w", err)
	}

	if analysis.Patterns == nil {
		// No patterns available, return empty analysis
		return &domain.FocusTimeAnalysis{
			UserEmail:         grantID,
			AnalyzedPeriod:    analysis.Period,
			GeneratedAt:       time.Now(),
			RecommendedBlocks: []domain.FocusTimeBlock{},
			Insights:          []string{"Not enough calendar history to analyze patterns"},
			Confidence:        0,
		}, nil
	}

	patterns := analysis.Patterns

	// Calculate deep work session stats
	deepWorkStats := f.calculateDeepWorkStats(patterns)

	// Find peak productivity times
	peakProductivity := f.findPeakProductivityBlocks(patterns)

	// Identify most/least productive days
	mostProductiveDay := f.findMostProductiveDay(patterns)
	leastProductiveDay := f.findLeastProductiveDay(patterns)

	// Generate recommended focus blocks based on patterns and settings
	recommendedBlocks := f.generateRecommendedBlocks(patterns, settings)

	// Calculate current protection (existing focus blocks)
	currentProtection := f.calculateCurrentProtection(ctx, grantID)

	// Generate insights
	insights := f.generateInsights(patterns, recommendedBlocks, settings)

	// Calculate confidence based on data quality
	confidence := f.calculateConfidence(patterns)

	focusAnalysis := &domain.FocusTimeAnalysis{
		UserEmail:          grantID, // Using grantID as user email
		AnalyzedPeriod:     analysis.Period,
		GeneratedAt:        time.Now(),
		PeakProductivity:   peakProductivity,
		DeepWorkSessions:   deepWorkStats,
		MostProductiveDay:  mostProductiveDay,
		LeastProductiveDay: leastProductiveDay,
		RecommendedBlocks:  recommendedBlocks,
		CurrentProtection:  currentProtection,
		TargetProtection:   settings.TargetHoursPerWeek,
		Insights:           insights,
		Confidence:         confidence,
	}

	return focusAnalysis, nil
}

// calculateDeepWorkStats calculates statistics about deep work sessions.
func (f *FocusOptimizer) calculateDeepWorkStats(patterns *domain.MeetingPattern) domain.DurationStats {
	// Analyze gaps between meetings to find typical deep work session lengths
	var deepWorkDurations []int

	// Look at focus blocks if available
	if len(patterns.Productivity.FocusBlocks) > 0 {
		for _, block := range patterns.Productivity.FocusBlocks {
			duration := f.calculateBlockDuration(block.StartTime, block.EndTime)
			deepWorkDurations = append(deepWorkDurations, duration)
		}
	}

	if len(deepWorkDurations) == 0 {
		// Default values if no data
		return domain.DurationStats{
			AverageScheduled: 120, // 2 hours
			AverageActual:    150, // 2.5 hours
			Variance:         30.0,
			OverrunRate:      0.0,
		}
	}

	// Calculate average
	total := 0
	for _, d := range deepWorkDurations {
		total += d
	}
	avg := total / len(deepWorkDurations)

	// Calculate variance
	variance := 0.0
	for _, d := range deepWorkDurations {
		diff := float64(d - avg)
		variance += diff * diff
	}
	variance = variance / float64(len(deepWorkDurations))

	return domain.DurationStats{
		AverageScheduled: avg,
		AverageActual:    avg,
		Variance:         variance,
		OverrunRate:      0.0,
	}
}

// calculateBlockDuration calculates duration in minutes between two time strings.
func (f *FocusOptimizer) calculateBlockDuration(startTime, endTime string) int {
	// Parse times (format: "09:00")
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return 120 // Default 2 hours
	}
	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return 120
	}

	duration := end.Sub(start)
	return int(duration.Minutes())
}

// findPeakProductivityBlocks identifies the most productive time blocks.
func (f *FocusOptimizer) findPeakProductivityBlocks(patterns *domain.MeetingPattern) []domain.TimeBlock {
	if len(patterns.Productivity.PeakFocus) > 0 {
		// Sort by score (highest first)
		blocks := make([]domain.TimeBlock, len(patterns.Productivity.PeakFocus))
		copy(blocks, patterns.Productivity.PeakFocus)

		slices.SortFunc(blocks, func(a, b domain.TimeBlock) int {
			if a.Score > b.Score {
				return -1
			}
			if a.Score < b.Score {
				return 1
			}
			return 0
		})

		// Return top 3 peak blocks
		if len(blocks) > 3 {
			return blocks[:3]
		}
		return blocks
	}

	// Default peak productivity times if no data
	return []domain.TimeBlock{
		{DayOfWeek: "Tuesday", StartTime: "10:00", EndTime: "12:00", Score: 90.0},
		{DayOfWeek: "Thursday", StartTime: "10:00", EndTime: "12:00", Score: 90.0},
		{DayOfWeek: "Wednesday", StartTime: "09:00", EndTime: "11:00", Score: 85.0},
	}
}

// findMostProductiveDay finds the day with the highest productivity.
func (f *FocusOptimizer) findMostProductiveDay(patterns *domain.MeetingPattern) string {
	// Find day with lowest meeting density and highest focus time
	minDensity := 999.0
	bestDay := "Wednesday" // Default

	for day, density := range patterns.Productivity.MeetingDensity {
		if density < minDensity {
			minDensity = density
			bestDay = day
		}
	}

	return bestDay
}

// findLeastProductiveDay finds the day with the lowest productivity.
func (f *FocusOptimizer) findLeastProductiveDay(patterns *domain.MeetingPattern) string {
	// Find day with highest meeting density
	maxDensity := 0.0
	worstDay := "Monday" // Default

	for day, density := range patterns.Productivity.MeetingDensity {
		if density > maxDensity {
			maxDensity = density
			worstDay = day
		}
	}

	return worstDay
}

// generateRecommendedBlocks generates AI-recommended focus time blocks.
func (f *FocusOptimizer) generateRecommendedBlocks(patterns *domain.MeetingPattern, settings *domain.FocusTimeSettings) []domain.FocusTimeBlock {
	var blocks []domain.FocusTimeBlock

	// Use productivity patterns to recommend blocks
	for _, peakBlock := range patterns.Productivity.PeakFocus {
		// Check if this day/time should be protected
		if !f.shouldProtectBlock(peakBlock, settings) {
			continue
		}

		duration := f.calculateBlockDuration(peakBlock.StartTime, peakBlock.EndTime)

		// Apply duration constraints
		if duration < settings.MinBlockDuration {
			continue
		}
		if settings.MaxBlockDuration > 0 && duration > settings.MaxBlockDuration {
			duration = settings.MaxBlockDuration
		}

		block := domain.FocusTimeBlock{
			DayOfWeek: peakBlock.DayOfWeek,
			StartTime: peakBlock.StartTime,
			EndTime:   peakBlock.EndTime,
			Duration:  duration,
			Score:     peakBlock.Score,
			Reason:    fmt.Sprintf("Peak productivity time (%.0f%% score)", peakBlock.Score),
			Conflicts: 0, // Will be calculated later
		}

		blocks = append(blocks, block)
	}

	// Sort by score (highest first)
	slices.SortFunc(blocks, func(a, b domain.FocusTimeBlock) int {
		if a.Score > b.Score {
			return -1
		}
		if a.Score < b.Score {
			return 1
		}
		return 0
	})

	// Limit to achieve target hours per week
	targetMinutes := int(settings.TargetHoursPerWeek * 60)
	var selectedBlocks []domain.FocusTimeBlock
	totalMinutes := 0

	for _, block := range blocks {
		if totalMinutes >= targetMinutes {
			break
		}
		selectedBlocks = append(selectedBlocks, block)
		totalMinutes += block.Duration
	}

	return selectedBlocks
}

// shouldProtectBlock checks if a block should be protected based on settings.
func (f *FocusOptimizer) shouldProtectBlock(block domain.TimeBlock, settings *domain.FocusTimeSettings) bool {
	// Check if day is in protected days list
	if len(settings.ProtectedDays) > 0 {
		if !slices.Contains(settings.ProtectedDays, block.DayOfWeek) {
			return false
		}
	}

	// Check if time overlaps with excluded ranges
	for _, excluded := range settings.ExcludedTimeRanges {
		if f.timesOverlap(block.StartTime, block.EndTime, excluded.StartTime, excluded.EndTime) {
			return false
		}
	}

	return true
}

// timesOverlap checks if two time ranges overlap.
func (f *FocusOptimizer) timesOverlap(start1, end1, start2, end2 string) bool {
	// Parse times
	s1, _ := time.Parse("15:04", start1)
	e1, _ := time.Parse("15:04", end1)
	s2, _ := time.Parse("15:04", start2)
	e2, _ := time.Parse("15:04", end2)

	return s1.Before(e2) && s2.Before(e1)
}

// calculateCurrentProtection calculates currently protected focus time hours per week.
func (f *FocusOptimizer) calculateCurrentProtection(ctx context.Context, grantID string) float64 {
	// This would query for existing focus time blocks in the calendar
	// For now, return 0 as placeholder
	return 0.0
}

// generateInsights generates AI insights about focus patterns.
func (f *FocusOptimizer) generateInsights(patterns *domain.MeetingPattern, blocks []domain.FocusTimeBlock, settings *domain.FocusTimeSettings) []string {
	var insights []string

	// Insight about peak productivity
	if len(patterns.Productivity.PeakFocus) > 0 {
		topBlock := patterns.Productivity.PeakFocus[0]
		insights = append(insights, fmt.Sprintf(
			"Your peak productivity is %s at %s--%s (%.0f%% focus score)",
			topBlock.DayOfWeek, topBlock.StartTime, topBlock.EndTime, topBlock.Score,
		))
	}

	// Insight about meeting density
	var highDensityDays []string
	for day, density := range patterns.Productivity.MeetingDensity {
		if density > 5.0 { // More than 5 meetings per day on average
			highDensityDays = append(highDensityDays, day)
		}
	}
	if len(highDensityDays) > 0 {
		sort.Strings(highDensityDays)
		insights = append(insights, fmt.Sprintf(
			"High meeting density on %v - consider protecting more focus time on these days",
			highDensityDays,
		))
	}

	// Insight about recommended blocks
	totalHours := 0.0
	for _, block := range blocks {
		totalHours += float64(block.Duration) / 60.0
	}
	if totalHours > 0 {
		insights = append(insights, fmt.Sprintf(
			"AI recommends %.1f hours/week of protected focus time across %d blocks",
			totalHours, len(blocks),
		))
	}

	// Insight about target gap
	if totalHours < settings.TargetHoursPerWeek {
		gap := settings.TargetHoursPerWeek - totalHours
		insights = append(insights, fmt.Sprintf(
			"Need %.1f more hours/week to reach your target of %.1f hours",
			gap, settings.TargetHoursPerWeek,
		))
	}

	return insights
}

// calculateConfidence calculates confidence in recommendations based on data quality.
func (f *FocusOptimizer) calculateConfidence(patterns *domain.MeetingPattern) float64 {
	confidence := 50.0 // Base confidence

	// Increase confidence if we have peak focus data
	if len(patterns.Productivity.PeakFocus) > 0 {
		confidence += 20.0
	}

	// Increase confidence if we have meeting density data
	if len(patterns.Productivity.MeetingDensity) > 0 {
		confidence += 15.0
	}

	// Increase confidence if we have participant patterns
	if len(patterns.Participants) > 10 {
		confidence += 15.0
	}

	// Cap at 100
	if confidence > 100.0 {
		confidence = 100.0
	}

	return confidence
}

// CreateProtectedBlocks creates protected focus time blocks in the calendar.
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
