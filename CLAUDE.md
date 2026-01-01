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

**Details:** See `.claude/rules/go-quality.md`, `.claude/rules/file-size-limits.md`

### Test & Doc Requirements:
| Change | Unit Test | Integration Test | Update Docs |
|--------|-----------|------------------|-------------|
| New feature | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Bug fix | ✅ REQUIRED | ⚠️ If API-related | ⚠️ If behavior changes |
| New command | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Flag change | ✅ REQUIRED | ❌ Not needed | ✅ REQUIRED |

### Test Coverage Goals:

**See:** `.claude/rules/testing.md` for coverage targets by package type.

**Check coverage:** `make test-coverage`

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

**Handler Groups:** All handlers follow `handlers_<feature>*.go` pattern (61 files), split by responsibility:
  - `handlers_email*.go`, `handlers_drafts.go` - Email operations
  - `handlers_calendar*.go`, `handlers_events.go` - Calendar/events
  - `handlers_contacts*.go` (4 files) - Contact CRUD, search, helpers
  - `handlers_ai_*.go` (6 files) - AI features (complete, config, smart, summarize, thread, types)
  - `handlers_scheduled_send.go`, `handlers_undo_send.go` - Send scheduling
  - `handlers_templates*.go` (2 files) - Email templates
  - `handlers_snooze_*.go` (3 files) - Snooze (handlers, parser, types)
  - `handlers_splitinbox_*.go` (3 files) - Split inbox (categorize, config, types)
  - `handlers_reply_later.go` - Reply later feature
  - `handlers_analytics.go` - Analytics tracking
  - `handlers_focus_mode.go` - Focus mode
  - `handlers_read_receipts.go` - Read receipts
  - `handlers_screener.go` - Email screener
  - `handlers_notetaker.go` - Notetaker feature
  - `handlers_availability.go`, `handlers_bundles.go`, `handlers_cache.go`, `handlers_config.go` - Other features

**Integration Tests:** `integration_*_test.go` - Split by feature (core, email, calendar, contacts, cache, ai, middleware, productivity, bundles)

**CLI Commands (19 directories in internal/cli/):**
| Directory | Files | Purpose |
|-----------|-------|---------|
| `admin/` | 6 | Applications, connectors, credentials, grants |
| `ai/` | 6 | Budget, usage, config, clear_data |
| `auth/` | 22 | Login, logout, add, remove, status, switch, whoami |
| `calendar/` | 35 | Events, availability, conflicts, focus time, AI scheduling |
| `common/` | 16 | Client, errors, format, logger, pagination, progress |
| `contacts/` | 13 | CRUD, groups, photos, sync |
| `demo/` | 19 | Interactive demos for all features |
| `email/` | 20 | CRUD, AI, drafts, attachments, threads |
| `inbound/` | 9 | Inbound email handling |
| `integration/` | 49 | Integration tests split by feature |
| `mcp/` | 7 | Install, serve, status, uninstall, assistants |
| `notetaker/` | 8 | CRUD, media management |
| `otp/` | 7 | OTP/SMS operations |
| `scheduler/` | 6 | Bookings, configurations, pages |
| `slack/` | 9 | Channels, messages, users, send, reply, search |
| `timezone/` | 10 | Convert, DST, find, info |
| `ui/` | 15 | UI server commands |
| `update/` | 5 | CLI update commands |
| `webhook/` | 11 | Webhook commands |

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

**Adapter Directories (12 in internal/adapters/):**
| Adapter | Files | Purpose |
|---------|-------|---------|
| `ai/` | 18 | AI clients (Claude, OpenAI, Groq, Ollama), email analyzer, pattern learner |
| `analytics/` | 14 | Focus optimizer, conflict resolver, meeting scorer |
| `browser/` | 2 | Browser automation |
| `config/` | 5 | Configuration validation |
| `keyring/` | 6 | Credential storage (file, grants) |
| `mcp/` | 7 | MCP proxy server |
| `nylas/` | 85 | Nylas API client (main adapter) |
| `oauth/` | 3 | OAuth server |
| `slack/` | 9 | Slack API client (channels, messages, users, search) |
| `tunnel/` | 2 | Cloudflare tunnel |
| `utilities/` | 12 | Services (contacts, email, scheduling, timezone, webhook) |
| `webhookserver/` | 2 | Webhook server |

**Nylas Adapter (heavily refactored):**
- `messages.go`, `messages_send.go` - Message operations
- `calendars_*.go` (5 files) - Calendar operations (calendars, converters, events, types, virtual)
- `demo/` subdir (16 files) - Demo client for screenshots
- `mock_*.go` (16 files) - Split mock implementations

