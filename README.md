# Nylas CLI

A unified command-line tool for Nylas API authentication, email management, calendar, contacts, webhooks, and OTP extraction.

## Features

- **Interactive TUI**: k9s-style terminal interface with vim-style commands, Google Calendar-style views, and email compose/reply
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

**Homebrew (macOS/Linux):**
```bash
brew install mqasimca/nylas/nylas
```

**Go Install:**
```bash
go install github.com/mqasimca/nylas/cmd/nylas@latest
```

**Download Binary:**

Download from [Releases](https://github.com/mqasimca/nylas/releases) and add to your PATH.

**Build from Source:**
```bash
make build
```

## Quick Start

```bash
# Configure with your Nylas credentials
nylas auth config

# Login with your email provider
nylas auth login

# Launch the interactive TUI
nylas tui

# Or use CLI commands directly
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

## Commands Overview

| Command | Description |
|---------|-------------|
| `nylas auth` | Authentication and account management |
| `nylas email` | Email operations (list, read, send, search) |
| `nylas calendar` | Calendar and event management |
| `nylas contacts` | Contact management |
| `nylas webhook` | Webhook configuration |
| `nylas otp` | OTP code extraction |
| `nylas tui` | Interactive terminal interface |
| `nylas doctor` | Diagnostic checks |

**[Full Command Reference](docs/COMMANDS.md)**

---

## TUI Highlights

![TUI Demo](docs/images/tui-demo.png)

```bash
nylas tui                    # Launch TUI at dashboard
nylas tui --demo             # Demo mode (no credentials needed)
nylas tui --theme amber      # Retro amber CRT theme
```

**Themes:** k9s, amber, green, apple2, vintage, ibm, futuristic, matrix, norton

**Vim-style keys:** `j/k` navigate, `gg/G` first/last, `dd` delete, `:q` quit, `/` search

**[Full TUI Documentation](docs/TUI.md)**

---

## Configuration

Credentials are stored securely in your system keyring:
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **macOS**: Keychain
- **Windows**: Windows Credential Manager

Config file location: `~/.config/nylas/config.yaml`

---

## Documentation

| Document | Description |
|----------|-------------|
| [Commands](docs/COMMANDS.md) | CLI command reference with examples |
| [TUI](docs/TUI.md) | Terminal UI themes, keys, customization |
| [Architecture](docs/ARCHITECTURE.md) | Hexagonal architecture overview |
| [Development](docs/DEVELOPMENT.md) | Testing, building, and contributing |
| [Security](docs/SECURITY.md) | Security practices and credential handling |

---

## Development

```bash
# Build
make build

# Test
make test

# Lint
make lint
```

**[Development Guide](docs/DEVELOPMENT.md)**

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
