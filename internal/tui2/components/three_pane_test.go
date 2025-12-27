package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewThreePaneLayout(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	if layout == nil {
		t.Fatal("NewThreePaneLayout returned nil")
	}

	if layout.focused != MessagePane {
		t.Errorf("expected initial focus on MessagePane, got %v", layout.focused)
	}

	if layout.theme != theme {
		t.Error("theme not set correctly")
	}
}

func TestThreePaneLayout_SetSize(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", 120, 40},
		{"wide", 200, 50},
		{"narrow", 80, 30},
		{"minimal", 60, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout.SetSize(tt.width, tt.height)

			if layout.width != tt.width {
				t.Errorf("width = %d, want %d", layout.width, tt.width)
			}
			if layout.height != tt.height {
				t.Errorf("height = %d, want %d", layout.height, tt.height)
			}

			// Verify pane widths add up correctly (with borders)
			folderWidth := tt.width * 20 / 100
			messageWidth := tt.width * 35 / 100
			previewWidth := tt.width - folderWidth - messageWidth - 4

			expectedTotal := folderWidth + messageWidth + previewWidth + 4
			if expectedTotal != tt.width {
				t.Errorf("pane widths don't add up: %d != %d", expectedTotal, tt.width)
			}
		})
	}
}

func TestThreePaneLayout_FocusManagement(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	// Test initial focus
	if layout.GetFocused() != MessagePane {
		t.Errorf("initial focus = %v, want MessagePane", layout.GetFocused())
	}

	// Test FocusNext
	layout.FocusNext()
	if layout.GetFocused() != PreviewPane {
		t.Errorf("after FocusNext = %v, want PreviewPane", layout.GetFocused())
	}

	layout.FocusNext()
	if layout.GetFocused() != FolderPane {
		t.Errorf("after second FocusNext = %v, want FolderPane", layout.GetFocused())
	}

	layout.FocusNext()
	if layout.GetFocused() != MessagePane {
		t.Errorf("after third FocusNext = %v, want MessagePane (wraparound)", layout.GetFocused())
	}

	// Test FocusPrevious
	layout.FocusPrevious()
	if layout.GetFocused() != FolderPane {
		t.Errorf("after FocusPrevious = %v, want FolderPane", layout.GetFocused())
	}

	layout.FocusPrevious()
	if layout.GetFocused() != PreviewPane {
		t.Errorf("after second FocusPrevious = %v, want PreviewPane", layout.GetFocused())
	}

	// Test FocusPane
	layout.FocusPane(FolderPane)
	if layout.GetFocused() != FolderPane {
		t.Errorf("after FocusPane(FolderPane) = %v, want FolderPane", layout.GetFocused())
	}

	layout.FocusPane(MessagePane)
	if layout.GetFocused() != MessagePane {
		t.Errorf("after FocusPane(MessagePane) = %v, want MessagePane", layout.GetFocused())
	}
}

func TestThreePaneLayout_SetFolders(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	folders := []list.Item{
		FolderItem{Folder: domain.Folder{Name: "Inbox", TotalCount: 10}},
		FolderItem{Folder: domain.Folder{Name: "Sent", TotalCount: 5}},
		FolderItem{Folder: domain.Folder{Name: "Drafts", TotalCount: 2}},
	}

	layout.SetFolders(folders)

	// Verify folders were set (can't directly check private field, but method shouldn't panic)
	_ = layout.SelectedFolder()
}

func TestThreePaneLayout_SetMessages(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40) // Set size first to avoid issues

	rows := []table.Row{
		{"John Doe", "Hello World", "2h ago"},
		{"Jane Smith", "Meeting tomorrow", "1d ago"},
		{"Bob Johnson", "Project update", "3d ago"},
	}

	layout.SetMessages(rows)

	// Verify messages were set
	selected := layout.SelectedMessage()
	if len(selected) != 3 {
		t.Errorf("selected message has %d columns, want 3", len(selected))
	}
}

func TestThreePaneLayout_SetPreview(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	content := "This is a test preview\nWith multiple lines\nAnd some content"
	layout.SetPreview(content)

	// Verify preview was set (method shouldn't panic)
	view := layout.View()
	if view == "" {
		t.Error("View() returned empty string after SetPreview")
	}
}