**AI Adapter (refactored):**
- `*_client.go` - AI clients (base, claude, openai, groq, ollama)
- `email_analyzer_core.go`, `email_analyzer_prompts.go` - Email analyzer
- `pattern_learner.go`, `pattern_learner_analysis.go` - Pattern learning

**Analytics Adapter (new):**
- `focus_optimizer_*.go` (3 files) - Focus time optimization
- `pattern_learner_*.go` (2 files) - Meeting pattern learning
- `conflict_resolver.go`, `meeting_scorer.go` - Conflict detection

**TUI (Terminal UI):**
- `internal/tui/` - tview-based TUI (77 files: commands, compose, views)
- **Themes:** k9s, amber, green, apple2, vintage, ibm, futuristic, matrix, norton

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

**Detailed guide:** Use `/add-command` skill

---

## Go Modernization

**See:** `.claude/rules/go-quality.md` for modern Go patterns (1.21+), error handling, and linting fixes.

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

**Shared test patterns:** `.claude/shared/patterns/`

---

## Quality Hooks

Hooks run automatically to enforce quality:

| Hook | Trigger | Purpose |
|------|---------|---------|
| `quality-gate.sh` | Stop | Blocks if Go code fails fmt/vet/lint/tests |
| `subagent-review.sh` | SubagentStop | Blocks if subagent finds critical issues |
| `pre-compact.sh` | PreCompact | Warns before context compaction |
| `context-injector.sh` | UserPromptSubmit | Injects context reminders |
| `file-size-check.sh` | PreToolUse (Write) | Blocks Go files >600 lines, warns >500 |
| `auto-format.sh` | PostToolUse (Edit) | Auto-runs gofmt on edited Go files |

**Full hook docs:** `.claude/HOOKS-CONFIG.md`

---

## Useful Commands

| Command | What It Does |
|---------|--------------|
| `/session-start` | Load context from previous sessions |
| `/add-command` | Add new CLI command |
| `/generate-tests` | Generate tests for code |
| `/fix-build` | Fix build errors |
| `/run-tests` | Run unit/integration tests |
| `/security-scan` | Security audit |
| `/correct "desc"` | Capture mistake for learning |
| `/diary` | Save session learnings |

**Full command list:** See `CLAUDE-QUICKSTART.md`

---

## Subagent Parallelization

Use parallel agents to explore or review the 745-file codebase without exhausting context.

### When to Use Parallel Agents

| Task | Agents | Why |
|------|--------|-----|
| Full codebase exploration | 5 | One per major directory |
| Feature search | 4 | Search cli, adapters, air, tui simultaneously |
| Multi-file PR review | 4 | Review different files in parallel |
| Test coverage analysis | 4 | Analyze different packages |

### Invocation Patterns

```
# Full exploration (4 agents)
"Explore using 4 parallel agents:
 - Agent 1: internal/cli/
 - Agent 2: internal/adapters/
 - Agent 3: internal/air/
 - Agent 4: internal/domain/ + ports/ + tui/"

# Feature search (4 agents)
"Find all email-related code using 4 agents across cli, adapters, air, tui"

# PR review (4 agents)
"Review these 8 files using 4 parallel code-reviewer agents"
```

### Directory Parallelization Value

| Directory | Files | Parallel Value |
|-----------|-------|----------------|
| `internal/cli/` | 268 | HIGH |
| `internal/adapters/` | 158 | HIGH |
| `internal/air/` | 117 | HIGH |
| `internal/tui/` | 77 | MEDIUM |
| `internal/domain/` | 21 | LOW (shared) |

### Safe vs Unsafe

**✅ SAFE:** Explore, review, search across different directories
**❌ AVOID:** Write to same file, modify domain/ or ports/nylas.go, parallel integration tests

### Existing Agents for Parallel Use

| Agent | Parallel Use |
|-------|--------------|
| `codebase-explorer` | ✅ Spawn 4-5 for large searches |
| `code-reviewer` | ✅ Review different files |
| `test-writer` | ⚠️ Different packages only |
| `code-writer` | ⚠️ Disjoint files only |

### Key Benefit

Parallel agents have **isolated context windows** - prevents "dumb Claude mid-session" when exploring large codebases.

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

**Dynamic local rules (check before operations):**

Before performing these operations, check if a matching `.local.md` rule exists and read it:

| Operation | Check for file |
|-----------|----------------|
| Git commits | `.claude/rules/git-commits.local.md` |
| Go cache cleanup | `.claude/rules/go-cache-cleanup.local.md` |
| Any operation | `.claude/rules/<operation>.local.md` |

