// =============================================================================
// Shared Variables for Command System
// =============================================================================

// Current command selection (tracks which command is active in each category)
let currentAuthCmd = '';
let currentEmailCmd = '';
let currentCalendarCmd = '';
let currentContactsCmd = '';
let currentInboundCmd = '';
let currentSchedulerCmd = '';
let currentTimezoneCmd = '';
let currentWebhookCmd = '';
let currentOtpCmd = '';
let currentAdminCmd = '';
let currentNotetakerCmd = '';

// =============================================================================
// Cached IDs from list commands (for autocomplete suggestions)
// =============================================================================

// Format: [{id: "abc123", label: "Display text"}, ...]
let cachedMessageIds = [];   // Email messages
let cachedFolderIds = [];    // Email folders
let cachedScheduleIds = [];  // Scheduled messages
let cachedThreadIds = [];    // Email threads
let cachedCalendarIds = [];  // Calendars
let cachedEventIds = [];     // Calendar events
let cachedGrantIds = [];     // Auth grants
let cachedContactIds = [];   // Contacts
let cachedInboxIds = [];     // Inbound inboxes
let cachedWebhookIds = [];   // Webhooks
let cachedNotetakerIds = []; // Notetaker sessions
