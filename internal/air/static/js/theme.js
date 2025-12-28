// ====================================
// THEME MANAGEMENT
// Handles dark/light/OLED mode with system preference detection
// ====================================

const ThemeManager = {
    // Available themes
    themes: ['dark', 'light', 'oled', 'system'],

    // Current theme
    currentTheme: 'system',

    // System preference media query
    systemPreferenceQuery: null,

    // Initialize theme system
    init() {
        // Set up system preference detection
        this.systemPreferenceQuery = window.matchMedia('(prefers-color-scheme: dark)');
        this.systemPreferenceQuery.addEventListener('change', (e) => {
            if (this.currentTheme === 'system') {
                this.applySystemTheme();
            }
        });

        // Load saved theme preference
        const savedTheme = localStorage.getItem('nylas-air-theme') || 'system';
        this.setTheme(savedTheme);

        // Set up keyboard shortcut (Ctrl/Cmd + Shift + T)
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'T') {
                e.preventDefault();
                this.cycleTheme();
            }
        });

        console.log('%c🎨 Theme system initialized', 'color: #8b5cf6;', { theme: savedTheme });
    },

    // Set theme
    setTheme(theme) {
        if (!this.themes.includes(theme)) {
            theme = 'system';
        }

        this.currentTheme = theme;
        localStorage.setItem('nylas-air-theme', theme);

        if (theme === 'system') {
            this.applySystemTheme();
        } else {
            this.applyTheme(theme);
        }

        // Update theme toggle UI if it exists
        this.updateToggleUI();

        // Dispatch custom event for other components
        window.dispatchEvent(new CustomEvent('themechange', { detail: { theme } }));
    },

    // Apply system theme based on preference
    applySystemTheme() {
        const prefersDark = this.systemPreferenceQuery.matches;
        this.applyTheme(prefersDark ? 'dark' : 'light');
    },

    // Apply specific theme to document
    applyTheme(theme) {
        const root = document.documentElement;

        // Remove all theme classes
        root.removeAttribute('data-theme');

        // Apply specific theme
        if (theme === 'oled') {
            root.setAttribute('data-theme', 'oled');
        } else if (theme === 'light') {
            root.setAttribute('data-theme', 'light');
        }
        // 'dark' is the default, no attribute needed

        // Update meta theme-color for mobile browsers
        const metaThemeColor = document.querySelector('meta[name="theme-color"]');
        if (metaThemeColor) {
            const colors = {
                dark: '#0a0a0c',
                light: '#ffffff',
                oled: '#000000'
            };
            metaThemeColor.content = colors[theme] || colors.dark;
        }
    },

    // Get current active theme (resolved if system)
    getActiveTheme() {
        if (this.currentTheme === 'system') {
            return this.systemPreferenceQuery.matches ? 'dark' : 'light';
        }
        return this.currentTheme;
    },

    // Cycle through themes
    cycleTheme() {
        const currentIndex = this.themes.indexOf(this.currentTheme);
        const nextIndex = (currentIndex + 1) % this.themes.length;
        this.setTheme(this.themes[nextIndex]);

        // Show toast notification
        if (typeof showToast === 'function') {
            const themeLabels = {
                dark: '🌙 Dark Mode',
                light: '☀️ Light Mode',
                oled: '🖤 OLED Black',
                system: '💻 System Theme'
            };
            showToast('info', 'Theme Changed', themeLabels[this.themes[nextIndex]]);
        }
    },

    // Update toggle button UI
    updateToggleUI() {
        const toggle = document.getElementById('themeToggle');
        if (!toggle) return;

        const icons = {
            dark: '🌙',
            light: '☀️',
            oled: '🖤',
            system: '💻'
        };

        const labels = {
            dark: 'Dark',
            light: 'Light',
            oled: 'OLED',
            system: 'Auto'
        };

        const icon = toggle.querySelector('.theme-icon');
        const label = toggle.querySelector('.theme-label');

        if (icon) icon.textContent = icons[this.currentTheme];
        if (label) label.textContent = labels[this.currentTheme];

        // Update active state on theme options
        document.querySelectorAll('.theme-option').forEach(option => {
            const optionTheme = option.dataset.theme;
            option.classList.toggle('active', optionTheme === this.currentTheme);
        });
    },

    // Create and inject theme toggle into header
    createThemeToggle(container) {
        if (!container) return;

        const toggleHTML = `
            <div class="theme-toggle-wrapper">
                <button id="themeToggle" class="theme-toggle" title="Change theme (Ctrl+Shift+T)">
                    <span class="theme-icon">💻</span>
                    <span class="theme-label">Auto</span>
                    <svg class="theme-dropdown-arrow" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="m6 9 6 6 6-6"/>
                    </svg>
                </button>
                <div class="theme-dropdown hidden">
                    <button class="theme-option" data-theme="system">
                        <span class="theme-option-icon">💻</span>
                        <span>System</span>
                        <span class="theme-option-desc">Follow OS preference</span>
                    </button>
                    <button class="theme-option" data-theme="dark">
                        <span class="theme-option-icon">🌙</span>
                        <span>Dark</span>
                        <span class="theme-option-desc">Obsidian theme</span>
                    </button>
                    <button class="theme-option" data-theme="light">
                        <span class="theme-option-icon">☀️</span>
                        <span>Light</span>
                        <span class="theme-option-desc">Clean & bright</span>
                    </button>
                    <button class="theme-option" data-theme="oled">
                        <span class="theme-option-icon">🖤</span>
                        <span>OLED Black</span>
                        <span class="theme-option-desc">True black for OLED</span>
                    </button>
                </div>
            </div>
        `;

        container.insertAdjacentHTML('beforeend', toggleHTML);

        // Set up event listeners
        const toggleBtn = document.getElementById('themeToggle');
        const dropdown = container.querySelector('.theme-dropdown');

        toggleBtn?.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdown?.classList.toggle('hidden');
        });

        // Theme option clicks
        container.querySelectorAll('.theme-option').forEach(option => {
            option.addEventListener('click', (e) => {
                e.stopPropagation();
                const theme = option.dataset.theme;
                this.setTheme(theme);
                dropdown?.classList.add('hidden');
            });
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', () => {
            dropdown?.classList.add('hidden');
        });

        // Update initial state
        this.updateToggleUI();
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();

    // Create theme toggle in header actions if container exists
    const headerActions = document.querySelector('.header-actions');
    if (headerActions) {
        ThemeManager.createThemeToggle(headerActions);
    }
});

// Export for use elsewhere
window.ThemeManager = ThemeManager;
