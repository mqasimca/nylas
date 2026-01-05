//go:build !integration

package air

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectMeetingProvider(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "zoom link",
			link:     "https://zoom.us/j/123456789",
			expected: "zoom",
		},
		{
			name:     "zoom link uppercase",
			link:     "https://ZOOM.US/j/123456789",
			expected: "zoom",
		},
		{
			name:     "google meet link",
			link:     "https://meet.google.com/abc-defg-hij",
			expected: "google_meet",
		},
		{
			name:     "google meet link uppercase",
			link:     "https://MEET.GOOGLE.COM/abc-defg-hij",
			expected: "google_meet",
		},
		{
			name:     "microsoft teams link",
			link:     "https://teams.microsoft.com/l/meetup-join/...",
			expected: "teams",
		},
		{
			name:     "microsoft teams link uppercase",
			link:     "https://TEAMS.MICROSOFT.COM/l/meetup-join/...",
			expected: "teams",
		},
		{
			name:     "unknown provider",
			link:     "https://webex.com/meeting/123",
			expected: "unknown",
		},
		{
			name:     "empty link",
			link:     "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMeetingProvider(tt.link)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDomainToNotetakerResponse(t *testing.T) {
	now := time.Now()
	joinTime := now.Add(-1 * time.Hour)
	createdAt := now.Add(-2 * time.Hour)

	tests := []struct {
		name     string
		input    *domain.Notetaker
		expected *NotetakerResponse
	}{
		{
			name: "basic notetaker with all fields",
			input: &domain.Notetaker{
				ID:           "nt-123",
				State:        "complete",
				MeetingLink:  "https://zoom.us/j/123456789",
				MeetingTitle: "Team Standup",
				JoinTime:     joinTime,
				CreatedAt:    createdAt,
				MediaData: &domain.MediaData{
					Recording:  &domain.MediaFile{URL: "https://example.com/recording.mp4"},
					Transcript: &domain.MediaFile{URL: "https://example.com/transcript.txt"},
				},
				MeetingInfo: &domain.MeetingInfo{
					Provider: "zoom_custom",
				},
			},
			expected: &NotetakerResponse{
				ID:            "nt-123",
				State:         "complete",
				MeetingLink:   "https://zoom.us/j/123456789",
				MeetingTitle:  "Team Standup",
				JoinTime:      joinTime.Format(time.RFC3339),
				CreatedAt:     createdAt.Format(time.RFC3339),
				HasRecording:  true,
				HasTranscript: true,
				Provider:      "zoom_custom",
			},
		},
		{
			name: "notetaker with no media data",
			input: &domain.Notetaker{
				ID:           "nt-456",
				State:        "scheduled",
				MeetingLink:  "https://meet.google.com/abc-defg-hij",
				MeetingTitle: "Planning Session",
			},
			expected: &NotetakerResponse{
				ID:            "nt-456",
				State:         "scheduled",
				MeetingLink:   "https://meet.google.com/abc-defg-hij",
				MeetingTitle:  "Planning Session",
				HasRecording:  false,
				HasTranscript: false,
				Provider:      "google_meet",
			},
		},
		{
			name: "notetaker with only recording",
			input: &domain.Notetaker{
				ID:           "nt-789",
				State:        "complete",
				MeetingLink:  "https://teams.microsoft.com/meeting",
				MeetingTitle: "Sprint Review",
				MediaData: &domain.MediaData{
					Recording: &domain.MediaFile{URL: "https://example.com/recording.mp4"},
				},
			},
			expected: &NotetakerResponse{
				ID:            "nt-789",
				State:         "complete",
				MeetingLink:   "https://teams.microsoft.com/meeting",
				MeetingTitle:  "Sprint Review",
				HasRecording:  true,
				HasTranscript: false,
				Provider:      "teams",
			},
		},
		{
			name: "notetaker with empty title gets default",
			input: &domain.Notetaker{
				ID:          "nt-abc",
				State:       "complete",
				MeetingLink: "https://zoom.us/j/123",
			},
			expected: &NotetakerResponse{
				ID:            "nt-abc",
				State:         "complete",
				MeetingLink:   "https://zoom.us/j/123",
				MeetingTitle:  "Meeting Recording",
				HasRecording:  false,
				HasTranscript: false,
				Provider:      "zoom",
			},
		},
		{
			name: "notetaker with meeting info provider",
			input: &domain.Notetaker{
				ID:           "nt-def",
				State:        "attending",
				MeetingLink:  "https://custom.com/meeting",
				MeetingTitle: "Custom Meeting",
				MeetingInfo: &domain.MeetingInfo{
					Provider: "custom_provider",
				},
			},
			expected: &NotetakerResponse{
				ID:            "nt-def",
				State:         "attending",
				MeetingLink:   "https://custom.com/meeting",
				MeetingTitle:  "Custom Meeting",
				HasRecording:  false,
				HasTranscript: false,
				Provider:      "custom_provider",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domainToNotetakerResponse(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.State, result.State)
			assert.Equal(t, tt.expected.MeetingLink, result.MeetingLink)
			assert.Equal(t, tt.expected.MeetingTitle, result.MeetingTitle)
			assert.Equal(t, tt.expected.HasRecording, result.HasRecording)
			assert.Equal(t, tt.expected.HasTranscript, result.HasTranscript)
			assert.Equal(t, tt.expected.Provider, result.Provider)

			if tt.expected.JoinTime != "" {
				assert.Equal(t, tt.expected.JoinTime, result.JoinTime)
			}
			if tt.expected.CreatedAt != "" {
				assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			}
		})
	}
}

