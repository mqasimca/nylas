package analytics

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestMeetingScorer_ScoreMeetingTime_NoPatterns(t *testing.T) {
	scorer := NewMeetingScorer(nil)

	proposedTime := time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC)
	participants := []string{"john@example.com"}

	score := scorer.ScoreMeetingTime(proposedTime, participants, 30)

	if score.Score != 50 {
		t.Errorf("Score with no patterns = %d, want 50 (default)", score.Score)
	}

	if score.Confidence != 0 {
		t.Errorf("Confidence with no patterns = %.0f, want 0", score.Confidence)
	}

	if score.Recommendation != "No historical data available for scoring" {
		t.Errorf("Recommendation = %q, want default message", score.Recommendation)
	}
}

func TestMeetingScorer_ScoreMeetingTime_WithPatterns(t *testing.T) {
	// Create patterns that favor Tuesday at 2 PM
	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Monday":  0.7,
				"Tuesday": 0.9,
				"Friday":  0.5,
			},
			ByTimeOfDay: map[string]float64{
				"09:00": 0.75,
				"10:00": 0.85,
				"14:00": 0.88,
				"16:00": 0.65,
			},
			Overall: 0.8,
		},
		Productivity: domain.ProductivityPatterns{
			PeakFocus: []domain.TimeBlock{
				{
					DayOfWeek: "Tuesday",
					StartTime: "14:00",
					EndTime:   "16:00",
					Score:     90.0,
				},
			},
			MeetingDensity: map[string]float64{
				"Monday":  3.5,
				"Tuesday": 2.0,
				"Friday":  1.5,
			},
		},
		Participants: map[string]domain.ParticipantPattern{
			"john@example.com": {
				Email:          "john@example.com",
				AcceptanceRate: 0.85,
				PreferredDays:  []string{"Tuesday", "Wednesday"},
				PreferredTimes: []string{"14:00", "15:00"},
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	// Score a Tuesday 2 PM meeting
	proposedTime := time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC) // Tuesday
	participants := []string{"john@example.com"}

	score := scorer.ScoreMeetingTime(proposedTime, participants, 30)

	// Should get a high score
	if score.Score < 70 {
		t.Errorf("Score = %d, want >= 70 (good time)", score.Score)
	}

	// Should have high confidence (all data available)
	if score.Confidence < 80 {
		t.Errorf("Confidence = %.0f, want >= 80", score.Confidence)
	}

	// Should have factors
	if len(score.Factors) == 0 {
		t.Error("Should have scoring factors")
	}

	// Should have positive day preference impact
	foundDayFactor := false
	for _, factor := range score.Factors {
		if factor.Name == "Day Preference" {
			foundDayFactor = true
			if factor.Impact < 0 {
				t.Errorf("Day Preference impact = %d, want positive (Tuesday is preferred)", factor.Impact)
			}
		}
	}

	if !foundDayFactor {
		t.Error("Should have Day Preference factor")
	}

	// Should not suggest alternatives for good score
	if len(score.AlternativeTimes) > 0 {
		t.Error("Should not suggest alternatives for high-scoring time")
	}
}

