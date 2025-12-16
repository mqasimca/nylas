// Package nylas provides the Nylas API client implementation.
package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	payload := map[string]string{
		"code":          code,
		"redirect_uri":  redirectURI,
		"grant_type":    "authorization_code",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v3/connect/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrGrantNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetMessages retrieves recent messages for a grant (simple version).
func (c *HTTPClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
	params := &domain.MessageQueryParams{Limit: limit}
	return c.GetMessagesWithParams(ctx, grantID, params)
}

// GetMessagesWithParams retrieves messages with query parameters.
func (c *HTTPClient) GetMessagesWithParams(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
	resp, err := c.GetMessagesWithCursor(ctx, grantID, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetMessagesWithCursor retrieves messages with pagination cursor support.
func (c *HTTPClient) GetMessagesWithCursor(ctx context.Context, grantID string, params *domain.MessageQueryParams) (*domain.MessageListResponse, error) {
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
	if params.To != "" {
		q.Set("to", params.To)
	}
	if params.ThreadID != "" {
		q.Set("thread_id", params.ThreadID)
	}
	if params.Unread != nil {
		q.Set("unread", strconv.FormatBool(*params.Unread))
	}
	if params.Starred != nil {
		q.Set("starred", strconv.FormatBool(*params.Starred))
	}
	if params.HasAttachment != nil {
		q.Set("has_attachment", strconv.FormatBool(*params.HasAttachment))
	}
	if params.ReceivedBefore > 0 {
		q.Set("received_before", strconv.FormatInt(params.ReceivedBefore, 10))
	}
	if params.ReceivedAfter > 0 {
		q.Set("received_after", strconv.FormatInt(params.ReceivedAfter, 10))
	}
	if params.SearchQuery != "" {
		q.Set("q", params.SearchQuery)
	}
	if len(params.In) > 0 {
		for _, folder := range params.In {
			q.Add("in", folder)
		}
	}
	if params.Fields != "" {
		q.Set("fields", params.Fields)
	}

	queryURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data       []messageResponse `json:"data"`
		NextCursor string            `json:"next_cursor,omitempty"`
		RequestID  string            `json:"request_id,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.MessageListResponse{
		Data: convertMessages(result.Data),
		Pagination: domain.Pagination{
			NextCursor: result.NextCursor,
			HasMore:    result.NextCursor != "",
		},
	}, nil
}

// GetMessage retrieves a single message by ID.
func (c *HTTPClient) GetMessage(ctx context.Context, grantID, messageID string) (*domain.Message, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages/%s", c.baseURL, grantID, messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: message not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data messageResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	msg := convertMessage(result.Data)
	return &msg, nil
}

// SendMessage sends an email.
func (c *HTTPClient) SendMessage(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages/send", c.baseURL, grantID)

	payload := map[string]interface{}{
		"subject": req.Subject,
		"body":    req.Body,
		"to":      convertContactsToAPI(req.To),
	}

	if len(req.From) > 0 {
		payload["from"] = convertContactsToAPI(req.From)
	}
	if len(req.Cc) > 0 {
		payload["cc"] = convertContactsToAPI(req.Cc)
	}
	if len(req.Bcc) > 0 {
		payload["bcc"] = convertContactsToAPI(req.Bcc)
	}
	if len(req.ReplyTo) > 0 {
		payload["reply_to"] = convertContactsToAPI(req.ReplyTo)
	}
	if req.ReplyToMsgID != "" {
		payload["reply_to_message_id"] = req.ReplyToMsgID
	}
	if req.TrackingOpts != nil {
		payload["tracking_options"] = req.TrackingOpts
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data messageResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	msg := convertMessage(result.Data)
	return &msg, nil
}

// UpdateMessage updates message properties.
func (c *HTTPClient) UpdateMessage(ctx context.Context, grantID, messageID string, req *domain.UpdateMessageRequest) (*domain.Message, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages/%s", c.baseURL, grantID, messageID)

	payload := make(map[string]interface{})
	if req.Unread != nil {
		payload["unread"] = *req.Unread
	}
	if req.Starred != nil {
		payload["starred"] = *req.Starred
	}
	if len(req.Folders) > 0 {
		payload["folders"] = req.Folders
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data messageResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	msg := convertMessage(result.Data)
	return &msg, nil
}

// DeleteMessage deletes a message (moves to trash).
func (c *HTTPClient) DeleteMessage(ctx context.Context, grantID, messageID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages/%s", c.baseURL, grantID, messageID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetThreads retrieves threads with query parameters.
func (c *HTTPClient) GetThreads(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
	if params == nil {
		params = &domain.ThreadQueryParams{Limit: 10}
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/threads", c.baseURL, grantID)
	q := url.Values{}
	q.Set("limit", strconv.Itoa(params.Limit))

	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Subject != "" {
		q.Set("subject", params.Subject)
	}
	if params.From != "" {
		q.Set("from", params.From)
	}
	if params.To != "" {
		q.Set("to", params.To)
	}
	if params.Unread != nil {
		q.Set("unread", strconv.FormatBool(*params.Unread))
	}
	if params.Starred != nil {
		q.Set("starred", strconv.FormatBool(*params.Starred))
	}
	if params.SearchQuery != "" {
		q.Set("q", params.SearchQuery)
	}

	queryURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []threadResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertThreads(result.Data), nil
}

// GetThread retrieves a single thread by ID.
func (c *HTTPClient) GetThread(ctx context.Context, grantID, threadID string) (*domain.Thread, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/threads/%s", c.baseURL, grantID, threadID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: thread not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data threadResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	thread := convertThread(result.Data)
	return &thread, nil
}

// UpdateThread updates thread properties.
func (c *HTTPClient) UpdateThread(ctx context.Context, grantID, threadID string, req *domain.UpdateMessageRequest) (*domain.Thread, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/threads/%s", c.baseURL, grantID, threadID)

	payload := make(map[string]interface{})
	if req.Unread != nil {
		payload["unread"] = *req.Unread
	}
	if req.Starred != nil {
		payload["starred"] = *req.Starred
	}
	if len(req.Folders) > 0 {
		payload["folders"] = req.Folders
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data threadResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	thread := convertThread(result.Data)
	return &thread, nil
}

// DeleteThread deletes a thread.
func (c *HTTPClient) DeleteThread(ctx context.Context, grantID, threadID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/threads/%s", c.baseURL, grantID, threadID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetDrafts retrieves drafts for a grant.
func (c *HTTPClient) GetDrafts(ctx context.Context, grantID string, limit int) ([]domain.Draft, error) {
	if limit <= 0 {
		limit = 10
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts?limit=%d", c.baseURL, grantID, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []draftResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertDrafts(result.Data), nil
}

// GetDraft retrieves a single draft by ID.
func (c *HTTPClient) GetDraft(ctx context.Context, grantID, draftID string) (*domain.Draft, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts/%s", c.baseURL, grantID, draftID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: draft not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data draftResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	draft := convertDraft(result.Data)
	return &draft, nil
}

// CreateDraft creates a new draft.
func (c *HTTPClient) CreateDraft(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts", c.baseURL, grantID)

	payload := map[string]interface{}{
		"subject": req.Subject,
		"body":    req.Body,
	}

	if len(req.To) > 0 {
		payload["to"] = convertContactsToAPI(req.To)
	}
	if len(req.Cc) > 0 {
		payload["cc"] = convertContactsToAPI(req.Cc)
	}
	if len(req.Bcc) > 0 {
		payload["bcc"] = convertContactsToAPI(req.Bcc)
	}
	if len(req.ReplyTo) > 0 {
		payload["reply_to"] = convertContactsToAPI(req.ReplyTo)
	}
	if req.ReplyToMsgID != "" {
		payload["reply_to_message_id"] = req.ReplyToMsgID
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data draftResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	draft := convertDraft(result.Data)
	return &draft, nil
}

// UpdateDraft updates an existing draft.
func (c *HTTPClient) UpdateDraft(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts/%s", c.baseURL, grantID, draftID)

	payload := map[string]interface{}{
		"subject": req.Subject,
		"body":    req.Body,
	}

	if len(req.To) > 0 {
		payload["to"] = convertContactsToAPI(req.To)
	}
	if len(req.Cc) > 0 {
		payload["cc"] = convertContactsToAPI(req.Cc)
	}
	if len(req.Bcc) > 0 {
		payload["bcc"] = convertContactsToAPI(req.Bcc)
	}
	if len(req.ReplyTo) > 0 {
		payload["reply_to"] = convertContactsToAPI(req.ReplyTo)
	}
	if req.ReplyToMsgID != "" {
		payload["reply_to_message_id"] = req.ReplyToMsgID
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data draftResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	draft := convertDraft(result.Data)
	return &draft, nil
}

// DeleteDraft deletes a draft.
func (c *HTTPClient) DeleteDraft(ctx context.Context, grantID, draftID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts/%s", c.baseURL, grantID, draftID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// SendDraft sends a draft as an email.
func (c *HTTPClient) SendDraft(ctx context.Context, grantID, draftID string) (*domain.Message, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/drafts/%s", c.baseURL, grantID, draftID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data messageResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	msg := convertMessage(result.Data)
	return &msg, nil
}

// GetFolders retrieves all folders for a grant.
func (c *HTTPClient) GetFolders(ctx context.Context, grantID string) ([]domain.Folder, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []folderResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertFolders(result.Data), nil
}

// GetFolder retrieves a single folder by ID.
func (c *HTTPClient) GetFolder(ctx context.Context, grantID, folderID string) (*domain.Folder, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders/%s", c.baseURL, grantID, folderID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: folder not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data folderResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	folder := convertFolder(result.Data)
	return &folder, nil
}

// CreateFolder creates a new folder.
func (c *HTTPClient) CreateFolder(ctx context.Context, grantID string, req *domain.CreateFolderRequest) (*domain.Folder, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders", c.baseURL, grantID)

	payload := map[string]interface{}{
		"name": req.Name,
	}

	if req.ParentID != "" {
		payload["parent_id"] = req.ParentID
	}
	if req.BackgroundColor != "" {
		payload["background_color"] = req.BackgroundColor
	}
	if req.TextColor != "" {
		payload["text_color"] = req.TextColor
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data folderResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	folder := convertFolder(result.Data)
	return &folder, nil
}

// UpdateFolder updates an existing folder.
func (c *HTTPClient) UpdateFolder(ctx context.Context, grantID, folderID string, req *domain.UpdateFolderRequest) (*domain.Folder, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders/%s", c.baseURL, grantID, folderID)

	payload := make(map[string]interface{})
	if req.Name != "" {
		payload["name"] = req.Name
	}
	if req.ParentID != "" {
		payload["parent_id"] = req.ParentID
	}
	if req.BackgroundColor != "" {
		payload["background_color"] = req.BackgroundColor
	}
	if req.TextColor != "" {
		payload["text_color"] = req.TextColor
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data folderResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	folder := convertFolder(result.Data)
	return &folder, nil
}

// DeleteFolder deletes a folder.
func (c *HTTPClient) DeleteFolder(ctx context.Context, grantID, folderID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders/%s", c.baseURL, grantID, folderID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetAttachment retrieves attachment metadata.
func (c *HTTPClient) GetAttachment(ctx context.Context, grantID, messageID, attachmentID string) (*domain.Attachment, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/attachments/%s?message_id=%s", c.baseURL, grantID, attachmentID, messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: attachment not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data attachmentResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.Attachment{
		ID:          result.Data.ID,
		GrantID:     result.Data.GrantID,
		Filename:    result.Data.Filename,
		ContentType: result.Data.ContentType,
		Size:        result.Data.Size,
		ContentID:   result.Data.ContentID,
		IsInline:    result.Data.IsInline,
	}, nil
}

// DownloadAttachment downloads attachment content.
func (c *HTTPClient) DownloadAttachment(ctx context.Context, grantID, messageID, attachmentID string) (io.ReadCloser, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/attachments/%s/download?message_id=%s", c.baseURL, grantID, attachmentID, messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("%w: attachment not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.parseError(resp)
	}

	return resp.Body, nil
}

// =============================================================================
// Calendar Operations
// =============================================================================

// GetCalendars retrieves all calendars for a grant.
func (c *HTTPClient) GetCalendars(ctx context.Context, grantID string) ([]domain.Calendar, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

// GetEvents retrieves events for a calendar.
func (c *HTTPClient) GetEvents(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) ([]domain.Event, error) {
	resp, err := c.GetEventsWithCursor(ctx, grantID, calendarID, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetEventsWithCursor retrieves events with pagination cursor support.
func (c *HTTPClient) GetEventsWithCursor(ctx context.Context, grantID, calendarID string, params *domain.EventQueryParams) (*domain.EventListResponse, error) {
	if params == nil {
		params = &domain.EventQueryParams{Limit: 50}
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/events", c.baseURL, grantID)
	q := url.Values{}
	q.Set("limit", strconv.Itoa(params.Limit))
	q.Set("calendar_id", calendarID)

	if params.PageToken != "" {
		q.Set("page_token", params.PageToken)
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
	if params.Start > 0 {
		q.Set("start", strconv.FormatInt(params.Start, 10))
	}
	if params.End > 0 {
		q.Set("end", strconv.FormatInt(params.End, 10))
	}
	if params.Busy != nil {
		q.Set("busy", strconv.FormatBool(*params.Busy))
	}
	if params.OrderBy != "" {
		q.Set("order_by", params.OrderBy)
	}
	if params.ExpandRecurring {
		q.Set("expand_recurring", "true")
	}

	queryURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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
		"busy":  req.Busy,
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

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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
		payload["when"] = *req.When
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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetFreeBusy retrieves free/busy information for calendars.
func (c *HTTPClient) GetFreeBusy(ctx context.Context, grantID string, freeBusyReq *domain.FreeBusyRequest) (*domain.FreeBusyResponse, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/calendars/free-busy", c.baseURL, grantID)

	body, err := json.Marshal(freeBusyReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result domain.FreeBusyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetAvailability finds available meeting times across participants.
func (c *HTTPClient) GetAvailability(ctx context.Context, availReq *domain.AvailabilityRequest) (*domain.AvailabilityResponse, error) {
	queryURL := fmt.Sprintf("%s/v3/calendars/availability", c.baseURL)

	body, err := json.Marshal(availReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result domain.AvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *HTTPClient) setAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

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

// Response types for API unmarshaling
type messageResponse struct {
	ID        string `json:"id"`
	GrantID   string `json:"grant_id"`
	ThreadID  string `json:"thread_id"`
	Subject   string `json:"subject"`
	From      []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"from"`
	To []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"to"`
	Cc []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"cc"`
	Bcc []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"bcc"`
	ReplyTo []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"reply_to"`
	Body        string   `json:"body"`
	Snippet     string   `json:"snippet"`
	Date        int64    `json:"date"`
	Unread      bool     `json:"unread"`
	Starred     bool     `json:"starred"`
	Folders     []string `json:"folders"`
	Attachments []struct {
		ID          string `json:"id"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size"`
		ContentID   string `json:"content_id"`
		IsInline    bool   `json:"is_inline"`
	} `json:"attachments"`
	CreatedAt int64  `json:"created_at"`
	Object    string `json:"object"`
}

type threadResponse struct {
	ID                    string `json:"id"`
	GrantID               string `json:"grant_id"`
	HasAttachments        bool   `json:"has_attachments"`
	HasDrafts             bool   `json:"has_drafts"`
	Starred               bool   `json:"starred"`
	Unread                bool   `json:"unread"`
	EarliestMessageDate   int64  `json:"earliest_message_date"`
	LatestMessageRecvDate int64  `json:"latest_message_received_date"`
	LatestMessageSentDate int64  `json:"latest_message_sent_date"`
	Participants          []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"participants"`
	MessageIDs []string `json:"message_ids"`
	DraftIDs   []string `json:"draft_ids"`
	FolderIDs  []string `json:"folders"`
	Snippet    string   `json:"snippet"`
	Subject    string   `json:"subject"`
}

type draftResponse struct {
	ID        string `json:"id"`
	GrantID   string `json:"grant_id"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	From      []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"from"`
	To []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"to"`
	Cc []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"cc"`
	Bcc []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"bcc"`
	ReplyTo []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"reply_to"`
	ReplyToMsgID string `json:"reply_to_message_id"`
	ThreadID     string `json:"thread_id"`
	Attachments  []struct {
		ID          string `json:"id"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size"`
	} `json:"attachments"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

type folderResponse struct {
	ID              string   `json:"id"`
	GrantID         string   `json:"grant_id"`
	Name            string   `json:"name"`
	SystemFolder    string   `json:"system_folder"`
	ParentID        string   `json:"parent_id"`
	BackgroundColor string   `json:"background_color"`
	TextColor       string   `json:"text_color"`
	TotalCount      int      `json:"total_count"`
	UnreadCount     int      `json:"unread_count"`
	ChildIDs        []string `json:"child_ids"`
	Attributes      []string `json:"attributes"`
}

type attachmentResponse struct {
	ID          string `json:"id"`
	GrantID     string `json:"grant_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	ContentID   string `json:"content_id"`
	IsInline    bool   `json:"is_inline"`
}

// Conversion helpers
func convertMessages(msgs []messageResponse) []domain.Message {
	result := make([]domain.Message, len(msgs))
	for i, m := range msgs {
		result[i] = convertMessage(m)
	}
	return result
}

func convertMessage(m messageResponse) domain.Message {
	from := make([]domain.EmailParticipant, len(m.From))
	for j, f := range m.From {
		from[j] = domain.EmailParticipant{Name: f.Name, Email: f.Email}
	}
	to := make([]domain.EmailParticipant, len(m.To))
	for j, t := range m.To {
		to[j] = domain.EmailParticipant{Name: t.Name, Email: t.Email}
	}
	cc := make([]domain.EmailParticipant, len(m.Cc))
	for j, c := range m.Cc {
		cc[j] = domain.EmailParticipant{Name: c.Name, Email: c.Email}
	}
	bcc := make([]domain.EmailParticipant, len(m.Bcc))
	for j, b := range m.Bcc {
		bcc[j] = domain.EmailParticipant{Name: b.Name, Email: b.Email}
	}
	replyTo := make([]domain.EmailParticipant, len(m.ReplyTo))
	for j, r := range m.ReplyTo {
		replyTo[j] = domain.EmailParticipant{Name: r.Name, Email: r.Email}
	}
	attachments := make([]domain.Attachment, len(m.Attachments))
	for j, a := range m.Attachments {
		attachments[j] = domain.Attachment{
			ID:          a.ID,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
			ContentID:   a.ContentID,
			IsInline:    a.IsInline,
		}
	}

	return domain.Message{
		ID:          m.ID,
		GrantID:     m.GrantID,
		ThreadID:    m.ThreadID,
		Subject:     m.Subject,
		From:        from,
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		ReplyTo:     replyTo,
		Body:        m.Body,
		Snippet:     m.Snippet,
		Date:        time.Unix(m.Date, 0),
		Unread:      m.Unread,
		Starred:     m.Starred,
		Folders:     m.Folders,
		Attachments: attachments,
		CreatedAt:   time.Unix(m.CreatedAt, 0),
		Object:      m.Object,
	}
}

func convertThreads(threads []threadResponse) []domain.Thread {
	result := make([]domain.Thread, len(threads))
	for i, t := range threads {
		result[i] = convertThread(t)
	}
	return result
}

func convertThread(t threadResponse) domain.Thread {
	participants := make([]domain.EmailParticipant, len(t.Participants))
	for j, p := range t.Participants {
		participants[j] = domain.EmailParticipant{Name: p.Name, Email: p.Email}
	}

	return domain.Thread{
		ID:                    t.ID,
		GrantID:               t.GrantID,
		HasAttachments:        t.HasAttachments,
		HasDrafts:             t.HasDrafts,
		Starred:               t.Starred,
		Unread:                t.Unread,
		EarliestMessageDate:   time.Unix(t.EarliestMessageDate, 0),
		LatestMessageRecvDate: time.Unix(t.LatestMessageRecvDate, 0),
		LatestMessageSentDate: time.Unix(t.LatestMessageSentDate, 0),
		Participants:          participants,
		MessageIDs:            t.MessageIDs,
		DraftIDs:              t.DraftIDs,
		FolderIDs:             t.FolderIDs,
		Snippet:               t.Snippet,
		Subject:               t.Subject,
	}
}

func convertDrafts(drafts []draftResponse) []domain.Draft {
	result := make([]domain.Draft, len(drafts))
	for i, d := range drafts {
		result[i] = convertDraft(d)
	}
	return result
}

func convertDraft(d draftResponse) domain.Draft {
	from := make([]domain.EmailParticipant, len(d.From))
	for j, f := range d.From {
		from[j] = domain.EmailParticipant{Name: f.Name, Email: f.Email}
	}
	to := make([]domain.EmailParticipant, len(d.To))
	for j, t := range d.To {
		to[j] = domain.EmailParticipant{Name: t.Name, Email: t.Email}
	}
	cc := make([]domain.EmailParticipant, len(d.Cc))
	for j, c := range d.Cc {
		cc[j] = domain.EmailParticipant{Name: c.Name, Email: c.Email}
	}
	bcc := make([]domain.EmailParticipant, len(d.Bcc))
	for j, b := range d.Bcc {
		bcc[j] = domain.EmailParticipant{Name: b.Name, Email: b.Email}
	}
	replyTo := make([]domain.EmailParticipant, len(d.ReplyTo))
	for j, r := range d.ReplyTo {
		replyTo[j] = domain.EmailParticipant{Name: r.Name, Email: r.Email}
	}
	attachments := make([]domain.Attachment, len(d.Attachments))
	for j, a := range d.Attachments {
		attachments[j] = domain.Attachment{
			ID:          a.ID,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
		}
	}

	return domain.Draft{
		ID:           d.ID,
		GrantID:      d.GrantID,
		Subject:      d.Subject,
		Body:         d.Body,
		From:         from,
		To:           to,
		Cc:           cc,
		Bcc:          bcc,
		ReplyTo:      replyTo,
		ReplyToMsgID: d.ReplyToMsgID,
		ThreadID:     d.ThreadID,
		Attachments:  attachments,
		CreatedAt:    time.Unix(d.CreatedAt, 0),
		UpdatedAt:    time.Unix(d.UpdatedAt, 0),
	}
}

func convertFolders(folders []folderResponse) []domain.Folder {
	result := make([]domain.Folder, len(folders))
	for i, f := range folders {
		result[i] = convertFolder(f)
	}
	return result
}

func convertFolder(f folderResponse) domain.Folder {
	return domain.Folder{
		ID:              f.ID,
		GrantID:         f.GrantID,
		Name:            f.Name,
		SystemFolder:    f.SystemFolder,
		ParentID:        f.ParentID,
		BackgroundColor: f.BackgroundColor,
		TextColor:       f.TextColor,
		TotalCount:      f.TotalCount,
		UnreadCount:     f.UnreadCount,
		ChildIDs:        f.ChildIDs,
		Attributes:      f.Attributes,
	}
}

func convertContactsToAPI(contacts []domain.EmailParticipant) []map[string]string {
	result := make([]map[string]string, len(contacts))
	for i, c := range contacts {
		result[i] = map[string]string{
			"name":  c.Name,
			"email": c.Email,
		}
	}
	return result
}

// Calendar response types
type calendarResponse struct {
	ID          string `json:"id"`
	GrantID     string `json:"grant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Timezone    string `json:"timezone"`
	ReadOnly    bool   `json:"read_only"`
	IsPrimary   bool   `json:"is_primary"`
	IsOwner     bool   `json:"is_owner"`
	HexColor    string `json:"hex_color"`
	Object      string `json:"object"`
}

type eventResponse struct {
	ID           string `json:"id"`
	GrantID      string `json:"grant_id"`
	CalendarID   string `json:"calendar_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	When         struct {
		StartTime     int64  `json:"start_time,omitempty"`
		EndTime       int64  `json:"end_time,omitempty"`
		StartTimezone string `json:"start_timezone,omitempty"`
		EndTimezone   string `json:"end_timezone,omitempty"`
		Date          string `json:"date,omitempty"`
		EndDate       string `json:"end_date,omitempty"`
		StartDate     string `json:"start_date,omitempty"`
		Object        string `json:"object,omitempty"`
	} `json:"when"`
	Participants []struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Status  string `json:"status"`
		Comment string `json:"comment"`
	} `json:"participants"`
	Organizer *struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Status  string `json:"status"`
		Comment string `json:"comment"`
	} `json:"organizer"`
	Status       string   `json:"status"`
	Busy         bool     `json:"busy"`
	ReadOnly     bool     `json:"read_only"`
	Visibility   string   `json:"visibility"`
	Recurrence   []string `json:"recurrence"`
	Conferencing *struct {
		Provider string `json:"provider"`
		Details  *struct {
			URL         string   `json:"url"`
			MeetingCode string   `json:"meeting_code"`
			Password    string   `json:"password"`
			Phone       []string `json:"phone"`
		} `json:"details"`
	} `json:"conferencing"`
	Reminders *struct {
		UseDefault bool `json:"use_default"`
		Overrides  []struct {
			ReminderMinutes int    `json:"reminder_minutes"`
			ReminderMethod  string `json:"reminder_method"`
		} `json:"overrides"`
	} `json:"reminders"`
	MasterEventID string `json:"master_event_id"`
	ICalUID       string `json:"ical_uid"`
	HtmlLink      string `json:"html_link"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
	Object        string `json:"object"`
}

func convertCalendars(cals []calendarResponse) []domain.Calendar {
	result := make([]domain.Calendar, len(cals))
	for i, c := range cals {
		result[i] = convertCalendar(c)
	}
	return result
}

func convertCalendar(c calendarResponse) domain.Calendar {
	return domain.Calendar{
		ID:          c.ID,
		GrantID:     c.GrantID,
		Name:        c.Name,
		Description: c.Description,
		Location:    c.Location,
		Timezone:    c.Timezone,
		ReadOnly:    c.ReadOnly,
		IsPrimary:   c.IsPrimary,
		IsOwner:     c.IsOwner,
		HexColor:    c.HexColor,
		Object:      c.Object,
	}
}

func convertEvents(events []eventResponse) []domain.Event {
	result := make([]domain.Event, len(events))
	for i, e := range events {
		result[i] = convertEvent(e)
	}
	return result
}

func convertEvent(e eventResponse) domain.Event {
	participants := make([]domain.Participant, len(e.Participants))
	for j, p := range e.Participants {
		participants[j] = domain.Participant{
			Name:    p.Name,
			Email:   p.Email,
			Status:  p.Status,
			Comment: p.Comment,
		}
	}

	var organizer *domain.Participant
	if e.Organizer != nil {
		organizer = &domain.Participant{
			Name:    e.Organizer.Name,
			Email:   e.Organizer.Email,
			Status:  e.Organizer.Status,
			Comment: e.Organizer.Comment,
		}
	}

	var conferencing *domain.Conferencing
	if e.Conferencing != nil {
		conferencing = &domain.Conferencing{
			Provider: e.Conferencing.Provider,
		}
		if e.Conferencing.Details != nil {
			conferencing.Details = &domain.ConferencingDetails{
				URL:         e.Conferencing.Details.URL,
				MeetingCode: e.Conferencing.Details.MeetingCode,
				Password:    e.Conferencing.Details.Password,
				Phone:       e.Conferencing.Details.Phone,
			}
		}
	}

	var reminders *domain.Reminders
	if e.Reminders != nil {
		overrides := make([]domain.Reminder, len(e.Reminders.Overrides))
		for j, o := range e.Reminders.Overrides {
			overrides[j] = domain.Reminder{
				ReminderMinutes: o.ReminderMinutes,
				ReminderMethod:  o.ReminderMethod,
			}
		}
		reminders = &domain.Reminders{
			UseDefault: e.Reminders.UseDefault,
			Overrides:  overrides,
		}
	}

	return domain.Event{
		ID:          e.ID,
		GrantID:     e.GrantID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		When: domain.EventWhen{
			StartTime:     e.When.StartTime,
			EndTime:       e.When.EndTime,
			StartTimezone: e.When.StartTimezone,
			EndTimezone:   e.When.EndTimezone,
			Date:          e.When.Date,
			EndDate:       e.When.EndDate,
			StartDate:     e.When.StartDate,
			Object:        e.When.Object,
		},
		Participants:  participants,
		Organizer:     organizer,
		Status:        e.Status,
		Busy:          e.Busy,
		ReadOnly:      e.ReadOnly,
		Visibility:    e.Visibility,
		Recurrence:    e.Recurrence,
		Conferencing:  conferencing,
		Reminders:     reminders,
		MasterEventID: e.MasterEventID,
		ICalUID:       e.ICalUID,
		HtmlLink:      e.HtmlLink,
		CreatedAt:     time.Unix(e.CreatedAt, 0),
		UpdatedAt:     time.Unix(e.UpdatedAt, 0),
		Object:        e.Object,
	}
}

// =============================================================================
// Contact Operations
// =============================================================================

// contactResponse represents a contact from the API.
type contactResponse struct {
	ID                string                   `json:"id"`
	GrantID           string                   `json:"grant_id"`
	Object            string                   `json:"object"`
	GivenName         string                   `json:"given_name"`
	MiddleName        string                   `json:"middle_name"`
	Surname           string                   `json:"surname"`
	Suffix            string                   `json:"suffix"`
	Nickname          string                   `json:"nickname"`
	Birthday          string                   `json:"birthday"`
	CompanyName       string                   `json:"company_name"`
	JobTitle          string                   `json:"job_title"`
	ManagerName       string                   `json:"manager_name"`
	Notes             string                   `json:"notes"`
	PictureURL        string                   `json:"picture_url"`
	Emails            []domain.ContactEmail    `json:"emails"`
	PhoneNumbers      []domain.ContactPhone    `json:"phone_numbers"`
	WebPages          []domain.ContactWebPage  `json:"web_pages"`
	IMAddresses       []domain.ContactIM       `json:"im_addresses"`
	PhysicalAddresses []domain.ContactAddress  `json:"physical_addresses"`
	Groups            []domain.ContactGroupInfo `json:"groups"`
	Source            string                   `json:"source"`
}

// contactGroupResponse represents a contact group from the API.
type contactGroupResponse struct {
	ID      string `json:"id"`
	GrantID string `json:"grant_id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Object  string `json:"object"`
}

