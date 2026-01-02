package domain

import (
	"testing"
	"time"
)

// =============================================================================
// DateRange Tests
// =============================================================================

func TestDateRange_Creation(t *testing.T) {
	now := time.Now()
	dr := DateRange{
		Start: now.AddDate(0, 0, -7),
		End:   now,
	}

	if dr.Start.IsZero() {
		t.Error("DateRange.Start should not be zero")
	}
	if dr.End.IsZero() {
		t.Error("DateRange.End should not be zero")
	}
	if dr.Start.After(dr.End) {
		t.Error("DateRange.Start should be before End")
	}
}

// =============================================================================
// MeetingFinderRequest Tests
// =============================================================================

func TestMeetingFinderRequest_Creation(t *testing.T) {
	now := time.Now()
	req := MeetingFinderRequest{
		TimeZones:         []string{"America/New_York", "Europe/London", "Asia/Tokyo"},
		Duration:          time.Hour,
		WorkingHoursStart: "09:00",
		WorkingHoursEnd:   "17:00",
		DateRange: DateRange{
			Start: now,
			End:   now.AddDate(0, 0, 7),
		},
		ExcludeWeekends: true,
	}

	if len(req.TimeZones) != 3 {
		t.Errorf("MeetingFinderRequest.TimeZones length = %d, want 3", len(req.TimeZones))
	}
	if req.Duration != time.Hour {
		t.Errorf("MeetingFinderRequest.Duration = %v, want 1h", req.Duration)
	}
	if req.WorkingHoursStart != "09:00" {
		t.Errorf("MeetingFinderRequest.WorkingHoursStart = %q, want %q", req.WorkingHoursStart, "09:00")
	}
	if !req.ExcludeWeekends {
		t.Error("MeetingFinderRequest.ExcludeWeekends should be true")
	}
}

// =============================================================================
// MeetingTimeSlots Tests
// =============================================================================

func TestMeetingTimeSlots_Creation(t *testing.T) {
	now := time.Now()
	slots := MeetingTimeSlots{
		Slots: []MeetingSlot{
			{
				StartTime: now,
				EndTime:   now.Add(time.Hour),
				Times: map[string]time.Time{
					"America/New_York": now,
					"Europe/London":    now.Add(5 * time.Hour),
				},
				Score: 0.95,
			},
		},
		TimeZones:  []string{"America/New_York", "Europe/London"},
		TotalSlots: 1,
	}

	if len(slots.Slots) != 1 {
		t.Errorf("MeetingTimeSlots.Slots length = %d, want 1", len(slots.Slots))
	}
	if slots.TotalSlots != 1 {
		t.Errorf("MeetingTimeSlots.TotalSlots = %d, want 1", slots.TotalSlots)
	}
	if slots.Slots[0].Score != 0.95 {
		t.Errorf("MeetingSlot.Score = %f, want 0.95", slots.Slots[0].Score)
	}
}

// =============================================================================
// MeetingSlot Tests
// =============================================================================

func TestMeetingSlot_Creation(t *testing.T) {
	now := time.Now()
	slot := MeetingSlot{
		StartTime: now,
		EndTime:   now.Add(30 * time.Minute),
		Times: map[string]time.Time{
			"UTC":              now,
			"America/New_York": now.Add(-5 * time.Hour),
		},
		Score: 0.85,
	}

	if slot.Score != 0.85 {
		t.Errorf("MeetingSlot.Score = %f, want 0.85", slot.Score)
	}
	if len(slot.Times) != 2 {
		t.Errorf("MeetingSlot.Times length = %d, want 2", len(slot.Times))
	}
}

// =============================================================================
// DSTTransition Tests
// =============================================================================

