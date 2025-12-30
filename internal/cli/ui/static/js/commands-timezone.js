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

function refreshTimezoneCmd() {
    if (currentTimezoneCmd) runTimezoneCmd();
}

function renderTimezoneCommands() {
    renderCommandSections('timezone-cmd-list', timezoneCommandSections, 'showTimezoneCmd');
}
