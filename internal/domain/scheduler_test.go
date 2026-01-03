package domain

import (
	"testing"
	"time"
)

// =============================================================================
// SchedulerConfiguration Tests
// =============================================================================

func TestSchedulerConfiguration_Creation(t *testing.T) {
	now := time.Now()
	config := SchedulerConfiguration{
		ID:                  "config-123",
		Name:                "30-Minute Meeting",
		Slug:                "30-min-meeting",
		RequiresSessionAuth: true,
		Participants: []ConfigurationParticipant{
			{
				Email:       "host@example.com",
				Name:        "Host User",
				IsOrganizer: true,
				Availability: ConfigurationAvailability{
					CalendarIDs: []string{"cal-primary"},
					OpenHours: []OpenHours{
						{
							Days:     []int{1, 2, 3, 4, 5},
							Start:    "09:00",
							End:      "17:00",
							Timezone: "America/New_York",
						},
					},
				},
				Booking: &ParticipantBooking{
					CalendarID: "cal-primary",
				},
			},
		},
		Availability: AvailabilityRules{
			DurationMinutes:    30,
			IntervalMinutes:    15,
			RoundTo:            15,
			AvailabilityMethod: "max-availability",
			Buffer: &AvailabilityBuffer{
				Before: 5,
				After:  5,
			},
		},
		EventBooking: EventBooking{
			Title:           "Meeting with {{guest_name}}",
			Description:     "A 30-minute meeting",
			Location:        "Video call",
			Timezone:        "America/New_York",
			BookingType:     "booking",
			DisableEmails:   false,
			ReminderMinutes: []int{15, 60},
		},
		Scheduler: SchedulerSettings{
			AvailableDaysInFuture: 30,
			MinBookingNotice:      60,
			MinCancellationNotice: 60,
			ConfirmationMethod:    "automatic",
		},
		AppearanceSettings: &AppearanceSettings{
			CompanyName:     "Acme Corp",
			Color:           "#4285f4",
			SubmitText:      "Book Meeting",
			ThankYouMessage: "Thanks for booking!",
		},
		CreatedAt:  &now,
		ModifiedAt: &now,
	}

	if config.Name != "30-Minute Meeting" {
		t.Errorf("SchedulerConfiguration.Name = %q, want %q", config.Name, "30-Minute Meeting")
	}
	if config.Slug != "30-min-meeting" {
		t.Errorf("SchedulerConfiguration.Slug = %q, want %q", config.Slug, "30-min-meeting")
	}
	if !config.RequiresSessionAuth {
		t.Error("SchedulerConfiguration.RequiresSessionAuth should be true")
	}
	if len(config.Participants) != 1 {
		t.Errorf("SchedulerConfiguration.Participants length = %d, want 1", len(config.Participants))
	}
	if config.Availability.DurationMinutes != 30 {
		t.Errorf("AvailabilityRules.DurationMinutes = %d, want 30", config.Availability.DurationMinutes)
	}
}

// =============================================================================
// ConfigurationParticipant Tests
// =============================================================================

func TestConfigurationParticipant_Creation(t *testing.T) {
	participant := ConfigurationParticipant{
		Email:       "participant@example.com",
		Name:        "Participant Name",
		IsOrganizer: false,
		Availability: ConfigurationAvailability{
			CalendarIDs: []string{"cal-1", "cal-2"},
			OpenHours: []OpenHours{
				{
					Days:  []int{1, 2, 3, 4, 5},
					Start: "08:00",
					End:   "18:00",
				},
			},
		},
		Booking: &ParticipantBooking{
			CalendarID: "cal-1",
		},
	}

	if participant.Email != "participant@example.com" {
		t.Errorf("ConfigurationParticipant.Email = %q, want %q", participant.Email, "participant@example.com")
	}
	if participant.IsOrganizer {
		t.Error("ConfigurationParticipant.IsOrganizer should be false")
	}
	if len(participant.Availability.CalendarIDs) != 2 {
		t.Errorf("ConfigurationAvailability.CalendarIDs length = %d, want 2", len(participant.Availability.CalendarIDs))
	}
}

// =============================================================================
// OpenHours Tests
// =============================================================================

