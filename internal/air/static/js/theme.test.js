/**
 * Theme Manager Unit Tests
 * Run in browser console: copy-paste this file or load as a script
 * Usage: ThemeTests.run()
 */

const ThemeTests = {
    passed: 0,
    failed: 0,
    results: [],

    assert(condition, message) {
        if (condition) {
            this.passed++;
            this.results.push({ status: 'PASS', message });
            console.log(`✓ ${message}`);
        } else {
            this.failed++;
            this.results.push({ status: 'FAIL', message });
            console.error(`✗ ${message}`);
        }
    },

    assertEquals(actual, expected, message) {
        this.assert(actual === expected, `${message} (expected: ${expected}, got: ${actual})`);
    },

    assertExists(value, message) {
        this.assert(value !== undefined && value !== null, message);
    },

    // Test: ThemeManager exists
    testThemeManagerExists() {
        this.assertExists(window.ThemeManager, 'ThemeManager should exist on window');
    },

    // Test: themes array is defined
    testThemesArray() {
        this.assert(Array.isArray(ThemeManager.themes), 'themes should be an array');
        this.assertEquals(ThemeManager.themes.length, 4, 'Should have 4 themes');
        this.assert(ThemeManager.themes.includes('dark'), 'Should include dark theme');
        this.assert(ThemeManager.themes.includes('light'), 'Should include light theme');
        this.assert(ThemeManager.themes.includes('oled'), 'Should include oled theme');
        this.assert(ThemeManager.themes.includes('system'), 'Should include system theme');
    },

    // Test: setTheme updates currentTheme
    testSetTheme() {
        const originalTheme = ThemeManager.currentTheme;

        ThemeManager.setTheme('dark');
        this.assertEquals(ThemeManager.currentTheme, 'dark', 'setTheme should update currentTheme to dark');

        ThemeManager.setTheme('light');
        this.assertEquals(ThemeManager.currentTheme, 'light', 'setTheme should update currentTheme to light');

        ThemeManager.setTheme('oled');
        this.assertEquals(ThemeManager.currentTheme, 'oled', 'setTheme should update currentTheme to oled');

        ThemeManager.setTheme('system');
        this.assertEquals(ThemeManager.currentTheme, 'system', 'setTheme should update currentTheme to system');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: setTheme handles invalid theme
    testSetThemeInvalid() {
        const originalTheme = ThemeManager.currentTheme;

        ThemeManager.setTheme('invalid-theme');
        this.assertEquals(ThemeManager.currentTheme, 'system', 'Invalid theme should fall back to system');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: setTheme persists to localStorage
    testSetThemePersistence() {
        const originalTheme = ThemeManager.currentTheme;

        ThemeManager.setTheme('oled');
        const stored = localStorage.getItem('nylas-air-theme');
        this.assertEquals(stored, 'oled', 'Theme should be saved to localStorage');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: applyTheme sets data-theme attribute
    testApplyTheme() {
        const root = document.documentElement;
        const originalTheme = ThemeManager.currentTheme;

        ThemeManager.setTheme('light');
        this.assertEquals(root.getAttribute('data-theme'), 'light', 'Light theme should set data-theme="light"');

        ThemeManager.setTheme('oled');
        this.assertEquals(root.getAttribute('data-theme'), 'oled', 'OLED theme should set data-theme="oled"');

        ThemeManager.setTheme('dark');
        const darkAttr = root.getAttribute('data-theme');
        this.assert(darkAttr === null || darkAttr === '', 'Dark theme should remove data-theme attribute');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: getActiveTheme returns correct theme
    testGetActiveTheme() {
        const originalTheme = ThemeManager.currentTheme;

        ThemeManager.setTheme('dark');
        this.assertEquals(ThemeManager.getActiveTheme(), 'dark', 'getActiveTheme should return dark');

        ThemeManager.setTheme('light');
        this.assertEquals(ThemeManager.getActiveTheme(), 'light', 'getActiveTheme should return light');

        ThemeManager.setTheme('oled');
        this.assertEquals(ThemeManager.getActiveTheme(), 'oled', 'getActiveTheme should return oled');

        // System theme should return dark or light based on system preference
        ThemeManager.setTheme('system');
        const activeTheme = ThemeManager.getActiveTheme();
        this.assert(activeTheme === 'dark' || activeTheme === 'light', 'System theme should resolve to dark or light');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: cycleTheme cycles through themes
    testCycleTheme() {
        const originalTheme = ThemeManager.currentTheme;
        const themes = ThemeManager.themes;

        // Start at known position
        ThemeManager.setTheme('dark');
        const startIndex = themes.indexOf('dark');

        ThemeManager.cycleTheme();
        const expectedNext = themes[(startIndex + 1) % themes.length];
        this.assertEquals(ThemeManager.currentTheme, expectedNext, 'cycleTheme should move to next theme');

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: systemPreferenceQuery exists
    testSystemPreferenceQuery() {
        this.assertExists(ThemeManager.systemPreferenceQuery, 'systemPreferenceQuery should exist');
        this.assert(ThemeManager.systemPreferenceQuery instanceof MediaQueryList, 'Should be MediaQueryList');
    },

    // Test: themechange event is dispatched
    testThemeChangeEvent() {
        const originalTheme = ThemeManager.currentTheme;
        let eventFired = false;
        let eventDetail = null;

        const handler = (e) => {
            eventFired = true;
            eventDetail = e.detail;
        };

        window.addEventListener('themechange', handler);
        ThemeManager.setTheme('light');

        this.assert(eventFired, 'themechange event should be dispatched');
        this.assertEquals(eventDetail?.theme, 'light', 'Event detail should contain theme');

        window.removeEventListener('themechange', handler);

        // Restore original
        ThemeManager.setTheme(originalTheme);
    },

    // Test: createThemeToggle creates DOM elements
    testCreateThemeToggle() {
        // Create a test container
        const container = document.createElement('div');
        container.id = 'testThemeContainer';
        document.body.appendChild(container);

        ThemeManager.createThemeToggle(container);

        const toggle = container.querySelector('#themeToggle');
        this.assertExists(toggle, 'Theme toggle button should be created');

        const dropdown = container.querySelector('.theme-dropdown');
        this.assertExists(dropdown, 'Theme dropdown should be created');

        const options = container.querySelectorAll('.theme-option');
        this.assertEquals(options.length, 4, 'Should have 4 theme options');

        // Clean up
        container.remove();
    },

    // Run all tests
    run() {
        console.log('🎨 Running Theme Manager Unit Tests...\n');

        this.passed = 0;
        this.failed = 0;
        this.results = [];

        // Run all test methods
        const testMethods = Object.getOwnPropertyNames(ThemeTests)
            .filter(name => name.startsWith('test'));

        for (const method of testMethods) {
            try {
                console.log(`\n--- ${method} ---`);
                this[method]();
            } catch (error) {
                this.failed++;
                this.results.push({ status: 'ERROR', message: `${method}: ${error.message}` });
                console.error(`✗ ${method}: ${error.message}`);
            }
        }

        // Summary
        console.log('\n========================================');
        console.log(`Tests completed: ${this.passed + this.failed}`);
        console.log(`Passed: ${this.passed}`);
        console.log(`Failed: ${this.failed}`);
        console.log('========================================\n');

        if (this.failed === 0) {
            console.log('🎨 All tests passed!');
        } else {
            console.log('❌ Some tests failed');
        }

        return { passed: this.passed, failed: this.failed, results: this.results };
    }
};

// Export for use
if (typeof window !== 'undefined') {
    window.ThemeTests = ThemeTests;
}

// Auto-run if loaded as script with ?run parameter
if (typeof location !== 'undefined' && location.search.includes('run')) {
    document.addEventListener('DOMContentLoaded', () => ThemeTests.run());
}
