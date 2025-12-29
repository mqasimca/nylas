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

**All files are ≤500 lines for maintainability.** Large files have been refactored into focused modules:

**Server Core** (refactored from server.go):
- `server.go` - Server struct definition
- `server_lifecycle.go` - Initialization, routing, lifecycle
- `server_stores.go` - Cache store accessors
- `server_sync.go` - Background sync logic
- `server_offline.go` - Offline queue processing
- `server_converters.go` - Domain to cache conversions
- `server_template.go` - Template handling
- `server_modules_test.go` - Unit tests

**Handlers** (organized by feature):
- Email: `handlers_email.go`, `handlers_drafts.go`, `handlers_bundles.go`
- Calendar: `handlers_calendars.go`, `handlers_events.go`, `handlers_calendar_helpers.go`
- Contacts: `handlers_contacts.go`, `handlers_contacts_crud.go`, `handlers_contacts_search.go`, `handlers_contacts_helpers.go`
- AI: `handlers_ai_types.go`, `handlers_ai_summarize.go`, `handlers_ai_smart.go`, `handlers_ai_thread.go`, `handlers_ai_complete.go`, `handlers_ai_config.go`
- Productivity: `handlers_scheduled_send.go`, `handlers_undo_send.go`, `handlers_templates.go`, `handlers_snooze_*.go`, `handlers_splitinbox_*.go`

**Other:**
- `middleware.go` - Middleware stack
- `data.go` - Data models
- `templates/` - HTML templates
- `integration_*.go` - Integration tests (organized by feature)

**Complete file listing:** See `CLAUDE.md` for detailed file structure with line counts

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
