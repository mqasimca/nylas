// =============================================================================
// Command Output Parsers
// =============================================================================
// These functions parse CLI command output to extract IDs for autocomplete

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
