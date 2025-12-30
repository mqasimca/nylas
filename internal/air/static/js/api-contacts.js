/**
 * API Contacts - Contact and contact group operations
 */
Object.assign(AirAPI, {
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
    }
});
