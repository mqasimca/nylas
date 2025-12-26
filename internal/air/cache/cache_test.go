package cache

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

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
		mgr.Close()
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

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"qasim.m@nylas.com", "qasim.m@nylas.com.db"},
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
	defer mgr.Close()

	email := "qasim.m@nylas.com"
	expected := filepath.Join(tmpDir, "qasim.m@nylas.com.db")
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
	defer mgr.Close()

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
	defer mgr.Close()

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
	defer mgr.Close()

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

func TestEmailStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewEmailStore(db)

	// Test Put and Get
	email := &CachedEmail{
		ID:        "email-123",
		ThreadID:  "thread-456",
		FolderID:  "inbox",
		Subject:   "Test Subject",
		Snippet:   "This is a test email...",
		FromName:  "John Doe",
		FromEmail: "john@example.com",
		To:        []string{"recipient@example.com"},
		Date:      time.Now().Add(-time.Hour),
		Unread:    true,
		Starred:   false,
	}

	if err := store.Put(email); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("email-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != email.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, email.ID)
	}
	if retrieved.Subject != email.Subject {
		t.Errorf("Retrieved Subject = %s, want %s", retrieved.Subject, email.Subject)
	}
	if retrieved.Unread != email.Unread {
		t.Errorf("Retrieved Unread = %v, want %v", retrieved.Unread, email.Unread)
	}

	// Test Count
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test UpdateFlags
	unread := false
	starred := true
	if err := store.UpdateFlags("email-123", &unread, &starred); err != nil {
		t.Fatalf("UpdateFlags failed: %v", err)
	}

	retrieved, _ = store.Get("email-123")
	if retrieved.Unread != false {
		t.Error("Email should be marked as read")
	}
	if retrieved.Starred != true {
		t.Error("Email should be marked as starred")
	}

	// Test Delete
	if err := store.Delete("email-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 0 {
		t.Error("Email should be deleted")
	}
}

func TestEmailStoreSearch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewEmailStore(db)

	// Add test emails
	emails := []*CachedEmail{
		{ID: "1", Subject: "Quarterly report review", FromName: "Alice", FromEmail: "alice@example.com", Date: time.Now()},
		{ID: "2", Subject: "Meeting tomorrow", FromName: "Bob", FromEmail: "bob@example.com", Date: time.Now()},
		{ID: "3", Subject: "Quarterly earnings call", FromName: "Charlie", FromEmail: "charlie@example.com", Date: time.Now()},
	}

	if err := store.PutBatch(emails); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Search for "quarterly"
	results, err := store.Search("quarterly", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search returned %d results, want 2", len(results))
	}

	// Search for "Bob"
	results, err = store.Search("Bob", 10)
	if err != nil {
		t.Fatalf("Search (Bob) failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search (Bob) returned %d results, want 1", len(results))
	}
}

func TestSyncStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewSyncStore(db)

	// Test Set and Get
	state := &SyncState{
		Resource: ResourceEmails,
		LastSync: time.Now(),
		Cursor:   "cursor-123",
		Metadata: map[string]string{"page": "5"},
	}

	if err := store.Set(state); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := store.Get(ResourceEmails)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Cursor != state.Cursor {
		t.Errorf("Cursor = %s, want %s", retrieved.Cursor, state.Cursor)
	}

	// Test NeedsSync
	needsSync, err := store.NeedsSync(ResourceEmails, time.Hour)
	if err != nil {
		t.Fatalf("NeedsSync failed: %v", err)
	}
	if needsSync {
		t.Error("Should not need sync (just synced)")
	}

	needsSync, err = store.NeedsSync(ResourceEmails, time.Millisecond)
	if err != nil {
		t.Fatalf("NeedsSync (2) failed: %v", err)
	}
	if !needsSync {
		t.Error("Should need sync (max age is 1ms)")
	}

	// Test non-existent resource
	needsSync, err = store.NeedsSync(ResourceContacts, time.Hour)
	if err != nil {
		t.Fatalf("NeedsSync (contacts) failed: %v", err)
	}
	if !needsSync {
		t.Error("Non-existent resource should need sync")
	}
}

func TestEventStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewEventStore(db)

	// Test Put and Get
	now := time.Now()
	event := &CachedEvent{
		ID:          "event-123",
		CalendarID:  "primary",
		Title:       "Team Meeting",
		Description: "Weekly sync",
		Location:    "Conference Room A",
		StartTime:   now,
		EndTime:     now.Add(time.Hour),
		AllDay:      false,
		Status:      "confirmed",
		Busy:        true,
	}

	if err := store.Put(event); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("event-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != event.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, event.ID)
	}
	if retrieved.Title != event.Title {
		t.Errorf("Retrieved Title = %s, want %s", retrieved.Title, event.Title)
	}
	if retrieved.CalendarID != event.CalendarID {
		t.Errorf("Retrieved CalendarID = %s, want %s", retrieved.CalendarID, event.CalendarID)
	}

	// Test Count
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test List with options
	events, err := store.List(EventListOptions{
		CalendarID: "primary",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("List returned %d events, want 1", len(events))
	}

	// Test ListByDateRange
	events, err = store.ListByDateRange(now.Add(-time.Hour), now.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("ListByDateRange failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("ListByDateRange returned %d events, want 1", len(events))
	}

	// Test Delete
	if err := store.Delete("event-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 0 {
		t.Error("Event should be deleted")
	}
}

func TestEventStorePutBatch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewEventStore(db)
	now := time.Now()

	events := []*CachedEvent{
		{ID: "1", CalendarID: "primary", Title: "Event 1", StartTime: now, EndTime: now.Add(time.Hour)},
		{ID: "2", CalendarID: "primary", Title: "Event 2", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)},
		{ID: "3", CalendarID: "work", Title: "Event 3", StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour)},
	}

	if err := store.PutBatch(events); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	count, _ := store.Count()
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}

	// Test filtering by calendar
	primary, err := store.List(EventListOptions{CalendarID: "primary", Limit: 10})
	if err != nil {
		t.Fatalf("List (primary) failed: %v", err)
	}
	if len(primary) != 2 {
		t.Errorf("Primary calendar has %d events, want 2", len(primary))
	}
}

func TestContactStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewContactStore(db)

	// Test Put and Get
	contact := &CachedContact{
		ID:          "contact-123",
		GivenName:   "John",
		Surname:     "Doe",
		DisplayName: "John Doe",
		Email:       "john@example.com",
		Phone:       "+1-555-123-4567",
		Company:     "Acme Corp",
		JobTitle:    "Engineer",
		Notes:       "Important client",
	}

	if err := store.Put(contact); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("contact-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != contact.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, contact.ID)
	}
	if retrieved.DisplayName != contact.DisplayName {
		t.Errorf("Retrieved DisplayName = %s, want %s", retrieved.DisplayName, contact.DisplayName)
	}
	if retrieved.Email != contact.Email {
		t.Errorf("Retrieved Email = %s, want %s", retrieved.Email, contact.Email)
	}

	// Test Count
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test List
	contacts, err := store.List(ContactListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(contacts) != 1 {
		t.Errorf("List returned %d contacts, want 1", len(contacts))
	}

	// Test Delete
	if err := store.Delete("contact-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 0 {
		t.Error("Contact should be deleted")
	}
}

func TestContactStoreSearch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewContactStore(db)

	// Add test contacts
	contacts := []*CachedContact{
		{ID: "1", GivenName: "Alice", Surname: "Smith", DisplayName: "Alice Smith", Email: "alice@example.com"},
		{ID: "2", GivenName: "Bob", Surname: "Jones", DisplayName: "Bob Jones", Email: "bob@example.com"},
		{ID: "3", GivenName: "Charlie", Surname: "Smith", DisplayName: "Charlie Smith", Email: "charlie@example.com"},
	}

	if err := store.PutBatch(contacts); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Search for "Smith"
	results, err := store.Search("Smith", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search returned %d results, want 2", len(results))
	}

	// Search for "Bob"
	results, err = store.Search("Bob", 10)
	if err != nil {
		t.Fatalf("Search (Bob) failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search (Bob) returned %d results, want 1", len(results))
	}
}

func TestFolderStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewFolderStore(db)

	// Test Put and Get
	folder := &CachedFolder{
		ID:          "folder-123",
		Name:        "INBOX",
		Type:        "inbox",
		UnreadCount: 10,
		TotalCount:  100,
	}

	if err := store.Put(folder); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("folder-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != folder.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, folder.ID)
	}
	if retrieved.Name != folder.Name {
		t.Errorf("Retrieved Name = %s, want %s", retrieved.Name, folder.Name)
	}
	if retrieved.UnreadCount != folder.UnreadCount {
		t.Errorf("Retrieved UnreadCount = %d, want %d", retrieved.UnreadCount, folder.UnreadCount)
	}

	// Test List
	folders, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(folders) != 1 {
		t.Errorf("List returned %d folders, want 1", len(folders))
	}

	// Test UpdateCounts
	if err := store.UpdateCounts("folder-123", 5, 105); err != nil {
		t.Fatalf("UpdateCounts failed: %v", err)
	}

	retrieved, _ = store.Get("folder-123")
	if retrieved.UnreadCount != 5 {
		t.Errorf("UnreadCount = %d, want 5", retrieved.UnreadCount)
	}
	if retrieved.TotalCount != 105 {
		t.Errorf("TotalCount = %d, want 105", retrieved.TotalCount)
	}

	// Test Delete
	if err := store.Delete("folder-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get("folder-123")
	if err == nil {
		t.Error("Folder should be deleted")
	}
}

func TestOfflineQueue(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Test Enqueue
	payload := MarkReadPayload{EmailID: "email-123", Unread: false}
	if err := queue.Enqueue(ActionMarkRead, "email-123", payload); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Test Count
	count, err := queue.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test HasPendingActions
	hasPending, err := queue.HasPendingActions()
	if err != nil {
		t.Fatalf("HasPendingActions failed: %v", err)
	}
	if !hasPending {
		t.Error("Should have pending actions")
	}

	// Test Peek
	peeked, err := queue.Peek()
	if err != nil {
		t.Fatalf("Peek failed: %v", err)
	}
	if peeked == nil {
		t.Fatal("Peek returned nil")
	}
	if peeked.Type != ActionMarkRead {
		t.Errorf("Peeked Type = %s, want %s", peeked.Type, ActionMarkRead)
	}
	if peeked.ResourceID != "email-123" {
		t.Errorf("Peeked ResourceID = %s, want email-123", peeked.ResourceID)
	}

	// Count should still be 1 after peek
	count, _ = queue.Count()
	if count != 1 {
		t.Error("Peek should not remove item from queue")
	}

	// Test List
	actions, err := queue.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(actions) != 1 {
		t.Errorf("List returned %d actions, want 1", len(actions))
	}

	// Test Dequeue
	dequeued, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if dequeued == nil {
		t.Fatal("Dequeue returned nil")
	}
	if dequeued.Type != ActionMarkRead {
		t.Errorf("Dequeued Type = %s, want %s", dequeued.Type, ActionMarkRead)
	}

	// Queue should be empty now
	count, _ = queue.Count()
	if count != 0 {
		t.Error("Queue should be empty after dequeue")
	}
}

func TestOfflineQueueMultipleActions(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Enqueue multiple actions
	actions := []struct {
		actionType ActionType
		resourceID string
		payload    any
	}{
		{ActionMarkRead, "email-1", MarkReadPayload{EmailID: "email-1", Unread: false}},
		{ActionStar, "email-2", StarPayload{EmailID: "email-2", Starred: true}},
		{ActionArchive, "email-3", nil},
	}

	for _, a := range actions {
		if err := queue.Enqueue(a.actionType, a.resourceID, a.payload); err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
	}

	count, _ := queue.Count()
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}

	// Test RemoveByResourceID
	if err := queue.RemoveByResourceID("email-2"); err != nil {
		t.Fatalf("RemoveByResourceID failed: %v", err)
	}

	count, _ = queue.Count()
	if count != 2 {
		t.Errorf("Count after RemoveByResourceID = %d, want 2", count)
	}

	// Test Clear
	if err := queue.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	count, _ = queue.Count()
	if count != 0 {
		t.Error("Queue should be empty after Clear")
	}
}

func TestSettings(t *testing.T) {
	tmpDir := t.TempDir()

	// Test LoadSettings (creates default if not exists)
	settings, err := LoadSettings(tmpDir)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify defaults
	if !settings.Enabled {
		t.Error("Default Enabled should be true")
	}
	if settings.MaxSizeMB != 500 {
		t.Errorf("Default MaxSizeMB = %d, want 500", settings.MaxSizeMB)
	}
	if settings.TTLDays != 30 {
		t.Errorf("Default TTLDays = %d, want 30", settings.TTLDays)
	}
	if settings.Theme != "dark" {
		t.Errorf("Default Theme = %s, want 'dark'", settings.Theme)
	}

	// Test Update
	err = settings.Update(func(s *Settings) {
		s.MaxSizeMB = 1000
		s.Theme = "light"
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	if settings.MaxSizeMB != 1000 {
		t.Errorf("Updated MaxSizeMB = %d, want 1000", settings.MaxSizeMB)
	}
	if settings.Theme != "light" {
		t.Errorf("Updated Theme = %s, want 'light'", settings.Theme)
	}

	// Test Get (returns copy)
	copy := settings.Get()
	if copy.MaxSizeMB != 1000 {
		t.Errorf("Get().MaxSizeMB = %d, want 1000", copy.MaxSizeMB)
	}

	// Test SetEnabled
	if err := settings.SetEnabled(false); err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}
	if settings.Enabled {
		t.Error("Enabled should be false")
	}

	// Test SetMaxSize
	if err := settings.SetMaxSize(2000); err != nil {
		t.Fatalf("SetMaxSize failed: %v", err)
	}
	if settings.MaxSizeMB != 2000 {
		t.Errorf("MaxSizeMB = %d, want 2000", settings.MaxSizeMB)
	}

	// Test SetMaxSize minimum
	if err := settings.SetMaxSize(10); err != nil {
		t.Fatalf("SetMaxSize (min) failed: %v", err)
	}
	if settings.MaxSizeMB != 50 {
		t.Errorf("MaxSizeMB should be clamped to minimum 50, got %d", settings.MaxSizeMB)
	}

	// Test SetTheme
	if err := settings.SetTheme("system"); err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}
	if settings.Theme != "system" {
		t.Errorf("Theme = %s, want 'system'", settings.Theme)
	}

	// Test SetTheme invalid (should default to dark)
	if err := settings.SetTheme("invalid"); err != nil {
		t.Fatalf("SetTheme (invalid) failed: %v", err)
	}
	if settings.Theme != "dark" {
		t.Errorf("Invalid theme should default to 'dark', got %s", settings.Theme)
	}

	// Test Reset
	if err := settings.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if settings.MaxSizeMB != 500 {
		t.Errorf("Reset MaxSizeMB = %d, want 500", settings.MaxSizeMB)
	}
	if settings.Theme != "dark" {
		t.Errorf("Reset Theme = %s, want 'dark'", settings.Theme)
	}
}

