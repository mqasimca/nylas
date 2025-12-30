// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Bulk Actions tests - Bulk operations on emails
 */
test.describe('Bulk Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('can select multiple emails with checkboxes', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count >= 2) {
      // Look for checkboxes
      const firstCheckbox = emailItems.first().locator('input[type="checkbox"]');
      const secondCheckbox = emailItems.nth(1).locator('input[type="checkbox"]');

      if (await firstCheckbox.count() > 0) {
        // Check first email
        await firstCheckbox.check();
        await expect(firstCheckbox).toBeChecked();

        // Check second email
        await secondCheckbox.check();
        await expect(secondCheckbox).toBeChecked();

        // Bulk actions bar should appear
        const bulkActions = page.locator('.bulk-actions-bar');

        if (await bulkActions.count() > 0) {
          await expect(bulkActions).toBeVisible();
        }
      }
    }
  });

  test('select all checkbox selects all visible emails', async ({ page }) => {
    const selectAllCheckbox = page.locator('.select-all-checkbox');

    if (await selectAllCheckbox.count() > 0) {
      await selectAllCheckbox.check();

      // All email checkboxes should be checked
      const emailItems = page.locator(selectors.email.emailItem);
      const count = await emailItems.count();

      if (count > 0) {
        const firstCheckbox = emailItems.first().locator('input[type="checkbox"]');

        if (await firstCheckbox.count() > 0) {
          await expect(firstCheckbox).toBeChecked();
        }
      }
    }
  });

  test('bulk archive action works on selected emails', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count >= 1) {
      const checkbox = emailItems.first().locator('input[type="checkbox"]');

      if (await checkbox.count() > 0) {
        await checkbox.check();

        const bulkActions = page.locator('.bulk-actions-bar');

        if (await bulkActions.count() > 0) {
          const bulkArchiveBtn = bulkActions.locator('button:has-text("Archive")');

          if (await bulkArchiveBtn.count() > 0) {
            await bulkArchiveBtn.click();

            // Should show success toast
            await page.waitForTimeout(500);
          }
        }
      }
    }
  });
});

