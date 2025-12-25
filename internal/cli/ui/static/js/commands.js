// =============================================================================
// Command Definitions and Execution
// =============================================================================

// Auth Commands Data - Grouped by section
const authCommandSections = [
    {
        title: 'Authentication',
        commands: {
            'login': {
                title: 'Login',
                cmd: 'auth login',
                desc: 'Authenticate with an email provider',
                flags: [
                    { name: 'provider', type: 'text', label: 'Provider', placeholder: 'google or microsoft', short: 'p' }
                ]
            },
            'logout': { title: 'Logout', cmd: 'auth logout', desc: 'Revoke current authentication' },
            'status': { title: 'Status', cmd: 'auth status', desc: 'Show authentication status' },
            'whoami': { title: 'Who Am I', cmd: 'auth whoami', desc: 'Show current user info' }
        }
    },
    {
        title: 'Accounts',
        commands: {
            'list': { title: 'List', cmd: 'auth list', desc: 'List all authenticated accounts' },
            'show': { title: 'Show', cmd: 'auth show', desc: 'Show detailed grant information', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'switch': { title: 'Switch', cmd: 'auth switch', desc: 'Switch between accounts', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'add': { title: 'Add', cmd: 'auth add', desc: 'Manually add an existing grant', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } },
            'remove': { title: 'Remove', cmd: 'auth remove', desc: 'Remove grant from local config', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'revoke': { title: 'Revoke', cmd: 'auth revoke', desc: 'Permanently revoke a grant', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } }
        }
    },
    {
        title: 'Configuration',
        commands: {
            'config': {
                title: 'Config',
                cmd: 'auth config',
                desc: 'Configure API credentials',
                flags: [
                    { name: 'api-key', type: 'text', label: 'API Key', placeholder: 'Your Nylas API key' },
                    { name: 'region', type: 'text', label: 'Region', placeholder: 'us or eu (default: us)', short: 'r' },
                    { name: 'client-id', type: 'text', label: 'Client ID', placeholder: 'Auto-detected if not provided' },
                    { name: 'reset', type: 'checkbox', label: 'Reset configuration' }
                ]
            },
            'providers': { title: 'Providers', cmd: 'auth providers', desc: 'List available providers' },
            'detect': { title: 'Detect', cmd: 'auth detect', desc: 'Detect provider from email', param: { name: 'email', placeholder: 'Enter email address...' } },
            'scopes': { title: 'Scopes', cmd: 'auth scopes', desc: 'Show OAuth scopes for a grant', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'token': { title: 'Token', cmd: 'auth token', desc: 'Show or copy API key' },
            'migrate': { title: 'Migrate', cmd: 'auth migrate', desc: 'Migrate to system keyring' }
        }
    }
];

// Flat lookup for auth commands
const authCommands = {};
authCommandSections.forEach(section => {
    Object.assign(authCommands, section.commands);
});

// Email Commands Data - Grouped by section
const emailCommandSections = [
    {
        title: 'Messages',
        commands: {
            'list': {
                title: 'List',
                cmd: 'email list',
                desc: 'List recent emails',
                flags: [
                    { name: 'id', type: 'checkbox', label: 'Show IDs', default: true },
                    { name: 'unread', type: 'checkbox', label: 'Unread only', short: 'u' },
                    { name: 'starred', type: 'checkbox', label: 'Starred only', short: 's' },
                    { name: 'all-folders', type: 'checkbox', label: 'All folders' },
                    { name: 'limit', type: 'number', label: 'Limit', placeholder: '10', short: 'l' },
                    { name: 'from', type: 'text', label: 'From', placeholder: 'sender@email.com', short: 'f' },
                    { name: 'folder', type: 'text', label: 'Folder', placeholder: 'INBOX, SENT, TRASH...' }
                ]
            },
            'read': { title: 'Read', cmd: 'email read', desc: 'Read a specific email', param: { name: 'message-id', placeholder: 'Enter message ID...' } },
            'send': { title: 'Send', cmd: 'email send', desc: 'Send an email' },
            'search': { title: 'Search', cmd: 'email search', desc: 'Search emails', param: { name: 'query', placeholder: 'Enter search query...' } },
            'delete': { title: 'Delete', cmd: 'email delete', desc: 'Delete an email', param: { name: 'message-id', placeholder: 'Enter message ID...' } },
            'mark': { title: 'Mark', cmd: 'email mark', desc: 'Mark as read/unread/starred', param: { name: 'message-id', placeholder: 'Enter message ID...' } }
        }
    },
    {
        title: 'Folders',
        commands: {
            'folders-list': {
                title: 'List',
                cmd: 'email folders list',
                desc: 'List all folders',
                flags: [
                    { name: 'id', type: 'checkbox', label: 'Show IDs', default: true }
                ]
            },
            'folders-show': { title: 'Show', cmd: 'email folders show', desc: 'Show folder details', param: { name: 'folder-id', placeholder: 'Enter folder ID...' } },
            'folders-create': { title: 'Create', cmd: 'email folders create', desc: 'Create a new folder' },
            'folders-rename': { title: 'Rename', cmd: 'email folders rename', desc: 'Rename a folder', param: { name: 'folder-id', placeholder: 'Enter folder ID...' } },
            'folders-delete': { title: 'Delete', cmd: 'email folders delete', desc: 'Delete a folder', param: { name: 'folder-id', placeholder: 'Enter folder ID...' } }
        }
    },
    {
        title: 'Drafts',
        commands: {
            'drafts-list': { title: 'List', cmd: 'email drafts list', desc: 'List drafts' },
            'drafts-show': { title: 'Show', cmd: 'email drafts show', desc: 'Show draft details', param: { name: 'draft-id', placeholder: 'Enter draft ID...' } },
            'drafts-create': {
                title: 'Create',
                cmd: 'email drafts create',
                desc: 'Create a new draft',
                flags: [
                    { name: 'to', type: 'text', label: 'To', placeholder: 'recipient@example.com', short: 't' },
                    { name: 'cc', type: 'text', label: 'CC', placeholder: 'cc@example.com (optional)' },
                    { name: 'subject', type: 'text', label: 'Subject', placeholder: 'Email subject', short: 's' },
                    { name: 'body', type: 'text', label: 'Body', placeholder: 'Email body...', short: 'b' }
                ]
            },
            'drafts-delete': { title: 'Delete', cmd: 'email drafts delete', desc: 'Delete a draft', param: { name: 'draft-id', placeholder: 'Enter draft ID...' } },
            'drafts-send': { title: 'Send', cmd: 'email drafts send', desc: 'Send a draft', param: { name: 'draft-id', placeholder: 'Enter draft ID...' } }
        }
    },
    {
        title: 'Threads',
        commands: {
            'threads-list': {
                title: 'List',
                cmd: 'email threads list',
                desc: 'List email threads',
                flags: [
                    { name: 'id', type: 'checkbox', label: 'Show IDs', default: true }
                ]
            },
            'threads-show': { title: 'Show', cmd: 'email threads show', desc: 'Show thread details', param: { name: 'thread-id', placeholder: 'Enter thread ID...' } },
            'threads-search': { title: 'Search', cmd: 'email threads search', desc: 'Search threads', param: { name: 'query', placeholder: 'Enter search query...' } },
            'threads-delete': { title: 'Delete', cmd: 'email threads delete', desc: 'Delete a thread', param: { name: 'thread-id', placeholder: 'Enter thread ID...' } },
            'threads-mark': { title: 'Mark', cmd: 'email threads mark', desc: 'Mark thread read/unread', param: { name: 'thread-id', placeholder: 'Enter thread ID...' } }
        }
    },
    {
        title: 'Scheduled',
        commands: {
            'scheduled-list': { title: 'List', cmd: 'email scheduled list', desc: 'List scheduled messages' },
            'scheduled-show': { title: 'Show', cmd: 'email scheduled show', desc: 'Show scheduled message', param: { name: 'schedule-id', placeholder: 'Enter schedule ID...' } },
            'scheduled-cancel': { title: 'Cancel', cmd: 'email scheduled cancel', desc: 'Cancel scheduled message', param: { name: 'schedule-id', placeholder: 'Enter schedule ID...' } }
        }
    },
    {
        title: 'Attachments',
        commands: {
            'attachments-list': { title: 'List', cmd: 'email attachments list', desc: 'List attachments', param: { name: 'message-id', placeholder: 'Enter message ID...' } },
            'attachments-show': { title: 'Show', cmd: 'email attachments show', desc: 'Show attachment metadata', param: { name: 'attachment-id', placeholder: 'Enter attachment ID...' } },
            'attachments-download': { title: 'Download', cmd: 'email attachments download', desc: 'Download attachment', param: { name: 'attachment-id', placeholder: 'Enter attachment ID...' } }
        }
    },
    {
        title: 'Other',
        commands: {
            'metadata': { title: 'Metadata', cmd: 'email metadata', desc: 'Manage message metadata', param: { name: 'message-id', placeholder: 'Enter message ID...' } },
            'tracking-info': { title: 'Tracking', cmd: 'email tracking-info', desc: 'Email tracking info' }
        }
    },
    {
        title: 'AI Features',
        commands: {
            'ai': { title: 'AI Assistant', cmd: 'email ai', desc: 'AI-powered email intelligence' },
            'smart-compose': { title: 'Smart Compose', cmd: 'email smart-compose', desc: 'Generate AI-powered drafts' }
        }
    }
];

// Flat lookup for email commands
const emailCommands = {};
emailCommandSections.forEach(section => {
    Object.assign(emailCommands, section.commands);
});

// Calendar Commands Data - Grouped by section
const calendarCommandSections = [
    {
        title: 'Calendars',
        commands: {
            'cal-list': { title: 'List', cmd: 'calendar list', desc: 'List all calendars' },
            'cal-show': { title: 'Show', cmd: 'calendar show', desc: 'Show calendar details', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } },
            'cal-create': { title: 'Create', cmd: 'calendar create', desc: 'Create a new calendar' },
            'cal-update': { title: 'Update', cmd: 'calendar update', desc: 'Update a calendar', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } },
            'cal-delete': { title: 'Delete', cmd: 'calendar delete', desc: 'Delete a calendar', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } }
        }
    },
    {
        title: 'Events',
        commands: {
            'events-list': {
                title: 'List',
                cmd: 'calendar events list',
                desc: 'List calendar events',
                flags: [
                    { name: 'days', type: 'number', label: 'Days', placeholder: '7', short: 'd' },
                    { name: 'limit', type: 'number', label: 'Limit', placeholder: '10', short: 'n' },
                    { name: 'show-tz', type: 'checkbox', label: 'Show timezone' },
                    { name: 'show-cancelled', type: 'checkbox', label: 'Show cancelled' },
                    { name: 'calendar', type: 'text', label: 'Calendar ID', placeholder: 'primary', short: 'c' },
                    { name: 'timezone', type: 'text', label: 'Timezone', placeholder: 'America/New_York' }
                ]
            },
            'events-show': { title: 'Show', cmd: 'calendar events show', desc: 'Show event details', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-create': { title: 'Create', cmd: 'calendar events create', desc: 'Create a new event' },
            'events-update': { title: 'Update', cmd: 'calendar events update', desc: 'Update an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-delete': { title: 'Delete', cmd: 'calendar events delete', desc: 'Delete an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-rsvp': { title: 'RSVP', cmd: 'calendar events rsvp', desc: 'RSVP to an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } }
        }
    },
    {
        title: 'Availability',
        commands: {
            'avail-check': { title: 'Check', cmd: 'calendar availability check', desc: 'Check free/busy status' },
            'avail-find': { title: 'Find', cmd: 'calendar availability find', desc: 'Find available meeting times' }
        }
    },
    {
        title: 'AI & Scheduling',
        commands: {
            'schedule': { title: 'Schedule', cmd: 'calendar schedule', desc: 'Schedule meetings with AI' },
            'find-time': { title: 'Find Time', cmd: 'calendar find-time', desc: 'Find optimal meeting times' },
            'ai': { title: 'AI Assistant', cmd: 'calendar ai', desc: 'AI calendar intelligence' },
            'recurring': { title: 'Recurring', cmd: 'calendar recurring', desc: 'Manage recurring events' },
            'virtual': { title: 'Virtual', cmd: 'calendar virtual', desc: 'Manage virtual calendars' }
        }
    }
];

// Flat lookup for calendar commands
const calendarCommands = {};
calendarCommandSections.forEach(section => {
    Object.assign(calendarCommands, section.commands);
});

// Current command selection
let currentAuthCmd = '';
let currentEmailCmd = '';
let currentCalendarCmd = '';

// =============================================================================
// Cached IDs from list commands (for suggestions in show/read/delete commands)
// =============================================================================
let cachedMessageIds = [];   // [{id: "abc123", label: "sender - subject"}, ...]
let cachedFolderIds = [];    // [{id: "folder-id", label: "INBOX (inbox)"}, ...]
let cachedScheduleIds = [];  // [{id: "schedule-id", label: "⏳ Jan 2, 2025 3:04 PM"}, ...]
let cachedThreadIds = [];    // [{id: "thread-id", label: "participants - subject"}, ...]
let cachedCalendarIds = [];  // [{id: "calendar-id", label: "Calendar Name"}, ...]
let cachedEventIds = [];     // [{id: "event-id", label: "Event Title"}, ...]
let cachedGrantIds = [];     // [{id: "grant-id", label: "email@example.com (Provider)"}, ...]

/**
 * Parse email list output to extract message IDs.
 * Format when --id flag is used:
 *   ● ★ sender@email.com    Subject line here...           2 hours ago
 *         ID: abc123def456...
 */
function parseMessageIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let lastEmailInfo = null;

    for (const line of lines) {
        // Check for ID line (starts with spaces then "ID:")
        const idMatch = line.match(/^\s+ID:\s*(\S+)/);
        if (idMatch) {
            const id = idMatch[1];
            ids.push({
                id: id,
                label: lastEmailInfo || id.substring(0, 20) + '...'
            });
            lastEmailInfo = null;
            continue;
        }

        // Try to capture email info from the line before ID
        // Format: status star from subject date
        // Look for lines that have content (not just whitespace)
        const trimmed = line.trim();
        if (trimmed && !trimmed.startsWith('Found') && !trimmed.startsWith('ID:')) {
            // Extract from and subject from the line
            // Remove status indicators (●, ★) and clean up
            const cleaned = trimmed.replace(/^[●★\s]+/, '').trim();
            if (cleaned.length > 5) {
                // Truncate for display
                lastEmailInfo = cleaned.length > 60 ? cleaned.substring(0, 57) + '...' : cleaned;
            }
        }
    }

    return ids;
}

/**
 * Parse folders list output to extract folder IDs.
 * Format when --id flag is used:
 *   ID                                   NAME                           TYPE         TOTAL   UNREAD
 *   ----------------------------------------------------------------------------------------------------
 *   abc123-def456...                     INBOX                          inbox          100        5
 */
function parseFolderIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    // Skip header lines (Folders:, empty line, column headers, separator)
    let dataStarted = false;

    for (const line of lines) {
        // Skip until we see the separator line
        if (line.includes('------')) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        // Parse data lines: ID (36 chars), NAME, TYPE, TOTAL, UNREAD
        // The ID is the first column, displayed in dim color but we can extract it
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Split by multiple spaces to get columns
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const id = parts[0].trim();
            const name = parts[1].trim();
            const type = parts.length > 2 ? parts[2].trim() : '';

            // Validate it looks like an ID (not a header)
            if (id && id.length > 10 && !id.includes('ID') && !id.includes('NAME')) {
                ids.push({
                    id: id,
                    label: `${name}${type ? ' (' + type + ')' : ''}`
                });
            }
        }
    }

    return ids;
}