func TestDSTTransition_Creation(t *testing.T) {
	tests := []struct {
		name       string
		transition DSTTransition
	}{
		{
			name: "spring forward",
			transition: DSTTransition{
				Date:      time.Date(2024, 3, 10, 2, 0, 0, 0, time.UTC),
				Offset:    -7 * 3600,
				Name:      "PDT",
				IsDST:     true,
				Direction: "forward",
			},
		},
		{
			name: "fall back",
			transition: DSTTransition{
				Date:      time.Date(2024, 11, 3, 2, 0, 0, 0, time.UTC),
				Offset:    -8 * 3600,
				Name:      "PST",
				IsDST:     false,
				Direction: "backward",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.transition.Name == "" {
				t.Error("DSTTransition.Name should not be empty")
			}
			if tt.transition.Direction != "forward" && tt.transition.Direction != "backward" {
				t.Errorf("DSTTransition.Direction = %q, want 'forward' or 'backward'", tt.transition.Direction)
			}
		})
	}
}

// =============================================================================
// TimeZoneInfo Tests
// =============================================================================

func TestTimeZoneInfo_Creation(t *testing.T) {
	nextDST := time.Date(2024, 3, 10, 2, 0, 0, 0, time.UTC)
	info := TimeZoneInfo{
		Name:         "America/Los_Angeles",
		Abbreviation: "PST",
		Offset:       -8 * 3600,
		IsDST:        false,
		NextDST:      &nextDST,
	}

	if info.Name != "America/Los_Angeles" {
		t.Errorf("TimeZoneInfo.Name = %q, want %q", info.Name, "America/Los_Angeles")
	}
	if info.Abbreviation != "PST" {
		t.Errorf("TimeZoneInfo.Abbreviation = %q, want %q", info.Abbreviation, "PST")
	}
	if info.Offset != -8*3600 {
		t.Errorf("TimeZoneInfo.Offset = %d, want %d", info.Offset, -8*3600)
	}
	if info.IsDST {
		t.Error("TimeZoneInfo.IsDST should be false")
	}
	if info.NextDST == nil {
		t.Error("TimeZoneInfo.NextDST should not be nil")
	}
}

// =============================================================================
// DSTWarning Tests
// =============================================================================

func TestDSTWarning_Creation(t *testing.T) {
	tests := []struct {
		name    string
		warning DSTWarning
	}{
		{
			name: "near transition warning",
			warning: DSTWarning{
				IsNearTransition: true,
				TransitionDate:   time.Date(2024, 3, 10, 2, 0, 0, 0, time.UTC),
				Direction:        "forward",
				DaysUntil:        3,
				TransitionName:   "PDT",
				InTransitionGap:  false,
				InDuplicateHour:  false,
				Warning:          "DST transition in 3 days",
				Severity:         "warning",
			},
		},
		{
			name: "in transition gap error",
			warning: DSTWarning{
				IsNearTransition: true,
				TransitionDate:   time.Date(2024, 3, 10, 2, 0, 0, 0, time.UTC),
				Direction:        "forward",
				DaysUntil:        0,
				InTransitionGap:  true,
				InDuplicateHour:  false,
				Warning:          "Time does not exist (spring forward gap)",
				Severity:         "error",
			},
		},
		{
			name: "in duplicate hour warning",
			warning: DSTWarning{
				IsNearTransition: true,
				TransitionDate:   time.Date(2024, 11, 3, 1, 0, 0, 0, time.UTC),
				Direction:        "backward",
				DaysUntil:        0,
				InTransitionGap:  false,
				InDuplicateHour:  true,
				Warning:          "Time occurs twice (fall back)",
				Severity:         "warning",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.warning.Warning == "" {
				t.Error("DSTWarning.Warning should not be empty")
			}
			if tt.warning.Severity != "error" && tt.warning.Severity != "warning" && tt.warning.Severity != "info" {
				t.Errorf("DSTWarning.Severity = %q, want 'error', 'warning', or 'info'", tt.warning.Severity)
			}
		})
	}
}

// =============================================================================
// WebhookServerConfig Tests
// =============================================================================