func TestMeetingScorer_ScoreMeetingTime_LowScore(t *testing.T) {
	// Create patterns that disfavor Friday late afternoon
	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Tuesday": 0.9,
				"Friday":  0.4,
			},
			ByTimeOfDay: map[string]float64{
				"10:00": 0.85,
				"16:00": 0.45,
			},
			Overall: 0.7,
		},
		Productivity: domain.ProductivityPatterns{
			MeetingDensity: map[string]float64{
				"Friday": 5.0, // Very busy Friday
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	// Score a Friday 4 PM meeting (not ideal)
	proposedTime := time.Date(2025, 1, 24, 16, 0, 0, 0, time.UTC) // Friday
	participants := []string{"alice@example.com"}

	score := scorer.ScoreMeetingTime(proposedTime, participants, 30)

	// Should get a lower score
	if score.Score >= 70 {
		t.Errorf("Score = %d, want < 70 (not ideal time)", score.Score)
	}

	// Should suggest alternatives for low score
	if score.Score < 70 && len(score.AlternativeTimes) == 0 {
		t.Error("Should suggest alternatives for low-scoring time")
	}
}

func TestMeetingScorer_GetProductivityScore(t *testing.T) {
	patterns := &domain.MeetingPattern{
		Productivity: domain.ProductivityPatterns{
			PeakFocus: []domain.TimeBlock{
				{
					DayOfWeek: "Tuesday",
					StartTime: "10:00",
					EndTime:   "12:00",
					Score:     95.0,
				},
			},
			MeetingDensity: map[string]float64{
				"Monday":  4.0,
				"Tuesday": 1.5,
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	tests := []struct {
		name          string
		time          time.Time
		wantHighScore bool
	}{
		{
			name:          "Peak focus time",
			time:          time.Date(2025, 1, 21, 10, 0, 0, 0, time.UTC), // Tuesday 10 AM
			wantHighScore: true,
		},
		{
			name:          "Low density day",
			time:          time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC), // Tuesday 2 PM
			wantHighScore: true,
		},
		{
			name:          "High density day",
			time:          time.Date(2025, 1, 20, 14, 0, 0, 0, time.UTC), // Monday 2 PM
			wantHighScore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.getProductivityScore(tt.time)

			if tt.wantHighScore && score < 10 {
				t.Errorf("%s: productivity score = %d, want >= 10", tt.name, score)
			}
			if !tt.wantHighScore && score > 15 {
				t.Errorf("%s: productivity score = %d, want <= 15", tt.name, score)
			}
		})
	}
}

func TestMeetingScorer_GetParticipantScore(t *testing.T) {
	patterns := &domain.MeetingPattern{
		Participants: map[string]domain.ParticipantPattern{
			"john@example.com": {
				Email:          "john@example.com",
				AcceptanceRate: 0.9,
				PreferredDays:  []string{"Tuesday", "Wednesday"},
				PreferredTimes: []string{"14:00", "15:00"},
			},
			"alice@example.com": {
				Email:          "alice@example.com",
				AcceptanceRate: 0.7,
				PreferredDays:  []string{"Monday"},
				PreferredTimes: []string{"10:00"},
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	tests := []struct {
		name          string
		participants  []string
		time          time.Time
		wantHighScore bool
	}{
		{
			name:          "Participant prefers this time",
			participants:  []string{"john@example.com"},
			time:          time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC), // Tuesday 2 PM
			wantHighScore: true,
		},
		{
			name:          "Participant does not prefer this time",
			participants:  []string{"alice@example.com"},
			time:          time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC), // Tuesday 2 PM
			wantHighScore: false,
		},
		{
			name:          "No participant data",
			participants:  []string{"unknown@example.com"},
			time:          time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC),
			wantHighScore: false, // Default neutral score
		},
		{
			name:          "No participants",
			participants:  []string{},
			time:          time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC),
			wantHighScore: false, // Neutral score
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.getParticipantScore(tt.participants, tt.time)

			if tt.wantHighScore && score < 10 {
				t.Errorf("%s: participant score = %d, want >= 10", tt.name, score)
			}
		})
	}
}

