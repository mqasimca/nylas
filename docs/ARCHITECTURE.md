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

### Quick Lookup

| Looking for | Location |
|-------------|----------|
| CLI helpers (context, config, colors) | `internal/cli/common/` |
| HTTP client | `internal/adapters/nylas/client.go` |
| AI clients (Claude, OpenAI, Groq) | `internal/adapters/ai/` |
| MCP server | `internal/adapters/mcp/` |
| Slack adapter | `internal/adapters/slack/` |
| Air web UI (port 7365) | `internal/air/` |
| UI web interface (port 7363) | `internal/cli/ui/` |
| TUI terminal interface | `internal/tui/` |
| Integration test helpers | `internal/cli/integration/helpers_test.go` |
| Air integration tests | `internal/air/integration_*_test.go` |

---

## Design Principles

### Hexagonal Architecture

**Three layers:**

1. **Domain** (`internal/domain/`) - 21 files
   - Pure business logic, no external dependencies
   - Core types: Message, Email, Calendar, Event, Contact, Grant, Webhook
   - Feature types: AI, Analytics, Admin, Scheduler, Notetaker, Slack, Inbound
   - Support types: Config, Errors, Provider, Utilities

2. **Ports** (`internal/ports/`) - 7 interface files
   - `nylas.go` - NylasClient interface (main API operations)
   - `secrets.go` - SecretStore interface (credential storage)
   - `llm.go` - LLM interface (AI providers)
   - `slack.go` - Slack interface
   - `config.go` - Config interface
   - `utilities.go` - Utilities interface
   - `webhook_server.go` - Webhook server interface

3. **Adapters** (`internal/adapters/`) - 12 adapter directories

   | Adapter | Files | Purpose |
   |---------|-------|---------|
   | `nylas/` | 85 | Nylas API client (messages, calendars, contacts, events) |
   | `ai/` | 18 | AI clients (Claude, OpenAI, Groq, Ollama), email analyzer |
   | `analytics/` | 14 | Focus optimizer, conflict resolver, meeting scorer |
   | `keyring/` | 6 | Credential storage (system keyring, file-based) |
   | `mcp/` | 7 | MCP proxy server for AI assistants |
   | `slack/` | 9 | Slack API client (channels, messages, users) |
   | `config/` | 5 | Configuration validation |
   | `oauth/` | 3 | OAuth callback server |
   | `utilities/` | 12 | Services (contacts, email, scheduling, timezone, webhook) |
   | `browser/` | 2 | Browser automation |
   | `tunnel/` | 2 | Cloudflare tunnel |
   | `webhookserver/` | 2 | Webhook server |

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

**Details:** See [commands/timezone.md](commands/timezone.md#working-hours--break-management)

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
| `integration_base_test.go` | 0 | Shared `testServer()` helper, utilities, rate limiting |
| `integration_core_test.go` | 5 | Config, Grants, Folders, Index page |
| `integration_email_test.go` | 4 | Email listing, filtering, drafts |
| `integration_calendar_test.go` | 11 | Calendars, events, availability, conflicts |
| `integration_contacts_test.go` | 4 | Contact CRUD operations |
| `integration_cache_test.go` | 4 | Cache store operations, invalidation |
| `integration_ai_test.go` | 15 | AI summarization, smart compose, thread analysis, config |
| `integration_middleware_test.go` | 6 | Compression, security headers, CORS |
| `integration_bundles_test.go` | 8 | Email bundles, categorization, bundle operations |
| `integration_productivity_test.go` | 8 | Scheduled send, undo send, snooze, reply later |

**Total:** 65 integration tests across 10 organized files

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
