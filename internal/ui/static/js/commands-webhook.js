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

            if (result.output && currentWebhookCmd === 'list') {
                const ids = parseWebhookIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedWebhookIds = ids;
                    showToast('Cached ' + ids.length + ' webhook IDs for quick access', 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('webhook');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshWebhookCmd() {
    if (currentWebhookCmd) runWebhookCmd();
}

function renderWebhookCommands() {
    renderCommandSections('webhook-cmd-list', webhookCommandSections, 'showWebhookCmd');
}
