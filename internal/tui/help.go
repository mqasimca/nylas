package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpView displays all available commands organized by category.
// It supports searching/filtering and executing commands directly.
type HelpView struct {
	*tview.Flex
	app          *App
	registry     *CommandRegistry
	table        *tview.Table
	searchInput  *tview.InputField
	filter       string
	categoryRows map[int]bool    // Rows that are category headers (not selectable)
	commandRows  map[int]Command // Map row index to command
	searching    bool            // Whether search input is active
	onClose      func()          // Callback when help is closed
	onExecute    func(string)    // Callback to execute a command
}

// NewHelpView creates a new help view with the command registry.
func NewHelpView(app *App, registry *CommandRegistry, onClose func(), onExecute func(string)) *HelpView {
	h := &HelpView{
		Flex:         tview.NewFlex(),
		app:          app,
		registry:     registry,
		categoryRows: make(map[int]bool),
		commandRows:  make(map[int]Command),
		onClose:      onClose,
		onExecute:    onExecute,
	}

	h.init()
	return h
}

func (h *HelpView) init() {
	styles := h.app.styles

	// Create search input
	h.searchInput = tview.NewInputField()
	h.searchInput.SetLabel(" / ")
	h.searchInput.SetLabelColor(styles.FilterColor)
	h.searchInput.SetFieldBackgroundColor(styles.BgColor)
	h.searchInput.SetFieldTextColor(styles.FgColor)
	h.searchInput.SetBackgroundColor(styles.BgColor)
	h.searchInput.SetPlaceholder("Filter commands...")
	h.searchInput.SetPlaceholderTextColor(styles.BorderColor)
	h.searchInput.SetChangedFunc(func(text string) {
		h.filter = text
		h.render()
	})

	// Create table for commands
	h.table = tview.NewTable()
	h.table.SetBackgroundColor(styles.BgColor)
	h.table.SetBorderPadding(0, 0, 1, 1)
	h.table.SetSelectable(true, false)
	h.table.SetSelectedStyle(tcell.StyleDefault.
		Background(styles.FocusColor).
		Foreground(styles.BgColor))

	// Handle table selection change - skip category headers
	h.table.SetSelectionChangedFunc(func(row, _ int) {
		if h.categoryRows[row] {
			// Skip to next selectable row
			if row < h.table.GetRowCount()-1 {
				h.table.Select(row+1, 0)
			} else if row > 0 {
				h.table.Select(row-1, 0)
			}
		}
	})

	// Create container with border
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetBorder(true)
	container.SetBorderColor(styles.FocusColor)
	container.SetTitle(" Help ")
	container.SetTitleColor(styles.TitleFg)
	container.SetBackgroundColor(styles.BgColor)

	// Add search input and table
	container.AddItem(h.searchInput, 1, 0, false)
	container.AddItem(h.table, 0, 1, true)

	// Footer with hints
	footer := tview.NewTextView()
	footer.SetDynamicColors(true)
	footer.SetBackgroundColor(styles.BgColor)
	keyHex := styles.Hex(styles.MenuKeyFg)
	_, _ = fmt.Fprintf(footer, " [%s]j/k[-] navigate  [%s]/[-] search  [%s]Enter[-] execute  [%s]Esc[-] close",
		keyHex, keyHex, keyHex, keyHex)

	container.AddItem(footer, 1, 0, false)

	h.SetDirection(tview.FlexRow)
	h.AddItem(container, 0, 1, true)

	// Initial render
	h.render()

	// Set up input capture
	h.SetInputCapture(h.handleInput)
}

