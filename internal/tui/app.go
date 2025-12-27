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
	GrantStore      ports.GrantStore // Optional: enables grant switching in TUI
	GrantID         string
	Email           string
	Provider        string
	RefreshInterval time.Duration
	InitialView     string    // Initial view to navigate to (messages, events, contacts, webhooks, grants)
	Theme           ThemeName // Theme name (k9s, amber, green, apple2, vintage, ibm, futuristic, matrix)
}

// App is the main TUI application using tview (like k9s).
type App struct {
	*tview.Application

	// Layout components (k9s style)
	main    *tview.Flex
	header  *tview.Flex
	logo    *Logo
	status  *StatusIndicator
	crumbs  *Crumbs
	menu    *Menu
	prompt  *Prompt         // For filter mode (/)
	palette *CommandPalette // For command mode (:) with autocomplete

	// Content area with page stack (like k9s)
	content *PageStack

	// Command registry for help and autocomplete
	cmdRegistry *CommandRegistry

	// State
	config      Config
	styles      *Styles
	running     bool
	mx          sync.RWMutex
	cmdActive   bool
	filterMode  bool
	lastKey     rune      // For vim-style 'gg' command
	lastKeyTime time.Time // Timeout for key sequences

	// View registry
	views map[string]ResourceView
}

// NewApp creates a new TUI application.
func NewApp(cfg Config) *App {
	// Use theme from config, default to k9s if not specified
	var styles *Styles
	if cfg.Theme != "" {
		styles = GetThemeStyles(cfg.Theme)
	} else {
		styles = DefaultStyles()
	}

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
	// Create command registry
	a.cmdRegistry = NewCommandRegistry()

	// Create components (k9s style)
	a.logo = NewLogo(a.styles)
	a.status = NewStatusIndicator(a.styles, a.config)
	a.crumbs = NewCrumbs(a.styles)
	a.menu = NewMenu(a.styles)
	a.prompt = NewPrompt(a.styles, a.onCommand, a.onFilter)
	a.palette = NewCommandPalette(a, a.cmdRegistry, a.onPaletteExecute, a.onPaletteCancel)
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

	// Initialize with specified view or dashboard
	initialView := a.config.InitialView
	if initialView == "" {
		initialView = "dashboard"
	}
	a.navigateTo(initialView)

	// Set root and enable mouse
	a.SetRoot(a.main, true)
	a.EnableMouse(true)
}

func (a *App) setupKeys() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If palette is active (command mode), let it handle input
		if a.cmdActive && a.palette.IsVisible() {
			// Palette handles its own input via SetInputCapture
			return event
		}

		// If filter mode is active, let prompt handle input
		if a.filterMode {
			return a.prompt.HandleKey(event)
		}

		// Check if we're in a detail view (compose, message-detail, etc.)
		// Detail views are pushed onto the stack but not registered in views map
		topPage := a.content.Top()
		currentView := a.getCurrentView()
		inDetailView := currentView == nil && a.content.Len() > 0

		// If in a detail view (compose, message-detail, help, etc.), only handle Escape and Ctrl+C
		if inDetailView {
			switch event.Key() {
			case tcell.KeyCtrlC:
				a.Stop()
				return nil
			case tcell.KeyEscape:
				// Pop the detail view
				a.content.Pop()
				// Re-focus the underlying view
				if view := a.getCurrentView(); view != nil {
					a.SetFocus(view.Primitive())
				}
				return nil
			}
			// Let the detail view handle all other keys (for typing in compose form)
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			// Quit with Ctrl+C
			a.Stop()
			return nil

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

		case tcell.KeyCtrlD:
			// Half page down (vim-style)
			a.pageMove(10, true)
			return nil

		case tcell.KeyCtrlU:
			// Half page up (vim-style)
			a.pageMove(10, false)
			return nil

		case tcell.KeyCtrlF:
			// Full page down (vim-style)
			a.pageMove(20, true)
			return nil

		case tcell.KeyCtrlB:
			// Full page up (vim-style)
			a.pageMove(20, false)
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case ':':
				// Enter command mode with palette (autocomplete)
				a.showPalette()
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

			case 'r':
				// Refresh (lowercase only - uppercase R is for reply)
				if currentView != nil {
					go func() {
						currentView.Refresh()
						a.QueueUpdateDraw(func() {})
					}()
				}
				return nil

			case 'g':
				// Handle 'gg' sequence for go-to-top (vim-style)
				now := time.Now()
				if a.lastKey == 'g' && now.Sub(a.lastKeyTime) < 500*time.Millisecond {
					// 'gg' pressed - go to top
					a.goToTop()
					a.lastKey = 0
					return nil
				}
				a.lastKey = 'g'
				a.lastKeyTime = now
				return nil

			case 'G':
				// Go to bottom (vim-style)
				a.goToBottom()
				return nil

			case 'd':
				// Handle 'dd' sequence for delete (vim-style)
				now := time.Now()
				if a.lastKey == 'd' && now.Sub(a.lastKeyTime) < 500*time.Millisecond {
					// 'dd' pressed - delete current item
					a.executeCommand("delete")
					a.lastKey = 0
					return nil
				}
				a.lastKey = 'd'
				a.lastKeyTime = now
				return nil

			case 'x':
				// Delete/archive current item (vim-style)
				a.executeCommand("delete")
				return nil
			}
		}

		// Pass to current view
		if currentView != nil {
			return currentView.HandleKey(event)
		}

		// Handle topPage for debugging
		_ = topPage

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

