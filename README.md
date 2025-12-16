# Nylas CLI

A unified command-line tool for Nylas API authentication, email management, calendar, contacts, webhooks, and OTP extraction.

## Features

- **Email Management**: List, read, send, search, and organize emails with scheduled sending support
- **Calendar Management**: View calendars, list/create/delete events, check availability
- **Contacts Management**: List, view, create, and delete contacts and contact groups
- **Webhook Management**: Create, update, delete, and test webhooks for event notifications
- **Draft Management**: Create, edit, and send drafts
- **Folder Management**: Create, rename, and delete folders/labels
- **Thread Management**: View and manage email conversations
- **OTP Extraction**: Automatically extract one-time passwords from emails
- **Multi-Account Support**: Manage multiple email accounts with grant switching
- **Secure Credential Storage**: Uses system keyring for credentials

## Installation

```bash
go install github.com/mqasimca/nylas/cmd/nylas@latest
```

Or build from source:

```bash
make build
```

## Quick Start

```bash
# Configure with your Nylas credentials
nylas auth config

# Login with your email provider
nylas auth login

# List recent emails
nylas email list

# Send an email (immediately)
nylas email send --to "recipient@example.com" --subject "Hello" --body "Hi there!"

# Send an email (scheduled for 2 hours from now)
nylas email send --to "recipient@example.com" --subject "Reminder" --schedule 2h

# List upcoming calendar events
nylas calendar events list

# Check calendar availability
nylas calendar availability check

# List contacts
nylas contacts list

# List webhooks
nylas webhook list

# Get the latest OTP code
nylas otp get
```

---

## Commands

### Authentication

Manage Nylas API authentication and multiple accounts.

```bash
nylas auth config     # Configure API credentials
nylas auth login      # Authenticate with email provider
nylas auth logout     # Revoke current authentication
nylas auth status     # Show authentication status
nylas auth whoami     # Show current user info
nylas auth list       # List all accounts
nylas auth switch     # Switch between accounts
nylas auth add        # Manually add an existing grant
nylas auth token      # Show/copy API key
nylas auth revoke     # Revoke specific grant
```

**Example: Check authentication status**
```bash
$ nylas auth status

Authentication Status
─────────────────────────────────────────────────────
  Status:    Authenticated
  Provider:  google
  Email:     user@gmail.com
  Grant ID:  abc123def456
  Region:    us
```

**Example: List all accounts**
```bash
$ nylas auth list

Authenticated Accounts
─────────────────────────────────────────────────────
  1. user@gmail.com (google) [default]
     Grant ID: abc123def456

  2. work@company.com (microsoft)
     Grant ID: xyz789ghi012
```

---

### Email Operations

Full email management including reading, sending, searching, and organizing.

#### List Emails

```bash
nylas email list [grant-id]           # List recent emails
nylas email list --limit 20           # Specify number of emails
nylas email list --unread             # Show only unread
nylas email list --starred            # Show only starred
nylas email list --from "sender@example.com"  # Filter by sender
```

**Example output:**
```bash
$ nylas email list --limit 5

Recent Emails
─────────────────────────────────────────────────────
  From: John Doe <john@example.com>
  Subject: Meeting Tomorrow
  Date: 2 hours ago
  ID: msg_abc123

  From: GitHub <noreply@github.com>
  Subject: [repo] New pull request #42
  Date: 5 hours ago
  ID: msg_def456

  From: Newsletter <news@company.com>
  Subject: Weekly Update
  Date: yesterday
  ID: msg_ghi789

Found 5 emails
```

#### Read Email

```bash
nylas email read <message-id>         # Read a specific email
nylas email read <id> --mark-read     # Mark as read after reading
```

**Example output:**
```bash
$ nylas email read msg_abc123

From: John Doe <john@example.com>
To: you@example.com
Subject: Meeting Tomorrow
Date: Mon, Dec 16, 2024 2:30 PM

Hi,

Just a reminder about our meeting tomorrow at 10am.
Please bring the quarterly report.

Best,
John

─────────────────────────────────────────────────────
ID: msg_abc123
Thread: thread_xyz789
```

