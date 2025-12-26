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
            'send': {
                title: 'Send',
                cmd: 'email send',
                desc: 'Send an email',
                flags: [
                    { name: 'to', type: 'text', label: 'To', placeholder: 'recipient@example.com', required: true, short: 't' },
                    { name: 'subject', type: 'text', label: 'Subject', placeholder: 'Email subject', required: true, short: 's' },
                    { name: 'body', type: 'textarea', label: 'Body', placeholder: 'Email body content', required: true, short: 'b' },
                    { name: 'cc', type: 'text', label: 'CC', placeholder: 'cc@example.com', short: 'c' },
                    { name: 'bcc', type: 'text', label: 'BCC', placeholder: 'bcc@example.com' },
                    { name: 'schedule', type: 'text', label: 'Schedule', placeholder: '2h or tomorrow 9am' },
                    { name: 'track-opens', type: 'checkbox', label: 'Track Opens' },
                    { name: 'track-links', type: 'checkbox', label: 'Track Links' }
                ]
            },
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
            'folders-create': { title: 'Create', cmd: 'email folders create', desc: 'Create a new folder', param: { name: 'folder-name', placeholder: 'Enter folder name...' } },
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
            'ai-analyze': { title: 'AI Analyze', cmd: 'email ai analyze', desc: 'AI inbox analysis' },
            'smart-compose': { title: 'Smart Compose', cmd: 'email smart-compose', desc: 'Generate AI-powered drafts', param: { name: 'prompt', placeholder: 'Enter prompt...' } }
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
            'events-create': {
                title: 'Create',
                cmd: 'calendar events create',
                desc: 'Create a new event',
                flags: [
                    { name: 'title', type: 'text', label: 'Title', placeholder: 'Event title', required: true, short: 't' },
                    { name: 'start', type: 'text', label: 'Start', placeholder: '2024-01-15 14:00', required: true, short: 's' },
                    { name: 'end', type: 'text', label: 'End', placeholder: '2024-01-15 15:00', short: 'e' },
                    { name: 'description', type: 'textarea', label: 'Description', placeholder: 'Event description', short: 'D' },
                    { name: 'location', type: 'text', label: 'Location', placeholder: 'Meeting room or address', short: 'l' },
                    { name: 'participant', type: 'text', label: 'Participants', placeholder: 'email1@example.com', short: 'p' },
                    { name: 'calendar', type: 'text', label: 'Calendar ID', placeholder: 'primary', short: 'c' },
                    { name: 'all-day', type: 'checkbox', label: 'All-Day Event' },
                    { name: 'busy', type: 'checkbox', label: 'Mark as Busy', default: true }
                ]
            },
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
            'ai-analyze': { title: 'AI Analyze', cmd: 'calendar ai analyze', desc: 'AI meeting analysis' },
            'ai-focus': { title: 'AI Focus', cmd: 'calendar ai focus-time', desc: 'AI focus time protection' },
            'ai-conflicts': { title: 'AI Conflicts', cmd: 'calendar ai conflicts', desc: 'AI conflict detection' }
        }
    },
    {
        title: 'Recurring',
        commands: {
            'recurring-list': { title: 'List', cmd: 'calendar recurring list', desc: 'List recurring instances', param: { name: 'event-id', placeholder: 'Enter master event ID...' } },
            'recurring-update': { title: 'Update', cmd: 'calendar recurring update', desc: 'Update instance', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'recurring-delete': { title: 'Delete', cmd: 'calendar recurring delete', desc: 'Delete instance', param: { name: 'event-id', placeholder: 'Enter event ID...' } }
        }
    },
    {
        title: 'Virtual Calendars',
        commands: {
            'virtual-list': { title: 'List', cmd: 'calendar virtual list', desc: 'List virtual calendars' },
            'virtual-show': { title: 'Show', cmd: 'calendar virtual show', desc: 'Show virtual calendar', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } },
            'virtual-create': {
                title: 'Create',
                cmd: 'calendar virtual create',
                desc: 'Create virtual calendar',
                flags: [
                    { name: 'email', type: 'text', label: 'Email', placeholder: 'Virtual calendar email identifier', required: true }
                ]
            },
            'virtual-delete': { title: 'Delete', cmd: 'calendar virtual delete', desc: 'Delete virtual calendar', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } }
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
let cachedContactIds = [];   // [{id: "contact-id", label: "John Doe (john@example.com)"}, ...]
let cachedInboxIds = [];     // [{id: "inbox-id", label: "support@app.nylas.email"}, ...]
let cachedWebhookIds = [];   // [{id: "webhook-id", label: "https://example.com/webhook"}, ...]
let cachedNotetakerIds = []; // [{id: "notetaker-id", label: "Team Standup"}, ...]

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
 * Parse contacts list output to extract contact IDs.
 * Format when --id flag is used (table format):
 *   Found X contact(s):
 *
 *   ID                                     NAME           EMAIL             PHONE    COMPANY
 *   abc123-def456-789...                   John Doe       john@example.com  +1234    Acme Inc
 */
function parseContactIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    // Skip until we find the header line with ID column
    let dataStarted = false;

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Look for header line starting with ID
        if (trimmed.startsWith('ID') && trimmed.includes('NAME')) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        // Skip separator lines
        if (trimmed.includes('───') || trimmed.includes('---')) continue;

        // Parse data rows - split by multiple spaces
        // Format: ID (36+ chars), NAME, EMAIL, PHONE, COMPANY
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const id = parts[0].trim();
            const name = parts[1].trim();
            const email = parts.length > 2 ? parts[2].trim() : '';

            // Validate it looks like a contact ID (alphanumeric string, at least 10 chars)
            if (id && id.length >= 10 && !id.includes('ID') && /^[a-zA-Z0-9_-]+$/.test(id)) {
                let label = name;
                if (email && email.includes('@')) {
                    label = `${name} (${email})`;
                }
                ids.push({
                    id: id,
                    label: label.length > 50 ? label.substring(0, 47) + '...' : label
                });
            }
        }
    }

    return ids;
}

