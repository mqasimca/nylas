// =============================================================================
// Timestamps
// =============================================================================

// Track last run times for timestamps
let lastRunTimes = {
    auth: null,
    email: null,
    calendar: null
};

function updateTimestamp(section) {
    lastRunTimes[section] = Date.now();
    updateTimestampDisplay(section);
}

function updateTimestampDisplay(section) {
    const el = document.getElementById(`${section}-timestamp`);
    if (!el || !lastRunTimes[section]) {
        if (el) el.textContent = '';
        return;
    }

    const elapsed = Date.now() - lastRunTimes[section];
    el.textContent = 'Last run: ' + formatRelativeTime(elapsed);
}

function updateAllTimestamps() {
    ['auth', 'email', 'calendar'].forEach(updateTimestampDisplay);
}

function formatRelativeTime(ms) {
    const seconds = Math.floor(ms / 1000);
    if (seconds < 60) return 'just now';
    const minutes = Math.floor(seconds / 60);
    if (minutes === 1) return '1 minute ago';
    if (minutes < 60) return `${minutes} minutes ago`;
    const hours = Math.floor(minutes / 60);
    if (hours === 1) return '1 hour ago';
    return `${hours} hours ago`;
}
