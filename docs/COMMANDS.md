# Nylas CLI Command Reference

Complete reference for all Nylas CLI commands.

> **Quick Links:** [README](../README.md) | [TUI](TUI.md) | [Architecture](ARCHITECTURE.md) | [Development](DEVELOPMENT.md) | [Security](SECURITY.md) | [Webhooks](WEBHOOKS.md)

---

## Authentication

Manage Nylas API authentication and multiple accounts.

```bash
nylas auth config     # Configure API credentials
nylas auth login      # Authenticate with email provider
nylas auth logout     # Revoke current authentication
nylas auth status     # Show authentication status
nylas auth whoami     # Show current user info
nylas auth list       # List all accounts
nylas auth show       # Show detailed grant information
nylas auth switch     # Switch between accounts
nylas auth add        # Manually add an existing grant
nylas auth token      # Show/copy API key
nylas auth revoke     # Revoke specific grant
nylas auth providers  # List available authentication providers
nylas auth detect     # Detect provider from email address
nylas auth scopes     # Show OAuth scopes for a grant
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

**Example: Show grant details**
```bash
$ nylas auth show abc123def456

════════════════════════════════════════════════════════════
Grant Details
════════════════════════════════════════════════════════════

Grant ID:    abc123def456
Email:       user@gmail.com
Provider:    Google
Status:      ✓ valid

Created:     Dec 15, 2024 10:00 AM
Updated:     Dec 16, 2024 2:30 PM

Scopes:
  • https://www.googleapis.com/auth/gmail.readonly
  • https://www.googleapis.com/auth/calendar
  • https://www.googleapis.com/auth/contacts.readonly

★ This is the default grant
```

**Example: List available providers**
```bash
$ nylas auth providers

Available Authentication Providers:

  google
    Name:       Google Workspace
    ID:         connector-abc123
    Scopes:     5 configured

  microsoft
    Name:       Microsoft 365
    ID:         connector-def456
    Scopes:     4 configured
```

**Example: Detect provider from email**
```bash
$ nylas auth detect user@gmail.com

Email:    user@gmail.com
Domain:   gmail.com
Provider: google

To authenticate:
  nylas auth login --provider google

$ nylas auth detect user@company.com

Email:    user@company.com
Domain:   company.com
Provider: imap

Note: Use IMAP for generic email providers. Configure IMAP/SMTP settings during authentication.

To authenticate:
  nylas auth login --provider imap
```

**Example: Show OAuth scopes for a grant**
```bash
$ nylas auth scopes abc123def456

Grant ID:  abc123def456
Email:     user@gmail.com
Provider:  google
Status:    valid

OAuth Scopes (3):
  1. https://www.googleapis.com/auth/gmail.readonly
     → Read-only access to Gmail
  2. https://www.googleapis.com/auth/calendar
     → Calendar access
  3. https://www.googleapis.com/auth/contacts.readonly
     → Read-only access to contacts
```

---

## Email Operations

Full email management including reading, sending, searching, and organizing.

### List Emails

```bash
nylas email list [grant-id]           # List recent emails
nylas email list --limit 20           # Specify number of emails
nylas email list --unread             # Show only unread
nylas email list --starred            # Show only starred
nylas email list --from "sender@example.com"  # Filter by sender
nylas email list --metadata key1:value  # Filter by metadata (key1-key5 only)
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

### Read Email

```bash
nylas email read <message-id>         # Read a specific email
nylas email show <message-id>         # Alias for read
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

### Send Email

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

# Send with tracking (opens and link clicks)
nylas email send --to "to@example.com" --subject "Newsletter" --body "..." \
  --track-opens --track-links --track-label "campaign-q4"

# Send with custom metadata
nylas email send --to "to@example.com" --subject "Order Confirmation" --body "..." \
  --metadata "order_id=12345" --metadata "customer_id=cust_abc"
```

**Tracking Options:**
- `--track-opens` - Track when recipients open the email
- `--track-links` - Track when recipients click links in the email
- `--track-label` - Label for grouping tracked emails (for analytics)
- `--metadata` - Custom key=value metadata pairs (can be specified multiple times)

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

### Search Emails

```bash
nylas email search "query"            # Search emails
nylas email search "query" --limit 50 # Search with custom limit
nylas email search "query" --from "sender@example.com"
nylas email search "query" --after "2024-01-01"
nylas email search "query" --before "2024-12-31"
nylas email search "query" --unread   # Only unread messages
nylas email search "query" --starred  # Only starred messages
nylas email search "query" --in INBOX # Search in specific folder
nylas email search "query" --has-attachment  # Only with attachments
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

### Mark Operations

```bash
nylas email mark read <message-id>      # Mark as read
nylas email mark unread <message-id>    # Mark as unread
nylas email mark starred <message-id>   # Star a message
nylas email mark unstarred <message-id> # Unstar a message
```

### Delete Email

```bash
nylas email delete <message-id>       # Delete an email
nylas email delete <message-id> -f    # Delete without confirmation
```

### Smart Compose (AI Email Generation)

Generate AI-powered email drafts using Nylas Smart Compose (requires Plus package):

```bash
# Generate a new email draft from scratch
nylas email smart-compose --prompt "Draft a thank you email for yesterday's meeting"

# Generate a reply to a specific message
nylas email smart-compose --message-id <msg-id> --prompt "Reply accepting the invitation"

# Output as JSON
nylas email smart-compose --prompt "Write a follow-up email" --json
```

**Features:**
- AI-powered email composition based on natural language prompts
- Context-aware replies to existing messages
- Max prompt length: 1000 tokens
- Requires Nylas Plus package subscription

**Note:** Smart Compose leverages AI to draft professional emails quickly. Always review and edit the generated content before sending.

### Email Tracking

Track email opens, link clicks, and replies via webhooks:

```bash
# View tracking information and setup guide
nylas email tracking-info

# Send an email with tracking enabled
nylas email send --to user@example.com \\
  --subject "Meeting Invite" \\
  --body "Let's schedule a meeting" \\
  --track-opens \\
  --track-links
```

**Tracking Features:**
- **Opens:** Track when recipients open your emails
- **Clicks:** Track when recipients click links in your emails
- **Replies:** Track when recipients reply to your messages

**Data Delivery:**
Tracking data is delivered via webhooks. Set up webhooks to receive notifications:

```bash
# Create webhook for tracking events
nylas webhook create --url https://your-server.com/webhooks \\
  --triggers message.opened,message.link_clicked,thread.replied
```

For detailed information about tracking setup and webhook payloads, run:
```bash
nylas email tracking-info
```

### Message Metadata

Manage custom metadata on messages for organization and filtering:

```bash
# View metadata information and usage guide
nylas email metadata info

# Show metadata for a specific message
nylas email metadata show <message-id>
nylas email metadata show <message-id> --json

# Filter messages by metadata (when listing)
nylas email list --metadata key1:project-alpha
nylas email list --metadata key2:urgent --limit 20
```

**Indexed Keys (Searchable):**
Only five keys support filtering in queries:
- `key1`, `key2`, `key3`, `key4`, `key5`

**Setting Metadata:**
Metadata can only be set when sending messages or creating drafts:

```bash
# Send with metadata
nylas email send --to user@example.com \\
  --subject "Project Update" \\
  --body "Status report" \\
  --metadata key1=project-alpha \\
  --metadata key2=status-update
```

**Features:**
- Store up to 50 custom key-value pairs per message
- Only `key1`-`key5` are indexed and searchable
- Cannot update metadata on existing messages
- Useful for categorization, tracking, and custom workflows

**Example filtering workflow:**
```bash
# Send emails with metadata tags
nylas email send --to team@company.com \\
  --subject "Sprint Planning" \\
  --metadata key1=sprint-23 \\
  --metadata key2=planning

# Later, filter by metadata
nylas email list --metadata key1:sprint-23
nylas email list --metadata key2:planning --unread
```

For detailed information about metadata usage and best practices, run:
```bash
nylas email metadata info
```

---

## Folder Management

Manage email folders and labels.

```bash
nylas email folders list              # List all folders
nylas email folders list --id         # List folders with IDs
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

## Thread Management

View and manage email conversations.

```bash
# List threads
nylas email threads list              # List threads
nylas email threads list --unread     # List unread threads
nylas email threads list --starred    # List starred threads
nylas email threads list --limit 20   # Limit results

# Show thread details
nylas email threads show <thread-id>  # Show a thread

# Search threads
nylas email threads search --subject "project"  # Search by subject
nylas email threads search --from "boss@company.com"  # Search by sender
nylas email threads search --unread --limit 10  # Search unread threads
nylas email threads search --after "2024-01-01" --before "2024-12-31"

# Mark threads
nylas email threads mark <thread-id> --read      # Mark as read
nylas email threads mark <thread-id> --unread    # Mark as unread
nylas email threads mark <thread-id> --star      # Star thread
nylas email threads mark <thread-id> --unstar    # Unstar thread

# Delete threads
nylas email threads delete <thread-id>           # Delete thread
nylas email threads delete <thread-id> --force   # Skip confirmation
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

## Draft Management

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

## Calendar Management

View calendars, manage events, and check availability.

### List Calendars

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

### Calendar Events

```bash
# List events
nylas calendar events list [grant-id]
nylas calendar events list --days 14        # Next 14 days
nylas calendar events list --limit 20       # Limit results
nylas calendar events list --calendar <id>  # Specific calendar
nylas calendar events list --show-cancelled # Include cancelled

# List events with timezone conversion (NEW)
nylas calendar events list --timezone America/Los_Angeles  # Convert to specific timezone
nylas calendar events list --show-tz                       # Show timezone abbreviations
nylas calendar events list --timezone Europe/London --show-tz  # Both

# Show event details
nylas calendar events show <event-id>
nylas calendar events show <event-id> --timezone Asia/Tokyo  # Show in specific timezone

# Create event
nylas calendar events create --title "Meeting" --start "2024-12-20 14:00" --end "2024-12-20 15:00"
nylas calendar events create --title "Vacation" --start "2024-12-25" --all-day
nylas calendar events create --title "Team Sync" --start "2024-12-20 10:00" \
  --participant "alice@example.com" --participant "bob@example.com"

# Create event with DST validation (automatically checks for conflicts)
nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM"

# Create event ignoring DST warnings
nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM" --ignore-dst-warning

# Delete event
nylas calendar events delete <event-id>
nylas calendar events delete <event-id> --force
```

**DST-Aware Event Creation (NEW):**

When creating events, the CLI automatically checks for Daylight Saving Time conflicts:
- **Spring Forward Gap**: Warns if time doesn't exist (e.g., 2:00-3:00 AM on DST start)
- **Fall Back Duplicate**: Warns if time occurs twice (e.g., 1:00-2:00 AM on DST end)
- Suggests alternative times
- Requires confirmation to proceed or use `--ignore-dst-warning` to skip

**Example DST Conflict Detection:**
```bash
$ nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM"

⚠️  DST Conflict Detected!

This time will not exist due to Daylight Saving Time (clocks spring forward)

Suggested alternatives:
  1. Schedule 1 hour earlier (before DST)
  2. Schedule at the requested time after DST
  3. Use a different date