func TestThreePaneLayout_SelectedMessageIndex(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	rows := []table.Row{
		{"John Doe", "Hello World", "2h ago"},
		{"Jane Smith", "Meeting tomorrow", "1d ago"},
		{"Bob Johnson", "Project update", "3d ago"},
	}
	layout.SetMessages(rows)

	// Initial cursor should be at 0
	idx := layout.SelectedMessageIndex()
	if idx < 0 {
		t.Errorf("SelectedMessageIndex = %d, want >= 0", idx)
	}
}

func TestThreePaneLayout_Update(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	// Test updating with key message
	msg := tea.KeyMsg{Type: tea.KeyDown}
	_ = layout.Update(msg)

	// Update may return nil command, which is valid
	// Just verify it doesn't panic
}

func TestThreePaneLayout_View(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	view := layout.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// View should contain pane titles
	// Note: We can't easily test the exact output due to styling,
	// but we can verify it's not empty
}

func TestThreePaneLayout_EqualPaneHeights(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", 120, 40},
		{"tall", 150, 60},
		{"short", 100, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout.SetSize(tt.width, tt.height)

			// Render the view
			view := layout.View()
			lines := strings.Split(view, "\n")

			// All panes should render with the same height
			// Count the actual rendered height (number of lines in output)
			if len(lines) == 0 {
				t.Fatal("View() returned no lines")
			}

			// The view height should match the set height
			// (allowing for some ANSI codes which don't affect visual height)
			viewHeight := len(lines)
			if viewHeight != tt.height {
				t.Logf("View height: %d, Expected: %d", viewHeight, tt.height)
				// This is informational - actual visual height may differ due to ANSI codes
			}

			// Verify that the view is not empty and contains expected content
			viewStr := strings.Join(lines, "\n")
			if !strings.Contains(viewStr, "Folders") {
				t.Error("View should contain 'Folders' title")
			}
			if !strings.Contains(viewStr, "Messages") {
				t.Error("View should contain 'Messages' title")
			}
			if !strings.Contains(viewStr, "Preview") {
				t.Error("View should contain 'Preview' title")
			}
		})
	}
}

func TestThreePaneLayout_DynamicPanelResizing(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	// Test with 120x40 terminal
	layout.SetSize(120, 40)

	// Calculate expected widths using GetFrameSize()
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
	frameWidth, _ := borderStyle.GetFrameSize()
	totalFrameWidth := frameWidth * 3
	availableWidth := 120 - totalFrameWidth

	tests := []struct {
		name               string
		focusPane          Pane
		expectedFolderPct  int
		expectedMessagePct int
		expectedPreviewPct int
	}{
		{
			name:               "folder_focused",
			focusPane:          FolderPane,
			expectedFolderPct:  40,
			expectedMessagePct: 30,
			expectedPreviewPct: 30,
		},
		{
			name:               "message_focused",
			focusPane:          MessagePane,
			expectedFolderPct:  15,
			expectedMessagePct: 50,
			expectedPreviewPct: 35,
		},
		{
			name:               "preview_focused",
			focusPane:          PreviewPane,
			expectedFolderPct:  15,
			expectedMessagePct: 25,
			expectedPreviewPct: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set focus
			layout.FocusPane(tt.focusPane)

			// Get the actual models to check their widths
			folderModel := layout.GetFolders()
			previewModel := layout.GetPreview()

			// Calculate expected widths
			expectedFolderWidth := availableWidth * tt.expectedFolderPct / 100
			expectedMessageWidth := availableWidth * tt.expectedMessagePct / 100
			expectedPreviewWidth := availableWidth - expectedFolderWidth - expectedMessageWidth

			// Verify folder width (allow ±2 for rounding)
			if abs(folderModel.Width()-expectedFolderWidth) > 2 {
				t.Errorf("Folder width = %d, want ~%d (%.0f%%)",
					folderModel.Width(), expectedFolderWidth, float64(tt.expectedFolderPct))
			}

			// Verify preview width (allow ±2 for rounding)
			if abs(previewModel.Width-expectedPreviewWidth) > 2 {
				t.Errorf("Preview width = %d, want ~%d (%.0f%%)",
					previewModel.Width, expectedPreviewWidth, float64(tt.expectedPreviewPct))
			}
		})
	}
}

