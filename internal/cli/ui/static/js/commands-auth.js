// =============================================================================
// Auth Commands
// =============================================================================

// Auth Commands Data - Grouped by section
const authCommandSections = [
    {
        title: 'Authentication',
        commands: {
            'login': {
                title: 'Login',
                cmd: 'auth login',
                desc: 'Authenticate with an email provider',
                flags: [
                    { name: 'provider', type: 'text', label: 'Provider', placeholder: 'google or microsoft', short: 'p' }
                ]
            },
            'logout': { title: 'Logout', cmd: 'auth logout', desc: 'Revoke current authentication' },
            'status': { title: 'Status', cmd: 'auth status', desc: 'Show authentication status' },
            'whoami': { title: 'Who Am I', cmd: 'auth whoami', desc: 'Show current user info' }
        }
    },
    {
        title: 'Accounts',
        commands: {
            'list': { title: 'List', cmd: 'auth list', desc: 'List all authenticated accounts' },
            'show': { title: 'Show', cmd: 'auth show', desc: 'Show detailed grant information', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'switch': { title: 'Switch', cmd: 'auth switch', desc: 'Switch between accounts', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'add': { title: 'Add', cmd: 'auth add', desc: 'Manually add an existing grant', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } },
            'remove': { title: 'Remove', cmd: 'auth remove', desc: 'Remove grant from local config', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'revoke': { title: 'Revoke', cmd: 'auth revoke', desc: 'Permanently revoke a grant', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } }
        }
    },
    {
        title: 'Configuration',
        commands: {
            'config': {
                title: 'Config',
                cmd: 'auth config',
                desc: 'Configure API credentials',
                flags: [
                    { name: 'api-key', type: 'text', label: 'API Key', placeholder: 'Your Nylas API key' },
                    { name: 'region', type: 'text', label: 'Region', placeholder: 'us or eu (default: us)', short: 'r' },
                    { name: 'client-id', type: 'text', label: 'Client ID', placeholder: 'Auto-detected if not provided' },
                    { name: 'reset', type: 'checkbox', label: 'Reset configuration' }
                ]
            },
            'providers': { title: 'Providers', cmd: 'auth providers', desc: 'List available providers' },
            'detect': { title: 'Detect', cmd: 'auth detect', desc: 'Detect provider from email', param: { name: 'email', placeholder: 'Enter email address...' } },
            'scopes': { title: 'Scopes', cmd: 'auth scopes', desc: 'Show OAuth scopes for a grant', param: { name: 'grant-id', placeholder: 'Enter grant ID or email...' } },
            'token': { title: 'Token', cmd: 'auth token', desc: 'Show or copy API key' },
            'migrate': { title: 'Migrate', cmd: 'auth migrate', desc: 'Migrate to system keyring' }
        }
    }
];

// Flat lookup for auth commands
const authCommands = {};
authCommandSections.forEach(section => {
    Object.assign(authCommands, section.commands);
});

function showAuthCmd(cmd) {
    const data = authCommands[cmd];
    if (!data) return;

    currentAuthCmd = cmd;

    document.querySelectorAll('#page-auth .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('auth-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('auth-detail-title').textContent = data.title;
    document.getElementById('auth-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('auth-detail-desc').textContent = data.desc || '';
    document.getElementById('auth-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('auth-output').className = 'output-pre';

    showParamInput('auth', data.param, data.flags);
}

async function runAuthCmd() {
    if (!currentAuthCmd) return;

    const data = authCommands[currentAuthCmd];
    const output = document.getElementById('auth-output');
    const btn = document.getElementById('auth-run-btn');
    const fullCmd = buildCommand(data.cmd, 'auth', data.flags);

    document.getElementById('auth-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentAuthCmd === 'list') {
                const ids = parseGrantIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedGrantIds = ids;
                    showToast(`Cached ${ids.length} grant IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('auth');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshAuthCmd() {
    if (currentAuthCmd) runAuthCmd();
}

function renderAuthCommands() {
    renderCommandSections('auth-cmd-list', authCommandSections, 'showAuthCmd');
}
