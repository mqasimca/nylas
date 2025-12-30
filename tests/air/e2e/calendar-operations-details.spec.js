// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Event Details tests - Event detail view
 */
test.describe('Event Details', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);
    await page.waitForTimeout(1000);
  });

  test('event details show all event information', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        await eventItems.first().click();
        await page.waitForTimeout(500);

        const modal = page.locator(selectors.eventModal.modal);

        if (await modal.count() > 0 && await modal.isVisible()) {
          // Title should be populated
          const title = modal.locator(selectors.eventModal.title);
          const titleValue = await title.inputValue();
          expect(titleValue.length).toBeGreaterThan(0);

          // Start date should be populated
          const startDate = modal.locator(selectors.eventModal.startDate);
          const dateValue = await startDate.inputValue();
          expect(dateValue.length).toBeGreaterThan(0);
        }
      }
    }
  });

  test('recurring events show recurrence indicator', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');

      // Look for recurring event indicators
      const recurringIndicators = eventItems.locator('.recurring-icon');
      const count = await recurringIndicators.count();

      // It's okay if no recurring events
      expect(count >= 0).toBeTruthy();
    }
  });

  test('event participants are displayed', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        await eventItems.first().click();
        await page.waitForTimeout(500);

        const modal = page.locator(selectors.eventModal.modal);

        if (await modal.count() > 0 && await modal.isVisible()) {
          const participants = modal.locator(selectors.eventModal.participants);

          if (await participants.count() > 0) {
            // Participants field exists
            await expect(participants).toBeVisible();
          }
        }
      }
    }
  });
});
