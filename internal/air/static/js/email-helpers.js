/* Email Helpers - Keyboard navigation and batch actions */

// ====================================

let currentEmailIndex = 0;
let selectedEmails = new Set();

// Virtual scrolling config
const ITEM_HEIGHT = 80;
const BUFFER_SIZE = 5;

// Get email items
function getEmailItems() {
    return document.querySelectorAll('.email-item');
}

// Select next email
function selectNextEmail() {
    const emailItems = getEmailItems();
    if (currentEmailIndex < emailItems.length - 1) {
        emailItems[currentEmailIndex]?.classList.remove('selected');
        currentEmailIndex++;
        emailItems[currentEmailIndex].classList.add('selected');
        emailItems[currentEmailIndex].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        if (typeof announce === 'function') {
            announce(`Email ${currentEmailIndex + 1} of ${emailItems.length}`);
        }
    }
}

// Select previous email
function selectPrevEmail() {
    const emailItems = getEmailItems();
    if (currentEmailIndex > 0) {
        emailItems[currentEmailIndex]?.classList.remove('selected');
        currentEmailIndex--;
        emailItems[currentEmailIndex].classList.add('selected');
        emailItems[currentEmailIndex].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        if (typeof announce === 'function') {
            announce(`Email ${currentEmailIndex + 1} of ${emailItems.length}`);
        }
    }
}

// Toggle email selection (for batch operations)
function toggleEmailSelection(index) {
    const emailItems = getEmailItems();
    if (selectedEmails.has(index)) {
        selectedEmails.delete(index);
        emailItems[index]?.classList.remove('batch-selected');
    } else {
        selectedEmails.add(index);
        emailItems[index]?.classList.add('batch-selected');
    }
    updateBatchActionsUI();
}

// Clear all selections
function clearEmailSelections() {
    const emailItems = getEmailItems();
    selectedEmails.clear();
    emailItems.forEach(item => item.classList.remove('batch-selected'));
    updateBatchActionsUI();
}

// Update batch actions UI
function updateBatchActionsUI() {
    const count = selectedEmails.size;
    const batchBar = document.getElementById('batchActionsBar');
    if (batchBar) {
        batchBar.style.display = count > 0 ? 'flex' : 'none';
        const countEl = batchBar.querySelector('.batch-count');
        if (countEl) countEl.textContent = `${count} selected`;
    }
}

// Archive selected emails
function archiveSelected() {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('success', 'Archived', `${selectedEmails.size} emails moved to archive`);
    }
    clearEmailSelections();
}

// Delete selected emails
function deleteSelected() {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('warning', 'Deleted', `${selectedEmails.size} emails moved to trash`);
    }
    clearEmailSelections();
}

// Mark selected as read/unread
function markSelectedAsRead(read = true) {
    if (selectedEmails.size === 0) return;
    if (typeof showToast === 'function') {
        showToast('info', read ? 'Marked Read' : 'Marked Unread', `${selectedEmails.size} emails updated`);
    }
    clearEmailSelections();
}

// Send email (legacy function for compatibility)
function sendEmail() {
    if (typeof showSendAnimation === 'function') {
        showSendAnimation();
    }
    setTimeout(() => {
        if (typeof showToast === 'function') {
            showToast('success', 'Email Sent', 'Your message has been delivered');
        }
        if (typeof toggleCompose === 'function') {
            toggleCompose();
        }
    }, 1000);
}

// Generate AI Summary (button-triggered)
function generateAISummary() {
    const btn = document.getElementById('aiSummaryBtn');
    const summaryDiv = document.getElementById('aiSummary');
    const summaryText = document.getElementById('aiSummaryText');

    if (!btn || !summaryDiv || !summaryText) return;

    // Show loading state
    btn.classList.add('loading');
    btn.innerHTML = '<span class="ai-icon">⏳</span><span>Generating summary...</span>';

    // Simulate AI processing
    setTimeout(() => {
        const summaries = [
            "This email discusses the Q4 product roadmap. Key points: 3 new features planned, design mockups attached, stakeholder meeting scheduled for next Tuesday.",
            "Sarah is requesting approval for the marketing budget. The proposal includes increased social media spend and a new influencer campaign targeting Gen Z.",
            "Technical review notes: Performance improvements show 40% faster load times. Minor bugs identified in mobile view need addressing before launch.",
            "Team sync summary: Sprint progress on track, two blockers identified for the auth module, new hire starting Monday needs onboarding setup."
        ];

        const randomSummary = summaries[Math.floor(Math.random() * summaries.length)];

        // Reset button
        btn.classList.remove('loading');
        btn.innerHTML = '<span class="ai-icon">✨</span><span>Summarize with AI</span>';

        // Show summary with typing effect
        summaryDiv.classList.remove('hidden');
        if (typeof showAITyping === 'function') {
            showAITyping(summaryText, randomSummary);
        } else {
            summaryText.textContent = randomSummary;
        }

    }, 1500 + Math.random() * 1000);
}

// Hide AI Summary
function hideAISummary() {
    const summaryDiv = document.getElementById('aiSummary');
    if (summaryDiv) {
        summaryDiv.classList.add('hidden');
    }
}

// Initialize email keyboard navigation
function initEmailKeyboard() {
    const emailList = document.querySelector('.email-list');
    if (!emailList) return;

    emailList.setAttribute('role', 'listbox');
    emailList.setAttribute('aria-label', 'Email messages');
    emailList.setAttribute('tabindex', '0');

    const items = emailList.querySelectorAll('.email-item');
    items.forEach((item, index) => {
        item.setAttribute('role', 'option');
        item.setAttribute('tabindex', '-1');
        item.setAttribute('aria-selected', item.classList.contains('selected') ? 'true' : 'false');
    });

    emailList.addEventListener('keydown', function(e) {
        const items = emailList.querySelectorAll('.email-item');
        const currentIndex = Array.from(items).findIndex(item =>
            item.classList.contains('focused') || item.classList.contains('selected')
        );

        switch(e.key) {
            case 'ArrowDown':
                e.preventDefault();
                if (currentIndex < items.length - 1) {
                    items[currentIndex]?.classList.remove('focused');
                    items[currentIndex + 1].classList.add('focused');
                    items[currentIndex + 1].focus();
                    if (typeof announce === 'function') {
                        announce(`Email ${currentIndex + 2} of ${items.length}`);
                    }
                }
                break;
            case 'ArrowUp':
                e.preventDefault();
                if (currentIndex > 0) {
                    items[currentIndex]?.classList.remove('focused');
                    items[currentIndex - 1].classList.add('focused');
                    items[currentIndex - 1].focus();
                    if (typeof announce === 'function') {
                        announce(`Email ${currentIndex} of ${items.length}`);
                    }
                }
                break;
            case 'Enter':
            case ' ':
                e.preventDefault();
                items[currentIndex]?.click();
                if (typeof announce === 'function') {
                    announce('Email opened');
                }
                break;
            case 'Delete':
            case 'Backspace':
                e.preventDefault();
                if (typeof announce === 'function') {
                    announce('Email deleted');
                }
                if (typeof showToast === 'function') {
                    showToast('warning', 'Deleted', 'Email moved to trash');
                }
                break;
            case 'x':
                e.preventDefault();
                toggleEmailSelection(currentIndex);
                break;
        }
    });
}

// Initialize email module
document.addEventListener('DOMContentLoaded', () => {
    // Init keyboard navigation
    initEmailKeyboard();

    // Init email list manager if we have the email list element
    if (document.querySelector('.email-list')) {
        EmailListManager.init();
    }
});
