// =============================================================================
// Email Commands
// =============================================================================

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
                flags: [{ name: 'id', type: 'checkbox', label: 'Show IDs', default: true }]
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
                flags: [{ name: 'id', type: 'checkbox', label: 'Show IDs', default: true }]
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

const emailCommands = {};
emailCommandSections.forEach(section => {
    Object.assign(emailCommands, section.commands);
});

let currentEmailCmd = '';

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

    setButtonLoading(btn, true);
    setOutputLoading(output);

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })
        });
        const result = await res.json();

        if (result.error) {
            setOutputError(output, result.error);
            showToast('Command failed', 'error');
        } else {
            setOutputSuccess(output, result.output);
            showToast('Command completed', 'success');

            if (result.output) {
                let cached = false;
                if (currentEmailCmd === 'list') {
                    const ids = parseMessageIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedMessageIds = ids;
                        showToast('Cached ' + ids.length + ' message IDs for quick access', 'info');
                        cached = true;
                    }
                } else if (currentEmailCmd === 'folders-list') {
                    const ids = parseFolderIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedFolderIds = ids;
                        showToast('Cached ' + ids.length + ' folder IDs for quick access', 'info');
                        cached = true;
                    }
                } else if (currentEmailCmd === 'scheduled-list') {
                    const ids = parseScheduleIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedScheduleIds = ids;
                        showToast('Cached ' + ids.length + ' schedule IDs for quick access', 'info');
                        cached = true;
                    }
                } else if (currentEmailCmd === 'threads-list') {
                    const ids = parseThreadIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedThreadIds = ids;
                        showToast('Cached ' + ids.length + ' thread IDs for quick access', 'info');
                        cached = true;
                    }
                }
                if (cached) updateCacheCountBadge();
            }
        }

        updateTimestamp('email');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshEmailCmd() {
    if (currentEmailCmd) runEmailCmd();
}

function renderEmailCommands() {
    renderCommandSections('email-cmd-list', emailCommandSections, 'showEmailCmd');
}
