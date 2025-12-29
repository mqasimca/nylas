# Nylas Air E2E Testing Guide

End-to-end testing for Nylas Air using Playwright.

## Prerequisites

- Node.js 18+
- Go 1.24+
- Playwright browsers installed

## Quick Start

```bash
# Install dependencies
cd tests
npm install

# Install Playwright browsers (first time only)
npx playwright install chromium

# Run all tests
npm run test:e2e

# Run tests with UI mode (interactive)
npm run test:e2e:ui

# Run tests for CI (generates HTML report)
npm run test:e2e:ci
```

## Test Structure

```
tests/
├── e2e/                    # Test spec files
│   ├── smoke.spec.js       # Basic smoke tests
│   ├── navigation.spec.js  # View switching tests
│   ├── modals.spec.js      # Modal interaction tests
│   └── keyboard.spec.js    # Keyboard shortcut tests
├── fixtures/               # Mock data
│   ├── mock-emails.json
│   ├── mock-events.json
│   └── mock-contacts.json
├── helpers/                # Shared utilities
│   ├── selectors.js        # Element selectors
│   └── api-mocks.js        # API mocking utilities
├── playwright.config.js    # Playwright configuration
└── package.json
```

## Running Tests

### All Tests
```bash
npm run test:e2e
```

### Specific Test File
```bash
npx playwright test smoke.spec.js
npx playwright test navigation.spec.js
```

### Specific Test
```bash
npx playwright test -g "page loads successfully"
```

### Headed Mode (See Browser)
```bash
npx playwright test --headed
```

### Debug Mode
```bash
npx playwright test --debug
```

### UI Mode (Interactive)
```bash
npm run test:e2e:ui
```

## Test Categories

### Smoke Tests (`smoke.spec.js`)
Basic functionality verification:
- Page loads successfully
- Main navigation is visible
- Email view is default
- Accessibility elements present

### Navigation Tests (`navigation.spec.js`)
View switching and navigation:
- Switch between Email, Calendar, Contacts, Notetaker
- Keyboard shortcuts (1, 2, 3 keys)
- Search overlay functionality
- ARIA attributes

### Modal Tests (`modals.spec.js`)
Modal interactions:
- Compose modal (open, close, fill form)
- Settings modal (theme selection)
- Command palette
- Keyboard shortcuts overlay

### Keyboard Tests (`keyboard.spec.js`)
Keyboard shortcut functionality:
- Global shortcuts (C, ?, Shift+F)
- Modifier keys (Cmd+K, Ctrl+K)
- Shortcuts blocked in input fields
- Navigation shortcuts

## Selectors Strategy

Tests use `data-testid` attributes for stable element selection:

```javascript
// Good - stable selector
page.locator('[data-testid="compose-modal"]')

// Fallback - semantic selector
page.locator('.compose-btn')

// Avoid - text-based (breaks with i18n)
page.getByText('Compose')
```

### Available Test IDs

| Element | data-testid |
|---------|-------------|
| Main nav | `main-nav` |
| Email tab | `nav-tab-email` |
| Calendar tab | `nav-tab-calendar` |
| Contacts tab | `nav-tab-contacts` |
| Email view | `email-view` |
| Calendar view | `calendar-view` |
| Contacts view | `contacts-view` |
| Compose modal | `compose-modal` |
| Compose To | `compose-to` |
| Compose Subject | `compose-subject` |
| Compose Body | `compose-body` |
| Compose Send | `compose-send` |
| Command palette | `command-palette` |
| Search overlay | `search-overlay` |
| Settings overlay | `settings-overlay` |
| Shortcut overlay | `shortcut-overlay` |

See `helpers/selectors.js` for the complete list.

## API Mocking

Use the `api-mocks.js` helpers to mock API responses:

```javascript
const { mockEmailsAPI, mockFoldersAPI } = require('../helpers/api-mocks');

test.beforeEach(async ({ page }) => {
  await mockEmailsAPI(page, require('../fixtures/mock-emails.json').messages);
  await mockFoldersAPI(page, require('../fixtures/mock-emails.json').folders);
});
```

### Available Mocks

- `mockEmailsAPI(page, messages)` - Mock messages endpoint
- `mockFoldersAPI(page, folders)` - Mock folders endpoint
- `mockCalendarsAPI(page, calendars)` - Mock calendars endpoint
- `mockEventsAPI(page, events)` - Mock events endpoint
- `mockContactsAPI(page, contacts)` - Mock contacts endpoint
- `mockAPIError(page, endpoint, status)` - Mock error response
- `mockNetworkFailure(page, pattern)` - Mock network failure
- `mockAllAPIsEmpty(page)` - Mock all APIs with empty data

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BASE_URL` | App URL | `http://localhost:7365` |
| `CI` | Running in CI | - |

### Playwright Config

Key settings in `playwright.config.js`:

```javascript
{
  testDir: './e2e',
  timeout: 30000,
  retries: process.env.CI ? 2 : 0,
  webServer: {
    command: 'cd .. && go run cmd/nylas/main.go air --no-browser',
    port: 7365,
    reuseExistingServer: !process.env.CI,
  }
}
```

## CI Integration

Tests run automatically in CI:

```yaml
# GitHub Actions example
- name: Run E2E Tests
  run: |
    cd tests
    npm ci
    npx playwright install chromium
    npm run test:e2e:ci
```

### Artifacts

On failure, these artifacts are saved:
- Screenshots in `test-results/`
- Traces in `test-results/` (on first retry)
- HTML report in `playwright-report/`

## Writing New Tests

### 1. Create Test File

```javascript
// tests/e2e/my-feature.spec.js
const { test, expect } = require('@playwright/test');
const selectors = require('../helpers/selectors');

test.describe('My Feature', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('does something', async ({ page }) => {
    // Test implementation
  });
});
```

### 2. Add Selectors (if needed)

Add new selectors to `helpers/selectors.js`:

```javascript
exports.myFeature = {
  button: '[data-testid="my-button"]',
  modal: '[data-testid="my-modal"]',
};
```

### 3. Add Test IDs to Templates

Add `data-testid` attributes to Go templates:

```html
<button data-testid="my-button">Click Me</button>
```

## Best Practices

1. **Use data-testid** - Stable, decoupled from styling
2. **Wait for elements** - Use `expect().toBeVisible()` before interacting
3. **Isolate tests** - Each test should work independently
4. **Mock API responses** - Don't depend on backend state
5. **Test user flows** - Focus on what users do, not implementation
6. **Keep tests fast** - Avoid unnecessary waits

## Troubleshooting

### Tests timeout
- Check if Air server is running
- Increase timeout in config
- Check network/API mocking

### Element not found
- Verify selector in DevTools
- Check if element is in DOM
- Add explicit wait

### Flaky tests
- Add proper waits for async operations
- Use `await expect().toBeVisible()`
- Check for race conditions

### Debug tips
```bash
# Run with headed mode
npx playwright test --headed

# Run with debug mode
npx playwright test --debug

# Show browser console
npx playwright test --headed --slow-mo=500
```

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Test API](https://playwright.dev/docs/api/class-test)
- [Playwright Selectors](https://playwright.dev/docs/selectors)
- [Nylas Air Architecture](../internal/air/README.md)
