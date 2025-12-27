package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CachedAttachment represents an attachment stored in the cache.
type CachedAttachment struct {
	ID          string    `json:"id"`
	EmailID     string    `json:"email_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"` // SHA256 of content
	LocalPath   string    `json:"local_path"`
	CachedAt    time.Time `json:"cached_at"`
	AccessedAt  time.Time `json:"accessed_at"`
}

// AttachmentStore provides attachment caching operations.
type AttachmentStore struct {
	db       *sql.DB
	basePath string
	maxSize  int64 // Maximum cache size in bytes
}

// NewAttachmentStore creates an attachment store.
func NewAttachmentStore(db *sql.DB, basePath string, maxSizeMB int) (*AttachmentStore, error) {
	// Create attachments table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS attachments (
			id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			content_type TEXT,
			size INTEGER NOT NULL,
			hash TEXT NOT NULL,
			local_path TEXT NOT NULL,
			cached_at INTEGER NOT NULL,
			accessed_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("create attachments table: %w", err)
	}

	// Create indexes
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_attachments_email ON attachments(email_id)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_attachments_hash ON attachments(hash)")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_attachments_accessed ON attachments(accessed_at)")

	// Ensure attachments directory exists
	attachmentsDir := filepath.Join(basePath, "attachments")
	if err := os.MkdirAll(attachmentsDir, 0700); err != nil {
		return nil, fmt.Errorf("create attachments directory: %w", err)
	}

	return &AttachmentStore{
		db:       db,
		basePath: attachmentsDir,
		maxSize:  int64(maxSizeMB) * 1024 * 1024,
	}, nil
}

// Put stores an attachment and its content.
func (s *AttachmentStore) Put(attachment *CachedAttachment, content io.Reader) error {
	// Read content and compute hash
	tempFile, err := os.CreateTemp(s.basePath, "temp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)

	size, err := io.Copy(writer, content)
	if err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("copy content: %w", err)
	}
	_ = tempFile.Close()

	hash := hex.EncodeToString(hasher.Sum(nil))
	attachment.Hash = hash
	attachment.Size = size

	// Check if content already exists (by hash)
	existingPath := filepath.Join(s.basePath, hash[:2], hash)
	if _, err := os.Stat(existingPath); err == nil {
		// Content exists, just update metadata
		attachment.LocalPath = existingPath
	} else {
		// Move temp file to final location
		dir := filepath.Join(s.basePath, hash[:2])
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create hash directory: %w", err)
		}

		attachment.LocalPath = existingPath
		if err := os.Rename(tempFile.Name(), existingPath); err != nil {
			return fmt.Errorf("move file: %w", err)
		}
	}

	now := time.Now()
	attachment.CachedAt = now
	attachment.AccessedAt = now

	// Save metadata to database
	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO attachments (
			id, email_id, filename, content_type, size, hash, local_path, cached_at, accessed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		attachment.ID, attachment.EmailID, attachment.Filename, attachment.ContentType,
		attachment.Size, attachment.Hash, attachment.LocalPath,
		attachment.CachedAt.Unix(), attachment.AccessedAt.Unix(),
	)

	return err
}

// Get retrieves an attachment by ID.
func (s *AttachmentStore) Get(id string) (*CachedAttachment, error) {
	row := s.db.QueryRow(`
		SELECT id, email_id, filename, content_type, size, hash, local_path, cached_at, accessed_at
		FROM attachments WHERE id = ?
	`, id)

	attachment, err := scanAttachment(row)
	if err != nil {
		return nil, err
	}

	// Update accessed time
	_, _ = s.db.Exec("UPDATE attachments SET accessed_at = ? WHERE id = ?", time.Now().Unix(), id)

	return attachment, nil
}

// GetByHash retrieves an attachment by content hash.
func (s *AttachmentStore) GetByHash(hash string) (*CachedAttachment, error) {
	row := s.db.QueryRow(`
		SELECT id, email_id, filename, content_type, size, hash, local_path, cached_at, accessed_at
		FROM attachments WHERE hash = ?
	`, hash)

	return scanAttachment(row)
}

// Open opens the attachment file for reading.
func (s *AttachmentStore) Open(id string) (*os.File, error) {
	attachment, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	return os.Open(attachment.LocalPath)
}

