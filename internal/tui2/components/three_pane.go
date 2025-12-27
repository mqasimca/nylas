// Package components provides reusable Bubble Tea components.
package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// Pane represents which pane has focus.
type Pane int

const (
	// FolderPane is the left folder list pane.
	FolderPane Pane = iota
	// MessagePane is the middle message list pane.
	MessagePane
	// PreviewPane is the right message preview pane.
	PreviewPane
)

// ThreePaneLayout is a three-pane email layout component.
type ThreePaneLayout struct {
	folders  list.Model
	messages table.Model
	preview  viewport.Model

	width  int
	height int
	theme  *styles.Theme

	focused        Pane
	expandedLayout bool // Whether to use expanded layout for focused pane
}

// NewThreePaneLayout creates a new three-pane layout.
func NewThreePaneLayout(theme *styles.Theme) *ThreePaneLayout {
	// Create folder list with custom delegate
	delegate := newFolderDelegate(theme)

	folders := list.New([]list.Item{}, delegate, 0, 0)
	folders.SetShowTitle(false) // Title is added by renderPane
	folders.SetShowStatusBar(true)
	folders.SetShowHelp(false)
	folders.SetFilteringEnabled(false)

	// Create message table
	columns := []table.Column{
		{Title: "From", Width: 20},
		{Title: "Subject", Width: 40},
		{Title: "Date", Width: 15},
	}
	messages := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	// Create preview viewport
	preview := viewport.New(0, 0)

	return &ThreePaneLayout{
		folders:        folders,
		messages:       messages,
		preview:        preview,
		theme:          theme,
		focused:        MessagePane, // Start with message pane focused
		expandedLayout: true,        // Enable dynamic resizing by default
	}
}

// calculatePaneWidths calculates the content width for each pane based on focus and layout mode.
func (t *ThreePaneLayout) calculatePaneWidths(availableContentWidth int) (folderWidth, messageWidth, previewWidth int) {
	if t.expandedLayout {
		switch t.focused {
		case FolderPane:
			// Folders expanded: 40% folders, 30% messages, 30% preview
			folderWidth = availableContentWidth * 40 / 100
			messageWidth = availableContentWidth * 30 / 100
			previewWidth = availableContentWidth - folderWidth - messageWidth
		case MessagePane:
			// Messages expanded (default): 15% folders, 50% messages, 35% preview
			folderWidth = availableContentWidth * 15 / 100
			messageWidth = availableContentWidth * 50 / 100
			previewWidth = availableContentWidth - folderWidth - messageWidth
		case PreviewPane:
			// Preview expanded: 15% folders, 25% messages, 60% preview
			folderWidth = availableContentWidth * 15 / 100
			messageWidth = availableContentWidth * 25 / 100
			previewWidth = availableContentWidth - folderWidth - messageWidth
		}
	} else {
		// Static layout: 20% folders, 35% messages, 45% preview
		folderWidth = availableContentWidth * 20 / 100
		messageWidth = availableContentWidth * 35 / 100
		previewWidth = availableContentWidth - folderWidth - messageWidth
	}
	return
}

// SetSize sets the size of the layout and recalculates pane sizes.
func (t *ThreePaneLayout) SetSize(width, height int) {
	t.width = width
	t.height = height

	// Calculate border/padding overhead using GetFrameSize() - Lipgloss best practice
	// This gives us the exact horizontal and vertical space used by borders/padding
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
	frameWidth, frameHeight := borderStyle.GetFrameSize()

	// Calculate content dimensions
	// Each pane needs: content width + frameWidth
	// Total available: width - (3 * frameWidth) for 3 panes
	totalFrameWidth := frameWidth * 3
	availableContentWidth := width - totalFrameWidth
	if availableContentWidth < 30 {
		availableContentWidth = 30
	}

	// Distribute width based on focus (adaptive layout)
	folderContentWidth, messageContentWidth, previewContentWidth := t.calculatePaneWidths(availableContentWidth)

	// Calculate content height
	// height includes: title (1 line) + frameHeight
	// Available for content: height - title - frameHeight
	contentHeight := height - 1 - frameHeight
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Update components with content dimensions
	t.folders.SetSize(folderContentWidth, contentHeight)

	t.updateMessageTableColumns(messageContentWidth)
	t.messages.SetHeight(contentHeight)
	t.messages.SetWidth(messageContentWidth)

	t.preview.Width = previewContentWidth
	t.preview.Height = contentHeight
}

