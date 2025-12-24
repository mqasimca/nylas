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

**For detailed implementation, see `CLAUDE.md` and `docs/DEVELOPMENT.md`**
