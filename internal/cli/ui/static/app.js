// =============================================================================
// Nylas CLI - Dashboard (Main Entry Point)
// =============================================================================

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Initialize all modules
    initTheme();
    initForm();
    initDropdowns();
    initNavigation();
    initKeyboardShortcuts();
    initToast();

    // Use server-provided initial state (hybrid SSR)
    if (window.__INITIAL_STATE__) {
        initFromServerState(window.__INITIAL_STATE__);
    } else {
        // Fallback to API call if no initial state
        checkConfig();
    }

    // Update timestamps every minute
    setInterval(updateAllTimestamps, 60000);
});
