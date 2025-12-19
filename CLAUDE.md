# AI Assistant Guide for Nylas CLI

Quick reference for AI assistants working on this codebase.

---

## ⛔ CRITICAL RULES - MUST FOLLOW

### NEVER DO:
- **NEVER run `git push`** - Only local commits allowed
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
# 3. Run the full verification suite:
make check   # Runs: lint → test → security → build

# 4. Before committing, verify no secrets:
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" || echo "✓ Clean"
```

### Test & Doc Requirements:
| Change | Unit Test | Integration Test | Update Docs |
|--------|-----------|------------------|-------------|
| New feature | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Bug fix | ✅ REQUIRED | ⚠️ If API-related | ⚠️ If behavior changes |
| New command | ✅ REQUIRED | ✅ REQUIRED | ✅ REQUIRED |
| Flag change | ✅ REQUIRED | ❌ Not needed | ✅ REQUIRED |

### Docs to Update (if applicable):
- `docs/COMMANDS.md` → New/changed commands or flags
- `plan.md` → Feature completed or API changes
- `README.md` → Major new features

### Commit Workflow:
```bash
# 1. Make changes
# 2. Write tests in *_test.go
# 3. Run: make check
# 4. Verify no secrets in diff
# 5. git add <files>
# 6. git commit -m "message"
# ⛔ DO NOT run git push
```

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
4. Test     →  "Run tests and fix any failures"
5. Commit   →  "Commit the changes"
```

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

```bash
# Unit tests
go test ./... -short

# Integration tests (requires credentials)
NYLAS_API_KEY="..." NYLAS_GRANT_ID="..." go test ./... -tags=integration

# Specific package
go test ./internal/cli/widget/... -v
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
