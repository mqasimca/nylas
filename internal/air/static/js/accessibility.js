// ====================================
// ACCESSIBILITY ENHANCEMENTS - Phase 7
// ====================================

/**
 * Keyboard navigation detection
 * Tracks whether user is using keyboard or mouse for navigation
 */
const KeyboardDetector = {
    isKeyboardUser: false,

    init() {
        // Detect keyboard navigation
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                this.setKeyboardUser(true);
            }
        });

        // Detect mouse navigation
        document.addEventListener('mousedown', () => {
            this.setKeyboardUser(false);
        });

        // Touch detection
        document.addEventListener('touchstart', () => {
            this.setKeyboardUser(false);
        });
    },

    setKeyboardUser(isKeyboard) {
        this.isKeyboardUser = isKeyboard;
        document.documentElement.setAttribute('data-keyboard-user', isKeyboard);
    }
};

/**
 * Focus management utilities
 */
const FocusManager = {
    // Trap focus within an element (for modals, dialogs)
    trapFocus(element) {
        const focusableElements = element.querySelectorAll(
            'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"]):not([disabled])'
        );

        const firstFocusable = focusableElements[0];
        const lastFocusable = focusableElements[focusableElements.length - 1];

        const handleTabKey = (e) => {
            if (e.key !== 'Tab') return;

            if (e.shiftKey) {
                // Shift + Tab
                if (document.activeElement === firstFocusable) {
                    lastFocusable.focus();
                    e.preventDefault();
                }
            } else {
                // Tab
                if (document.activeElement === lastFocusable) {
                    firstFocusable.focus();
                    e.preventDefault();
                }
            }
        };

        element.addEventListener('keydown', handleTabKey);

        // Return cleanup function
        return () => element.removeEventListener('keydown', handleTabKey);
    },

    // Focus first interactive element in container
    focusFirst(container) {
        const firstFocusable = container.querySelector(
            'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"]):not([disabled])'
        );
        if (firstFocusable) {
            firstFocusable.focus();
        }
    },

    // Restore focus to an element
    saveFocus() {
        return document.activeElement;
    },

    restoreFocus(element) {
        if (element && element.focus) {
            element.focus();
        }
    }
};

/**
 * ARIA live region announcer
 * Makes announcements to screen readers
 */
const Announcer = {
    liveRegion: null,

    init() {
        // Create ARIA live region if it doesn't exist
        if (!this.liveRegion) {
            this.liveRegion = document.createElement('div');
            this.liveRegion.setAttribute('role', 'status');
            this.liveRegion.setAttribute('aria-live', 'polite');
            this.liveRegion.setAttribute('aria-atomic', 'true');
            this.liveRegion.className = 'live-region sr-only';
            document.body.appendChild(this.liveRegion);
        }
    },

    // Announce a message to screen readers
    announce(message, priority = 'polite') {
        if (!this.liveRegion) this.init();

        this.liveRegion.setAttribute('aria-live', priority);

        // Clear previous message
        this.liveRegion.textContent = '';

        // Add new message after a brief delay (forces screen reader to announce)
        setTimeout(() => {
            this.liveRegion.textContent = message;
        }, 100);
    },

    // Announce assertive message (interrupts screen reader)
    announceAssertive(message) {
        this.announce(message, 'assertive');
    }
};

/**
 * Skip links helper
 */
const SkipLinks = {
    init() {
        // Create skip to main content link if it doesn't exist
        if (!document.querySelector('.skip-to-content')) {
            const skipLink = document.createElement('a');
            skipLink.href = '#main-content';
            skipLink.className = 'skip-to-content';
            skipLink.textContent = 'Skip to main content';
            document.body.insertBefore(skipLink, document.body.firstChild);

            // Ensure main content has ID
            const mainContent = document.querySelector('[role="main"], main, #main-content');
            if (mainContent && !mainContent.id) {
                mainContent.id = 'main-content';
                mainContent.setAttribute('tabindex', '-1'); // Make focusable
            }
        }
    }
};

/**
 * Keyboard shortcuts manager with documentation
 */
const KeyboardShortcuts = {
    shortcuts: new Map(),
    enabled: true,

    // Register a keyboard shortcut
    register(key, action, description, category = 'General') {
        this.shortcuts.set(key, { action, description, category });
    },

    // Handle keyboard events
    handleKeydown(e) {
        if (!this.enabled) return;

        // Don't trigger shortcuts when typing in inputs
        if (e.target.matches('input, textarea, [contenteditable="true"]')) {
            // Exception: Escape key
            if (e.key !== 'Escape') return;
        }

        const key = this.getKeyString(e);
        const shortcut = this.shortcuts.get(key);

        if (shortcut) {
            e.preventDefault();
            shortcut.action(e);
        }
    },

    // Get key string from event (e.g., "Ctrl+S", "j", "?")
    getKeyString(e) {
        const parts = [];

        if (e.ctrlKey) parts.push('Ctrl');
        if (e.altKey) parts.push('Alt');
        if (e.shiftKey && e.key !== 'Shift') parts.push('Shift');
        if (e.metaKey) parts.push('Meta');

        // Add the key itself (except modifier keys)
        if (!['Control', 'Alt', 'Shift', 'Meta'].includes(e.key)) {
            parts.push(e.key);
        }

        return parts.join('+');
    },

    // Get all shortcuts grouped by category
    getAll() {
        const grouped = {};
        this.shortcuts.forEach((shortcut, key) => {
            if (!grouped[shortcut.category]) {
                grouped[shortcut.category] = [];
            }
            grouped[shortcut.category].push({
                key,
                description: shortcut.description
            });
        });
        return grouped;
    },

    // Enable/disable all shortcuts
    setEnabled(enabled) {
        this.enabled = enabled;
    },

    init() {
        document.addEventListener('keydown', (e) => this.handleKeydown(e));
    }
};