func TestNotetakerResponse_Fields(t *testing.T) {
	resp := NotetakerResponse{
		ID:            "nt-123",
		State:         "complete",
		MeetingLink:   "https://zoom.us/j/123",
		MeetingTitle:  "Test Meeting",
		JoinTime:      "2024-01-15T10:00:00Z",
		Provider:      "zoom",
		HasRecording:  true,
		HasTranscript: true,
		CreatedAt:     "2024-01-15T09:00:00Z",
		IsExternal:    true,
		ExternalURL:   "https://example.com/recording",
		Attendees:     "user1@example.com, user2@example.com",
		Summary:       "Meeting summary text",
	}

	assert.Equal(t, "nt-123", resp.ID)
	assert.Equal(t, "complete", resp.State)
	assert.Equal(t, "https://zoom.us/j/123", resp.MeetingLink)
	assert.Equal(t, "Test Meeting", resp.MeetingTitle)
	assert.Equal(t, "2024-01-15T10:00:00Z", resp.JoinTime)
	assert.Equal(t, "zoom", resp.Provider)
	assert.True(t, resp.HasRecording)
	assert.True(t, resp.HasTranscript)
	assert.Equal(t, "2024-01-15T09:00:00Z", resp.CreatedAt)
	assert.True(t, resp.IsExternal)
	assert.Equal(t, "https://example.com/recording", resp.ExternalURL)
	assert.Equal(t, "user1@example.com, user2@example.com", resp.Attendees)
	assert.Equal(t, "Meeting summary text", resp.Summary)
}

func TestNotetakerSource_Fields(t *testing.T) {
	source := NotetakerSource{
		From:       "noreply@otter.ai",
		Subject:    "Meeting Notes",
		LinkDomain: "otter.ai",
	}

	assert.Equal(t, "noreply@otter.ai", source.From)
	assert.Equal(t, "Meeting Notes", source.Subject)
	assert.Equal(t, "otter.ai", source.LinkDomain)
}

func TestCreateNotetakerRequest_Fields(t *testing.T) {
	req := CreateNotetakerRequest{
		MeetingLink: "https://zoom.us/j/123456789",
		JoinTime:    1705320000,
		BotName:     "My Meeting Bot",
	}

	assert.Equal(t, "https://zoom.us/j/123456789", req.MeetingLink)
	assert.Equal(t, int64(1705320000), req.JoinTime)
	assert.Equal(t, "My Meeting Bot", req.BotName)
}

func TestMediaResponse_Fields(t *testing.T) {
	media := MediaResponse{
		RecordingURL:   "https://example.com/recording.mp4",
		TranscriptURL:  "https://example.com/transcript.txt",
		RecordingSize:  1024000,
		TranscriptSize: 5000,
		ExpiresAt:      1705406400,
	}

	assert.Equal(t, "https://example.com/recording.mp4", media.RecordingURL)
	assert.Equal(t, "https://example.com/transcript.txt", media.TranscriptURL)
	assert.Equal(t, int64(1024000), media.RecordingSize)
	assert.Equal(t, int64(5000), media.TranscriptSize)
	assert.Equal(t, int64(1705406400), media.ExpiresAt)
}

func TestExcludedNotetakerStates(t *testing.T) {
	// Check that failed_entry is excluded
	assert.True(t, excludedNotetakerStates["failed_entry"])

	// Check that normal states are not excluded
	assert.False(t, excludedNotetakerStates["complete"])
	assert.False(t, excludedNotetakerStates["scheduled"])
	assert.False(t, excludedNotetakerStates["attending"])
}

