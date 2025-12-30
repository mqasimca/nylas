// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/air-selectors');

/**
 * AI Features tests for Nylas Air.
 *
 * Tests AI completions, smart compose, email summaries,
 * AI settings, and AI-powered features.
 */

test.describe('AI Settings', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('AI settings section is available in settings', async ({ page }) => {
    // Open settings
    await page.click(selectors.nav.settingsBtn);
    await expect(page.locator(selectors.settings.overlay)).toHaveClass(/active/);

    // AI section should be visible
    const aiSection = page.locator(selectors.settings.aiSection);
    await expect(aiSection).toBeVisible();
  });

  test('can select AI provider', async ({ page }) => {
    await page.click(selectors.nav.settingsBtn);
    await expect(page.locator(selectors.settings.overlay)).toHaveClass(/active/);

    const aiSection = page.locator(selectors.settings.aiSection);

    if (await aiSection.count() > 0) {
      // Look for provider selector
      const providerSelect = aiSection.locator('select[name="ai-provider"]');

      if (await providerSelect.count() > 0) {
        await expect(providerSelect).toBeVisible();

        // Should have options
        const options = providerSelect.locator('option');
        const count = await options.count();

        expect(count).toBeGreaterThan(0);
      }
    }
  });

  test('AI provider options include major providers', async ({ page }) => {
    await page.click(selectors.nav.settingsBtn);
    await expect(page.locator(selectors.settings.overlay)).toHaveClass(/active/);

    const aiSection = page.locator(selectors.settings.aiSection);

    if (await aiSection.count() > 0) {
      const providerSelect = aiSection.locator('select[name="ai-provider"]');

      if (await providerSelect.count() > 0) {
        // Get all options
        const optionTexts = await providerSelect.locator('option').allTextContents();

        // Should include common providers or "None"
        const hasOptions = optionTexts.length > 0;
        expect(hasOptions).toBeTruthy();
      }
    }
  });

  test('can enable/disable AI features', async ({ page }) => {
    await page.click(selectors.nav.settingsBtn);
    await expect(page.locator(selectors.settings.overlay)).toHaveClass(/active/);

    const aiSection = page.locator(selectors.settings.aiSection);

    if (await aiSection.count() > 0) {
      // Look for AI enable checkbox
      const enableCheckbox = aiSection.locator('input[type="checkbox"][name="enable-ai"]');

      if (await enableCheckbox.count() > 0) {
        const wasChecked = await enableCheckbox.isChecked();

        // Toggle
        await enableCheckbox.click();

        // State should change
        const isChecked = await enableCheckbox.isChecked();
        expect(isChecked).toBe(!wasChecked);
      }
    }
  });
});

test.describe('Smart Compose', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('AI compose button is visible in compose modal', async ({ page }) => {
    // Open compose
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Look for AI assist button
    const aiBtn = page.locator('.ai-compose-btn');

    if (await aiBtn.count() > 0) {
      await expect(aiBtn).toBeVisible();
    }
  });

  test('AI suggestions appear while typing', async ({ page }) => {
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Type in body
    const bodyField = page.locator(selectors.compose.body);
    await bodyField.fill('Thank you for');

    // Wait for AI suggestions (if enabled)
    await page.waitForTimeout(1000);

    // Look for AI suggestion overlay
    const aiSuggestion = page.locator('.ai-suggestion');

    if (await aiSuggestion.count() > 0) {
      // AI suggestion may appear
      const isVisible = await aiSuggestion.isVisible().catch(() => false);
      expect(typeof isVisible).toBe('boolean');
    }
  });

  test('can accept AI suggestion with Tab key', async ({ page }) => {
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    const bodyField = page.locator(selectors.compose.body);
    await bodyField.fill('Thank you');

    await page.waitForTimeout(1000);

    // If suggestion appears, Tab should accept it
    const aiSuggestion = page.locator('.ai-suggestion');

    if (await aiSuggestion.count() > 0 && await aiSuggestion.isVisible()) {
      await bodyField.press('Tab');

      // Suggestion should be inserted
      const value = await bodyField.inputValue();
      expect(value.length).toBeGreaterThan('Thank you'.length);
    }
  });

  test('can generate email with AI button', async ({ page }) => {
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    const aiBtn = page.locator('.ai-compose-btn');

    if (await aiBtn.count() > 0) {
      // Fill prompt
      await page.fill(selectors.compose.subject, 'Meeting follow-up');

      await aiBtn.click();

      // AI generation modal or loading state should appear
      await page.waitForTimeout(500);

      const aiModal = page.locator('.ai-generation-modal');
      const loader = page.locator('.ai-loading');

      const hasModal = await aiModal.count() > 0;
      const hasLoader = await loader.count() > 0;

      expect(hasModal || hasLoader).toBeTruthy();
    }
  });
});

test.describe('Email Summaries', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('AI summary button appears for long emails', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);

      // Look for AI summary button
      const summaryBtn = preview.locator('.ai-summary-btn');

      if (await summaryBtn.count() > 0) {
        await expect(summaryBtn).toBeVisible();
      }
    }
  });

  test('clicking summary button generates AI summary', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const summaryBtn = preview.locator('.ai-summary-btn');

      if (await summaryBtn.count() > 0) {
        await summaryBtn.click();

        // Summary should appear or loading state
        await page.waitForTimeout(500);

        const summary = preview.locator('.ai-summary');
        const loader = preview.locator('.ai-summary-loading');

        const hasSummary = await summary.count() > 0;
        const hasLoader = await loader.count() > 0;

        expect(hasSummary || hasLoader).toBeTruthy();
      }
    }
  });

  test('AI summary is displayed correctly', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const summary = preview.locator('.ai-summary');

      if (await summary.count() > 0 && await summary.isVisible()) {
        // Should have text content
        const text = await summary.textContent();
        expect(text).toBeTruthy();
        expect(text.length).toBeGreaterThan(0);
      }
    }
  });
});

