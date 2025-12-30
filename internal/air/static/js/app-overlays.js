/**
 * App Overlays - Overlay management and keyboard handlers
 */
        // Keyboard Shortcut Overlay
        function showShortcutOverlay() {
            document.getElementById('shortcutOverlay').classList.add('active');
        }

        function closeShortcutOverlay() {
            document.getElementById('shortcutOverlay').classList.remove('active');
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
