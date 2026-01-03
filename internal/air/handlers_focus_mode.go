package air

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// FocusModeState represents the current focus mode state
type FocusModeState struct {
	IsActive      bool      `json:"isActive"`
	StartedAt     time.Time `json:"startedAt,omitempty"`
	EndsAt        time.Time `json:"endsAt,omitempty"`
	Duration      int       `json:"duration"` // minutes
	PomodoroMode  bool      `json:"pomodoroMode"`
	SessionCount  int       `json:"sessionCount"`
	BreakDuration int       `json:"breakDuration"` // minutes
	InBreak       bool      `json:"inBreak"`
}

// FocusModeSettings represents focus mode preferences
type FocusModeSettings struct {
	DefaultDuration    int      `json:"defaultDuration"`    // minutes
	PomodoroWork       int      `json:"pomodoroWork"`       // minutes
	PomodoroBreak      int      `json:"pomodoroBreak"`      // minutes
	PomodoroLongBreak  int      `json:"pomodoroLongBreak"`  // minutes
	SessionsBeforeLong int      `json:"sessionsBeforeLong"` // sessions before long break
	HideNotifications  bool     `json:"hideNotifications"`
	HideEmailList      bool     `json:"hideEmailList"`
	MuteSound          bool     `json:"muteSound"`
	AllowedSenders     []string `json:"allowedSenders"` // VIP list that can interrupt
	AutoReplyEnabled   bool     `json:"autoReplyEnabled"`
	AutoReplyMessage   string   `json:"autoReplyMessage"`
}

// focusModeStore holds focus mode state
type focusModeStore struct {
	state    *FocusModeState
	settings *FocusModeSettings
	mu       sync.RWMutex
}

var fmStore = &focusModeStore{
	state: &FocusModeState{
		IsActive:     false,
		Duration:     25,
		PomodoroMode: false,
	},
	settings: &FocusModeSettings{
		DefaultDuration:    25,
		PomodoroWork:       25,
		PomodoroBreak:      5,
		PomodoroLongBreak:  15,
		SessionsBeforeLong: 4,
		HideNotifications:  true,
		HideEmailList:      true,
		MuteSound:          false,
		AllowedSenders:     []string{},
		AutoReplyEnabled:   false,
		AutoReplyMessage:   "I'm currently in focus mode and will respond later.",
	},
}

// handleFocusModeRoute dispatches focus mode requests by method
func (s *Server) handleFocusModeRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetFocusModeState(w, r)
	case http.MethodPost:
		s.handleStartFocusMode(w, r)
	case http.MethodDelete:
		s.handleStopFocusMode(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFocusModeSettings dispatches focus mode settings requests by method
func (s *Server) handleFocusModeSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetFocusModeSettings(w, r)
	case http.MethodPut, http.MethodPost:
		s.handleUpdateFocusModeSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetFocusModeState returns current focus mode state
func (s *Server) handleGetFocusModeState(w http.ResponseWriter, r *http.Request) {
	fmStore.mu.RLock()
	defer fmStore.mu.RUnlock()

	// Check if session has ended
	state := *fmStore.state
	if state.IsActive && !state.EndsAt.IsZero() && time.Now().After(state.EndsAt) {
		state.IsActive = false
	}

	// Calculate remaining time
	response := map[string]any{
		"state": state,
	}

	if state.IsActive && !state.EndsAt.IsZero() {
		remaining := time.Until(state.EndsAt)
		if remaining > 0 {
			response["remainingMinutes"] = int(remaining.Minutes())
			response["remainingSeconds"] = int(remaining.Seconds()) % 60
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleStartFocusMode starts a focus mode session
func (s *Server) handleStartFocusMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Duration     int  `json:"duration,omitempty"`
		PomodoroMode bool `json:"pomodoroMode"`
	}

	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmStore.mu.Lock()
	defer fmStore.mu.Unlock()

	duration := req.Duration
	if duration <= 0 {
		if req.PomodoroMode {
			duration = fmStore.settings.PomodoroWork
		} else {
			duration = fmStore.settings.DefaultDuration
		}
	}

	now := time.Now()
	fmStore.state = &FocusModeState{
		IsActive:      true,
		StartedAt:     now,
		EndsAt:        now.Add(time.Duration(duration) * time.Minute),
		Duration:      duration,
		PomodoroMode:  req.PomodoroMode,
		SessionCount:  fmStore.state.SessionCount,
		BreakDuration: fmStore.settings.PomodoroBreak,
		InBreak:       false,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fmStore.state); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleStopFocusMode stops the current focus mode session
func (s *Server) handleStopFocusMode(w http.ResponseWriter, r *http.Request) {
	fmStore.mu.Lock()
	defer fmStore.mu.Unlock()

	if fmStore.state.IsActive && !fmStore.state.InBreak {
		fmStore.state.SessionCount++
	}
	fmStore.state.IsActive = false
	fmStore.state.InBreak = false

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"status":       "stopped",
		"sessionCount": fmStore.state.SessionCount,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleStartBreak starts a break in pomodoro mode
func (s *Server) handleStartBreak(w http.ResponseWriter, r *http.Request) {
	fmStore.mu.Lock()
	defer fmStore.mu.Unlock()

	if !fmStore.state.PomodoroMode {
		http.Error(w, "Not in pomodoro mode", http.StatusBadRequest)
		return
	}

	// Determine break duration
	breakDuration := fmStore.settings.PomodoroBreak
	if fmStore.state.SessionCount > 0 && fmStore.state.SessionCount%fmStore.settings.SessionsBeforeLong == 0 {
		breakDuration = fmStore.settings.PomodoroLongBreak
	}

	now := time.Now()
	fmStore.state.InBreak = true
	fmStore.state.StartedAt = now
	fmStore.state.EndsAt = now.Add(time.Duration(breakDuration) * time.Minute)
	fmStore.state.BreakDuration = breakDuration

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fmStore.state); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleGetFocusModeSettings returns focus mode settings
func (s *Server) handleGetFocusModeSettings(w http.ResponseWriter, r *http.Request) {
	fmStore.mu.RLock()
	defer fmStore.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fmStore.settings); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleUpdateFocusModeSettings updates focus mode settings
func (s *Server) handleUpdateFocusModeSettings(w http.ResponseWriter, r *http.Request) {
	var settings FocusModeSettings
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmStore.mu.Lock()
	fmStore.settings = &settings
	fmStore.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "updated"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// IsFocusModeActive returns whether focus mode is active
func IsFocusModeActive() bool {
	fmStore.mu.RLock()
	defer fmStore.mu.RUnlock()
	return fmStore.state.IsActive
}

// ShouldAllowNotification checks if notification should be shown
func ShouldAllowNotification(senderEmail string) bool {
	fmStore.mu.RLock()
	defer fmStore.mu.RUnlock()

	if !fmStore.state.IsActive || !fmStore.settings.HideNotifications {
		return true
	}

	// Check if sender is in allowed list
	for _, allowed := range fmStore.settings.AllowedSenders {
		if allowed == senderEmail {
			return true
		}
	}

	return false
}
