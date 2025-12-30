/**
 * App Shortcuts Extended - Additional keyboard shortcuts
 */
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