func TestSettingsValidate(t *testing.T) {
	tmpDir := t.TempDir()
	settings, _ := LoadSettings(tmpDir)

	// Valid settings
	if err := settings.Validate(); err != nil {
		t.Errorf("Valid settings should not return error: %v", err)
	}

	// Test invalid MaxSizeMB
	settings.MaxSizeMB = 10
	if err := settings.Validate(); err == nil {
		t.Error("MaxSizeMB < 50 should fail validation")
	}
	settings.MaxSizeMB = 500

	// Test invalid TTLDays
	settings.TTLDays = 0
	if err := settings.Validate(); err == nil {
		t.Error("TTLDays < 1 should fail validation")
	}
	settings.TTLDays = 30

	// Test invalid SyncIntervalMinutes
	settings.SyncIntervalMinutes = 0
	if err := settings.Validate(); err == nil {
		t.Error("SyncIntervalMinutes < 1 should fail validation")
	}
}

func TestSettingsHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	settings, _ := LoadSettings(tmpDir)

	// Test GetSyncInterval
	settings.SyncIntervalMinutes = 5
	interval := settings.GetSyncInterval()
	if interval != 5*time.Minute {
		t.Errorf("GetSyncInterval = %v, want 5m", interval)
	}

	// Test GetTTL
	settings.TTLDays = 30
	ttl := settings.GetTTL()
	if ttl != 30*24*time.Hour {
		t.Errorf("GetTTL = %v, want 720h", ttl)
	}

	// Test GetMaxSizeBytes
	settings.MaxSizeMB = 500
	maxBytes := settings.GetMaxSizeBytes()
	if maxBytes != 500*1024*1024 {
		t.Errorf("GetMaxSizeBytes = %d, want %d", maxBytes, 500*1024*1024)
	}

	// Test IsEncryptionEnabled
	settings.EncryptionEnabled = true
	if !settings.IsEncryptionEnabled() {
		t.Error("IsEncryptionEnabled should be true")
	}

	// Test IsCacheEnabled
	settings.Enabled = true
	if !settings.IsCacheEnabled() {
		t.Error("IsCacheEnabled should be true")
	}
}

func TestUnifiedSearch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	// Add test data
	emailStore := NewEmailStore(db)
	emails := []*CachedEmail{
		{ID: "e1", Subject: "Meeting notes", FromName: "Alice", FromEmail: "alice@test.com", Date: time.Now()},
		{ID: "e2", Subject: "Project update", FromName: "Bob", FromEmail: "bob@test.com", Date: time.Now()},
	}
	if err := emailStore.PutBatch(emails); err != nil {
		t.Fatalf("Put emails failed: %v", err)
	}

	eventStore := NewEventStore(db)
	now := time.Now()
	events := []*CachedEvent{
		{ID: "ev1", Title: "Team Meeting", StartTime: now, EndTime: now.Add(time.Hour)},
		{ID: "ev2", Title: "Project Review", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)},
	}
	if err := eventStore.PutBatch(events); err != nil {
		t.Fatalf("Put events failed: %v", err)
	}

	contactStore := NewContactStore(db)
	contacts := []*CachedContact{
		{ID: "c1", DisplayName: "Meeting Coordinator", Email: "coord@test.com"},
		{ID: "c2", DisplayName: "Project Manager", Email: "pm@test.com"},
	}
	if err := contactStore.PutBatch(contacts); err != nil {
		t.Fatalf("Put contacts failed: %v", err)
	}

	// Search for "Meeting"
	results, err := UnifiedSearch(db, "Meeting", 20)
	if err != nil {
		t.Fatalf("UnifiedSearch failed: %v", err)
	}

	// Should find: 1 email + 1 event + 1 contact = 3 results
	if len(results) != 3 {
		t.Errorf("UnifiedSearch returned %d results, want 3", len(results))
	}

	// Verify result types
	types := make(map[string]int)
	for _, r := range results {
		types[r.Type]++
	}

	if types["email"] != 1 {
		t.Errorf("Expected 1 email result, got %d", types["email"])
	}
	if types["event"] != 1 {
		t.Errorf("Expected 1 event result, got %d", types["event"])
	}
	if types["contact"] != 1 {
		t.Errorf("Expected 1 contact result, got %d", types["contact"])
	}

	// Search for "Project"
	results, err = UnifiedSearch(db, "Project", 20)
	if err != nil {
		t.Fatalf("UnifiedSearch (Project) failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("UnifiedSearch (Project) returned %d results, want 3", len(results))
	}
}

func TestManagerGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	email := "test@example.com"
	db, err := mgr.GetDB(email)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	// Add some test data
	emailStore := NewEmailStore(db)
	emails := []*CachedEmail{
		{ID: "1", Subject: "Email 1", Date: time.Now()},
		{ID: "2", Subject: "Email 2", Date: time.Now()},
	}
	if err := emailStore.PutBatch(emails); err != nil {
		t.Fatalf("Put emails failed: %v", err)
	}

	eventStore := NewEventStore(db)
	now := time.Now()
	events := []*CachedEvent{
		{ID: "1", Title: "Event 1", StartTime: now, EndTime: now.Add(time.Hour)},
	}
	if err := eventStore.PutBatch(events); err != nil {
		t.Fatalf("Put events failed: %v", err)
	}

	// Get stats
	stats, err := mgr.GetStats(email)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.EmailCount != 2 {
		t.Errorf("EmailCount = %d, want 2", stats.EmailCount)
	}
	if stats.EventCount != 1 {
		t.Errorf("EventCount = %d, want 1", stats.EventCount)
	}
	if stats.SizeBytes == 0 {
		t.Error("SizeBytes should be > 0")
	}
}

// ================================
// CALENDAR STORE TESTS
// ================================

func TestCalendarStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewCalendarStore(db)

	// Test Put and Get
	calendar := &CachedCalendar{
		ID:          "cal-123",
		Name:        "Work Calendar",
		Description: "My work calendar",
		IsPrimary:   true,
		ReadOnly:    false,
		HexColor:    "#4285f4",
	}

	if err := store.Put(calendar); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("cal-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != calendar.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, calendar.ID)
	}
	if retrieved.Name != calendar.Name {
		t.Errorf("Retrieved Name = %s, want %s", retrieved.Name, calendar.Name)
	}
	if retrieved.IsPrimary != calendar.IsPrimary {
		t.Errorf("Retrieved IsPrimary = %v, want %v", retrieved.IsPrimary, calendar.IsPrimary)
	}
	if retrieved.ReadOnly != calendar.ReadOnly {
		t.Errorf("Retrieved ReadOnly = %v, want %v", retrieved.ReadOnly, calendar.ReadOnly)
	}

	// Test GetPrimary
	primary, err := store.GetPrimary()
	if err != nil {
		t.Fatalf("GetPrimary failed: %v", err)
	}
	if primary.ID != "cal-123" {
		t.Errorf("GetPrimary ID = %s, want cal-123", primary.ID)
	}

	// Test Count
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test Delete
	if err := store.Delete("cal-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 0 {
		t.Errorf("Count after Delete = %d, want 0", count)
	}
}

func TestCalendarStorePutBatch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewCalendarStore(db)

	calendars := []*CachedCalendar{
		{ID: "cal-1", Name: "Primary", IsPrimary: true, ReadOnly: false},
		{ID: "cal-2", Name: "Holidays", IsPrimary: false, ReadOnly: true},
		{ID: "cal-3", Name: "Personal", IsPrimary: false, ReadOnly: false},
	}

	if err := store.PutBatch(calendars); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	count, _ := store.Count()
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestCalendarStoreList(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store := NewCalendarStore(db)

	calendars := []*CachedCalendar{
		{ID: "cal-1", Name: "Primary", IsPrimary: true, ReadOnly: false},
		{ID: "cal-2", Name: "Holidays", IsPrimary: false, ReadOnly: true},
		{ID: "cal-3", Name: "Personal", IsPrimary: false, ReadOnly: false},
	}

	if err := store.PutBatch(calendars); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Test List
	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List returned %d calendars, want 3", len(list))
	}

	// Primary should be first (sorted by is_primary DESC, name ASC)
	if list[0].ID != "cal-1" {
		t.Errorf("First calendar should be primary, got %s", list[0].ID)
	}

	// Test ListWritable (excludes read-only)
	writable, err := store.ListWritable()
	if err != nil {
		t.Fatalf("ListWritable failed: %v", err)
	}
	if len(writable) != 2 {
		t.Errorf("ListWritable returned %d calendars, want 2", len(writable))
	}

	// Verify no read-only calendars
	for _, cal := range writable {
		if cal.ReadOnly {
			t.Errorf("ListWritable returned read-only calendar: %s", cal.ID)
		}
	}
}

