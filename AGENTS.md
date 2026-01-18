# AI Coding Agent Guidelines

Quick reference for AI coding agents (Cursor, Copilot, Windsurf, etc.) working on this Go CLI codebase.

## Build, Lint, Test Commands

### Quick Validation (Use Before Commits)
```bash
make ci          # Format, vet, lint, unit tests, race detection, security, vuln, build
make ci-full     # Complete CI: ci + integration tests + cleanup
```

### Running Tests
```bash
# All unit tests
make test-unit

# Specific package
make test-pkg PKG=email

# Single test by name
go test ./internal/cli/email/... -v -run TestSpecificName

# With race detection
go test ./internal/cli/email/... -v -race -run TestSpecificName

# Integration tests (requires NYLAS_API_KEY, NYLAS_GRANT_ID)
make test-integration
```

### Build
```bash
make build       # Build binary to bin/nylas
make install     # Install to GOPATH/bin
```

## Code Style

### Go Version
- **Go 1.24.2** - Use modern features:
  - `any` instead of `interface{}`
  - `slices` and `maps` packages instead of manual loops
  - Generic functions where appropriate

### Imports (Ordered Groups)
```go
import (
    "context"           // 1. Standard library
    "fmt"

    "github.com/spf13/cobra"  // 2. External packages

    "github.com/mqasimca/nylas/internal/ports"  // 3. Internal packages
)
```

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to fetch emails: %w", err)
}

// Check errors immediately, don't defer
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### Testing
```go
// Always use table-driven tests with t.Run()
func TestFormatSize(t *testing.T) {
    tests := []struct {
        name     string
        input    int64
        expected string
    }{
        {"zero bytes", 0, "0 B"},
        {"kilobytes", 1024, "1.0 KB"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatSize(tt.input)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### File Size Limits
- **Ideal:** ≤500 lines per file
- **Maximum:** ≤600 lines per file
- Split large files by responsibility (helpers, types, handlers)

## Project Patterns

### Use Shared Helpers (Don't Create Duplicates)

```go
// CLI client - use directly, no package wrappers
client := common.GetNylasClient()

// Grant ID from args
grantID := common.GetGrantID(args)

// Output formatting
common.PrintSuccess("Email sent successfully")
common.PrintError("Failed to send email", err)
common.FormatSize(bytes)      // "1.5 MB"
common.FormatTimeAgo(time)    // "2 hours ago"
common.PrintJSON(data)        // Pretty-print JSON

// Structured output (use in list commands)
out := common.GetOutputWriter(cmd)  // Gets writer based on --json/--yaml/--quiet
out.Write(data)                     // Outputs in correct format

// Client helpers (reduce boilerplate)
common.WithClient(args, func(ctx, client, grantID) (T, error) {
    return client.DoSomething(ctx, grantID)
})
common.WithClientNoGrant(func(ctx, client) (T, error) {
    return client.DoSomething(ctx)
})

// Flag helpers (use instead of inline flag definitions)
common.AddJSONFlag(cmd, &jsonOutput)   // --json
common.AddLimitFlag(cmd, &limit, 25)   // --limit/-n
common.AddYesFlag(cmd, &yes)           // --yes/-y
common.AddFormatFlag(cmd, &format)     // --format/-f
common.AddIDFlag(cmd, &showID)         // --id
common.AddPageTokenFlag(cmd, &token)   // --page-token

// Validation helpers (use instead of inline checks)
common.ValidateRequired("event ID", eventID)
common.ValidateRequiredFlag("--to", toEmail)
common.ValidateRequiredArg(args, "message ID")
common.ValidateURL("webhook URL", webhookURL)
common.ValidateEmail("recipient", email)
common.ValidateOneOf("status", status, []string{"pending", "active"})
common.ValidateAtLeastOne("update field", url, description, status)

// HTTP handlers (in adapters)
httputil.WriteJSON(w, http.StatusOK, data)
body, err := httputil.LimitedBody(r, maxSize)
httputil.DecodeJSON(r, &target)
```

### AI Client Helpers
```go
// In adapters/ai/ - use shared base_client.go helpers
ConvertMessagesToMaps(messages)
ConvertToolsOpenAIFormat(tools)
FallbackStreamChat(ctx, messages, opts)
```

## Architecture

Hexagonal architecture with three layers:

```
CLI (internal/cli/)
    ↓ calls
Ports (internal/ports/) - Interfaces
    ↓ implemented by
Adapters (internal/adapters/) - Implementations
```

### Key Packages
| Package | Purpose |
|---------|---------|
| `internal/domain/` | Domain types (Email, Calendar, etc.) |
| `internal/ports/nylas.go` | Main NylasClient interface |
| `internal/ports/output.go` | OutputWriter interface |
| `internal/adapters/nylas/` | Nylas API client implementation |
| `internal/adapters/output/` | Table, JSON, YAML, Quiet formatters |
| `internal/httputil/` | HTTP response helpers |
| `internal/cli/common/` | Shared CLI helpers |
| `internal/air/` | Web email client |

### Adding a New Feature
1. **Domain:** `internal/domain/<feature>.go` - Define types
2. **Port:** `internal/ports/nylas.go` - Add interface methods
3. **Adapter:** `internal/adapters/nylas/<feature>.go` - Implement
4. **Mock:** `internal/adapters/nylas/mock.go` - Add mock methods
5. **CLI:** `internal/cli/<feature>/` - Add commands
6. **Register:** `cmd/nylas/main.go` - Wire command
7. **Tests:** Unit + integration tests
8. **Docs:** Update `docs/COMMANDS.md`

## Do Not Modify
- `.env*`, `**/secrets/**` - Contains secrets
- `*.pem`, `*.key` - Certificates
- `go.sum` - Auto-generated
- `.git/`, `vendor/` - Managed externally

## Credential Storage (Keyring)

Credentials are stored securely in the system keyring under service name `"nylas"`.

### Keys Stored
| Key | Constant | Description |
|-----|----------|-------------|
| `client_id` | `ports.KeyClientID` | Nylas Application/Client ID (auto-detected or manual) |
| `api_key` | `ports.KeyAPIKey` | Nylas API key (required, used for Bearer auth) |
| `client_secret` | `ports.KeyClientSecret` | Provider OAuth client secret (Google/Microsoft), optional |
| `org_id` | `ports.KeyOrgID` | Nylas Organization ID (auto-detected) |
| `grants` | `grantsKey` | JSON array of grant info (ID, email, provider) |
| `default_grant` | `defaultGrantKey` | Default grant ID for CLI operations |
| `grant_token_<id>` | `ports.GrantTokenKey()` | Per-grant access tokens |

### Key Files
| File | Purpose |
|------|---------|
| `internal/ports/secrets.go` | Key constants (`KeyClientID`, `KeyAPIKey`, etc.) |
| `internal/adapters/keyring/keyring.go` | System keyring implementation |
| `internal/adapters/keyring/grants.go` | Grant storage (`grants`, `default_grant` keys) |
| `internal/app/auth/config.go` | `SetupConfig()` saves credentials |

### Platform Storage
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **macOS**: Keychain
- **Windows**: Windows Credential Manager
- **Fallback**: Encrypted file store (`~/.config/nylas/`)

### Environment Override
Set `NYLAS_DISABLE_KEYRING=true` to force encrypted file store (useful for testing).

## API Reference
- **Nylas API v3 ONLY** - Never use v1/v2
- Docs: https://developer.nylas.com/docs/api/v3/
