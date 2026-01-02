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

## Web UI

Launch a local web interface for managing Nylas CLI:

```bash
nylas ui                         # Start on default port (3000)
nylas ui --port 8080             # Custom port
nylas ui --no-browser            # Don't auto-open browser
```

**Features:**
- Configure API credentials visually
- View and switch between authenticated accounts
- Execute email, calendar, and auth commands
- ID caching with autocomplete suggestions
- Command output with copy functionality

**Security:**
- Runs on localhost only (not accessible externally)
- Command whitelist prevents arbitrary execution
- Shell injection protection

**URL:** `http://localhost:3000` (default)

---

## Air (Modern Email Client)

Launch **Nylas Air** - a modern, keyboard-driven email client that runs in your browser:

```bash
nylas air                        # Start on default port (7365)
nylas air --port 8080            # Custom port
nylas air --no-browser           # Don't auto-open browser
nylas air --clear-cache          # Clear all cached data before starting
nylas air --encrypted            # Enable encryption for cached data
```

**Features:**
- **Three-pane interface:** Folders, message list, preview
- **Calendar & Contacts:** Full calendar and contact management
- **Keyboard shortcuts:** J/K navigate, C compose, E archive
- **Command palette:** Cmd+K for quick actions
- **Dark mode:** Customizable themes
- **AI-powered:** Email summaries, smart replies
- **Local caching:** Full-text search with offline support
- **Action queuing:** Queue actions when offline
- **Encryption:** Optional encryption for cached data (system keyring)

**Security:**
- Runs on localhost only (not accessible externally)
- All data stored locally on your machine
- Optional encryption for cached data using system keyring

**URL:** `http://localhost:7365` (default)

**Testing:**
```bash
make ci-full                     # Complete CI pipeline (includes Air tests + cleanup)
make test-air-integration        # Run Air integration tests only
```

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

## Slack Integration

Interact with Slack workspaces directly from the CLI.

### Authentication

```bash
nylas slack auth set --token xoxp-...      # Store Slack user token
nylas slack auth status                     # Check authentication status
nylas slack auth remove                     # Remove stored token
```

**Token sources (checked in order):**
1. `SLACK_USER_TOKEN` environment variable
2. System keyring (set via `nylas slack auth set`)

