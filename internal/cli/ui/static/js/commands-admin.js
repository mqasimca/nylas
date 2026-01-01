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
    document.getElementById('admin-detail-desc').textContent = data.desc || '';
    document.getElementById('admin-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('admin-output').className = 'output-pre';

    showParamInput('admin', data.param, data.flags);
}

async function runAdminCmd() {
    if (!currentAdminCmd) return;

    const data = adminCommands[currentAdminCmd];
    const output = document.getElementById('admin-output');
    const btn = document.getElementById('admin-run-btn');
    const fullCmd = buildCommand(data.cmd, 'admin', data.flags);

    document.getElementById('admin-detail-cmd').textContent = 'nylas ' + fullCmd;

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
        }

        updateTimestamp('admin');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshAdminCmd() {
    if (currentAdminCmd) runAdminCmd();
}

function renderAdminCommands() {
    renderCommandSections('admin-cmd-list', adminCommandSections, 'showAdminCmd');
}
