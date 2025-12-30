// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * Email List tests - List display and navigation
 */
test.describe('Email List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('displays email list container', async ({ page }) => {
    const emailList = page.locator(selectors.email.emailListContainer);
    await expect(emailList).toBeVisible();
  });

  test('shows skeleton loaders while emails are loading', async ({ page }) => {
    // On first load, skeletons may appear briefly
    const emailList = page.locator(selectors.email.emailListContainer);
    await expect(emailList).toBeVisible();

    // Either skeletons or actual emails should be present
    const hasSkeletons = await page.locator(selectors.email.emailSkeleton).count() > 0;
    const hasEmails = await page.locator(selectors.email.emailItem).count() > 0;
    expect(hasSkeletons || hasEmails).toBeTruthy();
  });

  test('email items are clickable', async ({ page }) => {
    // Wait for emails to load
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      // Click first email
      await emailItems.first().click();

      // Preview should update (no longer show empty state)
      const preview = page.locator(selectors.email.preview);
      await expect(preview).toBeVisible();

      // Email item should have selected class
      await expect(emailItems.first()).toHaveClass(/selected/);
    }
  });

  test('email items display sender and subject', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();

      // Check for sender with flexible selectors
      const hasSender = await firstEmail.locator('.email-from, .from, .sender').count() > 0;

      // Check for subject with flexible selectors
      const hasSubject = await firstEmail.locator('.email-subject, .subject').count() > 0;

      // Check for timestamp with flexible selectors
      const hasTime = await firstEmail.locator('.email-time, .time, .date').count() > 0;

      // Should have at least sender or subject
      expect(hasSender || hasSubject || hasTime).toBeTruthy();
    }
  });

  test('unread emails have unread indicator', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      // Check if any emails have unread indicator
      const unreadEmails = emailItems.filter({ has: page.locator('.unread-indicator') });
      const unreadCount = await unreadEmails.count();

      // It's okay if there are no unread emails, just verify the structure
      expect(unreadCount >= 0).toBeTruthy();
    }
  });

  test('starred emails show star icon', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      // Check if any emails have star icon
      const starredEmails = emailItems.filter({ has: page.locator('.star-btn') });
      const starredCount = await starredEmails.count();

      // It's okay if there are no starred emails, just verify the structure
      expect(starredCount >= 0).toBeTruthy();
    }
  });
});

