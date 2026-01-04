package domain

import (
	"testing"
	"time"
)

// =============================================================================
// FocusTimeSettings Tests
// =============================================================================

func TestFocusTimeSettings_Creation(t *testing.T) {
	settings := FocusTimeSettings{
		Enabled:             true,
		AutoBlock:           true,
		AutoDecline:         false,
		MinBlockDuration:    60,
		MaxBlockDuration:    180,
		TargetHoursPerWeek:  10.0,
		AllowUrgentOverride: true,
		RequireApproval:     true,
		ProtectedDays:       []string{"Wednesday", "Friday"},
		ExcludedTimeRanges: []TimeRange{
			{StartTime: "12:00", EndTime: "13:00"},
		},
		NotificationSettings: FocusTimeNotificationPrefs{
			NotifyOnDecline:  true,
			NotifyOnOverride: true,
			DailySummary:     false,
			WeeklySummary:    true,
		},
	}

	if !settings.Enabled {
		t.Error("FocusTimeSettings.Enabled should be true")
	}
	if settings.MinBlockDuration != 60 {
		t.Errorf("FocusTimeSettings.MinBlockDuration = %d, want 60", settings.MinBlockDuration)
	}
	if settings.TargetHoursPerWeek != 10.0 {
		t.Errorf("FocusTimeSettings.TargetHoursPerWeek = %f, want 10.0", settings.TargetHoursPerWeek)
	}
	if len(settings.ProtectedDays) != 2 {
		t.Errorf("FocusTimeSettings.ProtectedDays length = %d, want 2", len(settings.ProtectedDays))
	}
	if !settings.NotificationSettings.WeeklySummary {
		t.Error("NotificationSettings.WeeklySummary should be true")
	}
}

// =============================================================================
// TimeRange Tests
// =============================================================================

func TestTimeRange_Creation(t *testing.T) {
	tr := TimeRange{
		StartTime: "09:00",
		EndTime:   "17:00",
	}

	if tr.StartTime != "09:00" {
		t.Errorf("TimeRange.StartTime = %q, want %q", tr.StartTime, "09:00")
	}
	if tr.EndTime != "17:00" {
		t.Errorf("TimeRange.EndTime = %q, want %q", tr.EndTime, "17:00")
	}
}

// =============================================================================
// ProtectedBlock Tests
// =============================================================================

func TestProtectedBlock_Creation(t *testing.T) {
	now := time.Now()
	block := ProtectedBlock{
		ID:                "block-123",
		CalendarEventID:   "event-456",
		StartTime:         now,
		EndTime:           now.Add(2 * time.Hour),
		Duration:          120,
		IsRecurring:       true,
		RecurrencePattern: "weekly",
		Priority:          PriorityHigh,
		Reason:            "Deep work session",
		AllowOverride:     true,
		ProtectionRules: FocusProtectionRule{
			AutoDecline:         true,
			SuggestAlternatives: true,
			DeclineMessage:      "I'm in a focus block. Please suggest another time.",
		},
	}

	if block.Duration != 120 {
		t.Errorf("ProtectedBlock.Duration = %d, want 120", block.Duration)
	}
	if !block.IsRecurring {
		t.Error("ProtectedBlock.IsRecurring should be true")
	}
	if block.Priority != PriorityHigh {
		t.Errorf("ProtectedBlock.Priority = %v, want %v", block.Priority, PriorityHigh)
	}
	if !block.ProtectionRules.AutoDecline {
		t.Error("ProtectionRules.AutoDecline should be true")
	}
}

// =============================================================================
// AdaptiveScheduleChange Tests
// =============================================================================

