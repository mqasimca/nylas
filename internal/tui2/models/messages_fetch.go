package models

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/utils"
)

func (m *MessageList) fetchMessages() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch threads instead of individual messages for proper conversation grouping
		params := &domain.ThreadQueryParams{
			Limit: 50,
		}
		threads, err := m.global.Client.GetThreads(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return threadsLoadedMsg{threads}
	}
}

// fetchMessagesForFolder fetches threads filtered by folder ID.
func (m *MessageList) fetchMessagesForFolder(folderID string) tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch threads filtered by folder
		params := &domain.ThreadQueryParams{
			In:    []string{folderID},
			Limit: 50,
		}

		threads, err := m.global.Client.GetThreads(ctx, m.global.GrantID, params)
		if err != nil {
			return errMsg{err}
		}
		return threadsLoadedMsg{threads}
	}
}

// fetchFolders fetches the folder list (lazy loaded).
func (m *MessageList) fetchFolders() tea.Cmd {
	return func() tea.Msg {
		// Rate limit to avoid API errors
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		folders, err := m.global.Client.GetFolders(ctx, m.global.GrantID)
		if err != nil {
			return errMsg{err}
		}

		// Filter to show only important folders
		importantFolders := utils.FilterImportantFolders(folders)

		// Sort by importance
		sortedFolders := utils.SortFoldersByImportance(importantFolders)

		// Convert to list items with icons
		items := make([]list.Item, len(sortedFolders))
		for i, folder := range sortedFolders {
			items[i] = components.FolderItem{Folder: folder}
		}

		// Add "Show all folders..." option at the end if we filtered some
		if len(folders) > len(sortedFolders) {
			items = append(items, components.FolderItem{
				Folder: domain.Folder{
					Name:       fmt.Sprintf("📋 Show all %d folders...", len(folders)),
					ID:         "_show_all",
					TotalCount: len(folders) - len(sortedFolders),
				},
			})
		}

		return foldersLoadedMsg{folders: items}
	}
}

// updateMessageTable updates the message table in the layout.
