// Package utils provides utility functions for the TUI.
package utils

import (
	"strings"

	"github.com/mqasimca/nylas/internal/domain"
)

// ImportantFolderNames lists common important folder names.
var ImportantFolderNames = []string{
	"inbox",
	"sent",
	"drafts",
	"archive",
	"trash",
	"spam",
	"junk",
	"starred",
	"important",
	"all mail",
	"all",
}

// IsImportantFolder checks if a folder is considered important.
func IsImportantFolder(folder domain.Folder) bool {
	name := strings.ToLower(folder.Name)

	// Check against known important folder names
	for _, important := range ImportantFolderNames {
		if strings.Contains(name, important) {
			return true
		}
	}

	// Check common attributes
	if strings.HasPrefix(name, "[gmail]/") || strings.HasPrefix(name, "[google mail]/") {
		// Gmail system folders are important
		return true
	}

	return false
}

// GetFolderIcon returns an icon for the folder type.
func GetFolderIcon(folder domain.Folder) string {
	name := strings.ToLower(folder.Name)

	switch {
	case strings.Contains(name, "inbox"):
		return "📥"
	case strings.Contains(name, "sent"):
		return "📤"
	case strings.Contains(name, "draft"):
		return "✏️"
	case strings.Contains(name, "trash") || strings.Contains(name, "delete"):
		return "🗑️"
	case strings.Contains(name, "spam") || strings.Contains(name, "junk"):
		return "🚫"
	case strings.Contains(name, "archive"):
		return "📦"
	case strings.Contains(name, "starred") || strings.Contains(name, "important"):
		return "⭐"
	case strings.Contains(name, "all"):
		return "📧"
	default:
		return "📁"
	}
}

// FilterImportantFolders returns only important folders from the list.
func FilterImportantFolders(folders []domain.Folder) []domain.Folder {
	var important []domain.Folder

	for _, folder := range folders {
		if IsImportantFolder(folder) {
			important = append(important, folder)
		}
	}

	return important
}

// SortFoldersByImportance sorts folders with most important first.
func SortFoldersByImportance(folders []domain.Folder) []domain.Folder {
	// Priority order
	priority := map[string]int{
		"inbox":     1,
		"starred":   2,
		"important": 3,
		"sent":      4,
		"drafts":    5,
		"archive":   6,
		"all mail":  7,
		"spam":      8,
		"junk":      8,
		"trash":     9,
	}

	sorted := make([]domain.Folder, len(folders))
	copy(sorted, folders)

	// Simple bubble sort by priority
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			iPriority := getPriority(sorted[i].Name, priority)
			jPriority := getPriority(sorted[j].Name, priority)

			if iPriority > jPriority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func getPriority(name string, priorityMap map[string]int) int {
	lowerName := strings.ToLower(name)

	for key, priority := range priorityMap {
		if strings.Contains(lowerName, key) {
			return priority
		}
	}

	return 99 // Default low priority
}
