// =============================================================================
// Scheduler Commands
// =============================================================================

const schedulerCommandSections = [
    {
        title: 'Configurations',
        commands: {
            'config-list': { title: 'List', cmd: 'scheduler configurations list', desc: 'List scheduler configurations' },
            'config-show': { title: 'Show', cmd: 'scheduler configurations show', desc: 'Show configuration details', param: { name: 'config-id', placeholder: 'Enter configuration ID...' } },
            'config-create': {
                title: 'Create',
                cmd: 'scheduler configurations create',
                desc: 'Create a scheduler configuration',
                flags: [
                    { name: 'name', type: 'text', label: 'Name', placeholder: 'Configuration name', required: true },
                    { name: 'title', type: 'text', label: 'Title', placeholder: 'Event title', required: true },
                    { name: 'participants', type: 'text', label: 'Participants', placeholder: 'email1@example.com,email2@example.com', required: true },
                    { name: 'duration', type: 'number', label: 'Duration (min)', placeholder: '30' },
                    { name: 'location', type: 'text', label: 'Location', placeholder: 'Meeting location' }
                ]
            }
        }
    },
    {
        title: 'Pages',
        commands: {
            'page-list': { title: 'List', cmd: 'scheduler pages list', desc: 'List scheduler pages' },
            'page-show': { title: 'Show', cmd: 'scheduler pages show', desc: 'Show page details', param: { name: 'page-id', placeholder: 'Enter page ID...' } },
            'page-create': {
                title: 'Create',
                cmd: 'scheduler pages create',
                desc: 'Create a scheduler page',
                flags: [
                    { name: 'config-id', type: 'text', label: 'Config ID', placeholder: 'Configuration ID', required: true },
                    { name: 'name', type: 'text', label: 'Name', placeholder: 'Page name', required: true },
                    { name: 'slug', type: 'text', label: 'Slug', placeholder: 'URL slug (optional)' }
                ]
            }
        }
    },
    {
        title: 'Sessions',
        commands: {
            'session-create': {
                title: 'Create',
                cmd: 'scheduler sessions create',
                desc: 'Create a scheduling session',
                flags: [
                    { name: 'config-id', type: 'text', label: 'Config ID', placeholder: 'Configuration ID', required: true },
                    { name: 'ttl', type: 'number', label: 'TTL (min)', placeholder: '30' }
                ]
            },
            'session-show': { title: 'Show', cmd: 'scheduler sessions show', desc: 'Show session details', param: { name: 'session-id', placeholder: 'Enter session ID...' } }
        }
    },
    {
        title: 'Bookings',
        commands: {
            'booking-list': { title: 'List', cmd: 'scheduler bookings list', desc: 'List scheduler bookings' },
            'booking-show': { title: 'Show', cmd: 'scheduler bookings show', desc: 'Show booking details', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } },
            'booking-confirm': { title: 'Confirm', cmd: 'scheduler bookings confirm', desc: 'Confirm a booking', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } },
            'booking-cancel': { title: 'Cancel', cmd: 'scheduler bookings cancel', desc: 'Cancel a booking', param: { name: 'booking-id', placeholder: 'Enter booking ID...' } }
        }
    }
];

const schedulerCommands = {};
schedulerCommandSections.forEach(section => {
    Object.assign(schedulerCommands, section.commands);
});

let currentSchedulerCmd = '';


function showSchedulerCmd(cmd) {
    const data = schedulerCommands[cmd];
    if (!data) return;

    currentSchedulerCmd = cmd;

    document.querySelectorAll('#page-scheduler .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('scheduler-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('scheduler-detail-title').textContent = data.title;
    document.getElementById('scheduler-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('scheduler-detail-desc').textContent = data.desc || '';
    document.getElementById('scheduler-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('scheduler-output').className = 'output-pre';

    showParamInput('scheduler', data.param, data.flags);
}

async function runSchedulerCmd() {
    if (!currentSchedulerCmd) return;


async function runSchedulerCmd() {
    if (!currentSchedulerCmd) return;

    const data = schedulerCommands[currentSchedulerCmd];
    const output = document.getElementById('scheduler-output');
    const btn = document.getElementById('scheduler-run-btn');
    const fullCmd = buildCommand(data.cmd, 'scheduler', data.flags);

    document.getElementById('scheduler-detail-cmd').textContent = 'nylas ' + fullCmd;

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

        updateTimestamp('scheduler');
    } catch (err) {
        output.innerHTML = '<span class="ansi-red">Failed to execute command: ' + esc(err.message) + '</span>';
        output.className = 'output-pre error';
        showToast('Connection error', 'error');
    } finally {
        btn.classList.remove('loading');
        btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Run';
    }
}

function refreshSchedulerCmd() {
    if (currentSchedulerCmd) runSchedulerCmd();
}

function renderSchedulerCommands() {
    renderCommandSections('scheduler-cmd-list', schedulerCommandSections, 'showSchedulerCmd');
}

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

function refreshSchedulerCmd() {
    if (currentSchedulerCmd) runSchedulerCmd();
}

function renderSchedulerCommands() {
    renderCommandSections('scheduler-cmd-list', schedulerCommandSections, 'showSchedulerCmd');
}
