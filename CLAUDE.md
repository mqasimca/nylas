# AI Assistant Guide for Nylas CLI

Quick reference for AI assistants working on this codebase.

---

## ⛔ CRITICAL RULES - MUST FOLLOW

### NEVER DO:
- **NEVER run `git commit`** - User will commit changes manually
- **NEVER run `git push`** - User will push changes manually
- **NEVER commit secrets** - No API keys, tokens, passwords, .env files
- **NEVER skip tests** - All changes require passing tests
- **NEVER skip security scans** - Run `make security` before commits

### ALWAYS DO (every code change):

```bash
make ci-full   # Complete CI: quality checks → tests → cleanup
# OR for quick checks without integration tests:
make ci        # Runs: fmt → vet → lint → test-unit → test-race → security → vuln → build
```

**⚠️ CRITICAL: Never skip linting. Fix ALL linting errors in code you wrote.**

**Details:** See `.claude/rules/linting.md`, `.claude/rules/go-best-practices.md`

### Test & Doc Requirements:
| Change | Unit Test | Integration Test | Update Docs |
|--------|-----------|------------------|-------------|
| New feature | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Bug fix | ✅ REQUIRED | ⚠️ If API-related | ⚠️ If behavior changes |
| New command | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Flag change | ✅ REQUIRED | ❌ Not needed | ✅ REQUIRED |

### Test Coverage Goals:

| Package Type | Minimum | Target |
|--------------|---------|--------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

**Check coverage:** `make test-coverage` (opens coverage.html in browser)

### Docs to Update:
- `docs/COMMANDS.md` → New/changed commands or flags
- `docs/TIMEZONE.md` → Timezone-related changes, DST handling
- `docs/AI.md` → AI features, provider setup
- `docs/MCP.md` → MCP server, AI assistant integration
- `README.md` → Major new features

**📋 Complete rules:** `.claude/rules/documentation-maintenance.md`

### Do Not Touch:
| Path | Reason |
|------|--------|
| `.env*`, `**/secrets/**` | Contains secrets |
| `*.pem`, `*.key` | Certificates/keys |
| `go.sum` | Auto-generated |
| `.git/`, `vendor/` | Managed externally |

---

## Project Overview

- **Language**: Go 1.24.0 (use latest features!)
- **Architecture**: Hexagonal (ports and adapters)
- **CLI Framework**: Cobra
- **API**: Nylas v3 ONLY (never use v1/v2)
- **Timezone Support**: Offline utilities + calendar integration ✅

**Details:** See `docs/ARCHITECTURE.md`

---

## File Structure

**Standard pattern for all features:**

| Layer | Location | Example |
|-------|----------|---------|
| CLI | `internal/cli/<feature>/` | `internal/cli/email/` |
| Adapter | `internal/adapters/nylas/<feature>.go` | `internal/adapters/nylas/messages.go` |
| Domain | `internal/domain/<feature>.go` | `internal/domain/message.go` |
| Tests | `internal/cli/integration/<feature>_test.go` | `internal/cli/integration/email_test.go` |

**Core files:**
- `cmd/nylas/main.go` - Entry point, register commands
- `internal/ports/nylas.go` - Interface contract
- `internal/adapters/nylas/client.go` - HTTP client
- `internal/domain/config.go` - Configuration (working hours, breaks)

**MCP files:**
- `internal/cli/mcp/` - CLI commands (install, serve, status, uninstall)
- `internal/adapters/mcp/proxy.go` - MCP proxy server

**Air files (Web UI):**
- `internal/air/` - HTTP server, handlers, templates (all files ≤500 lines)

**Server Core (refactored from server.go):**
  - `server.go` (51 lines) - Server struct definition
  - `server_lifecycle.go` (315 lines) - Initialization, routing, lifecycle
  - `server_stores.go` (67 lines) - Cache store accessors
  - `server_sync.go` (187 lines) - Background sync logic
  - `server_offline.go` (98 lines) - Offline queue processing
  - `server_converters.go` (116 lines) - Domain to cache conversions
  - `server_template.go` (163 lines) - Template handling
  - `server_modules_test.go` (523 lines) - Unit tests for server modules

**Email Handlers:**
  - `handlers_email.go` (522 lines) - Email operations
  - `handlers_drafts.go` (531 lines) - Draft operations
  - `handlers_bundles.go` - Email categorization into smart bundles

