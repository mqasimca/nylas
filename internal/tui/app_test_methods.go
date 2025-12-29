package tui

import (
	"testing"

	"github.com/rivo/tview"
)

func TestApp_GetCurrentView(t *testing.T) {
	app := createTestApp(t)

	// Initially should have dashboard view
	view := app.getCurrentView()
	if view == nil {
		t.Fatal("getCurrentView() returned nil")
	}

	// Default initial view is dashboard
	if view.Name() != "dashboard" {
		t.Errorf("getCurrentView().Name() = %q, want %q", view.Name(), "dashboard")
	}
}

func TestApp_Flash(t *testing.T) {
	app := createTestApp(t)

	// Test all flash levels
	flashLevels := []FlashLevel{FlashInfo, FlashWarn, FlashError}

	for _, fl := range flashLevels {
		// Should not panic
		app.Flash(fl, "Test message %s", "arg")
	}
}

func TestApp_ShowConfirmDialog(t *testing.T) {
	app := createTestApp(t)

	called := false
	app.ShowConfirmDialog("Test Title", "Test message", func() {
		called = true
	})

	// Dialog should be shown (pushed to content stack)
	// We can't easily test the callback without simulating UI interaction
	// but we can verify it doesn't panic
	if called {
		t.Error("callback should not be called immediately")
	}
}

func TestApp_PopDetail(t *testing.T) {
	app := createTestApp(t)

	// Pop on empty detail should not panic
	app.PopDetail()

	// Push something to content stack first using PushDetail
	box := tview.NewBox()
	app.PushDetail("test-detail", box)

	// Now pop should work
	app.PopDetail()
}

func TestApp_GoBack(t *testing.T) {
	app := createTestApp(t)

	// goBack is unexported, but we can test the exported PopDetail
	// which internally handles going back in detail views

	// Push a detail view first
	box := tview.NewBox()
	app.PushDetail("test-detail", box)

	// PopDetail should work
	app.PopDetail()
}

func TestApp_PageNavigation(t *testing.T) {
	// pageMove, goToTop, goToBottom, goToRow are unexported methods
	// They work on the internal table selection
	// We can only test that table navigation works via exported methods

	styles := DefaultStyles()
	table := NewTable(styles)

	table.SetColumns([]Column{{Title: "Name", Expand: true}})
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}, {"Item 3"}},
		[]RowMeta{{ID: "id-1"}, {ID: "id-2"}, {ID: "id-3"}},
	)

	// Test table selection
	table.Select(1, 0)
	row, _ := table.GetSelection()
	if row != 1 {
		t.Errorf("After Select(1, 0), row = %d, want 1", row)
	}

	table.Select(2, 0)
	row, _ = table.GetSelection()
	if row != 2 {
		t.Errorf("After Select(2, 0), row = %d, want 2", row)
	}
}

func TestApp_NavigateView(t *testing.T) {
	// navigateTo is unexported, but we can test views creation directly
	views := []string{"messages", "events", "contacts", "webhooks", "grants", "dashboard", "drafts", "availability"}

	for _, viewName := range views {
		t.Run(viewName, func(t *testing.T) {
			app := createTestApp(t)

			// createView is unexported, but views are created automatically
			// via command processing. Test that getCurrentView works after init.
			view := app.getCurrentView()
			if view == nil {
				t.Fatal("getCurrentView() returned nil")
			}

			// Verify default is dashboard
			if view.Name() != "dashboard" {
				t.Errorf("Default view = %q, want dashboard", view.Name())
			}
		})
	}
}

func TestApp_SetFocus(t *testing.T) {
	app := createTestApp(t)

	// SetFocus on a primitive should not panic
	box := tview.NewBox()
	app.SetFocus(box)
}

func TestApp_Styles(t *testing.T) {
	app := createTestApp(t)

	// Styles should return the styles
	styles := app.Styles()
	if styles == nil {
		t.Error("Styles() returned nil")
	}
}