// ================================
// ATTACHMENT STORE TESTS
// ================================

func TestAttachmentStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100) // 100MB max
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Test Put
	attachment := &CachedAttachment{
		ID:          "att-123",
		EmailID:     "email-456",
		Filename:    "document.pdf",
		ContentType: "application/pdf",
	}

	content := strings.NewReader("This is test content for the attachment")
	if err := store.Put(attachment, content); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify attachment has hash and path
	if attachment.Hash == "" {
		t.Error("Hash should be set after Put")
	}
	if attachment.LocalPath == "" {
		t.Error("LocalPath should be set after Put")
	}
	if attachment.Size == 0 {
		t.Error("Size should be set after Put")
	}

	// Test Get
	retrieved, err := store.Get("att-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != attachment.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, attachment.ID)
	}
	if retrieved.Filename != attachment.Filename {
		t.Errorf("Retrieved Filename = %s, want %s", retrieved.Filename, attachment.Filename)
	}
	if retrieved.Hash != attachment.Hash {
		t.Errorf("Retrieved Hash = %s, want %s", retrieved.Hash, attachment.Hash)
	}

	// Test GetByHash
	byHash, err := store.GetByHash(attachment.Hash)
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	if byHash.ID != attachment.ID {
		t.Errorf("GetByHash ID = %s, want %s", byHash.ID, attachment.ID)
	}

	// Test Count
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Test TotalSize
	totalSize, err := store.TotalSize()
	if err != nil {
		t.Fatalf("TotalSize failed: %v", err)
	}
	if totalSize != attachment.Size {
		t.Errorf("TotalSize = %d, want %d", totalSize, attachment.Size)
	}

	// Test Delete
	if err := store.Delete("att-123"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count()
	if count != 0 {
		t.Errorf("Count after Delete = %d, want 0", count)
	}
}

func TestAttachmentStoreListByEmail(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100)
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add attachments for different emails
	att1 := &CachedAttachment{ID: "att-1", EmailID: "email-1", Filename: "file1.txt", ContentType: "text/plain"}
	att2 := &CachedAttachment{ID: "att-2", EmailID: "email-1", Filename: "file2.txt", ContentType: "text/plain"}
	att3 := &CachedAttachment{ID: "att-3", EmailID: "email-2", Filename: "file3.txt", ContentType: "text/plain"}

	_ = store.Put(att1, strings.NewReader("content 1"))
	_ = store.Put(att2, strings.NewReader("content 2"))
	_ = store.Put(att3, strings.NewReader("content 3"))

	// List attachments for email-1
	list, err := store.ListByEmail("email-1")
	if err != nil {
		t.Fatalf("ListByEmail failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListByEmail returned %d attachments, want 2", len(list))
	}

	// Test DeleteByEmail
	if err := store.DeleteByEmail("email-1"); err != nil {
		t.Fatalf("DeleteByEmail failed: %v", err)
	}

	list, _ = store.ListByEmail("email-1")
	if len(list) != 0 {
		t.Errorf("ListByEmail after DeleteByEmail = %d, want 0", len(list))
	}

	// email-2 should still have its attachment
	count, _ := store.Count()
	if count != 1 {
		t.Errorf("Count after DeleteByEmail = %d, want 1", count)
	}
}

func TestAttachmentStoreOpen(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100)
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	testContent := "This is the file content to read back"
	attachment := &CachedAttachment{
		ID:          "att-open",
		EmailID:     "email-1",
		Filename:    "test.txt",
		ContentType: "text/plain",
	}
	_ = store.Put(attachment, strings.NewReader(testContent))

	// Test Open
	file, err := store.Open("att-open")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Content = %q, want %q", string(content), testContent)
	}
}

func TestAttachmentStoreStats(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100) // 100MB
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add some attachments
	att1 := &CachedAttachment{ID: "att-1", EmailID: "email-1", Filename: "file1.txt"}
	att2 := &CachedAttachment{ID: "att-2", EmailID: "email-2", Filename: "file2.txt"}
	_ = store.Put(att1, strings.NewReader("content for file 1"))
	_ = store.Put(att2, strings.NewReader("content for file 2"))

	// Get stats
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Count != 2 {
		t.Errorf("Stats.Count = %d, want 2", stats.Count)
	}
	if stats.TotalSize == 0 {
		t.Error("Stats.TotalSize should be > 0")
	}
	if stats.MaxSize != 100*1024*1024 {
		t.Errorf("Stats.MaxSize = %d, want %d", stats.MaxSize, 100*1024*1024)
	}
	if stats.Usage <= 0 {
		t.Error("Stats.Usage should be > 0")
	}
}

func TestAttachmentStorePrune(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	// Create store with very small max size (1 byte)
	store, err := NewAttachmentStore(db, tmpDir, 1) // 1MB = 1048576 bytes
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add attachments that exceed the limit
	// Each attachment is small, well under 1MB, so no pruning needed
	att1 := &CachedAttachment{ID: "att-1", EmailID: "email-1", Filename: "file1.txt"}
	_ = store.Put(att1, strings.NewReader("small content"))

	// Prune should not remove anything since we're under limit
	pruned, err := store.Prune()
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}
	if pruned != 0 {
		t.Errorf("Prune should remove 0 attachments when under limit, removed %d", pruned)
	}
}

func TestAttachmentStoreLRUEvict(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100)
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add multiple attachments
	for i := 0; i < 5; i++ {
		att := &CachedAttachment{
			ID:       fmt.Sprintf("att-%d", i),
			EmailID:  "email-1",
			Filename: fmt.Sprintf("file%d.txt", i),
		}
		_ = store.Put(att, strings.NewReader(fmt.Sprintf("content %d", i)))
		time.Sleep(10 * time.Millisecond) // Ensure different access times
	}

	count, _ := store.Count()
	if count != 5 {
		t.Fatalf("Should have 5 attachments, got %d", count)
	}

	// Evict some attachments
	evicted, err := store.LRUEvict(100) // Evict to free 100 bytes
	if err != nil {
		t.Fatalf("LRUEvict failed: %v", err)
	}

	// Should have evicted at least 1
	if evicted < 1 {
		t.Errorf("LRUEvict should evict at least 1 attachment, evicted %d", evicted)
	}

	newCount, _ := store.Count()
	if newCount >= count {
		t.Errorf("Count after eviction = %d, should be less than %d", newCount, count)
	}
}

func TestAttachmentStoreDeduplication(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100)
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add two attachments with the same content
	sameContent := "identical content for deduplication test"

	att1 := &CachedAttachment{ID: "att-1", EmailID: "email-1", Filename: "file1.txt"}
	_ = store.Put(att1, strings.NewReader(sameContent))

	att2 := &CachedAttachment{ID: "att-2", EmailID: "email-2", Filename: "file2.txt"}
	_ = store.Put(att2, strings.NewReader(sameContent))

	// Both should have the same hash
	if att1.Hash != att2.Hash {
		t.Error("Identical content should produce same hash")
	}

	// Both should point to the same file
	if att1.LocalPath != att2.LocalPath {
		t.Error("Identical content should use same local path")
	}

	// Delete one - file should remain because other uses it
	if err := store.Delete("att-1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// att-2 should still be accessible
	retrieved, err := store.Get("att-2")
	if err != nil {
		t.Fatalf("Get att-2 after deleting att-1 failed: %v", err)
	}
	if retrieved.ID != "att-2" {
		t.Errorf("Retrieved ID = %s, want att-2", retrieved.ID)
	}
}

