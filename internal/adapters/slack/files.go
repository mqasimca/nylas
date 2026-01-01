// files.go provides file operations for Slack.

package slack

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/slack-go/slack"

	"github.com/mqasimca/nylas/internal/domain"
)

// ListFiles returns files uploaded to a channel or workspace.
func (c *Client) ListFiles(ctx context.Context, params *domain.SlackFileQueryParams) (*domain.SlackFileListResponse, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	apiParams := slack.ListFilesParameters{}

	if params != nil {
		apiParams.Channel = params.ChannelID
		apiParams.User = params.UserID
		if len(params.Types) > 0 {
			// Join types into comma-separated string
			apiParams.Types = strings.Join(params.Types, ",")
		}
		if params.Limit > 0 {
			apiParams.Limit = params.Limit
		}
		apiParams.Cursor = params.Cursor
	}

	if apiParams.Limit == 0 {
		apiParams.Limit = 20
	}

	files, nextParams, err := c.api.ListFilesContext(ctx, apiParams)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	result := make([]domain.SlackAttachment, len(files))
	for i, f := range files {
		result[i] = ConvertFile(f)
	}

	// Get next cursor from returned params
	nextCursor := ""
	if nextParams != nil && nextParams.Cursor != "" {
		nextCursor = nextParams.Cursor
	}

	return &domain.SlackFileListResponse{
		Files:      result,
		NextCursor: nextCursor,
	}, nil
}

// GetFileInfo returns metadata for a single file by its ID.
func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*domain.SlackAttachment, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if fileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", domain.ErrSlackMessageNotFound)
	}

	file, _, _, err := c.api.GetFileInfoContext(ctx, fileID, 0, 0)
	if err != nil {
		return nil, c.handleSlackError(err)
	}

	attachment := ConvertFile(*file)
	return &attachment, nil
}

// DownloadFile downloads file content from a private download URL.
// The caller must close the returned ReadCloser when done.
func (c *Client) DownloadFile(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("%w: download URL is required", domain.ErrSlackMessageNotFound)
	}

	// Use a buffer to download the file content
	var buf bytes.Buffer
	if err := c.api.GetFileContext(ctx, downloadURL, &buf); err != nil {
		return nil, c.handleSlackError(err)
	}

	return io.NopCloser(&buf), nil
}
