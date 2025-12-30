/* Email VIP - VIP sender management */
Object.assign(EmailListManager, {
async loadVIPSenders() {
    try {
        const response = await fetch('/api/inbox/vip');
        if (response.ok) {
            const data = await response.json();
            this.vipSenders = data.vip_senders || [];
            console.log('Loaded VIP senders:', this.vipSenders.length);
        }
    } catch (error) {
        console.error('Failed to load VIP senders:', error);
    }
},

// Set the current filter and update display
});
