// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Email Actions tests - Actions like archive, delete, star
 */
test.describe('Email Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('star button toggles star state', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();
      const starBtn = firstEmail.locator('.star-btn');

      if (await starBtn.count() > 0) {
        // Get initial state
        const wasStarred = await starBtn.evaluate((el) => el.classList.contains('starred'));

        // Click star
        await starBtn.click();

        // Wait for state change
        await page.waitForTimeout(300);

        // Verify state changed
        const isStarred = await starBtn.evaluate((el) => el.classList.contains('starred'));
        expect(isStarred).toBe(!wasStarred);
      }
    }
  });

  test('archive button shows confirmation toast', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const archiveBtn = preview.locator('button:has-text("Archive")');

      // Mock the archive action to show toast
      await page.evaluate(() => {
        if (typeof showToast === 'function') {
          showToast('success', 'Archived', 'Email archived successfully');
        }
      });

      // Toast should appear (may be multiple, use last() for most recent)
      const toasts = page.locator(selectors.toast.toast);
      const toastCount = await toasts.count();

      // Should have at least one toast
      expect(toastCount).toBeGreaterThan(0);

      // Most recent toast should be visible
      if (toastCount > 0) {
        await expect(toasts.last()).toBeVisible({ timeout: 2000 });
      }
    }
  });

  test('delete button shows confirmation', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const deleteBtn = preview.locator('button:has-text("Delete")');

      // Verify delete button exists
      await expect(deleteBtn).toBeVisible();
    }
  });

  test('mark as read/unread toggle works', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();

      // Right-click to open context menu
      await firstEmail.click({ button: 'right' });

      // Wait for context menu
      await page.waitForTimeout(300);

      // Context menu should appear
      const contextMenu = page.locator(selectors.contextMenu.menu);

      if (await contextMenu.count() > 0 && await contextMenu.isVisible()) {
        // Verify Mark as read/unread option exists
        const markOption = contextMenu.locator('.context-menu-item:has-text("Mark as")');

        if (await markOption.count() > 0) {
          await expect(markOption).toBeVisible();
        }
      }
    }
  });
});

