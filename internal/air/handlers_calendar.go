package air

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// handleListCalendars returns all calendars for the current account.
func (s *Server) handleListCalendars(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock calendars
	if s.demoMode {
		writeJSON(w, http.StatusOK, CalendarsResponse{
			Calendars: demoCalendars(),
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return
	}

	// Fetch calendars from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	calendars, err := s.nylasClient.GetCalendars(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch calendars: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := CalendarsResponse{
		Calendars: make([]CalendarResponse, 0, len(calendars)),
	}
	for _, c := range calendars {
		resp.Calendars = append(resp.Calendars, calendarToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleListEvents returns events for a calendar with optional date filtering.
func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	// Demo mode: return mock events
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventsResponse{
			Events:  demoEvents(),
			HasMore: false,
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Calendar ID is required
	calendarID := query.Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary" // Default to primary calendar
	}

	// Build query params
	params := &domain.EventQueryParams{
		Limit:           50,
		ExpandRecurring: true,
	}

	// Limit
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			params.Limit = parsed
		}
	}

	// Date range filtering
	if start := query.Get("start"); start != "" {
		if parsed, err := strconv.ParseInt(start, 10, 64); err == nil {
			params.Start = parsed
		}
	}
	if end := query.Get("end"); end != "" {
		if parsed, err := strconv.ParseInt(end, 10, 64); err == nil {
			params.End = parsed
		}
	}

	// Default to current week if no date range specified
	if params.Start == 0 && params.End == 0 {
		now := time.Now()
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		startOfWeek := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		// End of week (Saturday)
		endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Second)
		params.Start = startOfWeek.Unix()
		params.End = endOfWeek.Unix()
	}

	// Cursor for pagination
	cursor := query.Get("cursor")
	if cursor != "" {
		params.PageToken = cursor
	}

	// Get account email for cache lookup
	accountEmail := s.getAccountEmail(grantID)

	// Try cache first (only for first page)
	if cursor == "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEventStore(accountEmail); err == nil {
			cacheOpts := cache.EventListOptions{
				CalendarID: calendarID,
				Start:      time.Unix(params.Start, 0),
				End:        time.Unix(params.End, 0),
				Limit:      params.Limit,
			}
			if cached, err := store.List(cacheOpts); err == nil && len(cached) > 0 {
				resp := EventsResponse{
					Events:  make([]EventResponse, 0, len(cached)),
					HasMore: len(cached) >= params.Limit,
				}
				for _, e := range cached {
					resp.Events = append(resp.Events, cachedEventToResponse(e))
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch events from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetEventsWithCursor(ctx, grantID, calendarID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch events: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := EventsResponse{
		Events:     make([]EventResponse, 0, len(result.Data)),
		NextCursor: result.Pagination.NextCursor,
		HasMore:    result.Pagination.HasMore,
	}
	for _, e := range result.Data {
		resp.Events = append(resp.Events, eventToResponse(e))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleEventsRoute handles /api/events: GET (list) and POST (create).
func (s *Server) handleEventsRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListEvents(w, r)
	case http.MethodPost:
		s.handleCreateEvent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleEventByID handles single event operations: GET, PUT, DELETE.
func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	// Parse event ID and calendar ID from path: /api/events/{id}?calendar_id=xxx
	path := strings.TrimPrefix(r.URL.Path, "/api/events/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Event ID required", http.StatusBadRequest)
		return
	}
	eventID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.handleGetEvent(w, r, eventID)
	case http.MethodPut:
		s.handleUpdateEvent(w, r, eventID)
	case http.MethodDelete:
		s.handleDeleteEvent(w, r, eventID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateEvent creates a new event.
func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	// Demo mode: simulate success
	if s.demoMode {
		now := time.Now()
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Event: &EventResponse{
				ID:         "demo-event-new-" + now.Format("20060102150405"),
				CalendarID: "primary",
				Title:      "New Event",
				StartTime:  now.Add(1 * time.Hour).Unix(),
				EndTime:    now.Add(2 * time.Hour).Unix(),
				Status:     "confirmed",
				Busy:       true,
			},
			Message: "Event created (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req CreateEventRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Title is required",
		})
		return
	}

	calendarID := req.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	// Build domain request
	createReq := &domain.CreateEventRequest{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		Busy:        req.Busy,
	}

	// Set event time
	if req.IsAllDay {
		// All-day event: use date format
		startDate := time.Unix(req.StartTime, 0).Format("2006-01-02")
		endDate := time.Unix(req.EndTime, 0).Format("2006-01-02")
		createReq.When = domain.EventWhen{
			StartDate: startDate,
			EndDate:   endDate,
			Object:    "datespan",
		}
	} else {
		// Timed event
		createReq.When = domain.EventWhen{
			StartTime:     req.StartTime,
			EndTime:       req.EndTime,
			StartTimezone: req.Timezone,
			EndTimezone:   req.Timezone,
			Object:        "timespan",
		}
	}

	// Convert participants
	for _, p := range req.Participants {
		createReq.Participants = append(createReq.Participants, domain.Participant{
			Name:  p.Name,
			Email: p.Email,
		})
	}

	// Create event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.CreateEvent(ctx, grantID, calendarID, createReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to create event: " + err.Error(),
		})
		return
	}

	eventResp := eventToResponse(*event)
	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Event:   &eventResp,
		Message: "Event created successfully",
	})
}

