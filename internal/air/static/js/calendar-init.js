/**
 * Calendar Initialization - DOM ready setup
 */
// Data will load when user switches to calendar view (lazy loading)
document.addEventListener('DOMContentLoaded', () => {
    // Set up event listeners but don't load data yet
    if (document.getElementById('calendarView')) {
        CalendarManager.setupEventListeners();
    }

    // Wire up "New Event" button in calendar sidebar
    const newEventBtn = document.querySelector('#calendarView .compose-btn');
    if (newEventBtn) {
        newEventBtn.onclick = () => openEventModal();
    }

    // Wire up event card clicks for editing
    document.addEventListener('click', (e) => {
        const eventCard = e.target.closest('.event-card');
        if (eventCard) {
            const eventId = eventCard.getAttribute('data-event-id');
            if (eventId && CalendarManager) {
                const event = CalendarManager.events.find(ev => ev.id === eventId);
                if (event) {
                    openEventModal(event);
                }
            }
        }
    });

    // Close modal on Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && EventModal.isOpen) {
            closeEventModal();
        }
    });
});
