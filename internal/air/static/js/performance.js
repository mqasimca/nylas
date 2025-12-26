// ====================================
// PERFORMANCE UTILITIES
// ====================================

/**
 * Debounce function - delays execution until after wait time has passed
 * Use for: search inputs, resize handlers, scroll handlers
 * @param {Function} func - Function to debounce
 * @param {number} wait - Milliseconds to wait
 * @returns {Function} - Debounced function
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const context = this;
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(context, args), wait);
    };
}

/**
 * Throttle function - ensures function is called at most once per interval
 * Use for: scroll handlers, mouse move handlers
 * @param {Function} func - Function to throttle
 * @param {number} limit - Minimum milliseconds between calls
 * @returns {Function} - Throttled function
 */
function throttle(func, limit) {
    let inThrottle;
    return function executedFunction(...args) {
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

/**
 * Request animation frame wrapper for smooth animations
 * Use for: visual updates, DOM manipulations
 * @param {Function} callback - Function to execute on next frame
 */
function rafSchedule(callback) {
    let rafId = null;
    return function scheduledCallback(...args) {
        if (rafId !== null) {
            cancelAnimationFrame(rafId);
        }
        rafId = requestAnimationFrame(() => {
            callback(...args);
            rafId = null;
        });
    };
}

/**
 * Batch DOM reads and writes to avoid layout thrashing
 * Use for: Multiple DOM measurements or updates
 */
const DOMBatcher = {
    reads: [],
    writes: [],
    scheduled: false,

    read(fn) {
        this.reads.push(fn);
        this.schedule();
    },

    write(fn) {
        this.writes.push(fn);
        this.schedule();
    },

    schedule() {
        if (this.scheduled) return;
        this.scheduled = true;
        requestAnimationFrame(() => this.flush());
    },

    flush() {
        // Execute all reads first
        this.reads.forEach(fn => fn());
        this.reads = [];

        // Then execute all writes
        this.writes.forEach(fn => fn());
        this.writes = [];

        this.scheduled = false;
    }
};

/**
 * Lazy image loading with IntersectionObserver
 * Automatically loads images when they enter viewport
 */
const LazyLoader = {
    observer: null,

    init() {
        if (this.observer) return;

        // Check if IntersectionObserver is supported
        if (!('IntersectionObserver' in window)) {
            // Fallback: load all images immediately
            this.loadAll();
            return;
        }

        this.observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.loadImage(entry.target);
                    this.observer.unobserve(entry.target);
                }
            });
        }, {
            rootMargin: '50px', // Start loading 50px before entering viewport
            threshold: 0.01
        });

        // Observe all lazy images
        this.observeImages();
    },

    observeImages() {
        const lazyImages = document.querySelectorAll('img[data-src], [data-bg-src]');
        lazyImages.forEach(img => this.observer.observe(img));
    },

    loadImage(img) {
        // Load img src
        if (img.dataset.src) {
            img.src = img.dataset.src;
            delete img.dataset.src;
        }

        // Load background image
        if (img.dataset.bgSrc) {
            img.style.backgroundImage = `url('${img.dataset.bgSrc}')`;
            delete img.dataset.bgSrc;
        }

        img.classList.add('loaded');
    },

    loadAll() {
        // Fallback for browsers without IntersectionObserver
        const lazyImages = document.querySelectorAll('img[data-src], [data-bg-src]');
        lazyImages.forEach(img => this.loadImage(img));
    }
};

/**
 * Virtual scrolling for large lists
 * Only renders visible items plus buffer
 */
class VirtualScroller {
    constructor(container, items, renderItem, itemHeight = 60) {
        this.container = container;
        this.items = items;
        this.renderItem = renderItem;
        this.itemHeight = itemHeight;
        this.buffer = 5; // Number of items to render outside viewport
        this.scrollTop = 0;
        this.visibleStart = 0;
        this.visibleEnd = 0;

        this.init();
    }

    init() {
        // Create scroll container
        this.scrollContainer = document.createElement('div');
        this.scrollContainer.style.height = `${this.items.length * this.itemHeight}px`;
        this.scrollContainer.style.position = 'relative';

        // Create viewport
        this.viewport = document.createElement('div');
        this.viewport.style.position = 'relative';

        this.scrollContainer.appendChild(this.viewport);
        this.container.appendChild(this.scrollContainer);

        // Attach scroll handler (throttled for performance)
        this.container.addEventListener('scroll', throttle(() => {
            this.scrollTop = this.container.scrollTop;
            this.render();
        }, 16)); // ~60fps

        this.render();
    }

