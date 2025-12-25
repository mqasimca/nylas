// =============================================================================
// Output Formatting (ANSI parsing, table parsing)
// =============================================================================

// Table Parser - converts CLI table output to HTML table
function parseTable(text) {
    if (!text) return null;

    const lines = text.trim().split('\n');
    if (lines.length < 2) return null;

    // Check if first line looks like a header (has multiple words separated by spaces)
    const headerLine = lines[0].trim();

    // Common CLI table patterns - check for column headers
    const tablePatterns = [
        /^\s*(GRANT\s*ID|ID|EMAIL|NAME|SUBJECT|TITLE|CALENDAR)/i,
        /^\s*\w+\s{2,}\w+/  // At least two words separated by multiple spaces
    ];

    const looksLikeTable = tablePatterns.some(p => p.test(headerLine));
    if (!looksLikeTable) return null;

    // Parse header - split by 2+ spaces
    const headerParts = headerLine.split(/\s{2,}/).filter(h => h.trim());
    if (headerParts.length < 2) return null;

    // Find column positions based on header
    const colPositions = [];
    let pos = 0;
    for (const header of headerParts) {
        const idx = headerLine.indexOf(header, pos);
        colPositions.push(idx);
        pos = idx + header.length;
    }

    // Parse rows
    const rows = [];
    for (let i = 1; i < lines.length; i++) {
        const line = lines[i];
        if (!line.trim()) continue;

        const cells = [];
        for (let j = 0; j < colPositions.length; j++) {
            const start = colPositions[j];
            const end = j < colPositions.length - 1 ? colPositions[j + 1] : line.length;
            const cell = line.substring(start, end).trim();
            cells.push(cell);
        }

        if (cells.some(c => c)) {
            rows.push(cells);
        }
    }

    if (rows.length === 0) return null;

    // Build HTML table
    let html = '<table class="formatted-table"><thead><tr>';
    for (const header of headerParts) {
        html += `<th>${esc(header)}</th>`;
    }
    html += '</tr></thead><tbody>';

    for (const row of rows) {
        html += '<tr>';
        for (let i = 0; i < headerParts.length; i++) {
            const cell = row[i] || '';
            const headerLower = headerParts[i].toLowerCase();

            // Add special classes based on column type
            let cellClass = '';
            if (headerLower.includes('id') || headerLower.includes('grant')) {
                cellClass = 'cell-id';
            } else if (headerLower.includes('email')) {
                cellClass = 'cell-email';
            } else if (cell === '✓' || cell === '✔') {
                cellClass = 'cell-check';
            }

            html += `<td class="${cellClass}">${esc(cell)}</td>`;
        }
        html += '</tr>';
    }

    html += '</tbody></table>';
    return html;
}

// Format output - try table first, then ANSI
function formatOutput(text) {
    if (!text) return '';

    // Try to parse as table
    const tableHtml = parseTable(text);
    if (tableHtml) {
        return tableHtml;
    }

    // Fall back to ANSI parsing
    return parseAnsi(text);
}

// ANSI Color Parser
function parseAnsi(text) {
    if (!text) return '';

    // First escape HTML
    let html = esc(text);

    // ANSI escape code patterns
    const ansiMap = {
        // Reset
        '\\x1b\\[0m': '</span>',
        '\\x1b\\[m': '</span>',

        // Bold/Dim
        '\\x1b\\[1m': '<span class="ansi-bold">',
        '\\x1b\\[2m': '<span class="ansi-dim">',
        '\\x1b\\[4m': '<span class="ansi-underline">',

        // Foreground colors (standard)
        '\\x1b\\[30m': '<span class="ansi-gray">',
        '\\x1b\\[31m': '<span class="ansi-red">',
        '\\x1b\\[32m': '<span class="ansi-green">',
        '\\x1b\\[33m': '<span class="ansi-yellow">',
        '\\x1b\\[34m': '<span class="ansi-blue">',
        '\\x1b\\[35m': '<span class="ansi-magenta">',
        '\\x1b\\[36m': '<span class="ansi-cyan">',
        '\\x1b\\[37m': '<span class="ansi-white">',

        // Bright foreground colors
        '\\x1b\\[90m': '<span class="ansi-gray">',
        '\\x1b\\[91m': '<span class="ansi-red">',
        '\\x1b\\[92m': '<span class="ansi-green">',
        '\\x1b\\[93m': '<span class="ansi-yellow">',
        '\\x1b\\[94m': '<span class="ansi-blue">',
        '\\x1b\\[95m': '<span class="ansi-magenta">',
        '\\x1b\\[96m': '<span class="ansi-cyan">',
        '\\x1b\\[97m': '<span class="ansi-white">',

        // Bold + color combinations
        '\\x1b\\[1;32m': '<span class="ansi-bold ansi-green">',
        '\\x1b\\[1;31m': '<span class="ansi-bold ansi-red">',
        '\\x1b\\[1;33m': '<span class="ansi-bold ansi-yellow">',
        '\\x1b\\[1;34m': '<span class="ansi-bold ansi-blue">',
        '\\x1b\\[1;36m': '<span class="ansi-bold ansi-cyan">',
    };

    // Replace ANSI codes with HTML spans
    for (const [code, replacement] of Object.entries(ansiMap)) {
        html = html.replace(new RegExp(code, 'g'), replacement);
    }

    // Remove any remaining ANSI codes we didn't handle
    html = html.replace(/\x1b\[[0-9;]*m/g, '');

    // Also handle escaped versions that might come through
    html = html.replace(/\\x1b\[[0-9;]*m/g, '');
    html = html.replace(/&#x1b;\[[0-9;]*m/g, '');

    return html;
}
