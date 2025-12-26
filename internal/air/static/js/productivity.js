// ====================================
// PRODUCTIVITY MODULE
// Split Inbox, Snooze, Send Later, Undo Send, Templates
// ====================================

// =============================================================================
// SPLIT INBOX MANAGER
// =============================================================================

const SplitInboxManager = {
    config: null,
    categories: ['primary', 'vip', 'newsletters', 'updates', 'social', 'promotions'],
    currentCategory: 'primary',

    async init() {
        try {
            const response = await AirAPI.getSplitInboxConfig();
            this.config = response.config || response || {};
            this.setupCategoryTabs();
            console.log('%c📬 Split Inbox loaded', 'color: #22c55e;');
        } catch (error) {
            // Silently fail - split inbox config is optional
            console.log('%c📬 Split Inbox: using defaults', 'color: #a1a1aa;');
            this.config = { enabled: true };
        }
    },

    setupCategoryTabs() {
        const tabsContainer = document.querySelector('.filter-tabs');
        if (!tabsContainer) return;

        // Check if category tabs already exist
        if (tabsContainer.querySelector('[data-category]')) return;

        // Clear existing tabs and add category tabs
        tabsContainer.innerHTML = `
            <button class="filter-tab active" data-category="primary">Primary</button>
            <button class="filter-tab" data-category="vip">VIP</button>
            <button class="filter-tab" data-category="newsletters">Newsletters</button>
            <button class="filter-tab" data-category="updates">Updates</button>
        `;

        // Add click handlers
        tabsContainer.querySelectorAll('.filter-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                this.setCategory(tab.dataset.category);
            });
        });
    },

    async setCategory(category) {
        this.currentCategory = category;

        // Update tab UI
        document.querySelectorAll('.filter-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.category === category);
        });

        // Reload emails with category filter
        if (typeof EmailListManager !== 'undefined') {
            await this.loadCategorizedEmails(category);
        }
    },

    async loadCategorizedEmails(category) {
        // For VIP, use the existing filter
        if (category === 'vip') {
            if (typeof EmailListManager !== 'undefined') {
                EmailListManager.setFilter('vip');
            }
            return;
        }

        // For other categories, we need to categorize emails client-side
        if (typeof EmailListManager !== 'undefined') {
            EmailListManager.setFilter('all');

            // Filter displayed emails by category
            const emails = EmailListManager.emails;
            const filtered = [];

            for (const email of emails) {
                const from = email.from?.[0]?.email || '';
                const subject = email.subject || '';

                try {
                    const result = await AirAPI.categorizeEmail(email.id, from, subject);
                    if (result.category === category ||
                        (category === 'primary' && result.category === 'primary')) {
                        filtered.push(email);
                    }
                } catch (e) {
                    // On error, include in primary
                    if (category === 'primary') {
                        filtered.push(email);
                    }
                }
            }

            EmailListManager.filteredEmails = filtered;
            EmailListManager.renderEmails();
        }
    }
};

// =============================================================================
// SNOOZE MANAGER
// =============================================================================

