/**
 * API Config - Configuration, grants, and folders
 */
Object.assign(AirAPI, {
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
    }
});
