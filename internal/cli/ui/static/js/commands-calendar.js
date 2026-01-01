// =============================================================================
// Calendar Commands
// =============================================================================

const calendarCommandSections = [
    {
        title: 'Calendars',
        commands: {
            'cal-list': { title: 'List', cmd: 'calendar list', desc: 'List all calendars' },
            'cal-show': { title: 'Show', cmd: 'calendar show', desc: 'Show calendar details', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } },
            'cal-create': { title: 'Create', cmd: 'calendar create', desc: 'Create a new calendar' },
            'cal-update': { title: 'Update', cmd: 'calendar update', desc: 'Update a calendar', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } },
            'cal-delete': { title: 'Delete', cmd: 'calendar delete', desc: 'Delete a calendar', param: { name: 'calendar-id', placeholder: 'Enter calendar ID...' } }
        }
    },
    {
        title: 'Events',
        commands: {
            'events-list': {
                title: 'List',
                cmd: 'calendar events list',
                desc: 'List calendar events',
                flags: [
                    { name: 'days', type: 'number', label: 'Days', placeholder: '7', short: 'd' },
                    { name: 'limit', type: 'number', label: 'Limit', placeholder: '10', short: 'n' },
                    { name: 'show-tz', type: 'checkbox', label: 'Show timezone' },
                    { name: 'show-cancelled', type: 'checkbox', label: 'Show cancelled' },
                    { name: 'calendar', type: 'text', label: 'Calendar ID', placeholder: 'primary', short: 'c' },
                    { name: 'timezone', type: 'text', label: 'Timezone', placeholder: 'America/New_York' }
                ]
            },
            'events-show': { title: 'Show', cmd: 'calendar events show', desc: 'Show event details', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-create': {
                title: 'Create',
                cmd: 'calendar events create',
                desc: 'Create a new event',
                flags: [
                    { name: 'title', type: 'text', label: 'Title', placeholder: 'Event title', required: true, short: 't' },
                    { name: 'start', type: 'text', label: 'Start', placeholder: '2024-01-15 14:00', required: true, short: 's' },
                    { name: 'end', type: 'text', label: 'End', placeholder: '2024-01-15 15:00', short: 'e' },
                    { name: 'description', type: 'textarea', label: 'Description', placeholder: 'Event description', short: 'D' },
                    { name: 'location', type: 'text', label: 'Location', placeholder: 'Meeting room or address', short: 'l' },
                    { name: 'participant', type: 'text', label: 'Participants', placeholder: 'email1@example.com', short: 'p' },
                    { name: 'calendar', type: 'text', label: 'Calendar ID', placeholder: 'primary', short: 'c' },
                    { name: 'all-day', type: 'checkbox', label: 'All-Day Event' },
                    { name: 'busy', type: 'checkbox', label: 'Mark as Busy', default: true }
                ]
            },
            'events-update': { title: 'Update', cmd: 'calendar events update', desc: 'Update an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-delete': { title: 'Delete', cmd: 'calendar events delete', desc: 'Delete an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'events-rsvp': { title: 'RSVP', cmd: 'calendar events rsvp', desc: 'RSVP to an event', param: { name: 'event-id', placeholder: 'Enter event ID...' } }
        }
    },
    {
        title: 'Availability',
        commands: {
            'avail-check': { title: 'Check', cmd: 'calendar availability check', desc: 'Check free/busy status' },
            'avail-find': { title: 'Find', cmd: 'calendar availability find', desc: 'Find available meeting times' }
        }
    },
    {
        title: 'AI & Scheduling',
        commands: {
            'schedule': { title: 'Schedule', cmd: 'calendar schedule', desc: 'Schedule meetings with AI' },
            'find-time': { title: 'Find Time', cmd: 'calendar find-time', desc: 'Find optimal meeting times' },
            'ai-analyze': { title: 'AI Analyze', cmd: 'calendar ai analyze', desc: 'AI meeting analysis' },
            'ai-focus': { title: 'AI Focus', cmd: 'calendar ai focus-time', desc: 'AI focus time protection' },
            'ai-conflicts': { title: 'AI Conflicts', cmd: 'calendar ai conflicts', desc: 'AI conflict detection' }
        }
    },
    {
        title: 'Recurring',
        commands: {
            'recurring-list': { title: 'List', cmd: 'calendar recurring list', desc: 'List recurring instances', param: { name: 'event-id', placeholder: 'Enter master event ID...' } },
            'recurring-update': { title: 'Update', cmd: 'calendar recurring update', desc: 'Update instance', param: { name: 'event-id', placeholder: 'Enter event ID...' } },
            'recurring-delete': { title: 'Delete', cmd: 'calendar recurring delete', desc: 'Delete instance', param: { name: 'event-id', placeholder: 'Enter event ID...' } }
        }
    },
    {
        title: 'Virtual Calendars',
        commands: {
            'virtual-list': { title: 'List', cmd: 'calendar virtual list', desc: 'List virtual calendars' },
            'virtual-show': { title: 'Show', cmd: 'calendar virtual show', desc: 'Show virtual calendar', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } },
            'virtual-create': {
                title: 'Create',
                cmd: 'calendar virtual create',
                desc: 'Create virtual calendar',
                flags: [{ name: 'email', type: 'text', label: 'Email', placeholder: 'Virtual calendar email identifier', required: true }]
            },
            'virtual-delete': { title: 'Delete', cmd: 'calendar virtual delete', desc: 'Delete virtual calendar', param: { name: 'grant-id', placeholder: 'Enter grant ID...' } }
        }
    }
];

const calendarCommands = {};
calendarCommandSections.forEach(section => {
    Object.assign(calendarCommands, section.commands);
});

let currentCalendarCmd = '';

function showCalendarCmd(cmd) {
    const data = calendarCommands[cmd];
    if (!data) return;

    currentCalendarCmd = cmd;

    document.querySelectorAll('#page-calendar .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('calendar-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('calendar-detail-title').textContent = data.title;
    document.getElementById('calendar-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('calendar-detail-desc').textContent = data.desc || '';
    document.getElementById('calendar-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('calendar-output').className = 'output-pre';

    showParamInput('calendar', data.param, data.flags);
}

async function runCalendarCmd() {
    if (!currentCalendarCmd) return;

    const data = calendarCommands[currentCalendarCmd];
    const output = document.getElementById('calendar-output');
    const btn = document.getElementById('calendar-run-btn');
    const fullCmd = buildCommand(data.cmd, 'calendar', data.flags);

    document.getElementById('calendar-detail-cmd').textContent = 'nylas ' + fullCmd;

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

            if (result.output) {
                let cached = false;
                if (currentCalendarCmd === 'cal-list') {
                    const ids = parseCalendarIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedCalendarIds = ids;
                        showToast('Cached ' + ids.length + ' calendar IDs for quick access', 'info');
                        cached = true;
                    }
                } else if (currentCalendarCmd === 'events-list') {
                    const ids = parseEventIdsFromOutput(result.output);
                    if (ids.length > 0) {
                        cachedEventIds = ids;
                        showToast('Cached ' + ids.length + ' event IDs for quick access', 'info');
                        cached = true;
                    }
                }
                if (cached) updateCacheCountBadge();
            }
        }

        updateTimestamp('calendar');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshCalendarCmd() {
    if (currentCalendarCmd) runCalendarCmd();
}

function renderCalendarCommands() {
    renderCommandSections('calendar-cmd-list', calendarCommandSections, 'showCalendarCmd');
}