func TestHandleNotetakersRoute_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	// Test PUT method (not allowed)
	req := httptest.NewRequest(http.MethodPut, "/api/notetakers", nil)
	w := httptest.NewRecorder()

	s.handleNotetakersRoute(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleNotetakerByID_MethodNotAllowed(t *testing.T) {
	s := &Server{}

	// Test POST method (not allowed)
	req := httptest.NewRequest(http.MethodPost, "/api/notetakers/123", nil)
	w := httptest.NewRecorder()

	s.handleNotetakerByID(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandleListNotetakers_NilClient(t *testing.T) {
	s := &Server{nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/notetakers", nil)
	w := httptest.NewRecorder()

	s.handleListNotetakers(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var result map[string]string
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "Not configured. Run 'nylas auth login' first.", result["error"])
}

func TestHandleCreateNotetaker_NilClient(t *testing.T) {
	s := &Server{nylasClient: nil}

	req := httptest.NewRequest(http.MethodPost, "/api/notetakers", strings.NewReader(`{"meetingLink":"https://zoom.us/j/123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handleCreateNotetaker(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandleGetNotetaker_NilClient(t *testing.T) {
	s := &Server{nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/notetakers/123?id=nt-123", nil)
	w := httptest.NewRecorder()

	s.handleGetNotetaker(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandleGetNotetakerMedia_NilClient(t *testing.T) {
	s := &Server{nylasClient: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/notetakers/123/media?id=nt-123", nil)
	w := httptest.NewRecorder()

	s.handleGetNotetakerMedia(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandleDeleteNotetaker_NilClient(t *testing.T) {
	s := &Server{nylasClient: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/notetakers/123?id=nt-123", nil)
	w := httptest.NewRecorder()

	s.handleDeleteNotetaker(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestNotetakerResponse_JSONSerialization(t *testing.T) {
	resp := &NotetakerResponse{
		ID:            "nt-123",
		State:         "complete",
		MeetingLink:   "https://zoom.us/j/123",
		MeetingTitle:  "Test",
		HasRecording:  true,
		HasTranscript: false,
		Provider:      "zoom",
		IsExternal:    false,
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal back
	var decoded NotetakerResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.ID, decoded.ID)
	assert.Equal(t, resp.State, decoded.State)
	assert.Equal(t, resp.MeetingLink, decoded.MeetingLink)
	assert.Equal(t, resp.MeetingTitle, decoded.MeetingTitle)
	assert.Equal(t, resp.HasRecording, decoded.HasRecording)
	assert.Equal(t, resp.HasTranscript, decoded.HasTranscript)
	assert.Equal(t, resp.Provider, decoded.Provider)
	assert.Equal(t, resp.IsExternal, decoded.IsExternal)
}

func TestNotetakerSource_JSONSerialization(t *testing.T) {
	source := &NotetakerSource{
		From:       "noreply@otter.ai",
		Subject:    "Meeting Notes",
		LinkDomain: "otter.ai",
	}

	// Marshal to JSON
	data, err := json.Marshal(source)
	require.NoError(t, err)

	// Unmarshal back
	var decoded NotetakerSource
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, source.From, decoded.From)
	assert.Equal(t, source.Subject, decoded.Subject)
	assert.Equal(t, source.LinkDomain, decoded.LinkDomain)
}

func TestCreateNotetakerRequest_JSONSerialization(t *testing.T) {
	req := &CreateNotetakerRequest{
		MeetingLink: "https://zoom.us/j/123456789",
		JoinTime:    1705320000,
		BotName:     "Test Bot",
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal back
	var decoded CreateNotetakerRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.MeetingLink, decoded.MeetingLink)
	assert.Equal(t, req.JoinTime, decoded.JoinTime)
	assert.Equal(t, req.BotName, decoded.BotName)
}

func TestMediaResponse_JSONSerialization(t *testing.T) {
	media := &MediaResponse{
		RecordingURL:   "https://example.com/recording.mp4",
		TranscriptURL:  "https://example.com/transcript.txt",
		RecordingSize:  1024000,
		TranscriptSize: 5000,
		ExpiresAt:      1705406400,
	}

	// Marshal to JSON
	data, err := json.Marshal(media)
	require.NoError(t, err)

	// Unmarshal back
	var decoded MediaResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, media.RecordingURL, decoded.RecordingURL)
	assert.Equal(t, media.TranscriptURL, decoded.TranscriptURL)
	assert.Equal(t, media.RecordingSize, decoded.RecordingSize)
	assert.Equal(t, media.TranscriptSize, decoded.TranscriptSize)
	assert.Equal(t, media.ExpiresAt, decoded.ExpiresAt)
}