const SnoozeManager = {
    snoozedEmails: [],

    async init() {
        this.setupSnoozePicker();
        await this.loadSnoozedEmails();
        console.log('%c⏰ Snooze module loaded', 'color: #22c55e;');
    },

    setupSnoozePicker() {
        const picker = document.getElementById('snoozePicker');
        if (!picker) return;

        // Update picker with actual functionality
        picker.innerHTML = `
            <div class="snooze-header">
                <span class="snooze-title">Snooze until...</span>
                <button class="snooze-close" onclick="SnoozeManager.hidePicker()">&times;</button>
            </div>
            <div class="snooze-options">
                <button class="snooze-option" data-duration="later">
                    <span class="snooze-icon">☀️</span>
                    <span class="snooze-label">Later today</span>
                    <span class="snooze-time">4:00 PM</span>
                </button>
                <button class="snooze-option" data-duration="tonight">
                    <span class="snooze-icon">🌙</span>
                    <span class="snooze-label">Tonight</span>
                    <span class="snooze-time">8:00 PM</span>
                </button>
                <button class="snooze-option" data-duration="tomorrow">
                    <span class="snooze-icon">📅</span>
                    <span class="snooze-label">Tomorrow</span>
                    <span class="snooze-time">9:00 AM</span>
                </button>
                <button class="snooze-option" data-duration="this weekend">
                    <span class="snooze-icon">🎉</span>
                    <span class="snooze-label">This weekend</span>
                    <span class="snooze-time">Saturday 9:00 AM</span>
                </button>
                <button class="snooze-option" data-duration="next week">
                    <span class="snooze-icon">📆</span>
                    <span class="snooze-label">Next week</span>
                    <span class="snooze-time">Monday 9:00 AM</span>
                </button>
                <button class="snooze-option" data-duration="next month">
                    <span class="snooze-icon">🗓️</span>
                    <span class="snooze-label">Next month</span>
                    <span class="snooze-time">1st of next month</span>
                </button>
            </div>
            <div class="snooze-custom">
                <input type="text" id="snoozeCustom" placeholder="Or type: 2h, 3d, next tuesday...">
                <button class="snooze-custom-btn" onclick="SnoozeManager.snoozeCustom()">Snooze</button>
            </div>
        `;

        // Add click handlers
        picker.querySelectorAll('.snooze-option').forEach(option => {
            option.addEventListener('click', () => {
                this.snoozeSelected(option.dataset.duration);
            });
        });

        // Handle Enter key in custom input
        const customInput = picker.querySelector('#snoozeCustom');
        if (customInput) {
            customInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.snoozeCustom();
                }
            });
        }
    },

    // Current email being snoozed
    currentEmailId: null,

    open() {
        this.showPicker();
    },

    close() {
        this.hidePicker();
    },

    openForEmail(emailId) {
        this.currentEmailId = emailId;
        this.showPicker();
    },

    showPicker() {
        const picker = document.getElementById('snoozePicker');
        if (picker) {
            picker.style.display = 'block';
            picker.classList.add('active');
        }
    },

    hidePicker() {
        const picker = document.getElementById('snoozePicker');
        if (picker) {
            picker.style.display = 'none';
            picker.classList.remove('active');
        }
        this.currentEmailId = null;
    },

    // Called from HTML template buttons
    async snooze(duration) {
        const emailId = this.currentEmailId || this.getSelectedEmailId();
        if (!emailId) {
            if (typeof showToast === 'function') {
                showToast('warning', 'No email selected', 'Select an email to snooze');
            }
            return;
        }
        await this.snoozeEmail(emailId, duration);
    },

    async snoozeSelected(duration) {
        const emailId = this.getSelectedEmailId();
        if (!emailId) {
            if (typeof showToast === 'function') {
                showToast('warning', 'No email selected', 'Select an email to snooze');
            }
            return;
        }

        await this.snoozeEmail(emailId, duration);
    },

    async snoozeCustom() {
        const input = document.getElementById('snoozeCustom');
        const duration = input?.value?.trim();

        if (!duration) {
            if (typeof showToast === 'function') {
                showToast('warning', 'Enter a time', 'Type a snooze time like "2h" or "tomorrow"');
            }
            return;
        }

        const emailId = this.getSelectedEmailId();
        if (!emailId) {
            if (typeof showToast === 'function') {
                showToast('warning', 'No email selected', 'Select an email to snooze');
            }
            return;
        }

        await this.snoozeEmail(emailId, duration);
        if (input) input.value = '';
    },

    async snoozeEmail(emailId, duration) {
        try {
            const result = await AirAPI.snoozeEmail(emailId, duration);
            this.hidePicker();

            if (typeof showToast === 'function') {
                showToast('success', 'Snoozed', result.message || 'Email snoozed');
            }

            // Remove from current view
            if (typeof EmailListManager !== 'undefined') {
                EmailListManager.emails = EmailListManager.emails.filter(e => e.id !== emailId);
                EmailListManager.applyFilter();
                EmailListManager.renderEmails();
            }
        } catch (error) {
            console.error('Failed to snooze email:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to snooze: ' + error.message);
            }
        }
    },

    async loadSnoozedEmails() {
        try {
            const result = await AirAPI.getSnoozedEmails();
            this.snoozedEmails = result.snoozed || result || [];
            if (!Array.isArray(this.snoozedEmails)) {
                this.snoozedEmails = [];
            }
        } catch (error) {
            // Silently fail - snoozed emails are optional
            this.snoozedEmails = [];
        }
    },

    getSelectedEmailId() {
        if (typeof EmailListManager !== 'undefined') {
            return EmailListManager.selectedEmailId;
        }
        return null;
    }
};

