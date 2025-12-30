/**
 * Calendar Data - Loading and navigation
 */
Object.assign(CalendarManager, {
async loadCalendars() {
    try {
        const data = await AirAPI.getCalendars();
        this.calendars = data.calendars || [];
        this.renderCalendarList();
    } catch (error) {
        console.error('Failed to load calendars:', error);
    }
},

async loadEvents() {
    if (this.isLoading) return;
    this.isLoading = true;

    try {
        const { start, end } = this.getDateRange();

        const data = await AirAPI.getEvents({
            start: Math.floor(start.getTime() / 1000),
            end: Math.floor(end.getTime() / 1000)
        });

        this.events = data.events || [];
        this.renderEvents();
    } catch (error) {
        console.error('Failed to load events:', error);
        if (typeof showToast === 'function') {
            showToast('error', 'Error', 'Failed to load events');
        }
    } finally {
        this.isLoading = false;
    }
},

getDateRange() {
    const now = this.currentDate;
    let start, end;

    switch (this.currentView) {
        case 'today':
            start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
            end = new Date(start);
            end.setDate(end.getDate() + 1);
            break;
        case 'week':
            start = new Date(now);
            start.setDate(start.getDate() - start.getDay()); // Start of week (Sunday)
            start.setHours(0, 0, 0, 0);
            end = new Date(start);
            end.setDate(end.getDate() + 7);
            break;
        case 'month':
            start = new Date(now.getFullYear(), now.getMonth(), 1);
            end = new Date(now.getFullYear(), now.getMonth() + 1, 1);
            break;
        case 'agenda':
        default:
            start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
            end = new Date(start);
            end.setDate(end.getDate() + 14); // 2 weeks
            break;
    }

    return { start, end };
},

navigate(direction) {
    switch (this.currentView) {
        case 'today':
            this.currentDate.setDate(this.currentDate.getDate() + direction);
            break;
        case 'week':
            this.currentDate.setDate(this.currentDate.getDate() + (direction * 7));
            break;
        case 'month':
            this.currentDate.setMonth(this.currentDate.getMonth() + direction);
            break;
        case 'agenda':
            this.currentDate.setDate(this.currentDate.getDate() + (direction * 14));
            break;
    }

    this.updateTitle();
    this.loadEvents();
},

goToToday() {
    this.currentDate = new Date();
    this.updateTitle();
    this.loadEvents();
},

setView(view) {
    this.currentView = view;

    // Update sidebar active state
    const folderItems = document.querySelectorAll('#calendarView .sidebar .folder-item');
    folderItems.forEach(item => {
        const text = item.textContent.trim().toLowerCase();
        const isActive = text.includes(view) ||
            (view === 'today' && text.includes('today')) ||
            (view === 'week' && text.includes('week')) ||
            (view === 'month' && text.includes('month')) ||
            (view === 'agenda' && text.includes('agenda'));
        item.classList.toggle('active', isActive);
    });

    this.updateTitle();
    this.loadEvents();
},

updateTitle() {
    const titleEl = document.querySelector('.calendar-title');
    if (!titleEl) return;

    const options = { month: 'long', year: 'numeric' };
    if (this.currentView === 'today') {
        options.day = 'numeric';
    }

    titleEl.textContent = this.currentDate.toLocaleDateString('en-US', options);
},
});
