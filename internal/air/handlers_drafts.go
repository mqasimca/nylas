package air

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// handleDrafts handles POST /api/drafts (create) and GET /api/drafts (list).
func (s *Server) handleDrafts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListDrafts(w, r)
	case http.MethodPost:
		s.handleCreateDraft(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListDrafts returns all drafts.
func (s *Server) handleListDrafts(w http.ResponseWriter, r *http.Request) {
	// Demo mode: return mock drafts
	if s.demoMode {
		writeJSON(w, http.StatusOK, DraftsResponse{
			Drafts: demoDrafts(),
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Fetch drafts from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	drafts, err := s.nylasClient.GetDrafts(ctx, grantID, 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch drafts: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := DraftsResponse{
		Drafts: make([]DraftResponse, 0, len(drafts)),
	}
	for _, d := range drafts {
		resp.Drafts = append(resp.Drafts, draftToResponse(d))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleCreateDraft creates a new draft.
func (s *Server) handleCreateDraft(w http.ResponseWriter, r *http.Request) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, DraftResponse{
			ID:      "demo-draft-new",
			Subject: "New Draft",
			Date:    time.Now().Unix(),
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Parse request body
	var req DraftRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Create draft via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	createReq := &domain.CreateDraftRequest{
		Subject:      req.Subject,
		Body:         req.Body,
		To:           participantsToEmail(req.To),
		Cc:           participantsToEmail(req.Cc),
		Bcc:          participantsToEmail(req.Bcc),
		ReplyToMsgID: req.ReplyToMsgID,
	}

	draft, err := s.nylasClient.CreateDraft(ctx, grantID, createReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to create draft: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, draftToResponse(*draft))
}

// handleDraftByID handles single draft operations: GET, PUT, DELETE, and POST .../send.
func (s *Server) handleDraftByID(w http.ResponseWriter, r *http.Request) {
	// Parse draft ID from path: /api/drafts/{id} or /api/drafts/{id}/send
	path := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Draft ID required", http.StatusBadRequest)
		return
	}
	draftID := parts[0]

	// Check for /send action
	if len(parts) > 1 && parts[1] == "send" && r.Method == http.MethodPost {
		s.handleSendDraft(w, r, draftID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetDraft(w, r, draftID)
	case http.MethodPut:
		s.handleUpdateDraft(w, r, draftID)
	case http.MethodDelete:
		s.handleDeleteDraft(w, r, draftID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDraft retrieves a single draft.
func (s *Server) handleGetDraft(w http.ResponseWriter, r *http.Request, draftID string) {
	// Demo mode: return mock draft
	if s.demoMode {
		drafts := demoDrafts()
		for _, d := range drafts {
			if d.ID == draftID {
				writeJSON(w, http.StatusOK, d)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Draft not found"})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Fetch draft from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	draft, err := s.nylasClient.GetDraft(ctx, grantID, draftID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch draft: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, draftToResponse(*draft))
}

// handleUpdateDraft updates an existing draft.
func (s *Server) handleUpdateDraft(w http.ResponseWriter, r *http.Request, draftID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, DraftResponse{
			ID:      draftID,
			Subject: "Updated Draft",
			Date:    time.Now().Unix(),
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Parse request body
	var req DraftRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Update draft via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	updateReq := &domain.CreateDraftRequest{
		Subject:      req.Subject,
		Body:         req.Body,
		To:           participantsToEmail(req.To),
		Cc:           participantsToEmail(req.Cc),
		Bcc:          participantsToEmail(req.Bcc),
		ReplyToMsgID: req.ReplyToMsgID,
	}

	draft, err := s.nylasClient.UpdateDraft(ctx, grantID, draftID, updateReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to update draft: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, draftToResponse(*draft))
}

// handleDeleteDraft deletes a draft.
func (s *Server) handleDeleteDraft(w http.ResponseWriter, r *http.Request, draftID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, UpdateEmailResponse{
			Success: true,
			Message: "Draft deleted (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, UpdateEmailResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Delete draft via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.nylasClient.DeleteDraft(ctx, grantID, draftID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, UpdateEmailResponse{
			Success: false,
			Error:   "Failed to delete draft: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, UpdateEmailResponse{
		Success: true,
		Message: "Draft deleted",
	})
}

// handleSendDraft sends a draft.
func (s *Server) handleSendDraft(w http.ResponseWriter, r *http.Request, draftID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, SendMessageResponse{
			Success:   true,
			MessageID: "demo-sent-" + draftID,
			Message:   "Email sent (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Send draft via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	msg, err := s.nylasClient.SendDraft(ctx, grantID, draftID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Error:   "Failed to send draft: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, SendMessageResponse{
		Success:   true,
		MessageID: msg.ID,
		Message:   "Email sent successfully",
	})
}

// handleSendMessage sends a message directly without creating a draft first.
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, SendMessageResponse{
			Success:   true,
			MessageID: "demo-sent-" + time.Now().Format("20060102150405"),
			Message:   "Email sent (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req SendMessageRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate recipients
	if len(req.To) == 0 {
		writeJSON(w, http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Error:   "At least one recipient is required",
		})
		return
	}

	// Send message via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	sendReq := &domain.SendMessageRequest{
		Subject:      req.Subject,
		Body:         req.Body,
		To:           participantsToEmail(req.To),
		Cc:           participantsToEmail(req.Cc),
		Bcc:          participantsToEmail(req.Bcc),
		ReplyToMsgID: req.ReplyToMsgID,
	}

	msg, err := s.nylasClient.SendMessage(ctx, grantID, sendReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Error:   "Failed to send message: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, SendMessageResponse{
		Success:   true,
		MessageID: msg.ID,
		Message:   "Email sent successfully",
	})
}

// draftToResponse converts a domain draft to an API response.
func draftToResponse(d domain.Draft) DraftResponse {
	resp := DraftResponse{
		ID:      d.ID,
		Subject: d.Subject,
		Body:    d.Body,
		Date:    d.CreatedAt.Unix(),
	}

	for _, p := range d.To {
		resp.To = append(resp.To, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}
	for _, p := range d.Cc {
		resp.Cc = append(resp.Cc, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}
	for _, p := range d.Bcc {
		resp.Bcc = append(resp.Bcc, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}

	return resp
}

// participantsToEmail converts API participants to domain email participants.
func participantsToEmail(participants []EmailParticipantResponse) []domain.EmailParticipant {
	result := make([]domain.EmailParticipant, 0, len(participants))
	for _, p := range participants {
		result = append(result, domain.EmailParticipant{
			Name:  p.Name,
			Email: p.Email,
		})
	}
	return result
}

// demoDrafts returns demo draft data.
func demoDrafts() []DraftResponse {
	now := time.Now()
	return []DraftResponse{
		{
			ID:      "demo-draft-001",
			Subject: "Re: Project Update",
			Body:    "<p>Thanks for the update. I'll review and get back to you.</p>",
			To:      []EmailParticipantResponse{{Name: "Sarah Chen", Email: "sarah@example.com"}},
			Date:    now.Add(-1 * time.Hour).Unix(),
		},
		{
			ID:      "demo-draft-002",
			Subject: "Meeting Follow-up",
			Body:    "<p>Hi team,</p><p>Following up on our discussion...</p>",
			To:      []EmailParticipantResponse{{Name: "Team", Email: "team@example.com"}},
			Date:    now.Add(-2 * time.Hour).Unix(),
		},
	}
}

// demoFolders returns demo folder data.
func demoFolders() []FolderResponse {
	return []FolderResponse{
		{ID: "inbox", Name: "Inbox", SystemFolder: "inbox", TotalCount: 156, UnreadCount: 23},
		{ID: "sent", Name: "Sent", SystemFolder: "sent", TotalCount: 89, UnreadCount: 0},
		{ID: "drafts", Name: "Drafts", SystemFolder: "drafts", TotalCount: 3, UnreadCount: 0},
		{ID: "trash", Name: "Trash", SystemFolder: "trash", TotalCount: 12, UnreadCount: 0},
		{ID: "spam", Name: "Spam", SystemFolder: "spam", TotalCount: 5, UnreadCount: 0},
		{ID: "archive", Name: "Archive", SystemFolder: "archive", TotalCount: 234, UnreadCount: 0},
		{ID: "starred", Name: "Starred", SystemFolder: "", TotalCount: 8, UnreadCount: 0},
	}
}
