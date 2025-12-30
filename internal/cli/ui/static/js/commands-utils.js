// =============================================================================
// Command Utilities
// =============================================================================
// Shared utility functions for command rendering

/**
 * Generic function to render command sections.
 * Used by all command categories to render their command lists.
 *
 * @param {string} containerId - ID of the container element
 * @param {Array} sections - Array of command sections with title and commands
 * @param {string} showFn - Name of the show function to call on click
 */
function renderCommandSections(containerId, sections, showFn) {
    const container = document.getElementById(containerId);
    if (!container) return;

    let html = '';
    sections.forEach(section => {
        html += `<div class="cmd-section-title">${section.title}</div>`;
        Object.entries(section.commands).forEach(([key, data]) => {
            html += `
                <div class="cmd-item" data-cmd="${key}" onclick="${showFn}('${key}')">
                    <span class="cmd-name">${data.title}</span>
                    <button class="cmd-copy" onclick="event.stopPropagation(); copyText('nylas ${data.cmd}')">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="9" y="9" width="13" height="13" rx="2"/>
                            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                        </svg>
                    </button>
                </div>`;
        });
    });
    container.innerHTML = html;
}
