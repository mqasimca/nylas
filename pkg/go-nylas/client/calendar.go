package client

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/pkg/go-nylas/common"
	"github.com/mqasimca/nylas/pkg/go-nylas/config"
)

// CalendarClient provides calendar and event operations for plugins.
type CalendarClient struct {
	adapter *nylas.HTTPClient
	grantID string
	config  *config.Config
}

// NewCalendarClient creates a new calendar client for the given configuration.
func NewCalendarClient(cfg *config.Config) (*CalendarClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.GetAPIKey() == "" {
		return nil, fmt.Errorf("API key not configured")
	}
	if cfg.GetGrantID() == "" {
		return nil, fmt.Errorf("grant ID not configured")
	}

	// Create internal HTTP client
	adapter := nylas.NewHTTPClient(cfg.GetAPIKey(), cfg.GetGrantID(), cfg.GetRegion())

	return &CalendarClient{
		adapter: adapter,
		grantID: cfg.GetGrantID(),
		config:  cfg,
	}, nil
}

// Calendar represents a calendar.
type Calendar struct {
	ID          string
	Name        string
	Description string
	Location    string
	Timezone    string
	ReadOnly    bool
	IsPrimary   bool
	IsOwner     bool
	HexColor    string
}

// Event represents a calendar event.
type Event struct {
	ID           string
	CalendarID   string
	Title        string
	Description  string
	Location     string
	When         EventTime
	Participants []Participant
	Organizer    *Participant
	Status       string
	Busy         bool
	ReadOnly     bool
	Visibility   string
	Recurrence   []string
	Conferencing *ConferencingDetails
}

// EventTime represents event timing (either time-based or date-based).
type EventTime struct {
	StartTime     int64
	EndTime       int64
	StartTimezone string
	EndTimezone   string
	Date          string    // For all-day events
	StartDate     string    // For multi-day all-day events
	EndDate       string    // For multi-day all-day events
	Object        string    // "timespan" or "date" or "datespan"
}

// Participant represents an event participant.
type Participant struct {
	Name    string
	Email   string
	Status  string
	Comment string
}

// ConferencingDetails represents video conferencing details.
type ConferencingDetails struct {
	Provider    string
	URL         string
	MeetingCode string
	Password    string
	Phone       []string
}

// CreateEventOptions contains options for creating an event.
type CreateEventOptions struct {
	CalendarID   string
	Title        string
	Description  string
	Location     string
	When         EventTime
	Participants []Participant
	Busy         bool
	Conferencing *ConferencingDetails
}

// UpdateEventOptions contains options for updating an event.
type UpdateEventOptions struct {
	Title        *string
	Description  *string
	Location     *string
	When         *EventTime
	Participants []Participant
	Busy         *bool
}

// ListCalendars retrieves all calendars.
func (c *CalendarClient) ListCalendars(ctx context.Context) ([]*Calendar, error) {
	internalCalendars, err := c.adapter.GetCalendars(ctx, c.grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]*Calendar, len(internalCalendars))
	for i, cal := range internalCalendars {
		calendars[i] = convertCalendar(&cal)
	}

	return calendars, nil
}

// GetCalendar retrieves a single calendar by ID.
func (c *CalendarClient) GetCalendar(ctx context.Context, calendarID string) (*Calendar, error) {
	internalCalendar, err := c.adapter.GetCalendar(ctx, c.grantID, calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	return convertCalendar(internalCalendar), nil
}

// ListEvents retrieves events from a calendar.
func (c *CalendarClient) ListEvents(ctx context.Context, calendarID string, limit int) ([]*Event, error) {
	internalEvents, err := c.adapter.GetEvents(ctx, c.grantID, calendarID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]*Event, len(internalEvents))
	for i, evt := range internalEvents {
		events[i] = convertEvent(&evt)
	}

	return events, nil
}

// GetEvent retrieves a single event by ID.
func (c *CalendarClient) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	internalEvent, err := c.adapter.GetEvent(ctx, c.grantID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return convertEvent(internalEvent), nil
}

