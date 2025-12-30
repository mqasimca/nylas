/**
 * App Focus - Focus mode (Zen mode)
 */
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
