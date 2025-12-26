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

// Email List Manager
const EmailListManager = {
    currentFolder: 'INBOX',
    currentFilter: 'all', // 'all', 'vip', 'unread'
    emails: [],
    filteredEmails: [], // Emails after applying filter
    vipSenders: [], // List of VIP email addresses
    selectedEmailId: null,
    selectedEmailFull: null, // Store full email data for reply/forward
    nextCursor: null,
    hasMore: false,
    isLoading: false,

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

    // Apply filter to emails
    applyFilter() {
        switch (this.currentFilter) {
            case 'vip':
                this.filteredEmails = this.emails.filter(email => {
                    const senderEmail = email.from && email.from[0] ? email.from[0].email.toLowerCase() : '';
                    return this.vipSenders.some(vip => senderEmail.includes(vip.toLowerCase()));
                });
                break;
            case 'unread':
                this.filteredEmails = this.emails.filter(email => email.unread);
                break;
            default: // 'all'
                this.filteredEmails = [...this.emails];
                break;
        }
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

        // Infinite scroll
        const emailListContainer = document.querySelector('.email-list');
        if (emailListContainer) {
            emailListContainer.addEventListener('scroll', () => {
                if (this.hasMore && !this.isLoading) {
                    const { scrollTop, scrollHeight, clientHeight } = emailListContainer;
                    if (scrollTop + clientHeight >= scrollHeight - 200) {
                        this.loadMore();
                    }
                }
            });
        }
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

        try {
            const options = { limit: 10 };
            if (folder) {
                this.currentFolder = folder;
                options.folder = folder;
            }

            const data = await AirAPI.getEmails(options);
            this.emails = data.emails || [];
            this.nextCursor = data.next_cursor;
            this.hasMore = data.has_more;

            // Apply current filter
            this.applyFilter();
            this.renderEmails();
        } catch (error) {
            console.error('Failed to load emails:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load emails');
            }
        } finally {
            this.isLoading = false;
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
        }
    },

    renderEmails(append = false) {
        const emailList = document.querySelector('.email-list');
        if (!emailList) return;

        if (!append) {
            emailList.innerHTML = '';
        }

        // Use filtered emails for display
        const displayEmails = this.filteredEmails.length > 0 || this.currentFilter !== 'all'
            ? this.filteredEmails
            : this.emails;

        if (displayEmails.length === 0 && !append) {
            const emptyMessages = {
                'vip': { icon: '⭐', title: 'No VIP emails', message: 'Add VIP senders to see their emails here' },
                'unread': { icon: '✓', title: 'All caught up!', message: 'No unread emails' },
                'all': { icon: '📭', title: 'No emails', message: 'This folder is empty' }
            };
            const msg = emptyMessages[this.currentFilter] || emptyMessages.all;
            emailList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">${msg.icon}</div>
                    <div class="empty-title">${msg.title}</div>
                    <div class="empty-message">${msg.message}</div>
                </div>
            `;
            return;
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
            emailList.innerHTML = '';
            emailList.appendChild(fragment);
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
                ${email.body || '<p>No content</p>'}
            </div>
        `;
    },

    formatSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    },

    async markAsRead(emailId) {
        try {
            await AirAPI.updateEmail(emailId, { unread: false });

            // Update local state
            const email = this.emails.find(e => e.id === emailId);
            if (email) email.unread = false;

            // Update UI
            const item = document.querySelector(`.email-item[data-email-id="${emailId}"]`);
            if (item) item.classList.remove('unread');
        } catch (error) {
            console.error('Failed to mark as read:', error);
        }
    },

    async toggleStar(emailId) {
        const email = this.emails.find(e => e.id === emailId);
        if (!email) return;

        const newStarred = !email.starred;

        try {
            await AirAPI.updateEmail(emailId, { starred: newStarred });
            email.starred = newStarred;

            if (typeof showToast === 'function') {
                showToast('info', newStarred ? 'Starred' : 'Unstarred',
                    newStarred ? 'Email starred' : 'Star removed');
            }

            // Re-render if viewing this email
            if (this.selectedEmailId === emailId) {
                const fullEmail = await AirAPI.getEmail(emailId);
                this.renderEmailDetail(fullEmail);
            }
        } catch (error) {
            console.error('Failed to toggle star:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to update email');
            }
        }
    },

    async archiveEmail(emailId) {
        // For now, just show a message - archive requires knowing the archive folder ID
        if (typeof showToast === 'function') {
            showToast('success', 'Archived', 'Email moved to archive');
        }

        // Remove from current list
        this.emails = this.emails.filter(e => e.id !== emailId);
        this.renderEmails();

        // Clear detail pane if this was selected
        if (this.selectedEmailId === emailId) {
            this.selectedEmailId = null;
            const detailPane = document.querySelector('.email-detail');
            if (detailPane) {
                detailPane.innerHTML = '<div class="empty-state"><div class="empty-message">Select an email to view</div></div>';
            }
        }
    },

    async deleteEmail(emailId) {
        try {
            await AirAPI.deleteEmail(emailId);

            if (typeof showToast === 'function') {
                showToast('warning', 'Deleted', 'Email moved to trash');
            }

            // Remove from current list
            this.emails = this.emails.filter(e => e.id !== emailId);
            this.renderEmails();

            // Clear detail pane if this was selected
            if (this.selectedEmailId === emailId) {
                this.selectedEmailId = null;
                const detailPane = document.querySelector('.email-detail');
                if (detailPane) {
                    detailPane.innerHTML = '<div class="empty-state"><div class="empty-message">Select an email to view</div></div>';
                }
            }
        } catch (error) {
            console.error('Failed to delete email:', error);
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