// =============================================================================
// SCHEDULED SEND MANAGER
// =============================================================================

const ScheduledSendManager = {
    scheduledMessages: [],
    dropdownOpen: false,

    async init() {
        await this.loadScheduledMessages();
        this.setupSendLaterButton();
        console.log('%c📤 Scheduled Send module loaded', 'color: #22c55e;');
    },

    toggleDropdown(event) {
        if (event) event.stopPropagation();
        const dropdown = document.querySelector('.send-dropdown');
        if (dropdown) {
            this.dropdownOpen = !this.dropdownOpen;
            dropdown.classList.toggle('open', this.dropdownOpen);
        }
    },

    closeDropdown() {
        const dropdown = document.querySelector('.send-dropdown');
        if (dropdown) {
            dropdown.classList.remove('open');
            this.dropdownOpen = false;
        }
    },

    async scheduleFromCompose() {
        this.closeDropdown();
        if (typeof ComposeManager === 'undefined') {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Compose not available');
            }
            return;
        }

        const data = ComposeManager.getFormData();
        if (!data.to || data.to.length === 0) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Please add at least one recipient');
            }
            return;
        }

        this.showSchedulePicker(data);
    },

    setupSendLaterButton() {
        // Add "Send Later" option to compose modal
        const sendBtn = document.getElementById('composeSend');
        if (!sendBtn) return;

        // Create dropdown container
        const container = sendBtn.parentElement;
        if (!container || container.querySelector('.send-dropdown')) return;

        // Wrap send button in dropdown
        const dropdown = document.createElement('div');
        dropdown.className = 'send-dropdown';
        dropdown.innerHTML = `
            <button class="send-dropdown-toggle" title="Send options">▼</button>
            <div class="send-dropdown-menu">
                <button class="send-option" data-action="send-later">
                    <span>📅</span> Schedule send...
                </button>
                <button class="send-option" data-action="send-tomorrow">
                    <span>☀️</span> Send tomorrow 9 AM
                </button>
                <button class="send-option" data-action="send-monday">
                    <span>📆</span> Send Monday 9 AM
                </button>
            </div>
        `;

        container.appendChild(dropdown);

        // Toggle dropdown
        dropdown.querySelector('.send-dropdown-toggle').addEventListener('click', (e) => {
            e.stopPropagation();
            dropdown.classList.toggle('open');
        });

        // Handle options
        dropdown.querySelectorAll('.send-option').forEach(option => {
            option.addEventListener('click', (e) => {
                e.stopPropagation();
                dropdown.classList.remove('open');
                this.handleSendOption(option.dataset.action);
            });
        });

        // Close on click outside
        document.addEventListener('click', () => {
            dropdown.classList.remove('open');
        });
    },

    async handleSendOption(action) {
        if (typeof ComposeManager === 'undefined') return;

        const data = ComposeManager.getFormData();
        if (data.to.length === 0) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Please add at least one recipient');
            }
            return;
        }

        let sendAt;
        switch (action) {
            case 'send-tomorrow':
                sendAt = 'tomorrow';
                break;
            case 'send-monday':
                sendAt = 'next monday';
                break;
            case 'send-later':
                this.showSchedulePicker(data);
                return;
            default:
                return;
        }

        await this.scheduleMessage(data, sendAt);
    },

    showSchedulePicker(data) {
        const picker = document.createElement('div');
        picker.className = 'schedule-picker-modal';
        picker.innerHTML = `
            <div class="schedule-picker">
                <h3>Schedule send</h3>
                <div class="schedule-options">
                    <input type="text" id="scheduleInput" placeholder="e.g., tomorrow 2pm, next friday, in 3 hours">
                </div>
                <div class="schedule-actions">
                    <button class="btn-secondary" onclick="this.closest('.schedule-picker-modal').remove()">Cancel</button>
                    <button class="btn-primary" id="scheduleConfirm">Schedule</button>
                </div>
            </div>
        `;

        document.body.appendChild(picker);

        const input = picker.querySelector('#scheduleInput');
        const confirmBtn = picker.querySelector('#scheduleConfirm');

        input.focus();

        confirmBtn.addEventListener('click', async () => {
            const sendAt = input.value.trim();
            if (!sendAt) {
                if (typeof showToast === 'function') {
                    showToast('warning', 'Enter a time', 'Please enter when to send');
                }
                return;
            }
            picker.remove();
            await this.scheduleMessage(data, sendAt);
        });

        input.addEventListener('keypress', async (e) => {
            if (e.key === 'Enter') {
                confirmBtn.click();
            }
        });
    },

    async scheduleMessage(messageData, sendAt) {
        try {
            const result = await AirAPI.scheduleMessage(messageData, sendAt);

            if (typeof showToast === 'function') {
                showToast('success', 'Scheduled', result.message || 'Message scheduled');
            }

            if (typeof ComposeManager !== 'undefined') {
                ComposeManager.close();
            }
        } catch (error) {
            console.error('Failed to schedule message:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to schedule: ' + error.message);
            }
        }
    },

    async loadScheduledMessages() {
        try {
            const result = await AirAPI.getScheduledMessages();
            this.scheduledMessages = result.scheduled || result || [];
            if (!Array.isArray(this.scheduledMessages)) {
                this.scheduledMessages = [];
            }
        } catch (error) {
            // Silently fail - scheduled messages are optional
            this.scheduledMessages = [];
        }
    }
};