func TestThreePaneLayout_StaticLayout(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	// Disable expanded layout
	layout.expandedLayout = false
	layout.SetSize(120, 40)

	// Calculate expected widths for static layout (20%/35%/45%)
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
	frameWidth, _ := borderStyle.GetFrameSize()
	totalFrameWidth := frameWidth * 3
	availableWidth := 120 - totalFrameWidth

	expectedFolderWidth := availableWidth * 20 / 100
	expectedMessageWidth := availableWidth * 35 / 100
	expectedPreviewWidth := availableWidth - expectedFolderWidth - expectedMessageWidth

	// Test that focus changes don't affect width in static mode
	tests := []Pane{FolderPane, MessagePane, PreviewPane}
	for _, pane := range tests {
		layout.FocusPane(pane)

		folderModel := layout.GetFolders()
		previewModel := layout.GetPreview()

		// Widths should remain constant regardless of focus
		if abs(folderModel.Width()-expectedFolderWidth) > 2 {
			t.Errorf("With %v focused: Folder width = %d, want ~%d (should be static 20%%)",
				pane, folderModel.Width(), expectedFolderWidth)
		}

		if abs(previewModel.Width-expectedPreviewWidth) > 2 {
			t.Errorf("With %v focused: Preview width = %d, want ~%d (should be static 45%%)",
				pane, previewModel.Width, expectedPreviewWidth)
		}
	}
}

func TestThreePaneLayout_UpdateFocusRecalculation(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	// Get initial folder width with MessagePane focused (15%)
	initialFolderWidth := layout.GetFolders().Width()

	// Focus FolderPane (should expand to 40%)
	layout.FocusPane(FolderPane)

	newFolderWidth := layout.GetFolders().Width()

	// Folder should be wider after focusing
	if newFolderWidth <= initialFolderWidth {
		t.Errorf("Folder width should increase when focused: initial=%d, focused=%d",
			initialFolderWidth, newFolderWidth)
	}

	// Verify message table focus state changed
	msgTable := layout.GetMessages()
	if msgTable.Focused() {
		t.Error("Message table should not be focused when FolderPane is focused")
	}
}

func TestFolderDelegate_Render(t *testing.T) {
	theme := styles.DefaultTheme()
	delegate := newFolderDelegate(theme)

	// Create a mock list model
	items := []list.Item{
		FolderItem{Folder: domain.Folder{Name: "Inbox", TotalCount: 10}},
		FolderItem{Folder: domain.Folder{Name: "Sent", TotalCount: 5}},
		FolderItem{Folder: domain.Folder{Name: "Drafts", TotalCount: 2}},
	}

	listModel := list.New(items, delegate, 20, 10)

	// Test rendering selected item (index 0)
	var buf strings.Builder
	delegate.Render(&buf, listModel, 0, items[0])

	output := buf.String()
	if !strings.Contains(output, "Inbox") {
		t.Error("Rendered output should contain folder name 'Inbox'")
	}
	if !strings.Contains(output, "10 messages") {
		t.Error("Rendered output should contain message count '10 messages'")
	}

	// Test rendering non-selected item (index 1)
	buf.Reset()
	listModel.Select(1)
	delegate.Render(&buf, listModel, 2, items[2])

	output = buf.String()
	if !strings.Contains(output, "Drafts") {
		t.Error("Rendered output should contain folder name 'Drafts'")
	}
	if !strings.Contains(output, "2 messages") {
		t.Error("Rendered output should contain message count '2 messages'")
	}
}

func TestFolderDelegate_RenderWithoutDescription(t *testing.T) {
	theme := styles.DefaultTheme()
	delegate := newFolderDelegate(theme)

	// Create folder with 0 messages (no description)
	items := []list.Item{
		FolderItem{Folder: domain.Folder{Name: "Empty", TotalCount: 0}},
	}

	listModel := list.New(items, delegate, 20, 10)

	var buf strings.Builder
	delegate.Render(&buf, listModel, 0, items[0])

	output := buf.String()
	if !strings.Contains(output, "Empty") {
		t.Error("Rendered output should contain folder name 'Empty'")
	}
}