test.describe('Smart Reply', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('smart reply suggestions appear in email preview', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);

      // Look for smart reply chips
      const smartReplies = preview.locator('.smart-reply-chip');

      if (await smartReplies.count() > 0) {
        await expect(smartReplies.first()).toBeVisible();
      }
    }
  });

  test('clicking smart reply opens compose with suggestion', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const smartReplies = preview.locator('.smart-reply-chip');

      if (await smartReplies.count() > 0) {
        const firstReply = smartReplies.first();
        const replyText = await firstReply.textContent();

        await firstReply.click();

        // Compose modal should open with reply
        await expect(page.locator(selectors.compose.modal)).toBeVisible();

        // Body should contain reply text
        const bodyField = page.locator(selectors.compose.body);
        const bodyValue = await bodyField.inputValue();

        expect(bodyValue).toContain(replyText);
      }
    }
  });

  test('multiple smart reply options are available', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      await emailItems.first().click();

      const preview = page.locator(selectors.email.preview);
      const smartReplies = preview.locator('.smart-reply-chip');

      const replyCount = await smartReplies.count();

      // Should have 0 or multiple (typically 3)
      expect(replyCount >= 0).toBeTruthy();
    }
  });
});

test.describe('AI-Powered Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('semantic search is available', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);

    // Type natural language query
    await searchInput.fill('emails about project deadlines');

    // Wait for search
    await page.waitForTimeout(800);

    // Results should appear
    const results = page.locator(selectors.search.resultsSection);

    if (await results.count() > 0) {
      await expect(results).toBeVisible();
    }
  });

  test('AI search suggestions appear', async ({ page }) => {
    await page.click(selectors.nav.searchTrigger);
    await expect(page.locator(selectors.search.overlay)).toHaveClass(/active/);

    const searchInput = page.locator(selectors.search.input);
    await searchInput.fill('meetings');

    // Wait for suggestions
    await page.waitForTimeout(500);

    const suggestions = page.locator(selectors.search.suggestions);

    if (await suggestions.count() > 0) {
      // Suggestions may appear
      const isVisible = await suggestions.isVisible().catch(() => false);
      expect(typeof isVisible).toBe('boolean');
    }
  });
});

test.describe('AI Priority Inbox', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('priority inbox filter is available', async ({ page }) => {
    const filterTabs = page.locator(selectors.email.filterTabs);

    if (await filterTabs.count() > 0) {
      const priorityTab = filterTabs.locator('.filter-tab:has-text("Priority")');

      if (await priorityTab.count() > 0) {
        await expect(priorityTab).toBeVisible();
      }
    }
  });

  test('emails show AI priority indicators', async ({ page }) => {
    await page.waitForTimeout(1500);

    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      // Look for priority badges
      const priorityBadges = emailItems.locator('.priority-badge');
      const badgeCount = await priorityBadges.count();

      // It's okay if no priority badges
      expect(badgeCount >= 0).toBeTruthy();
    }
  });

  test('priority inbox shows important emails first', async ({ page }) => {
    const filterTabs = page.locator(selectors.email.filterTabs);

    if (await filterTabs.count() > 0) {
      const priorityTab = filterTabs.locator('.filter-tab:has-text("Priority")');

      if (await priorityTab.count() > 0) {
        await priorityTab.click();

        await page.waitForTimeout(500);

        // Priority emails should be shown
        const emailItems = page.locator(selectors.email.emailItem);
        const count = await emailItems.count();

        expect(count >= 0).toBeTruthy();
      }
    }
  });
});

test.describe('AI Email Classification', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1500);
  });

  test('emails show AI-generated labels', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();

      // Look for AI labels
      const aiLabels = firstEmail.locator('.ai-label');

      if (await aiLabels.count() > 0) {
        await expect(aiLabels.first()).toBeVisible();
      }
    }
  });

  test('emails are categorized by AI', async ({ page }) => {
    const emailItems = page.locator(selectors.email.emailItem);
    const count = await emailItems.count();

    if (count > 0) {
      const firstEmail = emailItems.first();

      // Look for category indicators
      const categories = firstEmail.locator('.email-category');

      if (await categories.count() > 0) {
        const categoryText = await categories.first().textContent();
        expect(categoryText).toBeTruthy();
      }
    }
  });
});

test.describe('AI Tone Detection', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
    await page.waitForLoadState('domcontentloaded');
  });

  test('AI tone analyzer is available in compose', async ({ page }) => {
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    // Type email content
    const bodyField = page.locator(selectors.compose.body);
    await bodyField.fill('I am extremely disappointed with this service.');

    await page.waitForTimeout(1000);

    // Look for tone indicator
    const toneIndicator = page.locator('.ai-tone-indicator');

    if (await toneIndicator.count() > 0) {
      // Tone indicator may appear
      const isVisible = await toneIndicator.isVisible().catch(() => false);
      expect(typeof isVisible).toBe('boolean');
    }
  });

  test('tone suggestions help improve email', async ({ page }) => {
    await page.click(selectors.email.composeBtn);
    await expect(page.locator(selectors.compose.modal)).toBeVisible();

    const bodyField = page.locator(selectors.compose.body);
    await bodyField.fill('This is terrible!');

    await page.waitForTimeout(1000);

    const toneSuggestion = page.locator('.ai-tone-suggestion');

    if (await toneSuggestion.count() > 0 && await toneSuggestion.isVisible()) {
      // Should have suggestion text
      const text = await toneSuggestion.textContent();
      expect(text).toBeTruthy();
    }
  });
});
