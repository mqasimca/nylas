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
    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('notetaker-detail-desc').textContent = data.desc || '';
    document.getElementById('notetaker-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('notetaker-output').className = 'output-pre';

    showParamInput('notetaker', data.param, data.flags);
}

async function runNotetakerCmd() {
    if (!currentNotetakerCmd) return;

    const data = notetakerCommands[currentNotetakerCmd];
    const output = document.getElementById('notetaker-output');
    const btn = document.getElementById('notetaker-run-btn');
    const fullCmd = buildCommand(data.cmd, 'notetaker', data.flags);

    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + fullCmd;

    btn.classList.add('loading');
    btn.innerHTML = '<span class="spinner"></span> Running...';
    output.innerHTML = '<span class="ansi-cyan">Running command...</span>';
    output.className = 'output-pre loading';

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })

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
    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('notetaker-detail-desc').textContent = data.desc || '';
    document.getElementById('notetaker-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('notetaker-output').className = 'output-pre';

    showParamInput('notetaker', data.param, data.flags);
}

async function runNotetakerCmd() {
    if (!currentNotetakerCmd) return;


async function runNotetakerCmd() {
    if (!currentNotetakerCmd) return;

    const data = notetakerCommands[currentNotetakerCmd];
    const output = document.getElementById('notetaker-output');
    const btn = document.getElementById('notetaker-run-btn');
    const fullCmd = buildCommand(data.cmd, 'notetaker', data.flags);

    document.getElementById('notetaker-detail-cmd').textContent = 'nylas ' + fullCmd;

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
            if (result.output && currentNotetakerCmd === 'list') {
                const ids = parseNotetakerIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedNotetakerIds = ids;
                    showToast(`Cached ${ids.length} notetaker IDs for quick access`, 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('notetaker');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshNotetakerCmd() {
    if (currentNotetakerCmd) runNotetakerCmd();
}

function renderNotetakerCommands() {
    renderCommandSections('notetaker-cmd-list', notetakerCommandSections, 'showNotetakerCmd');
}

function refreshNotetakerCmd() {
    if (currentNotetakerCmd) runNotetakerCmd();
}

function renderNotetakerCommands() {
    renderCommandSections('notetaker-cmd-list', notetakerCommandSections, 'showNotetakerCmd');
}
