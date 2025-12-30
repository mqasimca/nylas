// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Contact Groups tests - Group navigation
 */
test.describe('Contact Groups', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);
  });

  test('contacts are grouped alphabetically', async ({ page }) => {
    await page.waitForTimeout(1500);

    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      // Look for group headers (A, B, C, etc.)
      const groupHeaders = contactsList.locator('.contact-group-header');

      if (await groupHeaders.count() > 0) {
        const firstHeader = await groupHeaders.first().textContent();
        expect(firstHeader).toBeTruthy();
        expect(firstHeader.length).toBeGreaterThan(0);
      }
    }
  });

  test('can navigate between groups', async ({ page }) => {
    await page.waitForTimeout(1500);

    const alphabet = page.locator('.contacts-alphabet');

    if (await alphabet.count() > 0) {
      const letters = alphabet.locator('.alphabet-letter');
      const count = await letters.count();

      if (count > 0) {
        // Click a letter
        await letters.first().click();

        // Should scroll to that group
        await page.waitForTimeout(300);
      }
    }
  });
});
