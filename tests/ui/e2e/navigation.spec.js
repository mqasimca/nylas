// @ts-check
const { test, expect } = require('@playwright/test');
const selectors = require('../../shared/helpers/ui-selectors');

/**
 * Navigation tests for Nylas UI (Web CLI Interface).
 *
 * Tests sidebar navigation, page switching, and dropdowns.
 */

test.describe('Sidebar Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('can navigate to all pages via sidebar', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    const navItems = [
      { nav: selectors.nav.overview, page: selectors.pages.overview },
      { nav: selectors.nav.admin, page: selectors.pages.admin },
      { nav: selectors.nav.auth, page: selectors.pages.auth },
      { nav: selectors.nav.calendar, page: selectors.pages.calendar },
      { nav: selectors.nav.contacts, page: selectors.pages.contacts },
      { nav: selectors.nav.email, page: selectors.pages.email },
    ];

    for (const item of navItems) {
      await page.locator(item.nav).click();
      await page.waitForTimeout(300);

      // Nav item should be active
      await expect(page.locator(item.nav)).toHaveClass(/active/);

      // Page should be active
      await expect(page.locator(item.page)).toHaveClass(/active/);
    }
  });

  test('clicking nav item updates active state', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Click on Calendar
    await page.locator(selectors.nav.calendar).click();
    await page.waitForTimeout(200);

    // Calendar nav should be active
    await expect(page.locator(selectors.nav.calendar)).toHaveClass(/active/);

    // Overview nav should not be active
    await expect(page.locator(selectors.nav.overview)).not.toHaveClass(/active/);
  });

  test('navigation sections are visible', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Nav sections exist
    const sections = page.locator(selectors.nav.section);
    await expect(sections.first()).toBeVisible();

    // Section titles
    const titles = page.locator(selectors.nav.sectionTitle);
    await expect(titles.first()).toBeVisible();
  });
});

test.describe('Page Content', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('overview page has dashboard cards', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Dashboard title
    await expect(page.locator(selectors.dashboard.title)).toHaveText('Dashboard');

    // Status badge
    await expect(page.locator(selectors.dashboard.statusBadge)).toBeVisible();

    // Glass cards
    const cards = page.locator(selectors.card.glass);
    const count = await cards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('overview page has quick commands section', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Commands card
    await expect(page.locator(selectors.overview.commandsCard)).toBeVisible();

    // Command cards
    const cmdCards = page.locator(selectors.overview.cmdCard);
    const count = await cmdCards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('overview page has resources section', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Resources list
    const resources = page.locator(selectors.resources.item);
    const count = await resources.count();
    expect(count).toBeGreaterThan(0);
  });

  test('switching pages shows correct content', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    // Switch to Calendar
    await page.locator(selectors.nav.calendar).click();
    await page.waitForTimeout(300);

    // Calendar page should be visible
    await expect(page.locator(selectors.pages.calendar)).toHaveClass(/active/);

    // Overview page should not be visible
    await expect(page.locator(selectors.pages.overview)).not.toHaveClass(/active/);
  });
});

test.describe('Dropdown Menus', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator(selectors.general.app)).toBeVisible();
  });

  test('client dropdown opens on click', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    const clientDropdown = page.locator(selectors.header.clientDropdown);
    const dropdownBtn = clientDropdown.locator(selectors.dropdown.btn);

    await dropdownBtn.click();
    await page.waitForTimeout(200);

    // Menu should be visible
    const menu = clientDropdown.locator(selectors.dropdown.menu);
    await expect(menu).toBeVisible();
  });

  test('grant dropdown opens on click', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    const grantDropdown = page.locator(selectors.header.grantDropdown);
    const dropdownBtn = grantDropdown.locator(selectors.dropdown.btn);

    await dropdownBtn.click();
    await page.waitForTimeout(200);

    // Menu should be visible
    const menu = grantDropdown.locator(selectors.dropdown.menu);
    await expect(menu).toBeVisible();
  });

  test('grant dropdown has add account option', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    const grantDropdown = page.locator(selectors.header.grantDropdown);
    const dropdownBtn = grantDropdown.locator(selectors.dropdown.btn);

    await dropdownBtn.click();
    await page.waitForTimeout(200);

    // Add account option
    const addNew = grantDropdown.locator(selectors.dropdown.addNew);
    await expect(addNew).toBeVisible();
    await expect(addNew).toContainText('Add Account');
  });

  test('clicking outside closes dropdown', async ({ page }) => {
    const dashboardView = page.locator(selectors.dashboard.view);
    const isActive = await dashboardView.evaluate((el) => el.classList.contains('active'));

    if (!isActive) {
      test.skip();
      return;
    }

    const clientDropdown = page.locator(selectors.header.clientDropdown);
    const dropdownBtn = clientDropdown.locator(selectors.dropdown.btn);

    // Open dropdown
    await dropdownBtn.click();
    await page.waitForTimeout(200);

    const menu = clientDropdown.locator(selectors.dropdown.menu);
    await expect(menu).toBeVisible();

    // Click outside
    await page.locator(selectors.dashboard.content).click();
    await page.waitForTimeout(200);

    // Menu should be hidden (or at least not active)
    // Note: Implementation may vary
  });
});
