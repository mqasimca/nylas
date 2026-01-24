/* Email UI - Rendering and filtering */
Object.assign(EmailListManager, {
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

// Select a folder and load its emails
selectFolder(folderId, folderName) {
    // Update current folder and load emails
    this.currentFolder = folderId;
    this.loadEmails(folderId);

    // Update folder UI to show active state
    const folderItems = document.querySelectorAll('.folder-item');
    folderItems.forEach(item => {
        const itemFolderId = item.getAttribute('data-folder-id');
        if (itemFolderId === folderId) {
            item.classList.add('active');
            item.setAttribute('aria-current', 'true');
        } else {
            item.classList.remove('active');
            item.removeAttribute('aria-current');
        }
    });

    // Update header title with folder name
    const headerTitle = document.querySelector('.email-list-header h2');
    if (headerTitle && folderName) {
        headerTitle.textContent = folderName;
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
});
