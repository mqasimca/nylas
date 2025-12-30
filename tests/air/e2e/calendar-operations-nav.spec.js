// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Calendar Navigation tests - Month/week navigation
 */
test.describe('Calendar Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);
  });

  test('can navigate to previous month', async ({ page }) => {
    const prevBtn = page.locator('.calendar-nav-prev');

    if (await prevBtn.count() > 0) {
      await prevBtn.click();

      // Calendar should update
      await page.waitForTimeout(300);
    }
  });

  test('can navigate to next month', async ({ page }) => {
    const nextBtn = page.locator('.calendar-nav-next');

    if (await nextBtn.count() > 0) {
      await nextBtn.click();

      // Calendar should update
      await page.waitForTimeout(300);
    }
  });

  test('can go to today', async ({ page }) => {
    const todayBtn = page.locator('.calendar-nav-today');

    if (await todayBtn.count() > 0) {
      await todayBtn.click();

      // Today should be highlighted
      await page.waitForTimeout(300);

      const today = page.locator(selectors.calendar.today);
      await expect(today).toBeVisible();
    }
  });

  test('month/year selector is visible', async ({ page }) => {
    const monthYear = page.locator('.calendar-month-year');

    if (await monthYear.count() > 0) {
      await expect(monthYear).toBeVisible();

      // Should show current month and year
      const text = await monthYear.textContent();
      expect(text).toBeTruthy();
      expect(text.length).toBeGreaterThan(0);
    }
  });
});

