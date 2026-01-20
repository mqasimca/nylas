package air

import (
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

// handleGetContact retrieves a single contact.
func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request, contactID string) {
	// Special demo mode: return specific contact or 404
	if s.demoMode {
		for _, c := range demoContacts() {
			if c.ID == contactID {
				writeJSON(w, http.StatusOK, c)
				return
			}
		}
		writeError(w, http.StatusNotFound, "Contact not found")
		return
	}
	grantID := s.withAuthGrant(w, nil) // Demo mode already handled above
	if grantID == "" {
		return
	}

	ctx, cancel := s.withTimeout(r)
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
	grantID := s.withAuthGrant(w, ContactActionResponse{
		Success: true,
		Contact: &ContactResponse{ID: "demo-contact-new", DisplayName: "New Contact"},
		Message: "Contact created (demo mode)",
	})
	if grantID == "" {
		return
	}

	var req CreateContactRequest
	if !parseJSONBody(w, r, &req) {
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
	ctx, cancel := s.withTimeout(r)
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
	grantID := s.withAuthGrant(w, ContactActionResponse{
		Success: true,
		Contact: &ContactResponse{ID: contactID, DisplayName: "Updated Contact"},
		Message: "Contact updated (demo mode)",
	})
	if grantID == "" {
		return
	}

	var req UpdateContactRequest
	if !parseJSONBody(w, r, &req) {
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
	ctx, cancel := s.withTimeout(r)
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
	grantID := s.withAuthGrant(w, ContactActionResponse{Success: true, Message: "Contact deleted (demo mode)"})
	if grantID == "" {
		return
	}

	ctx, cancel := s.withTimeout(r)
	defer cancel()

	err := s.nylasClient.DeleteContact(ctx, grantID, contactID)
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
