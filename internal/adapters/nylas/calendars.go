package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// calendarResponse represents an API calendar response.
type calendarResponse struct {
	ID          string `json:"id"`
	GrantID     string `json:"grant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Timezone    string `json:"timezone"`
	ReadOnly    bool   `json:"read_only"`
	IsPrimary   bool   `json:"is_primary"`
	IsOwner     bool   `json:"is_owner"`
	HexColor    string `json:"hex_color"`
	Object      string `json:"object"`
}

// eventResponse represents an API event response.
type eventResponse struct {
	ID          string `json:"id"`
	GrantID     string `json:"grant_id"`
	CalendarID  string `json:"calendar_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	When        struct {
		StartTime     int64  `json:"start_time,omitempty"`
		EndTime       int64  `json:"end_time,omitempty"`
		StartTimezone string `json:"start_timezone,omitempty"`
		EndTimezone   string `json:"end_timezone,omitempty"`
		Date          string `json:"date,omitempty"`
		EndDate       string `json:"end_date,omitempty"`
		StartDate     string `json:"start_date,omitempty"`
		Object        string `json:"object,omitempty"`
	} `json:"when"`
	Participants []struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Status  string `json:"status"`
		Comment string `json:"comment"`
	} `json:"participants"`
	Organizer *struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Status  string `json:"status"`
		Comment string `json:"comment"`
	} `json:"organizer"`
	Status       string   `json:"status"`
	Busy         bool     `json:"busy"`
	ReadOnly     bool     `json:"read_only"`
	Visibility   string   `json:"visibility"`
	Recurrence   []string `json:"recurrence"`
	Conferencing *struct {
		Provider string `json:"provider"`
		Details  *struct {
			URL         string   `json:"url"`
			MeetingCode string   `json:"meeting_code"`
			Password    string   `json:"password"`
			Phone       []string `json:"phone"`
		} `json:"details"`
	} `json:"conferencing"`
	Reminders *struct {
		UseDefault bool `json:"use_default"`
		Overrides  []struct {
			ReminderMinutes int    `json:"reminder_minutes"`
			ReminderMethod  string `json:"reminder_method"`
		} `json:"overrides"`
	} `json:"reminders"`
	MasterEventID string `json:"master_event_id"`
	ICalUID       string `json:"ical_uid"`
	HtmlLink      string `json:"html_link"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
	Object        string `json:"object"`
}

// GetCalendars retrieves all calendars for a grant.
func (c *HTTPClient) GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []calendarResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertCalendars(result.Data), nil
}

// GetCalendar retrieves a single calendar by ID.
func (c *HTTPClient) GetCalendar(ctx context.Context, grantID, calendarID string) (*domain.Calendar, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars/%s", c.baseURL, grantID, calendarID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: calendar not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data calendarResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cal := convertCalendar(result.Data)
	return &cal, nil
}

// CreateCalendar creates a new calendar.
func (c *HTTPClient) CreateCalendar(ctx context.Context, grantID string, req *domain.CreateCalendarRequest) (*domain.Calendar, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars", c.baseURL, grantID)

	payload := map[string]interface{}{
		"name": req.Name,
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.Location != "" {
		payload["location"] = req.Location
	}
	if req.Timezone != "" {
		payload["timezone"] = req.Timezone
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data calendarResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cal := convertCalendar(result.Data)
	return &cal, nil
}

// UpdateCalendar updates an existing calendar.
func (c *HTTPClient) UpdateCalendar(ctx context.Context, grantID, calendarID string, req *domain.UpdateCalendarRequest) (*domain.Calendar, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars/%s", c.baseURL, grantID, calendarID)

	payload := make(map[string]interface{})
	if req.Name != nil {
		payload["name"] = *req.Name
	}
	if req.Description != nil {
		payload["description"] = *req.Description
	}
	if req.Location != nil {
		payload["location"] = *req.Location
	}
	if req.Timezone != nil {
		payload["timezone"] = *req.Timezone
	}
	if req.HexColor != nil {
		payload["hex_color"] = *req.HexColor
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data calendarResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cal := convertCalendar(result.Data)
	return &cal, nil
}

// DeleteCalendar deletes a calendar.
func (c *HTTPClient) DeleteCalendar(ctx context.Context, grantID, calendarID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars/%s", c.baseURL, grantID, calendarID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetEvents retrieves events for a calendar.
func (c *HTTPClient) GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	result, err := c.GetEventsWithCursor(ctx, grantID, calendarID, params)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEventsWithCursor retrieves events with pagination cursor support.
func (c *HTTPClient) GetEventsWithCursor(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) (*domain.EventListResponse, error) {
	if params == nil {
		params = &domain.EventQueryParams{Limit: 10}
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/events", c.baseURL, grantID)
	q := url.Values{}
	q.Set("calendar_id", calendarID)
	q.Set("limit", strconv.Itoa(params.Limit))

	if params.PageToken != "" {
		q.Set("page_token", params.PageToken)
	}
	if params.Start > 0 {
		q.Set("start", strconv.FormatInt(params.Start, 10))
	}
	if params.End > 0 {
		q.Set("end", strconv.FormatInt(params.End, 10))
	}
	if params.Title != "" {
		q.Set("title", params.Title)
	}
	if params.Location != "" {
		q.Set("location", params.Location)
	}
	if params.ShowCancelled {
		q.Set("show_cancelled", "true")
	}
	if params.ExpandRecurring {
		q.Set("expand_recurring", "true")
	}
	if params.Busy != nil {
		q.Set("busy", strconv.FormatBool(*params.Busy))
	}
	if params.OrderBy != "" {
		q.Set("order_by", params.OrderBy)
	}

	queryURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data       []eventResponse `json:"data"`
		NextCursor string          `json:"next_cursor,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.EventListResponse{
		Data: convertEvents(result.Data),
		Pagination: domain.Pagination{
			NextCursor: result.NextCursor,
			HasMore:    result.NextCursor != "",
		},
	}, nil
}