#### Send Email

```bash
# Send immediately
nylas email send --to "to@example.com" --subject "Subject" --body "Body"

# Send with CC and BCC
nylas email send --to "to@example.com" --cc "cc@example.com" --bcc "bcc@example.com" \
  --subject "Subject" --body "Body"

# Schedule to send in 2 hours
nylas email send --to "to@example.com" --subject "Reminder" --body "..." --schedule 2h

# Schedule for tomorrow at 9am
nylas email send --to "to@example.com" --subject "Morning" --schedule "tomorrow 9am"

# Schedule for a specific date/time
nylas email send --to "to@example.com" --subject "Meeting" --schedule "2024-12-20 14:30"

# Skip confirmation prompt
nylas email send --to "to@example.com" --subject "Quick" --body "..." --yes
```

**Example output (scheduled):**
```bash
$ nylas email send --to "user@example.com" --subject "Reminder" --body "Don't forget!" --schedule 2h --yes

Email preview:
  To:      user@example.com
  Subject: Reminder
  Body:    Don't forget!
  Scheduled: Mon Dec 16, 2024 4:30 PM PST

✓ Email scheduled successfully! Message ID: msg_scheduled_123
Scheduled to send: Mon Dec 16, 2024 4:30 PM PST
```

#### Search Emails

```bash
nylas email search "query"            # Search emails
nylas email search "query" --limit 50 # Search with custom limit
nylas email search "query" --from "sender@example.com"
nylas email search "query" --after "2024-01-01"
nylas email search "query" --before "2024-12-31"
```

**Example output:**
```bash
$ nylas email search "invoice" --limit 3

Search Results for "invoice"
─────────────────────────────────────────────────────
  From: Billing <billing@service.com>
  Subject: Your December Invoice
  Date: 3 days ago
  ID: msg_inv001

  From: Accounting <accounting@company.com>
  Subject: Invoice #2024-156 Approved
  Date: 1 week ago
  ID: msg_inv002

Found 3 matching emails
```

#### Mark Operations

```bash
nylas email mark read <message-id>      # Mark as read
nylas email mark unread <message-id>    # Mark as unread
nylas email mark starred <message-id>   # Star a message
nylas email mark unstarred <message-id> # Unstar a message
```

#### Delete Email

```bash
nylas email delete <message-id>       # Delete an email
nylas email delete <message-id> -f    # Delete without confirmation
```

---

### Folder Management

Manage email folders and labels.

```bash
nylas email folders list              # List all folders
nylas email folders create "Folder Name"  # Create a folder
nylas email folders delete <folder-id>    # Delete a folder
```

**Example output:**
```bash
$ nylas email folders list

Folders
─────────────────────────────────────────────────────
NAME              ID                    MESSAGES
INBOX             folder_inbox          1,234
Sent              folder_sent           567
Drafts            folder_drafts         12
Trash             folder_trash          45
Work              folder_custom_001     89
Personal          folder_custom_002     156

Found 6 folders
```

---

### Thread Management

View and manage email conversations.

```bash
nylas email threads list              # List threads
nylas email threads list --unread     # List unread threads
nylas email threads list --limit 20   # Limit results
nylas email threads read <thread-id>  # Read a thread
```

**Example output:**
```bash
$ nylas email threads list --limit 3

Email Threads
─────────────────────────────────────────────────────
  Subject: Project Discussion
  Participants: John, Jane, You
  Messages: 5
  Last: 1 hour ago
  ID: thread_abc123

  Subject: Re: Meeting Notes
  Participants: Team
  Messages: 12
  Last: 3 hours ago
  ID: thread_def456

Found 3 threads
```

---

### Draft Management

Create, view, and manage email drafts.

