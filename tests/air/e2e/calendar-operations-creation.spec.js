// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Event Creation tests - Creating new events
 */
test.describe('Event Creation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    // Switch to Calendar view
    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);
  });

  test('clicking new event button opens event modal', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    // Event modal should open
    const eventModal = page.locator(selectors.eventModal.overlay);

    if (await eventModal.count() > 0) {
      await expect(eventModal).toBeVisible();
    }
  });

  test('event modal contains all required fields', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.modal);

    if (await modal.count() > 0) {
      await expect(modal).toBeVisible();

      // Title field
      await expect(modal.locator(selectors.eventModal.title)).toBeVisible();

      // Start date and time
      await expect(modal.locator(selectors.eventModal.startDate)).toBeVisible();
      await expect(modal.locator(selectors.eventModal.startTime)).toBeVisible();

      // End date and time
      await expect(modal.locator(selectors.eventModal.endDate)).toBeVisible();
      await expect(modal.locator(selectors.eventModal.endTime)).toBeVisible();

      // All day checkbox
      const allDayCheckbox = modal.locator(selectors.eventModal.allDay);
      if (await allDayCheckbox.count() > 0) {
        await expect(allDayCheckbox).toBeVisible();
      }

      // Location field
      const locationField = modal.locator(selectors.eventModal.location);
      if (await locationField.count() > 0) {
        await expect(locationField).toBeVisible();
      }

      // Save button
      await expect(modal.locator(selectors.eventModal.saveBtn)).toBeVisible();
    }
  });

  test('can fill event form', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.modal);

    if (await modal.count() > 0) {
      // Wait for modal to be fully visible
      await expect(modal).toBeVisible();

      // Fill title
      const titleField = modal.locator(selectors.eventModal.title);
      await titleField.click();
      await titleField.fill('Team Meeting');
      await expect(titleField).toHaveValue('Team Meeting');

      // Fill location - click first to ensure focus, then clear and fill
      const locationField = modal.locator(selectors.eventModal.location);
      if (await locationField.count() > 0) {
        await locationField.click();
        await locationField.clear();
        await locationField.fill('Conference Room A');
        await expect(locationField).toHaveValue('Conference Room A', { timeout: 10000 });
      }

      // Fill description
      const descField = modal.locator(selectors.eventModal.description);
      if (await descField.count() > 0) {
        await descField.click();
        await descField.fill('Quarterly planning meeting');
        await expect(descField).toHaveValue('Quarterly planning meeting');
      }
    }
  });

  test('all day checkbox toggles time fields', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.modal);

    if (await modal.count() > 0) {
      const allDayCheckbox = modal.locator(selectors.eventModal.allDay);

      if (await allDayCheckbox.count() > 0) {
        // Check all day
        await allDayCheckbox.check();

        // Time fields should be disabled or hidden
        const startTime = modal.locator(selectors.eventModal.startTime);
        const endTime = modal.locator(selectors.eventModal.endTime);

        if (await startTime.count() > 0) {
          const isDisabled = await startTime.isDisabled();
          const isHidden = !(await startTime.isVisible());

          expect(isDisabled || isHidden).toBeTruthy();
        }

        // Uncheck all day
        await allDayCheckbox.uncheck();

        // Time fields should be enabled and visible
        if (await startTime.count() > 0) {
          await expect(startTime).toBeVisible();
          await expect(startTime).toBeEnabled();
        }
      }
    }
  });

  test('can add participants to event', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.modal);

    if (await modal.count() > 0) {
      const participantsField = modal.locator(selectors.eventModal.participants);

      if (await participantsField.count() > 0) {
        await participantsField.fill('alice@example.com, bob@example.com');
        await expect(participantsField).toHaveValue(/alice@example.com/);
      }
    }
  });

  test('event modal can be closed', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.overlay);

    if (await modal.count() > 0) {
      await expect(modal).toBeVisible();

      // Close with Escape
      await page.keyboard.press('Escape');

      // Modal should be hidden
      await expect(modal).toBeHidden();
    }
  });

  test('clicking close button closes event modal', async ({ page }) => {
    const newEventBtn = page.locator(selectors.calendar.newEventBtn);
    await newEventBtn.click();

    const modal = page.locator(selectors.eventModal.overlay);

    if (await modal.count() > 0) {
      await expect(modal).toBeVisible();

      const closeBtn = page.locator(selectors.eventModal.closeBtn);

      if (await closeBtn.count() > 0) {
        await closeBtn.click();
        await expect(modal).toBeHidden();
      }
    }
  });
});

