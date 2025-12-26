package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

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
