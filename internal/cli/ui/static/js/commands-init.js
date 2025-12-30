// =============================================================================
// Command System Initialization
// =============================================================================

// Initialize all command categories on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    renderAuthCommands();
    renderEmailCommands();
    renderCalendarCommands();
    renderContactsCommands();
    renderInboundCommands();
    renderSchedulerCommands();
    renderTimezoneCommands();
    renderWebhookCommands();
    renderOtpCommands();
    renderAdminCommands();
    renderNotetakerCommands();
});