// updateMessageTableColumns dynamically adjusts column widths based on available space.
func (t *ThreePaneLayout) updateMessageTableColumns(totalWidth int) {
	// Account for borders and padding
	availableWidth := totalWidth - 4
	if availableWidth < 30 {
		availableWidth = 30 // Minimum width
	}

	// Optimize column distribution: From (25%), Subject (55%), Date (20%)
	// Ensure date column is always visible with minimum 12 chars for "Jan 2" or "2h ago"
	dateWidth := max(12, availableWidth*20/100)
	fromWidth := max(15, availableWidth*25/100)
	subjectWidth := availableWidth - fromWidth - dateWidth - 2 // -2 for spacing

	if subjectWidth < 20 {
		subjectWidth = 20
	}

	columns := []table.Column{
		{Title: "From", Width: fromWidth},
		{Title: "Subject", Width: subjectWidth},
		{Title: "Date", Width: dateWidth},
	}

	t.messages.SetColumns(columns)
}

// SetFolders sets the folder list items.
func (t *ThreePaneLayout) SetFolders(items []list.Item) {
	t.folders.SetItems(items)
}

// SetMessages sets the message table rows.
func (t *ThreePaneLayout) SetMessages(rows []table.Row) {
	t.messages.SetRows(rows)
}

// SetPreview sets the preview content.
func (t *ThreePaneLayout) SetPreview(content string) {
	t.preview.SetContent(content)
}

// FocusNext moves focus to the next pane.
func (t *ThreePaneLayout) FocusNext() {
	t.focused = (t.focused + 1) % 3
	t.updateFocus()
}

// FocusPrevious moves focus to the previous pane.
func (t *ThreePaneLayout) FocusPrevious() {
	t.focused = (t.focused + 2) % 3 // +2 is the same as -1 mod 3
	t.updateFocus()
}

// FocusPane sets focus to a specific pane.
func (t *ThreePaneLayout) FocusPane(pane Pane) {
	t.focused = pane
	t.updateFocus()
}

// GetFocused returns the currently focused pane.
func (t *ThreePaneLayout) GetFocused() Pane {
	return t.focused
}

// Update updates the layout based on messages.
func (t *ThreePaneLayout) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Update the focused pane
	switch t.focused {
	case FolderPane:
		t.folders, cmd = t.folders.Update(msg)
		cmds = append(cmds, cmd)

	case MessagePane:
		t.messages, cmd = t.messages.Update(msg)
		cmds = append(cmds, cmd)

	case PreviewPane:
		t.preview, cmd = t.preview.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// View renders the three-pane layout.
func (t *ThreePaneLayout) View() string {
	// Calculate pane widths (must match SetSize calculations)
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
	frameWidth, _ := borderStyle.GetFrameSize()

	totalFrameWidth := frameWidth * 3
	availableContentWidth := t.width - totalFrameWidth

	// Use the same calculation logic as SetSize
	folderWidth, messageWidth, previewWidth := t.calculatePaneWidths(availableContentWidth)

	// Render each pane with explicit width and height
	folderView := t.renderPaneWithWidth(t.folders.View(), t.focused == FolderPane, "Folders", folderWidth+frameWidth, t.height)
	messageView := t.renderPaneWithWidth(t.messages.View(), t.focused == MessagePane, "Messages", messageWidth+frameWidth, t.height)
	previewView := t.renderPaneWithWidth(t.preview.View(), t.focused == PreviewPane, "Preview", previewWidth+frameWidth, t.height)

	// Join horizontally - should now use full width
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		folderView,
		messageView,
		previewView,
	)
}

// SelectedMessage returns the currently selected message row.
func (t *ThreePaneLayout) SelectedMessage() table.Row {
	return t.messages.SelectedRow()
}

