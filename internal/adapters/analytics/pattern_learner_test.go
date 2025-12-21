package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// testNylasClient is a minimal test implementation
type testNylasClient struct {
	getCalendarsFunc func(ctx context.Context, grantID string) ([]domain.Calendar, error)
	getEventsFunc    func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error)
}

func (t *testNylasClient) GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error) {
	if t.getCalendarsFunc != nil {
		return t.getCalendarsFunc(ctx, grantID)
	}
	return []domain.Calendar{}, nil
}

func (t *testNylasClient) GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	if t.getEventsFunc != nil {
		return t.getEventsFunc(ctx, grantID, calendarID, params)
	}
	return []domain.Event{}, nil
}

// Helper to create test events
func createTestEvent(title string, start time.Time, duration int, status string) domain.Event {
	end := start.Add(time.Duration(duration) * time.Minute)
	return domain.Event{
		ID:     "event_" + title,
		Title:  title,
		Status: status,
		When: domain.EventWhen{
			StartTime: start.Unix(),
			EndTime:   end.Unix(),
		},
		Participants: []domain.Participant{},
	}
}

func TestPatternLearner_AnalyzeHistory_NoEvents(t *testing.T) {
	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{
				{ID: "cal_1", Name: "Primary"},
			}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			return []domain.Event{}, nil
		},
	}

	learner := NewPatternLearner(client)
	ctx := context.Background()

	analysis, err := learner.AnalyzeHistory(ctx, "grant_123", 90)
	if err != nil {
		t.Fatalf("AnalyzeHistory() error = %v, want nil", err)
	}

	if analysis.TotalMeetings != 0 {
		t.Errorf("TotalMeetings = %d, want 0", analysis.TotalMeetings)
	}

	if len(analysis.Insights) == 0 {
		t.Error("Expected at least one insight for no meetings")
	}
}

func TestPatternLearner_AnalyzeHistory_WithEvents(t *testing.T) {
	now := time.Now()
	monday := now.AddDate(0, 0, -int(now.Weekday())+int(time.Monday))

	events := []domain.Event{
		// Monday 10 AM - confirmed
		createTestEvent("Monday Meeting", monday.Add(10*time.Hour), 30, "confirmed"),
		// Tuesday 2 PM - confirmed
		createTestEvent("Tuesday Meeting", monday.AddDate(0, 0, 1).Add(14*time.Hour), 60, "confirmed"),
		// Wednesday 10 AM - confirmed
		createTestEvent("Wednesday Meeting", monday.AddDate(0, 0, 2).Add(10*time.Hour), 30, "confirmed"),
		// Thursday 2 PM - cancelled
		createTestEvent("Thursday Meeting", monday.AddDate(0, 0, 3).Add(14*time.Hour), 30, "cancelled"),
		// Friday 4 PM - confirmed
		createTestEvent("Friday Meeting", monday.AddDate(0, 0, 4).Add(16*time.Hour), 30, "confirmed"),
	}

	client := &testNylasClient{}

	client.getCalendarsFunc = func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
		return []domain.Calendar{
			{ID: "cal_1", Name: "Primary"},
		}, nil
	}

	client.getEventsFunc = func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
		return events, nil
	}

	learner := NewPatternLearner(client)
	ctx := context.Background()

	analysis, err := learner.AnalyzeHistory(ctx, "grant_123", 90)
	if err != nil {
		t.Fatalf("AnalyzeHistory() error = %v, want nil", err)
	}

	if analysis.TotalMeetings != 5 {
		t.Errorf("TotalMeetings = %d, want 5", analysis.TotalMeetings)
	}

	if analysis.Patterns == nil {
		t.Fatal("Patterns should not be nil")
	}

	// Check acceptance patterns
	if analysis.Patterns.Acceptance.Overall == 0 {
		t.Error("Overall acceptance rate should not be 0")
	}

	// Should have acceptance data for days
	if len(analysis.Patterns.Acceptance.ByDayOfWeek) == 0 {
		t.Error("Should have acceptance patterns by day of week")
	}

	// Check that we have recommendations
	if len(analysis.Recommendations) == 0 {
		t.Error("Should have at least one recommendation")
	}

	// Check that we have insights
	if len(analysis.Insights) == 0 {
		t.Error("Should have at least one insight")
	}
}