// =============================================================================
// UNDO SEND MANAGER
// =============================================================================

const UndoSendManager = {
    config: { enabled: true, grace_period_sec: 10 },
    pendingSends: new Map(),
    currentPendingId: null,

    async init() {
        try {
            const response = await AirAPI.getUndoSendConfig();
            this.config = response.config || this.config;
        } catch (error) {
            console.error('Failed to load undo send config:', error);
        }
        console.log('%c↩️ Undo Send module loaded', 'color: #22c55e;');
    },

    // Called from HTML template button - send with undo from compose
    async sendWithUndoFromCompose() {
        // Close the dropdown first
        if (typeof ScheduledSendManager !== 'undefined') {
            ScheduledSendManager.closeDropdown();
        }

        if (typeof ComposeManager === 'undefined') {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Compose not available');
            }
            return;
        }

        const data = ComposeManager.getFormData();
        if (!data.to || data.to.length === 0) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Please add at least one recipient');
            }
            return;
        }

        try {
            await this.sendWithUndo(data);
            ComposeManager.close();
        } catch (error) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to send: ' + error.message);
            }
        }
    },

    // Undo the current pending send (called from undo toast button)
    async undo() {
        if (this.currentPendingId) {
            await this.cancelSend(this.currentPendingId);
        }
    },

    async sendWithUndo(messageData) {
        if (!this.config.enabled) {
            // Send immediately without undo
            return AirAPI.sendMessage(messageData);
        }

        try {
            const result = await AirAPI.sendWithUndo(messageData);

            if (result.pending_id) {
                this.showUndoToast(result.pending_id, result.send_at);
            }

            return result;
        } catch (error) {
            throw error;
        }
    },

    showUndoToast(pendingId, sendAt) {
        const gracePeriod = this.config.grace_period_sec || 10;
        let remaining = gracePeriod;

        // Store current pending ID for the undo() method
        this.currentPendingId = pendingId;

        // Use the existing toast from the HTML template if available
        let toast = document.getElementById('undoToast');
        if (!toast) {
            // Create undo toast
            toast = document.createElement('div');
            toast.className = 'undo-toast';
            toast.id = 'undoToast';
            toast.innerHTML = `
                <div class="undo-content">
                    <span class="undo-message">Message sent</span>
                    <span class="undo-timer">${remaining}s</span>
                </div>
                <button class="undo-btn" onclick="UndoSendManager.undo()">Undo</button>
            `;
            document.body.appendChild(toast);
        }

        // Update timer display
        const timerEl = toast.querySelector('.undo-timer') || toast.querySelector('#undoTimer');
        if (timerEl) timerEl.textContent = `${remaining}s`;

        setTimeout(() => toast.classList.add('active'), 10);

        // Store pending send
        this.pendingSends.set(pendingId, { toast, sendAt });

        // Countdown timer
        const interval = setInterval(() => {
            remaining--;
            if (timerEl) timerEl.textContent = `${remaining}s`;

            if (remaining <= 0) {
                clearInterval(interval);
                this.removePendingSend(pendingId, false);
            }
        }, 1000);

        // Store interval for cleanup
        this.pendingSends.get(pendingId).interval = interval;
    },

    async cancelSend(pendingId) {
        try {
            await AirAPI.cancelPendingSend(pendingId);
            this.removePendingSend(pendingId, true);

            if (typeof showToast === 'function') {
                showToast('info', 'Cancelled', 'Message cancelled');
            }

            // Reopen compose with the message content
            // (would need to store message data for this)
        } catch (error) {
            console.error('Failed to cancel send:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Could not cancel: ' + error.message);
            }
        }
    },

    removePendingSend(pendingId, cancelled) {
        const pending = this.pendingSends.get(pendingId);
        if (pending) {
            if (pending.interval) {
                clearInterval(pending.interval);
            }
            if (pending.toast) {
                pending.toast.classList.remove('active');
                // Don't remove the toast from DOM if it's the template one
                if (pending.toast.id !== 'undoToast') {
                    setTimeout(() => pending.toast.remove(), 300);
                }
            }
        }
        this.pendingSends.delete(pendingId);
        this.currentPendingId = null;
    }
};

