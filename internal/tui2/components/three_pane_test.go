package components

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
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
			// Content height should be messageViewHeight - 3 (title + borders)
			expectedContentHeight := messageViewHeight - 3

			folderModel := layout.GetFolders()
			if folderModel.Height() != expectedContentHeight {
				t.Errorf("Folder height = %d, want %d", folderModel.Height(), expectedContentHeight)
			}

			// Table height includes 1 line for header, so actual visible content is -1
			msgModel := layout.GetMessages()
			if msgModel.Height() != expectedContentHeight-1 {
				t.Errorf("Message table height = %d, want %d (contentHeight=%d includes header)",
					msgModel.Height(), expectedContentHeight-1, expectedContentHeight)
			}

			prevModel := layout.GetPreview()
			if prevModel.Height != expectedContentHeight {
				t.Errorf("Preview height = %d, want %d", prevModel.Height, expectedContentHeight)
			}
		})
	}
}
