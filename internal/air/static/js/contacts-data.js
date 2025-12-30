/**
 * Contacts Data - Loading contacts and groups
 */
Object.assign(ContactsManager, {
async loadContacts(append = false) {
    // Skip if already loaded (unless appending for pagination)
    if (!append && this.isInitialized && this.contacts.length > 0) return;
    if (this.isLoading) return;
    this.isLoading = true;

    if (!append) {
        this.showLoadingState();
        this.contacts = [];
        this.nextCursor = null;
    }

    try {
        const options = { limit: 50 };
        if (this.currentGroup) options.group = this.currentGroup;
        if (this.searchQuery) options.email = this.searchQuery;
        if (this.nextCursor) options.cursor = this.nextCursor;

        const result = await AirAPI.getContacts(options);

        if (append) {
            this.contacts = [...this.contacts, ...result.contacts];
        } else {
            this.contacts = result.contacts || [];
        }

        this.hasMore = result.has_more;
        this.nextCursor = result.next_cursor;

        this.renderContacts();
        this.isInitialized = true;
    } catch (error) {
        console.error('Failed to load contacts:', error);
        this.showErrorState('Failed to load contacts');
    } finally {
        this.isLoading = false;
    }
},

async loadContactGroups() {
    try {
        const result = await AirAPI.getContactGroups();
        this.groups = result.groups || [];
        this.renderGroupFilter();
    } catch (error) {
        console.error('Failed to load contact groups:', error);
    }
},
});
