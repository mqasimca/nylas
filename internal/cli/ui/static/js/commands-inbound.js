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

function refreshInboundCmd() {
    if (currentInboundCmd) runInboundCmd();
}

function renderInboundCommands() {
    renderCommandSections('inbound-cmd-list', inboundCommandSections, 'showInboundCmd');
}
