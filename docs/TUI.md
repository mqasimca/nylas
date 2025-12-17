# Terminal User Interface (TUI)

Interactive k9s-style terminal interface for managing your Nylas email, calendar, and more.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Architecture](ARCHITECTURE.md) | [Development](DEVELOPMENT.md) | [Security](SECURITY.md)

![TUI Demo](images/tui-demo.png)

---

## Quick Start

```bash
nylas tui                    # Launch TUI at dashboard
nylas tui messages           # Launch directly to messages view
nylas tui events             # Launch directly to calendar view
nylas tui contacts           # Launch directly to contacts view
nylas tui webhooks           # Launch directly to webhooks view
nylas tui grants             # Launch directly to grants view
nylas tui --refresh 5        # Custom refresh interval (seconds)
nylas tui --theme amber      # Launch with retro amber theme
nylas tui --theme green      # Launch with green phosphor theme
nylas tui --demo             # Launch in demo mode (no credentials needed)
nylas tui --demo --theme amber  # Demo mode with custom theme
```

---

## Demo Mode

Launch the TUI without any credentials using realistic sample data. Perfect for:
- Screenshots and documentation
- Demos and presentations
- Exploring the interface before configuring credentials

```bash
# Launch demo mode
nylas tui --demo

# Demo mode with a specific theme (great for screenshots)
nylas tui --demo --theme amber
nylas tui --demo --theme matrix

# Demo mode starting at a specific view
nylas tui messages --demo
nylas tui events --demo --theme green
```

Demo mode displays sample data including emails from realistic senders (GitHub, AWS, etc.), calendar events, contacts, and webhooks.

---

## Features

- **k9s-style Interface**: Familiar keyboard-driven navigation
- **Vim-style Commands**: Full vim keybindings (`gg`, `G`, `dd`, `Ctrl+d/u/f/b`, `:q`, `:wq`, etc.)
- **Retro Themes**: 9 color themes including amber/green phosphor CRT and Norton Commander DOS styles
- **Real-time Updates**: Auto-refresh with configurable interval
- **Multiple Views**: Messages, Calendar, Contacts, Webhooks, Grants
- **Google Calendar-style**: Month, Week, and Agenda views for events
- **Email Compose**: Write and send emails directly from the TUI
- **Reply/Reply All**: Quick reply to messages with quoted content
- **Mouse Support**: Click to select, double-click to open

---

## Themes