Create anyway? [y/N]: n
Cancelled.
```

**Working Hours Validation (NEW):**

The CLI validates event times against configured working hours:
- **Default Hours**: 9:00 AM - 5:00 PM (if not configured)
- **Per-Day Configuration**: Different hours for different days
- **Weekend Support**: Separate weekend hours or disable weekends
- Warns when scheduling outside working hours
- Use `--ignore-working-hours` to skip validation

**Configuration Example:**
```yaml
# ~/.config/nylas/config.yaml
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
  friday:
    enabled: true
    start: "09:00"
    end: "15:00"  # Short Fridays
  weekend:
    enabled: false  # No work on weekends
```

**Example Working Hours Warning:**
```bash
$ nylas calendar events create --title "Late Call" --start "2025-01-15 18:00" --end "2025-01-15 19:00"

⚠️  Working Hours Warning

This event is scheduled outside your working hours:
  • Your hours: 09:00 - 17:00
  • Event time: 6:00 PM Local
  • 1 hour(s) after end

Create anyway? [y/N]: n
Cancelled.

# Or skip validation:
$ nylas calendar events create --title "Late Call" --start "2025-01-15 18:00" --ignore-working-hours
✓ Event created successfully!
```

**Break Time Protection (NEW):**

Protect your lunch breaks and other break periods with hard-block enforcement:

- **Hard Block**: Cannot schedule events during breaks (unlike working hours which allow override)
- **Multiple Breaks**: Configure lunch, coffee breaks, and custom break periods
- **Per-Day Breaks**: Different break times for different days
- Use `--ignore-working-hours` to skip break validation

**Configuration Example:**
```yaml
# ~/.config/nylas/config.yaml
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
  friday:
    enabled: true
    start: "09:00"
    end: "15:00"
    breaks:
      - name: "Lunch"
        start: "11:30"
        end: "12:30"  # Earlier lunch on Fridays
        type: "lunch"
```

**Example Break Conflict:**
```bash
$ nylas calendar events create --title "Quick Sync" --start "2025-01-15 12:30" --end "2025-01-15 13:00"

⛔ Break Time Conflict

Event cannot be scheduled during Lunch (12:00 - 13:00)

Tip: Schedule the event outside of break times, or update your
     break configuration in ~/.nylas/config.yaml
Error: event conflicts with break time

# Break blocks are enforced - you must reschedule:
$ nylas calendar events create --title "Quick Sync" --start "2025-01-15 13:00" --end "2025-01-15 13:30"
✓ Event created successfully!
```

**Timezone Locking (NEW):**

Lock events to a specific timezone to prevent automatic conversion when viewing from different locations. Perfect for in-person events, conferences, or meetings in specific locations:

- **Lock on Creation**: Use `--lock-timezone` when creating events
- **Locked Display**: Shows 🔒 indicator next to time
- **No Auto-Convert**: Time always displays in locked timezone
- **Lock/Unlock**: Use `--lock-timezone` or `--unlock-timezone` in update command

**Example Timezone Locking:**
```bash
# Create event locked to NYC timezone (for in-person meeting)
$ nylas calendar events create \
    --title "NYC Office All-Hands" \
    --start "2025-01-15 09:00" \
    --location "New York Office" \
    --lock-timezone

✓ Event created successfully!

Title: NYC Office All-Hands
When: Wed, Jan 15, 2025, 9:00 AM - 10:00 AM
🔒 Timezone locked: America/New_York
     This event will always display in this timezone, regardless of viewer's location.
ID: event-123

# View locked event (shows lock indicator)
$ nylas calendar events show event-123

NYC Office All-Hands

When
  Wed, Jan 15, 2025, 9:00 AM - 10:00 AM EST 🔒
  (Your local: 6:00 AM PST)

Location
  New York Office

# Unlock timezone
$ nylas calendar events update event-123 --unlock-timezone

✓ Event updated successfully!
🔓 Timezone lock removed
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

**Example output (with timezone conversion):**
```bash
$ nylas calendar events list --timezone America/Los_Angeles

Found 3 event(s):

Team Standup
  When: Mon, Dec 16, 2024, 6:00 AM - 6:30 AM PST
        (Original: 9:00 AM - 9:30 AM EST)
  Location: Zoom
  Status: confirmed
  ID: event_abc123

Client Call
  When: Tue, Dec 17, 2024, 11:00 AM - 12:00 PM PST
        (Original: 7:00 PM - 8:00 PM GMT)
  Location: Google Meet
  Status: confirmed
  ID: event_def456
```

**Example output (show timezone info):**
```bash
$ nylas calendar events list --show-tz

Team Standup
  When: Mon, Dec 16, 2024, 9:00 AM - 9:30 AM EST
  Timezone: America/New_York (EST, UTC-5)
  Location: Conference Room A
  Status: confirmed
  ID: event_abc123
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

**Example output (list with timezone conversion):**
```bash
$ nylas calendar events list --timezone America/Los_Angeles --show-tz

Found 3 event(s):

Global Team Sync
  When: Mon, Dec 23, 2024, 6:00 AM - 7:00 AM PST
        (Original: Mon, Dec 23, 2024, 9:00 AM - 10:00 AM EST)
  Location: Zoom
  Status: confirmed
  Guests: 12 participant(s)
  ID: event_xyz123

Client Meeting
  When: Tue, Dec 24, 2024, 11:00 AM - 12:00 PM PST
        (Original: Tue, Dec 24, 2024, 2:00 PM - 3:00 PM EST)
  Status: confirmed
  Guests: 3 participant(s)
  ID: event_abc456

Holiday Party
  When: Fri, Dec 27, 2024 (all day)
  Location: Main Office
  Status: confirmed
  ID: event_def789
```

**Example output (show with timezone conversion):**
```bash
$ nylas calendar events show event_xyz123 --timezone Europe/London --show-tz

Global Team Sync

When
  Mon, Dec 23, 2024, 2:00 PM - 3:00 PM GMT
  (Original: Mon, Dec 23, 2024, 9:00 AM - 10:00 AM EST)

Location
  Zoom

Description
  Quarterly planning session with global team members.

Participants
  Alice (New York) <alice@company.com> ✓ accepted
  Bob (London) <bob@company.com> ✓ accepted
  Carol (Tokyo) <carol@company.com> ✓ accepted
  David (Sydney) <david@company.com> ? tentative

Video Conference
  Provider: zoom
  URL: https://zoom.us/j/987654321

Details
  Status: confirmed
  Busy: true
  ID: event_xyz123
  Calendar: cal_primary_123
```

### AI-Powered Scheduling

**NEW:** Schedule meetings using natural language with AI assistance. Supports multiple LLM providers including local privacy-first options.

```bash
# Basic AI scheduling
nylas calendar schedule ai "30-minute meeting with john@example.com next Tuesday afternoon"

# Use specific AI provider
nylas calendar schedule ai --provider claude "team meeting tomorrow morning"
nylas calendar schedule ai --provider openai "quarterly planning next week"
nylas calendar schedule ai --provider groq "quick sync with alice"

# Privacy mode (local LLM)
nylas calendar schedule ai --privacy "sensitive meeting about project X"

# Auto-confirm first option
nylas calendar schedule ai --yes "lunch with team next Friday"

# Specify your timezone
nylas calendar schedule ai --timezone America/Los_Angeles "call with UK team"

# Limit number of suggestions
nylas calendar schedule ai --max-options 5 "1-hour review meeting"
```

**Example Output:**
```bash
$ nylas calendar schedule ai "30-minute meeting with john@example.com next Tuesday afternoon"

🤖 AI Scheduling Assistant
Provider: Claude (Anthropic)

Processing your request: "30-minute meeting with john@example.com next Tuesday afternoon"

Top 3 AI-Suggested Times:

1. 🟢 Tuesday, Jan 21, 2:00 PM PST (Score: 94/100)
   you@example.com: 2:00 PM - 2:30 PM PST
   john@example.com: 5:00 PM - 5:30 PM EST

   Why this is good:
   • Both in working hours
   • No conflicts detected
   • Your calendar shows high productivity at 2 PM historically

2. 🟡 Tuesday, Jan 21, 1:00 PM PST (Score: 82/100)
   you@example.com: 1:00 PM - 1:30 PM PST
   john@example.com: 4:00 PM - 4:30 PM EST

   Why this is good:
   • Post-lunch slot, moderate energy
   • Late afternoon for John (still acceptable)

3. 🟢 Tuesday, Jan 21, 3:00 PM PST (Score: 90/100)
   you@example.com: 3:00 PM - 3:30 PM PST
   john@example.com: 6:00 PM - 6:30 PM EST

   ⚠️  Warnings:
   • Near end of working hours for John

Create meeting with option #1? [y/N/2/3]: y

Creating event...
✓ Event created
  Title: Meeting with john
  When: Tuesday, Jan 21, 2025, 2:00 PM - 2:30 PM PST
  Participants: john@example.com

💰 Estimated cost: ~$0.0150 (1500 tokens)
```

**Privacy Mode (Ollama - Local LLM):**
```bash
$ nylas calendar schedule ai --privacy "team standup tomorrow 9am"

🤖 AI Scheduling Assistant (Privacy Mode)
Provider: Ollama (Local LLM)

Processing locally... ✓

[... AI suggestions ...]

🔒 Privacy: All processing done locally, no data sent to cloud.
```

**Supported AI Providers:**
- `ollama` - Local LLM (privacy-first, free, no API key needed)
- `claude` - Anthropic Claude (best for complex scheduling)
- `openai` - OpenAI GPT-4 (well-balanced)
- `groq` - Groq (very fast, cheap)

**Configuration:**
Add AI configuration to `~/.nylas/config.yaml`:

```yaml
ai:
  default_provider: ollama  # Default provider

  fallback:
    enabled: true
    providers: [ollama, claude]  # Try in order

  ollama:
    host: http://localhost:11434
    model: mistral:latest

  claude:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022

  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4-turbo

  groq:
    api_key: ${GROQ_API_KEY}
    model: mixtral-8x7b-32768
```

**AI Features:**
- Natural language parsing
- Multi-timezone analysis
- Working hours validation
- DST transition detection
- Participant availability checking
- Meeting time scoring (0-100)
- Detailed reasoning for each option
- Function calling for calendar operations

### Predictive Calendar Analytics

**NEW:** Analyze your meeting history to learn patterns and get AI-powered recommendations for optimizing your calendar.

```bash
# Analyze last 90 days of meetings
nylas calendar analyze

# Analyze custom time period
nylas calendar analyze --days 60

# Score a specific meeting time
nylas calendar analyze --score-time "2025-01-15T14:00:00Z" \
  --participants john@example.com \
  --duration 30

# Show recommendations
nylas calendar analyze --apply
```

**Example Output:**
```bash
$ nylas calendar analyze

🔍 Analyzing 90 days of meeting history...

📊 Analysis Period: 2024-09-22 to 2024-12-21
📅 Total Meetings Analyzed: 156

✅ Meeting Acceptance Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall Acceptance Rate: 84.6%

By Day of Week:
    Monday: 78.3% ████████████████
   Tuesday: 92.1% ██████████████████
 Wednesday: 88.7% ██████████████████
  Thursday: 86.4% █████████████████
    Friday: 64.2% █████████████

