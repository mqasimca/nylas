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
# 1. Check Go docs for modern patterns (REQUIRED for Go code)
#    See: .claude/rules/go-best-practices.md
#    - Check go.dev/ref/spec for latest features
#    - Use WebSearch to verify best practices
#    - Apply modern Go idioms (slices, maps, clear, min/max, generics)

# 2. Write/update tests for your changes

# 3. Format code
go fmt ./...

# 4. Lint and fix ALL issues in your code (MANDATORY)
#    See: .claude/rules/linting.md for common fixes
golangci-lint run --timeout=5m
#    Fix errcheck, unused, staticcheck issues in files you modified

# 5. Run tests
go test ./... -short

# 6. Run the full verification suite:
make check   # Runs: lint → test → security → build

# 7. Before committing, verify no secrets:
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" || echo "✓ Clean"
```

**⚠️ CRITICAL: Never skip linting (step 4). Fix ALL linting errors in code you wrote.**

### Test & Doc Requirements:
| Change | Unit Test | Integration Test | Update Docs |
|--------|-----------|------------------|-------------|
| New feature | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Bug fix | ✅ REQUIRED | ⚠️ If API-related | ⚠️ If behavior changes |
| New command | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Flag change | ✅ REQUIRED | ❌ Not needed | ✅ REQUIRED |

### Test Coverage Goals:

| Package Type | Minimum Coverage | Target Coverage |
|--------------|------------------|-----------------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

**How to check coverage:**
```bash
# Generate coverage report
go test ./... -short -coverprofile=coverage.out

# View coverage by package
go tool cover -func=coverage.out

