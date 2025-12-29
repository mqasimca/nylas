// @ts-check
const { defineConfig, devices } = require('@playwright/test');

/**
 * Playwright configuration for Nylas Air E2E tests.
 *
 * @see https://playwright.dev/docs/test-configuration
 */
module.exports = defineConfig({
  // Test directory
  testDir: './e2e',

  // Run tests in parallel within files
  fullyParallel: true,

  // Fail the build on CI if accidentally left test.only
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Single worker on CI for stability, parallel locally
  workers: process.env.CI ? 1 : undefined,

  // Reporter configuration
  reporter: [
    ['list'],
    ['html', { open: 'never' }],
  ],

  // Shared settings for all projects
  use: {
    // Base URL for the Air server
    baseURL: process.env.AIR_BASE_URL || 'http://localhost:7365',

    // Collect trace on first retry
    trace: 'on-first-retry',

    // Take screenshot only on failure
    screenshot: 'only-on-failure',

    // Record video on first retry
    video: 'on-first-retry',

    // Timeout for each action
    actionTimeout: 10000,

    // Timeout for each navigation
    navigationTimeout: 30000,
  },

  // Global timeout for each test
  timeout: 30000,

  // Expect timeout
  expect: {
    timeout: 5000,
  },

  // Web server configuration
  webServer: {
    // Command to start the Air server
    command: 'cd .. && go run cmd/nylas/main.go air --no-browser --port 7365',

    // Port to wait for
    port: 7365,

    // Timeout for server startup
    timeout: 60000,

    // Reuse existing server in dev mode
    reuseExistingServer: !process.env.CI,

    // Environment variables for the server
    env: {
      AIR_TEST_MODE: 'true',
    },
  },

  // Projects (browser configurations)
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Viewport for consistent screenshots
        viewport: { width: 1280, height: 720 },
      },
    },

    // Uncomment for additional browser testing
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },

    // Uncomment for mobile testing
    // {
    //   name: 'Mobile Chrome',
    //   use: { ...devices['Pixel 5'] },
    // },
  ],

  // Output directory for test artifacts
  outputDir: 'test-results/',
});
