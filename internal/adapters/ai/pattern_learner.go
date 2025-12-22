package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// PatternLearner learns from calendar history to predict scheduling patterns.
type PatternLearner struct {
	nylasClient ports.NylasClient
	llmRouter   ports.LLMRouter
}

// NewPatternLearner creates a new pattern learner.
func NewPatternLearner(nylasClient ports.NylasClient, llmRouter ports.LLMRouter) *PatternLearner {
	return &PatternLearner{
		nylasClient: nylasClient,
		llmRouter:   llmRouter,
	}
}

// SchedulingPatterns represents discovered patterns from calendar history.
type SchedulingPatterns struct {
	UserID               string                `json:"user_id"`
	AnalysisPeriod       AnalysisPeriod        `json:"analysis_period"`
	AcceptancePatterns   []AcceptancePattern   `json:"acceptance_patterns"`
	DurationPatterns     []DurationPattern     `json:"duration_patterns"`
	TimezonePatterns     []TimezonePattern     `json:"timezone_patterns"`
	ProductivityInsights []ProductivityInsight `json:"productivity_insights"`
	Recommendations      []string              `json:"recommendations"`
	TotalEventsAnalyzed  int                   `json:"total_events_analyzed"`
	GeneratedAt          time.Time             `json:"generated_at"`
}

// AnalysisPeriod defines the time period analyzed.
type AnalysisPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Days      int       `json:"days"`
}

// AcceptancePattern represents meeting acceptance rates by time/day.
type AcceptancePattern struct {
	TimeSlot    string  `json:"time_slot"`   // e.g., "Monday 9-11 AM"
	AcceptRate  float64 `json:"accept_rate"` // 0-1
	EventCount  int     `json:"event_count"` // Number of events in this slot
	Description string  `json:"description"` // Human-readable explanation
	Confidence  float64 `json:"confidence"`  // 0-1, based on sample size
}

// DurationPattern represents typical meeting duration patterns.
type DurationPattern struct {
	MeetingType       string `json:"meeting_type"`       // e.g., "1-on-1", "Team standup"
	ScheduledDuration int    `json:"scheduled_duration"` // In minutes
	ActualDuration    int    `json:"actual_duration"`    // In minutes
	Variance          int    `json:"variance"`           // Difference
	EventCount        int    `json:"event_count"`        // Sample size
	Description       string `json:"description"`        // Pattern description
}

// TimezonePattern represents timezone preferences.
type TimezonePattern struct {
	Timezone      string  `json:"timezone"`       // e.g., "America/New_York"
	EventCount    int     `json:"event_count"`    // Number of events
	Percentage    float64 `json:"percentage"`     // % of total events
	PreferredTime string  `json:"preferred_time"` // e.g., "2-4 PM PST"
	Description   string  `json:"description"`    // Pattern description
}

// ProductivityInsight represents productivity patterns.
type ProductivityInsight struct {
	InsightType string   `json:"insight_type"` // e.g., "peak_focus", "low_energy"
	TimeSlot    string   `json:"time_slot"`    // e.g., "Tuesday 10 AM - 12 PM"
	Score       int      `json:"score"`        // 0-100
	Description string   `json:"description"`  // Explanation
	BasedOn     []string `json:"based_on"`     // What data this is based on
}

// LearnPatternsRequest represents a request to learn patterns.
type LearnPatternsRequest struct {
	GrantID          string  `json:"grant_id"`
	LookbackDays     int     `json:"lookback_days"`     // How far back to analyze
	MinConfidence    float64 `json:"min_confidence"`    // Minimum confidence threshold
	IncludeRecurring bool    `json:"include_recurring"` // Include recurring events
}

// LearnPatterns analyzes calendar history and learns scheduling patterns.
func (p *PatternLearner) LearnPatterns(ctx context.Context, req *LearnPatternsRequest) (*SchedulingPatterns, error) {
	// 1. Fetch historical events
	events, err := p.fetchHistoricalEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch historical events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found in the specified period")
	}

	// 2. Calculate analysis period
	analysisPeriod := p.calculateAnalysisPeriod(events, req.LookbackDays)

	// 3. Analyze acceptance patterns
	acceptancePatterns := p.analyzeAcceptancePatterns(events)

	// 4. Analyze duration patterns
	durationPatterns := p.analyzeDurationPatterns(events)

	// 5. Analyze timezone patterns
	timezonePatterns := p.analyzeTimezonePatterns(events)

	// 6. Analyze productivity patterns
	productivityInsights := p.analyzeProductivityPatterns(events)

	// 7. Use LLM to generate recommendations
	recommendations, err := p.generateRecommendations(ctx, events, acceptancePatterns, durationPatterns, timezonePatterns, productivityInsights)
	if err != nil {
		// Non-fatal: continue without LLM recommendations
		recommendations = []string{"Unable to generate AI recommendations"}
	}

	patterns := &SchedulingPatterns{
		UserID:               req.GrantID,
		AnalysisPeriod:       analysisPeriod,
		AcceptancePatterns:   acceptancePatterns,
		DurationPatterns:     durationPatterns,
		TimezonePatterns:     timezonePatterns,
		ProductivityInsights: productivityInsights,
		Recommendations:      recommendations,
		TotalEventsAnalyzed:  len(events),
		GeneratedAt:          time.Now(),
	}

	return patterns, nil
}

