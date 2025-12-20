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

// ListInboundInboxes lists all inbound inboxes (grants with provider=inbox).
func (c *HTTPClient) ListInboundInboxes(ctx context.Context) ([]domain.InboundInbox, error) {
	// Get all grants and filter by provider=inbox
	queryURL := fmt.Sprintf("%s/v3/grants?provider=inbox", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []struct {
			ID          string          `json:"id"`
			Email       string          `json:"email"`
			Provider    string          `json:"provider"`
			GrantStatus string          `json:"grant_status"`
			CreatedAt   domain.UnixTime `json:"created_at"`
			UpdatedAt   domain.UnixTime `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	inboxes := make([]domain.InboundInbox, 0, len(result.Data))
	for _, g := range result.Data {
		// Only include inboxes with provider=inbox
		if g.Provider == "inbox" {
			inboxes = append(inboxes, domain.InboundInbox{
				ID:          g.ID,
				Email:       g.Email,
				GrantStatus: g.GrantStatus,
				CreatedAt:   g.CreatedAt,
				UpdatedAt:   g.UpdatedAt,
			})
		}
	}

	return inboxes, nil
}

// GetInboundInbox retrieves a specific inbound inbox by grant ID.
func (c *HTTPClient) GetInboundInbox(ctx context.Context, grantID string) (*domain.InboundInbox, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: inbound inbox not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data struct {
			ID          string          `json:"id"`
			Email       string          `json:"email"`
			Provider    string          `json:"provider"`
			GrantStatus string          `json:"grant_status"`
			CreatedAt   domain.UnixTime `json:"created_at"`
			UpdatedAt   domain.UnixTime `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Verify it's an inbox provider
	if result.Data.Provider != "inbox" {
		return nil, fmt.Errorf("%w: grant is not an inbound inbox (provider=%s)", domain.ErrAPIError, result.Data.Provider)
	}

	return &domain.InboundInbox{
		ID:          result.Data.ID,
		Email:       result.Data.Email,
		GrantStatus: result.Data.GrantStatus,
		CreatedAt:   result.Data.CreatedAt,
		UpdatedAt:   result.Data.UpdatedAt,
	}, nil
}

// CreateInboundInbox creates a new inbound inbox with the given email address.
// The email parameter is the local part (e.g., "support" for support@app.nylas.email).
func (c *HTTPClient) CreateInboundInbox(ctx context.Context, email string) (*domain.InboundInbox, error) {
	queryURL := fmt.Sprintf("%s/v3/grants", c.baseURL)

	// Create the request payload for custom auth with inbox provider
	payload := map[string]interface{}{
		"provider": "inbox",
		"settings": map[string]string{
			"email": email,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data struct {
			ID          string          `json:"id"`
			Email       string          `json:"email"`
			Provider    string          `json:"provider"`
			GrantStatus string          `json:"grant_status"`
			CreatedAt   domain.UnixTime `json:"created_at"`
			UpdatedAt   domain.UnixTime `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.InboundInbox{
		ID:          result.Data.ID,
		Email:       result.Data.Email,
		GrantStatus: result.Data.GrantStatus,
		CreatedAt:   result.Data.CreatedAt,
		UpdatedAt:   result.Data.UpdatedAt,
	}, nil
}

// DeleteInboundInbox deletes an inbound inbox by revoking its grant.
func (c *HTTPClient) DeleteInboundInbox(ctx context.Context, grantID string) error {
	// First verify it's an inbox provider
	inbox, err := c.GetInboundInbox(ctx, grantID)
	if err != nil {
		return err
	}
	if inbox == nil {
		return fmt.Errorf("%w: inbound inbox not found", domain.ErrAPIError)
	}

	// Use RevokeGrant to delete the inbox
	return c.RevokeGrant(ctx, grantID)
}

// GetInboundMessages retrieves messages for an inbound inbox.
func (c *HTTPClient) GetInboundMessages(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.InboundMessage, error) {
	if params == nil {
		params = &domain.MessageQueryParams{Limit: 10}
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages", c.baseURL, grantID)
	q := url.Values{}
	q.Set("limit", strconv.Itoa(params.Limit))

	if params.PageToken != "" {
		q.Set("page_token", params.PageToken)
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Subject != "" {
		q.Set("subject", params.Subject)
	}
	if params.From != "" {
		q.Set("from", params.From)
	}
	if params.Unread != nil {
		q.Set("unread", strconv.FormatBool(*params.Unread))
	}
	if params.ReceivedBefore > 0 {
		q.Set("received_before", strconv.FormatInt(params.ReceivedBefore, 10))
	}
	if params.ReceivedAfter > 0 {
		q.Set("received_after", strconv.FormatInt(params.ReceivedAfter, 10))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []messageResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertMessages(result.Data), nil
}
