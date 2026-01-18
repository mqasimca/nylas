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

    const data = timezoneCommands[currentTimezoneCmd];
    const output = document.getElementById('timezone-output');
    const btn = document.getElementById('timezone-run-btn');
    const fullCmd = buildCommand(data.cmd, 'timezone', data.flags);

    document.getElementById('timezone-detail-cmd').textContent = 'nylas ' + fullCmd;

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

        updateTimestamp('timezone');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshTimezoneCmd() {
    if (currentTimezoneCmd) runTimezoneCmd();
}

function renderTimezoneCommands() {
    renderCommandSections('timezone-cmd-list', timezoneCommandSections, 'showTimezoneCmd');
}
