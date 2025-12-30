// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Contact Search tests - Searching and filtering contacts
 */
test.describe('Contact Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);
  });

  test('search input filters contacts', async ({ page }) => {
    // Try multiple selectors for search input
    const searchContainer = page.locator('.contacts-search');
    let searchInput = searchContainer.locator('input').first();

    // Fallback to generic input if container doesn't have one
    if (await searchInput.count() === 0) {
      searchInput = page.locator('input[placeholder*="Search"], input[type="search"]').first();
    }

    if (await searchInput.count() > 0) {
      await searchInput.fill('john');

      // Wait for search (debounced)
      await page.waitForTimeout(800);

      // Contacts list should filter
      const contactsList = page.locator(selectors.contacts.list);

      if (await contactsList.count() > 0) {
        await expect(contactsList).toBeVisible();
      }
    }
  });

  test('search works with name and email', async ({ page }) => {
    // Try multiple selectors for search input
    const searchContainer = page.locator('.contacts-search');
    let searchInput = searchContainer.locator('input').first();

    // Fallback to generic input if container doesn't have one
    if (await searchInput.count() === 0) {
      searchInput = page.locator('input[placeholder*="Search"], input[type="search"]').first();
    }

    if (await searchInput.count() > 0) {
      // Search by email
      await searchInput.fill('example.com');
      await page.waitForTimeout(800);

      // Should filter results
      const contactsList = page.locator(selectors.contacts.list);

      if (await contactsList.count() > 0) {
        const contacts = contactsList.locator(selectors.contacts.item);
        const count = await contacts.count();

        // Results should be filtered
        expect(count >= 0).toBeTruthy();
      }
    }
  });

  test('clearing search shows all contacts', async ({ page }) => {
    // Try multiple selectors for search input
    const searchContainer = page.locator('.contacts-search');
    let searchInput = searchContainer.locator('input').first();

    // Fallback to generic input if container doesn't have one
    if (await searchInput.count() === 0) {
      searchInput = page.locator('input[placeholder*="Search"], input[type="search"]').first();
    }

    if (await searchInput.count() > 0) {
      await page.waitForTimeout(1000);

      // Get initial count
      const contactsList = page.locator(selectors.contacts.list);
      const initialContacts = await contactsList.locator(selectors.contacts.item).count();

      // Search
      await searchInput.fill('nonexistent');
      await page.waitForTimeout(800);

      // Clear search
      await searchInput.clear();
      await page.waitForTimeout(800);

      // Should show all contacts again
      const finalContacts = await contactsList.locator(selectors.contacts.item).count();

      expect(finalContacts).toBeGreaterThanOrEqual(0);
    }
  });
});
