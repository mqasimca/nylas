package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// CalendarClient defines the minimal calendar interface needed for pattern learning.
type CalendarClient interface {
	GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error)
	GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error)
}

// PatternLearner analyzes calendar history to learn meeting patterns.
type PatternLearner struct {
	client       CalendarClient
	workingHours *domain.DaySchedule
}

// NewPatternLearner creates a new pattern learner.
func NewPatternLearner(client CalendarClient) *PatternLearner {
	return &PatternLearner{
		client: client,
	}
}

// NewPatternLearnerWithWorkingHours creates a pattern learner with working hours config.
func NewPatternLearnerWithWorkingHours(client CalendarClient, workingHours *domain.DaySchedule) *PatternLearner {
	return &PatternLearner{
		client:       client,
		workingHours: workingHours,
	}
}

// getWorkingHoursRange returns the start and end hours based on config or defaults.
func (p *PatternLearner) getWorkingHoursRange() (startHour, endHour int) {
	// Default working hours: 9-17
	startHour = 9
	endHour = 17

	if p.workingHours == nil || !p.workingHours.Enabled {
		return startHour, endHour
	}

	// Parse start time (format: "HH:MM")
	if p.workingHours.Start != "" {
		var h, m int
		if _, err := fmt.Sscanf(p.workingHours.Start, "%d:%d", &h, &m); err == nil {
			startHour = h
			// If minutes > 0, round up to next hour for block analysis
			if m > 0 {
				startHour = h + 1
			}
		}
	}

	// Parse end time (format: "HH:MM")
	if p.workingHours.End != "" {
		var h, m int
		if _, err := fmt.Sscanf(p.workingHours.End, "%d:%d", &h, &m); err == nil {
			endHour = h
		}
	}

	return startHour, endHour
}

// AnalyzeHistory analyzes meeting history to learn patterns.
func (p *PatternLearner) AnalyzeHistory(ctx context.Context, grantID string, days int) (*domain.MeetingAnalysis, error) {
	// Calculate date range
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	// Fetch events from the period
	events, err := p.fetchEvents(ctx, grantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		return &domain.MeetingAnalysis{
			Period:        domain.DateRange{Start: start, End: end},
			TotalMeetings: 0,
			Insights:      []string{"No meetings found in the analyzed period."},
		}, nil
	}

	// Learn patterns from events
	patterns := p.learnPatterns(events, start, end)

	// Generate recommendations
	recommendations := p.generateRecommendations(patterns, events)

	// Generate insights
	insights := p.generateInsights(patterns, events)

	analysis := &domain.MeetingAnalysis{
		Period:          domain.DateRange{Start: start, End: end},
		TotalMeetings:   len(events),
		Patterns:        patterns,
		Recommendations: recommendations,
		Insights:        insights,
	}

	return analysis, nil
}

// fetchEvents retrieves events for the specified period.
func (p *PatternLearner) fetchEvents(ctx context.Context, grantID string, start, end time.Time) ([]domain.Event, error) {
	// Get all calendars for the grant
	calendars, err := p.client.GetCalendars(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendars: %w", err)
	}

	// Aggregate events from all calendars
	var allEvents []domain.Event

	// Create query params for the date range
	params := &domain.EventQueryParams{
		Start: start.Unix(),
		End:   end.Unix(),
		Limit: 200, // Maximum allowed by Nylas API v3
	}

	// Fetch events from each calendar
	for _, calendar := range calendars {
		events, err := p.client.GetEvents(ctx, grantID, calendar.ID, params)
		if err != nil {
			// Log error but continue with other calendars
			continue
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// learnPatterns analyzes events to detect patterns.
func (p *PatternLearner) learnPatterns(events []domain.Event, start, end time.Time) *domain.MeetingPattern {
	pattern := &domain.MeetingPattern{
		AnalyzedPeriod: domain.DateRange{Start: start, End: end},
		LastUpdated:    time.Now(),
		Participants:   make(map[string]domain.ParticipantPattern),
	}

	// Learn acceptance patterns
	pattern.Acceptance = p.learnAcceptancePatterns(events)

	// Learn duration patterns
	pattern.Duration = p.learnDurationPatterns(events)

	// Learn timezone patterns
	pattern.Timezone = p.learnTimezonePatterns(events)

	// Learn productivity patterns
	pattern.Productivity = p.learnProductivityPatterns(events)

	// Learn participant-specific patterns
	pattern.Participants = p.learnParticipantPatterns(events)

	return pattern
}

// learnAcceptancePatterns analyzes when meetings are accepted.
func (p *PatternLearner) learnAcceptancePatterns(events []domain.Event) domain.AcceptancePatterns {
	byDayOfWeek := make(map[string]int)
	byTimeOfDay := make(map[string]int)
	byDayAndTime := make(map[string]int)
	totalByDay := make(map[string]int)
	totalByTime := make(map[string]int)
	totalByDayTime := make(map[string]int)

	acceptedCount := 0
	totalCount := len(events)

	for _, event := range events {
		// Count as accepted if status is confirmed
		isAccepted := event.Status == "confirmed"

		if isAccepted {
			acceptedCount++
		}

		// Extract time info
		eventTime := time.Unix(event.When.StartTime, 0)
		dayOfWeek := eventTime.Weekday().String()
		hourOfDay := fmt.Sprintf("%02d:00", eventTime.Hour())
		dayTimeKey := fmt.Sprintf("%s-%s", dayOfWeek, hourOfDay)

		totalByDay[dayOfWeek]++
		totalByTime[hourOfDay]++
		totalByDayTime[dayTimeKey]++

		if isAccepted {
			byDayOfWeek[dayOfWeek]++
			byTimeOfDay[hourOfDay]++
			byDayAndTime[dayTimeKey]++
		}
	}

	// Calculate rates
	dayRates := make(map[string]float64)
	for day, count := range totalByDay {
		dayRates[day] = float64(byDayOfWeek[day]) / float64(count)
	}

	timeRates := make(map[string]float64)
	for hour, count := range totalByTime {
		timeRates[hour] = float64(byTimeOfDay[hour]) / float64(count)
	}

	dayTimeRates := make(map[string]float64)
	for key, count := range totalByDayTime {
		dayTimeRates[key] = float64(byDayAndTime[key]) / float64(count)
	}

	overallRate := float64(acceptedCount) / float64(totalCount)

	return domain.AcceptancePatterns{
		ByDayOfWeek:  dayRates,
		ByTimeOfDay:  timeRates,
		ByDayAndTime: dayTimeRates,
		Overall:      overallRate,
	}
}

// learnDurationPatterns analyzes meeting durations.
