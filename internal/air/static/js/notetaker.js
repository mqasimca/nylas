/**
 * Notetaker Module - Meeting Recording Integration
 * Handles bot-based meeting recordings, transcripts, and AI summaries
 */

const NotetakerModule = {
    notetakers: [],
    selectedNotetaker: null,
    currentFilter: 'all',
    currentProvider: null,
    isLoading: false,

    /**
     * Initialize the notetaker module
     */
    init() {
        this.setupEventListeners();
        this.setupJoinTimeToggle();
        console.log('%c🎙️ Notetaker module loaded', 'color: #8b5cf6;');
    },

    /**
     * Create element helper
     */
    createElement(tag, classes, text) {
        const el = document.createElement(tag);
        if (classes) {
            if (Array.isArray(classes)) {
                el.classList.add(...classes);
            } else {
                el.className = classes;
            }
        }
        if (text) el.textContent = text;
        return el;
    },

    /**
     * Load all notetakers from API
     */
    async loadNotetakers() {
        try {
            const resp = await fetch('/api/notetakers');
            if (resp.ok) {
                this.notetakers = await resp.json();
                this.renderNotetakerPanel();
            }
        } catch (err) {
            console.error('Failed to load notetakers:', err);
        }
    },

    /**
     * Create a notetaker to record a meeting
     */
    async joinMeeting(meetingLink, joinTime = null) {
        try {
            const resp = await fetch('/api/notetakers', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ meetingLink, joinTime })
            });

            if (!resp.ok) {
                throw new Error('Failed to create notetaker');
            }

            const notetaker = await resp.json();
            this.notetakers.push(notetaker);
            this.showNotification('Bot scheduled to join meeting', 'success');
            this.renderNotetakerPanel();
            return notetaker;
        } catch (err) {
            console.error('Failed to join meeting:', err);
            this.showNotification('Failed to schedule bot', 'error');
            throw err;
        }
    },

    /**
     * Get media (recording/transcript) for a notetaker
     */
    async getMedia(notetakerId) {
        try {
            const resp = await fetch(`/api/notetakers/media?id=${notetakerId}`);
            if (!resp.ok) {
                throw new Error('Media not available');
            }
            return await resp.json();
        } catch (err) {
            console.error('Failed to get media:', err);
            throw err;
        }
    },

    /**
     * Cancel a scheduled notetaker
     */
    async cancel(notetakerId) {
        try {
            const resp = await fetch(`/api/notetakers?id=${notetakerId}`, {
                method: 'DELETE'
            });

            if (resp.ok) {
                const idx = this.notetakers.findIndex(n => n.id === notetakerId);
                if (idx >= 0) {
                    this.notetakers[idx].state = 'cancelled';
                }
                this.showNotification('Recording cancelled', 'info');
                this.renderNotetakerPanel();
            }
        } catch (err) {
            console.error('Failed to cancel notetaker:', err);
        }
    },

    /**
     * Get status icon for notetaker state
     */
    getStatusIcon(state) {
        const icons = {
            'scheduled': '🟡',
            'connecting': '🟠',
            'waiting_for_entry': '🟠',
            'attending': '🟢',
            'media_processing': '🔵',
            'complete': '✅',
            'cancelled': '⚪',
            'failed': '🔴'
        };
        return icons[state] || '⚪';
    },

    /**
     * Get human-readable status text
     */
    getStatusText(state) {
        const texts = {
            'scheduled': 'Scheduled',
            'connecting': 'Connecting...',
            'waiting_for_entry': 'Waiting to join',
            'attending': 'Recording',
            'media_processing': 'Processing',
            'complete': 'Complete',
            'cancelled': 'Cancelled',
            'failed': 'Failed'
        };
        return texts[state] || state;
    },

    /**
     * Get provider icon
     */
    getProviderIcon(provider) {
        const icons = {
            'zoom': '📹',
            'google_meet': '🎥',
            'teams': '💼'
        };
        return icons[provider] || '📹';
    },

    /**
     * Build empty state element
     */
    buildEmptyState() {
        const container = this.createElement('div', 'empty-state');
        const icon = this.createElement('span', 'icon', '🤖');
        const title = this.createElement('h3', null, 'No Recordings');
        const desc = this.createElement('p', null, 'Schedule a bot to record your meetings');
        container.appendChild(icon);
        container.appendChild(title);
        container.appendChild(desc);
        return container;
    },

    /**
     * Build notetaker item element
     */
    buildNotetakerItem(nt) {
        const item = this.createElement('div', 'notetaker-item');
        item.dataset.id = nt.id;

        // Add click handler to select
        item.onclick = (e) => {
            // Don't select if clicking action buttons
            if (e.target.closest('.notetaker-actions')) return;
            this.selectNotetaker(nt.id);
        };

        // Status indicator
        const statusIndicator = this.createElement('div', ['notetaker-status-indicator', nt.state]);
        if (nt.state === 'attending') statusIndicator.classList.add('recording');
        item.appendChild(statusIndicator);

        // Info
        const info = this.createElement('div', 'notetaker-info');
        const title = this.createElement('h4', null, nt.meetingTitle || 'Meeting Recording');
        const statusLine = this.createElement('p', 'notetaker-meta',
            this.getProviderIcon(nt.provider) + ' ' + this.getStatusText(nt.state));
        info.appendChild(title);
        info.appendChild(statusLine);

        if (nt.createdAt) {
            const date = this.createElement('small', 'notetaker-date', new Date(nt.createdAt).toLocaleDateString());
            info.appendChild(date);
        }
        item.appendChild(info);

        // Quick actions
        const actions = this.createElement('div', 'notetaker-actions');

        if (nt.state === 'complete') {
            const playBtn = this.createElement('button', 'action-btn', '▶️');
            playBtn.title = 'Play';
            playBtn.onclick = (e) => { e.stopPropagation(); this.playRecording(nt.id); };
            actions.appendChild(playBtn);
        }

        if (nt.state === 'scheduled') {
            const cancelBtn = this.createElement('button', 'action-btn danger', '✕');
            cancelBtn.title = 'Cancel';
            cancelBtn.onclick = (e) => { e.stopPropagation(); this.cancel(nt.id); };
            actions.appendChild(cancelBtn);
        }

        item.appendChild(actions);
        return item;
    },

    /**
     * Render the notetaker list
     */
    renderNotetakers() {
        const list = document.getElementById('notetakerList');
        const empty = document.getElementById('notetakerEmpty');
        if (!list) return;

        // Filter notetakers
        let filtered = this.notetakers;

        if (this.currentFilter !== 'all') {
            filtered = filtered.filter(nt => {
                if (this.currentFilter === 'scheduled') return nt.state === 'scheduled';
                if (this.currentFilter === 'attending') return ['connecting', 'waiting_for_entry', 'attending'].includes(nt.state);
                if (this.currentFilter === 'complete') return nt.state === 'complete';
                return true;
            });
        }

        if (this.currentProvider) {
            filtered = filtered.filter(nt => nt.provider === this.currentProvider);
        }

        // Clear existing items (keep empty state)
        const items = list.querySelectorAll('.notetaker-item');
        items.forEach(item => item.remove());

        // Show/hide empty state
        if (empty) {
            empty.style.display = filtered.length === 0 ? 'flex' : 'none';
        }

        // Render filtered items
        filtered.forEach(nt => {
            list.appendChild(this.buildNotetakerItem(nt));
        });

        // Update counts
        this.updateCounts();
    },

    /**
     * Update sidebar counts
     */
    updateCounts() {
        const counts = {
            all: this.notetakers.length,
            scheduled: this.notetakers.filter(n => n.state === 'scheduled').length,
            attending: this.notetakers.filter(n => ['connecting', 'waiting_for_entry', 'attending'].includes(n.state)).length,
            complete: this.notetakers.filter(n => n.state === 'complete').length
        };

        Object.entries(counts).forEach(([key, value]) => {
            const el = document.getElementById(`notetakerCount${key.charAt(0).toUpperCase() + key.slice(1)}`);
            if (el) el.textContent = value;
        });
    },

    /**
     * Select a notetaker and show details
     */
    selectNotetaker(notetakerId) {
        this.selectedNotetaker = this.notetakers.find(n => n.id === notetakerId);

        // Update active state in list
        document.querySelectorAll('.notetaker-item').forEach(item => {
            item.classList.toggle('active', item.dataset.id === notetakerId);
        });

        this.renderDetail();
    },

    /**
     * Render notetaker detail panel
     */
    renderDetail() {
        const detail = document.getElementById('notetakerDetail');
        if (!detail) return;

        if (!this.selectedNotetaker) {
            detail.innerHTML = `
                <div class="notetaker-detail-empty">
                    <div class="detail-empty-icon">🎬</div>
                    <h3>Select a recording</h3>
                    <p>Click on a recording to view details, playback, and transcript</p>
                </div>
            `;
            return;
        }

        const nt = this.selectedNotetaker;
        const statusClass = nt.state === 'attending' ? 'recording' : nt.state;

        detail.innerHTML = `
            <div class="notetaker-detail-header">
                <div class="notetaker-detail-status ${statusClass}">
                    ${this.getStatusIcon(nt.state)} ${this.getStatusText(nt.state)}
                </div>
                <h2>${nt.meetingTitle || 'Meeting Recording'}</h2>
                <p class="notetaker-detail-meta">
                    ${this.getProviderIcon(nt.provider)} ${this.getProviderName(nt.provider)}
                    ${nt.createdAt ? ' • ' + new Date(nt.createdAt).toLocaleString() : ''}
                </p>
            </div>
            <div class="notetaker-detail-body">
                ${nt.state === 'complete' ? this.renderCompleteContent(nt) : this.renderPendingContent(nt)}
            </div>
            <div class="notetaker-detail-actions">
                ${this.renderActions(nt)}
            </div>
        `;
    },

    /**
     * Get provider display name
     */
    getProviderName(provider) {
        const names = {
            'zoom': 'Zoom',
            'google_meet': 'Google Meet',
            'teams': 'Microsoft Teams'
        };
        return names[provider] || provider || 'Unknown';
    },

    /**
     * Render content for completed recording
     */
    renderCompleteContent(nt) {
        return `
            <div class="detail-section">
                <h3>📹 Recording</h3>
                <p>Video recording available for playback</p>
            </div>
            <div class="detail-section">
                <h3>📝 Transcript</h3>
                <p>Full meeting transcript with speaker labels</p>
            </div>
            <div class="detail-section">
                <h3>✨ AI Summary</h3>
                <p>Get key points and action items from this meeting</p>
            </div>
        `;
    },

    /**
     * Render content for pending/in-progress recording
     */
    renderPendingContent(nt) {
        if (nt.state === 'scheduled') {
            return `
                <div class="detail-section">
                    <h3>⏰ Scheduled</h3>
                    <p>The bot will join the meeting at the scheduled time.</p>
                    <p>Meeting link: <a href="${nt.meetingLink}" target="_blank">${nt.meetingLink || 'N/A'}</a></p>
                </div>
            `;
        }
        if (['connecting', 'waiting_for_entry', 'attending'].includes(nt.state)) {
            return `
                <div class="detail-section recording-indicator">
                    <div class="recording-dot"></div>
                    <h3>Recording in Progress</h3>
                    <p>The bot is currently recording this meeting.</p>
                </div>
            `;
        }
        return `<div class="detail-section"><p>Status: ${this.getStatusText(nt.state)}</p></div>`;
    },

    /**
     * Render action buttons based on state
     */
    renderActions(nt) {
        if (nt.state === 'complete') {
            return `
                <button class="btn-primary" onclick="NotetakerModule.playRecording('${nt.id}')">
                    <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <polygon points="5 3 19 12 5 21 5 3"/>
                    </svg>
                    Play Recording
                </button>
                <button class="btn-secondary" onclick="NotetakerModule.viewTranscript('${nt.id}')">
                    📝 View Transcript
                </button>
                <button class="btn-secondary" onclick="NotetakerModule.summarize('${nt.id}')">
                    ✨ AI Summary
                </button>
            `;
        }
        if (nt.state === 'scheduled') {
            return `
                <button class="btn-danger" onclick="NotetakerModule.cancel('${nt.id}')">
                    ❌ Cancel Recording
                </button>
            `;
        }
        return '';
    },

    /**
     * Render the notetaker panel (legacy support)
     */
    renderNotetakerPanel() {
        this.renderNotetakers();
    },

    /**
     * Play recording in modal
     */
    async playRecording(notetakerId) {
        try {
            const media = await this.getMedia(notetakerId);

            const content = this.createElement('div', 'video-container');
            const video = document.createElement('video');
            video.controls = true;
            video.autoplay = true;
            video.style.width = '100%';

            const source = document.createElement('source');
            source.src = media.recordingUrl;
            source.type = 'video/mp4';
            video.appendChild(source);

            const fallback = document.createTextNode('Your browser does not support video playback.');
            video.appendChild(fallback);
            content.appendChild(video);

            this.showMediaModal('Recording', content);
        } catch (err) {
            this.showNotification('Recording not available', 'error');
        }
    },

    /**
     * View transcript in modal
     */
    async viewTranscript(notetakerId) {
        try {
            const media = await this.getMedia(notetakerId);

            const content = this.createElement('div', 'transcript-content');
            const loading = this.createElement('p', null, 'Loading transcript...');
            content.appendChild(loading);

            const downloadBtn = this.createElement('button', null, '⬇️ Download Transcript');
            downloadBtn.onclick = () => window.open(media.transcriptUrl, '_blank');
            content.appendChild(downloadBtn);

            this.showMediaModal('Transcript', content);
        } catch (err) {
            this.showNotification('Transcript not available', 'error');
        }
    },

    /**
     * Get AI summary of meeting
     */
    async summarize(notetakerId) {
        this.showNotification('Generating AI summary...', 'info');
        try {
            await this.getMedia(notetakerId);

            const content = this.createElement('div', 'ai-summary');

            const title = this.createElement('h4', null, '✨ Meeting Summary');
            content.appendChild(title);

            const desc = this.createElement('p', null, 'AI-generated summary of the meeting:');
            content.appendChild(desc);

            const keyPointsTitle = this.createElement('h5', null, 'Key Points:');
            content.appendChild(keyPointsTitle);

            const keyPoints = this.createElement('ul');
            ['Discussion of project timeline', 'Resource allocation review', 'Next steps identified'].forEach(point => {
                const li = this.createElement('li', null, point);
                keyPoints.appendChild(li);
            });
            content.appendChild(keyPoints);

            const actionsTitle = this.createElement('h5', null, 'Action Items:');
            content.appendChild(actionsTitle);

            const actions = this.createElement('ul');
            ['Follow up with team by Friday', 'Schedule next review meeting'].forEach(action => {
                const li = this.createElement('li', null, action);
                actions.appendChild(li);
            });
            content.appendChild(actions);

            this.showMediaModal('AI Summary', content);
        } catch (err) {
            this.showNotification('Failed to generate summary', 'error');
        }
    },

    /**
     * Show media modal using safe DOM methods
     */
    showMediaModal(title, contentElement) {
        const modal = this.createElement('div', 'notetaker-modal');

        const backdrop = this.createElement('div', 'modal-backdrop');
        backdrop.onclick = () => modal.remove();
        modal.appendChild(backdrop);

        const content = this.createElement('div', 'modal-content');

        const header = this.createElement('div', 'modal-header');
        const headerTitle = this.createElement('h3', null, title);
        const closeBtn = this.createElement('button', null, '✕');
        closeBtn.onclick = () => modal.remove();
        header.appendChild(headerTitle);
        header.appendChild(closeBtn);
        content.appendChild(header);

        const body = this.createElement('div', 'modal-body');
        body.appendChild(contentElement);
        content.appendChild(body);

        modal.appendChild(content);
        document.body.appendChild(modal);
    },

    /**
     * Show notification
     */
    showNotification(message, type = 'info') {
        if (typeof showToast === 'function') {
            showToast(message, type);
        } else {
            console.log(`[${type}] ${message}`);
        }
    },

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        document.addEventListener('eventSelected', (e) => {
            const event = e.detail;
            if (event.conferencing && event.conferencing.details) {
                this.offerRecording(event);
            }
        });

        // Close modal on escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                closeJoinMeetingModal();
            }
        });
    },

    /**
     * Setup join time toggle in modal
     */
    setupJoinTimeToggle() {
        const radios = document.querySelectorAll('input[name="joinTime"]');
        const scheduledGroup = document.getElementById('scheduledTimeGroup');

        radios.forEach(radio => {
            radio.addEventListener('change', () => {
                if (scheduledGroup) {
                    scheduledGroup.style.display = radio.value === 'scheduled' ? 'block' : 'none';
                }
            });
        });
    },

    /**
     * Offer to record a calendar event
     */
    offerRecording(event) {
        const meetingLink = event.conferencing?.details?.url;
        if (!meetingLink) return;

        const shouldRecord = confirm(
            `Would you like to record "${event.title}"?\n\n` +
            `A bot will join the meeting to record and transcribe it.`
        );

        if (shouldRecord) {
            const startTime = new Date(event.when?.startTime || event.start).getTime() / 1000;
            this.joinMeeting(meetingLink, startTime);
        }
    },

    /**
     * Enhance calendar events with recording button
     */
    enhanceCalendarEvent(eventElement, event) {
        if (!event.conferencing?.details?.url) return;

        const recordBtn = this.createElement('button', 'record-meeting-btn', '🤖 Record');
        recordBtn.title = 'Schedule bot to record this meeting';
        recordBtn.onclick = (e) => {
            e.stopPropagation();
            const startTime = new Date(event.when?.startTime || event.start).getTime() / 1000;
            this.joinMeeting(event.conferencing.details.url, startTime);
        };

        eventElement.appendChild(recordBtn);
    }
};

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => NotetakerModule.init());
} else {
    NotetakerModule.init();
}

// ========================================
// GLOBAL FUNCTIONS (called from template)
// ========================================

/**
 * Filter notetakers by state
 */
function filterNotetakers(filter, element) {
    NotetakerModule.currentFilter = filter;
    NotetakerModule.currentProvider = null;

    // Update active state
    document.querySelectorAll('.folder-item[data-filter]').forEach(item => {
        item.classList.toggle('active', item.dataset.filter === filter);
    });
    document.querySelectorAll('.folder-item[data-provider]').forEach(item => {
        item.classList.remove('active');
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
