/**
 * API mocking utilities for Nylas Air E2E tests.
 *
 * Use these helpers to intercept API requests and provide
 * predictable responses for testing.
 */

/**
 * Mock the emails API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 * @param {Array} [options.emails] - Array of mock email objects
 * @param {number} [options.status] - HTTP status code
 * @param {number} [options.delay] - Response delay in ms
 */
exports.mockEmailsAPI = async (page, options = {}) => {
  const { emails = [], status = 200, delay = 0 } = options;

  await page.route('/api/emails*', async (route) => {
    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay));
    }

    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({
        emails,
        next_cursor: null,
      }),
    });
  });
};

/**
 * Mock the folders API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockFoldersAPI = async (page, options = {}) => {
  const { folders = [], status = 200 } = options;

  await page.route('/api/folders*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ folders }),
    });
  });
};

/**
 * Mock the calendars API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockCalendarsAPI = async (page, options = {}) => {
  const { calendars = [], status = 200 } = options;

  await page.route('/api/calendars*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ calendars }),
    });
  });
};

/**
 * Mock the events API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockEventsAPI = async (page, options = {}) => {
  const { events = [], status = 200 } = options;

  await page.route('/api/events*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ events }),
    });
  });
};

/**
 * Mock the contacts API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockContactsAPI = async (page, options = {}) => {
  const { contacts = [], status = 200 } = options;

  await page.route('/api/contacts*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ contacts }),
    });
  });
};

/**
 * Mock the config API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockConfigAPI = async (page, options = {}) => {
  const {
    configured = true,
    clientId = 'test-client-id',
    region = 'us',
    hasApiKey = true,
    status = 200,
  } = options;

  await page.route('/api/config*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({
        configured,
        client_id: clientId,
        region,
        has_api_key: hasApiKey,
      }),
    });
  });
};

/**
 * Mock the grants API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {Object} options
 */
exports.mockGrantsAPI = async (page, options = {}) => {
  const { grants = [], defaultGrantId = '', status = 200 } = options;

  await page.route('/api/grants*', async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({
        grants,
        default_grant_id: defaultGrantId,
      }),
    });
  });
};

/**
 * Mock an API error response.
 *
 * @param {import('@playwright/test').Page} page
 * @param {string} path - API path to mock (e.g., '/api/emails')
 * @param {Object} options
 */
exports.mockAPIError = async (page, path, options = {}) => {
  const { status = 500, message = 'Internal Server Error' } = options;

  await page.route(`${path}*`, async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ error: message }),
    });
  });
};

/**
 * Mock network failure for an API endpoint.
 *
 * @param {import('@playwright/test').Page} page
 * @param {string} path - API path to mock
 */
exports.mockNetworkFailure = async (page, path) => {
  await page.route(`${path}*`, async (route) => {
    await route.abort('failed');
  });
};

/**
 * Setup all API mocks with default empty data.
 * Useful for testing UI in a clean state.
 *
 * @param {import('@playwright/test').Page} page
 */
exports.mockAllAPIsEmpty = async (page) => {
  await exports.mockConfigAPI(page, { configured: true });
  await exports.mockGrantsAPI(page, { grants: [], defaultGrantId: '' });
  await exports.mockFoldersAPI(page, { folders: [] });
  await exports.mockEmailsAPI(page, { emails: [] });
  await exports.mockCalendarsAPI(page, { calendars: [] });
  await exports.mockEventsAPI(page, { events: [] });
  await exports.mockContactsAPI(page, { contacts: [] });
};

/**
 * Wait for an API request to complete.
 *
 * @param {import('@playwright/test').Page} page
 * @param {string} path - API path to wait for
 * @returns {Promise<import('@playwright/test').Response>}
 */
exports.waitForAPI = async (page, path) => {
  return page.waitForResponse((response) =>
    response.url().includes(path) && response.status() === 200
  );
};
