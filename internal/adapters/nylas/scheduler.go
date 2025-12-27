package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

// Scheduler Configurations

// ListSchedulerConfigurations retrieves all scheduler configurations.
func (c *HTTPClient) ListSchedulerConfigurations(ctx context.Context) ([]domain.SchedulerConfiguration, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/configurations", c.baseURL)

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
		Data []domain.SchedulerConfiguration `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetSchedulerConfiguration retrieves a specific scheduler configuration.
func (c *HTTPClient) GetSchedulerConfiguration(ctx context.Context, configID string) (*domain.SchedulerConfiguration, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/configurations/%s", c.baseURL, configID)

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
		return nil, fmt.Errorf("%w: configuration not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.SchedulerConfiguration `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CreateSchedulerConfiguration creates a new scheduler configuration.
func (c *HTTPClient) CreateSchedulerConfiguration(ctx context.Context, req *domain.CreateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/configurations", c.baseURL)

	body, _ := json.Marshal(req)
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
		Data domain.SchedulerConfiguration `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateSchedulerConfiguration updates an existing scheduler configuration.
func (c *HTTPClient) UpdateSchedulerConfiguration(ctx context.Context, configID string, req *domain.UpdateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/configurations/%s", c.baseURL, configID)

	body, _ := json.Marshal(req)
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
		Data domain.SchedulerConfiguration `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteSchedulerConfiguration deletes a scheduler configuration.
func (c *HTTPClient) DeleteSchedulerConfiguration(ctx context.Context, configID string) error {
	queryURL := fmt.Sprintf("%s/v3/scheduling/configurations/%s", c.baseURL, configID)

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

// Scheduler Sessions

// CreateSchedulerSession creates a new scheduler session.
func (c *HTTPClient) CreateSchedulerSession(ctx context.Context, req *domain.CreateSchedulerSessionRequest) (*domain.SchedulerSession, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/sessions", c.baseURL)

	body, _ := json.Marshal(req)
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
		Data domain.SchedulerSession `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetSchedulerSession retrieves a scheduler session.
func (c *HTTPClient) GetSchedulerSession(ctx context.Context, sessionID string) (*domain.SchedulerSession, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/sessions/%s", c.baseURL, sessionID)

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
		return nil, fmt.Errorf("%w: session not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.SchedulerSession `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// Bookings

// ListBookings retrieves all bookings for a configuration.
func (c *HTTPClient) ListBookings(ctx context.Context, configID string) ([]domain.Booking, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/bookings", c.baseURL)
	if configID != "" {
		queryURL = fmt.Sprintf("%s?configuration_id=%s", queryURL, configID)
	}

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
		Data []domain.Booking `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetBooking retrieves a specific booking.
func (c *HTTPClient) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/bookings/%s", c.baseURL, bookingID)

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
		return nil, fmt.Errorf("%w: booking not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.Booking `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ConfirmBooking confirms a booking.
func (c *HTTPClient) ConfirmBooking(ctx context.Context, bookingID string, req *domain.ConfirmBookingRequest) (*domain.Booking, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/bookings/%s", c.baseURL, bookingID)

	body, _ := json.Marshal(req)
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
		Data domain.Booking `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// RescheduleBooking reschedules a booking.
func (c *HTTPClient) RescheduleBooking(ctx context.Context, bookingID string, req *domain.ConfirmBookingRequest) (*domain.Booking, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/bookings/%s/reschedule", c.baseURL, bookingID)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", queryURL, bytes.NewReader(body))
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
		Data domain.Booking `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CancelBooking cancels a booking.
func (c *HTTPClient) CancelBooking(ctx context.Context, bookingID string, reason string) error {
	queryURL := fmt.Sprintf("%s/v3/scheduling/bookings/%s", c.baseURL, bookingID)
	if reason != "" {
		queryURL = fmt.Sprintf("%s?reason=%s", queryURL, reason)
	}

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

// Scheduler Pages

// ListSchedulerPages retrieves all scheduler pages.
func (c *HTTPClient) ListSchedulerPages(ctx context.Context) ([]domain.SchedulerPage, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/pages", c.baseURL)

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
		Data []domain.SchedulerPage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetSchedulerPage retrieves a specific scheduler page.
func (c *HTTPClient) GetSchedulerPage(ctx context.Context, pageID string) (*domain.SchedulerPage, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/pages/%s", c.baseURL, pageID)

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
		return nil, fmt.Errorf("%w: page not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.SchedulerPage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CreateSchedulerPage creates a new scheduler page.
func (c *HTTPClient) CreateSchedulerPage(ctx context.Context, req *domain.CreateSchedulerPageRequest) (*domain.SchedulerPage, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/pages", c.baseURL)

	body, _ := json.Marshal(req)
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
		Data domain.SchedulerPage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateSchedulerPage updates an existing scheduler page.
func (c *HTTPClient) UpdateSchedulerPage(ctx context.Context, pageID string, req *domain.UpdateSchedulerPageRequest) (*domain.SchedulerPage, error) {
	queryURL := fmt.Sprintf("%s/v3/scheduling/pages/%s", c.baseURL, pageID)

	body, _ := json.Marshal(req)
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
		Data domain.SchedulerPage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteSchedulerPage deletes a scheduler page.
func (c *HTTPClient) DeleteSchedulerPage(ctx context.Context, pageID string) error {
	queryURL := fmt.Sprintf("%s/v3/scheduling/pages/%s", c.baseURL, pageID)

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