func TestOpenHours_Creation(t *testing.T) {
	tests := []struct {
		name      string
		openHours OpenHours
	}{
		{
			name: "weekday hours",
			openHours: OpenHours{
				Days:     []int{1, 2, 3, 4, 5},
				Start:    "09:00",
				End:      "17:00",
				Timezone: "America/Los_Angeles",
			},
		},
		{
			name: "weekend hours with excluded dates",
			openHours: OpenHours{
				Days:     []int{0, 6},
				Start:    "10:00",
				End:      "14:00",
				Timezone: "America/New_York",
				ExDates:  []string{"2024-12-25", "2024-01-01"},
			},
		},
		{
			name: "split day hours",
			openHours: OpenHours{
				Days:  []int{1, 3, 5},
				Start: "14:00",
				End:   "20:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.openHours.Days) == 0 {
				t.Error("OpenHours.Days should not be empty")
			}
			if tt.openHours.Start == "" {
				t.Error("OpenHours.Start should not be empty")
			}
			if tt.openHours.End == "" {
				t.Error("OpenHours.End should not be empty")
			}
		})
	}
}

// =============================================================================
// AvailabilityRules Tests
// =============================================================================

func TestAvailabilityRules_Creation(t *testing.T) {
	rules := AvailabilityRules{
		DurationMinutes:    45,
		IntervalMinutes:    15,
		RoundTo:            30,
		AvailabilityMethod: "max-fairness",
		Buffer: &AvailabilityBuffer{
			Before: 10,
			After:  5,
		},
	}

	if rules.DurationMinutes != 45 {
		t.Errorf("AvailabilityRules.DurationMinutes = %d, want 45", rules.DurationMinutes)
	}
	if rules.AvailabilityMethod != "max-fairness" {
		t.Errorf("AvailabilityRules.AvailabilityMethod = %q, want %q", rules.AvailabilityMethod, "max-fairness")
	}
	if rules.Buffer == nil {
		t.Fatal("AvailabilityRules.Buffer should not be nil")
	}
	if rules.Buffer.Before != 10 {
		t.Errorf("AvailabilityBuffer.Before = %d, want 10", rules.Buffer.Before)
	}
}

// =============================================================================
// AvailabilityBuffer Tests
// =============================================================================

func TestAvailabilityBuffer_Creation(t *testing.T) {
	buffer := AvailabilityBuffer{
		Before: 15,
		After:  10,
	}

	if buffer.Before != 15 {
		t.Errorf("AvailabilityBuffer.Before = %d, want 15", buffer.Before)
	}
	if buffer.After != 10 {
		t.Errorf("AvailabilityBuffer.After = %d, want 10", buffer.After)
	}
}

// =============================================================================
// EventBooking Tests
// =============================================================================

func TestEventBooking_Creation(t *testing.T) {
	booking := EventBooking{
		Title:       "Consultation with {{guest_name}}",
		Description: "A consultation meeting to discuss your needs",
		Location:    "Zoom",
		Timezone:    "Europe/London",
		BookingType: "organizer-confirmation",
		Conferencing: &ConferencingSettings{
			Provider:   "Zoom",
			Autocreate: true,
		},
		DisableEmails:   false,
		ReminderMinutes: []int{10, 30, 1440},
		Metadata: map[string]string{
			"booking_type": "consultation",
		},
	}

	if booking.Title == "" {
		t.Error("EventBooking.Title should not be empty")
	}
	if booking.BookingType != "organizer-confirmation" {
		t.Errorf("EventBooking.BookingType = %q, want %q", booking.BookingType, "organizer-confirmation")
	}
	if booking.Conferencing == nil {
		t.Fatal("EventBooking.Conferencing should not be nil")
	}
	if booking.Conferencing.Provider != "Zoom" {
		t.Errorf("ConferencingSettings.Provider = %q, want %q", booking.Conferencing.Provider, "Zoom")
	}
	if len(booking.ReminderMinutes) != 3 {
		t.Errorf("EventBooking.ReminderMinutes length = %d, want 3", len(booking.ReminderMinutes))
	}
}

// =============================================================================
// ConferencingSettings Tests
// =============================================================================

