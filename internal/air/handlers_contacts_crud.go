package air

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

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

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
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
