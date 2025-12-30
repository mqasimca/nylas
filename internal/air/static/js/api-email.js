/**
 * API Email - Email, draft, and send operations
 */
Object.assign(AirAPI, {
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
    }
});
