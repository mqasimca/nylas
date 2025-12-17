// Package tui provides a k9s-style terminal user interface for Nylas.
package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/rivo/tview"
)

// Config holds the TUI configuration.
type Config struct {
	Client          ports.NylasClient
	GrantID         string
	Email           string
	Provider        string
	RefreshInterval time.Duration
}

// App is the main TUI application using tview (like k9s).
type App struct {
	*tview.Application

	// Layout components (k9s style)
	main   *tview.Flex
	header *tview.Flex
	logo   *Logo
	status *StatusIndicator
	crumbs *Crumbs
	menu   *Menu
	prompt *Prompt

	// Content area with page stack (like k9s)
	content *PageStack

	// State
	config       Config
	styles       *Styles
	running      bool
	mx           sync.RWMutex
	cmdActive    bool
	filterMode   bool

	// View registry
	views map[string]ResourceView
}

// NewApp creates a new TUI application.
func NewApp(cfg Config) *App {
	styles := DefaultStyles()

	app := &App{
		Application: tview.NewApplication(),
		config:      cfg,
		styles:      styles,
		views:       make(map[string]ResourceView),
	}

	app.init()
	return app
}

func (a *App) init() {
	// Create components (k9s style)
	a.logo = NewLogo(a.styles)
	a.status = NewStatusIndicator(a.styles, a.config)
	a.crumbs = NewCrumbs(a.styles)
	a.menu = NewMenu(a.styles)
	a.prompt = NewPrompt(a.styles, a.onCommand, a.onFilter)
	a.content = NewPageStack()

	// Header: Logo on left, Status on right (like k9s)
	a.header = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.logo, 12, 0, false).
		AddItem(a.status, 0, 1, false)

	// Main layout (vertical flex - like k9s)
	// Layout: Header -> Crumbs -> Content -> Menu
	a.main = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 1, 0, false).
		AddItem(a.crumbs, 1, 0, false).
		AddItem(a.content, 0, 1, true). // Content takes remaining space
		AddItem(a.menu, 1, 0, false)

	// Set up key bindings
	a.setupKeys()

	// Initialize with dashboard
	a.navigateTo("dashboard")

	// Set root and enable mouse
	a.SetRoot(a.main, true)
	a.EnableMouse(true)
}

func (a *App) setupKeys() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If prompt is active, let it handle input
		if a.cmdActive || a.filterMode {
			return a.prompt.HandleKey(event)
		}

		// Get current view
		currentView := a.getCurrentView()

		switch event.Key() {
		case tcell.KeyEscape:
			// First, let the current view handle Escape (for closing details, etc.)
			if currentView != nil {
				result := currentView.HandleKey(event)
				if result == nil {
					// View handled the Escape
					return nil
				}
			}

			// If view didn't handle it, go back in navigation
			return a.goBack()

		case tcell.KeyRune:
			switch event.Rune() {
			case ':':
				// Enter command mode
				a.cmdActive = true
				a.prompt.Activate(PromptCommand)
				a.showPrompt()
				return nil

			case '/':
				// Enter filter mode
				a.filterMode = true
				a.prompt.Activate(PromptFilter)
				a.showPrompt()
				return nil

			case '?':
				// Show help
				a.showHelp()
				return nil

			case 'q':
				// Quit (only if not in a sub-view)
				if a.content.Len() <= 1 {
					a.Stop()
					return nil
				}
				// Otherwise treat as back
				return a.goBack()

			case 'r':
				// Refresh
				if currentView != nil {
					go func() {
						currentView.Refresh()
						a.QueueUpdateDraw(func() {})
					}()
				}
				return nil
			}
		}

		// Pass to current view
		if currentView != nil {
			return currentView.HandleKey(event)
		}

		return event
	})
}

func (a *App) getCurrentView() ResourceView {
	name := a.content.Top()
	if view, ok := a.views[name]; ok {
		return view
	}
	return nil
}

func (a *App) showPrompt() {
	// Add prompt to layout (before menu)
	a.main.RemoveItem(a.menu)
	a.main.AddItem(a.prompt, 1, 0, true)
	a.main.AddItem(a.menu, 1, 0, false)
	a.SetFocus(a.prompt)
}

