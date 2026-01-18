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

            if (result.output && currentNotetakerCmd === 'list') {
                const ids = parseNotetakerIdsFromOutput(result.output);
                if (ids.length > 0) {
                    cachedNotetakerIds = ids;
                    showToast('Cached ' + ids.length + ' notetaker IDs for quick access', 'info');
                    updateCacheCountBadge();
                }
            }
        }

        updateTimestamp('notetaker');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshNotetakerCmd() {
    if (currentNotetakerCmd) runNotetakerCmd();
}

function renderNotetakerCommands() {
    renderCommandSections('notetaker-cmd-list', notetakerCommandSections, 'showNotetakerCmd');
}
