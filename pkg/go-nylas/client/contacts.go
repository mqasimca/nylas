package client

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/pkg/go-nylas/config"
)

// ContactsClient provides contacts operations for plugins.
type ContactsClient struct {
	adapter *nylas.HTTPClient
	grantID string
	config  *config.Config
}

// NewContactsClient creates a new contacts client for the given configuration.
func NewContactsClient(cfg *config.Config) (*ContactsClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.GetAPIKey() == "" {
		return nil, fmt.Errorf("API key not configured")
	}
	if cfg.GetGrantID() == "" {
		return nil, fmt.Errorf("grant ID not configured")
	}

	// Create internal HTTP client
	adapter := nylas.NewHTTPClient(cfg.GetAPIKey(), cfg.GetGrantID(), cfg.GetRegion())

	return &ContactsClient{
		adapter: adapter,
		grantID: cfg.GetGrantID(),
		config:  cfg,
	}, nil
}

// Contact represents a contact.
type Contact struct {
	ID                string
	GivenName         string
	MiddleName        string
	Surname           string
	Suffix            string
	Nickname          string
	Birthday          string
	CompanyName       string
	JobTitle          string
	ManagerName       string
	Notes             string
	PictureURL        string
	Emails            []ContactEmail
	PhoneNumbers      []ContactPhone
	WebPages          []ContactWebPage
	IMAddresses       []ContactIM
	PhysicalAddresses []ContactAddress
	Groups            []ContactGroupInfo
	Source            string
}

// ContactEmail represents an email address for a contact.
type ContactEmail struct {
	Type  string
	Email string
}

// ContactPhone represents a phone number for a contact.
type ContactPhone struct {
	Type   string
	Number string
}

// ContactWebPage represents a webpage for a contact.
type ContactWebPage struct {
	Type string
	URL  string
}

// ContactIM represents an instant messaging address.
type ContactIM struct {
	Type    string
	IMAddress string
}

// ContactAddress represents a physical address.
type ContactAddress struct {
	Type         string
	StreetAddress string
	City         string
	PostalCode   string
	State        string
	Country      string
}

// ContactGroupInfo represents contact group membership.
type ContactGroupInfo struct {
	ID   string
	Name string
	Path string
}

// ContactGroup represents a contact group.
type ContactGroup struct {
	ID   string
	Name string
	Path string
}

// ListContactsOptions contains options for listing contacts.
type ListContactsOptions struct {
	Limit          int
	PageToken      string
	Email          string
	PhoneNumber    string
	Source         string
	Group          string
	Recurse        bool
	ProfilePicture bool
}

// CreateContactOptions contains options for creating a contact.
type CreateContactOptions struct {
	GivenName         string
	MiddleName        string
	Surname           string
	Suffix            string
	Nickname          string
	Birthday          string
	CompanyName       string
	JobTitle          string
	ManagerName       string
	Notes             string
	Emails            []ContactEmail
	PhoneNumbers      []ContactPhone
	WebPages          []ContactWebPage
	IMAddresses       []ContactIM
	PhysicalAddresses []ContactAddress
}

// UpdateContactOptions contains options for updating a contact.
type UpdateContactOptions struct {
	GivenName         *string
	MiddleName        *string
	Surname           *string
	Suffix            *string
	Nickname          *string
	Birthday          *string
	CompanyName       *string
	JobTitle          *string
	ManagerName       *string
	Notes             *string
	Emails            []ContactEmail
	PhoneNumbers      []ContactPhone
	WebPages          []ContactWebPage
	IMAddresses       []ContactIM
	PhysicalAddresses []ContactAddress
}

