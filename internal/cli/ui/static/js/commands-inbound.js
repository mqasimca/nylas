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

            if (result.output && currentInboundCmd === 'list') {
                const ids = parseInboxIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedInboxIds = ids;
                    showToast('Cached ' + ids.length + ' inbox IDs for quick access', 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('inbound');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshInboundCmd() {
    if (currentInboundCmd) runInboundCmd();
}

function renderInboundCommands() {
    renderCommandSections('inbound-cmd-list', inboundCommandSections, 'showInboundCmd');
}
