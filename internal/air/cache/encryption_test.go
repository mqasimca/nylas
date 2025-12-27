package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// Tests for encryption.go functions that aren't covered in cache_test.go

// ================================
// CACHE.GO TESTS
// ================================

func TestOpenSharedDB(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := OpenSharedDB(tmpDir, "shared.db")
	if err != nil {
		t.Fatalf("OpenSharedDB() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db == nil {
		t.Fatal("OpenSharedDB() returned nil")
	}

	// Verify database is usable
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Errorf("Failed to create table in shared DB: %v", err)
	}

	// Verify file was created
	dbPath := filepath.Join(tmpDir, "shared.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file not created at %s", dbPath)
	}
}

func TestOpenSharedDB_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "path")

	db, err := OpenSharedDB(nestedPath, "test.db")
	if err != nil {
		t.Fatalf("OpenSharedDB() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify nested directory was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}

func TestOpenSharedDB_WALMode(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := OpenSharedDB(tmpDir, "wal_test.db")
	if err != nil {
		t.Fatalf("OpenSharedDB() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Check journal mode is WAL
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Journal mode = %q, want 'wal'", journalMode)
	}
}

// ================================
// SETTINGS.GO ADDITIONAL TESTS
// ================================

func TestLoadSettings_NewFile(t *testing.T) {
	tmpDir := t.TempDir()

	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}

	if settings == nil {
		t.Fatal("LoadSettings() returned nil")
	}

	// Should have default values
	if !settings.Enabled {
		t.Error("Default Enabled should be true")
	}
	if settings.MaxSizeMB != 500 {
		t.Errorf("Default MaxSizeMB = %d, want 500", settings.MaxSizeMB)
	}

	// File should exist now
	settingsPath := filepath.Join(tmpDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file should have been created")
	}
}

func TestLoadSettings_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom settings file
	settingsPath := filepath.Join(tmpDir, "settings.json")
	customSettings := `{
		"cache_enabled": false,
		"cache_max_size_mb": 1000,
		"cache_ttl_days": 60,
		"theme": "light"
	}`
	if err := os.WriteFile(settingsPath, []byte(customSettings), 0600); err != nil {
		t.Fatalf("Failed to create settings file: %v", err)
	}

	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}

	if settings.Enabled {
		t.Error("Enabled should be false from custom settings")
	}
	if settings.MaxSizeMB != 1000 {
		t.Errorf("MaxSizeMB = %d, want 1000", settings.MaxSizeMB)
	}
	if settings.TTLDays != 60 {
		t.Errorf("TTLDays = %d, want 60", settings.TTLDays)
	}
	if settings.Theme != "light" {
		t.Errorf("Theme = %q, want 'light'", settings.Theme)
	}
}

func TestLoadSettings_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid JSON file
	settingsPath := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte("not valid json{"), 0600); err != nil {
		t.Fatalf("Failed to create settings file: %v", err)
	}

	_, err := LoadSettings(tmpDir)
	if err == nil {
		t.Error("LoadSettings() should error on invalid JSON")
	}
}

func TestSettings_SetMaxSize_Bounds(t *testing.T) {
	tmpDir := t.TempDir()

	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}

	// Test minimum bound
	if err := settings.SetMaxSize(10); err != nil {
		t.Fatalf("SetMaxSize(10) error: %v", err)
	}
	if settings.MaxSizeMB != 50 {
		t.Errorf("MaxSizeMB should be clamped to minimum 50, got %d", settings.MaxSizeMB)
	}

	// Test maximum bound
	if err := settings.SetMaxSize(20000); err != nil {
		t.Fatalf("SetMaxSize(20000) error: %v", err)
	}
	if settings.MaxSizeMB != 10000 {
		t.Errorf("MaxSizeMB should be clamped to maximum 10000, got %d", settings.MaxSizeMB)
	}
}

func TestSettings_SetTheme_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}

	// Invalid theme should default to "dark"
	if err := settings.SetTheme("invalid_theme"); err != nil {
		t.Fatalf("SetTheme() error: %v", err)
	}
	if settings.Theme != "dark" {
		t.Errorf("Theme should default to 'dark', got %q", settings.Theme)
	}

	// Valid themes should work
	for _, theme := range []string{"light", "dark", "system"} {
		if err := settings.SetTheme(theme); err != nil {
			t.Fatalf("SetTheme(%q) error: %v", theme, err)
		}
		if settings.Theme != theme {
			t.Errorf("Theme = %q, want %q", settings.Theme, theme)
		}
	}
}

