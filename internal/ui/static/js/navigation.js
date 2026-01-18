// =============================================================================
// Navigation
// =============================================================================

function initNavigation() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const page = item.dataset.page;
            switchPage(page);
        });
    });
}

function switchPage(page) {
    // Update nav items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.toggle('active', item.dataset.page === page);
    });

    // Update pages
    document.querySelectorAll('.page').forEach(p => {
        p.classList.toggle('active', p.id === `page-${page}`);
    });
}

// Keyboard Shortcuts
function initKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Ignore if typing in input/textarea
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

        // Number keys 1-4 to switch pages
        if (e.key === '1') { switchPage('overview'); return; }
        if (e.key === '2') { switchPage('auth'); return; }
        if (e.key === '3') { switchPage('email'); return; }
        if (e.key === '4') { switchPage('calendar'); return; }

        // Enter to run current command
        if (e.key === 'Enter') {
            const activePage = document.querySelector('.page.active');
            if (!activePage) return;

            if (activePage.id === 'page-auth' && currentAuthCmd) {
                runAuthCmd();
            } else if (activePage.id === 'page-email' && currentEmailCmd) {
                runEmailCmd();
            } else if (activePage.id === 'page-calendar' && currentCalendarCmd) {
                runCalendarCmd();
            }
            return;
        }

        // Escape to clear selection
        if (e.key === 'Escape') {
            clearSelection();
            return;
        }
    });
}

function clearSelection() {
    // Clear auth
    currentAuthCmd = '';
    document.querySelectorAll('#page-auth .cmd-item').forEach(i => i.classList.remove('active'));
    const authDetail = document.getElementById('auth-detail');
    if (authDetail) {
        authDetail.querySelector('.detail-placeholder').style.display = 'flex';
        authDetail.querySelector('.detail-content').style.display = 'none';
    }

    // Clear email
    currentEmailCmd = '';
    document.querySelectorAll('#page-email .cmd-item').forEach(i => i.classList.remove('active'));
    const emailDetail = document.getElementById('email-detail');
    if (emailDetail) {
        emailDetail.querySelector('.detail-placeholder').style.display = 'flex';
        emailDetail.querySelector('.detail-content').style.display = 'none';
    }

    // Clear calendar
    currentCalendarCmd = '';
    document.querySelectorAll('#page-calendar .cmd-item').forEach(i => i.classList.remove('active'));
    const calendarDetail = document.getElementById('calendar-detail');
    if (calendarDetail) {
        calendarDetail.querySelector('.detail-placeholder').style.display = 'flex';
        calendarDetail.querySelector('.detail-content').style.display = 'none';
    }
}
