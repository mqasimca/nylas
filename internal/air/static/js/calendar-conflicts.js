/**
 * Calendar Conflicts - Detection and rendering
 */
Object.assign(CalendarManager, {
conflicts: [],
conflictsLoaded: false,

async loadConflicts() {
    try {
        const { start, end } = this.getDateRange();
        const data = await AirAPI.getConflicts({
            start_time: Math.floor(start.getTime() / 1000),
            end_time: Math.floor(end.getTime() / 1000)
        });
        this.conflicts = data.conflicts || [];
        this.conflictsLoaded = true;
        this.renderConflicts();
        return this.conflicts;
    } catch (error) {
        console.error('Failed to load conflicts:', error);
        return [];
    }
},

renderConflicts() {
    const container = document.getElementById('conflictsPanel');
    if (!container) return;

    if (this.conflicts.length === 0) {
        container.innerHTML = `
            <div class="conflicts-header">
                <span class="conflicts-icon">✅</span>
                <span>No scheduling conflicts</span>
            </div>
        `;
        return;
    }

    container.innerHTML = `
        <div class="conflicts-header warning">
            <span class="conflicts-icon">⚠️</span>
            <span>${this.conflicts.length} conflict${this.conflicts.length > 1 ? 's' : ''} detected</span>
        </div>
        <div class="conflicts-list">
            ${this.conflicts.map(c => this.renderConflictCard(c)).join('')}
        </div>
    `;
},

renderConflictCard(conflict) {
    const event1 = conflict.event1;
    const event2 = conflict.event2;
    return `
        <div class="conflict-card">
            <div class="conflict-event">
                <span class="conflict-time">${this.formatEventTime(event1.start_time)}</span>
                <span class="conflict-title">${this.escapeHtml(event1.title || '(No Title)')}</span>
            </div>
            <div class="conflict-overlap">↔️ overlaps with</div>
            <div class="conflict-event">
                <span class="conflict-time">${this.formatEventTime(event2.start_time)}</span>
                <span class="conflict-title">${this.escapeHtml(event2.title || '(No Title)')}</span>
            </div>
        </div>
    `;
},
});