/**
 * Parse inbound list output to extract inbox IDs.
 * Format:
 *   ID                     ADDRESS                           STATUS
 *   inbox-001              support@yourapp.nylas.email       Active
 */
function parseInboxIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let dataStarted = false;

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Look for header line with ID column
        if (trimmed.startsWith('ID') && trimmed.includes('ADDRESS')) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        // Skip separator lines and summary lines
        if (trimmed.includes('───') || trimmed.includes('---') || trimmed.includes('inboxes')) continue;

        // Parse data rows
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const id = parts[0].trim();
            const address = parts[1].trim();

            // Validate it looks like an inbox ID
            if (id && id.length >= 5 && !id.includes('ID') && address.includes('@')) {
                ids.push({
                    id: id,
                    label: address
                });
            }
        }
    }

    return ids;
}

/**
 * Parse webhook list output to extract webhook IDs.
 * Format:
 *   ID                     CALLBACK URL                           TRIGGERS        STATUS
 *   wh-001                 https://example.com/webhook/events     message.*       Active
 */
function parseWebhookIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let dataStarted = false;

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Look for header line with ID column
        if (trimmed.startsWith('ID') && (trimmed.includes('CALLBACK') || trimmed.includes('URL'))) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        // Skip separator lines and summary lines
        if (trimmed.includes('───') || trimmed.includes('---') || trimmed.includes('webhooks')) continue;

        // Parse data rows
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const id = parts[0].trim();
            const url = parts[1].trim();

            // Validate it looks like a webhook ID
            if (id && id.length >= 3 && !id.includes('ID')) {
                const label = url.length > 40 ? url.substring(0, 37) + '...' : url;
                ids.push({
                    id: id,
                    label: label
                });
            }
        }
    }

    return ids;
}

/**
 * Parse notetaker list output to extract notetaker IDs.
 * Format:
 *   ID                     MEETING                          STATUS        CREATED
 *   nt-001                 Team Standup                     Completed     Dec 24
 */
