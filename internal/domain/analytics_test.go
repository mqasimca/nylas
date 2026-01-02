package domain

import (
	"testing"
	"time"
)

// =============================================================================
// ConflictType Tests
// =============================================================================

func TestConflictType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		conflict ConflictType
		want     string
	}{
		{"hard conflict", ConflictTypeHard, "hard"},
		{"soft back-to-back", ConflictTypeSoftBackToBack, "soft_back_to_back"},
		{"soft focus time", ConflictTypeSoftFocusTime, "soft_focus_time"},
		{"soft travel time", ConflictTypeSoftTravelTime, "soft_travel_time"},
		{"soft overload", ConflictTypeSoftOverload, "soft_overload"},
		{"soft low priority", ConflictTypeSoftLowPriority, "soft_low_priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.conflict) != tt.want {
				t.Errorf("ConflictType = %q, want %q", string(tt.conflict), tt.want)
			}
		})
	}
}

// =============================================================================
// ConflictSeverity Tests
// =============================================================================

func TestConflictSeverity_Constants(t *testing.T) {
	tests := []struct {
		name     string
		severity ConflictSeverity
		want     string
	}{
		{"critical", SeverityCritical, "critical"},
		{"high", SeverityHigh, "high"},
		{"medium", SeverityMedium, "medium"},
		{"low", SeverityLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.severity) != tt.want {
				t.Errorf("ConflictSeverity = %q, want %q", string(tt.severity), tt.want)
			}
		})
	}
}

// =============================================================================
// MeetingPriority Tests
// =============================================================================

func TestMeetingPriority_Constants(t *testing.T) {
	tests := []struct {
		name     string
		priority MeetingPriority
		want     string
	}{
		{"critical", PriorityCritical, "critical"},
		{"high", PriorityHigh, "high"},
		{"medium", PriorityMedium, "medium"},
		{"low", PriorityLow, "low"},
		{"flexible", PriorityFlexible, "flexible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.priority) != tt.want {
				t.Errorf("MeetingPriority = %q, want %q", string(tt.priority), tt.want)
			}
		})
	}
}

// =============================================================================
// ApprovalStatus Tests
// =============================================================================

func TestApprovalStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status ApprovalStatus
		want   string
	}{
		{"pending", ApprovalPending, "pending"},
		{"approved", ApprovalApproved, "approved"},
		{"denied", ApprovalDenied, "denied"},
		{"expired", ApprovalExpired, "expired"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("ApprovalStatus = %q, want %q", string(tt.status), tt.want)
			}
		})
	}
}

// =============================================================================
// AdaptiveTrigger Tests
// =============================================================================

func TestAdaptiveTrigger_Constants(t *testing.T) {
	tests := []struct {
		name    string
		trigger AdaptiveTrigger
		want    string
	}{
		{"deadline change", TriggerDeadlineChange, "deadline_change"},
		{"meeting overload", TriggerMeetingOverload, "meeting_overload"},
		{"priority shift", TriggerPriorityShift, "priority_shift"},
		{"focus time at risk", TriggerFocusTimeAtRisk, "focus_time_at_risk"},
		{"conflict detected", TriggerConflictDetected, "conflict_detected"},
		{"pattern detected", TriggerPatternDetected, "pattern_detected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.trigger) != tt.want {
				t.Errorf("AdaptiveTrigger = %q, want %q", string(tt.trigger), tt.want)
			}
		})
	}
}

// =============================================================================
// AdaptiveChangeType Tests
// =============================================================================

func TestAdaptiveChangeType_Constants(t *testing.T) {
	tests := []struct {
		name       string
		changeType AdaptiveChangeType
		want       string
	}{
		{"increase focus time", ChangeTypeIncreaseFocusTime, "increase_focus_time"},
		{"reschedule meeting", ChangeTypeRescheduleMeeting, "reschedule_meeting"},
		{"shorten meeting", ChangeTypeShortenMeeting, "shorten_meeting"},
		{"decline meeting", ChangeTypeDeclineMeeting, "decline_meeting"},
		{"move meeting later", ChangeTypeMoveMeetingLater, "move_meeting_later"},
		{"protect block", ChangeTypeProtectBlock, "protect_block"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.changeType) != tt.want {
				t.Errorf("AdaptiveChangeType = %q, want %q", string(tt.changeType), tt.want)
			}
		})
	}
}

// =============================================================================
// MeetingPattern Tests
// =============================================================================

