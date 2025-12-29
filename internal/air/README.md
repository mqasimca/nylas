# Air - Web UI Server

This package implements a local web server providing a modern web interface for Nylas email and calendar.

## Usage

```bash
nylas air              # Start web UI on http://localhost:7365
nylas air --port 3000  # Custom port
nylas air --no-browser # Start without opening browser
nylas air --encrypted  # Enable encryption for cached data
nylas air --clear-cache # Clear all cached data before starting
```

## Architecture

Air is a self-contained HTTP server with embedded static assets (HTML, CSS, JS).

### Core Files

**All files are ≤500 lines for maintainability.**

**Server Core** (refactored from server.go):
- `server.go` (51 lines) - Server struct definition
- `server_lifecycle.go` (315 lines) - HTTP server setup, routing, lifecycle
- `server_stores.go` (67 lines) - Cache store accessors
- `server_sync.go` (187 lines) - Background sync logic
- `server_offline.go` (98 lines) - Offline queue processing
- `server_converters.go` (116 lines) - Domain to cache conversions
- `server_template.go` (163 lines) - Template handling
- `server_modules_test.go` (523 lines) - Unit tests for server modules

**Other Core:**
- `air.go` - Entry point, CLI command
- `middleware.go` - Auth, CORS, logging middleware
- `data.go` - Grant store, configuration persistence

### Handler Organization

**Email Handlers:**
- `handlers_email.go` (522 lines) - `/api/emails`, `/api/send`
- `handlers_drafts.go` (531 lines) - `/api/drafts`
- `handlers_bundles.go` - Email categorization into smart bundles

**Calendar Handlers** (refactored from handlers_calendar.go):
- `handlers_calendars.go` (108 lines) - `/api/calendars`
- `handlers_events.go` (520 lines) - `/api/events`
- `handlers_calendar_helpers.go` (159 lines) - Helpers and converters

**Contact Handlers** (refactored from handlers_contacts.go):
- `handlers_contacts.go` (190 lines) - `/api/contacts` listing
- `handlers_contacts_crud.go` (320 lines) - CRUD operations
- `handlers_contacts_search.go` (194 lines) - Search and groups
- `handlers_contacts_helpers.go` (308 lines) - Helpers and demo data

**AI Handlers** (refactored from handlers_ai.go):
- `handlers_ai_types.go` (92 lines) - Type definitions
- `handlers_ai_summarize.go` (165 lines) - `/api/ai/summarize`
- `handlers_ai_smart.go` (225 lines) - `/api/ai/smart-replies`
- `handlers_ai_thread.go` (192 lines) - `/api/ai/thread-summary`
- `handlers_ai_complete.go` (216 lines) - `/api/ai/complete`
- `handlers_ai_config.go` (227 lines) - `/api/ai/config`

**Productivity - Send Features** (refactored from handlers_productivity_send.go):
- `types_productivity_send.go` (82 lines) - Type definitions
- `handlers_scheduled_send.go` (177 lines) - `/api/scheduled`
- `handlers_undo_send.go` (148 lines) - `/api/undo-send`
- `handlers_templates.go` (259 lines) - `/api/templates`
- `handlers_templates_helpers.go` (100 lines) - Template utilities

**Productivity - Inbox Features** (refactored from handlers_productivity_inbox.go):
- `handlers_splitinbox_types.go` (84 lines) - Type definitions
- `handlers_splitinbox_config.go` (164 lines) - `/api/inbox/split`
- `handlers_splitinbox_categorize.go` (148 lines) - Categorization logic
- `handlers_snooze_types.go` (37 lines) - Type definitions
- `handlers_snooze_handlers.go` (126 lines) - `/api/snooze`
- `handlers_snooze_parser.go` (142 lines) - Natural language parser

**Other Handlers:**
- `handlers_types.go` - Shared types (EmailResponse, EventResponse, etc.)
- `handlers_config.go` - `/api/config`, `/api/grants`
- `handlers_availability.go` (527 lines) - `/api/availability`
- `handlers_cache.go` - `/api/cache`
- `handlers_productivity_*.go` - Focus mode, reply later, analytics, etc.

### Static Assets

| Directory | Contents |
|-----------|----------|
| `static/` | HTML templates |
| `static/css/` | Stylesheets |
| `static/js/` | JavaScript modules |

## API Endpoints

### Email
- `GET /api/emails` - List emails
- `GET /api/emails/:id` - Get single email
- `POST /api/send` - Send email
- `PUT /api/emails/:id` - Update (mark read, star, etc.)
- `DELETE /api/emails/:id` - Trash email

### Drafts
- `GET /api/drafts` - List drafts
- `POST /api/drafts` - Create draft
- `PUT /api/drafts/:id` - Update draft
- `DELETE /api/drafts/:id` - Delete draft

### Calendar
- `GET /api/events` - List events
- `POST /api/events` - Create event
- `GET /api/calendars` - List calendars
- `GET /api/availability` - Check availability

### Productivity Features
- `GET/PUT /api/inbox/split` - Split inbox configuration
- `POST /api/inbox/categorize` - Categorize email
- `GET/POST/DELETE /api/inbox/vip` - VIP senders
- `GET/POST/DELETE /api/snooze` - Snooze emails
- `GET/POST/DELETE /api/scheduled` - Scheduled send
- `GET/PUT/POST /api/undo-send` - Undo send config
- `GET/POST /api/templates` - Email templates

### AI Features
- `POST /api/ai/summarize` - Email summary
- `POST /api/ai/smart-replies` - Generate reply suggestions
- `POST /api/ai/enhanced-summary` - Detailed analysis
- `POST /api/ai/auto-label` - Auto-categorization
- `POST /api/ai/thread-summary` - Thread overview

## Test Organization

| File | Tests |
|------|-------|
| `*_test.go` | Unit tests (handler logic) |
| `integration_*.go` | Integration tests (require build tag) |

Integration test files:
- `integration_base_test.go` - Shared helpers (`testServer()`)
- `integration_core_test.go` - Config, grants, folders
- `integration_email_test.go` - Email operations
- `integration_calendar_test.go` - Calendar operations
- `integration_contacts_test.go` - Contact operations
- `integration_cache_test.go` - Cache operations
- `integration_ai_test.go` - AI features
- `integration_middleware_test.go` - Middleware tests
- `integration_productivity_test.go` - Productivity features (focus, reply later, notetaker, etc.)
- `integration_bundles_test.go` - Email bundle categorization

## Cache Subsystem

The `cache/` subdirectory contains the offline cache implementation:
- SQLite-based local storage
- Email, event, contact, and attachment caching
- Search query parsing
- Encryption support