// fetchHistoricalEvents fetches calendar events for pattern analysis.
func (p *PatternLearner) fetchHistoricalEvents(ctx context.Context, req *LearnPatternsRequest) ([]domain.Event, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -req.LookbackDays)

	// First get list of calendars to fetch events from all
	calendars, err := p.nylasClient.GetCalendars(ctx, req.GrantID)
	if err != nil {
		return nil, fmt.Errorf("fetch calendars: %w", err)
	}

	allEvents := []domain.Event{}

	// Fetch events from each calendar
	for _, calendar := range calendars {
		events, err := p.nylasClient.GetEvents(ctx, req.GrantID, calendar.ID, &domain.EventQueryParams{
			Start: startDate.Unix(),
			End:   now.Unix(),
			Limit: 200, // Maximum allowed by Nylas API v3
		})

		if err != nil {
			// Skip calendar if error (might be read-only, etc.)
			continue
		}

		allEvents = append(allEvents, events...)
	}

	// Filter out recurring events if not requested
	if !req.IncludeRecurring {
		filtered := []domain.Event{}
		for _, event := range allEvents {
			// Check if event is recurring (has recurrence or is part of series)
			if len(event.Recurrence) == 0 && event.MasterEventID == "" {
				filtered = append(filtered, event)
			}
		}
		return filtered, nil
	}

	return allEvents, nil
}

// calculateAnalysisPeriod calculates the actual period analyzed.
func (p *PatternLearner) calculateAnalysisPeriod(events []domain.Event, requestedDays int) AnalysisPeriod {
	if len(events) == 0 {
		return AnalysisPeriod{}
	}

	earliest := events[0].When.StartTime
	latest := events[0].When.EndTime

	for _, event := range events {
		if event.When.StartTime < earliest {
			earliest = event.When.StartTime
		}
		if event.When.EndTime > latest {
			latest = event.When.EndTime
		}
	}

	// Convert Unix timestamps to time.Time
	earliestTime := time.Unix(earliest, 0)
	latestTime := time.Unix(latest, 0)

	days := int(latestTime.Sub(earliestTime).Hours() / 24)

	return AnalysisPeriod{
		StartDate: earliestTime,
		EndDate:   latestTime,
		Days:      days,
	}
}

// analyzeAcceptancePatterns analyzes meeting acceptance rates by time slots.
func (p *PatternLearner) analyzeAcceptancePatterns(events []domain.Event) []AcceptancePattern {
	// Group events by day and time slot
	slotCounts := make(map[string]int)
	slotTotal := make(map[string]int)

	for _, event := range events {
		// Convert Unix timestamp to time.Time
		startTime := time.Unix(event.When.StartTime, 0)
		day := startTime.Weekday().String()
		hour := startTime.Hour()

		// Categorize into time blocks
		var timeBlock string
		if hour >= 9 && hour < 11 {
			timeBlock = "9-11 AM"
		} else if hour >= 11 && hour < 13 {
			timeBlock = "11 AM-1 PM"
		} else if hour >= 13 && hour < 15 {
			timeBlock = "1-3 PM"
		} else if hour >= 15 && hour < 17 {
			timeBlock = "3-5 PM"
		} else {
			timeBlock = "Outside hours"
		}

		slot := fmt.Sprintf("%s %s", day, timeBlock)
		slotTotal[slot]++

		// Consider event "accepted" if status is confirmed or busy is true
		if event.Status == "confirmed" || event.Busy {
			slotCounts[slot]++
		}
	}

	// Calculate acceptance rates
	patterns := []AcceptancePattern{}
	for slot, total := range slotTotal {
		if total < 3 {
			// Skip slots with too few samples
			continue
		}

		accepted := slotCounts[slot]
		acceptRate := float64(accepted) / float64(total)

		// Confidence based on sample size (higher samples = higher confidence)
		confidence := float64(total) / 20.0
		if confidence > 1.0 {
			confidence = 1.0
		}

		description := ""
		if acceptRate > 0.8 {
			description = "You prefer meetings during this time"
		} else if acceptRate < 0.4 {
			description = "You tend to avoid meetings during this time"
		} else {
			description = "Moderate acceptance rate"
		}

		patterns = append(patterns, AcceptancePattern{
			TimeSlot:    slot,
			AcceptRate:  acceptRate,
			EventCount:  total,
			Description: description,
			Confidence:  confidence,
		})
	}

	// Sort by accept rate (highest first)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].AcceptRate > patterns[j].AcceptRate
	})

	return patterns
}

