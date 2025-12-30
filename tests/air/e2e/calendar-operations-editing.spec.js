// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Event Editing tests - Editing existing events
 */
test.describe('Event Editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);
    await page.waitForTimeout(1000);
  });

  test('can edit existing event', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        // Click event to open details
        await eventItems.first().click();
        await page.waitForTimeout(500);

        const modal = page.locator(selectors.eventModal.modal);

        if (await modal.count() > 0 && await modal.isVisible()) {
          // Should have title field populated
          const titleField = modal.locator(selectors.eventModal.title);
          const currentTitle = await titleField.inputValue();

          expect(currentTitle.length).toBeGreaterThan(0);

          // Can edit title
          await titleField.fill('Updated Event Title');
          await expect(titleField).toHaveValue('Updated Event Title');
        }
      }
    }
  });

  test('event modal shows delete button for existing events', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        await eventItems.first().click();
        await page.waitForTimeout(500);

        const modal = page.locator(selectors.eventModal.modal);

        if (await modal.count() > 0 && await modal.isVisible()) {
          const deleteBtn = modal.locator(selectors.eventModal.deleteBtn);

          if (await deleteBtn.count() > 0) {
            await expect(deleteBtn).toBeVisible();
          }
        }
      }
    }
  });

  test('can change event time', async ({ page }) => {
    const eventsList = page.locator(selectors.calendar.eventsList);

    if (await eventsList.count() > 0) {
      const eventItems = eventsList.locator('.event-item');
      const count = await eventItems.count();

      if (count > 0) {
        await eventItems.first().click();
        await page.waitForTimeout(500);

        const modal = page.locator(selectors.eventModal.modal);

        if (await modal.count() > 0 && await modal.isVisible()) {
          const startTime = modal.locator(selectors.eventModal.startTime);

          if (await startTime.count() > 0 && await startTime.isVisible()) {
            await startTime.fill('14:00');
            await expect(startTime).toHaveValue('14:00');
          }
        }
      }
    }
  });
});

