package tui

import (
	"sort"
	"strings"
)

// CommandCategory groups related commands for organization in help and palette.
type CommandCategory string

const (
	CategoryNavigation CommandCategory = "Navigation"
	CategoryMessages   CommandCategory = "Messages"
	CategoryCalendar   CommandCategory = "Calendar"
	CategoryContacts   CommandCategory = "Contacts"
	CategoryWebhooks   CommandCategory = "Webhooks"
	CategoryFolders    CommandCategory = "Folders"
	CategoryVim        CommandCategory = "Vim Commands"
	CategorySystem     CommandCategory = "System"
)

// categoryOrder defines the display order for categories.
var categoryOrder = []CommandCategory{
	CategoryNavigation,
	CategoryMessages,
	CategoryCalendar,
	CategoryContacts,
	CategoryWebhooks,
	CategoryFolders,
	CategoryVim,
	CategorySystem,
}

// Command represents a TUI command with metadata.
type Command struct {
	Name        string          // Primary command name (e.g., "messages")
	Aliases     []string        // Short aliases (e.g., ["m", "msg"])
	Description string          // Human-readable description
	Category    CommandCategory // For grouping in help/palette
	Shortcut    string          // Direct key shortcut if any (e.g., "n" for compose)
	SubCommands []Command       // Nested sub-commands
	ContextView string          // View where this command is available ("" = all views)
}

// AllNames returns all names including aliases for this command.
func (c Command) AllNames() []string {
	names := []string{c.Name}
	names = append(names, c.Aliases...)
	return names
}

// DisplayAliases returns a formatted string of aliases for display.
func (c Command) DisplayAliases() string {
	if len(c.Aliases) == 0 {
		return ""
	}
	return strings.Join(c.Aliases, ", ")
}

// CommandRegistry holds all registered commands and provides lookup methods.
type CommandRegistry struct {
	commands   []Command
	byName     map[string]*Command           // Lookup by name or alias
	byCategory map[CommandCategory][]Command // Grouped by category
}

// NewCommandRegistry creates a new registry with all TUI commands registered.
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{
		commands:   make([]Command, 0),
		byName:     make(map[string]*Command),
		byCategory: make(map[CommandCategory][]Command),
	}

	r.registerAllCommands()
	return r
}

// Register adds a command to the registry.
func (r *CommandRegistry) Register(cmd Command) {
	r.commands = append(r.commands, cmd)

	// Index by primary name
	r.byName[cmd.Name] = &r.commands[len(r.commands)-1]

	// Index by aliases
	for _, alias := range cmd.Aliases {
		r.byName[alias] = &r.commands[len(r.commands)-1]
	}

	// Group by category
	r.byCategory[cmd.Category] = append(r.byCategory[cmd.Category], cmd)

	// Register sub-commands with parent prefix
	for _, sub := range cmd.SubCommands {
		subCmd := Command{
			Name:        cmd.Name + " " + sub.Name,
			Aliases:     make([]string, 0, len(sub.Aliases)),
			Description: sub.Description,
			Category:    cmd.Category,
			Shortcut:    sub.Shortcut,
			ContextView: sub.ContextView,
		}
		// Create aliased versions
		for _, subAlias := range sub.Aliases {
			subCmd.Aliases = append(subCmd.Aliases, cmd.Name+" "+subAlias)
		}
		r.commands = append(r.commands, subCmd)
		r.byName[subCmd.Name] = &r.commands[len(r.commands)-1]
		for _, alias := range subCmd.Aliases {
			r.byName[alias] = &r.commands[len(r.commands)-1]
		}
	}
}

// Get returns a command by name or alias, or nil if not found.
func (r *CommandRegistry) Get(name string) *Command {
	return r.byName[strings.ToLower(strings.TrimSpace(name))]
}

