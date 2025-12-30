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

function refreshAdminCmd() {
    if (currentAdminCmd) runAdminCmd();
}

function renderAdminCommands() {
    renderCommandSections('admin-cmd-list', adminCommandSections, 'showAdminCmd');
}