By Time of Day (working hours):
  09:00: 72.4% ███████████████
  10:00: 88.9% ██████████████████
  11:00: 91.2% ██████████████████
  14:00: 85.6% █████████████████
  15:00: 79.3% ████████████████

⏱️  Meeting Duration Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Average Scheduled: 34 minutes
Average Actual: 38 minutes
Overrun Rate: 41.7%

🌍 Timezone Distribution
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  America/Los_Angeles: 89 meetings
  America/New_York: 42 meetings
  Europe/London: 25 meetings

🎯 Productivity Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Peak Focus Times (recommended for deep work):
  1. Tuesday 10:00-12:00 (score: 92/100)
  2. Thursday 10:00-12:00 (score: 88/100)
  3. Wednesday 14:00-16:00 (score: 85/100)
  4. Tuesday 14:00-16:00 (score: 82/100)
  5. Thursday 14:00-16:00 (score: 79/100)

Meeting Density by Day:
    Monday: 3.2 meetings/day
   Tuesday: 2.8 meetings/day
 Wednesday: 3.1 meetings/day
  Thursday: 2.9 meetings/day
    Friday: 1.4 meetings/day

💡 AI Recommendations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. 🔴 Block Tuesday 10:00-12:00 for focus time [focus_time]
   Historical data shows you have few meetings during this time and accept 92% of meetings outside this block
   📌 Action: Create recurring focus time block
   📈 Impact: Increase productivity by 20-30%
   🎯 Confidence: 92%

2. 🟡 Adjust default meeting duration to 40 minutes [duration_adjustment]
   Your meetings typically run 4 minutes over the scheduled 30 minutes
   📌 Action: Update meeting templates
   📈 Impact: Reduce schedule overruns by 40%
   🎯 Confidence: 78%

3. 🟡 Prefer Tuesday/Wednesday afternoons for team meetings [scheduling_preference]
   Acceptance rate is 88% for Tuesday/Wednesday vs 71% for Monday/Friday
   📌 Action: Suggest Tuesday/Wednesday in meeting invites
   📈 Impact: Reduce declined meetings by 15%
   🎯 Confidence: 85%

📝 Key Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. You accept 92% of meetings on Tuesdays but only 64% on Fridays
2. Your meetings run 12% longer than scheduled on average
3. You have the most focus time on Tuesdays and Thursdays between 10-12 AM
4. Most of your meetings (57%) are with participants in Pacific timezone
```

**Scoring a Specific Meeting Time:**
```bash
$ nylas calendar analyze --score-time "2025-01-21T14:00:00Z" \
  --participants john@example.com \
  --duration 30

🔍 Analyzing historical patterns...

🎯 Meeting Score for Tuesday, Jan 21, 2025 at 2:00 PM PST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🟢 Overall Score: 86/100
   █████████████████

🎯 Confidence: 85%
📊 Historical Success Rate: 88%

Contributing Factors:
  ➕ Day Preference: +12
     88.7% acceptance rate on Tuesdays
  ➕ Time Preference: +10
     85.6% acceptance rate at 14:00
  ⚪ Productivity: +5
     Moderate productivity time
  ➕ Participant Match: +7
     Based on historical meetings with these participants
  ⚪ Timezone: +0
     Time works well for all timezones

💡 Good time - aligns well with your preferences
```

**Privacy & Local Storage:**
- All pattern learning happens locally
- No meeting data sent to cloud servers
- Patterns stored in `~/.nylas/patterns.json`
- GDPR/HIPAA compliant

**What Gets Analyzed:**
- Meeting acceptance/decline patterns by day and time
- Actual vs scheduled meeting durations
- Timezone distribution of participants
- Productivity windows (times with fewer meetings)
- Per-participant scheduling preferences

**How It Works:**
1. Fetches last 90 days of calendar events
2. Analyzes patterns using local ML algorithms
3. Generates personalized recommendations
4. All processing done locally (privacy-first)

### Conflict Detection & Smart Rescheduling

**NEW:** AI-powered conflict detection and intelligent meeting rescheduling with alternative time suggestions.

**Check for Conflicts:**
```bash
# Check conflicts for a proposed meeting
nylas calendar conflicts check \
  --title "Product Review" \
  --start "2025-01-22T14:00:00Z" \
  --duration 60 \
  --participants team@company.com

# Check and auto-select best alternative
nylas calendar conflicts check \
  --title "Team Sync" \
  --start "2025-01-23T10:00:00Z" \
  --duration 30 \
  --auto-resolve
```

**Example Output:**
```bash
$ nylas calendar conflicts check --title "Weekly Standup" \
  --start "2025-01-22T10:00:00Z" --duration 30

🔍 Analyzing your calendar patterns...
✓ Analyzed 156 meetings from last 90 days

⚙️  Detecting conflicts...

📊 Conflict Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔴 Hard Conflicts (1)

1. Overlaps with 'Executive Review'
   Event: Executive Review
   Time: Wed, Jan 22 at 10:15 AM PST
   Status: confirmed
   Impact: Cannot attend both meetings simultaneously
   Suggestion: Reschedule to avoid overlap

🟡 Soft Conflicts (2)

1. ⏱️ Back-to-back with 'Team Planning'
   Severity: medium
   Impact: No buffer time between meetings
   ✓ Can auto-resolve

2. 🎯 Interrupts focus time block
   Severity: high
   Impact: Conflicts with Tuesday 10:00-12:00 focus block
   ✓ Can auto-resolve

💡 Recommendations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Add 15-minute buffer before/after meetings
  Consider Tuesday afternoon instead (92% acceptance rate)

🔄 Suggested Alternative Times
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 🟢 Wed, Jan 22, 2025 at 2:00 PM PST (Score: 88/100)

   Pros:
   ✓ High acceptance rate on Wednesdays (88.7%)
   ✓ Preferred time slot (85.6% acceptance)
   ✓ No conflicts detected

   💡 This time aligns well with team availability patterns

2. 🟢 Thu, Jan 23, 2025 at 10:00 AM PST (Score: 85/100)

   Pros:
   ✓ High acceptance rate on Thursdays (86.4%)
   ✓ Good time for collaborative work

   Cons:
   ⚠️  Close to another meeting (11 min gap)

3. 🟡 Wed, Jan 22, 2025 at 3:00 PM PST (Score: 74/100)

   Pros:
   ✓ Same day as original
   ✓ No hard conflicts

   Cons:
   ⚠️  Lower acceptance rate for afternoon slots

🤖 AI Recommendation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Consider rescheduling to Wednesday 2:00 PM. This time has:
- 88% historical acceptance rate
- No scheduling conflicts
- Good match for team availability patterns
- Optimal for collaborative work based on past meetings

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Status: Cannot proceed (hard conflicts)
```

**AI-Powered Rescheduling:**
```bash
# Get AI suggestions for rescheduling an event
nylas calendar reschedule ai event_abc123 \
  --reason "Conflict with client meeting"

# Reschedule with constraints
nylas calendar reschedule ai event_abc123 \
  --max-delay-days 7 \
  --avoid-days Friday \
  --must-include john@company.com

# Auto-select best time and notify participants
nylas calendar reschedule ai event_abc123 \
  --reason "Calendar conflict" \
  --auto-select \
  --notify
```

**Example Reschedule Output:**
```bash
$ nylas calendar reschedule ai event_abc123

📅 Fetching event event_abc123...
✓ Found: Weekly Team Sync
  Current time: Wed, Jan 22, 2025 at 10:00 AM PST

🔍 Analyzing your calendar patterns...
✓ Analyzed 156 meetings from last 90 days

⚙️  Finding optimal alternative times...

📊 Reschedule Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Reason: Conflict with client meeting

🔄 Found 5 Alternative Time(s)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 🟢 Wed, Jan 22, 2025 at 2:00 PM PST (Score: 92/100)

   Pros:
   ✓ High acceptance rate on Wednesdays (88.7%)
   ✓ Preferred time slot (88% acceptance)
   ✓ Good match for team schedules

2. 🟢 Thu, Jan 23, 2025 at 10:00 AM PST (Score: 89/100)

   Pros:
   ✓ Same time, next day
   ✓ High acceptance rate on Thursdays

   ⚠️  1 soft conflict(s)

3. 🟡 Wed, Jan 22, 2025 at 11:00 AM PST (Score: 78/100)

   Pros:
   ✓ Same day as original
   ✓ One hour later

   Cons:
   ⚠️  Back-to-back with another meeting

💡 To apply a suggestion, use:
   nylas calendar events update event_abc123 --start 2025-01-22T14:00:00Z
```

**Available Flags:**

Conflict Check:
- `--title` - Meeting title (required)
- `--start` - Start time in RFC3339 format (required)
- `--end` - End time (optional, uses --duration if not set)
- `--duration` - Duration in minutes (default: 60)
- `--participants` - Participant email addresses
- `--auto-resolve` - Automatically select best alternative

AI Reschedule:
- `--reason` - Reason for rescheduling
- `--preferred-times` - Preferred alternative times (RFC3339 format)
- `--max-delay-days` - Maximum days to delay (default: 14)
- `--notify` - Send notification to participants
- `--auto-select` - Automatically apply best alternative
- `--must-include` - Emails that must be available
- `--avoid-days` - Days to avoid (e.g., Friday, Monday)

**Conflict Types Detected:**

Hard Conflicts (blocking):
- Overlapping meetings - Cannot attend both simultaneously

Soft Conflicts (warnings):
- Back-to-back meetings - No buffer time between events
- Focus time interruption - Conflicts with productive work blocks
- Meeting overload - Too many meetings in one day (6+)
- Close proximity - Less than 15 minutes between meetings

**How Conflict Detection Works:**
1. Analyzes proposed meeting time
2. Scans all calendars for conflicts
3. Uses learned patterns to detect soft conflicts
4. Scores alternative times using ML algorithm
5. Suggests top 3-5 alternative times with reasoning

**Reschedule Scoring Algorithm:**
The AI considers multiple factors when scoring alternatives:
- Historical acceptance patterns (day/time preferences)
- Participant availability and preferences
- Meeting density and calendar balance
- Focus time protection
- Timezone fairness for distributed teams

Score ranges:
- 🟢 85-100: Excellent match
- 🟡 70-84: Good option
- 🔴 0-69: Suboptimal (consider other times)

### AI Focus Time Protection

Automatically protect deep work time by analyzing productivity patterns and blocking focus time.

```bash
# Analyze productivity patterns and enable focus time protection
nylas calendar ai focus-time --enable

# Analyze patterns without enabling protection
nylas calendar ai focus-time --analyze

# Create recommended focus blocks
nylas calendar ai focus-time --create

# Customize target focus hours per week
nylas calendar ai focus-time --enable --target-hours 12

# Enable with auto-decline for meeting requests
nylas calendar ai focus-time --enable --auto-decline

# Allow urgent meeting overrides
nylas calendar ai focus-time --enable --allow-override
```

**Example output:**
```bash
$ nylas calendar ai focus-time --enable

🧠 AI Focus Time Protection

Analyzing your productivity patterns...

✨ Discovered Focus Patterns:

  • Peak productivity:
    - Tuesday: 10:00--12:00 (95% focus score) ⭐ Top
    - Thursday: 10:00--12:00 (92% focus score)
    - Wednesday: 09:00--11:00 (85% focus score)

  • Deep work sessions: Average 2.5 hours
  • Most productive day: Wednesday (fewest interruptions)

