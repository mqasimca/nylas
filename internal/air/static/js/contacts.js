// ====================================
// CONTACTS MODULE
// Handles contact list, details, and management
// ====================================

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
    },

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

    renderContacts() {
        const container = document.getElementById('contactsList');
        if (!container) return;

        if (this.contacts.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">👥</div>
                    <div class="empty-title">No contacts</div>
                    <div class="empty-message">
                        ${this.searchQuery ? 'No contacts match your search' : 'Your contacts will appear here'}
                    </div>
                </div>
            `;
            return;
        }

        container.innerHTML = this.contacts.map(contact => this.renderContactItem(contact)).join('');

        // Add load more button if there are more contacts
        if (this.hasMore) {
            container.innerHTML += `
                <div class="load-more">
                    <button class="btn btn-secondary" onclick="ContactsManager.loadContacts(true)">
                        Load More
                    </button>
                </div>
            `;
        }

        // Re-attach click handlers
        container.querySelectorAll('.contact-item').forEach(item => {
            item.addEventListener('click', () => {
                this.selectContact(item.dataset.contactId);
            });
        });
    },

    renderContactItem(contact) {
        const isSelected = contact.id === this.selectedContactId;
        const initials = this.getInitials(contact.display_name || contact.given_name || '?');
        const primaryEmail = contact.emails && contact.emails[0] ? contact.emails[0].email : '';
        const primaryPhone = contact.phone_numbers && contact.phone_numbers[0] ? contact.phone_numbers[0].number : '';

        return `
            <div class="contact-item ${isSelected ? 'selected' : ''}"
                 data-contact-id="${contact.id}">
                <div class="contact-avatar" style="background: ${this.getAvatarColor(contact.display_name || '')}">
                    ${contact.picture_url ?
                        `<img src="${contact.picture_url}" alt="${contact.display_name}" />` :
                        initials}
                </div>
                <div class="contact-info">
                    <div class="contact-name">${this.escapeHtml(contact.display_name || 'Unknown')}</div>
                    ${contact.job_title ? `<div class="contact-title">${this.escapeHtml(contact.job_title)}</div>` : ''}
                    ${primaryEmail ? `<div class="contact-email">${this.escapeHtml(primaryEmail)}</div>` : ''}
                </div>
                ${contact.company_name ?
                    `<div class="contact-company">${this.escapeHtml(contact.company_name)}</div>` : ''}
            </div>
        `;
    },

    renderGroupFilter() {
        const select = document.getElementById('contactGroupFilter');
        if (!select) return;

        select.innerHTML = '<option value="">All Contacts</option>';
        this.groups.forEach(group => {
            select.innerHTML += `<option value="${group.id}">${this.escapeHtml(group.name)}</option>`;
        });
    },

    async selectContact(contactId) {
        // Update selection state
        document.querySelectorAll('.contact-item').forEach(item => {
            item.classList.toggle('selected', item.dataset.contactId === contactId);
        });

        this.selectedContactId = contactId;

        try {
            const contact = await AirAPI.getContact(contactId);
            this.selectedContact = contact;
            this.renderContactDetail(contact);
        } catch (error) {
            console.error('Failed to load contact details:', error);
            this.showDetailError('Failed to load contact details');
        }
    },

    renderContactDetail(contact) {
        const container = document.getElementById('contactDetail');
        if (!container) return;

        const initials = this.getInitials(contact.display_name || contact.given_name || '?');
        const avatarColor = this.getAvatarColor(contact.display_name || '');

        container.innerHTML = `
            <div class="contact-detail-header">
                <div class="contact-detail-avatar" style="background: ${avatarColor}">
                    ${contact.picture_url ?
                        `<img src="${contact.picture_url}" alt="${contact.display_name}" />` :
                        initials}
                </div>
                <div class="contact-detail-name">${this.escapeHtml(contact.display_name || 'Unknown')}</div>
                ${contact.job_title ?
                    `<div class="contact-detail-title">${this.escapeHtml(contact.job_title)}${contact.company_name ? ` at ${this.escapeHtml(contact.company_name)}` : ''}</div>` :
                    (contact.company_name ? `<div class="contact-detail-title">${this.escapeHtml(contact.company_name)}</div>` : '')}
            </div>

            <div class="contact-detail-actions">
                <button class="action-btn primary" onclick="ContactsManager.emailContact('${contact.id}')">
                    <span>✉️</span> Email
                </button>
                <button class="action-btn" onclick="ContactsManager.editContact('${contact.id}')">
                    <span>✏️</span> Edit
                </button>
                <button class="action-btn" onclick="ContactsManager.deleteContact('${contact.id}')">
                    <span>🗑️</span> Delete
                </button>
            </div>

            <div class="contact-detail-sections">
                ${this.renderEmailSection(contact.emails)}
                ${this.renderPhoneSection(contact.phone_numbers)}
                ${this.renderAddressSection(contact.addresses)}
                ${this.renderNotesSection(contact.notes)}
                ${this.renderBirthdaySection(contact.birthday)}
            </div>
        `;
    },

    renderEmailSection(emails) {
        if (!emails || emails.length === 0) return '';

        return `
            <div class="contact-section">
                <div class="section-title">Email</div>
                ${emails.map(e => `
                    <div class="section-item">
                        <a href="mailto:${this.escapeHtml(e.email)}" class="section-value">${this.escapeHtml(e.email)}</a>
                        ${e.type ? `<span class="section-label">${this.escapeHtml(e.type)}</span>` : ''}
                    </div>
                `).join('')}
            </div>
        `;
    },

    renderPhoneSection(phones) {
        if (!phones || phones.length === 0) return '';

        return `
            <div class="contact-section">
                <div class="section-title">Phone</div>
                ${phones.map(p => `
                    <div class="section-item">
                        <a href="tel:${this.escapeHtml(p.number)}" class="section-value">${this.escapeHtml(p.number)}</a>
                        ${p.type ? `<span class="section-label">${this.escapeHtml(p.type)}</span>` : ''}
                    </div>
                `).join('')}
            </div>
        `;
    },

    renderAddressSection(addresses) {
        if (!addresses || addresses.length === 0) return '';

        return `
            <div class="contact-section">
                <div class="section-title">Address</div>
                ${addresses.map(a => {
                    const parts = [a.street_address, a.city, a.state, a.postal_code, a.country].filter(Boolean);
                    return `
                        <div class="section-item">
                            <div class="section-value">${parts.map(p => this.escapeHtml(p)).join(', ')}</div>
                            ${a.type ? `<span class="section-label">${this.escapeHtml(a.type)}</span>` : ''}
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    },

    renderNotesSection(notes) {
        if (!notes) return '';

        return `
            <div class="contact-section">
                <div class="section-title">Notes</div>
                <div class="section-item">
                    <div class="section-value notes">${this.escapeHtml(notes)}</div>
                </div>
            </div>
        `;
    },

    renderBirthdaySection(birthday) {
        if (!birthday) return '';

        return `
            <div class="contact-section">
                <div class="section-title">Birthday</div>
                <div class="section-item">
                    <div class="section-value">${this.escapeHtml(birthday)}</div>
                </div>
            </div>
        `;
    },

    showLoadingState() {
        const container = document.getElementById('contactsList');
        if (!container) return;

        container.innerHTML = `
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <div class="loading-text">Loading contacts...</div>
            </div>
        `;
    },

    showErrorState(message) {
        const container = document.getElementById('contactsList');
        if (!container) return;

        container.innerHTML = `
            <div class="error-state">
                <div class="error-icon">⚠️</div>
                <div class="error-message">${this.escapeHtml(message)}</div>
                <button class="btn btn-secondary" onclick="ContactsManager.loadContacts()">Retry</button>
            </div>
        `;
    },

    showDetailError(message) {
        const container = document.getElementById('contactDetail');
        if (!container) return;

        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">⚠️</div>
                <div class="empty-title">Error</div>
                <div class="empty-message">${this.escapeHtml(message)}</div>
            </div>
        `;
    },

    // Actions
    async emailContact(contactId) {
        const contact = this.contacts.find(c => c.id === contactId) || this.selectedContact;
        if (!contact || !contact.emails || contact.emails.length === 0) {
            if (typeof showToast === 'function') {
                showToast('warning', 'No Email', 'This contact has no email address');
            }
            return;
        }

        // Open compose with this contact's email
        if (typeof ComposeManager !== 'undefined') {
            ComposeManager.open('new');
            const els = ComposeManager.getElements();
            if (els.to) {
                els.to.value = contact.emails[0].email;
            }
        }
    },

    async editContact(contactId) {
        try {
            const contact = await AirAPI.getContact(contactId);
            this.editingContactId = contactId;

            document.getElementById('contactModalTitle').textContent = 'Edit Contact';
            document.getElementById('contactId').value = contactId;
            document.getElementById('contactGivenName').value = contact.given_name || '';
            document.getElementById('contactSurname').value = contact.surname || '';
            document.getElementById('contactNickname').value = contact.nickname || '';
            document.getElementById('contactCompany').value = contact.company_name || '';
            document.getElementById('contactJobTitle').value = contact.job_title || '';
            document.getElementById('contactNotes').value = contact.notes || '';
            document.getElementById('contactBirthday').value = contact.birthday || '';

            // Populate emails
            const emailsContainer = document.getElementById('contactEmails');
            if (contact.emails && contact.emails.length > 0) {
                emailsContainer.innerHTML = contact.emails.map((e, i) => `
                    <div class="contact-multi-row">
                        <input type="email" class="contact-input contact-email-input" placeholder="Email address" value="${this.escapeHtml(e.email || '')}">
                        <select class="contact-type-select">
                            <option value="personal" ${e.type === 'personal' ? 'selected' : ''}>Personal</option>
                            <option value="work" ${e.type === 'work' ? 'selected' : ''}>Work</option>
                            <option value="other" ${e.type === 'other' ? 'selected' : ''}>Other</option>
                        </select>
                        ${i === 0 ?
                            '<button type="button" class="contact-add-btn" onclick="ContactsManager.addEmailRow()">+</button>' :
                            '<button type="button" class="contact-remove-btn" onclick="this.parentElement.remove()">−</button>'}
                    </div>
                `).join('');
            } else {
                this.resetMultiInputs();
            }

            // Populate phones
            const phonesContainer = document.getElementById('contactPhones');
            if (contact.phone_numbers && contact.phone_numbers.length > 0) {
                phonesContainer.innerHTML = contact.phone_numbers.map((p, i) => `
                    <div class="contact-multi-row">
                        <input type="tel" class="contact-input contact-phone-input" placeholder="Phone number" value="${this.escapeHtml(p.number || '')}">
                        <select class="contact-type-select">
                            <option value="mobile" ${p.type === 'mobile' ? 'selected' : ''}>Mobile</option>
                            <option value="home" ${p.type === 'home' ? 'selected' : ''}>Home</option>
                            <option value="work" ${p.type === 'work' ? 'selected' : ''}>Work</option>
                            <option value="other" ${p.type === 'other' ? 'selected' : ''}>Other</option>
                        </select>
                        ${i === 0 ?
                            '<button type="button" class="contact-add-btn" onclick="ContactsManager.addPhoneRow()">+</button>' :
                            '<button type="button" class="contact-remove-btn" onclick="this.parentElement.remove()">−</button>'}
                    </div>
                `).join('');
            }

            document.getElementById('contactModalOverlay').classList.remove('hidden');
        } catch (error) {
            console.error('Failed to load contact for editing:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to load contact details');
            }
        }
    },

    async deleteContact(contactId) {
        if (!confirm('Are you sure you want to delete this contact?')) {
            return;
        }

        try {
            await AirAPI.deleteContact(contactId);

            if (typeof showToast === 'function') {
                showToast('success', 'Deleted', 'Contact deleted successfully');
            }

            // Clear selection if deleted contact was selected
            if (this.selectedContactId === contactId) {
                this.selectedContactId = null;
                this.selectedContact = null;
                const detailContainer = document.getElementById('contactDetail');
                if (detailContainer) {
                    detailContainer.innerHTML = `
                        <div class="empty-state">
                            <div class="empty-icon">👥</div>
                            <div class="empty-title">Select a contact</div>
                            <div class="empty-message">Choose a contact to view details</div>
                        </div>
                    `;
                }
            }

            // Reload contacts list
            await this.loadContacts();
        } catch (error) {
            console.error('Failed to delete contact:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to delete contact');
            }
        }
    },

    showCreateModal() {
        this.editingContactId = null;
        document.getElementById('contactModalTitle').textContent = 'New Contact';
        document.getElementById('contactForm').reset();
        document.getElementById('contactId').value = '';

        // Reset email/phone rows to single empty row
        this.resetMultiInputs();

        document.getElementById('contactModalOverlay').classList.remove('hidden');
    },

    closeModal() {
        document.getElementById('contactModalOverlay').classList.add('hidden');
        this.editingContactId = null;
    },

    resetMultiInputs() {
        // Reset emails to single row
        const emailsContainer = document.getElementById('contactEmails');
        emailsContainer.innerHTML = `
            <div class="contact-multi-row">
                <input type="email" class="contact-input contact-email-input" placeholder="Email address">
                <select class="contact-type-select">
                    <option value="personal">Personal</option>
                    <option value="work">Work</option>
                    <option value="other">Other</option>
                </select>
                <button type="button" class="contact-add-btn" onclick="ContactsManager.addEmailRow()">+</button>
            </div>
        `;

        // Reset phones to single row
        const phonesContainer = document.getElementById('contactPhones');
        phonesContainer.innerHTML = `
            <div class="contact-multi-row">
                <input type="tel" class="contact-input contact-phone-input" placeholder="Phone number">
                <select class="contact-type-select">
                    <option value="mobile">Mobile</option>
                    <option value="home">Home</option>
                    <option value="work">Work</option>
                    <option value="other">Other</option>
                </select>
                <button type="button" class="contact-add-btn" onclick="ContactsManager.addPhoneRow()">+</button>
            </div>
        `;
    },

    addEmailRow() {
        const container = document.getElementById('contactEmails');
        const row = document.createElement('div');
        row.className = 'contact-multi-row';
        row.innerHTML = `
            <input type="email" class="contact-input contact-email-input" placeholder="Email address">
            <select class="contact-type-select">
                <option value="personal">Personal</option>
                <option value="work">Work</option>
                <option value="other">Other</option>
            </select>
            <button type="button" class="contact-remove-btn" onclick="this.parentElement.remove()">−</button>
        `;
        container.appendChild(row);
    },

    addPhoneRow() {
        const container = document.getElementById('contactPhones');
        const row = document.createElement('div');
        row.className = 'contact-multi-row';
        row.innerHTML = `
            <input type="tel" class="contact-input contact-phone-input" placeholder="Phone number">
            <select class="contact-type-select">
                <option value="mobile">Mobile</option>
                <option value="home">Home</option>
                <option value="work">Work</option>
                <option value="other">Other</option>
            </select>
            <button type="button" class="contact-remove-btn" onclick="this.parentElement.remove()">−</button>
        `;
        container.appendChild(row);
    },

    async saveContact() {
        const contactId = document.getElementById('contactId').value;

        // Collect form data
        const contact = {
            given_name: document.getElementById('contactGivenName').value.trim(),
            surname: document.getElementById('contactSurname').value.trim(),
            nickname: document.getElementById('contactNickname').value.trim(),
            company_name: document.getElementById('contactCompany').value.trim(),
            job_title: document.getElementById('contactJobTitle').value.trim(),
            notes: document.getElementById('contactNotes').value.trim(),
            birthday: document.getElementById('contactBirthday').value || null,
            emails: [],
            phone_numbers: []
        };

        // Collect emails
        document.querySelectorAll('#contactEmails .contact-multi-row').forEach(row => {
            const email = row.querySelector('.contact-email-input').value.trim();
            const type = row.querySelector('.contact-type-select').value;
            if (email) {
                contact.emails.push({ email, type });
            }
        });

        // Collect phone numbers
        document.querySelectorAll('#contactPhones .contact-multi-row').forEach(row => {
            const number = row.querySelector('.contact-phone-input').value.trim();
            const type = row.querySelector('.contact-type-select').value;
            if (number) {
                contact.phone_numbers.push({ number, type });
            }
        });

        // Validate - need at least a name or email
        if (!contact.given_name && !contact.surname && contact.emails.length === 0) {
            if (typeof showToast === 'function') {
                showToast('warning', 'Required', 'Please enter a name or email address');
            }
            return;
        }

        try {
            if (contactId) {
                // Update existing contact
                await AirAPI.updateContact(contactId, contact);
                if (typeof showToast === 'function') {
                    showToast('success', 'Updated', 'Contact updated successfully');
                }
            } else {
                // Create new contact
                await AirAPI.createContact(contact);
                if (typeof showToast === 'function') {
                    showToast('success', 'Created', 'Contact created successfully');
                }
            }

            this.closeModal();
            this.isInitialized = false; // Force reload
            await this.loadContacts();
        } catch (error) {
            console.error('Failed to save contact:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', error.message || 'Failed to save contact');
            }
        }
    },

    // Utility functions
    getInitials(name) {
        if (!name) return '?';
        const parts = name.split(' ').filter(Boolean);
        if (parts.length >= 2) {
            return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    },

    getAvatarColor(name) {
        const colors = [
            '#e11d48', '#db2777', '#c026d3', '#9333ea',
            '#7c3aed', '#6366f1', '#2563eb', '#0284c7',
            '#0891b2', '#059669', '#16a34a', '#65a30d',
            '#ca8a04', '#ea580c', '#dc2626'
        ];
        let hash = 0;
        for (let i = 0; i < name.length; i++) {
            hash = name.charCodeAt(i) + ((hash << 5) - hash);
        }
        return colors[Math.abs(hash) % colors.length];
    },

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// Debounce utility
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    if (document.getElementById('contactsList') || document.querySelector('[data-tab="contacts"]')) {
        ContactsManager.init();
    }
});
