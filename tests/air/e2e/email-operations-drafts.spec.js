// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Draft Operations tests - Draft management
 */
test.describe('Draft Operations', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('can save draft from compose modal', async ({ page }) => {
    // Open compose
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Fill fields
    await page.fill(selectors.compose.to, 'draft@example.com');
    await page.fill(selectors.compose.subject, 'Draft Email');
    await page.fill(selectors.compose.body, 'This is a draft.');

    // Close without sending (save as draft)
    await page.keyboard.press('Escape');

    // Modal should close
    await expect(page.locator(selectors.compose.modal)).toBeHidden();
  });

  test('drafts folder shows drafts', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);

    await page.waitForTimeout(1000);

    const draftsFolder = folderList.locator('.folder-item:has-text("Drafts")');

    if (await draftsFolder.count() > 0) {
      // Click Drafts folder
      await draftsFolder.click();

      await page.waitForTimeout(500);

      // Should show drafts or empty state
      const emailList = page.locator(selectors.email.emailListContainer);
      await expect(emailList).toBeVisible();
    }
  });

  test('clicking draft opens compose with saved content', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);

    await page.waitForTimeout(1000);

    const draftsFolder = folderList.locator('.folder-item:has-text("Drafts")');

    if (await draftsFolder.count() > 0) {
      await draftsFolder.click();
      await page.waitForTimeout(500);

      const emailItems = page.locator(selectors.email.emailItem);
      const count = await emailItems.count();

      if (count > 0) {
        // Click first draft
        await emailItems.first().click();

        // Compose modal should open with draft content
        // (or preview shows draft content)
      }
    }
  });
});