func TestByAccessTime(t *testing.T) {
	now := time.Now()
	attachments := byAccessTime{
		{ID: "3", AccessedAt: now.Add(2 * time.Hour)},
		{ID: "1", AccessedAt: now},
		{ID: "2", AccessedAt: now.Add(1 * time.Hour)},
	}

	sort.Sort(attachments)

	if attachments[0].ID != "1" {
		t.Errorf("First should be oldest, got %s", attachments[0].ID)
	}
	if attachments[2].ID != "3" {
		t.Errorf("Last should be newest, got %s", attachments[2].ID)
	}
}

// ================================
// SEARCH QUERY PARSING TESTS
// ================================

func TestParseSearchQuery_Basic(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("hello world")

	if query.Text != "hello world" {
		t.Errorf("expected Text 'hello world', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_FromOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("from:john@example.com important email")

	if query.From != "john@example.com" {
		t.Errorf("expected From 'john@example.com', got '%s'", query.From)
	}
	if query.Text != "important email" {
		t.Errorf("expected Text 'important email', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_ToOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("to:recipient@test.com")

	if query.To != "recipient@test.com" {
		t.Errorf("expected To 'recipient@test.com', got '%s'", query.To)
	}
}

func TestParseSearchQuery_SubjectOperator(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("subject:meeting notes")

	if query.Subject != "meeting" {
		t.Errorf("expected Subject 'meeting', got '%s'", query.Subject)
	}
}

func TestParseSearchQuery_SubjectWithQuotes(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery(`subject:"meeting notes" from:john@test.com`)

	if query.Subject != "meeting notes" {
		t.Errorf("expected Subject 'meeting notes', got '%s'", query.Subject)
	}
	if query.From != "john@test.com" {
		t.Errorf("expected From 'john@test.com', got '%s'", query.From)
	}
}

func TestParseSearchQuery_HasAttachment(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("has:attachment")

	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
}

func TestParseSearchQuery_HasAttachments(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("has:attachments")

	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true for 'attachments'")
	}
}

func TestParseSearchQuery_IsUnread(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:unread")

	if query.IsUnread == nil || !*query.IsUnread {
		t.Error("expected IsUnread to be true")
	}
}

func TestParseSearchQuery_IsRead(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:read")

	if query.IsUnread == nil || *query.IsUnread {
		t.Error("expected IsUnread to be false (read)")
	}
}

func TestParseSearchQuery_IsStarred(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("is:starred")

	if query.IsStarred == nil || !*query.IsStarred {
		t.Error("expected IsStarred to be true")
	}
}

func TestParseSearchQuery_InFolder(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("in:INBOX")

	if query.In != "INBOX" {
		t.Errorf("expected In 'INBOX', got '%s'", query.In)
	}
}

func TestParseSearchQuery_DateAfter(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:2024-01-15")

	if query.After.IsZero() {
		t.Error("expected After date to be set")
	}
	if query.After.Year() != 2024 || query.After.Month() != 1 || query.After.Day() != 15 {
		t.Errorf("expected After date 2024-01-15, got %v", query.After)
	}
}

func TestParseSearchQuery_DateBefore(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("before:2024-12-31")

	if query.Before.IsZero() {
		t.Error("expected Before date to be set")
	}
	if query.Before.Year() != 2024 || query.Before.Month() != 12 || query.Before.Day() != 31 {
		t.Errorf("expected Before date 2024-12-31, got %v", query.Before)
	}
}

func TestParseSearchQuery_RelativeDateToday(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:today")

	if query.After.IsZero() {
		t.Error("expected After date to be set for 'today'")
	}
	// Should be today's date at midnight
	now := time.Now()
	if query.After.Year() != now.Year() || query.After.Month() != now.Month() || query.After.Day() != now.Day() {
		t.Errorf("expected After date to be today, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateYesterday(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:yesterday")

	if query.After.IsZero() {
		t.Error("expected After date to be set for 'yesterday'")
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	if query.After.Day() != yesterday.Day() {
		t.Errorf("expected After date to be yesterday, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateDays(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:7d")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '7d'")
	}
	expected := time.Now().AddDate(0, 0, -7)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~7 days ago, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateWeeks(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:2w")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '2w'")
	}
	expected := time.Now().AddDate(0, 0, -14)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~14 days ago, got %v", query.After)
	}
}

func TestParseSearchQuery_RelativeDateMonths(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("after:3m")

	if query.After.IsZero() {
		t.Error("expected After date to be set for '3m'")
	}
	expected := time.Now().AddDate(0, -3, 0)
	diff := query.After.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected After date ~3 months ago, got %v", query.After)
	}
}

func TestParseSearchQuery_MultipleOperators(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("from:sender@test.com to:recipient@test.com is:unread has:attachment important")

	if query.From != "sender@test.com" {
		t.Errorf("expected From 'sender@test.com', got '%s'", query.From)
	}
	if query.To != "recipient@test.com" {
		t.Errorf("expected To 'recipient@test.com', got '%s'", query.To)
	}
	if query.IsUnread == nil || !*query.IsUnread {
		t.Error("expected IsUnread to be true")
	}
	if query.HasAttachment == nil || !*query.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
	if query.Text != "important" {
		t.Errorf("expected Text 'important', got '%s'", query.Text)
	}
}

func TestParseSearchQuery_EmptyQuery(t *testing.T) {
	t.Parallel()

	query := ParseSearchQuery("")

	if query.Text != "" {
		t.Errorf("expected empty Text, got '%s'", query.Text)
	}
	if query.From != "" {
		t.Errorf("expected empty From, got '%s'", query.From)
	}
}

func TestParseSearchQuery_DateFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string // YYYY-MM-DD
	}{
		{"ISO format", "after:2024-06-15", "2024-06-15"},
		{"Slash format", "after:2024/06/15", "2024-06-15"},
		{"US format", "after:06/15/2024", "2024-06-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseSearchQuery(tt.input)
			if query.After.IsZero() {
				t.Errorf("expected After date to be set for %s", tt.input)
				return
			}
			got := query.After.Format("2006-01-02")
			if got != tt.expected {
				t.Errorf("expected date %s, got %s", tt.expected, got)
			}
		})
	}
}

// ================================
// PHOTO STORE TESTS
// ================================

func TestPhotoStore_PutAndGet(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Put a photo
	photoData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
	err = store.Put("contact-123", "image/png", photoData)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get the photo
	data, contentType, err := store.Get("contact-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if contentType != "image/png" {
		t.Errorf("expected content type 'image/png', got '%s'", contentType)
	}

	if len(data) != len(photoData) {
		t.Errorf("expected data length %d, got %d", len(photoData), len(data))
	}
}

func TestPhotoStore_GetNonExistent(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	data, contentType, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if data != nil || contentType != "" {
		t.Error("expected nil data and empty content type for nonexistent photo")
	}
}

func TestPhotoStore_Delete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Put a photo
	photoData := []byte{0x89, 0x50, 0x4E, 0x47}
	err = store.Put("contact-123", "image/png", photoData)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Delete it
	err = store.Delete("contact-123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should not be retrievable
	data, _, err := store.Get("contact-123")
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if data != nil {
		t.Error("expected nil data after delete")
	}
}

func TestPhotoStore_IsValid(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Should return false for nonexistent
	if store.IsValid("nonexistent") {
		t.Error("expected IsValid to return false for nonexistent photo")
	}

	// Put a photo
	err = store.Put("contact-123", "image/png", []byte{0x89, 0x50})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Should return true for existing
	if !store.IsValid("contact-123") {
		t.Error("expected IsValid to return true for existing photo")
	}
}

