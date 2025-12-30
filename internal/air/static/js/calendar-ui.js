/**
 * Calendar UI - Rendering
 */
Object.assign(CalendarManager, {
renderCalendarList() {
    const container = document.getElementById('calendarsList');
    if (!container) {
        console.error('Calendar list container not found');
        return;
    }

    // Clear skeleton loaders
    container.innerHTML = '';

    if (this.calendars.length === 0) {
        container.innerHTML = '<div class="folder-item"><span class="text-muted">No calendars found</span></div>';
        return;
    }

    this.calendars.forEach(cal => {
        const div = document.createElement('div');
        div.className = 'folder-item';
        div.setAttribute('data-calendar-id', cal.id);
        div.innerHTML = `
            <span class="label-dot" style="background: ${cal.hex_color || '#4285f4'}"></span>
            <span>${this.escapeHtml(cal.name)}</span>
        `;
        container.appendChild(div);
    });
},

renderEvents() {
    const eventsContainer = document.querySelector('.events-list');
    if (!eventsContainer) return;

    // Update header
    const dateEl = document.querySelector('.events-date');
    if (dateEl) {
        dateEl.textContent = this.currentDate.toLocaleDateString('en-US', {
            weekday: 'short',
            month: 'short',
            day: 'numeric'
        });
    }

    if (this.events.length === 0) {
        eventsContainer.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">📅</div>
                <div class="empty-message">No events</div>
            </div>
        `;
        return;
    }

    // Sort events by start time
    const sortedEvents = [...this.events].sort((a, b) => a.start_time - b.start_time);

    eventsContainer.innerHTML = sortedEvents.map(event => this.renderEventCard(event)).join('');
},

renderEventCard(event) {
    const startTime = this.formatEventTime(event.start_time);
    const endTime = this.formatEventTime(event.end_time);
    const isFocusTime = event.title?.toLowerCase().includes('focus');
    const hasConferencing = event.conferencing && event.conferencing.url;

    const participantsHtml = event.participants && event.participants.length > 0
        ? `<div class="event-attendees">
            ${event.participants.slice(0, 3).map(p => `
                <div class="attendee-avatar" style="background: var(--gradient-${Math.floor(Math.random() * 5) + 1})" title="${this.escapeHtml(p.name || p.email)}">
                    ${(p.name || p.email || '?')[0].toUpperCase()}
                </div>
            `).join('')}
            ${event.participants.length > 3 ? `<div class="attendee-more">+${event.participants.length - 3}</div>` : ''}
           </div>`
        : '';

    return `
        <div class="event-card${isFocusTime ? ' focus-time' : ''}" data-event-id="${event.id}">
            <div class="event-time">${event.is_all_day ? 'All Day' : `${startTime} - ${endTime}`}</div>
            <div class="event-title">${isFocusTime ? '🧘 ' : ''}${this.escapeHtml(event.title || '(No Title)')}</div>
            ${event.description ? `<div class="event-desc">${this.escapeHtml(event.description.substring(0, 100))}</div>` : ''}
            ${event.location ? `<div class="event-location">📍 ${this.escapeHtml(event.location)}</div>` : ''}
            ${participantsHtml}
            ${hasConferencing ? `
                <div class="event-meta">
                    <a href="${event.conferencing.url}" target="_blank" class="event-tag">
                        📹 ${event.conferencing.provider || 'Video Call'}
                    </a>
                </div>
            ` : ''}
        </div>
    `;
},

formatEventTime(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit',
        hour12: true
    });
},

escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
},
});