```bash
nylas email drafts list               # List drafts
nylas email drafts create --to "to@example.com" --subject "Subject" --body "Body"
nylas email drafts show <draft-id>    # Show a draft
nylas email drafts send <draft-id>    # Send a draft
nylas email drafts delete <draft-id>  # Delete a draft
```

**Example output:**
```bash
$ nylas email drafts list

Drafts
─────────────────────────────────────────────────────
  To: client@company.com
  Subject: Proposal Draft
  Last Modified: 2 hours ago
  ID: draft_abc123

  To: team@company.com
  Subject: Weekly Update
  Last Modified: yesterday
  ID: draft_def456

Found 2 drafts
```

---

### Calendar Management

View calendars, manage events, and check availability.

#### List Calendars

```bash
nylas calendar list [grant-id]        # List all calendars
nylas cal list                        # Alias
```

**Example output:**
```bash
$ nylas calendar list

Found 3 calendar(s):

NAME                    ID                      PRIMARY   READ-ONLY
Personal                cal_primary_123         Yes
Work                    cal_work_456
Holidays                cal_holidays_789                  Yes
```

#### Calendar Events

```bash
# List events
nylas calendar events list [grant-id]
nylas calendar events list --days 14        # Next 14 days
nylas calendar events list --limit 20       # Limit results
nylas calendar events list --calendar <id>  # Specific calendar
nylas calendar events list --show-cancelled # Include cancelled

# Show event details
nylas calendar events show <event-id>

# Create event
nylas calendar events create --title "Meeting" --start "2024-12-20 14:00" --end "2024-12-20 15:00"
nylas calendar events create --title "Vacation" --start "2024-12-25" --all-day
nylas calendar events create --title "Team Sync" --start "2024-12-20 10:00" \
  --participant "alice@example.com" --participant "bob@example.com"

# Delete event
nylas calendar events delete <event-id>
nylas calendar events delete <event-id> --force
```

**Example output (list events):**
```bash
$ nylas calendar events list --days 7

Found 4 event(s):

Team Standup
  When: Mon, Dec 16, 2024, 9:00 AM - 9:30 AM
  Location: Conference Room A
  Status: confirmed
  Guests: 5 participant(s)
  ID: event_abc123

Project Review
  When: Tue, Dec 17, 2024, 2:00 PM - 3:00 PM
  Status: confirmed
  Guests: 3 participant(s)
  ID: event_def456

Holiday Party
  When: Fri, Dec 20, 2024 (all day)
  Location: Main Office
  Status: confirmed
  ID: event_ghi789
```

**Example output (show event):**
```bash
$ nylas calendar events show event_abc123

Team Standup

When
  Mon, Dec 16, 2024, 9:00 AM - 9:30 AM

Location
  Conference Room A

Description
  Daily team standup meeting to discuss progress and blockers.

Organizer
  John Smith <john@company.com>

Participants
  Alice Johnson <alice@company.com> ✓ accepted
  Bob Wilson <bob@company.com> ✓ accepted
  Carol Davis <carol@company.com> ? tentative

Video Conference
  Provider: zoom
  URL: https://zoom.us/j/123456789

Details
  Status: confirmed
  Busy: true
  ID: event_abc123
  Calendar: cal_primary_123
```

#### Calendar Availability

```bash
# Check free/busy status
nylas calendar availability check [grant-id]
nylas calendar availability check --emails alice@example.com,bob@example.com
nylas calendar availability check --start "tomorrow 9am" --end "tomorrow 5pm"
nylas calendar availability check --duration 7d
nylas calendar availability check --format json

# Find available meeting times
nylas calendar availability find --participants alice@example.com,bob@example.com
nylas calendar availability find --participants team@example.com --duration 60
nylas calendar availability find --participants alice@example.com \
  --start "tomorrow 9am" --end "tomorrow 5pm" --interval 15
```

