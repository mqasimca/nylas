# TUI2 Testing Guide

## Automated Testing

### Test Coverage Summary

We have **517 comprehensive unit tests** covering all keyboard functionality:

```bash
# Run all TUI2 tests
go test ./internal/tui2/... -v

# Run specific test suites
go test ./internal/tui2/models/... -run TestDashboard_KeyboardShortcuts -v
go test ./internal/tui2/models/... -run TestHelp_KeyboardShortcuts -v
go test ./internal/tui2/models/... -run TestSplash_SkipWithKeyPress -v
```

### Test Files

| File | Tests | Coverage |
|------|-------|----------|
| `models/dashboard_keyboard_test.go` | Dashboard navigation (a, c, p, d, s, ?, t) | ✅ |
| `models/help_test.go` | Help screen (esc, q, ctrl+c) | ✅ |
| `models/splash_test.go` | Splash skip (any key) | ✅ |
| `models/calendar_test.go` | Calendar navigation (m, w, g, t, n, r, esc) | ✅ |
| `models/messages_test.go` | Messages navigation (tab, h, l, r, enter) | ✅ |
| `models/contacts_test.go` | Contacts functionality | ✅ |
| `models/settings_test.go` | Settings toggle | ✅ |
| `models/debug_test.go` | Debug panel | ✅ |

## Manual Testing

### Build and Run

```bash
# Build the application
make build

# Run with Bubble Tea engine
./bin/nylas tui --engine bubbletea
```

### Keyboard Test Checklist

#### 1. Splash Screen (First 3 seconds)
- [ ] Press **any key** → Should skip to dashboard immediately
- [ ] See "Press any key to continue..." hint
- [ ] Or wait 3 seconds → Auto-transition to dashboard

#### 2. Dashboard Screen
- [ ] Press **a** → Navigate to Air (Messages)
- [ ] Press **c** → Navigate to Calendar
- [ ] Press **p** → Navigate to People (Contacts)
- [ ] Press **d** → Navigate to Debug Panel
- [ ] Press **s** → Navigate to Settings
- [ ] Press **?** → Navigate to Help
- [ ] Press **t** → Cycle through themes
- [ ] Press **Ctrl+C** → Quit application

#### 3. Help Screen
- [ ] Press **esc** → Go back to dashboard
- [ ] Press **q** → Go back to dashboard
- [ ] Press **Ctrl+C** → Quit application

#### 4. Settings Screen
- [ ] Press **esc** → Go back to dashboard
- [ ] Press **↑/↓** → Navigate settings
- [ ] Press **enter/space** → Toggle setting
- [ ] Press **←/→** → Cycle theme
- [ ] Press **Ctrl+C** → Quit application

#### 5. Debug Panel
- [ ] Press **esc** → Go back to dashboard
- [ ] Press **↑/↓** → Scroll logs
- [ ] Press **t** → Add test logs
- [ ] Press **Ctrl+C** → Quit application

#### 6. Calendar Screen
- [ ] Press **m** → Month view
- [ ] Press **w** → Week view
- [ ] Press **g** → Agenda view
- [ ] Press **t** → Go to today
- [ ] Press **n** → New event
- [ ] Press **r** → Refresh
- [ ] Press **esc** → Go back to dashboard

#### 7. Messages Screen
- [ ] Press **tab** → Switch between folders and messages
- [ ] Press **h** → Focus folders pane
- [ ] Press **l** → Focus messages pane
- [ ] Press **↑/↓** → Navigate items
- [ ] Press **enter** → Select/open
- [ ] Press **r** → Refresh
- [ ] Press **esc** → Go back to dashboard

## Why Unit Tests Are Sufficient for Bubble Tea v2

Bubble Tea follows the **Elm Architecture** pattern where:
1. **Update()** receives messages and returns new state
2. **View()** renders the current state
3. **Init()** sets up initial commands

Our unit tests verify:
- ✅ Key press messages create correct navigation commands
- ✅ Commands return correct NavigateMsg/BackMsg
- ✅ Model state updates correctly
- ✅ View renders without errors

This is the **standard testing approach** for Bubble Tea applications.

## Test Coverage Report

```bash
# Generate coverage report
make test-coverage

# View in browser (opens coverage.html)
```

## CI/CD Testing

```bash
# Complete CI pipeline (recommended)
make ci-full

# Quick CI (no integration tests)
make ci
```

## Known Limitations

- **teatest** (interactive testing framework) only supports Bubble Tea v1, not v2
- Once teatest adds v2 support, we can add interactive tests
- Current unit tests provide comprehensive coverage of all functionality

## Debugging Failed Tests

```bash
# Run specific test with verbose output
go test ./internal/tui2/models -run TestDashboard_KeyboardShortcuts -v

# Run with race detector
go test ./internal/tui2/... -race

# Run with coverage
go test ./internal/tui2/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```