/**
 * Parse scheduled list output to extract schedule IDs.
 * Format:
 *   ⏳  Schedule ID: abc123-def456
 *      Status:      pending
 *      Send at:     Jan 2, 2025 3:04 PM
 */
function parseScheduleIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let currentId = null;
    let currentStatus = '';
    let currentSendAt = '';

    for (const line of lines) {
        // Check for Schedule ID line
        const idMatch = line.match(/Schedule ID:\s*(\S+)/);
        if (idMatch) {
            // Save previous entry if exists
            if (currentId) {
                ids.push({
                    id: currentId,
                    label: `${currentStatus} - ${currentSendAt}`.trim()
                });
            }
            currentId = idMatch[1];
            currentStatus = '';
            currentSendAt = '';
            continue;
        }

        // Check for Status line
        const statusMatch = line.match(/Status:\s*(\S+)/);
        if (statusMatch && currentId) {
            const status = statusMatch[1];
            currentStatus = status === 'pending' ? '⏳' : status === 'sent' ? '✅' : '❌';
            continue;
        }

        // Check for Send at line
        const sendAtMatch = line.match(/Send at:\s*(.+)$/);
        if (sendAtMatch && currentId) {
            currentSendAt = sendAtMatch[1].trim();
            continue;
        }
    }

    // Don't forget the last entry
    if (currentId) {
        ids.push({
            id: currentId,
            label: `${currentStatus} - ${currentSendAt}`.trim()
        });
    }

    return ids;
}