func TestSettings_Validate_AllErrors(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Settings)
		wantError string
	}{
		{
			name:      "MaxSizeMB too small",
			modify:    func(s *Settings) { s.MaxSizeMB = 10 },
			wantError: "cache_max_size_mb",
		},
		{
			name:      "TTLDays too small",
			modify:    func(s *Settings) { s.TTLDays = 0 },
			wantError: "cache_ttl_days",
		},
		{
			name:      "SyncIntervalMinutes too small",
			modify:    func(s *Settings) { s.SyncIntervalMinutes = 0 },
			wantError: "sync_interval_minutes",
		},
		{
			name:      "InitialSyncDays too small",
			modify:    func(s *Settings) { s.InitialSyncDays = 0 },
			wantError: "initial_sync_days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := DefaultSettings()
			tt.modify(settings)

			err := settings.Validate()
			if err == nil {
				t.Error("Validate() should return an error")
			}
		})
	}
}

// ================================
// ENCRYPTION.GO TESTS
// ================================

func TestGenerateKey(t *testing.T) {
	t.Parallel()
	key1, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() error: %v", err)
	}

	if len(key1) != keySize {
		t.Errorf("generateKey() returned key of length %d, want %d", len(key1), keySize)
	}

	// Generate another key - should be different
	key2, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() error: %v", err)
	}

	if string(key1) == string(key2) {
		t.Error("generateKey() returned identical keys on consecutive calls")
	}
}

func TestIsEncrypted_UnencryptedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create an unencrypted database
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create a simple table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		_ = db.Close()
		t.Fatalf("Failed to create table: %v", err)
	}
	_ = db.Close()

	// Check if detected as unencrypted
	isEnc, err := IsEncrypted(dbPath)
	if err != nil {
		t.Fatalf("IsEncrypted() error: %v", err)
	}

	if isEnc {
		t.Error("IsEncrypted() = true for unencrypted database, want false")
	}
}

func TestIsEncrypted_NonexistentDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db")

	// Should not error, but may return false or detect as not encrypted
	_, err := IsEncrypted(dbPath)
	// The function opens the DB which creates it, so this is expected behavior
	if err != nil {
		t.Fatalf("IsEncrypted() error: %v", err)
	}
}

