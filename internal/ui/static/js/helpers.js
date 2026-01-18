// =============================================================================
// Helper Functions
// =============================================================================

function setText(id, text) {
    const el = document.getElementById(id);
    if (el) el.textContent = text;
}

function truncate(s, len = 16) {
    if (!s || s.length <= len) return s || '';
    return s.slice(0, len - 4) + '...' + s.slice(-4);
}

function truncateEmail(email) {
    if (!email) return '';
    if (email.length <= 18) return email;
    const [local, domain] = email.split('@');
    if (local.length > 10) {
        return local.slice(0, 8) + '...@' + domain;
    }
    return email;
}

function formatProvider(p) {
    const m = { google: 'Google', microsoft: 'Microsoft', imap: 'IMAP', ews: 'Exchange', icloud: 'iCloud' };
    return m[p?.toLowerCase()] || p || 'Unknown';
}

function esc(s) {
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

function copyOutput(section) {
    const el = document.getElementById(section + '-output');
    if (!el) return;

    const text = el.textContent || el.innerText;
    const btn = event.target.closest('.copy-output-btn');

    navigator.clipboard.writeText(text).then(() => {
        if (btn) {
            btn.classList.add('copied');
            btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg> Copied!';

            setTimeout(() => {
                btn.classList.remove('copied');
                btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg> Copy';
            }, 2000);
        }
        showToast('Output copied to clipboard', 'success');
    });
}

function copyText(text) {
    navigator.clipboard.writeText(text).then(() => {
        event.target.closest('.cmd-copy').classList.add('copied');
        setTimeout(() => {
            document.querySelectorAll('.cmd-copy.copied').forEach(el => el.classList.remove('copied'));
        }, 1000);
    });
}

function copyCmd(el) {
    const code = el.querySelector('code');
    if (!code) return;

    const text = code.textContent.trim();

    navigator.clipboard.writeText(text).then(() => {
        el.classList.add('copied');

        const originalText = code.textContent;
        code.textContent = 'Copied!';

        setTimeout(() => {
            el.classList.remove('copied');
            code.textContent = originalText;
        }, 1000);
    }).catch(() => {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);

        el.classList.add('copied');
        setTimeout(() => el.classList.remove('copied'), 1000);
    });
}
