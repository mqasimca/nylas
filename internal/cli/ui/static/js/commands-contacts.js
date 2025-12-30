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


function refreshContactsCmd() {
    if (currentContactsCmd) runContactsCmd();
}

function renderContactsCommands() {
    renderCommandSections('contacts-cmd-list', contactsCommandSections, 'showContactsCmd');
}