/**
 * Parse threads list output to extract thread IDs.
 * Format when --id flag is used:
 *   ● ★   participants               subject                      (1)   date
 *         ID: abc123def456
 */
function parseThreadIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let lastThreadInfo = null;

    for (const line of lines) {
        // Check for ID line (starts with spaces then "ID:")
        const idMatch = line.match(/^\s+ID:\s*(\S+)/);
        if (idMatch) {
            const id = idMatch[1];
            ids.push({
                id: id,
                label: lastThreadInfo || id.substring(0, 20) + '...'
            });
            lastThreadInfo = null;
            continue;
        }

        // Try to capture thread info from the line before ID
        const trimmed = line.trim();
        if (trimmed && !trimmed.startsWith('Found') && !trimmed.startsWith('ID:')) {
            // Remove status indicators (●, ★, 📎) and clean up
            const cleaned = trimmed.replace(/^[●★📎\s]+/, '').trim();
            if (cleaned.length > 5) {
                // Truncate for display
                lastThreadInfo = cleaned.length > 60 ? cleaned.substring(0, 57) + '...' : cleaned;
            }
        }
    }

    return ids;
}

/**
 * Parse calendar list output to extract calendar IDs.
 * Format:
 *   NAME                           ID                                      PRIMARY  READ-ONLY
 *   ─────────────────────────────────────────────────────────────────────────────────────────
 *   My Calendar                    email@example.com                        Yes
 */
function parseCalendarIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    // Skip until we see the separator line
    let dataStarted = false;

    for (const line of lines) {
        if (line.includes('───') || line.includes('---')) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        const trimmed = line.trim();
        if (!trimmed) continue;

        // Parse table row - columns are separated by multiple spaces
        // Format: NAME (variable width) | ID | PRIMARY | READ-ONLY
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const name = parts[0].trim();
            const id = parts[1].trim();

            // Validate it looks like a calendar ID (contains @ or is a valid ID)
            if (id && (id.includes('@') || id.length > 10) && !id.includes('ID')) {
                const isPrimary = parts.length > 2 && parts[2].trim() === 'Yes';
                ids.push({
                    id: id,
                    label: `${name}${isPrimary ? ' (Primary)' : ''}`
                });
            }
        }
    }

    return ids;
}

/**
 * Parse events list output to extract event IDs.
 * Format:
 *   Event Title
 *     When: ...
 *     Status: ...
 *     ID: abc123def456
 */
function parseEventIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let currentTitle = null;

    for (const line of lines) {
        // Check for ID line
        const idMatch = line.match(/^\s+ID:\s*(\S+)/);
        if (idMatch) {
            const id = idMatch[1];
            ids.push({
                id: id,
                label: currentTitle || id.substring(0, 20) + '...'
            });
            currentTitle = null;
            continue;
        }

        // Capture event title (line that doesn't start with spaces and isn't "Found X event(s)")
        const trimmed = line.trim();
        if (trimmed && !line.startsWith(' ') && !trimmed.startsWith('Found') && !trimmed.startsWith('When:') && !trimmed.startsWith('Status:')) {
            currentTitle = trimmed.length > 50 ? trimmed.substring(0, 47) + '...' : trimmed;
        }
    }

    return ids;
}

