# Nylas E2E Testing Guide

End-to-end testing for Nylas web interfaces using Playwright.

## Test Targets

| Command | Type | Port | Description |
|---------|------|------|-------------|
| `nylas air` | Web | 7365 | Modern web email client |
| `nylas ui` | Web | 7363 | Web-based CLI admin interface |

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

# Run all tests (Air + UI)
npm test

# Run Air tests only
npm run test:air

# Run UI tests only
npm run test:ui
```

## Directory Structure

```
tests/
├── air/                      # Nylas Air (web email client) tests
│   └── e2e/                  # E2E test specs
│       ├── smoke.spec.js
│       ├── navigation.spec.js
│       ├── keyboard.spec.js
│       └── ...
├── ui/                       # Nylas UI (CLI admin interface) tests
│   └── e2e/                  # E2E test specs
│       ├── smoke.spec.js
│       └── navigation.spec.js
├── shared/                   # Shared utilities
│   ├── fixtures/             # Mock data
│   │   ├── mock-emails.json
│   │   ├── mock-events.json
│   │   └── mock-contacts.json
│   └── helpers/              # Test utilities
│       ├── air-selectors.js  # Air element selectors
│       ├── ui-selectors.js   # UI element selectors
│       └── api-mocks.js      # API mocking utilities
├── playwright.config.js      # Multi-project Playwright config
├── package.json
└── README.md
```

## Running Tests

### All Tests
```bash
npm test                      # Run all (Air + UI)
make test-e2e                 # Via Makefile
```

### Air Tests Only
```bash
npm run test:air              # Air tests
npm run test:air:ui           # Interactive UI mode
npm run test:air:headed       # Headed browser
npm run test:air:debug        # Debug mode
make test-e2e-air             # Via Makefile
```

### UI Tests Only
```bash
npm run test:ui               # UI tests
npm run test:ui:ui            # Interactive UI mode
npm run test:ui:headed        # Headed browser
npm run test:ui:debug         # Debug mode
make test-e2e-ui              # Via Makefile
```

### Interactive Mode
```bash
make test-playwright-interactive  # Interactive for all
```

## Test Categories

### Air Tests (Web Email Client)

| File | Description |
|------|-------------|
| `smoke.spec.js` | Basic functionality |
| `navigation.spec.js` | View switching |
| `keyboard.spec.js` | Keyboard shortcuts |
| `modals.spec.js` | Modal interactions |
| `email-operations-*.spec.js` | Email features |
| `calendar-operations-*.spec.js` | Calendar features |
| `contacts-operations-*.spec.js` | Contact features |

### UI Tests (CLI Admin Interface)

| File | Description |
|------|-------------|
| `smoke.spec.js` | Page loads, basic elements |
| `navigation.spec.js` | Sidebar navigation, page switching |

## Writing Tests

### Air Test Example

```javascript
// tests/air/e2e/my-feature.spec.js
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

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

### UI Test Example

```javascript
// tests/ui/e2e/my-feature.spec.js
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/ui-selectors');

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

## Selectors

### Air Selectors

```javascript
const selectors = require('../../shared/helpers/air-selectors');

// Navigation
selectors.nav.main           // '[data-testid="main-nav"]'
selectors.nav.tabEmail       // '[data-testid="nav-tab-email"]'

// Views
selectors.views.email        // '[data-testid="email-view"]'
selectors.views.calendar     // '[data-testid="calendar-view"]'
```

### UI Selectors

```javascript
const selectors = require('../../shared/helpers/ui-selectors');

// Header
selectors.header.header      // '.header'
selectors.header.themeBtn    // '.theme-btn'

// Navigation
selectors.nav.overview       // '[data-page="overview"]'
selectors.nav.email          // '[data-page="email"]'

// Pages
selectors.pages.overview     // '#page-overview'
selectors.dashboard.title    // '.dashboard-title'
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AIR_PORT` | Air server port | 7365 |
| `UI_PORT` | UI server port | 7363 |
| `CI` | Running in CI | - |

### Playwright Config

```javascript
{
  projects: [
    {
      name: 'air-chromium',
      testDir: './air/e2e',
      use: { baseURL: 'http://localhost:7365' },
    },
    {
      name: 'ui-chromium',
      testDir: './ui/e2e',
      use: { baseURL: 'http://localhost:7363' },
    },
  ],
  webServer: [
    { command: 'nylas air --no-browser', port: 7365 },
    { command: 'nylas ui --no-browser', port: 7363 },
  ],
}
```

## Make Targets

| Target | Description |
|--------|-------------|
| `make test-e2e` | Run all E2E tests |
| `make test-e2e-air` | Run Air tests only |
| `make test-e2e-ui` | Run UI tests only |
| `make test-playwright-interactive` | Interactive mode |
| `make test-playwright-headed` | Headed browser mode |

## Troubleshooting

**Tests timeout:**
- Check if servers are running
- Increase timeout in config

**Element not found:**
- Verify selector in DevTools
- Add explicit wait

### Debug Tips

```bash
# Run with headed mode
npx playwright test --headed

# Run with debug mode
npx playwright test --debug

# Slow motion
npx playwright test --headed --slow-mo=500
```

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
