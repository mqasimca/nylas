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
