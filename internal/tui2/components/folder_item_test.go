package components

import (
	"testing"

	"github.com/mqasimca/nylas/internal/domain"
)

func TestFolderItem_FilterValue(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       string
	}{
		{"inbox", "Inbox", "Inbox"},
		{"sent", "Sent Items", "Sent Items"},
		{"drafts", "Drafts", "Drafts"},
		{"custom", "My Custom Folder", "My Custom Folder"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := FolderItem{
				Folder: domain.Folder{Name: tt.folderName},
			}

			got := item.FilterValue()
			if got != tt.want {
				t.Errorf("FilterValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFolderItem_Title(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       string
	}{
		{"inbox", "Inbox", "📥 Inbox"},
		{"sent", "Sent Items", "📤 Sent Items"},
		{"drafts", "Drafts", "✏️ Drafts"},
		{"empty", "", "📁 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := FolderItem{
				Folder: domain.Folder{Name: tt.folderName},
			}

			got := item.Title()
			if got != tt.want {
				t.Errorf("Title() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFolderItem_Description(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int
		want       string
	}{
		{"with messages", 10, "10 messages"},
		{"one message", 1, "1 messages"},
		{"no messages", 0, ""},
		{"many messages", 999, "999 messages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := FolderItem{
				Folder: domain.Folder{
					Name:       "Test Folder",
					TotalCount: tt.totalCount,
				},
			}

			got := item.Description()
			if got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFolderItem_Implementation(t *testing.T) {
	// Verify FolderItem implements list.Item interface
	var _ interface {
		FilterValue() string
		Title() string
		Description() string
	} = FolderItem{}
}
