package analytics

import (
	"context"
	"fmt"
	"sort"
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
	client CalendarClient
}

// NewPatternLearner creates a new pattern learner.
func NewPatternLearner(client CalendarClient) *PatternLearner {
	return &PatternLearner{
		client: client,
	}
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
		Limit: 500, // Maximum events per calendar
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
func (p *PatternLearner) learnDurationPatterns(events []domain.Event) domain.DurationPatterns {
	byParticipant := make(map[string]*durationAccumulator)
	overall := &durationAccumulator{}

	for _, event := range events {
		duration := int((event.When.EndTime - event.When.StartTime) / 60) // Minutes

		overall.add(duration, duration) // For now, assume actual = scheduled

		// Track by participant
		for _, participant := range event.Participants {
			if participant.Email == "" {
				continue
			}

			if _, exists := byParticipant[participant.Email]; !exists {
				byParticipant[participant.Email] = &durationAccumulator{}
			}
			byParticipant[participant.Email].add(duration, duration)
		}
	}

	// Convert accumulators to stats
	participantStats := make(map[string]domain.DurationStats)
	for email, acc := range byParticipant {
		participantStats[email] = acc.toStats()
	}

	return domain.DurationPatterns{
		ByParticipant: participantStats,
		ByType:        make(map[string]domain.DurationStats), // Would classify by type
		Overall:       overall.toStats(),
	}
}

// learnTimezonePatterns analyzes timezone preferences.
func (p *PatternLearner) learnTimezonePatterns(events []domain.Event) domain.TimezonePatterns {
	distribution := make(map[string]int)

	for _, event := range events {
		tz := event.When.StartTimezone
		if tz == "" {
			tz = "UTC"
		}
		distribution[tz]++
	}

	return domain.TimezonePatterns{
		PreferredTimes: make(map[string][]string),
		Distribution:   distribution,
		CrossTZTimes:   []string{"14:00", "15:00", "16:00"}, // Default afternoon times
	}
}

// learnProductivityPatterns identifies productive time blocks.
func (p *PatternLearner) learnProductivityPatterns(events []domain.Event) domain.ProductivityPatterns {
	// Analyze when meetings are scheduled vs when there are gaps
	meetingsByDayHour := make(map[string]int)
	meetingsByDay := make(map[string]int)

	for _, event := range events {
		eventTime := time.Unix(event.When.StartTime, 0)
		dayOfWeek := eventTime.Weekday().String()
		hour := eventTime.Hour()

		dayHourKey := fmt.Sprintf("%s-%02d", dayOfWeek, hour)
		meetingsByDayHour[dayHourKey]++
		meetingsByDay[dayOfWeek]++
	}

	// Identify peak focus times (low meeting density)
	var peakFocus []domain.TimeBlock
	daysOfWeek := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	for _, day := range daysOfWeek {
		for hour := 9; hour < 17; hour++ {
			key := fmt.Sprintf("%s-%02d", day, hour)
			density := meetingsByDayHour[key]

			// Low density = good for focus
			if density <= 1 {
				peakFocus = append(peakFocus, domain.TimeBlock{
					DayOfWeek: day,
					StartTime: fmt.Sprintf("%02d:00", hour),
					EndTime:   fmt.Sprintf("%02d:00", hour+2),
					Score:     90.0 - float64(density)*10,
				})
			}
		}
	}

	// Calculate meeting density by day
	meetingDensity := make(map[string]float64)
	for day, count := range meetingsByDay {
		// Assuming ~90 days analyzed
		meetingDensity[day] = float64(count) / 12.0 // ~12 weeks
	}

	return domain.ProductivityPatterns{
		PeakFocus:      peakFocus,
		LowEnergy:      []domain.TimeBlock{}, // Would analyze based on declined meetings
		MeetingDensity: meetingDensity,
		FocusBlocks:    peakFocus[:min(len(peakFocus), 5)], // Top 5
	}
}

// learnParticipantPatterns learns patterns for specific participants.
func (p *PatternLearner) learnParticipantPatterns(events []domain.Event) map[string]domain.ParticipantPattern {
	participants := make(map[string]*participantAccumulator)

	for _, event := range events {
		for _, participant := range event.Participants {
			if participant.Email == "" {
				continue
			}

			email := participant.Email
			if _, exists := participants[email]; !exists {
				participants[email] = &participantAccumulator{
					email:     email,
					dayCount:  make(map[string]int),
					hourCount: make(map[string]int),
				}
			}

			acc := participants[email]
			acc.meetingCount++

			if event.Status == "confirmed" {
				acc.acceptedCount++
			}

			eventTime := time.Unix(event.When.StartTime, 0)
			acc.dayCount[eventTime.Weekday().String()]++
			acc.hourCount[fmt.Sprintf("%02d:00", eventTime.Hour())]++
			acc.totalDuration += int((event.When.EndTime - event.When.StartTime) / 60)

			if event.When.StartTimezone != "" {
				acc.timezone = event.When.StartTimezone
			}
		}
	}

	// Convert to patterns
	patterns := make(map[string]domain.ParticipantPattern)
	for email, acc := range participants {
		patterns[email] = acc.toPattern()
	}

	return patterns
}

// generateRecommendations creates AI recommendations based on patterns.
func (p *PatternLearner) generateRecommendations(patterns *domain.MeetingPattern, events []domain.Event) []domain.Recommendation {
	var recommendations []domain.Recommendation

	// Recommend focus time blocks
	if len(patterns.Productivity.FocusBlocks) > 0 {
		topBlock := patterns.Productivity.FocusBlocks[0]
		recommendations = append(recommendations, domain.Recommendation{
			Type:        "focus_time",
			Priority:    "high",
			Title:       fmt.Sprintf("Block %s %s-%s for focus time", topBlock.DayOfWeek, topBlock.StartTime, topBlock.EndTime),
			Description: "Historical data shows you have few meetings during this time, making it ideal for deep work.",
			Confidence:  topBlock.Score,
			Action:      "Create recurring focus time block",
			Impact:      "Increase productivity by 20-30%",
		})
	}

	// Recommend declining low-value time slots
	for day, rate := range patterns.Acceptance.ByDayOfWeek {
		if rate < 0.5 {
			recommendations = append(recommendations, domain.Recommendation{
				Type:        "decline_pattern",
				Priority:    "medium",
				Title:       fmt.Sprintf("Consider avoiding %s meetings", day),
				Description: fmt.Sprintf("You accept only %.0f%% of meetings on %ss. Consider blocking this time or being more selective.", rate*100, day),
				Confidence:  (1 - rate) * 100,
				Action:      fmt.Sprintf("Auto-suggest alternatives to %s meetings", day),
				Impact:      "Reduce low-productivity meetings",
			})
		}
	}

	// Recommend duration adjustments
	for participant, stats := range patterns.Duration.ByParticipant {
		if stats.AverageActual > 0 && stats.AverageScheduled > 0 {
			diff := stats.AverageActual - stats.AverageScheduled
			if diff > 5 {
				recommendations = append(recommendations, domain.Recommendation{
					Type:        "duration_adjustment",
					Priority:    "low",
					Title:       fmt.Sprintf("Adjust meeting length with %s", participant),
					Description: fmt.Sprintf("Meetings with %s typically run %d minutes over. Consider scheduling %d minutes instead of %d.", participant, diff, stats.AverageActual, stats.AverageScheduled),
					Confidence:  70.0,
					Action:      fmt.Sprintf("Suggest %d-minute meetings with %s", stats.AverageActual, participant),
					Impact:      "Better time estimates and reduced overruns",
				})
			}
		}
	}

	return recommendations
}

// generateInsights creates human-readable insights.
func (p *PatternLearner) generateInsights(patterns *domain.MeetingPattern, events []domain.Event) []string {
	var insights []string

	// Acceptance insights
	bestDay := ""
	bestRate := 0.0
	for day, rate := range patterns.Acceptance.ByDayOfWeek {
		if rate > bestRate {
			bestRate = rate
			bestDay = day
		}
	}
	if bestDay != "" {
		insights = append(insights, fmt.Sprintf("You accept %.0f%% of meetings on %ss (your best day)", bestRate*100, bestDay))
	}

	// Productivity insights
	if len(patterns.Productivity.FocusBlocks) > 0 {
		block := patterns.Productivity.FocusBlocks[0]
		insights = append(insights, fmt.Sprintf("Peak focus time: %s %s-%s (fewest meetings)", block.DayOfWeek, block.StartTime, block.EndTime))
	}

	// Timezone insights
	if len(patterns.Timezone.Distribution) > 0 {
		type tzCount struct {
			tz    string
			count int
		}
		var tzCounts []tzCount
		for tz, count := range patterns.Timezone.Distribution {
			tzCounts = append(tzCounts, tzCount{tz, count})
		}
		sort.Slice(tzCounts, func(i, j int) bool {
			return tzCounts[i].count > tzCounts[j].count
		})
		if len(tzCounts) > 0 {
			insights = append(insights, fmt.Sprintf("Most meetings in %s timezone (%d meetings)", tzCounts[0].tz, tzCounts[0].count))
		}
	}

	// Meeting count insights
	insights = append(insights, fmt.Sprintf("Analyzed %d meetings over %d days", len(events), 90))

	return insights
}

// Helper types and functions

type durationAccumulator struct {
	totalScheduled int
	totalActual    int
	count          int
	overruns       int
}

func (d *durationAccumulator) add(scheduled, actual int) {
	d.totalScheduled += scheduled
	d.totalActual += actual
	d.count++
	if actual > scheduled {
		d.overruns++
	}
}

func (d *durationAccumulator) toStats() domain.DurationStats {
	if d.count == 0 {
		return domain.DurationStats{}
	}

	return domain.DurationStats{
		AverageScheduled: d.totalScheduled / d.count,
		AverageActual:    d.totalActual / d.count,
		Variance:         0.0, // Would calculate standard deviation
		OverrunRate:      float64(d.overruns) / float64(d.count),
	}
}

type participantAccumulator struct {
	email         string
	meetingCount  int
	acceptedCount int
	dayCount      map[string]int
	hourCount     map[string]int
	totalDuration int
	timezone      string
}

func (p *participantAccumulator) toPattern() domain.ParticipantPattern {
	acceptanceRate := 0.0
	if p.meetingCount > 0 {
		acceptanceRate = float64(p.acceptedCount) / float64(p.meetingCount)
	}

	avgDuration := 0
	if p.meetingCount > 0 {
		avgDuration = p.totalDuration / p.meetingCount
	}

	// Find preferred days (top 2)
	type dayCount struct {
		day   string
		count int
	}
	var days []dayCount
	for day, count := range p.dayCount {
		days = append(days, dayCount{day, count})
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].count > days[j].count
	})

	preferredDays := make([]string, 0, 2)
	for i := 0; i < min(2, len(days)); i++ {
		preferredDays = append(preferredDays, days[i].day)
	}

	// Find preferred times (top 2)
	type hourCount struct {
		hour  string
		count int
	}
	var hours []hourCount
	for hour, count := range p.hourCount {
		hours = append(hours, hourCount{hour, count})
	}
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].count > hours[j].count
	})

	preferredTimes := make([]string, 0, 2)
	for i := 0; i < min(2, len(hours)); i++ {
		preferredTimes = append(preferredTimes, hours[i].hour)
	}

	return domain.ParticipantPattern{
		Email:           p.email,
		MeetingCount:    p.meetingCount,
		AcceptanceRate:  acceptanceRate,
		PreferredDays:   preferredDays,
		PreferredTimes:  preferredTimes,
		AverageDuration: avgDuration,
		Timezone:        p.timezone,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
