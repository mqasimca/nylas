// ====================================
// MOBILE MODULE
// ====================================

// Mobile state
let touchStartX = 0;
let touchStartY = 0;
let currentSwipeItem = null;
let pullStartY = 0;
let isPulling = false;

// Check if mobile
function isMobile() {
    return window.innerWidth <= 768;
}

// Toggle mobile sidebar
function toggleMobileSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.getElementById('sidebarOverlay');

    if (sidebar && overlay) {
        sidebar.classList.toggle('open');
        overlay.classList.toggle('active');
    }
}

// Close mobile sidebar
function closeMobileSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.getElementById('sidebarOverlay');

    if (sidebar && overlay) {
        sidebar.classList.remove('open');
        overlay.classList.remove('active');
    }
}

// Open email preview (mobile)
function openMobilePreview() {
    const preview = document.querySelector('.preview-pane');
    if (preview && isMobile()) {
        preview.classList.add('open');
    }
}

// Close email preview (mobile)
function closeMobilePreview() {
    const preview = document.querySelector('.preview-pane');
    if (preview) {
        preview.classList.remove('open');
    }
}

// Update mobile nav active state
function updateMobileNavActive(view) {
    document.querySelectorAll('.mobile-nav-item').forEach(item => {
        item.classList.remove('active');
    });

    const viewMap = {
        'email': 1,
        'calendar': 2,
        'contacts': 3
    };

    const index = viewMap[view];
    if (index !== undefined) {
        const items = document.querySelectorAll('.mobile-nav-item');
        items[index]?.classList.add('active');
    }
}

// Initialize swipe gestures
function initSwipeGestures() {
    const emailList = document.querySelector('.email-list');
    if (!emailList) return;

    emailList.addEventListener('touchstart', handleTouchStart, { passive: true });
    emailList.addEventListener('touchmove', handleTouchMove, { passive: false });
    emailList.addEventListener('touchend', handleTouchEnd, { passive: true });
}

function handleTouchStart(e) {
    const emailItem = e.target.closest('.email-item');
    if (!emailItem) return;

    touchStartX = e.touches[0].clientX;
    touchStartY = e.touches[0].clientY;
    currentSwipeItem = emailItem;
}

function handleTouchMove(e) {
    if (!currentSwipeItem) return;

    const touchX = e.touches[0].clientX;
    const touchY = e.touches[0].clientY;
    const diffX = touchX - touchStartX;
    const diffY = touchY - touchStartY;

    // If vertical scroll, don't handle swipe
    if (Math.abs(diffY) > Math.abs(diffX)) {
        currentSwipeItem = null;
        return;
    }

    // Prevent page scroll during swipe
    if (Math.abs(diffX) > 10) {
        e.preventDefault();
    }

    // Apply transform
    const maxSwipe = 100;
    const clampedDiff = Math.max(-maxSwipe, Math.min(maxSwipe, diffX));
    currentSwipeItem.style.transform = `translateX(${clampedDiff}px)`;

    // Show swipe indicator
    if (diffX > 50) {
        currentSwipeItem.classList.add('swipe-archive');
    } else if (diffX < -50) {
        currentSwipeItem.classList.add('swipe-delete');
    } else {
        currentSwipeItem.classList.remove('swipe-archive', 'swipe-delete');
    }
}

function handleTouchEnd(e) {
    if (!currentSwipeItem) return;

    const transform = currentSwipeItem.style.transform;
    const match = transform.match(/translateX\((-?\d+)px\)/);
    const swipeDistance = match ? parseInt(match[1]) : 0;

    if (swipeDistance > 80) {
        // Archive
        currentSwipeItem.style.transform = 'translateX(100%)';
        currentSwipeItem.style.opacity = '0';
        setTimeout(() => {
            showToast('success', 'Archived', 'Email moved to archive');
        }, 200);
    } else if (swipeDistance < -80) {
        // Delete
        currentSwipeItem.style.transform = 'translateX(-100%)';
        currentSwipeItem.style.opacity = '0';
        setTimeout(() => {
            showToast('warning', 'Deleted', 'Email moved to trash');
        }, 200);
    } else {
        // Reset
        currentSwipeItem.style.transform = '';
    }

    currentSwipeItem.classList.remove('swipe-archive', 'swipe-delete');
    currentSwipeItem = null;
}

// Initialize pull to refresh
function initPullToRefresh() {
    const emailList = document.querySelector('.email-list-container');
    if (!emailList) return;

    let pullIndicator = document.querySelector('.pull-to-refresh');
    if (!pullIndicator) {
        pullIndicator = document.createElement('div');
        pullIndicator.className = 'pull-to-refresh';
        pullIndicator.innerHTML = `
            <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="M23 4v6h-6M1 20v-6h6"/>
                <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
            </svg>
        `;
        emailList.style.position = 'relative';
        emailList.insertBefore(pullIndicator, emailList.firstChild);
    }

    emailList.addEventListener('touchstart', (e) => {
        if (emailList.scrollTop === 0) {
            pullStartY = e.touches[0].clientY;
            isPulling = true;
        }
    }, { passive: true });

    emailList.addEventListener('touchmove', (e) => {
        if (!isPulling) return;

        const pullY = e.touches[0].clientY;
        const pullDistance = pullY - pullStartY;

        if (pullDistance > 0 && pullDistance < 100) {
            pullIndicator.style.top = `${pullDistance - 50}px`;
            pullIndicator.style.opacity = pullDistance / 100;
        }

        if (pullDistance > 80) {
            pullIndicator.classList.add('active');
        }
    }, { passive: true });

    emailList.addEventListener('touchend', () => {
        if (pullIndicator.classList.contains('active')) {
            // Trigger refresh
            showToast('info', 'Refreshing', 'Checking for new emails...');
            setTimeout(() => {
                showToast('success', 'Updated', 'Inbox is up to date');
            }, 1500);
        }

        pullIndicator.style.top = '-50px';
        pullIndicator.style.opacity = '0';
        pullIndicator.classList.remove('active');
        isPulling = false;
    }, { passive: true });
}

// Handle orientation change
function handleOrientationChange() {
    // Close modals on orientation change
    closeMobileSidebar();
    closeMobilePreview();
}

// Handle resize
const handleResize = debounce(() => {
    if (!isMobile()) {
        closeMobileSidebar();
        closeMobilePreview();
    }
}, 250);

// Initialize mobile features
function initMobile() {
    if (!isMobile()) return;

    initSwipeGestures();
    initPullToRefresh();

    // Override showView to update mobile nav
    const originalShowView = window.showView;
    window.showView = function(view, event) {
        if (typeof originalShowView === 'function') {
            originalShowView(view, event);
        }
        updateMobileNavActive(view);
        closeMobileSidebar();
    };

    // Handle email item clicks on mobile
    document.querySelectorAll('.email-item').forEach(item => {
        item.addEventListener('click', () => {
            if (isMobile()) {
                openMobilePreview();
            }
        });
    });
}

// Event listeners
window.addEventListener('resize', handleResize);
window.addEventListener('orientationchange', handleOrientationChange);

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    initMobile();
});

console.log('%c📱 Mobile module loaded', 'color: #f59e0b;');
