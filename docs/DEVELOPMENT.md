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

Integration tests are organized in `internal/cli/integration/` directory:

| File | Description |
|------|-------------|
| `test.go` | Common setup, helpers |
| `auth_test.go` | Auth command tests |
| `email_test.go` | Email list, read, search, mark, send tests |
| `folders_test.go` | Folder command tests |
| `threads_test.go` | Thread command tests |
| `drafts_test.go` | Draft command tests |
| `calendar_test.go` | Calendar & availability tests |
| `contacts_test.go` | Contact command tests |
| `webhooks_test.go` | Webhook command tests |
| `misc_test.go` | OTP, help, errors, workflow, doctor tests |

```bash
# Run integration tests (requires NYLAS_API_KEY and NYLAS_GRANT_ID)
go test -tags=integration ./internal/cli/integration/...

# Optional: specify custom binary path
NYLAS_TEST_BINARY=/path/to/nylas go test -tags=integration ./internal/cli/integration/...

# Run specific test file
go test -tags=integration ./internal/cli/integration/ -run TestCLI_Email
```

**See**: `internal/cli/integration/README.md` for detailed testing documentation.

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
| Scheduling Validation | Break block enforcement, working hours validation, conflict detection |
| Contacts | List, get, create, delete |
| Contact Groups | List groups |
| Webhooks | CRUD, triggers, test events |
| Attachments | Get metadata, download content |
| Concurrency | Parallel requests |
| Error Handling | Invalid IDs, timeouts |

### Break Validation Tests

Calendar event creation includes comprehensive break time validation tests to ensure that configured break blocks are properly enforced.

**Test File:** `internal/cli/calendar/helpers_test.go`

**Run Break Validation Tests:**
```bash
# Run all calendar tests (includes break validation)
go test ./internal/cli/calendar/... -v

# Run only break validation tests
go test ./internal/cli/calendar/... -v -run TestCheckBreakViolation

# Run with coverage
go test ./internal/cli/calendar/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Test Coverage:**

The `TestCheckBreakViolation` test suite covers:

| Test Case | Purpose |
|-----------|---------|
| No config | Validates that missing config doesn't cause errors |
| No breaks | Events can be scheduled when no breaks are configured |
| Outside breaks | Events before/after breaks are allowed |
| During breaks | Events during breaks are rejected (hard block) |
| At break start | Events exactly at break start time are rejected |
| At break end | Events exactly at break end time are allowed |
| Multiple breaks | Validates all break blocks in a day |
| Day-specific breaks | Different breaks for different weekdays (e.g., Monday vs Friday) |
| Invalid config | Gracefully handles malformed break configurations |

**Example Test:**
```go
func TestCheckBreakViolation(t *testing.T) {
    eventTime := time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC) // 12:30 PM
    config := &domain.Config{
        WorkingHours: &domain.WorkingHoursConfig{
            Default: &domain.DaySchedule{
                Enabled: true,
                Start:   "09:00",
                End:     "17:00",
                Breaks: []domain.BreakBlock{
                    {Name: "Lunch", Start: "12:00", End: "13:00", Type: "lunch"},
                },
            },
        },
    }

    violation := checkBreakViolation(eventTime, config)
    // Should return: "Event cannot be scheduled during Lunch (12:00 - 13:00)"
}
```

**Integration with Event Creation:**

The `checkBreakViolation()` function is called during event creation in `internal/cli/calendar/events.go`:

1. **Before** working hours validation (breaks are more restrictive)
2. **Hard block** - returns error immediately if event conflicts with break
3. **User-friendly error** - shows break name and time range

**Testing Best Practices:**

- ✅ Test boundary conditions (start/end of breaks)
- ✅ Test multiple breaks in a single day
- ✅ Test day-specific overrides (Monday lunch vs Friday lunch)
- ✅ Test invalid configurations don't crash
- ✅ Verify exact error messages for user-facing output

For complete working hours and break configuration, see [Timezone & Working Hours Guide](TIMEZONE.md#working-hours--break-management).

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
6. Add integration tests in `internal/cli/integration/<command>_test.go`

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for linting
- Write table-driven tests where applicable
- Document exported functions and types
