# Claude Code Quick Start Guide

Your AI-powered development assistant with self-learning capabilities.

---

## TL;DR - Essential Commands

```bash
# Start of session
/session-start              # Load context from previous sessions

# During development
/generate-tests             # Generate tests for your code
/fix-build                  # Fix build errors
/run-tests                  # Run unit/integration tests
/security-scan              # Security analysis

# When mistakes happen
/correct "description"      # Capture mistake → adds to LEARNINGS

# End of session
/diary                      # Save session learnings
```

---

## Session Workflow

```
┌─────────────────────────────────────────────────────────────┐
│  START SESSION                                              │
│  /session-start                                             │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  DEVELOPMENT                                                │
│  Write code, use commands, Claude learns from mistakes      │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  MISTAKE HAPPENS?                                           │
│  /correct "what went wrong"                                 │
└─────────────────┬───────────────────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────────────────┐
│  END SESSION                                                │
│  /diary                                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Commands Reference

### Session Management

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `/session-start` | Load context from previous sessions | Start of every session |
| `/diary` | Save learnings to memory | End of session |
| `/reflect` | Review diary, propose CLAUDE.md updates | Weekly review |

### Self-Learning

| Command | Purpose | Example |
|---------|---------|---------|
| `/correct` | Capture a mistake for learning | `/correct "forgot to handle nil pointer"` |

**How it works:**
1. You notice Claude made a mistake
2. Run `/correct "description of what went wrong"`
3. Claude abstracts the pattern and adds it to LEARNINGS in CLAUDE.md
4. Future sessions avoid the same mistake

### Development Commands

| Command | Purpose | Tools Used |
|---------|---------|------------|
| `/add-command` | Create new CLI command | Read, Write, Edit, Glob, Grep |
| `/generate-crud-command` | Generate full CRUD command | Read, Write, Edit, Glob, Grep |
| `/add-flag` | Add flag to existing command | Read, Edit |
| `/add-domain-type` | Add new domain type | Read, Write, Edit |
| `/add-api-method` | Add new API method | Read, Write, Edit |

### Testing Commands

| Command | Purpose | Tools Used |
|---------|---------|------------|
| `/run-tests` | Run unit/integration tests | Read, Bash(go test, make test) |
| `/generate-tests` | Generate tests for code | Read, Write, Edit, Bash(go test) |
| `/add-integration-test` | Add integration test | Read, Write, Edit, Bash(go test) |
| `/debug-test-failure` | Analyze and fix test failures | Read, Edit, Bash(go test) |
| `/analyze-coverage` | Analyze test coverage | Read, Bash(go test) |

### Quality Commands

| Command | Purpose | Tools Used |
|---------|---------|------------|
| `/fix-build` | Fix Go build errors | Read, Edit, Write, Bash(go build, go vet) |
| `/security-scan` | Security vulnerability scan | Read, Grep, Glob, Bash(make security) |
| `/review-pr` | Review pull request | Read, Grep, Glob, Bash(git) |
| `/update-docs` | Update documentation | Read, Write, Edit |

### Parallel Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `/parallel-explore` | Explore codebase with 4-5 parallel agents | Large codebase search, cross-layer feature discovery |
| `/parallel-review` | Review code with parallel reviewer agents | Large PRs, multi-file reviews |

---

## Specialized Agents

Claude has specialized agents for different tasks:

### code-writer (Opus)
**Best for:** Production-ready code in Go, JavaScript, CSS

```
Expertise:
- Go: Hexagonal architecture, error wrapping, table-driven tests
- JavaScript: Vanilla JS, progressive enhancement, event delegation
- CSS: BEM naming, CSS custom properties, mobile-first
- Go Templates: .gohtml partials, semantic HTML
```

### codebase-explorer (Sonnet)
**Best for:** Fast codebase exploration without coding

```
Use for:
- Finding where functionality is implemented
- Understanding code patterns
- Answering "where is X?" questions
- Returns concise summaries (<200 words)
```

### test-writer (Opus)
**Best for:** All test generation (Go + Playwright)

```
Go Tests:
- Table-driven tests with t.Run()
- Testify assertions (require, assert)
- Mock structs with function fields
- Integration tests (rate-limited, parallel)

