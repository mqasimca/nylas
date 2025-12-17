package domain

import "time"

// Calendar represents a calendar from Nylas.
type Calendar struct {
	ID          string `json:"id"`
	GrantID     string `json:"grant_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	ReadOnly    bool   `json:"read_only"`
	IsPrimary   bool   `json:"is_primary,omitempty"`
	IsOwner     bool   `json:"is_owner,omitempty"`
	HexColor    string `json:"hex_color,omitempty"`
	Object      string `json:"object,omitempty"`
}

// Event represents a calendar event from Nylas.
type Event struct {
	ID             string          `json:"id"`
	GrantID        string          `json:"grant_id"`
	CalendarID     string          `json:"calendar_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description,omitempty"`
	Location       string          `json:"location,omitempty"`
	When           EventWhen       `json:"when"`
	Participants   []Participant   `json:"participants,omitempty"`
	Organizer      *Participant    `json:"organizer,omitempty"`
	Status         string          `json:"status,omitempty"` // confirmed, cancelled, tentative
	Busy           bool            `json:"busy"`
	ReadOnly       bool            `json:"read_only"`
	Visibility     string          `json:"visibility,omitempty"` // public, private
	Recurrence     []string        `json:"recurrence,omitempty"`
	Conferencing   *Conferencing   `json:"conferencing,omitempty"`
	Reminders      *Reminders      `json:"reminders,omitempty"`
	MasterEventID  string          `json:"master_event_id,omitempty"`
	ICalUID        string          `json:"ical_uid,omitempty"`
	HtmlLink       string          `json:"html_link,omitempty"`
	CreatedAt      time.Time       `json:"created_at,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`
	Object         string          `json:"object,omitempty"`
}

// EventWhen represents when an event occurs.
type EventWhen struct {
	// For timespan events
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	StartTimezone string `json:"start_timezone,omitempty"`
	EndTimezone   string `json:"end_timezone,omitempty"`

	// For date events (all-day)
	Date    string `json:"date,omitempty"`
	EndDate string `json:"end_date,omitempty"`

	// For datespan events (multi-day all-day)
	StartDate string `json:"start_date,omitempty"`
	// EndDate is shared with date events

	Object string `json:"object,omitempty"` // timespan, date, datespan
}

// StartDateTime returns the start time as a time.Time.
func (w EventWhen) StartDateTime() time.Time {
	if w.StartTime > 0 {
		return time.Unix(w.StartTime, 0)
	}
	if w.Date != "" {
		t, _ := time.Parse("2006-01-02", w.Date)
		return t
	}
	if w.StartDate != "" {
		t, _ := time.Parse("2006-01-02", w.StartDate)
		return t
	}
	return time.Time{}
}

// EndDateTime returns the end time as a time.Time.
func (w EventWhen) EndDateTime() time.Time {
	if w.EndTime > 0 {
		return time.Unix(w.EndTime, 0)
	}
	if w.EndDate != "" {
		t, _ := time.Parse("2006-01-02", w.EndDate)
		return t
	}
	if w.Date != "" {
		t, _ := time.Parse("2006-01-02", w.Date)
		return t
	}
	return time.Time{}
}

// IsAllDay returns true if this is an all-day event.
func (w EventWhen) IsAllDay() bool {
	return w.Object == "date" || w.Object == "datespan" || w.Date != "" || w.StartDate != ""
}

// Participant represents an event participant.
type Participant struct {
	Name    string `json:"name,omitempty"`
	Email   string `json:"email"`
	Status  string `json:"status,omitempty"` // yes, no, maybe, noreply
	Comment string `json:"comment,omitempty"`
}

// Conferencing represents video conferencing details.
type Conferencing struct {
	Provider string            `json:"provider,omitempty"` // Google Meet, Zoom, etc.
	Details  *ConferencingDetails `json:"details,omitempty"`
}

// ConferencingDetails contains conferencing URLs and info.
type ConferencingDetails struct {
	URL      string   `json:"url,omitempty"`
	MeetingCode string `json:"meeting_code,omitempty"`
	Password string   `json:"password,omitempty"`
	Phone    []string `json:"phone,omitempty"`
}

// Reminders represents event reminders.
type Reminders struct {
	UseDefault bool       `json:"use_default"`
	Overrides  []Reminder `json:"overrides,omitempty"`
}

// Reminder represents a single reminder.
type Reminder struct {
	ReminderMinutes int    `json:"reminder_minutes"`
	ReminderMethod  string `json:"reminder_method,omitempty"` // email, popup
}

