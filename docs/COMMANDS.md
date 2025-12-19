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