/**
 * Parse auth list output to extract grant IDs.
 * Format:
 *   GRANT ID                                EMAIL                     PROVIDER      STATUS        DEFAULT
 *   abc123-def456...                        user@example.com          Google        ✓ valid       ✓
 */
function parseGrantIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Skip header line
        if (trimmed.startsWith('GRANT ID')) continue;

        // Parse data lines - columns are: GRANT ID (38), EMAIL (24), PROVIDER (12), STATUS, DEFAULT
        // Split by multiple spaces
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 3) {
            const id = parts[0].trim();
            const email = parts[1].trim();
            const provider = parts[2].trim();

            // Validate it looks like a grant ID (long alphanumeric string)
            if (id && id.length > 20 && email.includes('@')) {
                ids.push({
                    id: id,
                    label: `${email} (${provider})`
                });
            }
        }
    }

    return ids;
}

/**
 * Get cached IDs by type.
 */
function getCachedMessageIds() {
    return cachedMessageIds;
}

function getCachedFolderIds() {
    return cachedFolderIds;
}

function getCachedScheduleIds() {
    return cachedScheduleIds;
}

function getCachedThreadIds() {
    return cachedThreadIds;
}

function getCachedCalendarIds() {
    return cachedCalendarIds;
}

function getCachedEventIds() {
    return cachedEventIds;
}

