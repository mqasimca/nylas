//go:build !integration
// +build !integration

package nylas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertNotetaker(t *testing.T) {
	now := time.Now().Unix()

	apiNotetaker := notetakerResponse{
		ID:           "notetaker-123",
		State:        "recording",
		MeetingLink:  "https://zoom.us/j/123456",
		JoinTime:     now - 600,
		MeetingTitle: "Team Standup",
		MediaData: &struct {
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
		}{
			Recording: &struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Size        int64  `json:"size"`
				ExpiresAt   int64  `json:"expires_at"`
			}{
				URL:         "https://example.com/recording.mp4",
				ContentType: "video/mp4",
				Size:        1024000,
				ExpiresAt:   now + 86400,
			},
			Transcript: &struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Size        int64  `json:"size"`
				ExpiresAt   int64  `json:"expires_at"`
			}{
				URL:         "https://example.com/transcript.txt",
				ContentType: "text/plain",
				Size:        5000,
				ExpiresAt:   now + 86400,
			},
		},
		BotConfig: &struct {
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		}{
			Name:      "Meeting Bot",
			AvatarURL: "https://example.com/avatar.png",
		},
		MeetingInfo: &struct {
			Provider    string `json:"provider"`
			MeetingCode string `json:"meeting_code"`
		}{
			Provider:    "zoom",
			MeetingCode: "123456",
		},
		CreatedAt: now - 3600,
		UpdatedAt: now,
		Object:    "notetaker",
	}

	notetaker := convertNotetaker(apiNotetaker)

	assert.Equal(t, "notetaker-123", notetaker.ID)
	assert.Equal(t, "recording", notetaker.State)
	assert.Equal(t, "https://zoom.us/j/123456", notetaker.MeetingLink)
	assert.Equal(t, "Team Standup", notetaker.MeetingTitle)
	assert.Equal(t, time.Unix(now-600, 0), notetaker.JoinTime)
	assert.Equal(t, time.Unix(now-3600, 0), notetaker.CreatedAt)
	assert.Equal(t, time.Unix(now, 0), notetaker.UpdatedAt)
	assert.Equal(t, "notetaker", notetaker.Object)

	// Test BotConfig
	assert.NotNil(t, notetaker.BotConfig)
	assert.Equal(t, "Meeting Bot", notetaker.BotConfig.Name)
	assert.Equal(t, "https://example.com/avatar.png", notetaker.BotConfig.AvatarURL)

	// Test MeetingInfo
	assert.NotNil(t, notetaker.MeetingInfo)
	assert.Equal(t, "zoom", notetaker.MeetingInfo.Provider)
	assert.Equal(t, "123456", notetaker.MeetingInfo.MeetingCode)

	// Test MediaData
	assert.NotNil(t, notetaker.MediaData)
	assert.NotNil(t, notetaker.MediaData.Recording)
	assert.Equal(t, "https://example.com/recording.mp4", notetaker.MediaData.Recording.URL)
	assert.Equal(t, "video/mp4", notetaker.MediaData.Recording.ContentType)
	assert.Equal(t, int64(1024000), notetaker.MediaData.Recording.Size)
	assert.Equal(t, int64(now+86400), notetaker.MediaData.Recording.ExpiresAt)

	assert.NotNil(t, notetaker.MediaData.Transcript)
	assert.Equal(t, "https://example.com/transcript.txt", notetaker.MediaData.Transcript.URL)
	assert.Equal(t, "text/plain", notetaker.MediaData.Transcript.ContentType)
	assert.Equal(t, int64(5000), notetaker.MediaData.Transcript.Size)
	assert.Equal(t, int64(now+86400), notetaker.MediaData.Transcript.ExpiresAt)
}

