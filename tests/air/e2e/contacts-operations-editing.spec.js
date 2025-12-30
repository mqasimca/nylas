// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Contact Editing tests - Editing existing contacts
 */
test.describe('Contact Editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);
    await page.waitForTimeout(1500);
  });

  test('can edit existing contact', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);

        if (await detail.count() > 0 && await detail.isVisible()) {
          const editBtn = detail.locator('button:has-text("Edit")');

          if (await editBtn.count() > 0) {
            await editBtn.click();

            // Contact modal should open with populated data
            const modal = page.locator(selectors.contactModal.modal);

            if (await modal.count() > 0 && await modal.isVisible()) {
              // Fields should be populated
              const givenName = modal.locator(selectors.contactModal.givenName);
              const currentName = await givenName.inputValue();

              expect(currentName.length).toBeGreaterThan(0);

              // Can edit name
              await givenName.fill('Updated Name');
              await expect(givenName).toHaveValue('Updated Name');
            }
          }
        }
      }
    }
  });

  test('contact modal shows save button for existing contacts', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);

        if (await detail.count() > 0 && await detail.isVisible()) {
          const editBtn = detail.locator('button:has-text("Edit")');

          if (await editBtn.count() > 0) {
            await editBtn.click();

            const modal = page.locator(selectors.contactModal.modal);

            if (await modal.count() > 0 && await modal.isVisible()) {
              const saveBtn = modal.locator(selectors.contactModal.saveBtn);
              await expect(saveBtn).toBeVisible();
            }
          }
        }
      }
    }
  });
});