# View detailed HTML report
go tool cover -html=coverage.out
```

**IMPORTANT**: New packages MUST have at least 70% test coverage before merging.

### Docs to Update (if applicable):
- `docs/COMMANDS.md` → New/changed commands or flags
- `docs/TIMEZONE.md` → Timezone-related changes, DST handling, calendar integration
- `docs/AI.md` → AI features, provider setup, privacy settings
- `plan.md` → Feature completed or API changes
- `AI_plan.md` → AI/timezone implementation status
- `README.md` → Major new features

**📋 IMPORTANT**: See `.claude/rules/documentation-maintenance.md` for complete documentation update requirements

### Workflow:
```bash
# 1. Make changes
# 2. Write tests in *_test.go
# 3. Format: go fmt ./...
# 4. Lint: golangci-lint run --timeout=5m
# 5. Fix ALL linting errors in your code (MANDATORY)
# 6. Test: go test ./... -short
# 7. Verify: make check
# 8. Verify no secrets in diff
# ⛔ DO NOT run git add, git commit, or git push
# → User will handle all git operations manually
```

**Quality Gate:** Code → Format → Lint → Fix → Test → Done
                                          ↑___|  (Loop until clean)

### Do Not Touch (without explicit permission):
| Path | Reason |
|------|--------|
| `.env*` | Contains secrets |
| `**/secrets/**` | Sensitive data |
| `*.pem`, `*.key` | Certificates/keys |
| `go.sum` | Auto-generated (only via `go mod tidy`) |
| `.git/` | Git internals |
| `vendor/` | Dependencies (if exists) |

### Repository Etiquette:

**Commit Messages:**
```
<type>: <short description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

Examples:
- `feat: add calendar availability check command`
- `fix: resolve nil pointer in email send`
- `docs: update COMMANDS.md with new flags`
- `test: add unit tests for contacts API`

**Branch Naming** (if creating branches):
- `feat/<feature-name>` - New features
- `fix/<bug-description>` - Bug fixes
- `docs/<what>` - Documentation updates

---

## Working with Claude (Tips)

### Recommended Workflow
```
1. Explore  →  "Read internal/cli/email/ and explain how send works"
2. Plan     →  "Think hard about how to add retry logic"
3. Code     →  "Implement the retry logic"
4. Lint     →  "Run golangci-lint and fix all issues"  ← NEW MANDATORY STEP
5. Test     →  "Run tests and fix any failures"
6. Commit   →  "Commit the changes"  ← NEVER auto-commit
```

**See `.claude/workflows/code-quality-checklist.md` for detailed linting guide.**

### Useful Commands
| Command | What It Does |
|---------|--------------|
| `/clear` | Reset context (use between unrelated tasks) |
| `/project:go-modernize` | Check Go docs & apply modern patterns |
| `/project:add-feature` | Structured feature workflow |
| `/project:fix-bug` | Bug fix workflow |
| `/project:review-pr` | Code review checklist |
| `/project:security-scan` | Security audit |
| `/project:smart-commit` | Generate commit message |

**IMPORTANT:** For all Go code changes, `/project:go-modernize` is automatically applied.
Claude will check go.dev/ref/spec and apply modern Go idioms before writing code.

**CRITICAL:** All skills (`add-feature`, `fix-bug`, etc.) now enforce linting:
- Linting runs automatically after code changes
- ALL linting errors in new/modified code MUST be fixed
- See `.claude/rules/linting.md` for common fixes

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `Escape` | Interrupt Claude mid-response |
| `Escape Escape` | Edit your previous message |
| `#` | Add instruction to CLAUDE.md |

### Test-Driven Development
```
1. "Write a failing test for <feature>"
2. Confirm test fails
3. "Implement the code to make the test pass"
4. Verify tests pass
5. "Commit with message: test: add tests for <feature>"
```

### Getting Fresh Reviews
For unbiased code review, use the `code-reviewer` subagent:
```
"Use the code-reviewer agent to review my changes"
```
This runs in isolated context (doesn't remember writing the code).

---

## Go Modernization Rules

**CRITICAL: Before writing ANY Go code, you MUST:**

### 1. Check Current Go Version
```bash
go version          # Check installed version
grep "^go " go.mod  # Check project version
```

### 2. Research Official Documentation
Use WebSearch to verify:
- **Go Spec**: https://go.dev/ref/spec - Language features
- **Pkg Docs**: https://pkg.go.dev/std - Standard library
- **Release Notes**: https://go.dev/doc/devel/release - Version features

### 3. Apply Modern Go Patterns (Go 1.21+)

| Instead of... | Use... | Since |
|---------------|--------|-------|
| `io/ioutil` | `os` package directly | Go 1.16+ |
| `interface{}` | `any` | Go 1.18+ |
| Manual slice ops | `slices` package | Go 1.21+ |
| Manual map ops | `maps` package | Go 1.21+ |
| Recreate to clear | `clear()` built-in | Go 1.21+ |
| Custom min/max | `min()`, `max()` built-ins | Go 1.21+ |
| Manual comparison | `cmp.Compare()` | Go 1.21+ |
| `sort.Slice` | `slices.SortFunc` | Go 1.21+ |

### 4. Examples

```go
// ✅ CORRECT (Modern Go 1.21+)
import (
    "os"
    "slices"
    "cmp"
)

// File operations
data, err := os.ReadFile("file.txt")

// Slice operations
found := slices.Contains(items, "target")

// Sorting
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Name, b.Name)
})

// Clearing
clear(myMap)

// Min/Max
smallest := min(a, b, c)

// ❌ WRONG (Deprecated/Verbose)
import "io/ioutil"

// Don't use deprecated packages
data, err := ioutil.ReadFile("file.txt")

// Don't write manual helpers
func Contains(items []string, target string) bool {
    for _, item := range items {
        if item == target {
            return true
        }
    }
    return false
}
```

### 5. Quality Checks (REQUIRED)
After any code changes:
```bash
go fmt ./...        # Format code
go vet ./...        # Vet code
golangci-lint run   # Lint (if available)
go test ./...       # Run tests
```

**See `.claude/rules/go-best-practices.md` for complete rules.**

---

## Project Overview

- **Language**: Go 1.24.0 (use modern features!)
- **Architecture**: Hexagonal (ports and adapters)
- **CLI Framework**: Cobra
- **API**: Nylas v3 ONLY (never use v1/v2)
- **AI Integration**: Multi-provider LLM support (Ollama, Claude, OpenAI, Groq) - **Planned** (see `AI_plan.md`)
- **Timezone Support**: Offline utilities + calendar integration with `--timezone` and `--show-tz` flags ✅

## Directory Structure

```
cmd/nylas/main.go          # Entry point - register commands here
internal/
  domain/                  # Business entities (Message, Event, Contact, etc.)
  ports/nylas.go           # Interface definitions
  adapters/nylas/          # API implementations
    client.go              # HTTP client
    mock.go                # Mock for testing
    demo.go                # Demo data for TUI
  cli/<feature>/           # CLI commands per feature
```

## Quick File Lookup

**When user asks about a feature, immediately know where to look:**

### By Feature / Command

| Feature | CLI Commands | Adapter | Domain Model | Tests |
|---------|-------------|---------|--------------|-------|
| **Email** | `internal/cli/email/` | `internal/adapters/nylas/messages.go`<br>`internal/adapters/nylas/drafts.go`<br>`internal/adapters/nylas/threads.go`<br>`internal/adapters/nylas/attachments.go` | `internal/domain/message.go`<br>`internal/domain/email.go` | `internal/cli/integration/email_test.go`<br>`internal/cli/integration/drafts_test.go`<br>`internal/cli/integration/threads_test.go` |
| **Calendar** (with timezone & breaks ⚡) | `internal/cli/calendar/`<br>`internal/cli/calendar/helpers.go` | `internal/adapters/nylas/calendars.go`<br>`internal/adapters/utilities/timezone/service.go` | `internal/domain/calendar.go`<br>`internal/domain/config.go` 📅 Working Hours & Breaks<br>`internal/domain/utilities.go` | `internal/cli/integration/calendar_test.go`<br>`internal/cli/calendar/helpers_test.go` |
| **Contacts** | `internal/cli/contacts/` | `internal/adapters/nylas/contacts.go` | `internal/domain/contact.go` | `internal/cli/integration/contacts_test.go` |
| **Auth** | `internal/cli/auth/` | `internal/adapters/nylas/auth.go` | `internal/domain/grant.go`<br>`internal/domain/provider.go` | `internal/cli/integration/auth_test.go` |
| **Webhooks** | `internal/cli/webhook/` | `internal/adapters/nylas/webhooks.go` | `internal/domain/webhook.go` | `internal/cli/integration/webhooks_test.go` |
| **Folders** | N/A (utility) | `internal/adapters/nylas/folders.go` | N/A | `internal/cli/integration/folders_test.go` |
| **Inbound** | `internal/cli/inbound/` | `internal/adapters/nylas/inbound.go` | `internal/domain/inbound.go` | `internal/cli/integration/inbound_test.go` |
| **Notetaker** | `internal/cli/notetaker/` | `internal/adapters/nylas/notetakers.go` | `internal/domain/notetaker.go` | N/A |
| **OTP** | `internal/cli/otp/` | `internal/adapters/nylas/otp.go` | N/A | `internal/adapters/nylas/otp_test.go` |
| **Timezone** ⚡ | `internal/cli/timezone/` | `internal/adapters/utilities/timezone/service.go` | `internal/domain/utilities.go` | `internal/cli/timezone/timezone_test.go`<br>`internal/cli/timezone/helpers_test.go`<br>`internal/adapters/utilities/timezone/service_test.go`<br>`internal/cli/integration/timezone_test.go` |
| **AI Scheduling** 🤖 | `internal/cli/calendar/ai_schedule.go`<br>`internal/cli/calendar/ai_*.go` | `internal/adapters/ai/`<br>`internal/ports/llm.go` | `internal/domain/ai.go` | `internal/adapters/ai/ollama_client_test.go`<br>`internal/adapters/ai/openai_client_test.go`<br>`internal/adapters/ai/claude_client_test.go`<br>`internal/adapters/ai/groq_client_test.go`<br>`internal/adapters/ai/router_test.go`<br>`internal/cli/integration/ai_test.go` |

### Core Files (Architecture Layers)

| Layer | File | Purpose |
|-------|------|---------|
| **Entry Point** | `cmd/nylas/main.go` | CLI entry point - register all commands here |
| **Root Command** | `internal/cli/root.go` | Root cobra command configuration |
| **Port Interface** | `internal/ports/nylas.go` | Interface contract - all adapter methods defined here |
| **HTTP Client** | `internal/adapters/nylas/client.go` | Base HTTP client, auth, request/response handling |
| **Mock Client** | `internal/adapters/nylas/mock.go` | Mock implementation for testing |
| **Demo Client** | `internal/adapters/nylas/demo.go` | Demo data for TUI mode |
| **Common Utils** | `internal/cli/common/` | Shared CLI utilities |
| **Errors** | `internal/domain/errors.go` | Domain-level error types |
| **Config** | `internal/domain/config.go` | Configuration models |

### CLI Package Pattern

Every CLI feature follows this pattern:
```
internal/cli/<feature>/
  ├── <feature>.go       # Main command definition
  ├── list.go            # List subcommand
  ├── show.go            # Show/Get subcommand
  ├── create.go          # Create subcommand
  ├── update.go          # Update subcommand
  ├── delete.go          # Delete subcommand
  └── helpers.go         # Shared helpers (getClient, getGrantID, etc.)
```

### Test Files

| Type | Pattern | Location |
|------|---------|----------|
| Unit tests | `*_test.go` | Alongside source files |
| Integration tests | `*_test.go` | `internal/cli/integration/` |
| Adapter tests | `*_test.go` | `internal/adapters/nylas/` |
| Domain tests | `domain_test.go` | `internal/domain/` |

### Utility Commands

| Command | File | Purpose |
|---------|------|---------|
| `nylas timezone` | `internal/cli/timezone/` | Offline timezone conversion, DST, meeting finder |
| `nylas doctor` | `internal/cli/doctor.go` | System diagnostics |
| `nylas version` | `internal/cli/version.go` | Version information |
| `nylas tui` | `internal/cli/tui.go` | Interactive TUI mode |

### Quick Navigation Examples

**User asks:** "Fix the email send command"
**Look in:** `internal/cli/email/send.go`

**User asks:** "Update the Calendar domain model"
**Look in:** `internal/domain/calendar.go`

**User asks:** "Add a new contact method to the API client"
**Look in:**
1. `internal/ports/nylas.go` (add interface method)
2. `internal/adapters/nylas/contacts.go` (implement method)
3. `internal/adapters/nylas/mock.go` (mock implementation)

**User asks:** "Where is authentication handled?"
**Look in:** `internal/cli/auth/` and `internal/adapters/nylas/auth.go`

**User asks:** "How do I test email functionality?"
**Look in:** `internal/cli/integration/email_test.go`

**User asks:** "How do I show calendar events in a different timezone?"
**Look in:** `internal/cli/calendar/helpers.go` for timezone conversion helpers, `internal/cli/calendar/events.go` for `--timezone` and `--show-tz` flags

## Adding a New Feature (Step-by-Step)

Example: Adding "widgets" feature

### 1. Domain Model
Create `internal/domain/widget.go`:
```go
package domain

type Widget struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type CreateWidgetRequest struct {
    Name string `json:"name"`
}
```

### 2. Port Interface
Add to `internal/ports/nylas.go`:
```go
// Widget operations
ListWidgets(ctx context.Context, grantID string) ([]domain.Widget, error)
GetWidget(ctx context.Context, grantID, widgetID string) (*domain.Widget, error)
CreateWidget(ctx context.Context, grantID string, req *domain.CreateWidgetRequest) (*domain.Widget, error)
DeleteWidget(ctx context.Context, grantID, widgetID string) error
```

### 3. Adapter Implementation
Create `internal/adapters/nylas/widgets.go`:
```go
package nylas

func (c *HTTPClient) ListWidgets(ctx context.Context, grantID string) ([]domain.Widget, error) {
    var resp struct {
        Data []domain.Widget `json:"data"`
    }
    if err := c.get(ctx, fmt.Sprintf("/grants/%s/widgets", grantID), &resp); err != nil {
        return nil, err
    }
    return resp.Data, nil
}
// ... implement other methods
```

### 4. Mock Implementation
Add to `internal/adapters/nylas/mock.go`:
```go
func (m *MockClient) ListWidgets(ctx context.Context, grantID string) ([]domain.Widget, error) {
    return []domain.Widget{{ID: "widget-1", Name: "Test Widget"}}, nil
}
```

### 5. Demo Implementation
Add to `internal/adapters/nylas/demo.go`:
```go
func (d *DemoClient) ListWidgets(ctx context.Context, grantID string) ([]domain.Widget, error) {
    return []domain.Widget{
        {ID: "demo-widget-1", Name: "Demo Widget"},
    }, nil
}
```

### 6. CLI Commands
Create `internal/cli/widget/widget.go`:
```go
package widget

import "github.com/spf13/cobra"

func NewWidgetCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "widget",
        Short: "Manage widgets",
    }
    cmd.AddCommand(newListCmd())
    cmd.AddCommand(newShowCmd())
    // ... add other subcommands
    return cmd
}
```

Create `internal/cli/widget/list.go`, `show.go`, etc.

### 7. Register Command
Add to `cmd/nylas/main.go`:
```go
import "github.com/mqasimca/nylas/internal/cli/widget"

rootCmd.AddCommand(widget.NewWidgetCmd())
```

### 8. Tests
- Unit tests: `internal/cli/widget/widget_test.go`
- Integration tests: Add to `internal/adapters/nylas/integration_test.go`

### 9. Documentation
Update `docs/COMMANDS.md` with new command examples.

## Common Patterns

### CLI Helper Functions
Each CLI package has helpers for:
- `getClient()` - Create authenticated Nylas client
- `getGrantID(args)` - Get grant ID from args or default
- `createContext()` - Create context with timeout

### Standard Flags
- `--json` - Output as JSON
- `--limit` - Limit results
- `--yes` / `-y` - Skip confirmation

### Error Handling
```go
if err != nil {
    return fmt.Errorf("failed to do X: %w", err)
}
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./... -short

# Run tests for specific package
go test ./internal/cli/widget/... -v

# Run with coverage
go test ./... -short -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

**Location**: All integration tests are in `internal/cli/integration/`

Integration tests require valid Nylas API credentials:

```bash
# Set environment variables
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"

# Run all integration tests
go test -tags=integration ./internal/cli/integration/...

# Run specific integration test
go test -tags=integration ./internal/cli/integration/ -run TestAuth

# Run integration tests with verbose output
go test -tags=integration -v ./internal/cli/integration/...

# Run with timeout for long-running tests
go test -tags=integration -v -timeout 30m ./internal/cli/integration/...
```

**Integration Test Checklist:**
- [ ] Tests tagged with `//go:build integration` and `// +build integration`
- [ ] Tests in `internal/cli/integration/` directory
- [ ] Tests use `package integration`
- [ ] Tests skip when credentials missing: `if testAPIKey == "" { t.Skip() }`
- [ ] Tests clean up resources using `t.Cleanup()`
- [ ] Tests handle API rate limits gracefully
- [ ] Tests don't assume test account state

**See**: `internal/cli/integration/README.md` for detailed documentation

## Pre-Commit Hook (Recommended)

To automatically run checks before each commit, create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Nylas CLI pre-commit hook

echo "Running pre-commit checks..."

# 1. Format code
echo "→ Formatting code..."
go fmt ./...

# 2. Run linter
echo "→ Running linter..."
if ! golangci-lint run --timeout=5m; then
    echo "❌ Linting failed. Fix errors before committing."
    exit 1
fi

# 3. Run tests
echo "→ Running tests..."
if ! go test ./... -short; then
    echo "❌ Tests failed. Fix tests before committing."
    exit 1
fi

# 4. Security check
echo "→ Running security scan..."
if ! make security; then
    echo "❌ Security scan failed. Check for secrets."
    exit 1
fi

echo "✅ All pre-commit checks passed!"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Quick Commands

```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linter
./bin/nylas --help  # Test CLI
```

## Files to Check When Debugging

1. `internal/ports/nylas.go` - Interface contract
2. `internal/adapters/nylas/client.go` - API base URL, auth
3. `internal/cli/<feature>/helpers.go` - Client creation, grant resolution

## API Reference

- Docs: https://developer.nylas.com/docs/api/v3/
- Base URL: `https://api.us.nylas.com/v3/`
- Auth: Bearer token via `Authorization` header
