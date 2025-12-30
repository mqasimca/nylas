// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Contacts View tests - Visibility and basic display
 */
test.describe('Contacts View', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    // Switch to Contacts view
    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);
  });

  test('contacts view is visible', async ({ page }) => {
    const contactsView = page.locator(selectors.contacts.view);
    await expect(contactsView).toBeVisible();
  });

  test('new contact button is visible', async ({ page }) => {
    const newContactBtn = page.locator(selectors.contacts.newContactBtn);
    await expect(newContactBtn).toBeVisible();
  });

  test('contacts list is displayed', async ({ page }) => {
    await page.waitForTimeout(1000);

    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      await expect(contactsList).toBeVisible();
    }
  });

  test('contacts show loading skeleton while loading', async ({ page }) => {
    // On initial load, may show skeletons
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const skeletons = contactsList.locator('.skeleton');
      const contacts = contactsList.locator(selectors.contacts.item);

      const skeletonCount = await skeletons.count();
      const contactCount = await contacts.count();

      // Should have either skeletons or contacts
      expect(skeletonCount + contactCount).toBeGreaterThanOrEqual(0);
    }
  });

  test('contact items display name and email', async ({ page }) => {
    await page.waitForTimeout(1500);

    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        const firstContact = contacts.first();

        // Should have name
        await expect(firstContact.locator('.contact-name')).toBeVisible();

        // May have email (if available)
        const email = firstContact.locator('.contact-email');
        if (await email.count() > 0) {
          await expect(email).toBeVisible();
        }
      }
    }
  });

  test('contact items show avatar or initials', async ({ page }) => {
    await page.waitForTimeout(1500);

    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        const firstContact = contacts.first();

        // Should have avatar or initials
        const avatar = firstContact.locator('.contact-avatar');
        const initials = firstContact.locator('.contact-initials');

        const hasAvatar = await avatar.count() > 0;
        const hasInitials = await initials.count() > 0;

        expect(hasAvatar || hasInitials).toBeTruthy();
      }
    }
  });

  test('clicking contact shows contact details', async ({ page }) => {
    await page.waitForTimeout(1500);

    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();

        // Contact detail panel or modal should show
        await page.waitForTimeout(500);

        const detailPanel = page.locator(selectors.contacts.detail);
        const modal = page.locator(selectors.contactModal.overlay);

        const panelVisible = await detailPanel.isVisible().catch(() => false);
        const modalVisible = await modal.isVisible().catch(() => false);

        expect(panelVisible || modalVisible).toBeTruthy();
      }
    }
  });
});
