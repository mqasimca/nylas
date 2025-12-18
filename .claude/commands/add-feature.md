# Add Feature

Add a complete new feature to the Nylas CLI following hexagonal architecture.

Feature: $ARGUMENTS

## Instructions

1. **Analyze the feature request** - Understand what API endpoints and CLI commands are needed

2. **Create domain model** in `internal/domain/<feature>.go`:
   - Define structs with JSON tags
   - Add request/response types
   - Add any constants (states, types)

3. **Add port interface** methods to `internal/ports/nylas.go`:
   - List, Get, Create, Update, Delete as needed
   - Use context.Context as first parameter
   - Return domain types

4. **Implement adapter** in `internal/adapters/nylas/<feature>.go`:
   - Implement all port methods
   - Use existing HTTP client patterns
   - Handle API response parsing

5. **Add mock implementation** to `internal/adapters/nylas/mock.go`:
   - Return test data for each method

6. **Add demo implementation** to `internal/adapters/nylas/demo.go`:
   - Return realistic demo data for TUI

7. **Create CLI package** `internal/cli/<feature>/`:
   - `<feature>.go` - Main command with NewXxxCmd()
   - `list.go` - List subcommand
   - `show.go` - Show/get subcommand
   - `create.go` - Create subcommand
   - `delete.go` - Delete subcommand
   - `helpers.go` - getClient(), getGrantID(), createContext()

8. **Register command** in `cmd/nylas/main.go`:
   - Import the new package
   - Add rootCmd.AddCommand()

9. **Add tests**:
   - Unit tests: `internal/cli/<feature>/<feature>_test.go`
   - Demo tests: `internal/adapters/nylas/demo_test.go`
   - Integration tests: `internal/adapters/nylas/integration_test.go`

10. **Update documentation**:
    - Add section to `docs/COMMANDS.md`
    - Update `plan.md` if applicable

## Checklist
- [ ] Domain model created
- [ ] Port interface added
- [ ] Adapter implemented
- [ ] Mock implementation added
- [ ] Demo implementation added
- [ ] CLI commands created
- [ ] Command registered in main.go
- [ ] **Unit tests written and pass**: `go test ./... -short`
- [ ] **Integration tests written and pass**: `go test ./... -tags=integration`
- [ ] **Linting passes**: `golangci-lint run`
- [ ] **Security scan passes**: `make security`
- [ ] **Documentation updated** (if user-facing changes):
  - [ ] `docs/COMMANDS.md` - New commands/flags/examples
  - [ ] `plan.md` - Mark feature complete, update API status
  - [ ] `README.md` - If major feature
- [ ] Build succeeds: `make build`
- [ ] Full check passes: `make check`

## ⛔ MANDATORY - Before Committing:
```bash
# Run full verification
make check

# Verify no secrets
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" && echo "STOP!" || echo "✓ OK"

# ⛔ NEVER run git push - only local commits
```