// analyzeDurationPatterns analyzes meeting duration patterns.
func (p *PatternLearner) analyzeDurationPatterns(events []domain.Event) []DurationPattern {
	// Group events by type (inferred from title patterns)
	typeMap := make(map[string][]domain.Event)

	for _, event := range events {
		meetingType := p.inferMeetingType(event.Title)
		typeMap[meetingType] = append(typeMap[meetingType], event)
	}

	patterns := []DurationPattern{}

	for meetingType, typeEvents := range typeMap {
		if len(typeEvents) < 3 {
			// Skip types with too few samples
			continue
		}

		// Calculate average scheduled duration
		var totalScheduled, totalActual int
		for _, event := range typeEvents {
			// Calculate duration from Unix timestamps (in seconds)
			durationSec := event.When.EndTime - event.When.StartTime
			scheduledDuration := int(durationSec / 60) // Convert to minutes
			totalScheduled += scheduledDuration

			// Actual duration is same as scheduled (we don't have end-time tracking)
			totalActual += scheduledDuration
		}

		avgScheduled := totalScheduled / len(typeEvents)
		avgActual := totalActual / len(typeEvents)

		patterns = append(patterns, DurationPattern{
			MeetingType:       meetingType,
			ScheduledDuration: avgScheduled,
			ActualDuration:    avgActual,
			Variance:          avgActual - avgScheduled,
			EventCount:        len(typeEvents),
			Description:       fmt.Sprintf("Average %d-minute %s meetings", avgScheduled, meetingType),
		})
	}

	return patterns
}

// inferMeetingType infers meeting type from title.
func (p *PatternLearner) inferMeetingType(title string) string {
	titleLower := strings.ToLower(title)

	if containsAny(titleLower, []string{"1:1", "1-on-1", "one-on-one"}) {
		return "1-on-1"
	}
	if containsAny(titleLower, []string{"standup", "daily", "scrum"}) {
		return "Standup"
	}
	if containsAny(titleLower, []string{"review", "retrospective", "retro"}) {
		return "Review"
	}
	if containsAny(titleLower, []string{"planning", "plan"}) {
		return "Planning"
	}
	if containsAny(titleLower, []string{"interview", "candidate"}) {
		return "Interview"
	}
	if containsAny(titleLower, []string{"client", "customer"}) {
		return "Client call"
	}

	return "General meeting"
}

// containsAny checks if string contains any of the substrings.
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			// Simple substring check
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// analyzeTimezonePatterns analyzes timezone preferences.
func (p *PatternLearner) analyzeTimezonePatterns(events []domain.Event) []TimezonePattern {
	tzCounts := make(map[string]int)
	totalEvents := len(events)

	for _, event := range events {
		tz := event.When.StartTimezone
		if tz == "" {
			tz = "UTC"
		}
		tzCounts[tz]++
	}

	patterns := []TimezonePattern{}
	for tz, count := range tzCounts {
		percentage := float64(count) / float64(totalEvents)

		description := fmt.Sprintf("%d%% of meetings in this timezone", int(percentage*100))

		patterns = append(patterns, TimezonePattern{
			Timezone:      tz,
			EventCount:    count,
			Percentage:    percentage,
			PreferredTime: "Varies", // Would need more analysis
			Description:   description,
		})
	}

	// Sort by event count (most common first)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].EventCount > patterns[j].EventCount
	})

	return patterns
}

// analyzeProductivityPatterns analyzes productivity patterns.
func (p *PatternLearner) analyzeProductivityPatterns(events []domain.Event) []ProductivityInsight {
	// Analyze meeting density by day
	dayDensity := make(map[string]int)
	for _, event := range events {
		startTime := time.Unix(event.When.StartTime, 0)
		day := startTime.Weekday().String()
		dayDensity[day]++
	}

	insights := []ProductivityInsight{}

	// Find peak and low days
	maxDay := ""
	maxCount := 0
	minDay := ""
	minCount := len(events) + 1

	for day, count := range dayDensity {
		if count > maxCount {
			maxCount = count
			maxDay = day
		}
		if count < minCount {
			minCount = count
			minDay = day
		}
	}

	if maxDay != "" {
		insights = append(insights, ProductivityInsight{
			InsightType: "high_meeting_density",
			TimeSlot:    maxDay,
			Score:       30, // Lower score for high meeting days
			Description: fmt.Sprintf("%s has the most meetings (%d) - may impact focus time", maxDay, maxCount),
			BasedOn:     []string{"Meeting count by day"},
		})
	}

	if minDay != "" {
		insights = append(insights, ProductivityInsight{
			InsightType: "low_meeting_density",
			TimeSlot:    minDay,
			Score:       90, // Higher score for low meeting days
			Description: fmt.Sprintf("%s has the fewest meetings (%d) - good for deep work", minDay, minCount),
			BasedOn:     []string{"Meeting count by day"},
		})
	}

	return insights
}

