# Nylas CLI - Claude Code Rules

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

### 3. Testing Requirements
- **Unit tests**: Test command structure, flags, and help output
- **Mock implementations**: Every adapter has a `mock.go` file
- **Integration tests**: Use build tag `//go:build integration`
- **Table-driven tests**: Use for parameter variations
- Run `go test ./...` before committing

### 4. File Organization
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
   - [ ] Add tests

4. **CLI** (`internal/cli/{resource}/`)
   - [ ] Create package with root command
   - [ ] Add subcommands (list, show, create, etc.)
   - [ ] Add `helpers.go` with getClient(), getGrantID()
   - [ ] Add tests in `{resource}_test.go`

5. **Registration**
   - [ ] Add to `cmd/nylas/main.go`: `rootCmd.AddCommand(resource.NewResourceCmd())`

6. **Documentation**
   - [ ] Update docs/COMMANDS.md with new commands (or docs/TUI.md for TUI changes)

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

## Common Files Reference

| File | Purpose |
|------|---------|
| `internal/domain/errors.go` | All domain errors |
| `internal/ports/nylas.go` | Main client interface |
| `internal/cli/common/errors.go` | CLI error wrapping |
| `internal/cli/common/format.go` | Output formatting utilities |
| `internal/cli/common/progress.go` | Spinner, progress bar |
| `internal/adapters/nylas/mock.go` | Mock client for tests |