Playwright E2E:
- Selector priority: getByRole > getByText > getByLabel > getByTestId
- Air (port 7365) + UI (port 7363) projects
- NEVER use CSS selectors or XPath
```

### frontend-agent (Sonnet)
**Best for:** JavaScript, CSS, Go templates

```
Tech stack:
- Vanilla ES6+ JavaScript (no npm in browser)
- CSS custom properties, BEM naming
- Go html/template (.gohtml)
```

### mistake-learner (Sonnet)
**Best for:** Abstracting mistakes into learnings

```
Process:
1. Understand what went wrong
2. Abstract the pattern (not specific instance)
3. Add to CLAUDE.md LEARNINGS section
```

### code-reviewer (Opus)
**Best for:** Independent code review for quality

```
Checks:
- Code quality and best practices
- Potential bugs and edge cases
- Security issues
- Performance concerns
```

**For deep security analysis:** Use `/security-scan` command

---

## Quality Hooks

Hooks run automatically to enforce quality:

### quality-gate.sh (Stop Hook)
**Runs:** When Claude tries to complete a task
**Blocks if:** Go code fails fmt, vet, lint, or tests

```bash
# What it checks:
go fmt ./...           # Code formatting
go vet ./...           # Static analysis
golangci-lint run      # Linting
go test -short ./...   # Unit tests
```

### subagent-review.sh (SubagentStop Hook)
**Runs:** When a subagent completes
**Blocks if:** Output contains CRITICAL, FAIL, or BUILD FAILED

### pre-compact.sh (PreCompact Hook)
**Runs:** Before context window compaction
**Action:** Warns to save learnings with `/diary`

### context-injector.sh (UserPromptSubmit Hook)
**Runs:** When you submit a prompt
**Action:** Injects relevant context based on keywords

```
Triggers:
- "test" → Testing patterns reminder
- "security" → Security scan reminder
- "api" → Nylas v3 API reminder
- "playwright" → Semantic selector reminder
- "css" → BEM naming reminder
- "commit" → Git rules reminder
```

### file-size-check.sh (PreToolUse Hook for Write)
**Runs:** Before writing Go files
**Blocks if:** File would exceed 600 lines
**Warns if:** File would exceed 500 lines

### auto-format.sh (PostToolUse Hook for Edit)
**Runs:** After editing Go files
**Action:** Auto-runs `gofmt -w` on the edited file

**Hook Configuration:** See `.claude/HOOKS-CONFIG.md` for settings.json setup

---

## Memory System

Claude remembers across sessions:

```
~/.claude/memory/
├── diary/                    # Session learnings
│   ├── 2025-12-30-session-1.md
│   └── 2025-12-30-session-2.md
└── nylas-cli/                # Project-specific memory
```

**Project tracking:**
```
claude-progress.txt           # What's done, in progress, next up
```

---

## LEARNINGS Section

CLAUDE.md has a self-updating LEARNINGS section:

### Project-Specific Gotchas
Things unique to this codebase:
- Playwright selectors: ALWAYS use semantic selectors
- Go tests: ALWAYS use table-driven tests
- Integration tests: ALWAYS use acquireRateLimit(t)

### Non-Obvious Workflows
Surprising sequences:
- Progressive disclosure: Keep main skill files under 100 lines
- Self-learning: Use "Reflect → Abstract → Generalize → Write"

### Time-Wasting Bugs Fixed
Hard-won knowledge:
- Go build cache corruption: Fix with `sudo rm -rf ~/.cache/go-build`

---

## Best Practices

### Do This

```bash
# Start every session with context
/session-start

# Capture mistakes immediately
/correct "what went wrong"

# Save learnings before ending
/diary

# Use specialized agents for their domain
# - test-writer for tests (Go + Playwright)
# - frontend-agent for CSS/JS
# - codebase-explorer for finding code
```

### Avoid This

```bash
# Don't skip session start - you lose context
# Don't ignore mistakes - they'll repeat
# Don't skip /diary - learnings are lost
# Don't use CSS selectors in Playwright - use semantic selectors
```

---

## Quick Reference Card

| Task | Command |
|------|---------|
| **Start session** | `/session-start` |
| **Create command** | `/add-command` |
| **Generate tests** | `/generate-tests` |
| **Fix build** | `/fix-build` |
| **Run tests** | `/run-tests` |
| **Security check** | `/security-scan` |
| **Explore codebase** | `/parallel-explore` |
| **Review PR** | `/parallel-review` |
| **Capture mistake** | `/correct "description"` |
| **End session** | `/diary` |
| **Review learnings** | `/reflect` |

---

## Getting Help

- **All commands:** Type `/` and see autocomplete
- **Command details:** Read `.claude/commands/<command>.md`
- **Agent details:** Read `.claude/agents/<agent>.md`
- **Hook setup:** Read `.claude/HOOKS-CONFIG.md`
- **Project rules:** Read `CLAUDE.md`
- **Improvement plan:** Read `c_plan.md`

---

**Welcome to self-learning Claude Code!**