func TestCopyTable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source database
	srcPath := filepath.Join(tmpDir, "src.db")
	srcDB, err := sql.Open(driverName, srcPath)
	if err != nil {
		t.Fatalf("Failed to create source database: %v", err)
	}
	defer func() { _ = srcDB.Close() }()

	// Initialize schema in source
	if err := initSchema(srcDB); err != nil {
		t.Fatalf("Failed to init schema in source: %v", err)
	}

	// Insert test data into emails table
	_, err = srcDB.Exec(`
		INSERT INTO emails (id, thread_id, folder_id, subject, snippet, from_name, from_email, to_json, cc_json, bcc_json, date, unread, starred, has_attachments, body_html, body_text, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "email-1", "thread-1", "inbox", "Test Subject", "Snippet", "John", "john@test.com", "[]", "[]", "[]", 1234567890, 1, 0, 0, "<p>Test body</p>", "Test body", 1234567890)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create destination database
	dstPath := filepath.Join(tmpDir, "dst.db")
	dstDB, err := sql.Open(driverName, dstPath)
	if err != nil {
		t.Fatalf("Failed to create destination database: %v", err)
	}
	defer func() { _ = dstDB.Close() }()

	// Initialize schema in destination
	if err := initSchema(dstDB); err != nil {
		t.Fatalf("Failed to init schema in destination: %v", err)
	}

	// Copy the emails table
	if err := copyTable(srcDB, dstDB, "emails"); err != nil {
		t.Fatalf("copyTable() error: %v", err)
	}

	// Verify data was copied
	var count int
	if err := dstDB.QueryRow("SELECT COUNT(*) FROM emails").Scan(&count); err != nil {
		t.Fatalf("Failed to count copied rows: %v", err)
	}

	if count != 1 {
		t.Errorf("copyTable() copied %d rows, want 1", count)
	}

	// Verify the data integrity
	var subject string
	if err := dstDB.QueryRow("SELECT subject FROM emails WHERE id = ?", "email-1").Scan(&subject); err != nil {
		t.Fatalf("Failed to query copied data: %v", err)
	}

	if subject != "Test Subject" {
		t.Errorf("Copied subject = %q, want %q", subject, "Test Subject")
	}
}

func TestCopyTable_InvalidTable(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "src.db")
	srcDB, err := sql.Open(driverName, srcPath)
	if err != nil {
		t.Fatalf("Failed to create source database: %v", err)
	}
	defer func() { _ = srcDB.Close() }()

	dstPath := filepath.Join(tmpDir, "dst.db")
	dstDB, err := sql.Open(driverName, dstPath)
	if err != nil {
		t.Fatalf("Failed to create destination database: %v", err)
	}
	defer func() { _ = dstDB.Close() }()

	// Attempt to copy an invalid table (SQL injection attempt)
	err = copyTable(srcDB, dstDB, "users; DROP TABLE emails;--")
	if err == nil {
		t.Error("copyTable() should reject invalid table names")
	}
}

func TestCopyTable_EmptyTable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source database with empty table
	srcPath := filepath.Join(tmpDir, "src.db")
	srcDB, err := sql.Open(driverName, srcPath)
	if err != nil {
		t.Fatalf("Failed to create source database: %v", err)
	}
	defer func() { _ = srcDB.Close() }()

	if err := initSchema(srcDB); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create destination database
	dstPath := filepath.Join(tmpDir, "dst.db")
	dstDB, err := sql.Open(driverName, dstPath)
	if err != nil {
		t.Fatalf("Failed to create destination database: %v", err)
	}
	defer func() { _ = dstDB.Close() }()

	if err := initSchema(dstDB); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Copy empty table - should not error
	if err := copyTable(srcDB, dstDB, "emails"); err != nil {
		t.Errorf("copyTable() on empty table error: %v", err)
	}
}

func TestNewEncryptedManager(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{BasePath: tmpDir}
	encCfg := EncryptionConfig{Enabled: false}

	mgr, err := NewEncryptedManager(cfg, encCfg)
	if err != nil {
		t.Fatalf("NewEncryptedManager() error: %v", err)
	}
	if mgr == nil {
		t.Fatal("NewEncryptedManager() returned nil")
	}
	defer func() { _ = mgr.Close() }()

	if mgr.Manager == nil {
		t.Error("NewEncryptedManager().Manager is nil")
	}

	if mgr.keys == nil {
		t.Error("NewEncryptedManager().keys is nil")
	}
}

func TestEncryptedManager_GetDB_EncryptionDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{BasePath: tmpDir}
	encCfg := EncryptionConfig{Enabled: false}

	mgr, err := NewEncryptedManager(cfg, encCfg)
	if err != nil {
		t.Fatalf("NewEncryptedManager() error: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// With encryption disabled, should use the regular Manager.GetDB
	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB() error: %v", err)
	}

	if db == nil {
		t.Fatal("GetDB() returned nil")
	}

	// Verify it's working
	_, err = db.Exec("SELECT 1")
	if err != nil {
		t.Errorf("Database query failed: %v", err)
	}
}

func TestEncryptedManager_ClearCache_EncryptionDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{BasePath: tmpDir}
	encCfg := EncryptionConfig{Enabled: false}

	mgr, err := NewEncryptedManager(cfg, encCfg)
	if err != nil {
		t.Fatalf("NewEncryptedManager() error: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	email := "test@example.com"

	// Create a database first
	_, err = mgr.GetDB(email)
	if err != nil {
		t.Fatalf("GetDB() error: %v", err)
	}

	// Clear the cache
	err = mgr.ClearCache(email)
	if err != nil {
		t.Fatalf("ClearCache() error: %v", err)
	}

	// Verify database file is gone
	dbPath := mgr.DBPath(email)
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("Database file should be deleted after ClearCache")
	}
}

func TestMigrateToEncrypted_NonexistentDB(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{BasePath: tmpDir}
	encCfg := EncryptionConfig{Enabled: true, KeyID: "test"}

	mgr, err := NewEncryptedManager(cfg, encCfg)
	if err != nil {
		t.Fatalf("NewEncryptedManager() error: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Migration of nonexistent database should not error
	err = mgr.MigrateToEncrypted("nonexistent@example.com")
	if err != nil {
		t.Errorf("MigrateToEncrypted() on nonexistent DB error: %v", err)
	}
}

func TestMigrateToUnencrypted_NonexistentDB(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{BasePath: tmpDir}
	encCfg := EncryptionConfig{Enabled: true, KeyID: "test"}

	mgr, err := NewEncryptedManager(cfg, encCfg)
	if err != nil {
		t.Fatalf("NewEncryptedManager() error: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Migration of nonexistent database should not error
	err = mgr.MigrateToUnencrypted("nonexistent@example.com")
	if err != nil {
		t.Errorf("MigrateToUnencrypted() on nonexistent DB error: %v", err)
	}
}

func TestCopyTable_AllowedTables(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "src.db")
	srcDB, err := sql.Open(driverName, srcPath)
	if err != nil {
		t.Fatalf("Failed to create source database: %v", err)
	}
	defer func() { _ = srcDB.Close() }()

	if err := initSchema(srcDB); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	dstPath := filepath.Join(tmpDir, "dst.db")
	dstDB, err := sql.Open(driverName, dstPath)
	if err != nil {
		t.Fatalf("Failed to create destination database: %v", err)
	}
	defer func() { _ = dstDB.Close() }()

	if err := initSchema(dstDB); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Test all allowed tables can be copied
	for table := range allowedTables {
		t.Run(table, func(t *testing.T) {
			err := copyTable(srcDB, dstDB, table)
			if err != nil {
				t.Errorf("copyTable(%q) error: %v", table, err)
			}
		})
	}
}