// handleGetEvent retrieves a single event.
func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: return mock event
	if s.demoMode {
		events := demoEvents()
		for _, e := range events {
			if e.ID == eventID {
				writeJSON(w, http.StatusOK, e)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found"})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return
	}

	// Fetch event from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.GetEvent(ctx, grantID, calendarID, eventID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch event: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, eventToResponse(*event))
}

// handleUpdateEvent updates an existing event.
func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Event: &EventResponse{
				ID:         eventID,
				CalendarID: calendarID,
				Title:      "Updated Event",
				Status:     "confirmed",
			},
			Message: "Event updated (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req UpdateEventRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Build domain update request
	updateReq := &domain.UpdateEventRequest{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		Busy:        req.Busy,
	}

	// Set event time if provided
	if req.StartTime != nil && req.EndTime != nil {
		when := &domain.EventWhen{}
		if req.IsAllDay != nil && *req.IsAllDay {
			// All-day event
			startDate := time.Unix(*req.StartTime, 0).Format("2006-01-02")
			endDate := time.Unix(*req.EndTime, 0).Format("2006-01-02")
			when.StartDate = startDate
			when.EndDate = endDate
			when.Object = "datespan"
		} else {
			// Timed event
			when.StartTime = *req.StartTime
			when.EndTime = *req.EndTime
			if req.Timezone != nil {
				when.StartTimezone = *req.Timezone
				when.EndTimezone = *req.Timezone
			}
			when.Object = "timespan"
		}
		updateReq.When = when
	}

	// Convert participants if provided
	if len(req.Participants) > 0 {
		for _, p := range req.Participants {
			updateReq.Participants = append(updateReq.Participants, domain.Participant{
				Name:  p.Name,
				Email: p.Email,
			})
		}
	}

	// Update event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.UpdateEvent(ctx, grantID, calendarID, eventID, updateReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to update event: " + err.Error(),
		})
		return
	}

	eventResp := eventToResponse(*event)
	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Event:   &eventResp,
		Message: "Event updated successfully",
	})
}

// handleDeleteEvent deletes an event.
func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Message: "Event deleted (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Delete event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.nylasClient.DeleteEvent(ctx, grantID, calendarID, eventID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to delete event: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Message: "Event deleted successfully",
	})
}

// calendarToResponse converts a domain calendar to an API response.
func calendarToResponse(c domain.Calendar) CalendarResponse {
	return CalendarResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Timezone:    c.Timezone,
		IsPrimary:   c.IsPrimary,
		ReadOnly:    c.ReadOnly,
		HexColor:    c.HexColor,
	}
}