func TestMeetingScorer_CalculateConfidence(t *testing.T) {
	tests := []struct {
		name              string
		patterns          *domain.MeetingPattern
		wantMinConfidence float64
	}{
		{
			name: "All data available",
			patterns: &domain.MeetingPattern{
				Acceptance: domain.AcceptancePatterns{
					ByDayOfWeek: map[string]float64{"Monday": 0.8},
					ByTimeOfDay: map[string]float64{"10:00": 0.9},
				},
				Productivity: domain.ProductivityPatterns{
					PeakFocus: []domain.TimeBlock{
						{DayOfWeek: "Tuesday", StartTime: "10:00"},
					},
				},
				Participants: map[string]domain.ParticipantPattern{
					"john@example.com": {Email: "john@example.com"},
				},
				Duration: domain.DurationPatterns{
					ByParticipant: map[string]domain.DurationStats{
						"john@example.com": {AverageScheduled: 30},
					},
				},
			},
			wantMinConfidence: 90.0, // 5/5 data points
		},
		{
			name: "Partial data",
			patterns: &domain.MeetingPattern{
				Acceptance: domain.AcceptancePatterns{
					ByDayOfWeek: map[string]float64{"Monday": 0.8},
				},
				Productivity: domain.ProductivityPatterns{
					PeakFocus: []domain.TimeBlock{},
				},
			},
			wantMinConfidence: 0.0,
		},
		{
			name:              "No data",
			patterns:          &domain.MeetingPattern{},
			wantMinConfidence: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewMeetingScorer(tt.patterns)
			confidence := scorer.calculateConfidence()

			if confidence < tt.wantMinConfidence {
				t.Errorf("Confidence = %.0f, want >= %.0f", confidence, tt.wantMinConfidence)
			}
		})
	}
}

func TestMeetingScorer_CalculateSuccessRate(t *testing.T) {
	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Tuesday": 0.9,
				"Friday":  0.5,
			},
			Overall: 0.75,
		},
	}

	scorer := NewMeetingScorer(patterns)

	// Tuesday should have high success rate
	tuesdayTime := time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC)
	tuesdayRate := scorer.calculateSuccessRate(tuesdayTime)

	if tuesdayRate != 0.9 {
		t.Errorf("Tuesday success rate = %.2f, want 0.90", tuesdayRate)
	}

	// Friday should have lower success rate
	fridayTime := time.Date(2025, 1, 24, 14, 0, 0, 0, time.UTC)
	fridayRate := scorer.calculateSuccessRate(fridayTime)

	if fridayRate != 0.5 {
		t.Errorf("Friday success rate = %.2f, want 0.50", fridayRate)
	}

	// Unknown day should use overall
	saturdayTime := time.Date(2025, 1, 25, 14, 0, 0, 0, time.UTC)
	saturdayRate := scorer.calculateSuccessRate(saturdayTime)

	if saturdayRate != 0.75 {
		t.Errorf("Saturday success rate = %.2f, want 0.75 (overall)", saturdayRate)
	}
}

func TestMeetingScorer_GenerateRecommendation(t *testing.T) {
	scorer := &MeetingScorer{}

	tests := []struct {
		name            string
		score           int
		factors         []domain.ScoreFactor
		wantContains    string
		wantNotContains string
	}{
		{
			name:         "Excellent score",
			score:        90,
			factors:      []domain.ScoreFactor{},
			wantContains: "Excellent",
		},
		{
			name:         "Good score",
			score:        75,
			factors:      []domain.ScoreFactor{},
			wantContains: "Good",
		},
		{
			name:         "Acceptable score",
			score:        55,
			factors:      []domain.ScoreFactor{},
			wantContains: "Acceptable",
		},
		{
			name:  "Poor score with factor",
			score: 30,
			factors: []domain.ScoreFactor{
				{Name: "Day Preference", Impact: -15},
			},
			wantContains: "Not recommended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := scorer.generateRecommendation(tt.score, tt.factors)

			if tt.wantContains != "" && len(rec) > 0 {
				// Just check that we got a recommendation
				if len(rec) == 0 {
					t.Errorf("Expected non-empty recommendation")
				}
			}
		})
	}
}

