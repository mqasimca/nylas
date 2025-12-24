# Nylas CLI Command Reference

Quick command reference. For detailed docs, see `docs/commands/<feature>.md`

> **Quick Links:** [README](../README.md) | [Development](DEVELOPMENT.md) | [Architecture](ARCHITECTURE.md)

---

## Global Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--json` | Output as JSON | `nylas email list --json` |
| `--limit N` | Limit results | `nylas email list --limit 10` |
| `--yes` / `-y` | Skip confirmations | `nylas email delete ID --yes` |
| `--help` / `-h` | Show help | `nylas email --help` |

---

## Authentication

```bash
nylas auth config                # Configure API credentials
nylas auth login                 # Authenticate with provider
nylas auth list                  # List connected accounts
nylas auth logout <grant-id>     # Disconnect account
```

---

## Email

```bash
nylas email list [grant-id]                                    # List emails
nylas email read <message-id>                                  # Read email
nylas email send --to EMAIL --subject SUBJECT --body BODY      # Send email
nylas email search --query "QUERY"                             # Search emails
nylas email delete <message-id>                                # Delete email
```

**Filters:** `--unread`, `--starred`, `--from`, `--to`, `--subject`

**AI features:**
```bash
nylas email ai analyze                    # AI-powered inbox summary
nylas email ai analyze --limit 25         # Analyze more emails
nylas email ai analyze --unread           # Only unread emails
nylas email ai analyze --provider claude  # Use specific AI provider
```

**Details:** `docs/commands/email.md`, `docs/AI.md`

---

## Folders & Threads

```bash
nylas folders list               # List folders
nylas folders create --name NAME # Create folder
nylas threads list               # List threads
nylas threads show <thread-id>   # Show thread
```

---

## Drafts

```bash
nylas drafts list                                 # List drafts
nylas drafts create --to EMAIL --subject SUBJECT  # Create draft
nylas drafts send <draft-id>                      # Send draft
nylas drafts delete <draft-id>                    # Delete draft
```

---

## Calendar

```bash
nylas calendar list                                              # List calendars
nylas calendar events list [--days N] [--timezone ZONE]          # List events
nylas calendar events create --title T --start TIME --end TIME   # Create event
nylas calendar events delete <event-id>                          # Delete event
```

**Timezone features:**
```bash
nylas calendar events list --timezone America/Los_Angeles --show-tz
```

**AI scheduling:**
```bash
nylas calendar schedule ai "meeting with John next Tuesday afternoon"
nylas calendar analyze         # AI-powered analytics
nylas calendar find-time --participants email1,email2 --duration 1h
```

**Key features:** DST detection, working hours validation, break protection, AI scheduling

**Details:** `docs/commands/calendar.md`, `docs/TIMEZONE.md`, `docs/AI.md`

---

## Contacts

```bash
nylas contacts list                                   # List contacts
nylas contacts create --name "NAME" --email "EMAIL"   # Create contact
nylas contacts update <id> --name "NEW NAME"          # Update contact
nylas contacts delete <contact-id>                    # Delete contact
```

**Details:** `docs/commands/contacts.md`

---

## Webhooks

```bash
nylas webhook create --url URL --triggers "event.created,event.updated"
nylas webhook list               # List webhooks
nylas webhook test <webhook-id>  # Test webhook
nylas webhook delete <webhook-id> # Delete webhook
```

**Details:** `docs/WEBHOOKS.md`, `docs/commands/webhooks.md`

---

## Timezone Utilities

```bash
nylas timezone convert --time "14:00" --from America/New_York --to Europe/London
nylas timezone list              # List timezones
nylas timezone now --zone "America/Los_Angeles"
```

**Details:** `docs/TIMEZONE.md`

---

## TUI (Terminal UI)

```bash
nylas tui                        # Launch interactive UI
```

**Navigation:** `↑/↓` navigate, `Enter` select, `q` quit, `?` help

**Details:** `docs/TUI.md`

---

## Utility Commands

```bash
nylas version                    # Show version
nylas doctor                     # System diagnostics
```

---

## Command Pattern

All commands follow consistent pattern:
- `nylas <resource> list` - List resources
- `nylas <resource> show <id>` - Show details
- `nylas <resource> create` - Create resource
- `nylas <resource> update <id>` - Update resource
- `nylas <resource> delete <id>` - Delete resource

---

**For detailed documentation on any feature, see:**
- Email: `docs/commands/email.md`
- Calendar: `docs/commands/calendar.md`
- Contacts: `docs/commands/contacts.md`
- Webhooks: `docs/commands/webhooks.md`
- Timezone: `docs/TIMEZONE.md`
- AI: `docs/AI.md`
