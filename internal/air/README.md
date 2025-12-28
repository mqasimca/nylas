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

| File | Purpose |
|------|---------|
| `air.go` | Entry point, CLI command |
| `server.go` | HTTP server setup, routing |
| `middleware.go` | Auth, CORS, logging middleware |
| `data.go` | Grant store, configuration persistence |

### Handler Organization

Handlers are split by domain:

| File | Handles |
|------|---------|
| `handlers_types.go` | Shared types (EmailResponse, EventResponse, etc.) |
| `handlers_config.go` | `/api/config`, `/api/grants` |
| `handlers_email.go` | `/api/emails`, `/api/send` |
| `handlers_drafts.go` | `/api/drafts` |
| `handlers_calendar.go` | `/api/events`, `/api/calendars` |
| `handlers_contacts.go` | `/api/contacts` |
| `handlers_availability.go` | `/api/availability` |
| `handlers_cache.go` | `/api/cache` |
| `handlers_ai.go` | `/api/ai/*` (summarize, smart-replies, etc.) |
| `handlers_productivity_inbox.go` | `/api/inbox/split`, `/api/snooze`, VIP senders |
| `handlers_productivity_send.go` | `/api/scheduled`, `/api/undo-send`, `/api/templates` |

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

## Cache Subsystem

The `cache/` subdirectory contains the offline cache implementation:
- SQLite-based local storage
- Email, event, contact, and attachment caching
- Search query parsing
- Encryption support
