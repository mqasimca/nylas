//go:build !integration
// +build !integration

package slack

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestMockClient_ListFiles(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		resp, err := mock.ListFiles(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, resp.Files, 2)
		assert.Equal(t, "test.png", resp.Files[0].Name)
		assert.Equal(t, "doc.pdf", resp.Files[1].Name)
	})

	t.Run("calls custom function with params", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListFilesFunc = func(ctx context.Context, params *domain.SlackFileQueryParams) (*domain.SlackFileListResponse, error) {
			assert.Equal(t, "C12345", params.ChannelID)
			assert.Equal(t, 10, params.Limit)
			return &domain.SlackFileListResponse{
				Files: []domain.SlackAttachment{
					{ID: "F1", Name: "custom.txt"},
				},
				NextCursor: "next-cursor",
			}, nil
		}

		resp, err := mock.ListFiles(context.Background(), &domain.SlackFileQueryParams{
			ChannelID: "C12345",
			Limit:     10,
		})
		require.NoError(t, err)
		assert.Len(t, resp.Files, 1)
		assert.Equal(t, "custom.txt", resp.Files[0].Name)
		assert.Equal(t, "next-cursor", resp.NextCursor)
	})

	t.Run("returns custom error", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListFilesFunc = func(ctx context.Context, params *domain.SlackFileQueryParams) (*domain.SlackFileListResponse, error) {
			return nil, domain.ErrSlackAuthFailed
		}

		resp, err := mock.ListFiles(context.Background(), nil)
		assert.ErrorIs(t, err, domain.ErrSlackAuthFailed)
		assert.Nil(t, resp)
	})

	t.Run("handles types filter", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListFilesFunc = func(ctx context.Context, params *domain.SlackFileQueryParams) (*domain.SlackFileListResponse, error) {
			assert.Equal(t, []string{"images", "pdfs"}, params.Types)
			return &domain.SlackFileListResponse{
				Files: []domain.SlackAttachment{
					{ID: "F1", Name: "photo.jpg", MimeType: "image/jpeg"},
					{ID: "F2", Name: "report.pdf", MimeType: "application/pdf"},
				},
			}, nil
		}

		resp, err := mock.ListFiles(context.Background(), &domain.SlackFileQueryParams{
			Types: []string{"images", "pdfs"},
		})
		require.NoError(t, err)
		assert.Len(t, resp.Files, 2)
	})
}

func TestMockClient_GetFileInfo(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		file, err := mock.GetFileInfo(context.Background(), "F12345")
		require.NoError(t, err)
		assert.Equal(t, "F12345", file.ID)
		assert.Equal(t, "test.png", file.Name)
		assert.Equal(t, "Test Image", file.Title)
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetFileInfoFunc = func(ctx context.Context, fileID string) (*domain.SlackAttachment, error) {
			if fileID == "F99999" {
				return &domain.SlackAttachment{
					ID:       "F99999",
					Name:     "custom.pdf",
					Title:    "Custom File",
					MimeType: "application/pdf",
					Size:     4096,
				}, nil
			}
			return nil, domain.ErrSlackMessageNotFound
		}

		file, err := mock.GetFileInfo(context.Background(), "F99999")
		require.NoError(t, err)
		assert.Equal(t, "custom.pdf", file.Name)
		assert.Equal(t, int64(4096), file.Size)

		file, err = mock.GetFileInfo(context.Background(), "F00000")
		assert.ErrorIs(t, err, domain.ErrSlackMessageNotFound)
		assert.Nil(t, file)
	})
}

func TestMockClient_DownloadFile(t *testing.T) {
	t.Run("returns default content when func not set", func(t *testing.T) {
		mock := NewMockClient()
		reader, err := mock.DownloadFile(context.Background(), "https://files.slack.com/test.png")
		require.NoError(t, err)
		defer reader.Close()

		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "mock file content", string(content))
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.DownloadFileFunc = func(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
			if strings.Contains(downloadURL, "restricted") {
				return nil, domain.ErrSlackPermissionDenied
			}
			return io.NopCloser(strings.NewReader("custom file data: " + downloadURL)), nil
		}

		reader, err := mock.DownloadFile(context.Background(), "https://files.slack.com/public.pdf")
		require.NoError(t, err)
		defer reader.Close()

		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Contains(t, string(content), "public.pdf")

		_, err = mock.DownloadFile(context.Background(), "https://files.slack.com/restricted/secret.pdf")
		assert.ErrorIs(t, err, domain.ErrSlackPermissionDenied)
	})

	t.Run("handles empty content", func(t *testing.T) {
		mock := NewMockClient()
		mock.DownloadFileFunc = func(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}

		reader, err := mock.DownloadFile(context.Background(), "https://files.slack.com/empty.txt")
		require.NoError(t, err)
		defer reader.Close()

		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Empty(t, content)
	})
}

func TestMockClient_ListMyChannels(t *testing.T) {
	t.Run("returns default data when func not set", func(t *testing.T) {
		mock := NewMockClient()
		resp, err := mock.ListMyChannels(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, resp.Channels, 2)
		for _, ch := range resp.Channels {
			assert.True(t, ch.IsMember)
		}
	})

	t.Run("calls custom function", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListMyChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
			return &domain.SlackChannelListResponse{
				Channels: []domain.SlackChannel{
					{ID: "C1", Name: "my-channel", IsMember: true},
				},
				NextCursor: "cursor",
			}, nil
		}

		resp, err := mock.ListMyChannels(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, resp.Channels, 1)
		assert.Equal(t, "my-channel", resp.Channels[0].Name)
	})

	t.Run("returns custom error", func(t *testing.T) {
		mock := NewMockClient()
		mock.ListMyChannelsFunc = func(ctx context.Context, params *domain.SlackChannelQueryParams) (*domain.SlackChannelListResponse, error) {
			return nil, domain.ErrSlackRateLimited
		}

		resp, err := mock.ListMyChannels(context.Background(), nil)
		assert.ErrorIs(t, err, domain.ErrSlackRateLimited)
		assert.Nil(t, resp)
	})
}

func TestMockClient_FileOperationsCompleteInterface(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Test ListFiles with nil params
	files, err := mock.ListFiles(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, files)

	// Test ListFiles with params
	files, err = mock.ListFiles(ctx, &domain.SlackFileQueryParams{
		ChannelID: "C1",
		UserID:    "U1",
		Limit:     50,
	})
	require.NoError(t, err)
	assert.NotNil(t, files)

	// Test GetFileInfo
	file, err := mock.GetFileInfo(ctx, "F1")
	require.NoError(t, err)
	assert.NotNil(t, file)

	// Test DownloadFile
	reader, err := mock.DownloadFile(ctx, "https://example.com/file.pdf")
	require.NoError(t, err)
	assert.NotNil(t, reader)
	reader.Close()
}
