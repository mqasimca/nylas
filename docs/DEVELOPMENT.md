# Development Guide

Guide for setting up the development environment, running tests, and building the CLI.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [TUI](TUI.md) | [Architecture](ARCHITECTURE.md) | [Security](SECURITY.md) | [Webhooks](WEBHOOKS.md)

---

## Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

## Building

```bash
# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

---

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests are split into multiple files by command for better maintainability:

| File | Description |
|------|-------------|
| `integration_test.go` | Common setup, helpers |
| `integration_auth_test.go` | Auth command tests |
| `integration_email_test.go` | Email list, read, search, mark, send tests |
| `integration_folders_test.go` | Folder command tests |
| `integration_threads_test.go` | Thread command tests |
| `integration_drafts_test.go` | Draft command tests |
| `integration_calendar_test.go` | Calendar & availability tests |
| `integration_contacts_test.go` | Contact command tests |
| `integration_webhooks_test.go` | Webhook command tests |
| `integration_misc_test.go` | OTP, help, errors, workflow, doctor tests |

```bash
# Run integration tests (requires NYLAS_API_KEY and NYLAS_GRANT_ID)
go test -tags=integration ./internal/cli/...

# Optional: specify custom binary path
NYLAS_TEST_BINARY=/path/to/nylas go test -tags=integration ./internal/cli/...
```

### Test Categories

The test suite includes:

| Category | Tests |
|----------|-------|
| Authentication | Login, logout, status, multi-account |
| Messages | List, get, search, filter, send, schedule |
| Mark Operations | Mark read/unread, star/unstar |
| Folders | List, create, rename, delete |
| Threads | List, get, update status |
| Drafts | Create, update, delete lifecycle |
| Calendar | List calendars, events CRUD |
| Availability | Free/busy check, find slots |
| Contacts | List, get, create, delete |
| Contact Groups | List groups |
| Webhooks | CRUD, triggers, test events |
| Attachments | Get metadata, download content |
| Concurrency | Parallel requests |
| Error Handling | Invalid IDs, timeouts |

### TUI Tests

The TUI has comprehensive integration tests:

```bash
# Run TUI tests
go test ./internal/tui/... -v

# Run with coverage
go test ./internal/tui/... -cover
```

| Category | Tests |
|----------|-------|
| Themes | All built-in themes, custom theme loading, color parsing, YAML config, validation, set-default |
| Theme CLI | Theme list, validate, init, set-default commands, error handling |
| Demo Mode | Demo client, demo data (messages, events, contacts, webhooks, grants), demo flag on commands |
| Vim Commands | `gg`, `dd`, `:q`, `:wq`, numeric jump, key sequences |
| App Initialization | NewApp with all themes, config, styles |
| Views | Dashboard, Messages, Events, Contacts, Webhooks, Grants |
| Table | Columns, data, selection, row metadata |
| Compose | New, Reply, Reply All, Forward modes |
| Utilities | parseRecipients, convertToHTML, formatDate, formatParticipants |
| Key Handling | Escape, Enter, Tab, vim keys (j/k/g/G), Ctrl+d/u/f/b |

---

## Project Structure

```
.
├── cmd/
│   └── nylas/
│       └── main.go          # Entry point
├── internal/
│   ├── domain/              # Domain models
│   ├── ports/               # Interfaces
│   ├── adapters/            # Implementations
│   ├── app/                 # Application services
│   └── cli/                 # CLI commands
│       ├── auth/            # Authentication commands
│       ├── email/           # Email commands
│       ├── calendar/        # Calendar commands
│       ├── contacts/        # Contacts commands
│       ├── webhook/         # Webhook commands
│       ├── otp/             # OTP commands
│       └── common/          # Shared utilities
├── docs/                    # Documentation
├── Makefile                 # Build automation
├── go.mod                   # Go modules
├── go.sum                   # Module checksums
└── README.md                # Project overview
```

---

## Adding New Commands

1. Create a new package under `internal/cli/<command>/`
2. Define the command using Cobra
3. Add domain types to `internal/domain/` if needed
4. Implement the adapter in `internal/adapters/nylas/`
5. Register the command in `cmd/nylas/main.go`
6. Add integration tests in `internal/cli/integration_<command>_test.go`

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for linting
- Write table-driven tests where applicable
- Document exported functions and types
