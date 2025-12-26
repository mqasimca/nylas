// ====================================
// CALENDAR MODULE
// Handles calendar view and events
// ====================================

const CalendarManager = {
    calendars: [],
    events: [],
    currentDate: new Date(),
    currentView: 'week', // 'today', 'week', 'month', 'agenda'
    selectedCalendarIds: [], // Empty means all calendars
    isLoading: false,
    isInitialized: false, // Track if data has been loaded

    async init() {
        // Only initialize once
        if (this.isInitialized) {
            return;
        }
        this.isInitialized = true;
        // Set up event listeners first (UI is ready immediately)
        this.setupEventListeners();

        // Load calendars first, then events (sequential to avoid rate limits)
        try {
            await this.loadCalendars();
        } catch (error) {
            console.error('Failed to load calendars:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load calendars. Will retry...');
            }
            // Retry after delay
            setTimeout(() => this.loadCalendars(), 3000);
        }

        try {
            await this.loadEvents();
        } catch (error) {
            console.error('Failed to load events:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load events. Will retry...');
            }
            // Retry after delay
            setTimeout(() => this.loadEvents(), 3000);
        }

        // Load conflicts after events
        try {
            await this.loadConflicts();
        } catch (error) {
            console.error('Failed to load conflicts:', error);
        }

        console.log('%c📅 Calendar module loaded', 'color: #22c55e;');
    },

    setupEventListeners() {
        // View selection
        const calendarSidebar = document.querySelector('#calendarView .sidebar');
        if (calendarSidebar) {
            calendarSidebar.addEventListener('click', (e) => {
                const folderItem = e.target.closest('.folder-item');
                if (folderItem) {
                    const text = folderItem.textContent.trim().toLowerCase();
                    if (text.includes('today')) this.setView('today');
                    else if (text.includes('week')) this.setView('week');
                    else if (text.includes('month')) this.setView('month');
                    else if (text.includes('agenda')) this.setView('agenda');
                }
            });
        }

        // Navigation buttons
        const navBtns = document.querySelectorAll('.calendar-nav-btn');
        if (navBtns.length >= 2) {
            navBtns[0].addEventListener('click', () => this.navigate(-1));
            navBtns[1].addEventListener('click', () => this.navigate(1));
        }

        // Today button
        const todayBtn = document.querySelector('.today-btn');
        if (todayBtn) {
            todayBtn.addEventListener('click', () => this.goToToday());
        }
    },

    async loadCalendars() {
        try {
            const data = await AirAPI.getCalendars();
            this.calendars = data.calendars || [];
            this.renderCalendarList();
        } catch (error) {
            console.error('Failed to load calendars:', error);
        }
    },

    async loadEvents() {
        if (this.isLoading) return;
        this.isLoading = true;

        try {
            const { start, end } = this.getDateRange();

            const data = await AirAPI.getEvents({
                start: Math.floor(start.getTime() / 1000),
                end: Math.floor(end.getTime() / 1000)
            });

            this.events = data.events || [];
            this.renderEvents();
        } catch (error) {
            console.error('Failed to load events:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load events');
            }
        } finally {
            this.isLoading = false;
        }
    },

    getDateRange() {
        const now = this.currentDate;
        let start, end;

        switch (this.currentView) {
            case 'today':
                start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
                end = new Date(start);
                end.setDate(end.getDate() + 1);
                break;
            case 'week':
                start = new Date(now);
                start.setDate(start.getDate() - start.getDay()); // Start of week (Sunday)
                start.setHours(0, 0, 0, 0);
                end = new Date(start);
                end.setDate(end.getDate() + 7);
                break;
            case 'month':
                start = new Date(now.getFullYear(), now.getMonth(), 1);
                end = new Date(now.getFullYear(), now.getMonth() + 1, 1);
                break;
            case 'agenda':
            default:
                start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
                end = new Date(start);
                end.setDate(end.getDate() + 14); // 2 weeks
                break;
        }

        return { start, end };
    },

    navigate(direction) {
        switch (this.currentView) {
            case 'today':
                this.currentDate.setDate(this.currentDate.getDate() + direction);
                break;
            case 'week':
                this.currentDate.setDate(this.currentDate.getDate() + (direction * 7));
                break;
            case 'month':
                this.currentDate.setMonth(this.currentDate.getMonth() + direction);
                break;
            case 'agenda':
                this.currentDate.setDate(this.currentDate.getDate() + (direction * 14));
                break;
        }

        this.updateTitle();
        this.loadEvents();
    },

    goToToday() {
        this.currentDate = new Date();
        this.updateTitle();
        this.loadEvents();
    },

    setView(view) {
        this.currentView = view;

        // Update sidebar active state
        const folderItems = document.querySelectorAll('#calendarView .sidebar .folder-item');
        folderItems.forEach(item => {
            const text = item.textContent.trim().toLowerCase();
            const isActive = text.includes(view) ||
                (view === 'today' && text.includes('today')) ||
                (view === 'week' && text.includes('week')) ||
                (view === 'month' && text.includes('month')) ||
                (view === 'agenda' && text.includes('agenda'));
            item.classList.toggle('active', isActive);
        });

        this.updateTitle();
        this.loadEvents();
    },

    updateTitle() {
        const titleEl = document.querySelector('.calendar-title');
        if (!titleEl) return;

        const options = { month: 'long', year: 'numeric' };
        if (this.currentView === 'today') {
            options.day = 'numeric';
        }

        titleEl.textContent = this.currentDate.toLocaleDateString('en-US', options);
    },

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

    // ====================================
    // CONFLICT DETECTION
    // ====================================

    conflicts: [],
    conflictsLoaded: false,

    async loadConflicts() {
        try {
            const { start, end } = this.getDateRange();
            const data = await AirAPI.getConflicts({
                start_time: Math.floor(start.getTime() / 1000),
                end_time: Math.floor(end.getTime() / 1000)
            });
            this.conflicts = data.conflicts || [];
            this.conflictsLoaded = true;
            this.renderConflicts();
            return this.conflicts;
        } catch (error) {
            console.error('Failed to load conflicts:', error);
            return [];
        }
    },

    renderConflicts() {
        const container = document.getElementById('conflictsPanel');
        if (!container) return;

        if (this.conflicts.length === 0) {
            container.innerHTML = `
                <div class="conflicts-header">
                    <span class="conflicts-icon">✅</span>
                    <span>No scheduling conflicts</span>
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="conflicts-header warning">
                <span class="conflicts-icon">⚠️</span>
                <span>${this.conflicts.length} conflict${this.conflicts.length > 1 ? 's' : ''} detected</span>
            </div>
            <div class="conflicts-list">
                ${this.conflicts.map(c => this.renderConflictCard(c)).join('')}
            </div>
        `;
    },

    renderConflictCard(conflict) {
        const event1 = conflict.event1;
        const event2 = conflict.event2;
        return `
            <div class="conflict-card">
                <div class="conflict-event">
                    <span class="conflict-time">${this.formatEventTime(event1.start_time)}</span>
                    <span class="conflict-title">${this.escapeHtml(event1.title || '(No Title)')}</span>
                </div>
                <div class="conflict-overlap">↔️ overlaps with</div>
                <div class="conflict-event">
                    <span class="conflict-time">${this.formatEventTime(event2.start_time)}</span>
                    <span class="conflict-title">${this.escapeHtml(event2.title || '(No Title)')}</span>
                </div>
            </div>
        `;
    },

    // ====================================
    // AVAILABILITY DISPLAY
    // ====================================

    availabilitySlots: [],

    async loadAvailability(options = {}) {
        try {
            const now = Math.floor(Date.now() / 1000);
            const data = await AirAPI.getAvailability({
                start_time: options.start_time || now,
                end_time: options.end_time || (now + 7 * 24 * 60 * 60),
                duration_minutes: options.duration_minutes || 30,
                participants: options.participants || [],
                interval_minutes: options.interval_minutes || 15
            });
            this.availabilitySlots = data.slots || [];
            return this.availabilitySlots;
        } catch (error) {
            console.error('Failed to load availability:', error);
            return [];
        }
    },

    renderAvailabilitySlots(container) {
        if (!container) return;

        if (this.availabilitySlots.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📅</div>
                    <div class="empty-message">No available slots found</div>
                </div>
            `;
            return;
        }

        // Group slots by day
        const slotsByDay = {};
        this.availabilitySlots.forEach(slot => {
            const date = new Date(slot.start_time * 1000);
            const dayKey = date.toDateString();
            if (!slotsByDay[dayKey]) {
                slotsByDay[dayKey] = [];
            }
            slotsByDay[dayKey].push(slot);
        });

        container.innerHTML = Object.entries(slotsByDay).map(([day, slots]) => `
            <div class="availability-day">
                <div class="availability-day-header">${day}</div>
                <div class="availability-slots">
                    ${slots.slice(0, 6).map(slot => `
                        <button class="availability-slot" data-start="${slot.start_time}" data-end="${slot.end_time}">
                            ${this.formatEventTime(slot.start_time)} - ${this.formatEventTime(slot.end_time)}
                        </button>
                    `).join('')}
                    ${slots.length > 6 ? `<span class="more-slots">+${slots.length - 6} more</span>` : ''}
                </div>
            </div>
        `).join('');
    }
};

// ====================================
// FIND TIME MODAL
// ====================================

const FindTimeModal = {
    isOpen: false,
    selectedSlot: null,

    open() {
        const overlay = document.getElementById('findTimeModalOverlay');
        if (!overlay) {
            this.createModal();
        }
        document.getElementById('findTimeModalOverlay').classList.remove('hidden');
        this.isOpen = true;
        this.reset();
    },

    close() {
        const overlay = document.getElementById('findTimeModalOverlay');
        if (overlay) {
            overlay.classList.add('hidden');
        }
        this.isOpen = false;
    },

    reset() {
        this.selectedSlot = null;
        const participantsInput = document.getElementById('findTimeParticipants');
        const durationSelect = document.getElementById('findTimeDuration');
        const resultsContainer = document.getElementById('findTimeResults');

        if (participantsInput) participantsInput.value = '';
        if (durationSelect) durationSelect.value = '30';
        if (resultsContainer) resultsContainer.innerHTML = `
            <div class="find-time-hint">
                Enter participant emails and click "Find Available Times"
            </div>
        `;
    },

    async search() {
        const participantsInput = document.getElementById('findTimeParticipants');
        const durationSelect = document.getElementById('findTimeDuration');
        const resultsContainer = document.getElementById('findTimeResults');
        const searchBtn = document.getElementById('findTimeSearchBtn');

        if (!participantsInput || !resultsContainer) return;

        const participants = participantsInput.value
            .split(',')
            .map(e => e.trim())
            .filter(e => e.length > 0);

        if (participants.length === 0) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Please enter at least one participant email');
            }
            return;
        }

        // Show loading state
        if (searchBtn) {
            searchBtn.disabled = true;
            searchBtn.textContent = 'Searching...';
        }
        resultsContainer.innerHTML = '<div class="loading">Finding available times...</div>';

        try {
            const duration = parseInt(durationSelect?.value || '30', 10);
            const slots = await AirAPI.findTime(participants, duration);

            CalendarManager.availabilitySlots = slots.slots || [];
            CalendarManager.renderAvailabilitySlots(resultsContainer);

            // Add click handlers for slot selection
            resultsContainer.querySelectorAll('.availability-slot').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.selectSlot(btn);
                });
            });

        } catch (error) {
            console.error('Find time error:', error);
            resultsContainer.innerHTML = `
                <div class="error-state">
                    <div class="error-icon">❌</div>
                    <div class="error-message">${error.message || 'Failed to find available times'}</div>
                </div>
            `;
        } finally {
            if (searchBtn) {
                searchBtn.disabled = false;
                searchBtn.textContent = 'Find Available Times';
            }
        }
    },

    selectSlot(btn) {
        // Remove previous selection
        document.querySelectorAll('.availability-slot.selected').forEach(el => {
            el.classList.remove('selected');
        });

        btn.classList.add('selected');
        this.selectedSlot = {
            start_time: parseInt(btn.dataset.start, 10),
            end_time: parseInt(btn.dataset.end, 10)
        };

        // Enable create event button
        const createBtn = document.getElementById('findTimeCreateBtn');
        if (createBtn) {
            createBtn.disabled = false;
        }
    },

    createEvent() {
        if (!this.selectedSlot) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Please select a time slot first');
            }
            return;
        }

        const participantsInput = document.getElementById('findTimeParticipants');
        const participants = participantsInput?.value
            .split(',')
            .map(e => e.trim())
            .filter(e => e.length > 0) || [];

        // Close find time modal
        this.close();

        // Open event modal with pre-filled data
        const eventData = {
            start_time: this.selectedSlot.start_time,
            end_time: this.selectedSlot.end_time,
            participants: participants.map(email => ({ email }))
        };

        EventModal.open(eventData);
    },

    createModal() {
        const overlay = document.createElement('div');
        overlay.id = 'findTimeModalOverlay';
        overlay.className = 'modal-overlay hidden';
        overlay.innerHTML = `
            <div class="modal find-time-modal">
                <div class="modal-header">
                    <h2>Find Available Time</h2>
                    <button class="modal-close" onclick="FindTimeModal.close()">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="form-group">
                        <label for="findTimeParticipants">Participants (comma-separated emails)</label>
                        <input type="text" id="findTimeParticipants"
                               placeholder="alice@example.com, bob@example.com">
                    </div>
                    <div class="form-group">
                        <label for="findTimeDuration">Meeting Duration</label>
                        <select id="findTimeDuration">
                            <option value="15">15 minutes</option>
                            <option value="30" selected>30 minutes</option>
                            <option value="45">45 minutes</option>
                            <option value="60">1 hour</option>
                            <option value="90">1.5 hours</option>
                            <option value="120">2 hours</option>
                        </select>
                    </div>
                    <button id="findTimeSearchBtn" class="btn btn-primary" onclick="FindTimeModal.search()">
                        Find Available Times
                    </button>
                    <div id="findTimeResults" class="find-time-results">
                        <div class="find-time-hint">
                            Enter participant emails and click "Find Available Times"
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="FindTimeModal.close()">Cancel</button>
                    <button id="findTimeCreateBtn" class="btn btn-primary" onclick="FindTimeModal.createEvent()" disabled>
                        Create Event
                    </button>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);

        // Close on backdrop click
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                this.close();
            }
        });
    }
};

