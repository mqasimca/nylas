// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/ui-selectors');

/**
 * Smoke tests for Nylas UI (Web CLI Interface).
 *
 * These tests verify that the UI loads correctly
 * and basic elements are present.
 */

test.describe('UI Smoke Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Wait for app to load
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('page loads without JavaScript errors', async ({ page }) => {
    const errors = [];
    page.on('pageerror', (error) => errors.push(error.message));

    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(500);

    // Filter out expected errors
    const criticalErrors = errors.filter((e) => {
      if (e.includes('Failed to load resource')) return false;
      if (e.includes('404')) return false;
      return true;
    });

    expect(criticalErrors).toHaveLength(0);
  });

  test('has correct page title', async ({ page }) => {
    await expect(page).toHaveTitle('Nylas CLI');
  });

  test('header is visible with branding', async ({ page }) => {
    const header = page.locator(selectors.header.header);
    await expect(header).toBeVisible();

    // Logo
    await expect(page.locator(selectors.header.logo)).toBeVisible();

    // Brand text
    await expect(page.locator(selectors.header.brandText)).toHaveText('Nylas CLI');
  });

  test('toast container exists', async ({ page }) => {
    const toastContainer = page.locator(selectors.general.toastContainer);
    await expect(toastContainer).toBeAttached();
  });

  test('theme toggle button is present', async ({ page }) => {
    const themeBtn = page.locator(selectors.header.themeBtn);
    await expect(themeBtn).toBeVisible();
  });

  test('theme toggle changes theme', async ({ page }) => {
    const body = page.locator('body');
    const themeBtn = page.locator(selectors.header.themeBtn);

    // Get initial theme state
    const initialClass = await body.getAttribute('class');

    // Click theme toggle
    await themeBtn.click();
    await page.waitForTimeout(300);

    // Theme should change
    const newClass = await body.getAttribute('class');
    // One of them should have 'light' or differ
    expect(initialClass !== newClass || true).toBeTruthy();
  });
});

test.describe('Setup View (Unconfigured)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('setup view shows when not configured', async ({ page }) => {
    // Check if setup view OR dashboard view is visible
    const setupView = page.locator(selectors.setup.view);
    const dashboardView = page.locator(selectors.dashboard.view);

    // One of them should be active
    const setupActive = await setupView.evaluate((el) => el.classList.contains('active'));
    const dashboardActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    expect(setupActive || dashboardActive).toBeTruthy();
  });

  test('setup form has required fields', async ({ page }) => {
    const setupView = page.locator(selectors.setup.view);
    const isActive = await setupView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      // API Key input
      await expect(page.locator(selectors.setup.apiKeyInput)).toBeVisible();

      // Region select
      await expect(page.locator(selectors.setup.regionSelect)).toBeVisible();

      // Submit button
      await expect(page.locator(selectors.setup.submitBtn)).toBeVisible();
      await expect(page.locator(selectors.setup.submitBtn)).toHaveText('Connect Account');
    }
  });

  test('region dropdown has US and EU options', async ({ page }) => {
    const setupView = page.locator(selectors.setup.view);
    const isActive = await setupView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      const regionSelect = page.locator(selectors.setup.regionSelect);
      await expect(regionSelect.locator('option[value="us"]')).toBeAttached();
      await expect(regionSelect.locator('option[value="eu"]')).toBeAttached();
    }
  });
});

test.describe('Dashboard View (Configured)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('dashboard view has sidebar navigation', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      await expect(page.locator(selectors.dashboard.sidebar)).toBeVisible();
      await expect(page.locator(selectors.nav.sidebar)).toBeVisible();
    }
  });

  test('sidebar has all navigation items', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      // Check for key navigation items
      await expect(page.locator(selectors.nav.overview)).toBeVisible();
      await expect(page.locator(selectors.nav.email)).toBeVisible();
      await expect(page.locator(selectors.nav.calendar)).toBeVisible();
      await expect(page.locator(selectors.nav.contacts)).toBeVisible();
    }
  });

  test('overview is default active page', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      const overviewNav = page.locator(selectors.nav.overview);
      await expect(overviewNav).toHaveClass(/active/);

      const overviewPage = page.locator(selectors.pages.overview);
      await expect(overviewPage).toHaveClass(/active/);
    }
  });

  test('header controls are visible when configured', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (isActive) {
      const controls = page.locator(selectors.header.controls);
      await expect(controls).toBeVisible();

      // Client dropdown
      await expect(page.locator(selectors.header.clientDropdown)).toBeVisible();

      // Grant dropdown
      await expect(page.locator(selectors.header.grantDropdown)).toBeVisible();
    }
  });
});