func (a *App) hidePrompt() {
	a.main.RemoveItem(a.prompt)
	a.cmdActive = false
	a.filterMode = false

	// Refocus current view
	if view := a.getCurrentView(); view != nil {
		a.SetFocus(view.Primitive())
	}
}

func (a *App) onCommand(cmd string) {
	a.hidePrompt()

	if cmd == "" {
		return
	}

	switch cmd {
	case "m", "messages", "msg":
		a.navigateTo("messages")
	case "e", "events", "ev":
		a.navigateTo("events")
	case "c", "contacts", "ct":
		a.navigateTo("contacts")
	case "w", "webhooks", "wh":
		a.navigateTo("webhooks")
	case "g", "grants", "gr":
		a.navigateTo("grants")
	case "d", "dashboard", "dash":
		a.navigateTo("dashboard")
	case "q", "quit":
		a.Stop()
	}
}

func (a *App) onFilter(filter string) {
	a.hidePrompt()
	if view := a.getCurrentView(); view != nil {
		view.Filter(filter)
		go func() {
			view.Refresh()
			a.QueueUpdateDraw(func() {})
		}()
	}
}

func (a *App) navigateTo(name string) {
	view, ok := a.views[name]
	if !ok {
		view = a.createView(name)
		a.views[name] = view
	}

	// Use page stack for navigation
	a.content.SwitchTo(name, view.Primitive())

	// Update UI
	a.crumbs.SetPath(view.Title())
	a.menu.SetHints(view.Hints())
	a.SetFocus(view.Primitive())

	// Load data asynchronously
	go func() {
		view.Load()
		a.QueueUpdateDraw(func() {})
	}()
}

func (a *App) goBack() *tcell.EventKey {
	// Need at least 2 items to go back (current + previous)
	if a.content.Len() <= 1 {
		return nil
	}

	// Pop current
	a.content.Pop()

	// Update UI for new top
	name := a.content.Top()
	if view, ok := a.views[name]; ok {
		a.crumbs.SetPath(view.Title())
		a.menu.SetHints(view.Hints())
		a.SetFocus(view.Primitive())
	}

	return nil
}

func (a *App) createView(name string) ResourceView {
	switch name {
	case "messages":
		return NewMessagesView(a)
	case "events":
		return NewEventsView(a)
	case "contacts":
		return NewContactsView(a)
	case "webhooks":
		return NewWebhooksView(a)
	case "grants":
		return NewGrantsView(a)
	default:
		return NewDashboardView(a)
	}
}

func (a *App) showHelp() {
	help := NewHelpView(a.styles)

	// Push help as a page
	a.content.Push("help", help)
	a.SetFocus(help)

	// Close on any key
	help.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
		return nil
	})
}

// PushDetail pushes a detail view onto the stack (for message detail, etc.)
func (a *App) PushDetail(name string, view tview.Primitive) {
	a.content.Push(name, view)
	a.SetFocus(view)
}

// PopDetail pops a detail view from the stack
func (a *App) PopDetail() {
	if a.content.Len() > 1 {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}
}

// Run starts the application.
func (a *App) Run() error {
	a.mx.Lock()
	a.running = true
	a.mx.Unlock()

	// Start status update ticker
	go a.statusTicker()

	return a.Application.Run()
}

func (a *App) statusTicker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		a.mx.RLock()
		running := a.running
		a.mx.RUnlock()

		if !running {
			return
		}

		select {
		case <-ticker.C:
			a.QueueUpdateDraw(func() {
				a.status.Update()
			})
		}
	}
}

// Stop stops the application.
func (a *App) Stop() {
	a.mx.Lock()
	a.running = false
	a.mx.Unlock()
	a.Application.Stop()
}

// Flash displays a temporary message.
func (a *App) Flash(level FlashLevel, msg string, args ...interface{}) {
	a.status.Flash(level, fmt.Sprintf(msg, args...))
}

// Styles returns the app styles.
func (a *App) Styles() *Styles {
	return a.styles
}

// Config returns the app config.
func (a *App) GetConfig() Config {
	return a.config
}