func TestConvertNotetakers(t *testing.T) {
	now := time.Now().Unix()

	apiNotetakers := []notetakerResponse{
		{
			ID:           "notetaker-1",
			State:        "recording",
			MeetingLink:  "https://zoom.us/j/111",
			MeetingTitle: "Meeting 1",
			JoinTime:     now,
			CreatedAt:    now,
			UpdatedAt:    now,
			Object:       "notetaker",
		},
		{
			ID:           "notetaker-2",
			State:        "completed",
			MeetingLink:  "https://meet.google.com/abc-def",
			MeetingTitle: "Meeting 2",
			JoinTime:     now - 1800,
			CreatedAt:    now - 3600,
			UpdatedAt:    now,
			Object:       "notetaker",
		},
	}

	// Test convertNotetakers uses util.Map
	notetakers := convertNotetakers(apiNotetakers)

	assert.Len(t, notetakers, 2)
	assert.Equal(t, "notetaker-1", notetakers[0].ID)
	assert.Equal(t, "recording", notetakers[0].State)
	assert.Equal(t, "Meeting 1", notetakers[0].MeetingTitle)

	assert.Equal(t, "notetaker-2", notetakers[1].ID)
	assert.Equal(t, "completed", notetakers[1].State)
	assert.Equal(t, "Meeting 2", notetakers[1].MeetingTitle)
}

func TestConvertNotetakers_Empty(t *testing.T) {
	// Test with empty slice
	notetakers := convertNotetakers([]notetakerResponse{})
	assert.NotNil(t, notetakers)
	assert.Len(t, notetakers, 0)
}