func TestWebhookServerConfig_Creation(t *testing.T) {
	config := WebhookServerConfig{
		Port:              8080,
		Host:              "localhost",
		PersistentURL:     "https://webhook.example.com",
		SaveToFile:        true,
		FilePath:          "/tmp/webhooks.json",
		ValidateSignature: true,
		Secret:            "webhook-secret",
		Headers: map[string]string{
			"X-Custom-Header": "value",
		},
	}

	if config.Port != 8080 {
		t.Errorf("WebhookServerConfig.Port = %d, want 8080", config.Port)
	}
	if config.Host != "localhost" {
		t.Errorf("WebhookServerConfig.Host = %q, want %q", config.Host, "localhost")
	}
	if !config.SaveToFile {
		t.Error("WebhookServerConfig.SaveToFile should be true")
	}
	if !config.ValidateSignature {
		t.Error("WebhookServerConfig.ValidateSignature should be true")
	}
}

// =============================================================================
// WebhookPayload Tests
// =============================================================================

func TestWebhookPayload_Creation(t *testing.T) {
	now := time.Now()
	payload := WebhookPayload{
		ID:        "payload-123",
		Timestamp: now,
		Method:    "POST",
		URL:       "https://webhook.example.com/endpoint",
		Headers: map[string]string{
			"Content-Type":      "application/json",
			"X-Nylas-Signature": "abc123",
		},
		Body:      []byte(`{"type": "message.created"}`),
		Signature: "abc123",
		Verified:  true,
	}

	if payload.Method != "POST" {
		t.Errorf("WebhookPayload.Method = %q, want %q", payload.Method, "POST")
	}
	if !payload.Verified {
		t.Error("WebhookPayload.Verified should be true")
	}
	if len(payload.Body) == 0 {
		t.Error("WebhookPayload.Body should not be empty")
	}
	if len(payload.Headers) != 2 {
		t.Errorf("WebhookPayload.Headers length = %d, want 2", len(payload.Headers))
	}
}

// =============================================================================
// TemplateRequest Tests
// =============================================================================

func TestTemplateRequest_Creation(t *testing.T) {
	req := TemplateRequest{
		Name:      "welcome-email",
		Subject:   "Welcome to {{company_name}}!",
		HTMLBody:  "<h1>Welcome, {{user_name}}!</h1>",
		TextBody:  "Welcome, {{user_name}}!",
		Variables: []string{"company_name", "user_name"},
		InlineCSS: true,
		Sanitize:  true,
		Metadata: map[string]string{
			"category": "onboarding",
		},
	}

	if req.Name != "welcome-email" {
		t.Errorf("TemplateRequest.Name = %q, want %q", req.Name, "welcome-email")
	}
	if len(req.Variables) != 2 {
		t.Errorf("TemplateRequest.Variables length = %d, want 2", len(req.Variables))
	}
	if !req.InlineCSS {
		t.Error("TemplateRequest.InlineCSS should be true")
	}
	if !req.Sanitize {
		t.Error("TemplateRequest.Sanitize should be true")
	}
}

// =============================================================================
// EmailTemplate Tests
// =============================================================================

func TestEmailTemplate_Creation(t *testing.T) {
	now := time.Now()
	template := EmailTemplate{
		ID:        "template-123",
		Name:      "newsletter",
		Subject:   "Weekly Newsletter",
		HTMLBody:  "<html><body>Newsletter content</body></html>",
		TextBody:  "Newsletter content",
		Variables: []string{"subscriber_name", "unsubscribe_link"},
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
		Metadata: map[string]string{
			"frequency": "weekly",
		},
	}

	if template.Name != "newsletter" {
		t.Errorf("EmailTemplate.Name = %q, want %q", template.Name, "newsletter")
	}
	if len(template.Variables) != 2 {
		t.Errorf("EmailTemplate.Variables length = %d, want 2", len(template.Variables))
	}
}

// =============================================================================
// DeliverabilityReport Tests
// =============================================================================

