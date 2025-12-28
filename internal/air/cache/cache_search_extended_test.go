package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	defer func() { _ = mgr.Close() }()

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
