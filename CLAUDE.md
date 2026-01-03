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

| Layer | Location |
|-------|----------|
| CLI | `internal/cli/<feature>/` |
| Adapter | `internal/adapters/nylas/<feature>.go` |
| Domain | `internal/domain/<feature>.go` |
| Tests | `internal/cli/integration/<feature>_test.go` |

**Core files:**
- `cmd/nylas/main.go` - Entry point
- `internal/ports/nylas.go` - Interface contract
- `internal/adapters/nylas/client.go` - HTTP client

**CLI pattern:**
```
internal/cli/<feature>/
  ├── <feature>.go    # Main command
  ├── list.go         # Subcommands
  └── helpers.go      # Shared utilities
```

**Quick lookup:**
| Looking for | Location |
|-------------|----------|
| CLI helpers (context, config, colors) | `internal/cli/common/` |
| HTTP client | `internal/adapters/nylas/client.go` |
| AI clients (Claude, OpenAI, Groq) | `internal/adapters/ai/` |
| MCP server | `internal/adapters/mcp/` |
| Slack adapter | `internal/adapters/slack/` |
| Air web UI (port 7365) | `internal/air/` |
| UI web interface (port 7363) | `internal/cli/ui/` |
| TUI terminal interface | `internal/tui/` |
| Integration test helpers | `internal/cli/integration/helpers_test.go` |
| Air integration tests | `internal/air/integration_*_test.go` |

**Full inventory:** `docs/ARCHITECTURE.md`

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

**Command:** `make ci-full` (complete CI: quality + tests + cleanup)

**Quick checks:** `make ci` (no integration tests)

**Details:** `.claude/rules/testing.md`

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

**Essential:** `/session-start`, `/run-tests`, `/correct "desc"`, `/diary`

**Development:** `/add-command`, `/generate-tests`, `/fix-build`, `/security-scan`

**Full list:** `CLAUDE-QUICKSTART.md`

---

## Subagent Parallelization

**Guide:** `.claude/agents/README.md`

**Quick reference:**
- Safe: `codebase-explorer`, `code-reviewer` (spawn 4-5 for large tasks)
- Limited: `code-writer`, `test-writer` (disjoint files only)
- Serial: `mistake-learner` (runs alone)

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
