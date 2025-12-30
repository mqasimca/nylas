package cache

import (
	"testing"
	"time"
)

// ================================
// ENCRYPTION HELPER TESTS
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
