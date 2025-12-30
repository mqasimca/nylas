package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// TestConflictResolver_DetectConflicts_NoConflicts tests when there are no conflicts
func TestConflictResolver_DetectConflicts_NoConflicts(t *testing.T) {
	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return an event far in the past
			yesterday := time.Now().AddDate(0, 0, -1)
			return []domain.Event{
				{
					ID:    "event_past",
					Title: "Past Meeting",
					When: domain.EventWhen{
						StartTime: yesterday.Add(-2 * time.Hour).Unix(),
						EndTime:   yesterday.Unix(),
					},
					Status: "confirmed",
				},
			}, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

	// Propose a meeting tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)
	proposedEvent := &domain.Event{
		Title: "New Meeting",
		When: domain.EventWhen{
			StartTime: tomorrow.Unix(),
			EndTime:   tomorrow.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	if len(analysis.HardConflicts) != 0 {
		t.Errorf("HardConflicts = %d, want 0", len(analysis.HardConflicts))
	}

	if len(analysis.SoftConflicts) != 0 {
		t.Errorf("SoftConflicts = %d, want 0", len(analysis.SoftConflicts))
	}

	if !analysis.CanProceed {
		t.Error("CanProceed = false, want true")
	}
}

// TestConflictResolver_DetectConflicts_HardConflict tests overlapping meetings
func TestConflictResolver_DetectConflicts_HardConflict(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return an overlapping event
			return []domain.Event{
				{
					ID:    "event_conflict",
					Title: "Existing Meeting",
					When: domain.EventWhen{
						StartTime: tomorrow.Add(15 * time.Minute).Unix(), // Overlaps with proposed
						EndTime:   tomorrow.Add(75 * time.Minute).Unix(),
					},
					Status: "confirmed",
				},
			}, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

	// Propose a meeting that overlaps
	proposedEvent := &domain.Event{
		Title: "New Meeting",
		When: domain.EventWhen{
			StartTime: tomorrow.Unix(),
			EndTime:   tomorrow.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	if len(analysis.HardConflicts) == 0 {
		t.Error("Expected hard conflicts, got none")
	}

	if analysis.HardConflicts[0].Type != domain.ConflictTypeHard {
		t.Errorf("ConflictType = %v, want %v", analysis.HardConflicts[0].Type, domain.ConflictTypeHard)
	}

	if analysis.HardConflicts[0].Severity != domain.SeverityCritical {
		t.Errorf("Severity = %v, want %v", analysis.HardConflicts[0].Severity, domain.SeverityCritical)
	}

	if analysis.CanProceed {
		t.Error("CanProceed = true, want false (hard conflict)")
	}

	// Should suggest alternatives
	if len(analysis.AlternativeTimes) == 0 {
		t.Error("Expected alternative times, got none")
	}
}

// TestConflictResolver_DetectConflicts_BackToBack tests soft conflict for back-to-back meetings
func TestConflictResolver_DetectConflicts_BackToBack(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return an event that ends exactly when proposed starts
			return []domain.Event{
				{
					ID:    "event_before",
					Title: "Previous Meeting",
					When: domain.EventWhen{
						StartTime: tomorrow.Add(-1 * time.Hour).Unix(),
						EndTime:   tomorrow.Unix(), // Ends exactly when proposed starts
					},
					Status: "confirmed",
				},
			}, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

	// Propose a meeting immediately after
	proposedEvent := &domain.Event{
		Title: "New Meeting",
		When: domain.EventWhen{
			StartTime: tomorrow.Unix(),
			EndTime:   tomorrow.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	if len(analysis.HardConflicts) != 0 {
		t.Errorf("HardConflicts = %d, want 0", len(analysis.HardConflicts))
	}

	// Should detect soft conflict for back-to-back
	foundBackToBack := false
	for _, conflict := range analysis.SoftConflicts {
		if conflict.Type == domain.ConflictTypeSoftBackToBack {
			foundBackToBack = true
			break
		}
	}

	if !foundBackToBack {
		t.Error("Expected back-to-back soft conflict, got none")
	}

	if !analysis.CanProceed {
		t.Error("CanProceed = false, want true (soft conflicts allow proceeding)")
	}
}

// TestConflictResolver_DetectConflicts_FocusTime tests focus time interruption
func TestConflictResolver_DetectConflicts_FocusTime(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	// Propose meeting at 9 AM on Monday
	proposedTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, time.Local)
	for proposedTime.Weekday() != time.Monday {
		proposedTime = proposedTime.AddDate(0, 0, 1)
	}

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			return []domain.Event{}, nil
		},
	}

	// Create patterns with focus time blocks
	patterns := &domain.MeetingPattern{
		Productivity: domain.ProductivityPatterns{
			FocusBlocks: []domain.TimeBlock{
				{
					DayOfWeek: "Monday",
					StartTime: "09:00",
					EndTime:   "11:00",
					Score:     90.0,
				},
			},
		},
	}

	resolver := NewConflictResolver(client, patterns)

	proposedEvent := &domain.Event{
		Title: "New Meeting",
		When: domain.EventWhen{
			StartTime: proposedTime.Unix(),
			EndTime:   proposedTime.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, patterns)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	// Should detect focus time conflict
	foundFocusTime := false
	for _, conflict := range analysis.SoftConflicts {
		if conflict.Type == domain.ConflictTypeSoftFocusTime {
			foundFocusTime = true
			if conflict.Severity != domain.SeverityHigh {
				t.Errorf("FocusTime Severity = %v, want %v", conflict.Severity, domain.SeverityHigh)
			}
			break
		}
	}

	if !foundFocusTime {
		t.Error("Expected focus time soft conflict, got none")
	}
}
