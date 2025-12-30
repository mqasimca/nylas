/**
 * App Account - Account switcher and management
 */
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
