/**
 * App Context - Context menu and legacy snooze handlers
 */
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