// Global function for Find Time
function openFindTimeModal() {
    FindTimeModal.open();
}

// ====================================
// EVENT MODAL MANAGER
// Handles create/edit/delete events
// ====================================

const EventModal = {
    isOpen: false,
    isEditing: false,
    currentEventId: null,
    currentCalendarId: null,

    open(event = null) {
        const overlay = document.getElementById('eventModalOverlay');
        const titleEl = document.getElementById('eventModalTitle');
        const deleteBtn = document.getElementById('eventDeleteBtn');
        const form = document.getElementById('eventForm');

        if (!overlay || !form) return;

        // Reset form
        form.reset();
        document.getElementById('eventId').value = '';
        document.getElementById('eventCalendarId').value = '';

        // Set defaults
        const now = new Date();
        const startDate = now.toISOString().split('T')[0];
        const startHour = now.getHours();
        const startTime = `${String(startHour + 1).padStart(2, '0')}:00`;
        const endTime = `${String(startHour + 2).padStart(2, '0')}:00`;

        document.getElementById('eventStartDate').value = startDate;
        document.getElementById('eventEndDate').value = startDate;
        document.getElementById('eventStartTime').value = startTime;
        document.getElementById('eventEndTime').value = endTime;
        document.getElementById('eventBusy').checked = true;
        document.getElementById('eventAllDay').checked = false;
        this.toggleTimeInputs(false);

        if (event) {
            // Editing existing event
            this.isEditing = true;
            this.currentEventId = event.id;
            this.currentCalendarId = event.calendar_id || 'primary';
            titleEl.textContent = 'Edit Event';
            deleteBtn.classList.remove('hidden');

            // Populate form
            document.getElementById('eventId').value = event.id;
            document.getElementById('eventCalendarId').value = event.calendar_id || '';
            document.getElementById('eventTitle').value = event.title || '';
            document.getElementById('eventDescription').value = event.description || '';
            document.getElementById('eventLocation').value = event.location || '';
            document.getElementById('eventBusy').checked = event.busy !== false;
            document.getElementById('eventAllDay').checked = event.is_all_day || false;

            // Set dates/times
            if (event.start_time) {
                const start = new Date(event.start_time * 1000);
                const end = new Date(event.end_time * 1000);

                document.getElementById('eventStartDate').value = start.toISOString().split('T')[0];
                document.getElementById('eventEndDate').value = end.toISOString().split('T')[0];

                if (!event.is_all_day) {
                    document.getElementById('eventStartTime').value =
                        `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`;
                    document.getElementById('eventEndTime').value =
                        `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`;
                }
            }

            this.toggleTimeInputs(event.is_all_day);

            // Set participants
            if (event.participants && event.participants.length > 0) {
                const emails = event.participants.map(p => p.email).filter(Boolean).join(', ');
                document.getElementById('eventParticipants').value = emails;
            }
        } else {
            // Creating new event
            this.isEditing = false;
            this.currentEventId = null;
            this.currentCalendarId = 'primary';
            titleEl.textContent = 'New Event';
            deleteBtn.classList.add('hidden');
        }

        // Show modal
        overlay.classList.remove('hidden');
        this.isOpen = true;

        // Focus title input
        setTimeout(() => {
            document.getElementById('eventTitle').focus();
        }, 100);
    },

    close() {
        const overlay = document.getElementById('eventModalOverlay');
        if (overlay) {
            overlay.classList.add('hidden');
        }
        this.isOpen = false;
        this.isEditing = false;
        this.currentEventId = null;
    },

    toggleTimeInputs(allDay) {
        const startTime = document.getElementById('eventStartTime');
        const endTime = document.getElementById('eventEndTime');
        if (startTime) startTime.style.display = allDay ? 'none' : 'block';
        if (endTime) endTime.style.display = allDay ? 'none' : 'block';
    },

    getFormData() {
        const title = document.getElementById('eventTitle').value.trim();
        const description = document.getElementById('eventDescription').value.trim();
        const location = document.getElementById('eventLocation').value.trim();
        const isAllDay = document.getElementById('eventAllDay').checked;
        const busy = document.getElementById('eventBusy').checked;
        const participantsStr = document.getElementById('eventParticipants').value.trim();

        const startDate = document.getElementById('eventStartDate').value;
        const endDate = document.getElementById('eventEndDate').value;
        const startTime = document.getElementById('eventStartTime').value || '00:00';
        const endTime = document.getElementById('eventEndTime').value || '23:59';

        // Build timestamps
        let startTimestamp, endTimestamp;
        if (isAllDay) {
            startTimestamp = new Date(startDate + 'T00:00:00').getTime() / 1000;
            endTimestamp = new Date(endDate + 'T23:59:59').getTime() / 1000;
        } else {
            startTimestamp = new Date(`${startDate}T${startTime}:00`).getTime() / 1000;
            endTimestamp = new Date(`${endDate}T${endTime}:00`).getTime() / 1000;
        }

        // Parse participants
        const participants = participantsStr
            .split(',')
            .map(email => email.trim())
            .filter(email => email.length > 0)
            .map(email => ({ email }));

        return {
            title,
            description,
            location,
            start_time: startTimestamp,
            end_time: endTimestamp,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            is_all_day: isAllDay,
            busy,
            participants,
            calendar_id: this.currentCalendarId || 'primary'
        };
    },

    async save() {
        const data = this.getFormData();

        if (!data.title) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Title is required');
            }
            return;
        }

        const saveBtn = document.getElementById('eventSaveBtn');
        if (saveBtn) {
            saveBtn.disabled = true;
            saveBtn.textContent = 'Saving...';
        }

        try {
            let result;
            if (this.isEditing && this.currentEventId) {
                // Update event
                result = await AirAPI.updateEvent(this.currentEventId, data, data.calendar_id);
            } else {
                // Create event
                result = await AirAPI.createEvent(data);
            }

            if (result.success) {
                if (typeof showToast === 'function') {
                    showToast('success', 'Success', this.isEditing ? 'Event updated' : 'Event created');
                }
                this.close();
                // Reload events
                if (CalendarManager) {
                    CalendarManager.loadEvents();
                }
            } else {
                throw new Error(result.error || 'Failed to save event');
            }
        } catch (error) {
            console.error('Failed to save event:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', error.message || 'Failed to save event');
            }
        } finally {
            if (saveBtn) {
                saveBtn.disabled = false;
                saveBtn.textContent = 'Save';
            }
        }
    },

    async delete() {
        if (!this.currentEventId) return;

        if (!confirm('Are you sure you want to delete this event?')) {
            return;
        }

        const deleteBtn = document.getElementById('eventDeleteBtn');
        if (deleteBtn) {
            deleteBtn.disabled = true;
            deleteBtn.textContent = 'Deleting...';
        }

        try {
            const result = await AirAPI.deleteEvent(this.currentEventId, this.currentCalendarId);

            if (result.success) {
                if (typeof showToast === 'function') {
                    showToast('success', 'Success', 'Event deleted');
                }
                this.close();
                // Reload events
                if (CalendarManager) {
                    CalendarManager.loadEvents();
                }
            } else {
                throw new Error(result.error || 'Failed to delete event');
            }
        } catch (error) {
            console.error('Failed to delete event:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', error.message || 'Failed to delete event');
            }
        } finally {
            if (deleteBtn) {
                deleteBtn.disabled = false;
                deleteBtn.textContent = 'Delete';
            }
        }
    }
};

