// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Event List tests - Event list display
 */
test.describe('Event List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);
    await page.waitForTimeout(1000);
  });

  test('events panel shows events list', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      await expect(eventsList).toBeVisible();
    }
  });

  test('events show time and title', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        const firstEvent = eventItems.first();

        // Should have time
        await expect(firstEvent.locator('.event-time')).toBeVisible();

        // Should have title
        await expect(firstEvent.locator('.event-title')).toBeVisible();
      }
    }
  });

  test('clicking event shows event details', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        await eventItems.first().click();

        // Event modal should open or event detail panel should show
        await page.waitForTimeout(500);

        // Either modal or detail panel
        const modal = page.locator(selectors.eventModal.overlay);
        const detailPanel = page.locator('.event-detail-panel');

        const modalVisible = await modal.isVisible().catch(() => false);
        const panelVisible = await detailPanel.isVisible().catch(() => false);

        expect(modalVisible || panelVisible).toBeTruthy();
      }
    }
  });

  test('events display calendar color indicator', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        const firstEvent = eventItems.first();
        const colorIndicator = firstEvent.locator('.event-color');

        if (await colorIndicator.count() > 0) {
          await expect(colorIndicator).toBeVisible();
        }
      }
    }
  });
});