/**
 * Roving tab index for lists (keyboard navigation)
 */
class RovingTabIndex {
    constructor(container, itemSelector) {
        this.container = container;
        this.itemSelector = itemSelector;
        this.currentIndex = 0;
        this.items = [];

        this.init();
    }

    init() {
        this.updateItems();

        this.container.addEventListener('keydown', (e) => {
            switch (e.key) {
                case 'ArrowDown':
                case 'j':
                    e.preventDefault();
                    this.next();
                    break;
                case 'ArrowUp':
                case 'k':
                    e.preventDefault();
                    this.previous();
                    break;
                case 'Home':
                    e.preventDefault();
                    this.first();
                    break;
                case 'End':
                    e.preventDefault();
                    this.last();
                    break;
            }
        });

        // Update when items change
        const observer = new MutationObserver(() => this.updateItems());
        observer.observe(this.container, { childList: true, subtree: true });
    }

    updateItems() {
        this.items = Array.from(this.container.querySelectorAll(this.itemSelector));

        // Set initial tab indices
        this.items.forEach((item, index) => {
            item.setAttribute('tabindex', index === this.currentIndex ? '0' : '-1');
        });
    }

    focusItem(index) {
        if (index < 0 || index >= this.items.length) return;

        // Update tab indices
        this.items[this.currentIndex]?.setAttribute('tabindex', '-1');
        this.items[index]?.setAttribute('tabindex', '0');

        // Focus and scroll into view
        this.items[index]?.focus();
        this.items[index]?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });

        this.currentIndex = index;
    }

    next() {
        this.focusItem((this.currentIndex + 1) % this.items.length);
    }

    previous() {
        this.focusItem((this.currentIndex - 1 + this.items.length) % this.items.length);
    }

    first() {
        this.focusItem(0);
    }

    last() {
        this.focusItem(this.items.length - 1);
    }
}

/**
 * Color contrast checker (WCAG 2.1 compliance)
 */
const ContrastChecker = {
    // Calculate relative luminance
    getLuminance(rgb) {
        const [r, g, b] = rgb.map(val => {
            val /= 255;
            return val <= 0.03928 ? val / 12.92 : Math.pow((val + 0.055) / 1.055, 2.4);
        });
        return 0.2126 * r + 0.7152 * g + 0.0722 * b;
    },

    // Calculate contrast ratio
    getContrastRatio(rgb1, rgb2) {
        const lum1 = this.getLuminance(rgb1);
        const lum2 = this.getLuminance(rgb2);
        const lighter = Math.max(lum1, lum2);
        const darker = Math.min(lum1, lum2);
        return (lighter + 0.05) / (darker + 0.05);
    },

    // Check if contrast meets WCAG level (AA = 4.5:1, AAA = 7:1 for normal text)
    checkContrast(foreground, background, level = 'AA') {
        const ratio = this.getContrastRatio(foreground, background);
        const threshold = level === 'AAA' ? 7 : 4.5;
        return {
            ratio: ratio.toFixed(2),
            passes: ratio >= threshold,
            level
        };
    },

    // Parse CSS color to RGB
    parseColor(color) {
        const ctx = document.createElement('canvas').getContext('2d');
        ctx.fillStyle = color;
        const computed = ctx.fillStyle;

        // Extract RGB values from computed color
        const match = computed.match(/\d+/g);
        return match ? match.slice(0, 3).map(Number) : [0, 0, 0];
    }
};

// ====================================
// INITIALIZATION
// ====================================

// Initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        KeyboardDetector.init();
        Announcer.init();
        SkipLinks.init();
        KeyboardShortcuts.init();
    });
} else {
    KeyboardDetector.init();
    Announcer.init();
    SkipLinks.init();
    KeyboardShortcuts.init();
}

// Export for global use
window.A11y = {
    KeyboardDetector,
    FocusManager,
    Announcer,
    SkipLinks,
    KeyboardShortcuts,
    RovingTabIndex,
    ContrastChecker
};

// Alias for convenience
window.announce = (message) => Announcer.announce(message);

console.log('%c♿ Accessibility utilities loaded', 'color: #22c55e;');
