package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// webhookResponse represents a webhook from the API.
type webhookResponse struct {
	ID                         string   `json:"id"`
	Description                string   `json:"description"`
	TriggerTypes               []string `json:"trigger_types"`
	WebhookURL                 string   `json:"webhook_url"`
	WebhookSecret              string   `json:"webhook_secret"`
	Status                     string   `json:"status"`
	NotificationEmailAddresses []string `json:"notification_email_addresses"`
	StatusUpdatedAt            int64    `json:"status_updated_at"`
	CreatedAt                  int64    `json:"created_at"`
	UpdatedAt                  int64    `json:"updated_at"`
}

// ListWebhooks retrieves all webhooks.
func (c *HTTPClient) ListWebhooks(ctx context.Context) ([]domain.Webhook, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks", c.baseURL)

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
		Data []webhookResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	webhooks := make([]domain.Webhook, len(result.Data))
	for i, w := range result.Data {
		webhooks[i] = convertWebhook(w)
	}

	return webhooks, nil
}

// GetWebhook retrieves a single webhook by ID.
func (c *HTTPClient) GetWebhook(ctx context.Context, webhookID string) (*domain.Webhook, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks/%s", c.baseURL, webhookID)

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
		return nil, fmt.Errorf("%w: webhook not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data webhookResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	webhook := convertWebhook(result.Data)
	return &webhook, nil
}

// CreateWebhook creates a new webhook.
func (c *HTTPClient) CreateWebhook(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks", c.baseURL)

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
		Data webhookResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	webhook := convertWebhook(result.Data)
	return &webhook, nil
}

// UpdateWebhook updates an existing webhook.
func (c *HTTPClient) UpdateWebhook(ctx context.Context, webhookID string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks/%s", c.baseURL, webhookID)

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
		Data webhookResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	webhook := convertWebhook(result.Data)
	return &webhook, nil
}

// DeleteWebhook deletes a webhook.
func (c *HTTPClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	queryURL := fmt.Sprintf("%s/v3/webhooks/%s", c.baseURL, webhookID)

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

// SendWebhookTestEvent sends a test event to a webhook URL.
func (c *HTTPClient) SendWebhookTestEvent(ctx context.Context, webhookURL string) error {
	queryURL := fmt.Sprintf("%s/v3/webhooks/send-test-event", c.baseURL)

	payload := map[string]string{"webhook_url": webhookURL}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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

// GetWebhookMockPayload gets a mock payload for a trigger type.
func (c *HTTPClient) GetWebhookMockPayload(ctx context.Context, triggerType string) (map[string]any, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks/mock-payload", c.baseURL)

	payload := map[string]string{"trigger_type": triggerType}
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

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// convertWebhook converts an API webhook response to domain model.
func convertWebhook(w webhookResponse) domain.Webhook {
	return domain.Webhook{
		ID:                         w.ID,
		Description:                w.Description,
		TriggerTypes:               w.TriggerTypes,
		WebhookURL:                 w.WebhookURL,
		WebhookSecret:              w.WebhookSecret,
		Status:                     w.Status,
		NotificationEmailAddresses: w.NotificationEmailAddresses,
		StatusUpdatedAt:            unixToTime(w.StatusUpdatedAt),
		CreatedAt:                  unixToTime(w.CreatedAt),
		UpdatedAt:                  unixToTime(w.UpdatedAt),
	}
}

// unixToTime converts a Unix timestamp to time.Time, returning zero time if timestamp is 0.
func unixToTime(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}
