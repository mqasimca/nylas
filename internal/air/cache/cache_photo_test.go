package cache

import (
	"os"
	"testing"
	"time"
)

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