func TestConvertNotetaker_MinimalData(t *testing.T) {
	now := time.Now().Unix()

	// Notetaker with minimal required fields
	apiNotetaker := notetakerResponse{
		ID:          "notetaker-min",
		State:       "pending",
		MeetingLink: "https://zoom.us/j/minimal",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	notetaker := convertNotetaker(apiNotetaker)

	assert.Equal(t, "notetaker-min", notetaker.ID)
	assert.Equal(t, "pending", notetaker.State)
	assert.Equal(t, "https://zoom.us/j/minimal", notetaker.MeetingLink)
	assert.Equal(t, "", notetaker.MeetingTitle)
	assert.Nil(t, notetaker.BotConfig)
	assert.Nil(t, notetaker.MeetingInfo)
	assert.Nil(t, notetaker.MediaData)
}

func TestConvertNotetaker_ZeroJoinTime(t *testing.T) {
	now := time.Now().Unix()

	apiNotetaker := notetakerResponse{
		ID:          "notetaker-nojoin",
		State:       "pending",
		MeetingLink: "https://zoom.us/j/test",
		JoinTime:    0, // Not yet joined
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	notetaker := convertNotetaker(apiNotetaker)

	assert.Equal(t, "notetaker-nojoin", notetaker.ID)
	// When JoinTime is 0, it should be the zero time
	assert.True(t, notetaker.JoinTime.IsZero())
}

func TestHTTPClient_ListNotetakers(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-123/notetakers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check for default limit
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		response := map[string]any{
			"data": []map[string]any{
				{
					"id":            "notetaker-1",
					"state":         "recording",
					"meeting_link":  "https://zoom.us/j/123",
					"meeting_title": "Team Meeting",
					"join_time":     now - 600,
					"created_at":    now - 3600,
					"updated_at":    now,
					"object":        "notetaker",
				},
				{
					"id":            "notetaker-2",
					"state":         "completed",
					"meeting_link":  "https://meet.google.com/abc",
					"meeting_title": "Client Call",
					"join_time":     now - 3600,
					"created_at":    now - 7200,
					"updated_at":    now - 600,
					"object":        "notetaker",
				},
			},
			"request_id": "req-123",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	notetakers, err := client.ListNotetakers(ctx, "grant-123", nil)

	require.NoError(t, err)
	assert.Len(t, notetakers, 2)
	assert.Equal(t, "notetaker-1", notetakers[0].ID)
	assert.Equal(t, "recording", notetakers[0].State)
	assert.Equal(t, "Team Meeting", notetakers[0].MeetingTitle)
	assert.Equal(t, "notetaker-2", notetakers[1].ID)
	assert.Equal(t, "completed", notetakers[1].State)
}

func TestHTTPClient_ListNotetakers_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-456/notetakers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check query params
		assert.Equal(t, "recording", r.URL.Query().Get("state"))
		assert.Equal(t, "next-page", r.URL.Query().Get("page_token"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))

		response := map[string]any{
			"data":       []map[string]any{},
			"request_id": "req-456",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	params := &domain.NotetakerQueryParams{
		State:     "recording",
		PageToken: "next-page",
		Limit:     20,
	}
	notetakers, err := client.ListNotetakers(ctx, "grant-456", params)

	require.NoError(t, err)
	assert.Len(t, notetakers, 0)
}

func TestHTTPClient_GetNotetaker(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-789/notetakers/notetaker-abc", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]any{
			"data": map[string]any{
				"id":            "notetaker-abc",
				"state":         "recording",
				"meeting_link":  "https://zoom.us/j/production",
				"meeting_title": "Production Review",
				"join_time":     now - 300,
				"media_data": map[string]any{
					"recording": map[string]any{
						"url":          "https://api.example.com/recording.mp4",
						"content_type": "video/mp4",
						"size":         int64(2048000),
						"expires_at":   now + 86400,
					},
					"transcript": map[string]any{
						"url":          "https://api.example.com/transcript.txt",
						"content_type": "text/plain",
						"size":         int64(10000),
						"expires_at":   now + 86400,
					},
				},
				"bot_config": map[string]any{
					"name":       "Production Bot",
					"avatar_url": "https://example.com/bot.png",
				},
				"meeting_info": map[string]any{
					"provider":     "zoom",
					"meeting_code": "production123",
				},
				"created_at": now - 3600,
				"updated_at": now,
				"object":     "notetaker",
			},
			"request_id": "req-789",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	notetaker, err := client.GetNotetaker(ctx, "grant-789", "notetaker-abc")

	require.NoError(t, err)
	assert.Equal(t, "notetaker-abc", notetaker.ID)
	assert.Equal(t, "recording", notetaker.State)
	assert.Equal(t, "Production Review", notetaker.MeetingTitle)
	assert.NotNil(t, notetaker.MediaData)
	assert.NotNil(t, notetaker.MediaData.Recording)
	assert.Equal(t, "https://api.example.com/recording.mp4", notetaker.MediaData.Recording.URL)
	assert.NotNil(t, notetaker.BotConfig)
	assert.Equal(t, "Production Bot", notetaker.BotConfig.Name)
	assert.NotNil(t, notetaker.MeetingInfo)
	assert.Equal(t, "zoom", notetaker.MeetingInfo.Provider)
}

func TestHTTPClient_CreateNotetaker(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-create/notetakers", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "https://zoom.us/j/newmeeting", body["meeting_link"])

		response := map[string]any{
			"data": map[string]any{
				"id":            "notetaker-new",
				"state":         "pending",
				"meeting_link":  "https://zoom.us/j/newmeeting",
				"meeting_title": "",
				"created_at":    now,
				"updated_at":    now,
				"object":        "notetaker",
			},
			"request_id": "req-create",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateNotetakerRequest{
		MeetingLink: "https://zoom.us/j/newmeeting",
	}

	notetaker, err := client.CreateNotetaker(ctx, "grant-create", req)

	require.NoError(t, err)
	assert.Equal(t, "notetaker-new", notetaker.ID)
	assert.Equal(t, "pending", notetaker.State)
	assert.Equal(t, "https://zoom.us/j/newmeeting", notetaker.MeetingLink)
}

func TestHTTPClient_CreateNotetaker_WithBotConfig(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-bot/notetakers", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "https://meet.google.com/xyz", body["meeting_link"])

		botConfig, ok := body["bot_config"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Custom Bot", botConfig["name"])

		response := map[string]any{
			"data": map[string]any{
				"id":           "notetaker-bot",
				"state":        "pending",
				"meeting_link": "https://meet.google.com/xyz",
				"bot_config": map[string]any{
					"name":       "Custom Bot",
					"avatar_url": "https://example.com/custom.png",
				},
				"created_at": now,
				"updated_at": now,
				"object":     "notetaker",
			},
			"request_id": "req-bot",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	req := &domain.CreateNotetakerRequest{
		MeetingLink: "https://meet.google.com/xyz",
		BotConfig: &domain.BotConfig{
			Name:      "Custom Bot",
			AvatarURL: "https://example.com/custom.png",
		},
	}

	notetaker, err := client.CreateNotetaker(ctx, "grant-bot", req)

	require.NoError(t, err)
	assert.Equal(t, "notetaker-bot", notetaker.ID)
	assert.NotNil(t, notetaker.BotConfig)
	assert.Equal(t, "Custom Bot", notetaker.BotConfig.Name)
}

func TestHTTPClient_DeleteNotetaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-delete/notetakers/notetaker-del", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	err := client.DeleteNotetaker(ctx, "grant-delete", "notetaker-del")

	require.NoError(t, err)
}

func TestHTTPClient_GetNotetakerMedia(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-media/notetakers/notetaker-media/media", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]any{
			"data": map[string]any{
				"recording": map[string]any{
					"url":          "https://storage.example.com/rec.mp4",
					"content_type": "video/mp4",
					"size":         int64(3072000),
					"expires_at":   now + 172800,
				},
				"transcript": map[string]any{
					"url":          "https://storage.example.com/transcript.txt",
					"content_type": "text/plain",
					"size":         int64(15000),
					"expires_at":   now + 172800,
				},
			},
			"request_id": "req-media",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	media, err := client.GetNotetakerMedia(ctx, "grant-media", "notetaker-media")

	require.NoError(t, err)
	assert.NotNil(t, media)
	assert.NotNil(t, media.Recording)
	assert.Equal(t, "https://storage.example.com/rec.mp4", media.Recording.URL)
	assert.Equal(t, int64(3072000), media.Recording.Size)
	assert.NotNil(t, media.Transcript)
	assert.Equal(t, "https://storage.example.com/transcript.txt", media.Transcript.URL)
	assert.Equal(t, int64(15000), media.Transcript.Size)
}

func TestHTTPClient_GetNotetakerMedia_OnlyRecording(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-rec/notetakers/notetaker-rec/media", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]any{
			"data": map[string]any{
				"recording": map[string]any{
					"url":          "https://storage.example.com/only-rec.mp4",
					"content_type": "video/mp4",
					"size":         int64(2560000),
					"expires_at":   now + 86400,
				},
			},
			"request_id": "req-rec",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	media, err := client.GetNotetakerMedia(ctx, "grant-rec", "notetaker-rec")

	require.NoError(t, err)
	assert.NotNil(t, media)
	assert.NotNil(t, media.Recording)
	assert.Equal(t, "https://storage.example.com/only-rec.mp4", media.Recording.URL)
	assert.Nil(t, media.Transcript)
}

func TestHTTPClient_GetNotetakerMedia_OnlyTranscript(t *testing.T) {
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/grants/grant-txt/notetakers/notetaker-txt/media", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]any{
			"data": map[string]any{
				"transcript": map[string]any{
					"url":          "https://storage.example.com/only-transcript.txt",
					"content_type": "text/plain",
					"size":         int64(12000),
					"expires_at":   now + 86400,
				},
			},
			"request_id": "req-txt",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	media, err := client.GetNotetakerMedia(ctx, "grant-txt", "notetaker-txt")

	require.NoError(t, err)
	assert.NotNil(t, media)
	assert.Nil(t, media.Recording)
	assert.NotNil(t, media.Transcript)
	assert.Equal(t, "https://storage.example.com/only-transcript.txt", media.Transcript.URL)
}
