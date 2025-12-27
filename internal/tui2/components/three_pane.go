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

	focused Pane
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
		folders:  folders,
		messages: messages,
		preview:  preview,
		theme:    theme,
		focused:  MessagePane, // Start with message pane focused
	}
}

// SetSize sets the size of the layout and recalculates pane sizes.
func (t *ThreePaneLayout) SetSize(width, height int) {
	t.width = width
	t.height = height

	// Calculate pane widths: 20% folders, 35% messages, 45% preview
	folderWidth := width * 20 / 100
	messageWidth := width * 35 / 100
	previewWidth := width - folderWidth - messageWidth - 4 // -4 for borders

	// Content height accounts for:
	// - 1 line for title (added by renderPane)
	// - 2 lines for top/bottom border (added by renderPane)
	// Total: height - 3
	contentHeight := height - 3

	// Update folder list
	t.folders.SetSize(folderWidth, contentHeight)

	// Update message table
	t.messages.SetHeight(contentHeight)
	t.messages.SetWidth(messageWidth)

	// Update preview viewport
	t.preview.Width = previewWidth
	t.preview.Height = contentHeight
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
	// Render each pane with styling
	folderView := t.renderPane(t.folders.View(), t.focused == FolderPane, "Folders")
	messageView := t.renderPane(t.messages.View(), t.focused == MessagePane, "Messages")
	previewView := t.renderPane(t.preview.View(), t.focused == PreviewPane, "Preview")

	// Join horizontally
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

// renderPane renders a pane with border and title.
func (t *ThreePaneLayout) renderPane(content string, focused bool, title string) string {
	borderColor := t.theme.Dimmed.GetForeground()
	titleColor := t.theme.Secondary

	if focused {
		borderColor = t.theme.Primary
		titleColor = t.theme.Primary
	}

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	// Add title
	titleBar := titleStyle.Render(title)
	contentWithTitle := lipgloss.JoinVertical(lipgloss.Left, titleBar, content)

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