func TestMeetingPattern_Creation(t *testing.T) {
	now := time.Now()
	pattern := MeetingPattern{
		UserEmail: "user@example.com",
		AnalyzedPeriod: DateRange{
			Start: now.AddDate(0, -1, 0),
			End:   now,
		},
		LastUpdated: now,
		Acceptance: AcceptancePatterns{
			ByDayOfWeek: map[string]float64{
				"Monday":  0.92,
				"Tuesday": 0.88,
			},
			ByTimeOfDay: map[string]float64{
				"09:00": 0.85,
				"14:00": 0.78,
			},
			Overall: 0.85,
		},
		Duration: DurationPatterns{
			Overall: DurationStats{
				AverageScheduled: 45,
				AverageActual:    50,
				Variance:         5.5,
				OverrunRate:      0.15,
			},
		},
		Productivity: ProductivityPatterns{
			PeakFocus: []TimeBlock{
				{DayOfWeek: "Tuesday", StartTime: "09:00", EndTime: "11:00", Score: 95.0},
			},
			MeetingDensity: map[string]float64{
				"Monday": 3.5,
				"Friday": 2.0,
			},
		},
		Participants: map[string]ParticipantPattern{
			"colleague@example.com": {
				Email:          "colleague@example.com",
				MeetingCount:   10,
				AcceptanceRate: 0.9,
				PreferredDays:  []string{"Monday", "Wednesday"},
			},
		},
	}

	if pattern.UserEmail != "user@example.com" {
		t.Errorf("MeetingPattern.UserEmail = %q, want %q", pattern.UserEmail, "user@example.com")
	}
	if pattern.Acceptance.Overall != 0.85 {
		t.Errorf("Acceptance.Overall = %f, want 0.85", pattern.Acceptance.Overall)
	}
	if pattern.Duration.Overall.AverageScheduled != 45 {
		t.Errorf("Duration.Overall.AverageScheduled = %d, want 45", pattern.Duration.Overall.AverageScheduled)
	}
	if len(pattern.Productivity.PeakFocus) != 1 {
		t.Errorf("Productivity.PeakFocus length = %d, want 1", len(pattern.Productivity.PeakFocus))
	}
	if _, ok := pattern.Participants["colleague@example.com"]; !ok {
		t.Error("Participants should contain colleague@example.com")
	}
}

// =============================================================================
// TimeBlock Tests
// =============================================================================

func TestTimeBlock_Creation(t *testing.T) {
	block := TimeBlock{
		DayOfWeek: "Wednesday",
		StartTime: "09:00",
		EndTime:   "12:00",
		Score:     92.5,
	}

	if block.DayOfWeek != "Wednesday" {
		t.Errorf("TimeBlock.DayOfWeek = %q, want %q", block.DayOfWeek, "Wednesday")
	}
	if block.StartTime != "09:00" {
		t.Errorf("TimeBlock.StartTime = %q, want %q", block.StartTime, "09:00")
	}
	if block.Score != 92.5 {
		t.Errorf("TimeBlock.Score = %f, want 92.5", block.Score)
	}
}

// =============================================================================
// DurationStats Tests
// =============================================================================

func TestDurationStats_Creation(t *testing.T) {
	stats := DurationStats{
		AverageScheduled: 30,
		AverageActual:    35,
		Variance:         7.5,
		OverrunRate:      0.20,
	}

	if stats.AverageScheduled != 30 {
		t.Errorf("DurationStats.AverageScheduled = %d, want 30", stats.AverageScheduled)
	}
	if stats.AverageActual != 35 {
		t.Errorf("DurationStats.AverageActual = %d, want 35", stats.AverageActual)
	}
	if stats.OverrunRate != 0.20 {
		t.Errorf("DurationStats.OverrunRate = %f, want 0.20", stats.OverrunRate)
	}
}

// =============================================================================
// Recommendation Tests
// =============================================================================

func TestRecommendation_Creation(t *testing.T) {
	rec := Recommendation{
		Type:        "focus_time",
		Priority:    "high",
		Title:       "Block morning focus time",
		Description: "Based on your patterns, mornings are your most productive time.",
		Confidence:  85.0,
		Action:      "Create recurring focus block 9-11 AM",
		Impact:      "Estimated 2 hours more focus time per week",
	}

	if rec.Type != "focus_time" {
		t.Errorf("Recommendation.Type = %q, want %q", rec.Type, "focus_time")
	}
	if rec.Priority != "high" {
		t.Errorf("Recommendation.Priority = %q, want %q", rec.Priority, "high")
	}
	if rec.Confidence != 85.0 {
		t.Errorf("Recommendation.Confidence = %f, want 85.0", rec.Confidence)
	}
}

// =============================================================================
// MeetingScore Tests
// =============================================================================

