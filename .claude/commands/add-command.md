# Add New CLI Command

Create a new CLI command following the nylas CLI patterns and hexagonal architecture.

## Instructions

1. First, ask me for:
   - Command name (e.g., "resource", "attachment")
   - Parent command if subcommand (e.g., "email" for "email attachments")
   - What operations it needs (list, show, create, update, delete)
   - Brief description of what it does

2. Then create the following files in order:

### If new domain types needed:
- `internal/domain/{resource}.go` - Domain types
- Update `internal/domain/domain_test.go` - Add tests

### If new API methods needed:
- Update `internal/ports/nylas.go` - Add interface methods
- `internal/adapters/nylas/{resource}.go` - Implement methods
- Update `internal/adapters/nylas/mock.go` - Add mock methods

### CLI package:
- `internal/cli/{resource}/{resource}.go` - Root command with New{Resource}Cmd()
- `internal/cli/{resource}/list.go` - newListCmd() if needed
- `internal/cli/{resource}/show.go` - newShowCmd() if needed
- `internal/cli/{resource}/create.go` - newCreateCmd() if needed
- `internal/cli/{resource}/helpers.go` - getClient(), getGrantID(), createContext()
- `internal/cli/{resource}/{resource}_test.go` - Unit tests

### Registration:
- Update `cmd/nylas/main.go` to add the command

3. Follow these patterns:
   - Use cobra command structure from existing commands
   - Support --format flag (table, json, yaml) for list/show
   - Use context with 30s timeout for API calls
   - Use spinner for long operations
   - Provide helpful error messages with suggestions

4. After creating, run:
   - `go build ./...` to verify compilation
   - `go test ./internal/cli/{resource}/...` to run tests
