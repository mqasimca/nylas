package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// ConfirmDialogResult represents the user's choice.
type ConfirmDialogResult int

const (
	ConfirmDialogResultPending ConfirmDialogResult = iota
	ConfirmDialogResultConfirm
	ConfirmDialogResultCancel
)

// ConfirmDialogMsg is sent when the dialog is closed.
type ConfirmDialogMsg struct {
	Result ConfirmDialogResult
	Data   any // Optional data passed through
}

// ConfirmDialog is a confirmation dialog component.
type ConfirmDialog struct {
	theme   *styles.Theme
	title   string
	message string
	data    any

	confirmText string
	cancelText  string

	focusedButton int // 0 = cancel, 1 = confirm
	visible       bool

	width  int
	height int
}

// NewConfirmDialog creates a new confirmation dialog.
func NewConfirmDialog(theme *styles.Theme, title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		theme:         theme,
		title:         title,
		message:       message,
		confirmText:   "Confirm",
		cancelText:    "Cancel",
		focusedButton: 0, // Cancel is focused by default (safer)
		visible:       true,
	}
}

// SetData sets optional data to pass through with the result.
func (d *ConfirmDialog) SetData(data any) {
	d.data = data
}

// SetButtonLabels sets custom button labels.
func (d *ConfirmDialog) SetButtonLabels(confirm, cancel string) {
	d.confirmText = confirm
	d.cancelText = cancel
}

// SetSize sets the dialog size.
func (d *ConfirmDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// Show shows the dialog.
func (d *ConfirmDialog) Show() {
	d.visible = true
	d.focusedButton = 0 // Reset to cancel (safer default)
}

// Hide hides the dialog.
func (d *ConfirmDialog) Hide() {
	d.visible = false
}

// IsVisible returns true if the dialog is visible.
func (d *ConfirmDialog) IsVisible() bool {
	return d.visible
}

// Init implements tea.Model.
func (d *ConfirmDialog) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d *ConfirmDialog) Update(msg tea.Msg) (*ConfirmDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		// Handle escape - cancel
		if key.Code == tea.KeyEsc {
			d.visible = false
			return d, func() tea.Msg {
				return ConfirmDialogMsg{
					Result: ConfirmDialogResultCancel,
					Data:   d.data,
				}
			}
		}

		// Handle tab - switch focus
		if key.Code == tea.KeyTab || keyStr == "shift+tab" {
			d.focusedButton = (d.focusedButton + 1) % 2
			return d, nil
		}

		// Handle left/right arrows
		if key.Code == tea.KeyLeft || keyStr == "h" {
			d.focusedButton = 0
			return d, nil
		}
		if key.Code == tea.KeyRight || keyStr == "l" {
			d.focusedButton = 1
			return d, nil
		}

		// Handle enter
		if key.Code == tea.KeyEnter {
			d.visible = false
			result := ConfirmDialogResultCancel
			if d.focusedButton == 1 {
				result = ConfirmDialogResultConfirm
			}
			return d, func() tea.Msg {
				return ConfirmDialogMsg{
					Result: result,
					Data:   d.data,
				}
			}
		}

		// Handle y/n shortcuts
		if keyStr == "y" || keyStr == "Y" {
			d.visible = false
			return d, func() tea.Msg {
				return ConfirmDialogMsg{
					Result: ConfirmDialogResultConfirm,
					Data:   d.data,
				}
			}
		}
		if keyStr == "n" || keyStr == "N" {
			d.visible = false
			return d, func() tea.Msg {
				return ConfirmDialogMsg{
					Result: ConfirmDialogResultCancel,
					Data:   d.data,
				}
			}
		}
	}

	return d, nil
}

// View implements tea.Model.
func (d *ConfirmDialog) View() string {
	if !d.visible {
		return ""
	}

	var b strings.Builder

	// Dialog box style
	dialogWidth := 50
	if d.width > 0 && d.width < dialogWidth+10 {
		dialogWidth = d.width - 10
	}
	if dialogWidth < 30 {
		dialogWidth = 30
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Primary).
		Padding(1, 2).
		Width(dialogWidth)

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(d.theme.Warning)
	b.WriteString(titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// Message
	messageStyle := lipgloss.NewStyle().Foreground(d.theme.Foreground)
	b.WriteString(messageStyle.Render(d.message))
	b.WriteString("\n\n")

	// Buttons
	cancelStyle := lipgloss.NewStyle().Padding(0, 2)
	confirmStyle := lipgloss.NewStyle().Padding(0, 2)

	if d.focusedButton == 0 {
		cancelStyle = cancelStyle.Background(d.theme.Primary).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	} else {
		cancelStyle = cancelStyle.Border(lipgloss.NormalBorder()).BorderForeground(d.theme.Dimmed.GetForeground())
	}

	if d.focusedButton == 1 {
		confirmStyle = confirmStyle.Background(d.theme.Error).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	} else {
		confirmStyle = confirmStyle.Border(lipgloss.NormalBorder()).BorderForeground(d.theme.Dimmed.GetForeground())
	}

	b.WriteString(cancelStyle.Render(d.cancelText) + "  " + confirmStyle.Render(d.confirmText))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(d.theme.Dimmed.GetForeground())
	b.WriteString(helpStyle.Render("y: confirm  n: cancel  Tab: switch  Enter: select"))

	return boxStyle.Render(b.String())
}

// GetData returns the dialog data.
func (d *ConfirmDialog) GetData() any {
	return d.data
}
