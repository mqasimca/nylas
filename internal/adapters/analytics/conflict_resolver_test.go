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

// TestConflictResolver_DetectConflicts_MeetingOverload tests too many meetings in a day
func TestConflictResolver_DetectConflicts_MeetingOverload(t *testing.T) {
	// Use a fixed date to ensure consistency
	proposedTime := time.Date(2025, 1, 22, 14, 0, 0, 0, time.UTC) // 2 PM

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return 6 meetings on the same day (Jan 22)
			// Spaced throughout the day to ensure they're all on the same calendar day
			day := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)
			events := []domain.Event{
				{
					ID:    "event_1",
					Title: "Morning Standup",
					When: domain.EventWhen{
						StartTime: day.Add(9 * time.Hour).Unix(),
						EndTime:   day.Add(9*time.Hour + 30*time.Minute).Unix(),
					},
					Status: "confirmed",
				},
				{
					ID:    "event_2",
					Title: "Team Sync",
					When: domain.EventWhen{
						StartTime: day.Add(10 * time.Hour).Unix(),
						EndTime:   day.Add(11 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
				{
					ID:    "event_3",
					Title: "Client Call",
					When: domain.EventWhen{
						StartTime: day.Add(11 * time.Hour).Unix(),
						EndTime:   day.Add(12 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
				{
					ID:    "event_4",
					Title: "Lunch Meeting",
					When: domain.EventWhen{
						StartTime: day.Add(12 * time.Hour).Unix(),
						EndTime:   day.Add(13 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
				{
					ID:    "event_5",
					Title: "Design Review",
					When: domain.EventWhen{
						StartTime: day.Add(15 * time.Hour).Unix(),
						EndTime:   day.Add(16 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
				{
					ID:    "event_6",
					Title: "Sprint Planning",
					When: domain.EventWhen{
						StartTime: day.Add(16 * time.Hour).Unix(),
						EndTime:   day.Add(17 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
			}
			return events, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

	// Propose a 7th meeting at 2 PM on the same day
	proposedEvent := &domain.Event{
		Title: "Product Review",
		When: domain.EventWhen{
			StartTime: proposedTime.Unix(),
			EndTime:   proposedTime.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	// Should detect overload conflict (6 existing meetings on the same day)
	foundOverload := false
	for _, conflict := range analysis.SoftConflicts {
		if conflict.Type == domain.ConflictTypeSoftOverload {
			foundOverload = true
			break
		}
	}

	if !foundOverload {
		t.Errorf("Expected overload soft conflict (6 meetings on Jan 22), got none. Total soft conflicts: %d", len(analysis.SoftConflicts))
		for i, c := range analysis.SoftConflicts {
			t.Logf("Conflict %d: Type=%s, Description=%s", i, c.Type, c.Description)
		}
	}
}

// TestConflictResolver_SuggestAlternatives tests alternative time suggestions
func TestConflictResolver_SuggestAlternatives(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return an overlapping event at 10 AM
			return []domain.Event{
				{
					ID:    "event_conflict",
					Title: "Existing Meeting",
					When: domain.EventWhen{
						StartTime: tomorrow.Unix(),
						EndTime:   tomorrow.Add(1 * time.Hour).Unix(),
					},
					Status: "confirmed",
				},
			}, nil
		},
	}

	patterns := &domain.MeetingPattern{
		Acceptance: domain.AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				tomorrow.Weekday().String(): 0.85,
			},
			ByTimeOfDay: map[string]float64{
				"11:00": 0.90, // 11 AM has high acceptance
				"14:00": 0.85, // 2 PM has good acceptance
			},
		},
	}

	resolver := NewConflictResolver(client, patterns)

	// Propose a meeting that conflicts
	proposedEvent := &domain.Event{
		Title: "New Meeting",
		When: domain.EventWhen{
			StartTime: tomorrow.Unix(),
			EndTime:   tomorrow.Add(1 * time.Hour).Unix(),
		},
	}

	analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, patterns)
	if err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}

	// Should suggest alternatives
	if len(analysis.AlternativeTimes) == 0 {
		t.Fatal("Expected alternative times, got none")
	}

	// Alternatives should be sorted by score (highest first)
	for i := 0; i < len(analysis.AlternativeTimes)-1; i++ {
		if analysis.AlternativeTimes[i].Score < analysis.AlternativeTimes[i+1].Score {
			t.Errorf("Alternatives not sorted by score: %d < %d",
				analysis.AlternativeTimes[i].Score,
				analysis.AlternativeTimes[i+1].Score)
		}
	}

	// Alternatives should not have hard conflicts
	for i, alt := range analysis.AlternativeTimes {
		hasHardConflict := false
		for _, conflict := range alt.Conflicts {
			if conflict.Type == domain.ConflictTypeHard {
				hasHardConflict = true
				break
			}
		}
		if hasHardConflict {
			t.Errorf("Alternative %d has hard conflict", i)
		}
	}
}

// TestConflictResolver_GenerateRecommendations tests recommendation generation
func TestConflictResolver_GenerateRecommendations(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{{ID: "cal_1", Name: "Primary"}}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Return back-to-back meetings
			return []domain.Event{
				{
					ID:    "event_before",
					Title: "Previous Meeting",
					When: domain.EventWhen{
						StartTime: tomorrow.Add(-1 * time.Hour).Unix(),
						EndTime:   tomorrow.Unix(),
					},
					Status: "confirmed",
				},
			}, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

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

	// Verify the analysis structure is present
	if analysis == nil {
		t.Fatal("Expected analysis, got nil")
	}

	// Should have soft conflicts detected
	if len(analysis.SoftConflicts) == 0 {
		t.Error("Expected soft conflicts, got none")
	}
}

// TestConflictResolver_MultipleCalendars tests conflict detection across multiple calendars
func TestConflictResolver_MultipleCalendars(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Add(10 * time.Hour)

	client := &testNylasClient{
		getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
			return []domain.Calendar{
				{ID: "cal_1", Name: "Work"},
				{ID: "cal_2", Name: "Personal"},
			}, nil
		},
		getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
			// Only personal calendar has a conflicting event
			if calendarID == "cal_2" {
				return []domain.Event{
					{
						ID:    "event_personal",
						Title: "Personal Appointment",
						When: domain.EventWhen{
							StartTime: tomorrow.Add(15 * time.Minute).Unix(),
							EndTime:   tomorrow.Add(75 * time.Minute).Unix(),
						},
						Status: "confirmed",
					},
				}, nil
			}
			return []domain.Event{}, nil
		},
	}

	resolver := NewConflictResolver(client, nil)

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

	// Should detect conflict from personal calendar
	if len(analysis.HardConflicts) == 0 {
		t.Error("Expected hard conflicts from personal calendar, got none")
	}

	// Conflicting event should be from personal calendar
	if analysis.HardConflicts[0].ConflictingEvent == nil {
		t.Error("Expected conflicting event reference, got nil")
	}

	if analysis.HardConflicts[0].ConflictingEvent.Title != "Personal Appointment" {
		t.Errorf("Conflicting event title = %q, want %q",
			analysis.HardConflicts[0].ConflictingEvent.Title,
			"Personal Appointment")
	}
}

// TestConflictResolver_ErrorHandling tests error handling
func TestConflictResolver_ErrorHandling(t *testing.T) {
	t.Run("GetCalendars error", func(t *testing.T) {
		client := &testNylasClient{
			getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
				return nil, domain.ErrAPIError
			},
		}

		resolver := NewConflictResolver(client, nil)

		proposedEvent := &domain.Event{
			Title: "New Meeting",
			When: domain.EventWhen{
				StartTime: time.Now().Unix(),
				EndTime:   time.Now().Add(1 * time.Hour).Unix(),
			},
		}

		_, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
		if err == nil {
			t.Error("Expected error from GetCalendars, got nil")
		}
	})

	t.Run("GetEvents error", func(t *testing.T) {
		client := &testNylasClient{
			getCalendarsFunc: func(ctx context.Context, grantID string) ([]domain.Calendar, error) {
				return []domain.Calendar{{ID: "cal_1"}}, nil
			},
			getEventsFunc: func(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
				return nil, domain.ErrNetworkError
			},
		}

		resolver := NewConflictResolver(client, nil)

		proposedEvent := &domain.Event{
			Title: "New Meeting",
			When: domain.EventWhen{
				StartTime: time.Now().Unix(),
				EndTime:   time.Now().Add(1 * time.Hour).Unix(),
			},
		}

		// Should still work even if GetEvents fails (empty events)
		analysis, err := resolver.DetectConflicts(context.Background(), "grant_123", proposedEvent, nil)
		if err != nil {
			t.Fatalf("Expected no error when GetEvents fails, got %v", err)
		}

		// Should have no conflicts since no events were retrieved
		if len(analysis.HardConflicts) != 0 || len(analysis.SoftConflicts) != 0 {
			t.Error("Expected no conflicts when GetEvents fails")
		}
	})
}
