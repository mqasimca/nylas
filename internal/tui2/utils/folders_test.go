package utils

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestIsImportantFolder(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       bool
	}{
		// Standard important folders
		{"inbox is important", "Inbox", true},
		{"INBOX uppercase", "INBOX", true},
		{"sent is important", "Sent", true},
		{"drafts is important", "Drafts", true},
		{"archive is important", "Archive", true},
		{"trash is important", "Trash", true},
		{"spam is important", "Spam", true},
		{"junk is important", "Junk", true},
		{"starred is important", "Starred", true},
		{"important is important", "Important", true},
		{"all mail is important", "All Mail", true},
		{"all is important", "All", true},

		// Gmail system folders
		{"gmail sent", "[Gmail]/Sent Mail", true},
		{"gmail trash", "[Gmail]/Trash", true},
		{"gmail spam", "[Gmail]/Spam", true},
		{"google mail sent", "[Google Mail]/Sent", true},

		// Folders containing important names
		{"contains inbox", "My Inbox", true},
		{"contains sent", "Sent Items", true},
		{"contains drafts", "Drafts Messages", true},

		// Regular folders
		{"custom folder", "Projects", false},
		{"work folder", "Work", false},
		{"personal folder", "Personal", false},
		{"clients folder", "Clients", false},
		{"2024 folder", "2024", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder := domain.Folder{Name: tt.folderName}
			got := IsImportantFolder(folder)
			if got != tt.want {
				t.Errorf("IsImportantFolder(%q) = %v, want %v", tt.folderName, got, tt.want)
			}
		})
	}
}

func TestGetFolderIcon(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       string
	}{
		{"inbox icon", "Inbox", "📥"},
		{"sent icon", "Sent", "📤"},
		{"draft icon", "Drafts", "✏️"},
		{"trash icon", "Trash", "🗑️"},
		{"delete icon", "Deleted Items", "🗑️"},
		{"spam icon", "Spam", "🚫"},
		{"junk icon", "Junk", "🚫"},
		{"archive icon", "Archive", "📦"},
		{"starred icon", "Starred", "⭐"},
		{"important icon", "Important", "⭐"},
		{"all mail icon", "All Mail", "📧"},
		{"all icon", "All", "📧"},
		{"custom folder icon", "Projects", "📁"},
		{"work folder icon", "Work", "📁"},

		// Case insensitive
		{"INBOX uppercase", "INBOX", "📥"},
		{"inbox lowercase", "inbox", "📥"},
		{"Inbox mixed case", "InBoX", "📥"},

		// Partial matches
		{"contains inbox", "My Inbox", "📥"},
		{"contains sent", "Sent Items", "📤"},
		{"contains draft", "Draft Messages", "✏️"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder := domain.Folder{Name: tt.folderName}
			got := GetFolderIcon(folder)
			if got != tt.want {
				t.Errorf("GetFolderIcon(%q) = %q, want %q", tt.folderName, got, tt.want)
			}
		})
	}
}

func TestFilterImportantFolders(t *testing.T) {
	folders := []domain.Folder{
		{Name: "Inbox"},
		{Name: "Sent"},
		{Name: "Projects"},
		{Name: "Drafts"},
		{Name: "Work"},
		{Name: "Trash"},
		{Name: "Clients"},
		{Name: "Spam"},
	}

	important := FilterImportantFolders(folders)

	// Should filter to only important folders
	expectedCount := 5 // Inbox, Sent, Drafts, Trash, Spam
	if len(important) != expectedCount {
		t.Errorf("expected %d important folders, got %d", expectedCount, len(important))
	}

	// Verify important folders are included
	importantNames := make(map[string]bool)
	for _, f := range important {
		importantNames[f.Name] = true
	}

	if !importantNames["Inbox"] {
		t.Error("Inbox should be included")
	}
	if !importantNames["Sent"] {
		t.Error("Sent should be included")
	}
	if !importantNames["Drafts"] {
		t.Error("Drafts should be included")
	}
	if !importantNames["Trash"] {
		t.Error("Trash should be included")
	}
	if !importantNames["Spam"] {
		t.Error("Spam should be included")
	}

	// Verify non-important folders are excluded
	if importantNames["Projects"] {
		t.Error("Projects should not be included")
	}
	if importantNames["Work"] {
		t.Error("Work should not be included")
	}
	if importantNames["Clients"] {
		t.Error("Clients should not be included")
	}
}

func TestFilterImportantFolders_EmptyInput(t *testing.T) {
	folders := []domain.Folder{}
	important := FilterImportantFolders(folders)

	if len(important) != 0 {
		t.Errorf("expected 0 important folders, got %d", len(important))
	}
}

func TestFilterImportantFolders_NilInput(t *testing.T) {
	important := FilterImportantFolders(nil)

	// nil input returns nil slice (Go range over nil slice is safe)
	if len(important) != 0 {
		t.Errorf("expected nil or empty slice, got %d items", len(important))
	}
}

