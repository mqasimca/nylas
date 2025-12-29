# Code Navigation Guide for Claude Code

**Purpose**: Help AI assistants navigate this 732-file codebase efficiently.

---

## Quick File Lookup

### Finding Files by Feature

```bash
# Email functionality
internal/cli/email/           # CLI commands
internal/adapters/nylas/messages.go  # API adapter
internal/air/handlers_email.go       # Web UI handler

# Calendar functionality
internal/cli/calendar/        # CLI commands
internal/adapters/nylas/calendars_*.go  # API adapters (split)
internal/air/handlers_events.go      # Web UI handler

# Contacts functionality
internal/cli/contacts/        # CLI commands
internal/adapters/nylas/contacts.go  # API adapter
internal/air/handlers_contacts*.go   # Web UI handlers (split)

# AI features
internal/cli/ai/              # CLI commands
internal/adapters/ai/         # AI provider adapters
internal/air/handlers_ai_*.go # Web UI handlers (split)

# MCP server
internal/cli/mcp/             # CLI commands
internal/adapters/mcp/proxy.go # MCP proxy server

# Webhooks
internal/cli/webhook/         # CLI commands
internal/adapters/nylas/webhooks.go  # API adapter
```

### Finding Test Files

```bash
# Unit tests (co-located)
internal/cli/email/email_test.go  # Same directory as source

# Integration tests (centralized)
internal/cli/integration/email_test.go  # CLI integration
internal/air/integration_email_test.go  # Web UI integration
```

---

## Architecture Layers

```
┌─────────────────────────────────────────┐
│  Entry Points                           │
├─────────────────────────────────────────┤
│  cmd/nylas/main.go          CLI entry   │
│  internal/air/server*.go    Web entry   │
│  internal/tui/app*.go       TUI entry   │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│  Command Layer                          │
├─────────────────────────────────────────┤
│  internal/cli/<feature>/    197 files   │
│  internal/air/handlers*.go  61 files    │
│  internal/tui/views*.go     Varies      │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│  Business Logic (Ports)                 │
├─────────────────────────────────────────┤
│  internal/ports/nylas.go    Interface   │
│  internal/domain/*.go       Types       │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│  Adapter Layer                          │
├─────────────────────────────────────────┤
│  internal/adapters/nylas/   83 files    │
│  internal/adapters/ai/      17 files    │
│  internal/adapters/mcp/     6 files     │
└─────────────────────────────────────────┘
```

---

## Common Patterns

### Pattern 1: Adding a New CLI Command

**Files to Touch:**
1. `internal/cli/<feature>/<command>.go` - Implementation
2. `internal/cli/<feature>/<feature>.go` - Register subcommand
3. `cmd/nylas/main.go` - Register root command (if new feature)
4. `internal/cli/integration/<feature>_test.go` - Add test
5. `docs/COMMANDS.md` - Document command

**Example**: See `internal/cli/email/` for reference

### Pattern 2: Adding a New API Method

**Files to Touch:**
1. `internal/ports/nylas.go` - Add interface method
2. `internal/adapters/nylas/<resource>.go` - Implement method
3. `internal/adapters/nylas/mock_<resource>.go` - Add mock
4. `internal/domain/<resource>.go` - Add/update types
5. Add tests in appropriate test file

**Example**: See `internal/adapters/nylas/messages.go`

### Pattern 3: Adding Web UI Handler

**Files to Touch:**
1. `internal/air/handlers_<feature>.go` - Add handler
2. `internal/air/server_lifecycle.go` - Register route
3. `internal/air/templates/<feature>.gohtml` - Add template
4. `internal/air/integration_<feature>_test.go` - Add test

**Example**: See `internal/air/handlers_email.go`

---

## File Organization Principles

### By Size (Post-Refactoring)

- **Target**: All files ≤500 lines (ideal), ≤600 lines (acceptable)
- **Current**: ~20 files still 590-626 lines (refactoring candidates)
- **Method**: Split by responsibility/functionality

### By Type

```
<feature>/
  ├── <feature>.go          # Main command/entrypoint
  ├── list.go               # List subcommand
  ├── create.go             # Create subcommand
  ├── update.go             # Update subcommand
  ├── delete.go             # Delete subcommand
  ├── helpers.go            # Shared helpers
  └── <feature>_test.go     # Tests
```

### Split Patterns Used

**handlers_<feature>.go → Multiple files**:
- `handlers_<feature>_types.go` - Type definitions
- `handlers_<feature>_crud.go` - CRUD operations
- `handlers_<feature>_search.go` - Search functionality
- `handlers_<feature>_helpers.go` - Helper functions

