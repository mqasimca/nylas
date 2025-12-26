package air

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// Grant represents an authenticated account for API responses.
type Grant struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
}

// grantFromDomain converts a domain.GrantInfo to a Grant for API responses.
func grantFromDomain(g domain.GrantInfo) Grant {
	return Grant{
		ID:       g.ID,
		Email:    g.Email,
		Provider: string(g.Provider),
	}
}

// ConfigStatusResponse represents the config status API response.
type ConfigStatusResponse struct {
	Configured   bool   `json:"configured"`
	Region       string `json:"region"`
	ClientID     string `json:"client_id,omitempty"`
	HasAPIKey    bool   `json:"has_api_key"`
	GrantCount   int    `json:"grant_count"`
	DefaultGrant string `json:"default_grant,omitempty"`
}

// GrantsResponse represents the grants list API response.
type GrantsResponse struct {
	Grants       []Grant `json:"grants"`
	DefaultGrant string  `json:"default_grant"`
}

// SetDefaultGrantRequest represents the request to set default grant.
type SetDefaultGrantRequest struct {
	GrantID string `json:"grant_id"`
}

// SetDefaultGrantResponse represents the response for setting default grant.
type SetDefaultGrantResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// FolderResponse represents a folder in API responses.
type FolderResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SystemFolder string `json:"system_folder,omitempty"`
	TotalCount   int    `json:"total_count"`
	UnreadCount  int    `json:"unread_count"`
}

// FoldersResponse represents the folders list API response.
type FoldersResponse struct {
	Folders []FolderResponse `json:"folders"`
}

// EmailParticipantResponse represents an email participant.
type EmailParticipantResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AttachmentResponse represents an email attachment.
type AttachmentResponse struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// EmailResponse represents an email in API responses.
type EmailResponse struct {
	ID          string                     `json:"id"`
	ThreadID    string                     `json:"thread_id,omitempty"`
	Subject     string                     `json:"subject"`
	Snippet     string                     `json:"snippet"`
	Body        string                     `json:"body,omitempty"`
	From        []EmailParticipantResponse `json:"from"`
	To          []EmailParticipantResponse `json:"to,omitempty"`
	Cc          []EmailParticipantResponse `json:"cc,omitempty"`
	Date        int64                      `json:"date"` // Unix timestamp
	Unread      bool                       `json:"unread"`
	Starred     bool                       `json:"starred"`
	Folders     []string                   `json:"folders,omitempty"`
	Attachments []AttachmentResponse       `json:"attachments,omitempty"`
}