Inspired by [cool-retro-term](https://github.com/Swordfish90/cool-retro-term):

| Theme | Description | Color |
|-------|-------------|-------|
| `k9s` | Default k9s style | Blue/Orange |
| `amber` | Amber phosphor CRT | #ff8100 |
| `green` | Green phosphor CRT | #0ccc68 |
| `apple2` | Apple ][ style | #00d56d |
| `vintage` | Vintage neon green | #00ff3e |
| `ibm` | IBM DOS white | #ffffff |
| `futuristic` | Steel blue futuristic | #729fcf |
| `matrix` | Matrix green | #00ff00 |
| `norton` | Norton Commander DOS | #0000AA |

---

## Custom Themes

Create your own themes using the same YAML format as k9s:

```bash
# Create a custom theme
nylas tui theme init mytheme

# List available themes (built-in and custom)
nylas tui theme list

# Validate a custom theme file
nylas tui theme validate mytheme

# Set a theme as the default (persists across sessions)
nylas tui theme set-default mytheme

# Use your custom theme (one-time)
nylas tui --theme mytheme
```

### Setting a Default Theme

```bash
# Set amber as your default theme
nylas tui theme set-default amber

# Set a custom theme as default
nylas tui theme set-default mytheme

# Reset to the built-in k9s theme
nylas tui theme set-default k9s
```

The default theme is saved to `~/.config/nylas/config.yaml`. You can override it anytime with the `--theme` flag.

### Theme Validation

```bash
# Validate a custom theme
nylas tui theme validate mytheme

# Output shows:
#   - File path and size
#   - Colors found in the theme
#   - Any errors or warnings
#   - Suggestions for fixing common issues
```

### Theme File Format

Custom themes are stored in `~/.config/nylas/themes/<name>.yaml` and use the k9s skin format:

```yaml
# ~/.config/nylas/themes/mytheme.yaml
foreground: "#c0caf5"
background: "#1a1b26"
red: "#f7768e"
green: "#9ece6a"
yellow: "#e0af68"
blue: "#7aa2f7"
magenta: "#bb9af7"
cyan: "#7dcfff"

k9s:
  body:
    fgColor: "#c0caf5"
    bgColor: "#1a1b26"
    logoColor: "#bb9af7"
  frame:
    border:
      fgColor: "#3b4261"
      focusColor: "#7aa2f7"
    menu:
      fgColor: "#7dcfff"
      keyColor: "#9ece6a"
  views:
    table:
      fgColor: "#c0caf5"
      header:
        fgColor: "#7aa2f7"
      selected:
        fgColor: "#1a1b26"
        bgColor: "#7aa2f7"
```

See [k9s skins documentation](https://k9scli.io/topics/skins/) for more color options.

---

## Full Retro CRT Experience

For the authentic retro CRT look with scanlines, phosphor glow, and screen curvature, run the TUI inside [cool-retro-term](https://github.com/Swordfish90/cool-retro-term):

```bash
# Install cool-retro-term (macOS)
brew install --cask cool-retro-term

# Open cool-retro-term
open -a "cool-retro-term"

# Inside cool-retro-term, run:
nylas tui --theme amber
```

Right-click inside cool-retro-term to access settings and switch between CRT profiles (Amber, Green, Apple II, etc.).

---

## View Aliases

```bash
nylas tui m                  # Messages (alias)
nylas tui e                  # Events (alias)
nylas tui cal                # Calendar (alias)
nylas tui c                  # Contacts (alias)
nylas tui w                  # Webhooks (alias)
nylas tui g                  # Grants (alias)
```

---

## Keyboard Reference

### Vim-style Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `gg` | Go to first row |
| `G` | Go to last row |
| `Ctrl+d` | Half page down |
| `Ctrl+u` | Half page up |
| `Ctrl+f` | Full page down |
| `Ctrl+b` | Full page up |
| `:N` | Jump to row N (e.g., `:5`) |
| `Enter` | Open/select |
| `Esc` | Go back |
| `Tab` | Switch panels |

### Vim-style Commands

| Command | Action |
|---------|--------|
| `:` | Enter command mode |
| `/` | Filter/search |
| `?` | Show help |
| `r` | Refresh |
| `:q` | Quit |
| `:q!` | Force quit |
| `:wq` / `:x` | Save and quit |
| `:h` / `:help` | Show help |
| `:e <view>` | Open view (e.g., `:e messages`) |

### Vim-style Actions

| Key/Command | Action |
|-------------|--------|
| `dd` | Delete current item |
| `x` | Delete current item |
| `:star` | Star message |
| `:unread` | Mark as unread |
| `:reply` | Reply to message |
| `:compose` | Compose new message |

### Resource Commands (in command mode)

| Command | Action |
|---------|--------|
| `:m` or `:messages` | Go to Messages |
| `:e` or `:events` | Go to Calendar |
| `:c` or `:contacts` | Go to Contacts |
| `:w` or `:webhooks` | Go to Webhooks |
| `:g` or `:grants` | Go to Grants |
| `:d` or `:dashboard` | Go to Dashboard |

### Messages View Keys

| Key | Action |
|-----|--------|
| `n` | Compose new email |
| `R` | Reply to message |
| `A` | Reply all |
| `s` | Toggle star |
| `u` | Mark as unread |

### Calendar View Keys

| Key | Action |
|-----|--------|
| `m` | Month view |
| `w` | Week view |
| `a` | Agenda view |
| `t` | Go to today |
| `H/L` | Previous/next month |
| `h/l` | Previous/next day |
| `c` | Cycle calendars |
| `C` | Show calendar list |

### Compose Email Keys

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Ctrl+S` | Send email |
| `Esc` | Cancel |

---

## Views

### Messages View
![Messages View](images/tui-view-messages.png)

### Calendar View
![Calendar View](images/tui-view-calendar.png)

### Contacts View
![Contacts View](images/tui-view-contacts.png)

### Webhooks View
![Webhooks View](images/tui-view-webhooks.png)
