package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/util"
)

// messageResponse represents an API message response.
type messageResponse struct {
	ID       string `json:"id"`
	GrantID  string `json:"grant_id"`
	ThreadID string `json:"thread_id"`
	Subject  string `json:"subject"`
	From     []struct {
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
	Metadata  map[string]string `json:"metadata"`
	CreatedAt int64             `json:"created_at"`
	Object    string            `json:"object"`
}

// GetMessages retrieves recent messages for a grant (simple version).
func (c *HTTPClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
	if err := validateGrantID(grantID); err != nil {
		return nil, err
	}
	params := &domain.MessageQueryParams{Limit: limit}
	return c.GetMessagesWithParams(ctx, grantID, params)
}

// GetMessagesWithParams retrieves messages with query parameters.
func (c *HTTPClient) GetMessagesWithParams(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
	if err := validateGrantID(grantID); err != nil {
		return nil, err
	}
	resp, err := c.GetMessagesWithCursor(ctx, grantID, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetMessagesWithCursor retrieves messages with pagination cursor support.
func (c *HTTPClient) GetMessagesWithCursor(ctx context.Context, grantID string, params *domain.MessageQueryParams) (*domain.MessageListResponse, error) {
	if err := validateGrantID(grantID); err != nil {
		return nil, err
	}
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
	if params.MetadataPair != "" {
		q.Set("metadata_pair", params.MetadataPair)
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
	if err := validateGrantID(grantID); err != nil {
		return nil, err
	}
	if err := validateMessageID(messageID); err != nil {
		return nil, err
	}
	queryURL := fmt.Sprintf("%s/v3/grants/%s/messages/%s", c.baseURL, grantID, messageID)

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

	resp, err := c.doRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// convertMessages converts API message responses to domain models.
func convertMessages(msgs []messageResponse) []domain.Message {
	return util.Map(msgs, convertMessage)
}

// convertMessage converts an API message response to domain model.
func convertMessage(m messageResponse) domain.Message {
	convertParticipant := func(p struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}) domain.EmailParticipant {
		return domain.EmailParticipant{Name: p.Name, Email: p.Email}
	}
	convertAttachment := func(a struct {
		ID          string `json:"id"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size"`
		ContentID   string `json:"content_id"`
		IsInline    bool   `json:"is_inline"`
	}) domain.Attachment {
		return domain.Attachment{
			ID:          a.ID,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
			ContentID:   a.ContentID,
			IsInline:    a.IsInline,
		}
	}

	from := util.Map(m.From, convertParticipant)
	to := util.Map(m.To, convertParticipant)
	cc := util.Map(m.Cc, convertParticipant)
	bcc := util.Map(m.Bcc, convertParticipant)
	replyTo := util.Map(m.ReplyTo, convertParticipant)
	attachments := util.Map(m.Attachments, convertAttachment)

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
		Metadata:    m.Metadata,
		CreatedAt:   time.Unix(m.CreatedAt, 0),
		Object:      m.Object,
	}
}
