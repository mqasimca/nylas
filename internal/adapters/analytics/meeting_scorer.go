package analytics

import (
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// MeetingScorer scores meetings based on learned patterns.
type MeetingScorer struct {
	patterns *domain.MeetingPattern
}

// NewMeetingScorer creates a new meeting scorer.
func NewMeetingScorer(patterns *domain.MeetingPattern) *MeetingScorer {
	return &MeetingScorer{
		patterns: patterns,
	}
}

// ScoreMeetingTime scores a proposed meeting time based on historical patterns.
func (s *MeetingScorer) ScoreMeetingTime(proposedTime time.Time, participants []string, duration int) *domain.MeetingScore {
	if s.patterns == nil {
		return &domain.MeetingScore{
			Score:          50,
			Confidence:     0,
			Recommendation: "No historical data available for scoring",
		}
	}

	var factors []domain.ScoreFactor
	totalScore := 0
	maxScore := 0

	// Factor 1: Acceptance rate for day of week (weight: 25)
	dayOfWeek := proposedTime.Weekday().String()
	if rate, exists := s.patterns.Acceptance.ByDayOfWeek[dayOfWeek]; exists {
		impact := int(rate * 25)
		totalScore += impact
		maxScore += 25

		factors = append(factors, domain.ScoreFactor{
			Name:        "Day Preference",
			Impact:      impact - 13, // Relative to average (12.5)
			Description: fmt.Sprintf("%.0f%% acceptance rate on %ss", rate*100, dayOfWeek),
		})
	}

	// Factor 2: Acceptance rate for time of day (weight: 25)
	hourOfDay := fmt.Sprintf("%02d:00", proposedTime.Hour())
	if rate, exists := s.patterns.Acceptance.ByTimeOfDay[hourOfDay]; exists {
		impact := int(rate * 25)
		totalScore += impact
		maxScore += 25

		factors = append(factors, domain.ScoreFactor{
			Name:        "Time Preference",
			Impact:      impact - 13,
			Description: fmt.Sprintf("%.0f%% acceptance rate at %s", rate*100, hourOfDay),
		})
	}

	// Factor 3: Productivity score (weight: 20)
	productivityScore := s.getProductivityScore(proposedTime)
	totalScore += productivityScore
	maxScore += 20

	factors = append(factors, domain.ScoreFactor{
		Name:        "Productivity",
		Impact:      productivityScore - 10,
		Description: s.getProductivityDescription(proposedTime),
	})

	// Factor 4: Participant compatibility (weight: 15)
	participantScore := s.getParticipantScore(participants, proposedTime)
	totalScore += participantScore
	maxScore += 15

	if participantScore > 0 {
		factors = append(factors, domain.ScoreFactor{
			Name:        "Participant Match",
			Impact:      participantScore - 8,
			Description: "Based on historical meetings with these participants",
		})
	}

	// Factor 5: Timezone fairness (weight: 15)
	timezoneScore := 15 // Default: assume fair
	totalScore += timezoneScore
	maxScore += 15

	factors = append(factors, domain.ScoreFactor{
		Name:        "Timezone",
		Impact:      0,
		Description: "Time works well for all timezones",
	})

	// Calculate final score (0-100)
	finalScore := 0
	if maxScore > 0 {
		finalScore = (totalScore * 100) / maxScore
	}

	// Calculate confidence based on data availability
	confidence := s.calculateConfidence()

	// Calculate success rate
	successRate := s.calculateSuccessRate(proposedTime)

	// Generate recommendation
	recommendation := s.generateRecommendation(finalScore, factors)

	// Suggest alternative times if score is low
	var alternativeTimes []time.Time
	if finalScore < 70 {
		alternativeTimes = s.suggestAlternatives(proposedTime, participants)
	}

	return &domain.MeetingScore{
		Score:            finalScore,
		Confidence:       confidence,
		SuccessRate:      successRate,
		Factors:          factors,
		Recommendation:   recommendation,
		AlternativeTimes: alternativeTimes,
	}
}

// getProductivityScore calculates productivity score for a time slot.
func (s *MeetingScorer) getProductivityScore(t time.Time) int {
	dayOfWeek := t.Weekday().String()
	hour := t.Hour()

	// Check if it's a peak focus time
	for _, block := range s.patterns.Productivity.PeakFocus {
		if block.DayOfWeek == dayOfWeek {
			blockHour := parseHour(block.StartTime)
			if hour == blockHour {
				return int(block.Score / 5) // Scale to 0-20
			}
		}
	}

	// Check meeting density (less density = better productivity)
	if density, exists := s.patterns.Productivity.MeetingDensity[dayOfWeek]; exists {
		// Lower density = higher score
		if density < 2 {
			return 18
		} else if density < 4 {
			return 12
		} else {
			return 6
		}
	}

	return 10 // Default neutral score
}

// getProductivityDescription generates a description for productivity score.
func (s *MeetingScorer) getProductivityDescription(t time.Time) string {
	dayOfWeek := t.Weekday().String()

	for _, block := range s.patterns.Productivity.PeakFocus {
		if block.DayOfWeek == dayOfWeek {
			return "Peak focus time - fewer meetings scheduled"
		}
	}

	if density, exists := s.patterns.Productivity.MeetingDensity[dayOfWeek]; exists {
		return fmt.Sprintf("Average %.1f meetings on %ss", density, dayOfWeek)
	}

	return "Standard productivity time"
}

// getParticipantScore calculates compatibility score with participants.
func (s *MeetingScorer) getParticipantScore(participants []string, t time.Time) int {
	if len(participants) == 0 {
		return 8 // Neutral score if no participants
	}

	totalScore := 0
	participantCount := 0

	dayOfWeek := t.Weekday().String()
	hourOfDay := fmt.Sprintf("%02d:00", t.Hour())

	for _, email := range participants {
		if pattern, exists := s.patterns.Participants[email]; exists {
			participantCount++

			// Check if day matches preferences
			for _, prefDay := range pattern.PreferredDays {
				if prefDay == dayOfWeek {
					totalScore += 8
					break
				}
			}

			// Check if time matches preferences
			for _, prefTime := range pattern.PreferredTimes {
				if prefTime == hourOfDay {
					totalScore += 7
					break
				}
			}
		}
	}

	if participantCount == 0 {
		return 8 // No data available
	}

	return totalScore / participantCount
}

// calculateConfidence calculates confidence in the score based on data availability.
func (s *MeetingScorer) calculateConfidence() float64 {
	dataPoints := 0
	maxDataPoints := 5

	if len(s.patterns.Acceptance.ByDayOfWeek) > 0 {
		dataPoints++
	}
	if len(s.patterns.Acceptance.ByTimeOfDay) > 0 {
		dataPoints++
	}
	if len(s.patterns.Productivity.PeakFocus) > 0 {
		dataPoints++
	}
	if len(s.patterns.Participants) > 0 {
		dataPoints++
	}
	if len(s.patterns.Duration.ByParticipant) > 0 {
		dataPoints++
	}

	return (float64(dataPoints) / float64(maxDataPoints)) * 100
}

// calculateSuccessRate estimates historical success rate for similar meetings.
func (s *MeetingScorer) calculateSuccessRate(t time.Time) float64 {
	dayOfWeek := t.Weekday().String()

	// Use day-based acceptance rate as proxy for success rate
	if rate, exists := s.patterns.Acceptance.ByDayOfWeek[dayOfWeek]; exists {
		return rate
	}

	return s.patterns.Acceptance.Overall
}

// generateRecommendation generates a text recommendation based on score.
func (s *MeetingScorer) generateRecommendation(score int, factors []domain.ScoreFactor) string {
	if score >= 85 {
		return "Excellent time - highly recommended based on historical patterns"
	} else if score >= 70 {
		return "Good time - aligns well with your preferences"
	} else if score >= 50 {
		return "Acceptable time - consider alternatives if available"
	} else {
		// Find the biggest negative factor
		worstFactor := ""
		worstImpact := 0
		for _, factor := range factors {
			if factor.Impact < worstImpact {
				worstImpact = factor.Impact
				worstFactor = factor.Name
			}
		}

		if worstFactor != "" {
			return fmt.Sprintf("Not recommended - %s is suboptimal. Consider alternative times.", worstFactor)
		}
		return "Not recommended - consider alternative times"
	}
}

// suggestAlternatives suggests better meeting times.
func (s *MeetingScorer) suggestAlternatives(original time.Time, participants []string) []time.Time {
	var alternatives []time.Time

	// Find best day based on acceptance rate
	bestDay := ""
	bestRate := 0.0
	for day, rate := range s.patterns.Acceptance.ByDayOfWeek {
		if rate > bestRate {
			bestRate = rate
			bestDay = day
		}
	}

	// Find best time based on acceptance rate
	bestHour := ""
	bestHourRate := 0.0
	for hour, rate := range s.patterns.Acceptance.ByTimeOfDay {
		hourNum := parseHour(hour)
		// Only consider working hours
		if hourNum >= 9 && hourNum <= 17 && rate > bestHourRate {
			bestHourRate = rate
			bestHour = hour
		}
	}

	if bestDay != "" && bestHour != "" {
		// Create alternative time on best day at best hour
		targetDay := dayNameToWeekday(bestDay)
		currentDay := original.Weekday()
		daysUntilTarget := (int(targetDay) - int(currentDay) + 7) % 7
		if daysUntilTarget == 0 {
			daysUntilTarget = 7 // Next week
		}

		alternativeDate := original.AddDate(0, 0, daysUntilTarget)
		alternativeHour := parseHour(bestHour)

		alternative := time.Date(
			alternativeDate.Year(),
			alternativeDate.Month(),
			alternativeDate.Day(),
			alternativeHour,
			0,
			0,
			0,
			original.Location(),
		)

		alternatives = append(alternatives, alternative)
	}

	return alternatives
}

// Helper functions

func parseHour(timeStr string) int {
	// Parse "HH:MM" format
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	var hour int
	_, _ = fmt.Sscanf(parts[0], "%d", &hour) // Parse hour, default 0 on error
	return hour
}

func dayNameToWeekday(name string) time.Weekday {
	switch strings.ToLower(name) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Monday
	}
}
