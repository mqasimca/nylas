package tui

import (
	"testing"
)

func TestNewStack(t *testing.T) {
	stack := NewStack()

	if stack == nil {
		t.Fatal("NewStack() returned nil")
	}

	if stack.Len() != 0 {
		t.Errorf("New stack Len() = %d, want 0", stack.Len())
	}

	if stack.Top() != "" {
		t.Errorf("New stack Top() = %q, want empty", stack.Top())
	}
}

func TestStack_Push(t *testing.T) {
	stack := NewStack()

	stack.Push("item1")
	if stack.Len() != 1 {
		t.Errorf("After Push, Len() = %d, want 1", stack.Len())
	}
	if stack.Top() != "item1" {
		t.Errorf("After Push, Top() = %q, want %q", stack.Top(), "item1")
	}

	stack.Push("item2")
	if stack.Len() != 2 {
		t.Errorf("After second Push, Len() = %d, want 2", stack.Len())
	}
	if stack.Top() != "item2" {
		t.Errorf("After second Push, Top() = %q, want %q", stack.Top(), "item2")
	}

	stack.Push("item3")
	if stack.Len() != 3 {
		t.Errorf("After third Push, Len() = %d, want 3", stack.Len())
	}
	if stack.Top() != "item3" {
		t.Errorf("After third Push, Top() = %q, want %q", stack.Top(), "item3")
	}
}

func TestStack_Pop(t *testing.T) {
	stack := NewStack()

	// Pop on empty stack
	result := stack.Pop()
	if result != "" {
		t.Errorf("Pop on empty stack = %q, want empty", result)
	}

	// Push and pop
	stack.Push("item1")
	stack.Push("item2")
	stack.Push("item3")

	result = stack.Pop()
	if result != "item3" {
		t.Errorf("Pop() = %q, want %q", result, "item3")
	}
	if stack.Len() != 2 {
		t.Errorf("After Pop, Len() = %d, want 2", stack.Len())
	}

	result = stack.Pop()
	if result != "item2" {
		t.Errorf("Pop() = %q, want %q", result, "item2")
	}
	if stack.Len() != 1 {
		t.Errorf("After Pop, Len() = %d, want 1", stack.Len())
	}

	result = stack.Pop()
	if result != "item1" {
		t.Errorf("Pop() = %q, want %q", result, "item1")
	}
	if stack.Len() != 0 {
		t.Errorf("After Pop, Len() = %d, want 0", stack.Len())
	}

	// Pop on now-empty stack
	result = stack.Pop()
	if result != "" {
		t.Errorf("Pop on empty stack = %q, want empty", result)
	}
}

func TestStack_Top(t *testing.T) {
	stack := NewStack()

	// Top on empty stack
	if stack.Top() != "" {
		t.Errorf("Top on empty stack = %q, want empty", stack.Top())
	}

	stack.Push("item1")
	if stack.Top() != "item1" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "item1")
	}

	stack.Push("item2")
	if stack.Top() != "item2" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "item2")
	}

	// Top doesn't remove item
	if stack.Len() != 2 {
		t.Errorf("After Top(), Len() = %d, want 2", stack.Len())
	}
}

func TestStack_Len(t *testing.T) {
	stack := NewStack()

	if stack.Len() != 0 {
		t.Errorf("Empty stack Len() = %d, want 0", stack.Len())
	}

	for i := 1; i <= 5; i++ {
		stack.Push("item")
		if stack.Len() != i {
			t.Errorf("After %d pushes, Len() = %d, want %d", i, stack.Len(), i)
		}
	}
}

func TestStack_Clear(t *testing.T) {
	stack := NewStack()

	// Clear empty stack
	stack.Clear()
	if stack.Len() != 0 {
		t.Errorf("After Clear on empty stack, Len() = %d, want 0", stack.Len())
	}

	// Add items then clear
	stack.Push("item1")
	stack.Push("item2")
	stack.Push("item3")

	if stack.Len() != 3 {
		t.Errorf("Before Clear, Len() = %d, want 3", stack.Len())
	}

	stack.Clear()
	if stack.Len() != 0 {
		t.Errorf("After Clear, Len() = %d, want 0", stack.Len())
	}

	if stack.Top() != "" {
		t.Errorf("After Clear, Top() = %q, want empty", stack.Top())
	}

	if stack.Pop() != "" {
		t.Errorf("After Clear, Pop() = %q, want empty", stack.Pop())
	}
}

func TestStack_PushPopSequence(t *testing.T) {
	stack := NewStack()

	// Interleaved push and pop
	stack.Push("a")
	stack.Push("b")
	if stack.Pop() != "b" {
		t.Error("Expected 'b'")
	}
	stack.Push("c")
	stack.Push("d")
	if stack.Pop() != "d" {
		t.Error("Expected 'd'")
	}
	if stack.Pop() != "c" {
		t.Error("Expected 'c'")
	}
	if stack.Pop() != "a" {
		t.Error("Expected 'a'")
	}
	if stack.Pop() != "" {
		t.Error("Expected empty string")
	}
}
