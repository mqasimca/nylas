package tui

import (
	"cmp"
	"slices"
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
	slices.SortFunc(result, func(a, b Command) int {
		return cmp.Compare(a.Name, b.Name)
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
	slices.SortFunc(results, func(a, b scored) int {
		if a.score != b.score {
			return cmp.Compare(a.score, b.score)
		}
		return cmp.Compare(a.cmd.Name, b.cmd.Name)
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
