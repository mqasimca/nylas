/**
 * App Initialization - Service worker, startup, and productivity managers
 */
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

        // Expose key functions for E2E testing
        window.toggleCompose = toggleCompose;
        window.toggleFocusMode = toggleFocusMode;
        window.showToast = showToast;