func TestConferencingSettings_Creation(t *testing.T) {
	tests := []struct {
		name     string
		settings ConferencingSettings
	}{
		{
			name: "Google Meet auto-create",
			settings: ConferencingSettings{
				Provider:   "Google Meet",
				Autocreate: true,
			},
		},
		{
			name: "Zoom with details",
			settings: ConferencingSettings{
				Provider:   "Zoom",
				Autocreate: false,
				Details: &ConferencingDetails{
					URL:         "https://zoom.us/j/123456789",
					MeetingCode: "123456789",
					Password:    "password123",
				},
			},
		},
		{
			name: "Microsoft Teams",
			settings: ConferencingSettings{
				Provider:   "Microsoft Teams",
				Autocreate: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.settings.Provider == "" {
				t.Error("ConferencingSettings.Provider should not be empty")
			}
		})
	}
}

// =============================================================================
// SchedulerSettings Tests
// =============================================================================

func TestSchedulerSettings_Creation(t *testing.T) {
	settings := SchedulerSettings{
		AvailableDaysInFuture: 60,
		MinBookingNotice:      120,
		MinCancellationNotice: 1440,
		ConfirmationMethod:    "manual",
		ReschedulingURL:       "https://scheduler.example.com/reschedule",
		CancellationURL:       "https://scheduler.example.com/cancel",
		AdditionalFields: map[string]any{
			"phone":   "required",
			"company": "optional",
		},
		CancellationPolicy: "24 hours notice required",
	}

	if settings.AvailableDaysInFuture != 60 {
		t.Errorf("SchedulerSettings.AvailableDaysInFuture = %d, want 60", settings.AvailableDaysInFuture)
	}
	if settings.ConfirmationMethod != "manual" {
		t.Errorf("SchedulerSettings.ConfirmationMethod = %q, want %q", settings.ConfirmationMethod, "manual")
	}
	if settings.CancellationPolicy == "" {
		t.Error("SchedulerSettings.CancellationPolicy should not be empty")
	}
}

// =============================================================================
// AppearanceSettings Tests
// =============================================================================

func TestAppearanceSettings_Creation(t *testing.T) {
	appearance := AppearanceSettings{
		CompanyName:     "Tech Startup Inc",
		Logo:            "https://example.com/logo.png",
		Color:           "#00ff00",
		SubmitText:      "Schedule Now",
		ThankYouMessage: "Your meeting has been scheduled!",
	}

	if appearance.CompanyName != "Tech Startup Inc" {
		t.Errorf("AppearanceSettings.CompanyName = %q, want %q", appearance.CompanyName, "Tech Startup Inc")
	}
	if appearance.Color != "#00ff00" {
		t.Errorf("AppearanceSettings.Color = %q, want %q", appearance.Color, "#00ff00")
	}
}

// =============================================================================
// SchedulerSession Tests
// =============================================================================

