/**
 * API Calendar - Calendar, event, availability, and conflict operations
 */
Object.assign(AirAPI, {
    // ========================================
    // CALENDARS
    // ========================================

    async getCalendars() {
        return this.request('/calendars');
    },

    // ========================================
    // EVENTS
    // ========================================

    async getEvents(options = {}) {
        const params = new URLSearchParams();
        if (options.calendarId) params.append('calendar_id', options.calendarId);
        if (options.start) params.append('start', options.start.toString());
        if (options.end) params.append('end', options.end.toString());
        if (options.limit) params.append('limit', options.limit.toString());

        const queryString = params.toString();
        return this.request(`/events${queryString ? '?' + queryString : ''}`);
    },

    async getEvent(id) {
        return this.request(`/events/${id}`);
    },

    async createEvent(event) {
        return this.request('/events', {
            method: 'POST',
            body: JSON.stringify(event)
        });
    },

    async updateEvent(id, updates, calendarId = 'primary') {
        const params = new URLSearchParams();
        if (calendarId) params.append('calendar_id', calendarId);
        const queryString = params.toString();
        return this.request(`/events/${id}${queryString ? '?' + queryString : ''}`, {
            method: 'PUT',
            body: JSON.stringify(updates)
        });
    },

    async deleteEvent(id, calendarId = 'primary') {
        const params = new URLSearchParams();
        if (calendarId) params.append('calendar_id', calendarId);
        const queryString = params.toString();
        return this.request(`/events/${id}${queryString ? '?' + queryString : ''}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // CALENDAR - AVAILABILITY & FIND TIME
    // ========================================

    /**
     * Find available time slots for scheduling meetings.
     * @param {Object} options - Availability options
     * @param {number} options.start_time - Start of search window (Unix timestamp)
     * @param {number} options.end_time - End of search window (Unix timestamp)
     * @param {number} options.duration_minutes - Meeting duration in minutes
     * @param {string[]} options.participants - Array of participant emails
     * @param {number} options.interval_minutes - Slot interval (default: 15)
     * @returns {Promise<{slots: Array<{start_time: number, end_time: number, emails: string[]}>}>}
     */
    async getAvailability(options = {}) {
        if (options.start_time && options.end_time) {
            // Use GET with query params for simple requests
            const params = new URLSearchParams();
            if (options.start_time) params.set('start_time', options.start_time);
            if (options.end_time) params.set('end_time', options.end_time);
            if (options.duration_minutes) params.set('duration_minutes', options.duration_minutes);
            if (options.participants?.length) params.set('participants', options.participants.join(','));
            if (options.interval_minutes) params.set('interval_minutes', options.interval_minutes);
            return this.request(`/availability?${params.toString()}`);
        }
        // Use POST for complex requests
        return this.request('/availability', {
            method: 'POST',
            body: JSON.stringify(options)
        });
    },

    /**
     * Find mutually available times between multiple participants.
     * Alias for getAvailability with required participants.
     * @param {string[]} participants - Array of participant emails
     * @param {number} durationMinutes - Meeting duration in minutes
     * @param {Object} options - Additional options
     * @returns {Promise<{slots: Array}>}
     */
    async findTime(participants, durationMinutes = 30, options = {}) {
        const now = Math.floor(Date.now() / 1000);
        return this.getAvailability({
            start_time: options.start_time || now,
            end_time: options.end_time || (now + 7 * 24 * 60 * 60), // Default: next 7 days
            duration_minutes: durationMinutes,
            participants: participants,
            interval_minutes: options.interval_minutes || 15
        });
    },

    /**
     * Get free/busy information for participants.
     * @param {string[]} emails - Array of participant emails
     * @param {number} startTime - Start time (Unix timestamp)
     * @param {number} endTime - End time (Unix timestamp)
     * @returns {Promise<{data: Array<{email: string, time_slots: Array}>}>}
     */
    async getFreeBusy(emails, startTime, endTime) {
        const params = new URLSearchParams();
        params.set('emails', emails.join(','));
        params.set('start_time', startTime);
        params.set('end_time', endTime);
        return this.request(`/freebusy?${params.toString()}`);
    },

    // ========================================
    // CALENDAR - CONFLICT DETECTION
    // ========================================

    /**
     * Detect scheduling conflicts in a time range.
     * @param {Object} options - Query options
     * @param {string} options.calendar_id - Calendar ID (default: 'primary')
     * @param {number} options.start_time - Start of range (Unix timestamp)
     * @param {number} options.end_time - End of range (Unix timestamp)
     * @returns {Promise<{conflicts: Array<{event1: Object, event2: Object}>}>}
     */
    async getConflicts(options = {}) {
        const params = new URLSearchParams();
        if (options.calendar_id) params.set('calendar_id', options.calendar_id);
        if (options.start_time) params.set('start_time', options.start_time);
        if (options.end_time) params.set('end_time', options.end_time);
        const queryString = params.toString();
        return this.request(`/events/conflicts${queryString ? '?' + queryString : ''}`);
    },

    /**
     * Check for conflicts in the current week.
     * @param {string} calendarId - Calendar ID (default: 'primary')
     * @returns {Promise<{conflicts: Array}>}
     */
    async getWeeklyConflicts(calendarId = 'primary') {
        const now = new Date();
        const dayOfWeek = now.getDay();
        const startOfWeek = new Date(now);
        startOfWeek.setDate(now.getDate() - dayOfWeek);
        startOfWeek.setHours(0, 0, 0, 0);
        const endOfWeek = new Date(startOfWeek);
        endOfWeek.setDate(startOfWeek.getDate() + 7);

        return this.getConflicts({
            calendar_id: calendarId,
            start_time: Math.floor(startOfWeek.getTime() / 1000),
            end_time: Math.floor(endOfWeek.getTime() / 1000)
        });
    }
});
