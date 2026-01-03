package air

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// ScreenedSender represents a sender pending approval
type ScreenedSender struct {
	Email       string    `json:"email"`
	Name        string    `json:"name,omitempty"`
	Domain      string    `json:"domain"`
	FirstSeen   time.Time `json:"firstSeen"`
	EmailCount  int       `json:"emailCount"`
	SampleSubj  string    `json:"sampleSubject,omitempty"`
	Status      string    `json:"status"`                // pending, allowed, blocked
	Destination string    `json:"destination,omitempty"` // inbox, feed, paper_trail
}

// ScreenerStore manages screened senders
type ScreenerStore struct {
	senders map[string]*ScreenedSender
	mu      sync.RWMutex
}

var screenerStore = &ScreenerStore{
	senders: make(map[string]*ScreenedSender),
}

// handleGetScreenedSenders returns pending senders
func (s *Server) handleGetScreenedSenders(w http.ResponseWriter, r *http.Request) {
	screenerStore.mu.RLock()
	defer screenerStore.mu.RUnlock()

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	senders := make([]*ScreenedSender, 0)
	for _, sender := range screenerStore.senders {
		if sender.Status == status {
			senders = append(senders, sender)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(senders); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleScreenerAllow allows a sender
func (s *Server) handleScreenerAllow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Destination string `json:"destination"` // inbox, feed, paper_trail
	}

	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Destination == "" {
		req.Destination = "inbox"
	}

	screenerStore.mu.Lock()
	defer screenerStore.mu.Unlock()

	if sender, ok := screenerStore.senders[req.Email]; ok {
		sender.Status = "allowed"
		sender.Destination = req.Destination
	} else {
		screenerStore.senders[req.Email] = &ScreenedSender{
			Email:       req.Email,
			Status:      "allowed",
			Destination: req.Destination,
			FirstSeen:   time.Now(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "allowed", "destination": req.Destination}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleScreenerBlock blocks a sender
func (s *Server) handleScreenerBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	screenerStore.mu.Lock()
	defer screenerStore.mu.Unlock()

	if sender, ok := screenerStore.senders[req.Email]; ok {
		sender.Status = "blocked"
	} else {
		screenerStore.senders[req.Email] = &ScreenedSender{
			Email:     req.Email,
			Status:    "blocked",
			FirstSeen: time.Now(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "blocked"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// handleAddToScreener adds a new sender for screening
func (s *Server) handleAddToScreener(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email   string `json:"email"`
		Name    string `json:"name,omitempty"`
		Subject string `json:"subject,omitempty"`
	}

	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	domain := extractDomain(req.Email)

	screenerStore.mu.Lock()
	defer screenerStore.mu.Unlock()

	if sender, ok := screenerStore.senders[req.Email]; ok {
		sender.EmailCount++
		if req.Subject != "" {
			sender.SampleSubj = req.Subject
		}
	} else {
		screenerStore.senders[req.Email] = &ScreenedSender{
			Email:      req.Email,
			Name:       req.Name,
			Domain:     domain,
			FirstSeen:  time.Now(),
			EmailCount: 1,
			SampleSubj: req.Subject,
			Status:     "pending",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "pending"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode", http.StatusInternalServerError)
	}
}

// IsSenderAllowed checks if a sender is allowed
func IsSenderAllowed(email string) (bool, string) {
	screenerStore.mu.RLock()
	defer screenerStore.mu.RUnlock()

	if sender, ok := screenerStore.senders[email]; ok {
		if sender.Status == "allowed" {
			return true, sender.Destination
		}
		return sender.Status != "blocked", ""
	}
	// Unknown sender - needs screening
	return false, ""
}