📅 AI-Recommended Focus Time Blocks:

Weekly Schedule:
  Monday:    ░░░░░░░░░░░░████████░░░░░░░░ 2.0 hrs
  Tuesday:   ████████████░░░░░░░░░░░░░░░░ 2.0 hrs ⭐ Peak
  Wednesday: ████████████████████░░░░░░░░ 4.0 hrs 🎯
  Thursday:  ████████████░░░░░░░░░░░░░░░░ 2.0 hrs ⭐ Peak
  Friday:    ░░░░░░░░░░░░░░░░░░░░████████ 2.0 hrs

Total: 14.0 hours/week protected for focus time

🛡️  Protection Rules:
  1. Auto-decline meeting requests during focus blocks
  2. Suggest alternative times when requests come in
  3. Allow override for "urgent" meetings (you approve)
  4. Dynamically adjust if deadline pressure increases

💡 AI Insights:

  • Your peak productivity is Tuesday at 10:00--12:00 (95% focus score)
  • High meeting density on [Monday Friday] - consider protecting more focus time on these days
  • AI recommends 14.0 hours/week of protected focus time across 5 blocks

📊 Confidence: 100%
   Based on 90 days of calendar history

✅ Focus time protection is enabled!

To create these focus blocks in your calendar, run:
  nylas calendar ai focus-time --create
```

**Creating Focus Blocks:**
```bash
$ nylas calendar ai focus-time --create

🔨 Creating Focus Time Blocks...

✅ Created 5 focus time blocks:

1. Peak productivity time (95% score)
   📅 Tuesday, 10:00 AM--12:00 PM (120 min)
   🔒 Protected with auto-decline: true
   📆 Calendar Event ID: evt_abc123

2. Peak productivity time (92% score)
   📅 Thursday, 10:00 AM--12:00 PM (120 min)
   🔒 Protected with auto-decline: true
   📆 Calendar Event ID: evt_def456

...

✨ Focus time blocks are now protected in your calendar!

To view adaptive schedule recommendations, run:
  nylas calendar ai adapt
```

### Adaptive Schedule Optimization

Real-time adaptive schedule optimization based on changing priorities and workload.

```bash
# Detect and suggest adaptive changes
nylas calendar ai adapt

# Adapt for specific triggers
nylas calendar ai adapt --trigger overload      # Meeting overload
nylas calendar ai adapt --trigger deadline      # Deadline change
nylas calendar ai adapt --trigger focus-risk    # Focus time at risk

# Automatically apply recommended changes
nylas calendar ai adapt --auto-apply
```

**Example output:**
```bash
$ nylas calendar ai adapt

🔄 AI Adaptive Scheduling

Analyzing schedule changes and workload...

📊 Detected Changes:

  • Trigger: Meeting overload detected
  • Affected events: 3
  • Confidence: 85%

📈 Predicted Impact:

  • Focus time gained: 2.0 hours
  • Meetings to reschedule: 2
  • Time saved: 30 minutes
  • Conflicts resolved: 1

  Predicted benefit: Improved focus time availability

🤖 AI Adaptive Actions:

1. Move low-priority meeting to reduce meeting overload
   Event ID: evt_123

2. Move low-priority meeting to reduce meeting overload
   Event ID: evt_456

3. Add additional focus blocks due to deadline pressure

⏸️  Changes require approval (use --auto-apply to apply automatically)

To approve these changes, run:
  nylas calendar ai adapt --auto-apply
```

**How Adaptive Scheduling Works:**
1. Monitors schedule changes and workload patterns
2. Detects triggers (deadline changes, meeting overload, focus time erosion)
3. Analyzes impact of proposed changes
4. Suggests optimizations to protect focus time and reduce overload
5. Learns from historical patterns to improve recommendations

**Adaptive Triggers:**
- **Meeting Overload**: Too many meetings scheduled (18+ hours/week)
- **Deadline Change**: Project deadline moved up, need more focus time
- **Focus Time At Risk**: Protected focus blocks being eroded by meetings
- **Priority Shift**: Task priorities changed, schedule needs adjustment

### Calendar Availability

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

### Smart Meeting Finder (Multi-Timezone)

**NEW:** Find optimal meeting times across multiple timezones with intelligent scoring.

The smart meeting finder analyzes participant timezones and suggests meeting times using a 100-point scoring algorithm that considers:
- **Working Hours (40 pts)**: All participants within working hours
- **Time Quality (25 pts)**: Quality of time for participants (morning/afternoon preference)
- **Cultural Considerations (15 pts)**: Respects cultural norms (no Friday PM, no lunch hour, no Monday early AM)
- **Weekday Preference (10 pts)**: Prefers mid-week meetings (Tuesday/Wednesday best)
- **Holiday Check (10 pts)**: Avoids holidays

```bash
# Find optimal meeting time for multiple participants
nylas calendar find-time --participants alice@example.com,bob@example.com --duration 1h

# Custom working hours and date range
nylas calendar find-time \
  --participants alice@example.com,bob@example.com,carol@example.com \
  --duration 1h \
  --working-start 09:00 \
  --working-end 17:00 \
  --days 7

# 30-minute meeting with weekend availability
nylas calendar find-time \
  --participants team@example.com \
  --duration 30m \
  --exclude-weekends=false
```

**Example output:**
```bash
$ nylas calendar find-time --participants alice@example.com,bob@example.com --duration 1h

🌍 Multi-Timezone Meeting Finder

Participants:
  • alice@example.com: America/New_York
  • bob@example.com: Europe/London

Top 3 Suggested Times:

1. 🟢 Tuesday, Jan 7, 10:00 AM PST (Score: 94/100)
   alice: 1:00 PM - 2:00 PM America/New_York (Good)
   bob: 6:00 PM - 7:00 PM Europe/London (Poor ⚠️)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 22/25
   • Cultural: 15/15
   • Weekday: 10/10
   • Holidays: 7/10

2. 🟢 Wednesday, Jan 8, 11:00 AM PST (Score: 92/100)
   alice: 2:00 PM - 3:00 PM America/New_York (Good)
   bob: 7:00 PM - 8:00 PM Europe/London (Bad 🔴)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 20/25
   • Cultural: 15/15
   • Weekday: 10/10
   • Holidays: 7/10

3. 🟡 Thursday, Jan 9, 9:00 AM PST (Score: 75/100)
   alice: 12:00 PM - 1:00 PM America/New_York (Good)
   bob: 5:00 PM - 6:00 PM Europe/London (Poor ⚠️)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 18/25
   • Cultural: 12/15
   • Weekday: 8/10
   • Holidays: 7/10

💡 Recommendation: Book option #1 for best overall experience
```

**Scoring Legend:**
- 🟢 Excellent (85-100): Great time for all participants
- 🟡 Good (70-84): Acceptable with minor compromises
- 🔴 Poor (<70): Significant compromises, consider alternatives

**Time Quality Labels:**
- ✨ Excellent: 9-11 AM
- Good: 11 AM - 2 PM
- Fair: 2-5 PM
- ⚠️ Poor: 8-9 AM or 5-6 PM
- 🔴 Bad: Outside working hours

### Virtual Calendars

Virtual calendars allow scheduling without connecting to a third-party provider. They're perfect for conference rooms, equipment, or external contractors.

**Features:**
- No OAuth required
- Never expire
- Support calendar and event operations only (no email/contacts)
- Maximum 10 calendars per virtual account

```bash
# List all virtual calendar grants
nylas calendar virtual list
nylas calendar virtual list --json

# Create a virtual calendar grant
nylas calendar virtual create --email conference-room-a@company.com
nylas calendar virtual create --email projector-1@company.com

# Show virtual calendar grant details
nylas calendar virtual show <grant-id>
nylas calendar virtual show <grant-id> --json

# Delete a virtual calendar grant
nylas calendar virtual delete <grant-id>
nylas calendar virtual delete <grant-id> -y  # Skip confirmation
```

**Example workflow:**
```bash
# 1. Create a virtual calendar grant for a conference room
$ nylas calendar virtual create --email conference-room-a@company.com
✓ Created virtual calendar grant
  ID:     vcal-grant-123abc
  Email:  conference-room-a@company.com
  Status: valid

# 2. Create a calendar for this virtual grant
$ nylas calendar create vcal-grant-123abc --name "Conference Room A"
✓ Created calendar
  ID:   primary
  Name: Conference Room A

# 3. Create events on the virtual calendar
$ nylas calendar events create vcal-grant-123abc primary \
  --title "Board Meeting" \
  --start "2024-01-15T14:00:00" \
  --end "2024-01-15T16:00:00"
✓ Created event
```

### Recurring Events

Manage recurring calendar events, including viewing all instances and updating or deleting specific occurrences.

**Supported recurrence patterns:**
- Daily: `RRULE:FREQ=DAILY;COUNT=5`
- Weekly: `RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=10`
- Monthly: `RRULE:FREQ=MONTHLY;COUNT=12`
- Yearly: `RRULE:FREQ=YEARLY;COUNT=3`

```bash
# List all instances of a recurring event
nylas calendar recurring list <master-event-id> --calendar <calendar-id>
nylas calendar recurring list event-123 --calendar cal-456 --limit 100
nylas calendar recurring list event-123 --calendar cal-456 \
  --start 1704067200 --end 1706745600

# Update a single instance
nylas calendar recurring update <instance-id> --calendar <calendar-id> \
  --title "Updated Meeting Title"
nylas calendar recurring update instance-789 --calendar cal-456 \
  --start "2024-01-15T14:00:00" --end "2024-01-15T15:30:00" \
  --location "Conference Room B"

# Delete a single instance (creates an exception)
nylas calendar recurring delete <instance-id> --calendar <calendar-id>
nylas calendar recurring delete instance-789 --calendar cal-456 -y
```

**Example output (list):**
```bash
$ nylas calendar recurring list event-master-123 --calendar primary

INSTANCE ID        TITLE                   START TIME        STATUS
event-inst-1       Weekly Team Meeting     2024-01-08 10:00  confirmed
event-inst-2       Weekly Team Meeting     2024-01-15 10:00  confirmed
event-inst-3       Weekly Team Meeting     2024-01-22 10:00  confirmed
event-inst-4       Weekly Team Meeting     2024-01-29 10:00  confirmed

Total instances: 4
Master Event ID: event-master-123
```

**Understanding recurring events:**
- **Master Event ID**: The parent event that defines the recurrence pattern
- **Instance**: A single occurrence of the recurring series
- **Exception**: An instance that has been modified or deleted
- When you update an instance, it becomes an exception with custom properties
- When you delete an instance, it adds an EXDATE to the recurrence rule

---

## Contacts Management

Manage contacts and contact groups.

### List Contacts

```bash
nylas contacts list [grant-id]
nylas contacts list --limit 100
nylas contacts list --id                      # Show contact IDs
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

### Show Contact

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

### Create Contact

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

### Update Contact

```bash
nylas contacts update <contact-id> [grant-id]
nylas contacts update <contact-id> --given-name "John" --surname "Smith"
nylas contacts update <contact-id> --company "Acme Inc" --job-title "Engineer"
nylas contacts update <contact-id> --email "new@example.com" --phone "+1-555-0123"
nylas contacts update <contact-id> --birthday "1990-05-15" --notes "Updated notes"
```