// =============================================================================
// TEMPLATES MANAGER
// =============================================================================

const TemplatesManager = {
    templates: [],
    currentTemplate: null,
    currentVariables: {},

    async init() {
        await this.loadTemplates();
        this.setupTemplateButton();
        console.log('%c📋 Templates module loaded', 'color: #22c55e;');
    },

    // Methods called from HTML templates
    open() {
        this.showTemplatesPicker();
    },

    close() {
        const picker = document.getElementById('templatesPicker');
        if (picker) {
            picker.classList.remove('active');
        }
    },

    showCreate() {
        this.showCreateTemplate();
    },

    hideCreate() {
        const modal = document.getElementById('createTemplateModal');
        if (modal) {
            modal.classList.remove('active');
        }
    },

    filter(query) {
        const templatesList = document.getElementById('templatesList');
        if (!templatesList) return;

        const lowerQuery = (query || '').toLowerCase();
        templatesList.querySelectorAll('.template-item').forEach(item => {
            const name = item.querySelector('.template-name')?.textContent?.toLowerCase() || '';
            item.style.display = name.includes(lowerQuery) ? '' : 'none';
        });
    },

    cancelVariables() {
        const picker = document.getElementById('variablesPicker');
        if (picker) {
            picker.classList.remove('active');
        }
        this.currentTemplate = null;
        this.currentVariables = {};
    },

    applyVariables() {
        const picker = document.getElementById('variablesPicker');
        if (!picker || !this.currentTemplate) return;

        const variables = {};
        picker.querySelectorAll('[data-var]').forEach(input => {
            variables[input.dataset.var] = input.value || '';
        });

        // Expand template with variables
        const expanded = { ...this.currentTemplate };
        let body = expanded.body || '';
        let subject = expanded.subject || '';

        for (const [key, value] of Object.entries(variables)) {
            const regex = new RegExp(`{{${key}}}`, 'g');
            body = body.replace(regex, value);
            subject = subject.replace(regex, value);
        }

        expanded.body = body;
        expanded.subject = subject;

        this.applyTemplate(expanded);
        this.cancelVariables();
    },

    async save() {
        const nameInput = document.getElementById('templateName');
        const categorySelect = document.getElementById('templateCategory');
        const subjectInput = document.getElementById('templateSubject');
        const bodyInput = document.getElementById('templateBody');

        const name = nameInput?.value?.trim();
        const category = categorySelect?.value || '';
        const subject = subjectInput?.value?.trim() || '';
        const body = bodyInput?.value?.trim();

        if (!name || !body) {
            if (typeof showToast === 'function') {
                showToast('warning', 'Required', 'Name and body are required');
            }
            return;
        }

        try {
            const result = await AirAPI.createTemplate({ name, category, subject, body });
            this.templates.push(result.template || result);

            if (typeof showToast === 'function') {
                showToast('success', 'Created', 'Template saved');
            }

            this.hideCreate();
        } catch (error) {
            console.error('Failed to create template:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to save template');
            }
        }
    },

    setupTemplateButton() {
        // Add template button to compose toolbar
        const composeForm = document.getElementById('composeForm');
        if (!composeForm) return;

        const toolbar = composeForm.querySelector('.compose-toolbar');
        if (!toolbar || toolbar.querySelector('.template-btn')) return;

        const templateBtn = document.createElement('button');
        templateBtn.type = 'button';
        templateBtn.className = 'template-btn';
        templateBtn.innerHTML = '📋 Templates';
        templateBtn.title = 'Insert template';
        templateBtn.addEventListener('click', () => this.showTemplatesPicker());

        toolbar.insertBefore(templateBtn, toolbar.firstChild);
    },

    showTemplatesPicker() {
        // Use the existing picker from HTML template
        let picker = document.getElementById('templatesPicker');

        if (!picker) {
            // Fallback: create picker if not in template
            picker = document.createElement('div');
            picker.className = 'templates-picker';
            picker.id = 'templatesPicker';
            picker.innerHTML = `
                <div class="templates-header">
                    <h4>Email Templates</h4>
                    <button class="templates-close" onclick="TemplatesManager.close()">&times;</button>
                </div>
                <div class="templates-search">
                    <input type="text" id="templatesSearch" placeholder="Search templates...">
                </div>
                <div class="templates-list" id="templatesList"></div>
                <div class="templates-footer">
                    <button class="btn-primary" onclick="TemplatesManager.showCreate()">+ Create Template</button>
                </div>
            `;
            document.body.appendChild(picker);
        }

        // Populate templates list
        this.renderTemplatesList();

        // Show picker
        setTimeout(() => picker.classList.add('active'), 10);
    },

    renderTemplatesList() {
        const templatesList = document.getElementById('templatesList');
        if (!templatesList) return;

        if (this.templates.length === 0) {
            templatesList.innerHTML = `
                <div class="no-templates">
                    <p>No templates yet</p>
                    <p>Create your first template to speed up email composition</p>
                </div>
            `;
            return;
        }

        templatesList.innerHTML = this.templates.map(t => `
            <div class="template-item" data-id="${t.id}" onclick="TemplatesManager.insertTemplate('${t.id}')">
                <div class="template-name">${this.escapeHtml(t.name)}</div>
                <div class="template-preview">${this.escapeHtml((t.body || '').substring(0, 100))}${(t.body || '').length > 100 ? '...' : ''}</div>
            </div>
        `).join('');
    },

    async insertTemplate(templateId) {
        try {
            const template = this.templates.find(t => t.id === templateId);
            if (!template) return;

            // Close the templates picker
            this.close();

            // Extract variables from template body ({{variableName}} pattern)
            const variableMatches = (template.body || '').match(/\{\{(\w+)\}\}/g) || [];
            const variables = [...new Set(variableMatches.map(v => v.replace(/\{\{|\}\}/g, '')))];

            // Check for variables
            if (variables.length > 0) {
                template.variables = variables;
                this.promptForVariables(template);
            } else {
                this.applyTemplate(template);
            }
        } catch (error) {
            console.error('Failed to insert template:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Failed to insert template');
            }
        }
    },

    promptForVariables(template) {
        // Store current template for applyVariables method
        this.currentTemplate = template;

        // Use existing picker from HTML or create one
        let picker = document.getElementById('variablesPicker');

        if (!picker) {
            picker = document.createElement('div');
            picker.className = 'variables-picker';
            picker.id = 'variablesPicker';
            document.body.appendChild(picker);
        }

        // Populate variables fields
        const variablesFields = picker.querySelector('#variablesFields') || picker;
        variablesFields.innerHTML = `
            <div class="variables-form">
                <div class="variables-header">
                    <h4>Fill in variables for: ${this.escapeHtml(template.name)}</h4>
                </div>
                ${template.variables.map(v => `
                    <div class="variable-field">
                        <label>${this.escapeHtml(v)}</label>
                        <input type="text" data-var="${this.escapeHtml(v)}" placeholder="Enter ${this.escapeHtml(v)}">
                    </div>
                `).join('')}
                <div class="variables-actions">
                    <button class="btn-secondary" onclick="TemplatesManager.cancelVariables()">Cancel</button>
                    <button class="btn-primary" onclick="TemplatesManager.applyVariables()">Apply</button>
                </div>
            </div>
        `;

        setTimeout(() => picker.classList.add('active'), 10);

        // Focus first input
        const firstInput = picker.querySelector('input[data-var]');
        if (firstInput) firstInput.focus();
    },

    applyTemplate(template) {
        if (typeof ComposeManager === 'undefined') return;

        const els = ComposeManager.getElements();

        if (template.subject && els.subject) {
            els.subject.value = template.subject;
        }
        if (template.body && els.body) {
            // Append to existing content or replace
            const currentBody = els.body.value;
            els.body.value = currentBody ? currentBody + '\n\n' + template.body : template.body;
        }

        if (typeof showToast === 'function') {
            showToast('success', 'Template applied', template.name);
        }
    },

    showCreateTemplate() {
        // Close templates picker first
        this.close();

        // Use existing modal from HTML template
        let modal = document.getElementById('createTemplateModal');

        if (!modal) {
            // Fallback: create modal if not in template
            modal = document.createElement('div');
            modal.className = 'create-template-modal';
            modal.id = 'createTemplateModal';
            modal.innerHTML = `
                <div class="create-template">
                    <h3>Create Template</h3>
                    <form class="template-form" onsubmit="event.preventDefault(); TemplatesManager.save()">
                        <div class="form-field">
                            <label for="templateName">Template Name</label>
                            <input type="text" id="templateName" placeholder="e.g., Meeting Follow-up" required>
                        </div>
                        <div class="form-field">
                            <label for="templateCategory">Category</label>
                            <select id="templateCategory">
                                <option value="">No category</option>
                                <option value="work">Work</option>
                                <option value="personal">Personal</option>
                                <option value="sales">Sales</option>
                                <option value="support">Support</option>
                            </select>
                        </div>
                        <div class="form-field">
                            <label for="templateSubject">Subject Line</label>
                            <input type="text" id="templateSubject" placeholder="Use {{name}} for variables">
                        </div>
                        <div class="form-field">
                            <label for="templateBody">Body</label>
                            <textarea id="templateBody" placeholder="Hi {{name}},&#10;&#10;Use {{variable}} for dynamic content..." required></textarea>
                        </div>
                        <div class="template-actions">
                            <button type="button" class="btn-secondary" onclick="TemplatesManager.hideCreate()">Cancel</button>
                            <button type="submit" class="btn-primary">Save Template</button>
                        </div>
                    </form>
                </div>
            `;
            document.body.appendChild(modal);
        }

        // Clear form fields
        const nameInput = document.getElementById('templateName');
        const categorySelect = document.getElementById('templateCategory');
        const subjectInput = document.getElementById('templateSubject');
        const bodyInput = document.getElementById('templateBody');

        if (nameInput) nameInput.value = '';
        if (categorySelect) categorySelect.value = '';
        if (subjectInput) subjectInput.value = '';
        if (bodyInput) bodyInput.value = '';

        // Show modal
        setTimeout(() => modal.classList.add('active'), 10);

        // Focus name input
        if (nameInput) nameInput.focus();
    },

    async loadTemplates() {
        try {
            const result = await AirAPI.getTemplates();
            this.templates = result.templates || result || [];
            // Ensure templates is always an array
            if (!Array.isArray(this.templates)) {
                this.templates = [];
            }
        } catch (error) {
            // Silently fail - templates are optional
            console.log('%c📋 Templates: none available (this is OK)', 'color: #a1a1aa;');
            this.templates = [];
        }
    },

    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
};

// =============================================================================
// INITIALIZATION
// =============================================================================

document.addEventListener('DOMContentLoaded', () => {
    // Initialize all productivity modules
    setTimeout(() => {
        SplitInboxManager.init();
        SnoozeManager.init();
        ScheduledSendManager.init();
        UndoSendManager.init();
        TemplatesManager.init();
    }, 500); // Wait for other modules to load first
});

// Override the old snooze handler
window.showSnoozePicker = () => SnoozeManager.showPicker();
window.handleSnooze = (time) => SnoozeManager.snoozeSelected(time);

// Export managers globally
window.SplitInboxManager = SplitInboxManager;
window.SnoozeManager = SnoozeManager;
window.ScheduledSendManager = ScheduledSendManager;
window.UndoSendManager = UndoSendManager;
window.TemplatesManager = TemplatesManager;

console.log('%c🚀 Productivity module loaded', 'color: #22c55e;');