func TestDeliverabilityReport_Creation(t *testing.T) {
	report := DeliverabilityReport{
		Score: 85,
		Issues: []DeliverabilityIssue{
			{
				Severity: "warning",
				Category: "content",
				Message:  "Subject line may trigger spam filters",
				Fix:      "Avoid using all caps in subject",
			},
		},
		SPFStatus:       "pass",
		DKIMStatus:      "pass",
		DMARCStatus:     "pass",
		SpamScore:       2.5,
		MobileOptimized: true,
		Recommendations: []string{"Add preheader text", "Include plain text version"},
	}

	if report.Score != 85 {
		t.Errorf("DeliverabilityReport.Score = %d, want 85", report.Score)
	}
	if report.SPFStatus != "pass" {
		t.Errorf("DeliverabilityReport.SPFStatus = %q, want %q", report.SPFStatus, "pass")
	}
	if len(report.Issues) != 1 {
		t.Errorf("DeliverabilityReport.Issues length = %d, want 1", len(report.Issues))
	}
	if !report.MobileOptimized {
		t.Error("DeliverabilityReport.MobileOptimized should be true")
	}
}

// =============================================================================
// DeliverabilityIssue Tests
// =============================================================================

func TestDeliverabilityIssue_Creation(t *testing.T) {
	tests := []struct {
		name  string
		issue DeliverabilityIssue
	}{
		{
			name: "critical issue",
			issue: DeliverabilityIssue{
				Severity: "critical",
				Category: "authentication",
				Message:  "DKIM signature invalid",
				Fix:      "Verify DKIM configuration",
			},
		},
		{
			name: "warning issue",
			issue: DeliverabilityIssue{
				Severity: "warning",
				Category: "content",
				Message:  "Image to text ratio too high",
				Fix:      "Add more text content",
			},
		},
		{
			name: "info issue",
			issue: DeliverabilityIssue{
				Severity: "info",
				Category: "formatting",
				Message:  "Consider adding alt text to images",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issue.Severity == "" {
				t.Error("DeliverabilityIssue.Severity should not be empty")
			}
			if tt.issue.Category == "" {
				t.Error("DeliverabilityIssue.Category should not be empty")
			}
		})
	}
}

// =============================================================================
// ParsedEmail Tests
// =============================================================================

func TestParsedEmail_Creation(t *testing.T) {
	now := time.Now()
	parsed := ParsedEmail{
		Headers: map[string]string{
			"Message-ID": "<msg-123@example.com>",
			"Date":       now.Format(time.RFC1123),
		},
		From:     "sender@example.com",
		To:       []string{"recipient1@example.com", "recipient2@example.com"},
		Cc:       []string{"cc@example.com"},
		Subject:  "Test Email",
		Date:     now,
		HTMLBody: "<p>HTML content</p>",
		TextBody: "Text content",
		Attachments: []Attachment{
			{Filename: "file.pdf", Size: 1024},
		},
	}

	if parsed.From != "sender@example.com" {
		t.Errorf("ParsedEmail.From = %q, want %q", parsed.From, "sender@example.com")
	}
	if len(parsed.To) != 2 {
		t.Errorf("ParsedEmail.To length = %d, want 2", len(parsed.To))
	}
	if len(parsed.Headers) != 2 {
		t.Errorf("ParsedEmail.Headers length = %d, want 2", len(parsed.Headers))
	}
}

// =============================================================================
// EmailValidation Tests
// =============================================================================

func TestEmailValidation_Creation(t *testing.T) {
	tests := []struct {
		name       string
		validation EmailValidation
	}{
		{
			name: "valid email",
			validation: EmailValidation{
				Email:       "user@example.com",
				Valid:       true,
				FormatValid: true,
				MXExists:    true,
				Disposable:  false,
			},
		},
		{
			name: "invalid format",
			validation: EmailValidation{
				Email:       "not-an-email",
				Valid:       false,
				FormatValid: false,
				MXExists:    false,
				Disposable:  false,
			},
		},
		{
			name: "disposable email with suggestion",
			validation: EmailValidation{
				Email:       "user@tempmail.com",
				Valid:       false,
				FormatValid: true,
				MXExists:    true,
				Disposable:  true,
				Suggestion:  "Please use a non-disposable email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.validation.Email == "" {
				t.Error("EmailValidation.Email should not be empty")
			}
		})
	}
}