// ListByEmail retrieves all attachments for an email.
func (s *AttachmentStore) ListByEmail(emailID string) ([]*CachedAttachment, error) {
	rows, err := s.db.Query(`
		SELECT id, email_id, filename, content_type, size, hash, local_path, cached_at, accessed_at
		FROM attachments WHERE email_id = ?
		ORDER BY filename ASC
	`, emailID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var attachments []*CachedAttachment
	for rows.Next() {
		attachment, err := scanAttachmentRow(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

// Delete removes an attachment.
func (s *AttachmentStore) Delete(id string) error {
	attachment, err := s.Get(id)
	if err != nil {
		return err
	}

	// Check if any other attachments use the same hash
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM attachments WHERE hash = ? AND id != ?", attachment.Hash, id).Scan(&count)
	if err != nil {
		return err
	}

	// Delete from database
	_, err = s.db.Exec("DELETE FROM attachments WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Only delete file if no other attachments reference it
	if count == 0 {
		_ = os.Remove(attachment.LocalPath)
	}

	return nil
}

// DeleteByEmail removes all attachments for an email.
func (s *AttachmentStore) DeleteByEmail(emailID string) error {
	// Get all attachments for this email
	attachments, err := s.ListByEmail(emailID)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		if err := s.Delete(attachment.ID); err != nil {
			return err
		}
	}

	return nil
}

// TotalSize returns the total size of cached attachments.
func (s *AttachmentStore) TotalSize() (int64, error) {
	var size int64
	err := s.db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM attachments").Scan(&size)
	return size, err
}

// Count returns the number of cached attachments.
func (s *AttachmentStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM attachments").Scan(&count)
	return count, err
}

// Prune removes least recently used attachments to stay under maxSize.
func (s *AttachmentStore) Prune() (int, error) {
	currentSize, err := s.TotalSize()
	if err != nil {
		return 0, err
	}

	if currentSize <= s.maxSize {
		return 0, nil
	}

	// Get all attachments ordered by last accessed time
	rows, err := s.db.Query(`
		SELECT id, email_id, filename, content_type, size, hash, local_path, cached_at, accessed_at
		FROM attachments
		ORDER BY accessed_at ASC
	`)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	var toDelete []*CachedAttachment
	targetSize := s.maxSize * 80 / 100 // Target 80% of max

	for rows.Next() && currentSize > targetSize {
		attachment, err := scanAttachmentRow(rows)
		if err != nil {
			continue
		}
		toDelete = append(toDelete, attachment)
		currentSize -= attachment.Size
	}

	// Delete the selected attachments
	for _, attachment := range toDelete {
		_ = s.Delete(attachment.ID)
	}

	return len(toDelete), nil
}

// RemoveOrphaned removes attachment files not referenced in database.
func (s *AttachmentStore) RemoveOrphaned() (int, error) {
	// Get all known hashes
	rows, err := s.db.Query("SELECT DISTINCT hash FROM attachments")
	if err != nil {
		return 0, err
	}

	knownHashes := make(map[string]bool)
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err == nil {
			knownHashes[hash] = true
		}
	}
	_ = rows.Close()

	// Walk the attachments directory and remove unknown files
	count := 0
	err = filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if file is known
		hash := filepath.Base(path)
		if !knownHashes[hash] {
			_ = os.Remove(path)
			count++
		}

		return nil
	})

	return count, err
}

func scanAttachment(row *sql.Row) (*CachedAttachment, error) {
	var attachment CachedAttachment
	var cachedAtUnix, accessedAtUnix int64

	err := row.Scan(
		&attachment.ID, &attachment.EmailID, &attachment.Filename, &attachment.ContentType,
		&attachment.Size, &attachment.Hash, &attachment.LocalPath,
		&cachedAtUnix, &accessedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	attachment.CachedAt = time.Unix(cachedAtUnix, 0)
	attachment.AccessedAt = time.Unix(accessedAtUnix, 0)
	return &attachment, nil
}

func scanAttachmentRow(rows *sql.Rows) (*CachedAttachment, error) {
	var attachment CachedAttachment
	var cachedAtUnix, accessedAtUnix int64

	err := rows.Scan(
		&attachment.ID, &attachment.EmailID, &attachment.Filename, &attachment.ContentType,
		&attachment.Size, &attachment.Hash, &attachment.LocalPath,
		&cachedAtUnix, &accessedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	attachment.CachedAt = time.Unix(cachedAtUnix, 0)
	attachment.AccessedAt = time.Unix(accessedAtUnix, 0)
	return &attachment, nil
}

// AttachmentCacheStats contains statistics about the attachment cache.
type AttachmentCacheStats struct {
	Count     int
	TotalSize int64
	MaxSize   int64
	Usage     float64 // Percentage used
	Oldest    time.Time
	Newest    time.Time
}

// GetStats returns attachment cache statistics.
func (s *AttachmentStore) GetStats() (*AttachmentCacheStats, error) {
	stats := &AttachmentCacheStats{MaxSize: s.maxSize}

	// Count and size
	count, err := s.Count()
	if err != nil {
		return nil, err
	}
	stats.Count = count

	size, err := s.TotalSize()
	if err != nil {
		return nil, err
	}
	stats.TotalSize = size
	stats.Usage = float64(size) / float64(s.maxSize) * 100

	// Oldest and newest
	var oldestUnix, newestUnix int64
	_ = s.db.QueryRow("SELECT MIN(cached_at) FROM attachments").Scan(&oldestUnix)
	_ = s.db.QueryRow("SELECT MAX(cached_at) FROM attachments").Scan(&newestUnix)

	if oldestUnix > 0 {
		stats.Oldest = time.Unix(oldestUnix, 0)
	}
	if newestUnix > 0 {
		stats.Newest = time.Unix(newestUnix, 0)
	}

	return stats, nil
}

// LRUEvict removes the least recently used attachments to free up space.
func (s *AttachmentStore) LRUEvict(bytesToFree int64) (int, error) {
	// Get attachments ordered by access time
	rows, err := s.db.Query(`
		SELECT id, size FROM attachments
		ORDER BY accessed_at ASC
	`)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	var toDelete []string
	var freedBytes int64

	for rows.Next() && freedBytes < bytesToFree {
		var id string
		var size int64
		if err := rows.Scan(&id, &size); err == nil {
			toDelete = append(toDelete, id)
			freedBytes += size
		}
	}

	// Delete selected attachments
	for _, id := range toDelete {
		_ = s.Delete(id)
	}

	return len(toDelete), nil
}

// Helper for sorting by access time
type byAccessTime []*CachedAttachment

func (a byAccessTime) Len() int           { return len(a) }
func (a byAccessTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAccessTime) Less(i, j int) bool { return a[i].AccessedAt.Before(a[j].AccessedAt) }

var _ sort.Interface = byAccessTime{}
