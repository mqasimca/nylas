package nylas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

// folderResponse represents an API folder response.
type folderResponse struct {
	ID              string   `json:"id"`
	GrantID         string   `json:"grant_id"`
	Name            string   `json:"name"`
	SystemFolder    any      `json:"system_folder"` // Can be string or bool depending on provider
	ParentID        string   `json:"parent_id"`
	BackgroundColor string   `json:"background_color"`
	TextColor       string   `json:"text_color"`
	TotalCount      int      `json:"total_count"`
	UnreadCount     int      `json:"unread_count"`
	ChildIDs        []string `json:"child_ids"`
	Attributes      []string `json:"attributes"`
}

// GetFolders retrieves all folders for a grant.
func (c *HTTPClient) GetFolders(ctx context.Context, grantID string) ([]domain.Folder, error) {
	queryURL := fmt.Sprintf("%s/v3/grants/%s/folders", c.baseURL, grantID)

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

	resp, err := c.doRequest(ctx, req)
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

	resp, err := c.doRequest(ctx, httpReq)
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

	resp, err := c.doRequest(ctx, httpReq)
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

// convertFolders converts API folder responses to domain models.
func convertFolders(folders []folderResponse) []domain.Folder {
	result := make([]domain.Folder, len(folders))
	for i, f := range folders {
		result[i] = convertFolder(f)
	}
	return result
}

// convertFolder converts an API folder response to domain model.
func convertFolder(f folderResponse) domain.Folder {
	// SystemFolder can be a string or bool depending on provider
	var systemFolder string
	switch v := f.SystemFolder.(type) {
	case string:
		systemFolder = v
	case bool:
		if v {
			systemFolder = "true"
		}
		// If false, leave as empty string
	}

	return domain.Folder{
		ID:              f.ID,
		GrantID:         f.GrantID,
		Name:            f.Name,
		SystemFolder:    systemFolder,
		ParentID:        f.ParentID,
		BackgroundColor: f.BackgroundColor,
		TextColor:       f.TextColor,
		TotalCount:      f.TotalCount,
		UnreadCount:     f.UnreadCount,
		ChildIDs:        f.ChildIDs,
		Attributes:      f.Attributes,
	}
}
