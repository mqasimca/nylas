package cache

import (
	"testing"
	"time"
)

// ================================
// EMAIL STORE TESTS
// ================================

func TestEmailStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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

// ================================
// SYNC STORE TESTS
// ================================

func TestSyncStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

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

// ================================
// EVENT STORE TESTS
// ================================

func TestEventStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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

// ================================
// CONTACT STORE TESTS
// ================================

func TestContactStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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

// ================================
// FOLDER STORE TESTS
// ================================

func TestFolderStore(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(Config{BasePath: tmpDir})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer func() { _ = mgr.Close() }()

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