// SelectedMessageIndex returns the index of the currently selected message.
func (t *ThreePaneLayout) SelectedMessageIndex() int {
	return t.messages.Cursor()
}

// SelectedFolder returns the currently selected folder.
func (t *ThreePaneLayout) SelectedFolder() list.Item {
	return t.folders.SelectedItem()
}

// renderPaneWithWidth renders a pane with border, title, and explicit width and height.
func (t *ThreePaneLayout) renderPaneWithWidth(content string, focused bool, title string, totalWidth, totalHeight int) string {
	borderColor := t.theme.Dimmed.GetForeground()
	titleColor := t.theme.Secondary

	if focused {
		borderColor = t.theme.Primary
		titleColor = t.theme.Primary
	}

	// Create border style WITHOUT .Width() to avoid lipgloss v1.1.0 bug
	// where .Width() doesn't account for border size (fixed in v2.0.0)
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	// Calculate content dimensions manually: total - borders - padding
	frameWidth, frameHeight := borderStyle.GetFrameSize()
	contentWidth := totalWidth - frameWidth
	if contentWidth < 10 {
		contentWidth = 10 // Minimum content width
	}

	// Calculate content height: totalHeight - title (1 line) - frame
	contentHeight := totalHeight - 1 - frameHeight
	if contentHeight < 5 {
		contentHeight = 5 // Minimum content height
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	// Pad content to exact width AND height using lipgloss utilities
	paddedContent := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Render(content)

	// Add title
	titleBar := titleStyle.Render(title)
	contentWithTitle := lipgloss.JoinVertical(lipgloss.Left, titleBar, paddedContent)

	return borderStyle.Render(contentWithTitle)
}

// updateFocus updates the focus state of all panes.
func (t *ThreePaneLayout) updateFocus() {
	// Update table focus
	if t.focused == MessagePane {
		t.messages.Focus()
	} else {
		t.messages.Blur()
	}

	// Recalculate sizes when focus changes (for dynamic resizing)
	if t.expandedLayout && t.width > 0 && t.height > 0 {
		t.SetSize(t.width, t.height)
	}

	// Folders and preview don't have explicit focus methods,
	// but we handle their appearance in renderPane
}

// GetMessages returns the message table for direct manipulation.
func (t *ThreePaneLayout) GetMessages() *table.Model {
	return &t.messages
}

// GetFolders returns the folder list for direct manipulation.
func (t *ThreePaneLayout) GetFolders() *list.Model {
	return &t.folders
}

// GetPreview returns the preview viewport for direct manipulation.
func (t *ThreePaneLayout) GetPreview() *viewport.Model {
	return &t.preview
}

// folderDelegate is a custom delegate for rendering folder items.
type folderDelegate struct {
	theme *styles.Theme
}

// newFolderDelegate creates a new folder delegate.
func newFolderDelegate(theme *styles.Theme) folderDelegate {
	return folderDelegate{theme: theme}
}

// Height returns the height of a folder item.
func (d folderDelegate) Height() int {
	return 2 // Title + description
}

// Spacing returns the spacing between items.
func (d folderDelegate) Spacing() int {
	return 1
}

// Update handles item updates.
func (d folderDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders a folder item.
func (d folderDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	folderItem, ok := item.(FolderItem)
	if !ok {
		return
	}

	// Get item title and description
	title := folderItem.Title()
	desc := folderItem.Description()

	// Style based on selection
	var titleStyle, descStyle lipgloss.Style
	if index == m.Index() {
		// Selected item
		titleStyle = lipgloss.NewStyle().
			Foreground(d.theme.Primary).
			Bold(true)
		descStyle = lipgloss.NewStyle().
			Foreground(d.theme.Secondary)
	} else {
		// Normal item
		titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
		descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
	}

	// Render title
	_, _ = fmt.Fprintf(w, "%s\n", titleStyle.Render(title))

	// Render description if present
	if desc != "" {
		_, _ = fmt.Fprintf(w, "  %s", descStyle.Render(desc))
	}
}