**Example output:**
```bash
$ nylas contacts update contact_abc123 --given-name "John" --surname "Smith"

✓ Contact updated successfully!

Name: John Smith
ID: contact_abc123
```

### Delete Contact

```bash
nylas contacts delete <contact-id> [grant-id]
nylas contacts delete <contact-id> --force   # Skip confirmation
```

### Contact Groups

Manage contact groups with full CRUD operations.

```bash
# List groups
nylas contacts groups list [grant-id]

# Show group details
nylas contacts groups show <group-id> [grant-id]

# Create group
nylas contacts groups create "VIP Clients" [grant-id]

# Update group
nylas contacts groups update <group-id> --name "Premium Clients"

# Delete group
nylas contacts groups delete <group-id> [grant-id]
nylas contacts groups delete <group-id> --force   # Skip confirmation
```

**Example output:**
```bash
$ nylas contacts groups list

Found 4 contact group(s):

NAME                ID                    PATH
Family              group_abc123          /Family
Work                group_def456          /Work
Friends             group_ghi789          /Friends
VIP                 group_jkl012          /VIP
```

**Example: Create a contact group**
```bash
$ nylas contacts groups create "VIP Clients"

✓ Contact group created successfully!

Name: VIP Clients
ID: group_new_123
```

### Advanced Contact Search

Search contacts with advanced filtering options including company name, email, phone, and more.

```bash
# Basic search
nylas contacts search [grant-id]

# Search by company name (partial match, case-insensitive)
nylas contacts search --company "Acme"

# Search by email address
nylas contacts search --email "john@example.com"

# Search by phone number
nylas contacts search --phone "+1-555-0101"

# Filter by contact source
nylas contacts search --source address_book
nylas contacts search --source inbox
nylas contacts search --source domain

# Only show contacts with email addresses
nylas contacts search --has-email

# Combine multiple filters
nylas contacts search --company "Corp" --has-email --limit 20

# Output as JSON
nylas contacts search --json
```

**Example output:**
```bash
$ nylas contacts search --company "Acme" --has-email

ID              Name              Email                   Company         Job Title
---             ----              -----                   -------         ---------
contact_001     Alice Johnson     alice@company.com       Acme Corp       Engineer
contact_002     Eve Martinez      eve@company.com         Acme Corp       Designer

Found 2 contacts
```

**Available filters:**
- `--company` - Filter by company name (partial match)
- `--email` - Filter by email address
- `--phone` - Filter by phone number
- `--source` - Filter by source (address_book, inbox, domain)
- `--group` - Filter by contact group ID
- `--has-email` - Only show contacts with email addresses
- `--limit` - Maximum number of contacts to retrieve (default: 50)
- `--json` - Output as JSON

### Profile Picture Management

Download and manage contact profile pictures.

#### Download Profile Picture

```bash
# Get Base64-encoded profile picture data
nylas contacts photo download <contact-id> [grant-id]

# Save profile picture to file (automatically decodes Base64)
nylas contacts photo download <contact-id> --output photo.jpg

# Get as JSON
nylas contacts photo download <contact-id> --json
```

**Example output:**
```bash
$ nylas contacts photo download contact_abc123 --output alice.jpg

Profile picture saved to: alice.jpg
Size: 15234 bytes
```

**Example output (Base64):**
```bash
$ nylas contacts photo download contact_abc123

Base64-encoded profile picture:
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==

To save to a file, use the --output flag
```

#### Profile Picture Information

```bash
# View information about how profile pictures work in Nylas API v3
nylas contacts photo info
```

**Key points:**
- Profile pictures are retrieved using `?profile_picture=true` query parameter
- API returns Base64-encoded image data
- Images come directly from email provider (Gmail, Outlook, etc.)
- **Upload not supported** - pictures must be managed through provider
- Not all contacts have profile pictures
- Cache pictures locally if using frequently

### Contact Synchronization Info

View information about how contact synchronization works in Nylas API v3.

```bash
# View sync architecture and best practices
nylas contacts sync
```

**Key changes in v3:**
- **No more traditional sync model** - v3 eliminated local data storage
- **Direct provider access** - Requests forwarded to email providers
- **Provider-native IDs** - Contact IDs come from provider
- **Real-time data** - No stale cached data
- **No sync delays** - Instant access to new contacts

**Provider-specific behavior:**
- **Google/Gmail**: Real-time via Google Contacts API (5 min polling)
- **Microsoft/Outlook**: Real-time via Microsoft Graph
- **IMAP**: Depends on provider support
- **Virtual calendars**: Nylas-managed (no provider sync)

**Webhook events for change notifications:**
- `contact.created` - New contact added
- `contact.updated` - Contact modified
- `contact.deleted` - Contact removed

---

## Webhook Management

Create and manage webhooks for real-time event notifications.

### List Webhooks

```bash
nylas webhook list
nylas webhook list --full-ids         # Show full webhook IDs (for copy/paste)
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

### Show Webhook

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

### Create Webhook

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

### Update Webhook

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

### Delete Webhook

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

### List Trigger Types

```bash
nylas webhook triggers
nylas webhook triggers --format json
nylas webhook triggers --format list
nylas webhook triggers --category message   # Filter by category
nylas webhook triggers --category notetaker # Filter by notetaker
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
   • grant.imap_sync_completed

📧 Message
   Email message events

   • message.created
   • message.updated
   • message.opened
   • message.bounce_detected
   • message.send_success
   • message.send_failed
   • message.opened.truncated
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

📝 Notetaker
   Meeting notetaker events

   • notetaker.media

Usage:
  nylas webhook create --url <URL> --triggers message.created
  nylas webhook create --url <URL> --triggers message.created,event.created
```

### Test Webhook

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

## Notetaker Management

Manage Nylas Notetaker bots for meeting recording and transcription.

### List Notetakers

```bash
nylas notetaker list [grant-id]         # List all notetakers
nylas notetaker ls                      # Alias
nylas notetaker list --limit 10         # Limit results
nylas notetaker list --state scheduled  # Filter by state
nylas notetaker list --json             # Output as JSON
```

**Example output:**
```bash
$ nylas notetaker list

Found 3 notetaker(s):

ID: notetaker_abc123
  State:   complete
  Title:   Q4 Planning Meeting
  Link:    https://zoom.us/j/123456789
  Created: 2 hours ago

ID: notetaker_def456
  State:   attending
  Title:   Weekly Standup
  Link:    https://meet.google.com/abc-defg-hij
  Created: 30 minutes ago

ID: notetaker_ghi789
  State:   scheduled
  Title:   Client Demo
  Join:    Mon Dec 23, 2024 2:00 PM
  Created: yesterday
```

### Show Notetaker

```bash
nylas notetaker show <notetaker-id> [grant-id]  # Show details
nylas notetaker show <id> --json                 # Output as JSON
```

**Example output:**
```bash
$ nylas notetaker show notetaker_abc123

Notetaker: notetaker_abc123
State:     complete
Title:     Q4 Planning Meeting
Link:      https://zoom.us/j/123456789
Provider:  zoom
Bot Name:  Meeting Bot

Media:
  Recording: https://storage.nylas.com/recordings/abc123.mp4
    Size: 120.5 MB
  Transcript: https://storage.nylas.com/transcripts/abc123.json
    Size: 50.0 KB

Created: Mon Dec 16, 2024 10:00 AM PST
Updated: Mon Dec 16, 2024 11:30 AM PST
```

### Create Notetaker

```bash
# Create notetaker to join immediately
nylas notetaker create --meeting-link "https://zoom.us/j/123456789"

# Create with scheduled join time
nylas notetaker create --meeting-link "https://meet.google.com/abc-defg-hij" \
  --join-time "2024-12-20 14:00"

# Create with custom bot name
nylas notetaker create --meeting-link "https://zoom.us/j/123" \
  --bot-name "Meeting Recorder"

# Join in 30 minutes
nylas notetaker create --meeting-link "https://zoom.us/j/123" --join-time "30m"
```

**Supported meeting providers:**
- Zoom
- Google Meet
- Microsoft Teams

**Example output:**
```bash
$ nylas notetaker create --meeting-link "https://zoom.us/j/123456789" --bot-name "My Bot"

✓ Notetaker created successfully!

ID:    notetaker_new_123
State: scheduled
Link:  https://zoom.us/j/123456789
```

### Delete Notetaker

```bash
nylas notetaker delete <notetaker-id> [grant-id]  # Delete with confirmation
nylas notetaker delete <id> --yes                  # Skip confirmation
nylas notetaker rm <id>                            # Alias
nylas notetaker cancel <id>                        # Alias
```

**Example output:**
```bash
$ nylas notetaker delete notetaker_abc123

Delete notetaker notetaker_abc123?
  Title: Q4 Planning Meeting
  State: scheduled

This action cannot be undone. Continue? [y/N]: y
✓ Notetaker notetaker_abc123 deleted successfully
```

### Get Notetaker Media

Retrieve recording and transcript URLs from a completed notetaker session.

```bash
nylas notetaker media <notetaker-id> [grant-id]  # Get media URLs
nylas notetaker media <id> --json                 # Output as JSON
```

**Example output:**
```bash
$ nylas notetaker media notetaker_abc123

Notetaker Media:

Recording:
  URL:  https://storage.nylas.com/recordings/abc123.mp4
  Type: video/mp4
  Size: 120.5 MB
  Expires: Mon Dec 23, 2024 10:00 AM PST

Transcript:
  URL:  https://storage.nylas.com/transcripts/abc123.json
  Type: application/json
  Size: 50.0 KB
  Expires: Mon Dec 23, 2024 10:00 AM PST
```

**Note:** Media URLs have an expiration time. Download them promptly after retrieval.

---

## Inbound Email Management

Manage Nylas Inbound email inboxes for receiving emails at managed addresses without OAuth flows.

### What is Nylas Inbound?

Nylas Inbound enables your application to receive emails at dedicated managed addresses (e.g., `support@yourapp.nylas.email`) and process them via webhooks. It's designed for:

- Capturing messages sent to specific addresses (intake@, leads@, tickets@)
- Triggering automated workflows from incoming mail
- Real-time message delivery to workers, LLMs, or downstream systems

### List Inbound Inboxes

```bash
nylas inbound list                    # List all inbound inboxes
nylas inbox list                      # Alias
nylas inbound list --json             # Output as JSON
```

**Example output:**
```bash
$ nylas inbound list

Inbound Inboxes (3)

1. support@yourapp.nylas.email  30 days ago  active
   ID: inbox_abc123

2. sales@yourapp.nylas.email  14 days ago  active
   ID: inbox_def456

3. info@yourapp.nylas.email  7 days ago  active
   ID: inbox_ghi789

Use 'nylas inbound messages <inbox-id>' to view messages
```

### Show Inbox Details

```bash
nylas inbound show <inbox-id>         # Show inbox details
nylas inbound show <inbox-id> --json  # Output as JSON

# Use environment variable for inbox ID
export NYLAS_INBOUND_GRANT_ID=inbox_abc123
nylas inbound show
```

**Example output:**
```bash
$ nylas inbound show inbox_abc123

