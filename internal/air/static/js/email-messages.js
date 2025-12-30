/* Email Messages - Loading and pagination */
Object.assign(EmailListManager, {
async loadEmails(folderOrOptions = null) {
    if (this.isLoading) return;
    this.isLoading = true;

    // Support both string (folder) and object (options) parameter
    let folder = null;
    let search = null;
    let from = null;
    if (typeof folderOrOptions === 'string') {
        folder = folderOrOptions;
    } else if (folderOrOptions && typeof folderOrOptions === 'object') {
        folder = folderOrOptions.folder || null;
        search = folderOrOptions.search || null;
        from = folderOrOptions.from || null;
    }

    console.log('[loadEmails] Starting...', { folder, search, from, limit: 50 });

    try {
        const options = { limit: 50 }; // Increased from 10 to 50 to fill viewport
        if (folder) {
            this.currentFolder = folder;
            options.folder = folder;
        }
        if (search) {
            this.currentSearch = search;
            options.search = search;
        }
        if (from) {
            options.from = from;
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

        const isInbox = this.currentFolder === this.inboxFolderId ||
                       this.currentFolder === 'INBOX' ||
                       this.currentFolder === 'inbox';
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
});
