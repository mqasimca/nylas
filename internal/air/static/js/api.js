// ====================================
// CORE API CLIENT
// Base HTTP client for Nylas Air UI
// With retry logic and rate limit handling
// ====================================

const AirAPI = {
    baseURL: '/api',

    // Rate limiting configuration
    _requestQueue: [],
    _isProcessingQueue: false,
    _minRequestInterval: 200, // Minimum 200ms between requests
    _lastRequestTime: 0,

    // Retry configuration
    _maxRetries: 3,
    _baseDelay: 1000, // 1 second base delay

    // Sleep utility
    _sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    },

    // Queue a request to prevent rate limiting
    async _queueRequest(requestFn) {
        return new Promise((resolve, reject) => {
            this._requestQueue.push({ requestFn, resolve, reject });
            this._processQueue();
        });
    },

    // Process queued requests with rate limiting
    async _processQueue() {
        if (this._isProcessingQueue || this._requestQueue.length === 0) {
            return;
        }

        this._isProcessingQueue = true;

        while (this._requestQueue.length > 0) {
            const { requestFn, resolve, reject } = this._requestQueue.shift();

            // Ensure minimum interval between requests
            const now = Date.now();
            const timeSinceLastRequest = now - this._lastRequestTime;
            if (timeSinceLastRequest < this._minRequestInterval) {
                await this._sleep(this._minRequestInterval - timeSinceLastRequest);
            }

            try {
                this._lastRequestTime = Date.now();
                const result = await requestFn();
                resolve(result);
            } catch (error) {
                reject(error);
            }
        }

        this._isProcessingQueue = false;
    },

    // Generic request method with retry and backoff
    async request(endpoint, options = {}) {
        const requestFn = async () => {
            return this._executeRequest(endpoint, options);
        };

        // Queue the request to prevent overwhelming the API
        return this._queueRequest(requestFn);
    },

    // Execute a single request with retry logic
    async _executeRequest(endpoint, options = {}, retryCount = 0) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);

            // Handle rate limiting (429)
            if (response.status === 429) {
                if (retryCount < this._maxRetries) {
                    const retryAfter = response.headers.get('Retry-After');
                    const delay = retryAfter
                        ? parseInt(retryAfter, 10) * 1000
                        : this._baseDelay * Math.pow(2, retryCount);

                    console.warn(`Rate limited on ${endpoint}, retrying in ${delay}ms...`);
                    await this._sleep(delay);
                    return this._executeRequest(endpoint, options, retryCount + 1);
                }
                throw new Error('Rate limit exceeded. Please try again later.');
            }

            // Handle server errors with retry
            if (response.status >= 500 && retryCount < this._maxRetries) {
                const delay = this._baseDelay * Math.pow(2, retryCount);
                console.warn(`Server error on ${endpoint}, retrying in ${delay}ms...`);
                await this._sleep(delay);
                return this._executeRequest(endpoint, options, retryCount + 1);
            }

            // Parse response
            let data;
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await response.json();
            } else {
                data = { message: await response.text() };
            }

            if (!response.ok) {
                throw new Error(data.error || `HTTP ${response.status}`);
            }

            return data;
        } catch (error) {
            // Retry on network errors
            if (error.name === 'TypeError' && retryCount < this._maxRetries) {
                const delay = this._baseDelay * Math.pow(2, retryCount);
                console.warn(`Network error on ${endpoint}, retrying in ${delay}ms...`);
                await this._sleep(delay);
                return this._executeRequest(endpoint, options, retryCount + 1);
            }

            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    },

    // ========================================
    // CONFIG & GRANTS
    // ========================================

    async getConfig() {
        return this.request('/config');
    },

    async getGrants() {
        return this.request('/grants');
    },

    async setDefaultGrant(grantId) {
        return this.request('/grants/default', {
            method: 'POST',
            body: JSON.stringify({ grant_id: grantId })
        });
    },

    // ========================================
    // FOLDERS
    // ========================================

    async getFolders() {
        return this.request('/folders');
    },

    // ========================================
    // EMAILS
    // ========================================

    async getEmails(options = {}) {
        const params = new URLSearchParams();
        if (options.folder) params.append('folder', options.folder);
        if (options.unread) params.append('unread', 'true');
        if (options.starred) params.append('starred', 'true');
        if (options.limit) params.append('limit', options.limit.toString());
        if (options.cursor) params.append('cursor', options.cursor);
        if (options.search) params.append('search', options.search);
        if (options.from) params.append('from', options.from);

        const queryString = params.toString();
        return this.request(`/emails${queryString ? '?' + queryString : ''}`);
    },

    async getEmail(id) {
        return this.request(`/emails/${id}`);
    },

    async updateEmail(id, updates) {
        return this.request(`/emails/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates)
        });
    },

    async deleteEmail(id) {
        return this.request(`/emails/${id}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // DRAFTS
    // ========================================

    async getDrafts() {
        return this.request('/drafts');
    },

    async createDraft(draft) {
        return this.request('/drafts', {
            method: 'POST',
            body: JSON.stringify(draft)
        });
    },

    async getDraft(id) {
        return this.request(`/drafts/${id}`);
    },

    async updateDraft(id, draft) {
        return this.request(`/drafts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(draft)
        });
    },

    async deleteDraft(id) {
        return this.request(`/drafts/${id}`, {
            method: 'DELETE'
        });
    },

    async sendDraft(id) {
        return this.request(`/drafts/${id}/send`, {
            method: 'POST'
        });
    },

    // ========================================
    // SEND MESSAGE (Direct Send)
    // ========================================

    async sendMessage(message) {
        return this.request('/send', {
            method: 'POST',
            body: JSON.stringify(message)
        });
    },

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
    // CONTACTS
    // ========================================

    async getContacts(options = {}) {
        const params = new URLSearchParams();
        if (options.email) params.append('email', options.email);
        if (options.source) params.append('source', options.source);
        if (options.group) params.append('group', options.group);
        if (options.limit) params.append('limit', options.limit.toString());
        if (options.cursor) params.append('cursor', options.cursor);

        const queryString = params.toString();
        return this.request(`/contacts${queryString ? '?' + queryString : ''}`);
    },

    async getContact(id) {
        return this.request(`/contacts/${id}`);
    },

    async createContact(contact) {
        return this.request('/contacts', {
            method: 'POST',
            body: JSON.stringify(contact)
        });
    },

    async updateContact(id, updates) {
        return this.request(`/contacts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates)
        });
    },

    async deleteContact(id) {
        return this.request(`/contacts/${id}`, {
            method: 'DELETE'
        });
    },

    async getContactGroups() {
        return this.request('/contact-groups');
    },

    // ========================================
    // PRODUCTIVITY - SPLIT INBOX
    // ========================================

    async getSplitInboxConfig() {
        return this.request('/inbox/split');
    },

    async updateSplitInboxConfig(config) {
        return this.request('/inbox/split', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
    },

    async categorizeEmail(emailId, from, subject, headers = {}) {
        return this.request('/inbox/categorize', {
            method: 'POST',
            body: JSON.stringify({ email_id: emailId, from, subject, headers })
        });
    },

    async getVIPSenders() {
        return this.request('/inbox/vip');
    },

    async addVIPSender(email) {
        return this.request('/inbox/vip', {
            method: 'POST',
            body: JSON.stringify({ email })
        });
    },

    async removeVIPSender(email) {
        return this.request(`/inbox/vip?email=${encodeURIComponent(email)}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // PRODUCTIVITY - SNOOZE
    // ========================================

    async getSnoozedEmails() {
        return this.request('/snooze');
    },

    async snoozeEmail(emailId, duration) {
        const body = { email_id: emailId };
        if (typeof duration === 'number') {
            body.snooze_until = duration;
        } else {
            body.duration = duration; // Natural language like "tomorrow", "1h", "next week"
        }
        return this.request('/snooze', {
            method: 'POST',
            body: JSON.stringify(body)
        });
    },

    async unsnoozeEmail(emailId) {
        return this.request(`/snooze?email_id=${encodeURIComponent(emailId)}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // PRODUCTIVITY - SCHEDULED SEND
    // ========================================

    async getScheduledMessages() {
        return this.request('/scheduled');
    },

    async scheduleMessage(message, sendAt) {
        const body = { ...message };
        if (typeof sendAt === 'number') {
            body.send_at = sendAt;
        } else {
            body.send_at_natural = sendAt; // Natural language
        }
        return this.request('/scheduled', {
            method: 'POST',
            body: JSON.stringify(body)
        });
    },

    async cancelScheduledMessage(scheduleId) {
        return this.request(`/scheduled?schedule_id=${encodeURIComponent(scheduleId)}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // PRODUCTIVITY - UNDO SEND
    // ========================================

    async getUndoSendConfig() {
        return this.request('/undo-send');
    },

    async updateUndoSendConfig(config) {
        return this.request('/undo-send', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
    },

    async sendWithUndo(message) {
        return this.request('/undo-send', {
            method: 'POST',
            body: JSON.stringify(message)
        });
    },

    async getPendingSends() {
        return this.request('/pending-sends');
    },

    async cancelPendingSend(messageId) {
        return this.request(`/pending-sends?message_id=${encodeURIComponent(messageId)}`, {
            method: 'DELETE'
        });
    },

    // ========================================
    // PRODUCTIVITY - EMAIL TEMPLATES
    // ========================================

    async getTemplates(category = '') {
        const params = category ? `?category=${encodeURIComponent(category)}` : '';
        return this.request(`/templates${params}`);
    },

    async createTemplate(template) {
        return this.request('/templates', {
            method: 'POST',
            body: JSON.stringify(template)
        });
    },

    async getTemplate(id) {
        return this.request(`/templates/${id}`);
    },

    async updateTemplate(id, template) {
        return this.request(`/templates/${id}`, {
            method: 'PUT',
            body: JSON.stringify(template)
        });
    },

    async deleteTemplate(id) {
        return this.request(`/templates/${id}`, {
            method: 'DELETE'
        });
    },

    async expandTemplate(id, variables = {}) {
        return this.request(`/templates/${id}/expand`, {
            method: 'POST',
            body: JSON.stringify({ variables })
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
};

// Make AirAPI available globally
window.AirAPI = AirAPI;

console.log('%c🔌 API client loaded (with rate limiting)', 'color: #22c55e;');
