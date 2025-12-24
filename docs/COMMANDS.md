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

## Demo Mode (No Account Required)

Explore the CLI with sample data before connecting your accounts:

```bash
nylas demo email list            # Browse sample emails
nylas demo calendar list         # View sample events
nylas demo contacts list         # See sample contacts
nylas demo notetaker list        # Explore AI notetaker
nylas demo tui                   # Interactive demo UI
```

All demo commands mirror real CLI structure: `nylas demo <feature> <command>`

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

## MCP (Model Context Protocol)

Enable AI assistants (Claude Desktop, Cursor, Windsurf, VS Code) to interact with your email and calendar.

```bash
nylas mcp install                          # Interactive assistant selection
nylas mcp install --assistant claude-code  # Install for Claude Code
nylas mcp install --assistant cursor       # Install for Cursor
nylas mcp install --all                    # Install for all detected assistants
nylas mcp status                           # Check installation status
nylas mcp uninstall --assistant cursor     # Remove configuration
nylas mcp serve                            # Start MCP server (used by assistants)
```

**Supported assistants:**
| Assistant | Config Location |
|-----------|-----------------|
| Claude Desktop | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Claude Code | `~/.claude.json` + permissions in `~/.claude/settings.json` |
| Cursor | `~/.cursor/mcp.json` |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` |
| VS Code | `.vscode/mcp.json` (project-level) |

**Features:**
- Auto-detects system timezone for consistent time display
- Auto-configures Claude Code permissions (`mcp__nylas__*`)
- Injects default grant ID for seamless authentication
- Local grant lookup (no email required for `get_grant`)

**Available MCP tools:** `list_messages`, `list_threads`, `list_calendars`, `list_events`, `create_event`, `update_event`, `send_message`, `create_draft`, `availability`, `get_grant`, `epoch_to_datetime`, `current_time`

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
- MCP: `docs/MCP.md`
