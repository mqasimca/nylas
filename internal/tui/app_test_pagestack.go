package tui

import (
	"testing"

	"github.com/rivo/tview"
)

func TestPageStack(t *testing.T) {
	stack := NewPageStack()

	if stack == nil {
		t.Fatal("NewPageStack() returned nil")
	}

	// Test empty stack
	if stack.Len() != 0 {
		t.Errorf("New stack Len() = %d, want 0", stack.Len())
	}
	if stack.Top() != "" {
		t.Errorf("Empty stack Top() = %q, want empty string", stack.Top())
	}
}

func TestPageStack_PushPop(t *testing.T) {
	stack := NewPageStack()

	// Create dummy primitives
	box1 := tview.NewBox()
	box2 := tview.NewBox()
	box3 := tview.NewBox()

	// Push items
	stack.Push("page1", box1)
	stack.Push("page2", box2)
	stack.Push("page3", box3)

	if stack.Len() != 3 {
		t.Errorf("Len() = %d, want 3", stack.Len())
	}

	if stack.Top() != "page3" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "page3")
	}

	// Pop items
	popped := stack.Pop()
	if popped != "page3" {
		t.Errorf("Pop() = %q, want %q", popped, "page3")
	}

	if stack.Len() != 2 {
		t.Errorf("After pop, Len() = %d, want 2", stack.Len())
	}

	// Pop remaining
	stack.Pop()
	stack.Pop()

	// Pop on empty should return empty string
	popped = stack.Pop()
	if popped != "" {
		t.Errorf("Pop() on empty stack = %q, want empty string", popped)
	}
}

func TestPageStack_HasPage(t *testing.T) {
	stack := NewPageStack()

	box1 := tview.NewBox()
	box2 := tview.NewBox()

	stack.Push("page1", box1)
	stack.Push("page2", box2)

	// PageStack embeds tview.Pages which has HasPage method
	if !stack.HasPage("page1") {
		t.Error("HasPage(page1) = false, want true")
	}

	if !stack.HasPage("page2") {
		t.Error("HasPage(page2) = false, want true")
	}

	if stack.HasPage("page3") {
		t.Error("HasPage(page3) = true, want false")
	}
}

func TestPageStack_SwitchTo(t *testing.T) {
	stack := NewPageStack()

	// Push pages
	box1 := tview.NewBox()
	box2 := tview.NewBox()

	stack.Push("page1", box1)
	stack.Push("page2", box2)

	if stack.Len() != 2 {
		t.Errorf("Len() = %d, want 2", stack.Len())
	}

	// Top should return current page name
	if stack.Top() != "page2" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "page2")
	}

	// SwitchTo an existing page
	stack.SwitchTo("page1", box1)
	if stack.Top() != "page1" {
		t.Errorf("After SwitchTo(page1), Top() = %q, want page1", stack.Top())
	}

	// SwitchTo a new page
	box3 := tview.NewBox()
	stack.SwitchTo("page3", box3)
	if stack.Top() != "page3" {
		t.Errorf("After SwitchTo(page3), Top() = %q, want page3", stack.Top())
	}

	// Pop
	stack.Pop()
	if stack.Top() != "page1" {
		t.Errorf("After pop, Top() = %q, want page1", stack.Top())
	}
}