// Global functions for event modal (called from HTML)
function openEventModal(event = null) {
    EventModal.open(event);
}

function closeEventModal() {
    EventModal.close();
}

function toggleAllDay() {
    const isAllDay = document.getElementById('eventAllDay').checked;
    EventModal.toggleTimeInputs(isAllDay);
}

function saveEvent() {
    EventModal.save();
}

function deleteEvent() {
    EventModal.delete();
}

// Initialize when DOM is ready - but DON'T load data immediately
// Data will load when user switches to calendar view (lazy loading)
document.addEventListener('DOMContentLoaded', () => {
    // Set up event listeners but don't load data yet
    if (document.getElementById('calendarView')) {
        CalendarManager.setupEventListeners();
    }

    // Wire up "New Event" button in calendar sidebar
    const newEventBtn = document.querySelector('#calendarView .compose-btn');
    if (newEventBtn) {
        newEventBtn.onclick = () => openEventModal();
    }

    // Wire up event card clicks for editing
    document.addEventListener('click', (e) => {
        const eventCard = e.target.closest('.event-card');
        if (eventCard) {
            const eventId = eventCard.getAttribute('data-event-id');
            if (eventId && CalendarManager) {
                const event = CalendarManager.events.find(ev => ev.id === eventId);
                if (event) {
                    openEventModal(event);
                }
            }
        }
    });

    // Close modal on Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && EventModal.isOpen) {
            closeEventModal();
        }
    });
});
