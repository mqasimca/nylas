package air

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

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
	// Parse contact ID from path: /api/contacts/{id} or /api/contacts/{id}/photo
	path := strings.TrimPrefix(r.URL.Path, "/api/contacts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Contact ID required", http.StatusBadRequest)
		return
	}
	contactID := parts[0]

	// Check for /photo suffix
	if len(parts) > 1 && parts[1] == "photo" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleContactPhoto(w, r, contactID)
		return
	}

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

// handleContactPhoto returns the contact's profile picture as an image.
// Photos are cached locally for 30 days to reduce API calls.
func (s *Server) handleContactPhoto(w http.ResponseWriter, r *http.Request, contactID string) {
	// Demo mode: return a placeholder image
	if s.demoMode {
		// Return a 1x1 transparent PNG as placeholder
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		// 1x1 transparent PNG
		transparentPNG := []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
			0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
			0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
		}
		_, _ = w.Write(transparentPNG)
		return
	}

	// Check if configured
	if s.nylasClient == nil {
		http.Error(w, "Not configured", http.StatusServiceUnavailable)
		return
	}

	// Try to serve from cache first
	if s.photoStore != nil {
		if imageData, contentType, err := s.photoStore.Get(contactID); err == nil && imageData != nil {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=86400")
			w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
			w.Header().Set("X-Cache", "HIT")
			_, _ = w.Write(imageData)
			return
		}
	}

	// Get default grant
	grantID, err := s.grantStore.GetDefaultGrant()
	if err != nil || grantID == "" {
		http.Error(w, "No default account", http.StatusBadRequest)
		return
	}

	// Fetch contact with picture from Nylas API
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	contact, err := s.nylasClient.GetContactWithPicture(ctx, grantID, contactID, true)
	if err != nil {
		http.Error(w, "Failed to fetch contact photo", http.StatusInternalServerError)
		return
	}

	// Check if contact has a picture
	if contact.Picture == "" {
		// No picture available - return 404
		http.Error(w, "No photo available", http.StatusNotFound)
		return
	}

	// Parse the base64 data URL (format: data:image/jpeg;base64,/9j/4AAQ...)
	pictureData := contact.Picture
	var contentType string
	var imageData []byte

	if strings.HasPrefix(pictureData, "data:") {
		// Parse data URL
		parts := strings.SplitN(pictureData, ",", 2)
		if len(parts) != 2 {
			http.Error(w, "Invalid image data format", http.StatusInternalServerError)
			return
		}
		// Extract content type from "data:image/jpeg;base64"
		metaParts := strings.SplitN(parts[0], ";", 2)
		contentType = strings.TrimPrefix(metaParts[0], "data:")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		// Decode base64
		imageData, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http.Error(w, "Failed to decode image data", http.StatusInternalServerError)
			return
		}
	} else {
		// Assume raw base64
		contentType = "image/jpeg"
		imageData, err = base64.StdEncoding.DecodeString(pictureData)
		if err != nil {
			http.Error(w, "Failed to decode image data", http.StatusInternalServerError)
			return
		}
	}

	// Cache the photo for future requests (30 days)
	if s.photoStore != nil {
		_ = s.photoStore.Put(contactID, contentType, imageData)
	}

	// Set headers and write image
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Browser cache for 1 day
	w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(imageData)
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

// contactToResponse converts a domain contact to an API response.
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

// demoContacts returns demo contact data.
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

// demoContactGroups returns demo contact group data.
func demoContactGroups() []ContactGroupResponse {
	return []ContactGroupResponse{
		{ID: "group-001", Name: "Work", Path: "/Work"},
		{ID: "group-002", Name: "Family", Path: "/Family"},
		{ID: "group-003", Name: "Friends", Path: "/Friends"},
		{ID: "group-004", Name: "VIP Clients", Path: "/Work/VIP Clients"},
	}
}