func TestSchedulerSession_Creation(t *testing.T) {
	now := time.Now()
	session := SchedulerSession{
		SessionID:       "session-123",
		ConfigurationID: "config-456",
		BookingURL:      "https://scheduler.example.com/book/session-123",
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	if session.SessionID != "session-123" {
		t.Errorf("SchedulerSession.SessionID = %q, want %q", session.SessionID, "session-123")
	}
	if session.ConfigurationID != "config-456" {
		t.Errorf("SchedulerSession.ConfigurationID = %q, want %q", session.ConfigurationID, "config-456")
	}
	if session.BookingURL == "" {
		t.Error("SchedulerSession.BookingURL should not be empty")
	}
}

// =============================================================================
// CreateSchedulerSessionRequest Tests
// =============================================================================

func TestCreateSchedulerSessionRequest_Creation(t *testing.T) {
	req := CreateSchedulerSessionRequest{
		ConfigurationID: "config-789",
		TimeToLive:      60,
		Slug:            "quick-chat",
		AdditionalFields: map[string]any{
			"email":   "guest@example.com",
			"company": "Guest Corp",
		},
	}

	if req.ConfigurationID != "config-789" {
		t.Errorf("CreateSchedulerSessionRequest.ConfigurationID = %q, want %q", req.ConfigurationID, "config-789")
	}
	if req.TimeToLive != 60 {
		t.Errorf("CreateSchedulerSessionRequest.TimeToLive = %d, want 60", req.TimeToLive)
	}
}

// =============================================================================
// Booking Tests
// =============================================================================

func TestBooking_Creation(t *testing.T) {
	now := time.Now()
	booking := Booking{
		BookingID: "booking-123",
		EventID:   "event-456",
		Title:     "Strategy Meeting",
		Organizer: Participant{
			Person: Person{Name: "Host User", Email: "host@example.com"},
		},
		Participants: []Participant{
			{Person: Person{Name: "Guest User", Email: "guest@example.com"}, Status: "yes"},
		},
		StartTime:   now.Add(24 * time.Hour),
		EndTime:     now.Add(25 * time.Hour),
		Status:      "confirmed",
		Description: "Discuss Q1 strategy",
		Location:    "Conference Room A",
		Timezone:    "America/New_York",
		Conferencing: &ConferencingDetails{
			URL:         "https://meet.google.com/abc-defg-hij",
			MeetingCode: "abc-defg-hij",
		},
		AdditionalFields: map[string]any{
			"guest_phone": "+1-555-123-4567",
		},
		Metadata: map[string]string{
			"source": "website",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if booking.BookingID != "booking-123" {
		t.Errorf("Booking.BookingID = %q, want %q", booking.BookingID, "booking-123")
	}
	if booking.Status != "confirmed" {
		t.Errorf("Booking.Status = %q, want %q", booking.Status, "confirmed")
	}
	if booking.Organizer.Email != "host@example.com" {
		t.Errorf("Booking.Organizer.Email = %q, want %q", booking.Organizer.Email, "host@example.com")
	}
	if len(booking.Participants) != 1 {
		t.Errorf("Booking.Participants length = %d, want 1", len(booking.Participants))
	}
}

func TestBooking_StatusValues(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"confirmed", "confirmed"},
		{"cancelled", "cancelled"},
		{"pending", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := Booking{Status: tt.status}
			if booking.Status != tt.status {
				t.Errorf("Booking.Status = %q, want %q", booking.Status, tt.status)
			}
		})
	}
}

// =============================================================================
// ConfirmBookingRequest Tests
// =============================================================================

func TestConfirmBookingRequest_Creation(t *testing.T) {
	tests := []struct {
		name string
		req  ConfirmBookingRequest
	}{
		{
			name: "confirm booking",
			req: ConfirmBookingRequest{
				Status: "confirmed",
			},
		},
		{
			name: "cancel booking with reason",
			req: ConfirmBookingRequest{
				Status: "cancelled",
				Reason: "Schedule conflict",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req.Status == "" {
				t.Error("ConfirmBookingRequest.Status should not be empty")
			}
		})
	}
}

// =============================================================================
// RescheduleBookingRequest Tests
// =============================================================================

func TestRescheduleBookingRequest_Creation(t *testing.T) {
	now := time.Now()
	req := RescheduleBookingRequest{
		StartTime: now.Add(48 * time.Hour).Unix(),
		EndTime:   now.Add(49 * time.Hour).Unix(),
		Timezone:  "America/Chicago",
		Reason:    "Guest requested different time",
	}

	if req.StartTime == 0 {
		t.Error("RescheduleBookingRequest.StartTime should not be zero")
	}
	if req.EndTime == 0 {
		t.Error("RescheduleBookingRequest.EndTime should not be zero")
	}
	if req.Timezone != "America/Chicago" {
		t.Errorf("RescheduleBookingRequest.Timezone = %q, want %q", req.Timezone, "America/Chicago")
	}
}

// =============================================================================
// SchedulerPage Tests
// =============================================================================

func TestSchedulerPage_Creation(t *testing.T) {
	now := time.Now()
	page := SchedulerPage{
		ID:              "page-123",
		ConfigurationID: "config-456",
		Name:            "Book a Demo",
		Slug:            "book-demo",
		URL:             "https://scheduler.example.com/book-demo",
		CustomDomain:    "booking.example.com",
		CreatedAt:       now.Add(-7 * 24 * time.Hour),
		ModifiedAt:      now,
	}

	if page.Name != "Book a Demo" {
		t.Errorf("SchedulerPage.Name = %q, want %q", page.Name, "Book a Demo")
	}
	if page.Slug != "book-demo" {
		t.Errorf("SchedulerPage.Slug = %q, want %q", page.Slug, "book-demo")
	}
	if page.CustomDomain != "booking.example.com" {
		t.Errorf("SchedulerPage.CustomDomain = %q, want %q", page.CustomDomain, "booking.example.com")
	}
}