func TestMeetingScore_Creation(t *testing.T) {
	score := MeetingScore{
		Score:       75,
		Confidence:  80.0,
		SuccessRate: 0.85,
		Factors: []ScoreFactor{
			{Name: "Time of day", Impact: 10, Description: "Optimal afternoon slot"},
			{Name: "Participant availability", Impact: -5, Description: "One participant often declines"},
		},
		Recommendation: "Consider scheduling earlier in the week",
	}

	if score.Score != 75 {
		t.Errorf("MeetingScore.Score = %d, want 75", score.Score)
	}
	if len(score.Factors) != 2 {
		t.Errorf("MeetingScore.Factors length = %d, want 2", len(score.Factors))
	}
	if score.Factors[0].Impact != 10 {
		t.Errorf("ScoreFactor.Impact = %d, want 10", score.Factors[0].Impact)
	}
}

// =============================================================================
// Conflict Tests
// =============================================================================

func TestConflict_Creation(t *testing.T) {
	conflict := Conflict{
		ID:               "conflict-123",
		Type:             ConflictTypeHard,
		Severity:         SeverityCritical,
		ProposedEvent:    &Event{ID: "event-1", Title: "Team Meeting"},
		ConflictingEvent: &Event{ID: "event-2", Title: "1:1 with Manager"},
		Description:      "Both events overlap from 2-3 PM",
		Impact:           "Cannot attend both meetings",
		Suggestion:       "Move team meeting to 3 PM",
		CanAutoResolve:   false,
	}

	if conflict.Type != ConflictTypeHard {
		t.Errorf("Conflict.Type = %v, want %v", conflict.Type, ConflictTypeHard)
	}
	if conflict.Severity != SeverityCritical {
		t.Errorf("Conflict.Severity = %v, want %v", conflict.Severity, SeverityCritical)
	}
	if conflict.ProposedEvent == nil {
		t.Error("Conflict.ProposedEvent should not be nil")
	}
	if conflict.CanAutoResolve {
		t.Error("Conflict.CanAutoResolve should be false")
	}
}

// =============================================================================
// ConflictAnalysis Tests
// =============================================================================

func TestConflictAnalysis_Creation(t *testing.T) {
	analysis := ConflictAnalysis{
		ProposedEvent: &Event{ID: "event-1", Title: "New Meeting"},
		HardConflicts: []Conflict{
			{ID: "hard-1", Type: ConflictTypeHard},
		},
		SoftConflicts: []Conflict{
			{ID: "soft-1", Type: ConflictTypeSoftBackToBack},
			{ID: "soft-2", Type: ConflictTypeSoftFocusTime},
		},
		TotalConflicts:   3,
		CanProceed:       false,
		Recommendations:  []string{"Reschedule to avoid overlap", "Add buffer time"},
		AIRecommendation: "Consider moving to Thursday afternoon",
	}

	if len(analysis.HardConflicts) != 1 {
		t.Errorf("ConflictAnalysis.HardConflicts length = %d, want 1", len(analysis.HardConflicts))
	}
	if len(analysis.SoftConflicts) != 2 {
		t.Errorf("ConflictAnalysis.SoftConflicts length = %d, want 2", len(analysis.SoftConflicts))
	}
	if analysis.TotalConflicts != 3 {
		t.Errorf("ConflictAnalysis.TotalConflicts = %d, want 3", analysis.TotalConflicts)
	}
	if analysis.CanProceed {
		t.Error("ConflictAnalysis.CanProceed should be false")
	}
}

// =============================================================================
// RescheduleOption Tests
// =============================================================================

func TestRescheduleOption_Creation(t *testing.T) {
	now := time.Now()
	option := RescheduleOption{
		ProposedTime:     now.Add(24 * time.Hour),
		EndTime:          now.Add(25 * time.Hour),
		Score:            85,
		Confidence:       90.0,
		Pros:             []string{"No conflicts", "Good time for all participants"},
		Cons:             []string{"Less notice than usual"},
		ParticipantMatch: 1.0,
		AIInsight:        "This time works well based on historical patterns",
	}

	if option.Score != 85 {
		t.Errorf("RescheduleOption.Score = %d, want 85", option.Score)
	}
	if option.Confidence != 90.0 {
		t.Errorf("RescheduleOption.Confidence = %f, want 90.0", option.Confidence)
	}
	if len(option.Pros) != 2 {
		t.Errorf("RescheduleOption.Pros length = %d, want 2", len(option.Pros))
	}
	if option.ParticipantMatch != 1.0 {
		t.Errorf("RescheduleOption.ParticipantMatch = %f, want 1.0", option.ParticipantMatch)
	}
}

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
