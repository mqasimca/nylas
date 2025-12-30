// =============================================================================
// Cache Management Functions
// =============================================================================
// Functions for managing cached IDs from command output

/**
 * Get cached IDs by type - simple accessors
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
 * Note: This function depends on command objects being loaded (e.g., emailCommands, calendarCommands)
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
    // These command objects are defined in their respective command files
    if (typeof currentEmailCmd !== 'undefined' && currentEmailCmd && typeof emailCommands !== 'undefined') {
        const data = emailCommands[currentEmailCmd];
        if (data && data.param) {
            showParamInput('email', data.param, data.flags);
        }
    }
    if (typeof currentCalendarCmd !== 'undefined' && currentCalendarCmd && typeof calendarCommands !== 'undefined') {
        const data = calendarCommands[currentCalendarCmd];
        if (data && data.param) {
            showParamInput('calendar', data.param, data.flags);
        }
    }
    if (typeof currentAuthCmd !== 'undefined' && currentAuthCmd && typeof authCommands !== 'undefined') {
        const data = authCommands[currentAuthCmd];
        if (data && data.param) {
            showParamInput('auth', data.param, data.flags);
        }
    }
    if (typeof currentContactsCmd !== 'undefined' && currentContactsCmd && typeof contactsCommands !== 'undefined') {
        const data = contactsCommands[currentContactsCmd];
        if (data && data.param) {
            showParamInput('contacts', data.param, data.flags);
        }
    }
    if (typeof currentInboundCmd !== 'undefined' && currentInboundCmd && typeof inboundCommands !== 'undefined') {
        const data = inboundCommands[currentInboundCmd];
        if (data && data.param) {
            showParamInput('inbound', data.param, data.flags);
        }
    }
    if (typeof currentWebhookCmd !== 'undefined' && currentWebhookCmd && typeof webhookCommands !== 'undefined') {
        const data = webhookCommands[currentWebhookCmd];
        if (data && data.param) {
            showParamInput('webhook', data.param, data.flags);
        }
    }
    if (typeof currentNotetakerCmd !== 'undefined' && currentNotetakerCmd && typeof notetakerCommands !== 'undefined') {
        const data = notetakerCommands[currentNotetakerCmd];
        if (data && data.param) {
            showParamInput('notetaker', data.param, data.flags);
        }
    }
}
