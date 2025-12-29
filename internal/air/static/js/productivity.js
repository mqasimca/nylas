/* Productivity Core - Split Inbox functionality */

// ====================================
// PRODUCTIVITY MODULE
// Split Inbox, Snooze, Send Later, Undo Send, Templates
// ====================================

// =============================================================================
// SPLIT INBOX MANAGER
// =============================================================================

// Split Inbox Manager - DISABLED (using simplified filters in EmailListManager)
// Filter tabs are now: All, VIP, Unread - handled directly by EmailListManager
const SplitInboxManager = {
    config: null,

    async init() {
        // Filter handling is now done by EmailListManager
        // This module just loads the VIP config
        try {
            const response = await AirAPI.getSplitInboxConfig();
            this.config = response.config || response || {};
            console.log('%c📬 Inbox filters ready', 'color: #22c55e;');
        } catch (error) {
            console.log('%c📬 Inbox filters: using defaults', 'color: #a1a1aa;');
            this.config = { enabled: true };
        }
    }
};

// Legacy compatibility - do not use
const _legacySplitInbox = {
    async loadCategorizedEmails(category) {
        // Redirect to EmailListManager
        if (typeof EmailListManager !== 'undefined') {
            EmailListManager.setFilter(category);
        }
    }
};

// REMOVED: Complex category tab system
// The simplified filter system (All, VIP, Unread) is now in the HTML template
// and handled by EmailListManager.setFilter()

// Note: Complex category filtering was removed in favor of simplified All/VIP/Unread tabs

// =============================================================================
// SNOOZE MANAGER
// =============================================================================