**<large_test>.go → Multiple files**:
- `<feature>_test_basic.go` - Basic functionality tests
- `<feature>_test_advanced.go` - Advanced/edge case tests
- `<feature>_test_integration.go` - Integration tests

---

## Key Entry Points for Common Tasks

### Debugging API Issues
1. Start: `internal/adapters/nylas/client.go` - HTTP client
2. Check: `internal/adapters/nylas/<resource>.go` - Specific adapter
3. Verify: `internal/ports/nylas.go` - Interface contract

### Debugging CLI Issues
1. Start: `cmd/nylas/main.go` - CLI entry
2. Check: `internal/cli/<feature>/<command>.go` - Command implementation
3. Verify: `internal/cli/<feature>/helpers.go` - Helper functions

### Debugging Web UI Issues
1. Start: `internal/air/server_lifecycle.go` - Routes
2. Check: `internal/air/handlers_<feature>.go` - Handler implementation
3. Verify: `internal/air/templates/<feature>.gohtml` - Template

### Adding AI Features
1. Config: `internal/adapters/ai/<provider>_client.go` - Provider adapter
2. Integration: `internal/cli/ai/` - CLI commands
3. Web UI: `internal/air/handlers_ai_*.go` - Web handlers

---

## Package-Level Organization

### Most Active Packages (by file count)

1. **internal/air** (86 files)
   - Web server and HTTP handlers
   - Templates in `templates/`
   - Cache in `cache/`

2. **internal/adapters/nylas** (83 files)
   - API client implementation
   - Resource adapters (split by type)
   - Mock implementations

3. **internal/tui** (75 files)
   - TView-based terminal UI
   - Views and pages
   - Theme configuration

4. **internal/cli/integration** (48 files)
   - Integration test suite
   - Organized by feature

5. **internal/tui2/models** (37 files)
   - Bubble Tea UI models
   - Experimental TUI

---

## Testing Strategy

### Test Locations

| Test Type | Location | Pattern |
|-----------|----------|---------|
| Unit tests | Co-located | `*_test.go` in same directory |
| CLI integration | `internal/cli/integration/` | `<feature>_test.go` |
| Web integration | `internal/air/` | `integration_<feature>_test.go` |
| Adapter integration | `internal/adapters/nylas/` | `integration_<feature>_test.go` |

### Running Tests

```bash
# All tests with cleanup
make ci-full

# Just unit tests
make test-unit

# Just CLI integration
make test-integration

# Just Web UI integration
make test-air-integration

# Cleanup test resources
make test-cleanup
```

---

## Documentation Hierarchy

### Auto-Loaded (Always in Context)
- `CLAUDE.md` - AI assistant guide
- `CODE_NAVIGATION.md` - This file
- `docs/ARCHITECTURE.md` - Architecture overview
- `docs/COMMANDS.md` - Command reference

### Load On-Demand
- `docs/commands/*.md` - Detailed command docs
- `docs/ai/*.md` - AI provider setup
- `docs/development/*.md` - Development guides
- `docs/examples/*.md` - Usage examples

### Search Strategy
1. Check `CLAUDE.md` for overview
2. Check this file for navigation
3. Use `Grep` to find specific code
4. Use `Read` for specific files
5. Load detailed docs only when needed

---

## Token Optimization Tips

### What NOT to Load
- Test fixtures: `tests/fixtures/` (claudeignored)
- VHS tapes: `internal/tui2/vhs-tests/` (claudeignored)
- Build artifacts: `bin/`, `*.test` (claudeignored)
- Detailed docs: Load only when needed

### What to Prioritize
- Interface definitions: `internal/ports/`
- Domain types: `internal/domain/`
- Core adapters: `internal/adapters/nylas/`
- Command structure: `internal/cli/`

### Efficient Patterns
- Use `Grep` for searching across files
- Use `Glob` for finding files by pattern
- Use `Read` sparingly, with limits when possible
- Rely on documentation first

---

## Common Search Patterns

```bash
# Find all handlers for a feature
grep -r "func.*Handle.*Email" internal/air/

# Find where a type is defined
grep -r "^type EmailMessage " internal/

# Find all tests for a feature
find . -name "*email*test.go"

# Find where a command is registered
grep -r "AddCommand.*email" internal/cli/

# Find API method implementations
grep -r "func.*SendMessage" internal/adapters/
```

---

## Refactoring Reference

See `REFACTORING_GUIDE.md` for detailed patterns on:
- How to split large files
- Naming conventions for split files
- When to split vs when to keep together
- Examples of successful refactorings

---

**Last Updated**: 2025-12-29
**Maintained By**: Automated during refactoring
**Version**: Post-35-file-refactoring (626-794 lines → 89 files)