// GetContacts retrieves contacts for a grant.
func (c *HTTPClient) GetContacts(ctx context.Context, grantID string, params *domain.ContactQueryParams) ([]domain.Contact, error) {
	result, err := c.GetContactsWithCursor(ctx, grantID, params)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetContactsWithCursor retrieves contacts with pagination cursor.
func (c *HTTPClient) GetContactsWithCursor(ctx context.Context, grantID string, params *domain.ContactQueryParams) (*domain.ContactListResponse, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts", c.baseURL, grantID)

	queryParams := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			queryParams.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.PageToken != "" {
			queryParams.Set("page_token", params.PageToken)
		}
		if params.Email != "" {
			queryParams.Set("email", params.Email)
		}
		if params.PhoneNumber != "" {
			queryParams.Set("phone_number", params.PhoneNumber)
		}
		if params.Source != "" {
			queryParams.Set("source", params.Source)
		}
		if params.Group != "" {
			queryParams.Set("group", params.Group)
		}
		if params.Recurse {
			queryParams.Set("recurse", "true")
		}
	}
	if len(queryParams) > 0 {
		queryURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data       []contactResponse `json:"data"`
		NextCursor string            `json:"next_cursor"`
		RequestID  string            `json:"request_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	contacts := make([]domain.Contact, len(result.Data))
	for i, cr := range result.Data {
		contacts[i] = convertContact(cr)
	}

	return &domain.ContactListResponse{
		Data: contacts,
		Pagination: domain.Pagination{
			NextCursor: result.NextCursor,
		},
	}, nil
}

// GetContact retrieves a single contact by ID.
func (c *HTTPClient) GetContact(ctx context.Context, grantID, contactID string) (*domain.Contact, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts/%s", c.baseURL, grantID, contactID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: contact not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data contactResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	contact := convertContact(result.Data)
	return &contact, nil
}

// CreateContact creates a new contact.
func (c *HTTPClient) CreateContact(ctx context.Context, grantID string, req *domain.CreateContactRequest) (*domain.Contact, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts", c.baseURL, grantID)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data contactResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	contact := convertContact(result.Data)
	return &contact, nil
}

// UpdateContact updates an existing contact.
func (c *HTTPClient) UpdateContact(ctx context.Context, grantID, contactID string, req *domain.UpdateContactRequest) (*domain.Contact, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts/%s", c.baseURL, grantID, contactID)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data contactResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	contact := convertContact(result.Data)
	return &contact, nil
}

// DeleteContact deletes a contact.
func (c *HTTPClient) DeleteContact(ctx context.Context, grantID, contactID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts/%s", c.baseURL, grantID, contactID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetContactGroups retrieves contact groups for a grant.
func (c *HTTPClient) GetContactGroups(ctx context.Context, grantID string) ([]domain.ContactGroup, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/contacts/groups", c.baseURL, grantID)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data []contactGroupResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	groups := make([]domain.ContactGroup, len(result.Data))
	for i, g := range result.Data {
		groups[i] = domain.ContactGroup{
			ID:      g.ID,
			GrantID: g.GrantID,
			Name:    g.Name,
			Path:    g.Path,
			Object:  g.Object,
		}
	}

	return groups, nil
}

// convertContact converts an API contact response to domain model.
func convertContact(c contactResponse) domain.Contact {
	return domain.Contact{
		ID:                c.ID,
		GrantID:           c.GrantID,
		Object:            c.Object,
		GivenName:         c.GivenName,
		MiddleName:        c.MiddleName,
		Surname:           c.Surname,
		Suffix:            c.Suffix,
		Nickname:          c.Nickname,
		Birthday:          c.Birthday,
		CompanyName:       c.CompanyName,
		JobTitle:          c.JobTitle,
		ManagerName:       c.ManagerName,
		Notes:             c.Notes,
		PictureURL:        c.PictureURL,
		Emails:            c.Emails,
		PhoneNumbers:      c.PhoneNumbers,
		WebPages:          c.WebPages,
		IMAddresses:       c.IMAddresses,
		PhysicalAddresses: c.PhysicalAddresses,
		Groups:            c.Groups,
		Source:            c.Source,
	}
}

// =============================================================================
// Webhook Operations
// =============================================================================

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetWebhookMockPayload gets a mock payload for a trigger type.
func (c *HTTPClient) GetWebhookMockPayload(ctx context.Context, triggerType string) (map[string]interface{}, error) {
	queryURL := fmt.Sprintf("%s/v3/webhooks/mock-payload", c.baseURL)

	payload := map[string]string{"trigger_type": triggerType}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result map[string]interface{}
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
		StatusUpdatedAt:            time.Unix(w.StatusUpdatedAt, 0),
		CreatedAt:                  time.Unix(w.CreatedAt, 0),
		UpdatedAt:                  time.Unix(w.UpdatedAt, 0),
	}
}