────────────────────────────────────────────────────────────
Inbox: support@yourapp.nylas.email
────────────────────────────────────────────────────────────
ID:          inbox_abc123
Email:       support@yourapp.nylas.email
Status:      active
Created:     Dec 1, 2024 10:00 AM (30 days ago)
Updated:     Dec 16, 2024 2:30 PM (1 hour ago)
```

### Create Inbound Inbox

```bash
# Create a new inbound inbox
nylas inbound create <email-prefix>

# Examples
nylas inbound create support          # Creates: support@yourapp.nylas.email
nylas inbound create leads            # Creates: leads@yourapp.nylas.email
nylas inbound create tickets --json   # Output as JSON
```

**Example output:**
```bash
$ nylas inbound create support

Inbound inbox created successfully!

────────────────────────────────────────────────────────────
Inbox: support@yourapp.nylas.email
────────────────────────────────────────────────────────────
ID:          inbox_new_123
Email:       support@yourapp.nylas.email
Status:      active
Created:     Dec 16, 2024 3:00 PM (just now)

Next steps:
  1. Set up a webhook: nylas webhooks create --url <your-url> --triggers message.created
  2. View messages: nylas inbound messages inbox_new_123
  3. Monitor in real-time: nylas inbound monitor inbox_new_123
```

### View Inbound Messages

```bash
nylas inbound messages <inbox-id>           # List messages
nylas inbound messages <inbox-id> --limit 5 # Limit results
nylas inbound messages <inbox-id> --unread  # Show only unread
nylas inbound messages <inbox-id> --json    # Output as JSON

# Use environment variable
export NYLAS_INBOUND_GRANT_ID=inbox_abc123
nylas inbound messages
```

**Example output:**
```bash
$ nylas inbound messages inbox_abc123

Messages (5 total, 2 unread)

● ★ John Smith           New Lead: Enterprise Plan Inquiry      10 minutes ago
      ID: msg_001

●   Sarah Johnson        Support Request: Integration Help      1 hour ago
      ID: msg_002

  ★ Mike Chen            Partnership Opportunity                 3 hours ago
      ID: msg_003

    Lisa Park            Billing Question                        yesterday
      ID: msg_004

    Alex Rivera          Feature Request: Dark Mode              2 days ago
      ID: msg_005

Use 'nylas email read <inbox-id> <message-id>' to view full message
```

### Monitor Inbound Messages (Real-time)

Monitor for new incoming emails in real-time using webhooks.

```bash
nylas inbound monitor <inbox-id>              # Start monitoring
nylas inbound monitor <inbox-id> --tunnel cloudflared  # With public tunnel
nylas inbound monitor <inbox-id> --port 8080  # Custom port
nylas inbound monitor <inbox-id> --json       # Output events as JSON
nylas inbound monitor <inbox-id> --quiet      # Only show events

# Use environment variable
export NYLAS_INBOUND_GRANT_ID=inbox_abc123
nylas inbound monitor --tunnel cloudflared
```

**Example output:**
```bash
$ nylas inbound monitor inbox_abc123 --tunnel cloudflared

╔══════════════════════════════════════════════════════════════╗
║            Nylas Inbound Monitor                             ║
╚══════════════════════════════════════════════════════════════╝

Monitoring: support@yourapp.nylas.email

Monitor started successfully!

  Local URL:    http://localhost:3000/webhook
  Public URL:   https://abc123.trycloudflare.com/webhook

  Tunnel:       cloudflared (connected)

To receive events, register this webhook URL with Nylas:
  nylas webhooks create --url https://abc123.trycloudflare.com/webhook --triggers message.created

Press Ctrl+C to stop

─────────────────────────────────────────────────────────────────
Incoming Messages:

[14:32:15] NEW MESSAGE [verified]
  Subject: New Lead: Enterprise Plan Inquiry
  From: John Smith <john@bigcorp.com>
  Preview: Hi, I'm interested in learning more about your enterprise plan...
  ID: msg_new_001

[14:35:42] NEW MESSAGE [verified]
  Subject: Support Request
  From: Sarah <sarah@startup.io>
  Preview: We're having trouble connecting our calendar integration...
  ID: msg_new_002
```

### Delete Inbound Inbox

```bash
nylas inbound delete <inbox-id>        # Delete with confirmation
nylas inbound delete <inbox-id> --yes  # Skip confirmation
nylas inbound delete <inbox-id> -f     # Force delete (alias for --yes)
```

**Example output:**
```bash
$ nylas inbound delete inbox_abc123

You are about to delete the inbound inbox:
  Email: support@yourapp.nylas.email
  ID:    inbox_abc123

This action cannot be undone. All messages in this inbox will be deleted.

Type 'delete' to confirm: delete
Inbox support@yourapp.nylas.email deleted successfully!
```

### Environment Variables

You can use environment variables to avoid passing the inbox ID repeatedly:

```bash
# Set the inbound grant ID
export NYLAS_INBOUND_GRANT_ID=inbox_abc123

# Now commands will use this ID by default
nylas inbound show
nylas inbound messages
nylas inbound messages --unread
nylas inbound monitor --tunnel cloudflared
```

---

## Scheduler Management

Manage Nylas Scheduler for creating booking pages, configurations, sessions, and appointments.

### What is Nylas Scheduler?

Nylas Scheduler enables you to create customizable booking workflows for scheduling meetings. Key features include:
- **Configurations**: Define meeting types with availability rules and settings
- **Sessions**: Generate temporary booking sessions for specific configurations
- **Bookings**: Manage scheduled appointments (view, confirm, reschedule, cancel)
- **Pages**: Create and manage hosted scheduling pages

### Scheduler Configurations

Manage scheduling configurations (meeting types):

```bash
# List all scheduler configurations
nylas scheduler configurations list
nylas scheduler configs list              # Alias
nylas scheduler configurations list --json

# Show configuration details
nylas scheduler configurations show <config-id>
nylas scheduler configs show <config-id>

# Create a new configuration
nylas scheduler configurations create \\
  --name "30 Min Meeting" \\
  --duration 30 \\
  --interval 15

# Update a configuration
nylas scheduler configurations update <config-id> \\
  --name "Updated Name" \\
  --duration 60

# Delete a configuration
nylas scheduler configurations delete <config-id>
nylas scheduler configs delete <config-id> -f    # Skip confirmation
```

**Configuration Features:**
- Duration and interval settings
- Availability rules and windows
- Buffer times before/after meetings
- Booking limits and restrictions
- Custom event settings

### Scheduler Sessions

Create temporary booking sessions for configurations:

```bash
# Create a session for a configuration
nylas scheduler sessions create <config-id>

# Show session details
nylas scheduler sessions show <session-id>
```

**Session Features:**
- Temporary booking URLs with expiration
- Configuration-specific availability
- Session-based booking tracking

### Scheduler Bookings

Manage scheduled appointments:

```bash
# List all bookings
nylas scheduler bookings list
nylas scheduler bookings list --json

# Show booking details
nylas scheduler bookings show <booking-id>

# Confirm a booking
nylas scheduler bookings confirm <booking-id>

# Reschedule a booking
nylas scheduler bookings reschedule <booking-id> \\
  --start-time "2024-03-20T10:00:00Z"

# Cancel a booking
nylas scheduler bookings cancel <booking-id>
nylas scheduler bookings cancel <booking-id> \\
  --reason "Meeting no longer needed"
```

**Booking Information Includes:**
- Event ID and configuration ID
- Start and end times
- Participant details
- Status (pending, confirmed, cancelled)

### Scheduler Pages

Create and manage hosted booking pages:

```bash
# List all scheduler pages
nylas scheduler pages list
nylas scheduler pages list --json

# Show page details
nylas scheduler pages show <page-id>

# Create a new page
nylas scheduler pages create \\
  --config-id <config-id> \\
  --slug "meet-me"

# Update a page
nylas scheduler pages update <page-id> \\
  --slug "new-slug" \\
  --name "Updated Page"

# Delete a page
nylas scheduler pages delete <page-id>
```

**Page Features:**
- Custom slugs for friendly URLs
- Configuration-based availability
- Optional custom domain support
- Appearance customization

**Example Workflow:**

```bash
# 1. Create a meeting type
nylas scheduler configs create \\
  --name "Product Demo" \\
  --duration 30

# 2. Create a booking page
nylas scheduler pages create \\
  --config-id <config-id> \\
  --slug "product-demo"

# 3. Share the booking URL with prospects
# URL format: https://schedule.nylas.com/product-demo

# 4. View bookings
nylas scheduler bookings list

# 5. Manage bookings
nylas scheduler bookings confirm <booking-id>
nylas scheduler bookings reschedule <booking-id> --start-time "..."
```

**Note:** Some scheduler features may not be available in all Nylas API versions or require specific subscription tiers.

---

## Admin Commands

Administrative commands for managing Nylas platform resources at an organizational level. These commands require API key authentication.

### Applications

Manage Nylas applications in your organization.

```bash
# List applications
nylas admin applications list
nylas admin apps list              # Alias
nylas admin app list --json        # Output as JSON

# Show application details
nylas admin applications show <app-id>
nylas admin app show <app-id> --json

# Create application
nylas admin applications create --name "My App" --region us
nylas admin app create --name "My App" --region eu \
  --branding-name "MyApp" \
  --website-url "https://myapp.com" \
  --callback-uris "https://myapp.com/oauth/callback,https://myapp.com/oauth/redirect"

# Update application
nylas admin applications update <app-id> --name "Updated Name"
nylas admin app update <app-id> --branding-name "NewBrand" --website-url "https://new.com"

# Delete application
nylas admin applications delete <app-id>
nylas admin app delete <app-id> --yes  # Skip confirmation
```

**Example: List applications**
```bash
$ nylas admin applications list

Found 2 application(s):

APP ID              REGION    ENVIRONMENT
myapp-prod-123      us        production
myapp-dev-456       us        development
```

**Example: Show application details**
```bash
$ nylas admin applications show myapp-prod-123

Application Details
  ID: app_abc123
  Application ID: myapp-prod-123
  Organization ID: org_xyz789
  Region: us
  Environment: production

Branding:
  Name: MyApp
  Website: https://myapp.com

Callback URIs (2):
  1. https://myapp.com/oauth/callback
  2. https://myapp.com/oauth/redirect
```

### Connectors

Manage email provider connectors (Google, Microsoft, IMAP, etc.).

```bash
# List connectors
nylas admin connectors list
nylas admin conn list              # Alias
nylas admin connectors list --json

# Show connector details
nylas admin connectors show <connector-id>
nylas admin conn show <connector-id> --json

# Create OAuth connector (Google/Microsoft)
nylas admin connectors create --name "Gmail" --provider google \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "GOCSPX-xxx" \
  --scopes "https://www.googleapis.com/auth/gmail.readonly,https://www.googleapis.com/auth/calendar"

# Create IMAP connector
nylas admin connectors create --name "Custom IMAP" --provider imap \
  --imap-host "imap.example.com" --imap-port 993 \
  --smtp-host "smtp.example.com" --smtp-port 587

# Update connector
nylas admin connectors update <connector-id> --name "Updated Name"
nylas admin conn update <connector-id> --scopes "new,scopes,list"