function getCachedGrantIds() {
    return cachedGrantIds;
}

/**
 * Clear all cached IDs.
 */
function clearAllCachedIds() {
    cachedMessageIds = [];
    cachedFolderIds = [];
    cachedScheduleIds = [];
    cachedThreadIds = [];
    cachedCalendarIds = [];
    cachedEventIds = [];
    cachedGrantIds = [];
}

/**
 * Get total count of cached IDs.
 */
function getTotalCachedCount() {
    return cachedMessageIds.length + cachedFolderIds.length + cachedScheduleIds.length +
           cachedThreadIds.length + cachedCalendarIds.length + cachedEventIds.length +
           cachedGrantIds.length;
}

/**
 * Update the cache count badge display.
 */
function updateCacheCountBadge() {
    const count = getTotalCachedCount();

    // Update email badge
    const emailBadge = document.getElementById('cache-count-badge');
    if (emailBadge) {
        if (count > 0) {
            emailBadge.textContent = count;
            emailBadge.style.display = 'inline-flex';
        } else {
            emailBadge.style.display = 'none';
        }
    }

    // Update calendar badge
    const calendarBadge = document.getElementById('calendar-cache-count-badge');
    if (calendarBadge) {
        if (count > 0) {
            calendarBadge.textContent = count;
            calendarBadge.style.display = 'inline-flex';
        } else {
            calendarBadge.style.display = 'none';
        }
    }

    // Update auth badge
    const authBadge = document.getElementById('auth-cache-count-badge');
    if (authBadge) {
        if (count > 0) {
            authBadge.textContent = count;
            authBadge.style.display = 'inline-flex';
        } else {
            authBadge.style.display = 'none';
        }
    }
}

/**
 * Clear cache and show notification.
 */
function clearCacheAndNotify() {
    const count = getTotalCachedCount();
    if (count === 0) {
        showToast('No cached IDs to clear', 'info');
        return;
    }

    clearAllCachedIds();
    updateCacheCountBadge();
    showToast(`Cleared ${count} cached IDs`, 'success');

    // Refresh the current command's param input to remove suggestions
    if (currentEmailCmd) {
        const data = emailCommands[currentEmailCmd];
        if (data && data.param) {
            showParamInput('email', data.param, data.flags);
        }
    }
    if (currentCalendarCmd) {
        const data = calendarCommands[currentCalendarCmd];
        if (data && data.param) {
            showParamInput('calendar', data.param, data.flags);
        }
    }
    if (currentAuthCmd) {
        const data = authCommands[currentAuthCmd];
        if (data && data.param) {
            showParamInput('auth', data.param, data.flags);
        }
    }
}

// =============================================================================
// Auth Commands
// =============================================================================