// GetAll returns all top-level commands sorted alphabetically.
func (r *CommandRegistry) GetAll() []Command {
	result := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		// Skip sub-commands (they contain spaces)
		if !strings.Contains(cmd.Name, " ") {
			result = append(result, cmd)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// GetByCategory returns commands grouped by category in display order.
func (r *CommandRegistry) GetByCategory() []CategoryGroup {
	result := make([]CategoryGroup, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		if cmds, ok := r.byCategory[cat]; ok && len(cmds) > 0 {
			result = append(result, CategoryGroup{
				Category: cat,
				Commands: cmds,
			})
		}
	}
	return result
}

// CategoryGroup holds a category and its commands.
type CategoryGroup struct {
	Category CommandCategory
	Commands []Command
}

// GetSubCommands returns sub-commands for a parent command.
func (r *CommandRegistry) GetSubCommands(parent string) []Command {
	parent = strings.ToLower(strings.TrimSpace(parent))
	prefix := parent + " "
	result := make([]Command, 0)

	for _, cmd := range r.commands {
		if strings.HasPrefix(cmd.Name, prefix) {
			// Extract just the sub-command part
			subName := strings.TrimPrefix(cmd.Name, prefix)
			if !strings.Contains(subName, " ") { // Only direct children
				result = append(result, Command{
					Name:        subName,
					Aliases:     extractSubAliases(cmd.Aliases, prefix),
					Description: cmd.Description,
					Category:    cmd.Category,
					Shortcut:    cmd.Shortcut,
				})
			}
		}
	}
	return result
}

// extractSubAliases extracts the sub-command part from aliased names.
func extractSubAliases(aliases []string, prefix string) []string {
	result := make([]string, 0)
	for _, alias := range aliases {
		if strings.HasPrefix(alias, prefix) {
			result = append(result, strings.TrimPrefix(alias, prefix))
		}
	}
	return result
}

// Search returns commands matching the query using fuzzy matching.
// Results are sorted by relevance (exact match > prefix match > contains).
func (r *CommandRegistry) Search(query string) []Command {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return r.GetAll()
	}

	type scored struct {
		cmd   Command
		score int // Lower is better
	}

	var results []scored

	for _, cmd := range r.commands {
		// Skip sub-commands in general search
		if strings.Contains(cmd.Name, " ") {
			continue
		}

		bestScore := -1

		// Check primary name
		if score := matchScore(cmd.Name, query); score >= 0 {
			bestScore = score
		}

		// Check aliases
		for _, alias := range cmd.Aliases {
			if score := matchScore(alias, query); score >= 0 {
				if bestScore < 0 || score < bestScore {
					bestScore = score
				}
			}
		}

		if bestScore >= 0 {
			results = append(results, scored{cmd: cmd, score: bestScore})
		}
	}

	// Sort by score (lower is better), then alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score < results[j].score
		}
		return results[i].cmd.Name < results[j].cmd.Name
	})

	// Extract commands
	cmds := make([]Command, len(results))
	for i, r := range results {
		cmds[i] = r.cmd
	}
	return cmds
}

// matchScore returns a score for how well target matches query.
// Returns -1 if no match, 0 for exact match, 1 for prefix, 2 for contains.
func matchScore(target, query string) int {
	target = strings.ToLower(target)
	if target == query {
		return 0 // Exact match
	}
	if strings.HasPrefix(target, query) {
		return 1 // Prefix match
	}
	if strings.Contains(target, query) {
		return 2 // Contains match
	}
	// Fuzzy match: check if all query chars appear in order
	if fuzzyMatch(target, query) {
		return 3
	}
	return -1
}

// fuzzyMatch checks if all characters in query appear in target in order.
func fuzzyMatch(target, query string) bool {
	ti := 0
	for _, qc := range query {
		found := false
		for ti < len(target) {
			if rune(target[ti]) == qc {
				found = true
				ti++
				break
			}
			ti++
		}
		if !found {
			return false
		}
	}
	return true
}

