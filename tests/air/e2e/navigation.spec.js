// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Navigation tests for Nylas Air.
 *
 * Tests view switching between Email, Calendar, Contacts, and Notetaker.
 */

test.describe('View Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Wait for initial load
    await expect(page.locator(selectors.general.app)).toBeVisible();
    // Wait for DOM to be fully ready
    await page.waitForLoadState('domcontentloaded');
    // Ensure page body is focused for keyboard events
    await page.locator('body').click();
    await page.waitForTimeout(200);
  });

  test('can switch from Email to Calendar view', async ({ page }) => {
    // Email should be active initially
    await expect(page.locator(selectors.views.email)).toHaveClass(/active/);

    // Click Calendar tab
    await page.click(selectors.nav.tabCalendar);

    // Calendar view should now be active
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);

    // Email view should no longer be active
    await expect(page.locator(selectors.views.email)).not.toHaveClass(/active/);

    // Calendar tab should be marked active
    await expect(page.locator(selectors.nav.tabCalendar)).toHaveClass(/active/);
  });

  test('can switch from Email to Contacts view', async ({ page }) => {
    // Click Contacts tab
    await page.click(selectors.nav.tabContacts);

    // Contacts view should be active
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);

    // Other views should not be active
    await expect(page.locator(selectors.views.email)).not.toHaveClass(/active/);
    await expect(page.locator(selectors.views.calendar)).not.toHaveClass(/active/);
  });

  test('can switch to Notetaker view', async ({ page }) => {
    // Click Notetaker tab
    await page.click(selectors.nav.tabNotetaker);

    // Notetaker view should be active
    await expect(page.locator(selectors.views.notetaker)).toHaveClass(/active/);
  });

  test('can cycle through all views', async ({ page }) => {
    // Start with Email (default)
    await expect(page.locator(selectors.views.email)).toHaveClass(/active/);

    // Go to Calendar
    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);

    // Go to Contacts
    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);

    // Go to Notetaker
    await page.click(selectors.nav.tabNotetaker);
    await expect(page.locator(selectors.views.notetaker)).toHaveClass(/active/);

    // Back to Email
    await page.click(selectors.nav.tabEmail);
    await expect(page.locator(selectors.views.email)).toHaveClass(/active/);
  });

  test('tab aria-selected attribute updates correctly', async ({ page }) => {
    // Email tab should have aria-selected="true" initially
    await expect(page.locator(selectors.nav.tabEmail)).toHaveAttribute(
      'aria-selected',
      'true'
    );

    // Switch to Calendar
    await page.click(selectors.nav.tabCalendar);

    // Calendar tab should now have aria-selected="true"
    await expect(page.locator(selectors.nav.tabCalendar)).toHaveAttribute(
      'aria-selected',
      'true'
    );

    // Email tab should have aria-selected="false"
    await expect(page.locator(selectors.nav.tabEmail)).toHaveAttribute(
      'aria-selected',
      'false'
    );
  });

  test('keyboard navigation: number keys switch views', async ({ page }) => {
    // Press '2' for Calendar
    await page.keyboard.press('2');
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);

    // Press '3' for Contacts
    await page.keyboard.press('3');
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);

    // Press '1' for Email
    await page.keyboard.press('1');
    await expect(page.locator(selectors.views.email)).toHaveClass(/active/);
  });

  test('keyboard shortcuts do not trigger in input fields', async ({ page }) => {
    // Open compose modal to get an input field
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Focus on the To field
    const toField = page.locator(selectors.compose.to);
    await toField.focus();

    // Type '2' - should NOT switch views, should type in field
    await page.keyboard.type('2');

    // View should still be Email (compose modal is over it)
    await expect(page.locator(selectors.views.email)).toHaveClass(/active/);

    // Field should contain '2'
    await expect(toField).toHaveValue('2');

    // Close modal
    await page.keyboard.press('Escape');
  });

  test('calendar view contains calendar-specific elements', async ({ page }) => {
    await page.click(selectors.nav.tabCalendar);
    await expect(page.locator(selectors.views.calendar)).toHaveClass(/active/);

    // Calendar grid
    await expect(page.locator(selectors.calendar.grid)).toBeVisible();

    // Day headers (Sun, Mon, etc.)
    const headers = page.locator(selectors.calendar.dayHeader);
    await expect(headers).toHaveCount(7);

    // Events panel
    await expect(page.locator(selectors.calendar.eventsPanel)).toBeVisible();

    // New Event button
    await expect(page.locator(selectors.calendar.newEventBtn)).toBeVisible();
  });

  test('contacts view contains contacts-specific elements', async ({ page }) => {
    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);

    // Contacts view is visible
    const contactsView = page.locator(selectors.views.contacts);
    await expect(contactsView).toBeVisible();

    // New Contact button (uses same compose-btn class)
    const newContactBtn = contactsView.locator('.compose-btn');
    await expect(newContactBtn).toBeVisible();
  });
});

test.describe('Search Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.locator('body').click();
    await page.waitForTimeout(200);
  });

  test('clicking search trigger opens search overlay', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);

    // Search overlay should be visible
    const overlay = page.locator(selectors.search.overlay);
    await expect(overlay).toHaveClass(/active/);

    // Search input should be focused
    await expect(page.locator(selectors.search.input)).toBeFocused();
  });

  test('Cmd+K opens command palette', async ({ page }) => {
    // Press Cmd+K (Meta+k on Mac)
    await page.keyboard.press('Meta+k');

    // Command palette should be visible
    const palette = page.locator(selectors.commandPalette.overlay);
    await expect(palette).not.toHaveClass(/hidden/);

    // Input should be present
    await expect(page.locator(selectors.commandPalette.input)).toBeVisible();
  });

  test('Escape closes search overlay', async ({ page }) => {
    // Open search
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    // Press Escape
    await page.keyboard.press('Escape');

    // Search should be closed
    await expect(page.locator(selectors.search.overlay)).not.toHaveClass(
      /active/
    );
  });

  test('search filter chips are interactive', async ({ page }) => {
    // Open search
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    // All filter should be active by default
    const allFilter = page.locator(selectors.search.filterChip).filter({
      hasText: 'All',
    });
    await expect(allFilter).toHaveClass(/active/);

    // Click Emails filter
    const emailsFilter = page.locator(selectors.search.filterChip).filter({
      hasText: 'Emails',
    });
    await emailsFilter.click();

    // Emails filter should now be active
    await expect(emailsFilter).toHaveClass(/active/);

    // All filter should no longer be active
    await expect(allFilter).not.toHaveClass(/active/);
  });
});
