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

// notetakerResponse represents an API notetaker response.
type notetakerResponse struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	MeetingLink  string `json:"meeting_link"`
	JoinTime     int64  `json:"join_time"`
	MeetingTitle string `json:"meeting_title"`
	MediaData    *struct {
		Recording *struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			Size        int64  `json:"size"`
			ExpiresAt   int64  `json:"expires_at"`
		} `json:"recording"`
		Transcript *struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			Size        int64  `json:"size"`
			ExpiresAt   int64  `json:"expires_at"`
		} `json:"transcript"`
	} `json:"media_data"`
	BotConfig *struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	} `json:"bot_config"`
	MeetingInfo *struct {
		Provider    string `json:"provider"`
		MeetingCode string `json:"meeting_code"`
	} `json:"meeting_info"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Object    string `json:"object"`
}

// ListNotetakers retrieves all notetakers for a grant.
func (c *HTTPClient) ListNotetakers(ctx context.Context, grantID string, params *domain.NotetakerQueryParams) ([]domain.Notetaker, error) {
	if params == nil {
		params = &domain.NotetakerQueryParams{Limit: 10}
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	queryURL := fmt.Sprintf("%s/v3/grants/%s/notetakers", c.baseURL, grantID)
	q := url.Values{}
	q.Set("limit", strconv.Itoa(params.Limit))

	if params.PageToken != "" {
		q.Set("page_token", params.PageToken)
	}
	if params.State != "" {
		q.Set("state", params.State)
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
		Data []notetakerResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return convertNotetakers(result.Data), nil
}

// GetNotetaker retrieves a single notetaker by ID.
func (c *HTTPClient) GetNotetaker(ctx context.Context, grantID, notetakerID string) (*domain.Notetaker, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/notetakers/%s", c.baseURL, grantID, notetakerID)

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
		return nil, fmt.Errorf("%w: notetaker not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data notetakerResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	notetaker := convertNotetaker(result.Data)
	return &notetaker, nil
}

// CreateNotetaker creates a new notetaker to join a meeting.
func (c *HTTPClient) CreateNotetaker(ctx context.Context, grantID string, req *domain.CreateNotetakerRequest) (*domain.Notetaker, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/notetakers", c.baseURL, grantID)

	payload := map[string]interface{}{
		"meeting_link": req.MeetingLink,
	}

	if req.JoinTime > 0 {
		payload["join_time"] = req.JoinTime
	}
	if req.BotConfig != nil {
		botConfig := map[string]interface{}{}
		if req.BotConfig.Name != "" {
			botConfig["name"] = req.BotConfig.Name
		}
		if req.BotConfig.AvatarURL != "" {
			botConfig["avatar_url"] = req.BotConfig.AvatarURL
		}
		if len(botConfig) > 0 {
			payload["bot_config"] = botConfig
		}
	}

	body, _ := json.Marshal(payload)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data notetakerResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	notetaker := convertNotetaker(result.Data)
	return &notetaker, nil
}

// DeleteNotetaker deletes/cancels a notetaker.
func (c *HTTPClient) DeleteNotetaker(ctx context.Context, grantID, notetakerID string) error {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/notetakers/%s", c.baseURL, grantID, notetakerID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", queryURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// GetNotetakerMedia retrieves the media (recording/transcript) for a notetaker.
func (c *HTTPClient) GetNotetakerMedia(ctx context.Context, grantID, notetakerID string) (*domain.MediaData, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/notetakers/%s/media", c.baseURL, grantID, notetakerID)

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
		return nil, fmt.Errorf("%w: notetaker media not found", domain.ErrAPIError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data struct {
			Recording *struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Size        int64  `json:"size"`
				ExpiresAt   int64  `json:"expires_at"`
			} `json:"recording"`
			Transcript *struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Size        int64  `json:"size"`
				ExpiresAt   int64  `json:"expires_at"`
			} `json:"transcript"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	media := &domain.MediaData{}
	if result.Data.Recording != nil {
		media.Recording = &domain.MediaFile{
			URL:         result.Data.Recording.URL,
			ContentType: result.Data.Recording.ContentType,
			Size:        result.Data.Recording.Size,
			ExpiresAt:   result.Data.Recording.ExpiresAt,
		}
	}
	if result.Data.Transcript != nil {
		media.Transcript = &domain.MediaFile{
			URL:         result.Data.Transcript.URL,
			ContentType: result.Data.Transcript.ContentType,
			Size:        result.Data.Transcript.Size,
			ExpiresAt:   result.Data.Transcript.ExpiresAt,
		}
	}

	return media, nil
}

// convertNotetakers converts API notetaker responses to domain models.
func convertNotetakers(notetakers []notetakerResponse) []domain.Notetaker {
	return util.Map(notetakers, convertNotetaker)
}

// convertNotetaker converts an API notetaker response to domain model.
func convertNotetaker(n notetakerResponse) domain.Notetaker {
	notetaker := domain.Notetaker{
		ID:           n.ID,
		State:        n.State,
		MeetingLink:  n.MeetingLink,
		MeetingTitle: n.MeetingTitle,
		Object:       n.Object,
	}

	if n.JoinTime > 0 {
		notetaker.JoinTime = time.Unix(n.JoinTime, 0)
	}
	if n.CreatedAt > 0 {
		notetaker.CreatedAt = time.Unix(n.CreatedAt, 0)
	}
	if n.UpdatedAt > 0 {
		notetaker.UpdatedAt = time.Unix(n.UpdatedAt, 0)
	}

	if n.BotConfig != nil {
		notetaker.BotConfig = &domain.BotConfig{
			Name:      n.BotConfig.Name,
			AvatarURL: n.BotConfig.AvatarURL,
		}
	}

	if n.MeetingInfo != nil {
		notetaker.MeetingInfo = &domain.MeetingInfo{
			Provider:    n.MeetingInfo.Provider,
			MeetingCode: n.MeetingInfo.MeetingCode,
		}
	}

	if n.MediaData != nil {
		notetaker.MediaData = &domain.MediaData{}
		if n.MediaData.Recording != nil {
			notetaker.MediaData.Recording = &domain.MediaFile{
				URL:         n.MediaData.Recording.URL,
				ContentType: n.MediaData.Recording.ContentType,
				Size:        n.MediaData.Recording.Size,
				ExpiresAt:   n.MediaData.Recording.ExpiresAt,
			}
		}
		if n.MediaData.Transcript != nil {
			notetaker.MediaData.Transcript = &domain.MediaFile{
				URL:         n.MediaData.Transcript.URL,
				ContentType: n.MediaData.Transcript.ContentType,
				Size:        n.MediaData.Transcript.Size,
				ExpiresAt:   n.MediaData.Transcript.ExpiresAt,
			}
		}
	}

	return notetaker
}
