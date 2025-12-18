# Nylas CLI - Claude Code Rules

## ⛔ MANDATORY REQUIREMENTS - READ FIRST

### DO NOT EVER:
1. **NEVER push to GitHub** - Only create local commits. NEVER run `git push`.
2. **NEVER commit secrets** - No API keys, passwords, tokens, .env files, or credentials.
3. **NEVER skip tests** - All code changes require tests to pass.
4. **NEVER skip security scans** - All changes must pass security analysis.

### ALWAYS DO (in this order):

#### After ANY Code Change:
```bash
# 1. Run unit tests
go test ./... -short

# 2. Run linting
golangci-lint run

# 3. Run security scan
make security

# 4. Build verification
make build
```

#### For API/Adapter Changes:
```bash
# Also run integration tests (if credentials available)
go test ./... -tags=integration -v
```

#### Before ANY Commit:
```bash
# 1. Full check (lint + test + security + build)
make check

# 2. Verify no secrets in staged files
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" && echo "⛔ SECRETS DETECTED - DO NOT COMMIT" || echo "✓ No secrets found"

# 3. Check for sensitive file types
git diff --cached --name-only | grep -E "\.(env|pem|key|credentials)$" && echo "⛔ SENSITIVE FILE - DO NOT COMMIT" || echo "✓ No sensitive files"
```

### Test Requirements by Change Type:

| Change Type | Unit Tests | Integration Tests | Security Scan | Update Docs |
|------------|------------|-------------------|---------------|-------------|
| New feature | ✅ Required | ✅ Required | ✅ Required | ✅ Required |
| Bug fix | ✅ Required | ⚠️ If API-related | ✅ Required | ⚠️ If behavior changes |
| Refactor | ✅ Required | ⚠️ If API-related | ✅ Required | ⚠️ If API changes |
| New command | ✅ Required | ✅ Required | ✅ Required | ✅ Required |
| Flag change | ✅ Required | ❌ Not needed | ✅ Required | ✅ Required |
| Docs only | ❌ Not needed | ❌ Not needed | ❌ Not needed | N/A |

### Documentation Updates (MANDATORY when applicable):

Update these docs when code changes affect user-facing behavior:

| Doc File | Update When |
|----------|-------------|
| `docs/COMMANDS.md` | New/changed commands, flags, or examples |
| `plan.md` | Feature completed or API status changes |
| `README.md` | Major new features or installation changes |
| `docs/TUI.md` | TUI changes, new keybindings, new views |

**Check if docs need updating:**
```bash
# If you changed CLI commands, check COMMANDS.md
git diff --name-only | grep -E "internal/cli/" && echo "→ Review docs/COMMANDS.md"

# If you added new feature, check plan.md
git diff --name-only | grep -E "internal/(adapters|domain)/" && echo "→ Review plan.md"
```

### Writing Tests:

When adding a feature or fixing a bug, you MUST:
1. **Write unit tests** in `*_test.go` alongside the code
2. **Use table-driven tests** for multiple scenarios
3. **Test error cases** not just happy paths
4. **Update mock.go** if adding new interface methods

Example test file location:
- Feature code: `internal/cli/email/send.go`
- Test file: `internal/cli/email/send_test.go` or `internal/cli/email/email_test.go`

---

## Project Overview

This is a Go CLI application for the **Nylas v3 API** following **hexagonal architecture** (ports and adapters pattern). The CLI provides email, calendar, contacts, webhooks, and OTP management functionality.

**IMPORTANT: This CLI supports ONLY Nylas v3 API. Do NOT use v1 or v2 API documentation.**

## Nylas API Documentation

When looking up Nylas API documentation, ONLY use v3 API docs:

- **Official v3 API Docs**: https://developer.nylas.com/docs/api/v3/
- **v3 API Reference**: https://developer.nylas.com/docs/api/v3/ecc/
- **v3 Authentication**: https://developer.nylas.com/docs/v3/auth/
- **v3 Quickstart**: https://developer.nylas.com/docs/v3/quickstart/

**Base URLs (v3 only):**
- US Region: `https://api.us.nylas.com/v3/`
- EU Region: `https://api.eu.nylas.com/v3/`

**DO NOT reference:**
- Any `/v1/` or `/v2/` endpoints
- Legacy Nylas documentation
- Deprecated authentication methods

## Architecture Layers

```
CLI Commands → App Services → Adapters → Ports (interfaces)
                    ↑
               Domain Models
```

### Layer Responsibilities