// generateRecommendations uses LLM to generate actionable recommendations.
func (p *PatternLearner) generateRecommendations(ctx context.Context, events []domain.Event, acceptance []AcceptancePattern, duration []DurationPattern, timezone []TimezonePattern, productivity []ProductivityInsight) ([]string, error) {
	// Build context for LLM
	context := p.buildPatternContext(events, acceptance, duration, timezone, productivity)

	// Create chat request
	chatReq := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert productivity coach analyzing calendar patterns. Provide 3-5 actionable recommendations to improve scheduling and productivity.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Based on the following calendar analysis, provide specific recommendations:\n\n%s", context),
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	// Call LLM
	response, err := p.llmRouter.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// Parse recommendations (simple line-based parsing)
	recommendations := []string{}
	lines := splitLines(response.Content)
	for _, line := range lines {
		trimmed := trimSpace(line)
		if trimmed != "" && len(trimmed) > 10 {
			// Remove numbering if present
			if len(trimmed) > 3 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
				trimmed = trimSpace(trimmed[3:])
			}
			recommendations = append(recommendations, trimmed)
		}
	}

	if len(recommendations) == 0 {
		recommendations = []string{"No specific recommendations available"}
	}

	return recommendations, nil
}

// buildPatternContext builds context string for LLM.
func (p *PatternLearner) buildPatternContext(events []domain.Event, acceptance []AcceptancePattern, duration []DurationPattern, timezone []TimezonePattern, productivity []ProductivityInsight) string {
	context := fmt.Sprintf("Calendar Analysis (%d events analyzed):\n\n", len(events))

	// Acceptance patterns
	if len(acceptance) > 0 {
		context += "Meeting Acceptance Patterns:\n"
		for i, pattern := range acceptance {
			if i >= 5 {
				break // Top 5
			}
			context += fmt.Sprintf("- %s: %.0f%% acceptance (%d events) - %s\n",
				pattern.TimeSlot, pattern.AcceptRate*100, pattern.EventCount, pattern.Description)
		}
		context += "\n"
	}

	// Duration patterns
	if len(duration) > 0 {
		context += "Meeting Duration Patterns:\n"
		for _, pattern := range duration {
			context += fmt.Sprintf("- %s: avg %d minutes (%d events)\n",
				pattern.MeetingType, pattern.ScheduledDuration, pattern.EventCount)
		}
		context += "\n"
	}

	// Timezone patterns
	if len(timezone) > 0 {
		context += "Timezone Patterns:\n"
		for i, pattern := range timezone {
			if i >= 3 {
				break // Top 3
			}
			context += fmt.Sprintf("- %s: %.0f%% of meetings (%d events)\n",
				pattern.Timezone, pattern.Percentage*100, pattern.EventCount)
		}
		context += "\n"
	}

	// Productivity insights
	if len(productivity) > 0 {
		context += "Productivity Insights:\n"
		for _, insight := range productivity {
			context += fmt.Sprintf("- %s\n", insight.Description)
		}
		context += "\n"
	}

	return context
}

// Helper functions for string operations

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func trimSpace(s string) string {
	// Trim leading and trailing whitespace
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// SavePatterns saves learned patterns (stub for future storage implementation).
func (p *PatternLearner) SavePatterns(ctx context.Context, patterns *SchedulingPatterns) error {
	// Future: Save to local storage/database
	// For now, this is a no-op
	return nil
}

// LoadPatterns loads previously learned patterns (stub for future storage implementation).
func (p *PatternLearner) LoadPatterns(ctx context.Context, userID string) (*SchedulingPatterns, error) {
	// Future: Load from local storage/database
	return nil, fmt.Errorf("pattern storage not yet implemented")
}

// ExportPatterns exports patterns to JSON.
func (p *PatternLearner) ExportPatterns(patterns *SchedulingPatterns) ([]byte, error) {
	return json.MarshalIndent(patterns, "", "  ")
}
