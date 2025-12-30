package cache

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"testing"
	"time"
)

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