func TestPhotoStore_CountAndTotalSize(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Initially empty
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	size, err := store.TotalSize()
	if err != nil {
		t.Fatalf("TotalSize failed: %v", err)
	}
	if size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	// Add photos
	_ = store.Put("contact-1", "image/png", []byte{1, 2, 3, 4, 5})
	_ = store.Put("contact-2", "image/jpeg", []byte{1, 2, 3})

	count, _ = store.Count()
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	size, _ = store.TotalSize()
	if size != 8 { // 5 + 3
		t.Errorf("expected size 8, got %d", size)
	}
}

func TestPhotoStore_Prune(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Add a photo normally
	_ = store.Put("contact-keep", "image/png", []byte{1, 2, 3})

	// Manually insert an "old" photo directly into the database
	// with a cached_at timestamp from 1 year ago
	oldTime := time.Now().Add(-365 * 24 * time.Hour).Unix()
	_, err = db.Exec(`INSERT INTO photos (contact_id, content_type, local_path, size, cached_at, accessed_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"contact-expired", "image/png", tmpDir+"/photos/old-photo", 100, oldTime, oldTime)
	if err != nil {
		t.Fatalf("insert old photo failed: %v", err)
	}

	// Prune should remove the old one but keep the recent one
	pruned, err := store.Prune()
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 photo pruned, got %d", pruned)
	}

	// Old photo should no longer exist in DB
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM photos WHERE contact_id = ?", "contact-expired").Scan(&count)
	if count != 0 {
		t.Error("expected expired photo to be pruned from database")
	}

	// Recent photo should still exist
	data, _, _ := store.Get("contact-keep")
	if data == nil {
		t.Error("expected recent photo to still exist")
	}
}

func TestPhotoStore_GetStats(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Add photos
	_ = store.Put("contact-1", "image/png", []byte{1, 2, 3, 4, 5})
	_ = store.Put("contact-2", "image/jpeg", []byte{1, 2, 3})

	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Count != 2 {
		t.Errorf("expected Count 2, got %d", stats.Count)
	}
	if stats.TotalSize != 8 {
		t.Errorf("expected TotalSize 8, got %d", stats.TotalSize)
	}
	if stats.TTLDays != 30 {
		t.Errorf("expected TTLDays 30, got %d", stats.TTLDays)
	}
}

func TestPhotoStore_RemoveOrphaned(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	tmpDir := t.TempDir()
	store, err := NewPhotoStore(db, tmpDir, DefaultPhotoTTL)
	if err != nil {
		t.Fatalf("NewPhotoStore failed: %v", err)
	}

	// Add a photo through the store
	_ = store.Put("contact-known", "image/png", []byte{1, 2, 3})

	// Create an orphaned file directly
	orphanPath := tmpDir + "/photos/orphan-file"
	_ = os.WriteFile(orphanPath, []byte{4, 5, 6}, 0600)

	// Remove orphaned files
	removed, err := store.RemoveOrphaned()
	if err != nil {
		t.Fatalf("RemoveOrphaned failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 orphan removed, got %d", removed)
	}

	// Known photo should still exist
	data, _, _ := store.Get("contact-known")
	if data == nil {
		t.Error("known photo should still exist")
	}
}

// ================================
// ENCRYPTION HELPER TESTS
// ================================

func TestTableNames(t *testing.T) {
	t.Parallel()

	names := tableNames()

	if len(names) == 0 {
		t.Error("expected non-empty table names")
	}

	// Check that expected tables are present
	expectedTables := map[string]bool{
		"emails":   false,
		"events":   false,
		"contacts": false,
		"folders":  false,
	}

	for _, name := range names {
		if _, ok := expectedTables[name]; ok {
			expectedTables[name] = true
		}
	}

	for table, found := range expectedTables {
		if !found {
			t.Errorf("expected table '%s' in tableNames()", table)
		}
	}
}

func TestAllowedTables(t *testing.T) {
	t.Parallel()

	// Test that the allowedTables map is populated
	if len(allowedTables) == 0 {
		t.Error("allowedTables should not be empty")
	}

	// Check specific tables
	if !allowedTables["emails"] {
		t.Error("emails should be in allowedTables")
	}
	if !allowedTables["events"] {
		t.Error("events should be in allowedTables")
	}
	if !allowedTables["contacts"] {
		t.Error("contacts should be in allowedTables")
	}

	// Check that random strings are not allowed
	if allowedTables["malicious_table"] {
		t.Error("malicious_table should not be in allowedTables")
	}
	if allowedTables["DROP TABLE emails"] {
		t.Error("SQL injection attempt should not be in allowedTables")
	}
}

func TestEncryptionConfig_Struct(t *testing.T) {
	t.Parallel()

	config := EncryptionConfig{
		Enabled: true,
		KeyID:   "user@example.com",
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.KeyID != "user@example.com" {
		t.Errorf("expected KeyID 'user@example.com', got '%s'", config.KeyID)
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	if config.BasePath == "" {
		t.Error("expected BasePath to be set")
	}
	if config.TTLDays <= 0 {
		t.Error("expected TTLDays to be positive")
	}
	if config.MaxSizeMB <= 0 {
		t.Error("expected MaxSizeMB to be positive")
	}
	if config.SyncIntervalMinutes <= 0 {
		t.Error("expected SyncIntervalMinutes to be positive")
	}
}

// ================================
// ADDITIONAL EMAIL STORE TESTS
// ================================

func TestEmailStoreList(t *testing.T) {
	db := setupTestDB(t)
	store := NewEmailStore(db)

	// Add test emails
	now := time.Now()
	emails := []*CachedEmail{
		{ID: "1", FolderID: "inbox", Subject: "Email 1", Unread: true, Starred: false, Date: now.Add(-3 * time.Hour)},
		{ID: "2", FolderID: "inbox", Subject: "Email 2", Unread: false, Starred: true, Date: now.Add(-2 * time.Hour)},
		{ID: "3", FolderID: "sent", Subject: "Email 3", Unread: false, Starred: false, Date: now.Add(-1 * time.Hour)},
		{ID: "4", ThreadID: "thread-1", FolderID: "inbox", Subject: "Email 4", Unread: true, Starred: false, Date: now},
	}

	if err := store.PutBatch(emails); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Test List with FolderID filter
	list, err := store.List(ListOptions{FolderID: "inbox", Limit: 10})
	if err != nil {
		t.Fatalf("List (folder) failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List (folder) returned %d, want 3", len(list))
	}

	// Test List with UnreadOnly filter
	list, err = store.List(ListOptions{UnreadOnly: true, Limit: 10})
	if err != nil {
		t.Fatalf("List (unread) failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List (unread) returned %d, want 2", len(list))
	}

	// Test List with StarredOnly filter
	list, err = store.List(ListOptions{StarredOnly: true, Limit: 10})
	if err != nil {
		t.Fatalf("List (starred) failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List (starred) returned %d, want 1", len(list))
	}

	// Test List with ThreadID filter
	list, err = store.List(ListOptions{ThreadID: "thread-1", Limit: 10})
	if err != nil {
		t.Fatalf("List (thread) failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List (thread) returned %d, want 1", len(list))
	}

	// Test List with Since/Before filters
	// Since 2.5 hours ago should include emails at -2h, -1h, and now (3 total)
	list, err = store.List(ListOptions{Since: now.Add(-2*time.Hour - 30*time.Minute), Limit: 10})
	if err != nil {
		t.Fatalf("List (since) failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List (since) returned %d, want 3", len(list))
	}

	// Test List with Offset
	list, err = store.List(ListOptions{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("List (offset) failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List (offset) returned %d, want 2", len(list))
	}
}

func TestEmailStoreCountUnread(t *testing.T) {
	db := setupTestDB(t)
	store := NewEmailStore(db)

	// Add emails with different read status
	emails := []*CachedEmail{
		{ID: "1", Subject: "Unread 1", Unread: true, Date: time.Now()},
		{ID: "2", Subject: "Unread 2", Unread: true, Date: time.Now()},
		{ID: "3", Subject: "Read 1", Unread: false, Date: time.Now()},
	}

	if err := store.PutBatch(emails); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	count, err := store.CountUnread()
	if err != nil {
		t.Fatalf("CountUnread failed: %v", err)
	}
	if count != 2 {
		t.Errorf("CountUnread = %d, want 2", count)
	}
}

// ================================
// ADDITIONAL CONTACT STORE TESTS
// ================================

func TestContactStoreGetByEmail(t *testing.T) {
	db := setupTestDB(t)
	store := NewContactStore(db)

	// Add test contacts
	contacts := []*CachedContact{
		{ID: "1", DisplayName: "Alice", Email: "alice@example.com"},
		{ID: "2", DisplayName: "Bob", Email: "bob@example.com"},
	}

	if err := store.PutBatch(contacts); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Test GetByEmail
	contact, err := store.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if contact == nil {
		t.Fatal("GetByEmail returned nil")
	}
	if contact.DisplayName != "Alice" {
		t.Errorf("DisplayName = %s, want Alice", contact.DisplayName)
	}

	// Test GetByEmail - not found (returns sql.ErrNoRows)
	notFound, err := store.GetByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("GetByEmail should return error for non-existent email")
	}
	if notFound != nil {
		t.Error("GetByEmail should return nil for non-existent email")
	}
}

func TestContactStoreListGroups(t *testing.T) {
	db := setupTestDB(t)
	store := NewContactStore(db)

	// Add contacts with groups
	contacts := []*CachedContact{
		{ID: "1", DisplayName: "Alice", Email: "alice@example.com", Groups: []string{"Work", "Friends"}},
		{ID: "2", DisplayName: "Bob", Email: "bob@example.com", Groups: []string{"Work"}},
		{ID: "3", DisplayName: "Charlie", Email: "charlie@example.com", Groups: []string{"Family"}},
		{ID: "4", DisplayName: "Dave", Email: "dave@example.com", Groups: []string{}},
	}

	if err := store.PutBatch(contacts); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Test ListGroups
	groups, err := store.ListGroups()
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("ListGroups returned %d groups, want 3", len(groups))
	}

	// Verify groups
	groupSet := make(map[string]bool)
	for _, g := range groups {
		groupSet[g] = true
	}
	if !groupSet["Work"] {
		t.Error("Work group not found")
	}
	if !groupSet["Friends"] {
		t.Error("Friends group not found")
	}
	if !groupSet["Family"] {
		t.Error("Family group not found")
	}
}

// ================================
// ADDITIONAL EVENT STORE TESTS
// ================================

func TestEventStoreDeleteByCalendar(t *testing.T) {
	db := setupTestDB(t)
	store := NewEventStore(db)
	now := time.Now()

	// Add events for different calendars
	events := []*CachedEvent{
		{ID: "1", CalendarID: "cal-1", Title: "Event 1", StartTime: now, EndTime: now.Add(time.Hour)},
		{ID: "2", CalendarID: "cal-1", Title: "Event 2", StartTime: now, EndTime: now.Add(time.Hour)},
		{ID: "3", CalendarID: "cal-2", Title: "Event 3", StartTime: now, EndTime: now.Add(time.Hour)},
	}

	if err := store.PutBatch(events); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Delete by calendar
	err := store.DeleteByCalendar("cal-1")
	if err != nil {
		t.Fatalf("DeleteByCalendar failed: %v", err)
	}

	// Verify cal-1 events deleted
	count, _ := store.Count()
	if count != 1 {
		t.Errorf("Count after DeleteByCalendar = %d, want 1", count)
	}

	// Verify cal-2 event remains
	event, err := store.Get("3")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if event.CalendarID != "cal-2" {
		t.Error("Wrong event remaining")
	}
}

func TestEventStoreGetUpcoming(t *testing.T) {
	db := setupTestDB(t)
	store := NewEventStore(db)
	now := time.Now()

	// Add past and future events
	events := []*CachedEvent{
		{ID: "1", CalendarID: "primary", Title: "Past Event", StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Hour)},
		{ID: "2", CalendarID: "primary", Title: "Future Event 1", StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour)},
		{ID: "3", CalendarID: "primary", Title: "Future Event 2", StartTime: now.Add(3 * time.Hour), EndTime: now.Add(4 * time.Hour)},
	}

	if err := store.PutBatch(events); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Get upcoming events
	upcoming, err := store.GetUpcoming(10)
	if err != nil {
		t.Fatalf("GetUpcoming failed: %v", err)
	}

	if len(upcoming) != 2 {
		t.Errorf("GetUpcoming returned %d events, want 2", len(upcoming))
	}

	// Verify only future events
	for _, e := range upcoming {
		if e.StartTime.Before(now) {
			t.Error("GetUpcoming returned past event")
		}
	}
}

// ================================
// ADDITIONAL FOLDER STORE TESTS
// ================================

func TestFolderStorePutBatch(t *testing.T) {
	db := setupTestDB(t)
	store := NewFolderStore(db)

	folders := []*CachedFolder{
		{ID: "1", Name: "INBOX", Type: "inbox", UnreadCount: 10},
		{ID: "2", Name: "Sent", Type: "sent", UnreadCount: 0},
		{ID: "3", Name: "Drafts", Type: "drafts", UnreadCount: 2},
	}

	if err := store.PutBatch(folders); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestFolderStoreGetByType(t *testing.T) {
	db := setupTestDB(t)
	store := NewFolderStore(db)

	folders := []*CachedFolder{
		{ID: "1", Name: "INBOX", Type: "inbox", UnreadCount: 10},
		{ID: "2", Name: "Sent", Type: "sent", UnreadCount: 0},
		{ID: "3", Name: "Trash", Type: "trash", UnreadCount: 5},
	}

	if err := store.PutBatch(folders); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	// Test GetByType
	folder, err := store.GetByType("inbox")
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}
	if folder.Name != "INBOX" {
		t.Errorf("GetByType Name = %s, want INBOX", folder.Name)
	}

	// Test GetByType - not found
	notFoundFolder, err := store.GetByType("nonexistent")
	if err == nil {
		t.Error("GetByType should return error for non-existent type")
	}
	if notFoundFolder != nil {
		t.Error("GetByType should return nil for non-existent type")
	}
}

func TestFolderStoreIncrementUnread(t *testing.T) {
	db := setupTestDB(t)
	store := NewFolderStore(db)

	folder := &CachedFolder{ID: "1", Name: "INBOX", Type: "inbox", UnreadCount: 10}
	if err := store.Put(folder); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Increment
	if err := store.IncrementUnread("1", 5); err != nil {
		t.Fatalf("IncrementUnread failed: %v", err)
	}

	retrieved, _ := store.Get("1")
	if retrieved.UnreadCount != 15 {
		t.Errorf("UnreadCount after increment = %d, want 15", retrieved.UnreadCount)
	}

	// Decrement
	if err := store.IncrementUnread("1", -3); err != nil {
		t.Fatalf("IncrementUnread (decrement) failed: %v", err)
	}

	retrieved, _ = store.Get("1")
	if retrieved.UnreadCount != 12 {
		t.Errorf("UnreadCount after decrement = %d, want 12", retrieved.UnreadCount)
	}
}

func TestFolderStoreCount(t *testing.T) {
	db := setupTestDB(t)
	store := NewFolderStore(db)

	// Initially empty
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}

	// Add folders
	folders := []*CachedFolder{
		{ID: "1", Name: "INBOX", Type: "inbox"},
		{ID: "2", Name: "Sent", Type: "sent"},
	}
	_ = store.PutBatch(folders)

	count, _ = store.Count()
	if count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestFolderStoreGetTotalUnread(t *testing.T) {
	db := setupTestDB(t)
	store := NewFolderStore(db)

	folders := []*CachedFolder{
		{ID: "1", Name: "INBOX", Type: "inbox", UnreadCount: 10},
		{ID: "2", Name: "Work", Type: "label", UnreadCount: 5},
		{ID: "3", Name: "Sent", Type: "sent", UnreadCount: 0},
	}

	if err := store.PutBatch(folders); err != nil {
		t.Fatalf("PutBatch failed: %v", err)
	}

	total, err := store.GetTotalUnread()
	if err != nil {
		t.Fatalf("GetTotalUnread failed: %v", err)
	}
	if total != 15 {
		t.Errorf("GetTotalUnread = %d, want 15", total)
	}
}

// ================================
// ADDITIONAL OFFLINE QUEUE TESTS
// ================================

func TestOfflineQueueMarkFailed(t *testing.T) {
	db := setupTestDB(t)
	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Enqueue an action
	payload := MarkReadPayload{EmailID: "email-1", Unread: false}
	if err := queue.Enqueue(ActionMarkRead, "email-1", payload); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Peek to get the ID
	action, _ := queue.Peek()

	// Mark as failed
	testErr := fmt.Errorf("test error")
	if err := queue.MarkFailed(action.ID, testErr); err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}

	// Verify attempts incremented and error recorded
	actions, _ := queue.List()
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	if actions[0].Attempts != 1 {
		t.Errorf("Attempts = %d, want 1", actions[0].Attempts)
	}
	if actions[0].LastError != "test error" {
		t.Errorf("LastError = %s, want 'test error'", actions[0].LastError)
	}
}

func TestOfflineQueueRemove(t *testing.T) {
	db := setupTestDB(t)
	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Enqueue actions
	_ = queue.Enqueue(ActionMarkRead, "email-1", nil)
	_ = queue.Enqueue(ActionStar, "email-2", nil)

	// Get first action ID
	action, _ := queue.Peek()

	// Remove it
	if err := queue.Remove(action.ID); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify count
	count, _ := queue.Count()
	if count != 1 {
		t.Errorf("Count after Remove = %d, want 1", count)
	}
}

func TestOfflineQueueRemoveStale(t *testing.T) {
	db := setupTestDB(t)
	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Enqueue an action
	_ = queue.Enqueue(ActionMarkRead, "email-1", nil)

	// Remove stale with short max age (should remove nothing since just created)
	removed, err := queue.RemoveStale(time.Hour)
	if err != nil {
		t.Fatalf("RemoveStale failed: %v", err)
	}
	if removed != 0 {
		t.Errorf("RemoveStale removed %d, want 0", removed)
	}

	// Manually insert an old action
	oldTime := time.Now().Add(-24 * time.Hour).Unix()
	_, err = db.Exec(`INSERT INTO offline_queue (type, resource_id, payload, created_at) VALUES (?, ?, ?, ?)`,
		ActionArchive, "old-email", "{}", oldTime)
	if err != nil {
		t.Fatalf("insert old action failed: %v", err)
	}

	// Remove stale with 1 hour max age
	removed, err = queue.RemoveStale(time.Hour)
	if err != nil {
		t.Fatalf("RemoveStale failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("RemoveStale removed %d, want 1", removed)
	}
}

func TestOfflineQueueGetActionData(t *testing.T) {
	db := setupTestDB(t)
	queue, err := NewOfflineQueue(db)
	if err != nil {
		t.Fatalf("NewOfflineQueue failed: %v", err)
	}

	// Enqueue with payload
	payload := MarkReadPayload{EmailID: "email-123", Unread: true}
	_ = queue.Enqueue(ActionMarkRead, "email-123", payload)

	// Get the action
	action, _ := queue.Peek()

	// Parse payload
	var retrieved MarkReadPayload
	if err := action.GetActionData(&retrieved); err != nil {
		t.Fatalf("GetActionData failed: %v", err)
	}

	if retrieved.EmailID != "email-123" {
		t.Errorf("EmailID = %s, want email-123", retrieved.EmailID)
	}
	if !retrieved.Unread {
		t.Error("Unread should be true")
	}
}

// ================================
// ADDITIONAL SYNC STORE TESTS
// ================================

func TestSyncStoreUpdateCursor(t *testing.T) {
	db := setupTestDB(t)
	store := NewSyncStore(db)

	// Set initial state
	state := &SyncState{
		Resource: ResourceEmails,
		LastSync: time.Now().Add(-time.Hour),
		Cursor:   "cursor-1",
	}
	if err := store.Set(state); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Update cursor
	if err := store.UpdateCursor(ResourceEmails, "cursor-2"); err != nil {
		t.Fatalf("UpdateCursor failed: %v", err)
	}

	// Verify
	retrieved, _ := store.Get(ResourceEmails)
	if retrieved.Cursor != "cursor-2" {
		t.Errorf("Cursor = %s, want cursor-2", retrieved.Cursor)
	}
}

func TestSyncStoreMarkSynced(t *testing.T) {
	db := setupTestDB(t)
	store := NewSyncStore(db)

	// MarkSynced for new resource
	if err := store.MarkSynced(ResourceContacts); err != nil {
		t.Fatalf("MarkSynced failed: %v", err)
	}

	// Verify state exists
	state, err := store.Get(ResourceContacts)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if state == nil {
		t.Fatal("State should exist")
	}
	if time.Since(state.LastSync) > time.Second {
		t.Error("LastSync should be recent")
	}

	// MarkSynced again (update existing) - uses Unix second precision
	if err := store.MarkSynced(ResourceContacts); err != nil {
		t.Fatalf("MarkSynced (2nd) failed: %v", err)
	}

	state2, _ := store.Get(ResourceContacts)
	// Verify state2 exists and LastSync is within last second
	if state2 == nil {
		t.Fatal("State should still exist after second MarkSynced")
	}
	if time.Since(state2.LastSync) > time.Second {
		t.Error("LastSync should be recent after update")
	}
}

func TestSyncStoreDelete(t *testing.T) {
	db := setupTestDB(t)
	store := NewSyncStore(db)

	// Set state
	state := &SyncState{
		Resource: ResourceEvents,
		LastSync: time.Now(),
		Cursor:   "cursor-1",
	}
	_ = store.Set(state)

	// Delete
	if err := store.Delete(ResourceEvents); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	retrieved, err := store.Get(ResourceEvents)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved != nil {
		t.Error("State should be deleted")
	}
}

// ================================
// ADDITIONAL ATTACHMENT STORE TESTS
// ================================

func TestAttachmentStoreRemoveOrphaned(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	db, err := mgr.GetDB("test@example.com")
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	store, err := NewAttachmentStore(db, tmpDir, 100)
	if err != nil {
		t.Fatalf("NewAttachmentStore failed: %v", err)
	}

	// Add a tracked attachment
	att := &CachedAttachment{ID: "att-tracked", EmailID: "email-1", Filename: "tracked.txt"}
	_ = store.Put(att, strings.NewReader("tracked content"))

	// Create orphan file directly in attachments directory
	attachDir := filepath.Join(tmpDir, "attachments")
	orphanPath := filepath.Join(attachDir, "orphan-file.txt")
	_ = os.WriteFile(orphanPath, []byte("orphan content"), 0600)

	// Remove orphaned
	removed, err := store.RemoveOrphaned()
	if err != nil {
		t.Fatalf("RemoveOrphaned failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("RemoveOrphaned removed %d, want 1", removed)
	}

	// Verify orphan file is deleted
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Error("Orphan file should be deleted")
	}

	// Verify tracked attachment still exists
	retrieved, err := store.Get("att-tracked")
	if err != nil {
		t.Fatalf("Get tracked failed: %v", err)
	}
	if retrieved == nil {
		t.Error("Tracked attachment should still exist")
	}
}

// ================================
// ADDITIONAL SETTINGS TESTS
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
	defer mgr.Close()

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