**Example output (check):**
```bash
$ nylas calendar availability check --emails alice@example.com,bob@example.com

Free/Busy Status: Mon Dec 16 2:30 PM - Tue Dec 17 2:30 PM
────────────────────────────────────────────────────────────

📧 alice@example.com
   Busy times:
   ● Mon Dec 16 3:00 PM - 4:00 PM
   ● Tue Dec 17 9:00 AM - 10:00 AM

📧 bob@example.com
   ✓ Free during this period
```

**Example output (find):**
```bash
$ nylas calendar availability find --participants alice@example.com,bob@example.com --duration 30

Available 30-minute Meeting Slots
────────────────────────────────────────

📅 Mon, Dec 16
   1. 9:00 AM - 9:30 AM
   2. 9:30 AM - 10:00 AM
   3. 11:00 AM - 11:30 AM
   4. 2:00 PM - 2:30 PM

📅 Tue, Dec 17
   5. 10:30 AM - 11:00 AM
   6. 1:00 PM - 1:30 PM
   7. 3:00 PM - 3:30 PM

Found 7 available slots
```

---

### Contacts Management

Manage contacts and contact groups.

#### List Contacts

```bash
nylas contacts list [grant-id]
nylas contacts list --limit 100
nylas contacts list --email "john@example.com"
nylas contacts list --source address_book
```

**Example output:**
```bash
$ nylas contacts list --limit 5

Found 5 contact(s):

NAME                EMAIL                      PHONE            COMPANY
Alice Johnson       alice@company.com          +1-555-0101      Acme Corp - Engineer
Bob Wilson          bob@example.com            +1-555-0102
Carol Davis         carol@startup.io           +1-555-0103      Startup Inc - CEO
David Brown         david@email.com                             Freelancer
Eve Martinez        eve@company.com            +1-555-0105      Acme Corp - Designer
```

#### Show Contact

```bash
nylas contacts show <contact-id> [grant-id]
nylas contacts get <contact-id>  # Alias
```

**Example output:**
```bash
$ nylas contacts show contact_abc123

Alice Johnson

Work
  Job Title: Software Engineer
  Company: Acme Corporation
  Manager: John Smith

Email Addresses
  alice@company.com (work)
  alice.personal@gmail.com (personal)

Phone Numbers
  +1-555-0101 (mobile)
  +1-555-0102 (work)

Addresses
  (work)
    123 Main Street
    San Francisco, CA 94102
    United States

Web Pages
  https://linkedin.com/in/alice (linkedin)
  https://github.com/alice (profile)

Personal
  Nickname: Ali
  Birthday: 1990-05-15

Notes
  Met at the tech conference in 2023.

Details
  ID: contact_abc123
  Source: address_book
```

#### Create Contact

```bash
nylas contacts create [grant-id]
nylas contacts create --first-name "John" --last-name "Doe" --email "john@example.com"
nylas contacts create --first-name "Jane" --last-name "Smith" \
  --email "jane@company.com" --phone "+1-555-123-4567" \
  --company "Acme Corp" --job-title "Engineer"
```

**Example output:**
```bash
$ nylas contacts create --first-name "John" --last-name "Doe" --email "john@example.com"

✓ Contact created successfully!

Name: John Doe
Email: john@example.com
ID: contact_new_123
```

#### Delete Contact

```bash
nylas contacts delete <contact-id> [grant-id]
nylas contacts delete <contact-id> --force   # Skip confirmation
```

#### Contact Groups

```bash
nylas contacts groups [grant-id]
```

**Example output:**
```bash
$ nylas contacts groups

Found 4 contact group(s):

NAME                ID                    PATH
Family              group_abc123          /Family
Work                group_def456          /Work
Friends             group_ghi789          /Friends
VIP                 group_jkl012          /VIP
```

---

### Webhook Management

Create and manage webhooks for real-time event notifications.

#### List Webhooks

```bash
nylas webhook list
nylas webhook list --format json
nylas webhook list --format yaml
nylas webhook list --format csv
```

