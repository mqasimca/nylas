package tui

import "github.com/rivo/tview"

// Stack manages a stack of components (like k9s PageStack).
type Stack struct {
	items []string
}

// NewStack creates a new stack.
func NewStack() *Stack {
	return &Stack{
		items: make([]string, 0),
	}
}

// Push adds an item to the stack.
func (s *Stack) Push(name string) {
	s.items = append(s.items, name)
}

// Pop removes and returns the top item.
func (s *Stack) Pop() string {
	if len(s.items) == 0 {
		return ""
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

// Top returns the top item without removing it.
func (s *Stack) Top() string {
	if len(s.items) == 0 {
		return ""
	}
	return s.items[len(s.items)-1]
}

// Len returns the stack size.
func (s *Stack) Len() int {
	return len(s.items)
}

// Clear empties the stack.
func (s *Stack) Clear() {
	s.items = s.items[:0]
}

// PageStack combines tview.Pages with a navigation stack (like k9s).
type PageStack struct {
	*tview.Pages
	stack *Stack
}

// NewPageStack creates a new page stack.
func NewPageStack() *PageStack {
	return &PageStack{
		Pages: tview.NewPages(),
		stack: NewStack(),
	}
}

// Push adds a page and shows it.
func (p *PageStack) Push(name string, page tview.Primitive) {
	p.stack.Push(name)
	p.AddPage(name, page, true, true)
}

// Pop removes the top page and shows the previous one.
func (p *PageStack) Pop() string {
	name := p.stack.Pop()
	if name != "" {
		p.RemovePage(name)
	}
	// Show the new top
	if top := p.stack.Top(); top != "" {
		p.SwitchToPage(top)
	}
	return name
}

// Top returns the current page name.
func (p *PageStack) Top() string {
	return p.stack.Top()
}

// Len returns the stack depth.
func (p *PageStack) Len() int {
	return p.stack.Len()
}

// SwitchTo shows a page (adds to stack if not top).
func (p *PageStack) SwitchTo(name string, page tview.Primitive) {
	// If already exists, just switch
	if p.HasPage(name) {
		p.SwitchToPage(name)
		// Update stack - remove if exists and push to top
		newItems := make([]string, 0)
		for _, item := range p.stack.items {
			if item != name {
				newItems = append(newItems, item)
			}
		}
		p.stack.items = append(newItems, name)
	} else {
		p.Push(name, page)
	}
}
