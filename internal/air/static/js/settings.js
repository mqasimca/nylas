// ====================================
// SETTINGS MODULE
// ====================================

// Settings state
const settingsState = {
    aiProvider: 'nylas',
    theme: 'dark',
    accentColor: 'purple',
    threading: true,
    avatars: true,
    previewPane: true,
    refreshInterval: 60 // Default 60 seconds
};

// Refresh interval timer
let refreshTimer = null;
let lastRefreshTime = Date.now();

// Color values
const accentColors = {
    purple: '#8b5cf6',
    blue: '#3b82f6',
    green: '#22c55e',
    orange: '#f59e0b',
    pink: '#ec4899',
    red: '#ef4444'
};

// Load settings from localStorage
function loadSettings() {
    const saved = storage.get('nylasClientSettings');
    if (saved) {
        Object.assign(settingsState, saved);
        applySettings();
    }
}

// Save settings to localStorage
function saveSettings() {
    // Collect toggle states
    settingsState.threading = document.getElementById('threadingToggle')?.checked ?? true;
    settingsState.avatars = document.getElementById('avatarsToggle')?.checked ?? true;
    settingsState.previewPane = document.getElementById('previewToggle')?.checked ?? true;

    storage.set('nylasClientSettings', settingsState);
    showToast('success', 'Settings Saved', 'Your preferences have been saved');
    toggleSettings();
}

// Reset settings to defaults
function resetSettings() {
    settingsState.aiProvider = 'nylas';
    settingsState.theme = 'dark';
    settingsState.accentColor = 'purple';
    settingsState.threading = true;
    settingsState.avatars = true;
    settingsState.previewPane = true;

    applySettings();
    updateSettingsUI();
    showToast('info', 'Settings Reset', 'All settings restored to defaults');
}

// Apply settings to UI
function applySettings() {
    setAccentColor(settingsState.accentColor, false);
    setTheme(settingsState.theme, false);
}

// Update settings UI to reflect current state
function updateSettingsUI() {
    // Update AI provider selection
    document.querySelectorAll('.settings-option').forEach(opt => {
        opt.classList.remove('selected');
        const input = opt.querySelector('input');
        if (input && input.value === settingsState.aiProvider) {
            opt.classList.add('selected');
            input.checked = true;
        }
    });

    // Update theme buttons
    document.querySelectorAll('.theme-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.theme === settingsState.theme);
    });

    // Update color options
    document.querySelectorAll('.color-option').forEach(opt => {
        opt.classList.toggle('active', opt.dataset.color === settingsState.accentColor);
    });

    // Update toggles
    const threadingToggle = document.getElementById('threadingToggle');
    const avatarsToggle = document.getElementById('avatarsToggle');
    const previewToggle = document.getElementById('previewToggle');

    if (threadingToggle) threadingToggle.checked = settingsState.threading;
    if (avatarsToggle) avatarsToggle.checked = settingsState.avatars;
    if (previewToggle) previewToggle.checked = settingsState.previewPane;
}

// Toggle settings modal
function toggleSettings() {
    const overlay = document.getElementById('settingsOverlay');
    if (!overlay) return;

    if (overlay.classList.contains('active')) {
        overlay.classList.remove('active');
        setTimeout(() => overlay.style.display = 'none', 200);
    } else {
        overlay.style.display = 'flex';
        updateSettingsUI();
        requestAnimationFrame(() => overlay.classList.add('active'));
    }
}

// Close settings when clicking overlay
function closeSettingsOnOverlay(event) {
    if (event.target.id === 'settingsOverlay') {
        toggleSettings();
    }
}

// Set theme mode
function setTheme(theme, notify = true) {
    settingsState.theme = theme;

    document.querySelectorAll('.theme-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.theme === theme);
    });

    if (theme === 'light') {
        document.body.classList.add('light-theme');
        document.body.classList.remove('dark-theme');
    } else if (theme === 'dark') {
        document.body.classList.remove('light-theme');
        document.body.classList.add('dark-theme');
    } else {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        document.body.classList.toggle('dark-theme', prefersDark);
        document.body.classList.toggle('light-theme', !prefersDark);
    }

    if (notify) {
        showToast('info', 'Theme Updated', `Switched to ${theme} mode`);
    }
}

// Set accent color
function setAccentColor(color, notify = true) {
    settingsState.accentColor = color;

    document.querySelectorAll('.color-option').forEach(opt => {
        opt.classList.toggle('active', opt.dataset.color === color);
    });

    const colorValue = accentColors[color] || accentColors.purple;
    document.documentElement.style.setProperty('--accent', colorValue);
    document.documentElement.style.setProperty('--gradient-accent',
        `linear-gradient(135deg, ${colorValue} 0%, ${adjustColor(colorValue, -20)} 100%)`);

    if (notify) {
        showToast('info', 'Color Updated', `Accent color set to ${color}`);
    }
}

