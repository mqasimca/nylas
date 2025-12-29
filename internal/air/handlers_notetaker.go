package air

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// NotetakerResponse represents a notetaker for the UI
type NotetakerResponse struct {
	ID            string `json:"id"`
	State         string `json:"state"`
	MeetingLink   string `json:"meetingLink"`
	MeetingTitle  string `json:"meetingTitle"`
	JoinTime      string `json:"joinTime,omitempty"`
	Provider      string `json:"provider,omitempty"`
	HasRecording  bool   `json:"hasRecording"`
	HasTranscript bool   `json:"hasTranscript"`
	CreatedAt     string `json:"createdAt,omitempty"`
}

// CreateNotetakerRequest for creating a notetaker
type CreateNotetakerRequest struct {
	MeetingLink string `json:"meetingLink"`
	JoinTime    int64  `json:"joinTime,omitempty"`
	BotName     string `json:"botName,omitempty"`
}

// MediaResponse for notetaker media
type MediaResponse struct {
	RecordingURL   string `json:"recordingUrl,omitempty"`
	TranscriptURL  string `json:"transcriptUrl,omitempty"`
	RecordingSize  int64  `json:"recordingSize,omitempty"`
	TranscriptSize int64  `json:"transcriptSize,omitempty"`
	ExpiresAt      int64  `json:"expiresAt,omitempty"`
}

// notetakerStore holds notetakers in memory for demo
type notetakerStore struct {
	notetakers map[string]*NotetakerResponse
}

var ntStore = &notetakerStore{
	notetakers: make(map[string]*NotetakerResponse),
}

// handleNotetakersRoute dispatches notetaker requests by method
func (s *Server) handleNotetakersRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListNotetakers(w, r)
	case http.MethodPost:
		s.handleCreateNotetaker(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleNotetakerByID dispatches requests for individual notetakers
func (s *Server) handleNotetakerByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetNotetaker(w, r)
	case http.MethodDelete:
		s.handleDeleteNotetaker(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListNotetakers returns all notetakers
func (s *Server) handleListNotetakers(w http.ResponseWriter, r *http.Request) {
	notetakers := make([]*NotetakerResponse, 0, len(ntStore.notetakers))
	for _, nt := range ntStore.notetakers {
		notetakers = append(notetakers, nt)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notetakers); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleCreateNotetaker creates a new notetaker
func (s *Server) handleCreateNotetaker(w http.ResponseWriter, r *http.Request) {
	var req CreateNotetakerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MeetingLink == "" {
		http.Error(w, "meetingLink is required", http.StatusBadRequest)
		return
	}

	// Detect provider from meeting link
	provider := detectMeetingProvider(req.MeetingLink)

	// Generate ID
	id := generateNotetakerID()

	nt := &NotetakerResponse{
		ID:           id,
		State:        "scheduled",
		MeetingLink:  req.MeetingLink,
		MeetingTitle: "Meeting Recording",
		Provider:     provider,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	if req.JoinTime > 0 {
		nt.JoinTime = time.Unix(req.JoinTime, 0).Format(time.RFC3339)
	}

	ntStore.notetakers[id] = nt

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nt); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleGetNotetaker returns a single notetaker
func (s *Server) handleGetNotetaker(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	nt, ok := ntStore.notetakers[id]
	if !ok {
		http.Error(w, "Notetaker not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nt); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleGetNotetakerMedia returns media for a notetaker
func (s *Server) handleGetNotetakerMedia(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	nt, ok := ntStore.notetakers[id]
	if !ok {
		http.Error(w, "Notetaker not found", http.StatusNotFound)
		return
	}

	// Only return media if complete
	if nt.State != "complete" {
		http.Error(w, "Recording not yet available", http.StatusNotFound)
		return
	}

	media := MediaResponse{
		RecordingURL:   "/api/notetakers/recording/" + id + ".mp4",
		TranscriptURL:  "/api/notetakers/transcript/" + id + ".txt",
		RecordingSize:  1024 * 1024 * 50, // 50MB example
		TranscriptSize: 1024 * 10,        // 10KB example
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(media); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleDeleteNotetaker cancels a notetaker
func (s *Server) handleDeleteNotetaker(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	nt, ok := ntStore.notetakers[id]
	if !ok {
		http.Error(w, "Notetaker not found", http.StatusNotFound)
		return
	}

	nt.State = "cancelled"
	w.WriteHeader(http.StatusNoContent)
}

// detectMeetingProvider detects the meeting provider from URL
func detectMeetingProvider(link string) string {
	link = strings.ToLower(link)
	switch {
	case strings.Contains(link, "zoom.us"):
		return "zoom"
	case strings.Contains(link, "meet.google.com"):
		return "google_meet"
	case strings.Contains(link, "teams.microsoft.com"):
		return "teams"
	default:
		return "unknown"
	}
}

// generateNotetakerID generates a unique ID
func generateNotetakerID() string {
	return "nt_" + time.Now().Format("20060102150405")
}