// CreateEvent creates a new calendar event.
func (c *CalendarClient) CreateEvent(ctx context.Context, opts *CreateEventOptions) (*Event, error) {
	if opts == nil {
		return nil, fmt.Errorf("create options cannot be nil")
	}

	// Convert to internal format
	req := &domain.CreateEventRequest{
		CalendarID:  opts.CalendarID,
		Title:       opts.Title,
		Description: opts.Description,
		Location:    opts.Location,
		Busy:        opts.Busy,
	}

	// Convert when
	req.When = convertEventTimeToInternal(&opts.When)

	// Convert participants
	req.Participants = make([]domain.Participant, len(opts.Participants))
	for i, p := range opts.Participants {
		req.Participants[i] = domain.Participant{
			Name:  p.Name,
			Email: p.Email,
		}
	}

	// Create via internal adapter
	createdEvent, err := c.adapter.CreateEvent(ctx, c.grantID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return convertEvent(createdEvent), nil
}

// UpdateEvent updates an existing event.
func (c *CalendarClient) UpdateEvent(ctx context.Context, eventID string, opts *UpdateEventOptions) (*Event, error) {
	if opts == nil {
		return nil, fmt.Errorf("update options cannot be nil")
	}

	// Convert to internal format
	req := &domain.UpdateEventRequest{
		Title:       opts.Title,
		Description: opts.Description,
		Location:    opts.Location,
		Busy:        opts.Busy,
	}

	if opts.When != nil {
		req.When = convertEventTimeToInternal(opts.When)
	}

	// Update via internal adapter
	updatedEvent, err := c.adapter.UpdateEvent(ctx, c.grantID, eventID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return convertEvent(updatedEvent), nil
}

// DeleteEvent deletes an event.
func (c *CalendarClient) DeleteEvent(ctx context.Context, eventID string) error {
	err := c.adapter.DeleteEvent(ctx, c.grantID, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// Helper functions

func convertCalendar(internal *domain.Calendar) *Calendar {
	if internal == nil {
		return nil
	}

	return &Calendar{
		ID:          internal.ID,
		Name:        internal.Name,
		Description: internal.Description,
		Location:    internal.Location,
		Timezone:    internal.Timezone,
		ReadOnly:    internal.ReadOnly,
		IsPrimary:   internal.IsPrimary,
		IsOwner:     internal.IsOwned,
		HexColor:    internal.HexColor,
	}
}

func convertEvent(internal *domain.Event) *Event {
	if internal == nil {
		return nil
	}

	evt := &Event{
		ID:          internal.ID,
		CalendarID:  internal.CalendarID,
		Title:       internal.Title,
		Description: internal.Description,
		Location:    internal.Location,
		Status:      internal.Status,
		Busy:        internal.Busy,
		ReadOnly:    internal.ReadOnly,
		Visibility:  internal.Visibility,
		Recurrence:  internal.Recurrence,
	}

	// Convert when
	if internal.When != nil {
		evt.When = EventTime{
			StartTime:     internal.When.StartTime,
			EndTime:       internal.When.EndTime,
			StartTimezone: internal.When.StartTimezone,
			EndTimezone:   internal.When.EndTimezone,
			Date:          internal.When.Date,
			StartDate:     internal.When.StartDate,
			EndDate:       internal.When.EndDate,
			Object:        internal.When.Object,
		}
	}

	// Convert participants
	evt.Participants = make([]Participant, len(internal.Participants))
	for i, p := range internal.Participants {
		evt.Participants[i] = Participant{
			Name:   p.Name,
			Email:  p.Email,
			Status: p.Status,
		}
	}

	// Convert organizer
	if internal.Organizer != nil {
		evt.Organizer = &Participant{
			Name:   internal.Organizer.Name,
			Email:  internal.Organizer.Email,
			Status: internal.Organizer.Status,
		}
	}

	return evt
}

func convertEventTimeToInternal(when *EventTime) *domain.When {
	if when == nil {
		return nil
	}

	return &domain.When{
		StartTime:     when.StartTime,
		EndTime:       when.EndTime,
		StartTimezone: when.StartTimezone,
		EndTimezone:   when.EndTimezone,
		Date:          when.Date,
		StartDate:     when.StartDate,
		EndDate:       when.EndDate,
		Object:        when.Object,
	}
}
