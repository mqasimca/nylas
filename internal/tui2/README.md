# TUI2 - Bubble Tea TUI (Experimental)

This package contains an **experimental** TUI implementation using [Bubble Tea](https://github.com/charmbracelet/bubbletea), an Elm-inspired framework.

## Status: Experimental

This is an alternative TUI implementation alongside the primary `internal/tui/` (tview-based).
Use `--engine bubbletea` flag to try it:

```bash
nylas tui --engine bubbletea
```

The primary TUI (`internal/tui/`) remains the default and recommended option.

## Architecture

Uses the **Elm Architecture** pattern:
- **Model**: Application state
- **Update**: State transitions via messages
- **View**: Render state to terminal

### Directory Structure

| Directory | Purpose |
|-----------|---------|
| `models/` | Screen models (Dashboard, Messages, Calendar, Compose, etc.) |
| `components/` | Reusable Bubble Tea components (dialogs, forms, grids) |
| `styles/` | Lip Gloss themes and styling |
| `state/` | Global state management |
| `utils/` | Utilities (rate limiting, folder helpers) |

### Key Files

| File | Purpose |
|------|---------|
| `app.go` | Root application, model stack pattern |
| `messages.go` | Message types for Elm Architecture |

## Models (Screens)

| Model | File | Purpose |
|-------|------|---------|
| Dashboard | `models/dashboard.go` | Overview with stats and quick actions |
| Messages | `models/messages.go` | Email inbox with three-pane layout |
| MessageDetail | `models/message_detail.go` | Single email view |
| Compose | `models/compose.go` | Email composition |
| Calendar | `models/calendar.go` | Calendar with event grid |

## Components

| Component | File | Purpose |
|-----------|------|---------|
| ThreePane | `components/three_pane.go` | Folders / List / Detail layout |
| CalendarGrid | `components/calendar_grid.go` | Month/week/day calendar views |
| SearchDialog | `components/search_dialog.go` | Search with filters |
| ConfirmDialog | `components/confirm_dialog.go` | Yes/No confirmations |
| EventForm | `components/event_form.go` | Event creation/editing |

## Comparison with Primary TUI

| Aspect | tui (Primary) | tui2 (Experimental) |
|--------|---------------|---------------------|
| Framework | tview | Bubble Tea |
| Pattern | Widget-based | Elm Architecture |
| Status | Stable, default | Experimental |
| Command | `nylas tui` | `nylas tui --engine bubbletea` |

## Development Notes

When working on this package:
1. Follow Elm Architecture principles
2. Use Lip Gloss for styling (not raw ANSI)
3. All state changes go through messages
4. Keep models focused on single screens
5. Extract reusable UI to `components/`