function parseNotetakerIdsFromOutput(output) {
    const ids = [];
    const lines = output.split('\n');

    let dataStarted = false;

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        // Look for header line with ID column
        if (trimmed.startsWith('ID') && trimmed.includes('MEETING')) {
            dataStarted = true;
            continue;
        }

        if (!dataStarted) continue;

        // Skip separator lines and summary lines
        if (trimmed.includes('───') || trimmed.includes('---') || trimmed.includes('notetakers')) continue;

        // Parse data rows
        const parts = trimmed.split(/\s{2,}/);
        if (parts.length >= 2) {
            const id = parts[0].trim();
            const meeting = parts[1].trim();

            // Validate it looks like a notetaker ID
            if (id && id.length >= 3 && !id.includes('ID')) {
                ids.push({
                    id: id,
                    label: meeting.length > 40 ? meeting.substring(0, 37) + '...' : meeting
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

function getCachedContactIds() {
    return cachedContactIds;
}

function getCachedInboxIds() {
    return cachedInboxIds;
}

function getCachedWebhookIds() {
    return cachedWebhookIds;
}

function getCachedNotetakerIds() {
    return cachedNotetakerIds;
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
    cachedContactIds = [];
    cachedInboxIds = [];
    cachedWebhookIds = [];
    cachedNotetakerIds = [];
}

/**
 * Get total count of cached IDs.
 */
function getTotalCachedCount() {
    return cachedMessageIds.length + cachedFolderIds.length + cachedScheduleIds.length +
           cachedThreadIds.length + cachedCalendarIds.length + cachedEventIds.length +
           cachedGrantIds.length + cachedContactIds.length + cachedInboxIds.length +
           cachedWebhookIds.length + cachedNotetakerIds.length;
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

    // Update contacts badge
    const contactsBadge = document.getElementById('contacts-cache-count-badge');
    if (contactsBadge) {
        if (count > 0) {
            contactsBadge.textContent = count;
            contactsBadge.style.display = 'inline-flex';
        } else {
            contactsBadge.style.display = 'none';
        }
    }

    // Update inbound badge
    const inboundBadge = document.getElementById('inbound-cache-count-badge');
    if (inboundBadge) {
        if (count > 0) {
            inboundBadge.textContent = count;
            inboundBadge.style.display = 'inline-flex';
        } else {
            inboundBadge.style.display = 'none';
        }
    }

    // Update webhook badge
    const webhookBadge = document.getElementById('webhook-cache-count-badge');
    if (webhookBadge) {
        if (count > 0) {
            webhookBadge.textContent = count;
            webhookBadge.style.display = 'inline-flex';
        } else {
            webhookBadge.style.display = 'none';
        }
    }

    // Update notetaker badge
    const notetakerBadge = document.getElementById('notetaker-cache-count-badge');
    if (notetakerBadge) {
        if (count > 0) {
            notetakerBadge.textContent = count;
            notetakerBadge.style.display = 'inline-flex';
        } else {
            notetakerBadge.style.display = 'none';
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
    if (currentContactsCmd) {
        const data = contactsCommands[currentContactsCmd];
        if (data && data.param) {
            showParamInput('contacts', data.param, data.flags);
        }
    }
    if (currentInboundCmd) {
        const data = inboundCommands[currentInboundCmd];
        if (data && data.param) {
            showParamInput('inbound', data.param, data.flags);
        }
    }
    if (currentWebhookCmd) {
        const data = webhookCommands[currentWebhookCmd];
        if (data && data.param) {
            showParamInput('webhook', data.param, data.flags);
        }
    }
    if (currentNotetakerCmd) {
        const data = notetakerCommands[currentNotetakerCmd];
        if (data && data.param) {
            showParamInput('notetaker', data.param, data.flags);
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
    renderContactsCommands();
    renderInboundCommands();
    renderSchedulerCommands();
    renderTimezoneCommands();
    renderWebhookCommands();
    renderOtpCommands();
    renderAdminCommands();
    renderNotetakerCommands();
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

// =============================================================================
// Contacts Commands
// =============================================================================

const contactsCommandSections = [
    {
        title: 'Contacts',
        commands: {
            'list': {
                title: 'List',
                cmd: 'contacts list',
                desc: 'List all contacts',
                flags: [
                    { name: 'id', type: 'checkbox', label: 'Show IDs', default: true }
                ]
            },
            'show': { title: 'Show', cmd: 'contacts show', desc: 'Show contact details', param: { name: 'contact-id', placeholder: 'Enter contact ID...' } },
            'search': { title: 'Search', cmd: 'contacts search', desc: 'Search contacts', param: { name: 'query', placeholder: 'Enter search query...' } },
            'create': {
                title: 'Create',
                cmd: 'contacts create',
                desc: 'Create a new contact',
                flags: [
                    { name: 'first-name', type: 'text', label: 'First Name', placeholder: 'John', short: 'f' },
                    { name: 'last-name', type: 'text', label: 'Last Name', placeholder: 'Doe', short: 'l' },
                    { name: 'email', type: 'text', label: 'Email', placeholder: 'john@example.com', short: 'e' },
                    { name: 'phone', type: 'text', label: 'Phone', placeholder: '+1-555-123-4567', short: 'p' },
                    { name: 'company', type: 'text', label: 'Company', placeholder: 'Acme Corp', short: 'c' },
                    { name: 'job-title', type: 'text', label: 'Job Title', placeholder: 'Engineer', short: 'j' },
                    { name: 'notes', type: 'textarea', label: 'Notes', placeholder: 'Notes about the contact', short: 'n' }
                ]
            },
            'update': { title: 'Update', cmd: 'contacts update', desc: 'Update a contact', param: { name: 'contact-id', placeholder: 'Enter contact ID...' } },
            'delete': { title: 'Delete', cmd: 'contacts delete', desc: 'Delete a contact', param: { name: 'contact-id', placeholder: 'Enter contact ID...' } }
        }
    },
    {
        title: 'Groups',
        commands: {
            'groups-list': { title: 'List', cmd: 'contacts groups list', desc: 'List contact groups' },
            'groups-show': { title: 'Show', cmd: 'contacts groups show', desc: 'Show group details', param: { name: 'group-id', placeholder: 'Enter group ID...' } },
            'groups-create': { title: 'Create', cmd: 'contacts groups create', desc: 'Create a contact group', param: { name: 'group-name', placeholder: 'Enter group name...' } }
        }
    },
    {
        title: 'Other',
        commands: {
            'photo-info': { title: 'Photo Info', cmd: 'contacts photo info', desc: 'Show photo info', param: { name: 'contact-id', placeholder: 'Enter contact ID...' } },
            'photo-download': { title: 'Download Photo', cmd: 'contacts photo download', desc: 'Download contact photo', param: { name: 'contact-id', placeholder: 'Enter contact ID...' } },
            'sync': { title: 'Sync Info', cmd: 'contacts sync', desc: 'Contact sync information' }
        }
    }
];

const contactsCommands = {};
contactsCommandSections.forEach(section => {
    Object.assign(contactsCommands, section.commands);
});

let currentContactsCmd = '';

function showContactsCmd(cmd) {
    const data = contactsCommands[cmd];
    if (!data) return;

    currentContactsCmd = cmd;

    document.querySelectorAll('#page-contacts .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('contacts-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('contacts-detail-title').textContent = data.title;
    document.getElementById('contacts-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('contacts-detail-desc').textContent = data.desc || '';
    document.getElementById('contacts-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('contacts-output').className = 'output-pre';

    showParamInput('contacts', data.param, data.flags);
}

async function runContactsCmd() {
    if (!currentContactsCmd) return;

    const data = contactsCommands[currentContactsCmd];
    const output = document.getElementById('contacts-output');
    const btn = document.getElementById('contacts-run-btn');
    const fullCmd = buildCommand(data.cmd, 'contacts', data.flags);

    document.getElementById('contacts-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentContactsCmd === 'list') {
                const ids = parseContactIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedContactIds = ids;
                    showToast(`Cached ${ids.length} contact IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('contacts');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshContactsCmd() {
    if (currentContactsCmd) runContactsCmd();
}

function renderContactsCommands() {
    renderCommandSections('contacts-cmd-list', contactsCommandSections, 'showContactsCmd');
}

// =============================================================================
// Inbound Commands
// =============================================================================

const inboundCommandSections = [
    {
        title: 'Inboxes',
        commands: {
            'list': { title: 'List', cmd: 'inbound list', desc: 'List all inbound inboxes' },
            'show': { title: 'Show', cmd: 'inbound show', desc: 'Show inbox details', param: { name: 'inbox-id', placeholder: 'Enter inbox ID...' } },
            'create': { title: 'Create', cmd: 'inbound create', desc: 'Create a new inbox', param: { name: 'name', placeholder: 'Enter inbox name (e.g., support)...' } },
            'delete': { title: 'Delete', cmd: 'inbound delete', desc: 'Delete an inbox', param: { name: 'inbox-id', placeholder: 'Enter inbox ID...' } }
        }
    },
    {
        title: 'Messages',
        commands: {
            'messages': { title: 'Messages', cmd: 'inbound messages', desc: 'View inbox messages', param: { name: 'inbox-id', placeholder: 'Enter inbox ID...' } },
            'monitor': { title: 'Monitor', cmd: 'inbound monitor', desc: 'Monitor for new messages', param: { name: 'inbox-id', placeholder: 'Enter inbox ID...' } }
        }
    }
];

const inboundCommands = {};
inboundCommandSections.forEach(section => {
    Object.assign(inboundCommands, section.commands);
});

let currentInboundCmd = '';

function showInboundCmd(cmd) {
    const data = inboundCommands[cmd];
    if (!data) return;

    currentInboundCmd = cmd;

    document.querySelectorAll('#page-inbound .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('inbound-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('inbound-detail-title').textContent = data.title;
    document.getElementById('inbound-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('inbound-detail-desc').textContent = data.desc || '';
    document.getElementById('inbound-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('inbound-output').className = 'output-pre';

    showParamInput('inbound', data.param, data.flags);
}

async function runInboundCmd() {
    if (!currentInboundCmd) return;

    const data = inboundCommands[currentInboundCmd];
    const output = document.getElementById('inbound-output');
    const btn = document.getElementById('inbound-run-btn');
    const fullCmd = buildCommand(data.cmd, 'inbound', data.flags);

    document.getElementById('inbound-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentInboundCmd === 'list') {
                const ids = parseInboxIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedInboxIds = ids;
                    showToast(`Cached ${ids.length} inbox IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('inbound');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshInboundCmd() {
    if (currentInboundCmd) runInboundCmd();
}

function renderInboundCommands() {
    renderCommandSections('inbound-cmd-list', inboundCommandSections, 'showInboundCmd');
}

// =============================================================================
// Scheduler Commands
// =============================================================================

const schedulerCommandSections = [
    {
        title: 'Configurations',
        commands: {
            'config-list': { title: 'List', cmd: 'scheduler configurations list', desc: 'List scheduler configurations' },
            'config-show': { title: 'Show', cmd: 'scheduler configurations show', desc: 'Show configuration details', param: { name: 'config-id', placeholder: 'Enter configuration ID...' } },
            'config-create': {
                title: 'Create',
                cmd: 'scheduler configurations create',
                desc: 'Create a scheduler configuration',
                flags: [
                    { name: 'name', type: 'text', label: 'Name', placeholder: 'Configuration name', required: true },
                    { name: 'title', type: 'text', label: 'Title', placeholder: 'Event title', required: true },
                    { name: 'participants', type: 'text', label: 'Participants', placeholder: 'email1@example.com,email2@example.com', required: true },
                    { name: 'duration', type: 'number', label: 'Duration (min)', placeholder: '30' },
                    { name: 'location', type: 'text', label: 'Location', placeholder: 'Meeting location' }
                ]
            }
        }
    },
    {
        title: 'Pages',
        commands: {
            'page-list': { title: 'List', cmd: 'scheduler pages list', desc: 'List scheduler pages' },
            'page-show': { title: 'Show', cmd: 'scheduler pages show', desc: 'Show page details', param: { name: 'page-id', placeholder: 'Enter page ID...' } },
            'page-create': {
                title: 'Create',
                cmd: 'scheduler pages create',
                desc: 'Create a scheduler page',
                flags: [
                    { name: 'config-id', type: 'text', label: 'Config ID', placeholder: 'Configuration ID', required: true },
                    { name: 'name', type: 'text', label: 'Name', placeholder: 'Page name', required: true },
                    { name: 'slug', type: 'text', label: 'Slug', placeholder: 'URL slug (optional)' }
                ]
            }
        }
    },
    {
        title: 'Sessions',
        commands: {
            'session-create': {
                title: 'Create',
                cmd: 'scheduler sessions create',
                desc: 'Create a scheduling session',
                flags: [
                    { name: 'config-id', type: 'text', label: 'Config ID', placeholder: 'Configuration ID', required: true },
                    { name: 'ttl', type: 'number', label: 'TTL (min)', placeholder: '30' }
                ]
            },
            'session-show': { title: 'Show', cmd: 'scheduler sessions show', desc: 'Show session details', param: { name: 'session-id', placeholder: 'Enter session ID...' } }
        }
    },
    {
        title: 'Bookings',
        commands: {
            'booking-list': { title: 'List', cmd: 'scheduler bookings list', desc: 'List scheduler bookings' },
            'booking-show': { title: 'Show', cmd: 'scheduler bookings show', desc: 'Show booking details', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } },
            'booking-confirm': { title: 'Confirm', cmd: 'scheduler bookings confirm', desc: 'Confirm a booking', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } },
            'booking-cancel': { title: 'Cancel', cmd: 'scheduler bookings cancel', desc: 'Cancel a booking', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } }
        }
    }
];

const schedulerCommands = {};
schedulerCommandSections.forEach(section => {
    Object.assign(schedulerCommands, section.commands);
});

let currentSchedulerCmd = '';

function showSchedulerCmd(cmd) {
    const data = schedulerCommands[cmd];
    if (!data) return;

    currentSchedulerCmd = cmd;

    document.querySelectorAll('#page-scheduler .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('scheduler-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('scheduler-detail-title').textContent = data.title;
    document.getElementById('scheduler-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('scheduler-detail-desc').textContent = data.desc || '';
    document.getElementById('scheduler-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('scheduler-output').className = 'output-pre';

    showParamInput('scheduler', data.param, data.flags);
}

async function runSchedulerCmd() {
    if (!currentSchedulerCmd) return;

    const data = schedulerCommands[currentSchedulerCmd];
    const output = document.getElementById('scheduler-output');
    const btn = document.getElementById('scheduler-run-btn');
    const fullCmd = buildCommand(data.cmd, 'scheduler', data.flags);

    document.getElementById('scheduler-detail-cmd').textContent = 'nylas ' + fullCmd;

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
        }

        updateTimestamp('scheduler');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshSchedulerCmd() {
    if (currentSchedulerCmd) runSchedulerCmd();
}

function renderSchedulerCommands() {
    renderCommandSections('scheduler-cmd-list', schedulerCommandSections, 'showSchedulerCmd');
}

// =============================================================================
// Timezone Commands
// =============================================================================

const timezoneCommandSections = [
    {
        title: 'Information',
        commands: {
            'list': { title: 'List', cmd: 'timezone list', desc: 'List all time zones' },
            'info': { title: 'Info', cmd: 'timezone info', desc: 'Get time zone info', param: { name: 'zone', placeholder: 'e.g., America/New_York' } }
        }
    },
    {
        title: 'Conversion',
        commands: {
            'convert': {
                title: 'Convert',
                cmd: 'timezone convert',
                desc: 'Convert time between zones',
                flags: [
                    { name: 'from', type: 'text', label: 'From Zone', placeholder: 'America/New_York', short: 'f' },
                    { name: 'to', type: 'text', label: 'To Zone', placeholder: 'Asia/Tokyo', short: 't' },
                    { name: 'time', type: 'text', label: 'Time', placeholder: '2024-01-15 10:00' }
                ]
            },
            'find-meeting': {
                title: 'Find Meeting',
                cmd: 'timezone find-meeting',
                desc: 'Find meeting times across zones',
                flags: [
                    { name: 'zones', type: 'text', label: 'Zones', placeholder: 'America/New_York,Europe/London,Asia/Tokyo', short: 'z' },
                    { name: 'duration', type: 'text', label: 'Duration', placeholder: '1h', short: 'd' }
                ]
            }
        }
    },
    {
        title: 'DST',
        commands: {
            'dst': {
                title: 'DST Transitions',
                cmd: 'timezone dst',
                desc: 'Check DST transitions',
                flags: [
                    { name: 'zone', type: 'text', label: 'Zone', placeholder: 'America/New_York', short: 'z' },
                    { name: 'year', type: 'number', label: 'Year', placeholder: '2025', short: 'y' }
                ]
            }
        }
    }
];

const timezoneCommands = {};
timezoneCommandSections.forEach(section => {
    Object.assign(timezoneCommands, section.commands);
});

let currentTimezoneCmd = '';

function showTimezoneCmd(cmd) {
    const data = timezoneCommands[cmd];
    if (!data) return;

    currentTimezoneCmd = cmd;

    document.querySelectorAll('#page-timezone .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('timezone-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('timezone-detail-title').textContent = data.title;
    document.getElementById('timezone-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('timezone-detail-desc').textContent = data.desc || '';
    document.getElementById('timezone-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('timezone-output').className = 'output-pre';

    showParamInput('timezone', data.param, data.flags);
}

async function runTimezoneCmd() {
    if (!currentTimezoneCmd) return;

    const data = timezoneCommands[currentTimezoneCmd];
    const output = document.getElementById('timezone-output');
    const btn = document.getElementById('timezone-run-btn');
    const fullCmd = buildCommand(data.cmd, 'timezone', data.flags);

    document.getElementById('timezone-detail-cmd').textContent = 'nylas ' + fullCmd;

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
        }

        updateTimestamp('timezone');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshTimezoneCmd() {
    if (currentTimezoneCmd) runTimezoneCmd();
}

function renderTimezoneCommands() {
    renderCommandSections('timezone-cmd-list', timezoneCommandSections, 'showTimezoneCmd');
}

// =============================================================================
// Webhook Commands
// =============================================================================

const webhookCommandSections = [
    {
        title: 'Webhooks',
        commands: {
            'list': { title: 'List', cmd: 'webhook list', desc: 'List all webhooks' },
            'show': { title: 'Show', cmd: 'webhook show', desc: 'Show webhook details', param: { name: 'webhook-id', placeholder: 'Enter webhook ID...' } },
            'create': {
                title: 'Create',
                cmd: 'webhook create',
                desc: 'Create a new webhook',
                flags: [
                    { name: 'url', type: 'text', label: 'Webhook URL', placeholder: 'https://example.com/webhook', required: true },
                    { name: 'triggers', type: 'text', label: 'Triggers', placeholder: 'message.created,event.created', required: true },
                    { name: 'description', type: 'text', label: 'Description', placeholder: 'My webhook description', short: 'd' },
                    { name: 'notify', type: 'text', label: 'Notify Email', placeholder: 'admin@example.com' }
                ]
            },
            'update': { title: 'Update', cmd: 'webhook update', desc: 'Update a webhook', param: { name: 'webhook-id', placeholder: 'Enter webhook ID...' } },
            'delete': { title: 'Delete', cmd: 'webhook delete', desc: 'Delete a webhook', param: { name: 'webhook-id', placeholder: 'Enter webhook ID...' } }
        }
    },
    {
        title: 'Tools',
        commands: {
            'triggers': { title: 'Triggers', cmd: 'webhook triggers', desc: 'List available trigger types' },
            'test': { title: 'Test', cmd: 'webhook test', desc: 'Test webhook functionality' },
            'server': { title: 'Server', cmd: 'webhook server', desc: 'Start local webhook server' }
        }
    }
];

const webhookCommands = {};
webhookCommandSections.forEach(section => {
    Object.assign(webhookCommands, section.commands);
});

let currentWebhookCmd = '';

function showWebhookCmd(cmd) {
    const data = webhookCommands[cmd];
    if (!data) return;

    currentWebhookCmd = cmd;

    document.querySelectorAll('#page-webhook .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('webhook-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('webhook-detail-title').textContent = data.title;
    document.getElementById('webhook-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('webhook-detail-desc').textContent = data.desc || '';
    document.getElementById('webhook-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('webhook-output').className = 'output-pre';

    showParamInput('webhook', data.param, data.flags);
}

async function runWebhookCmd() {
    if (!currentWebhookCmd) return;

    const data = webhookCommands[currentWebhookCmd];
    const output = document.getElementById('webhook-output');
    const btn = document.getElementById('webhook-run-btn');
    const fullCmd = buildCommand(data.cmd, 'webhook', data.flags);

    document.getElementById('webhook-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentWebhookCmd === 'list') {
                const ids = parseWebhookIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedWebhookIds = ids;
                    showToast(`Cached ${ids.length} webhook IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('webhook');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshWebhookCmd() {
    if (currentWebhookCmd) runWebhookCmd();
}

function renderWebhookCommands() {
    renderCommandSections('webhook-cmd-list', webhookCommandSections, 'showWebhookCmd');
}

// =============================================================================
// OTP Commands
// =============================================================================

const otpCommandSections = [
    {
        title: 'OTP Management',
        commands: {
            'get': { title: 'Get', cmd: 'otp get', desc: 'Get the latest OTP code' },
            'watch': { title: 'Watch', cmd: 'otp watch', desc: 'Watch for new OTP codes' },
            'list': { title: 'List', cmd: 'otp list', desc: 'List configured accounts' },
            'messages': { title: 'Messages', cmd: 'otp messages', desc: 'Show recent OTP messages' }
        }
    }
];

const otpCommands = {};
otpCommandSections.forEach(section => {
    Object.assign(otpCommands, section.commands);
});

let currentOtpCmd = '';

function showOtpCmd(cmd) {
    const data = otpCommands[cmd];
    if (!data) return;

    currentOtpCmd = cmd;

    document.querySelectorAll('#page-otp .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('otp-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('otp-detail-title').textContent = data.title;
    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('otp-detail-desc').textContent = data.desc || '';
    document.getElementById('otp-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('otp-output').className = 'output-pre';

    showParamInput('otp', data.param, data.flags);
}

async function runOtpCmd() {
    if (!currentOtpCmd) return;

    const data = otpCommands[currentOtpCmd];
    const output = document.getElementById('otp-output');
    const btn = document.getElementById('otp-run-btn');
    const fullCmd = buildCommand(data.cmd, 'otp', data.flags);

    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + fullCmd;

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
        }

        updateTimestamp('otp');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshOtpCmd() {
    if (currentOtpCmd) runOtpCmd();
}

function renderOtpCommands() {
    renderCommandSections('otp-cmd-list', otpCommandSections, 'showOtpCmd');
}

// =============================================================================
// Admin Commands
// =============================================================================

const adminCommandSections = [
    {
        title: 'Applications',
        commands: {
            'apps-list': { title: 'List', cmd: 'admin applications list', desc: 'List applications' },
            'apps-show': { title: 'Show', cmd: 'admin applications show', desc: 'Show application details', param: { name: 'app-id', placeholder: 'Enter application ID...' } },
            'apps-create': {
                title: 'Create',
                cmd: 'admin applications create',
                desc: 'Create an application',
                flags: [
                    { name: 'name', type: 'text', label: 'Name', placeholder: 'Application name', required: true },
                    { name: 'region', type: 'text', label: 'Region', placeholder: 'us or eu' },
                    { name: 'callback-uris', type: 'text', label: 'Callback URIs', placeholder: 'https://example.com/callback' }
                ]
            }
        }
    },
    {
        title: 'Connectors',
        commands: {
            'connectors-list': { title: 'List', cmd: 'admin connectors list', desc: 'List connectors' },
            'connectors-show': { title: 'Show', cmd: 'admin connectors show', desc: 'Show connector details', param: { name: 'connector-id', placeholder: 'Enter connector ID...' } }
        }
    },
    {
        title: 'Credentials',
        commands: {
            'credentials-list': { title: 'List', cmd: 'admin credentials list', desc: 'List credentials' },
            'credentials-show': { title: 'Show', cmd: 'admin credentials show', desc: 'Show credential details', param: { name: 'credential-id', placeholder: 'Enter credential ID...' } }
        }
    },
    {
        title: 'Grants',
        commands: {
            'grants-list': { title: 'List', cmd: 'admin grants list', desc: 'List grants' },
            'grants-stats': { title: 'Stats', cmd: 'admin grants stats', desc: 'Show grant statistics' }
        }
    }
];

const adminCommands = {};
adminCommandSections.forEach(section => {
    Object.assign(adminCommands, section.commands);
});

let currentAdminCmd = '';

function showAdminCmd(cmd) {
    const data = adminCommands[cmd];
    if (!data) return;

    currentAdminCmd = cmd;

    document.querySelectorAll('#page-admin .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('admin-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('admin-detail-title').textContent = data.title;
    document.getElementById('admin-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('admin-detail-desc').textContent = data.desc || '';
    document.getElementById('admin-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('admin-output').className = 'output-pre';

    showParamInput('admin', data.param, data.flags);
}

async function runAdminCmd() {
    if (!currentAdminCmd) return;

    const data = adminCommands[currentAdminCmd];
    const output = document.getElementById('admin-output');
    const btn = document.getElementById('admin-run-btn');
    const fullCmd = buildCommand(data.cmd, 'admin', data.flags);

    document.getElementById('admin-detail-cmd').textContent = 'nylas ' + fullCmd;

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
        }

        updateTimestamp('admin');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshAdminCmd() {
    if (currentAdminCmd) runAdminCmd();
}

function renderAdminCommands() {
    renderCommandSections('admin-cmd-list', adminCommandSections, 'showAdminCmd');
}

// =============================================================================
// Notetaker Commands
// =============================================================================

const notetakerCommandSections = [
    {
        title: 'Notetakers',
        commands: {
            'list': { title: 'List', cmd: 'notetaker list', desc: 'List all notetakers' },
            'show': { title: 'Show', cmd: 'notetaker show', desc: 'Show notetaker details', param: { name: 'notetaker-id', placeholder: 'Enter notetaker ID...' } },
            'create': { title: 'Create', cmd: 'notetaker create', desc: 'Create a new notetaker' },
            'delete': { title: 'Delete', cmd: 'notetaker delete', desc: 'Delete a notetaker', param: { name: 'notetaker-id', placeholder: 'Enter notetaker ID...' } }
        }
    },
    {
        title: 'Media',
        commands: {
            'media': { title: 'Media', cmd: 'notetaker media', desc: 'Get recording/transcript', param: { name: 'notetaker-id', placeholder: 'Enter notetaker ID...' } }
        }
    }
];

const notetakerCommands = {};
notetakerCommandSections.forEach(section => {
    Object.assign(notetakerCommands, section.commands);
});

let currentNotetakerCmd = '';

function showNotetakerCmd(cmd) {
    const data = notetakerCommands[cmd];
    if (!data) return;

    currentNotetakerCmd = cmd;

    document.querySelectorAll('#page-notetaker .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('notetaker-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('notetaker-detail-title').textContent = data.title;
    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('notetaker-detail-desc').textContent = data.desc || '';
    document.getElementById('notetaker-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('notetaker-output').className = 'output-pre';

    showParamInput('notetaker', data.param, data.flags);
}

async function runNotetakerCmd() {
    if (!currentNotetakerCmd) return;

    const data = notetakerCommands[currentNotetakerCmd];
    const output = document.getElementById('notetaker-output');
    const btn = document.getElementById('notetaker-run-btn');
    const fullCmd = buildCommand(data.cmd, 'notetaker', data.flags);

    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentNotetakerCmd === 'list') {
                const ids = parseNotetakerIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedNotetakerIds = ids;
                    showToast(`Cached ${ids.length} notetaker IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('notetaker');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshNotetakerCmd() {
    if (currentNotetakerCmd) runNotetakerCmd();
}

function renderNotetakerCommands() {
    renderCommandSections('notetaker-cmd-list', notetakerCommandSections, 'showNotetakerCmd');
}
