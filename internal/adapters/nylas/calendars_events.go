package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mqasimca/nylas/internal/domain"
)

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
