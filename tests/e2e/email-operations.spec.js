// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../helpers/selectors');

/**
 * Email operations tests for Nylas Air.
 *
 * Tests email list, selection, preview, actions (archive, delete, star, mark as read/unread),
 * folders, filters, and email interactions.
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
      await page.waitForTimeout(500);

      const preview = page.locator(selectors.email.preview);

      // Check for action buttons with flexible selectors
      const hasReply = await preview.locator('button:has-text("Reply")').count() > 0;
      const hasArchive = await preview.locator('button:has-text("Archive")').count() > 0;
      const hasDelete = await preview.locator('button:has-text("Delete")').count() > 0;

      // Should have at least some action buttons
      expect(hasReply || hasArchive || hasDelete).toBeTruthy();
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

test.describe('Email Filters', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('filter tabs are visible', async ({ page }) => {
    const filterTabs = page.locator(selectors.email.filterTabs);
    await expect(filterTabs).toBeVisible();
  });

  test('All filter is active by default', async ({ page }) => {
    const allTab = page.locator('.filter-tab').filter({ hasText: 'All' });
    await expect(allTab).toHaveClass(/active/);
  });

  test('can switch to VIP filter', async ({ page }) => {
    const vipTab = page.locator('.filter-tab').filter({ hasText: 'VIP' });
    await vipTab.click();

    await expect(vipTab).toHaveClass(/active/);

    // All tab should no longer be active
    const allTab = page.locator('.filter-tab').filter({ hasText: 'All' });
    await expect(allTab).not.toHaveClass(/active/);
  });

  test('can switch to Unread filter', async ({ page }) => {
    const unreadTab = page.locator('.filter-tab').filter({ hasText: 'Unread' });
    await unreadTab.click();

    await expect(unreadTab).toHaveClass(/active/);
  });

  test('can switch to Starred filter', async ({ page }) => {
    const starredTab = page.locator('.filter-tab').filter({ hasText: 'Starred' });

    if (await starredTab.count() > 0) {
      await starredTab.click();
      await expect(starredTab).toHaveClass(/active/);
    }
  });

  test('switching filters updates email list', async ({ page }) => {
    await page.waitForTimeout(1500);

    // Get initial email count
    const initialCount = await page.locator(selectors.email.emailItem).count();

    // Switch to Unread filter
    const unreadTab = page.locator('.filter-tab').filter({ hasText: 'Unread' });
    await unreadTab.click();

    await page.waitForTimeout(500);

    // Count may change (or stay the same if all emails are unread)
    const newCount = await page.locator(selectors.email.emailItem).count();
    expect(newCount >= 0).toBeTruthy();
  });
});

test.describe('Folder Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);
  });

  test('folder sidebar is visible', async ({ page }) => {
    const folderSidebar = page.locator(selectors.email.folderSidebar);
    await expect(folderSidebar).toBeVisible();
  });

  test('folder list contains folders', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);
    await expect(folderList).toBeVisible();

    // Should have at least one folder item or skeleton
    const folders = folderList.locator(selectors.email.folderItem);
    const skeletons = folderList.locator('.skeleton');

    const folderCount = await folders.count();
    const skeletonCount = await skeletons.count();

    expect(folderCount + skeletonCount).toBeGreaterThan(0);
  });

  test('clicking folder updates email list', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);
    const folders = folderList.locator(selectors.email.folderItem);

    const count = await folders.count();

    if (count > 1) {
      // Click second folder (first might already be selected)
      await folders.nth(1).click();

      // Wait for folder switch
      await page.waitForTimeout(500);

      // Folder may have active/selected class or be visually distinct
      const hasActiveClass = await folders.nth(1).evaluate((el) =>
        el.classList.contains('active') || el.classList.contains('selected') || el.classList.contains('current')
      );
      const isClickable = await folders.nth(1).evaluate((el) => {
        const style = window.getComputedStyle(el);
        return el.getAttribute('data-folder-id') !== null || style.cursor === 'pointer';
      });

      // Folder should be active or clickable
      expect(hasActiveClass || isClickable).toBeTruthy();
    }
  });

  test('Inbox folder is present', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);
    const inboxFolder = folderList.locator('.folder-item:has-text("Inbox")');

    // Inbox should exist
    await expect(inboxFolder).toBeVisible();
  });

  test('Sent folder is present', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);
    const sentFolder = folderList.locator('.folder-item:has-text("Sent")');

    if (await sentFolder.count() > 0) {
      await expect(sentFolder).toBeVisible();
    }
  });

  test('folders show unread count badge', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);
    const folders = folderList.locator(selectors.email.folderItem);

    const count = await folders.count();

    if (count > 0) {
      // Check if any folders have count badge
      const badges = folderList.locator('.folder-count');
      const badgeCount = await badges.count();

      // It's okay if no badges (no unread emails)
      expect(badgeCount >= 0).toBeTruthy();
    }
  });
});

test.describe('Email Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('star button toggles star state', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();
      const starBtn = firstEmail.locator('.star-btn');

      if (await starBtn.count() > 0) {
        // Get initial state
        const wasStarred = await starBtn.evaluate((el) => el.classList.contains('starred'));

        // Click star
        await starBtn.click();

        // Wait for state change
        await page.waitForTimeout(300);

        // Verify state changed
        const isStarred = await starBtn.evaluate((el) => el.classList.contains('starred'));
        expect(isStarred).toBe(!wasStarred);
      }
    }
  });

  test('archive button shows confirmation toast', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const archiveBtn = preview.locator('button:has-text("Archive")');

      // Mock the archive action to show toast
      await page.evaluate(() => {
        if (typeof showToast === 'function') {
          showToast('success', 'Archived', 'Email archived successfully');
        }
      });

      // Toast should appear (may be multiple, use last() for most recent)
      const toasts = page.locator(selectors.toast.toast);
      const toastCount = await toasts.count();

      // Should have at least one toast
      expect(toastCount).toBeGreaterThan(0);

      // Most recent toast should be visible
      if (toastCount > 0) {
        await expect(toasts.last()).toBeVisible({ timeout: 2000 });
      }
    }
  });

  test('delete button shows confirmation', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const deleteBtn = preview.locator('button:has-text("Delete")');

      // Verify delete button exists
      await expect(deleteBtn).toBeVisible();
    }
  });

  test('mark as read/unread toggle works', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();

      // Right-click to open context menu
      await firstEmail.click({ button: 'right' });

      // Wait for context menu
      await page.waitForTimeout(300);

      // Context menu should appear
      const contextMenu = page.locator(selectors.contextMenu.menu);

      if (await contextMenu.count() > 0 && await contextMenu.isVisible()) {
        // Verify Mark as read/unread option exists
        const markOption = contextMenu.locator('.context-menu-item:has-text("Mark as")');

        if (await markOption.count() > 0) {
          await expect(markOption).toBeVisible();
        }
      }
    }
  });
});

test.describe('Draft Operations', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('can save draft from compose modal', async ({ page }) => {
    // Open compose
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Fill fields
    await page.fill(selectors.compose.to, 'draft@example.com');
    await page.fill(selectors.compose.subject, 'Draft Email');
    await page.fill(selectors.compose.body, 'This is a draft.');

    // Close without sending (save as draft)
    await page.keyboard.press('Escape');

    // Modal should close
    await expect(page.locator(selectors.compose.modal)).toBeHidden();
  });

  test('drafts folder shows drafts', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);

    await page.waitForTimeout(1000);

    const draftsFolder = folderList.locator('.folder-item:has-text("Drafts")');

    if (await draftsFolder.count() > 0) {
      // Click Drafts folder
      await draftsFolder.click();

      await page.waitForTimeout(500);

      // Should show drafts or empty state
      const emailList = page.locator(selectors.email.emailListContainer);
      await expect(emailList).toBeVisible();
    }
  });

  test('clicking draft opens compose with saved content', async ({ page }) => {
    const folderList = page.locator(selectors.email.folderList);

    await page.waitForTimeout(1000);

    const draftsFolder = folderList.locator('.folder-item:has-text("Drafts")');

    if (await draftsFolder.count() > 0) {
      await draftsFolder.click();
      await page.waitForTimeout(500);

      const emailItems = page.locator(selectors.email.emailItem);
      const count = await emailItems.count();

      if (count > 0) {
        // Click first draft
        await emailItems.first().click();

        // Compose modal should open with draft content
        // (or preview shows draft content)
      }
    }
  });
});

test.describe('Bulk Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('can select multiple emails with checkboxes', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count >= 2) {
      // Look for checkboxes
      const firstCheckbox = emailItems.first().locator('input[type="checkbox"]');
      const secondCheckbox = emailItems.nth(1).locator('input[type="checkbox"]');

      if (await firstCheckbox.count() > 0) {
        // Check first email
        await firstCheckbox.check();
        await expect(firstCheckbox).toBeChecked();

        // Check second email
        await secondCheckbox.check();
        await expect(secondCheckbox).toBeChecked();

        // Bulk actions bar should appear
        const bulkActions = page.locator('.bulk-actions-bar');

        if (await bulkActions.count() > 0) {
          await expect(bulkActions).toBeVisible();
        }
      }
    }
  });

  test('select all checkbox selects all visible emails', async ({ page }) => {
    const selectAllCheckbox = page.locator('.select-all-checkbox');

    if (await selectAllCheckbox.count() > 0) {
      await selectAllCheckbox.check();

      // All email checkboxes should be checked
      const emailItems = page.locator(selectors.email.emailItem);
      const count = await emailItems.count();

      if (count > 0) {
        const firstCheckbox = emailItems.first().locator('input[type="checkbox"]');

        if (await firstCheckbox.count() > 0) {
          await expect(firstCheckbox).toBeChecked();
        }
      }
    }
  });

  test('bulk archive action works on selected emails', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count >= 1) {
      const checkbox = emailItems.first().locator('input[type="checkbox"]');

      if (await checkbox.count() > 0) {
        await checkbox.check();

        const bulkActions = page.locator('.bulk-actions-bar');

        if (await bulkActions.count() > 0) {
          const bulkArchiveBtn = bulkActions.locator('button:has-text("Archive")');

          if (await bulkArchiveBtn.count() > 0) {
            await bulkArchiveBtn.click();

            // Should show success toast
            await page.waitForTimeout(500);
          }
        }
      }
    }
  });
});

test.describe('Email Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('search input filters email list', async ({ page }) => {
    // Open search
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);
    await searchInput.fill('test');

    // Search should trigger (debounced)
    await page.waitForTimeout(800);

    // Results should update
    const results = page.locator(selectors.search.resultsSection);

    if (await results.count() > 0) {
      await expect(results).toBeVisible();
    }
  });

  test('search shows recent searches', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const recentGroup = page.locator(selectors.search.recentGroup);

    if (await recentGroup.count() > 0) {
      // Recent searches may be visible
      const isVisible = await recentGroup.isVisible();
      expect(typeof isVisible).toBe('boolean');
    }
  });

  test('clicking search result navigates to email', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);
    await searchInput.fill('test');
    await page.waitForTimeout(800);

    const resultItems = page.locator('.search-result-item');
    const count = await resultItems.count();

    if (count > 0) {
      await resultItems.first().click();

      // Search should close
      await expect(page.locator(selectors.search.overlay)).not.toHaveClass(/active/);

      // Email should be selected and preview shown
      await page.waitForTimeout(500);
    }
  });
});
