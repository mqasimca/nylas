/**
 * Skeleton Loader Module
 * Provides loading placeholders for improved perceived performance
 * Shows animated skeletons while content loads
 */

const SkeletonLoader = {
    /**
     * Create a skeleton element with classes
     * @param {string} tag - HTML tag name
     * @param {string[]} classes - CSS classes to add
     * @returns {HTMLElement}
     */
    createElement(tag, classes = []) {
        const el = document.createElement(tag);
        classes.forEach(cls => el.classList.add(cls));
        return el;
    },

    /**
     * Build email skeleton using DOM methods
     * @returns {HTMLElement}
     */
    buildEmailSkeleton() {
        const container = this.createElement('div', ['skeleton-email']);

        const avatar = this.createElement('div', ['skeleton-avatar']);
        container.appendChild(avatar);

        const content = this.createElement('div', ['skeleton-content']);
        content.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-sender']));
        content.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-subject']));
        content.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-preview']));
        container.appendChild(content);

        const date = this.createElement('div', ['skeleton-date']);
        container.appendChild(date);

        return container;
    },

    /**
     * Build email preview skeleton using DOM methods
     * @returns {HTMLElement}
     */
    buildPreviewSkeleton() {
        const container = this.createElement('div', ['skeleton-preview-pane']);

        const header = this.createElement('div', ['skeleton-header']);
        header.appendChild(this.createElement('div', ['skeleton-avatar-lg']));

        const headerText = this.createElement('div', ['skeleton-header-text']);
        headerText.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-title']));
        headerText.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-meta']));
        header.appendChild(headerText);
        container.appendChild(header);

        const body = this.createElement('div', ['skeleton-body']);
        body.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-full']));
        body.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-full']));
        body.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-partial']));
        body.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-full']));
        body.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-short']));
        container.appendChild(body);

        return container;
    },

    /**
     * Build event skeleton using DOM methods
     * @returns {HTMLElement}
     */
    buildEventSkeleton() {
        const container = this.createElement('div', ['skeleton-event']);
        container.appendChild(this.createElement('div', ['skeleton-time']));

        const content = this.createElement('div', ['skeleton-event-content']);
        content.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-event-title']));
        content.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-event-location']));
        container.appendChild(content);

        return container;
    },

    /**
     * Build contact skeleton using DOM methods
     * @returns {HTMLElement}
     */
    buildContactSkeleton() {
        const container = this.createElement('div', ['skeleton-contact']);
        container.appendChild(this.createElement('div', ['skeleton-avatar']));

        const info = this.createElement('div', ['skeleton-contact-info']);
        info.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-name']));
        info.appendChild(this.createElement('div', ['skeleton-line', 'skeleton-email-addr']));
        container.appendChild(info);

        return container;
    },

    /**
     * Get skeleton builder for type
     * @param {string} type - Skeleton type
     * @returns {Function}
     */
    getBuilder(type) {
        const builders = {
            email: () => this.buildEmailSkeleton(),
            emailPreview: () => this.buildPreviewSkeleton(),
            event: () => this.buildEventSkeleton(),
            contact: () => this.buildContactSkeleton(),
        };
        return builders[type] || builders.email;
    },

    /**
     * Show skeleton loaders in a container
     * @param {HTMLElement|string} container - Container or selector
     * @param {string} type - Type of skeleton (email, event, contact)
     * @param {number} count - Number of skeletons to show
     */
    show(container, type = 'email', count = 10) {
        const el = typeof container === 'string'
            ? document.querySelector(container)
            : container;

        if (!el) return;

        const builder = this.getBuilder(type);

        // Create skeleton container
        const skeletonContainer = this.createElement('div', ['skeleton-container']);
        skeletonContainer.setAttribute('role', 'status');
        skeletonContainer.setAttribute('aria-label', 'Loading content');

        // Add screen reader text
        const srText = this.createElement('span', ['sr-only']);
        srText.textContent = 'Loading...';
        skeletonContainer.appendChild(srText);

        // Add skeletons
        for (let i = 0; i < count; i++) {
            const skeleton = builder();
            skeleton.style.animationDelay = `${i * 0.05}s`;
            skeletonContainer.appendChild(skeleton);
        }

        // Clear container and add skeletons
        while (el.firstChild) {
            el.removeChild(el.firstChild);
        }
        el.appendChild(skeletonContainer);
        el.classList.add('loading');
    },

    /**
     * Hide skeleton loaders
     * @param {HTMLElement|string} container - Container or selector
     */
    hide(container) {
        const el = typeof container === 'string'
            ? document.querySelector(container)
            : container;

        if (!el) return;

        const skeletonContainer = el.querySelector('.skeleton-container');
        if (skeletonContainer) {
            skeletonContainer.classList.add('fade-out');
            setTimeout(() => {
                skeletonContainer.remove();
                el.classList.remove('loading');
            }, 200);
        }
    },

    /**
     * Replace skeleton with actual content element
     * @param {HTMLElement|string} container - Container or selector
     * @param {HTMLElement} content - Content element to insert
     */
    replace(container, content) {
        const el = typeof container === 'string'
            ? document.querySelector(container)
            : container;

        if (!el || !content) return;

        this.hide(el);

        setTimeout(() => {
            while (el.firstChild) {
                el.removeChild(el.firstChild);
            }
            el.appendChild(content);
            el.classList.add('content-loaded');
            setTimeout(() => el.classList.remove('content-loaded'), 300);
        }, 200);
    },

    /**
     * Create inline skeleton text element
     * @param {number|string} width - Width in pixels or percentage
     * @returns {HTMLElement}
     */
    inlineText(width = '100%') {
        const span = this.createElement('span', ['skeleton-inline']);
        span.style.width = typeof width === 'number' ? `${width}px` : width;
        return span;
    },

    /**
     * Wrap async load with skeleton
     * @param {Function} loadFn - Async function that loads content
     * @param {HTMLElement} container - Container element
     * @param {string} type - Skeleton type
     */
    async wrap(loadFn, container, type = 'email') {
        this.show(container, type);

        try {
            const result = await loadFn();
            this.hide(container);
            return result;
        } catch (error) {
            this.hide(container);
            throw error;
        }
    },
};

// Export for use
if (typeof window !== 'undefined') {
    window.SkeletonLoader = SkeletonLoader;
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = SkeletonLoader;
}
