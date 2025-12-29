// Package models provides screen models for the TUI.
package models

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

const helpMarkdown = `# Nylas CLI - Keyboard Shortcuts

## Global Commands

- **Ctrl+C** - Quit the application
- **Esc** - Go back to previous screen
- **?** - Show this help screen

## Dashboard

- **a** - Open Air (Messages view)
- **c** - Open Calendar view
- **p** - Open Contacts (People) view
- **s** - Open Settings
- **?** - Show help

## Messages View

- **j/k** or **↓/↑** - Navigate messages
- **Enter** - Open selected message
- **c** - Compose new message
- **r** - Reply to message
- **f** - Forward message
- **d** - Delete message
- **/** - Search messages
- **Esc** - Back to dashboard

## Calendar View

- **n** - Create new event
- **e** - Edit selected event
- **d** - Delete selected event
- **j/k** or **↓/↑** - Navigate events
- **Enter** - View event details
- **Esc** - Back to dashboard

## Compose Screen

- **Tab** - Switch between fields
- **Ctrl+S** - Send message
- **Ctrl+A** - Add attachment
- **Esc** - Cancel and go back

## Message Detail

- **r** - Reply
- **a** - Reply all
- **f** - Forward
- **d** - Delete
- **j/k** or **↓/↑** - Scroll content
- **Esc** - Back to messages

## Settings

- **Tab** - Navigate form fields
- **Enter** - Save settings
- **Esc** - Cancel and go back

---

**Version:** Nylas CLI v2.0
**TUI Engine:** Bubble Tea
**Theme:** K9s (default)

Press **Esc** to return to the previous screen.
`

// HelpScreen shows keyboard shortcuts and help information.
type HelpScreen struct {
	global   *state.GlobalState
	theme    *styles.Theme
	viewport viewport.Model
	content  string
	ready    bool
}

// NewHelpScreen creates a new help screen.
func NewHelpScreen(global *state.GlobalState) *HelpScreen {
	theme := styles.GetTheme(global.Theme)

	// Render markdown with glamour
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	content, _ := renderer.Render(helpMarkdown)

	return &HelpScreen{
		global:  global,
		theme:   theme,
		content: content,
		ready:   false,
	}
}

// Init implements tea.Model.
func (h *HelpScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (h *HelpScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.global.SetWindowSize(msg.Width, msg.Height)

		if !h.ready {
			vp := viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-4),
			)
			vp.SetContent(h.content)
			h.viewport = vp
			h.ready = true
		} else {
			h.viewport.SetWidth(msg.Width)
			h.viewport.SetHeight(msg.Height - 4)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return h, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return h, tea.Quit
		}
	}

	// Update viewport
	var cmd tea.Cmd
	h.viewport, cmd = h.viewport.Update(msg)
	return h, cmd
}

// View implements tea.Model.
func (h *HelpScreen) View() tea.View {
	if !h.ready {
		return tea.NewView("Loading help...")
	}

	// Header
	header := h.theme.Title.Render("Help - Keyboard Shortcuts")

	// Footer
	footer := h.theme.Help.Render("Press Esc to go back")

	// Combine
	content := header + "\n" + h.viewport.View() + "\n" + footer

	return tea.NewView(content)
}
