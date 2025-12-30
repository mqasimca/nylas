/**
 * Notetaker Standalone Functions
 */
function filterNotetakers(filter, element) {
    NotetakerModule.currentFilter = filter;
    // Update tab active state
    document.querySelectorAll('.nt-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.filter === filter);
    });
    NotetakerModule.renderNotetakers();
}

/**
 * Filter notetakers by provider
 */
function filterByProvider(provider, element) {
    NotetakerModule.currentProvider = provider;
    NotetakerModule.currentFilter = 'all';

    // Update active state
    document.querySelectorAll('.folder-item[data-filter]').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelectorAll('.folder-item[data-provider]').forEach(item => {
        item.classList.toggle('active', item.dataset.provider === provider);
    });

    NotetakerModule.renderNotetakers();
}

/**
 * Open join meeting modal
 */
function openJoinMeetingModal() {
    const modal = document.getElementById('joinMeetingModal');
    if (modal) {
        modal.style.display = 'flex';
        const input = document.getElementById('meetingLink');
        if (input) input.focus();
    }
}

/**
 * Close join meeting modal
 */
function closeJoinMeetingModal() {
    const modal = document.getElementById('joinMeetingModal');
    if (modal) {
        modal.style.display = 'none';
        // Reset form
        const form = modal.querySelector('form');
        if (form) form.reset();
        document.getElementById('meetingLink').value = '';
        document.getElementById('botName').value = '';
        document.getElementById('scheduledTimeGroup').style.display = 'none';
    }
}

/**
 * Join meeting (called from modal)
 */
async function joinMeeting() {
    const meetingLink = document.getElementById('meetingLink').value.trim();
    const botName = document.getElementById('botName').value.trim() || 'Nylas Notetaker';
    const joinTimeRadio = document.querySelector('input[name="joinTime"]:checked');
    const scheduledTime = document.getElementById('scheduledTime').value;

    if (!meetingLink) {
        NotetakerModule.showNotification('Please enter a meeting link', 'error');
        return;
    }

    // Validate meeting link
    const validPlatforms = ['zoom.us', 'meet.google.com', 'teams.microsoft.com'];
    const isValid = validPlatforms.some(platform => meetingLink.includes(platform));
    if (!isValid) {
        NotetakerModule.showNotification('Please enter a valid Zoom, Google Meet, or Teams link', 'error');
        return;
    }

    let joinTime = null;
    if (joinTimeRadio && joinTimeRadio.value === 'scheduled' && scheduledTime) {
        joinTime = Math.floor(new Date(scheduledTime).getTime() / 1000);
    }

    try {
        await NotetakerModule.joinMeeting(meetingLink, joinTime);
        closeJoinMeetingModal();
    } catch (err) {
        // Error already shown in joinMeeting
    }
}

/**
 * Refresh notetakers list
 */
function refreshNotetakers() {
    NotetakerModule.loadNotetakers();
}

/**
 * Show help/tips modal
 */
function showNotetakerHelp() {
    const content = NotetakerModule.createElement('div', 'help-content');

    const sections = [
        {
            title: '🎙️ Recording Meetings',
            text: 'Paste a Zoom, Google Meet, or Teams link to have our bot join and record the meeting automatically.'
        },
        {
            title: '📝 Transcripts',
            text: 'Once the meeting ends, a full transcript with speaker labels will be available within minutes.'
        },
        {
            title: '✨ AI Summaries',
            text: 'Get key points, action items, and decisions automatically extracted from your meetings.'
        },
        {
            title: '⏰ Scheduling',
            text: 'Schedule the bot to join future meetings. It will automatically join at the specified time.'
        }
    ];

    sections.forEach(section => {
        const sectionEl = NotetakerModule.createElement('div', 'help-section');
        const title = NotetakerModule.createElement('h4', null, section.title);
        const text = NotetakerModule.createElement('p', null, section.text);
        sectionEl.appendChild(title);
        sectionEl.appendChild(text);
        content.appendChild(sectionEl);
    });

    NotetakerModule.showMediaModal('Notetaker Help', content);
}
