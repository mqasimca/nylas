package cache

import (
	"testing"
)

// ================================
// CALENDAR STORE TESTS
// ================================

func TestCalendarStore(t *testing.T) {
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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