// SearchSubCommands searches sub-commands for a parent command.
func (r *CommandRegistry) SearchSubCommands(parent, query string) []Command {
	subs := r.GetSubCommands(parent)
	if query == "" {
		return subs
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []Command

	for _, cmd := range subs {
		if matchScore(cmd.Name, query) >= 0 {
			results = append(results, cmd)
		}
		for _, alias := range cmd.Aliases {
			if matchScore(alias, query) >= 0 {
				results = append(results, cmd)
				break
			}
		}
	}
	return results
}

// HasSubCommands returns true if the command has sub-commands.
func (r *CommandRegistry) HasSubCommands(name string) bool {
	return len(r.GetSubCommands(name)) > 0
}

// registerAllCommands registers all TUI commands.
func (r *CommandRegistry) registerAllCommands() {
	// =========================================================================
	// Navigation Commands
	// =========================================================================
	r.Register(Command{
		Name:        "messages",
		Aliases:     []string{"m", "msg"},
		Description: "Go to messages view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "events",
		Aliases:     []string{"e", "ev", "cal", "calendar"},
		Description: "Go to calendar events view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "contacts",
		Aliases:     []string{"c", "ct"},
		Description: "Go to contacts view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "webhooks",
		Aliases:     []string{"w", "wh"},
		Description: "Go to webhooks view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "webhook-server",
		Aliases:     []string{"ws", "whs", "server"},
		Description: "Go to webhook server view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "grants",
		Aliases:     []string{"g", "gr"},
		Description: "Go to grants/accounts view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "inbound",
		Aliases:     []string{"i", "in", "inbox"},
		Description: "Go to inbound inboxes view",
		Category:    CategoryNavigation,
	})
	r.Register(Command{
		Name:        "dashboard",
		Aliases:     []string{"d", "dash", "home"},
		Description: "Go to dashboard",
		Category:    CategoryNavigation,
	})

	// =========================================================================
	// Message Commands
	// =========================================================================
	r.Register(Command{
		Name:        "compose",
		Aliases:     []string{"n", "new"},
		Description: "Compose new email",
		Category:    CategoryMessages,
		Shortcut:    "n",
	})
	r.Register(Command{
		Name:        "reply",
		Aliases:     []string{"r"},
		Description: "Reply to current message",
		Category:    CategoryMessages,
		Shortcut:    "R",
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "replyall",
		Aliases:     []string{"ra", "reply-all"},
		Description: "Reply all to message",
		Category:    CategoryMessages,
		Shortcut:    "A",
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "forward",
		Aliases:     []string{"f", "fwd"},
		Description: "Forward message",
		Category:    CategoryMessages,
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "star",
		Aliases:     []string{"s"},
		Description: "Toggle star on message",
		Category:    CategoryMessages,
		Shortcut:    "s",
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "unstar",
		Aliases:     []string{},
		Description: "Remove star from message",
		Category:    CategoryMessages,
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "read",
		Aliases:     []string{"mr"},
		Description: "Mark as read",
		Category:    CategoryMessages,
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "unread",
		Aliases:     []string{"mu"},
		Description: "Mark as unread",
		Category:    CategoryMessages,
		Shortcut:    "u",
		ContextView: "messages",
	})
	r.Register(Command{
		Name:        "delete",
		Aliases:     []string{"del", "rm"},
		Description: "Delete current item",
		Category:    CategoryMessages,
		Shortcut:    "dd",
	})
	r.Register(Command{
		Name:        "archive",
		Aliases:     []string{},
		Description: "Archive message",
		Category:    CategoryMessages,
		ContextView: "messages",
	})

	// =========================================================================
	// Calendar Commands (with sub-commands)
	// =========================================================================
	r.Register(Command{
		Name:        "event",
		Aliases:     []string{},
		Description: "Event management",
		Category:    CategoryCalendar,
		ContextView: "events",
		SubCommands: []Command{
			{Name: "new", Aliases: []string{"create"}, Description: "Create new event"},
			{Name: "edit", Aliases: []string{"update"}, Description: "Edit current event"},
			{Name: "delete", Aliases: []string{"del"}, Description: "Delete current event"},
		},
	})
	r.Register(Command{
		Name:        "rsvp",
		Aliases:     []string{},
		Description: "RSVP to event",
		Category:    CategoryCalendar,
		ContextView: "events",
		SubCommands: []Command{
			{Name: "yes", Description: "RSVP yes to event"},
			{Name: "no", Description: "RSVP no to event"},
			{Name: "maybe", Description: "RSVP maybe to event"},
		},
	})
	r.Register(Command{
		Name:        "availability",
		Aliases:     []string{"avail"},
		Description: "Check availability",
		Category:    CategoryCalendar,
	})
	r.Register(Command{
		Name:        "find-time",
		Aliases:     []string{"findtime"},
		Description: "Find meeting time",
		Category:    CategoryCalendar,
	})

	// =========================================================================
	// Contact Commands (with sub-commands)
	// =========================================================================
	r.Register(Command{
		Name:        "contact",
		Aliases:     []string{},
		Description: "Contact management",
		Category:    CategoryContacts,
		ContextView: "contacts",
		SubCommands: []Command{
			{Name: "new", Aliases: []string{"create"}, Description: "Create new contact"},
			{Name: "edit", Aliases: []string{"update"}, Description: "Edit current contact"},
			{Name: "delete", Aliases: []string{"del"}, Description: "Delete current contact"},
		},
	})

	// =========================================================================
	// Webhook Commands (with sub-commands)
	// =========================================================================
	r.Register(Command{
		Name:        "webhook",
		Aliases:     []string{},
		Description: "Webhook management",
		Category:    CategoryWebhooks,
		ContextView: "webhooks",
		SubCommands: []Command{
			{Name: "new", Aliases: []string{"create"}, Description: "Create new webhook"},
			{Name: "edit", Aliases: []string{"update"}, Description: "Edit current webhook"},
			{Name: "delete", Aliases: []string{"del"}, Description: "Delete current webhook"},
			{Name: "test", Description: "Test current webhook"},
		},
	})

	// =========================================================================
	// Folder Commands (with sub-commands)
	// =========================================================================
	r.Register(Command{
		Name:        "folder",
		Aliases:     []string{},
		Description: "Folder management",
		Category:    CategoryFolders,
		ContextView: "messages",
		SubCommands: []Command{
			{Name: "list", Aliases: []string{"ls"}, Description: "List all folders"},
			{Name: "create", Aliases: []string{"new"}, Description: "Create new folder"},
			{Name: "delete", Aliases: []string{"del"}, Description: "Delete folder"},
		},
	})
	r.Register(Command{
		Name:        "inbox",
		Aliases:     []string{},
		Description: "Go to inbox folder",
		Category:    CategoryFolders,
	})
	r.Register(Command{
		Name:        "sent",
		Aliases:     []string{},
		Description: "Go to sent folder",
		Category:    CategoryFolders,
	})
	r.Register(Command{
		Name:        "trash",
		Aliases:     []string{},
		Description: "Go to trash folder",
		Category:    CategoryFolders,
	})
	r.Register(Command{
		Name:        "drafts",
		Aliases:     []string{"dr"},
		Description: "Go to drafts",
		Category:    CategoryFolders,
	})

	// =========================================================================
	// Vim Commands
	// =========================================================================
	r.Register(Command{
		Name:        "quit",
		Aliases:     []string{"q", "exit"},
		Description: "Quit application",
		Category:    CategoryVim,
	})
	r.Register(Command{
		Name:        "quit!",
		Aliases:     []string{"q!"},
		Description: "Force quit",
		Category:    CategoryVim,
	})
	r.Register(Command{
		Name:        "wq",
		Aliases:     []string{"x"},
		Description: "Save and quit",
		Category:    CategoryVim,
	})
	r.Register(Command{
		Name:        "help",
		Aliases:     []string{"h"},
		Description: "Show help",
		Category:    CategoryVim,
		Shortcut:    "?",
	})
	r.Register(Command{
		Name:        "top",
		Aliases:     []string{"first", "gg"},
		Description: "Go to first row",
		Category:    CategoryVim,
		Shortcut:    "gg",
	})
	r.Register(Command{
		Name:        "bottom",
		Aliases:     []string{"last", "G"},
		Description: "Go to last row",
		Category:    CategoryVim,
		Shortcut:    "G",
	})

	// =========================================================================
	// System Commands
	// =========================================================================
	r.Register(Command{
		Name:        "refresh",
		Aliases:     []string{"reload"},
		Description: "Refresh current view",
		Category:    CategorySystem,
		Shortcut:    "r",
	})
}