| Layer | Location | Purpose |
|-------|----------|---------|
| **Domain** | `internal/domain/` | Business entities, errors, constants |
| **Ports** | `internal/ports/` | Interface contracts (never implementations) |
| **Adapters** | `internal/adapters/` | External implementations (API, keyring, config) |
| **App** | `internal/app/` | Business logic orchestration |
| **CLI** | `internal/cli/` | Cobra commands and user interaction |

## Critical Rules

### 1. Dependency Direction
- **ALWAYS** depend on interfaces (ports), never on concrete implementations
- Services receive ports via constructor injection
- CLI commands use helper functions to create services with proper dependencies

### 2. Error Handling
- Define sentinel errors in `internal/domain/errors.go`
- Wrap errors with context using `fmt.Errorf("%w: details", domain.ErrXxx, ...)`
- Map domain errors to user-friendly CLI errors in `internal/cli/common/errors.go`
- Always provide actionable suggestions in error messages

### 3. Testing Requirements (MANDATORY)
- **ALWAYS run tests after ANY code change**: `go test ./... -short`
- **ALWAYS run integration tests for API changes**: `go test ./... -tags=integration`
- **Unit tests**: Test command structure, flags, and help output
- **Mock implementations**: Every adapter has a `mock.go` file
- **Integration tests**: Use build tag `//go:build integration`
- **Table-driven tests**: Use for parameter variations
- Run `make check` before committing (lint + test + build)

### 4. Security Requirements (MANDATORY)
- **ALWAYS run security scan after ANY code change**
- **Check for hardcoded credentials**: `grep -rE "nyk_v0|api_key\s*=|password\s*=|secret\s*=" --include="*.go" .`
- **Check for credential logging**: `grep -rE "fmt\.(Print|Log).*([Aa]pi[Kk]ey|[Pp]assword|[Ss]ecret)" --include="*.go" .`
- **Verify no secrets in git history**: `git log --all --name-only --pretty=format: | grep -E "\.(sh|json)$" | sort -u`
- **Never commit**: API keys, passwords, tokens, .env files, credential files
- **Always use**: Environment variables for test credentials, keyring for storage
- **Before pushing**: Verify no sensitive files with `git diff --name-only origin/main...HEAD`

### 5. File Organization
```
internal/cli/{feature}/
├── {feature}.go      # Root command with NewXxxCmd()
├── list.go           # newListCmd()
├── show.go           # newShowCmd()
├── create.go         # newCreateCmd()
├── update.go         # newUpdateCmd()
├── delete.go         # newDeleteCmd()
├── helpers.go        # getClient(), getGrantID(), createContext()
└── {feature}_test.go # Unit tests
```

## Naming Conventions

### Functions
```go
// Public command constructors
func NewAuthCmd() *cobra.Command      // Exported, returns root command
func NewEmailCmd() *cobra.Command

// Private command helpers
func newLoginCmd() *cobra.Command     // Unexported, individual commands
func newListCmd() *cobra.Command
```

### Types
```go
// Domain models - singular nouns
type Grant struct { ... }
type Message struct { ... }

// Request/Response types
type CreateWebhookRequest struct { ... }
type MessageListResponse struct { ... }

// Query parameters
type MessageQueryParams struct { ... }
```

### Errors
```go
// Sentinel errors in domain package
var ErrNotConfigured = errors.New("nylas not configured")
var ErrGrantNotFound = errors.New("grant not found")

// CLI error codes
const ErrCodeNotConfigured = "E001"
```

## Command Template

When creating new commands, follow this pattern:

```go
func newListCmd() *cobra.Command {
    var (
        limit  int
        format string
    )

    cmd := &cobra.Command{
        Use:     "list [grant-id]",
        Aliases: []string{"ls"},
        Short:   "Short description (one line)",
        Long:    `Detailed multi-line description...`,
        Example: `  # Example 1
  nylas resource list

  # Example 2
  nylas resource list --format json`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            client, err := getClient()
            if err != nil {
                return err
            }

            grantID, err := getGrantID(args)
            if err != nil {
                return err
            }

            ctx, cancel := createContext()
            defer cancel()

            // Business logic here
            return nil
        },
    }

    cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum items to return")
    cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")

    return cmd
}
```

## Documentation Updates (MANDATORY)

**CRITICAL: Always update documentation when adding or modifying commands/features.**

When you add, modify, or remove ANY command or feature:

1. **docs/COMMANDS.md** - Update command reference
   - Add new commands with syntax and examples
   - Update existing command flags/options
   - Add example output for new features

2. **docs/plan.md** - Update feature status
   - Mark completed items as `[x]`
   - Update API endpoint status tables
   - Update implementation status in header table

3. **README.md** - Update if major features added
   - Update feature list
   - Update quick start examples if needed

4. **docs/TUI.md** - Update if TUI changes
   - New keyboard shortcuts
   - New views or panels

