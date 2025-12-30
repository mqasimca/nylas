/**
 * Calendar Events - Event modal and CRUD operations
 */
const EventModal = {
    isOpen: false,
    isEditing: false,
    currentEventId: null,
    currentCalendarId: null,

    open(event = null) {
        const overlay = document.getElementById('eventModalOverlay');
        const titleEl = document.getElementById('eventModalTitle');
        const deleteBtn = document.getElementById('eventDeleteBtn');
        const form = document.getElementById('eventForm');

        if (!overlay || !form) return;

        // Reset form
        form.reset();
        document.getElementById('eventId').value = '';
        document.getElementById('eventCalendarId').value = '';

        // Set defaults
        const now = new Date();
        const startDate = now.toISOString().split('T')[0];
        const startHour = now.getHours();
        const startTime = `${String(startHour + 1).padStart(2, '0')}:00`;
        const endTime = `${String(startHour + 2).padStart(2, '0')}:00`;

        document.getElementById('eventStartDate').value = startDate;
        document.getElementById('eventEndDate').value = startDate;
        document.getElementById('eventStartTime').value = startTime;
        document.getElementById('eventEndTime').value = endTime;
        document.getElementById('eventBusy').checked = true;
        document.getElementById('eventAllDay').checked = false;
        this.toggleTimeInputs(false);

        if (event) {
            // Editing existing event
            this.isEditing = true;
            this.currentEventId = event.id;
            this.currentCalendarId = event.calendar_id || 'primary';
            titleEl.textContent = 'Edit Event';
            deleteBtn.classList.remove('hidden');

            // Populate form
            document.getElementById('eventId').value = event.id;
            document.getElementById('eventCalendarId').value = event.calendar_id || '';
            document.getElementById('eventTitle').value = event.title || '';
            document.getElementById('eventDescription').value = event.description || '';
            document.getElementById('eventLocation').value = event.location || '';
            document.getElementById('eventBusy').checked = event.busy !== false;
            document.getElementById('eventAllDay').checked = event.is_all_day || false;

            // Set dates/times
            if (event.start_time) {
                const start = new Date(event.start_time * 1000);
                const end = new Date(event.end_time * 1000);

                document.getElementById('eventStartDate').value = start.toISOString().split('T')[0];
                document.getElementById('eventEndDate').value = end.toISOString().split('T')[0];

                if (!event.is_all_day) {
                    document.getElementById('eventStartTime').value =
                        `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`;
                    document.getElementById('eventEndTime').value =
                        `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`;
                }
            }

            this.toggleTimeInputs(event.is_all_day);

            // Set participants
            if (event.participants && event.participants.length > 0) {
                const emails = event.participants.map(p => p.email).filter(Boolean).join(', ');
                document.getElementById('eventParticipants').value = emails;
            }
        } else {
            // Creating new event
            this.isEditing = false;
            this.currentEventId = null;
            this.currentCalendarId = 'primary';
            titleEl.textContent = 'New Event';
            deleteBtn.classList.add('hidden');
        }

        // Show modal
        overlay.classList.remove('hidden');
        this.isOpen = true;

        // Focus title input
        setTimeout(() => {
            document.getElementById('eventTitle').focus();
        }, 100);
    },

    close() {
        const overlay = document.getElementById('eventModalOverlay');
        if (overlay) {
            overlay.classList.add('hidden');
        }
        this.isOpen = false;
        this.isEditing = false;
        this.currentEventId = null;
    },

    toggleTimeInputs(allDay) {
        const startTime = document.getElementById('eventStartTime');
        const endTime = document.getElementById('eventEndTime');
        if (startTime) startTime.style.display = allDay ? 'none' : 'block';
        if (endTime) endTime.style.display = allDay ? 'none' : 'block';
    },

    getFormData() {
        const title = document.getElementById('eventTitle').value.trim();
        const description = document.getElementById('eventDescription').value.trim();
        const location = document.getElementById('eventLocation').value.trim();
        const isAllDay = document.getElementById('eventAllDay').checked;
        const busy = document.getElementById('eventBusy').checked;
        const participantsStr = document.getElementById('eventParticipants').value.trim();

        const startDate = document.getElementById('eventStartDate').value;
        const endDate = document.getElementById('eventEndDate').value;
        const startTime = document.getElementById('eventStartTime').value || '00:00';
        const endTime = document.getElementById('eventEndTime').value || '23:59';

        // Build timestamps
        let startTimestamp, endTimestamp;
        if (isAllDay) {
            startTimestamp = new Date(startDate + 'T00:00:00').getTime() / 1000;
            endTimestamp = new Date(endDate + 'T23:59:59').getTime() / 1000;
        } else {
            startTimestamp = new Date(`${startDate}T${startTime}:00`).getTime() / 1000;
            endTimestamp = new Date(`${endDate}T${endTime}:00`).getTime() / 1000;
        }

        // Parse participants
        const participants = participantsStr
            .split(',')
            .map(email => email.trim())
            .filter(email => email.length > 0)
            .map(email => ({ email }));

        return {
            title,
            description,
            location,
            start_time: startTimestamp,
            end_time: endTimestamp,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            is_all_day: isAllDay,
            busy,
            participants,
            calendar_id: this.currentCalendarId || 'primary'
        };
    },

    async save() {
        const data = this.getFormData();

        if (!data.title) {
            if (typeof showToast === 'function') {
                showToast('error', 'Error', 'Title is required');
            }
            return;
        }

        const saveBtn = document.getElementById('eventSaveBtn');
        if (saveBtn) {
            saveBtn.disabled = true;
            saveBtn.textContent = 'Saving...';
        }

        try {
            let result;
            if (this.isEditing && this.currentEventId) {
                // Update event
                result = await AirAPI.updateEvent(this.currentEventId, data, data.calendar_id);
            } else {
                // Create event
                result = await AirAPI.createEvent(data);
            }

            if (result.success) {
                if (typeof showToast === 'function') {
                    showToast('success', 'Success', this.isEditing ? 'Event updated' : 'Event created');
                }
                this.close();
                // Reload events
                if (CalendarManager) {
                    CalendarManager.loadEvents();
                }
            } else {
                throw new Error(result.error || 'Failed to save event');
            }
        } catch (error) {
            console.error('Failed to save event:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', error.message || 'Failed to save event');
            }
        } finally {
            if (saveBtn) {
                saveBtn.disabled = false;
                saveBtn.textContent = 'Save';
            }
        }
    },

    async delete() {
        if (!this.currentEventId) return;

        if (!confirm('Are you sure you want to delete this event?')) {
            return;
        }

        const deleteBtn = document.getElementById('eventDeleteBtn');
        if (deleteBtn) {
            deleteBtn.disabled = true;
            deleteBtn.textContent = 'Deleting...';
        }

        try {
            const result = await AirAPI.deleteEvent(this.currentEventId, this.currentCalendarId);

            if (result.success) {
                if (typeof showToast === 'function') {
                    showToast('success', 'Success', 'Event deleted');
                }
                this.close();
                // Reload events
                if (CalendarManager) {
                    CalendarManager.loadEvents();
                }
            } else {
                throw new Error(result.error || 'Failed to delete event');
            }
        } catch (error) {
            console.error('Failed to delete event:', error);
            if (typeof showToast === 'function') {
                showToast('error', 'Error', error.message || 'Failed to delete event');
            }
        } finally {
            if (deleteBtn) {
                deleteBtn.disabled = false;
                deleteBtn.textContent = 'Delete';
            }
        }
    }
};

// Global functions for event modal (called from HTML)
function openEventModal(event = null) {
    EventModal.open(event);
}

function closeEventModal() {
    EventModal.close();
}

function toggleAllDay() {
    const isAllDay = document.getElementById('eventAllDay').checked;
    EventModal.toggleTimeInputs(isAllDay);
}

function saveEvent() {
    EventModal.save();
}

function deleteEvent() {
    EventModal.delete();
}
