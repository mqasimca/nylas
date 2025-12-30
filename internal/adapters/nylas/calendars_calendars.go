package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

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