**Get your token:** [api.slack.com/apps](https://api.slack.com/apps) → Your App → OAuth & Permissions → User OAuth Token

### Channels

```bash
# List channels you're a member of
nylas slack channels list                   # List your channels
nylas slack channels list --type public_channel  # List public channels only
nylas slack channels list --type private_channel # List private channels
nylas slack channels list --exclude-archived     # Exclude archived channels
nylas slack channels list --limit 20             # Limit results
nylas slack channels list --id                   # Show channel IDs

# Filter by creation date
nylas slack channels list --created-after 24h    # Channels created in last 24 hours
nylas slack channels list --created-after 7d     # Channels created in last 7 days
nylas slack channels list --created-after 2w     # Channels created in last 2 weeks

# Workspace-wide listing (slower, may hit rate limits)
nylas slack channels list --all-workspace        # List all workspace channels
nylas slack channels list --all                  # Fetch all pages

# Get channel info
nylas slack channels info C01234567890           # Get detailed channel info
```

### Messages

```bash
nylas slack messages list --channel general       # List messages from channel
nylas slack messages list --channel-id C01234567  # Use channel ID
nylas slack messages list --channel general --limit 10  # Limit results
nylas slack messages list --channel general --id  # Show message timestamps
nylas slack messages list --channel general --thread 1234567890.123456  # Show thread replies
```

### Send & Reply

```bash
# Send a message
nylas slack send --channel general --text "Hello team!"
nylas slack send --channel general --text "Message" --yes  # Skip confirmation

# Reply to a thread
nylas slack reply --channel general --thread 1234567890.123456 --text "Reply"
```

### Users

```bash
nylas slack users list                      # List all users
nylas slack users list --limit 50           # Limit results
nylas slack users list --id                 # Show user IDs
```

### Search

```bash
nylas slack search --query "project update" # Search messages
nylas slack search --query "from:@john"     # Search with Slack modifiers
nylas slack search --query "in:#general"    # Search in specific channel
nylas slack search --query "meeting" --limit 20
```

**Required OAuth Scopes:**
- `channels:read`, `groups:read`, `im:read`, `mpim:read` - List channels
- `channels:history`, `groups:history`, `im:history`, `mpim:history` - Read messages
- `chat:write` - Send messages
- `users:read` - List users
- `search:read` - Search messages

---

## Notetaker (AI Meeting Bot)

Manage Nylas Notetaker bots that join video meetings to record and transcribe.

```bash
# List all notetakers
nylas notetaker list                              # List all notetakers
nylas notetaker list --state scheduled            # Filter by state
nylas notetaker list --limit 10                   # Limit results

# Create a notetaker to join a meeting
nylas notetaker create --meeting-link "https://zoom.us/j/123456789"
nylas notetaker create --meeting-link "https://meet.google.com/abc-defg" --join-time "tomorrow 2pm"
nylas notetaker create --meeting-link "https://zoom.us/j/123" --bot-name "Meeting Recorder"

# Show notetaker details
nylas notetaker show <notetaker-id>

# Get recording and transcript URLs
nylas notetaker media <notetaker-id>

# Delete/cancel a notetaker
nylas notetaker delete <notetaker-id>             # With confirmation
nylas notetaker delete <notetaker-id> --yes       # Skip confirmation
```

**Aliases:** `nylas nt`, `nylas bot`

**Supported Providers:** Zoom, Google Meet, Microsoft Teams

**States:** `scheduled`, `connecting`, `waiting`, `attending`, `processing`, `complete`, `cancelled`, `failed`

**Join Time Formats:**
- ISO: `2024-01-15 14:00`
- Relative: `30m`, `2h`, `1d`
- Natural: `tomorrow 9am`, `tomorrow 2:30pm`

---

## OTP (One-Time Password)

Retrieve OTP/verification codes from email automatically.

```bash
# Get the latest OTP code (auto-copies to clipboard)
nylas otp get                                     # From default account
nylas otp get user@example.com                    # From specific account
nylas otp get --no-copy                           # Don't copy to clipboard
nylas otp get --raw                               # Output only the code

# Watch for new OTP codes (continuous polling)
nylas otp watch                                   # Poll every 10 seconds
nylas otp watch --interval 5                      # Poll every 5 seconds
nylas otp watch user@example.com                  # Watch specific account

# List configured accounts
nylas otp list

# Debug: Show recent messages with OTP detection
nylas otp messages                                # Show last 10 messages
nylas otp messages --limit 20                     # Show more messages
```

**Features:**
- Auto-copies OTP to clipboard
- Supports 4-8 digit codes
- Detects OTPs from common providers (Google, Microsoft, banks, etc.)
- Pretty-printed display with sender info

---

## Utility Commands

```bash
nylas version                    # Show version
nylas doctor                     # System diagnostics
nylas update                     # Update CLI to latest version
nylas update --check             # Check for updates without installing
nylas update --force             # Force update even if on latest
nylas update --yes               # Skip confirmation prompt
```

**Update command features:**
- Downloads from GitHub releases
- SHA256 checksum verification
- Automatic backup and restore on failure
- Detects Homebrew installs (redirects to `brew upgrade`)

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
- Slack: `docs/COMMANDS.md#slack-integration`
- Notetaker: `docs/COMMANDS.md#notetaker-ai-meeting-bot`
- OTP: `docs/COMMANDS.md#otp-one-time-password`
- Timezone: `docs/TIMEZONE.md`
- AI: `docs/AI.md`
- MCP: `docs/MCP.md`
