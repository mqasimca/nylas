package cache

import (
	"testing"
	"time"
)

// ================================
// ENCRYPTION HELPER TESTS
// ================================

func TestSettingsSetEncryption(t *testing.T) {
	tmpDir := t.TempDir()
	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Enable encryption
	if err := settings.SetEncryption(true); err != nil {
		t.Fatalf("SetEncryption failed: %v", err)
	}
	if !settings.EncryptionEnabled {
		t.Error("EncryptionEnabled should be true")
	}

	// Disable encryption
	if err := settings.SetEncryption(false); err != nil {
		t.Fatalf("SetEncryption (disable) failed: %v", err)
	}
	if settings.EncryptionEnabled {
		t.Error("EncryptionEnabled should be false")
	}
}

func TestSettingsToConfig(t *testing.T) {
	tmpDir := t.TempDir()
	settings, _ := LoadSettings(tmpDir)
	settings.MaxSizeMB = 1000
	settings.TTLDays = 60
	settings.SyncIntervalMinutes = 10

	config := settings.ToConfig(tmpDir)

	if config.BasePath != tmpDir {
		t.Errorf("BasePath = %s, want %s", config.BasePath, tmpDir)
	}
	if config.MaxSizeMB != 1000 {
		t.Errorf("MaxSizeMB = %d, want 1000", config.MaxSizeMB)
	}
	if config.TTLDays != 60 {
		t.Errorf("TTLDays = %d, want 60", config.TTLDays)
	}
	if config.SyncIntervalMinutes != 10 {
		t.Errorf("SyncIntervalMinutes = %d, want 10", config.SyncIntervalMinutes)
	}
}

func TestSettingsToEncryptionConfig(t *testing.T) {
	tmpDir := t.TempDir()
	settings, _ := LoadSettings(tmpDir)
	settings.EncryptionEnabled = true

	encConfig := settings.ToEncryptionConfig()

	if !encConfig.Enabled {
		t.Error("Enabled should be true")
	}

	// Test with encryption disabled
	settings.EncryptionEnabled = false
	encConfig = settings.ToEncryptionConfig()
	if encConfig.Enabled {
		t.Error("Enabled should be false")
	}
}

// ================================
// ADDITIONAL SEARCH TESTS
// ================================

func TestSearchWithQuery(t *testing.T) {
	db := setupTestDB(t)

	// Add test data
	emailStore := NewEmailStore(db)
	now := time.Now()
	emails := []*CachedEmail{
		{ID: "1", Subject: "Meeting notes", FromName: "Alice", FromEmail: "alice@example.com", Unread: true, Starred: false, FolderID: "inbox", Date: now},
		{ID: "2", Subject: "Project update", FromName: "Bob", FromEmail: "bob@example.com", Unread: false, Starred: true, FolderID: "inbox", Date: now.Add(-time.Hour)},
		{ID: "3", Subject: "Meeting reminder", FromName: "Charlie", FromEmail: "charlie@example.com", Unread: true, Starred: false, FolderID: "sent", Date: now.Add(-2 * time.Hour)},
	}
	_ = emailStore.PutBatch(emails)

	// Test with parsed query - filter by from and folder
	query := &SearchQuery{
		From: "alice",
		In:   "inbox",
	}
	results, err := emailStore.SearchWithQuery(query, 10)
	if err != nil {
		t.Fatalf("SearchWithQuery failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("SearchWithQuery returned %d results, want 1", len(results))
	}

	// Test with unread filter
	query2 := &SearchQuery{
		IsUnread: boolPtr(true),
	}
	results2, err := emailStore.SearchWithQuery(query2, 10)
	if err != nil {
		t.Fatalf("SearchWithQuery (unread) failed: %v", err)
	}
	if len(results2) != 2 {
		t.Errorf("SearchWithQuery (unread) returned %d results, want 2", len(results2))
	}
}

func TestSearchAdvanced(t *testing.T) {
	db := setupTestDB(t)

	// Add test data
	emailStore := NewEmailStore(db)
	now := time.Now()
	emails := []*CachedEmail{
		{ID: "1", Subject: "Quarterly report", FromName: "Alice", FromEmail: "alice@example.com", Unread: true, FolderID: "inbox", Date: now},
		{ID: "2", Subject: "Weekly update", FromName: "Bob", FromEmail: "bob@example.com", Unread: false, FolderID: "inbox", Date: now.Add(-time.Hour)},
		{ID: "3", Subject: "Quarterly review", FromName: "Charlie", FromEmail: "charlie@example.com", Unread: false, FolderID: "sent", Date: now.Add(-2 * time.Hour)},
	}
	_ = emailStore.PutBatch(emails)

	// Test SearchAdvanced with operator string "in:inbox from:alice"
	results, err := emailStore.SearchAdvanced("from:alice in:inbox", 10)
	if err != nil {
		t.Fatalf("SearchAdvanced failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("SearchAdvanced returned %d results, want 1", len(results))
	}

	// Test with is:unread operator
	results2, err := emailStore.SearchAdvanced("is:unread", 10)
	if err != nil {
		t.Fatalf("SearchAdvanced (unread) failed: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("SearchAdvanced (unread) returned %d results, want 1", len(results2))
	}
}

// ================================
// MANAGER TESTS
// ================================

func TestManagerClearAllCaches(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Create multiple databases
	emails := []string{"user1@example.com", "user2@example.com"}
	for _, email := range emails {
		db, err := mgr.GetDB(email)
		if err != nil {
			t.Fatalf("GetDB(%s) failed: %v", email, err)
		}
		// Add some data
		store := NewEmailStore(db)
		_ = store.Put(&CachedEmail{ID: "1", Subject: "Test", Date: time.Now()})
	}

	// Clear all caches
	err = mgr.ClearAllCaches()
	if err != nil {
		t.Fatalf("ClearAllCaches failed: %v", err)
	}

	// Verify all databases deleted
	accounts, _ := mgr.ListCachedAccounts()
	if len(accounts) != 0 {
		t.Errorf("ListCachedAccounts returned %d, want 0", len(accounts))
	}
}
