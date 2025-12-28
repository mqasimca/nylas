// ====================================
// EMAIL MODULE
// Handles email list, rendering, and navigation
// ====================================

// Email rendering helpers
const EmailRenderer = {
    // Format timestamp to relative time
    formatTime(timestamp) {
        const date = new Date(timestamp * 1000);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m`;
        if (diffHours < 24) return `${diffHours}h`;
        if (diffDays < 7) return `${diffDays}d`;

        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    },

    // Get sender display info
    getSenderInfo(from) {
        if (!from || from.length === 0) {
            return { name: 'Unknown', initials: '?', email: '' };
        }
        const sender = from[0];
        const name = sender.name || sender.email.split('@')[0];
        const initials = name.split(' ')
            .map(n => n[0])
            .join('')
            .substring(0, 2)
            .toUpperCase();
        return { name, initials, email: sender.email };
    },

    // Render a single email item for the list
    renderEmailItem(email, isSelected = false) {
        const sender = this.getSenderInfo(email.from);
        const time = this.formatTime(email.date);
        const hasAttachment = email.attachments && email.attachments.length > 0;

        const div = document.createElement('div');
        div.className = `email-item${isSelected ? ' selected' : ''}${email.unread ? ' unread' : ''}`;
        div.setAttribute('data-email-id', email.id);
        div.setAttribute('role', 'option');
        div.setAttribute('tabindex', '-1');
        div.setAttribute('aria-selected', isSelected ? 'true' : 'false');

        div.innerHTML = `
            <div class="email-avatar" style="background: var(--gradient-${Math.floor(Math.random() * 5) + 1})">
                ${sender.initials}
            </div>
            <div class="email-content">
                <div class="email-header">
                    <span class="email-sender">${this.escapeHtml(sender.name)}</span>
                    <span class="email-time">${time}</span>
                </div>
                <div class="email-subject">${this.escapeHtml(email.subject || '(No Subject)')}</div>
                <div class="email-preview">${this.escapeHtml(email.snippet || '')}</div>
            </div>
            <div class="email-actions-mini">
                ${email.starred ? '<span class="starred" title="Starred">&#9733;</span>' : ''}
                ${hasAttachment ? '<span class="attachment" title="Has attachments">&#128206;</span>' : ''}
            </div>
        `;

        return div;
    },

    // Render folder item
    renderFolderItem(folder, isActive = false) {
        const icons = {
            inbox: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="M2 12h6l2 2h4l2-2h6"/></svg>',
            sent: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/></svg>',
            drafts: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 3H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
            trash: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>',
            spam: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/></svg>',
            archive: '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="2" y="4" width="20" height="5" rx="1"/><path d="M4 9v9a2 2 0 002 2h12a2 2 0 002-2V9M10 13h4"/></svg>'
        };

        const icon = icons[folder.system_folder] || icons.inbox;

        const li = document.createElement('li');
        li.className = `folder-item${isActive ? ' active' : ''}`;
        li.setAttribute('data-folder-id', folder.id);

        li.innerHTML = `
            <span class="folder-icon">${icon}</span>
            <span class="folder-name">${this.escapeHtml(folder.name)}</span>
            ${folder.unread_count > 0 ? `<span class="folder-count">${folder.unread_count}</span>` : ''}
        `;

        return li;
    },

    // Escape HTML to prevent XSS
    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
};

// Email List Manager with Virtual Scrolling & Optimistic Updates
const EmailListManager = {
    currentFolder: 'INBOX',
    currentFilter: 'all', // 'all', 'primary', 'vip', 'newsletters', 'updates', 'unread'
    emails: [],
    filteredEmails: [], // Emails after applying filter
    vipSenders: [], // List of VIP email addresses
    selectedEmailId: null,
    selectedEmailFull: null, // Store full email data for reply/forward
    nextCursor: null,
    hasMore: false,
    isLoading: false,

    // Virtual scrolling state
    virtualScroll: {
        enabled: true,
        itemHeight: 76,           // Height of each email item in pixels
        bufferSize: 5,            // Extra items to render above/below viewport
        visibleStart: 0,
        visibleEnd: 0,
        scrollContainer: null,
        totalHeight: 0
    },

    // Request cache for performance
    cache: {
        emails: new Map(),        // emailId -> full email data
        folders: null,
        foldersTimestamp: 0,
        cacheDuration: 60000      // 1 minute cache
    },

    // Pending operations for optimistic updates
    pendingOperations: new Map(), // operationId -> { type, emailId, originalState }

    // Simplified filters: All, VIP, Unread

    async init() {
        // Set up event listeners first (UI is ready immediately)
        this.setupEventListeners();
        this.setupFilterTabs();

        // Load folders first, then emails from inbox (sequential to avoid rate limits)
        try {
            await this.loadFolders();
        } catch (error) {
            console.error('Failed to load folders:', error);
        }

        // Load VIP senders list in background
        this.loadVIPSenders().catch(err => console.error('Failed to load VIP senders:', err));

        try {
            // Load inbox emails by default with limit 10
            await this.loadEmails('INBOX');
        } catch (error) {
            console.error('Failed to load emails:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load emails. Will retry...');
            }
            // Retry after delay
            setTimeout(() => this.loadEmails('INBOX'), 3000);
        }

        console.log('%c📧 Email module loaded', 'color: #22c55e;');
    },

    // Set up filter tab click handlers
    setupFilterTabs() {
        const filterTabs = document.querySelectorAll('.filter-tab');
        filterTabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const filter = tab.textContent.toLowerCase().trim();
                this.setFilter(filter);
            });
        });
    },

    // Load VIP senders from backend
    async loadVIPSenders() {
        try {
            const response = await fetch('/api/inbox/vip');
            if (response.ok) {
                const data = await response.json();
                this.vipSenders = data.vip_senders || [];
                console.log('Loaded VIP senders:', this.vipSenders.length);
            }
        } catch (error) {
            console.error('Failed to load VIP senders:', error);
        }
    },

    // Set the current filter and update display
    setFilter(filter) {
        this.currentFilter = filter;

        // Update tab UI
        const filterTabs = document.querySelectorAll('.filter-tab');
        filterTabs.forEach(tab => {
            const tabFilter = tab.textContent.toLowerCase().trim();
            tab.classList.toggle('active', tabFilter === filter);
        });

        // Apply filter and re-render
        this.applyFilter();
        this.renderEmails();
    },

    // Check if email is from a VIP sender
    isVIP(email) {
        const senderEmail = email.from && email.from[0] ? email.from[0].email.toLowerCase() : '';
        return this.vipSenders.some(vip => senderEmail.includes(vip.toLowerCase()));
    },

    // Apply filter to emails (simplified: All, VIP, Unread)
    applyFilter() {
        switch (this.currentFilter) {
            case 'vip':
                this.filteredEmails = this.emails.filter(email => this.isVIP(email));
                break;
            case 'unread':
                this.filteredEmails = this.emails.filter(email => email.unread);
                break;
            default: // 'all'
                this.filteredEmails = [...this.emails];
                break;
        }

        // Update filter tab counts
        this.updateFilterCounts();
    },

    // Update counts on filter tabs
    updateFilterCounts() {
        const counts = {
            all: this.emails.length,
            vip: this.emails.filter(e => this.isVIP(e)).length,
            unread: this.emails.filter(e => e.unread).length
        };

        // Update DOM
        const tabs = document.querySelectorAll('.filter-tab');
        tabs.forEach(tab => {
            const filter = tab.dataset.filter || tab.textContent.toLowerCase().trim();
            const count = counts[filter];
            let countBadge = tab.querySelector('.filter-count');

            if (count > 0 && filter !== 'all') {
                if (!countBadge) {
                    countBadge = document.createElement('span');
                    countBadge.className = 'filter-count';
                    tab.appendChild(countBadge);
                }
                countBadge.textContent = count > 99 ? '99+' : count;
            } else if (countBadge) {
                countBadge.remove();
            }
        });
    },

    setupEventListeners() {
        // Folder click handling - use #folderList or .folder-group
        const folderList = document.getElementById('folderList') || document.querySelector('.folder-group');
        if (folderList) {
            folderList.addEventListener('click', (e) => {
                const folderItem = e.target.closest('.folder-item');
                if (folderItem) {
                    const folderId = folderItem.getAttribute('data-folder-id');
                    const folderName = folderItem.getAttribute('data-folder-name');
                    if (folderId) {
                        this.selectFolder(folderId, folderName);
                    }
                }
            });
        }

        // Email click handling
        const emailList = document.querySelector('.email-list');
        if (emailList) {
            emailList.addEventListener('click', (e) => {
                const emailItem = e.target.closest('.email-item');
                if (emailItem) {
                    const emailId = emailItem.getAttribute('data-email-id');
                    this.selectEmail(emailId);
                }
            });
        }

        // Virtual scroll + infinite scroll handling
        const emailListContainer = document.querySelector('.email-list');
        if (emailListContainer) {
            this.virtualScroll.scrollContainer = emailListContainer;

            // Debounced scroll handler for performance
            let scrollTimeout = null;
            emailListContainer.addEventListener('scroll', () => {
                // Cancel previous timeout
                if (scrollTimeout) cancelAnimationFrame(scrollTimeout);

                // Use requestAnimationFrame for smooth scrolling
                scrollTimeout = requestAnimationFrame(() => {
                    // Virtual scroll update
                    if (this.virtualScroll.enabled) {
                        this.updateVirtualScroll();
                    }

                    // Infinite scroll - load more when near bottom
                    if (this.hasMore && !this.isLoading) {
                        const { scrollTop, scrollHeight, clientHeight } = emailListContainer;
                        if (scrollTop + clientHeight >= scrollHeight - 200) {
                            this.loadMore();
                        }
                    }
                });
            });
        }
    },

    // Initialize virtual scroll container
    initVirtualScroll() {
        const container = this.virtualScroll.scrollContainer;
        if (!container) return;

        // Create spacer elements for virtual scroll
        let topSpacer = container.querySelector('.virtual-spacer-top');
        let bottomSpacer = container.querySelector('.virtual-spacer-bottom');

        if (!topSpacer) {
            topSpacer = document.createElement('div');
            topSpacer.className = 'virtual-spacer-top';
            container.prepend(topSpacer);
        }

        if (!bottomSpacer) {
            bottomSpacer = document.createElement('div');
            bottomSpacer.className = 'virtual-spacer-bottom';
            container.append(bottomSpacer);
        }
    },

    // Update virtual scroll visible range
    updateVirtualScroll() {
        const container = this.virtualScroll.scrollContainer;
        if (!container) return;

        const { itemHeight, bufferSize } = this.virtualScroll;
        const scrollTop = container.scrollTop;
        const viewportHeight = container.clientHeight;

        const displayEmails = this.getDisplayEmails();
        const totalItems = displayEmails.length;

        // Calculate visible range
        const visibleStart = Math.max(0, Math.floor(scrollTop / itemHeight) - bufferSize);
        const visibleEnd = Math.min(totalItems, Math.ceil((scrollTop + viewportHeight) / itemHeight) + bufferSize);

        // Only re-render if visible range changed significantly
        if (visibleStart !== this.virtualScroll.visibleStart || visibleEnd !== this.virtualScroll.visibleEnd) {
            this.virtualScroll.visibleStart = visibleStart;
            this.virtualScroll.visibleEnd = visibleEnd;
            this.renderVirtualEmails();
        }
    },

    // Render only visible emails (virtual scrolling)
    renderVirtualEmails() {
        const container = this.virtualScroll.scrollContainer;
        if (!container) return;

        const { itemHeight, visibleStart, visibleEnd } = this.virtualScroll;
        const displayEmails = this.getDisplayEmails();
        const totalItems = displayEmails.length;

        // Update spacers
        const topSpacer = container.querySelector('.virtual-spacer-top');
        const bottomSpacer = container.querySelector('.virtual-spacer-bottom');

        if (topSpacer) {
            topSpacer.style.height = `${visibleStart * itemHeight}px`;
        }
        if (bottomSpacer) {
            bottomSpacer.style.height = `${Math.max(0, (totalItems - visibleEnd) * itemHeight)}px`;
        }

        // Get visible emails
        const visibleEmails = displayEmails.slice(visibleStart, visibleEnd);

        // Create/update email items between spacers
        const fragment = document.createDocumentFragment();
        visibleEmails.forEach((email, i) => {
            const isSelected = email.id === this.selectedEmailId;
            const item = EmailRenderer.renderEmailItem(email, isSelected);
            item.dataset.virtualIndex = visibleStart + i;
            fragment.appendChild(item);
        });

        // Remove old email items (but keep spacers)
        Array.from(container.children).forEach(child => {
            if (!child.classList.contains('virtual-spacer-top') &&
                !child.classList.contains('virtual-spacer-bottom') &&
                !child.classList.contains('empty-state')) {
                child.remove();
            }
        });

        // Insert new items after top spacer
        if (topSpacer && topSpacer.nextSibling !== bottomSpacer) {
            topSpacer.after(fragment);
        } else if (topSpacer) {
            topSpacer.after(fragment);
        } else {
            container.prepend(fragment);
        }
    },

    // Get emails to display (with filter applied)
    getDisplayEmails() {
        return this.filteredEmails.length > 0 || this.currentFilter !== 'all'
            ? this.filteredEmails
            : this.emails;
    },

    async loadFolders() {
        try {
            const data = await AirAPI.getFolders();
            this.renderFolders(data.folders || []);
        } catch (error) {
            console.error('Failed to load folders:', error);
            // Keep using template-rendered folders
        }
    },

    renderFolders(folders) {
        const folderList = document.getElementById('folderList') || document.querySelector('.folder-group');
        if (!folderList || folders.length === 0) return;

        // Primary folders to show directly (in order)
        const primaryFolderIds = ['INBOX', 'STARRED', 'SENT', 'DRAFT', 'TRASH', 'SPAM'];

        // Filter out Gmail category folders and system folders
        const filteredFolders = folders.filter(f => {
            const id = (f.id || '').toUpperCase();
            if (id.startsWith('CATEGORY_')) return false;
            if (id === 'UNREAD' || id === 'CHAT' || id === 'IMPORTANT' || id === 'SNOOZED' || id === 'SCHEDULED') return false;
            return true;
        });

        // Separate primary and other folders
        const primaryFolders = [];
        const otherFolders = [];

        filteredFolders.forEach(f => {
            const id = (f.id || '').toUpperCase();
            if (primaryFolderIds.includes(id)) {
                primaryFolders.push(f);
            } else {
                otherFolders.push(f);
            }
        });

        // Sort primary folders by predefined order
        primaryFolders.sort((a, b) => {
            const aIdx = primaryFolderIds.indexOf((a.id || '').toUpperCase());
            const bIdx = primaryFolderIds.indexOf((b.id || '').toUpperCase());
            return aIdx - bIdx;
        });

        // Sort other folders alphabetically
        otherFolders.sort((a, b) => (a.name || a.id).localeCompare(b.name || b.id));

        folderList.innerHTML = '';

        // Render primary folders
        primaryFolders.forEach(folder => {
            const isActive = folder.id === this.currentFolder || (folder.id.toUpperCase() === 'INBOX' && !this.currentFolder);
            const item = this.createFolderElement(folder, isActive);
            folderList.appendChild(item);
        });

        // Add "More" dropdown if there are other folders
        if (otherFolders.length > 0) {
            const moreContainer = this.createMoreDropdown(otherFolders);
            folderList.appendChild(moreContainer);
        }
    },

    createMoreDropdown(folders) {
        const container = document.createElement('div');
        container.className = 'folder-more-container';

        const trigger = document.createElement('div');
        trigger.className = 'folder-item folder-more-trigger';
        trigger.innerHTML = `
            <span class="folder-icon">📂</span>
            <span>More</span>
            <span class="folder-count">${folders.length}</span>
            <svg class="folder-more-arrow" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="m6 9 6 6 6-6"/>
            </svg>
        `;

        const dropdown = document.createElement('div');
        dropdown.className = 'folder-more-dropdown hidden';

        folders.forEach(folder => {
            const item = this.createFolderElement(folder, false);
            item.classList.add('folder-dropdown-item');
            dropdown.appendChild(item);
        });

        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdown.classList.toggle('hidden');
            trigger.classList.toggle('expanded');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!container.contains(e.target)) {
                dropdown.classList.add('hidden');
                trigger.classList.remove('expanded');
            }
        });

        container.appendChild(trigger);
        container.appendChild(dropdown);
        return container;
    },

    createFolderElement(folder, isActive = false) {
        const icons = {
            'INBOX': '📥',
            'SENT': '📤',
            'DRAFT': '📝',
            'DRAFTS': '📝',
            'TRASH': '🗑️',
            'SPAM': '⚠️',
            'STARRED': '⭐',
            'SNOOZED': '🕐',
            'SCHEDULED': '📅',
            'ARCHIVE': '📦'
        };

        const folderId = (folder.id || '').toUpperCase();
        const icon = icons[folderId] || '📁';

        // Clean up display name
        let displayName = folder.name || folder.id;
        // Capitalize first letter, lowercase rest
        displayName = displayName.charAt(0).toUpperCase() + displayName.slice(1).toLowerCase();

        const div = document.createElement('div');
        div.className = `folder-item${isActive ? ' active' : ''}`;
        div.setAttribute('data-folder-id', folder.id);
        div.setAttribute('data-folder-name', displayName);
        div.setAttribute('role', 'listitem');
        div.setAttribute('tabindex', '0');
        if (isActive) div.setAttribute('aria-current', 'true');

        const count = folder.unread_count || folder.total_count || 0;
        div.innerHTML = `
            <span class="folder-icon">${icon}</span>
            <span>${this.escapeHtml(displayName)}</span>
            ${count > 0 ? `<span class="folder-count${folder.unread_count > 0 ? ' unread' : ''}">${count}</span>` : ''}
        `;

        return div;
    },

    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    async loadEmails(folder = null) {
        if (this.isLoading) return;
        this.isLoading = true;

        console.log('[loadEmails] Starting...', { folder, limit: 50 });

        try {
            const options = { limit: 50 }; // Increased from 10 to 50 to fill viewport
            if (folder) {
                this.currentFolder = folder;
                options.folder = folder;
            }

            const data = await AirAPI.getEmails(options);
            console.log('[loadEmails] API response:', {
                emailCount: data.emails?.length || 0,
                hasMore: data.has_more,
                nextCursor: data.next_cursor
            });

            this.emails = data.emails || [];
            this.nextCursor = data.next_cursor;
            this.hasMore = data.has_more;

            // Apply current filter
            this.applyFilter();
            this.renderEmails();
        } catch (error) {
            console.error('[loadEmails] Failed to load emails:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load emails');
            }
        } finally {
            this.isLoading = false;
            console.log('[loadEmails] Complete. Triggering ensureScrollable...');
            // Auto-load more AFTER isLoading is set to false
            // Use setTimeout to ensure DOM has updated
            setTimeout(() => this.ensureScrollable(), 100);
        }
    },

    async loadMore() {
        if (!this.hasMore || !this.nextCursor || this.isLoading) return;
        this.isLoading = true;

        try {
            const data = await AirAPI.getEmails({
                folder: this.currentFolder,
                cursor: this.nextCursor
            });

            this.emails = [...this.emails, ...(data.emails || [])];
            this.nextCursor = data.next_cursor;
            this.hasMore = data.has_more;

            // Apply current filter
            this.applyFilter();
            this.renderEmails(true); // Append mode
        } catch (error) {
            console.error('Failed to load more emails:', error);
        } finally {
            this.isLoading = false;
            // Auto-load more AFTER isLoading is set to false
            // Use setTimeout to ensure DOM has updated
            setTimeout(() => this.ensureScrollable(), 100);
        }
    },

    // Ensure the email list has enough content to be scrollable
    // Auto-loads more emails if content doesn't fill viewport
    ensureScrollable() {
        const emailList = document.querySelector('.email-list');

        console.log('[ensureScrollable] Starting check...', {
            hasEmailList: !!emailList,
            hasMore: this.hasMore,
            isLoading: this.isLoading,
            emailCount: this.emails.length,
            nextCursor: this.nextCursor
        });

        if (!emailList) {
            console.warn('[ensureScrollable] Email list element not found');
            return;
        }

        if (!this.hasMore) {
            console.log('[ensureScrollable] No more emails to load (hasMore=false)');
            return;
        }

        if (this.isLoading) {
            console.log('[ensureScrollable] Already loading, skipping');
            return;
        }

        // Check if content fills viewport (has scrollbar)
        const scrollHeight = emailList.scrollHeight;
        const clientHeight = emailList.clientHeight;
        const needsMore = scrollHeight <= clientHeight;

        console.log('[ensureScrollable] Viewport check:', {
            scrollHeight,
            clientHeight,
            needsMore,
            hasScrollbar: scrollHeight > clientHeight
        });

        if (needsMore) {
            console.log('[ensureScrollable] Loading more emails to fill viewport...');
            setTimeout(() => this.loadMore(), 100);
        } else {
            console.log('[ensureScrollable] Viewport is full, stopping auto-load');
        }
    },

    renderEmails(append = false) {
        const emailList = document.querySelector('.email-list');
        if (!emailList) return;

        // Use filtered emails for display
        const displayEmails = this.getDisplayEmails();

        console.log('[renderEmails]', {
            totalEmails: this.emails.length,
            filteredEmails: this.filteredEmails.length,
            currentFilter: this.currentFilter,
            displayCount: displayEmails.length,
            append,
            virtualScrollEnabled: this.virtualScroll.enabled
        });

        if (displayEmails.length === 0 && !append) {
            // Clear spacers for empty state
            emailList.innerHTML = '';

            const isInbox = this.currentFolder === 'INBOX' || this.currentFolder === 'inbox';
            const emptyMessages = {
                'vip': { icon: '⭐', title: 'No VIP emails', message: 'Add VIP senders to see their emails here' },
                'unread': { icon: '✓', title: 'All caught up!', message: 'No unread emails', celebrate: isInbox },
                'all': isInbox
                    ? { icon: '🎉', title: 'Inbox Zero!', message: 'You\'ve conquered your inbox. Take a moment to celebrate!', celebrate: true }
                    : { icon: '📭', title: 'No emails', message: 'This folder is empty' }
            };
            const msg = emptyMessages[this.currentFilter] || emptyMessages.all;
            emailList.innerHTML = `
                <div class="empty-state inbox-zero${msg.celebrate ? ' celebration' : ''}">
                    <div class="empty-icon">${msg.icon}</div>
                    <div class="empty-title">${msg.title}</div>
                    <div class="empty-message">${msg.message}</div>
                    ${msg.celebrate ? '<div class="inbox-zero-subtitle">✨ Enjoy the moment</div>' : ''}
                </div>
            `;

            // Trigger celebration confetti for Inbox Zero
            if (msg.celebrate && typeof window.celebrateInboxZero === 'function') {
                setTimeout(() => window.celebrateInboxZero(), 300);
            }
            return;
        }

        // Use virtual scrolling for large lists
        if (this.virtualScroll.enabled && displayEmails.length > 20) {
            if (!append) {
                emailList.innerHTML = '';
                this.initVirtualScroll();
            }

            // Calculate initial visible range
            const { itemHeight, bufferSize } = this.virtualScroll;
            const viewportHeight = emailList.clientHeight || 600;
            this.virtualScroll.visibleStart = 0;
            this.virtualScroll.visibleEnd = Math.min(
                displayEmails.length,
                Math.ceil(viewportHeight / itemHeight) + bufferSize * 2
            );

            this.renderVirtualEmails();
        } else {
            // Standard rendering for small lists
            if (!append) {
                emailList.innerHTML = '';
            }

            const fragment = document.createDocumentFragment();
            displayEmails.forEach(email => {
                const isSelected = email.id === this.selectedEmailId;
                const item = EmailRenderer.renderEmailItem(email, isSelected);
                fragment.appendChild(item);
            });

            if (append) {
                emailList.appendChild(fragment);
            } else {
                emailList.appendChild(fragment);
            }
        }
    },

    async selectFolder(folderId, folderName = null) {
        // Update folder UI - mark active
        document.querySelectorAll('.folder-item').forEach(item => {
            item.classList.toggle('active', item.getAttribute('data-folder-id') === folderId);
        });

        // Update title
        const titleEl = document.querySelector('.list-title');
        if (titleEl && folderName) {
            titleEl.textContent = folderName;
        }

        this.currentFolder = folderId;
        this.selectedEmailId = null;
        this.emails = [];
        this.nextCursor = null;

        // Clear email detail pane
        const detailPane = document.querySelector('.email-detail');
        if (detailPane) {
            detailPane.innerHTML = '<div class="empty-state"><div class="empty-icon">📬</div><div class="empty-title">Select an email</div><div class="empty-message">Click on an email to view its contents</div></div>';
        }

        await this.loadEmails(folderId);
    },

    async selectEmail(emailId) {
        this.selectedEmailId = emailId;

        // Update list selection
        document.querySelectorAll('.email-item').forEach(item => {
            const isSelected = item.getAttribute('data-email-id') === emailId;
            item.classList.toggle('selected', isSelected);
            item.setAttribute('aria-selected', isSelected ? 'true' : 'false');
        });

        // Load full email
        try {
            const email = await AirAPI.getEmail(emailId);
            // Store full email for reply/forward (includes thread_id)
            this.selectedEmailFull = email;
            this.renderEmailDetail(email);

            // Render email body in sandboxed iframe (async, after DOM update)
            requestAnimationFrame(() => {
                this.renderEmailBodyIframe(email.id, email.body);
            });

            // Mark as read if unread
            if (email.unread) {
                await this.markAsRead(emailId);
            }
        } catch (error) {
            console.error('Failed to load email:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load email');
            }
        }
    },

    renderEmailDetail(email) {
        const detailPane = document.querySelector('.email-detail');
        if (!detailPane) return;

        const sender = EmailRenderer.getSenderInfo(email.from);
        const time = new Date(email.date * 1000).toLocaleString();

        const toList = (email.to || [])
            .map(p => p.name || p.email)
            .join(', ');

        detailPane.innerHTML = `
            <div class="email-detail-header">
                <div class="email-detail-subject">${EmailRenderer.escapeHtml(email.subject || '(No Subject)')}</div>
                <div class="email-detail-meta">
                    <div class="email-detail-avatar" style="background: var(--gradient-1)">${sender.initials}</div>
                    <div class="email-detail-info">
                        <div class="email-detail-from">${EmailRenderer.escapeHtml(sender.name)} <span class="email-detail-email">&lt;${EmailRenderer.escapeHtml(sender.email)}&gt;</span></div>
                        <div class="email-detail-to">To: ${EmailRenderer.escapeHtml(toList || 'me')}</div>
                    </div>
                    <div class="email-detail-time">${time}</div>
                </div>
            </div>
            <div class="email-detail-actions">
                <button class="action-btn" onclick="EmailListManager.replyToEmail('${email.id}')" title="Reply">
                    <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path d="M9 17H4a2 2 0 01-2-2V5a2 2 0 012-2h16a2 2 0 012 2v10a2 2 0 01-2 2h-5l-5 5v-5z"/>
                    </svg>
                    Reply
                </button>
                <button class="action-btn" onclick="EmailListManager.toggleStar('${email.id}')" title="Star">
                    <svg width="16" height="16" fill="${email.starred ? 'currentColor' : 'none'}" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
                    </svg>
                    ${email.starred ? 'Starred' : 'Star'}
                </button>
                <button class="action-btn" onclick="EmailListManager.archiveEmail('${email.id}')" title="Archive">
                    <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <rect x="2" y="4" width="20" height="5" rx="1"/>
                        <path d="M4 9v9a2 2 0 002 2h12a2 2 0 002-2V9M10 13h4"/>
                    </svg>
                    Archive
                </button>
                <button class="action-btn" onclick="EmailListManager.deleteEmail('${email.id}')" title="Delete">
                    <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                    </svg>
                    Delete
                </button>
                <button class="action-btn ai-btn" id="summarizeBtn-${email.id}" onclick="EmailListManager.summarizeWithAI('${email.id}')" title="Summarize with AI">
                    <svg class="ai-icon" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path d="M12 2a10 10 0 100 20 10 10 0 000-20z"/>
                        <path d="M12 6v6l4 2"/>
                    </svg>
                    <svg class="ai-spinner" width="16" height="16" viewBox="0 0 24 24" style="display:none">
                        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none" stroke-dasharray="31.4" stroke-dashoffset="10">
                            <animateTransform attributeName="transform" type="rotate" from="0 12 12" to="360 12 12" dur="1s" repeatCount="indefinite"/>
                        </circle>
                    </svg>
                    <span class="ai-btn-text">✨ Summarize</span>
                </button>
            </div>
            <div class="smart-replies-container" id="smartReplies-${email.id}">
                <button class="smart-replies-trigger" onclick="EmailListManager.loadSmartReplies('${email.id}')">
                    <span class="smart-replies-icon">💬</span>
                    <span>Get smart reply suggestions</span>
                </button>
            </div>
            ${email.attachments && email.attachments.length > 0 ? `
                <div class="email-detail-attachments">
                    <div class="attachments-header">Attachments (${email.attachments.length})</div>
                    <div class="attachments-list">
                        ${email.attachments.map(a => `
                            <div class="attachment-item">
                                <span class="attachment-icon">&#128206;</span>
                                <span class="attachment-name">${EmailRenderer.escapeHtml(a.filename)}</span>
                                <span class="attachment-size">${this.formatSize(a.size)}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            ` : ''}
            <div class="email-detail-body">
                <div class="email-iframe-container" id="emailBodyContainer-${email.id}">
                    <div class="email-loading-state">
                        <div class="email-loading-spinner"></div>
                        <span>Loading email...</span>
                    </div>
                </div>
            </div>
        `;
    },

    // Render email body into a sandboxed iframe for security and proper HTML rendering
    renderEmailBodyIframe(emailId, bodyHtml) {
        const container = document.getElementById(`emailBodyContainer-${emailId}`);
        if (!container) return;

        // Create sandboxed iframe - no scripts, no forms, no popups
        const iframe = document.createElement('iframe');
        iframe.className = 'email-body-iframe';
        iframe.setAttribute('sandbox', 'allow-same-origin'); // Minimal permissions - no scripts
        iframe.setAttribute('title', 'Email content');
        iframe.setAttribute('loading', 'lazy');

        // Build the email content with embedded styles for proper rendering
        const emailContent = this.buildEmailIframeContent(bodyHtml);

        // Use srcdoc for security - content is isolated
        iframe.srcdoc = emailContent;

        // Handle iframe load
        iframe.onload = () => {
            // Process and hide broken/tracking images first
            this.processIframeImages(iframe);
            // Auto-resize iframe to content height
            this.resizeIframeToContent(iframe);
            // Add loaded class for fade-in animation
            container.classList.add('loaded');
            // Make links open in new tab
            this.processIframeLinks(iframe);
        };

        iframe.onerror = () => {
            container.innerHTML = `
                <div class="email-error-state">
                    <span class="error-icon">⚠️</span>
                    <span>Failed to load email content</span>
                </div>
            `;
        };

        // Clear loading state and add iframe
        container.innerHTML = '';
        container.appendChild(iframe);
    },

    // Build the HTML content for the email iframe with embedded styles
    buildEmailIframeContent(bodyHtml) {
        // Default to empty paragraph if no content
        const content = bodyHtml || '<p style="color: #71717a; font-style: italic;">No content</p>';

        return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        /* Reset and base styles */
        *, *::before, *::after {
            box-sizing: border-box;
        }

        html, body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            font-size: 15px;
            line-height: 1.65;
            color: #1a1a2e;
            background: transparent;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }

        body {
            padding: 4px;
        }

        /* Typography */
        p {
            margin: 0 0 1em 0;
        }

        p:last-child {
            margin-bottom: 0;
        }

        h1, h2, h3, h4, h5, h6 {
            margin: 0 0 0.5em 0;
            font-weight: 600;
            line-height: 1.3;
        }

        h1 { font-size: 1.75em; }
        h2 { font-size: 1.5em; }
        h3 { font-size: 1.25em; }

        /* Links */
        a {
            color: #6366f1;
            text-decoration: none;
            transition: color 0.15s ease;
        }

        a:hover {
            color: #4f46e5;
            text-decoration: underline;
        }

        /* Images */
        img {
            max-width: 100%;
            height: auto;
        }

        /* Hide broken images - applied via JS */
        img.broken-image,
        img.tracking-pixel {
            display: none !important;
            visibility: hidden !important;
            width: 0 !important;
            height: 0 !important;
            opacity: 0 !important;
        }

        /* Tables - common in HTML emails */
        table {
            border-collapse: collapse;
            max-width: 100%;
            width: auto;
        }

        td, th {
            padding: 8px 12px;
            text-align: left;
            vertical-align: top;
        }

        /* Blockquotes - for email threads */
        blockquote {
            margin: 1em 0;
            padding: 0.5em 0 0.5em 1em;
            border-left: 3px solid #e5e7eb;
            color: #52525b;
        }

        /* Code blocks */
        pre, code {
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.9em;
            background: #f4f4f5;
            border-radius: 4px;
        }

        code {
            padding: 0.15em 0.4em;
        }

        pre {
            padding: 1em;
            overflow-x: auto;
            white-space: pre-wrap;
            word-wrap: break-word;
        }

        pre code {
            padding: 0;
            background: transparent;
        }

        /* Lists */
        ul, ol {
            margin: 0.5em 0;
            padding-left: 1.5em;
        }

        li {
            margin-bottom: 0.25em;
        }

        /* Horizontal rules */
        hr {
            border: none;
            border-top: 1px solid #e5e7eb;
            margin: 1.5em 0;
        }

        /* Hide tracking pixels and invisible images */
        img[width="1"], img[height="1"],
        img[width="0"], img[height="0"],
        img[width="1px"], img[height="1px"],
        img[style*="display:none"],
        img[style*="display: none"],
        img[style*="width: 1px"],
        img[style*="height: 1px"],
        img[style*="width:1px"],
        img[style*="height:1px"],
        img[src*="tracking"],
        img[src*="beacon"],
        img[src*="pixel"],
        img[src*="open."],
        img[src*="/o/"],
        img[src*="mailtrack"] {
            display: none !important;
            width: 0 !important;
            height: 0 !important;
        }

        /* Force word wrapping for long URLs/text */
        * {
            word-wrap: break-word;
            overflow-wrap: break-word;
        }
    </style>
</head>
<body>${content}</body>
</html>`;
    },

    // Resize iframe to fit its content
    resizeIframeToContent(iframe) {
        try {
            const doc = iframe.contentDocument || iframe.contentWindow?.document;
            if (doc && doc.body) {
                // Get the actual content height
                const height = Math.max(
                    doc.body.scrollHeight,
                    doc.body.offsetHeight,
                    doc.documentElement?.scrollHeight || 0,
                    doc.documentElement?.offsetHeight || 0
                );
                // Set minimum height and add small buffer
                iframe.style.height = Math.max(height + 20, 100) + 'px';
            }
        } catch (e) {
            // Cross-origin restrictions - use default height
            console.warn('Could not resize iframe:', e);
            iframe.style.height = '400px';
        }
    },

    // Process links in iframe to open in new tab
    processIframeLinks(iframe) {
        try {
            const doc = iframe.contentDocument || iframe.contentWindow?.document;
            if (doc) {
                const links = doc.querySelectorAll('a[href]');
                links.forEach(link => {
                    link.setAttribute('target', '_blank');
                    link.setAttribute('rel', 'noopener noreferrer');
                });
            }
        } catch (e) {
            console.warn('Could not process iframe links:', e);
        }
    },

    // Process images in iframe - hide broken and tracking images
    processIframeImages(iframe) {
        try {
            const doc = iframe.contentDocument || iframe.contentWindow?.document;
            if (!doc) return;

            const images = doc.querySelectorAll('img');
            images.forEach(img => {
                // Check if image is a tracking pixel by size
                const isTrackingPixel = (
                    img.naturalWidth <= 3 ||
                    img.naturalHeight <= 3 ||
                    img.width <= 3 ||
                    img.height <= 3 ||
                    (img.getAttribute('width') && parseInt(img.getAttribute('width')) <= 3) ||
                    (img.getAttribute('height') && parseInt(img.getAttribute('height')) <= 3)
                );

                if (isTrackingPixel) {
                    img.classList.add('tracking-pixel');
                    img.style.display = 'none';
                    return;
                }

                // Handle broken images
                img.onerror = () => {
                    img.classList.add('broken-image');
                    img.style.display = 'none';
                };

                // Check if already broken (naturalWidth is 0 for broken images)
                if (img.complete && img.naturalWidth === 0) {
                    img.classList.add('broken-image');
                    img.style.display = 'none';
                }
            });

            // Re-run check after a short delay for images still loading
            setTimeout(() => {
                images.forEach(img => {
                    if (img.complete && img.naturalWidth === 0 && !img.classList.contains('broken-image')) {
                        img.classList.add('broken-image');
                        img.style.display = 'none';
                    }
                    // Final check for tiny images
                    if (img.complete && (img.naturalWidth <= 3 || img.naturalHeight <= 3)) {
                        img.classList.add('tracking-pixel');
                        img.style.display = 'none';
                    }
                });
                // Resize iframe after hiding images
                this.resizeIframeToContent(iframe);
            }, 500);

        } catch (e) {
            console.warn('Could not process iframe images:', e);
        }
    },

    formatSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    },

    // Optimistic update helper - applies change immediately, rolls back on failure
    async optimisticUpdate(emailId, updateFn, apiCall, successMsg, rollbackFn) {
        const operationId = `${emailId}-${Date.now()}`;
        const email = this.emails.find(e => e.id === emailId);
        if (!email) return;

        // Store original state for rollback
        const originalState = JSON.parse(JSON.stringify(email));
        this.pendingOperations.set(operationId, { emailId, originalState });

        // Apply optimistic update immediately
        updateFn(email);
        this.updateEmailInUI(emailId);

        try {
            // Make API call
            await apiCall();

            // Success - show toast and clear pending operation
            this.pendingOperations.delete(operationId);
            if (successMsg && typeof showToast === 'function') {
                showToast('success', successMsg.title, successMsg.message);
            }
        } catch (error) {
            console.error(`Optimistic update failed for ${emailId}:`, error);

            // Rollback to original state
            if (rollbackFn) {
                rollbackFn(email, originalState);
            } else {
                Object.assign(email, originalState);
            }
            this.updateEmailInUI(emailId);
            this.pendingOperations.delete(operationId);

            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to update. Changes reverted.');
            }
        }
    },

    // Update a single email item in the UI without full re-render
    updateEmailInUI(emailId) {
        const email = this.emails.find(e => e.id === emailId);
        const item = document.querySelector(`.email-item[data-email-id="${emailId}"]`);
        if (!item || !email) return;

        // Update unread class
        item.classList.toggle('unread', email.unread);

        // Update starred indicator
        let starredEl = item.querySelector('.starred');
        if (email.starred && !starredEl) {
            const actionsEl = item.querySelector('.email-actions-mini');
            if (actionsEl) {
                const span = document.createElement('span');
                span.className = 'starred';
                span.title = 'Starred';
                span.innerHTML = '&#9733;';
                actionsEl.prepend(span);
            }
        } else if (!email.starred && starredEl) {
            starredEl.remove();
        }

        // Add visual feedback for pending operation
        if (this.hasPendingOperation(emailId)) {
            item.classList.add('pending-update');
        } else {
            item.classList.remove('pending-update');
        }
    },

    // Check if email has pending operation
    hasPendingOperation(emailId) {
        for (const [, op] of this.pendingOperations) {
            if (op.emailId === emailId) return true;
        }
        return false;
    },

    async markAsRead(emailId) {
        // Optimistic update - instant UI feedback
        await this.optimisticUpdate(
            emailId,
            (email) => { email.unread = false; },
            () => AirAPI.updateEmail(emailId, { unread: false }),
            null, // Silent - no toast for mark as read
            (email, original) => { email.unread = original.unread; }
        );

        // Also update filtered emails
        this.applyFilter();
    },

    async toggleStar(emailId) {
        const email = this.emails.find(e => e.id === emailId);
        if (!email) return;

        const newStarred = !email.starred;

        // Optimistic update with immediate feedback
        await this.optimisticUpdate(
            emailId,
            (e) => { e.starred = newStarred; },
            () => AirAPI.updateEmail(emailId, { starred: newStarred }),
            { title: newStarred ? 'Starred' : 'Unstarred', message: newStarred ? 'Email starred' : 'Star removed' },
            (e, original) => { e.starred = original.starred; }
        );

        // Re-render detail view if viewing this email
        if (this.selectedEmailId === emailId && this.selectedEmailFull) {
            this.selectedEmailFull.starred = email.starred;
            this.renderEmailDetail(this.selectedEmailFull);
        }
    },

    async archiveEmail(emailId) {
        const email = this.emails.find(e => e.id === emailId);
        if (!email) return;

        // Store original position for undo
        const originalIndex = this.emails.indexOf(email);
        const originalEmail = { ...email };

        // Optimistic removal - instant UI feedback
        this.emails = this.emails.filter(e => e.id !== emailId);
        this.applyFilter();
        this.renderEmails();

        // Clear detail pane if this was selected
        if (this.selectedEmailId === emailId) {
            this.selectedEmailId = null;
            const detailPane = document.querySelector('.email-detail');
            if (detailPane) {
                detailPane.innerHTML = '<div class="empty-state"><div class="empty-message">Select an email to view</div></div>';
            }
        }

        // Show toast with undo option
        if (typeof showToast === 'function') {
            showToast('success', 'Archived', 'Email moved to archive', {
                action: 'Undo',
                onAction: () => {
                    // Restore email
                    this.emails.splice(originalIndex, 0, originalEmail);
                    this.applyFilter();
                    this.renderEmails();
                    showToast('info', 'Restored', 'Email restored');
                }
            });
        }

        // Note: Archive API call would go here if available
    },

    async deleteEmail(emailId) {
        const email = this.emails.find(e => e.id === emailId);
        if (!email) return;

        // Store original for undo
        const originalIndex = this.emails.indexOf(email);
        const originalEmail = { ...email };

        // Optimistic removal - instant UI feedback
        this.emails = this.emails.filter(e => e.id !== emailId);
        this.applyFilter();
        this.renderEmails();

        // Clear detail pane if this was selected
        if (this.selectedEmailId === emailId) {
            this.selectedEmailId = null;
            const detailPane = document.querySelector('.email-detail');
            if (detailPane) {
                detailPane.innerHTML = '<div class="empty-state"><div class="empty-message">Select an email to view</div></div>';
            }
        }

        try {
            await AirAPI.deleteEmail(emailId);

            if (typeof showToast === 'function') {
                showToast('warning', 'Deleted', 'Email moved to trash', {
                    action: 'Undo',
                    onAction: async () => {
                        // Note: Undo delete would require undelete API
                        showToast('info', 'Note', 'Check trash to restore');
                    }
                });
            }
        } catch (error) {
            console.error('Failed to delete email:', error);

            // Rollback - restore email to list
            this.emails.splice(originalIndex, 0, originalEmail);
            this.applyFilter();
            this.renderEmails();

            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to delete email');
            }
        }
    },

    replyToEmail(emailId) {
        // Use full email data (includes thread_id) if available
        const email = this.selectedEmailFull && this.selectedEmailFull.id === emailId
            ? this.selectedEmailFull
            : this.emails.find(e => e.id === emailId);
        if (email && typeof ComposeManager !== 'undefined') {
            ComposeManager.openReply(email);
        }
    },

    forwardEmail(emailId) {
        // Use full email data if available
        const email = this.selectedEmailFull && this.selectedEmailFull.id === emailId
            ? this.selectedEmailFull
            : this.emails.find(e => e.id === emailId);
        if (email && typeof ComposeManager !== 'undefined') {
            ComposeManager.openForward(email);
        }
    },

    // Load smart reply suggestions
    async loadSmartReplies(emailId) {
        const email = this.selectedEmailFull && this.selectedEmailFull.id === emailId
            ? this.selectedEmailFull
            : this.emails.find(e => e.id === emailId);

        if (!email) return;

        const container = document.getElementById(`smartReplies-${emailId}`);
        if (!container) return;

        // Show loading state
        container.innerHTML = `
            <div class="smart-replies-loading">
                <span class="loading-spinner"></span>
                <span>Generating smart replies...</span>
            </div>
        `;

        // Extract plain text from email body
        const parser = new DOMParser();
        const doc = parser.parseFromString(email.body || '', 'text/html');
        const plainText = doc.body?.textContent || '';

        try {
            const response = await fetch('/api/ai/smart-replies', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email_id: emailId,
                    subject: email.subject || '',
                    from: email.from?.map(f => f.name || f.email).join(', ') || '',
                    body: plainText.substring(0, 2000)
                })
            });

            const result = await response.json();

            if (result.success && result.replies && result.replies.length > 0) {
                container.innerHTML = `
                    <div class="smart-replies-header">
                        <span class="smart-replies-icon">💬</span>
                        <span>Smart Replies</span>
                    </div>
                    <div class="smart-replies-list">
                        ${result.replies.map((reply, i) => `
                            <button class="smart-reply-chip" onclick="EmailListManager.useSmartReply('${emailId}', ${i})" data-reply="${this.escapeHtml(reply)}">
                                ${this.escapeHtml(reply)}
                            </button>
                        `).join('')}
                    </div>
                `;
                // Store replies for use
                this.smartReplies = result.replies;
            } else {
                container.innerHTML = `
                    <button class="smart-replies-trigger" onclick="EmailListManager.loadSmartReplies('${emailId}')">
                        <span class="smart-replies-icon">💬</span>
                        <span>Get smart reply suggestions</span>
                    </button>
                `;
                if (result.error) {
                    if (typeof showToast === 'function') {
                        showToast('error', 'AI Error', result.error);
                    }
                }
            }
        } catch (err) {
            console.error('Smart replies error:', err);
            container.innerHTML = `
                <button class="smart-replies-trigger" onclick="EmailListManager.loadSmartReplies('${emailId}')">
                    <span class="smart-replies-icon">💬</span>
                    <span>Get smart reply suggestions</span>
                </button>
            `;
        }
    },

    // Use a smart reply suggestion
    useSmartReply(emailId, replyIndex) {
        if (!this.smartReplies || !this.smartReplies[replyIndex]) return;

        const reply = this.smartReplies[replyIndex];
        const email = this.selectedEmailFull && this.selectedEmailFull.id === emailId
            ? this.selectedEmailFull
            : this.emails.find(e => e.id === emailId);

        if (email && typeof ComposeManager !== 'undefined') {
            ComposeManager.openReply(email, reply);
        }
    },

    // Summarize email with AI (enhanced version)
    async summarizeWithAI(emailId) {
        const email = this.selectedEmailFull && this.selectedEmailFull.id === emailId
            ? this.selectedEmailFull
            : this.emails.find(e => e.id === emailId);

        if (!email) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Email not found');
            }
            return;
        }

        // Get button and show loading state
        const btn = document.getElementById(`summarizeBtn-${emailId}`);
        if (btn) {
            btn.classList.add('loading');
            btn.disabled = true;
            const icon = btn.querySelector('.ai-icon');
            const spinner = btn.querySelector('.ai-spinner');
            const text = btn.querySelector('.ai-btn-text');
            if (icon) icon.style.display = 'none';
            if (spinner) spinner.style.display = 'block';
            if (text) text.textContent = 'Analyzing...';
        }

        // Extract plain text from email body safely using DOMParser
        const parser = new DOMParser();
        const doc = parser.parseFromString(email.body || '', 'text/html');
        const plainText = doc.body?.textContent || '';

        try {
            // Use enhanced summary endpoint
            const response = await fetch('/api/ai/enhanced-summary', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email_id: emailId,
                    subject: email.subject || '',
                    from: email.from?.map(f => f.name || f.email).join(', ') || '',
                    body: plainText.substring(0, 3000)
                })
            });

            const result = await response.json();

            if (result.success) {
                // Show enhanced summary modal
                this.showEnhancedSummaryModal(email.subject, result);
            } else {
                if (typeof showToast === 'function') {
                    showToast('error', 'AI Error', result.error || 'Failed to summarize');
                }
            }
        } catch (err) {
            console.error('AI summarize error:', err);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to connect to Claude Code');
            }
        } finally {
            // Reset button state
            if (btn) {
                btn.classList.remove('loading');
                btn.disabled = false;
                const icon = btn.querySelector('.ai-icon');
                const spinner = btn.querySelector('.ai-spinner');
                const text = btn.querySelector('.ai-btn-text');
                if (icon) icon.style.display = 'block';
                if (spinner) spinner.style.display = 'none';
                if (text) text.textContent = '✨ Summarize';
            }
        }
    },

    // Show enhanced AI summary modal with action items and sentiment
    showEnhancedSummaryModal(subject, result) {
        const sentimentIcons = {
            positive: '😊',
            neutral: '😐',
            negative: '😟',
            urgent: '🚨'
        };

        const categoryIcons = {
            meeting: '📅',
            task: '✅',
            fyi: 'ℹ️',
            question: '❓',
            social: '👋'
        };

        const sentimentIcon = sentimentIcons[result.sentiment] || '😐';
        const categoryIcon = categoryIcons[result.category] || 'ℹ️';

        // Build action items HTML
        const actionItemsHtml = result.action_items && result.action_items.length > 0
            ? `<div class="summary-section">
                <div class="summary-section-title">📋 Action Items</div>
                <ul class="action-items-list">
                    ${result.action_items.map(item => `<li>${this.escapeHtml(item)}</li>`).join('')}
                </ul>
               </div>`
            : '';

        let modal = document.getElementById('aiSummaryModal');
        if (!modal) {
            modal = document.createElement('div');
            modal.id = 'aiSummaryModal';
            modal.className = 'modal-overlay';
            document.body.appendChild(modal);
        }

        modal.innerHTML = `
            <div class="modal ai-summary-modal">
                <div class="modal-header">
                    <h3>✨ AI Analysis</h3>
                    <button class="close-btn" onclick="EmailListManager.closeAISummaryModal()">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="summary-subject">${this.escapeHtml(subject || '(No Subject)')}</div>
                    <div class="summary-badges">
                        <span class="summary-badge sentiment-${result.sentiment}">${sentimentIcon} ${result.sentiment}</span>
                        <span class="summary-badge category-${result.category}">${categoryIcon} ${result.category}</span>
                    </div>
                    <div class="summary-section">
                        <div class="summary-section-title">📝 Summary</div>
                        <div class="summary-content">${this.escapeHtml(result.summary)}</div>
                    </div>
                    ${actionItemsHtml}
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="EmailListManager.copyAISummary()">Copy Summary</button>
                    <button class="btn btn-primary" onclick="EmailListManager.closeAISummaryModal()">Close</button>
                </div>
            </div>
        `;

        modal.style.display = 'flex';
        modal.classList.add('active');

        // Store summary for copying
        this.currentSummary = result.summary;
        if (result.action_items && result.action_items.length > 0) {
            this.currentSummary += '\n\nAction Items:\n' + result.action_items.map(item => '• ' + item).join('\n');
        }
    },

    // Show AI summary in a modal (legacy - for basic summarize)
    showAISummaryModal(subject, summary) {
        // Check if modal already exists
        let modal = document.getElementById('aiSummaryModal');
        if (!modal) {
            modal = document.createElement('div');
            modal.id = 'aiSummaryModal';
            modal.className = 'modal-overlay';
            modal.innerHTML = `
                <div class="modal ai-summary-modal">
                    <div class="modal-header">
                        <h3>✨ AI Summary</h3>
                        <button class="close-btn" onclick="EmailListManager.closeAISummaryModal()">&times;</button>
                    </div>
                    <div class="modal-body">
                        <div class="summary-subject"></div>
                        <div class="summary-content"></div>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-secondary" onclick="EmailListManager.copyAISummary()">Copy Summary</button>
                        <button class="btn btn-primary" onclick="EmailListManager.closeAISummaryModal()">Close</button>
                    </div>
                </div>
            `;
            document.body.appendChild(modal);
        }

        // Update content
        modal.querySelector('.summary-subject').textContent = subject || '(No Subject)';
        modal.querySelector('.summary-content').textContent = summary;
        modal.style.display = 'flex';
        modal.classList.add('active');

        // Store summary for copying
        this.currentSummary = summary;
    },

    // Close AI summary modal
    closeAISummaryModal() {
        const modal = document.getElementById('aiSummaryModal');
        if (modal) {
            modal.classList.remove('active');
            setTimeout(() => modal.style.display = 'none', 200);
        }
    },

    // Copy AI summary to clipboard
    async copyAISummary() {
        if (this.currentSummary) {
            try {
                await navigator.clipboard.writeText(this.currentSummary);
                if (typeof showToast === 'function') {
                    showToast('success', 'Copied', 'Summary copied to clipboard');
                }
            } catch (err) {
                console.error('Copy error:', err);
            }
        }
    }
};

// ====================================
// EMAIL NAVIGATION (keyboard support)
// ====================================

let currentEmailIndex = 0;
let selectedEmails = new Set();

// Virtual scrolling config
const ITEM_HEIGHT = 80;
const BUFFER_SIZE = 5;

// Get email items
function getEmailItems() {
    return document.querySelectorAll('.email-item');
}

// Select next email
function selectNextEmail() {
    const emailItems = getEmailItems();
    if (currentEmailIndex < emailItems.length - 1) {
        emailItems[currentEmailIndex]?.classList.remove('selected');
        currentEmailIndex++;
        emailItems[currentEmailIndex].classList.add('selected');
        emailItems[currentEmailIndex].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        if (typeof announce === 'function') {
            announce(`Email ${currentEmailIndex + 1} of ${emailItems.length}`);
        }
    }
}

// Select previous email
function selectPrevEmail() {
    const emailItems = getEmailItems();
    if (currentEmailIndex > 0) {
        emailItems[currentEmailIndex]?.classList.remove('selected');
        currentEmailIndex--;
        emailItems[currentEmailIndex].classList.add('selected');
        emailItems[currentEmailIndex].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        if (typeof announce === 'function') {
            announce(`Email ${currentEmailIndex + 1} of ${emailItems.length}`);
        }
    }
}

// Toggle email selection (for batch operations)
function toggleEmailSelection(index) {
    const emailItems = getEmailItems();
    if (selectedEmails.has(index)) {
        selectedEmails.delete(index);
        emailItems[index]?.classList.remove('batch-selected');
    } else {
        selectedEmails.add(index);
        emailItems[index]?.classList.add('batch-selected');
    }
    updateBatchActionsUI();
}

// Clear all selections
function clearEmailSelections() {
    const emailItems = getEmailItems();
    selectedEmails.clear();
    emailItems.forEach(item => item.classList.remove('batch-selected'));
    updateBatchActionsUI();
}

// Update batch actions UI
function updateBatchActionsUI() {
    const count = selectedEmails.size;
    const batchBar = document.getElementById('batchActionsBar');
    if (batchBar) {
        batchBar.style.display = count > 0 ? 'flex' : 'none';
        const countEl = batchBar.querySelector('.batch-count');
        if (countEl) countEl.textContent = `${count} selected`;
    }
}

// Archive selected emails
function archiveSelected() {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('success', 'Archived', `${selectedEmails.size} emails moved to archive`);
    }
    clearEmailSelections();
}

// Delete selected emails
function deleteSelected() {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('warning', 'Deleted', `${selectedEmails.size} emails moved to trash`);
    }
    clearEmailSelections();
}

// Mark selected as read/unread
function markSelectedAsRead(read = true) {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('info', read ? 'Marked Read' : 'Marked Unread', `${selectedEmails.size} emails updated`);
    }
    clearEmailSelections();
}

// Send email (legacy function for compatibility)
function sendEmail() {
    if (typeof showSendAnimation === 'function') {
        showSendAnimation();
    }
    setTimeout(() => {
        if (typeof showToast === 'function') {
            showToast('success', 'Email Sent', 'Your message has been delivered');
        }
        if (typeof toggleCompose === 'function') {
            toggleCompose();
        }
    }, 1000);
}

// Generate AI Summary (button-triggered)
function generateAISummary() {
    const btn = document.getElementById('aiSummaryBtn');
    const summaryDiv = document.getElementById('aiSummary');
    const summaryText = document.getElementById('aiSummaryText');

    if (!btn || !summaryDiv || !summaryText) return;

    // Show loading state
    btn.classList.add('loading');
    btn.innerHTML = '<span class="ai-icon">⏳</span><span>Generating summary...</span>';

    // Simulate AI processing
    setTimeout(() => {
        const summaries = [
            "This email discusses the Q4 product roadmap. Key points: 3 new features planned, design mockups attached, stakeholder meeting scheduled for next Tuesday.",
            "Sarah is requesting approval for the marketing budget. The proposal includes increased social media spend and a new influencer campaign targeting Gen Z.",
            "Technical review notes: Performance improvements show 40% faster load times. Minor bugs identified in mobile view need addressing before launch.",
            "Team sync summary: Sprint progress on track, two blockers identified for the auth module, new hire starting Monday needs onboarding setup."
        ];

        const randomSummary = summaries[Math.floor(Math.random() * summaries.length)];

        // Reset button
        btn.classList.remove('loading');
        btn.innerHTML = '<span class="ai-icon">✨</span><span>Summarize with AI</span>';

        // Show summary with typing effect
        summaryDiv.classList.remove('hidden');
        if (typeof showAITyping === 'function') {
            showAITyping(summaryText, randomSummary);
        } else {
            summaryText.textContent = randomSummary;
        }

    }, 1500 + Math.random() * 1000);
}

// Hide AI Summary
function hideAISummary() {
    const summaryDiv = document.getElementById('aiSummary');
    if (summaryDiv) {
        summaryDiv.classList.add('hidden');
    }
}

// Initialize email keyboard navigation
function initEmailKeyboard() {
    const emailList = document.querySelector('.email-list');
    if (!emailList) return;

    emailList.setAttribute('role', 'listbox');
    emailList.setAttribute('aria-label', 'Email messages');
    emailList.setAttribute('tabindex', '0');

    const items = emailList.querySelectorAll('.email-item');
    items.forEach((item, index) => {
        item.setAttribute('role', 'option');
        item.setAttribute('tabindex', '-1');
        item.setAttribute('aria-selected', item.classList.contains('selected') ? 'true' : 'false');
    });

    emailList.addEventListener('keydown', function(e) {
        const items = emailList.querySelectorAll('.email-item');
        const currentIndex = Array.from(items).findIndex(item =>
            item.classList.contains('focused') || item.classList.contains('selected')
        );

        switch(e.key) {
            case 'ArrowDown':
                e.preventDefault();
                if (currentIndex < items.length - 1) {
                    items[currentIndex]?.classList.remove('focused');
                    items[currentIndex + 1].classList.add('focused');
                    items[currentIndex + 1].focus();
                    if (typeof announce === 'function') {
                        announce(`Email ${currentIndex + 2} of ${items.length}`);
                    }
                }
                break;
            case 'ArrowUp':
                e.preventDefault();
                if (currentIndex > 0) {
                    items[currentIndex]?.classList.remove('focused');
                    items[currentIndex - 1].classList.add('focused');
                    items[currentIndex - 1].focus();
                    if (typeof announce === 'function') {
                        announce(`Email ${currentIndex} of ${items.length}`);
                    }
                }
                break;
            case 'Enter':
            case ' ':
                e.preventDefault();
                items[currentIndex]?.click();
                if (typeof announce === 'function') {
                    announce('Email opened');
                }
                break;
            case 'Delete':
            case 'Backspace':
                e.preventDefault();
                if (typeof announce === 'function') {
                    announce('Email deleted');
                }
                if (typeof showToast === 'function') {
                    showToast('warning', 'Deleted', 'Email moved to trash');
                }
                break;
            case 'x':
                e.preventDefault();
                toggleEmailSelection(currentIndex);
                break;
        }
    });
}

// Initialize email module
document.addEventListener('DOMContentLoaded', () => {
    // Init keyboard navigation
    initEmailKeyboard();

    // Init email list manager if we have the email list element
    if (document.querySelector('.email-list')) {
        EmailListManager.init();
    }
});
