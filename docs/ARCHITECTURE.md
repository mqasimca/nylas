# Architecture

The Nylas CLI follows hexagonal (ports and adapters) architecture.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [TUI](TUI.md) | [Development](DEVELOPMENT.md) | [Security](SECURITY.md)

---

## Project Structure

```
cmd/nylas/           # Entry point
internal/
  domain/            # Business entities and errors
    message.go       # Message, Contact types
    email.go         # Thread, Draft, Folder, Attachment types
    calendar.go      # Calendar, Event, Availability types
    contacts.go      # Contact, ContactGroup types
    webhook.go       # Webhook, TriggerTypes
    grant.go         # Grant, Provider types
    config.go        # Configuration types: Config, WorkingHoursConfig, DaySchedule, BreakBlock 📅
    errors.go        # Domain errors
  ports/             # Interfaces (contracts)
    nylas.go         # NylasClient interface
    secrets.go       # SecretStore interface
  adapters/          # External implementations
    keyring/         # Secret storage (system keyring)
    config/          # Configuration files
    nylas/           # Nylas API HTTP client
      client.go      # HTTP client implementation
      mock.go        # Mock client for testing
    oauth/           # OAuth callback server
    browser/         # Browser launcher
  app/               # Application services
    auth/            # Authentication logic
    otp/             # OTP extraction logic
  cli/               # CLI commands
    auth/            # Auth subcommands
    email/           # Email subcommands
    calendar/        # Calendar subcommands
    contacts/        # Contacts subcommands
    webhook/         # Webhook subcommands
    otp/             # OTP subcommands
    common/          # Shared CLI utilities
  tui/               # Terminal User Interface
    app.go           # Main TUI application
    views.go         # Resource views (Messages, Events, etc.)
    calendar.go      # Google Calendar-style component
    compose.go       # Email compose/reply form
    table.go         # k9s-style table component
    styles.go        # k9s color scheme
```

## Design Principles

### Hexagonal Architecture

The codebase separates concerns into three layers:

1. **Domain Layer** (`internal/domain/`)
   - Pure business logic and entities
   - No external dependencies
   - Defines core types: Message, Event, Contact, Webhook, Grant

2. **Ports Layer** (`internal/ports/`)
   - Interfaces that define contracts
   - `NylasClient` - API operations
   - `SecretStore` - Credential storage

3. **Adapters Layer** (`internal/adapters/`)
   - Concrete implementations of ports
   - `keyring/` - System keyring for secrets
   - `nylas/` - HTTP client for Nylas API
   - `oauth/` - OAuth callback server

### Benefits

- **Testability**: Mock adapters for unit testing
- **Flexibility**: Swap implementations without changing business logic
- **Maintainability**: Clear separation of concerns
- **Portability**: Domain logic is independent of infrastructure

### Working Hours and Break Management

The calendar system supports flexible working hours and break time management through the `WorkingHoursConfig` domain model:

**Domain Models:**
- **WorkingHoursConfig** - Per-day working hours configuration with multiple break blocks
- **DaySchedule** - Working hours for a specific weekday (Monday, Tuesday, etc.) with optional break periods
- **BreakBlock** - Break periods (lunch, coffee break, etc.) that enforce hard scheduling constraints

**Configuration:**
Working hours and breaks are stored in `~/.nylas/config.yaml` under the `working_hours` section. The configuration supports:
- Default hours that apply to all weekdays
- Day-specific overrides (e.g., shorter Fridays)
- Weekend configuration
- Multiple break blocks per day

**Validation:**
Event creation respects break blocks as **hard constraints** (cannot be overridden), while working hours serve as **soft warnings** (can be overridden with user confirmation).

**Example Configuration:**
```yaml
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"
      - name: "Afternoon Coffee"
        start: "15:00"
        end: "15:15"
        type: "coffee"
```

**Implementation:**
- Validation logic: `internal/cli/calendar/helpers.go` (`checkBreakViolation()`)
- Event creation: `internal/cli/calendar/events.go` (enforces break blocks)
- Tests: `internal/cli/calendar/helpers_test.go` (`TestCheckBreakViolation`)

For complete usage documentation, see [Timezone & Working Hours Guide](TIMEZONE.md#working-hours--break-management).
