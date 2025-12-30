/**
 * API Productivity - Productivity features (split inbox, snooze, scheduled send, undo send, templates)
 */
Object.assign(AirAPI, {
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
    }
});