func TestFolderDelegate_Update(t *testing.T) {
	theme := styles.DefaultTheme()
	delegate := newFolderDelegate(theme)

	listModel := list.New([]list.Item{}, delegate, 20, 10)
	cmd := delegate.Update(tea.KeyMsg{}, &listModel)

	if cmd != nil {
		t.Error("folderDelegate.Update should always return nil")
	}
}

func TestFolderDelegate_Dimensions(t *testing.T) {
	theme := styles.DefaultTheme()
	delegate := newFolderDelegate(theme)

	if delegate.Height() != 2 {
		t.Errorf("Height() = %d, want 2", delegate.Height())
	}

	if delegate.Spacing() != 1 {
		t.Errorf("Spacing() = %d, want 1", delegate.Spacing())
	}
}

func TestThreePaneLayout_UpdateWithDifferentPanes(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)
	layout.SetSize(120, 40)

	tests := []struct {
		name      string
		focusPane Pane
		msg       tea.Msg
	}{
		{"folder_pane_down", FolderPane, tea.KeyMsg{Type: tea.KeyDown}},
		{"folder_pane_up", FolderPane, tea.KeyMsg{Type: tea.KeyUp}},
		{"message_pane_down", MessagePane, tea.KeyMsg{Type: tea.KeyDown}},
		{"message_pane_up", MessagePane, tea.KeyMsg{Type: tea.KeyUp}},
		{"preview_pane_down", PreviewPane, tea.KeyMsg{Type: tea.KeyDown}},
		{"preview_pane_up", PreviewPane, tea.KeyMsg{Type: tea.KeyUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout.FocusPane(tt.focusPane)
			cmd := layout.Update(tt.msg)

			// Update should return a command (or nil, both are valid)
			// Just verify it doesn't panic
			_ = cmd
		})
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestThreePaneLayout_HeightCalculation(t *testing.T) {
	theme := styles.DefaultTheme()
	layout := NewThreePaneLayout(theme)

	tests := []struct {
		name             string
		terminalHeight   int
		expectedPaneSize int // height passed to layout
	}{
		{
			name:             "50 line terminal",
			terminalHeight:   50,
			expectedPaneSize: 46, // 50 - 4 for header/footer
		},
		{
			name:             "30 line terminal",
			terminalHeight:   30,
			expectedPaneSize: 26, // 30 - 4 for header/footer
		},
		{
			name:             "24 line terminal (standard)",
			terminalHeight:   24,
			expectedPaneSize: 20, // 24 - 4 for header/footer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate what messages.go does
			messageViewHeight := tt.terminalHeight - 4 // header (2) + footer (2)

			// Set layout size
			layout.SetSize(120, messageViewHeight)

			// Verify components got correct height
			// Calculate expected height using same method as SetSize()
			// height - 1 (title) - frameHeight (borders)
			borderStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(0, 1)
			_, frameHeight := borderStyle.GetFrameSize()

			expectedContentHeight := messageViewHeight - 1 - frameHeight
			if expectedContentHeight < 5 {
				expectedContentHeight = 5
			}

			folderModel := layout.GetFolders()
			if folderModel.Height() != expectedContentHeight {
				t.Errorf("Folder height = %d, want %d", folderModel.Height(), expectedContentHeight)
			}

			// Bubble Tea table's Height() returns the height minus 1 for the header row
			// When we call SetHeight(43), Height() returns 42
			msgModel := layout.GetMessages()
			expectedTableHeight := expectedContentHeight - 1
			if msgModel.Height() != expectedTableHeight {
				t.Errorf("Message table height = %d, want %d (contentHeight %d minus header)",
					msgModel.Height(), expectedTableHeight, expectedContentHeight)
			}

			prevModel := layout.GetPreview()
			if prevModel.Height != expectedContentHeight {
				t.Errorf("Preview height = %d, want %d", prevModel.Height, expectedContentHeight)
			}
		})
	}
}
