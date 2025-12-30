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
- **NEVER create files >600 lines** - Split logically by responsibility (types, helpers, handlers)

### ALWAYS DO (every code change):

```bash
make ci-full   # Complete CI: quality checks → tests → cleanup
# OR for quick checks without integration tests:
make ci        # Runs: fmt → vet → lint → test-unit → test-race → security → vuln → build
```

**⚠️ CRITICAL: Never skip linting. Fix ALL linting errors in code you wrote.**

**⚠️ CRITICAL: Enforce file size limits. Files must be ≤500 lines (ideal) or ≤600 lines (max).**

**Details:** See `.claude/rules/linting.md`, `.claude/rules/go-best-practices.md`, `.claude/rules/file-size-limits.md`

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
- `internal/adapters/mcp/proxy_response.go` - Response handling helpers

**Air files (Web UI):**
- `internal/air/` - HTTP server, handlers, templates (all files ≤500 lines)
- `internal/air/query_helpers.go` - Query parameter parsing utilities

**Air CSS (refactored for maintainability):**
```
internal/air/static/css/
  ├── main.css                 # Core styles and imports
  ├── accessibility-*.css      # Accessibility (core, aria)
  ├── calendar-*.css           # Calendar (grid, modal)
  ├── components-*.css         # UI components (account, skeleton, ui)
  ├── contacts-*.css           # Contacts (list, modal)
  ├── features-*.css           # Features (ui, widgets)
  ├── productivity-*.css       # Productivity (send, ui)
  └── settings-*.css           # Settings (ai, modal, notetaker)
```

**Server Core:** `server*.go` - Split by responsibility (lifecycle, stores, sync, offline, converters, templates)

**Handler Groups:** All handlers follow `handlers_<feature>*.go` pattern, split by responsibility:
  - `handlers_email*.go`, `handlers_drafts.go` - Email operations
  - `handlers_calendar*.go`, `handlers_events.go` - Calendar/events
  - `handlers_contacts*.go` - Contact CRUD, search, helpers
  - `handlers_ai_*.go` - AI features (summarize, compose, thread, config)
  - `handlers_*productivity*.go` - Productivity (scheduled send, undo, templates, split inbox, snooze)
  - `handlers_availability.go`, `handlers_bundles.go` - Other features

**Integration Tests:** `integration_*_test.go` - Split by feature (core, email, calendar, contacts, cache, ai, middleware, productivity, bundles)

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

**Nylas Adapter (refactored):**
- `internal/adapters/nylas/messages.go` - Message operations
- `internal/adapters/nylas/messages_send.go` - Send message helpers

**AI Adapter (refactored):**
- `internal/adapters/ai/pattern_learner.go` - Pattern learning core
- `internal/adapters/ai/pattern_learner_analysis.go` - Analysis methods

**TUI (Terminal UI):**
- `internal/tui/` - tview-based TUI (commands, compose, views)
- `internal/tui2/` - Bubble Tea TUI (models/, components/, styles/, state/)
- `internal/tui2/vhs-tests/` - VHS visual testing (tapes/, output/)

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

**Auto-loaded:**
- `CLAUDE.md` - This guide
- `.claude/rules/*.md` - Development rules
- Select docs (DEVELOPMENT, SECURITY, TIMEZONE, TUI, WEBHOOKS)

**On-demand (use Read tool when needed):**
| Doc | When to Load |
|-----|--------------|
| `docs/COMMANDS.md` | Adding/modifying CLI commands |
| `docs/ARCHITECTURE.md` | Understanding project structure |
| `docs/MCP.md` | Working on MCP server |
| `docs/AI.md` | Working on AI features |
| `docs/commands/*.md` | Detailed command guides |
| `docs/ai/*.md` | AI provider setup |

**Never loaded (excluded via .claudeignore):**
- Build artifacts, coverage reports, IDE files

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
