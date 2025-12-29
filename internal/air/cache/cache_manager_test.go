package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// ================================
// TEST HELPERS
// ================================

// setupTestDB creates a new in-memory database for testing.
// Returns a *sql.DB that should be closed when done.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Close()
	})

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	return db
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// ================================
// MANAGER TESTS
// ================================

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"test.user@example.com", "test.user@example.com.db"},
		{"user@example.com", "user@example.com.db"},
		{"test/path@bad.com", "test_path@bad.com.db"},
		{"back\\slash@bad.com", "back_slash@bad.com.db"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := sanitizeEmail(tt.email)
			if result != tt.expected {
				t.Errorf("sanitizeEmail(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}

func TestManagerDBPath(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	email := "test.user@example.com"
	expected := filepath.Join(tmpDir, "test.user@example.com.db")
	if got := mgr.DBPath(email); got != expected {
		t.Errorf("DBPath(%q) = %q, want %q", email, got, expected)
	}
}

func TestManagerGetDB(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	email := "test@example.com"

	// Get a database (should create it)
	db1, err := mgr.GetDB(email)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}
	if db1 == nil {
		t.Fatal("GetDB returned nil")
	}

	// Get the same database again (should return cached)
	db2, err := mgr.GetDB(email)
	if err != nil {
		t.Fatalf("GetDB (2nd) failed: %v", err)
	}
	if db1 != db2 {
		t.Error("GetDB should return the same instance")
	}

	// Verify database file exists
	dbPath := mgr.DBPath(email)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file not created at %s", dbPath)
	}
}

func TestManagerClearCache(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	email := "test@example.com"

	// Create database
	_, err = mgr.GetDB(email)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	dbPath := mgr.DBPath(email)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file should exist")
	}

	// Clear cache
	err = mgr.ClearCache(email)
	if err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("Database file should be deleted after ClearCache")
	}
}

func TestManagerListCachedAccounts(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Create multiple databases
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range emails {
		if _, err := mgr.GetDB(email); err != nil {
			t.Fatalf("GetDB(%s) failed: %v", email, err)
		}
	}

	// List cached accounts
	accounts, err := mgr.ListCachedAccounts()
	if err != nil {
		t.Fatalf("ListCachedAccounts failed: %v", err)
	}

	if len(accounts) != len(emails) {
		t.Errorf("ListCachedAccounts returned %d accounts, want %d", len(accounts), len(emails))
	}
}
