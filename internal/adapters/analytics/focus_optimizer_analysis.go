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