```bash
# Pattern: Check if local rule exists before operation
ls .claude/rules/<operation>.local.md 2>/dev/null && Read it
```

---

## Session Continuity (Auto-Update)

**CRITICAL:** Automatically maintain `claude-progress.txt` for session continuity.

### When to Update `claude-progress.txt`:

| Trigger | Action |
|---------|--------|
| After completing a major task | Update with what was done |
| After making significant progress | Update current state |
| Before context feels full | Save everything important |
| When switching focus areas | Document where you left off |

### Update Format:

```markdown
# Claude Progress - [DATE]

## Current Branch
[branch name]

## Last Session Summary
[1-3 sentences of what was accomplished]

## In Progress
- [Current task being worked on]

## Next Steps
1. [Immediate next action]
2. [Following action]

## Key Decisions Made
- [Important choices and rationale]

## Blockers/Notes
- [Any issues or things to remember]
```

### Auto-Update Rule:

**After completing any task or making significant progress, proactively update `claude-progress.txt` without being asked.** This ensures session continuity when a new conversation starts.

Quick update command pattern:
```
Write current progress to claude-progress.txt
```

---

## Quick Reference

### Essential Make Targets

**See:** `docs/DEVELOPMENT.md` for complete make target list.

| Target | Description |
|--------|-------------|
| `make ci-full` | **Complete CI pipeline** (quality + tests + cleanup) |
| `make ci` | Quality checks only (no integration) |
| `make build` | Build binary |
| `make test-unit` | Unit tests only |

```bash
# Before committing code
make ci-full
```

**Debugging:**
1. Check `internal/ports/nylas.go` - Interface contract
2. Check `internal/adapters/nylas/client.go` - HTTP client
3. Check `internal/cli/<feature>/helpers.go` - CLI helpers

**API:** https://developer.nylas.com/docs/api/v3/

---

## LEARNINGS (Self-Updating)

> **When Claude makes a mistake, use:** "Reflect on this mistake. Abstract and generalize the learning. Write it to CLAUDE.md."

This section captures lessons learned from mistakes. Claude updates this section when errors are caught.

### Project-Specific Gotchas
- Playwright selectors: ALWAYS use semantic selectors (getByRole > getByText > getByLabel > getByTestId), NEVER CSS/XPath
- Go tests: ALWAYS use table-driven tests with t.Run() for multiple scenarios
- Air handlers: ALWAYS return after error responses (prevents writing to closed response)
- Integration tests: ALWAYS use acquireRateLimit(t) before API calls in parallel tests
- Frontend JS: ALWAYS use textContent for user data, NEVER innerHTML (XSS prevention)

### Non-Obvious Workflows
- Progressive disclosure: Keep main skill files under 100 lines, use references/ for details
- Self-learning: Use "Reflect → Abstract → Generalize → Write" when mistakes occur
- Session continuity: Read claude-progress.txt at session start, update at session end
- Hook debugging: Check ~/.claude/logs/ for hook execution errors

### Time-Wasting Bugs Fixed
- Go build cache corruption: Fix with `sudo rm -rf ~/.cache/go-build ~/go/pkg/mod && go clean -cache`
- Playwright MCP not connecting: Run `claude mcp add playwright` to install plugin
- Quality gate timeout: Add `timeout 120` before golangci-lint in hooks

### Curation Rules
- Maximum 30 items per category
- Remove obsolete entries when adding new
- One imperative line per item
- Monthly review to prune stale advice

---

## META - MAINTAINING THIS DOCUMENT

> This section governs how Claude updates CLAUDE.md itself.

### When to Update CLAUDE.md

| Trigger | Action |
|---------|--------|
| Mistake caught by user | Add to LEARNINGS with abstracted pattern |
| New workflow discovered | Add to Quick Reference if reusable |
| Rule violation | Strengthen rule with concrete example |
| Obsolete workaround | Remove from LEARNINGS |

### Writing Principles

**Core (Always Apply):**
- Use absolute directives ("ALWAYS"/"NEVER") for critical rules
- Lead with rationale before solution (1-3 bullets max)
- Be concrete with actual commands/code examples
- One clear point per bullet
- Use bullets over paragraphs

**Anti-Bloat Rules:**
- Don't add "Warning Signs" to obvious rules
- Don't show bad examples for trivial mistakes
- Don't create decision trees for binary choices
- Remove entries that haven't been relevant in 30 days

### Process for Adding Rules

1. Add new rule to appropriate section
2. Verify rule doesn't duplicate existing guidance
3. Test that rule is actionable (not vague)
4. Keep entry under 2 lines
