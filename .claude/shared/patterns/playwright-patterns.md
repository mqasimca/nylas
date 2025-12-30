# Playwright E2E Test Patterns

Shared patterns for Playwright E2E tests in the Nylas CLI project.

---

## Two Test Projects

| Project | Port | Location | Purpose |
|---------|------|----------|---------|
| **Air** | 7365 | `tests/air/e2e/` | Modern web email client |
| **UI** | 7363 | `tests/ui/e2e/` | CLI admin interface |

---

## Selector Strategy (Priority Order)

| Priority | Selector Type | Example |
|----------|--------------|---------|
| 1 | Role selectors | `getByRole('button', { name: 'Send' })` |
| 2 | Text selectors | `getByText('Welcome')` |
| 3 | Label selectors | `getByLabel('Email')` |
| 4 | Test IDs | `getByTestId('submit-btn')` (last resort) |

**NEVER use:** CSS selectors, XPath, fragile DOM paths

---

## Air Test Template

```javascript
// tests/air/e2e/feature.spec.js
const { test, expect } = require('@playwright/test');

test.describe('Feature Name', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/');
        await page.waitForSelector('.app-loaded');
    });

    test('user can perform action', async ({ page }) => {
        // Arrange - navigate to correct state
        await page.getByRole('link', { name: 'Inbox' }).click();

        // Act - perform the action
        await page.getByRole('button', { name: 'Compose' }).click();
        await page.getByLabel('To').fill('test@example.com');
        await page.getByLabel('Subject').fill('Test Subject');
        await page.getByRole('button', { name: 'Send' }).click();

        // Assert - verify the result
        await expect(page.getByText('Message sent')).toBeVisible();
    });

    test('handles error gracefully', async ({ page }) => {
        // Test error scenarios
        await page.getByRole('button', { name: 'Send' }).click();
        await expect(page.getByText('Please fill required fields')).toBeVisible();
    });
});
```

---

## UI Test Template

```javascript
// tests/ui/e2e/admin.spec.js
const { test, expect } = require('@playwright/test');

test.describe('Admin Panel', () => {
    test('should load dashboard', async ({ page }) => {
        await page.goto('/');
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    });

    test('should display grants list', async ({ page }) => {
        await page.goto('/grants');
        await expect(page.getByRole('table')).toBeVisible();
    });
});
```

---

## Shared Fixtures & Helpers

**Fixtures:** `tests/shared/fixtures/`
- `mock-contacts.json`
- `mock-emails.json`
- `mock-events.json`

**Helpers:** `tests/shared/helpers/`
- `air-selectors.js`
- `ui-selectors.js`
- `api-mocks.js`

---

## Test Categories

| Category | What to Test |
|----------|--------------|
| Happy Path | Normal user workflows |
| Error Cases | Invalid inputs, network failures |
| Edge Cases | Empty states, long strings, unicode |
| Boundary | Pagination, limits, first/last items |

---

## Commands

```bash
npx playwright test                              # Run all tests
npx playwright test --project=air                # Air only
npx playwright test --project=ui                 # UI only
npx playwright test tests/air/e2e/email.spec.js  # Specific file
npx playwright test --ui                         # Interactive mode
npx playwright test --debug                      # Debug mode
```

---

## Rules

1. **Semantic selectors** - No CSS/XPath
2. **Independent tests** - No shared state
3. **Descriptive names** - Test name describes scenario
4. **Arrange-Act-Assert** - Clear test structure
5. **Test behavior** - Not implementation details