function showAuthCmd(cmd) {
    const data = authCommands[cmd];
    if (!data) return;

    currentAuthCmd = cmd;

    document.querySelectorAll('#page-auth .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('auth-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('auth-detail-title').textContent = data.title;
    document.getElementById('auth-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('auth-detail-desc').textContent = data.desc || '';
    document.getElementById('auth-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('auth-output').className = 'output-pre';

    showParamInput('auth', data.param, data.flags);
}

async function runAuthCmd() {
    if (!currentAuthCmd) return;

    const data = authCommands[currentAuthCmd];
    const output = document.getElementById('auth-output');
    const btn = document.getElementById('auth-run-btn');
    const fullCmd = buildCommand(data.cmd, 'auth', data.flags);

    document.getElementById('auth-detail-cmd').textContent = 'nylas ' + fullCmd;

    btn.classList.add('loading');
    btn.innerHTML = '<span class="spinner"></span> Running...';
    output.innerHTML = '<span class="ansi-cyan">Running command...</span>';
    output.className = 'output-pre loading';

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })
        });
        const result = await res.json();

        if (result.error) {
            output.innerHTML = '<span class="ansi-red">' + esc(result.error) + '</span>';
            output.className = 'output-pre error';
            showToast('Command failed', 'error');
        } else {
            output.innerHTML = formatOutput(result.output) || '<span class="ansi-green">Command completed successfully.</span>';
            output.className = 'output-pre';
            showToast('Command completed', 'success');

            // Cache IDs from list command for suggestions
            if (result.output && currentAuthCmd === 'list') {
                const ids = parseGrantIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedGrantIds = ids;
                    showToast(`Cached ${ids.length} grant IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('auth');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

// =============================================================================
// Email Commands
// =============================================================================

function showEmailCmd(cmd) {
    const data = emailCommands[cmd];
    if (!data) return;

    currentEmailCmd = cmd;

    document.querySelectorAll('#page-email .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('email-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('email-detail-title').textContent = data.title;
    document.getElementById('email-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('email-detail-desc').textContent = data.desc || '';
    document.getElementById('email-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('email-output').className = 'output-pre';

    showParamInput('email', data.param, data.flags);
}

async function runEmailCmd() {
    if (!currentEmailCmd) return;

    const data = emailCommands[currentEmailCmd];
    const output = document.getElementById('email-output');
    const btn = document.getElementById('email-run-btn');
    const fullCmd = buildCommand(data.cmd, 'email', data.flags);

    document.getElementById('email-detail-cmd').textContent = 'nylas ' + fullCmd;

    btn.classList.add('loading');
    btn.innerHTML = '<span class="spinner"></span> Running...';
    output.innerHTML = '<span class="ansi-cyan">Running command...</span>';
    output.className = 'output-pre loading';

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })
        });
        const result = await res.json();

        if (result.error) {
            output.innerHTML = '<span class="ansi-red">' + esc(result.error) + '</span>';
            output.className = 'output-pre error';
            showToast('Command failed', 'error');
        } else {
            output.innerHTML = formatOutput(result.output) || '<span class="ansi-green">Command completed successfully.</span>';
            output.className = 'output-pre';
            showToast('Command completed', 'success');

            // Cache IDs from list commands for suggestions
            if (result.output) {
                let cached = false;
                // Cache message IDs when running email list with --id flag
                if (currentEmailCmd === 'list') {
                    const ids = parseMessageIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedMessageIds = ids;
                        showToast(`Cached ${ids.length} message IDs for quick access`, 'info');
                        cached = true;
                    }
                }
                // Cache folder IDs when running folders-list with --id flag
                else if (currentEmailCmd === 'folders-list') {
                    const ids = parseFolderIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedFolderIds = ids;
                        showToast(`Cached ${ids.length} folder IDs for quick access`, 'info');
                        cached = true;
                    }
                }
                // Cache schedule IDs when running scheduled-list
                else if (currentEmailCmd === 'scheduled-list') {
                    const ids = parseScheduleIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedScheduleIds = ids;
                        showToast(`Cached ${ids.length} schedule IDs for quick access`, 'info');
                        cached = true;
                    }
                }
                // Cache thread IDs when running threads-list with --id flag
                else if (currentEmailCmd === 'threads-list') {
                    const ids = parseThreadIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedThreadIds = ids;
                        showToast(`Cached ${ids.length} thread IDs for quick access`, 'info');
                        cached = true;
                    }
                }

                // Update badge if we cached anything
                if (cached) {
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('email');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

// =============================================================================
// Calendar Commands
// =============================================================================

// Generic function to render command sections
function renderCommandSections(containerId, sections, showFn) {
    const container = document.getElementById(containerId);
    if (!container) return;

    let html = '';
    sections.forEach(section => {
        html += `<div class="cmd-section-title">${section.title}</div>`;
        Object.entries(section.commands).forEach(([key, data]) => {
            html += `
                <div class="cmd-item" data-cmd="${key}" onclick="${showFn}('${key}')">
                    <span class="cmd-name">${data.title}</span>
                    <button class="cmd-copy" onclick="event.stopPropagation(); copyText('nylas ${data.cmd}')">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="9" y="9" width="13" height="13" rx="2"/>
                            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                        </svg>
                    </button>
                </div>`;
        });
    });
    container.innerHTML = html;
}

// Render command sections on page load
function renderAuthCommands() {
    renderCommandSections('auth-cmd-list', authCommandSections, 'showAuthCmd');
}

function renderEmailCommands() {
    renderCommandSections('email-cmd-list', emailCommandSections, 'showEmailCmd');
}

function renderCalendarCommands() {
    renderCommandSections('calendar-cmd-list', calendarCommandSections, 'showCalendarCmd');
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    renderAuthCommands();
    renderEmailCommands();
    renderCalendarCommands();
});

function showCalendarCmd(cmd) {
    const data = calendarCommands[cmd];
    if (!data) return;

    currentCalendarCmd = cmd;

    document.querySelectorAll('#page-calendar .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('calendar-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('calendar-detail-title').textContent = data.title;
    document.getElementById('calendar-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('calendar-detail-desc').textContent = data.desc || '';
    document.getElementById('calendar-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('calendar-output').className = 'output-pre';

    showParamInput('calendar', data.param, data.flags);
}

async function runCalendarCmd() {
    if (!currentCalendarCmd) return;

    const data = calendarCommands[currentCalendarCmd];
    const output = document.getElementById('calendar-output');
    const btn = document.getElementById('calendar-run-btn');
    const fullCmd = buildCommand(data.cmd, 'calendar', data.flags);

    document.getElementById('calendar-detail-cmd').textContent = 'nylas ' + fullCmd;

    btn.classList.add('loading');
    btn.innerHTML = '<span class="spinner"></span> Running...';
    output.innerHTML = '<span class="ansi-cyan">Running command...</span>';
    output.className = 'output-pre loading';

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })
        });
        const result = await res.json();

        if (result.error) {
            output.innerHTML = '<span class="ansi-red">' + esc(result.error) + '</span>';
            output.className = 'output-pre error';
            showToast('Command failed', 'error');
        } else {
            output.innerHTML = formatOutput(result.output) || '<span class="ansi-green">Command completed successfully.</span>';
            output.className = 'output-pre';
            showToast('Command completed', 'success');

            // Cache IDs from list commands for suggestions
            if (result.output) {
                let cached = false;
                // Cache calendar IDs when running calendar list
                if (currentCalendarCmd === 'cal-list') {
                    const ids = parseCalendarIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedCalendarIds = ids;
                        showToast(`Cached ${ids.length} calendar IDs for quick access`, 'info');
                        cached = true;
                    }
                }
                // Cache event IDs when running events list
                else if (currentCalendarCmd === 'events-list') {
                    const ids = parseEventIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedEventIds = ids;
                        showToast(`Cached ${ids.length} event IDs for quick access`, 'info');
                        cached = true;
                    }
                }

                // Update badge if we cached anything
                if (cached) {
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('calendar');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

// =============================================================================
// Refresh Commands
// =============================================================================

function refreshAuthCmd() {
    if (currentAuthCmd) runAuthCmd();
}

function refreshEmailCmd() {
    if (currentEmailCmd) runEmailCmd();
}

function refreshCalendarCmd() {
    if (currentCalendarCmd) runCalendarCmd();
}