func TestPatternLearner_AnalyzeHistory_MultipleCalendars(t *testing.T) {
	now := time.Now()

	cal1Events := []domain.Event{
		createTestEvent("Cal1 Meeting", now.Add(-24*time.Hour), 30, "confirmed"),
	}

	cal2Events := []domain.Event{
		createTestEvent("Cal2 Meeting", now.Add(-48*time.Hour), 60, "confirmed"),
	}

	client := &testNylasClient{}

	client.getCalendarsFunc = func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
		return []domain.Calendar{
			{ID: "cal_1", Name: "Primary"},
			{ID: "cal_2", Name: "Work"},
		}, nil
	}

	client.getEventsFunc = func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
		if calendarID == "cal_1" {
			return cal1Events, nil
		}
		if calendarID == "cal_2" {
			return cal2Events, nil
		}
		return []domain.Event{}, nil
	}

	learner := NewPatternLearner(client)
	ctx := context.Background()

	analysis, err := learner.AnalyzeHistory(ctx, "grant_123", 90)
	if err != nil {
		t.Fatalf("AnalyzeHistory() error = %v, want nil", err)
	}

	// Should aggregate events from both calendars
	if analysis.TotalMeetings != 2 {
		t.Errorf("TotalMeetings = %d, want 2", analysis.TotalMeetings)
	}
}

func TestPatternLearner_LearnAcceptancePatterns(t *testing.T) {
	monday := time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)

	events := []domain.Event{
		// 4 confirmed Monday meetings
		createTestEvent("Mon1", monday, 30, "confirmed"),
		createTestEvent("Mon2", monday.Add(2*time.Hour), 30, "confirmed"),
		createTestEvent("Mon3", monday.Add(4*time.Hour), 30, "confirmed"),
		createTestEvent("Mon4", monday.Add(6*time.Hour), 30, "confirmed"),
		// 1 cancelled Monday meeting
		createTestEvent("Mon5", monday.Add(8*time.Hour), 30, "cancelled"),
		// Overall: 4/5 = 80% acceptance on Monday
	}

	learner := &PatternLearner{}
	patterns := learner.learnAcceptancePatterns(events)

	if patterns.Overall == 0 {
		t.Error("Overall acceptance rate should not be 0")
	}

	mondayRate, exists := patterns.ByDayOfWeek["Monday"]
	if !exists {
		t.Fatal("Should have Monday acceptance rate")
	}

	expectedRate := 0.8 // 4 confirmed / 5 total
	tolerance := 0.01
	if mondayRate < expectedRate-tolerance || mondayRate > expectedRate+tolerance {
		t.Errorf("Monday acceptance rate = %.2f, want ~%.2f", mondayRate, expectedRate)
	}

	// Should have time-based patterns
	if len(patterns.ByTimeOfDay) == 0 {
		t.Error("Should have acceptance patterns by time of day")
	}
}

func TestPatternLearner_LearnProductivityPatterns(t *testing.T) {
	monday := time.Date(2025, 1, 6, 9, 0, 0, 0, time.UTC)

	// Create events that leave Tuesday 10-12 free (good for focus)
	events := []domain.Event{
		// Monday: 3 meetings
		createTestEvent("Mon1", monday, 30, "confirmed"),
		createTestEvent("Mon2", monday.Add(2*time.Hour), 30, "confirmed"),
		createTestEvent("Mon3", monday.Add(4*time.Hour), 30, "confirmed"),
		// Tuesday: 1 meeting at 2 PM (leaves morning free)
		createTestEvent("Tue1", monday.AddDate(0, 0, 1).Add(5*time.Hour), 30, "confirmed"),
		// Wednesday: 4 meetings (busy day)
		createTestEvent("Wed1", monday.AddDate(0, 0, 2), 30, "confirmed"),
		createTestEvent("Wed2", monday.AddDate(0, 0, 2).Add(2*time.Hour), 30, "confirmed"),
		createTestEvent("Wed3", monday.AddDate(0, 0, 2).Add(4*time.Hour), 30, "confirmed"),
		createTestEvent("Wed4", monday.AddDate(0, 0, 2).Add(6*time.Hour), 30, "confirmed"),
	}

	learner := &PatternLearner{}
	patterns := learner.learnProductivityPatterns(events)

	// Should have meeting density data
	if len(patterns.MeetingDensity) == 0 {
		t.Error("Should have meeting density data")
	}

	// Wednesday should have higher density than Tuesday
	wedDensity, wedExists := patterns.MeetingDensity["Wednesday"]
	tueDensity, tueExists := patterns.MeetingDensity["Tuesday"]

	if !wedExists || !tueExists {
		t.Fatal("Should have density data for Tuesday and Wednesday")
	}

	if wedDensity <= tueDensity {
		t.Errorf("Wednesday density (%.1f) should be higher than Tuesday (%.1f)", wedDensity, tueDensity)
	}

	// Should identify focus blocks
	if len(patterns.FocusBlocks) == 0 {
		t.Error("Should have identified at least one focus block")
	}
}