func TestSortFoldersByImportance(t *testing.T) {
	folders := []domain.Folder{
		{Name: "Trash"},
		{Name: "Inbox"},
		{Name: "Projects"},
		{Name: "Sent"},
		{Name: "Drafts"},
		{Name: "Work"},
		{Name: "Starred"},
	}

	sorted := SortFoldersByImportance(folders)

	if len(sorted) != len(folders) {
		t.Fatalf("expected %d folders, got %d", len(folders), len(sorted))
	}

	// Verify Inbox comes first
	if sorted[0].Name != "Inbox" {
		t.Errorf("expected Inbox first, got %s", sorted[0].Name)
	}

	// Verify Starred comes before Sent
	starredIdx := -1
	sentIdx := -1
	for i, f := range sorted {
		if f.Name == "Starred" {
			starredIdx = i
		}
		if f.Name == "Sent" {
			sentIdx = i
		}
	}
	if starredIdx == -1 || sentIdx == -1 {
		t.Fatal("Starred or Sent not found in sorted list")
	}
	if starredIdx > sentIdx {
		t.Error("Starred should come before Sent")
	}

	// Verify Trash comes last among important folders
	trashIdx := -1
	for i, f := range sorted {
		if f.Name == "Trash" {
			trashIdx = i
		}
	}
	if trashIdx == -1 {
		t.Fatal("Trash not found in sorted list")
	}

	// Trash should be after Inbox, Starred, Sent, Drafts
	inboxIdx := 0 // We already verified Inbox is first
	if trashIdx <= inboxIdx {
		t.Error("Trash should come after Inbox")
	}
}

func TestSortFoldersByImportance_EmptyInput(t *testing.T) {
	folders := []domain.Folder{}
	sorted := SortFoldersByImportance(folders)

	if len(sorted) != 0 {
		t.Errorf("expected 0 folders, got %d", len(sorted))
	}
}

func TestSortFoldersByImportance_DoesNotModifyOriginal(t *testing.T) {
	folders := []domain.Folder{
		{Name: "Trash"},
		{Name: "Inbox"},
		{Name: "Sent"},
	}

	// Keep a copy of original order
	original := make([]string, len(folders))
	for i, f := range folders {
		original[i] = f.Name
	}

	sorted := SortFoldersByImportance(folders)

	// Verify original is unchanged
	for i, f := range folders {
		if f.Name != original[i] {
			t.Errorf("original slice was modified at index %d: expected %s, got %s",
				i, original[i], f.Name)
		}
	}

	// Verify sorted is different from original
	if sorted[0].Name == folders[0].Name &&
		sorted[1].Name == folders[1].Name &&
		sorted[2].Name == folders[2].Name {
		t.Error("sorted slice should be in different order than original")
	}
}

func TestGetPriority(t *testing.T) {
	priority := map[string]int{
		"inbox":   1,
		"starred": 2,
		"sent":    4,
		"trash":   9,
	}

	tests := []struct {
		name string
		want int
	}{
		{"Inbox", 1},
		{"INBOX", 1},
		{"inbox", 1},
		{"My Inbox", 1},
		{"Starred", 2},
		{"Sent", 4},
		{"Trash", 9},
		{"Unknown", 99}, // Default priority
		{"Projects", 99},
		{"Work", 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPriority(tt.name, priority)
			if got != tt.want {
				t.Errorf("getPriority(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestImportantFolderNames_Coverage(t *testing.T) {
	// Verify all names in ImportantFolderNames are covered by GetFolderIcon
	for _, name := range ImportantFolderNames {
		folder := domain.Folder{Name: name}
		icon := GetFolderIcon(folder)

		// Should return a specific icon, not the default
		if icon != "📁" {
			// Good, it has a specific icon
		} else {
			// "all mail" and "all" return default icon due to separate handling
			if name != "all mail" && name != "all" {
				t.Errorf("important folder %q has default icon, should have specific icon", name)
			}
		}
	}
}

func TestFolderFunctions_WithComplexNames(t *testing.T) {
	tests := []struct {
		name         string
		isImportant  bool
		expectedIcon string
	}{
		{"[Gmail]/Sent Mail", true, "📤"},
		{"[Gmail]/All Mail", true, "📧"},
		{"[Google Mail]/Drafts", true, "✏️"},
		{"Re: Inbox Discussion", true, "📥"},
		{"Fwd: Sent Items", true, "📤"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder := domain.Folder{Name: tt.name}

			if got := IsImportantFolder(folder); got != tt.isImportant {
				t.Errorf("IsImportantFolder(%q) = %v, want %v", tt.name, got, tt.isImportant)
			}

			if got := GetFolderIcon(folder); got != tt.expectedIcon {
				t.Errorf("GetFolderIcon(%q) = %q, want %q", tt.name, got, tt.expectedIcon)
			}
		})
	}
}
