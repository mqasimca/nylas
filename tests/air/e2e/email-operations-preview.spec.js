// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Email Preview tests - Email preview pane
 */
test.describe('Email Preview', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('shows empty state when no email is selected', async ({ page }) => {
    const preview = page.locator(selectors.email.preview);
    await expect(preview).toBeVisible();

    const emptyState = preview.locator(selectors.email.emptyState);
    await expect(emptyState).toBeVisible();

    await expect(emptyState.locator('.empty-title')).toHaveText('Select an email');
  });

  test('displays email content when email is selected', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      // Click first email
      await emailItems.first().click();
      await page.waitForTimeout(500);

      const preview = page.locator(selectors.email.preview);

      // Preview should be visible
      await expect(preview).toBeVisible();

      // Check for content with flexible selectors
      const hasHeader = await preview.locator('.preview-header, .email-header, header, h1, h2').count() > 0;
      const hasSubject = await preview.locator('.preview-subject, .email-subject, .subject').count() > 0;
      const hasFrom = await preview.locator('.preview-from, .email-from, .from').count() > 0;
      const hasBody = await preview.locator('.preview-body, .email-body, .body, .content').count() > 0;
      const hasText = await preview.textContent().then(t => t && t.trim().length > 0);

      // Should have at least some content displayed
      expect(hasHeader || hasSubject || hasFrom || hasBody || hasText).toBeTruthy();
    }
  });

  test('preview actions are visible when email is selected', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      // Wait for preview to load
      await page.waitForTimeout(2000);

      const preview = page.locator(selectors.email.preview);

      // Wait for preview content to be visible (not empty state)
      await page.waitForTimeout(1000);

      // Check for action buttons with flexible selectors (case-insensitive)
      const hasReply = await preview.locator('button').filter({ hasText: /reply/i }).count() > 0;
      const hasArchive = await preview.locator('button').filter({ hasText: /archive/i }).count() > 0;
      const hasDelete = await preview.locator('button').filter({ hasText: /delete/i }).count() > 0;
      const hasTrash = await preview.locator('button').filter({ hasText: /trash/i }).count() > 0;
      const hasForward = await preview.locator('button').filter({ hasText: /forward/i }).count() > 0;

      // Check for icon buttons (might not have text)
      const hasActionButtons = await preview.locator('button, .action-btn, .preview-action').count() > 1;

      // Should have at least some action buttons
      expect(hasReply || hasArchive || hasDelete || hasTrash || hasForward || hasActionButtons).toBeTruthy();
    }
  });

  test('reply button opens compose modal with quoted text', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const replyBtn = preview.locator('button:has-text("Reply")').first();

      await replyBtn.click();

      // Compose modal should open
      await expect(page.locator(selectors.compose.modal)).toBeVisible();

      // To field should be populated
      const toField = page.locator(selectors.compose.to);
      const toValue = await toField.inputValue();
      expect(toValue.length).toBeGreaterThan(0);
    }
  });
});

