package analytics

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

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

	// Calculate average meeting density per hour to use as baseline
	totalMeetings := 0
	totalHours := 0
	for _, count := range meetingsByDayHour {
		totalMeetings += count
		totalHours++
	}
	avgDensity := 0.0
	if totalHours > 0 {
		avgDensity = float64(totalMeetings) / float64(totalHours)
	}

	// Get working hours from config
	startHour, endHour := p.getWorkingHoursRange()

	// Identify peak focus times (below-average meeting density)
	var peakFocus []domain.TimeBlock
	daysOfWeek := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	for _, day := range daysOfWeek {
		for hour := startHour; hour < endHour; hour++ {
			key := fmt.Sprintf("%s-%02d", day, hour)
			density := meetingsByDayHour[key]

			// Calculate score based on how much better this slot is than average
			// Lower density = higher score
			// Score formula: 100 - (density/avgDensity * 50) capped at 0-100
			score := 100.0
			if avgDensity > 0 {
				densityRatio := float64(density) / avgDensity
				score = 100.0 - (densityRatio * 50.0)
				// Cap between 0-100
				if score < 0 {
					score = 0
				} else if score > 100 {
					score = 100
				}
			}

			// Include blocks with score >= 50 (better than or equal to average)
			if score >= 50 {
				peakFocus = append(peakFocus, domain.TimeBlock{
					DayOfWeek: day,
					StartTime: fmt.Sprintf("%02d:00", hour),
					EndTime:   fmt.Sprintf("%02d:00", hour+2),
					Score:     score,
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

	// Select the BEST focus block for each day (1 per day)
	// Group blocks by day and pick the highest-scoring one
	bestByDay := make(map[string]domain.TimeBlock)
	for _, block := range peakFocus {
		existing, exists := bestByDay[block.DayOfWeek]
		if !exists || block.Score > existing.Score {
			bestByDay[block.DayOfWeek] = block
		}
	}

	// Convert map to slice and sort by day order (Mon-Fri)
	var topFocusBlocks []domain.TimeBlock
	dayOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	for _, day := range dayOrder {
		if block, exists := bestByDay[day]; exists {
			topFocusBlocks = append(topFocusBlocks, block)
		}
	}

	return domain.ProductivityPatterns{
		PeakFocus:      peakFocus,
		LowEnergy:      []domain.TimeBlock{}, // Would analyze based on declined meetings
		MeetingDensity: meetingDensity,
		FocusBlocks:    topFocusBlocks, // 1 best block per day (max 5 for Mon-Fri)
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
	// Generate recommendations for all high-scoring focus blocks (score >= 70)
	for _, block := range patterns.Productivity.FocusBlocks {
		if block.Score >= 70 {
			// Determine priority based on score
			priority := "medium"
			if block.Score >= 85 {
				priority = "high"
			}

			recommendations = append(recommendations, domain.Recommendation{
				Type:        "focus_time",
				Priority:    priority,
				Title:       fmt.Sprintf("Block %s %s-%s for focus time", block.DayOfWeek, block.StartTime, block.EndTime),
				Description: fmt.Sprintf("Historical data shows you have few meetings during this time (score: %.0f/100), making it ideal for deep work.", block.Score),
				Confidence:  block.Score,
				Action:      "Create recurring focus time block",
				Impact:      "Increase productivity by 20-30%",
			})
		}
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
		slices.SortFunc(tzCounts, func(a, b tzCount) int {
			return cmp.Compare(b.count, a.count) // Descending order
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
	slices.SortFunc(days, func(a, b dayCount) int {
		return cmp.Compare(b.count, a.count) // Descending order
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
	slices.SortFunc(hours, func(a, b hourCount) int {
		return cmp.Compare(b.count, a.count) // Descending order
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