// List retrieves a list of contacts.
func (c *ContactsClient) List(ctx context.Context, opts *ListContactsOptions) ([]*Contact, error) {
	if opts == nil {
		opts = &ListContactsOptions{Limit: 50}
	}

	// Convert to internal params
	params := &domain.ContactQueryParams{
		Limit:          opts.Limit,
		PageToken:      opts.PageToken,
		Email:          opts.Email,
		PhoneNumber:    opts.PhoneNumber,
		Source:         opts.Source,
		Group:          opts.Group,
		Recurse:        opts.Recurse,
		ProfilePicture: opts.ProfilePicture,
	}

	// Call internal adapter
	internalContacts, err := c.adapter.GetContacts(ctx, c.grantID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	// Convert to public format
	contacts := make([]*Contact, len(internalContacts))
	for i, con := range internalContacts {
		contacts[i] = convertContact(&con)
	}

	return contacts, nil
}

// Get retrieves a single contact by ID.
func (c *ContactsClient) Get(ctx context.Context, contactID string) (*Contact, error) {
	internalContact, err := c.adapter.GetContact(ctx, c.grantID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return convertContact(internalContact), nil
}

// Create creates a new contact.
func (c *ContactsClient) Create(ctx context.Context, opts *CreateContactOptions) (*Contact, error) {
	if opts == nil {
		return nil, fmt.Errorf("create options cannot be nil")
	}

	// Convert to internal format
	req := &domain.CreateContactRequest{
		GivenName:   opts.GivenName,
		MiddleName:  opts.MiddleName,
		Surname:     opts.Surname,
		Suffix:      opts.Suffix,
		Nickname:    opts.Nickname,
		Birthday:    opts.Birthday,
		CompanyName: opts.CompanyName,
		JobTitle:    opts.JobTitle,
		ManagerName: opts.ManagerName,
		Notes:       opts.Notes,
	}

	// Convert emails
	req.Emails = make([]domain.ContactEmail, len(opts.Emails))
	for i, e := range opts.Emails {
		req.Emails[i] = domain.ContactEmail{Type: e.Type, Email: e.Email}
	}

	// Convert phone numbers
	req.PhoneNumbers = make([]domain.ContactPhone, len(opts.PhoneNumbers))
	for i, p := range opts.PhoneNumbers {
		req.PhoneNumbers[i] = domain.ContactPhone{Type: p.Type, Number: p.Number}
	}

	// Convert web pages
	req.WebPages = make([]domain.ContactWebPage, len(opts.WebPages))
	for i, w := range opts.WebPages {
		req.WebPages[i] = domain.ContactWebPage{Type: w.Type, URL: w.URL}
	}

	// Create via internal adapter
	createdContact, err := c.adapter.CreateContact(ctx, c.grantID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return convertContact(createdContact), nil
}

// Update updates an existing contact.
func (c *ContactsClient) Update(ctx context.Context, contactID string, opts *UpdateContactOptions) (*Contact, error) {
	if opts == nil {
		return nil, fmt.Errorf("update options cannot be nil")
	}

	// Convert to internal format
	req := &domain.UpdateContactRequest{
		GivenName:   opts.GivenName,
		MiddleName:  opts.MiddleName,
		Surname:     opts.Surname,
		Suffix:      opts.Suffix,
		Nickname:    opts.Nickname,
		Birthday:    opts.Birthday,
		CompanyName: opts.CompanyName,
		JobTitle:    opts.JobTitle,
		ManagerName: opts.ManagerName,
		Notes:       opts.Notes,
	}

	// Update via internal adapter
	updatedContact, err := c.adapter.UpdateContact(ctx, c.grantID, contactID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return convertContact(updatedContact), nil
}

// Delete deletes a contact.
func (c *ContactsClient) Delete(ctx context.Context, contactID string) error {
	err := c.adapter.DeleteContact(ctx, c.grantID, contactID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	return nil
}

// ListGroups retrieves all contact groups.
func (c *ContactsClient) ListGroups(ctx context.Context) ([]*ContactGroup, error) {
	internalGroups, err := c.adapter.GetContactGroups(ctx, c.grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact groups: %w", err)
	}

	groups := make([]*ContactGroup, len(internalGroups))
	for i, g := range internalGroups {
		groups[i] = &ContactGroup{
			ID:   g.ID,
			Name: g.Name,
			Path: g.Path,
		}
	}

	return groups, nil
}

// Helper functions

func convertContact(internal *domain.Contact) *Contact {
	if internal == nil {
		return nil
	}

	contact := &Contact{
		ID:          internal.ID,
		GivenName:   internal.GivenName,
		MiddleName:  internal.MiddleName,
		Surname:     internal.Surname,
		Suffix:      internal.Suffix,
		Nickname:    internal.Nickname,
		Birthday:    internal.Birthday,
		CompanyName: internal.CompanyName,
		JobTitle:    internal.JobTitle,
		ManagerName: internal.ManagerName,
		Notes:       internal.Notes,
		PictureURL:  internal.PictureURL,
		Source:      internal.Source,
	}

	// Convert emails
	contact.Emails = make([]ContactEmail, len(internal.Emails))
	for i, e := range internal.Emails {
		contact.Emails[i] = ContactEmail{Type: e.Type, Email: e.Email}
	}

	// Convert phone numbers
	contact.PhoneNumbers = make([]ContactPhone, len(internal.PhoneNumbers))
	for i, p := range internal.PhoneNumbers {
		contact.PhoneNumbers[i] = ContactPhone{Type: p.Type, Number: p.Number}
	}

	// Convert web pages
	contact.WebPages = make([]ContactWebPage, len(internal.WebPages))
	for i, w := range internal.WebPages {
		contact.WebPages[i] = ContactWebPage{Type: w.Type, URL: w.URL}
	}

	// Convert IM addresses
	contact.IMAddresses = make([]ContactIM, len(internal.IMAddresses))
	for i, im := range internal.IMAddresses {
		contact.IMAddresses[i] = ContactIM{Type: im.Type, IMAddress: im.IMAddress}
	}

	// Convert physical addresses
	contact.PhysicalAddresses = make([]ContactAddress, len(internal.PhysicalAddresses))
	for i, addr := range internal.PhysicalAddresses {
		contact.PhysicalAddresses[i] = ContactAddress{
			Type:          addr.Type,
			StreetAddress: addr.StreetAddress,
			City:          addr.City,
			PostalCode:    addr.PostalCode,
			State:         addr.State,
			Country:       addr.Country,
		}
	}

	// Convert groups
	contact.Groups = make([]ContactGroupInfo, len(internal.Groups))
	for i, g := range internal.Groups {
		contact.Groups[i] = ContactGroupInfo{
			ID:   g.ID,
			Name: g.Name,
			Path: g.Path,
		}
	}

	return contact
}
