// Package nylas provides the Nylas API client implementation.
package nylas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

const (
	baseURLUS = "https://api.us.nylas.com"
	baseURLEU = "https://api.eu.nylas.com"
)

// HTTPClient implements the NylasClient interface.
type HTTPClient struct {
	httpClient   *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	apiKey       string
}

// NewHTTPClient creates a new Nylas HTTP client.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURLUS,
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
