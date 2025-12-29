/**
 * Email Bundles Module
 * Smart email categorization inspired by Shortwave/Google Inbox
 * Groups emails by type: newsletters, receipts, social, etc.
 */

const Bundles = {
    // Bundle state
    bundles: [],
    emailBundles: new Map(), // emailId -> bundleId
    isLoaded: false,

    // Bundle icons for UI
    icons: {
        newsletters: '📰',
        receipts: '🧾',
        social: '👥',
        updates: '🔔',
        promotions: '🏷️',
        finance: '💰',
        travel: '✈️',
        primary: '📥',
    },

    /**
     * Initialize bundles module
     */
    async init() {
        try {
            await this.loadBundles();
            this.setupUI();
            this.isLoaded = true;
            console.log('%c📦 Bundles module loaded', 'color: #22c55e;');
        } catch (error) {
            console.error('Failed to initialize bundles:', error);
        }
    },

    /**
     * Load bundles from server
     */
    async loadBundles() {
        try {
            const response = await fetch('/api/bundles');
            if (response.ok) {
                this.bundles = await response.json();
            }
        } catch (error) {
            console.error('Failed to load bundles:', error);
            this.bundles = this.getDefaultBundles();
        }
    },

    /**
     * Get default bundles (fallback)
     */
    getDefaultBundles() {
        return [
            { id: 'newsletters', name: 'Newsletters', icon: '📰', collapsed: true, count: 0 },
            { id: 'receipts', name: 'Receipts & Orders', icon: '🧾', collapsed: true, count: 0 },
            { id: 'social', name: 'Social', icon: '👥', collapsed: true, count: 0 },
            { id: 'updates', name: 'Updates', icon: '🔔', collapsed: true, count: 0 },
            { id: 'promotions', name: 'Promotions', icon: '🏷️', collapsed: true, count: 0 },
        ];
    },

    /**
     * Setup bundle UI in sidebar
     */
    setupUI() {
        const sidebar = document.querySelector('.sidebar-bundles');
        if (!sidebar) return;

        // Clear existing
        while (sidebar.firstChild) {
            sidebar.removeChild(sidebar.firstChild);
        }

        // Add bundle items
        this.bundles.forEach(bundle => {
            if (bundle.count > 0) {
                sidebar.appendChild(this.createBundleItem(bundle));
            }
        });
    },

    /**
     * Create bundle list item
     * @param {Object} bundle - Bundle data
     * @returns {HTMLElement}
     */
    createBundleItem(bundle) {
        const item = document.createElement('div');
        item.className = 'bundle-item';
        item.dataset.bundleId = bundle.id;

        const icon = document.createElement('span');
        icon.className = 'bundle-icon';
        icon.textContent = bundle.icon || this.icons[bundle.id] || '📁';
        item.appendChild(icon);

        const name = document.createElement('span');
        name.className = 'bundle-name';
        name.textContent = bundle.name;
        item.appendChild(name);

        if (bundle.count > 0) {
            const count = document.createElement('span');
            count.className = 'bundle-count';
            count.textContent = bundle.count;
            item.appendChild(count);
        }

        const toggle = document.createElement('button');
        toggle.className = 'bundle-toggle';
        toggle.setAttribute('aria-label', bundle.collapsed ? 'Expand' : 'Collapse');
        toggle.textContent = bundle.collapsed ? '▶' : '▼';
        item.appendChild(toggle);

        // Click to expand/view bundle
        item.addEventListener('click', () => this.viewBundle(bundle.id));
        toggle.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleBundle(bundle.id);
        });

        return item;
    },

    /**
     * Categorize an email into a bundle
     * @param {Object} email - Email object with from, subject
     * @returns {string|null} Bundle ID or null
     */
    async categorize(email) {
        if (!email || !email.from) return null;

        try {
            const response = await fetch('/api/bundles/categorize', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    from: email.from,
                    subject: email.subject || '',
                    emailId: email.id,
                }),
            });

            if (response.ok) {
                const result = await response.json();
                if (result.bundleId) {
                    this.emailBundles.set(email.id, result.bundleId);
                    this.updateBundleCount(result.bundleId, 1);
                    return result.bundleId;
                }
            }
        } catch (error) {
            console.error('Failed to categorize email:', error);
        }

        return null;
    },

    /**
     * Categorize multiple emails
     * @param {Array} emails - Array of email objects
     */
    async categorizeAll(emails) {
        const promises = emails.map(email => this.categorize(email));
        await Promise.all(promises);
        this.setupUI();
    },

    /**
     * Update bundle count
     * @param {string} bundleId - Bundle ID
     * @param {number} delta - Change in count
     */
    updateBundleCount(bundleId, delta) {
        const bundle = this.bundles.find(b => b.id === bundleId);
        if (bundle) {
            bundle.count = (bundle.count || 0) + delta;
        }
    },

    /**
     * View emails in a bundle
     * @param {string} bundleId - Bundle ID
     */
    viewBundle(bundleId) {
        const bundle = this.bundles.find(b => b.id === bundleId);
        if (!bundle) return;

        // Dispatch event for email list to filter
        const event = new CustomEvent('bundleSelected', {
            detail: { bundleId, bundleName: bundle.name },
        });
        document.dispatchEvent(event);

        // Update active state
        document.querySelectorAll('.bundle-item').forEach(item => {
            item.classList.toggle('active', item.dataset.bundleId === bundleId);
        });
    },

    /**
     * Toggle bundle collapsed state
     * @param {string} bundleId - Bundle ID
     */
    async toggleBundle(bundleId) {
        const bundle = this.bundles.find(b => b.id === bundleId);
        if (!bundle) return;

        bundle.collapsed = !bundle.collapsed;

        try {
            await fetch('/api/bundles', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(bundle),
            });
        } catch (error) {
            console.error('Failed to update bundle:', error);
        }

        this.setupUI();
    },

    /**
     * Get bundle for an email
     * @param {string} emailId - Email ID
     * @returns {Object|null} Bundle or null
     */
    getBundleForEmail(emailId) {
        const bundleId = this.emailBundles.get(emailId);
        return bundleId ? this.bundles.find(b => b.id === bundleId) : null;
    },

    /**
     * Check if email is in a collapsed bundle
     * @param {string} emailId - Email ID
     * @returns {boolean}
     */
    isInCollapsedBundle(emailId) {
        const bundle = this.getBundleForEmail(emailId);
        return bundle ? bundle.collapsed : false;
    },

    /**
     * Get all emails in a bundle
     * @param {string} bundleId - Bundle ID
     * @returns {string[]} Array of email IDs
     */
    getEmailsInBundle(bundleId) {
        const emailIds = [];
        this.emailBundles.forEach((bId, emailId) => {
            if (bId === bundleId) {
                emailIds.push(emailId);
            }
        });
        return emailIds;
    },
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    Bundles.init();
});

// Export for use
if (typeof window !== 'undefined') {
    window.Bundles = Bundles;
}
