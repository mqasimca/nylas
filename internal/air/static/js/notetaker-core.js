/**
 * Notetaker Core - Initialization and event handling
 */
const NotetakerModule = {
    notetakers: [],
    selectedNotetaker: null,
    currentFilter: 'past',
    currentProvider: null,
    isLoading: false
};
init() {
    this.setupEventListeners();
    this.setupJoinTimeToggle();
    console.log('%c🎙️ Notetaker module loaded', 'color: #8b5cf6;');
},

/**
 * Get notetaker sources from global settings
 * Falls back to default if not configured
 */
getSources() {
    if (typeof settingsState !== 'undefined' && settingsState.notetakerSources && settingsState.notetakerSources.length > 0) {
        return settingsState.notetakerSources;
    }
    // No default source - user must configure in Settings
    return [];
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
 * Uses notetaker sources from global settings
 */
async loadNotetakers() {
    try {
        const sources = this.getSources();
        const params = new URLSearchParams();
        // Pass sources as JSON array
        params.set('sources', JSON.stringify(sources));
        const url = '/api/notetakers?' + params.toString();
        const resp = await fetch(url);
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
 * Filter notetakers by past/upcoming