func TestPatternLearner_GenerateRecommendations(t *testing.T) {
	pattern := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Monday":    0.6,
				"Tuesday":   0.9,
				"Wednesday": 0.85,
				"Friday":    0.4,
			},
			Overall: 0.75,
		},
		Duration: domain.DurationPatterns{
			Overall: domain.DurationStats{
				AverageScheduled: 30,
				AverageActual:    35,
				OverrunRate:      0.6,
			},
		},
		Productivity: domain.ProductivityPatterns{
			FocusBlocks: []domain.TimeBlock{
				{
					DayOfWeek: "Tuesday",
					StartTime: "10:00",
					EndTime:   "12:00",
					Score:     95.0,
				},
			},
		},
	}

	learner := &PatternLearner{}
	recommendations := learner.generateRecommendations(pattern, []domain.Event{})

	if len(recommendations) == 0 {
		t.Error("Should generate at least one recommendation")
	}

	// Should recommend blocking Tuesday 10-12 for focus time
	foundFocusRec := false
	for _, rec := range recommendations {
		if rec.Type == "focus_time" {
			foundFocusRec = true
			if rec.Priority != "high" && rec.Priority != "medium" {
				t.Errorf("Focus time recommendation priority = %s, want high or medium", rec.Priority)
			}
			if rec.Confidence == 0 {
				t.Error("Recommendation should have confidence score")
			}
		}
	}

	if !foundFocusRec {
		t.Error("Should recommend focus time blocking")
	}
}

func TestPatternLearner_GenerateInsights(t *testing.T) {
	pattern := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Monday": 0.9,
				"Friday": 0.4,
			},
			Overall: 0.7,
		},
		Duration: domain.DurationPatterns{
			Overall: domain.DurationStats{
				AverageScheduled: 30,
				AverageActual:    40,
			},
		},
	}

	learner := &PatternLearner{}
	insights := learner.generateInsights(pattern, []domain.Event{})

	if len(insights) == 0 {
		t.Error("Should generate at least one insight")
	}

	// Insights should mention acceptance patterns
	foundAcceptanceInsight := false
	for _, insight := range insights {
		if len(insight) > 0 {
			foundAcceptanceInsight = true
			break
		}
	}

	if !foundAcceptanceInsight {
		t.Error("Should generate insights about patterns")
	}
}

func TestPatternLearner_FetchEvents_Error(t *testing.T) {
	client := &testNylasClient{}

	client.getCalendarsFunc = func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
		return nil, errors.New("unauthorized")
	}

	learner := NewPatternLearner(client)
	ctx := context.Background()

	_, err := learner.AnalyzeHistory(ctx, "grant_123", 90)
	if err == nil {
		t.Error("Expected error when client fails, got nil")
	}
}

func TestPatternLearner_DateRangePeriod(t *testing.T) {
	client := &testNylasClient{}

	client.getCalendarsFunc = func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
		return []domain.Calendar{
			{ID: "cal_1", Name: "Primary"},
		}, nil
	}

	client.getEventsFunc = func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
		return []domain.Event{}, nil
	}

	learner := NewPatternLearner(client)
	ctx := context.Background()

	analysis, err := learner.AnalyzeHistory(ctx, "grant_123", 60)
	if err != nil {
		t.Fatalf("AnalyzeHistory() error = %v, want nil", err)
	}

	// Check that period is approximately 60 days
	periodDuration := analysis.Period.End.Sub(analysis.Period.Start)
	expectedDuration := 60 * 24 * time.Hour

	// Allow 1 day tolerance for timezone differences
	tolerance := 24 * time.Hour
	if periodDuration < expectedDuration-tolerance || periodDuration > expectedDuration+tolerance {
		t.Errorf("Period duration = %v, want ~%v", periodDuration, expectedDuration)
	}
}
