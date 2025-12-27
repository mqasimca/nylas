// Package nylas provides the Nylas API client implementation.
package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mqasimca/nylas/internal/domain"
)

// BuildAuthURL builds the OAuth authorization URL.
func (c *HTTPClient) BuildAuthURL(provider domain.Provider, redirectURI string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("provider", string(provider))
	params.Set("access_type", "offline")

	return fmt.Sprintf("%s/v3/connect/auth?%s", c.baseURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for tokens.
func (c *HTTPClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*domain.Grant, error) {
	// In Nylas v3, client_secret is the API key
	secret := c.clientSecret
	if secret == "" {
		secret = c.apiKey
	}

	payload := map[string]string{
		"code":          code,
		"redirect_uri":  redirectURI,
		"grant_type":    "authorization_code",
		"client_id":     c.clientID,
		"client_secret": secret,
		"code_verifier": "nylas",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v3/connect/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		GrantID      string `json:"grant_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Email        string `json:"email"`
		Provider     string `json:"provider"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.Grant{
		ID:           result.GrantID,
		Email:        result.Email,
		Provider:     domain.Provider(result.Provider),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		GrantStatus:  "valid",
	}, nil
}

// ListGrants lists all grants for the application.
func (c *HTTPClient) ListGrants(ctx context.Context) ([]domain.Grant, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v3/grants", nil)
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

// GetGrant retrieves a specific grant.
func (c *HTTPClient) GetGrant(ctx context.Context, grantID string) (*domain.Grant, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v3/grants/"+grantID, nil)
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
		return nil, domain.ErrGrantNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data domain.Grant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// RevokeGrant revokes a grant.
func (c *HTTPClient) RevokeGrant(ctx context.Context, grantID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/v3/grants/"+grantID, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrGrantNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}
