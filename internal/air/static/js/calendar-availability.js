/**
 * Calendar Availability - Slots and rendering
 */
Object.assign(CalendarManager, {
availabilitySlots: [],

async loadAvailability(options = {}) {
    try {
        const now = Math.floor(Date.now() / 1000);
        const data = await AirAPI.getAvailability({
            start_time: options.start_time || now,
            end_time: options.end_time || (now + 7 * 24 * 60 * 60),
            duration_minutes: options.duration_minutes || 30,
            participants: options.participants || [],
            interval_minutes: options.interval_minutes || 15
        });
        this.availabilitySlots = data.slots || [];
        return this.availabilitySlots;
    } catch (error) {
        console.error('Failed to load availability:', error);
        return [];
    }
},

renderAvailabilitySlots(container) {
    if (!container) return;

    if (this.availabilitySlots.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">📅</div>
                <div class="empty-message">No available slots found</div>
            </div>
        `;
        return;
    }

    // Group slots by day
    const slotsByDay = {};
    this.availabilitySlots.forEach(slot => {
        const date = new Date(slot.start_time * 1000);
        const dayKey = date.toDateString();
        if (!slotsByDay[dayKey]) {
            slotsByDay[dayKey] = [];
        }
        slotsByDay[dayKey].push(slot);
    });

    container.innerHTML = Object.entries(slotsByDay).map(([day, slots]) => `
        <div class="availability-day">
            <div class="availability-day-header">${day}</div>
            <div class="availability-slots">
                ${slots.slice(0, 6).map(slot => `
                    <button class="availability-slot" data-start="${slot.start_time}" data-end="${slot.end_time}">
                        ${this.formatEventTime(slot.start_time)} - ${this.formatEventTime(slot.end_time)}
                    </button>
                `).join('')}
                ${slots.length > 6 ? `<span class="more-slots">+${slots.length - 6} more</span>` : ''}
            </div>
        </div>
    `).join('');
}
});
