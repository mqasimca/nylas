package cache

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
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

func TestAttachmentStore(t *testing.T) {
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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = file.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
	defer func() { _ = mgr.Close() }()

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
