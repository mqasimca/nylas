// =============================================================================
// Theme Management
// =============================================================================

function initTheme() {
    const theme = localStorage.getItem('theme') || 'dark';
    if (theme === 'light') document.documentElement.classList.add('light');
}

function toggleTheme() {
    document.documentElement.classList.toggle('light');
    const isLight = document.documentElement.classList.contains('light');
    localStorage.setItem('theme', isLight ? 'light' : 'dark');
}