// GetEvent retrieves a single event by ID.
func (c *HTTPClient) GetEvent(ctx context.Context, grantID, calendarID, eventID string) (*domain.Event, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/events/%s?calendar_id=%s", c.baseURL, grantID, eventID, calendarID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: event not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data eventResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	event := convertEvent(result.Data)
	return &event, nil
}

// CreateEvent creates a new event.
func (c *HTTPClient) CreateEvent(ctx context.Context, grantID, calendarID string, req *domain.CreateEventRequest) (*domain.Event, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/events?calendar_id=%s", c.baseURL, grantID, calendarID)

	payload := map[string]interface{}{
		"title": req.Title,
		"when":  req.When,
	}

	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.Location != "" {
		payload["location"] = req.Location
	}
	if len(req.Participants) > 0 {
		payload["participants"] = req.Participants
	}
	payload["busy"] = req.Busy
	if req.Visibility != "" {
		payload["visibility"] = req.Visibility
	}
	if len(req.Recurrence) > 0 {
		payload["recurrence"] = req.Recurrence
	}
	if req.Conferencing != nil {
		payload["conferencing"] = req.Conferencing
	}
	if req.Reminders != nil {
		payload["reminders"] = req.Reminders
	}
	if len(req.Metadata) > 0 {
		payload["metadata"] = req.Metadata
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data eventResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	event := convertEvent(result.Data)
	return &event, nil
}

// UpdateEvent updates an existing event.
func (c *HTTPClient) UpdateEvent(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/events/%s?calendar_id=%s", c.baseURL, grantID, eventID, calendarID)

	payload := make(map[string]interface{})
	if req.Title != nil {
		payload["title"] = *req.Title
	}
	if req.Description != nil {
		payload["description"] = *req.Description
	}
	if req.Location != nil {
		payload["location"] = *req.Location
	}
	if req.When != nil {
		payload["when"] = req.When
	}
	if len(req.Participants) > 0 {
		payload["participants"] = req.Participants
	}
	if req.Busy != nil {
		payload["busy"] = *req.Busy
	}
	if req.Visibility != nil {
		payload["visibility"] = *req.Visibility
	}
	if len(req.Recurrence) > 0 {
		payload["recurrence"] = req.Recurrence
	}
	if req.Conferencing != nil {
		payload["conferencing"] = req.Conferencing
	}
	if req.Reminders != nil {
		payload["reminders"] = req.Reminders
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data eventResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	event := convertEvent(result.Data)
	return &event, nil
}

// DeleteEvent deletes an event.
func (c *HTTPClient) DeleteEvent(ctx context.Context, grantID, calendarID, eventID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/events/%s?calendar_id=%s", c.baseURL, grantID, eventID, calendarID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// SendRSVP sends an RSVP response to an event invitation.
func (c *HTTPClient) SendRSVP(ctx context.Context, grantID, calendarID, eventID string, req *domain.SendRSVPRequest) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/events/%s/send-rsvp?calendar_id=%s", c.baseURL, grantID, eventID, calendarID)

	payload := map[string]interface{}{
		"status": req.Status,
	}
	if req.Comment != "" {
		payload["comment"] = req.Comment
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return c.parseError(resp)
	}

	return nil
}

// GetFreeBusy retrieves free/busy information.
func (c *HTTPClient) GetFreeBusy(ctx context.Context, grantID string, freeBusyReq *domain.FreeBusyRequest) (*domain.FreeBusyResponse, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars/free-busy", c.baseURL, grantID)

	body, _ := json.Marshal(freeBusyReq)
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result domain.FreeBusyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAvailability retrieves availability information.
func (c *HTTPClient) GetAvailability(ctx context.Context, availReq *domain.AvailabilityRequest) (*domain.AvailabilityResponse, error) {
	queryURL := fmt.Sprintf("%s/v3/calendars/availability", c.baseURL)

	body, _ := json.Marshal(availReq)
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result domain.AvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CreateVirtualCalendarGrant creates a virtual calendar grant (account).
func (c *HTTPClient) CreateVirtualCalendarGrant(ctx context.Context, email string) (*domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/connect/custom", c.baseURL)

	payload := map[string]interface{}{
		"provider": "virtual-calendar",
		"settings": map[string]interface{}{
			"email": email,
		},
		"scope": []string{"calendar"},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result domain.VirtualCalendarGrant
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListVirtualCalendarGrants lists all virtual calendar grants.
func (c *HTTPClient) ListVirtualCalendarGrants(ctx context.Context) ([]domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/grants?provider=virtual-calendar", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []domain.VirtualCalendarGrant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetVirtualCalendarGrant retrieves a single virtual calendar grant by ID.
func (c *HTTPClient) GetVirtualCalendarGrant(ctx context.Context, grantID string) (*domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: virtual calendar grant not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.VirtualCalendarGrant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// DeleteVirtualCalendarGrant deletes a virtual calendar grant.
func (c *HTTPClient) DeleteVirtualCalendarGrant(ctx context.Context, grantID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetRecurringEventInstances retrieves all instances of a recurring event.
func (c *HTTPClient) GetRecurringEventInstances(ctx context.Context, grantID, calendarID, masterEventID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	if params == nil {
		params = &domain.EventQueryParams{
			ExpandRecurring: true,
			Limit:           50,
		}
	} else {
		params.ExpandRecurring = true
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/events", c.baseURL, grantID)
	q := url.Values{}
	q.Set("calendar_id", calendarID)
	q.Set("event_id", masterEventID)
	q.Set("expand_recurring", "true")
	q.Set("limit", strconv.Itoa(params.Limit))

	if params.Start > 0 {
		q.Set("start", strconv.FormatInt(params.Start, 10))
	}
	if params.End > 0 {
		q.Set("end", strconv.FormatInt(params.End, 10))
	}

	queryURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []eventResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertEvents(result.Data), nil
}

// UpdateRecurringEventInstance updates a single instance of a recurring event.
func (c *HTTPClient) UpdateRecurringEventInstance(ctx context.Context, grantID, calendarID, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {
	return c.UpdateEvent(ctx, grantID, calendarID, eventID, req)
}

// DeleteRecurringEventInstance deletes a single instance of a recurring event.
func (c *HTTPClient) DeleteRecurringEventInstance(ctx context.Context, grantID, calendarID, eventID string) error {
	return c.DeleteEvent(ctx, grantID, calendarID, eventID)
}

// convertCalendars converts API calendar responses to domain models.
func convertCalendars(cals []calendarResponse) []domain.Calendar {
	result := make([]domain.Calendar, len(cals))
	for i, c := range cals {
		result[i] = convertCalendar(c)
	}
	return result
}

// convertCalendar converts an API calendar response to domain model.
func convertCalendar(c calendarResponse) domain.Calendar {
	return domain.Calendar{
		ID:          c.ID,
		GrantID:     c.GrantID,
		Name:        c.Name,
		Description: c.Description,
		Location:    c.Location,
		Timezone:    c.Timezone,
		ReadOnly:    c.ReadOnly,
		IsPrimary:   c.IsPrimary,
		IsOwner:     c.IsOwner,
		HexColor:    c.HexColor,
		Object:      c.Object,
	}
}

// convertEvents converts API event responses to domain models.
func convertEvents(events []eventResponse) []domain.Event {
	result := make([]domain.Event, len(events))
	for i, e := range events {
		result[i] = convertEvent(e)
	}
	return result
}

// convertEvent converts an API event response to domain model.
func convertEvent(e eventResponse) domain.Event {
	participants := make([]domain.Participant, len(e.Participants))
	for j, p := range e.Participants {
		participants[j] = domain.Participant{
			Name:    p.Name,
			Email:   p.Email,
			Status:  p.Status,
			Comment: p.Comment,
		}
	}

	var organizer *domain.Participant
	if e.Organizer != nil {
		organizer = &domain.Participant{
			Name:    e.Organizer.Name,
			Email:   e.Organizer.Email,
			Status:  e.Organizer.Status,
			Comment: e.Organizer.Comment,
		}
	}

	var conferencing *domain.Conferencing
	if e.Conferencing != nil {
		conferencing = &domain.Conferencing{
			Provider: e.Conferencing.Provider,
		}
		if e.Conferencing.Details != nil {
			conferencing.Details = &domain.ConferencingDetails{
				URL:         e.Conferencing.Details.URL,
				MeetingCode: e.Conferencing.Details.MeetingCode,
				Password:    e.Conferencing.Details.Password,
				Phone:       e.Conferencing.Details.Phone,
			}
		}
	}

	var reminders *domain.Reminders
	if e.Reminders != nil {
		overrides := make([]domain.Reminder, len(e.Reminders.Overrides))
		for j, o := range e.Reminders.Overrides {
			overrides[j] = domain.Reminder{
				ReminderMinutes: o.ReminderMinutes,
				ReminderMethod:  o.ReminderMethod,
			}
		}
		reminders = &domain.Reminders{
			UseDefault: e.Reminders.UseDefault,
			Overrides:  overrides,
		}
	}

	return domain.Event{
		ID:          e.ID,
		GrantID:     e.GrantID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		When: domain.EventWhen{
			StartTime:     e.When.StartTime,
			EndTime:       e.When.EndTime,
			StartTimezone: e.When.StartTimezone,
			EndTimezone:   e.When.EndTimezone,
			Date:          e.When.Date,
			EndDate:       e.When.EndDate,
			StartDate:     e.When.StartDate,
			Object:        e.When.Object,
		},
		Participants:  participants,
		Organizer:     organizer,
		Status:        e.Status,
		Busy:          e.Busy,
		ReadOnly:      e.ReadOnly,
		Visibility:    e.Visibility,
		Recurrence:    e.Recurrence,
		Conferencing:  conferencing,
		Reminders:     reminders,
		MasterEventID: e.MasterEventID,
		ICalUID:       e.ICalUID,
		HtmlLink:      e.HtmlLink,
		CreatedAt:     time.Unix(e.CreatedAt, 0),
		UpdatedAt:     time.Unix(e.UpdatedAt, 0),
		Object:        e.Object,
	}
}
