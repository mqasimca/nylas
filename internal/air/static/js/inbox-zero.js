/**
 * Inbox Zero Celebration Module
 * Gamification and motivation features for email management
 * Celebrates achievements and tracks streaks
 */

const InboxZero = {
    // State
    state: {
        streak: 0,
        lastZeroDate: null,
        achievements: [],
        totalZeros: 0,
    },

    // Achievement definitions
    achievements: {
        firstZero: { id: 'firstZero', name: 'First Zero', icon: '🎯', desc: 'Reached Inbox Zero for the first time' },
        streak3: { id: 'streak3', name: 'Hat Trick', icon: '🎩', desc: '3-day Inbox Zero streak' },
        streak7: { id: 'streak7', name: 'Week Warrior', icon: '⚔️', desc: '7-day Inbox Zero streak' },
        streak30: { id: 'streak30', name: 'Monthly Master', icon: '👑', desc: '30-day Inbox Zero streak' },
        earlyBird: { id: 'earlyBird', name: 'Early Bird', icon: '🐦', desc: 'Inbox Zero before 9am' },
        speedster: { id: 'speedster', name: 'Speedster', icon: '⚡', desc: 'Cleared inbox in under 5 minutes' },
    },

    /**
     * Initialize inbox zero tracking
     */
    init() {
        this.loadState();
        this.setupListeners();
        console.log('%c🎯 Inbox Zero tracker initialized', 'color: #22c55e;');
    },

    /**
     * Load state from localStorage
     */
    loadState() {
        try {
            const saved = localStorage.getItem('inboxZeroState');
            if (saved) {
                const parsed = JSON.parse(saved);
                this.state = { ...this.state, ...parsed };
            }
        } catch (e) {
            console.error('Failed to load inbox zero state:', e);
        }
    },

    /**
     * Save state to localStorage
     */
    saveState() {
        try {
            localStorage.setItem('inboxZeroState', JSON.stringify(this.state));
        } catch (e) {
            console.error('Failed to save inbox zero state:', e);
        }
    },

    /**
     * Setup event listeners
     */
    setupListeners() {
        // Listen for email list updates
        document.addEventListener('emailsLoaded', (e) => {
            this.checkInboxStatus(e.detail.count);
        });
    },

    /**
     * Check if inbox is zero and trigger celebration
     * @param {number} emailCount - Current inbox email count
     */
    checkInboxStatus(emailCount) {
        if (emailCount === 0) {
            this.celebrate();
        }
    },

    /**
     * Trigger inbox zero celebration
     */
    celebrate() {
        const today = new Date().toDateString();

        // Check if already celebrated today
        if (this.state.lastZeroDate === today) {
            return;
        }

        // Update streak
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        const wasYesterday = this.state.lastZeroDate === yesterday.toDateString();

        if (wasYesterday) {
            this.state.streak++;
        } else {
            this.state.streak = 1;
        }

        this.state.lastZeroDate = today;
        this.state.totalZeros++;

        // Check for new achievements
        this.checkAchievements();

        // Show celebration UI
        this.showCelebration();

        // Save state
        this.saveState();
    },

    /**
     * Check and award achievements
     */
    checkAchievements() {
        const newAchievements = [];

        // First zero
        if (this.state.totalZeros === 1 && !this.hasAchievement('firstZero')) {
            newAchievements.push('firstZero');
        }

        // Streak achievements
        if (this.state.streak >= 3 && !this.hasAchievement('streak3')) {
            newAchievements.push('streak3');
        }
        if (this.state.streak >= 7 && !this.hasAchievement('streak7')) {
            newAchievements.push('streak7');
        }
        if (this.state.streak >= 30 && !this.hasAchievement('streak30')) {
            newAchievements.push('streak30');
        }

        // Early bird
        const hour = new Date().getHours();
        if (hour < 9 && !this.hasAchievement('earlyBird')) {
            newAchievements.push('earlyBird');
        }

        // Award new achievements
        newAchievements.forEach(id => {
            this.state.achievements.push(id);
            this.showAchievement(this.achievements[id]);
        });
    },

    /**
     * Check if user has achievement
     * @param {string} id - Achievement ID
     * @returns {boolean}
     */
    hasAchievement(id) {
        return this.state.achievements.includes(id);
    },

    /**
     * Show celebration UI
     */
    showCelebration() {
        // Trigger confetti
        if (typeof Confetti !== 'undefined') {
            Confetti.fire({
                particleCount: 100,
                spread: 70,
                origin: { y: 0.6 },
            });
        }

        // Show celebration modal
        this.showCelebrationModal();
    },

    /**
     * Show celebration modal
     */
    showCelebrationModal() {
        const modal = document.createElement('div');
        modal.className = 'inbox-zero-modal';
        modal.setAttribute('role', 'dialog');
        modal.setAttribute('aria-label', 'Inbox Zero Celebration');

        const content = document.createElement('div');
        content.className = 'inbox-zero-content';

        const icon = document.createElement('div');
        icon.className = 'inbox-zero-icon';
        icon.textContent = '🎉';
        content.appendChild(icon);

        const title = document.createElement('h2');
        title.textContent = 'Inbox Zero!';
        content.appendChild(title);

        const message = document.createElement('p');
        message.textContent = this.getCelebrationMessage();
        content.appendChild(message);

        if (this.state.streak > 1) {
            const streak = document.createElement('div');
            streak.className = 'inbox-zero-streak';
            streak.textContent = `🔥 ${this.state.streak} day streak!`;
            content.appendChild(streak);
        }

        const closeBtn = document.createElement('button');
        closeBtn.className = 'inbox-zero-close';
        closeBtn.textContent = 'Awesome!';
        closeBtn.addEventListener('click', () => modal.remove());
        content.appendChild(closeBtn);

        modal.appendChild(content);
        document.body.appendChild(modal);

        // Auto-close after 5 seconds
        setTimeout(() => {
            if (modal.parentNode) {
                modal.classList.add('fade-out');
                setTimeout(() => modal.remove(), 300);
            }
        }, 5000);
    },

    /**
     * Get celebration message based on streak
     * @returns {string}
     */
    getCelebrationMessage() {
        const messages = [
            "You're all caught up! Time for a coffee ☕",
            "Inbox conquered! You're on fire! 🔥",
            "Zero emails! Go do something fun! 🎮",
            "Clean inbox, clear mind! 🧘",
            "Productivity champion! 🏆",
        ];

        if (this.state.streak >= 7) {
            return "A week of inbox zero! You're unstoppable! 🚀";
        }
        if (this.state.streak >= 3) {
            return "Three days in a row! Keep it going! 💪";
        }

        return messages[Math.floor(Math.random() * messages.length)];
    },

    /**
     * Show achievement unlock
     * @param {Object} achievement - Achievement data
     */
    showAchievement(achievement) {
        const toast = document.createElement('div');
        toast.className = 'achievement-toast';

        const icon = document.createElement('span');
        icon.className = 'achievement-icon';
        icon.textContent = achievement.icon;
        toast.appendChild(icon);

        const info = document.createElement('div');
        info.className = 'achievement-info';

        const name = document.createElement('strong');
        name.textContent = achievement.name;
        info.appendChild(name);

        const desc = document.createElement('span');
        desc.textContent = achievement.desc;
        info.appendChild(desc);

        toast.appendChild(info);
        document.body.appendChild(toast);

        // Animate and remove
        setTimeout(() => toast.classList.add('show'), 100);
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 4000);
    },

    /**
     * Get current stats
     * @returns {Object}
     */
    getStats() {
        return {
            streak: this.state.streak,
            totalZeros: this.state.totalZeros,
            achievements: this.state.achievements.length,
            lastZero: this.state.lastZeroDate,
        };
    },
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    InboxZero.init();
});

// Export for use
if (typeof window !== 'undefined') {
    window.InboxZero = InboxZero;
}
