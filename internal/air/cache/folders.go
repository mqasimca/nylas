package cache

import (
	"database/sql"
	"fmt"
	"time"
)

// CachedFolder represents an email folder stored in the cache.
type CachedFolder struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type,omitempty"` // inbox, sent, drafts, trash, etc.
	UnreadCount int       `json:"unread_count"`
	TotalCount  int       `json:"total_count"`
	CachedAt    time.Time `json:"cached_at"`
}

// FolderStore provides folder caching operations.
type FolderStore struct {
	db *sql.DB
}

// NewFolderStore creates a folder store for a database.
func NewFolderStore(db *sql.DB) *FolderStore {
	return &FolderStore{db: db}
}

// Put stores a folder in the cache.
func (s *FolderStore) Put(folder *CachedFolder) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO folders (
			id, name, type, unread_count, total_count, cached_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		folder.ID, folder.Name, folder.Type,
		folder.UnreadCount, folder.TotalCount, time.Now().Unix(),
	)
	return err
}

// PutBatch stores multiple folders in a transaction.
func (s *FolderStore) PutBatch(folders []*CachedFolder) error {
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
		INSERT OR REPLACE INTO folders (
			id, name, type, unread_count, total_count, cached_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, folder := range folders {
		_, err = stmt.Exec(
			folder.ID, folder.Name, folder.Type,
			folder.UnreadCount, folder.TotalCount, now,
		)
		if err != nil {
			return fmt.Errorf("insert folder %s: %w", folder.ID, err)
		}
	}

	return tx.Commit()
}

// Get retrieves a folder by ID.
func (s *FolderStore) Get(id string) (*CachedFolder, error) {
	row := s.db.QueryRow(`
		SELECT id, name, type, unread_count, total_count, cached_at
		FROM folders WHERE id = ?
	`, id)

	return scanFolder(row)
}

// GetByType retrieves a folder by type (inbox, sent, drafts, trash).
func (s *FolderStore) GetByType(folderType string) (*CachedFolder, error) {
	row := s.db.QueryRow(`
		SELECT id, name, type, unread_count, total_count, cached_at
		FROM folders WHERE type = ?
	`, folderType)

	return scanFolder(row)
}

// List retrieves all folders.
func (s *FolderStore) List() ([]*CachedFolder, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, unread_count, total_count, cached_at
		FROM folders
		ORDER BY
			CASE type
				WHEN 'inbox' THEN 1
				WHEN 'drafts' THEN 2
				WHEN 'sent' THEN 3
				WHEN 'trash' THEN 4
				WHEN 'spam' THEN 5
				ELSE 6
			END,
			name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query folders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var folders []*CachedFolder
	for rows.Next() {
		folder, err := scanFolderRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	return folders, rows.Err()
}

// UpdateCounts updates the unread and total counts for a folder.
func (s *FolderStore) UpdateCounts(id string, unreadCount, totalCount int) error {
	_, err := s.db.Exec(`
		UPDATE folders SET unread_count = ?, total_count = ?
		WHERE id = ?
	`, unreadCount, totalCount, id)
	return err
}

// IncrementUnread increments or decrements the unread count.
func (s *FolderStore) IncrementUnread(id string, delta int) error {
	_, err := s.db.Exec(`
		UPDATE folders SET unread_count = unread_count + ?
		WHERE id = ?
	`, delta, id)
	return err
}

// Delete removes a folder from the cache.
func (s *FolderStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM folders WHERE id = ?", id)
	return err
}

// Count returns the number of cached folders.
func (s *FolderStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM folders").Scan(&count)
	return count, err
}

// GetTotalUnread returns the total unread count across all folders.
func (s *FolderStore) GetTotalUnread() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COALESCE(SUM(unread_count), 0) FROM folders").Scan(&count)
	return count, err
}

func scanFolder(row *sql.Row) (*CachedFolder, error) {
	var folder CachedFolder
	var cachedAtUnix int64

	err := row.Scan(
		&folder.ID, &folder.Name, &folder.Type,
		&folder.UnreadCount, &folder.TotalCount, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	folder.CachedAt = time.Unix(cachedAtUnix, 0)
	return &folder, nil
}

func scanFolderRow(rows *sql.Rows) (*CachedFolder, error) {
	var folder CachedFolder
	var cachedAtUnix int64

	err := rows.Scan(
		&folder.ID, &folder.Name, &folder.Type,
		&folder.UnreadCount, &folder.TotalCount, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	folder.CachedAt = time.Unix(cachedAtUnix, 0)
	return &folder, nil
}