func TestAdaptiveScheduleChange_Creation(t *testing.T) {
	now := time.Now()
	change := AdaptiveScheduleChange{
		ID:             "change-123",
		Timestamp:      now,
		Trigger:        TriggerFocusTimeAtRisk,
		ChangeType:     ChangeTypeRescheduleMeeting,
		AffectedEvents: []string{"event-1", "event-2"},
		Changes: []ScheduleModification{
			{
				EventID:      "event-1",
				Action:       "reschedule",
				OldStartTime: now,
				NewStartTime: now.Add(24 * time.Hour),
				Description:  "Moved to protect focus time",
			},
		},
		Reason: "Focus time protection triggered",
		Impact: AdaptiveImpact{
			FocusTimeGained:     2.0,
			MeetingsRescheduled: 1,
			DurationSaved:       0,
			ConflictsResolved:   1,
		},
		UserApproval: ApprovalPending,
		Confidence:   85.0,
	}

	if change.Trigger != TriggerFocusTimeAtRisk {
		t.Errorf("AdaptiveScheduleChange.Trigger = %v, want %v", change.Trigger, TriggerFocusTimeAtRisk)
	}
	if change.ChangeType != ChangeTypeRescheduleMeeting {
		t.Errorf("AdaptiveScheduleChange.ChangeType = %v, want %v", change.ChangeType, ChangeTypeRescheduleMeeting)
	}
	if len(change.AffectedEvents) != 2 {
		t.Errorf("AdaptiveScheduleChange.AffectedEvents length = %d, want 2", len(change.AffectedEvents))
	}
	if change.Impact.FocusTimeGained != 2.0 {
		t.Errorf("AdaptiveImpact.FocusTimeGained = %f, want 2.0", change.Impact.FocusTimeGained)
	}
}

// =============================================================================
// DurationOptimization Tests
// =============================================================================

func TestDurationOptimization_Creation(t *testing.T) {
	opt := DurationOptimization{
		EventID:             "event-123",
		CurrentDuration:     60,
		RecommendedDuration: 45,
		HistoricalData: DurationStats{
			AverageScheduled: 60,
			AverageActual:    42,
			Variance:         5.0,
			OverrunRate:      0.05,
		},
		TimeSavings:    15,
		Confidence:     90.0,
		Reason:         "Historical data shows meetings typically end early",
		Recommendation: "Consider reducing to 45 minutes",
	}

	if opt.CurrentDuration != 60 {
		t.Errorf("DurationOptimization.CurrentDuration = %d, want 60", opt.CurrentDuration)
	}
	if opt.RecommendedDuration != 45 {
		t.Errorf("DurationOptimization.RecommendedDuration = %d, want 45", opt.RecommendedDuration)
	}
	if opt.TimeSavings != 15 {
		t.Errorf("DurationOptimization.TimeSavings = %d, want 15", opt.TimeSavings)
	}
	if opt.HistoricalData.AverageActual != 42 {
		t.Errorf("HistoricalData.AverageActual = %d, want 42", opt.HistoricalData.AverageActual)
	}
}

// =============================================================================
// FocusTimeBlock Tests
// =============================================================================

func TestFocusTimeBlock_Creation(t *testing.T) {
	block := FocusTimeBlock{
		DayOfWeek: "Tuesday",
		StartTime: "09:00",
		EndTime:   "11:00",
		Duration:  120,
		Score:     95.0,
		Reason:    "Historically high productivity",
		Conflicts: 0,
	}

	if block.DayOfWeek != "Tuesday" {
		t.Errorf("FocusTimeBlock.DayOfWeek = %q, want %q", block.DayOfWeek, "Tuesday")
	}
	if block.Duration != 120 {
		t.Errorf("FocusTimeBlock.Duration = %d, want 120", block.Duration)
	}
	if block.Score != 95.0 {
		t.Errorf("FocusTimeBlock.Score = %f, want 95.0", block.Score)
	}
	if block.Conflicts != 0 {
		t.Errorf("FocusTimeBlock.Conflicts = %d, want 0", block.Conflicts)
	}
}

// =============================================================================
// MeetingMetadata Tests
// =============================================================================

func TestMeetingMetadata_Creation(t *testing.T) {
	now := time.Now()
	meta := MeetingMetadata{
		EventID:           "event-123",
		Priority:          PriorityMedium,
		IsRecurring:       true,
		ParticipantCount:  5,
		HistoricalMoves:   2,
		LastMoved:         now.AddDate(0, 0, -7),
		DeclineRate:       0.10,
		AvgRescheduleLead: 3,
	}

	if meta.Priority != PriorityMedium {
		t.Errorf("MeetingMetadata.Priority = %v, want %v", meta.Priority, PriorityMedium)
	}
	if !meta.IsRecurring {
		t.Error("MeetingMetadata.IsRecurring should be true")
	}
	if meta.ParticipantCount != 5 {
		t.Errorf("MeetingMetadata.ParticipantCount = %d, want 5", meta.ParticipantCount)
	}
	if meta.HistoricalMoves != 2 {
		t.Errorf("MeetingMetadata.HistoricalMoves = %d, want 2", meta.HistoricalMoves)
	}
}