// eventToResponse converts a domain event to an API response.
func eventToResponse(e domain.Event) EventResponse {
	resp := EventResponse{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartTime:   e.When.StartTime,
		EndTime:     e.When.EndTime,
		Timezone:    e.When.StartTimezone,
		IsAllDay:    e.When.IsAllDay(),
		Status:      e.Status,
		Busy:        e.Busy,
		HtmlLink:    e.HtmlLink,
	}

	// Handle all-day events
	if resp.IsAllDay {
		if e.When.Date != "" {
			t, _ := time.Parse("2006-01-02", e.When.Date)
			resp.StartTime = t.Unix()
			resp.EndTime = t.Add(24 * time.Hour).Unix()
		} else if e.When.StartDate != "" {
			st, _ := time.Parse("2006-01-02", e.When.StartDate)
			resp.StartTime = st.Unix()
			if e.When.EndDate != "" {
				et, _ := time.Parse("2006-01-02", e.When.EndDate)
				resp.EndTime = et.Unix()
			}
		}
	}

	// Convert participants
	for _, p := range e.Participants {
		resp.Participants = append(resp.Participants, EventParticipantResponse{
			Name:   p.Name,
			Email:  p.Email,
			Status: p.Status,
		})
	}

	// Convert conferencing
	if e.Conferencing != nil && e.Conferencing.Details != nil {
		resp.Conferencing = &ConferencingResponse{
			Provider: e.Conferencing.Provider,
			URL:      e.Conferencing.Details.URL,
		}
	}

	return resp
}

// cachedEventToResponse converts a cached event to response format.
func cachedEventToResponse(e *cache.CachedEvent) EventResponse {
	return EventResponse{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartTime:   e.StartTime.Unix(),
		EndTime:     e.EndTime.Unix(),
		IsAllDay:    e.AllDay,
		Status:      e.Status,
		Busy:        e.Busy,
	}
}

// demoCalendars returns demo calendar data.
func demoCalendars() []CalendarResponse {
	return []CalendarResponse{
		{
			ID:        "primary",
			Name:      "Personal Calendar",
			Timezone:  "America/New_York",
			IsPrimary: true,
			ReadOnly:  false,
			HexColor:  "#4285f4",
		},
		{
			ID:        "work",
			Name:      "Work Calendar",
			Timezone:  "America/New_York",
			IsPrimary: false,
			ReadOnly:  false,
			HexColor:  "#0b8043",
		},
		{
			ID:          "holidays",
			Name:        "US Holidays",
			Description: "Public holidays in the United States",
			Timezone:    "America/New_York",
			IsPrimary:   false,
			ReadOnly:    true,
			HexColor:    "#f6bf26",
		},
	}
}

// demoEvents returns demo event data.
func demoEvents() []EventResponse {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return []EventResponse{
		{
			ID:          "demo-event-001",
			CalendarID:  "primary",
			Title:       "Team Standup",
			Description: "Daily team sync",
			Location:    "",
			StartTime:   today.Add(9 * time.Hour).Unix(),
			EndTime:     today.Add(9*time.Hour + 30*time.Minute).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
			Participants: []EventParticipantResponse{
				{Name: "Sarah Chen", Email: "sarah@example.com", Status: "yes"},
				{Name: "Alex Johnson", Email: "alex@example.com", Status: "yes"},
			},
			Conferencing: &ConferencingResponse{
				Provider: "Google Meet",
				URL:      "https://meet.google.com/abc-defg-hij",
			},
		},
		{
			ID:          "demo-event-002",
			CalendarID:  "work",
			Title:       "Product Review",
			Description: "Weekly product roadmap review with stakeholders",
			Location:    "Conference Room A",
			StartTime:   today.Add(14 * time.Hour).Unix(),
			EndTime:     today.Add(15 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
			Participants: []EventParticipantResponse{
				{Name: "Product Team", Email: "product@example.com", Status: "yes"},
			},
		},
		{
			ID:          "demo-event-003",
			CalendarID:  "primary",
			Title:       "Lunch with Client",
			Description: "Discuss Q1 partnership opportunities",
			Location:    "Cafe Milano",
			StartTime:   today.Add(12 * time.Hour).Unix(),
			EndTime:     today.Add(13 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
		},
		{
			ID:          "demo-event-004",
			CalendarID:  "primary",
			Title:       "Focus Time",
			Description: "Deep work - no meetings",
			StartTime:   today.Add(15 * time.Hour).Unix(),
			EndTime:     today.Add(17 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
		},
		{
			ID:         "demo-event-005",
			CalendarID: "holidays",
			Title:      "Christmas Day",
			StartTime:  time.Date(now.Year(), 12, 25, 0, 0, 0, 0, now.Location()).Unix(),
			EndTime:    time.Date(now.Year(), 12, 26, 0, 0, 0, 0, now.Location()).Unix(),
			Timezone:   "America/New_York",
			IsAllDay:   true,
			Status:     "confirmed",
			Busy:       false,
		},
	}
}
