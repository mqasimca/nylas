// =============================================================================
// Otp Commands
// =============================================================================

const otpCommandSections = [
    {
        title: 'OTP Management',
        commands: {
            'get': { title: 'Get', cmd: 'otp get', desc: 'Get the latest OTP code' },
            'watch': { title: 'Watch', cmd: 'otp watch', desc: 'Watch for new OTP codes' },
            'list': { title: 'List', cmd: 'otp list', desc: 'List configured accounts' },
            'messages': { title: 'Messages', cmd: 'otp messages', desc: 'Show recent OTP messages' }
        }
    }
];

const otpCommands = {};
otpCommandSections.forEach(section => {
    Object.assign(otpCommands, section.commands);
});

let currentOtpCmd = '';

function showOtpCmd(cmd) {
    const data = otpCommands[cmd];
    if (!data) return;

    currentOtpCmd = cmd;

    document.querySelectorAll('#page-otp .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('otp-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('otp-detail-title').textContent = data.title;
    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + data.cmd;

function showOtpCmd(cmd) {
    const data = otpCommands[cmd];
    if (!data) return;

    currentOtpCmd = cmd;

    document.querySelectorAll('#page-otp .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('otp-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('otp-detail-title').textContent = data.title;
    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('otp-detail-desc').textContent = data.desc || '';
    document.getElementById('otp-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('otp-output').className = 'output-pre';

    showParamInput('otp', data.param, data.flags);
}

async function runOtpCmd() {
    if (!currentOtpCmd) return;


async function runOtpCmd() {
    if (!currentOtpCmd) return;

    const data = otpCommands[currentOtpCmd];
    const output = document.getElementById('otp-output');
    const btn = document.getElementById('otp-run-btn');
    const fullCmd = buildCommand(data.cmd, 'otp', data.flags);

    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + fullCmd;

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

        updateTimestamp('otp');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshOtpCmd() {
    if (currentOtpCmd) runOtpCmd();
}

function renderOtpCommands() {
    renderCommandSections('otp-cmd-list', otpCommandSections, 'showOtpCmd');
}

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

function refreshOtpCmd() {
    if (currentOtpCmd) runOtpCmd();
}

function renderOtpCommands() {
    renderCommandSections('otp-cmd-list', otpCommandSections, 'showOtpCmd');
}
