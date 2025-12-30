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
