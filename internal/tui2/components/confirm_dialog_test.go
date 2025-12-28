package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewConfirmDialog(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Delete Item", "Are you sure?")

	if dialog == nil {
		t.Fatal("expected dialog to be created")
	}
	if dialog.title != "Delete Item" {
		t.Errorf("expected title 'Delete Item', got '%s'", dialog.title)
	}
	if dialog.message != "Are you sure?" {
		t.Errorf("expected message 'Are you sure?', got '%s'", dialog.message)
	}
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible by default")
	}
	if dialog.focusedButton != 0 {
		t.Error("expected cancel button (0) to be focused by default for safety")
	}
}

func TestConfirmDialog_SetData(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	testData := "test-id-123"
	dialog.SetData(testData)

	if dialog.GetData() != testData {
		t.Errorf("expected data '%s', got '%v'", testData, dialog.GetData())
	}
}

func TestConfirmDialog_SetButtonLabels(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	dialog.SetButtonLabels("Yes, Delete", "No, Keep")

	if dialog.confirmText != "Yes, Delete" {
		t.Errorf("expected confirm text 'Yes, Delete', got '%s'", dialog.confirmText)
	}
	if dialog.cancelText != "No, Keep" {
		t.Errorf("expected cancel text 'No, Keep', got '%s'", dialog.cancelText)
	}
}

func TestConfirmDialog_ShowHide(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	dialog.Hide()
	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden")
	}

	dialog.Show()
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible")
	}
}

func TestConfirmDialog_Update_TabSwitchesFocus(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	// Initially cancel (0) is focused
	if dialog.focusedButton != 0 {
		t.Errorf("expected initial focus on cancel (0), got %d", dialog.focusedButton)
	}

	// Tab should switch to confirm (1)
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	dialog, _ = dialog.Update(tabKey)

	if dialog.focusedButton != 1 {
		t.Errorf("expected focus on confirm (1) after tab, got %d", dialog.focusedButton)
	}

	// Tab again should switch back to cancel (0)
	dialog, _ = dialog.Update(tabKey)
	if dialog.focusedButton != 0 {
		t.Errorf("expected focus on cancel (0) after second tab, got %d", dialog.focusedButton)
	}
}

func TestConfirmDialog_Update_ArrowKeysFocus(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	// Right arrow should switch to confirm
	rightKey := tea.KeyPressMsg{Code: tea.KeyRight}
	dialog, _ = dialog.Update(rightKey)

	if dialog.focusedButton != 1 {
		t.Errorf("expected focus on confirm (1) after right arrow, got %d", dialog.focusedButton)
	}

	// Left arrow should switch back to cancel
	leftKey := tea.KeyPressMsg{Code: tea.KeyLeft}
	dialog, _ = dialog.Update(leftKey)

	if dialog.focusedButton != 0 {
		t.Errorf("expected focus on cancel (0) after left arrow, got %d", dialog.focusedButton)
	}
}

func TestConfirmDialog_Update_EscapeCancel(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")
	dialog.SetData("test-data")

	escKey := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := dialog.Update(escKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(ConfirmDialogMsg)
	if !ok {
		t.Fatalf("expected ConfirmDialogMsg, got %T", msg)
	}

	if result.Result != ConfirmDialogResultCancel {
		t.Errorf("expected cancel result, got %d", result.Result)
	}
	if result.Data != "test-data" {
		t.Errorf("expected data 'test-data', got '%v'", result.Data)
	}
	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden after escape")
	}
}

func TestConfirmDialog_Update_EnterConfirm(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")
	dialog.SetData("item-id")

	// Focus on confirm button
	dialog.focusedButton = 1

	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := dialog.Update(enterKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(ConfirmDialogMsg)
	if !ok {
		t.Fatalf("expected ConfirmDialogMsg, got %T", msg)
	}

	if result.Result != ConfirmDialogResultConfirm {
		t.Errorf("expected confirm result, got %d", result.Result)
	}
}

func TestConfirmDialog_Update_EnterCancel(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	// Focus on cancel button (default)
	dialog.focusedButton = 0

	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := dialog.Update(enterKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(ConfirmDialogMsg)
	if !ok {
		t.Fatalf("expected ConfirmDialogMsg, got %T", msg)
	}

	if result.Result != ConfirmDialogResultCancel {
		t.Errorf("expected cancel result, got %d", result.Result)
	}
}

func TestConfirmDialog_Update_YKeyConfirms(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	yKey := tea.KeyPressMsg{Text: "y"}
	_, cmd := dialog.Update(yKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(ConfirmDialogMsg)
	if !ok {
		t.Fatalf("expected ConfirmDialogMsg, got %T", msg)
	}

	if result.Result != ConfirmDialogResultConfirm {
		t.Errorf("expected confirm result from 'y' key, got %d", result.Result)
	}
}

func TestConfirmDialog_Update_NKeyCancels(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	nKey := tea.KeyPressMsg{Text: "n"}
	_, cmd := dialog.Update(nKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(ConfirmDialogMsg)
	if !ok {
		t.Fatalf("expected ConfirmDialogMsg, got %T", msg)
	}

	if result.Result != ConfirmDialogResultCancel {
		t.Errorf("expected cancel result from 'n' key, got %d", result.Result)
	}
}

func TestConfirmDialog_Update_NotVisibleNoOp(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")
	dialog.Hide()

	// When not visible, updates should be no-ops
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	_, cmd := dialog.Update(tabKey)

	if cmd != nil {
		t.Error("expected no cmd when dialog is hidden")
	}
}

func TestConfirmDialog_View(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Delete Event", "Are you sure you want to delete this event?")
	dialog.SetSize(80, 40)

	view := dialog.View()

	// Check for expected content
	if !strings.Contains(view, "Delete Event") {
		t.Error("expected title 'Delete Event' in view")
	}
	if !strings.Contains(view, "Are you sure") {
		t.Error("expected message in view")
	}
	if !strings.Contains(view, "Confirm") {
		t.Error("expected 'Confirm' button in view")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("expected 'Cancel' button in view")
	}
	if !strings.Contains(view, "y: confirm") {
		t.Error("expected help text in view")
	}
}

func TestConfirmDialog_View_Hidden(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")
	dialog.Hide()

	view := dialog.View()

	if view != "" {
		t.Error("expected empty view when dialog is hidden")
	}
}

func TestConfirmDialog_View_CustomButtons(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")
	dialog.SetButtonLabels("Yes, Delete", "No, Keep")
	dialog.SetSize(80, 40)

	view := dialog.View()

	if !strings.Contains(view, "Yes, Delete") {
		t.Error("expected custom confirm button text in view")
	}
	if !strings.Contains(view, "No, Keep") {
		t.Error("expected custom cancel button text in view")
	}
}

func TestConfirmDialog_Init(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	cmd := dialog.Init()
	if cmd != nil {
		t.Error("expected Init to return nil (no initialization needed)")
	}
}

func TestConfirmDialog_VimKeysNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewConfirmDialog(theme, "Test", "Test message")

	// 'l' key should focus confirm
	lKey := tea.KeyPressMsg{Text: "l"}
	dialog, _ = dialog.Update(lKey)

	if dialog.focusedButton != 1 {
		t.Errorf("expected focus on confirm (1) after 'l' key, got %d", dialog.focusedButton)
	}

	// 'h' key should focus cancel
	hKey := tea.KeyPressMsg{Text: "h"}
	dialog, _ = dialog.Update(hKey)

	if dialog.focusedButton != 0 {
		t.Errorf("expected focus on cancel (0) after 'h' key, got %d", dialog.focusedButton)
	}
}
