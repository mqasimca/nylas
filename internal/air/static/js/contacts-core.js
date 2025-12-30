/**
 * Contacts Core - State, initialization, and event listeners
 */
const ContactsManager = {
    contacts: [],
    groups: [],
    selectedContactId: null,
    selectedContact: null,
    currentGroup: null,
    searchQuery: '',
    hasMore: false,
    nextCursor: null,
    isLoading: false,
    isInitialized: false,
    editingContactId: null,

    async init() {
    // Set up event listeners
    this.setupEventListeners();

    // Load contacts when contacts tab is active
    const contactsTab = document.querySelector('[data-tab="contacts"]');
    if (contactsTab) {
        contactsTab.addEventListener('click', () => this.loadContacts());
    }

    console.log('%c👥 Contacts module loaded', 'color: #22c55e;');
},

setupEventListeners() {
    // Search input
    const searchInput = document.getElementById('contactSearch');
    if (searchInput) {
        searchInput.addEventListener('input', debounce(() => {
            this.searchQuery = searchInput.value.trim();
            this.loadContacts();
        }, 300));
    }

    // Add contact button
    const addBtn = document.getElementById('addContactBtn');
    if (addBtn) {
        addBtn.addEventListener('click', () => this.showCreateModal());
    }

    // Contact group filter
    const groupFilter = document.getElementById('contactGroupFilter');
    if (groupFilter) {
        groupFilter.addEventListener('change', () => {
            this.currentGroup = groupFilter.value || null;
            this.loadContacts();
        });
    }
}
};