func (a *App) showPalette() {
	// Add palette to layout (before menu)
	a.main.RemoveItem(a.menu)
	a.main.AddItem(a.palette, 12, 0, true) // Give palette more height for dropdown
	a.main.AddItem(a.menu, 1, 0, false)
	a.palette.Show()
	a.cmdActive = true
}

func (a *App) hidePalette() {
	a.main.RemoveItem(a.palette)
	a.palette.Hide()
	a.cmdActive = false

	// Refocus current view
	if view := a.getCurrentView(); view != nil {
		a.SetFocus(view.Primitive())
	}
}

func (a *App) onPaletteExecute(cmd string) {
	a.hidePalette()
	if cmd != "" {
		a.onCommand(cmd)
	}
}

func (a *App) onPaletteCancel() {
	a.hidePalette()
}

func (a *App) onCommand(cmd string) {
	a.hidePrompt()

	if cmd == "" {
		return
	}

	// Handle numeric commands (go to row number)
	if isNumeric(cmd) {
		a.goToRow(parseInt(cmd))
		return
	}

	switch cmd {
	// Navigation - vim style
	case "m", "messages", "msg":
		a.navigateTo("messages")
	case "dr", "drafts":
		a.navigateTo("drafts")
	case "e", "events", "ev", "cal", "calendar":
		a.navigateTo("events")
	case "av", "avail", "availability":
		a.navigateTo("availability")
	case "c", "contacts", "ct":
		a.navigateTo("contacts")
	case "w", "webhooks", "wh":
		a.navigateTo("webhooks")
	case "ws", "webhook-server", "whs", "server":
		a.navigateTo("webhook-server")
	case "g", "grants", "gr":
		a.navigateTo("grants")
	case "i", "in", "inbound", "inbox":
		a.navigateTo("inbound")
	case "d", "dashboard", "dash", "home":
		a.navigateTo("dashboard")

	// Quit commands - vim style
	case "q", "quit", "exit":
		a.Stop()
	case "q!", "quit!":
		a.Stop() // Force quit
	case "wq", "x":
		a.Stop() // Write and quit (just quit for TUI)

	// Help - vim style
	case "h", "help":
		a.showHelp()

	// Actions on current item
	case "delete", "del", "rm":
		a.executeCommand("delete")
	case "star", "s":
		a.executeCommand("star")
	case "unstar":
		a.executeCommand("unstar")
	case "read", "mr":
		a.executeCommand("read")
	case "unread", "mu":
		a.executeCommand("unread")

	// Compose/Reply - vim style
	case "new", "n", "compose":
		a.executeCommand("compose")
	case "reply", "r":
		a.executeCommand("reply")
	case "replyall", "ra", "reply-all":
		a.executeCommand("replyall")
	case "forward", "f", "fwd":
		a.executeCommand("forward")

	// View commands
	case "refresh", "reload":
		if view := a.getCurrentView(); view != nil {
			go func() {
				view.Refresh()
				a.QueueUpdateDraw(func() {})
			}()
		}
	case "top", "first", "gg":
		a.goToTop()
	case "bottom", "last", "G":
		a.goToBottom()

	// Set commands (vim-style :set)
	default:
		// Check for :e <view> pattern (vim-style edit)
		if len(cmd) > 2 && cmd[:2] == "e " {
			viewName := cmd[2:]
			switch viewName {
			case "messages", "m":
				a.navigateTo("messages")
			case "drafts", "dr":
				a.navigateTo("drafts")
			case "events", "ev", "cal":
				a.navigateTo("events")
			case "availability", "av", "avail":
				a.navigateTo("availability")
			case "contacts", "c":
				a.navigateTo("contacts")
			case "webhooks", "w":
				a.navigateTo("webhooks")
			case "grants", "g":
				a.navigateTo("grants")
			}
		}
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
	case "drafts":
		return NewDraftsView(a)
	case "events":
		return NewEventsView(a)
	case "availability":
		return NewAvailabilityView(a)
	case "contacts":
		return NewContactsView(a)
	case "webhooks":
		return NewWebhooksView(a)
	case "webhook-server":
		return NewWebhookServerView(a)
	case "grants":
		return NewGrantsView(a)
	case "inbound":
		return NewInboundView(a)
	default:
		return NewDashboardView(a)
	}
}