**Calendar Handlers (refactored from handlers_calendar.go):**
  - `handlers_calendars.go` (108 lines) - Calendar operations
  - `handlers_events.go` (520 lines) - Event CRUD
  - `handlers_calendar_helpers.go` (159 lines) - Helpers and converters

**Contact Handlers (refactored from handlers_contacts.go):**
  - `handlers_contacts.go` (190 lines) - Routes and listing
  - `handlers_contacts_crud.go` (320 lines) - CRUD operations
  - `handlers_contacts_search.go` (194 lines) - Search and groups
  - `handlers_contacts_helpers.go` (308 lines) - Helpers and demo data

**AI Handlers (refactored from handlers_ai.go):**
  - `handlers_ai_types.go` (92 lines) - AI types
  - `handlers_ai_summarize.go` (165 lines) - Email summarization
  - `handlers_ai_smart.go` (225 lines) - Smart compose/reply
  - `handlers_ai_thread.go` (192 lines) - Thread analysis
  - `handlers_ai_complete.go` (216 lines) - Autocomplete
  - `handlers_ai_config.go` (227 lines) - AI configuration

**Productivity - Send Features (refactored from handlers_productivity_send.go):**
  - `types_productivity_send.go` (82 lines) - Type definitions
  - `handlers_scheduled_send.go` (177 lines) - Scheduled sends
  - `handlers_undo_send.go` (148 lines) - Undo send feature
  - `handlers_templates.go` (259 lines) - Template CRUD
  - `handlers_templates_helpers.go` (100 lines) - Template utilities

**Productivity - Inbox Features (refactored from handlers_productivity_inbox.go):**
  - `handlers_splitinbox_types.go` (84 lines) - Split inbox types
  - `handlers_splitinbox_config.go` (164 lines) - Configuration
  - `handlers_splitinbox_categorize.go` (148 lines) - Categorization logic
  - `handlers_snooze_types.go` (37 lines) - Snooze types
  - `handlers_snooze_handlers.go` (126 lines) - Snooze operations
  - `handlers_snooze_parser.go` (142 lines) - Natural language parser

**Other Handlers:**
  - `handlers_availability.go` (527 lines) - Availability checking
  - `handlers_productivity_*.go` - Focus mode, reply later, analytics, etc.

**Integration Tests:**
  - `integration_base_test.go` - Shared `testServer()` helper and utilities
  - `integration_core_test.go` - Config, Grants, Folders, Index tests
  - `integration_email_test.go` - Email and draft operations
  - `integration_calendar_test.go` - Calendar, events, availability, conflicts
  - `integration_contacts_test.go` - Contact operations
  - `integration_cache_test.go` - Cache operations
  - `integration_ai_test.go` - AI features
  - `integration_middleware_test.go` - Middleware tests
  - `integration_productivity_test.go` - Productivity features
  - `integration_bundles_test.go` - Bundle categorization

**CLI pattern:**
```
internal/cli/<feature>/
  ├── <feature>.go    # Main command
  ├── list.go         # List subcommand
  ├── create.go       # Create subcommand
  ├── update.go       # Update subcommand
  ├── delete.go       # Delete subcommand
  └── helpers.go      # Shared helpers
```

**TUI (Terminal UI) implementations:**
```
internal/tui/        # Current tview-based TUI (k9s-style)
internal/tui2/       # New Bubble Tea TUI (experimental)
  ├── models/        # Screen models (Dashboard, Messages, etc.)
  ├── components/    # Custom Bubble Tea components
  ├── styles/        # Lip Gloss themes and styling
  ├── state/         # State management (GlobalState)
  ├── utils/         # Utilities
  ├── vhs-tests/     # VHS visual testing (screenshots, GIFs)
  │   ├── tapes/     # VHS tape files (.tape)
  │   ├── output/    # Generated screenshots and GIFs
  │   └── README.md  # VHS testing documentation
  ├── app.go         # Root application (Model Stack pattern)
  ├── messages.go    # Message types for Elm Architecture
  └── app_test.go    # Tests
```

**Switch between TUI engines:**
```bash
nylas tui                      # Default: tview
nylas tui --engine bubbletea   # Experimental: Bubble Tea
```

**Visual testing with VHS:**
```bash
make test-vhs                  # Run dashboard visual test
make test-vhs-all              # Run all visual tests (splash, dashboard, navigation)
make test-vhs-clean            # Clean test outputs
```

---

## Adding a New Feature

