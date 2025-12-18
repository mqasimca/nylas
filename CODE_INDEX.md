# Code Index

Quick reference for finding code by functionality. Use for fast searches.

## By Feature

| Feature | Domain | Adapter | CLI | Tests |
|---------|--------|---------|-----|-------|
| **Email/Messages** | `domain/email.go`, `domain/message.go` | `adapters/nylas/messages.go` | `cli/email/` | `cli/email/email_test.go` |
| **Threads** | `domain/email.go` | `adapters/nylas/threads.go` | `cli/email/threads.go` | `cli/email/email_test.go` |
| **Drafts** | `domain/email.go` | `adapters/nylas/drafts.go` | `cli/email/drafts.go` | `cli/email/email_test.go` |
| **Folders** | `domain/email.go` | `adapters/nylas/folders.go` | `cli/email/folders.go` | `cli/email/email_test.go` |
| **Attachments** | `domain/attachment.go` | `adapters/nylas/attachments.go` | `cli/email/attachments.go` | `cli/email/email_test.go` |
| **Calendar** | `domain/calendar.go` | `adapters/nylas/calendars.go` | `cli/calendar/` | `cli/calendar/calendar_test.go` |
| **Events** | `domain/calendar.go` | `adapters/nylas/events.go` | `cli/calendar/events.go` | `cli/calendar/calendar_test.go` |
| **Contacts** | `domain/contacts.go` | `adapters/nylas/contacts.go` | `cli/contacts/` | `cli/contacts/contacts_test.go` |
| **Contact Groups** | `domain/contacts.go` | `adapters/nylas/contact_groups.go` | `cli/contacts/groups.go` | `cli/contacts/contacts_test.go` |
| **Webhooks** | `domain/webhook.go` | `adapters/nylas/webhooks.go` | `cli/webhook/` | `cli/webhook/webhook_test.go` |
| **Notetaker** | `domain/notetaker.go` | `adapters/nylas/notetakers.go` | `cli/notetaker/` | `cli/notetaker/notetaker_test.go` |
| **Auth/Grants** | `domain/grant.go` | `adapters/nylas/auth.go`, `grants.go` | `cli/auth/` | `cli/auth/auth_test.go` |
| **OTP** | - | - | `cli/otp/` | `cli/otp/otp_test.go` |

## By File Type

### Domain Models (`internal/domain/`)
```
email.go      - Message, Thread, Draft, Folder, SendMessageRequest, TrackingOptions
message.go    - (legacy, see email.go)
calendar.go   - Calendar, Event, CreateEventRequest, Availability
contacts.go   - Contact, ContactGroup, CreateContactRequest
webhook.go    - Webhook, TriggerType, CreateWebhookRequest
notetaker.go  - Notetaker, MediaData, CreateNotetakerRequest
grant.go      - Grant, Provider, GrantStatus
attachment.go - Attachment
config.go     - Config types
errors.go     - Domain errors (ErrNotFound, ErrUnauthorized, etc.)
```

### Port Interfaces (`internal/ports/`)
```
nylas.go      - NylasClient interface (ALL API methods)
secrets.go    - SecretStore interface
```

### Adapters (`internal/adapters/nylas/`)
```
client.go         - HTTPClient, SetCredentials, get/post/put/delete helpers
messages.go       - GetMessages, SendMessage, UpdateMessage
threads.go        - GetThreads, GetThread, UpdateThread
drafts.go         - GetDrafts, CreateDraft, SendDraft
folders.go        - GetFolders, CreateFolder, UpdateFolder
attachments.go    - GetAttachment, DownloadAttachment
calendars.go      - GetCalendars, CreateCalendar, CreateEvent
events.go         - GetEvents, UpdateEvent, SendRSVP
contacts.go       - GetContacts, CreateContact, UpdateContact
contact_groups.go - GetContactGroups, CreateContactGroup
webhooks.go       - GetWebhooks, CreateWebhook, UpdateWebhook
notetakers.go     - ListNotetakers, CreateNotetaker, GetNotetakerMedia
grants.go         - ListGrants, GetGrant, DeleteGrant
auth.go           - OAuth flow, token exchange
scheduled.go      - Scheduled messages
mock.go           - MockClient (for unit tests)
demo.go           - DemoClient (for TUI demo mode)
integration_test.go - Integration tests
```

### CLI Commands (`internal/cli/`)
```
email/        - email list, read, send, search, mark, delete, threads, drafts, folders
calendar/     - calendar list, events, availability
contacts/     - contacts list, show, create, update, delete, groups
webhook/      - webhook list, show, create, update, delete, triggers, server
notetaker/    - notetaker list, show, create, delete, media
auth/         - auth login, logout, status, config, list, show, switch
otp/          - otp get, watch, list
tui.go        - TUI launcher command
doctor.go     - Diagnostic command
root.go       - Root command setup
```

## Common Search Patterns

```bash
# Find where a CLI flag is defined
grep -r "Flags().*\"flag-name\"" internal/cli/

# Find API endpoint implementation
grep -r "grants/%s/endpoint" internal/adapters/nylas/

# Find domain type definition
grep -r "type TypeName struct" internal/domain/

# Find port interface method
grep -r "MethodName.*context.Context" internal/ports/

# Find command registration
grep -r "AddCommand" cmd/nylas/main.go

# Find error handling
grep -r "ErrNotFound\|ErrUnauthorized" internal/
```

## Quick Lookups

### "Where is X defined?"
| Looking for... | Location |
|---------------|----------|
| API base URL | `adapters/nylas/client.go:17` |
| HTTP client | `adapters/nylas/client.go` |
| All API methods | `ports/nylas.go` |
| Domain errors | `domain/errors.go` |
| CLI root command | `cli/root.go` |
| Main entry point | `cmd/nylas/main.go` |
| TUI application | `tui/app.go` |
| Secret storage | `adapters/keyring/` |
| OAuth callback | `adapters/oauth/` |

### "How do I add X?"
| Task | See Example |
|------|-------------|
| New CLI command | `cli/notetaker/notetaker.go` |
| New API method | `adapters/nylas/notetakers.go` |
| New domain type | `domain/notetaker.go` |
| New flag | `cli/email/send.go` (tracking flags) |
| New subcommand | `cli/email/threads.go` |