func (a *App) showHelp() {
	// Create help view with callbacks
	onClose := func() {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}
	onExecute := func(cmd string) {
		a.executeCommand(cmd)
	}

	help := NewHelpView(a, a.cmdRegistry, onClose, onExecute)

	// Push help as a page
	a.content.Push("help", help)
	a.SetFocus(help)
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

		<-ticker.C
		a.QueueUpdateDraw(func() {
			a.status.Update()
		})
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
func (a *App) Flash(level FlashLevel, msg string, args ...any) {
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

// SwitchGrant switches to a different grant and updates the UI.
// Returns an error if GrantStore is not configured or the switch fails.
func (a *App) SwitchGrant(grantID, email, provider string) error {
	if a.config.GrantStore == nil {
		return fmt.Errorf("grant switching not available (no grant store)")
	}

	// Set the new default grant
	if err := a.config.GrantStore.SetDefaultGrant(grantID); err != nil {
		return fmt.Errorf("failed to switch grant: %w", err)
	}

	// Update config
	a.config.GrantID = grantID
	a.config.Email = email
	a.config.Provider = provider

	// Update status indicator
	a.status.UpdateGrant(email, provider, grantID)

	// Refresh the current view to load data for the new grant
	if view := a.getCurrentView(); view != nil {
		go func() {
			a.QueueUpdateDraw(func() {
				view.Refresh()
			})
		}()
	}

	return nil
}

// CanSwitchGrant returns true if grant switching is available.
func (a *App) CanSwitchGrant() bool {
	return a.config.GrantStore != nil
}

// ============================================================================
// Vim-style Navigation Helpers
// ============================================================================

// pageMove moves the selection up or down by the specified amount.
func (a *App) pageMove(lines int, down bool) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	// Get the table from the view if it's a table-based view
	if tableView, ok := view.Primitive().(*Table); ok {
		row, col := tableView.GetSelection()
		rowCount := tableView.GetRowCount()

		var newRow int
		if down {
			newRow = row + lines
			if newRow > rowCount {
				newRow = rowCount
			}
		} else {
			newRow = row - lines
			if newRow < 1 {
				newRow = 1 // Skip header
			}
		}

		tableView.Select(newRow, col)
	}
}

// goToTop moves selection to the first row.
func (a *App) goToTop() {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		tableView.Select(1, 0) // Row 1 is first data row (0 is header)
	}
}

// goToBottom moves selection to the last row.
func (a *App) goToBottom() {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		rowCount := tableView.GetRowCount()
		if rowCount > 0 {
			tableView.Select(rowCount, 0)
		}
	}
}

// goToRow moves selection to a specific row number (1-indexed).
func (a *App) goToRow(row int) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	if tableView, ok := view.Primitive().(*Table); ok {
		rowCount := tableView.GetRowCount()
		if row < 1 {
			row = 1
		}
		if row > rowCount {
			row = rowCount
		}
		tableView.Select(row, 0)
	}
}

// executeCommand executes a command on the current view/item.
func (a *App) executeCommand(cmd string) {
	view := a.getCurrentView()
	if view == nil {
		return
	}

	// Create a synthetic key event based on the command
	var event *tcell.EventKey

	switch cmd {
	case "delete":
		// For now, just flash a message - delete requires view-specific handling
		a.Flash(FlashWarn, "Delete not implemented for this view")
		return
	case "star":
		event = tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	case "unstar":
		event = tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	case "read":
		// Mark as read - no direct key, flash message
		a.Flash(FlashInfo, "Use 'u' to toggle unread status")
		return
	case "unread":
		event = tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone)
	case "compose":
		event = tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone)
	case "reply":
		event = tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
	case "replyall":
		event = tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone)
	case "forward":
		a.Flash(FlashInfo, "Forward: coming soon")
		return
	default:
		return
	}

	if event != nil {
		view.HandleKey(event)
	}
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseInt converts a string to int, returning 0 on error.
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
