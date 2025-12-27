package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CachedContact represents a contact stored in the cache.
type CachedContact struct {
	ID          string    `json:"id"`
	GivenName   string    `json:"given_name,omitempty"`
	Surname     string    `json:"surname,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Company     string    `json:"company,omitempty"`
	JobTitle    string    `json:"job_title,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	Groups      []string  `json:"groups,omitempty"`
	CachedAt    time.Time `json:"cached_at"`
}

// ContactStore provides contact caching operations.
type ContactStore struct {
	db *sql.DB
}

// NewContactStore creates a contact store for a database.
func NewContactStore(db *sql.DB) *ContactStore {
	return &ContactStore{db: db}
}

// Put stores a contact in the cache.
func (s *ContactStore) Put(contact *CachedContact) error {
	groupsJSON, _ := json.Marshal(contact.Groups)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO contacts (
			id, given_name, surname, display_name, email,
			phone, company, job_title, notes, photo_url,
			groups_json, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		contact.ID, contact.GivenName, contact.Surname, contact.DisplayName, contact.Email,
		contact.Phone, contact.Company, contact.JobTitle, contact.Notes, contact.PhotoURL,
		string(groupsJSON), time.Now().Unix(),
	)
	return err
}

// PutBatch stores multiple contacts in a transaction.
func (s *ContactStore) PutBatch(contacts []*CachedContact) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO contacts (
			id, given_name, surname, display_name, email,
			phone, company, job_title, notes, photo_url,
			groups_json, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, contact := range contacts {
		groupsJSON, _ := json.Marshal(contact.Groups)

		_, err = stmt.Exec(
			contact.ID, contact.GivenName, contact.Surname, contact.DisplayName, contact.Email,
			contact.Phone, contact.Company, contact.JobTitle, contact.Notes, contact.PhotoURL,
			string(groupsJSON), now,
		)
		if err != nil {
			return fmt.Errorf("insert contact %s: %w", contact.ID, err)
		}
	}

	return tx.Commit()
}

// Get retrieves a contact by ID.
func (s *ContactStore) Get(id string) (*CachedContact, error) {
	row := s.db.QueryRow(`
		SELECT id, given_name, surname, display_name, email,
			phone, company, job_title, notes, photo_url,
			groups_json, cached_at
		FROM contacts WHERE id = ?
	`, id)

	return scanContact(row)
}

// GetByEmail retrieves a contact by email address.
func (s *ContactStore) GetByEmail(email string) (*CachedContact, error) {
	row := s.db.QueryRow(`
		SELECT id, given_name, surname, display_name, email,
			phone, company, job_title, notes, photo_url,
			groups_json, cached_at
		FROM contacts WHERE email = ?
	`, email)

	return scanContact(row)
}

// ContactListOptions configures contact listing.
type ContactListOptions struct {
	Group  string
	Limit  int
	Offset int
}

// List retrieves contacts with pagination.
func (s *ContactStore) List(opts ContactListOptions) ([]*CachedContact, error) {
	query := `
		SELECT id, given_name, surname, display_name, email,
			phone, company, job_title, notes, photo_url,
			groups_json, cached_at
		FROM contacts
	`
	var args []any

	if opts.Group != "" {
		query += " WHERE groups_json LIKE ?"
		args = append(args, "%"+opts.Group+"%")
	}

	query += " ORDER BY display_name ASC, given_name ASC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query contacts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var contacts []*CachedContact
	for rows.Next() {
		contact, err := scanContactRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, rows.Err()
}

// Search performs full-text search on contacts.
func (s *ContactStore) Search(query string, limit int) ([]*CachedContact, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`
		SELECT c.id, c.given_name, c.surname, c.display_name, c.email,
			c.phone, c.company, c.job_title, c.notes, c.photo_url,
			c.groups_json, c.cached_at
		FROM contacts c
		WHERE c.rowid IN (
			SELECT rowid FROM contacts_fts WHERE contacts_fts MATCH ?
		)
		ORDER BY c.display_name ASC
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var contacts []*CachedContact
	for rows.Next() {
		contact, err := scanContactRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, rows.Err()
}

// Delete removes a contact from the cache.
func (s *ContactStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM contacts WHERE id = ?", id)
	return err
}

// Count returns the number of cached contacts.
func (s *ContactStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
	return count, err
}

// ListGroups returns all unique groups.
func (s *ContactStore) ListGroups() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT groups_json FROM contacts WHERE groups_json != '[]' AND groups_json != ''")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	groupSet := make(map[string]bool)
	for rows.Next() {
		var groupsJSON string
		if err := rows.Scan(&groupsJSON); err != nil {
			continue
		}
		var groups []string
		if err := json.Unmarshal([]byte(groupsJSON), &groups); err != nil {
			continue
		}
		for _, g := range groups {
			groupSet[g] = true
		}
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	return groups, nil
}

func scanContact(row *sql.Row) (*CachedContact, error) {
	var contact CachedContact
	var groupsJSON sql.NullString
	var cachedAtUnix int64

	err := row.Scan(
		&contact.ID, &contact.GivenName, &contact.Surname, &contact.DisplayName, &contact.Email,
		&contact.Phone, &contact.Company, &contact.JobTitle, &contact.Notes, &contact.PhotoURL,
		&groupsJSON, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	contact.CachedAt = time.Unix(cachedAtUnix, 0)

	if groupsJSON.Valid {
		_ = json.Unmarshal([]byte(groupsJSON.String), &contact.Groups)
	}

	return &contact, nil
}

func scanContactRow(rows *sql.Rows) (*CachedContact, error) {
	var contact CachedContact
	var groupsJSON sql.NullString
	var cachedAtUnix int64

	err := rows.Scan(
		&contact.ID, &contact.GivenName, &contact.Surname, &contact.DisplayName, &contact.Email,
		&contact.Phone, &contact.Company, &contact.JobTitle, &contact.Notes, &contact.PhotoURL,
		&groupsJSON, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	contact.CachedAt = time.Unix(cachedAtUnix, 0)

	if groupsJSON.Valid {
		_ = json.Unmarshal([]byte(groupsJSON.String), &contact.Groups)
	}

	return &contact, nil
}