// =============================================================================
// CreateSchedulerPageRequest Tests
// =============================================================================

func TestCreateSchedulerPageRequest_Creation(t *testing.T) {
	req := CreateSchedulerPageRequest{
		ConfigurationID: "config-789",
		Name:            "Sales Meeting",
		Slug:            "sales-meeting",
		CustomDomain:    "meetings.sales.example.com",
	}

	if req.ConfigurationID != "config-789" {
		t.Errorf("CreateSchedulerPageRequest.ConfigurationID = %q, want %q", req.ConfigurationID, "config-789")
	}
	if req.Name != "Sales Meeting" {
		t.Errorf("CreateSchedulerPageRequest.Name = %q, want %q", req.Name, "Sales Meeting")
	}
}

// =============================================================================
// UpdateSchedulerPageRequest Tests
// =============================================================================

func TestUpdateSchedulerPageRequest_Creation(t *testing.T) {
	name := "Updated Page Name"
	slug := "updated-slug"
	customDomain := "new.example.com"

	req := UpdateSchedulerPageRequest{
		Name:         &name,
		Slug:         &slug,
		CustomDomain: &customDomain,
	}

	if req.Name == nil || *req.Name != "Updated Page Name" {
		t.Errorf("UpdateSchedulerPageRequest.Name = %v, want %q", req.Name, "Updated Page Name")
	}
	if req.Slug == nil || *req.Slug != "updated-slug" {
		t.Errorf("UpdateSchedulerPageRequest.Slug = %v, want %q", req.Slug, "updated-slug")
	}
}

// =============================================================================
// CreateSchedulerConfigurationRequest Tests
// =============================================================================

func TestCreateSchedulerConfigurationRequest_Creation(t *testing.T) {
	req := CreateSchedulerConfigurationRequest{
		Name: "New Configuration",
		Slug: "new-config",
		Participants: []ConfigurationParticipant{
			{Email: "host@example.com", IsOrganizer: true},
		},
		Availability: AvailabilityRules{
			DurationMinutes: 60,
		},
		EventBooking: EventBooking{
			Title: "Meeting",
		},
		Scheduler: SchedulerSettings{
			AvailableDaysInFuture: 14,
		},
	}

	if req.Name != "New Configuration" {
		t.Errorf("CreateSchedulerConfigurationRequest.Name = %q, want %q", req.Name, "New Configuration")
	}
	if len(req.Participants) != 1 {
		t.Errorf("CreateSchedulerConfigurationRequest.Participants length = %d, want 1", len(req.Participants))
	}
}

// =============================================================================
// UpdateSchedulerConfigurationRequest Tests
// =============================================================================

func TestUpdateSchedulerConfigurationRequest_Creation(t *testing.T) {
	name := "Updated Configuration"
	requiresAuth := false

	req := UpdateSchedulerConfigurationRequest{
		Name:                &name,
		RequiresSessionAuth: &requiresAuth,
		Availability: &AvailabilityRules{
			DurationMinutes: 45,
		},
	}

	if req.Name == nil || *req.Name != "Updated Configuration" {
		t.Errorf("UpdateSchedulerConfigurationRequest.Name = %v, want %q", req.Name, "Updated Configuration")
	}
	if req.RequiresSessionAuth == nil || *req.RequiresSessionAuth {
		t.Error("UpdateSchedulerConfigurationRequest.RequiresSessionAuth should be false")
	}
	if req.Availability == nil {
		t.Fatal("UpdateSchedulerConfigurationRequest.Availability should not be nil")
	}
	if req.Availability.DurationMinutes != 45 {
		t.Errorf("AvailabilityRules.DurationMinutes = %d, want 45", req.Availability.DurationMinutes)
	}
}