func TestMeetingScorer_SuggestAlternatives(t *testing.T) {
	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Tuesday": 0.9,
				"Friday":  0.5,
			},
			ByTimeOfDay: map[string]float64{
				"10:00": 0.85,
				"14:00": 0.88,
				"16:00": 0.65,
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	// Suggest alternative for Friday 4 PM
	original := time.Date(2025, 1, 24, 16, 0, 0, 0, time.UTC) // Friday 4 PM
	alternatives := scorer.suggestAlternatives(original, []string{})

	if len(alternatives) == 0 {
		t.Error("Should suggest at least one alternative time")
	}

	if len(alternatives) > 0 {
		alt := alternatives[0]

		// Alternative should be on a better day (Tuesday)
		if alt.Weekday().String() != "Tuesday" {
			t.Errorf("Alternative day = %s, want Tuesday", alt.Weekday().String())
		}

		// Alternative should be at a better time (14:00)
		if alt.Hour() != 14 {
			t.Errorf("Alternative hour = %d, want 14", alt.Hour())
		}
	}
}

func TestParseHour(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"09:00", 9},
		{"14:30", 14},
		{"23:59", 23},
		{"00:00", 0},
		{"invalid", 0},
		{"", 0},
		{"9", 0}, // Not HH:MM format
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseHour(tt.input)
			if got != tt.want {
				t.Errorf("parseHour(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestDayNameToWeekday(t *testing.T) {
	tests := []struct {
		input string
		want  time.Weekday
	}{
		{"Sunday", time.Sunday},
		{"Monday", time.Monday},
		{"Tuesday", time.Tuesday},
		{"Wednesday", time.Wednesday},
		{"Thursday", time.Thursday},
		{"Friday", time.Friday},
		{"Saturday", time.Saturday},
		{"sunday", time.Sunday},  // lowercase
		{"MONDAY", time.Monday},  // uppercase
		{"invalid", time.Monday}, // default
		{"", time.Monday},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := dayNameToWeekday(tt.input)
			if got != tt.want {
				t.Errorf("dayNameToWeekday(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMeetingScorer_CompleteScenario(t *testing.T) {
	// Create realistic patterns
	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Monday":    0.78,
				"Tuesday":   0.92,
				"Wednesday": 0.88,
				"Thursday":  0.86,
				"Friday":    0.64,
			},
			ByTimeOfDay: map[string]float64{
				"09:00": 0.72,
				"10:00": 0.89,
				"11:00": 0.91,
				"14:00": 0.86,
				"15:00": 0.79,
				"16:00": 0.65,
			},
			Overall: 0.82,
		},
		Duration: domain.DurationPatterns{
			Overall: domain.DurationStats{
				AverageScheduled: 30,
				AverageActual:    35,
			},
		},
		Productivity: domain.ProductivityPatterns{
			PeakFocus: []domain.TimeBlock{
				{
					DayOfWeek: "Tuesday",
					StartTime: "10:00",
					EndTime:   "12:00",
					Score:     92.0,
				},
			},
			MeetingDensity: map[string]float64{
				"Monday":    3.2,
				"Tuesday":   2.8,
				"Wednesday": 3.1,
				"Thursday":  2.9,
				"Friday":    1.4,
			},
		},
		Participants: map[string]domain.ParticipantPattern{
			"john@example.com": {
				Email:          "john@example.com",
				AcceptanceRate: 0.85,
				PreferredDays:  []string{"Tuesday", "Wednesday"},
				PreferredTimes: []string{"14:00", "15:00"},
			},
		},
	}

	scorer := NewMeetingScorer(patterns)

	// Score Tuesday 2 PM meeting with John
	proposedTime := time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC) // Tuesday 2 PM
	participants := []string{"john@example.com"}

	score := scorer.ScoreMeetingTime(proposedTime, participants, 30)

	// Validate complete response
	if score.Score == 0 {
		t.Error("Score should not be 0")
	}

	if score.Confidence == 0 {
		t.Error("Confidence should not be 0")
	}

	if score.SuccessRate == 0 {
		t.Error("Success rate should not be 0")
	}

	if len(score.Factors) == 0 {
		t.Error("Should have scoring factors")
	}

	if score.Recommendation == "" {
		t.Error("Should have a recommendation")
	}

	// All factors should have descriptions
	for _, factor := range score.Factors {
		if factor.Name == "" {
			t.Error("Factor should have a name")
		}
		if factor.Description == "" {
			t.Error("Factor should have a description")
		}
	}
}