**Quick pattern:**
1. Domain: `internal/domain/<feature>.go` - Define types
2. Port: `internal/ports/nylas.go` - Add interface methods
3. Adapter: `internal/adapters/nylas/<feature>.go` - Implement methods
4. Mock: `internal/adapters/nylas/mock.go` - Add mock methods
5. CLI: `internal/cli/<feature>/` - Add commands
6. Register: `cmd/nylas/main.go` - Add command
7. Tests: `internal/cli/integration/<feature>_test.go`
8. Docs: `docs/COMMANDS.md` - Add examples

**Detailed guide:** Use `/project:add-feature` skill

---

## Go Modernization (Go 1.21+)

**Always use modern patterns:**

| Instead of | Use | Since |
|------------|-----|-------|
| `io/ioutil` | `os` package | Go 1.16+ |
| `interface{}` | `any` | Go 1.18+ |
| Manual slice ops | `slices` package | Go 1.21+ |
| Manual map ops | `maps` package | Go 1.21+ |
| Recreate to clear | `clear()` | Go 1.21+ |
| Custom min/max | `min()`, `max()` | Go 1.21+ |
| `sort.Slice` | `slices.SortFunc` | Go 1.21+ |

**Before writing Go code:** Check `go.dev/ref/spec` using WebSearch

**Complete rules:** `.claude/rules/go-best-practices.md`

---

## Testing

### 🚀 Primary Command (Recommended)
```bash
make ci-full   # Complete CI pipeline:
               # • All code quality checks (fmt, vet, lint, vuln, security)
               # • All unit tests (test-unit, test-race)
               # • All integration tests (CLI + Air)
               # • Automatic cleanup of test resources
               # • Output saved to ci-full.txt
```

### Granular Testing (When Needed)

**Unit Tests:**
```bash
make test-unit                   # Run unit tests (-short)
make test-race                   # Run with race detector
make test-coverage               # Generate coverage report
go test ./internal/cli/email/... # Test specific package
```

**Integration Tests:**
```bash
make test-integration            # CLI integration tests
make test-integration-fast       # Fast tests (skip LLM)
make test-air-integration        # Air web UI integration tests
```

**Cleanup:**
```bash
make test-cleanup                # Clean up test resources
                                 # (emails, events, grants)
```

**Location:**
- CLI integration tests: `internal/cli/integration/`
- Air integration tests: `internal/air/integration_*.go`

**Details:** `.claude/rules/testing.md`

---

## Useful Commands

| Command | What It Does |
|---------|--------------|
| `/clear` | Reset context |
| `/project:add-feature` | Add new feature workflow |
| `/project:fix-bug` | Bug fix workflow |
| `/project:review-pr` | Code review |
| `/project:security-scan` | Security audit |

---

## Context Loading Strategy

**Auto-loaded (always in context):**
- `CLAUDE.md` - This guide
- `docs/COMMANDS.md` - Command reference
- `docs/ARCHITECTURE.md` - Architecture overview
- `.claude/rules/*.md` - Development rules

**Load on-demand (use Read tool):**
- `docs/commands/*.md` - Detailed command guides
- `docs/ai/*.md` - AI provider setup
- `plan.md`, `AI_plan.md` - Planning documents

**Never loaded (excluded via .claudeignore):**
- `local/` - Historical docs (233KB)
- Build artifacts, coverage reports, IDE files
- `docs/commands/`, `docs/ai/` - Detailed guides

---

## Quick Reference

### Essential Make Targets

| Target | Description | When to Use |
|--------|-------------|-------------|
| `make ci-full` | **Complete CI pipeline** (quality + tests + cleanup) | Before PRs, releases |
| `make ci` | Quality checks only (no integration) | Quick pre-commit |
| `make build` | Build binary | Development |
| `make test-unit` | Unit tests only | Fast feedback |
| `make test-coverage` | Coverage report | Check test coverage |
| `make clean` | Remove artifacts | Clean workspace |

**Run `make help` for all available targets**

### Common Workflows

```bash
# Before committing code
make ci-full

# Quick pre-commit check
make ci

# Build and test locally
make build
./bin/nylas --help

# Check test coverage
make test-coverage
```

**Debugging:**
1. Check `internal/ports/nylas.go` - Interface contract
2. Check `internal/adapters/nylas/client.go` - HTTP client
3. Check `internal/cli/<feature>/helpers.go` - CLI helpers

**API:** https://developer.nylas.com/docs/api/v3/
