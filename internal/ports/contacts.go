package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// ContactClient defines the interface for contact and contact group operations.
type ContactClient interface {
	// ================================
	// CONTACT OPERATIONS
	// ================================

	// GetContacts retrieves contacts with query parameters.
	GetContacts(ctx context.Context, grantID string, params *domain.ContactQueryParams) ([]domain.Contact, error)

	// GetContactsWithCursor retrieves contacts with cursor-based pagination.
	GetContactsWithCursor(ctx context.Context, grantID string, params *domain.ContactQueryParams) (*domain.ContactListResponse, error)

	// GetContact retrieves a specific contact.
	GetContact(ctx context.Context, grantID, contactID string) (*domain.Contact, error)

	// GetContactWithPicture retrieves a contact with optional picture data.
	GetContactWithPicture(ctx context.Context, grantID, contactID string, includePicture bool) (*domain.Contact, error)

	// CreateContact creates a new contact.
	CreateContact(ctx context.Context, grantID string, req *domain.CreateContactRequest) (*domain.Contact, error)

	// UpdateContact updates an existing contact.
	UpdateContact(ctx context.Context, grantID, contactID string, req *domain.UpdateContactRequest) (*domain.Contact, error)

	// DeleteContact deletes a contact.
	DeleteContact(ctx context.Context, grantID, contactID string) error

	// ================================
	// CONTACT GROUP OPERATIONS
	// ================================

	// GetContactGroups retrieves all contact groups.
	GetContactGroups(ctx context.Context, grantID string) ([]domain.ContactGroup, error)

	// GetContactGroup retrieves a specific contact group.
	GetContactGroup(ctx context.Context, grantID, groupID string) (*domain.ContactGroup, error)

	// CreateContactGroup creates a new contact group.
	CreateContactGroup(ctx context.Context, grantID string, req *domain.CreateContactGroupRequest) (*domain.ContactGroup, error)

	// UpdateContactGroup updates an existing contact group.
	UpdateContactGroup(ctx context.Context, grantID, groupID string, req *domain.UpdateContactGroupRequest) (*domain.ContactGroup, error)

	// DeleteContactGroup deletes a contact group.
	DeleteContactGroup(ctx context.Context, grantID, groupID string) error
}