    render() {
        const containerHeight = this.container.clientHeight;
        const start = Math.max(0, Math.floor(this.scrollTop / this.itemHeight) - this.buffer);
        const end = Math.min(
            this.items.length,
            Math.ceil((this.scrollTop + containerHeight) / this.itemHeight) + this.buffer
        );

        // Only update if visible range changed
        if (start === this.visibleStart && end === this.visibleEnd) {
            return;
        }

        this.visibleStart = start;
        this.visibleEnd = end;

        // Clear viewport
        this.viewport.innerHTML = '';

        // Render visible items
        for (let i = start; i < end; i++) {
            const item = this.renderItem(this.items[i], i);
            item.style.position = 'absolute';
            item.style.top = `${i * this.itemHeight}px`;
            item.style.width = '100%';
            this.viewport.appendChild(item);
        }
    }

    update(items) {
        this.items = items;
        this.scrollContainer.style.height = `${items.length * this.itemHeight}px`;
        this.render();
    }
}

/**
 * Performance monitoring utility
 * Tracks page performance metrics
 */
const PerformanceMonitor = {
    metrics: {},

    // Mark the start of an operation
    start(label) {
        this.metrics[label] = performance.now();
    },

    // Mark the end and return duration
    end(label) {
        if (!this.metrics[label]) {
            console.warn(`No start mark for: ${label}`);
            return 0;
        }

        const duration = performance.now() - this.metrics[label];
        delete this.metrics[label];

        // Log slow operations (>100ms threshold for Phase 7)
        if (duration > 100) {
            console.warn(`⚠️ Slow operation: ${label} took ${duration.toFixed(2)}ms`);
        }

        return duration;
    },

    // Measure a function execution time
    measure(label, fn) {
        this.start(label);
        const result = fn();
        this.end(label);
        return result;
    },

    // Measure async function execution time
    async measureAsync(label, fn) {
        this.start(label);
        const result = await fn();
        this.end(label);
        return result;
    },

    // Get Core Web Vitals
    getCoreWebVitals() {
        if (!window.performance || !window.performance.getEntriesByType) {
            return {};
        }

        const paintEntries = performance.getEntriesByType('paint');
        const fcp = paintEntries.find(entry => entry.name === 'first-contentful-paint');

        return {
            FCP: fcp ? fcp.startTime : null, // First Contentful Paint
            LCP: null, // Largest Contentful Paint (requires PerformanceObserver)
            CLS: null, // Cumulative Layout Shift (requires PerformanceObserver)
        };
    }
};

/**
 * Memory-efficient cache with automatic cleanup
 */
class LRUCache {
    constructor(maxSize = 100) {
        this.maxSize = maxSize;
        this.cache = new Map();
    }

    get(key) {
        if (!this.cache.has(key)) return null;

        // Move to end (most recently used)
        const value = this.cache.get(key);
        this.cache.delete(key);
        this.cache.set(key, value);

        return value;
    }

    set(key, value) {
        // Delete if exists (to update position)
        if (this.cache.has(key)) {
            this.cache.delete(key);
        }

        // Add to end
        this.cache.set(key, value);

        // Remove oldest if over limit
        if (this.cache.size > this.maxSize) {
            const firstKey = this.cache.keys().next().value;
            this.cache.delete(firstKey);
        }
    }

    has(key) {
        return this.cache.has(key);
    }

    clear() {
        this.cache.clear();
    }
}

/**
 * Request deduplication
 * Prevents duplicate API requests for the same resource
 */
const RequestDeduplicator = {
    pending: new Map(),

    async fetch(key, fetchFn) {
        // Return existing promise if request is in flight
        if (this.pending.has(key)) {
            return this.pending.get(key);
        }

        // Create new request
        const promise = fetchFn()
            .finally(() => {
                // Clean up after request completes
                this.pending.delete(key);
            });

        this.pending.set(key, promise);
        return promise;
    }
};

// ====================================
// INITIALIZATION
// ====================================

// Initialize lazy loading on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => LazyLoader.init());
} else {
    LazyLoader.init();
}

// Export for use in other modules
window.PerformanceUtils = {
    debounce,
    throttle,
    rafSchedule,
    DOMBatcher,
    LazyLoader,
    VirtualScroller,
    PerformanceMonitor,
    LRUCache,
    RequestDeduplicator
};

console.log('%c⚡ Performance utilities loaded', 'color: #22c55e;');
