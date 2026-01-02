// Package nylas provides the Nylas API client implementation.
package nylas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"golang.org/x/time/rate"
)

const (
	baseURLUS = "https://api.us.nylas.com"
	baseURLEU = "https://api.eu.nylas.com"

	// defaultRequestTimeout is the default timeout for individual API requests
	defaultRequestTimeout = 30 * time.Second

	// defaultRateLimit is the default rate limit (requests per second)
	// Set to 10 requests per second to avoid API quota exhaustion
	defaultRateLimit = 10
)

// HTTPClient implements the NylasClient interface.
type HTTPClient struct {
	httpClient     *http.Client
	baseURL        string
	clientID       string
	clientSecret   string
	apiKey         string
	rateLimiter    *rate.Limiter
	requestTimeout time.Duration
}

// NewHTTPClient creates a new Nylas HTTP client with rate limiting.
// Rate limiting prevents API quota exhaustion and temporary account suspension.
// Default: 10 requests/second with burst capacity of 20 requests.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			// Remove global timeout since we use per-request context timeouts
			Timeout: 0,
		},
		baseURL: baseURLUS,
		// Create token bucket rate limiter: 10 requests/second, burst of 20
		rateLimiter:    rate.NewLimiter(rate.Limit(defaultRateLimit), defaultRateLimit*2),
		requestTimeout: defaultRequestTimeout,
	}
}

// SetRegion sets the API region (us or eu).
func (c *HTTPClient) SetRegion(region string) {
	if region == "eu" {
		c.baseURL = baseURLEU
	} else {
		c.baseURL = baseURLUS
	}
}

// SetCredentials sets the API credentials.
func (c *HTTPClient) SetCredentials(clientID, clientSecret, apiKey string) {
	c.clientID = clientID
	c.clientSecret = clientSecret
	c.apiKey = apiKey
}

// SetBaseURL sets the base URL (for testing purposes).
func (c *HTTPClient) SetBaseURL(url string) {
	c.baseURL = url
}

// setAuthHeader sets the authorization header on the request.
func (c *HTTPClient) setAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// parseError parses an error response from the API.
func (c *HTTPClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return fmt.Errorf("%w: %s", domain.ErrAPIError, errResp.Error.Message)
	}

	return fmt.Errorf("%w: status %d", domain.ErrAPIError, resp.StatusCode)
}

// validateGrantID validates that a grant ID is not empty.
func validateGrantID(grantID string) error {
	if grantID == "" {
		return fmt.Errorf("%w: grant ID is required", domain.ErrInvalidInput)
	}
	return nil
}

// validateCalendarID validates that a calendar ID is not empty.
func validateCalendarID(calendarID string) error {
	if calendarID == "" {
		return fmt.Errorf("%w: calendar ID is required", domain.ErrInvalidInput)
	}
	return nil
}

// validateMessageID validates that a message ID is not empty.
func validateMessageID(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("%w: message ID is required", domain.ErrInvalidInput)
	}
	return nil
}

// validateEventID validates that an event ID is not empty.
func validateEventID(eventID string) error {
	if eventID == "" {
		return fmt.Errorf("%w: event ID is required", domain.ErrInvalidInput)
	}
	return nil
}

// getRequestID extracts the request ID from response headers.
func getRequestID(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	// Nylas uses X-Request-Id header
	if id := resp.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	return resp.Header.Get("Request-Id")
}

// ensureContext ensures a context has a timeout.
// If the context already has a deadline, it's returned as-is.
// Otherwise, a new context with the default timeout is created.
func (c *HTTPClient) ensureContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		// Context already has timeout, use as-is
		return ctx, func() {}
	}
	// Add default timeout
	return context.WithTimeout(ctx, c.requestTimeout)
}

// doRequest executes an HTTP request with rate limiting and timeout.
// This method applies rate limiting before making the request and ensures
// the context has a timeout to prevent hanging requests.
func (c *HTTPClient) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Apply rate limiting - wait for permission to proceed
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	// Ensure context has timeout
	ctxWithTimeout, cancel := c.ensureContext(ctx)
	defer cancel()

	// Update request context
	req = req.WithContext(ctxWithTimeout)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}

	return resp, nil
}