**Example output:**
```bash
$ nylas webhook list

ID                    Description              URL                                    Status    Triggers
────────────────────────────────────────────────────────────────────────────────────────────────────────────
webhook_abc123        Message notifications    https://api.myapp.com/webhooks/nylas   ● active  message.created, message.updated
webhook_def456        Calendar sync            https://api.myapp.com/calendar         ● active  event.created, event.updated
webhook_ghi789        Contact updates          https://api.myapp.com/contacts         ○ inactive contact.created

Total: 3 webhooks
```

#### Show Webhook

```bash
nylas webhook show <webhook-id>
nylas webhook show <webhook-id> --format json
```

**Example output:**
```bash
$ nylas webhook show webhook_abc123

Webhook: webhook_abc123
────────────────────────────────────────────────────────────
Description:  Message notifications
URL:          https://api.myapp.com/webhooks/nylas
Status:       ● active
Secret:       wh_s****************************cret

Trigger Types:
  • message.created
  • message.updated
  • message.opened

Notification Emails:
  • admin@myapp.com

Timestamps:
  Created:        2024-12-01T10:00:00Z
  Updated:        2024-12-15T14:30:00Z
  Status Updated: 2024-12-15T14:30:00Z
```

#### Create Webhook

```bash
# Create with required fields
nylas webhook create --url https://example.com/webhook --triggers message.created

# Create with multiple triggers
nylas webhook create --url https://example.com/webhook \
  --triggers message.created,event.created,contact.created

# Create with description and notification email
nylas webhook create --url https://example.com/webhook \
  --triggers message.created \
  --description "My message webhook" \
  --notify admin@example.com
```

**Example output:**
```bash
$ nylas webhook create --url https://api.myapp.com/webhook --triggers message.created --description "New messages"

✓ Webhook created successfully!

  ID:     webhook_new_123
  URL:    https://api.myapp.com/webhook
  Status: active

Triggers:
  • message.created

Important: Save your webhook secret - it won't be shown again:
  Secret: wh_secret_abc123xyz789

Use this secret to verify webhook signatures.
```

#### Update Webhook

```bash
# Update URL
nylas webhook update <webhook-id> --url https://new.example.com/webhook

# Update triggers
nylas webhook update <webhook-id> --triggers message.created,message.updated

# Pause/resume webhook
nylas webhook update <webhook-id> --status inactive
nylas webhook update <webhook-id> --status active

# Update multiple properties
nylas webhook update <webhook-id> \
  --description "Updated webhook" \
  --triggers event.created,event.updated
```

**Example output:**
```bash
$ nylas webhook update webhook_abc123 --status inactive

✓ Webhook updated successfully!

  ID:     webhook_abc123
  URL:    https://api.myapp.com/webhooks/nylas
  Status: ○ inactive

Triggers:
  • message.created
  • message.updated
```

#### Delete Webhook

```bash
nylas webhook delete <webhook-id>
nylas webhook delete <webhook-id> --force   # Skip confirmation
```

**Example output:**
```bash
$ nylas webhook delete webhook_abc123

Webhook to delete:
  ID:  webhook_abc123
  URL: https://api.myapp.com/webhooks/nylas
  Description: Message notifications
  Triggers: [message.created message.updated]

Are you sure you want to delete this webhook? [y/N] y
✓ Webhook deleted successfully!
```

#### List Trigger Types

```bash
nylas webhook triggers
nylas webhook triggers --format json
nylas webhook triggers --format list
nylas webhook triggers --category message   # Filter by category
```