// EmailsResponse represents the emails list API response.
type EmailsResponse struct {
	Emails     []EmailResponse `json:"emails"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

// UpdateEmailRequest represents a request to update an email.
type UpdateEmailRequest struct {
	Unread  *bool    `json:"unread,omitempty"`
	Starred *bool    `json:"starred,omitempty"`
	Folders []string `json:"folders,omitempty"`
}

// UpdateEmailResponse represents the response for updating an email.
type UpdateEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DraftRequest represents a request to create or update a draft.
type DraftRequest struct {
	To           []EmailParticipantResponse `json:"to"`
	Cc           []EmailParticipantResponse `json:"cc,omitempty"`
	Bcc          []EmailParticipantResponse `json:"bcc,omitempty"`
	Subject      string                     `json:"subject"`
	Body         string                     `json:"body"`
	ReplyToMsgID string                     `json:"reply_to_message_id,omitempty"`
}

// DraftResponse represents a draft in API responses.
type DraftResponse struct {
	ID      string                     `json:"id"`
	Subject string                     `json:"subject"`
	Body    string                     `json:"body,omitempty"`
	To      []EmailParticipantResponse `json:"to,omitempty"`
	Cc      []EmailParticipantResponse `json:"cc,omitempty"`
	Bcc     []EmailParticipantResponse `json:"bcc,omitempty"`
	Date    int64                      `json:"date"`
}

// DraftsResponse represents the drafts list API response.
type DraftsResponse struct {
	Drafts []DraftResponse `json:"drafts"`
}

// SendMessageRequest represents a request to send a message directly.
type SendMessageRequest struct {
	To           []EmailParticipantResponse `json:"to"`
	Cc           []EmailParticipantResponse `json:"cc,omitempty"`
	Bcc          []EmailParticipantResponse `json:"bcc,omitempty"`
	Subject      string                     `json:"subject"`
	Body         string                     `json:"body"`
	ReplyToMsgID string                     `json:"reply_to_message_id,omitempty"`
}

// SendMessageResponse represents the response for sending a message.
type SendMessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CalendarResponse represents a calendar in API responses.
type CalendarResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	IsPrimary   bool   `json:"is_primary"`
	ReadOnly    bool   `json:"read_only"`
	HexColor    string `json:"hex_color,omitempty"`
}

// CalendarsResponse represents the calendars list API response.
type CalendarsResponse struct {
	Calendars []CalendarResponse `json:"calendars"`
}

// EventParticipantResponse represents an event participant.
type EventParticipantResponse struct {
	Name   string `json:"name,omitempty"`
	Email  string `json:"email"`
	Status string `json:"status,omitempty"`
}

// ConferencingResponse represents video conferencing details.
type ConferencingResponse struct {
	Provider string `json:"provider,omitempty"`
	URL      string `json:"url,omitempty"`
}

// EventResponse represents an event in API responses.
type EventResponse struct {
	ID           string                     `json:"id"`
	CalendarID   string                     `json:"calendar_id"`
	Title        string                     `json:"title"`
	Description  string                     `json:"description,omitempty"`
	Location     string                     `json:"location,omitempty"`
	StartTime    int64                      `json:"start_time"`
	EndTime      int64                      `json:"end_time"`
	Timezone     string                     `json:"timezone,omitempty"`
	IsAllDay     bool                       `json:"is_all_day"`
	Status       string                     `json:"status,omitempty"`
	Busy         bool                       `json:"busy"`
	Participants []EventParticipantResponse `json:"participants,omitempty"`
	Conferencing *ConferencingResponse      `json:"conferencing,omitempty"`
	HtmlLink     string                     `json:"html_link,omitempty"`
}

// EventsResponse represents the events list API response.
type EventsResponse struct {
	Events     []EventResponse `json:"events"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

// CreateEventRequest represents a request to create an event.
type CreateEventRequest struct {
	CalendarID   string                     `json:"calendar_id"`
	Title        string                     `json:"title"`
	Description  string                     `json:"description,omitempty"`
	Location     string                     `json:"location,omitempty"`
	StartTime    int64                      `json:"start_time"`
	EndTime      int64                      `json:"end_time"`
	Timezone     string                     `json:"timezone,omitempty"`
	IsAllDay     bool                       `json:"is_all_day"`
	Busy         bool                       `json:"busy"`
	Participants []EventParticipantResponse `json:"participants,omitempty"`
}

// UpdateEventRequest represents a request to update an event.
type UpdateEventRequest struct {
	Title        *string                    `json:"title,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Location     *string                    `json:"location,omitempty"`
	StartTime    *int64                     `json:"start_time,omitempty"`
	EndTime      *int64                     `json:"end_time,omitempty"`
	Timezone     *string                    `json:"timezone,omitempty"`
	IsAllDay     *bool                      `json:"is_all_day,omitempty"`
	Busy         *bool                      `json:"busy,omitempty"`
	Participants []EventParticipantResponse `json:"participants,omitempty"`
}

// EventActionResponse represents the response for event actions (create/update/delete).
type EventActionResponse struct {
	Success bool           `json:"success"`
	Event   *EventResponse `json:"event,omitempty"`
	Message string         `json:"message,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ====================================
// CONTACTS TYPES
// ====================================

// ContactEmailResponse represents a contact email in API responses.
type ContactEmailResponse struct {
	Email string `json:"email"`
	Type  string `json:"type,omitempty"`
}

// ContactPhoneResponse represents a contact phone number in API responses.
type ContactPhoneResponse struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

// ContactAddressResponse represents a contact address in API responses.
type ContactAddressResponse struct {
	Type          string `json:"type,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

// ContactResponse represents a contact in API responses.
type ContactResponse struct {
	ID           string                   `json:"id"`
	GivenName    string                   `json:"given_name,omitempty"`
	Surname      string                   `json:"surname,omitempty"`
	DisplayName  string                   `json:"display_name"`
	Nickname     string                   `json:"nickname,omitempty"`
	CompanyName  string                   `json:"company_name,omitempty"`
	JobTitle     string                   `json:"job_title,omitempty"`
	Birthday     string                   `json:"birthday,omitempty"`
	Notes        string                   `json:"notes,omitempty"`
	PictureURL   string                   `json:"picture_url,omitempty"`
	Emails       []ContactEmailResponse   `json:"emails,omitempty"`
	PhoneNumbers []ContactPhoneResponse   `json:"phone_numbers,omitempty"`
	Addresses    []ContactAddressResponse `json:"addresses,omitempty"`
	Source       string                   `json:"source,omitempty"`
}

// ContactsResponse represents the contacts list API response.
type ContactsResponse struct {
	Contacts   []ContactResponse `json:"contacts"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

// CreateContactRequest represents a request to create a contact.
type CreateContactRequest struct {
	GivenName    string                   `json:"given_name,omitempty"`
	Surname      string                   `json:"surname,omitempty"`
	Nickname     string                   `json:"nickname,omitempty"`
	CompanyName  string                   `json:"company_name,omitempty"`
	JobTitle     string                   `json:"job_title,omitempty"`
	Birthday     string                   `json:"birthday,omitempty"`
	Notes        string                   `json:"notes,omitempty"`
	Emails       []ContactEmailResponse   `json:"emails,omitempty"`
	PhoneNumbers []ContactPhoneResponse   `json:"phone_numbers,omitempty"`
	Addresses    []ContactAddressResponse `json:"addresses,omitempty"`
}

// UpdateContactRequest represents a request to update a contact.
type UpdateContactRequest struct {
	GivenName    *string                  `json:"given_name,omitempty"`
	Surname      *string                  `json:"surname,omitempty"`
	Nickname     *string                  `json:"nickname,omitempty"`
	CompanyName  *string                  `json:"company_name,omitempty"`
	JobTitle     *string                  `json:"job_title,omitempty"`
	Birthday     *string                  `json:"birthday,omitempty"`
	Notes        *string                  `json:"notes,omitempty"`
	Emails       []ContactEmailResponse   `json:"emails,omitempty"`
	PhoneNumbers []ContactPhoneResponse   `json:"phone_numbers,omitempty"`
	Addresses    []ContactAddressResponse `json:"addresses,omitempty"`
}

// ContactActionResponse represents the response for contact actions (create/update/delete).
type ContactActionResponse struct {
	Success bool             `json:"success"`
	Contact *ContactResponse `json:"contact,omitempty"`
	Message string           `json:"message,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// ContactGroupResponse represents a contact group in API responses.
type ContactGroupResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

// ContactGroupsResponse represents the contact groups list API response.
type ContactGroupsResponse struct {
	Groups []ContactGroupResponse `json:"groups"`
}

// maxRequestBodySize is the maximum allowed request body size (1MB).
const maxRequestBodySize = 1 << 20

// limitedBody wraps a request body with a size limit.
func limitedBody(w http.ResponseWriter, r *http.Request) io.ReadCloser {
	return http.MaxBytesReader(w, r.Body, maxRequestBodySize)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// handleConfigStatus returns the current configuration status.
func (s *Server) handleConfigStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return sample configured status
	if s.demoMode {
		writeJSON(w, http.StatusOK, ConfigStatusResponse{
			Configured:   true,
			Region:       "us",
			ClientID:     "demo-client-id",
			HasAPIKey:    true,
			GrantCount:   3,
			DefaultGrant: demoDefaultGrant(),
		})
		return
	}

	status, err := s.configSvc.GetStatus()
	if err != nil {
		writeJSON(w, http.StatusOK, ConfigStatusResponse{Configured: false})
		return
	}

	resp := ConfigStatusResponse{
		Configured:   status.IsConfigured,
		Region:       status.Region,
		ClientID:     status.ClientID,
		HasAPIKey:    status.HasAPIKey,
		GrantCount:   status.GrantCount,
		DefaultGrant: status.DefaultGrant,
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleListGrants returns all authenticated accounts.
func (s *Server) handleListGrants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return sample grants
	if s.demoMode {
		writeJSON(w, http.StatusOK, GrantsResponse{
			Grants:       demoGrants(),
			DefaultGrant: demoDefaultGrant(),
		})
		return
	}

	grants, err := s.grantStore.ListGrants()
	if err != nil {
		writeJSON(w, http.StatusOK, GrantsResponse{Grants: []Grant{}})
		return
	}

	// Filter to only supported providers (Google, Microsoft)
	var grantList []Grant
	for _, g := range grants {
		if g.Provider.IsSupportedByAir() {
			grantList = append(grantList, grantFromDomain(g))
		}
	}

	defaultID, _ := s.grantStore.GetDefaultGrant()

	// If default grant is not a supported provider, pick the first supported account as default
	defaultIsSupported := false
	for _, g := range grantList {
		if g.ID == defaultID {
			defaultIsSupported = true
			break
		}
	}
	if !defaultIsSupported && len(grantList) > 0 {
		defaultID = grantList[0].ID
	}

	writeJSON(w, http.StatusOK, GrantsResponse{
		Grants:       grantList,
		DefaultGrant: defaultID,
	})
}

// handleSetDefaultGrant sets the default account.
func (s *Server) handleSetDefaultGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, SetDefaultGrantResponse{
			Success: true,
			Message: "Default account updated (demo mode)",
		})
		return
	}

	var req SetDefaultGrantRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SetDefaultGrantResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.GrantID == "" {
		writeJSON(w, http.StatusBadRequest, SetDefaultGrantResponse{
			Success: false,
			Error:   "Grant ID is required",
		})
		return
	}

	// Verify grant exists
	grants, err := s.grantStore.ListGrants()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SetDefaultGrantResponse{
			Success: false,
			Error:   "Failed to list grants",
		})
		return
	}

	found := slices.ContainsFunc(grants, func(g domain.GrantInfo) bool {
		return g.ID == req.GrantID
	})

	if !found {
		writeJSON(w, http.StatusNotFound, SetDefaultGrantResponse{
			Success: false,
			Error:   "Grant not found",
		})
		return
	}

	if err := s.grantStore.SetDefaultGrant(req.GrantID); err != nil {
		writeJSON(w, http.StatusInternalServerError, SetDefaultGrantResponse{
			Success: false,
			Error:   "Failed to set default grant: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, SetDefaultGrantResponse{
		Success: true,
		Message: "Default account updated",
	})
}

// Demo mode helpers.
func demoGrants() []Grant {
	return []Grant{
		{ID: "demo-grant-001", Email: "alice@example.com", Provider: "google"},
		{ID: "demo-grant-002", Email: "bob@work.com", Provider: "microsoft"},
		{ID: "demo-grant-003", Email: "carol@company.org", Provider: "google"},
	}
}

func demoDefaultGrant() string {
	return "demo-grant-001"
}

// handleListFolders returns all folders for the current account.
func (s *Server) handleListFolders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock folders
	if s.demoMode {
		writeJSON(w, http.StatusOK, FoldersResponse{
			Folders: demoFolders(),
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

	// Fetch folders from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	folders, err := s.nylasClient.GetFolders(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch folders: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := FoldersResponse{
		Folders: make([]FolderResponse, 0, len(folders)),
	}
	for _, f := range folders {
		resp.Folders = append(resp.Folders, FolderResponse{
			ID:           f.ID,
			Name:         f.Name,
			SystemFolder: f.SystemFolder,
			TotalCount:   f.TotalCount,
			UnreadCount:  f.UnreadCount,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

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
	query := r.URL.Query()
	limit := 50
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	params := &domain.MessageQueryParams{
		Limit: limit,
	}

	// Filter by folder
	folderID := query.Get("folder")
	if folderID != "" {
		params.In = []string{folderID}
	}

	// Filter by unread
	unreadFilter := query.Get("unread") == "true"
	if unreadFilter {
		unreadBool := true
		params.Unread = &unreadBool
	}

	// Filter by starred
	starredFilter := query.Get("starred") == "true"
	if starredFilter {
		starredBool := true
		params.Starred = &starredBool
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
				Limit:       limit,
				FolderID:    folderID,
				UnreadOnly:  unreadFilter,
				StarredOnly: starredFilter,
			}
			if cached, err := store.List(cacheOpts); err == nil && len(cached) > 0 {
				resp := EmailsResponse{
					Emails:  make([]EmailResponse, 0, len(cached)),
					HasMore: len(cached) >= limit,
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
				cacheOpts := cache.ListOptions{Limit: limit, FolderID: folderID}
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

// Conversion helpers

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

// ====================================
// DRAFT & SEND HANDLERS
// ====================================

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

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
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

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
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

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
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

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No default account. Please select an account first.",
		})
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

// Conversion helpers for drafts

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

// Demo data helpers

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
			From:    []EmailParticipantResponse{{Name: "Alex Johnson", Email: "alex.johnson@nylas.com"}},
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

// ====================================
// CALENDAR HANDLERS
// ====================================

// handleListCalendars returns all calendars for the current account.
func (s *Server) handleListCalendars(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock calendars
	if s.demoMode {
		writeJSON(w, http.StatusOK, CalendarsResponse{
			Calendars: demoCalendars(),
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

	// Fetch calendars from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	calendars, err := s.nylasClient.GetCalendars(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch calendars: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := CalendarsResponse{
		Calendars: make([]CalendarResponse, 0, len(calendars)),
	}
	for _, c := range calendars {
		resp.Calendars = append(resp.Calendars, calendarToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleListEvents returns events for a calendar with optional date filtering.
func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	// Demo mode: return mock events
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventsResponse{
			Events:  demoEvents(),
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
	query := r.URL.Query()

	// Calendar ID is required
	calendarID := query.Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary" // Default to primary calendar
	}

	// Build query params
	params := &domain.EventQueryParams{
		Limit:           50,
		ExpandRecurring: true,
	}

	// Limit
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			params.Limit = parsed
		}
	}

	// Date range filtering
	if start := query.Get("start"); start != "" {
		if parsed, err := strconv.ParseInt(start, 10, 64); err == nil {
			params.Start = parsed
		}
	}
	if end := query.Get("end"); end != "" {
		if parsed, err := strconv.ParseInt(end, 10, 64); err == nil {
			params.End = parsed
		}
	}

	// Default to current week if no date range specified
	if params.Start == 0 && params.End == 0 {
		now := time.Now()
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		startOfWeek := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		// End of week (Saturday)
		endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Second)
		params.Start = startOfWeek.Unix()
		params.End = endOfWeek.Unix()
	}

	// Cursor for pagination
	cursor := query.Get("cursor")
	if cursor != "" {
		params.PageToken = cursor
	}

	// Get account email for cache lookup
	accountEmail := s.getAccountEmail(grantID)

	// Try cache first (only for first page)
	if cursor == "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getEventStore(accountEmail); err == nil {
			cacheOpts := cache.EventListOptions{
				CalendarID: calendarID,
				Start:      time.Unix(params.Start, 0),
				End:        time.Unix(params.End, 0),
				Limit:      params.Limit,
			}
			if cached, err := store.List(cacheOpts); err == nil && len(cached) > 0 {
				resp := EventsResponse{
					Events:  make([]EventResponse, 0, len(cached)),
					HasMore: len(cached) >= params.Limit,
				}
				for _, e := range cached {
					resp.Events = append(resp.Events, cachedEventToResponse(e))
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch events from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetEventsWithCursor(ctx, grantID, calendarID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch events: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := EventsResponse{
		Events:     make([]EventResponse, 0, len(result.Data)),
		NextCursor: result.Pagination.NextCursor,
		HasMore:    result.Pagination.HasMore,
	}
	for _, e := range result.Data {
		resp.Events = append(resp.Events, eventToResponse(e))
	}

	writeJSON(w, http.StatusOK, resp)
}

// Conversion helpers for calendar

func calendarToResponse(c domain.Calendar) CalendarResponse {
	return CalendarResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Timezone:    c.Timezone,
		IsPrimary:   c.IsPrimary,
		ReadOnly:    c.ReadOnly,
		HexColor:    c.HexColor,
	}
}

func eventToResponse(e domain.Event) EventResponse {
	resp := EventResponse{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartTime:   e.When.StartTime,
		EndTime:     e.When.EndTime,
		Timezone:    e.When.StartTimezone,
		IsAllDay:    e.When.IsAllDay(),
		Status:      e.Status,
		Busy:        e.Busy,
		HtmlLink:    e.HtmlLink,
	}

	// Handle all-day events
	if resp.IsAllDay {
		if e.When.Date != "" {
			t, _ := time.Parse("2006-01-02", e.When.Date)
			resp.StartTime = t.Unix()
			resp.EndTime = t.Add(24 * time.Hour).Unix()
		} else if e.When.StartDate != "" {
			st, _ := time.Parse("2006-01-02", e.When.StartDate)
			resp.StartTime = st.Unix()
			if e.When.EndDate != "" {
				et, _ := time.Parse("2006-01-02", e.When.EndDate)
				resp.EndTime = et.Unix()
			}
		}
	}

	// Convert participants
	for _, p := range e.Participants {
		resp.Participants = append(resp.Participants, EventParticipantResponse{
			Name:   p.Name,
			Email:  p.Email,
			Status: p.Status,
		})
	}

	// Convert conferencing
	if e.Conferencing != nil && e.Conferencing.Details != nil {
		resp.Conferencing = &ConferencingResponse{
			Provider: e.Conferencing.Provider,
			URL:      e.Conferencing.Details.URL,
		}
	}

	return resp
}

// Demo data for calendars

func demoCalendars() []CalendarResponse {
	return []CalendarResponse{
		{
			ID:        "primary",
			Name:      "Personal Calendar",
			Timezone:  "America/New_York",
			IsPrimary: true,
			ReadOnly:  false,
			HexColor:  "#4285f4",
		},
		{
			ID:        "work",
			Name:      "Work Calendar",
			Timezone:  "America/New_York",
			IsPrimary: false,
			ReadOnly:  false,
			HexColor:  "#0b8043",
		},
		{
			ID:          "holidays",
			Name:        "US Holidays",
			Description: "Public holidays in the United States",
			Timezone:    "America/New_York",
			IsPrimary:   false,
			ReadOnly:    true,
			HexColor:    "#f6bf26",
		},
	}
}

func demoEvents() []EventResponse {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return []EventResponse{
		{
			ID:          "demo-event-001",
			CalendarID:  "primary",
			Title:       "Team Standup",
			Description: "Daily team sync",
			Location:    "",
			StartTime:   today.Add(9 * time.Hour).Unix(),
			EndTime:     today.Add(9*time.Hour + 30*time.Minute).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
			Participants: []EventParticipantResponse{
				{Name: "Sarah Chen", Email: "sarah@example.com", Status: "yes"},
				{Name: "Alex Johnson", Email: "alex@example.com", Status: "yes"},
			},
			Conferencing: &ConferencingResponse{
				Provider: "Google Meet",
				URL:      "https://meet.google.com/abc-defg-hij",
			},
		},
		{
			ID:          "demo-event-002",
			CalendarID:  "work",
			Title:       "Product Review",
			Description: "Weekly product roadmap review with stakeholders",
			Location:    "Conference Room A",
			StartTime:   today.Add(14 * time.Hour).Unix(),
			EndTime:     today.Add(15 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
			Participants: []EventParticipantResponse{
				{Name: "Product Team", Email: "product@example.com", Status: "yes"},
			},
		},
		{
			ID:          "demo-event-003",
			CalendarID:  "primary",
			Title:       "Lunch with Client",
			Description: "Discuss Q1 partnership opportunities",
			Location:    "Cafe Milano",
			StartTime:   today.Add(12 * time.Hour).Unix(),
			EndTime:     today.Add(13 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
		},
		{
			ID:          "demo-event-004",
			CalendarID:  "primary",
			Title:       "Focus Time",
			Description: "Deep work - no meetings",
			StartTime:   today.Add(15 * time.Hour).Unix(),
			EndTime:     today.Add(17 * time.Hour).Unix(),
			Timezone:    "America/New_York",
			IsAllDay:    false,
			Status:      "confirmed",
			Busy:        true,
		},
		{
			ID:         "demo-event-005",
			CalendarID: "holidays",
			Title:      "Christmas Day",
			StartTime:  time.Date(now.Year(), 12, 25, 0, 0, 0, 0, now.Location()).Unix(),
			EndTime:    time.Date(now.Year(), 12, 26, 0, 0, 0, 0, now.Location()).Unix(),
			Timezone:   "America/New_York",
			IsAllDay:   true,
			Status:     "confirmed",
			Busy:       false,
		},
	}
}

// ====================================
// EVENT CRUD HANDLERS
// ====================================

// handleEventsRoute handles /api/events: GET (list) and POST (create).
func (s *Server) handleEventsRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListEvents(w, r)
	case http.MethodPost:
		s.handleCreateEvent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleEventByID handles single event operations: GET, PUT, DELETE.
func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	// Parse event ID and calendar ID from path: /api/events/{id}?calendar_id=xxx
	path := strings.TrimPrefix(r.URL.Path, "/api/events/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Event ID required", http.StatusBadRequest)
		return
	}
	eventID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.handleGetEvent(w, r, eventID)
	case http.MethodPut:
		s.handleUpdateEvent(w, r, eventID)
	case http.MethodDelete:
		s.handleDeleteEvent(w, r, eventID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateEvent creates a new event.
func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	// Demo mode: simulate success
	if s.demoMode {
		now := time.Now()
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Event: &EventResponse{
				ID:         "demo-event-new-" + now.Format("20060102150405"),
				CalendarID: "primary",
				Title:      "New Event",
				StartTime:  now.Add(1 * time.Hour).Unix(),
				EndTime:    now.Add(2 * time.Hour).Unix(),
				Status:     "confirmed",
				Busy:       true,
			},
			Message: "Event created (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req CreateEventRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Title is required",
		})
		return
	}

	calendarID := req.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	// Build domain request
	createReq := &domain.CreateEventRequest{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		Busy:        req.Busy,
	}

	// Set event time
	if req.IsAllDay {
		// All-day event: use date format
		startDate := time.Unix(req.StartTime, 0).Format("2006-01-02")
		endDate := time.Unix(req.EndTime, 0).Format("2006-01-02")
		createReq.When = domain.EventWhen{
			StartDate: startDate,
			EndDate:   endDate,
			Object:    "datespan",
		}
	} else {
		// Timed event
		createReq.When = domain.EventWhen{
			StartTime:     req.StartTime,
			EndTime:       req.EndTime,
			StartTimezone: req.Timezone,
			EndTimezone:   req.Timezone,
			Object:        "timespan",
		}
	}

	// Convert participants
	for _, p := range req.Participants {
		createReq.Participants = append(createReq.Participants, domain.Participant{
			Name:  p.Name,
			Email: p.Email,
		})
	}

	// Create event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.CreateEvent(ctx, grantID, calendarID, createReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to create event: " + err.Error(),
		})
		return
	}

	eventResp := eventToResponse(*event)
	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Event:   &eventResp,
		Message: "Event created successfully",
	})
}

// handleGetEvent retrieves a single event.
func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: return mock event
	if s.demoMode {
		events := demoEvents()
		for _, e := range events {
			if e.ID == eventID {
				writeJSON(w, http.StatusOK, e)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found"})
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

	// Fetch event from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.GetEvent(ctx, grantID, calendarID, eventID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch event: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, eventToResponse(*event))
}

// handleUpdateEvent updates an existing event.
func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Event: &EventResponse{
				ID:         eventID,
				CalendarID: calendarID,
				Title:      "Updated Event",
				Status:     "confirmed",
			},
			Message: "Event updated (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req UpdateEventRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Build domain update request
	updateReq := &domain.UpdateEventRequest{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		Busy:        req.Busy,
	}

	// Set event time if provided
	if req.StartTime != nil && req.EndTime != nil {
		when := &domain.EventWhen{}
		if req.IsAllDay != nil && *req.IsAllDay {
			// All-day event
			startDate := time.Unix(*req.StartTime, 0).Format("2006-01-02")
			endDate := time.Unix(*req.EndTime, 0).Format("2006-01-02")
			when.StartDate = startDate
			when.EndDate = endDate
			when.Object = "datespan"
		} else {
			// Timed event
			when.StartTime = *req.StartTime
			when.EndTime = *req.EndTime
			if req.Timezone != nil {
				when.StartTimezone = *req.Timezone
				when.EndTimezone = *req.Timezone
			}
			when.Object = "timespan"
		}
		updateReq.When = when
	}

	// Convert participants if provided
	if len(req.Participants) > 0 {
		for _, p := range req.Participants {
			updateReq.Participants = append(updateReq.Participants, domain.Participant{
				Name:  p.Name,
				Email: p.Email,
			})
		}
	}

	// Update event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	event, err := s.nylasClient.UpdateEvent(ctx, grantID, calendarID, eventID, updateReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to update event: " + err.Error(),
		})
		return
	}

	eventResp := eventToResponse(*event)
	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Event:   &eventResp,
		Message: "Event updated successfully",
	})
}

// handleDeleteEvent deletes an event.
func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	calendarID := r.URL.Query().Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, EventActionResponse{
			Success: true,
			Message: "Event deleted (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, EventActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, EventActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Delete event via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.nylasClient.DeleteEvent(ctx, grantID, calendarID, eventID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, EventActionResponse{
			Success: false,
			Error:   "Failed to delete event: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, EventActionResponse{
		Success: true,
		Message: "Event deleted successfully",
	})
}

// ====================================
// CONTACTS HANDLERS
// ====================================

// handleContactsRoute handles /api/contacts: GET (list) and POST (create).
func (s *Server) handleContactsRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListContacts(w, r)
	case http.MethodPost:
		s.handleCreateContact(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListContacts returns contacts with optional filtering.
func (s *Server) handleListContacts(w http.ResponseWriter, r *http.Request) {
	// Demo mode: return mock contacts
	if s.demoMode {
		writeJSON(w, http.StatusOK, ContactsResponse{
			Contacts: demoContacts(),
			HasMore:  false,
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
	query := r.URL.Query()
	limit := 50
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	params := &domain.ContactQueryParams{
		Limit: limit,
	}

	// Filter by email
	if email := query.Get("email"); email != "" {
		params.Email = email
	}

	// Filter by source
	if source := query.Get("source"); source != "" {
		params.Source = source
	}

	// Filter by group
	group := query.Get("group")
	if group != "" {
		params.Group = group
	}

	// Cursor for pagination
	cursor := query.Get("cursor")
	if cursor != "" {
		params.PageToken = cursor
	}

	// Get account email for cache lookup
	accountEmail := s.getAccountEmail(grantID)

	// Try cache first (only for first page without complex filters)
	if cursor == "" && params.Email == "" && params.Source == "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getContactStore(accountEmail); err == nil {
			cacheOpts := cache.ContactListOptions{
				Group: group,
				Limit: limit,
			}
			if cached, err := store.List(cacheOpts); err == nil && len(cached) > 0 {
				resp := ContactsResponse{
					Contacts: make([]ContactResponse, 0, len(cached)),
					HasMore:  len(cached) >= limit,
				}
				for _, c := range cached {
					resp.Contacts = append(resp.Contacts, cachedContactToResponse(c))
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch contacts from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetContactsWithCursor(ctx, grantID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch contacts: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := ContactsResponse{
		Contacts:   make([]ContactResponse, 0, len(result.Data)),
		NextCursor: result.Pagination.NextCursor,
		HasMore:    result.Pagination.HasMore,
	}
	for _, c := range result.Data {
		resp.Contacts = append(resp.Contacts, contactToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleContactByID handles single contact operations: GET, PUT, DELETE.
func (s *Server) handleContactByID(w http.ResponseWriter, r *http.Request) {
	// Parse contact ID from path: /api/contacts/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/contacts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Contact ID required", http.StatusBadRequest)
		return
	}
	contactID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.handleGetContact(w, r, contactID)
	case http.MethodPut:
		s.handleUpdateContact(w, r, contactID)
	case http.MethodDelete:
		s.handleDeleteContact(w, r, contactID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetContact retrieves a single contact.
func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request, contactID string) {
	// Demo mode: return mock contact
	if s.demoMode {
		contacts := demoContacts()
		for _, c := range contacts {
			if c.ID == contactID {
				writeJSON(w, http.StatusOK, c)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Contact not found"})
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

	// Fetch contact from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	contact, err := s.nylasClient.GetContact(ctx, grantID, contactID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch contact: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, contactToResponse(*contact))
}

// handleCreateContact creates a new contact.
func (s *Server) handleCreateContact(w http.ResponseWriter, r *http.Request) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, ContactActionResponse{
			Success: true,
			Contact: &ContactResponse{
				ID:          "demo-contact-new",
				DisplayName: "New Contact",
			},
			Message: "Contact created (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, ContactActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, ContactActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req CreateContactRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ContactActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Build domain request
	createReq := &domain.CreateContactRequest{
		GivenName:   req.GivenName,
		Surname:     req.Surname,
		Nickname:    req.Nickname,
		CompanyName: req.CompanyName,
		JobTitle:    req.JobTitle,
		Birthday:    req.Birthday,
		Notes:       req.Notes,
	}

	// Convert emails
	for _, e := range req.Emails {
		createReq.Emails = append(createReq.Emails, domain.ContactEmail{
			Email: e.Email,
			Type:  e.Type,
		})
	}

	// Convert phone numbers
	for _, p := range req.PhoneNumbers {
		createReq.PhoneNumbers = append(createReq.PhoneNumbers, domain.ContactPhone{
			Number: p.Number,
			Type:   p.Type,
		})
	}

	// Convert addresses
	for _, a := range req.Addresses {
		createReq.PhysicalAddresses = append(createReq.PhysicalAddresses, domain.ContactAddress{
			Type:          a.Type,
			StreetAddress: a.StreetAddress,
			City:          a.City,
			State:         a.State,
			PostalCode:    a.PostalCode,
			Country:       a.Country,
		})
	}

	// Create contact via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	contact, err := s.nylasClient.CreateContact(ctx, grantID, createReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ContactActionResponse{
			Success: false,
			Error:   "Failed to create contact: " + err.Error(),
		})
		return
	}

	contactResp := contactToResponse(*contact)
	writeJSON(w, http.StatusOK, ContactActionResponse{
		Success: true,
		Contact: &contactResp,
		Message: "Contact created successfully",
	})
}

// handleUpdateContact updates an existing contact.
func (s *Server) handleUpdateContact(w http.ResponseWriter, r *http.Request, contactID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, ContactActionResponse{
			Success: true,
			Contact: &ContactResponse{
				ID:          contactID,
				DisplayName: "Updated Contact",
			},
			Message: "Contact updated (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, ContactActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, ContactActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Parse request body
	var req UpdateContactRequest
	if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ContactActionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Build domain update request
	updateReq := &domain.UpdateContactRequest{
		GivenName:   req.GivenName,
		Surname:     req.Surname,
		Nickname:    req.Nickname,
		CompanyName: req.CompanyName,
		JobTitle:    req.JobTitle,
		Birthday:    req.Birthday,
		Notes:       req.Notes,
	}

	// Convert emails if provided
	if len(req.Emails) > 0 {
		for _, e := range req.Emails {
			updateReq.Emails = append(updateReq.Emails, domain.ContactEmail{
				Email: e.Email,
				Type:  e.Type,
			})
		}
	}

	// Convert phone numbers if provided
	if len(req.PhoneNumbers) > 0 {
		for _, p := range req.PhoneNumbers {
			updateReq.PhoneNumbers = append(updateReq.PhoneNumbers, domain.ContactPhone{
				Number: p.Number,
				Type:   p.Type,
			})
		}
	}

	// Convert addresses if provided
	if len(req.Addresses) > 0 {
		for _, a := range req.Addresses {
			updateReq.PhysicalAddresses = append(updateReq.PhysicalAddresses, domain.ContactAddress{
				Type:          a.Type,
				StreetAddress: a.StreetAddress,
				City:          a.City,
				State:         a.State,
				PostalCode:    a.PostalCode,
				Country:       a.Country,
			})
		}
	}

	// Update contact via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	contact, err := s.nylasClient.UpdateContact(ctx, grantID, contactID, updateReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ContactActionResponse{
			Success: false,
			Error:   "Failed to update contact: " + err.Error(),
		})
		return
	}

	contactResp := contactToResponse(*contact)
	writeJSON(w, http.StatusOK, ContactActionResponse{
		Success: true,
		Contact: &contactResp,
		Message: "Contact updated successfully",
	})
}

// handleDeleteContact deletes a contact.
func (s *Server) handleDeleteContact(w http.ResponseWriter, r *http.Request, contactID string) {
	// Demo mode: simulate success
	if s.demoMode {
		writeJSON(w, http.StatusOK, ContactActionResponse{
			Success: true,
			Message: "Contact deleted (demo mode)",
		})
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, ContactActionResponse{
			Success: false,
			Error:   "Not configured. Run 'nylas auth login' first.",
		})
		return
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		writeJSON(w, http.StatusBadRequest, ContactActionResponse{
			Success: false,
			Error:   "No default account. Please select an account first.",
		})
		return
	}

	// Delete contact via Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err = s.nylasClient.DeleteContact(ctx, grantID, contactID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ContactActionResponse{
			Success: false,
			Error:   "Failed to delete contact: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, ContactActionResponse{
		Success: true,
		Message: "Contact deleted successfully",
	})
}

// handleContactGroups returns contact groups.
func (s *Server) handleContactGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock groups
	if s.demoMode {
		writeJSON(w, http.StatusOK, ContactGroupsResponse{
			Groups: demoContactGroups(),
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

	// Fetch contact groups from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	groups, err := s.nylasClient.GetContactGroups(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch contact groups: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := ContactGroupsResponse{
		Groups: make([]ContactGroupResponse, 0, len(groups)),
	}
	for _, g := range groups {
		resp.Groups = append(resp.Groups, ContactGroupResponse{
			ID:   g.ID,
			Name: g.Name,
			Path: g.Path,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleContactSearch searches contacts with text query.
func (s *Server) handleContactSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	q := query.Get("q")

	// Demo mode: filter mock contacts
	if s.demoMode {
		contacts := demoContacts()
		if q != "" {
			q = strings.ToLower(q)
			filtered := make([]ContactResponse, 0)
			for _, c := range contacts {
				if strings.Contains(strings.ToLower(c.DisplayName), q) ||
					strings.Contains(strings.ToLower(c.GivenName), q) ||
					strings.Contains(strings.ToLower(c.Surname), q) ||
					strings.Contains(strings.ToLower(c.CompanyName), q) ||
					containsEmail(c.Emails, q) {
					filtered = append(filtered, c)
				}
			}
			contacts = filtered
		}
		writeJSON(w, http.StatusOK, ContactsResponse{
			Contacts: contacts,
			HasMore:  false,
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
	limit := 50
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	params := &domain.ContactQueryParams{
		Limit: limit,
	}

	// Set email filter if query looks like email
	if strings.Contains(q, "@") {
		params.Email = q
	}

	// Cursor for pagination
	if cursor := query.Get("cursor"); cursor != "" {
		params.PageToken = cursor
	}

	// Get account email for cache search
	accountEmail := s.getAccountEmail(grantID)

	// Try cache search first
	if q != "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getContactStore(accountEmail); err == nil {
			cached, err := store.Search(q, limit)
			if err == nil && len(cached) > 0 {
				resp := ContactsResponse{
					Contacts: make([]ContactResponse, 0, len(cached)),
					HasMore:  len(cached) >= limit,
				}
				for _, c := range cached {
					resp.Contacts = append(resp.Contacts, cachedContactToResponse(c))
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	// Fetch contacts from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetContactsWithCursor(ctx, grantID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch contacts: " + err.Error(),
		})
		return
	}

	// Filter by query if provided (for non-email queries)
	var contacts []ContactResponse
	for _, c := range result.Data {
		contact := contactToResponse(c)
		if q == "" || matchesContactQuery(contact, q) {
			contacts = append(contacts, contact)
		}
	}

	resp := ContactsResponse{
		Contacts:   contacts,
		NextCursor: result.Pagination.NextCursor,
		HasMore:    result.Pagination.HasMore,
	}

	writeJSON(w, http.StatusOK, resp)
}

// containsEmail checks if any email in the list contains the query.
func containsEmail(emails []ContactEmailResponse, q string) bool {
	for _, e := range emails {
		if strings.Contains(strings.ToLower(e.Email), q) {
			return true
		}
	}
	return false
}

// matchesContactQuery checks if a contact matches the search query.
func matchesContactQuery(c ContactResponse, q string) bool {
	q = strings.ToLower(q)
	if strings.Contains(strings.ToLower(c.DisplayName), q) ||
		strings.Contains(strings.ToLower(c.GivenName), q) ||
		strings.Contains(strings.ToLower(c.Surname), q) ||
		strings.Contains(strings.ToLower(c.CompanyName), q) ||
		strings.Contains(strings.ToLower(c.Notes), q) {
		return true
	}
	return containsEmail(c.Emails, q)
}

// roundUpTo5Min rounds a Unix timestamp up to the next 5-minute boundary.
// This is required by the Nylas API for availability requests.
func roundUpTo5Min(unixTime int64) int64 {
	const fiveMinutes = 5 * 60 // 300 seconds
	remainder := unixTime % fiveMinutes
	if remainder == 0 {
		return unixTime
	}
	return unixTime + (fiveMinutes - remainder)
}

// Conversion helper for contacts

func contactToResponse(c domain.Contact) ContactResponse {
	resp := ContactResponse{
		ID:          c.ID,
		GivenName:   c.GivenName,
		Surname:     c.Surname,
		DisplayName: c.DisplayName(),
		Nickname:    c.Nickname,
		CompanyName: c.CompanyName,
		JobTitle:    c.JobTitle,
		Birthday:    c.Birthday,
		Notes:       c.Notes,
		PictureURL:  c.PictureURL,
		Source:      c.Source,
	}

	// Convert emails
	for _, e := range c.Emails {
		resp.Emails = append(resp.Emails, ContactEmailResponse{
			Email: e.Email,
			Type:  e.Type,
		})
	}

	// Convert phone numbers
	for _, p := range c.PhoneNumbers {
		resp.PhoneNumbers = append(resp.PhoneNumbers, ContactPhoneResponse{
			Number: p.Number,
			Type:   p.Type,
		})
	}

	// Convert addresses
	for _, a := range c.PhysicalAddresses {
		resp.Addresses = append(resp.Addresses, ContactAddressResponse{
			Type:          a.Type,
			StreetAddress: a.StreetAddress,
			City:          a.City,
			State:         a.State,
			PostalCode:    a.PostalCode,
			Country:       a.Country,
		})
	}

	return resp
}

// Demo data for contacts

func demoContacts() []ContactResponse {
	return []ContactResponse{
		{
			ID:          "demo-contact-001",
			GivenName:   "Sarah",
			Surname:     "Chen",
			DisplayName: "Sarah Chen",
			CompanyName: "Nylas Inc",
			JobTitle:    "Product Manager",
			Emails: []ContactEmailResponse{
				{Email: "sarah.chen@company.com", Type: "work"},
				{Email: "sarah@personal.com", Type: "home"},
			},
			PhoneNumbers: []ContactPhoneResponse{
				{Number: "+1-555-123-4567", Type: "mobile"},
			},
		},
		{
			ID:          "demo-contact-002",
			GivenName:   "Alex",
			Surname:     "Johnson",
			DisplayName: "Alex Johnson",
			CompanyName: "Nylas Inc",
			JobTitle:    "Senior Engineer",
			Emails: []ContactEmailResponse{
				{Email: "alex.johnson@nylas.com", Type: "work"},
			},
			PhoneNumbers: []ContactPhoneResponse{
				{Number: "+1-555-234-5678", Type: "work"},
			},
		},
		{
			ID:          "demo-contact-003",
			GivenName:   "Maria",
			Surname:     "Garcia",
			DisplayName: "Maria Garcia",
			CompanyName: "Acme Corp",
			JobTitle:    "VP of Sales",
			Emails: []ContactEmailResponse{
				{Email: "maria.g@acme.com", Type: "work"},
			},
			PhoneNumbers: []ContactPhoneResponse{
				{Number: "+1-555-345-6789", Type: "mobile"},
				{Number: "+1-555-345-0000", Type: "work"},
			},
			Addresses: []ContactAddressResponse{
				{
					Type:          "work",
					StreetAddress: "123 Business St",
					City:          "San Francisco",
					State:         "CA",
					PostalCode:    "94107",
					Country:       "USA",
				},
			},
		},
		{
			ID:          "demo-contact-004",
			GivenName:   "James",
			Surname:     "Wilson",
			DisplayName: "James Wilson",
			CompanyName: "Tech Solutions",
			JobTitle:    "CTO",
			Emails: []ContactEmailResponse{
				{Email: "jwilson@techsolutions.io", Type: "work"},
			},
		},
		{
			ID:          "demo-contact-005",
			GivenName:   "Emily",
			Surname:     "Brown",
			DisplayName: "Emily Brown",
			Nickname:    "Em",
			Birthday:    "1990-03-15",
			Emails: []ContactEmailResponse{
				{Email: "emily.brown@email.com", Type: "home"},
			},
			PhoneNumbers: []ContactPhoneResponse{
				{Number: "+1-555-456-7890", Type: "mobile"},
			},
		},
	}
}

func demoContactGroups() []ContactGroupResponse {
	return []ContactGroupResponse{
		{ID: "group-001", Name: "Work", Path: "/Work"},
		{ID: "group-002", Name: "Family", Path: "/Family"},
		{ID: "group-003", Name: "Friends", Path: "/Friends"},
		{ID: "group-004", Name: "VIP Clients", Path: "/Work/VIP Clients"},
	}
}

// getAccountEmail returns the email address for a grant ID.
func (s *Server) getAccountEmail(grantID string) string {
	if s.grantStore == nil {
		return ""
	}
	grants, err := s.grantStore.ListGrants()
	if err != nil {
		return ""
	}
	for _, g := range grants {
		if g.ID == grantID {
			return g.Email
		}
	}
	// Fall back to first grant
	if len(grants) > 0 {
		return grants[0].Email
	}
	return ""
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

// cachedEventToResponse converts a cached event to response format.
func cachedEventToResponse(e *cache.CachedEvent) EventResponse {
	return EventResponse{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartTime:   e.StartTime.Unix(),
		EndTime:     e.EndTime.Unix(),
		IsAllDay:    e.AllDay,
		Status:      e.Status,
		Busy:        e.Busy,
	}
}

// cachedContactToResponse converts a cached contact to response format.
func cachedContactToResponse(c *cache.CachedContact) ContactResponse {
	return ContactResponse{
		ID:          c.ID,
		GivenName:   c.GivenName,
		Surname:     c.Surname,
		DisplayName: c.DisplayName,
		Emails: []ContactEmailResponse{
			{Email: c.Email, Type: "personal"},
		},
		PhoneNumbers: []ContactPhoneResponse{
			{Number: c.Phone, Type: "mobile"},
		},
		CompanyName: c.Company,
		JobTitle:    c.JobTitle,
		Notes:       c.Notes,
	}
}

// ====================================
// AVAILABILITY & FIND TIME HANDLERS (Phase 4)
// ====================================

// AvailabilityRequest represents a request to find available times.
type AvailabilityRequest struct {
	StartTime       int64    `json:"start_time"`
	EndTime         int64    `json:"end_time"`
	DurationMinutes int      `json:"duration_minutes"`
	Participants    []string `json:"participants"` // Email addresses
	IntervalMinutes int      `json:"interval_minutes,omitempty"`
}

// AvailabilityResponse represents available meeting slots.
type AvailabilityResponse struct {
	Slots   []AvailableSlotResponse `json:"slots"`
	Message string                  `json:"message,omitempty"`
}

// AvailableSlotResponse represents a single available time slot.
type AvailableSlotResponse struct {
	StartTime int64    `json:"start_time"`
	EndTime   int64    `json:"end_time"`
	Emails    []string `json:"emails,omitempty"`
}

// FreeBusyRequest represents a request to get free/busy info.
type FreeBusyRequest struct {
	StartTime int64    `json:"start_time"`
	EndTime   int64    `json:"end_time"`
	Emails    []string `json:"emails"`
}

// FreeBusyResponse represents free/busy data for participants.
type FreeBusyResponse struct {
	Data []FreeBusyCalendarResponse `json:"data"`
}

// FreeBusyCalendarResponse represents a calendar's busy times.
type FreeBusyCalendarResponse struct {
	Email     string             `json:"email"`
	TimeSlots []TimeSlotResponse `json:"time_slots"`
}

// TimeSlotResponse represents a busy or free time slot.
type TimeSlotResponse struct {
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Status    string `json:"status"` // busy, free
}

// ConflictsResponse represents conflicting events.
type ConflictsResponse struct {
	Conflicts []EventConflict `json:"conflicts"`
	HasMore   bool            `json:"has_more"`
}

// EventConflict represents a scheduling conflict.
type EventConflict struct {
	Event1 EventResponse `json:"event1"`
	Event2 EventResponse `json:"event2"`
}

// handleAvailability finds available meeting times.
func (s *Server) handleAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock availability
	if s.demoMode {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		writeJSON(w, http.StatusOK, AvailabilityResponse{
			Slots: []AvailableSlotResponse{
				{StartTime: today.Add(10 * time.Hour).Unix(), EndTime: today.Add(11 * time.Hour).Unix()},
				{StartTime: today.Add(14 * time.Hour).Unix(), EndTime: today.Add(15 * time.Hour).Unix()},
				{StartTime: today.Add(24*time.Hour + 9*time.Hour).Unix(), EndTime: today.Add(24*time.Hour + 10*time.Hour).Unix()},
				{StartTime: today.Add(24*time.Hour + 11*time.Hour).Unix(), EndTime: today.Add(24*time.Hour + 12*time.Hour).Unix()},
				{StartTime: today.Add(24*time.Hour + 15*time.Hour).Unix(), EndTime: today.Add(24*time.Hour + 16*time.Hour).Unix()},
			},
			Message: "Demo mode: showing sample availability",
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

	// Parse request
	var req AvailabilityRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}
	} else {
		// Parse from query params for GET
		query := r.URL.Query()
		if startStr := query.Get("start_time"); startStr != "" {
			req.StartTime, _ = strconv.ParseInt(startStr, 10, 64)
		}
		if endStr := query.Get("end_time"); endStr != "" {
			req.EndTime, _ = strconv.ParseInt(endStr, 10, 64)
		}
		if durationStr := query.Get("duration_minutes"); durationStr != "" {
			req.DurationMinutes, _ = strconv.Atoi(durationStr)
		}
		if participants := query.Get("participants"); participants != "" {
			req.Participants = strings.Split(participants, ",")
		}
		if intervalStr := query.Get("interval_minutes"); intervalStr != "" {
			req.IntervalMinutes, _ = strconv.Atoi(intervalStr)
		}
	}

	// Validate request
	if req.StartTime == 0 || req.EndTime == 0 {
		// Default to next 7 days
		now := time.Now()
		req.StartTime = now.Unix()
		req.EndTime = now.Add(7 * 24 * time.Hour).Unix()
	}
	if req.DurationMinutes == 0 {
		req.DurationMinutes = 30 // Default 30 minutes
	}
	if req.IntervalMinutes == 0 {
		req.IntervalMinutes = 15 // Default 15 min intervals
	}

	// Round times to 5-minute intervals (Nylas API requirement)
	req.StartTime = roundUpTo5Min(req.StartTime)
	req.EndTime = roundUpTo5Min(req.EndTime)

	// Get current user's email if no participants specified
	if len(req.Participants) == 0 {
		email := s.getCurrentUserEmail()
		if email != "" {
			req.Participants = []string{email}
		}
	}

	if len(req.Participants) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "At least one participant email is required",
		})
		return
	}

	// Build domain request
	domainReq := &domain.AvailabilityRequest{
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		DurationMinutes: req.DurationMinutes,
		IntervalMinutes: req.IntervalMinutes,
		Participants:    make([]domain.AvailabilityParticipant, 0, len(req.Participants)),
	}
	for _, email := range req.Participants {
		domainReq.Participants = append(domainReq.Participants, domain.AvailabilityParticipant{
			Email: strings.TrimSpace(email),
		})
	}

	// Call Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetAvailability(ctx, domainReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to get availability: " + err.Error(),
		})
		return
	}

	// Convert to response
	resp := AvailabilityResponse{
		Slots: make([]AvailableSlotResponse, 0, len(result.Data.TimeSlots)),
	}
	for _, slot := range result.Data.TimeSlots {
		resp.Slots = append(resp.Slots, AvailableSlotResponse{
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
			Emails:    slot.Emails,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleFreeBusy returns free/busy information for participants.
func (s *Server) handleFreeBusy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock free/busy data
	if s.demoMode {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		writeJSON(w, http.StatusOK, FreeBusyResponse{
			Data: []FreeBusyCalendarResponse{
				{
					Email: "demo@example.com",
					TimeSlots: []TimeSlotResponse{
						{StartTime: today.Add(9 * time.Hour).Unix(), EndTime: today.Add(10 * time.Hour).Unix(), Status: "busy"},
						{StartTime: today.Add(12 * time.Hour).Unix(), EndTime: today.Add(13 * time.Hour).Unix(), Status: "busy"},
						{StartTime: today.Add(14 * time.Hour).Unix(), EndTime: today.Add(15 * time.Hour).Unix(), Status: "busy"},
					},
				},
			},
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

	// Parse request
	var req FreeBusyRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(limitedBody(w, r)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}
	} else {
		// Parse from query params for GET
		query := r.URL.Query()
		if startStr := query.Get("start_time"); startStr != "" {
			req.StartTime, _ = strconv.ParseInt(startStr, 10, 64)
		}
		if endStr := query.Get("end_time"); endStr != "" {
			req.EndTime, _ = strconv.ParseInt(endStr, 10, 64)
		}
		if emails := query.Get("emails"); emails != "" {
			req.Emails = strings.Split(emails, ",")
		}
	}

	// Validate and set defaults
	if req.StartTime == 0 || req.EndTime == 0 {
		// Default to next 7 days
		now := time.Now()
		req.StartTime = now.Unix()
		req.EndTime = now.Add(7 * 24 * time.Hour).Unix()
	}

	// Get current user's email if no emails specified
	if len(req.Emails) == 0 {
		email := s.getCurrentUserEmail()
		if email != "" {
			req.Emails = []string{email}
		}
	}

	if len(req.Emails) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "At least one email is required",
		})
		return
	}

	// Build domain request
	domainReq := &domain.FreeBusyRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Emails:    req.Emails,
	}

	// Call Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.nylasClient.GetFreeBusy(ctx, grantID, domainReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to get free/busy: " + err.Error(),
		})
		return
	}

	// Convert to response
	resp := FreeBusyResponse{
		Data: make([]FreeBusyCalendarResponse, 0, len(result.Data)),
	}
	for _, cal := range result.Data {
		calResp := FreeBusyCalendarResponse{
			Email:     cal.Email,
			TimeSlots: make([]TimeSlotResponse, 0, len(cal.TimeSlots)),
		}
		for _, slot := range cal.TimeSlots {
			calResp.TimeSlots = append(calResp.TimeSlots, TimeSlotResponse{
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
				Status:    slot.Status,
			})
		}
		resp.Data = append(resp.Data, calResp)
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleConflicts detects scheduling conflicts in events.
func (s *Server) handleConflicts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return sample conflicts
	if s.demoMode {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		writeJSON(w, http.StatusOK, ConflictsResponse{
			Conflicts: []EventConflict{
				{
					Event1: EventResponse{
						ID:        "demo-event-conflict-1",
						Title:     "Team Meeting",
						StartTime: today.Add(14 * time.Hour).Unix(),
						EndTime:   today.Add(15 * time.Hour).Unix(),
						Busy:      true,
					},
					Event2: EventResponse{
						ID:        "demo-event-conflict-2",
						Title:     "Client Call",
						StartTime: today.Add(14*time.Hour + 30*time.Minute).Unix(),
						EndTime:   today.Add(15*time.Hour + 30*time.Minute).Unix(),
						Busy:      true,
					},
				},
			},
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

	// Parse query params
	query := r.URL.Query()
	calendarID := query.Get("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	// Parse time range
	var startTime, endTime int64
	if startStr := query.Get("start_time"); startStr != "" {
		startTime, _ = strconv.ParseInt(startStr, 10, 64)
	}
	if endStr := query.Get("end_time"); endStr != "" {
		endTime, _ = strconv.ParseInt(endStr, 10, 64)
	}

	// Default to current week
	if startTime == 0 || endTime == 0 {
		now := time.Now()
		weekday := int(now.Weekday())
		startOfWeek := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		endOfWeek := startOfWeek.AddDate(0, 0, 7)
		startTime = startOfWeek.Unix()
		endTime = endOfWeek.Unix()
	}

	// Fetch events
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	params := &domain.EventQueryParams{
		Limit:           200,
		Start:           startTime,
		End:             endTime,
		ExpandRecurring: true,
	}

	result, err := s.nylasClient.GetEventsWithCursor(ctx, grantID, calendarID, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch events: " + err.Error(),
		})
		return
	}

	// Find conflicts (overlapping busy events)
	conflicts := findConflicts(result.Data)

	resp := ConflictsResponse{
		Conflicts: conflicts,
		HasMore:   false,
	}

	writeJSON(w, http.StatusOK, resp)
}

// findConflicts detects overlapping events.
func findConflicts(events []domain.Event) []EventConflict {
	var conflicts []EventConflict

	// Filter to only busy events
	var busyEvents []domain.Event
	for _, e := range events {
		if e.Busy && e.Status != "cancelled" {
			busyEvents = append(busyEvents, e)
		}
	}

	// Check each pair for overlap
	for i := 0; i < len(busyEvents); i++ {
		for j := i + 1; j < len(busyEvents); j++ {
			e1, e2 := busyEvents[i], busyEvents[j]

			// Get start/end times
			start1, end1 := e1.When.StartTime, e1.When.EndTime
			start2, end2 := e2.When.StartTime, e2.When.EndTime

			// Handle all-day events
			if e1.When.IsAllDay() {
				if e1.When.Date != "" {
					t, _ := time.Parse("2006-01-02", e1.When.Date)
					start1 = t.Unix()
					end1 = t.Add(24 * time.Hour).Unix()
				} else if e1.When.StartDate != "" {
					t, _ := time.Parse("2006-01-02", e1.When.StartDate)
					start1 = t.Unix()
					if e1.When.EndDate != "" {
						et, _ := time.Parse("2006-01-02", e1.When.EndDate)
						end1 = et.Unix()
					} else {
						end1 = start1 + 24*60*60
					}
				}
			}
			if e2.When.IsAllDay() {
				if e2.When.Date != "" {
					t, _ := time.Parse("2006-01-02", e2.When.Date)
					start2 = t.Unix()
					end2 = t.Add(24 * time.Hour).Unix()
				} else if e2.When.StartDate != "" {
					t, _ := time.Parse("2006-01-02", e2.When.StartDate)
					start2 = t.Unix()
					if e2.When.EndDate != "" {
						et, _ := time.Parse("2006-01-02", e2.When.EndDate)
						end2 = et.Unix()
					} else {
						end2 = start2 + 24*60*60
					}
				}
			}

			// Check for overlap: start1 < end2 && start2 < end1
			if start1 < end2 && start2 < end1 {
				conflicts = append(conflicts, EventConflict{
					Event1: eventToResponse(e1),
					Event2: eventToResponse(e2),
				})
			}
		}
	}

	return conflicts
}
