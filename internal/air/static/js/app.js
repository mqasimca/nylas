        // Toast System
        function showToast(type, title, message, action) {
            const container = document.getElementById('toastContainer');
            const toast = document.createElement('div');
            toast.className = `toast ${type}`;

            const icons = { success: '✅', info: '💡', warning: '⏰', error: '❌' };

            toast.innerHTML = `
                <span class="toast-icon">${icons[type]}</span>
                <div class="toast-message"><strong>${title}</strong> — ${message}</div>
                ${action ? `<button class="toast-action" onclick="${action.onclick}">${action.label}</button>` : ''}
            `;

            container.appendChild(toast);

            setTimeout(() => {
                toast.style.opacity = '0';
                toast.style.transform = 'translateY(20px)';
                setTimeout(() => toast.remove(), 300);
            }, 4000);
        }

        // Send Animation
        function showSendAnimation() {
            const anim = document.getElementById('sendAnimation');
            anim.classList.add('active');
            setTimeout(() => anim.classList.remove('active'), 1000);
        }

        // View Switching
        function showView(view, event) {
            // Update nav tabs
            document.querySelectorAll('.nav-tab').forEach(tab => {
                tab.classList.remove('active');
                tab.setAttribute('aria-selected', 'false');
            });

            // Find and activate the clicked tab
            if (event && event.target) {
                const clickedTab = event.target.closest('.nav-tab');
                if (clickedTab) {
                    clickedTab.classList.add('active');
                    clickedTab.setAttribute('aria-selected', 'true');
                }
            } else {
                // Fallback: activate based on view name
                const tabs = document.querySelectorAll('.nav-tab');
                const viewIndex = view === 'email' ? 0 : view === 'calendar' ? 1 : 2;
                if (tabs[viewIndex]) {
                    tabs[viewIndex].classList.add('active');
                    tabs[viewIndex].setAttribute('aria-selected', 'true');
                }
            }

            // Hide all views
            const emailView = document.getElementById('emailView');
            const calendarView = document.getElementById('calendarView');
            const contactsView = document.getElementById('contactsView');

            if (emailView) emailView.classList.remove('active');
            if (calendarView) calendarView.classList.remove('active');
            if (contactsView) contactsView.classList.remove('active');

            // Show selected view
            const targetView = document.getElementById(view + 'View');
            if (targetView) {
                targetView.classList.add('active');
            }

            // Update mobile nav if present
            document.querySelectorAll('.mobile-nav-item').forEach(item => {
                item.classList.remove('active');
            });
            const mobileNavItems = document.querySelectorAll('.mobile-nav-item');
            const mobileIndex = view === 'email' ? 1 : view === 'calendar' ? 2 : 3;
            if (mobileNavItems[mobileIndex]) {
                mobileNavItems[mobileIndex].classList.add('active');
            }

            // Announce view change for screen readers
            if (typeof announce === 'function') {
                announce(`Switched to ${view} view`);
            }

            // Lazy load data for the view (only loads once)
            if (view === 'calendar' && typeof CalendarManager !== 'undefined') {
                CalendarManager.init(); // Will only load data once due to isInitialized flag
            }
            if (view === 'contacts' && typeof ContactsManager !== 'undefined') {
                ContactsManager.loadContacts();
            }
        }

        function toggleCommandPalette() {
            const palette = document.getElementById('commandPalette');
            palette.classList.toggle('hidden');
            if (!palette.classList.contains('hidden')) {
                palette.querySelector('input').focus();
            }
        }

        function toggleCompose() {
            // Use ComposeManager if available
            if (typeof ComposeManager !== 'undefined') {
                if (ComposeManager.isOpen) {
                    ComposeManager.close();
                } else {
                    ComposeManager.open();
                }
            } else {
                // Fallback to simple toggle
                document.getElementById('composeModal').classList.toggle('hidden');
            }
        }

        // Keyboard Shortcuts
        document.addEventListener('keydown', function(e) {
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                toggleCommandPalette();
            }
            if (e.key === 'Escape') {
                document.getElementById('commandPalette').classList.add('hidden');
                document.getElementById('composeModal').classList.add('hidden');
            }
            if (!e.target.matches('input, textarea, [contenteditable]')) {
                if (e.key === 'c') { e.preventDefault(); toggleCompose(); }
                if (e.key === '1') { document.querySelector('.nav-tab').click(); }
                if (e.key === '2') { document.querySelectorAll('.nav-tab')[1].click(); }
                if (e.key === '3') { document.querySelectorAll('.nav-tab')[2].click(); }
                if (e.key === 'e') { showToast('success', 'Archived', 'Moved to archive'); }
                if (e.key === 'r') { toggleCompose(); }
                if (e.key === 's') { showToast('info', 'Starred', 'Conversation starred'); }
                if (e.key === '#') { showToast('warning', 'Deleted', 'Moved to trash'); }
                if (e.key === 'j') { selectNextEmail(); }
                if (e.key === 'k') { selectPrevEmail(); }
            }
            // Send email: Cmd+Enter (handled by ComposeManager in api.js)
        });

        // NOTE: Email navigation (selectNextEmail, selectPrevEmail, sendEmail)
        // is defined in js/email.js

        // Demo: Show toast on page load
        setTimeout(() => {
            showToast('info', 'Welcome back!', '3 new messages since you left');
        }, 1500);

        // NOTE: Email item click handlers are managed in js/email.js

        // ================================
        // NEW CUTTING-EDGE FEATURES 2025
        // ================================

        // Focus Mode / Zen Mode
        let focusModeActive = false;

        function toggleFocusMode() {
            focusModeActive = !focusModeActive;
            document.querySelector('.app').classList.toggle('focus-mode-active', focusModeActive);
            document.getElementById('focusModeToggle').classList.toggle('active', focusModeActive);

            if (focusModeActive) {
                showToast('info', 'Focus Mode', 'Distractions hidden. Press F to exit.');
            } else {
                showToast('info', 'Focus Mode Off', 'Full interface restored');
            }
        }

        // Advanced Search
        function openSearch() {
            const overlay = document.getElementById('searchOverlay');
            overlay.classList.add('active');
            setTimeout(() => {
                document.getElementById('searchInput').focus();
            }, 100);
        }

        function closeSearch() {
            document.getElementById('searchOverlay').classList.remove('active');
            document.getElementById('searchInput').value = '';
        }

        function handleSearch(query) {
            const suggestions = document.getElementById('searchSuggestions');
            if (query.length > 0) {
                // Demo: show search results
                suggestions.innerHTML = `
                    <div class="search-suggestion-group">
                        <div class="search-suggestion-title">Results for "${query}"</div>
                        <div class="search-suggestion-item" onclick="executeSearch('${query}')">
                            <div class="search-suggestion-icon">📧</div>
                            <div class="search-suggestion-content">
                                <div class="search-suggestion-text">Email containing <mark>${query}</mark></div>
                                <div class="search-suggestion-meta">From Sarah Chen • 2 hours ago</div>
                            </div>
                        </div>
                        <div class="search-suggestion-item" onclick="executeSearch('${query}')">
                            <div class="search-suggestion-icon">📧</div>
                            <div class="search-suggestion-content">
                                <div class="search-suggestion-text">Re: <mark>${query}</mark> discussion</div>
                                <div class="search-suggestion-meta">From Alex Johnson • Yesterday</div>
                            </div>
                        </div>
                    </div>
                `;
            }
        }

        function executeSearch(query) {
            closeSearch();
            showToast('info', 'Searching', `Finding emails matching "${query}"...`);
        }

        function toggleSearchFilter(btn) {
            btn.classList.toggle('active');
        }

        // Keyboard Shortcut Overlay
        function showShortcutOverlay() {
            document.getElementById('shortcutOverlay').classList.add('active');
        }

        function closeShortcutOverlay() {
            document.getElementById('shortcutOverlay').classList.remove('active');
        }

        // Context Menu
        let contextMenuTarget = null;

        document.addEventListener('contextmenu', function(e) {
            const emailItem = e.target.closest('.email-item');
            if (emailItem) {
                e.preventDefault();
                contextMenuTarget = emailItem;
                // Store email ID for context menu actions (used by snooze)
                window.contextMenuEmailId = emailItem.getAttribute('data-email-id');
                const contextMenu = document.getElementById('contextMenu');
                contextMenu.style.left = e.clientX + 'px';
                contextMenu.style.top = e.clientY + 'px';
                contextMenu.classList.add('active');
            }
        });

        document.addEventListener('click', function(e) {
            const contextMenu = document.getElementById('contextMenu');
            if (!e.target.closest('.context-menu')) {
                contextMenu.classList.remove('active');
            }
        });

        function handleContextAction(action) {
            document.getElementById('contextMenu').classList.remove('active');

            // Get the email ID from the context menu target
            const emailId = contextMenuTarget?.getAttribute('data-email-id');

            const actions = {
                reply: () => {
                    if (emailId && typeof EmailListManager !== 'undefined') {
                        EmailListManager.replyToEmail(emailId);
                    } else {
                        toggleCompose();
                    }
                },
                replyAll: () => { toggleCompose(); showToast('info', 'Reply All', 'Replying to all...'); },
                forward: () => {
                    if (emailId && typeof EmailListManager !== 'undefined') {
                        EmailListManager.forwardEmail(emailId);
                    } else {
                        toggleCompose();
                    }
                },
                archive: () => {
                    if (emailId && typeof EmailListManager !== 'undefined') {
                        EmailListManager.archiveEmail(emailId);
                    } else {
                        showToast('success', 'Archived', 'Conversation archived');
                    }
                },
                snooze: () => {
                    if (emailId && typeof SnoozeManager !== 'undefined') {
                        SnoozeManager.openForEmail(emailId);
                    }
                },
                star: () => {
                    if (emailId && typeof EmailListManager !== 'undefined') {
                        EmailListManager.toggleStar(emailId);
                    } else {
                        showToast('info', 'Starred', 'Conversation starred');
                    }
                },
                markUnread: () => showToast('info', 'Marked Unread', 'Conversation marked as unread'),
                label: () => showToast('info', 'Label', 'Label picker would open...'),
                move: () => showToast('info', 'Move', 'Folder picker would open...'),
                delete: () => {
                    if (emailId && typeof EmailListManager !== 'undefined') {
                        EmailListManager.deleteEmail(emailId);
                    } else {
                        showToast('warning', 'Deleted', 'Moved to trash');
                    }
                }
            };

            if (actions[action]) actions[action]();
        }

        // Snooze Picker (legacy - now uses SnoozeManager from productivity.js)
        function showSnoozePicker() {
            if (typeof SnoozeManager !== 'undefined') {
                SnoozeManager.open();
            }
        }

        function handleSnooze(time) {
            if (typeof SnoozeManager !== 'undefined') {
                SnoozeManager.snooze(time);
            }
        }

        // Extended Keyboard Shortcuts
        document.addEventListener('keydown', function(e) {
            if (e.target.matches('input, textarea, [contenteditable]')) return;

            // Focus mode: Shift+F
            if (e.shiftKey && e.key === 'F') {
                e.preventDefault();
                toggleFocusMode();
            }

            // Show shortcuts: ?
            if (e.key === '?') {
                e.preventDefault();
                showShortcutOverlay();
            }

            // Snooze: B
            if (e.key === 'b') {
                e.preventDefault();
                showSnoozePicker();
            }
        });

        // Update the command palette to use the new search
        document.querySelector('.search-trigger').addEventListener('click', function(e) {
            e.stopPropagation();
            openSearch();
        });

        // Parallax effect on cards
        document.querySelectorAll('.email-item').forEach(card => {
            card.addEventListener('mousemove', function(e) {
                const rect = card.getBoundingClientRect();
                const x = e.clientX - rect.left;
                const y = e.clientY - rect.top;
                const centerX = rect.width / 2;
                const centerY = rect.height / 2;
                const rotateX = (y - centerY) / 20;
                const rotateY = (centerX - x) / 20;
                card.style.setProperty('--rotateX', rotateX + 'deg');
                card.style.setProperty('--rotateY', rotateY + 'deg');
            });

            card.addEventListener('mouseleave', function() {
                card.style.setProperty('--rotateX', '0deg');
                card.style.setProperty('--rotateY', '0deg');
            });
        });

        // Demo: AI Typing Animation
        function showAITyping(element, text) {
            element.classList.add('ai-streaming-text');
            let i = 0;
            const interval = setInterval(() => {
                element.textContent = text.substring(0, i);
                i++;
                if (i > text.length) {
                    clearInterval(interval);
                    element.classList.remove('ai-streaming-text');
                }
            }, 30);
        }

        // Magnetic button effect
        document.querySelectorAll('.magnetic-btn, .compose-btn, .action-btn').forEach(btn => {
            btn.addEventListener('mousemove', function(e) {
                const rect = btn.getBoundingClientRect();
                const x = e.clientX - rect.left - rect.width / 2;
                const y = e.clientY - rect.top - rect.height / 2;
                btn.style.transform = `translate(${x * 0.1}px, ${y * 0.1}px)`;
            });

            btn.addEventListener('mouseleave', function() {
                btn.style.transform = '';
            });
        });

        // Spring animation on new elements
        function springAnimate(element) {
            element.classList.add('spring-in');
            element.addEventListener('animationend', () => {
                element.classList.remove('spring-in');
            }, { once: true });
        }

        // Close overlays on Escape
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape') {
                closeSearch();
                closeShortcutOverlay();
                document.getElementById('contextMenu').classList.remove('active');

                // Close productivity modals
                if (typeof SnoozeManager !== 'undefined') SnoozeManager.close();
                if (typeof TemplatesManager !== 'undefined') {
                    TemplatesManager.close();
                    TemplatesManager.hideCreate();
                    TemplatesManager.cancelVariables();
                }
                if (typeof ScheduledSendManager !== 'undefined') ScheduledSendManager.closeDropdown();

                // Close settings if open
                const settingsOverlay = document.getElementById('settingsOverlay');
                if (settingsOverlay && settingsOverlay.classList.contains('active')) {
                    toggleSettings();
                }
            }
        });

        // ====================================
        // NOTE: Settings functionality is in js/settings.js
        // ====================================

        // ====================================
        // NOTE: AI Summary functionality is in js/email.js
        // ====================================

        // ====================================
        // ACCESSIBILITY HELPERS
        // NOTE: announce() is in js/utils.js
        // NOTE: Email keyboard navigation is in js/email.js (initEmailKeyboard)
        // ====================================

        // Trap focus within modal (unique to app.js)
        function trapFocus(element) {
            const focusableElements = element.querySelectorAll(
                'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
            );
            const firstFocusable = focusableElements[0];
            const lastFocusable = focusableElements[focusableElements.length - 1];

            element.addEventListener('keydown', function(e) {
                if (e.key === 'Tab') {
                    if (e.shiftKey) {
                        if (document.activeElement === firstFocusable) {
                            lastFocusable.focus();
                            e.preventDefault();
                        }
                    } else {
                        if (document.activeElement === lastFocusable) {
                            firstFocusable.focus();
                            e.preventDefault();
                        }
                    }
                }
            });

            // Focus first element
            firstFocusable?.focus();
        }

        // ====================================
        // SERVICE WORKER REGISTRATION
        // ====================================

        if ('serviceWorker' in navigator) {
            window.addEventListener('load', () => {
                navigator.serviceWorker.register('sw.js')
                    .then(registration => {
                        console.log('%c🔧 Service Worker registered', 'color: #22c55e;');

                        // Check for updates
                        registration.addEventListener('updatefound', () => {
                            const newWorker = registration.installing;
                            newWorker.addEventListener('statechange', () => {
                                if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                                    showToast('info', 'Update Available', 'Refresh to get the latest version');
                                }
                            });
                        });
                    })
                    .catch(error => {
                        console.log('%c⚠️ Service Worker registration failed', 'color: #f59e0b;', error);
                    });
            });
        }

        // ====================================
        // ACCOUNT SWITCHER (Phase 2)
        // ====================================

        function toggleAccountDropdown() {
            const dropdown = document.getElementById('accountDropdown');
            if (dropdown) {
                dropdown.classList.toggle('hidden');

                // Update aria-expanded
                const switcher = document.querySelector('.account-switcher');
                if (switcher) {
                    switcher.setAttribute('aria-expanded', !dropdown.classList.contains('hidden'));
                }
            }
        }

        function switchAccount(grantId) {
            // Close dropdown
            const dropdown = document.getElementById('accountDropdown');
            if (dropdown) dropdown.classList.add('hidden');

            // Call API to switch account
            fetch('/api/grants/default', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ grant_id: grantId })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showToast('success', 'Account Switched', 'Reloading...');
                    // Reload to show new account data
                    setTimeout(() => window.location.reload(), 500);
                } else {
                    showToast('error', 'Error', data.error || 'Failed to switch account');
                }
            })
            .catch(err => {
                showToast('error', 'Error', 'Failed to switch account');
                console.error('Switch account error:', err);
            });
        }

        function addAccount() {
            // Close dropdown
            const dropdown = document.getElementById('accountDropdown');
            if (dropdown) dropdown.classList.add('hidden');

            showToast('info', 'Add Account', 'Run: nylas auth login --provider google');
        }

        function showSetupInstructions() {
            showToast('info', 'Setup Required', 'Run: nylas auth login in your terminal');
        }

        function closeSetupBanner() {
            const banner = document.getElementById('setupBanner');
            if (banner) {
                banner.style.display = 'none';
            }
        }

        // Close account dropdown when clicking outside
        document.addEventListener('click', function(e) {
            const dropdown = document.getElementById('accountDropdown');
            const container = document.querySelector('.account-switcher-container');
            if (dropdown && container && !container.contains(e.target)) {
                dropdown.classList.add('hidden');
            }
        });

        // ====================================
        // APP INITIALIZATION
        // ====================================

        console.log('%c✨ Nylas Air - Modern Email Client',
            'color: #8b5cf6; font-size: 16px; font-weight: bold;');
        console.log('%c📦 Modules: Utils, Settings, Email, Mobile, Productivity',
            'color: #a1a1aa; font-size: 12px;');
        console.log('%c🚀 Features: PWA, Offline Support, Accessibility, Responsive, Split Inbox',
            'color: #a1a1aa; font-size: 12px;');

        // ====================================
        // PRODUCTIVITY MANAGERS INITIALIZATION
        // ====================================

        // Initialize Split Inbox tab handlers
        document.addEventListener('DOMContentLoaded', function() {
            const filterTabs = document.querySelectorAll('#emailFilterTabs .filter-tab');
            filterTabs.forEach(tab => {
                tab.addEventListener('click', function() {
                    // Update active state
                    filterTabs.forEach(t => t.classList.remove('active'));
                    this.classList.add('active');

                    // Get filter type
                    const filter = this.getAttribute('data-filter');

                    // Call SplitInboxManager if available
                    if (typeof SplitInboxManager !== 'undefined') {
                        SplitInboxManager.filterByCategory(filter);
                    } else if (typeof EmailListManager !== 'undefined') {
                        // Fallback to email list filtering
                        if (filter === 'unread') {
                            EmailListManager.loadEmails({ unread: true });
                        } else if (filter === 'vip') {
                            EmailListManager.filterVIP();
                        } else {
                            EmailListManager.loadEmails({});
                        }
                    }
                });
            });

            console.log('%c📊 Split Inbox tabs initialized', 'color: #22c55e;');
        });

        // Keyboard shortcut for snooze: Z key
        document.addEventListener('keydown', function(e) {
            if (e.target.matches('input, textarea, [contenteditable]')) return;

            // Z for snooze on selected email
            if (e.key === 'z' && typeof SnoozeManager !== 'undefined') {
                e.preventDefault();
                const selectedEmail = document.querySelector('.email-item.selected');
                if (selectedEmail) {
                    const emailId = selectedEmail.getAttribute('data-email-id');
                    if (emailId) {
                        SnoozeManager.openForEmail(emailId);
                    }
                }
            }

            // T for templates in compose mode
            if (e.key === 't' && typeof TemplatesManager !== 'undefined') {
                const composeModal = document.getElementById('composeModal');
                if (composeModal && !composeModal.classList.contains('hidden')) {
                    e.preventDefault();
                    TemplatesManager.open();
                }
            }
        });
