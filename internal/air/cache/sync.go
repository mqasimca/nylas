package cache

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SyncState tracks sync progress for a resource.
type SyncState struct {
	Resource string
	LastSync time.Time
	Cursor   string
	Metadata map[string]string
}

// SyncStore provides sync state operations.
type SyncStore struct {
	db *sql.DB
}

// NewSyncStore creates a sync store for a database.
func NewSyncStore(db *sql.DB) *SyncStore {
	return &SyncStore{db: db}
}

// Get retrieves the sync state for a resource.
func (s *SyncStore) Get(resource string) (*SyncState, error) {
	row := s.db.QueryRow(`
		SELECT resource, last_sync, cursor, metadata_json
		FROM sync_state
		WHERE resource = ?
	`, resource)

	var state SyncState
	var lastSync int64
	var cursor, metadataJSON sql.NullString

	err := row.Scan(&state.Resource, &lastSync, &cursor, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state.LastSync = time.Unix(lastSync, 0)
	state.Cursor = cursor.String

	if metadataJSON.Valid && metadataJSON.String != "" {
		_ = json.Unmarshal([]byte(metadataJSON.String), &state.Metadata)
	}

	return &state, nil
}

// Set updates the sync state for a resource.
func (s *SyncStore) Set(state *SyncState) error {
	metadataJSON := ""
	if state.Metadata != nil {
		data, _ := json.Marshal(state.Metadata)
		metadataJSON = string(data)
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sync_state (resource, last_sync, cursor, metadata_json)
		VALUES (?, ?, ?, ?)
	`, state.Resource, state.LastSync.Unix(), state.Cursor, metadataJSON)
	return err
}

// UpdateCursor updates just the cursor for a resource.
func (s *SyncStore) UpdateCursor(resource, cursor string) error {
	_, err := s.db.Exec(`
		UPDATE sync_state SET cursor = ?, last_sync = ?
		WHERE resource = ?
	`, cursor, time.Now().Unix(), resource)
	return err
}

// MarkSynced updates the last sync time for a resource.
func (s *SyncStore) MarkSynced(resource string) error {
	_, err := s.db.Exec(`
		INSERT INTO sync_state (resource, last_sync, cursor, metadata_json)
		VALUES (?, ?, '', '')
		ON CONFLICT(resource) DO UPDATE SET last_sync = excluded.last_sync
	`, resource, time.Now().Unix())
	return err
}

// Delete removes sync state for a resource.
func (s *SyncStore) Delete(resource string) error {
	_, err := s.db.Exec("DELETE FROM sync_state WHERE resource = ?", resource)
	return err
}

// NeedsSync returns true if the resource needs to be synced.
func (s *SyncStore) NeedsSync(resource string, maxAge time.Duration) (bool, error) {
	state, err := s.Get(resource)
	if err != nil {
		return true, err
	}
	if state == nil {
		return true, nil
	}
	return time.Since(state.LastSync) > maxAge, nil
}

// Resource types for sync state.
const (
	ResourceEmails    = "emails"
	ResourceEvents    = "events"
	ResourceContacts  = "contacts"
	ResourceFolders   = "folders"
	ResourceCalendars = "calendars"
)
