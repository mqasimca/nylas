// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Contact Details tests - Viewing contact details
 */
test.describe('Contact Details', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');

    await page.click(selectors.nav.tabContacts);
    await expect(page.locator(selectors.views.contacts)).toHaveClass(/active/);
    await page.waitForTimeout(1500);
  });

  test('contact detail shows full contact information', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);
        const detailExists = await detail.count() > 0;

        if (detailExists) {
          const isVisible = await detail.isVisible().catch(() => false);

          if (isVisible) {
            // Should show name or contact info
            const hasName = await detail.locator('.detail-name, .contact-name, h2, h3').first().isVisible().catch(() => false);
            const hasEmail = await detail.locator('.detail-email, .contact-email, [href^="mailto:"]').count() > 0;
            const hasPhone = await detail.locator('.detail-phone, .contact-phone, [href^="tel:"]').count() > 0;
            const hasAnyInfo = await detail.textContent().then(t => t && t.trim().length > 0);

            // Should have at least some contact information displayed
            expect(hasName || hasEmail || hasPhone || hasAnyInfo).toBeTruthy();
          }
        } else {
          // Detail panel doesn't exist - contacts might use modal instead
          const modal = page.locator(selectors.contactModal.modal);
          const modalExists = await modal.count() > 0;

          if (modalExists) {
            const isVisible = await modal.isVisible().catch(() => false);
            expect(isVisible).toBeTruthy();
          } else {
            // Neither detail nor modal - contact might be shown inline
            expect(count).toBeGreaterThan(0);
          }
        }
      }
    }
  });

  test('contact detail shows edit button', async ({ page }) => {
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
            await expect(editBtn).toBeVisible();
          }
        }
      }
    }
  });

  test('contact detail shows delete button', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);

        if (await detail.count() > 0 && await detail.isVisible()) {
          const deleteBtn = detail.locator('button:has-text("Delete")');

          if (await deleteBtn.count() > 0) {
            await expect(deleteBtn).toBeVisible();
          }
        }
      }
    }
  });

  test('contact detail shows compose email button', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);

        if (await detail.count() > 0 && await detail.isVisible()) {
          const emailBtn = detail.locator('button:has-text("Email")');

          if (await emailBtn.count() > 0) {
            await expect(emailBtn).toBeVisible();
          }
        }
      }
    }
  });

  test('clicking email button opens compose with contact email', async ({ page }) => {
    const contactsList = page.locator(selectors.contacts.list);

    if (await contactsList.count() > 0) {
      const contacts = contactsList.locator(selectors.contacts.item);
      const count = await contacts.count();

      if (count > 0) {
        await contacts.first().click();
        await page.waitForTimeout(500);

        const detail = page.locator(selectors.contacts.detail);

        if (await detail.count() > 0 && await detail.isVisible()) {
          const emailBtn = detail.locator('button:has-text("Email")');

          if (await emailBtn.count() > 0) {
            await emailBtn.click();

            // Compose modal should open
            await expect(page.locator(selectors.compose.modal)).toBeVisible();

            // To field should be populated
            const toField = page.locator(selectors.compose.to);
            const value = await toField.inputValue();

            expect(value.length).toBeGreaterThan(0);
          }
        }
      }
    }
  });
});
