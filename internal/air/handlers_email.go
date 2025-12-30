package air

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// handleListEmails returns emails with optional filtering.
func (s *Server) handleListEmails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock emails
	if s.demoMode {
		writeJSON(w, http.StatusOK, EmailsResponse{
			Emails:  demoEmails(),
			HasMore: false,
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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return
	}

	// Parse query parameters
	query := NewQueryParams(r.URL.Query())

	params := &domain.MessageQueryParams{
		Limit: query.GetLimit(50),
	}

	// Filter by folder
	folderID := query.Get("folder")
	if folderID != "" {
		params.In = []string{folderID}
	}

	// Filter by unread
	if query.GetBool("unread") {
		unreadBool := true
		params.Unread = &unreadBool
	}

	// Filter by starred
	if query.GetBool("starred") {
		starredBool := true
		params.Starred = &starredBool
	}

	// Search by sender email (from)
	fromFilter := query.Get("from")
	if fromFilter != "" {
		params.From = fromFilter
	}

	// Full-text search query
	searchQuery := query.Get("search")
	if searchQuery != "" {
		params.SearchQuery = searchQuery
	}

	// Cursor for pagination
	cursor := query.Get("cursor")
	if cursor != "" {
		params.PageToken = cursor
	}

	// Get account email for cache lookup
	accountEmail := s.getAccountEmail(grantID)

	// Try cache first (only for first page without complex filters)
	if cursor == "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEmailStore(accountEmail); err == nil {
			cacheOpts := cache.ListOptions{
				Limit:       params.Limit,
				FolderID:    folderID,
				UnreadOnly:  params.Unread != nil && *params.Unread,
				StarredOnly: params.Starred != nil && *params.Starred,
			}
			if cached, err := store.List(cacheOpts); err == nil && len(cached) > 0 {
				resp := EmailsResponse{
					Emails:  make([]EmailResponse, 0, len(cached)),
					HasMore: len(cached) >= params.Limit,
				}
				for _, e := range cached {
					resp.Emails = append(resp.Emails, cachedEmailToResponse(e))
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch messages from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetMessagesWithCursor(ctx, grantID, params)
	if err != nil {
		// If offline and cache available, try cache as fallback
		if s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
			if store, storeErr := s.getEmailStore(accountEmail); storeErr == nil {
				cacheOpts := cache.ListOptions{Limit: params.Limit, FolderID: folderID}
				if cached, cacheErr := store.List(cacheOpts); cacheErr == nil && len(cached) > 0 {
					resp := EmailsResponse{
						Emails:  make([]EmailResponse, 0, len(cached)),
						HasMore: false,
					}
					for _, e := range cached {
						resp.Emails = append(resp.Emails, cachedEmailToResponse(e))
					}
					writeJSON(w, http.StatusOK, resp)
					return
				}
			}
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch emails: " + err.Error(),
		})
		return
	}

	// Cache the results
	if s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEmailStore(accountEmail); err == nil {
			for i := range result.Data {
				_ = store.Put(domainMessageToCached(&result.Data[i]))
			}
		}
	}

	// Convert to response format
	resp := EmailsResponse{
		Emails:     make([]EmailResponse, 0, len(result.Data)),
		NextCursor: result.Pagination.NextCursor,
		HasMore:    result.Pagination.HasMore,
	}
	for _, m := range result.Data {
		resp.Emails = append(resp.Emails, emailToResponse(m, false))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleEmailByID handles single email operations: GET, PUT, DELETE.
func (s *Server) handleEmailByID(w http.ResponseWriter, r *http.Request) {
	// Parse email ID from path: /api/emails/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/emails/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Email ID required", http.StatusBadRequest)
		return
	}
	emailID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.handleGetEmail(w, r, emailID)
	case http.MethodPut:
		s.handleUpdateEmail(w, r, emailID)
	case http.MethodDelete:
		s.handleDeleteEmail(w, r, emailID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetEmail retrieves a single email with full body.
func (s *Server) handleGetEmail(w http.ResponseWriter, r *http.Request, emailID string) {
	// Demo mode: return mock email
	if s.demoMode {
		emails := demoEmails()
		for _, e := range emails {
			if e.ID == emailID {
				writeJSON(w, http.StatusOK, e)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Email not found"})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
		return
	}

	// Get account email for cache lookup
	accountEmail := s.getAccountEmail(grantID)

	// Try cache first
	if s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEmailStore(accountEmail); err == nil {
			if cached, err := store.Get(emailID); err == nil && cached != nil {
				resp := cachedEmailToResponse(cached)
				resp.Body = cached.BodyHTML // Include full body
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch message from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	msg, err := s.nylasClient.GetMessage(ctx, grantID, emailID)
	if err != nil {
		// Try cache as fallback on error
		if s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
			if store, storeErr := s.getEmailStore(accountEmail); storeErr == nil {
				if cached, cacheErr := store.Get(emailID); cacheErr == nil && cached != nil {
					resp := cachedEmailToResponse(cached)
					resp.Body = cached.BodyHTML
					writeJSON(w, http.StatusOK, resp)
					return
				}
			}
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch email: " + err.Error(),
		})
		return
	}

	// Cache the result
	if s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEmailStore(accountEmail); err == nil {
			_ = store.Put(domainMessageToCached(msg))
		}
	}

	writeJSON(w, http.StatusOK, emailToResponse(*msg, true))
}

// handleUpdateEmail updates an email (mark read/unread, star/unstar).
func (s *Server) handleUpdateEmail(w http.ResponseWriter, r *http.Request, emailID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, UpdateEmailResponse{
			Success: true,
			Message: "Email updated (demo mode)",
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

	// Parse request body
	var req UpdateEmailRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, UpdateEmailResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Update message via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	updateReq := &domain.UpdateMessageRequest{
		Unread:  req.Unread,
		Starred: req.Starred,
		Folders: req.Folders,
	}

	_, err = s.nylasClient.UpdateMessage(ctx, grantID, emailID, updateReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, UpdateEmailResponse{
			Success: false,
			Error:   "Failed to update email: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, UpdateEmailResponse{
		Success: true,
		Message: "Email updated",
	})
}

// handleDeleteEmail moves an email to trash.
func (s *Server) handleDeleteEmail(w http.ResponseWriter, r *http.Request, emailID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, UpdateEmailResponse{
			Success: true,
			Message: "Email deleted (demo mode)",
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

	// Delete message via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.nylasClient.DeleteMessage(ctx, grantID, emailID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, UpdateEmailResponse{
			Success: false,
			Error:   "Failed to delete email: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, UpdateEmailResponse{
		Success: true,
		Message: "Email deleted",
	})
}

// emailToResponse converts a domain message to an API response.
func emailToResponse(m domain.Message, includeBody bool) EmailResponse {
	resp := EmailResponse{
		ID:       m.ID,
		ThreadID: m.ThreadID,
		Subject:  m.Subject,
		Snippet:  m.Snippet,
		Date:     m.Date.Unix(),
		Unread:   m.Unread,
		Starred:  m.Starred,
		Folders:  m.Folders,
	}

	if includeBody {
		resp.Body = m.Body
	}

	// Convert participants
	for _, p := range m.From {
		resp.From = append(resp.From, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}
	for _, p := range m.To {
		resp.To = append(resp.To, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}
	for _, p := range m.Cc {
		resp.Cc = append(resp.Cc, EmailParticipantResponse{
			Name:  p.Name,
			Email: p.Email,
		})
	}

	// Convert attachments
	for _, a := range m.Attachments {
		resp.Attachments = append(resp.Attachments, AttachmentResponse{
			ID:          a.ID,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Size:        a.Size,
		})
	}

	return resp
}

// cachedEmailToResponse converts a cached email to response format.
func cachedEmailToResponse(e *cache.CachedEmail) EmailResponse {
	return EmailResponse{
		ID:       e.ID,
		ThreadID: e.ThreadID,
		Subject:  e.Subject,
		Snippet:  e.Snippet,
		From: []EmailParticipantResponse{
			{Name: e.FromName, Email: e.FromEmail},
		},
		Date:    e.Date.Unix(),
		Unread:  e.Unread,
		Starred: e.Starred,
		Folders: []string{e.FolderID},
	}
}

// demoEmails returns demo email data.
func demoEmails() []EmailResponse {
	now := time.Now()
	return []EmailResponse{
		{
			ID:      "demo-email-001",
			Subject: "Q4 Product Roadmap Review",
			Snippet: "Hi team, I've attached the updated roadmap for Q4...",
			Body:    "<p>Hi team,</p><p>I've attached the updated roadmap for Q4. Please review the timeline changes and let me know if you have any concerns.</p>",
			From:    []EmailParticipantResponse{{Name: "Sarah Chen", Email: "sarah.chen@company.com"}},
			To:      []EmailParticipantResponse{{Name: "Team", Email: "team@company.com"}},
			Date:    now.Add(-2 * time.Minute).Unix(),
			Unread:  true,
			Starred: true,
			Folders: []string{"inbox"},
			Attachments: []AttachmentResponse{
				{ID: "att-001", Filename: "Q4_Roadmap_v2.pdf", ContentType: "application/pdf", Size: 2516582},
			},
		},
		{
			ID:      "demo-email-002",
			Subject: "[nylas/cli] PR #142: Add focus time feature",
			Snippet: "mergify[bot] merged 1 commit into main...",
			From:    []EmailParticipantResponse{{Name: "GitHub", Email: "notifications@github.com"}},
			To:      []EmailParticipantResponse{{Name: "You", Email: "you@example.com"}},
			Date:    now.Add(-15 * time.Minute).Unix(),
			Unread:  true,
			Starred: false,
			Folders: []string{"inbox"},
		},
		{
			ID:      "demo-email-003",
			Subject: "Re: Meeting Tomorrow",
			Snippet: "That works for me. I'll send a calendar invite...",
			From:    []EmailParticipantResponse{{Name: "Alex Johnson", Email: "demo@example.com"}},
			To:      []EmailParticipantResponse{{Name: "You", Email: "you@example.com"}},
			Date:    now.Add(-1 * time.Hour).Unix(),
			Unread:  false,
			Starred: false,
			Folders: []string{"inbox"},
		},
		{
			ID:      "demo-email-004",
			Subject: "Your December invoice is ready",
			Snippet: "Your invoice for December 2024 is now available...",
			From:    []EmailParticipantResponse{{Name: "Stripe", Email: "billing@stripe.com"}},
			To:      []EmailParticipantResponse{{Name: "You", Email: "you@example.com"}},
			Date:    now.Add(-3 * time.Hour).Unix(),
			Unread:  false,
			Starred: true,
			Folders: []string{"inbox"},
		},
		{
			ID:      "demo-email-005",
			Subject: "This week in design: AI tools reshaping...",
			Snippet: "The latest trends, tools, and inspiration...",
			From:    []EmailParticipantResponse{{Name: "Design Weekly", Email: "newsletter@designweekly.com"}},
			To:      []EmailParticipantResponse{{Name: "You", Email: "you@example.com"}},
			Date:    now.Add(-5 * time.Hour).Unix(),
			Unread:  false,
			Starred: false,
			Folders: []string{"inbox"},
		},
	}
}
