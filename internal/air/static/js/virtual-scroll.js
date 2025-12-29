/**
 * Virtual Scroll Module
 * Efficiently renders large lists by only rendering visible items
 * Implements <100ms render target inspired by Superhuman
 */

const VirtualScroll = {
    // Configuration
    config: {
        itemHeight: 72, // Default email item height in pixels
        overscan: 5, // Extra items to render above/below viewport
        bufferSize: 20, // Items to preload
        throttleMs: 16, // ~60fps
    },

    // State
    state: {
        items: [],
        visibleStart: 0,
        visibleEnd: 0,
        scrollTop: 0,
        containerHeight: 0,
        isScrolling: false,
        scrollTimeout: null,
    },

    // DOM references
    container: null,
    content: null,
    spacerTop: null,
    spacerBottom: null,

    /**
     * Initialize virtual scroll on a container
     * @param {HTMLElement} container - The scrollable container
     * @param {Array} items - Array of items to render
     * @param {Function} renderItem - Function to render each item
     * @param {Object} options - Configuration options
     */
    init(container, items, renderItem, options = {}) {
        this.container = container;
        this.state.items = items;
        this.renderItem = renderItem;
        this.config = { ...this.config, ...options };

        this.setupDOM();
        this.setupListeners();
        this.calculateVisibleRange();
        this.render();

        console.log('%c Virtual scroll initialized', 'color: #22c55e;', {
            items: items.length,
            itemHeight: this.config.itemHeight,
        });

        return this;
    },

    /**
     * Setup DOM structure for virtual scrolling
     */
    setupDOM() {
        // Create content wrapper if not exists
        this.content = this.container.querySelector('.virtual-content');
        if (!this.content) {
            this.content = document.createElement('div');
            this.content.className = 'virtual-content';

            // Move existing children into content
            while (this.container.firstChild) {
                this.content.appendChild(this.container.firstChild);
            }
            this.container.appendChild(this.content);
        }

        // Create spacers for maintaining scroll position
        this.spacerTop = document.createElement('div');
        this.spacerTop.className = 'virtual-spacer-top';
        this.spacerTop.style.height = '0px';

        this.spacerBottom = document.createElement('div');
        this.spacerBottom.className = 'virtual-spacer-bottom';
        this.spacerBottom.style.height = '0px';

        this.content.insertBefore(this.spacerTop, this.content.firstChild);
        this.content.appendChild(this.spacerBottom);

        // Set container styles
        this.container.style.overflowY = 'auto';
        this.container.style.position = 'relative';
    },

    /**
     * Setup scroll and resize listeners
     */
    setupListeners() {
        // Throttled scroll handler
        let lastScroll = 0;
        this.container.addEventListener('scroll', () => {
            const now = performance.now();
            if (now - lastScroll < this.config.throttleMs) return;
            lastScroll = now;

            this.state.scrollTop = this.container.scrollTop;
            this.state.isScrolling = true;
            this.calculateVisibleRange();
            this.render();

            // Clear scrolling state after delay
            clearTimeout(this.state.scrollTimeout);
            this.state.scrollTimeout = setTimeout(() => {
                this.state.isScrolling = false;
            }, 150);
        }, { passive: true });

        // Resize observer for container
        if (typeof ResizeObserver !== 'undefined') {
            const resizeObserver = new ResizeObserver(() => {
                this.state.containerHeight = this.container.clientHeight;
                this.calculateVisibleRange();
                this.render();
            });
            resizeObserver.observe(this.container);
        }

        this.state.containerHeight = this.container.clientHeight;
    },

    /**
     * Calculate visible item range based on scroll position
     */
    calculateVisibleRange() {
        const { itemHeight, overscan } = this.config;
        const { scrollTop, containerHeight, items } = this.state;

        const start = Math.floor(scrollTop / itemHeight);
        const visibleCount = Math.ceil(containerHeight / itemHeight);
        const end = start + visibleCount;

        this.state.visibleStart = Math.max(0, start - overscan);
        this.state.visibleEnd = Math.min(items.length, end + overscan);
    },

    /**
     * Render visible items
     */
    render() {
        const startTime = performance.now();
        const { items, visibleStart, visibleEnd } = this.state;
        const { itemHeight } = this.config;

        // Update spacers
        this.spacerTop.style.height = `${visibleStart * itemHeight}px`;
        this.spacerBottom.style.height = `${(items.length - visibleEnd) * itemHeight}px`;

        // Get visible items
        const visibleItems = items.slice(visibleStart, visibleEnd);

        // Clear existing items (except spacers)
        const existingItems = this.content.querySelectorAll('.virtual-item');
        existingItems.forEach(item => item.remove());

        // Render visible items
        const fragment = document.createDocumentFragment();
        visibleItems.forEach((item, index) => {
            const el = this.renderItem(item, visibleStart + index);
            el.classList.add('virtual-item');
            el.style.height = `${itemHeight}px`;
            fragment.appendChild(el);
        });

        // Insert after top spacer
        this.spacerTop.after(fragment);

        // Log performance in dev mode
        const renderTime = performance.now() - startTime;
        if (renderTime > 16) {
            console.warn('Virtual scroll render slow:', renderTime.toFixed(2) + 'ms');
        }
    },

    /**
     * Update items and re-render
     * @param {Array} newItems - New items array
     */
    updateItems(newItems) {
        this.state.items = newItems;
        this.calculateVisibleRange();
        this.render();
    },

    /**
     * Scroll to specific item index
     * @param {number} index - Item index to scroll to
     * @param {string} behavior - 'smooth' or 'instant'
     */
    scrollToIndex(index, behavior = 'smooth') {
        const { itemHeight } = this.config;
        const scrollTop = index * itemHeight;

        this.container.scrollTo({
            top: scrollTop,
            behavior,
        });
    },

    /**
     * Get current visible items
     * @returns {Array} Currently visible items
     */
    getVisibleItems() {
        const { items, visibleStart, visibleEnd } = this.state;
        return items.slice(visibleStart, visibleEnd);
    },

    /**
     * Destroy virtual scroll instance
     */
    destroy() {
        this.spacerTop?.remove();
        this.spacerBottom?.remove();
        this.container = null;
        this.content = null;
    },
};

// Export for use
if (typeof window !== 'undefined') {
    window.VirtualScroll = VirtualScroll;
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = VirtualScroll;
}
