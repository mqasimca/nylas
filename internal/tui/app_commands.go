// Package tui provides a k9s-style terminal user interface for Nylas.
package tui

import (
	"github.com/gdamore/tcell/v2"
)

func (a *App) pageMove(lines int, down bool) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	// Get the table from the view if it's a table-based view
	if tableView, ok := view.Primitive().(*Table); ok {
		row, col := tableView.GetSelection()
		rowCount := tableView.GetRowCount()

		var newRow int
		if down {
			newRow = row + lines
			if newRow > rowCount {
				newRow = rowCount
			}
		} else {
			newRow = row - lines
			if newRow < 1 {
				newRow = 1 // Skip header
			}
		}

		tableView.Select(newRow, col)
	}
}

// goToTop moves selection to the first row.
func (a *App) goToTop() {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		tableView.Select(1, 0) // Row 1 is first data row (0 is header)
	}
}

// goToBottom moves selection to the last row.
func (a *App) goToBottom() {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		rowCount := tableView.GetRowCount()
		if rowCount > 0 {
			tableView.Select(rowCount, 0)
		}
	}
}

// goToRow moves selection to a specific row number (1-indexed).
func (a *App) goToRow(row int) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		rowCount := tableView.GetRowCount()
		if row < 1 {
			row = 1
		}
		if row > rowCount {
			row = rowCount
		}
		tableView.Select(row, 0)
	}
}

// executeCommand executes a command on the current view/item.
func (a *App) executeCommand(cmd string) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	// Create a synthetic key event based on the command
	var event *tcell.EventKey

	switch cmd {
	case "delete":
		// For now, just flash a message - delete requires view-specific handling
		a.Flash(FlashWarn, "Delete not implemented for this view")
		return
	case "star":
		event = tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	case "unstar":
		event = tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	case "read":
		// Mark as read - no direct key, flash message
		a.Flash(FlashInfo, "Use 'u' to toggle unread status")
		return
	case "unread":
		event = tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone)
	case "compose":
		event = tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone)
	case "reply":
		event = tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
	case "replyall":
		event = tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone)
	case "forward":
		a.Flash(FlashInfo, "Forward: coming soon")
		return
	default:
		return
	}

	if event != nil {
		view.HandleKey(event)
	}
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseInt converts a string to int, returning 0 on error.
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
