# Architecture

Hexagonal (ports and adapters) architecture for clean separation of concerns.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Development](DEVELOPMENT.md)

---

## Project Structure

```
cmd/nylas/           # Entry point
internal/
  domain/            # Business entities (Message, Calendar, Contact, etc.)
  ports/             # Interface contracts (NylasClient, SecretStore)
  adapters/          # Implementations (HTTP client, keyring, OAuth)
    nylas/           # Nylas API client
    keyring/         # Secret storage
    config/          # Configuration
  cli/<feature>/     # CLI commands by feature
  tui/               # Terminal UI
  air/               # Web UI server
```

---

## Design Principles

### Hexagonal Architecture

**Three layers:**

1. **Domain** (`internal/domain/`)
   - Pure business logic
   - No external dependencies
   - Core types: Message, Event, Contact, Calendar, Webhook

2. **Ports** (`internal/ports/`)
   - Interface contracts
   - `NylasClient` - API operations
   - `SecretStore` - Credential storage

3. **Adapters** (`internal/adapters/`)
   - Concrete implementations
   - `nylas/` - HTTP client for Nylas API
   - `keyring/` - System keyring
   - `oauth/` - OAuth callback server

**Benefits:**
- Testability (mock adapters)
- Flexibility (swap implementations)
- Clean separation of concerns

---

## Working Hours and Breaks

Calendar enforces working hours (soft warnings) and break blocks (hard constraints).

**Domain models:**
- `WorkingHoursConfig` - Per-day working hours with break periods
- `DaySchedule` - Working hours for specific weekday
- `BreakBlock` - Break periods (lunch, coffee) with hard constraints

**Configuration:** `~/.nylas/config.yaml`
**Implementation:** `internal/cli/calendar/helpers.go` (`checkBreakViolation()`)
**Tests:** `internal/cli/calendar/helpers_test.go`

**Details:** See [TIMEZONE.md](TIMEZONE.md#working-hours--break-management)

---

## CLI Pattern

Each feature follows consistent structure:

```
internal/cli/<feature>/
  ├── <feature>.go    # Main command
  ├── list.go         # List subcommand
  ├── create.go       # Create subcommand
  ├── update.go       # Update subcommand
  ├── delete.go       # Delete subcommand
  └── helpers.go      # Shared helpers
```

---

## Air (Web UI)

**Air** is the web-based UI for Nylas CLI, providing a browser interface for email, calendar, and productivity features.

### Architecture

- **Location:** `internal/air/`
- **Server:** HTTP server with middleware stack (CORS, compression, security, caching)
- **Handlers:** Feature-specific HTTP handlers (email, calendar, contacts, AI)
- **Templates:** Go templates with Tailwind CSS
- **Port:** Default `:7365` (configurable)

### File Organization

```
internal/air/
  ├── server.go              # HTTP server setup
  ├── handlers_*.go          # Feature handlers (email, calendar, contacts, AI)
  ├── middleware.go          # Middleware stack
  ├── data.go               # Data models
  ├── templates/             # HTML templates
  └── integration_*.go       # Integration tests (organized by feature)
```

### Integration Tests

Air integration tests are **split by feature** for better maintainability:

| File | Tests | Purpose |
|------|-------|---------|
| `integration_base_test.go` | 0 | Shared `testServer()` helper and utilities |
| `integration_core_test.go` | 5 | Config, Grants, Folders, Index page |
| `integration_email_test.go` | 4 | Email listing, filtering, drafts |
| `integration_calendar_test.go` | 11 | Calendars, events, availability, conflicts |
| `integration_contacts_test.go` | 4 | Contact operations |
| `integration_cache_test.go` | 4 | Cache operations |
| `integration_ai_test.go` | 3 | AI summarization features |
| `integration_middleware_test.go` | 6 | Middleware (compression, security, CORS) |

**Total:** 37 integration tests across 8 organized files

**Running tests:**
```bash
make ci-full                     # RECOMMENDED: Complete CI with automatic cleanup
make test-air-integration        # Run Air integration tests only
make test-cleanup                # Manual cleanup if needed
```

**Why cleanup?** Air tests create real resources (drafts, events, contacts) in the connected Nylas account. The `make ci-full` target automatically runs cleanup after all tests.

**Pattern:** Air tests use `httptest` to test HTTP handlers directly:
```go
func TestIntegration_Feature(t *testing.T) {
    server := testServer(t)  // Shared helper
    req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
    w := httptest.NewRecorder()
    server.handleEndpoint(w, req)
    // Assertions...
}
```

---

**For detailed implementation, see `CLAUDE.md` and `docs/DEVELOPMENT.md`**