# Delete connector
nylas admin connectors delete <connector-id>
nylas admin conn delete <connector-id> --yes
```

**Example: List connectors**
```bash
$ nylas admin connectors list

Found 3 connector(s):

NAME                ID                    PROVIDER      SCOPES
Gmail               conn_google_123       google        3
Microsoft 365       conn_ms365_456        microsoft     4
Custom IMAP         conn_imap_789         imap          0
```

**Example: Show connector details**
```bash
$ nylas admin connectors show conn_google_123

Connector: Gmail
  ID: conn_google_123
  Provider: google

Scopes (3):
  1. https://www.googleapis.com/auth/gmail.readonly
  2. https://www.googleapis.com/auth/calendar
  3. https://www.googleapis.com/auth/contacts.readonly

Settings:
  Client ID: 123456789.apps.googleusercontent.com
```

### Credentials

Manage authentication credentials for connectors.

```bash
# List credentials
nylas admin credentials list
nylas admin creds list             # Alias
nylas admin credentials list --json

# Show credential details
nylas admin credentials show <credential-id>
nylas admin cred show <credential-id> --json

# Create credential
nylas admin credentials create --connector-id <connector-id> \
  --name "Production Credentials" \
  --credential-type oauth

# Create credential with data
nylas admin cred create --connector-id <connector-id> \
  --name "Service Account" \
  --credential-type service_account \
  --credential-data '{"private_key":"..."}'

# Update credential
nylas admin credentials update <credential-id> --name "Updated Name"

# Delete credential
nylas admin credentials delete <credential-id>
nylas admin cred delete <credential-id> --yes
```

**Example: List credentials**
```bash
$ nylas admin credentials list

Found 2 credential(s):

NAME                    ID                    CONNECTOR          TYPE
Production OAuth        cred_oauth_123        conn_google_123    oauth
Service Account         cred_sa_456           conn_google_123    service_account
```

**Example: Show credential details**
```bash
$ nylas admin credentials show cred_oauth_123

Credential: Production OAuth
  ID: cred_oauth_123
  Connector ID: conn_google_123
  Name: Production OAuth
  Type: oauth

Created: Dec 1, 2024 10:00 AM
Updated: Dec 15, 2024 2:30 PM
```

### Grants

View and manage grants across all applications.

```bash
# List grants
nylas admin grants list
nylas admin grant list             # Alias
nylas admin grants list --limit 100 --offset 0
nylas admin grants list --connector-id <connector-id>
nylas admin grants list --status valid
nylas admin grants list --status invalid
nylas admin grants list --json

# Show grant statistics
nylas admin grants stats
nylas admin grants stats --json
```

**Example: List grants**
```bash
$ nylas admin grants list --limit 5

Found 5 grant(s):

EMAIL                   ID                    PROVIDER    STATUS
user@gmail.com          grant_abc123          google      valid
work@company.com        grant_def456          microsoft   valid
john@example.com        grant_ghi789          google      invalid
-                       grant_jkl012          imap        valid
alice@startup.io        grant_mno345          google      valid
```

**Example: Grant statistics**
```bash
$ nylas admin grants stats

Grant Statistics
  Total Grants: 150
  Valid: 142
  Invalid: 8

By Provider:
PROVIDER          COUNT
google            95
microsoft         42
imap              13

By Status:
STATUS            COUNT
valid             142
invalid           8
```

**Filter options:**
- `--limit` - Maximum number of grants to return (default: 50)
- `--offset` - Offset for pagination (default: 0)
- `--connector-id` - Filter by connector ID
- `--status` - Filter by status (valid, invalid)
- `--json` - Output as JSON

---

## OTP Management

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

## Time Zone Utilities

Offline time zone conversion and meeting scheduling tools. Works 100% offline—no API access required.

> **📚 For comprehensive timezone documentation, see the [Timezone Guide](TIMEZONE.md).**

> **💡 Pro Tip:** All timezone commands work instantly without network calls. Perfect for remote teams, travel planning, and global coordination.

### Quick Examples

```bash
# Convert current time between zones
nylas timezone convert --from PST --to IST

# Check DST transitions for planning
nylas timezone dst --zone America/New_York --year 2026

# Find meeting times across multiple zones
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"

# List all available time zones
nylas timezone list --filter America

# Get detailed zone information
nylas timezone info UTC
```

---

### Convert Time Between Zones

Convert time from one timezone to another with automatic DST handling.

```bash
nylas timezone convert --from <zone> --to <zone>           # Convert current time
nylas timezone convert --from <zone> --to <zone> --time <RFC3339>  # Convert specific time
nylas timezone convert --from <zone> --to <zone> --json    # JSON output
```

**Flags:**
- `--from` (required) - Source time zone (IANA name or abbreviation)
- `--to` (required) - Target time zone (IANA name or abbreviation)
- `--time` - Specific time to convert (RFC3339 format: 2025-01-01T12:00:00Z)
- `--json` - Output as JSON

**Supported Abbreviations:**
- PST/PDT → America/Los_Angeles
- EST/EDT → America/New_York
- CST/CDT → America/Chicago
- MST/MDT → America/Denver
- GMT/BST → Europe/London
- IST → Asia/Kolkata
- JST → Asia/Tokyo
- AEST/AEDT → Australia/Sydney

**Example: Convert current time**
```bash
$ nylas timezone convert --from PST --to IST

Time Zone Conversion

From: America/Los_Angeles (PST)
  Time:   2025-12-20 18:00:00
  Offset: UTC-8
  DST:    No (Standard Time)

To: Asia/Kolkata (IST)
  Time:   2025-12-21 07:30:00
  Offset: UTC+5:30
  DST:    No (Standard Time)

Time Difference: Asia/Kolkata is 13 hour(s) ahead of America/Los_Angeles
```

**Example: Convert specific time**
```bash
$ nylas timezone convert --from UTC --to America/New_York --time "2025-01-01T12:00:00Z"

Time Zone Conversion

From: UTC (UTC)
  Time:   2025-01-01 12:00:00
  Offset: UTC+0
  DST:    No (Standard Time)

To: America/New_York (EST)
  Time:   2025-01-01 07:00:00
  Offset: UTC-5
  DST:    No (Standard Time)

Time Difference: America/New_York is 5 hour(s) behind UTC
```

**Example: Using abbreviations**
```bash
$ nylas timezone convert --from PST --to EST

Time Zone Conversion

From: America/Los_Angeles (PST)
  Time:   2025-12-20 18:00:00
  Offset: UTC-8

To: America/New_York (EST)
  Time:   2025-12-20 21:00:00
  Offset: UTC-5

Time Difference: America/New_York is 3 hour(s) ahead of America/Los_Angeles
```

**Example: JSON output for scripting**
```bash
$ nylas timezone convert --from UTC --to EST --json
{
  "from": {
    "zone": "UTC",
    "time": "2025-12-21T02:00:00Z",
    "abbr": "UTC",
    "offset": "UTC+0",
    "is_dst": false
  },
  "to": {
    "zone": "America/New_York",
    "time": "2025-12-20T21:00:00-05:00",
    "abbr": "EST",
    "offset": "UTC-5",
    "is_dst": false
  }
}
```

---

### Find Meeting Times Across Zones

Find overlapping working hours across multiple time zones for scheduling meetings.

```bash
nylas timezone find-meeting --zones <zones>                # Basic meeting finder
nylas timezone find-meeting --zones <zones> --duration <duration>  # Specify duration
nylas timezone find-meeting --zones <zones> --start-hour <HH:MM> --end-hour <HH:MM>  # Custom hours
nylas timezone find-meeting --zones <zones> --exclude-weekends  # Skip weekends
```

**Flags:**
- `--zones` (required) - Comma-separated list of time zones
- `--duration` - Meeting duration (default: 1h). Format: 30m, 1h, 1h30m
- `--start-hour` - Working hours start (default: 09:00). Format: HH:MM
- `--end-hour` - Working hours end (default: 17:00). Format: HH:MM
- `--start-date` - Search start date (default: today). Format: YYYY-MM-DD
- `--end-date` - Search end date (default: 7 days from start). Format: YYYY-MM-DD
- `--exclude-weekends` - Skip Saturday and Sunday
- `--json` - Output as JSON

**Example: Basic meeting finder**
```bash
$ nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"

Meeting Time Finder

Time Zones: America/New_York,Europe/London,Asia/Tokyo
Duration: 1h
Working Hours: 09:00 - 17:00
Date Range: 2025-12-21 to 2025-12-28

⚠️  NOTE: Meeting time finder logic is not yet fully implemented.
          The service will return available slots once the algorithm is complete.

Planned features:
  • Identify overlapping working hours across all zones
  • Calculate quality scores (middle of day = higher score)
  • Filter by meeting duration
  • Respect weekend exclusions
```

**Example: Custom working hours**
```bash
$ nylas timezone find-meeting \
  --zones "PST,EST,IST" \
  --duration 30m \
  --start-hour 10:00 \
  --end-hour 16:00 \
  --exclude-weekends

Meeting Time Finder

Time Zones: PST,EST,IST
Duration: 30m
Working Hours: 10:00 - 16:00
Date Range: 2025-12-21 to 2025-12-28
Excluding: Weekends
```

**Example: Specific date range**
```bash
$ nylas timezone find-meeting \
  --zones "America/Los_Angeles,Europe/Paris" \
  --duration 1h \
  --start-date 2026-01-15 \
  --end-date 2026-01-22
```

> **Note:** The meeting finder algorithm is planned but not yet implemented. The CLI and service interfaces are complete and ready for the algorithm implementation.

---

### Check DST Transitions

Display Daylight Saving Time transitions for a specific time zone and year.

```bash
nylas timezone dst --zone <zone>                # Check current year
nylas timezone dst --zone <zone> --year <year>  # Check specific year
nylas timezone dst --zone <zone> --json         # JSON output
```

**Flags:**
- `--zone` (required) - Time zone to check (IANA name or abbreviation)
- `--year` - Year to check (default: current year)
- `--json` - Output as JSON

**Example: Zone with DST**
```bash
$ nylas timezone dst --zone America/New_York --year 2026

DST Transitions for America/New_York in 2026

Found 2 transition(s):

Date          Time      Direction       Name  Offset
────────────────────────────────────────────────────────
⏰ 2026-03-08  02:00:00  Spring Forward  EDT   UTC-4
🕐 2026-11-01  02:00:00  Fall Back       EST   UTC-5

Legend:
  ⏰ Spring Forward: Clocks move ahead (lose 1 hour)
  🕐 Fall Back: Clocks move back (gain 1 hour)

⚠️  WARNING: DST transition in 77 days (March 8)
   Be mindful when scheduling meetings around this date.
```

**Example: Zone without DST**
```bash
$ nylas timezone dst --zone America/Phoenix --year 2026

DST Transitions for America/Phoenix in 2026

❌ No DST transitions found

This time zone likely does not observe Daylight Saving Time.
It stays on standard time throughout the year.

Examples of non-DST zones:
  • America/Phoenix (Arizona)
  • Pacific/Honolulu (Hawaii)
  • Asia/Tokyo (Japan)
  • Asia/Kolkata (India)
```

**Example: Using abbreviation**
```bash
$ nylas timezone dst --zone PST

