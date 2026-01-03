package nylas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mqasimca/nylas/internal/domain"
)

func (c *HTTPClient) CreateVirtualCalendarGrant(ctx context.Context, email string) (*domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/connect/custom", c.baseURL)

	payload := map[string]any{
		"provider": "virtual-calendar",
		"settings": map[string]any{
			"email": email,
		},
		"scope": []string{"calendar"},
	}

	resp, err := c.doJSONRequest(ctx, "POST", queryURL, payload)
	if err != nil {
		return nil, err
	}

	var result domain.VirtualCalendarGrant
	if err := c.decodeJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListVirtualCalendarGrants lists all virtual calendar grants.
func (c *HTTPClient) ListVirtualCalendarGrants(ctx context.Context) ([]domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/grants?provider=virtual-calendar", c.baseURL)

	var result struct {
		Data []domain.VirtualCalendarGrant `json:"data"`
	}
	if err := c.doGet(ctx, queryURL, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetVirtualCalendarGrant retrieves a single virtual calendar grant by ID.
func (c *HTTPClient) GetVirtualCalendarGrant(ctx context.Context, grantID string) (*domain.VirtualCalendarGrant, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s", c.baseURL, grantID)

	var result struct {
		Data domain.VirtualCalendarGrant `json:"data"`
	}
	if err := c.doGetWithNotFound(ctx, queryURL, &result, fmt.Errorf("%w: virtual calendar grant not found", domain.ErrAPIError)); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// DeleteVirtualCalendarGrant deletes a virtual calendar grant.
func (c *HTTPClient) DeleteVirtualCalendarGrant(ctx context.Context, grantID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s", c.baseURL, grantID)
	return c.doDelete(ctx, queryURL)
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
