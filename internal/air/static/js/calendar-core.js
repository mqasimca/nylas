/**
 * Calendar Core - State, initialization, and event listeners
 */
const CalendarManager = {
    calendars: [],
    events: [],
    currentDate: new Date(),
    currentView: 'week',
    selectedCalendarIds: [],
    isLoading: false,
    isInitialized: false,

    async init() {
    // Only initialize once
    if (this.isInitialized) {
        return;
    }
    this.isInitialized = true;
    // Set up event listeners first (UI is ready immediately)
    this.setupEventListeners();

    // Update title and render calendar grid with current date
    this.updateTitle();
    this.renderCalendarGrid();

    // Load calendars first, then events (sequential to avoid rate limits)
    try {
        await this.loadCalendars();
    } catch (error) {
        console.error('Failed to load calendars:', error);
        if (typeof showToast === 'function') {
            showToast('error', 'Error', 'Failed to load calendars. Will retry...');
        }
        // Retry after delay
        setTimeout(() => this.loadCalendars(), 3000);
    }

    try {
        await this.loadEvents();
    } catch (error) {
        console.error('Failed to load events:', error);
        if (typeof showToast === 'function') {
            showToast('error', 'Error', 'Failed to load events. Will retry...');
        }
        // Retry after delay
        setTimeout(() => this.loadEvents(), 3000);
    }

    // Load conflicts after events
    try {
        await this.loadConflicts();
    } catch (error) {
        console.error('Failed to load conflicts:', error);
    }

    console.log('%c📅 Calendar module loaded', 'color: #22c55e;');
},

setupEventListeners() {
    // View selection
    const calendarSidebar = document.querySelector('#calendarView .sidebar');
    if (calendarSidebar) {
        calendarSidebar.addEventListener('click', (e) => {
            const folderItem = e.target.closest('.folder-item');
            if (folderItem) {
                const text = folderItem.textContent.trim().toLowerCase();
                if (text.includes('today')) this.setView('today');
                else if (text.includes('week')) this.setView('week');
                else if (text.includes('month')) this.setView('month');
                else if (text.includes('agenda')) this.setView('agenda');
            }
        });
    }

    // Navigation buttons
    const navBtns = document.querySelectorAll('.calendar-nav-btn');
    if (navBtns.length >= 2) {
        navBtns[0].addEventListener('click', () => this.navigate(-1));
        navBtns[1].addEventListener('click', () => this.navigate(1));
    }

    // Today button
    const todayBtn = document.querySelector('.today-btn');
    if (todayBtn) {
        todayBtn.addEventListener('click', () => this.goToToday());
    }
}
};
