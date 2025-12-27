package components

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

// FolderItem wraps a domain.Folder to implement list.Item.
type FolderItem struct {
	Folder domain.Folder
}

// FilterValue returns the value to filter on.
func (f FolderItem) FilterValue() string {
	return f.Folder.Name
}

// Title returns the folder title with icon.
func (f FolderItem) Title() string {
	// Get icon for folder
	icon := utils.GetFolderIcon(f.Folder)
	return fmt.Sprintf("%s %s", icon, f.Folder.Name)
}

// Description returns the folder description.
func (f FolderItem) Description() string {
	// Show unread count if available
	if f.Folder.TotalCount > 0 {
		return fmt.Sprintf("%d messages", f.Folder.TotalCount)
	}
	return ""
}