**Documentation files:**
- `docs/COMMANDS.md` - Complete CLI command reference
- `docs/plan.md` - Feature roadmap and API coverage
- `docs/TUI.md` - TUI documentation
- `docs/ARCHITECTURE.md` - Architecture documentation
- `README.md` - Project overview

## Adding New Features Checklist

1. **Domain** (`internal/domain/`)
   - [ ] Add types in `{resource}.go`
   - [ ] Add errors if needed in `errors.go`
   - [ ] Add tests in `domain_test.go`

2. **Ports** (`internal/ports/nylas.go`)
   - [ ] Add methods to `NylasClient` interface

3. **Adapter** (`internal/adapters/nylas/`)
   - [ ] Implement methods in `client.go` or new file
   - [ ] Add mock methods in `mock.go`
   - [ ] Add demo methods in `demo.go`
   - [ ] Add tests

4. **CLI** (`internal/cli/{resource}/`)
   - [ ] Create package with root command
   - [ ] Add subcommands (list, show, create, etc.)
   - [ ] Add `helpers.go` with getClient(), getGrantID()
   - [ ] Add unit tests in `{resource}_test.go`
   - [ ] Add integration tests in `integration_{resource}_test.go`

5. **Registration**
   - [ ] Add to `cmd/nylas/main.go`: `rootCmd.AddCommand(resource.NewResourceCmd())`

6. **Documentation** (MANDATORY)
   - [ ] Update `docs/COMMANDS.md` with new commands, flags, and examples
   - [ ] Update `docs/plan.md` to mark features complete and update API status
   - [ ] Update `README.md` if adding major features

## Output Formatting

Support multiple output formats in list/show commands:

```go
switch format {
case "json":
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(data)
case "yaml":
    return yaml.NewEncoder(os.Stdout).Encode(data)
case "csv":
    return outputCSV(data)
default:
    return outputTable(data)
}
```

## Common Patterns

### Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### Optional Boolean Flags
```go
var unread bool
cmd.Flags().BoolVar(&unread, "unread", false, "Filter unread only")

// In RunE:
params := &domain.MessageQueryParams{}
if cmd.Flags().Changed("unread") {
    params.Unread = &unread  // Only set if explicitly provided
}
```

### Progress Indication
```go
spinner := common.NewSpinner("Fetching data...")
spinner.Start()
defer spinner.Stop()

// Long operation here
```

### Color Output
```go
green := color.New(color.FgGreen)
cyan := color.New(color.FgCyan)
dim := color.New(color.Faint)

green.Printf("✓ Success!\n")
cyan.Printf("ID: %s\n", id)
```

## Integration Tests

Integration tests require environment variables:

```bash
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"
export NYLAS_TEST_BINARY="/path/to/bin/nylas"

go test -tags=integration ./internal/cli/...
```

## Code Quality

- **No hardcoded credentials** - Use keyring/config
- **Context everywhere** - All API calls accept context
- **Graceful degradation** - Handle missing optional features
- **Consistent formatting** - Run `gofmt` before commit
- **Lint clean** - Run `golangci-lint run`

## File Size Limits

**CRITICAL: Files must not exceed 25,000 lines.**

When a file approaches this limit:
1. **Split into focused files** - One responsibility per file
2. **Group by feature** - e.g., `messages.go`, `drafts.go`, `attachments.go`
3. **Keep related code together** - Tests in `*_test.go` files
4. **Share common code** - Extract helpers to `helpers.go` or `common/`

Example split for a growing `client.go`:
```
internal/adapters/nylas/
├── client.go          # Core client, auth, base URL
├── messages.go        # Message CRUD operations
├── drafts.go          # Draft CRUD operations
├── attachments.go     # Attachment operations
├── threads.go         # Thread operations
├── folders.go         # Folder operations
├── calendars.go       # Calendar operations
├── events.go          # Event operations
├── contacts.go        # Contact operations
├── webhooks.go        # Webhook operations
├── mock.go            # Mock implementations
├── demo.go            # Demo data for screenshots
└── *_test.go          # Test files
```

**Before writing to any file:**
1. Check current line count
2. If approaching 20,000 lines, proactively split
3. Never write to a file that would exceed 25,000 lines

## Common Files Reference

| File | Purpose |
|------|---------|
| `internal/domain/errors.go` | All domain errors |
| `internal/ports/nylas.go` | Main client interface |
| `internal/cli/common/errors.go` | CLI error wrapping |
| `internal/cli/common/format.go` | Output formatting utilities |
| `internal/cli/common/progress.go` | Spinner, progress bar |
| `internal/adapters/nylas/mock.go` | Mock client for tests |