DST Transitions for America/Los_Angeles in 2025

Found 2 transition(s):

Date          Time      Direction       Name  Offset
────────────────────────────────────────────────────────
⏰ 2025-03-09  02:00:00  Spring Forward  PDT   UTC-7
🕐 2025-11-02  02:00:00  Fall Back       PST   UTC-8
```

**Example: JSON output**
```bash
$ nylas timezone dst --zone EST --json
{
  "zone": "America/New_York",
  "year": 2025,
  "transitions": [
    {
      "date": "2025-03-09T07:00:00Z",
      "direction": "forward",
      "name": "EDT",
      "offset": -14400
    },
    {
      "date": "2025-11-02T06:00:00Z",
      "direction": "backward",
      "name": "EST",
      "offset": -18000
    }
  ],
  "count": 2
}
```

---

### List Available Time Zones

Display all IANA time zones with current time and offset information.

```bash
nylas timezone list                    # List all zones
nylas timezone list --filter <text>    # Filter by name
nylas timezone list --json             # JSON output
```

**Flags:**
- `--filter` - Filter zones by name (case-insensitive)
- `--json` - Output as JSON

**Example: List all zones**
```bash
$ nylas timezone list

IANA Time Zones

═══ Africa (58) ═══
  • Africa/Abidjan                           UTC+0   02:44 GMT
  • Africa/Accra                             UTC+0   02:44 GMT
  • Africa/Cairo                             UTC+2   04:44 EET
  ...

═══ America (142) ═══
  • America/New_York                         UTC-5   21:44 EST
  • America/Chicago                          UTC-6   20:44 CST
  • America/Denver                           UTC-7   19:44 MST
  • America/Los_Angeles                      UTC-8   18:44 PST
  • America/Phoenix                          UTC-7   19:44 MST
  ...

═══ Asia (88) ═══
  • Asia/Tokyo                               UTC+9   11:44 JST
  • Asia/Kolkata                             UTC+5:30 08:14 IST
  • Asia/Shanghai                            UTC+8   10:44 CST
  ...

═══ Europe (62) ═══
  • Europe/London                            UTC+0   02:44 GMT
  • Europe/Paris                             UTC+1   03:44 CET
  • Europe/Berlin                            UTC+1   03:44 CET
  ...

Total: 593 time zone(s)
```

**Example: Filter by region**
```bash
$ nylas timezone list --filter America

IANA Time Zones (filtered by 'America')

═══ America (142) ═══
  • America/New_York                         UTC-5   21:44 EST
  • America/Chicago                          UTC-6   20:44 CST
  • America/Denver                           UTC-7   19:44 MST
  • America/Los_Angeles                      UTC-8   18:44 PST
  • America/Phoenix                          UTC-7   19:44 MST
  • America/Anchorage                        UTC-9   17:44 AKST
  • America/Halifax                          UTC-4   22:44 AST
  • America/Sao_Paulo                        UTC-3   23:44 -03
  • America/Mexico_City                      UTC-6   20:44 CST
  ...

Total: 142 time zone(s)
```

**Example: Filter by city**
```bash
$ nylas timezone list --filter Tokyo

IANA Time Zones (filtered by 'Tokyo')

═══ Asia (1) ═══
  • Asia/Tokyo                               UTC+9   11:44 JST

Total: 1 time zone(s)
```

**Example: JSON output**
```bash
$ nylas timezone list --filter UTC --json
{
  "zones": [
    "UTC",
    "Etc/UTC"
  ],
  "count": 2
}
```

**Example: No results**
```bash
$ nylas timezone list --filter NonExistent

IANA Time Zones (filtered by 'NonExistent')

No time zones found matching the filter.
```

---

### Get Time Zone Information

Display detailed information about a specific time zone.

```bash
nylas timezone info <zone>                     # Get info for zone
nylas timezone info --zone <zone>              # Alternative syntax
nylas timezone info --zone <zone> --time <RFC3339>  # Info at specific time
nylas timezone info --zone <zone> --json       # JSON output
```

**Flags:**
- `--zone` - Time zone to query (IANA name or abbreviation)
- `--time` - Check info at specific time (RFC3339 format)
- `--json` - Output as JSON

> **Note:** Zone can be provided as a positional argument or via `--zone` flag.

**Example: Get zone information**
```bash
$ nylas timezone info America/New_York

Time Zone Information

Zone: America/New_York
Abbreviation: EST
Current Time: 2025-12-20 21:44:03 (EST)
UTC Offset: UTC-5 (-18000 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-08 02:00:00 EST
  Days Until: 77
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 5 hour(s) behind UTC
```

**Example: Using abbreviation**
```bash
$ nylas timezone info PST

Time Zone Information

Zone: America/Los_Angeles (expanded from 'PST')
Abbreviation: PST
Current Time: 2025-12-20 18:44:03 (PST)
UTC Offset: UTC-8 (-28800 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-09 02:00:00 PST
  Days Until: 78
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 8 hour(s) behind UTC
```

**Example: Zone without DST**
```bash
$ nylas timezone info Asia/Tokyo

Time Zone Information

Zone: Asia/Tokyo
Abbreviation: JST
Current Time: 2025-12-21 11:44:03 (JST)
UTC Offset: UTC+9 (32400 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition: None found in next 365 days
  (This zone may not observe DST)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 9 hour(s) ahead of UTC
```

**Example: Check at specific time**
```bash
$ nylas timezone info --zone America/New_York --time "2026-07-01T12:00:00Z"

Time Zone Information

Zone: America/New_York
Abbreviation: EDT
Current Time: 2026-07-01 08:00:00 (EDT)
UTC Offset: UTC-4 (-14400 seconds)
DST Status: ✓ Currently observing Daylight Saving Time

Next DST Transition:
  Date: 2026-11-01 02:00:00 EDT
  Days Until: 123
  Change: Fall Back (DST ends, gain 1 hour)

UTC Comparison:
  UTC Time: 2026-07-01 12:00:00 (UTC)
  Difference: 4 hour(s) behind UTC
```

**Example: JSON output**
```bash
$ nylas timezone info UTC --json
{
  "zone": "UTC",
  "abbreviation": "UTC",
  "offset": "UTC+0",
  "offset_seconds": 0,
  "is_dst": false,
  "local_time": "2025-12-21T02:44:03Z",
  "next_dst": null
}
```

**Example: Using flag instead of positional arg**
```bash
$ nylas timezone info --zone Europe/London

Time Zone Information

Zone: Europe/London
Abbreviation: GMT
Current Time: 2025-12-21 02:44:03 (GMT)
UTC Offset: UTC+0 (0 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-29 01:00:00 GMT
  Days Until: 97
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: Same as UTC
```

---

### Tips & Tricks

**Use Abbreviations for Speed**
```bash
# Instead of full IANA names:
nylas timezone convert --from America/Los_Angeles --to Asia/Kolkata

# Use common abbreviations:
nylas timezone convert --from PST --to IST
```

**JSON Output for Scripting**
```bash
# Parse with jq
nylas timezone info UTC --json | jq '.offset_seconds'
# Output: 0

# Get all America zones
nylas timezone list --filter America --json | jq '.zones[]'
```

**Check Multiple Zones Quickly**
```bash
# Loop through zones
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  echo "=== $zone ==="
  nylas timezone info $zone | grep "Current Time"
done
```

**DST Planning for Meetings**
```bash
# Check if DST change affects your meeting
nylas timezone dst --zone America/New_York --year 2026

# Plan around the transition dates
```

**Combine with Other Commands**
```bash
# Get current time in client's timezone before calling
CLIENT_ZONE="Europe/London"
nylas timezone info $CLIENT_ZONE | grep "Current Time"

# Then make your call
```

**Save Common Conversions as Aliases**
```bash
# Add to ~/.bashrc or ~/.zshrc
alias pst2ist='nylas timezone convert --from PST --to IST'
alias est2pst='nylas timezone convert --from EST --to PST'
alias utc2local='nylas timezone convert --from UTC --to $(date +%Z)'
```

**Offline Usage**
```bash
# Works anywhere - plane, train, no WiFi needed
# All calculations are local, instant, and private
nylas timezone convert --from PST --to EST
```

---

### Common Use Cases

**1. Remote Team Standups**
```bash
# "What time is 9 AM PST for my team in India?"
nylas timezone convert --from PST --to IST --time "2025-12-21T09:00:00-08:00"
```

**2. Client Calls**
```bash
# "Is it business hours in London right now?"
nylas timezone info Europe/London
```

**3. Travel Planning**
```bash
# "When does my flight land in local time?"
nylas timezone convert --from UTC --to America/Los_Angeles --time "2025-12-25T14:30:00Z"
```

**4. Meeting Scheduling**
```bash
# "Find time that works for NYC, London, and Tokyo"
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```

**5. DST Change Awareness**
```bash
# "Will DST affect my recurring meeting in March?"
nylas timezone dst --zone America/New_York --year 2026
```

**6. Multi-Region Deployments**
```bash
# "What time is 2 AM UTC in all our datacenter regions?"
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  nylas timezone convert --from UTC --to $zone --time "2025-12-21T02:00:00Z"
done
```

---

### Troubleshooting

**Invalid Time Zone Error**
```bash
$ nylas timezone info Invalid/Zone
Error: get time zone info: unknown time zone Invalid/Zone

# Use list to find valid zones:
nylas timezone list --filter <search>
```

**Invalid Time Format**
```bash
$ nylas timezone convert --from UTC --to EST --time "invalid"
Error: invalid time format (use RFC3339, e.g., 2025-01-01T12:00:00Z)

# Use RFC3339 format:
# YYYY-MM-DDTHH:MM:SSZ (UTC)
# YYYY-MM-DDTHH:MM:SS±HH:MM (with offset)
```

**Missing Required Flag**
```bash
$ nylas timezone convert --from PST
Error: required flag(s) "to" not set

# Both --from and --to are required
nylas timezone convert --from PST --to EST
```

**Abbreviation Not Recognized**
```bash
# If abbreviation isn't in the built-in list, use full IANA name
nylas timezone list --filter <region>
# Then use the full name from the list
```

---

### Performance Notes

- **Instant execution** - All operations are local calculations
- **No network calls** - Works 100% offline
- **No rate limits** - Use as frequently as needed
- **Privacy-first** - No data ever sent to external servers
- **Minimal resources** - Uses OS timezone database

---

### Related Commands

- `nylas auth detect` - Detect email provider timezone
- `nylas calendar list` - View events (which may have timezone info)
- `nylas tui` - Interactive terminal interface

---

## Terminal User Interface (TUI)

See **[TUI Documentation](TUI.md)** for complete TUI reference including themes, keyboard shortcuts, and customization.

```bash
nylas tui                    # Launch TUI at dashboard
nylas tui --demo             # Demo mode (no credentials needed)
nylas tui --theme amber      # Retro amber CRT theme
```

---

## Diagnostic Commands

### Doctor

Run diagnostic checks to verify your Nylas CLI setup.

```bash
nylas doctor            # Run all diagnostic checks
nylas doctor --verbose  # Show detailed information
```

---

## Global Flags

Available on all commands:

```
--json       Output in JSON format
--no-color   Disable color output
-v, --verbose Enable verbose output
--config     Custom config file path
```
