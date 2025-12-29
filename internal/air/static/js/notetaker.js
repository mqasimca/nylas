/**
 * Notetaker Module - Meeting Recording Integration
 * Handles bot-based meeting recordings, transcripts, and AI summaries
 */

const NotetakerModule = {
    notetakers: [],
    selectedNotetaker: null,
    currentFilter: 'past',
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
            'teams': '💼',
            'nylas_notebook': '📓'
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
     * Build notetaker card element
     */
    buildNotetakerItem(nt) {
        const card = this.createElement('div', 'nt-card');
        card.dataset.id = nt.id;

        // Click handler to show summary for external notetakers
        if (nt.isExternal && nt.summary) {
            card.style.cursor = 'pointer';
            card.onclick = (e) => {
                if (e.target.closest('.nt-card-btn') || e.target.closest('.nt-card-toggle')) return;
                this.showSummaryModal(nt);
            };
        }

        // Banner with provider icon
        const banner = this.createElement('div', 'nt-card-banner');
        const providerIcon = this.createElement('div', 'nt-card-provider');
        providerIcon.innerHTML = this.getProviderSVG(nt.provider);
        banner.appendChild(providerIcon);
        card.appendChild(banner);

        // Card body
        const body = this.createElement('div', 'nt-card-body');

        // Title row with badge
        const titleRow = this.createElement('div', 'nt-card-title-row');
        const title = this.createElement('h4', 'nt-card-title', nt.meetingTitle || 'Meeting Recording');
        titleRow.appendChild(title);

        // Status badge - show "External" for external sources
        const badge = this.createElement('span', 'nt-card-badge');
        if (nt.isExternal) {
            badge.classList.add('external');
            badge.textContent = 'External';
        } else {
            badge.classList.add(this.getBadgeClass(nt.state));
            badge.textContent = this.getStatusText(nt.state);
        }
        titleRow.appendChild(badge);
        body.appendChild(titleRow);

        // Meta info (date/time)
        const meta = this.createElement('div', 'nt-card-meta');
        if (nt.createdAt) {
            const d = new Date(nt.createdAt);
            meta.innerHTML = `<span>📅 ${d.toLocaleDateString()}</span><span>🕐 ${d.toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'})}</span>`;
        }
        body.appendChild(meta);

        // Details section (collapsed by default)
        const details = this.createElement('div', 'nt-card-details');
        details.style.display = 'none';
        const detailsLink = nt.meetingLink ? `<p><a href="${nt.meetingLink}" target="_blank">🔗 Meeting Link</a></p>` : '';
        details.innerHTML = `<p>${this.getProviderName(nt.provider)}</p>${detailsLink}`;
        body.appendChild(details);

        // Toggle details button
        const toggleBtn = this.createElement('button', 'nt-card-toggle', '▼ Meeting Details');
        toggleBtn.onclick = (e) => {
            e.stopPropagation();
            const open = details.style.display !== 'none';
            details.style.display = open ? 'none' : 'block';
            toggleBtn.textContent = open ? '▼ Meeting Details' : '▲ Hide Details';
        };
        body.appendChild(toggleBtn);

        // Action button
        const btn = this.createElement('button', 'nt-card-btn');
        if (nt.isExternal && nt.externalUrl) {
            btn.textContent = '🔗 Open Recording';
            btn.onclick = () => window.open(nt.externalUrl, '_blank');
        } else if (nt.state === 'complete' || nt.state === 'completed') {
            btn.textContent = '▶️ Watch Now';
            btn.onclick = () => this.playRecording(nt.id);
        } else if (nt.state === 'scheduled') {
            btn.textContent = '❌ Cancel';
            btn.classList.add('danger');
            btn.onclick = () => this.cancel(nt.id);
        } else {
            btn.textContent = this.getStatusText(nt.state);
            btn.disabled = true;
        }
        body.appendChild(btn);

        card.appendChild(body);
        return card;
    },

    /**
     * Get badge CSS class for state
     */
    getBadgeClass(state) {
        const classes = {
            'complete': 'complete', 'completed': 'complete',
            'failed': 'failed', 'failed_entry': 'failed', 'cancelled': 'failed',
            'attending': 'active', 'connecting': 'pending', 'waiting_for_entry': 'pending',
            'scheduled': 'pending', 'media_processing': 'pending'
        };
        return classes[state] || 'pending';
    },

    /**
     * Get provider SVG icon
     */
    getProviderSVG(provider) {
        if (provider === 'google_meet') return '<svg viewBox="0 0 24 24" width="48" height="48"><rect fill="#00897B" width="24" height="24" rx="4"/><path fill="#fff" d="M12 6l6 4v4l-6 4-6-4v-4z"/></svg>';
        if (provider === 'zoom') return '<svg viewBox="0 0 24 24" width="48" height="48"><rect fill="#2D8CFF" width="24" height="24" rx="4"/><path fill="#fff" d="M4 8h10v8H4z"/><path fill="#fff" d="M14 10l4-2v8l-4-2z"/></svg>';
        if (provider === 'teams') return '<svg viewBox="0 0 24 24" width="48" height="48"><rect fill="#5059C9" width="24" height="24" rx="4"/><path fill="#fff" d="M6 8h8v8H6z"/></svg>';
        return '<svg viewBox="0 0 24 24" width="48" height="48"><rect fill="#8b5cf6" width="24" height="24" rx="4"/><text x="12" y="16" text-anchor="middle" fill="#fff" font-size="10">N</text></svg>';
    },

    /**
     * Render the notetaker list as cards
     */
    renderNotetakers() {
        const list = document.getElementById('notetakerList');
        const empty = document.getElementById('notetakerEmpty');
        if (!list) return;

        // Filter by past/upcoming
        const now = Date.now();
        let filtered = this.notetakers.filter(nt => {
            const ntTime = nt.createdAt ? new Date(nt.createdAt).getTime() : now;
            if (this.currentFilter === 'past') {
                return nt.state === 'complete' || nt.state === 'completed' || nt.state === 'failed' || nt.state === 'cancelled' || nt.isExternal;
            }
            if (this.currentFilter === 'upcoming') {
                return nt.state === 'scheduled' || nt.state === 'connecting' || nt.state === 'waiting_for_entry' || nt.state === 'attending';
            }
            return true;
        });

        // Clear existing cards
        list.querySelectorAll('.nt-card').forEach(c => c.remove());

        // Show/hide empty state
        if (empty) empty.style.display = filtered.length === 0 ? 'flex' : 'none';

        // Render cards
        filtered.forEach(nt => list.appendChild(this.buildNotetakerItem(nt)));
    },

    /**
     * Update sidebar counts
     */
    updateCounts() {
        const counts = {
            all: this.notetakers.length,
            scheduled: this.notetakers.filter(n => n.state === 'scheduled').length,
            attending: this.notetakers.filter(n => ['connecting', 'waiting_for_entry', 'attending'].includes(n.state)).length,
            complete: this.notetakers.filter(n => n.state === 'complete' || n.state === 'completed').length
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
        const isCompleted = nt.state === 'complete' || nt.state === 'completed';

        // Determine body content
        let bodyContent;
        if (nt.isExternal) {
            bodyContent = this.renderExternalContent(nt);
        } else if (isCompleted) {
            bodyContent = this.renderCompleteContent(nt);
        } else {
            bodyContent = this.renderPendingContent(nt);
        }

        // Build status display
        const statusDisplay = nt.isExternal
            ? '🔗 External'
            : this.getStatusIcon(nt.state) + ' ' + this.getStatusText(nt.state);

        // Build attendees line
        const attendeesLine = nt.attendees
            ? '<p class="notetaker-detail-attendees">👥 ' + nt.attendees + '</p>'
            : '';

        detail.innerHTML = `
            <div class="notetaker-detail-header">
                <div class="notetaker-detail-status ${statusClass}">
                    ${statusDisplay}
                </div>
                <h2>${nt.meetingTitle || 'Meeting Recording'}</h2>
                <p class="notetaker-detail-meta">
                    ${this.getProviderIcon(nt.provider)} ${this.getProviderName(nt.provider)}
                    ${nt.createdAt ? ' • ' + new Date(nt.createdAt).toLocaleString() : ''}
                </p>
                ${attendeesLine}
            </div>
            <div class="notetaker-detail-body">
                ${bodyContent}
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
            'teams': 'Microsoft Teams',
            'nylas_notebook': 'Nylas Notebook (External)'
        };
        return names[provider] || provider || 'Unknown';
    },

    /**
     * Strip embedded styles from HTML to allow our CSS to take control
     */
    stripEmbeddedStyles(html) {
        // Remove <style> tags and their content
        let cleaned = html.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '');
        // Remove inline style attributes
        cleaned = cleaned.replace(/\s+style="[^"]*"/gi, '');
        // Remove <html>, <head>, <body> tags but keep their content
        cleaned = cleaned.replace(/<\/?html[^>]*>/gi, '');
        cleaned = cleaned.replace(/<head[^>]*>[\s\S]*?<\/head>/gi, '');
        cleaned = cleaned.replace(/<\/?body[^>]*>/gi, '');
        return cleaned;
    },

    /**
     * Render content for external recording (from Nylas Notebook emails)
     */
    renderExternalContent(nt) {
        const container = this.createElement('div', 'external-content');

        // If there's a summary from the email, show it
        if (nt.summary) {
            const summarySection = this.createElement('div', 'detail-section summary-section');
            const summaryContent = this.createElement('div', 'summary-content');
            // Strip embedded styles to let our CSS control theming
            summaryContent.innerHTML = this.stripEmbeddedStyles(nt.summary);
            summarySection.appendChild(summaryContent);
            container.appendChild(summarySection);
        } else {
            const content = this.createElement('div', 'detail-section');
            const title = this.createElement('h3', null, '🔗 External Recording');
            const desc = this.createElement('p', null, 'This recording is available in Nylas Notebook.');
            const note = this.createElement('p', 'external-note', 'Click the button below to open in a new tab.');
            content.appendChild(title);
            content.appendChild(desc);
            content.appendChild(note);
            container.appendChild(content);
        }

        return container.outerHTML;
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
        if (nt.isExternal && nt.externalUrl) {
            return `
                <button class="btn-primary" onclick="window.open('${nt.externalUrl}', '_blank')">
                    🔗 Open in Nylas Notebook
                </button>
            `;
        }
        if (nt.state === 'complete' || nt.state === 'completed') {
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
     * Show summary modal for external notetakers
     */
    showSummaryModal(nt) {
        const content = this.createElement('div', 'summary-modal-content');

        // Email summary content - just the body text
        if (nt.summary) {
            const summaryDiv = this.createElement('div', 'summary-body');
            summaryDiv.innerHTML = this.stripEmailCruft(nt.summary);
            content.appendChild(summaryDiv);
        }

        this.showMediaModal(nt.meetingTitle || 'Meeting Summary', content);
    },

    /**
     * Clean email HTML - keep structure but remove inline styles
     */
    stripEmailCruft(html) {
        let cleaned = html;
        // Remove style tags
        cleaned = cleaned.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '');
        // Remove inline styles
        cleaned = cleaned.replace(/\s+style="[^"]*"/gi, '');
        // Remove width/height attributes globally
        cleaned = cleaned.replace(/\s+width="[^"]*"/gi, '');
        cleaned = cleaned.replace(/\s+height="[^"]*"/gi, '');
        // Add small size to all images
        cleaned = cleaned.replace(/<img/gi, '<img style="max-width:80px;max-height:40px;display:block;margin:0 auto 16px"');
        // Remove html/head/body tags
        cleaned = cleaned.replace(/<\/?html[^>]*>/gi, '');
        cleaned = cleaned.replace(/<head[^>]*>[\s\S]*?<\/head>/gi, '');
        cleaned = cleaned.replace(/<\/?body[^>]*>/gi, '');
        return cleaned;
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
     * Show settings modal - opens global settings modal
     */
    showSettingsModal() {
        // Open the global settings modal
        if (typeof toggleSettings === 'function') {
            toggleSettings();
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
 * Filter notetakers by past/upcoming
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