**Example output:**
```bash
$ nylas webhook triggers

Available Webhook Trigger Types
================================

🔑 Grant
   Authentication grant events

   • grant.created
   • grant.updated
   • grant.deleted
   • grant.expired

📧 Message
   Email message events

   • message.created
   • message.updated
   • message.opened
   • message.link_clicked

💬 Thread
   Email thread events

   • thread.replied

📅 Event
   Calendar event events

   • event.created
   • event.updated
   • event.deleted

👤 Contact
   Contact events

   • contact.created
   • contact.updated
   • contact.deleted

📆 Calendar
   Calendar events

   • calendar.created
   • calendar.updated
   • calendar.deleted

📁 Folder
   Email folder events

   • folder.created
   • folder.updated
   • folder.deleted

Usage:
  nylas webhook create --url <URL> --triggers message.created
  nylas webhook create --url <URL> --triggers message.created,event.created
```

#### Test Webhook

```bash
# Send a test event to a URL
nylas webhook test send https://example.com/webhook

# Get mock payload for a trigger type
nylas webhook test payload message.created
nylas webhook test payload event.created --format json
```

**Example output:**
```bash
$ nylas webhook test send https://api.myapp.com/webhook

✓ Test event sent successfully!

  URL: https://api.myapp.com/webhook

Check your webhook endpoint logs to verify the event was received.
```

---

### OTP Management

Extract one-time passwords from emails automatically.

```bash
nylas otp get [email]      # Get the latest OTP code
nylas otp get --raw        # Output just the code (for scripting)
nylas otp watch [email]    # Watch for new OTP codes
nylas otp watch -i 10      # Check every 10 seconds
nylas otp list             # List configured accounts
nylas otp messages [email] # Show recent messages (debug)
```

**Example output:**
```bash
$ nylas otp get

OTP Code Found
─────────────────────────────────────────────────────
  Code:    847293
  From:    noreply@service.com
  Subject: Your verification code
  Time:    2 minutes ago

Code copied to clipboard!
```

**Example output (watch):**
```bash
$ nylas otp watch

Watching for OTP codes... (Press Ctrl+C to stop)

[14:32:15] Checking for new codes...
[14:32:45] Checking for new codes...
[14:33:15] ✓ New OTP: 592847 from auth@service.com
           Subject: Your login code
           Code copied to clipboard!
```

---

### Global Flags

Available on all commands:

```
--json       Output in JSON format
--no-color   Disable color output
-v, --verbose Enable verbose output
--config     Custom config file path
```

---

## Configuration

Credentials are stored securely in your system keyring:
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **macOS**: Keychain
- **Windows**: Windows Credential Manager

Config file location: `~/.config/nylas/config.yaml`

---

## Security

This project follows security best practices:

- **No hardcoded credentials**: All API keys and secrets are stored in the system keyring
- **Comprehensive .gitignore**: Prevents accidental commit of sensitive files
- **Environment variables for testing**: Integration tests use environment variables for credentials
- **No credential files in repository**: The `.gitignore` blocks all common credential file patterns

### Running Integration Tests

Integration tests require Nylas API credentials. Set them via environment variables:

```bash
# Required
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"

# Optional
export NYLAS_CLIENT_ID="your-client-id"

# Run integration tests
go test -tags=integration ./internal/cli/...

# Run with verbose output
go test -tags=integration -v ./internal/cli/...
```

### Destructive Test Operations

Some tests can modify data (send emails, delete messages). These require explicit opt-in:

```bash
# Enable send email tests
export NYLAS_TEST_SEND_EMAIL=true
export NYLAS_TEST_EMAIL="your-test-email@example.com"

# Enable delete message tests
export NYLAS_TEST_DELETE_MESSAGE=true
```

---

## Architecture

The CLI follows hexagonal (ports and adapters) architecture:

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
    config.go        # Configuration types
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

```bash
# Run integration tests (requires NYLAS_API_KEY and NYLAS_GRANT_ID)
go test -tags=integration ./internal/cli/...
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

---

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Building

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

### Project Structure

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
├── Makefile                 # Build automation
├── go.mod                   # Go modules
├── go.sum                   # Module checksums
└── README.md                # This file
```

---

## API Reference

This CLI uses the [Nylas v3 API](https://developer.nylas.com/docs/api/v3/).

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT
