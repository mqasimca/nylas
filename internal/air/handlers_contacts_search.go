package air

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

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

	query := NewQueryParams(r.URL.Query())
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
	params := &domain.ContactQueryParams{
		Limit:     query.GetLimit(50),
		PageToken: query.Get("cursor"),
	}

	// Set email filter if query looks like email
	if strings.Contains(q, "@") {
		params.Email = q
	}

	// Get account email for cache search
	accountEmail := s.getAccountEmail(grantID)

	// Try cache search first
	if q != "" && s.cacheManager != nil && s.cacheSettings != nil && s.cacheSettings.IsCacheEnabled() {
		if store, err := s.getContactStore(accountEmail); err == nil {
			cached, err := store.Search(q, params.Limit)
			if err == nil && len(cached) > 0 {
				resp := ContactsResponse{
					Contacts: make([]ContactResponse, 0, len(cached)),
					HasMore:  len(cached) >= params.Limit,
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