// =============================================================================
// SpamAnalysis Tests
// =============================================================================

func TestSpamAnalysis_Creation(t *testing.T) {
	analysis := SpamAnalysis{
		Score:  3.5,
		IsSpam: false,
		Triggers: []SpamTrigger{
			{
				Rule:        "CAPS_SUBJECT",
				Description: "Subject contains excessive capital letters",
				Score:       1.5,
				Severity:    "medium",
			},
			{
				Rule:        "URGENT_LANGUAGE",
				Description: "Message contains urgent language",
				Score:       2.0,
				Severity:    "high",
			},
		},
		Passed:      []string{"SPF_PASS", "DKIM_VALID", "DMARC_PASS"},
		Suggestions: []string{"Reduce urgency language", "Use mixed case in subject"},
	}

	if analysis.Score != 3.5 {
		t.Errorf("SpamAnalysis.Score = %f, want 3.5", analysis.Score)
	}
	if analysis.IsSpam {
		t.Error("SpamAnalysis.IsSpam should be false")
	}
	if len(analysis.Triggers) != 2 {
		t.Errorf("SpamAnalysis.Triggers length = %d, want 2", len(analysis.Triggers))
	}
	if len(analysis.Passed) != 3 {
		t.Errorf("SpamAnalysis.Passed length = %d, want 3", len(analysis.Passed))
	}
}

// =============================================================================
// DeduplicationRequest Tests
// =============================================================================

func TestDeduplicationRequest_Creation(t *testing.T) {
	req := DeduplicationRequest{
		Contacts: []Contact{
			{ID: "contact-1", GivenName: "John", Surname: "Doe"},
			{ID: "contact-2", GivenName: "John", Surname: "Doe"},
		},
		FuzzyThreshold: 0.85,
		MatchFields:    []string{"email", "phone", "name"},
		AutoMerge:      true,
		MergeStrategy:  "most_complete",
	}

	if len(req.Contacts) != 2 {
		t.Errorf("DeduplicationRequest.Contacts length = %d, want 2", len(req.Contacts))
	}
	if req.FuzzyThreshold != 0.85 {
		t.Errorf("DeduplicationRequest.FuzzyThreshold = %f, want 0.85", req.FuzzyThreshold)
	}
	if !req.AutoMerge {
		t.Error("DeduplicationRequest.AutoMerge should be true")
	}
	if req.MergeStrategy != "most_complete" {
		t.Errorf("DeduplicationRequest.MergeStrategy = %q, want %q", req.MergeStrategy, "most_complete")
	}
}

// =============================================================================
// DeduplicationResult Tests
// =============================================================================

func TestDeduplicationResult_Creation(t *testing.T) {
	result := DeduplicationResult{
		OriginalCount:     100,
		DeduplicatedCount: 85,
		DuplicateGroups: []DuplicateGroup{
			{
				Contacts: []Contact{
					{ID: "contact-1", GivenName: "John"},
					{ID: "contact-2", GivenName: "John"},
				},
				MatchScore:    0.95,
				MatchedFields: []string{"email", "name"},
				Suggested:     &Contact{ID: "merged-1", GivenName: "John"},
			},
		},
		MergedContacts: []Contact{
			{ID: "merged-1", GivenName: "John"},
		},
	}

	if result.OriginalCount != 100 {
		t.Errorf("DeduplicationResult.OriginalCount = %d, want 100", result.OriginalCount)
	}
	if result.DeduplicatedCount != 85 {
		t.Errorf("DeduplicationResult.DeduplicatedCount = %d, want 85", result.DeduplicatedCount)
	}
	if len(result.DuplicateGroups) != 1 {
		t.Errorf("DeduplicationResult.DuplicateGroups length = %d, want 1", len(result.DuplicateGroups))
	}
	if result.DuplicateGroups[0].MatchScore != 0.95 {
		t.Errorf("DuplicateGroup.MatchScore = %f, want 0.95", result.DuplicateGroups[0].MatchScore)
	}
}
