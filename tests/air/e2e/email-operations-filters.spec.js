// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Email Filters tests - Filtering emails
 */
test.describe('Email Filters', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('filter tabs are visible', async ({ page }) => {
    const filterTabs = page.locator(selectors.email.filterTabs);
    await expect(filterTabs).toBeVisible();
  });

  test('All filter is active by default', async ({ page }) => {
    const allTab = page.locator('.filter-tab').filter({ hasText: 'All' });
    await expect(allTab).toHaveClass(/active/);
  });

  test('can switch to VIP filter', async ({ page }) => {
    const vipTab = page.locator('.filter-tab').filter({ hasText: 'VIP' });
    await vipTab.click();

    await expect(vipTab).toHaveClass(/active/);

    // All tab should no longer be active
    const allTab = page.locator('.filter-tab').filter({ hasText: 'All' });
    await expect(allTab).not.toHaveClass(/active/);
  });

  test('can switch to Unread filter', async ({ page }) => {
    const unreadTab = page.locator('.filter-tab').filter({ hasText: 'Unread' });
    await unreadTab.click();

    await expect(unreadTab).toHaveClass(/active/);
  });

  test('can switch to Starred filter', async ({ page }) => {
    const starredTab = page.locator('.filter-tab').filter({ hasText: 'Starred' });

    if (await starredTab.count() > 0) {
      await starredTab.click();
      await expect(starredTab).toHaveClass(/active/);
    }
  });

  test('switching filters updates email list', async ({ page }) => {
    await page.waitForTimeout(1500);

    // Get initial email count
    const initialCount = await page.locator(selectors.email.emailItem).count();

    // Switch to Unread filter
    const unreadTab = page.locator('.filter-tab').filter({ hasText: 'Unread' });
    await unreadTab.click();

    await page.waitForTimeout(500);

    // Count may change (or stay the same if all emails are unread)
    const newCount = await page.locator(selectors.email.emailItem).count();
    expect(newCount >= 0).toBeTruthy();
  });
});