func (h *HelpView) render() {
	h.table.Clear()
	h.categoryRows = make(map[int]bool)
	h.commandRows = make(map[int]Command)

	styles := h.app.styles
	row := 0

	// Get commands to display
	var groups []CategoryGroup
	if h.filter == "" {
		groups = h.registry.GetByCategory()
	} else {
		// Filter commands and group results
		matched := h.registry.Search(h.filter)
		if len(matched) > 0 {
			// Group filtered results by category
			byCategory := make(map[CommandCategory][]Command)
			for _, cmd := range matched {
				byCategory[cmd.Category] = append(byCategory[cmd.Category], cmd)
			}
			for _, cat := range categoryOrder {
				if cmds, ok := byCategory[cat]; ok {
					groups = append(groups, CategoryGroup{Category: cat, Commands: cmds})
				}
			}
		}
	}

	titleColor := styles.Hex(styles.TitleFg)
	cmdColor := styles.Hex(styles.InfoColor)
	aliasColor := styles.Hex(styles.BorderColor)
	descColor := styles.Hex(styles.FgColor)

	for _, group := range groups {
		// Category header
		h.table.SetCell(row, 0,
			tview.NewTableCell(fmt.Sprintf("[%s::b]%s[-::-]", titleColor, string(group.Category))).
				SetSelectable(false).
				SetExpansion(1))
		h.categoryRows[row] = true
		row++

		// Commands in category
		for _, cmd := range group.Commands {
			// Format: :command, :alias    Description    [shortcut]
			cmdText := fmt.Sprintf("[%s]:%s[-]", cmdColor, cmd.Name)
			if len(cmd.Aliases) > 0 {
				cmdText += fmt.Sprintf("[%s], :%s[-]", aliasColor, strings.Join(cmd.Aliases, ", :"))
			}

			descText := fmt.Sprintf("[%s]%s[-]", descColor, cmd.Description)

			shortcutText := ""
			if cmd.Shortcut != "" {
				shortcutText = fmt.Sprintf("[%s]<%s>[-]", aliasColor, cmd.Shortcut)
			}

			h.table.SetCell(row, 0,
				tview.NewTableCell(fmt.Sprintf("  %-35s %s %s", cmdText, descText, shortcutText)).
					SetExpansion(1))
			h.commandRows[row] = cmd
			row++
		}

		// Empty row between categories
		if row > 0 {
			h.table.SetCell(row, 0, tview.NewTableCell("").SetSelectable(false))
			h.categoryRows[row] = true
			row++
		}
	}

	// Select first command row
	for i := 0; i < h.table.GetRowCount(); i++ {
		if !h.categoryRows[i] {
			h.table.Select(i, 0)
			break
		}
	}
}

func (h *HelpView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if h.searching {
		return h.handleSearchInput(event)
	}

	switch event.Key() {
	case tcell.KeyEscape:
		if h.filter != "" {
			// Clear filter first
			h.filter = ""
			h.searchInput.SetText("")
			h.render()
			return nil
		}
		if h.onClose != nil {
			h.onClose()
		}
		return nil

	case tcell.KeyEnter:
		// Execute selected command
		row, _ := h.table.GetSelection()
		if cmd, ok := h.commandRows[row]; ok {
			if h.onExecute != nil {
				h.onExecute(cmd.Name)
			}
			if h.onClose != nil {
				h.onClose()
			}
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			// Enter search mode
			h.searching = true
			h.app.SetFocus(h.searchInput)
			return nil

		case 'q':
			if h.onClose != nil {
				h.onClose()
			}
			return nil

		case 'j':
			// Move down
			h.moveSelection(1)
			return nil

		case 'k':
			// Move up
			h.moveSelection(-1)
			return nil

		case 'g':
			// Go to top (would need double-g detection, skip for now)
			return event

		case 'G':
			// Go to bottom
			h.goToBottom()
			return nil
		}

	case tcell.KeyDown:
		h.moveSelection(1)
		return nil

	case tcell.KeyUp:
		h.moveSelection(-1)
		return nil
	}

	return event
}

func (h *HelpView) handleSearchInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		h.searching = false
		h.app.SetFocus(h.table)
		return nil

	case tcell.KeyEnter:
		h.searching = false
		h.app.SetFocus(h.table)
		return nil

	case tcell.KeyDown:
		h.searching = false
		h.app.SetFocus(h.table)
		h.moveSelection(1)
		return nil

	case tcell.KeyUp:
		h.searching = false
		h.app.SetFocus(h.table)
		h.moveSelection(-1)
		return nil
	}

	return event
}

func (h *HelpView) moveSelection(delta int) {
	row, _ := h.table.GetSelection()
	newRow := row + delta

	// Skip category headers
	for newRow >= 0 && newRow < h.table.GetRowCount() {
		if !h.categoryRows[newRow] {
			h.table.Select(newRow, 0)
			return
		}
		newRow += delta
	}
}

func (h *HelpView) goToBottom() {
	for i := h.table.GetRowCount() - 1; i >= 0; i-- {
		if !h.categoryRows[i] {
			h.table.Select(i, 0)
			return
		}
	}
}

// Focus sets focus to the table.
func (h *HelpView) Focus(delegate func(p tview.Primitive)) {
	delegate(h.table)
}
