package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

// Admin Applications

// ListApplications retrieves all applications.
func (c *HTTPClient) ListApplications(ctx context.Context) ([]domain.Application, error) {
	queryURL := fmt.Sprintf("%s/v3/applications", c.baseURL)

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

	// Read body once
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to decode as an array first
	var multiResult struct {
		Data []domain.Application `json:"data"`
	}
	if err := json.Unmarshal(body, &multiResult); err == nil && len(multiResult.Data) > 0 {
		return multiResult.Data, nil
	}

	// Try to decode as a single application object (v3 API returns single app)
	var singleResult struct {
		Data domain.Application `json:"data"`
	}
	if err := json.Unmarshal(body, &singleResult); err == nil {
		// Check if we got valid application data (ID or ApplicationID set)
		if singleResult.Data.ID != "" || singleResult.Data.ApplicationID != "" {
			return []domain.Application{singleResult.Data}, nil
		}
	}

	// If both fail, return error with response body for debugging
	return nil, fmt.Errorf("failed to decode applications response: %s", string(body))
}

// GetApplication retrieves a specific application.
func (c *HTTPClient) GetApplication(ctx context.Context, appID string) (*domain.Application, error) {
	queryURL := fmt.Sprintf("%s/v3/applications/%s", c.baseURL, appID)

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
		return nil, fmt.Errorf("%w: application not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.Application `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CreateApplication creates a new application.
func (c *HTTPClient) CreateApplication(ctx context.Context, req *domain.CreateApplicationRequest) (*domain.Application, error) {
	queryURL := fmt.Sprintf("%s/v3/applications", c.baseURL)

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
		Data domain.Application `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateApplication updates an existing application.
func (c *HTTPClient) UpdateApplication(ctx context.Context, appID string, req *domain.UpdateApplicationRequest) (*domain.Application, error) {
	queryURL := fmt.Sprintf("%s/v3/applications/%s", c.baseURL, appID)

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
		Data domain.Application `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteApplication deletes an application.
func (c *HTTPClient) DeleteApplication(ctx context.Context, appID string) error {
	queryURL := fmt.Sprintf("%s/v3/applications/%s", c.baseURL, appID)

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

// Admin Connectors

// ListConnectors retrieves all connectors.
func (c *HTTPClient) ListConnectors(ctx context.Context) ([]domain.Connector, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors", c.baseURL)

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
		Data []domain.Connector `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetConnector retrieves a specific connector.
func (c *HTTPClient) GetConnector(ctx context.Context, connectorID string) (*domain.Connector, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors/%s", c.baseURL, connectorID)

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
		return nil, fmt.Errorf("%w: connector not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.Connector `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CreateConnector creates a new connector.
func (c *HTTPClient) CreateConnector(ctx context.Context, req *domain.CreateConnectorRequest) (*domain.Connector, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors", c.baseURL)

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
		Data domain.Connector `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateConnector updates an existing connector.
func (c *HTTPClient) UpdateConnector(ctx context.Context, connectorID string, req *domain.UpdateConnectorRequest) (*domain.Connector, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors/%s", c.baseURL, connectorID)

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
		Data domain.Connector `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteConnector deletes a connector.
func (c *HTTPClient) DeleteConnector(ctx context.Context, connectorID string) error {
	queryURL := fmt.Sprintf("%s/v3/connectors/%s", c.baseURL, connectorID)

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

// Admin Credentials

// ListCredentials retrieves all credentials for a connector.
func (c *HTTPClient) ListCredentials(ctx context.Context, connectorID string) ([]domain.ConnectorCredential, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors/%s/credentials", c.baseURL, connectorID)

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
		Data []domain.ConnectorCredential `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetCredential retrieves a specific credential.
func (c *HTTPClient) GetCredential(ctx context.Context, credentialID string) (*domain.ConnectorCredential, error) {
	queryURL := fmt.Sprintf("%s/v3/credentials/%s", c.baseURL, credentialID)

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
		return nil, fmt.Errorf("%w: credential not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.ConnectorCredential `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CreateCredential creates a new credential.
func (c *HTTPClient) CreateCredential(ctx context.Context, connectorID string, req *domain.CreateCredentialRequest) (*domain.ConnectorCredential, error) {
	queryURL := fmt.Sprintf("%s/v3/connectors/%s/credentials", c.baseURL, connectorID)

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
		Data domain.ConnectorCredential `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateCredential updates an existing credential.
func (c *HTTPClient) UpdateCredential(ctx context.Context, credentialID string, req *domain.UpdateCredentialRequest) (*domain.ConnectorCredential, error) {
	queryURL := fmt.Sprintf("%s/v3/credentials/%s", c.baseURL, credentialID)

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
		Data domain.ConnectorCredential `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteCredential deletes a credential.
func (c *HTTPClient) DeleteCredential(ctx context.Context, credentialID string) error {
	queryURL := fmt.Sprintf("%s/v3/credentials/%s", c.baseURL, credentialID)

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

// Admin Grant Operations

// ListAllGrants retrieves all grants with optional filtering.
func (c *HTTPClient) ListAllGrants(ctx context.Context, params *domain.GrantsQueryParams) ([]domain.Grant, error) {
	queryURL := fmt.Sprintf("%s/v3/grants", c.baseURL)

	if params != nil {
		query := ""
		if params.Limit > 0 {
			query = fmt.Sprintf("%s&limit=%d", query, params.Limit)
		}
		if params.Offset > 0 {
			query = fmt.Sprintf("%s&offset=%d", query, params.Offset)
		}
		if params.ConnectorID != "" {
			query = fmt.Sprintf("%s&connector_id=%s", query, params.ConnectorID)
		}
		if params.Status != "" {
			query = fmt.Sprintf("%s&status=%s", query, params.Status)
		}
		if query != "" {
			queryURL = fmt.Sprintf("%s?%s", queryURL, query[1:]) // Remove leading &
		}
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
		Data []domain.Grant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetGrantStats retrieves grant statistics.
func (c *HTTPClient) GetGrantStats(ctx context.Context) (*domain.GrantStats, error) {
	// Get all grants
	grants, err := c.ListAllGrants(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := &domain.GrantStats{
		Total:      len(grants),
		ByProvider: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	for _, grant := range grants {
		// Count by provider
		stats.ByProvider[string(grant.Provider)]++

		// Count by status
		if grant.GrantStatus != "" {
			stats.ByStatus[grant.GrantStatus]++
			switch grant.GrantStatus {
			case "valid":
				stats.Valid++
			case "invalid":
				stats.Invalid++
			}
		}
	}

	return stats, nil
}