// Helper to adjust color brightness
function adjustColor(hex, percent) {
    const num = parseInt(hex.slice(1), 16);
    const amt = Math.round(2.55 * percent);
    const R = Math.min(255, Math.max(0, (num >> 16) + amt));
    const G = Math.min(255, Math.max(0, ((num >> 8) & 0x00FF) + amt));
    const B = Math.min(255, Math.max(0, (num & 0x0000FF) + amt));
    return `#${(0x1000000 + R * 0x10000 + G * 0x100 + B).toString(16).slice(1)}`;
}

// Initialize AI provider selection listeners
function initSettingsListeners() {
    document.querySelectorAll('.settings-option input[name="ai-provider"]').forEach(input => {
        input.addEventListener('change', function() {
            settingsState.aiProvider = this.value;
            document.querySelectorAll('.settings-option').forEach(opt => opt.classList.remove('selected'));
            this.closest('.settings-option').classList.add('selected');
        });
    });

    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (settingsState.theme === 'system') {
            setTheme('system', false);
        }
    });
}

// ====================================
// REFRESH INTERVAL FUNCTIONALITY
// ====================================

// Set refresh interval
function setRefreshInterval(seconds) {
    settingsState.refreshInterval = seconds;

    // Update UI
    document.querySelectorAll('.interval-btn').forEach(btn => {
        btn.classList.toggle('active', parseInt(btn.dataset.interval) === seconds);
    });

    // Update status text
    updateRefreshStatus();

    // Restart timer with new interval
    startRefreshTimer();
}

// Update refresh status display
function updateRefreshStatus() {
    const statusEl = document.getElementById('refreshStatus');
    if (!statusEl) return;

    const indicator = statusEl.querySelector('.refresh-indicator');
    const text = statusEl.querySelector('.refresh-text');

    if (settingsState.refreshInterval === 0) {
        indicator.classList.add('paused');
        text.textContent = 'Manual refresh only';
    } else {
        indicator.classList.remove('paused');
        const interval = settingsState.refreshInterval;
        if (interval < 60) {
            text.textContent = `Auto-refresh every ${interval} seconds`;
        } else if (interval === 60) {
            text.textContent = 'Auto-refresh every 1 minute';
        } else {
            text.textContent = `Auto-refresh every ${interval / 60} minutes`;
        }
    }
}

// Start the refresh timer
function startRefreshTimer() {
    // Clear existing timer
    if (refreshTimer) {
        clearInterval(refreshTimer);
        refreshTimer = null;
    }

    // Don't start if manual mode
    if (settingsState.refreshInterval === 0) {
        console.log('%c⏸️ Auto-refresh disabled', 'color: #f59e0b;');
        return;
    }

    // Start new timer
    refreshTimer = setInterval(() => {
        refreshEmails();
    }, settingsState.refreshInterval * 1000);

    console.log(`%c🔄 Auto-refresh started: every ${settingsState.refreshInterval}s`, 'color: #22c55e;');
}

// Refresh emails function
function refreshEmails() {
    lastRefreshTime = Date.now();

    // Show refresh indicator in status bar if it exists
    const syncStatus = document.querySelector('.sync-status');
    if (syncStatus) {
        syncStatus.classList.add('syncing');
        setTimeout(() => syncStatus.classList.remove('syncing'), 1000);
    }

    // Simulate fetching new emails
    console.log('%c📬 Checking for new emails...', 'color: #3b82f6;');

    // Random chance of new email for demo
    if (Math.random() > 0.7) {
        setTimeout(() => {
            if (typeof showToast === 'function') {
                showToast('info', 'New Email', '1 new message received');
            }
            // Update unread count
            const badge = document.querySelector('.mobile-nav-badge');
            if (badge) {
                const count = parseInt(badge.textContent) + 1;
                badge.textContent = count;
                badge.classList.add('new-mail');
                setTimeout(() => badge.classList.remove('new-mail'), 300);
            }
        }, 500);
    }
}

// Manual refresh
function manualRefresh() {
    refreshEmails();
    if (typeof showToast === 'function') {
        showToast('info', 'Refreshing', 'Checking for new emails...');
    }
}

// Update settings UI to include refresh interval
const originalUpdateSettingsUI = updateSettingsUI;
updateSettingsUI = function() {
    originalUpdateSettingsUI();

    // Update refresh interval buttons
    document.querySelectorAll('.interval-btn').forEach(btn => {
        btn.classList.toggle('active', parseInt(btn.dataset.interval) === settingsState.refreshInterval);
    });

    updateRefreshStatus();
};

// Initialize settings on load
document.addEventListener('DOMContentLoaded', () => {
    loadSettings();
    initSettingsListeners();
    startRefreshTimer();
});

console.log('%c⚙️ Settings module loaded', 'color: #8b5cf6;');
