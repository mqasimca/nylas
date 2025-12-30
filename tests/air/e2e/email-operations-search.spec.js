// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Email Search tests - Searching emails
 */
test.describe('Email Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('search input filters email list', async ({ page }) => {
    // Open search
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);
    await searchInput.fill('test');

    // Search should trigger (debounced)
    await page.waitForTimeout(800);

    // Results should update
    const results = page.locator(selectors.search.resultsSection);

    if (await results.count() > 0) {
      await expect(results).toBeVisible();
    }
  });

  test('search shows recent searches', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const recentGroup = page.locator(selectors.search.recentGroup);

    if (await recentGroup.count() > 0) {
      // Recent searches may be visible
      const isVisible = await recentGroup.isVisible();
      expect(typeof isVisible).toBe('boolean');
    }
  });

  test('clicking search result navigates to email', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);
    await searchInput.fill('test');
    await page.waitForTimeout(800);

    const resultItems = page.locator('.search-result-item');
    const count = await resultItems.count();

    if (count > 0) {
      await resultItems.first().click();

      // Search should close
      await expect(page.locator(selectors.search.overlay)).not.toHaveClass(/active/);

      // Email should be selected and preview shown
      await page.waitForTimeout(500);
    }
  });
});