// EventQueryParams for filtering events.
type EventQueryParams struct {
	Limit       int    `json:"limit,omitempty"`
	PageToken   string `json:"page_token,omitempty"`
	CalendarID  string `json:"calendar_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Location    string `json:"location,omitempty"`
	ShowCancelled bool `json:"show_cancelled,omitempty"`
	Start       int64  `json:"start,omitempty"` // Unix timestamp
	End         int64  `json:"end,omitempty"`   // Unix timestamp
	MetadataPair string `json:"metadata_pair,omitempty"`
	Busy        *bool  `json:"busy,omitempty"`
	OrderBy     string `json:"order_by,omitempty"` // start, end
	ExpandRecurring bool `json:"expand_recurring,omitempty"`
}

// CreateEventRequest for creating a new event.
type CreateEventRequest struct {
	Title        string        `json:"title"`
	Description  string        `json:"description,omitempty"`
	Location     string        `json:"location,omitempty"`
	When         EventWhen     `json:"when"`
	Participants []Participant `json:"participants,omitempty"`
	Busy         bool          `json:"busy"`
	Visibility   string        `json:"visibility,omitempty"`
	Recurrence   []string      `json:"recurrence,omitempty"`
	Conferencing *Conferencing `json:"conferencing,omitempty"`
	Reminders    *Reminders    `json:"reminders,omitempty"`
	CalendarID   string        `json:"calendar_id,omitempty"`
}

// UpdateEventRequest for updating an event.
type UpdateEventRequest struct {
	Title        *string       `json:"title,omitempty"`
	Description  *string       `json:"description,omitempty"`
	Location     *string       `json:"location,omitempty"`
	When         *EventWhen    `json:"when,omitempty"`
	Participants []Participant `json:"participants,omitempty"`
	Busy         *bool         `json:"busy,omitempty"`
	Visibility   *string       `json:"visibility,omitempty"`
	Recurrence   []string      `json:"recurrence,omitempty"`
	Conferencing *Conferencing `json:"conferencing,omitempty"`
	Reminders    *Reminders    `json:"reminders,omitempty"`
}

// CalendarListResponse represents a paginated calendar list response.
type CalendarListResponse struct {
	Data       []Calendar `json:"data"`
	Pagination Pagination `json:"pagination,omitempty"`
}

// EventListResponse represents a paginated event list response.
type EventListResponse struct {
	Data       []Event    `json:"data"`
	Pagination Pagination `json:"pagination,omitempty"`
}

// FreeBusyRequest for checking availability.
type FreeBusyRequest struct {
	StartTime int64    `json:"start_time"` // Unix timestamp
	EndTime   int64    `json:"end_time"`   // Unix timestamp
	Emails    []string `json:"emails"`
}

// FreeBusyResponse represents availability data.
type FreeBusyResponse struct {
	Data []FreeBusyCalendar `json:"data"`
}

// FreeBusyCalendar represents a calendar's availability.
type FreeBusyCalendar struct {
	Email      string       `json:"email"`
	TimeSlots  []TimeSlot   `json:"time_slots,omitempty"`
	Object     string       `json:"object,omitempty"`
}

// TimeSlot represents a busy time slot.
type TimeSlot struct {
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Status    string `json:"status,omitempty"` // busy, free
	Object    string `json:"object,omitempty"`
}

// AvailabilityRequest for finding available meeting times.
type AvailabilityRequest struct {
	StartTime      int64                  `json:"start_time"`
	EndTime        int64                  `json:"end_time"`
	DurationMinutes int                   `json:"duration_minutes"`
	Participants   []AvailabilityParticipant `json:"participants"`
	IntervalMinutes int                   `json:"interval_minutes,omitempty"`
	RoundTo        int                    `json:"round_to,omitempty"`
}

// AvailabilityParticipant represents a participant in availability check.
type AvailabilityParticipant struct {
	Email       string   `json:"email"`
	CalendarIDs []string `json:"calendar_ids,omitempty"`
}

// AvailabilityResponse contains available time slots.
type AvailabilityResponse struct {
	Data []AvailableSlot `json:"data"`
}

// AvailableSlot represents an available meeting slot.
type AvailableSlot struct {
	StartTime    int64                       `json:"start_time"`
	EndTime      int64                       `json:"end_time"`
	Participants []AvailabilityParticipant   `json:"participants,omitempty"`
}
